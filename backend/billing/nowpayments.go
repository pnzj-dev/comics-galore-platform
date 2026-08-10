// Package billing implements the PaymentsProvider interface using the
// NowPayments REST API.
//
// Full API reference: backend/billing/nowpayments-openapi.yaml
// Base URL: https://api.nowpayments.io
package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const npBaseURL = "https://api.nowpayments.io/v1"

type NowPaymentsProvider struct {
	apiKey   string
	ipnKey   string
	email    string
	password string

	jwtToken     string
	jwtExpiresAt time.Time
	jwtMutex     sync.Mutex

	http *http.Client
}

func NewNowPaymentsProvider(apiKey, ipnKey, email, password string) *NowPaymentsProvider {
	return &NowPaymentsProvider{
		apiKey:   apiKey,
		ipnKey:   ipnKey,
		email:    email,
		password: password,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// getAuthToken returns a cached JWT token, refreshing if expired.
// Tokens from NowPayments expire in 5 minutes; we cache for 4 minutes.
func (p *NowPaymentsProvider) getAuthToken(ctx context.Context) (string, error) {
	p.jwtMutex.Lock()
	defer p.jwtMutex.Unlock()

	if p.jwtToken != "" && time.Now().Add(time.Minute).Before(p.jwtExpiresAt) {
		return p.jwtToken, nil
	}

	body := map[string]string{"email": p.email, "password": p.password}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", npBaseURL+"/auth", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("nowpayments auth failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("nowpayments auth error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("nowpayments auth parse: %w", err)
	}

	p.jwtToken = result.Token
	p.jwtExpiresAt = time.Now().Add(4 * time.Minute)
	return p.jwtToken, nil
}

// doJWTRequest sends a request authenticated with both JWT (Bearer) and API key.
func (p *NowPaymentsProvider) doJWTRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	token, err := p.getAuthToken(ctx)
	if err != nil {
		return nil, err
	}
	return p.doRequestWithAuth(ctx, method, url, body, token)
}

func (p *NowPaymentsProvider) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	return p.doRequestWithAuth(ctx, method, url, body, "")
}

func (p *NowPaymentsProvider) doRequestWithAuth(ctx context.Context, method, url string, body interface{}, jwtToken string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("x-api-key", p.apiKey)
	if jwtToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("nowpayments error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ----- EstimatePrice (API key only) -----

func (p *NowPaymentsProvider) EstimatePrice(ctx context.Context, req EstimateRequest) (*EstimateResponse, error) {
	url := fmt.Sprintf("%s/estimate?amount=%.2f&currency_from=%s&currency_to=%s",
		npBaseURL, req.Amount, req.Currency, req.Crypto)

	resp, err := p.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		EstimatedAmount json.Number `json:"estimated_amount"`
		CurrencyFrom    string      `json:"currency_from"`
		CurrencyTo      string      `json:"currency_to"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	amt, _ := result.EstimatedAmount.Float64()
	return &EstimateResponse{
		EstimatedAmount: amt,
		FromCurrency:    result.CurrencyFrom,
		ToCurrency:      result.CurrencyTo,
	}, nil
}

// ----- CheckBalance (API key only) -----

func (p *NowPaymentsProvider) CheckBalance(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error) {
	url := fmt.Sprintf("%s/sub-partner/balance/%s", npBaseURL, subPartnerID)

	resp, err := p.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var raw map[string]struct {
		Amount        json.Number `json:"amount"`
		PendingAmount json.Number `json:"pendingAmount"`
	}
	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]BalanceEntry)
	for k, v := range raw {
		amt, _ := v.Amount.Float64()
		pend, _ := v.PendingAmount.Float64()
		result[k] = BalanceEntry{Amount: amt, PendingAmount: pend}
	}
	return result, nil
}

// ----- CreateCustomer (JWT + API key) -----

func (p *NowPaymentsProvider) CreateCustomer(ctx context.Context, name string) (string, error) {
	body := map[string]interface{}{
		"name": name,
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/balance", body)
	if err != nil {
		return "", err
	}

	var result struct {
		ID json.Number `json:"id"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("create customer parse: %w", err)
	}

	return result.ID.String(), nil
}

// ----- CreateSubscription (JWT + API key) -----

func (p *NowPaymentsProvider) CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
	planID, _ := strconv.Atoi(req.SubscriptionPlanID)
	subPartnerID, _ := strconv.Atoi(req.SubPartnerID)
	body := map[string]interface{}{
		"subscription_plan_id": planID,
		"sub_partner_id":       subPartnerID,
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/subscriptions", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ID     json.Number `json:"id"`
		Status string      `json:"status"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &SubscriptionResponse{
		SubscriptionID: result.ID.String(),
		Status:         result.Status,
	}, nil
}

// ----- CreateDeposit (JWT + API key) -----

func (p *NowPaymentsProvider) CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error) {
	subPartnerID, _ := strconv.Atoi(req.SubPartnerID)
	body := map[string]interface{}{
		"currency":            req.Crypto,
		"amount":              req.AmountUSD,
		"sub_partner_id":      subPartnerID,
		"ipn_callback_url":    req.IPNCallbackURL,
		"is_fixed_rate":       false,
		"is_fee_paid_by_user": false,
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/payment", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		PaymentID   json.Number `json:"payment_id"`
		PayAddress  string      `json:"pay_address"`
		PayAmount   json.Number `json:"pay_amount"`
		PayCurrency string      `json:"pay_currency"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	payAmt, _ := result.PayAmount.Float64()
	return &DepositResponse{
		PaymentID:  result.PaymentID.String(),
		PayAddress: result.PayAddress,
		PayAmount:  payAmt,
		PayCurrency: result.PayCurrency,
	}, nil
}

// ----- CreatePlan (JWT + API key) -----

func periodToDays(period string) int {
	switch period {
	case "day":
		return 1
	case "week":
		return 7
	case "month":
		return 30
	case "quarter":
		return 90
	case "semester":
		return 180
	case "year":
		return 365
	default:
		return 30
	}
}

func (p *NowPaymentsProvider) CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error) {
	body := map[string]interface{}{
		"title":          	req.Name,
		"amount":   		req.PriceAmount,
		"currency": 		"usd",
		"interval_day":     periodToDays(req.Period),
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/subscriptions/plans", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ID json.Number `json:"id"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("create plan parse: %w", err)
	}

	return &CreatePlanResponse{
		ProviderPlanID: result.ID.String(),
	}, nil
}
