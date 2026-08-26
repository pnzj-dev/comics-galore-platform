package reading

import (
	"context"

	"encore.dev/beta/errs"
)

// exhaustedUserID matches the dedicated demo user seeded by the auth service
// (member-exhausted@pnzj.dev), free tier (5 downloads/month).
const exhaustedUserID = "10000000-0000-0000-0000-000000000009"

type SeedParams struct {
	Token string `json:"token"`
}

type SeedQuotaResponse struct {
	Inserted int    `json:"inserted"`
	UserID   string `json:"user_id"`
	Message  string `json:"message"`
}

// DevSeedQuota exhausts the free-tier download quota for the dedicated demo
// user so the boost-quota flow can be tested manually.
//encore:api public method=POST path=/dev/seed-quota
func DevSeedQuota(ctx context.Context, p *SeedParams) (*SeedQuotaResponse, error) {
	if p.Token != "dev-secret" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	var count int
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM download_logs
		WHERE user_id = $1 AND created_at >= date_trunc('month', now())
	`, exhaustedUserID).Scan(&count); err != nil {
		return nil, err
	}

	// Free quota is 5 downloads/month; ensure 6 records exist this month.
	inserted := 0
	for i := count; i < 6; i++ {
		if _, err := db.Exec(ctx, `
			INSERT INTO download_logs (user_id, comic_id)
			VALUES ($1, '00000000-0000-0000-0000-000000000000')
		`, exhaustedUserID); err != nil {
			return nil, err
		}
		inserted++
	}

	return &SeedQuotaResponse{
		Inserted: inserted,
		UserID:   exhaustedUserID,
		Message:  "member-exhausted@pnzj.dev quota is now exhausted (free tier, 5 downloads/month). Sign in as this user to test the boost flow.",
	}, nil
}
