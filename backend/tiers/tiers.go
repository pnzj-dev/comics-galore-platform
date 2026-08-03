package tiers

import (
	"context"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("tiersdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type Tier struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Plan struct {
	ID              string    `json:"id"`
	TierID          string    `json:"tier_id"`
	Interval        string    `json:"interval"`
	PriceUsdCents   int       `json:"price_usd_cents"`
	QuotaDownloads  int       `json:"quota_downloads"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

//encore:api public method=GET path=/tiers
func ListTiers(ctx context.Context) (*ListTiersResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, description, sort_order, created_at
		FROM tiers
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []Tier
	for rows.Next() {
		var t Tier
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.SortOrder, &t.CreatedAt); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ListTiersResponse{Tiers: tiers}, nil
}

type ListTiersResponse struct {
	Tiers []Tier `json:"tiers"`
}

//encore:api public method=GET path=/tiers/:id
func GetTier(ctx context.Context, id string) (*Tier, error) {
	var t Tier
	err := db.QueryRow(ctx, `
		SELECT id, name, description, sort_order, created_at
		FROM tiers WHERE id = $1
	`, id).Scan(&t.ID, &t.Name, &t.Description, &t.SortOrder, &t.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "tier not found",
			}
		}
		return nil, err
	}
	return &t, nil
}

//encore:api public method=GET path=/plans
func ListPlans(ctx context.Context) (*ListPlansResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, tier_id, interval, price_usd_cents, quota_downloads, is_active, created_at
		FROM plans
		ORDER BY price_usd_cents ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.ID, &p.TierID, &p.Interval, &p.PriceUsdCents, &p.QuotaDownloads, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ListPlansResponse{Plans: plans}, nil
}

type ListPlansResponse struct {
	Plans []Plan `json:"plans"`
}
