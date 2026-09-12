package billing

import (
	"context"
	"strconv"
	"strings"
	"time"

	myauth "comics-galore/backend/auth"
	myreading "comics-galore/backend/reading"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

// parseTimeLimit converts a NowPayments time_limit value (ISO-8601 datetime or
// Unix seconds) into a time.Time. Falls back to a 30-minute window when empty
// or unparseable.
func parseTimeLimit(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().Add(30 * time.Minute)
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if sec, err := strconv.ParseInt(s, 10, 64); err == nil && sec > 0 {
		return time.Unix(sec, 0)
	}
	return time.Now().Add(30 * time.Minute)
}

// checkBalanceForUser reports whether a user has a non-zero NowPayments balance
// for the given crypto.
func checkBalanceForUser(ctx context.Context, userID string) (*CheckBalanceResponse, error) {
	subResp, err := ensureSubPartnerID(ctx, &myauth.EnsureSubPartnerIDParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	if subResp.SubPartnerID == "" {
		return nil, &errs.Error{Code: errs.NotFound, Message: "no sub_partner_id configured"}
	}
	balances, err := provider.CheckBalance(ctx, subResp.SubPartnerID)
	if err != nil {
		return nil, err
	}
	return &CheckBalanceResponse{Balances: balances}, nil
}

// createSubscriptionForUser creates a NowPayments subscription and stores the
// local row, optionally linked to a checkout workflow. It does not start any
// workflow itself.
func createSubscriptionForUser(ctx context.Context, userID, planID, checkoutID string) (*CreateSubResponse, error) {
	subResp, err := ensureSubPartnerID(ctx, &myauth.EnsureSubPartnerIDParams{UserID: userID})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "ensure sub_partner_id failed: " + err.Error()}
	}
	subPartnerID := subResp.SubPartnerID

	plan, err := getPlan(ctx, planID)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found or no sub_partner_id"}
	}
	providerPlanID := plan.ProviderPlanID
	interval := plan.Interval
	tierName := plan.TierName

	if subPartnerID == "" {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found or no sub_partner_id"}
	}
	if providerPlanID == "" {
		return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "this plan is not yet configured with a provider plan ID — contact admin"}
	}
	if strings.ToLower(tierName) == "free" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "cannot subscribe to the free tier"}
	}

	npResp, err := provider.CreateSubscription(ctx, SubscriptionRequest{
		PlanID:             planID,
		SubPartnerID:       subPartnerID,
		SubscriptionPlanID: providerPlanID,
	})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "nowpayments subscription creation failed: " + err.Error()}
	}

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
			tier, status, active, expires_at, checkout_id)
		VALUES ($1, $2, 'nowpayments', $3, $4, $5, false, $6, NULLIF($7, ''))
		RETURNING id
	`, userID, planID, npResp.SubscriptionID, tierName, normalizeSubStatus(npResp.Status), expiresAt, checkoutID).Scan(&subID)
	if err != nil {
		return nil, err
	}

	return &CreateSubResponse{
		SubscriptionID: subID,
		Status:         npResp.Status,
	}, nil
}

// createDepositForUser creates a NowPayments deposit and stores the local row
// (including the payment screen fields + expiry), optionally linked to a
// checkout workflow.
func createDepositForUser(ctx context.Context, userID, planID, crypto, checkoutID string) (*CreateDepositResponse, error) {
	subResp, err := ensureSubPartnerID(ctx, &myauth.EnsureSubPartnerIDParams{UserID: userID})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "ensure sub_partner_id failed: " + err.Error()}
	}
	subPartnerID := subResp.SubPartnerID

	plan, err := getPlan(ctx, planID)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found or no sub_partner_id"}
	}
	priceCents := plan.PriceUsdCents

	if subPartnerID == "" {
		return nil, &errs.Error{Code: errs.NotFound, Message: "plan not found or no sub_partner_id"}
	}

	var depositID string
	err = db.QueryRow(ctx, `
		INSERT INTO deposits (user_id, currency_crypto, amount_usd_cents, checkout_id)
		VALUES ($1, $2, $3, NULLIF($4, '')) RETURNING id
	`, userID, crypto, priceCents, checkoutID).Scan(&depositID)
	if err != nil {
		return nil, err
	}

	callbackURL := buildCallbackURL("/webhooks/nowpayments/deposit?deposit_id=" + depositID)

	npResp, err := provider.CreateDeposit(ctx, DepositRequest{
		Crypto:         crypto,
		AmountUSD:      float64(priceCents) / 100,
		SubPartnerID:   subPartnerID,
		IPNCallbackURL: callbackURL,
	})
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "nowpayments deposit creation failed: " + err.Error()}
	}

	expiresAt := parseTimeLimit(npResp.TimeLimit)
	qr, uri := buildDepositQR(npResp)

	db.Exec(ctx, `
		UPDATE deposits SET provider_deposit_id = $1, pay_address = $2, amount_crypto = $3,
			payin_extra_id = $4, network = $5, qr_code_url = $6, expires_at = $7
		WHERE id = $8
	`, npResp.PaymentID, npResp.PayAddress, fmtNum(npResp.PayAmount), npResp.PayinExtraID, npResp.Network, qr, expiresAt, depositID)

	return &CreateDepositResponse{
		DepositID:    depositID,
		PayAddress:   npResp.PayAddress,
		PayAmount:    npResp.PayAmount,
		PayCurrency:  npResp.PayCurrency,
		PlanID:       planID,
		PayinExtraID: npResp.PayinExtraID,
		Network:      npResp.Network,
		QrDataURL:    qr,
		PaymentURI:   uri,
	}, nil
}

// ----- Internal (worker-facing) private endpoints -----

type InternalCheckBalanceParams struct {
	UserID string `json:"user_id"`
	Crypto string `json:"crypto"`
}

type InternalCheckBalanceResponse struct {
	HasBalance bool `json:"has_balance"`
}

//encore:api private method=POST path=/billing/internal/check-balance
func InternalCheckBalance(ctx context.Context, p *InternalCheckBalanceParams) (*InternalCheckBalanceResponse, error) {
	balances, err := checkBalanceForUser(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	amount := balances.Balances[p.Crypto]
	if amount.Amount <= 0 {
		if upper, ok := balances.Balances[strings.ToUpper(p.Crypto)]; ok {
			amount = upper
		}
	}
	return &InternalCheckBalanceResponse{HasBalance: amount.Amount > 0}, nil
}

type InternalCreateDepositParams struct {
	UserID     string `json:"user_id"`
	PlanID     string `json:"plan_id"`
	Crypto     string `json:"crypto"`
	CheckoutID string `json:"checkout_id"`
}

type InternalCreateDepositResponse struct {
	DepositID      string `json:"deposit_id"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

//encore:api private method=POST path=/billing/internal/create-deposit
func InternalCreateDeposit(ctx context.Context, p *InternalCreateDepositParams) (*InternalCreateDepositResponse, error) {
	res, err := createDepositForUser(ctx, p.UserID, p.PlanID, p.Crypto, p.CheckoutID)
	if err != nil {
		return nil, err
	}
	return &InternalCreateDepositResponse{
		DepositID:      res.DepositID,
		TimeoutSeconds: depositTimeoutSeconds(ctx, res.DepositID),
	}, nil
}

type InternalCreateSubscriptionParams struct {
	UserID     string `json:"user_id"`
	PlanID     string `json:"plan_id"`
	CheckoutID string `json:"checkout_id"`
}

type InternalCreateSubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
}

//encore:api private method=POST path=/billing/internal/create-subscription
func InternalCreateSubscription(ctx context.Context, p *InternalCreateSubscriptionParams) (*InternalCreateSubscriptionResponse, error) {
	res, err := createSubscriptionForUser(ctx, p.UserID, p.PlanID, p.CheckoutID)
	if err != nil {
		return nil, err
	}
	return &InternalCreateSubscriptionResponse{SubscriptionID: res.SubscriptionID, Status: res.Status}, nil
}

// ExpireDeposit marks a deposit expired. Called by the worker; idempotent.
//
//encore:api private method=POST path=/billing/deposits/:id/expire
func ExpireDeposit(ctx context.Context, id string) error {
	db.Exec(ctx, `UPDATE deposits SET status = 'expired', updated_at = now() WHERE id = $1`, id)
	return nil
}

// CompleteDeposit marks a deposit completed and grants its quota boost (if any)
// exactly once. Called by the worker; idempotent.
//
//encore:api private method=POST path=/billing/deposits/:id/complete
func CompleteDeposit(ctx context.Context, id string) error {
	return completeDeposit(ctx, id)
}

func completeDeposit(ctx context.Context, depositID string) error {
	db.Exec(ctx, `UPDATE deposits SET status = 'completed', completed_at = now(), updated_at = now() WHERE id = $1`, depositID)

	var boostUserID string
	var boostDownloads int
	db.QueryRow(ctx, `SELECT user_id, boost_downloads FROM deposits WHERE id = $1 AND boost_downloads > 0 AND boost_granted = false`, depositID).Scan(&boostUserID, &boostDownloads)
	if boostDownloads > 0 && boostUserID != "" {
		if err := myreading.GrantBoost(ctx, &myreading.GrantBoostParams{UserID: boostUserID, Downloads: boostDownloads}); err == nil {
			db.Exec(ctx, `UPDATE deposits SET boost_granted = true WHERE id = $1`, depositID)
		}
	}
	return nil
}

// depositTimeoutSeconds returns the seconds remaining on a deposit's payment
// window (from its stored expires_at), floored at zero.
func depositTimeoutSeconds(ctx context.Context, depositID string) int {
	var expiresAt time.Time
	if err := db.QueryRow(ctx, `SELECT expires_at FROM deposits WHERE id = $1`, depositID).Scan(&expiresAt); err != nil {
		return 0
	}
	sec := int(time.Until(expiresAt).Seconds())
	if sec < 0 {
		return 0
	}
	return sec
}

// ----- Frontend-facing auth endpoints (combined checkout) -----

type StartSubscriptionParams struct {
	PlanID string `json:"plan_id"`
	Crypto string `json:"crypto"`
}

type StartSubscriptionResponse struct {
	CheckoutID string `json:"checkout_id"`
}

//encore:api auth method=POST path=/billing/start-subscription
func StartSubscription(ctx context.Context, p *StartSubscriptionParams) (*StartSubscriptionResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	checkoutID := uuid.NewString()

	if err := startSubscribeWorkflow(ctx, checkoutID, ad.UserID, p.PlanID, p.Crypto, waitingPayTimeout(ctx)); err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "failed to start checkout: " + err.Error()}
	}
	return &StartSubscriptionResponse{CheckoutID: checkoutID}, nil
}

type SubscriptionStateResponse struct {
	Step           string `json:"step"`
	PayAddress     string `json:"pay_address,omitempty"`
	PayAmount      string `json:"pay_amount,omitempty"`
	PayCurrency    string `json:"pay_currency,omitempty"`
	PayinExtraID   string `json:"payin_extra_id,omitempty"`
	Network        string `json:"network,omitempty"`
	QrDataURL      string `json:"qr_data_url,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
}

//encore:api auth method=GET path=/billing/checkout/:id
func GetSubscriptionState(ctx context.Context, id string) (*SubscriptionStateResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	// Latest subscription for this checkout.
	var sub struct {
		SubID  string
		Status string
		Active bool
	}
	subErr := db.QueryRow(ctx, `
		SELECT id, status, active FROM subscriptions WHERE checkout_id = $1 AND user_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, id, ad.UserID).Scan(&sub.SubID, &sub.Status, &sub.Active)

	if subErr == nil {
		switch {
		case sub.Active || sub.Status == "active":
			return &SubscriptionStateResponse{Step: "active", SubscriptionID: sub.SubID}, nil
		case sub.Status == "expired":
			return &SubscriptionStateResponse{Step: "expired", SubscriptionID: sub.SubID}, nil
		case sub.Status == "failed" || sub.Status == "cancelled":
			return &SubscriptionStateResponse{Step: "failed", SubscriptionID: sub.SubID}, nil
		default:
			return &SubscriptionStateResponse{Step: "awaiting_subscription", SubscriptionID: sub.SubID}, nil
		}
	}

	// Otherwise, look for a deposit awaiting payment.
	var dep struct {
		DepID        string
		Status       string
		PayAddress   string
		AmountCrypto string
		Currency     string
		PayinExtraID string
		Network      string
		QrCodeURL    string
		ExpiresAt    time.Time
	}
	depErr := db.QueryRow(ctx, `
		SELECT id, status, COALESCE(pay_address, ''), COALESCE(amount_crypto, ''),
			currency_crypto, COALESCE(payin_extra_id, ''), COALESCE(network, ''),
			COALESCE(qr_code_url, ''), expires_at
		FROM deposits WHERE checkout_id = $1 AND user_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, id, ad.UserID).Scan(&dep.DepID, &dep.Status, &dep.PayAddress, &dep.AmountCrypto, &dep.Currency, &dep.PayinExtraID, &dep.Network, &dep.QrCodeURL, &dep.ExpiresAt)

	if depErr == nil {
		switch dep.Status {
		case "expired", "failed":
			return &SubscriptionStateResponse{Step: "expired"}, nil
		case "completed":
			return &SubscriptionStateResponse{Step: "awaiting_subscription"}, nil
		default:
			return &SubscriptionStateResponse{
				Step:         "awaiting_deposit",
				PayAddress:   dep.PayAddress,
				PayAmount:    dep.AmountCrypto,
				PayCurrency:  dep.Currency,
				PayinExtraID: dep.PayinExtraID,
				Network:      dep.Network,
				QrDataURL:    dep.QrCodeURL,
				ExpiresAt:    dep.ExpiresAt.Format(time.RFC3339),
			}, nil
		}
	}

	return &SubscriptionStateResponse{Step: "checking"}, nil
}
