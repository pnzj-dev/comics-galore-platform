// Package nowpayments implements the NowPayments REST API client shared by
// the auth, billing and tiers services.
//
// Full API reference: backend/billing/nowpayments-openapi.yaml
// Base URL: https://api.nowpayments.io
package nowpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"encore.dev"
)

const npBaseURL = "https://api.nowpayments.io/v1"

type Provider struct {
	apiKey   string
	ipnKey   string
	email    string
	password string

	jwtToken     string
	jwtExpiresAt time.Time
	jwtMutex     sync.Mutex

	http *http.Client
}

func NewProvider(apiKey, ipnKey, email, password string) *Provider {
	return &Provider{
		apiKey:   apiKey,
		ipnKey:   ipnKey,
		email:    email,
		password: password,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// BuildCallbackURL builds a webhook callback URL from the backend's own API
// base URL (encore.Meta().APIBaseURL). When running locally and a valid ngrok
// tunnel URL is configured, the tunnel is used instead so NowPayments can reach
// the local machine.
func BuildCallbackURL(ngrokURL, path string) string {
	base := encore.Meta().APIBaseURL
	host := base.Hostname()

	if (host == "localhost" || host == "127.0.0.1") && ngrokURL != "" {
		if u, err := url.Parse(ngrokURL); err == nil && u.Scheme != "" && u.Host != "" {
			return strings.TrimRight(ngrokURL, "/") + path
		}
	}

	return strings.TrimRight(base.String(), "/") + path
}

// getAuthToken returns a cached JWT token, refreshing if expired.
// Tokens from NowPayments expire in 5 minutes; we cache for 4 minutes.
func (p *Provider) getAuthToken(ctx context.Context) (string, error) {
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
func (p *Provider) doJWTRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	token, err := p.getAuthToken(ctx)
	if err != nil {
		return nil, err
	}
	return p.doRequestWithAuth(ctx, method, url, body, token)
}

func (p *Provider) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	return p.doRequestWithAuth(ctx, method, url, body, "")
}

func (p *Provider) doRequestWithAuth(ctx context.Context, method, url string, body interface{}, jwtToken string) ([]byte, error) {
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

func (p *Provider) EstimatePrice(ctx context.Context, req EstimateRequest) (*EstimateResponse, error) {
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

// ----- ListCurrencies (API key only) -----

// ListCurrencies returns the cryptocurrencies the merchant accepts, preferring
// the "checked" (enabled) coins and falling back to the full platform list.
func (p *Provider) ListCurrencies(ctx context.Context) ([]string, error) {
	for _, path := range []string{"/merchant/coins", "/currencies"} {
		resp, err := p.doRequest(ctx, "GET", npBaseURL+path, nil)
		if err != nil {
			return nil, err
		}
		var result struct {
			Currencies []string `json:"currencies"`
		}
		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, err
		}
		if len(result.Currencies) > 0 {
			return result.Currencies, nil
		}
	}
	return []string{}, nil
}

// ----- CheckBalance (API key only) -----

func (p *Provider) CheckBalance(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error) {
	url := fmt.Sprintf("%s/sub-partner/balance/%s", npBaseURL, subPartnerID)

	resp, err := p.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	result := make(map[string]BalanceEntry)

	// Docs show `balances` as a map keyed by currency; the live API returns an
	// array of {currency, amount, pendingAmount} objects. Handle both.
	var obj struct {
		Result struct {
			Balances map[string]struct {
				Amount        json.Number `json:"amount"`
				PendingAmount json.Number `json:"pendingAmount"`
			} `json:"balances"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &obj); err == nil && len(obj.Result.Balances) > 0 {
		for k, v := range obj.Result.Balances {
			amt, _ := v.Amount.Float64()
			pend, _ := v.PendingAmount.Float64()
			result[strings.ToLower(k)] = BalanceEntry{Amount: amt, PendingAmount: pend}
		}
		return result, nil
	}

	var arr struct {
		Result struct {
			Balances []struct {
				Currency      string      `json:"currency"`
				CurrencyCode  string      `json:"currency_code"`
				Amount        json.Number `json:"amount"`
				PendingAmount json.Number `json:"pendingAmount"`
			} `json:"balances"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &arr); err != nil {
		return nil, err
	}
	for _, v := range arr.Result.Balances {
		code := v.Currency
		if code == "" {
			code = v.CurrencyCode
		}
		code = strings.ToLower(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		amt, _ := v.Amount.Float64()
		pend, _ := v.PendingAmount.Float64()
		result[code] = BalanceEntry{Amount: amt, PendingAmount: pend}
	}
	return result, nil
}

// ----- CreateCustomer (JWT + API key) -----

func (p *Provider) CreateCustomer(ctx context.Context, name string) (string, error) {
	body := map[string]interface{}{
		"name": name,
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/balance", body)
	if err != nil {
		return "", err
	}

	var result struct {
		Result struct {
			ID json.Number `json:"id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("create customer parse: %w", err)
	}

	id := result.Result.ID.String()
	if id == "" {
		return "", fmt.Errorf("create customer: nowpayments returned no id (response: %s)", string(resp))
	}
	return id, nil
}

// ----- CreateSubscription (JWT + API key) -----

func (p *Provider) CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
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

	// The response wraps the created subscription in `result`, which is either
	// a single object or an array of objects depending on the account type.
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(resp, &envelope); err != nil {
		return nil, fmt.Errorf("create subscription parse: %w", err)
	}

	trimmed := bytes.TrimSpace(envelope.Result)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, fmt.Errorf("create subscription: nowpayments returned empty result (response: %s)", string(resp))
	}

	sub := struct {
		ID     json.Number `json:"id"`
		Status string      `json:"status"`
	}{}
	if trimmed[0] == '[' {
		var subs []struct {
			ID     json.Number `json:"id"`
			Status string      `json:"status"`
		}
		if err := json.Unmarshal(envelope.Result, &subs); err != nil {
			return nil, fmt.Errorf("create subscription parse: %w", err)
		}
		if len(subs) == 0 {
			return nil, fmt.Errorf("create subscription: nowpayments returned empty result (response: %s)", string(resp))
		}
		sub = subs[0]
	} else {
		if err := json.Unmarshal(envelope.Result, &sub); err != nil {
			return nil, fmt.Errorf("create subscription parse: %w", err)
		}
	}

	if sub.ID.String() == "" {
		return nil, fmt.Errorf("create subscription: nowpayments returned no id (response: %s)", string(resp))
	}

	return &SubscriptionResponse{
		SubscriptionID: sub.ID.String(),
		Status:         sub.Status,
	}, nil
}

// ----- CreateDeposit (JWT + API key) -----

func (p *Provider) CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error) {
	body := map[string]interface{}{
		"currency":            req.Crypto,
		"amount":              req.AmountUSD,
		"sub_partner_id":      req.SubPartnerID,
		"ipn_callback_url":    req.IPNCallbackURL,
		"is_fixed_rate":       false,
		"is_fee_paid_by_user": false,
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/payment", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result struct {
			PaymentID        json.Number `json:"payment_id"`
			PayAddress       string      `json:"pay_address"`
			PayAmount        json.Number `json:"pay_amount"`
			PayCurrency      string      `json:"pay_currency"`
			PayinExtraID     string      `json:"payin_extra_id"`
			Network          string      `json:"network"`
			NetworkPrecision int         `json:"network_precision"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	payAmt, _ := result.Result.PayAmount.Float64()
	return &DepositResponse{
		PaymentID:       result.Result.PaymentID.String(),
		PayAddress:      result.Result.PayAddress,
		PayAmount:       payAmt,
		PayCurrency:     result.Result.PayCurrency,
		PayinExtraID:    result.Result.PayinExtraID,
		Network:         result.Result.Network,
		NetworkPrecision: result.Result.NetworkPrecision,
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

func (p *Provider) CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error) {
	body := map[string]interface{}{
		"title":            req.Name,
		"amount":           req.PriceAmount,
		"currency":         "usd",
		"interval_day":     periodToDays(req.Period),
		"ipn_callback_url": req.IPNCallbackURL,
	}

	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/subscriptions/plans", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result struct {
			ID json.Number `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("create plan parse: %w", err)
	}

	providerPlanID := result.Result.ID.String()
	if providerPlanID == "" {
		return nil, fmt.Errorf("create plan: nowpayments returned no plan id (response: %s)", string(resp))
	}

	return &CreatePlanResponse{
		ProviderPlanID: providerPlanID,
	}, nil
}
