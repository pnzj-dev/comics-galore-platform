package auth

import (
	"context"
	"fmt"

	"encore.dev/beta/errs"
)

func isDevTokenValid(token string) bool {
	return token != "" && token == "dev-secret"
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
			INSERT INTO users (id, email, password_hash, role, tier, username, sub_partner_id)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
		`, u.ID, u.Email, hash, u.Role, u.Tier, u.Username, u.SubPartnerID)
		if err != nil {
			return nil, err
		}
		created++
	}

	return &SeedUsersResponse{
		Created: created,
		Skipped: skipped,
		Message: fmt.Sprintf("Seeded %d users, skipped %d (already exist). Default password: %s", created, skipped, defaultPassword),
	}, nil
}
