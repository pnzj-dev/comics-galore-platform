package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"net/http"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

const oauthStateTTL = 10 * time.Minute

// generatePKCE returns a code_verifier and its code_challenge (S256).
func generatePKCE() (verifier, challenge string) {
	b := make([]byte, 32)
	rand.Read(b)
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge
}

func storeOAuthState(ctx context.Context, state, provider, verifier, linkUserID string) error {
	var uid interface{}
	if linkUserID != "" {
		uid = linkUserID
	}
	_, err := db.Exec(ctx, `
		INSERT INTO oauth_states (state, provider, code_verifier, link_user_id, expires_at)
		VALUES ($1, $2, $3, $4, now() + $5::interval)
	`, state, provider, verifier, uid, oauthStateTTL.String())
	return err
}

func consumeOAuthState(ctx context.Context, state, provider string) (verifier, linkUserID string, err error) {
	var v string
	var link interface{}
	var expiresAt time.Time
	err = db.QueryRow(ctx, `
		SELECT code_verifier, link_user_id, expires_at FROM oauth_states
		WHERE state = $1 AND provider = $2
	`, state, provider).Scan(&v, &link, &expiresAt)
	if err != nil {
		if isNoRows(err) {
			return "", "", &errs.Error{Code: errs.InvalidArgument, Message: "invalid or expired state"}
		}
		return "", "", err
	}
	if time.Now().After(expiresAt) {
		_, _ = db.Exec(ctx, `DELETE FROM oauth_states WHERE state = $1`, state)
		return "", "", &errs.Error{Code: errs.InvalidArgument, Message: "state expired"}
	}
	// Single-use.
	if _, err := db.Exec(ctx, `DELETE FROM oauth_states WHERE state = $1`, state); err != nil {
		return "", "", err
	}
	if l, ok := link.(string); ok {
		linkUserID = l
	}
	return v, linkUserID, nil
}

// ----- Start (raw redirect to provider) -----

//encore:api public raw method=GET path=/auth/oauth/:provider
func OAuthStart(w http.ResponseWriter, req *http.Request) {
	provider := req.PathValue("provider")
	if !isKnownProvider(provider) {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}

	state := randomToken(32)
	verifier, _ := generatePKCE()

	// If already logged in, mark intent to link (account linking flow).
	linkUserID := ""
	if data, ok := auth.Data().(*AuthData); ok {
		linkUserID = data.UserID
	}

	if err := storeOAuthState(req.Context(), state, provider, verifier, linkUserID); err != nil {
		http.Error(w, "could not start login", http.StatusInternalServerError)
		return
	}

	prov := oauthProviders[provider]
	url := prov.authURL(state, redirectURI(req.Host, provider), verifier)
	http.Redirect(w, req, url, http.StatusFound)
}

// ----- Callback (raw, provider redirects here) -----

//encore:api public raw method=GET path=/auth/oauth/:provider/callback
func OAuthCallback(w http.ResponseWriter, req *http.Request) {
	provider := req.PathValue("provider")
	if !isKnownProvider(provider) {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}

	q := req.URL.Query()
	if errMsg := q.Get("error"); errMsg != "" {
		redirectError(w, req, "oauth_cancelled")
		return
	}

	code := q.Get("code")
	state := q.Get("state")
	if code == "" || state == "" {
		redirectError(w, req, "invalid_callback")
		return
	}

	verifier, linkUserID, err := consumeOAuthState(req.Context(), state, provider)
	if err != nil {
		redirectError(w, req, "invalid_state")
		return
	}

	prov := oauthProviders[provider]
	identity, err := prov.exchange(req.Context(), code, verifier, redirectURI(req.Host, provider))
	if err != nil {
		redirectError(w, req, "provider_error")
		return
	}

	userID, err := resolveOAuthUser(req.Context(), provider, identity, linkUserID)
	if err != nil {
		redirectError(w, req, "oauth_failed")
		return
	}

	exchangeCode, err := issueExchangeCode(req.Context(), userID)
	if err != nil {
		redirectError(w, req, "oauth_failed")
		return
	}

	// Redirect the browser to the SvelteKit server with the one-time code.
	frontend := secrets.FrontendURL
	if frontend == "" {
		frontend = "http://localhost:5173"
	}
	redirect := frontend + "/auth/oauth/callback?code=" + exchangeCode
	http.Redirect(w, req, redirect, http.StatusFound)
}

func redirectError(w http.ResponseWriter, req *http.Request, reason string) {
	frontend := secrets.FrontendURL
	if frontend == "" {
		frontend = "http://localhost:5173"
	}
	http.Redirect(w, req, frontend+"/auth/oauth/callback?error="+reason, http.StatusFound)
}

// ----- Exchange (browser hands one-time code to SvelteKit server) -----

type OAuthExchangeParams struct {
	Code string `json:"code"`
}

//encore:api public method=POST path=/auth/oauth/exchange
func OAuthExchange(ctx context.Context, p *OAuthExchangeParams) (*AuthResponse, error) {
	if p.Code == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "code is required"}
	}
	userID, err := consumeExchangeCode(ctx, p.Code)
	if err != nil {
		return nil, err
	}
	user, err := getUserByID(ctx, userID)
	if err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "user not found"}
	}
	token, err := createSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: *user}, nil
}

// resolveOAuthUser finds-or-creates the app user for a verified provider
// identity, honoring account linking rules:
//   - If linkUserID is set (logged-in user linking), bind to that user.
//   - Else if (provider, provider_id) already exists, log that user in.
//   - Else create a new user (email populated only if not already taken).
func resolveOAuthUser(ctx context.Context, provider string, identity *oauthIdentity, linkUserID string) (string, error) {
	// 1. Explicit linking intent (user was logged in when starting).
	if linkUserID != "" {
		existing, err := findAccountUser(ctx, provider, identity.ProviderID)
		if err == nil && existing != "" && existing != linkUserID {
			return "", &errs.Error{Code: errs.AlreadyExists, Message: "this " + provider + " account is already linked to another user"}
		}
		if err := linkAccount(ctx, linkUserID, provider, identity); err != nil {
			return "", err
		}
		return linkUserID, nil
	}

	// 2. Existing account → login.
	if uid, err := findAccountUser(ctx, provider, identity.ProviderID); err == nil && uid != "" {
		return uid, nil
	}

	// 3. New user.
	email := ""
	if identity.Email != "" {
		if existingID, err := findUserByEmailID(ctx, identity.Email); err == nil && existingID == "" {
			email = identity.Email
		}
	}

	var user User
	var role = "user"
	insertEmail := interface{}(nil)
	if email != "" {
		insertEmail = email
	}
	var emailOut sql.NullString
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, terms_accepted_at)
		VALUES ($1, NULL, $2, now())
		RETURNING id, email, role, tier, created_at
	`, insertEmail, role).Scan(&user.ID, &emailOut, &user.Role, &user.Tier, &user.CreatedAt)
	if err != nil {
		return "", err
	}
	user.Email = emailOut.String

	if err := linkAccount(ctx, user.ID, provider, identity); err != nil {
		return "", err
	}
	return user.ID, nil
}

func findAccountUser(ctx context.Context, provider, providerAccountID string) (string, error) {
	var userID string
	err := db.QueryRow(ctx, `
		SELECT user_id FROM auth_accounts WHERE provider = $1 AND provider_account_id = $2
	`, provider, providerAccountID).Scan(&userID)
	return userID, err
}

func findUserByEmailID(ctx context.Context, email string) (string, error) {
	var id string
	err := db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	return id, err
}

func linkAccount(ctx context.Context, userID, provider string, identity *oauthIdentity) error {
	_, err := db.Exec(ctx, `
		INSERT INTO auth_accounts (user_id, provider, provider_account_id, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_account_id) DO NOTHING
	`, userID, provider, identity.ProviderID, identity.Email)
	return err
}
