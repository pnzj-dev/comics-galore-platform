package auth

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"encore.dev/beta/errs"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// --- Logto token validation -------------------------------------------------
//
// Logto is the identity provider: it owns passwords, social connectors,
// passkeys, MFA, and password reset. The Encore auth handler validates the
// Logto-issued access token and maps its `sub` to an internal `users` row.

var jwksCache struct {
	mu  sync.Mutex
	set jwk.Set
	at  time.Time
}

// jwksProvider fetches the Logto JWKS. Package-level so tests can override it
// with a locally generated key set (no network in unit tests).
var jwksProvider = func(ctx context.Context) (jwk.Set, error) {
	jwksCache.mu.Lock()
	defer jwksCache.mu.Unlock()
	if jwksCache.set != nil && time.Since(jwksCache.at) < 15*time.Minute {
		return jwksCache.set, nil
	}
	uri := strings.TrimSpace(secrets.LogtoJWKSURI)
	if uri == "" {
		return nil, &errs.Error{Code: errs.Internal, Message: "Logto JWKS URI not configured"}
	}
	set, err := jwk.Fetch(ctx, uri)
	if err != nil {
		return nil, err
	}
	jwksCache.set = set
	jwksCache.at = time.Now()
	return set, nil
}

// logtoClaims is the subset of token claims the auth handler needs.
type logtoClaims struct {
	Sub   string
	Email string
}

// validateLogtoToken verifies a Logto access token and returns its claims.
func validateLogtoToken(ctx context.Context, token string) (*logtoClaims, error) {
	set, err := jwksProvider(ctx)
	if err != nil {
		return nil, err
	}

	opts := []jwt.ParseOption{jwt.WithKeySet(set)}
	if issuer := strings.TrimSpace(secrets.LogtoIssuer); issuer != "" {
		opts = append(opts, jwt.WithIssuer(issuer))
	}
	if aud := strings.TrimSpace(secrets.LogtoAudience); aud != "" {
		opts = append(opts, jwt.WithAudience(aud))
	}

	parsed, err := jwt.Parse([]byte(token), opts...)
	if err != nil {
		return nil, err
	}

	sub, _ := parsed.Subject()
	claims := &logtoClaims{Sub: strings.TrimSpace(sub)}

	var email string
	if parsed.Get("email", &email) == nil {
		claims.Email = strings.TrimSpace(email)
	}
	return claims, nil
}

// resolveLogtoUser resolves (and lazily provisions) the internal user for a
// Logto identity. Linking order: logto_id → email → provision.
func resolveLogtoUser(ctx context.Context, sub, email string) (*userRow, error) {
	if sub != "" {
		u, err := getUserByLogtoID(ctx, sub)
		if err == nil {
			return u, nil
		}
		if !isNoRows(err) {
			return nil, err
		}
	}
	if email != "" {
		u, err := getUserByEmail(ctx, email)
		if err == nil {
			if sub != "" {
				// Link the existing account to its Logto identity.
				db.Exec(ctx, `UPDATE users SET logto_id = $1 WHERE id = $2 AND logto_id IS NULL`, sub, u.ID)
			}
			return u, nil
		}
		if !isNoRows(err) {
			return nil, err
		}
	}
	return provisionUser(ctx, sub, email)
}

func getUserByLogtoID(ctx context.Context, sub string) (*userRow, error) {
	var u userRow
	err := db.QueryRow(ctx, `
		SELECT id, email, role, tier, COALESCE(username, ''), COALESCE(avatar_key::text, ''), banned_at, suspended_at, created_at
		FROM users WHERE logto_id = $1
	`, sub).Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.Username, &u.AvatarKey, &u.BannedAt, &u.SuspendedAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// provisionUser creates a fresh internal user for a new Logto identity and
// eagerly provisions the NowPayments sub-partner (best-effort).
func provisionUser(ctx context.Context, sub, email string) (*userRow, error) {
	var u userRow
	var emailArg any
	if email != "" {
		emailArg = email
	}
	var subArg any
	if sub != "" {
		subArg = sub
	}
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, logto_id, role, tier)
		VALUES ($1, $2, 'user', 'free')
		RETURNING id, email, role, tier, COALESCE(username, ''), COALESCE(avatar_key::text, ''), banned_at, suspended_at, created_at
	`, emailArg, subArg).Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.Username, &u.AvatarKey, &u.BannedAt, &u.SuspendedAt, &u.CreatedAt)
	if err != nil {
		// Lost a concurrent-provision race; recover by re-resolving.
		if sub != "" {
			if u2, e2 := getUserByLogtoID(ctx, sub); e2 == nil {
				return u2, nil
			}
		}
		if email != "" {
			if u2, e2 := getUserByEmail(ctx, email); e2 == nil {
				return u2, nil
			}
		}
		return nil, err
	}

	// Best-effort NowPayments customer provisioning, fired-and-forgotten so a
	// slow provider never blocks first sign-in. Skips when no key is configured.
	if secrets.NowPaymentsAPIKey != "" {
		uid := u.ID
		go func() {
			if _, err := ensureSubPartnerID(context.Background(), uid); err != nil {
				log.Printf("[auth] provision sub_partner_id for %s: %v", uid, err)
			}
		}()
	}
	return &u, nil
}
