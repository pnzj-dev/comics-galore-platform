package nowpayments

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// buildURL appends query params to a v1 API path.
func buildURL(path string, q url.Values) string {
	if len(q) == 0 {
		return npBaseURL + path
	}
	return npBaseURL + path + "?" + q.Encode()
}

// unwrapResult decodes a NowPayments response whose payload is wrapped in a
// top-level `result` field.
func unwrapResult[T any](resp []byte) (T, error) {
	var zero T
	var envelope struct {
		Result T `json:"result"`
	}
	if err := json.Unmarshal(resp, &envelope); err != nil {
		return zero, err
	}
	return envelope.Result, nil
}

// ----- Auth & status -----

// Status checks the API availability (GET /v1/status).
func (p *Provider) Status(ctx context.Context) (*StatusResponse, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/status", nil)
	if err != nil {
		return nil, err
	}
	var result StatusResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ----- Currencies -----

// ListAllCurrencies returns every cryptocurrency available for payments
// (GET /v1/currencies).
func (p *Provider) ListAllCurrencies(ctx context.Context) ([]string, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/currencies", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Currencies []string `json:"currencies"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Currencies, nil
}

// ListFullCurrencies returns detailed metadata for every available currency
// (GET /v1/full-currencies).
func (p *Provider) ListFullCurrencies(ctx context.Context) ([]FullCurrency, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/full-currencies", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Currencies []FullCurrency `json:"currencies"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Currencies, nil
}

// ----- Payments -----

// GetMinAmount returns the minimum payment amount for a currency pair
// (GET /v1/min-amount).
func (p *Provider) GetMinAmount(ctx context.Context, req MinAmountRequest) (*MinAmountResponse, error) {
	q := url.Values{}
	q.Set("currency_from", req.CurrencyFrom)
	q.Set("currency_to", req.CurrencyTo)
	if req.FiatEquivalent != "" {
		q.Set("fiat_equivalent", req.FiatEquivalent)
	}
	if req.IsFixedRate {
		q.Set("is_fixed_rate", "true")
	}
	if req.IsFeePaidByUser {
		q.Set("is_fee_paid_by_user", "true")
	}

	resp, err := p.doRequest(ctx, "GET", buildURL("/min-amount", q), nil)
	if err != nil {
		return nil, err
	}
	var result MinAmountResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateInvoice creates a payment link (POST /v1/invoice).
func (p *Provider) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	resp, err := p.doRequest(ctx, "POST", npBaseURL+"/invoice", req)
	if err != nil {
		return nil, err
	}
	var result Invoice
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreatePayment creates a payment and returns the deposit address
// (POST /v1/payment).
func (p *Provider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*Payment, error) {
	resp, err := p.doRequest(ctx, "POST", npBaseURL+"/payment", req)
	if err != nil {
		return nil, err
	}
	var result Payment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateInvoicePayment creates a payment for an existing invoice
// (POST /v1/invoice-payment).
func (p *Provider) CreateInvoicePayment(ctx context.Context, req CreateInvoicePaymentRequest) (*Payment, error) {
	resp, err := p.doRequest(ctx, "POST", npBaseURL+"/invoice-payment", req)
	if err != nil {
		return nil, err
	}
	var result Payment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMerchantEstimate refreshes the payment estimate
// (POST /v1/payment/{id}/update-merchant-estimate).
func (p *Provider) UpdateMerchantEstimate(ctx context.Context, paymentID string) (*UpdateMerchantEstimateResponse, error) {
	resp, err := p.doRequest(ctx, "POST", npBaseURL+"/payment/"+paymentID+"/update-merchant-estimate", nil)
	if err != nil {
		return nil, err
	}
	var result UpdateMerchantEstimateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPaymentStatus returns the current status of a payment
// (GET /v1/payment/{payment_id}).
func (p *Provider) GetPaymentStatus(ctx context.Context, paymentID string) (*Payment, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/payment/"+paymentID, nil)
	if err != nil {
		return nil, err
	}
	var result Payment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPayments returns the list of payments made to the account
// (GET /v1/payment/ — JWT + API key).
func (p *Provider) ListPayments(ctx context.Context, req ListPaymentsRequest) (*ListPaymentsResponse, error) {
	q := url.Values{}
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Page > 0 {
		q.Set("page", strconv.Itoa(req.Page))
	}
	if req.SortBy != "" {
		q.Set("sortBy", req.SortBy)
	}
	if req.OrderBy != "" {
		q.Set("orderBy", req.OrderBy)
	}
	if req.DateFrom != "" {
		q.Set("dateFrom", req.DateFrom)
	}
	if req.DateTo != "" {
		q.Set("dateTo", req.DateTo)
	}
	if req.InvoiceID != "" {
		q.Set("invoiceId", req.InvoiceID)
	}

	resp, err := p.doJWTRequest(ctx, "GET", buildURL("/payment/", q), nil)
	if err != nil {
		return nil, err
	}
	var result ListPaymentsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ----- Balance -----

// GetBalance returns the master account balances (GET /v1/balance).
func (p *Provider) GetBalance(ctx context.Context) (map[string]BalanceEntry, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/balance", nil)
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
	result := make(map[string]BalanceEntry, len(raw))
	for k, v := range raw {
		amt, _ := v.Amount.Float64()
		pend, _ := v.PendingAmount.Float64()
		result[k] = BalanceEntry{Amount: amt, PendingAmount: pend}
	}
	return result, nil
}

// ----- Customer management (sub-partner) -----

// GetCustomers returns the list of customer (sub-partner) accounts
// (GET /v1/sub-partner — JWT).
func (p *Provider) GetCustomers(ctx context.Context, req GetCustomersRequest) ([]Customer, error) {
	q := url.Values{}
	if req.ID != "" {
		q.Set("id", req.ID)
	}
	if req.Offset > 0 {
		q.Set("offset", strconv.Itoa(req.Offset))
	}
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Order != "" {
		q.Set("order", req.Order)
	}
	resp, err := p.doJWTRequest(ctx, "GET", buildURL("/sub-partner", q), nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[[]Customer](resp)
}

// GetTransfers returns the list of customer transfers
// (GET /v1/sub-partner/transfers — JWT).
func (p *Provider) GetTransfers(ctx context.Context, req GetTransfersRequest) ([]Transfer, error) {
	q := url.Values{}
	if req.ID != "" {
		q.Set("id", req.ID)
	}
	if req.Status != "" {
		q.Set("status", req.Status)
	}
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Offset > 0 {
		q.Set("offset", strconv.Itoa(req.Offset))
	}
	if req.Order != "" {
		q.Set("order", req.Order)
	}
	resp, err := p.doJWTRequest(ctx, "GET", buildURL("/sub-partner/transfers", q), nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[[]Transfer](resp)
}

// GetTransfer returns a single transfer (GET /v1/sub-partner/transfer/{id} — JWT).
func (p *Provider) GetTransfer(ctx context.Context, id string) (*Transfer, error) {
	resp, err := p.doJWTRequest(ctx, "GET", npBaseURL+"/sub-partner/transfer/"+id, nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Transfer](resp)
}

// CreateTransfer creates a transfer between customer accounts
// (POST /v1/sub-partner/transfer — JWT).
func (p *Provider) CreateTransfer(ctx context.Context, req CreateTransferRequest) (*Transfer, error) {
	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/transfer", req)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Transfer](resp)
}

// GetCustomerPayments returns the payments generated for a customer
// (GET /v1/sub-partner/payments — JWT). The response shape is not fully
// documented upstream; both `result` and `data` are accepted.
func (p *Provider) GetCustomerPayments(ctx context.Context, req GetCustomerPaymentsRequest) ([]Payment, error) {
	q := url.Values{}
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Page > 0 {
		q.Set("page", strconv.Itoa(req.Page))
	}
	if req.ID != "" {
		q.Set("id", req.ID)
	}
	if req.PayCurrency != "" {
		q.Set("pay_currency", req.PayCurrency)
	}
	if req.Status != "" {
		q.Set("status", req.Status)
	}
	if req.SubPartnerID != "" {
		q.Set("sub_partner_id", req.SubPartnerID)
	}
	if req.DateFrom != "" {
		q.Set("date_from", req.DateFrom)
	}
	if req.DateTo != "" {
		q.Set("date_to", req.DateTo)
	}
	if req.OrderBy != "" {
		q.Set("orderBy", req.OrderBy)
	}
	if req.SortBy != "" {
		q.Set("sortBy", req.SortBy)
	}

	resp, err := p.doJWTRequest(ctx, "GET", buildURL("/sub-partner/payments", q), nil)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Result []Payment `json:"result"`
		Data   []Payment `json:"data"`
	}
	if err := json.Unmarshal(resp, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Result) > 0 {
		return envelope.Result, nil
	}
	return envelope.Data, nil
}

// DepositToSubPartner transfers funds from the master account to a customer
// (POST /v1/sub-partner/deposit — JWT).
func (p *Provider) DepositToSubPartner(ctx context.Context, req SubPartnerTransferRequest) (*Transfer, error) {
	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/deposit", req)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Transfer](resp)
}

// WriteOffFromSubPartner withdraws funds from a customer to the master account
// (POST /v1/sub-partner/write-off — JWT).
func (p *Provider) WriteOffFromSubPartner(ctx context.Context, req SubPartnerTransferRequest) (*Transfer, error) {
	resp, err := p.doJWTRequest(ctx, "POST", npBaseURL+"/sub-partner/write-off", req)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Transfer](resp)
}

// ----- Recurring payments (subscriptions) -----

// ListSubscriptions returns recurring payments filtered by status/plan
// (GET /v1/subscriptions).
func (p *Provider) ListSubscriptions(ctx context.Context, req ListSubscriptionsRequest) ([]Subscription, error) {
	q := url.Values{}
	if req.Status != "" {
		q.Set("status", req.Status)
	}
	if req.SubscriptionPlanID != "" {
		q.Set("subscription_plan_id", req.SubscriptionPlanID)
	}
	if req.IsActive != nil {
		q.Set("is_active", strconv.FormatBool(*req.IsActive))
	}
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Offset > 0 {
		q.Set("offset", strconv.Itoa(req.Offset))
	}
	resp, err := p.doRequest(ctx, "GET", buildURL("/subscriptions", q), nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[[]Subscription](resp)
}

// GetSubscription returns a single recurring payment
// (GET /v1/subscriptions/{sub_id}).
func (p *Provider) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/subscriptions/"+id, nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Subscription](resp)
}

// DeleteSubscription removes a recurring payment
// (DELETE /v1/subscriptions/{sub_id} — JWT).
func (p *Provider) DeleteSubscription(ctx context.Context, id string) error {
	_, err := p.doJWTRequest(ctx, "DELETE", npBaseURL+"/subscriptions/"+id, nil)
	return err
}

// ListPlans returns all subscription plans (GET /v1/subscriptions/plans).
func (p *Provider) ListPlans(ctx context.Context) ([]Plan, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/subscriptions/plans", nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[[]Plan](resp)
}

// GetPlan returns a single subscription plan
// (GET /v1/subscriptions/plans/{plan-id}).
func (p *Provider) GetPlan(ctx context.Context, id string) (*Plan, error) {
	resp, err := p.doRequest(ctx, "GET", npBaseURL+"/subscriptions/plans/"+id, nil)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Plan](resp)
}

// UpdatePlan updates a subscription plan
// (PATCH /v1/subscriptions/plans/{plan-id} — JWT).
func (p *Provider) UpdatePlan(ctx context.Context, id string, req UpdatePlanRequest) (*Plan, error) {
	resp, err := p.doJWTRequest(ctx, "PATCH", npBaseURL+"/subscriptions/plans/"+id, req)
	if err != nil {
		return nil, err
	}
	return unwrapResult[*Plan](resp)
}
