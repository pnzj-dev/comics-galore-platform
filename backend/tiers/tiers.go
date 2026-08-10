package tiers

import (
	"context"
	"strings"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"
	billing "comics-galore/backend/billing"
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
	ID                   string    `json:"id"`
	TierID               string    `json:"tier_id"`
	Name                 string    `json:"name"`
	Interval             string    `json:"interval"`
	PriceUsdCents        int       `json:"price_usd_cents"`
	QuotaDownloads       int       `json:"quota_downloads"`
	Features             []string  `json:"features"`
	IsActive             bool      `json:"is_active"`
	ProviderPlanID       string    `json:"provider_plan_id,omitempty"`
	ProviderIntervalDays int       `json:"provider_interval_days"`
	CreatedAt            time.Time `json:"created_at"`
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
		SELECT plans.id, plans.tier_id, COALESCE(plans.name, t.name || ' ' || plans.interval) as name, plans.interval,
			plans.price_usd_cents, plans.quota_downloads, COALESCE(plans.features, '[]'),
			plans.is_active, COALESCE(plans.provider_plan_id, ''), COALESCE(plans.provider_interval_days, 0), plans.created_at
		FROM plans LEFT JOIN tiers t ON t.id = plans.tier_id
		ORDER BY plans.price_usd_cents ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		var featuresJSON []byte
		if err := rows.Scan(&p.ID, &p.TierID, &p.Name, &p.Interval, &p.PriceUsdCents,
			&p.QuotaDownloads, &featuresJSON, &p.IsActive, &p.ProviderPlanID,
			&p.ProviderIntervalDays, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Features = parseFeatures(featuresJSON)
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

type MatrixStatus struct {
	Complete bool `json:"complete"`
}

//encore:api public method=GET path=/plans/ready
func PlansReady(ctx context.Context) (*MatrixStatus, error) {
	var count int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM plans WHERE provider_plan_id IS NULL OR provider_plan_id = ''`).Scan(&count)
	return &MatrixStatus{Complete: count == 0}, nil
}

//encore:api auth method=GET path=/admin/plans/matrix-status
func PlanMatrixStatus(ctx context.Context) (*MatrixStatus, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var count int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM plans WHERE provider_plan_id IS NULL OR provider_plan_id = ''`).Scan(&count)
	return &MatrixStatus{Complete: count == 0}, nil
}

type UpdatePlanProviderParams struct {
	ProviderPlanID string `json:"provider_plan_id"`
}

//encore:api auth method=PATCH path=/admin/plans/:id
func UpdatePlanProviderID(ctx context.Context, id string, p *UpdatePlanProviderParams) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var existing string
	err := db.QueryRow(ctx, `SELECT COALESCE(provider_plan_id, '') FROM plans WHERE id = $1`, id).Scan(&existing)
	if err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "plan not found"}
		}
		return err
	}
	if existing != "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "plan already linked to provider plan ID: " + existing}
	}

	_, err = db.Exec(ctx, `UPDATE plans SET provider_plan_id = $1 WHERE id = $2`, p.ProviderPlanID, id)
	return err
}

type AutoLinkPlanResponse struct {
	ProviderPlanID string `json:"provider_plan_id"`
	PlanName       string `json:"plan_name"`
}

//encore:api auth method=POST path=/admin/plans/link/:id
func AutoLinkPlan(ctx context.Context, id string) (*AutoLinkPlanResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var plan struct {
		Name          string
		TierID        string
		Interval      string
		PriceUsdCents int
		ProviderPlanID string
	}
	err := db.QueryRow(ctx, `
		SELECT COALESCE(plans.name, ''), plans.tier_id, plans.interval,
			plans.price_usd_cents, COALESCE(plans.provider_plan_id, '')
		FROM plans WHERE id = $1
	`, id).Scan(&plan.Name, &plan.TierID, &plan.Interval, &plan.PriceUsdCents, &plan.ProviderPlanID)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found"}
		}
		return nil, err
	}
	if plan.ProviderPlanID != "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "plan already linked to provider plan ID: " + plan.ProviderPlanID}
	}
	if plan.PriceUsdCents <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "free plan cannot be linked to NowPayments (zero price)"}
	}

	period := intervalToPeriod(plan.Interval)
	displayName := plan.Name
	if displayName == "" {
		displayName = plan.TierID
	}
	displayName += " - " + plan.Interval

	resp, err := billing.CreatePlan(ctx, billing.CreatePlanRequest{
		Name:        displayName,
		PriceAmount: float64(plan.PriceUsdCents) / 100.0,
		Period:      period,
	})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "nowpayments plan creation failed: " + err.Error()}
	}

	_, err = db.Exec(ctx, `UPDATE plans SET provider_plan_id = $1 WHERE id = $2`, resp.ProviderPlanID, id)
	if err != nil {
		return nil, err
	}

	return &AutoLinkPlanResponse{
		ProviderPlanID: resp.ProviderPlanID,
		PlanName:       displayName,
	}, nil
}

type UnlinkAllPlansResponse struct {
	Count int `json:"count"`
}

//encore:api auth method=POST path=/admin/plans/unlink
func UnlinkAllPlans(ctx context.Context) (*UnlinkAllPlansResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	result, err := db.Exec(ctx, `UPDATE plans SET provider_plan_id = NULL WHERE provider_plan_id IS NOT NULL AND provider_plan_id != ''`)
	if err != nil {
		return nil, err
	}
	count := result.RowsAffected()
	return &UnlinkAllPlansResponse{Count: int(count)}, nil
}

func intervalToPeriod(interval string) string {
	switch strings.ToLower(interval) {
	case "daily":
		return "day"
	case "weekly":
		return "week"
	case "monthly":
		return "month"
	case "yearly":
		return "year"
	default:
		return "month"
	}
}
