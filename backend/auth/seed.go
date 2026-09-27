package auth

import (
	"context"
	"fmt"
	"strings"

	"encore.dev"
	"encore.dev/beta/errs"
)

func isDevTokenValid(token string) bool {
	return token != "" && token == "dev-secret"
}

// isProductionEnv reports whether the app is running in the production
// environment, where the dev seed endpoints must be inert (no-op).
func isProductionEnv() bool {
	name := strings.ToLower(encore.Meta().Environment.Name)
	return name == "production" || name == "prod"
}

type SeedParams struct {
	Token string `json:"token"`
}

type SeedUsersResponse struct {
	Created int    `json:"created"`
	Skipped int    `json:"skipped"`
	Message string `json:"message"`
}

type demoUser struct {
	ID            string
	Email         string
	Role          string
	Tier          string
	Username      string
	SubPartnerID  string
}

//encore:api public method=POST path=/dev/seed-users
func DevSeedUsers(ctx context.Context, p *SeedParams) (*SeedUsersResponse, error) {
	if isProductionEnv() {
		return nil, &errs.Error{Code: errs.Unavailable, Message: "dev seed is disabled in production"}
	}
	if !isDevTokenValid(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	demoUsers := []demoUser{
		{ID: "10000000-0000-0000-0000-000000000001", Email: "admin@comics-galore.dev", Role: "admin", Tier: "platinum", Username: "admin"},
		{ID: "10000000-0000-0000-0000-000000000002", Email: "author-free@pnzj.dev", Role: "uploader", Tier: "free", Username: "author_free"},
		{ID: "10000000-0000-0000-0000-000000000003", Email: "author-gold@pnzj.dev", Role: "uploader", Tier: "gold", Username: "author_gold"},
		{ID: "10000000-0000-0000-0000-000000000004", Email: "member-free@pnzj.dev", Role: "user", Tier: "free", SubPartnerID: "254825522", Username: "member_free"},
		{ID: "10000000-0000-0000-0000-000000000005", Email: "member-bronze@pnzj.dev", Role: "user", Tier: "bronze", Username: "member_bronze"},
		{ID: "10000000-0000-0000-0000-000000000006", Email: "member-silver@pnzj.dev", Role: "user", Tier: "silver", Username: "member_silver"},
		{ID: "10000000-0000-0000-0000-000000000007", Email: "member-gold@pnzj.dev", Role: "user", Tier: "gold", Username: "member_gold"},
		{ID: "10000000-0000-0000-0000-000000000008", Email: "member-platinum@pnzj.dev", Role: "user", Tier: "platinum", Username: "member_platinum"},
		{ID: "10000000-0000-0000-0000-000000000009", Email: "member-exhausted@pnzj.dev", Role: "user", Tier: "free", Username: "member_exhausted"},
		// One login-able account per role (Logto identity linked on first sign-in by email).
		{ID: "10000000-0000-0000-0000-000000000010", Email: "moderator@comics-galore.dev", Role: "moderator", Tier: "platinum", Username: "moderator"},
		{ID: "10000000-0000-0000-0000-000000000011", Email: "uploader@comics-galore.dev", Role: "uploader", Tier: "free", Username: "uploader"},
		{ID: "10000000-0000-0000-0000-000000000012", Email: "user@comics-galore.dev", Role: "user", Tier: "free", Username: "user"},
	}

	created := 0
	skipped := 0

	for _, u := range demoUsers {
		var exists bool
		db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 OR email = $2)`, u.ID, u.Email).Scan(&exists)
		if exists {
			skipped++
			continue
		}

		_, err := db.Exec(ctx, `
			INSERT INTO users (id, email, role, tier, username, sub_partner_id, email_verified_at)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), now())
		`, u.ID, u.Email, u.Role, u.Tier, u.Username, u.SubPartnerID)
		if err != nil {
			return nil, err
		}
		created++
	}

	return &SeedUsersResponse{
		Created: created,
		Skipped: skipped,
		Message: fmt.Sprintf("Seeded %d users, skipped %d (already exist).", created, skipped),
	}, nil
}

// DevLinkLogtoParams links pre-seeded internal users (by email) to their Logto
// identity, so the first sign-in resolves to the seeded role regardless of
// whether Logto emits the email claim.
type DevLinkLogtoParams struct {
	Token string            `json:"token"`
	Users []DevLinkLogtoUser `json:"users"`
}

type DevLinkLogtoUser struct {
	Email   string `json:"email"`
	LogtoID string `json:"logto_id"`
}

type DevLinkLogtoResponse struct {
	Linked int `json:"linked"`
}

//encore:api public method=POST path=/dev/link-logto
func DevLinkLogto(ctx context.Context, p *DevLinkLogtoParams) (*DevLinkLogtoResponse, error) {
	if isProductionEnv() {
		return nil, &errs.Error{Code: errs.Unavailable, Message: "dev seed is disabled in production"}
	}
	if !isDevTokenValid(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	linked := 0
	for _, u := range p.Users {
		email := strings.ToLower(strings.TrimSpace(u.Email))
		logtoID := strings.TrimSpace(u.LogtoID)
		if email == "" || logtoID == "" {
			continue
		}
		res, err := db.Exec(ctx, `UPDATE users SET logto_id = $1 WHERE email = $2 AND logto_id IS NULL`, logtoID, email)
		if err != nil {
			return nil, err
		}
		linked += int(res.RowsAffected())
	}

	return &DevLinkLogtoResponse{Linked: linked}, nil
}
