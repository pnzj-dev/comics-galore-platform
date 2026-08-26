package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

// oauthIdentity is the minimal verified identity a provider returns.
type oauthIdentity struct {
	ProviderID string // stable provider account id
	Email      string
	Verified   bool // whether the provider guarantees the email belongs to the user
	Name       string
}

// oauthProvider abstracts the four supported providers behind one interface.
type oauthProvider interface {
	name() string
	// authURL returns the provider's authorization URL (state + PKCE challenge).
	authURL(state, redirectURL, challenge string) string
	// exchange exchanges the code for a verified identity. `verifier` is the
	// PKCE code verifier (empty for providers that don't support PKCE).
	exchange(ctx context.Context, code, verifier, redirectURL string) (*oauthIdentity, error)
}

var oauthProviders = map[string]oauthProvider{
	"google":   googleOAuth{},
	"facebook": facebookOAuth{},
	"twitter":  twitterOAuth{},
	"apple":    appleOAuth{},
}

func isKnownProvider(p string) bool {
	_, ok := oauthProviders[p]
	return ok
}

// schemeFor derives http vs https from the request host (localhost → http).
func schemeFor(host string) string {
	h := strings.ToLower(host)
	if strings.HasPrefix(h, "localhost") || strings.HasPrefix(h, "127.") || strings.HasPrefix(h, "0.0.0.0") {
		return "http"
	}
	return "https"
}

// redirectURI builds the callback URL for a provider given the request host.
func redirectURI(host, provider string) string {
	return schemeFor(host) + "://" + host + "/auth/oauth/" + provider + "/callback"
}

// exchangeCode issues a short-lived single-use code the browser hands to the
// SvelteKit server, which exchanges it for a session token.
func issueExchangeCode(ctx context.Context, userID string) (string, error) {
	code := randomToken(32)
	_, err := db.Exec(ctx, `
		INSERT INTO oauth_exchange_codes (code, user_id, expires_at)
		VALUES ($1, $2, now() + interval '60 seconds')
	`, code, userID)
	return code, err
}

func consumeExchangeCode(ctx context.Context, code string) (string, error) {
	var userID string
	var expiresAt time.Time
	err := db.QueryRow(ctx, `
		DELETE FROM oauth_exchange_codes
		WHERE code = $1 AND expires_at > now()
		RETURNING user_id, expires_at
	`, code).Scan(&userID, &expiresAt)
	if err != nil {
		if isNoRows(err) {
			return "", &errs.Error{Code: errs.Unauthenticated, Message: "invalid or expired login code"}
		}
		return "", err
	}
	return userID, nil
}

// parseUserInfo unmarshals a provider userinfo JSON body into a map with the
// given field keys.
func parseUserInfo(body []byte, idKey, emailKey, verifiedKey string) (*oauthIdentity, error) {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	id, _ := m[idKey].(string)
	if id == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "provider returned no account id"}
	}
	email, _ := m[emailKey].(string)
	verified, _ := m[verifiedKey].(bool)
	name, _ := m["name"].(string)
	return &oauthIdentity{ProviderID: id, Email: email, Verified: verified, Name: name}, nil
}

// fetchJSON is a tiny helper for provider userinfo GETs.
func fetchJSON(ctx context.Context, url, bearer string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "provider rejected token"}
	}
	return body, nil
}

// httpPostForm POSTs urlencoded form values and returns the response.
func httpPostForm(ctx context.Context, urlStr string, values map[string]string) (*http.Response, error) {
	form := url.Values{}
	for k, v := range values {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return http.DefaultClient.Do(req)
}
