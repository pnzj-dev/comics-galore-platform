package billing

import (
	"context"
	"fmt"
	"time"

	"encore.dev/beta/errs"
)

type SeedBillingParams struct {
	Token string `json:"token"`
}

type SeedBillingResponse struct {
	Created int    `json:"created"`
	Skipped int    `json:"skipped"`
	Message string `json:"message"`
}

//encore:api public method=POST path=/dev/seed-billing
func DevSeedBilling(ctx context.Context, p *SeedBillingParams) (*SeedBillingResponse, error) {
	if !isBillingDevToken(p.Token) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid dev seed token"}
	}

	var count int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE provider_subscription_id LIKE 'seed-%'`).Scan(&count)
	if count > 0 {
		return &SeedBillingResponse{Created: 0, Skipped: count, Message: "billing data already seeded"}, nil
	}

	// User tiers and sub_partner_id are owned by the auth service (see
	// backend/auth/seed.go); billing seed only creates billing-local data.

	// Plan IDs from tiers service (must match the seed)
	// We just need ANY monthly plan ID for each tier; the exact ID doesn't matter for demo data
	type subEntry struct {
		UserID string
		Tier   string
	}

	subs := []subEntry{
		{"10000000-0000-0000-0000-000000000003", "gold"},
		{"10000000-0000-0000-0000-000000000005", "bronze"},
		{"10000000-0000-0000-0000-000000000006", "silver"},
		{"10000000-0000-0000-0000-000000000007", "gold"},
		{"10000000-0000-0000-0000-000000000008", "platinum"},
	}

	created := 0
	for _, s := range subs {
		// Use a placeholder plan_id — the subscription still works for demo purposes
		_, err := db.Exec(ctx, `
			INSERT INTO subscriptions (id, user_id, plan_id, provider, provider_subscription_id,
				status, active, tier, activated_at, expires_at, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, gen_random_uuid(), 'nowpayments', $3,
				'active', true, $2, now(), now() + interval '30 days', now(), now())
		`, s.UserID, s.Tier, "seed-"+s.UserID)
		if err != nil {
			return nil, err
		}
		created++
	}

	// Sample deposits
	type depoSample struct {
		UserID    string
		Crypto    string
		AmountUSD int
		Status    string
	}

	deps := []depoSample{
		{"10000000-0000-0000-0000-000000000007", "btc", 999, "completed"},
		{"10000000-0000-0000-0000-000000000005", "eth", 399, "completed"},
		{"10000000-0000-0000-0000-000000000004", "usdt", 0, "pending"},
	}

	for _, d := range deps {
		var completedAt interface{}
		if d.Status == "completed" {
			completedAt = time.Now().AddDate(0, 0, -7)
		}
		db.Exec(ctx, `
			INSERT INTO deposits (id, user_id, provider, provider_deposit_id, currency_crypto, amount_usd_cents, status, pay_address, created_at, completed_at)
			VALUES (gen_random_uuid(), $1, 'nowpayments', $6, $2, $3, $4, 'seed-address-'||$2, now() - interval '14 days', $5)
		`, d.UserID, d.Crypto, d.AmountUSD, d.Status, completedAt, "seed-deposit-"+d.UserID)
	}

	return &SeedBillingResponse{
		Created: created,
		Skipped: 0,
		Message: fmt.Sprintf("Seeded %d subscriptions + 3 deposits", created),
	}, nil
}

func isBillingDevToken(token string) bool {
	return token != "" && token == "dev-secret"
}

var _ = time.Now
