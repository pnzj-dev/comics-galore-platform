package nowpayments

import "context"

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
	PaymentID   string  `json:"payment_id"`
	PayAddress  string  `json:"pay_address"`
	PayAmount   float64 `json:"pay_amount"`
	PayCurrency string  `json:"pay_currency"`
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

type PaymentsProvider interface {
	EstimatePrice(ctx context.Context, req EstimateRequest) (*EstimateResponse, error)
	ListCurrencies(ctx context.Context) ([]string, error)
	CheckBalance(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error)
	CreateCustomer(ctx context.Context, name string) (string, error)
	CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error)
	CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error)
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)
}
