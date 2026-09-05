package nowpayments

import (
	"context"
	"encoding/json"
)

// EstimateRequest is the input for EstimatePrice.
type EstimateRequest struct {
	Amount   float64
	Currency string
	Crypto   string
}

type EstimateResponse struct {
	EstimatedAmount float64 `json:"estimated_amount"`
	FromCurrency    string  `json:"from_currency"`
	ToCurrency      string  `json:"to_currency"`
}

type BalanceEntry struct {
	Amount        float64 `json:"amount"`
	PendingAmount float64 `json:"pending_amount"`
}

type SubscriptionRequest struct {
	PlanID             string
	SubPartnerID       string
	SubscriptionPlanID string
}

type SubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
	Status         string `json:"status"`
}

type DepositRequest struct {
	Crypto         string
	AmountUSD      float64
	SubPartnerID   string
	IPNCallbackURL string
}

type DepositResponse struct {
	PaymentID       string  `json:"payment_id"`
	PayAddress      string  `json:"pay_address"`
	PayAmount       float64 `json:"pay_amount"`
	PayCurrency     string  `json:"pay_currency"`
	PayinExtraID    string  `json:"payin_extra_id,omitempty"`
	Network         string  `json:"network,omitempty"`
	NetworkPrecision int    `json:"network_precision,omitempty"`
}

type CreatePlanRequest struct {
	Name           string
	PriceAmount    float64
	Period         string
	IPNCallbackURL string
}

type CreatePlanResponse struct {
	ProviderPlanID string
}

// PaymentsProvider is the minimal set of NowPayments operations the app's
// services depend on. The full client (Provider) exposes additional methods
// that are not part of this interface.
type PaymentsProvider interface {
	EstimatePrice(ctx context.Context, req EstimateRequest) (*EstimateResponse, error)
	ListCurrencies(ctx context.Context) ([]string, error)
	CheckBalance(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error)
	CreateCustomer(ctx context.Context, name string) (string, error)
	CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error)
	CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error)
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)
}

// ----- Auth & status -----

type StatusResponse struct {
	Message string `json:"message"`
}

// ----- Currencies -----

type FullCurrency struct {
	ID               int     `json:"id"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	Enable           bool    `json:"enable"`
	WalletRegex      string  `json:"wallet_regex"`
	Priority         int     `json:"priority"`
	ExtraIDExists    bool    `json:"extra_id_exists"`
	ExtraIDRegex     *string `json:"extra_id_regex"`
	LogoURL          string  `json:"logo_url"`
	Track            bool    `json:"track"`
	CGID             string  `json:"cg_id"`
	IsMaxLimit       bool    `json:"is_maxlimit"`
	Network          string  `json:"network"`
	SmartContract    *string `json:"smart_contract"`
	NetworkPrecision *int    `json:"network_precision"`
}

// ----- Payments -----

type MinAmountRequest struct {
	CurrencyFrom    string
	CurrencyTo      string
	FiatEquivalent  string
	IsFixedRate     bool
	IsFeePaidByUser bool
}

type MinAmountResponse struct {
	CurrencyFrom   string      `json:"currency_from"`
	CurrencyTo     string      `json:"currency_to"`
	MinAmount      json.Number `json:"min_amount"`
	FiatEquivalent json.Number `json:"fiat_equivalent"`
}

type CreateInvoiceRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency,omitempty"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	SuccessURL       string  `json:"success_url,omitempty"`
	CancelURL        string  `json:"cancel_url,omitempty"`
	PartiallyPaidURL string  `json:"partially_paid_url,omitempty"`
	IsFixedRate      bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"`
}

type Invoice struct {
	ID               json.Number `json:"id"`
	OrderID          string      `json:"order_id"`
	OrderDescription string `json:"order_description"`
	PriceAmount      string `json:"price_amount"`
	PriceCurrency    string `json:"price_currency"`
	PayCurrency      string `json:"pay_currency"`
	IPNCallbackURL   string `json:"ipn_callback_url"`
	InvoiceURL       string `json:"invoice_url"`
	SuccessURL       string `json:"success_url"`
	CancelURL        string `json:"cancel_url"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type CreatePaymentRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency"`
	PayAmount        float64 `json:"pay_amount,omitempty"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	IsFixedRate      bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"`
}

type CreateInvoicePaymentRequest struct {
	IID              string `json:"iid"`
	PayCurrency      string `json:"pay_currency"`
	PurchaseID       string `json:"purchase_id,omitempty"`
	OrderDescription string `json:"order_description,omitempty"`
	CustomerEmail    string `json:"customer_email,omitempty"`
	PayoutAddress    string `json:"payout_address,omitempty"`
	PayoutExtraID    string `json:"payout_extra_id,omitempty"`
	PayoutCurrency   string `json:"payout_currency,omitempty"`
}

// Payment is the shared shape returned by create payment, invoice-payment and
// payment status endpoints. ID/amount fields use json.Number because
// NowPayments returns them inconsistently as numbers or strings.
type Payment struct {
	PaymentID              json.Number  `json:"payment_id"`
	InvoiceID              json.Number  `json:"invoice_id"`
	PaymentStatus          string       `json:"payment_status"`
	PayAddress             string       `json:"pay_address"`
	PayinExtraID           string       `json:"payin_extra_id"`
	PriceAmount            json.Number  `json:"price_amount"`
	PriceCurrency          string       `json:"price_currency"`
	PayAmount              json.Number  `json:"pay_amount"`
	ActuallyPaid           json.Number  `json:"actually_paid"`
	PayCurrency            string       `json:"pay_currency"`
	OrderID                string       `json:"order_id"`
	OrderDescription       string       `json:"order_description"`
	PurchaseID             json.Number  `json:"purchase_id"`
	OutcomeAmount          json.Number  `json:"outcome_amount"`
	OutcomeCurrency        string       `json:"outcome_currency"`
	PayoutHash             string       `json:"payout_hash"`
	PayinHash              string       `json:"payin_hash"`
	AmountReceived         json.Number  `json:"amount_received"`
	SmartContract          string       `json:"smart_contract"`
	Network                string       `json:"network"`
	NetworkPrecision       int          `json:"network_precision"`
	TimeLimit              string       `json:"time_limit"`
	BurningPercent         string       `json:"burning_percent"`
	ExpirationEstimateDate string       `json:"expiration_estimate_date"`
	Type                   string       `json:"type"`
	PaymentExtraIDs        []json.Number `json:"payment_extra_ids"`
}

type UpdateMerchantEstimateResponse struct {
	ID                     json.Number `json:"id"`
	TokenID                json.Number `json:"token_id"`
	PayAmount              json.Number `json:"pay_amount"`
	ExpirationEstimateDate string      `json:"expiration_estimate_date"`
}

type ListPaymentsRequest struct {
	Limit     int
	Page      int
	SortBy    string
	OrderBy   string
	DateFrom  string
	DateTo    string
	InvoiceID string
}

type ListPaymentsResponse struct {
	Data       []Payment `json:"data"`
	Limit      int       `json:"limit"`
	Page       int       `json:"page"`
	PagesCount int       `json:"pagesCount"`
	Total      int       `json:"total"`
}

// ----- Customer management (sub-partner) -----

type Customer struct {
	ID        json.Number `json:"id"`
	Name      string      `json:"name"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

type GetCustomersRequest struct {
	ID     string
	Offset int
	Limit  int
	Order  string
}

type GetTransfersRequest struct {
	ID     string
	Status string
	Limit  int
	Offset int
	Order  string
}

type Transfer struct {
	ID        json.Number `json:"id"`
	FromSubID json.Number `json:"from_sub_id"`
	ToSubID   json.Number `json:"to_sub_id"`
	Status    string      `json:"status"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	Amount    json.Number `json:"amount"`
	Currency  string      `json:"currency"`
}

type CreateTransferRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	FromID   string  `json:"from_id"`
	ToID     string  `json:"to_id"`
}

type SubPartnerTransferRequest struct {
	Currency     string  `json:"currency"`
	Amount       float64 `json:"amount"`
	SubPartnerID string  `json:"sub_partner_id"`
}

type GetCustomerPaymentsRequest struct {
	Limit        int
	Page         int
	ID           string
	PayCurrency  string
	Status       string
	SubPartnerID string
	DateFrom     string
	DateTo       string
	OrderBy      string
	SortBy       string
}

// ----- Recurring payments (subscriptions) -----

type ListSubscriptionsRequest struct {
	Status             string
	SubscriptionPlanID string
	IsActive           *bool
	Limit              int
	Offset             int
}

type Subscription struct {
	ID                 json.Number `json:"id"`
	SubscriptionPlanID json.Number `json:"subscription_plan_id"`
	IsActive           bool        `json:"is_active"`
	Status             string      `json:"status"`
	ExpireDate         string      `json:"expire_date"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
}

type Plan struct {
	ID          json.Number `json:"id"`
	Title       string      `json:"title"`
	IntervalDay json.Number `json:"interval_day"`
	Amount      json.Number `json:"amount"`
	Currency    string      `json:"currency"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

type UpdatePlanRequest struct {
	Title       string  `json:"title,omitempty"`
	IntervalDay int     `json:"interval_day,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Currency    string  `json:"currency,omitempty"`
}
