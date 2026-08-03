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
	"time"
)

const npBaseURL = "https://api.nowpayments.io/v1"

type NowPaymentsProvider struct {
	apiKey string
	ipnKey string
	http   *http.Client
}

func NewNowPaymentsProvider(apiKey, ipnKey string) *NowPaymentsProvider {
	return &NowPaymentsProvider{
		apiKey: apiKey,
		ipnKey: ipnKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

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

func (p *NowPaymentsProvider) CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error) {
	planID, _ := strconv.Atoi(req.SubscriptionPlanID)
	subPartnerID, _ := strconv.Atoi(req.SubPartnerID)
	body := map[string]interface{}{
		"subscription_plan_id": planID,
		"sub_partner_id":       subPartnerID,
	}

	resp, err := p.doRequest(ctx, "POST", npBaseURL+"/subscriptions", body)
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

func (p *NowPaymentsProvider) CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error) {
	subPartnerID, _ := strconv.Atoi(req.SubPartnerID)
	body := map[string]interface{}{
		"currency":           req.Crypto,
		"amount":             req.AmountUSD,
		"sub_partner_id":     subPartnerID,
		"ipn_callback_url":   req.IPNCallbackURL,
		"is_fixed_rate":      false,
		"is_fee_paid_by_user": false,
	}

	resp, err := p.doRequest(ctx, "POST", npBaseURL+"/sub-partner/payment", body)
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

func (p *NowPaymentsProvider) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
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
