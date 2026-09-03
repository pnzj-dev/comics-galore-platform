package auth

import (
	"context"
	"crypto/subtle"
	"strings"

	"encore.dev/beta/errs"
)

type BootstrapAdminParams struct {
	Token    string `json:"token" encore:"sensitive"`
	Email    string `json:"email"`
	Password string `json:"password" encore:"sensitive"`
}

type BootstrapAdminResponse struct {
	User User `json:"user"`
}

//encore:api public method=POST path=/auth/bootstrap
func BootstrapAdmin(ctx context.Context, p *BootstrapAdminParams) (*BootstrapAdminResponse, error) {
	secret := strings.TrimSpace(secrets.BootstrapSecret)
	if secret == "" {
		return nil, &errs.Error{Code: errs.Unavailable, Message: "bootstrap is disabled"}
	}
	if subtle.ConstantTimeCompare([]byte(p.Token), []byte(secret)) != 1 {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid bootstrap token"}
	}

	email := strings.ToLower(strings.TrimSpace(p.Email))
	if email == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "email is required"}
	}
	if len(p.Password) < 8 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "password must be at least 8 characters"}
	}

	var adminCount int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount); err != nil {
		return nil, err
	}
	if adminCount > 0 {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "an admin already exists; bootstrap is one-time only"}
	}

	hash, err := hashPassword(p.Password)
	if err != nil {
		return nil, err
	}

	var user User
	err = db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, tier, username, terms_accepted_at, email_verified_at)
		VALUES ($1, $2, 'admin', 'platinum', $3, now(), now())
		RETURNING id, email, role, tier, COALESCE(username, ''), created_at
	`, email, hash, adminUsername(email)).Scan(&user.ID, &user.Email, &user.Role, &user.Tier, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &BootstrapAdminResponse{User: user}, nil
}

// adminUsername derives a valid lowercase public handle from an email address.
// Falls back to "admin" when the local part is too short after sanitization.
func adminUsername(email string) string {
	local := email
	if i := strings.Index(local, "@"); i >= 0 {
		local = local[:i]
	}

	var b strings.Builder
	lastSep := false
	for _, r := range strings.ToLower(local) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastSep = false
		case r == '-' || r == '_':
			if b.Len() > 0 && !lastSep {
				b.WriteRune(r)
				lastSep = true
			}
		}
	}

	s := strings.TrimRight(b.String(), "-_")
	if len(s) < 3 {
		s = "admin"
	}
	if len(s) > 20 {
		s = strings.TrimRight(s[:20], "-_")
	}
	return s
}
