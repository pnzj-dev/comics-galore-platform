package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	myauth "comics-galore/backend/auth"
	"comics-galore/backend/tiers"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

// signPayload computes the NowPayments IPN signature (HMAC-SHA512 over the
// sorted-key JSON payload), mirroring verifySignature.
func signPayload(secret string, obj map[string]interface{}) string {
	sorted := sortObject(obj)
	sortedJSON, _ := json.Marshal(sorted)
	h := hmac.New(sha512.New, []byte(secret))
	h.Write(sortedJSON)
	return hex.EncodeToString(h.Sum(nil))
}

func buildSubscriptionWebhookPayload(providerSubID, status string, amountUSD float64) map[string]interface{} {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]interface{}{
		"id":               providerSubID,
		"status":           status,
		"currency":         "usd",
		"amount":           fmtNum(amountUSD),
		"ipn_callback_url": buildCallbackURL("/webhooks/nowpayments/subscription"),
		"created_at":       now,
		"updated_at":       now,
	}
}

func buildDepositWebhookPayload(providerDepositID, status, amountCrypto, currencyCrypto string, amountUsdCents int) map[string]interface{} {
	now := time.Now().UTC().Format(time.RFC3339)
	price := float64(amountUsdCents) / 100
	return map[string]interface{}{
		"payment_id":     providerDepositID,
		"payment_status": status,
		"price_amount":   price,
		"price_currency": "usd",
		"pay_amount":     amountCrypto,
		"pay_currency":   currencyCrypto,
		"actually_paid":  amountCrypto,
		"purchase_id":    providerDepositID,
		"created_at":     now,
		"updated_at":     now,
	}
}

// ----- Simulate Webhook -----

type SimulateWebhookParams struct {
	Type   string `json:"type"`   // "deposit" | "subscription"
	ID     string `json:"id"`     // local deposit/subscription id
	Status string `json:"status"` // "finished" | "waiting" | "expired" | "failed" | "partially_paid" | ...
	DryRun bool   `json:"dry_run"`
}

type SimulateWebhookRequestDetail struct {
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Signature string              `json:"signature"`
	Payload   json.RawMessage     `json:"payload"`
}

type SimulateWebhookResponseDetail struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

type SimulateWebhookOutcome struct {
	SubscriptionActive bool   `json:"subscription_active"`
	SubscriptionStatus string `json:"subscription_status"`
	DepositCompleted   bool   `json:"deposit_completed"`
	BoostGranted       bool   `json:"boost_granted"`
}

type SimulateWebhookResponse struct {
	Request  SimulateWebhookRequestDetail  `json:"request"`
	Response SimulateWebhookResponseDetail `json:"response"`
	Outcome  SimulateWebhookOutcome        `json:"outcome"`
}

//encore:api auth method=POST path=/admin/simulate-webhook
func SimulateWebhook(ctx context.Context, p *SimulateWebhookParams) (*SimulateWebhookResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	status := strings.ToLower(strings.TrimSpace(p.Status))
	if status == "" {
		status = "finished"
	}

	var (
		path      string
		query     map[string][]string
		payload   map[string]interface{}
		invokeSub bool
		depositID string
	)

	switch strings.ToLower(strings.TrimSpace(p.Type)) {
	case "subscription":
		var sub struct {
			ID                     string
			PlanID                 string
			ProviderSubscriptionID string
		}
		err := db.QueryRow(ctx, `
			SELECT id, plan_id, COALESCE(provider_subscription_id, '')
			FROM subscriptions WHERE id = $1
		`, p.ID).Scan(&sub.ID, &sub.PlanID, &sub.ProviderSubscriptionID)
		if err != nil {
			if isNoRows(err) {
				return nil, &errs.Error{Code: errs.NotFound, Message: "subscription not found"}
			}
			return nil, err
		}
		if sub.ProviderSubscriptionID == "" {
			return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "subscription has no provider_subscription_id"}
		}

		price := 0.0
		if plan, perr := tiers.GetPlan(ctx, sub.PlanID); perr == nil {
			price = float64(plan.PriceUsdCents) / 100
		}
		payload = buildSubscriptionWebhookPayload(sub.ProviderSubscriptionID, strings.ToUpper(status), price)
		path = "/webhooks/nowpayments/subscription"
		query = map[string][]string{}
		invokeSub = true
	case "deposit":
		var dep struct {
			ID                string
			ProviderDepositID string
			AmountCrypto      string
			CurrencyCrypto    string
			AmountUsdCents    int
		}
		err := db.QueryRow(ctx, `
			SELECT id, COALESCE(provider_deposit_id, ''), COALESCE(amount_crypto, ''), currency_crypto, COALESCE(amount_usd_cents, 0)
			FROM deposits WHERE id = $1
		`, p.ID).Scan(&dep.ID, &dep.ProviderDepositID, &dep.AmountCrypto, &dep.CurrencyCrypto, &dep.AmountUsdCents)
		if err != nil {
			if isNoRows(err) {
				return nil, &errs.Error{Code: errs.NotFound, Message: "deposit not found"}
			}
			return nil, err
		}
		if dep.ProviderDepositID == "" {
			return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "deposit has no provider_deposit_id"}
		}

		depositID = dep.ID
		payload = buildDepositWebhookPayload(dep.ProviderDepositID, status, dep.AmountCrypto, dep.CurrencyCrypto, dep.AmountUsdCents)
		path = "/webhooks/nowpayments/deposit"
		query = map[string][]string{"deposit_id": {dep.ID}}
	default:
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "type must be 'deposit' or 'subscription'"}
	}

	body, _ := json.Marshal(payload)
	sig := signPayload(secrets.NowPaymentsIPNKey, payload)

	reqDetail := SimulateWebhookRequestDetail{
		Method:    "POST",
		Path:      path,
		Query:     query,
		Signature: sig,
		Payload:   json.RawMessage(body),
	}

	if p.DryRun {
		return &SimulateWebhookResponse{
			Request:  reqDetail,
			Response: SimulateWebhookResponseDetail{StatusCode: 0, Body: "(dry run — webhook not invoked)"},
			Outcome:  SimulateWebhookOutcome{},
		}, nil
	}

	// Invoke the core processing path (signature verification included).
	var statusCode int
	var respBody string
	if invokeSub {
		statusCode, respBody = processSubscriptionWebhook(ctx, body, sig)
	} else {
		statusCode, respBody = processDepositWebhook(ctx, body, sig, depositID)
	}

	outcome := SimulateWebhookOutcome{}
	if invokeSub {
		var active bool
		var subStatus string
		if err := db.QueryRow(ctx, `SELECT active, status FROM subscriptions WHERE id = $1`, p.ID).Scan(&active, &subStatus); err == nil {
			outcome.SubscriptionActive = active
			outcome.SubscriptionStatus = subStatus
		}
	} else {
		var depStatus string
		var boostGranted bool
		if err := db.QueryRow(ctx, `SELECT status, boost_granted FROM deposits WHERE id = $1`, p.ID).Scan(&depStatus, &boostGranted); err == nil {
			outcome.DepositCompleted = depStatus == "completed"
			outcome.BoostGranted = boostGranted
		}
	}

	return &SimulateWebhookResponse{
		Request:  reqDetail,
		Response: SimulateWebhookResponseDetail{StatusCode: statusCode, Body: respBody},
		Outcome:  outcome,
	}, nil
}
