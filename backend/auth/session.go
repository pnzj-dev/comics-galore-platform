package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"encore.dev/beta/errs"
)

// sessionDuration is how long a fresh session lasts before requiring a new
// sign-in. The browser cookie max-age mirrors this.
const sessionDuration = 30 * 24 * time.Hour

// createSession creates a new opaque session for the user and returns its
// token (the session id). Callers are responsible for issuing the cookie.
func createSession(ctx context.Context, userID string) (string, error) {
	token := randomToken(32)
	_, err := db.Exec(ctx, `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, now() + $3::interval)
	`, token, userID, sessionDuration.String())
	if err != nil {
		return "", err
	}
	return token, nil
}

// createImpersonationSession creates a session for `userID` that records the
// admin who initiated it. The auth handler surfaces ImpersonatedBy.
func createImpersonationSession(ctx context.Context, userID, impersonatedBy string) (string, error) {
	token := randomToken(32)
	_, err := db.Exec(ctx, `
		INSERT INTO sessions (id, user_id, impersonated_by, expires_at)
		VALUES ($1, $2, $3, now() + $4::interval)
	`, token, userID, impersonatedBy, sessionDuration.String())
	if err != nil {
		return "", err
	}
	return token, nil
}

// sessionUser is the resolved identity of a validated session.
type sessionUser struct {
	UserID         string
	ImpersonatedBy string
	Expiry         time.Time
}

// validateSession resolves an opaque session token to its user, enforcing
// revocation and expiry. Returns an errs.Error suitable for the auth handler.
func validateSession(ctx context.Context, token string) (*sessionUser, error) {
	if token == "" {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing session"}
	}
	var s sessionUser
	var revoked sql.NullTime
	var impersonatedBy sql.NullString
	err := db.QueryRow(ctx, `
		SELECT user_id, COALESCE(impersonated_by::text, ''), expires_at, revoked_at
		FROM sessions WHERE id = $1
	`, token).Scan(&s.UserID, &impersonatedBy, &s.Expiry, &revoked)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid session"}
		}
		return nil, err
	}
	if revoked.Valid || time.Now().After(s.Expiry) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "session expired"}
	}
	s.ImpersonatedBy = impersonatedBy.String
	return &s, nil
}

// touchSession best-effort updates last_seen_at for an active session.
func touchSession(ctx context.Context, token string) {
	_, _ = db.Exec(ctx, `UPDATE sessions SET last_seen_at = now() WHERE id = $1`, token)
}

// revokeAllSessions invalidates every active session for a user.
func revokeAllSessions(ctx context.Context, userID string) error {
	_, err := db.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

var errSessionNotFound = errors.New("session not found")

// deleteSession hard-removes a session (used on logout).
func deleteSession(ctx context.Context, token string) error {
	res, err := db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, token)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errSessionNotFound
	}
	return nil
}
