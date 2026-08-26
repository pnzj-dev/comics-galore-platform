package billing

import (
	"context"
	"time"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

// ----- Coupons -----

type Coupon struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	PercentOff int        `json:"percent_off"`
	Tier       string     `json:"tier"`
	MaxUses    int        `json:"max_uses"`
	Used       int        `json:"used"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateCouponParams struct {
	Code       string `json:"code"`
	PercentOff int    `json:"percent_off"`
	Tier       string `json:"tier"`
	MaxUses    int    `json:"max_uses"`
}

type ListCouponsResponse struct {
	Coupons []Coupon `json:"coupons"`
}

//encore:api auth method=GET path=/admin/coupons
func AdminListCoupons(ctx context.Context) (*ListCouponsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `SELECT id, code, percent_off, tier, max_uses, used, expires_at, created_at FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []Coupon
	for rows.Next() {
		var c Coupon
		if err := rows.Scan(&c.ID, &c.Code, &c.PercentOff, &c.Tier, &c.MaxUses, &c.Used, &c.ExpiresAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	if coupons == nil {
		coupons = []Coupon{}
	}
	return &ListCouponsResponse{Coupons: coupons}, rows.Err()
}

//encore:api auth method=POST path=/admin/coupons
func AdminCreateCoupon(ctx context.Context, p *CreateCouponParams) (*Coupon, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	if p.Code == "" || p.PercentOff <= 0 || p.PercentOff > 100 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "code and a 1-100 percent_off are required"}
	}

	var c Coupon
	err := db.QueryRow(ctx, `
		INSERT INTO coupons (code, percent_off, tier, max_uses)
		VALUES ($1, $2, $3, $4)
		RETURNING id, code, percent_off, tier, max_uses, used, expires_at, created_at
	`, p.Code, p.PercentOff, p.Tier, p.MaxUses).Scan(&c.ID, &c.Code, &c.PercentOff, &c.Tier, &c.MaxUses, &c.Used, &c.ExpiresAt, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ----- Manual grant / revoke -----

type GrantSubscriptionParams struct {
	UserID     string `json:"user_id"`
	Tier       string `json:"tier"`
	DurationDays int  `json:"duration_days"`
}

//encore:api auth method=POST path=/admin/subscriptions/:id/grant
func AdminGrantSubscription(ctx context.Context, id string, p *GrantSubscriptionParams) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	if p.Tier == "" || p.UserID == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "user_id and tier are required"}
	}
	days := p.DurationDays
	if days <= 0 {
		days = 30
	}

	_, err := db.Exec(ctx, `
		INSERT INTO subscriptions (id, user_id, plan_id, provider, provider_subscription_id, status, active, tier, activated_at, expires_at)
		VALUES ($1, $2, gen_random_uuid(), 'manual', $1, 'active', true, $3, now(), now() + make_interval(days => $4))
	`, id, p.UserID, p.Tier, days)
	if err != nil {
		return err
	}

	// Update the user's tier via the auth service (ADR 0016).
	myauth.SetUserTier(ctx, &myauth.SetUserTierParams{UserID: p.UserID, Tier: p.Tier})
	return nil
}

//encore:api auth method=POST path=/admin/subscriptions/:id/revoke
func AdminRevokeSubscription(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `UPDATE subscriptions SET active = false, status = 'cancelled', updated_at = now() WHERE id = $1`, id)
	return err
}

// ----- User-facing subscription management -----

type MySubscription struct {
	ID       string     `json:"id"`
	Tier     string     `json:"tier"`
	Status   string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

//encore:api auth method=GET path=/billing/my-subscription
func GetMySubscription(ctx context.Context) (*MySubscription, error) {
	ad := auth.Data().(*myauth.AuthData)
	var s MySubscription
	err := db.QueryRow(ctx, `
		SELECT id, tier, status, expires_at
		FROM subscriptions
		WHERE user_id = $1 AND active = true
		ORDER BY created_at DESC LIMIT 1
	`, ad.UserID).Scan(&s.ID, &s.Tier, &s.Status, &s.ExpiresAt)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

//encore:api auth method=POST path=/billing/cancel-subscription
func CancelMySubscription(ctx context.Context) error {
	ad := auth.Data().(*myauth.AuthData)
	res, err := db.Exec(ctx, `
		UPDATE subscriptions SET active = false, status = 'cancelled', updated_at = now()
		WHERE user_id = $1 AND active = true
	`, ad.UserID)
	if err != nil {
		return err
	}
	if res.RowsAffected() > 0 {
		myauth.SetUserTier(ctx, &myauth.SetUserTierParams{UserID: ad.UserID, Tier: "free"})
	}
	return nil
}

// ----- Past-due / failed payments -----

type PastDuePayment struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	SubscriptionID string    `json:"subscription_id"`
	Tier           string    `json:"tier"`
	AmountUsdCents int       `json:"amount_usd_cents"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type ListPastDueResponse struct {
	Payments []PastDuePayment `json:"payments"`
}

//encore:api auth method=GET path=/admin/payments/past-due
func AdminPastDuePayments(ctx context.Context) (*ListPastDueResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `
		SELECT id, COALESCE(user_id::text, ''), COALESCE(subscription_id::text, ''), tier,
			COALESCE(amount_usd_cents, 0), status, created_at
		FROM payments WHERE status IN ('failed', 'expired')
		ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []PastDuePayment
	for rows.Next() {
		var p PastDuePayment
		if err := rows.Scan(&p.ID, &p.UserID, &p.SubscriptionID, &p.Tier, &p.AmountUsdCents, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []PastDuePayment{}
	}
	return &ListPastDueResponse{Payments: payments}, rows.Err()
}
