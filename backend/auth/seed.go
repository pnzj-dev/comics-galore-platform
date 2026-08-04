package auth

import (
	"context"
	"fmt"

	encoreauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

var seedSecrets struct {
	DevSeedToken string
}

func isDevTokenValid(ctx context.Context, token string) bool {
	if seedSecrets.DevSeedToken == "" {
		return false
	}
	return token == seedSecrets.DevSeedToken
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
	ID    string
	Email string
	Role  string
	Tier  string
}

//encore:api public method=POST path=/dev/seed-users
func DevSeedUsers(ctx context.Context, p *SeedParams) (*SeedUsersResponse, error) {
	if !isDevTokenValid(ctx, p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	demoUsers := []demoUser{
		{ID: "10000000-0000-0000-0000-000000000001", Email: "admin@comics-galore.dev", Role: "admin", Tier: "platinum"},
		{ID: "10000000-0000-0000-0000-000000000002", Email: "author-free@pnzj.dev", Role: "uploader", Tier: "free"},
		{ID: "10000000-0000-0000-0000-000000000003", Email: "author-gold@pnzj.dev", Role: "uploader", Tier: "gold"},
		{ID: "10000000-0000-0000-0000-000000000004", Email: "member-free@pnzj.dev", Role: "user", Tier: "free"},
		{ID: "10000000-0000-0000-0000-000000000005", Email: "member-bronze@pnzj.dev", Role: "user", Tier: "bronze"},
		{ID: "10000000-0000-0000-0000-000000000006", Email: "member-silver@pnzj.dev", Role: "user", Tier: "silver"},
		{ID: "10000000-0000-0000-0000-000000000007", Email: "member-gold@pnzj.dev", Role: "user", Tier: "gold"},
		{ID: "10000000-0000-0000-0000-000000000008", Email: "member-platinum@pnzj.dev", Role: "user", Tier: "platinum"},
	}

	defaultPassword := "devpassword"
	hash, err := hashPassword(defaultPassword)
	if err != nil {
		return nil, err
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
			INSERT INTO users (id, email, password_hash, role, tier)
			VALUES ($1, $2, $3, $4, $5)
		`, u.ID, u.Email, hash, u.Role, u.Tier)
		if err != nil {
			return nil, err
		}
		created++
	}

	// In dev, ensure at least one admin exists
	var adminCount int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount)
	if adminCount == 0 {
		return nil, fmt.Errorf("no admin user exists after seed")
	}

	return &SeedUsersResponse{
		Created: created,
		Skipped: skipped,
		Message: fmt.Sprintf("Seeded %d users, skipped %d (already exist). Default password: %s", created, skipped, defaultPassword),
	}, nil
}

var _ = encoreauth.Data // suppress unused import warning
