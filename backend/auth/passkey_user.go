package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/go-webauthn/webauthn/webauthn"

	"encore.dev/beta/errs"
)

// webauthnConfig holds the relying-party configuration. RPID/Origins are
// environment-aware: localhost in dev, the real domain in staging/production.
// Values are loaded from the service config (see encore.gen.cue).
type webauthnCfg struct {
	RPID          string   `json:"rp_id"`
	RPDisplayName string   `json:"rp_display_name"`
	RPOrigins     []string `json:"rp_origins"`
}

var waCfg = loadWebAuthnConfig()

func loadWebAuthnConfig() webauthnCfg {
	// Defaults are localhost-safe so `encore run` works out of the box;
	// production values are set per environment via Encore config/secrets.
	rpID := secrets.WebAuthnRPID
	if rpID == "" {
		rpID = "localhost"
	}
	return webauthnCfg{
		RPID:          rpID,
		RPDisplayName: "Comics Galore",
		RPOrigins:     webauthnOrigins(),
	}
}

func webauthnOrigins() []string {
	if secrets.WebAuthnOrigins != "" {
		return splitCSV(secrets.WebAuthnOrigins)
	}
	if secrets.FrontendURL != "" {
		return []string{secrets.FrontendURL}
	}
	return []string{"http://localhost:5173", "http://localhost:5174"}
}

// wa is the lazily-initialised WebAuthn instance.
var wa *webauthn.WebAuthn

func webAuthn() (*webauthn.WebAuthn, error) {
	if wa != nil {
		return wa, nil
	}
	w, err := webauthn.New(&webauthn.Config{
		RPID:          waCfg.RPID,
		RPDisplayName: waCfg.RPDisplayName,
		RPOrigins:     waCfg.RPOrigins,
	})
	if err != nil {
		return nil, err
	}
	wa = w
	return wa, nil
}

// isLocalhostOrigin reports whether an origin's hostname is a loopback
// address. Uses the parsed hostname (not a substring match) so lookalikes
// such as "localhost.evil.com" are never accepted.
func isLocalhostOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

// webAuthnForOrigin returns a WebAuthn instance whose allowed origins include
// `origin` when it is a loopback origin (dev). The frontend port can vary on
// localhost, so we accept the actual origin rather than a hardcoded port.
// Non-localhost origins reuse the shared singleton unchanged.
func webAuthnForOrigin(origin string) (*webauthn.WebAuthn, error) {
	w, err := webAuthn()
	if err != nil {
		return nil, err
	}
	if !isLocalhostOrigin(origin) {
		return w, nil
	}
	for _, o := range w.Config.RPOrigins {
		if o == origin {
			return w, nil
		}
	}
	cfg := *w.Config
	cfg.RPOrigins = append(append([]string(nil), cfg.RPOrigins...), origin)
	return webauthn.New(&cfg)
}

// webauthnUser adapts a users row to the go-webauthn User interface.
type webauthnUser struct {
	ID          string
	Email       string
	Credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte                       { return []byte(u.ID) }
func (u *webauthnUser) WebAuthnName() string                     { return u.Email }
func (u *webauthnUser) WebAuthnDisplayName() string              { return u.Email }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }

// loadWebAuthnUser fetches a user and their passkeys, returning a webauthnUser.
func loadWebAuthnUser(ctx context.Context, userID string) (*webauthnUser, error) {
	var email string
	err := db.QueryRow(ctx, `SELECT COALESCE(email, '') FROM users WHERE id = $1`, userID).Scan(&email)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, `SELECT credential FROM passkeys WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	creds := []webauthn.Credential{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var c webauthn.Credential
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return &webauthnUser{ID: userID, Email: email, Credentials: creds}, rows.Err()
}

// findUserByPasskey resolves a discoverable credential's user via credential id.
func findUserByPasskey(ctx context.Context, credentialID string) (string, error) {
	var userID string
	err := db.QueryRow(ctx, `SELECT user_id FROM passkeys WHERE credential_id = $1`, credentialID).Scan(&userID)
	if err != nil {
		if isNoRows(err) {
			return "", &errs.Error{Code: errs.Unauthenticated, Message: "passkey not found"}
		}
		return "", err
	}
	return userID, nil
}

// discoverableUserHandler resolves a user handle (or credential id) to a
// webauthnUser during a discoverable (usernameless) login ceremony.
func discoverableUserHandler(ctx context.Context) webauthn.DiscoverableUserHandler {
	return func(rawID, userHandle []byte) (webauthn.User, error) {
		// Prefer the credential id (rawID): it's always present and uniquely
		// identifies the passkey and thus its owner.
		userID, err := findUserByPasskey(ctx, base64.RawURLEncoding.EncodeToString(rawID))
		if err != nil {
			// Fall back to the user handle if the credential id is unknown.
			if len(userHandle) > 0 {
				if u, err2 := loadWebAuthnUser(ctx, string(userHandle)); err2 == nil {
					return u, nil
				}
			}
			return nil, err
		}
		return loadWebAuthnUser(ctx, userID)
	}
}
