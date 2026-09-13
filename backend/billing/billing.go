package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	myauth "comics-galore/backend/auth"
	"comics-galore/backend/nowpayments"
	"comics-galore/backend/tiers"

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

	// WorkerSecret authenticates the Temporal worker when it calls the
	// worker-facing endpoints. Empty = disabled (local dev).
	WorkerSecret string

	// Temporal connection (Temporal Cloud uses mTLS cert/key OR an API key;
	// local dev uses the address/namespace only, defaulting to localhost).
	TemporalAddress   string
	TemporalNamespace string
	TemporalAPIKey    string
	TemporalCert      string
	TemporalKey       string
}

var provider nowpayments.PaymentsProvider

// Cross-service dependencies, overridable in tests. Encore API functions
// can't be referenced directly, so each is wrapped in a closure.
var (
	getPlan = func(ctx context.Context, id string) (*tiers.PlanDetail, error) {
		return tiers.GetPlan(ctx, id)
	}
	ensureSubPartnerID = func(ctx context.Context, p *myauth.EnsureSubPartnerIDParams) (*myauth.SubPartnerIDResponse, error) {
		return myauth.EnsureSubPartnerID(ctx, p)
	}
)

func init() {
	provider = nowpayments.NewProvider(secrets.NowPaymentsAPIKey, secrets.NowPaymentsIPNKey,
		secrets.NowPaymentsEmail, secrets.NowPaymentsPassword)
}

func buildCallbackURL(path string) string {
	return nowpayments.BuildCallbackURL(secrets.NgrokURL, path)
}

// normalizeSubStatus maps a raw NowPayments subscription/payment status to the
// local subscriptions.status enum.
func normalizeSubStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "waiting", "waiting_pay", "waitingpay", "partially_paid", "partiallypaid":
		return "waiting_pay"
	case "active", "finished":
		return "active"
	case "expired":
		return "expired"
	case "failed":
		return "failed"
	case "canceled", "cancelled":
		return "cancelled"
	default:
		return "pending"
	}
}

// normalizeDepositStatus maps a NowPayments payment_status to the local
// deposits.status enum.
func normalizeDepositStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "finished", "confirmed", "sending":
		return "completed"
	case "failed", "refunded":
		return "failed"
	case "expired":
		return "expired"
	default:
		return "pending" // waiting, confirming, partially_paid, ...
	}
}

// ----- Estimate Price -----

type EstimatePriceParams struct {
	PlanID string `json:"plan_id"`
	Crypto string `json:"crypto"`
}

//encore:api auth method=POST path=/billing/estimate-price
func EstimatePrice(ctx context.Context, p *EstimatePriceParams) (*nowpayments.EstimateResponse, error) {
	plan, err := getPlan(ctx, p.PlanID)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found"}
	}

	return provider.EstimatePrice(ctx, EstimateRequest{
		Amount:   float64(plan.PriceUsdCents) / 100,
		Currency: "usd",
		Crypto:   p.Crypto,
	})
}

// ----- List Currencies -----

type ListCurrenciesResponse struct {
	Currencies []string `json:"currencies"`
}

//encore:api public method=GET path=/billing/currencies
func ListCurrencies(ctx context.Context) (*ListCurrenciesResponse, error) {
	currencies, err := provider.ListCurrencies(ctx)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "failed to list currencies: " + err.Error()}
	}

	// Restrict to the checkout menu configured in app settings (when set).
	if cfg, err := myauth.GetAppConfig(ctx); err == nil && strings.TrimSpace(cfg.CryptoCurrencies) != "" {
		currencies = filterCurrencies(currencies, cfg.CryptoCurrencies)
	}

	return &ListCurrenciesResponse{Currencies: currencies}, nil
}

// filterCurrencies keeps only the configured codes, preserving the configured
// order. Both sides are trimmed and lowercased.
func filterCurrencies(available []string, configured string) []string {
	allowed := make(map[string]bool)
	var order []string
	for _, c := range strings.Split(configured, ",") {
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "" || allowed[c] {
			continue
		}
		allowed[c] = true
		order = append(order, c)
	}

	availableSet := make(map[string]bool, len(available))
	for _, a := range available {
		availableSet[strings.ToLower(strings.TrimSpace(a))] = true
	}

	var out []string
	for _, c := range order {
		if availableSet[c] {
			out = append(out, c)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

// ----- Check Balance -----

type CheckBalanceResponse struct {
	Balances map[string]BalanceEntry `json:"balances"`
}

type CreateSubResponse struct {
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
}

type CreateDepositResponse struct {
	DepositID    string  `json:"deposit_id"`
	PayAddress   string  `json:"pay_address"`
	PayAmount    float64 `json:"pay_amount"`
	PayCurrency  string  `json:"pay_currency"`
	PayinExtraID string  `json:"payin_extra_id,omitempty"`
	Network      string  `json:"network,omitempty"`
	QrDataURL    string  `json:"qr_data_url,omitempty"`
	PaymentURI   string  `json:"payment_uri,omitempty"`
}

// ----- Quota Boost -----

type BoostOption struct {
	Downloads int     `json:"downloads"`
	PriceUSD  float64 `json:"price_usd"`
}

type BoostOptionsResponse struct {
	Boosts []BoostOption `json:"boosts"`
}

//encore:api public method=GET path=/billing/quota-boosts
func GetBoostOptions(ctx context.Context) (*BoostOptionsResponse, error) {
	cfg, err := myauth.GetBoostConfig(ctx)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "quota config unavailable"}
	}
	return &BoostOptionsResponse{Boosts: []BoostOption{
		{Downloads: cfg.Boost1Downloads, PriceUSD: cfg.Boost1Price},
		{Downloads: cfg.Boost2Downloads, PriceUSD: cfg.Boost2Price},
		{Downloads: cfg.Boost3Downloads, PriceUSD: cfg.Boost3Price},
	}}, nil
}

// ----- Subscription Webhook -----

// processSubscriptionWebhook runs the full subscription webhook processing
// (signature verification included) and returns the HTTP status + body. Kept
// separate from the raw handler so the webhook simulator can invoke the same
// path in-process.
func processSubscriptionWebhook(ctx context.Context, body []byte, sig string) (int, string) {
	if sig == "" || !verifySignature(secrets.NowPaymentsIPNKey, body, sig) {
		return http.StatusUnauthorized, "invalid signature"
	}

	db.Exec(ctx, `
		INSERT INTO webhook_events (provider, event_type, payload)
		VALUES ('nowpayments', 'subscription', $1)
	`, string(body))

	var event struct {
		ID            json.Number `json:"id"`
		Status        string      `json:"status"`
		PaymentStatus string      `json:"payment_status"`
		Amount        json.Number `json:"amount"`
		Currency      string      `json:"currency"`
		PriceAmount   json.Number `json:"price_amount"`
	}
	json.Unmarshal(body, &event)

	// Look up subscription info (billing-local); plan details come from tiers
	var sub struct {
		SubID      string
		UserID     string
		Tier       string
		PlanID     string
		CheckoutID string
	}
	err := db.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.tier, s.plan_id, COALESCE(s.checkout_id, '')
		FROM subscriptions s
		WHERE s.provider_subscription_id = $1
		ORDER BY s.created_at DESC LIMIT 1
	`, event.ID.String()).Scan(&sub.SubID, &sub.UserID, &sub.Tier, &sub.PlanID, &sub.CheckoutID)

	interval := "monthly"
	priceCents := 0
	if plan, perr := getPlan(ctx, sub.PlanID); perr == nil {
		interval = plan.Interval
		priceCents = plan.PriceUsdCents
	}

	status := strings.ToLower(event.Status)
	if status == "" {
		status = strings.ToLower(event.PaymentStatus)
	}
	amtCrypto, _ := event.Amount.Float64()
	amtUSD, _ := event.PriceAmount.Float64()

	// Calculate amount_usd_cents from price_amount in payload or from plan
	usdCents := int(amtUSD * 100)
	if usdCents == 0 {
		usdCents = priceCents
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
			sub.Tier, interval, amtCrypto, event.Currency, usdCents,
			status, string(body))
	}

	// Signal the durable SubscribeWorkflow (best-effort). It activates on
	// "finished" and expires on timeout or a non-finished status.
	if err == nil && sub.CheckoutID != "" {
		if wfErr := signalSubscribeSubscription(ctx, sub.CheckoutID, status); wfErr != nil {
			log.Printf("signal subscribe workflow (best-effort): %v", wfErr)
		}
	}

	return http.StatusOK, `{"ok":true}`
}

//encore:api public raw path=/webhooks/nowpayments/subscription method=POST
func SubscriptionWebhook(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	status, respBody := processSubscriptionWebhook(req.Context(), body, req.Header.Get("x-nowpayments-sig"))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(respBody))
}

// ActivateSubscription marks a subscription active and promotes the user's
// tier. Called by the Temporal worker's activity; idempotent.
//
//encore:api public method=POST path=/billing/subscriptions/:id/activate
func ActivateSubscription(ctx context.Context, id string) error {
	if err := requireWorker(); err != nil {
		return err
	}
	var userID, tier string
	if err := db.QueryRow(ctx, `SELECT user_id, tier FROM subscriptions WHERE id = $1`, id).Scan(&userID, &tier); err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "subscription not found"}
		}
		return err
	}

	db.Exec(ctx, `UPDATE subscriptions SET active = true, status = 'active', activated_at = now(), updated_at = now() WHERE id = $1`, id)
	return myauth.SetUserTier(ctx, &myauth.SetUserTierParams{UserID: userID, Tier: tier})
}

// ExpireSubscription expires a subscription and downgrades the user to the
// free tier. Called by the Temporal worker's activity; idempotent.
//
//encore:api public method=POST path=/billing/subscriptions/:id/expire
func ExpireSubscription(ctx context.Context, id string) error {
	if err := requireWorker(); err != nil {
		return err
	}
	var userID string
	if err := db.QueryRow(ctx, `SELECT user_id FROM subscriptions WHERE id = $1`, id).Scan(&userID); err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "subscription not found"}
		}
		return err
	}

	if err := myauth.SetUserTier(ctx, &myauth.SetUserTierParams{UserID: userID, Tier: "free"}); err != nil {
		return err
	}
	db.Exec(ctx, `UPDATE subscriptions SET status = 'expired', active = false, updated_at = now() WHERE id = $1`, id)
	return nil
}

// ----- Deposit Webhook -----

// processDepositWebhook runs the full deposit webhook processing (signature
// verification included) and returns the HTTP status + body.
func processDepositWebhook(ctx context.Context, body []byte, sig, depositID string) (int, string) {
	if sig == "" || !verifySignature(secrets.NowPaymentsIPNKey, body, sig) {
		return http.StatusUnauthorized, "invalid signature"
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

	status := normalizeDepositStatus(event.PaymentStatus)

	// Resolve the deposit row id + checkout_id.
	actualID := depositID
	if actualID == "" {
		db.QueryRow(ctx, `SELECT id FROM deposits WHERE provider_deposit_id = $1`, event.PaymentID.String()).Scan(&actualID)
	}
	checkoutID := ""
	if actualID != "" {
		db.QueryRow(ctx, `SELECT COALESCE(checkout_id, '') FROM deposits WHERE id = $1`, actualID).Scan(&checkoutID)
	}

	// Record the raw payload on the deposit (audit).
	if actualID != "" {
		db.Exec(ctx, `UPDATE deposits SET raw_payload = $1, updated_at = now() WHERE id = $2`, string(body), actualID)
	}

	// Signal the durable workflow; it completes/expires via activities.
	if checkoutID != "" {
		if wfErr := signalSubscribeDeposit(ctx, checkoutID, status); wfErr != nil {
			log.Printf("signal subscribe deposit (best-effort): %v", wfErr)
		}
	}

	return http.StatusOK, `{"ok":true}`
}

//encore:api public raw path=/webhooks/nowpayments/deposit method=POST
func DepositWebhook(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	status, respBody := processDepositWebhook(req.Context(), body, req.Header.Get("x-nowpayments-sig"), req.URL.Query().Get("deposit_id"))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(respBody))
}

// ----- Billing Stats -----

type RevenueByTier struct {
	Tier    string `json:"tier"`
	Revenue int    `json:"revenue"`
}

type BillingStats struct {
	TotalRevenue  int             `json:"total_revenue"`
	ActiveRevenue int             `json:"active_revenue"`
	RecentRevenue int             `json:"recent_revenue"`
	TotalDeposits int             `json:"total_deposits"`
	TotalPayments int             `json:"total_payments"`
	ActiveSubs    int             `json:"active_subscriptions"`
	RevenueByTier []RevenueByTier `json:"revenue_by_tier"`
}

//encore:api private
func GetBillingStats(ctx context.Context) (*BillingStats, error) {
	var stats BillingStats
	db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_usd_cents), 0) FROM payments WHERE status = 'finished'`).Scan(&stats.TotalRevenue)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_usd_cents), 0) FROM payments WHERE status = 'finished' AND interval = 'monthly'`).Scan(&stats.ActiveRevenue)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_usd_cents), 0) FROM payments WHERE status = 'finished' AND created_at > now() - interval '30 days'`).Scan(&stats.RecentRevenue)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM deposits`).Scan(&stats.TotalDeposits)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE status = 'finished'`).Scan(&stats.TotalPayments)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE active = true`).Scan(&stats.ActiveSubs)

	rows, err := db.Query(ctx, `
		SELECT tier, COALESCE(SUM(amount_usd_cents), 0) AS revenue
		FROM payments WHERE status = 'finished'
		GROUP BY tier ORDER BY revenue DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r RevenueByTier
			if rows.Scan(&r.Tier, &r.Revenue) == nil {
				stats.RevenueByTier = append(stats.RevenueByTier, r)
			}
		}
	}

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
	if page <= 0 {
		page = 1
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := "created_at"
	sortDir := "DESC"
	switch p.Sort {
	case "created_at":
		sortCol = "created_at"
	case "status":
		sortCol = "status"
	case "tier":
		sortCol = "tier"
	case "expires_at":
		sortCol = "expires_at"
	}
	if strings.ToLower(p.SortDir) == "asc" {
		sortDir = "ASC"
	}

	where := "WHERE tier <> 'free' AND (user_id::text ILIKE $1 OR plan_id::text ILIKE $1)"
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
		SELECT id, user_id, plan_id, COALESCE(provider_subscription_id, ''), status, active, tier, activated_at, expires_at, created_at
		FROM subscriptions %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []AdminSubscription
	for rows.Next() {
		var s AdminSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.PlanID, &s.ProviderSubscriptionID, &s.Status,
			&s.Active, &s.Tier, timePtr(&s.ActivatedAt), timePtr(&s.ExpiresAt), &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return &AdminSubList{Subscriptions: subs, Total: total}, rows.Err()
}

// ----- Admin Deposits -----

type AdminListDepositsParams struct {
	Page         int    `query:"page"`
	Limit        int    `query:"limit"`
	Search       string `query:"search"`
	Sort         string `query:"sort"`
	SortDir      string `query:"sort_dir"`
	FilterStatus string `query:"filter_status"`
}

//encore:api auth method=GET path=/admin/deposits
func AdminListDeposits(ctx context.Context, p *AdminListDepositsParams) (*AdminDepositList, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := p.Page
	if page <= 0 {
		page = 1
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := "created_at"
	sortDir := "DESC"
	switch p.Sort {
	case "created_at":
		sortCol = "created_at"
	case "status":
		sortCol = "status"
	case "amount_usd_cents":
		sortCol = "amount_usd_cents"
	case "completed_at":
		sortCol = "completed_at"
	}
	if strings.ToLower(p.SortDir) == "asc" {
		sortDir = "ASC"
	}

	where := "WHERE (user_id::text ILIKE $1 OR pay_address ILIKE $1 OR provider_deposit_id ILIKE $1)"
	args := []interface{}{search}
	argIdx := 2

	if p.FilterStatus != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, p.FilterStatus)
		argIdx++
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM deposits `+where, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT id, user_id, COALESCE(provider_deposit_id,''), COALESCE(amount_crypto,''),
			currency_crypto, COALESCE(amount_usd_cents,0), status,
			COALESCE(pay_address,''), created_at, completed_at
		FROM deposits %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []AdminDeposit
	for rows.Next() {
		var d AdminDeposit
		if err := rows.Scan(&d.ID, &d.UserID, &d.ProviderDepositID, &d.AmountCrypto,
			&d.CurrencyCrypto, &d.AmountUsdCents, &d.Status,
			&d.PayAddress, &d.CreatedAt, timePtr(&d.CompletedAt)); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}

	return &AdminDepositList{Deposits: deps, Total: total}, rows.Err()
}

// ----- Admin Payments -----

type AdminListPaymentsParams struct {
	Page         int    `query:"page"`
	Limit        int    `query:"limit"`
	Search       string `query:"search"`
	Sort         string `query:"sort"`
	SortDir      string `query:"sort_dir"`
	FilterStatus string `query:"filter_status"`
	FilterTier   string `query:"filter_tier"`
}

//encore:api auth method=GET path=/admin/payments
func AdminListPayments(ctx context.Context, p *AdminListPaymentsParams) (*AdminPaymentList, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := p.Page
	if page <= 0 {
		page = 1
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := "created_at"
	sortDir := "DESC"
	switch p.Sort {
	case "created_at":
		sortCol = "created_at"
	case "status":
		sortCol = "status"
	case "tier":
		sortCol = "tier"
	case "amount_usd_cents":
		sortCol = "amount_usd_cents"
	}
	if strings.ToLower(p.SortDir) == "asc" {
		sortDir = "ASC"
	}

	where := "WHERE (user_id::text ILIKE $1 OR provider_payment_id ILIKE $1)"
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

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM payments `+where, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT id, provider_payment_id, COALESCE(user_id::text,''), COALESCE(subscription_id::text,''),
			tier, COALESCE(interval,''), COALESCE(amount_crypto,0), COALESCE(currency_crypto,''),
			COALESCE(amount_usd_cents,0), status, created_at
		FROM payments %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pays []AdminPayment
	for rows.Next() {
		var pm AdminPayment
		var amountCrypto float64
		if err := rows.Scan(&pm.ID, &pm.ProviderPaymentID, &pm.UserID, &pm.SubscriptionID,
			&pm.Tier, &pm.Interval, &amountCrypto, &pm.CurrencyCrypto,
			&pm.AmountUsdCents, &pm.Status, &pm.CreatedAt); err != nil {
			return nil, err
		}
		pm.AmountCrypto = strconv.FormatFloat(amountCrypto, 'f', -1, 64)
		pays = append(pays, pm)
	}

	return &AdminPaymentList{Payments: pays, Total: total}, rows.Err()
}

// ----- Types -----

type AdminSubList struct {
	Subscriptions []AdminSubscription `json:"subscriptions"`
	Total         int                 `json:"total"`
}

type AdminSubscription struct {
	ID                     string    `json:"id"`
	UserID                 string    `json:"user_id"`
	PlanID                 string    `json:"plan_id"`
	ProviderSubscriptionID string    `json:"provider_subscription_id"`
	Status                 string    `json:"status"`
	Active                 bool      `json:"active"`
	Tier                   string    `json:"tier"`
	ActivatedAt            time.Time `json:"activated_at,omitempty"`
	ExpiresAt              time.Time `json:"expires_at,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}

type AdminDepositList struct {
	Deposits []AdminDeposit `json:"deposits"`
	Total    int            `json:"total"`
}

type AdminDeposit struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	ProviderDepositID string    `json:"provider_deposit_id"`
	AmountCrypto      string    `json:"amount_crypto"`
	CurrencyCrypto    string    `json:"currency_crypto"`
	AmountUsdCents    int       `json:"amount_usd_cents"`
	Status            string    `json:"status"`
	PayAddress        string    `json:"pay_address"`
	CreatedAt         time.Time `json:"created_at"`
	CompletedAt       time.Time `json:"completed_at,omitempty"`
}

type AdminPaymentList struct {
	Payments []AdminPayment `json:"payments"`
	Total    int            `json:"total"`
}

type AdminPayment struct {
	ID                string    `json:"id"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	UserID            string    `json:"user_id"`
	SubscriptionID    string    `json:"subscription_id"`
	Tier              string    `json:"tier"`
	Interval          string    `json:"interval"`
	AmountCrypto      string    `json:"amount_crypto"`
	CurrencyCrypto    string    `json:"currency_crypto"`
	AmountUsdCents    int       `json:"amount_usd_cents"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
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
