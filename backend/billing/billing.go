package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("billingdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var secrets struct {
	NowPaymentsAPIKey   string
	NowPaymentsIPNKey   string
	NowPaymentsEmail    string
	NowPaymentsPassword string
	NgrokURL            string
}

var provider PaymentsProvider

func init() {
	provider = NewNowPaymentsProvider(secrets.NowPaymentsAPIKey, secrets.NowPaymentsIPNKey,
		secrets.NowPaymentsEmail, secrets.NowPaymentsPassword)
}

func buildCallbackURL(host, path string) string {
	if (strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")) && secrets.NgrokURL != "" {
		return strings.TrimRight(secrets.NgrokURL, "/") + path
	}
	scheme := "https"
	if host == "" || strings.Contains(host, "localhost") || strings.Contains(host, ":4000") {
		scheme = "http"
	}
	if host == "" {
		host = "localhost:4000"
	}
	return scheme + "://" + host + path
}

// ----- Estimate Price -----

type EstimatePriceParams struct {
	PlanID string `json:"plan_id"`
	Crypto string `json:"crypto"`
}

//encore:api auth method=POST path=/billing/estimate-price
func EstimatePrice(ctx context.Context, p *EstimatePriceParams) (*EstimateResponse, error) {
	var priceCents int
	err := db.QueryRow(ctx, `SELECT price_usd_cents FROM plans WHERE id = $1`, p.PlanID).Scan(&priceCents)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found"}
	}

	return provider.EstimatePrice(ctx, EstimateRequest{
		Amount:   float64(priceCents) / 100,
		Currency: "usd",
		Crypto:   p.Crypto,
	})
}

// ----- Check Balance -----

type CheckBalanceResponse struct {
	Balances map[string]BalanceEntry `json:"balances"`
}

//encore:api auth method=GET path=/billing/check-balance
func CheckBalance(ctx context.Context) (*CheckBalanceResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var subID string
	err := db.QueryRow(ctx, `SELECT COALESCE(sub_partner_id, '') FROM users WHERE id = $1`, ad.UserID).Scan(&subID)
	if err != nil || subID == "" {
		return nil, &errs.Error{Code: errs.NotFound, Message: "no sub_partner_id configured"}
	}

	balances, err := provider.CheckBalance(ctx, subID)
	if err != nil {
		return nil, err
	}

	return &CheckBalanceResponse{Balances: balances}, nil
}

// ----- Create Subscription (atomic, Step 4A) -----

type CreateSubParams struct {
	PlanID string `json:"plan_id"`
}

type CreateSubResponse struct {
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
}

//encore:api auth method=POST path=/billing/create-subscription
func CreateSubscription(ctx context.Context, p *CreateSubParams) (*CreateSubResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var subPartnerID, providerPlanID, interval, tierName string
	var priceCents int
	err := db.QueryRow(ctx, `
		SELECT COALESCE(u.sub_partner_id,''), COALESCE(p.provider_plan_id,''), p.interval, t.name, p.price_usd_cents
		FROM plans p JOIN tiers t ON t.id = p.tier_id
		CROSS JOIN (SELECT sub_partner_id FROM users WHERE id = $1) u
		WHERE p.id = $2
	`, ad.UserID, p.PlanID).Scan(&subPartnerID, &providerPlanID, &interval, &tierName, &priceCents)
	if err != nil || subPartnerID == "" {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found or no sub_partner_id"}
	}
	if providerPlanID == "" {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "this plan is not yet configured with a provider plan ID — contact admin"}
	}

	// Call NowPayments to create the subscription
	npResp, err := provider.CreateSubscription(ctx, SubscriptionRequest{
		PlanID:             p.PlanID,
		SubPartnerID:       subPartnerID,
		SubscriptionPlanID: providerPlanID,
	})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "nowpayments subscription creation failed: " + err.Error()}
	}

	// Atomically save locally
	expiresAt := time.Now()
	switch interval {
	case "monthly":
		expiresAt = expiresAt.AddDate(0, 1, 0)
	case "quarterly":
		expiresAt = expiresAt.AddDate(0, 3, 0)
	case "semesterly":
		expiresAt = expiresAt.AddDate(0, 6, 0)
	case "yearly":
		expiresAt = expiresAt.AddDate(1, 0, 0)
	default:
		expiresAt = expiresAt.AddDate(0, 1, 0)
	}

	var subID string
	err = db.QueryRow(ctx, `
		INSERT INTO subscriptions (user_id, plan_id, provider, provider_subscription_id,
			tier, active, expires_at)
		VALUES ($1, $2, 'nowpayments', $3, $4, false, $5)
		RETURNING id
	`, ad.UserID, p.PlanID, npResp.SubscriptionID, tierName, expiresAt).Scan(&subID)
	if err != nil {
		return nil, err
	}

	return &CreateSubResponse{
		SubscriptionID: subID,
		Status:         npResp.Status,
	}, nil
}

// ----- Create Deposit (Step 4B) -----

type CreateDepositParams struct {
	PlanID    string `json:"plan_id"`
	Crypto    string `json:"crypto"`
	Host      string `header:"Host"`
}

type CreateDepositResponse struct {
	DepositID   string  `json:"deposit_id"`
	PayAddress  string  `json:"pay_address"`
	PayAmount   float64 `json:"pay_amount"`
	PayCurrency string  `json:"pay_currency"`
	PlanID      string  `json:"plan_id"`
}

//encore:api auth method=POST path=/billing/create-deposit
func CreateDeposit(ctx context.Context, p *CreateDepositParams) (*CreateDepositResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var subPartnerID string
	var priceCents int
	err := db.QueryRow(ctx, `
		SELECT COALESCE(u.sub_partner_id,''), p.price_usd_cents
		FROM users u, plans p
		WHERE u.id = $1 AND p.id = $2
	`, ad.UserID, p.PlanID).Scan(&subPartnerID, &priceCents)
	if err != nil || subPartnerID == "" {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found or no sub_partner_id"}
	}

	var depositID string
	err = db.QueryRow(ctx, `
		INSERT INTO deposits (user_id, currency_crypto, amount_usd_cents)
		VALUES ($1, $2, $3) RETURNING id
	`, ad.UserID, p.Crypto, priceCents).Scan(&depositID)
	if err != nil {
		return nil, err
	}

	callbackURL := buildCallbackURL(p.Host, "/webhooks/nowpayments/deposit?deposit_id="+depositID)

	npResp, err := provider.CreateDeposit(ctx, DepositRequest{
		Crypto:        p.Crypto,
		AmountUSD:     float64(priceCents) / 100,
		SubPartnerID:  subPartnerID,
		IPNCallbackURL: callbackURL,
	})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "nowpayments deposit creation failed: " + err.Error()}
	}

	db.Exec(ctx, `
		UPDATE deposits SET provider_deposit_id = $1, pay_address = $2, amount_crypto = $3
		WHERE id = $4
	`, npResp.PaymentID, npResp.PayAddress, fmtNum(npResp.PayAmount), depositID)

	return &CreateDepositResponse{
		DepositID:   depositID,
		PayAddress:  npResp.PayAddress,
		PayAmount:   npResp.PayAmount,
		PayCurrency: npResp.PayCurrency,
		PlanID:      p.PlanID,
	}, nil
}

// ----- Poll Subscription -----

type PollSubResponse struct {
	Active bool `json:"active"`
}

//encore:api auth method=GET path=/billing/subscription/:id/poll
func PollSubscription(ctx context.Context, id string) (*PollSubResponse, error) {
	var finished bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM payments WHERE subscription_id = $1 AND status = 'finished'
		)
	`, id).Scan(&finished)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "subscription not found"}
		}
		return nil, err
	}
	return &PollSubResponse{Active: finished}, nil
}

// ----- Poll Deposit -----

type PollDepositResponse struct {
	Completed bool `json:"completed"`
}

//encore:api auth method=GET path=/billing/deposit/:id/poll
func PollDeposit(ctx context.Context, id string) (*PollDepositResponse, error) {
	var status string
	err := db.QueryRow(ctx, `SELECT status FROM deposits WHERE id = $1`, id).Scan(&status)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "deposit not found"}
		}
		return nil, err
	}
	return &PollDepositResponse{Completed: status == "completed"}, nil
}

// ----- Subscription Webhook -----

//encore:api public raw path=/webhooks/nowpayments/subscription method=POST
func SubscriptionWebhook(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	sig := req.Header.Get("x-nowpayments-sig")
	if sig == "" || !verifySignature(secrets.NowPaymentsIPNKey, body, sig) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	db.Exec(ctx, `
		INSERT INTO webhook_events (provider, event_type, payload)
		VALUES ('nowpayments', 'subscription', $1)
	`, string(body))

	var event struct {
		ID            json.Number `json:"id"`
		PaymentStatus string      `json:"payment_status"`
		Amount        json.Number `json:"amount"`
		Currency      string      `json:"currency"`
		PriceAmount   json.Number `json:"price_amount"`
	}
	json.Unmarshal(body, &event)

	// Look up subscription, plan, and user info
	var sub struct {
		SubID      string
		UserID     string
		Tier       string
		PlanID     string
		Interval   string
		PriceCents int
	}
	err = db.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.tier, s.plan_id,
			COALESCE(p.interval, 'monthly'),
			COALESCE(p.price_usd_cents, 0)
		FROM subscriptions s
		LEFT JOIN plans p ON p.id = s.plan_id
		WHERE s.provider_subscription_id = $1
		ORDER BY s.created_at DESC LIMIT 1
	`, event.ID.String()).Scan(&sub.SubID, &sub.UserID, &sub.Tier, &sub.PlanID, &sub.Interval, &sub.PriceCents)

	status := strings.ToLower(event.PaymentStatus)
	amtCrypto, _ := event.Amount.Float64()
	amtUSD, _ := event.PriceAmount.Float64()

	// Calculate amount_usd_cents from price_amount in payload or from plan
	usdCents := int(amtUSD * 100)
	if usdCents == 0 {
		usdCents = sub.PriceCents
	}

		if err == nil && event.ID.String() != "" {
		db.Exec(ctx, `
			INSERT INTO payments
				(provider, provider_payment_id, user_id, subscription_id, plan_id,
				 tier, interval, amount_crypto, currency_crypto, amount_usd_cents,
				 status, raw_payload, updated_at)
			VALUES
				('nowpayments', $1, $2, $3, $4,
				 $5, $6, $7, $8, $9,
				 $10, $11, now())
			ON CONFLICT (provider_payment_id) DO UPDATE SET
				amount_crypto = EXCLUDED.amount_crypto,
				amount_usd_cents = EXCLUDED.amount_usd_cents,
				status = EXCLUDED.status,
				raw_payload = EXCLUDED.raw_payload,
				updated_at = now()
		`, event.ID.String(), sub.UserID, sub.SubID, sub.PlanID,
			sub.Tier, sub.Interval, amtCrypto, event.Currency, usdCents,
			status, string(body))
	}

	if status == "finished" {
		if err == nil {
			db.Exec(ctx, `UPDATE subscriptions SET active = true, status = 'active' WHERE provider_subscription_id = $1`, event.ID.String())
			db.Exec(ctx, `UPDATE users SET tier = $1 WHERE id = $2`, sub.Tier, sub.UserID)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// ----- Deposit Webhook -----

//encore:api public raw path=/webhooks/nowpayments/deposit method=POST
func DepositWebhook(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	sig := req.Header.Get("x-nowpayments-sig")
	if sig == "" || !verifySignature(secrets.NowPaymentsIPNKey, body, sig) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	db.Exec(ctx, `
		INSERT INTO webhook_events (provider, event_type, payload)
		VALUES ('nowpayments', 'deposit', $1)
	`, string(body))

	var event struct {
		PaymentStatus string      `json:"payment_status"`
		PaymentID     json.Number `json:"payment_id"`
	}
	json.Unmarshal(body, &event)

	status := strings.ToLower(event.PaymentStatus)
	depositID := req.URL.Query().Get("deposit_id")

	// Always update deposit by deposit_id (exact match from callback URL)
	if depositID != "" {
		db.Exec(ctx, `
			UPDATE deposits SET status = $1, raw_payload = $2, updated_at = now()
			WHERE id = $3
		`, status, string(body), depositID)
	}

	// Also try matching by provider_deposit_id
	db.Exec(ctx, `
		UPDATE deposits SET status = $1, raw_payload = $2, updated_at = now()
		WHERE provider_deposit_id = $3
	`, status, string(body), event.PaymentID.String())

	// If finished, also set completed_at
	if status == "finished" {
		if depositID != "" {
			db.Exec(ctx, `UPDATE deposits SET completed_at = now() WHERE id = $1`, depositID)
		}
		db.Exec(ctx, `UPDATE deposits SET completed_at = now() WHERE provider_deposit_id = $1`, event.PaymentID.String())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// ----- Billing Stats -----

type BillingStats struct {
	TotalRevenue  int `json:"total_revenue"`
	ActiveRevenue int `json:"active_revenue"`
	RecentRevenue int `json:"recent_revenue"`
}

//encore:api auth method=GET path=/admin/billing-stats
func GetBillingStats(ctx context.Context) (*BillingStats, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var stats BillingStats
	db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_usd_cents), 0) FROM payments WHERE status = 'finished'`).Scan(&stats.TotalRevenue)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_usd_cents), 0) FROM payments WHERE status = 'finished' AND interval = 'monthly'`).Scan(&stats.ActiveRevenue)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_usd_cents), 0) FROM payments WHERE status = 'finished' AND created_at > now() - interval '30 days'`).Scan(&stats.RecentRevenue)

	return &stats, nil
}

// ----- Admin -----

type AdminListSubscriptionsParams struct {
	Page         int    `query:"page"`
	Limit        int    `query:"limit"`
	Search       string `query:"search"`
	Sort         string `query:"sort"`
	SortDir      string `query:"sort_dir"`
	FilterStatus string `query:"filter_status"`
	FilterTier   string `query:"filter_tier"`
	FilterUserID string `query:"filter_user_id"`
}

//encore:api auth method=GET path=/admin/subscriptions
func AdminListSubscriptions(ctx context.Context, p *AdminListSubscriptionsParams) (*AdminSubList, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := p.Page
	if page <= 0 { page = 1 }
	limit := p.Limit
	if limit <= 0 { limit = 20 }
	if limit > 100 { limit = 100 }
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := "created_at"
	sortDir := "DESC"
	switch p.Sort {
	case "created_at": sortCol = "created_at"
	case "status": sortCol = "status"
	case "tier": sortCol = "tier"
	case "expires_at": sortCol = "expires_at"
	}
	if strings.ToLower(p.SortDir) == "asc" { sortDir = "ASC" }

	where := "WHERE (user_id::text ILIKE $1 OR plan_id::text ILIKE $1)"
	args := []interface{}{search}
	argIdx := 2

	if p.FilterStatus != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, p.FilterStatus)
		argIdx++
	}
	if p.FilterTier != "" {
		where += fmt.Sprintf(" AND tier = $%d", argIdx)
		args = append(args, p.FilterTier)
		argIdx++
	}
	if p.FilterUserID != "" {
		where += fmt.Sprintf(" AND user_id::text ILIKE $%d", argIdx)
		args = append(args, "%"+p.FilterUserID+"%")
		argIdx++
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions `+where, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT id, user_id, plan_id, status, active, tier, activated_at, expires_at, created_at
		FROM subscriptions %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var subs []AdminSubscription
	for rows.Next() {
		var s AdminSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Status,
			&s.Active, &s.Tier, timePtr(&s.ActivatedAt), timePtr(&s.ExpiresAt), &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return &AdminSubList{Subscriptions: subs, Total: total}, rows.Err()
}

// ----- Types -----

type AdminSubList struct {
	Subscriptions []AdminSubscription `json:"subscriptions"`
	Total         int                 `json:"total"`
}

type AdminSubscription struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	PlanID      string    `json:"plan_id"`
	Status      string    `json:"status"`
	Active      bool      `json:"active"`
	Tier        string    `json:"tier"`
	ActivatedAt time.Time `json:"activated_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ----- Helpers -----

func verifySignature(secret string, body []byte, sig string) bool {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return false
	}
	sorted := sortObject(obj)
	sortedJSON, _ := json.Marshal(sorted)
	h := hmac.New(sha512.New, []byte(secret))
	h.Write(sortedJSON)
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func sortObject(obj map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v, ok := obj[k].(map[string]interface{}); ok {
			result[k] = sortObject(v)
		} else {
			result[k] = obj[k]
		}
	}
	return result
}

func fmtNum(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func CreatePlan(ctx context.Context, req CreatePlanRequest, host string) (*CreatePlanResponse, error) {
	req.IPNCallbackURL = buildCallbackURL(host, "/webhooks/nowpayments/subscription")
	return provider.CreatePlan(ctx, req)
}
