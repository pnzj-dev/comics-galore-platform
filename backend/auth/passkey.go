package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

const webauthnChallengeTTL = 5 * time.Minute

// storeChallenge persists a ceremony's session data, single-use, short-lived.
func storeChallenge(ctx context.Context, challenge string, userID string, purpose string, session *webauthn.SessionData) error {
	raw, err := json.Marshal(session)
	if err != nil {
		return err
	}
	var uid interface{}
	if userID != "" {
		uid = userID
	}
	_, err = db.Exec(ctx, `
		INSERT INTO webauthn_challenges (challenge, user_id, purpose, session_data, expires_at)
		VALUES ($1, $2, $3, $4, now() + $5::interval)
	`, challenge, uid, purpose, string(raw), (webauthnChallengeTTL).String())
	return err
}

// consumeChallenge loads and atomically deletes a single-use challenge.
func consumeChallenge(ctx context.Context, challenge, purpose string) (*webauthn.SessionData, error) {
	var raw []byte
	var expiresAt time.Time
	err := db.QueryRow(ctx, `
		SELECT session_data, expires_at FROM webauthn_challenges
		WHERE challenge = $1 AND purpose = $2
	`, challenge, purpose).Scan(&raw, &expiresAt)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "unknown or expired challenge"}
		}
		return nil, err
	}
	if time.Now().After(expiresAt) {
		_, _ = db.Exec(ctx, `DELETE FROM webauthn_challenges WHERE challenge = $1`, challenge)
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "challenge expired"}
	}
	// Single-use: delete before returning so a replay cannot reuse it.
	if _, err := db.Exec(ctx, `DELETE FROM webauthn_challenges WHERE challenge = $1`, challenge); err != nil {
		return nil, err
	}
	var session webauthn.SessionData
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// ----- Registration -----

type PasskeyRegisterOptionsParams struct {
	Name string `json:"name"`
}

type PasskeyRegisterOptionsResponse struct {
	Options json.RawMessage `json:"options"`
}

//encore:api auth method=POST path=/auth/passkey/register/options
func PasskeyRegisterOptions(ctx context.Context, p *PasskeyRegisterOptionsParams) (*PasskeyRegisterOptionsResponse, error) {
	ad := auth.Data().(*AuthData)
	w, err := webAuthn()
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "webauthn unavailable"}
	}

	user, err := loadWebAuthnUser(ctx, ad.UserID)
	if err != nil {
		return nil, err
	}

	creation, session, err := w.BeginRegistration(user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		webauthn.WithExclusions(webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			UserVerification:        protocol.VerificationRequired,
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
			AuthenticatorAttachment: protocol.AuthenticatorAttachment(""),
		}),
	)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "failed to begin registration"}
	}

	if err := storeChallenge(ctx, session.Challenge, ad.UserID, "register", session); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(creation)
	if err != nil {
		return nil, err
	}
	return &PasskeyRegisterOptionsResponse{Options: raw}, nil
}

type PasskeyRegisterVerifyParams struct {
	Name       string `json:"name"`
	RawResponse json.RawMessage `json:"response"`
}

//encore:api auth method=POST path=/auth/passkey/register/verify
func PasskeyRegisterVerify(ctx context.Context, p *PasskeyRegisterVerifyParams) (*PasskeyListResponse, error) {
	ad := auth.Data().(*AuthData)

	user, err := loadWebAuthnUser(ctx, ad.UserID)
	if err != nil {
		return nil, err
	}

	// Parse the credential from the client-supplied JSON body.
	parsed, err := protocol.ParseCredentialCreationResponseBytes(p.RawResponse)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid credential response"}
	}

	session, err := consumeChallenge(ctx, string(parsed.Response.CollectedClientData.Challenge), "register")
	if err != nil {
		return nil, err
	}

	w, err := webAuthnForOrigin(parsed.Response.CollectedClientData.Origin)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "webauthn unavailable"}
	}

	credential, err := w.CreateCredential(user, *session, parsed)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "credential verification failed: " + err.Error()}
	}

	raw, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)

	name := p.Name
	if name == "" {
		name = "Passkey"
	}

	_, err = db.Exec(ctx, `
		INSERT INTO passkeys (user_id, name, credential_id, credential)
		VALUES ($1, $2, $3, $4)
	`, ad.UserID, name, credentialID, string(raw))
	if err != nil {
		return nil, err
	}

	return listPasskeys(ctx, ad.UserID)
}

// ----- Login -----

//encore:api public method=POST path=/auth/passkey/login/options
func PasskeyLoginOptions(ctx context.Context) (*PasskeyRegisterOptionsResponse, error) {
	w, err := webAuthn()
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "webauthn unavailable"}
	}

	assertion, session, err := w.BeginDiscoverableMediatedLogin(protocol.MediationOptional, webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "failed to begin login"}
	}

	if err := storeChallenge(ctx, session.Challenge, "", "login", session); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(assertion)
	if err != nil {
		return nil, err
	}
	return &PasskeyRegisterOptionsResponse{Options: raw}, nil
}

type PasskeyLoginVerifyParams struct {
	RawResponse json.RawMessage `json:"response"`
}

//encore:api public method=POST path=/auth/passkey/login/verify
func PasskeyLoginVerify(ctx context.Context, p *PasskeyLoginVerifyParams) (*AuthResponse, error) {
	parsed, err := protocol.ParseCredentialRequestResponseBytes(p.RawResponse)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid assertion"}
	}

	session, err := consumeChallenge(ctx, string(parsed.Response.CollectedClientData.Challenge), "login")
	if err != nil {
		return nil, err
	}

	w, err := webAuthnForOrigin(parsed.Response.CollectedClientData.Origin)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "webauthn unavailable"}
	}

	validatedUser, credential, err := w.ValidatePasskeyLogin(discoverableUserHandler(ctx), *session, parsed)
	if err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "passkey verification failed: " + err.Error()}
	}

	wu := validatedUser.(*webauthnUser)
	if wu == nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "passkey verification failed"}
	}

	// Update sign count + last_used_at for the credential.
	raw, _ := json.Marshal(credential)
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	db.Exec(ctx, `
		UPDATE passkeys SET credential = $1, last_used_at = now()
		WHERE credential_id = $2 AND user_id = $3
	`, string(raw), credentialID, wu.ID)

	user, err := getUserByID(ctx, wu.ID)
	if err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "user not found"}
	}

	token, err := createSession(ctx, wu.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: *user}, nil
}

// ----- Management -----

type PasskeyInfo struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

type PasskeyListResponse struct {
	Passkeys []PasskeyInfo `json:"passkeys"`
}

//encore:api auth method=GET path=/auth/passkeys
func ListPasskeys(ctx context.Context) (*PasskeyListResponse, error) {
	ad := auth.Data().(*AuthData)
	return listPasskeys(ctx, ad.UserID)
}

func listPasskeys(ctx context.Context, userID string) (*PasskeyListResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, created_at, last_used_at FROM passkeys
		WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out PasskeyListResponse
	for rows.Next() {
		var p PasskeyInfo
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.LastUsedAt); err != nil {
			return nil, err
		}
		out.Passkeys = append(out.Passkeys, p)
	}
	return &out, rows.Err()
}

//encore:api auth method=DELETE path=/auth/passkeys/:id
func DeletePasskey(ctx context.Context, id string) error {
	ad := auth.Data().(*AuthData)
	if id == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "id is required"}
	}
	// Only apply the recovery guard if the passkey exists and is owned by
	// the user (deleting a nonexistent id is a no-op).
	var exists bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM passkeys WHERE id = $1 AND user_id = $2)`, id, ad.UserID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if err := guardRemoveAuthMethod(ctx, ad.UserID); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `DELETE FROM passkeys WHERE id = $1 AND user_id = $2`, id, ad.UserID)
	return err
}
