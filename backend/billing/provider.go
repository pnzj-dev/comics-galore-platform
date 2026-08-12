package billing

import "context"

type EstimateRequest struct {
	Amount    float64
	Currency  string
	Crypto    string
}

type EstimateResponse struct {
	EstimatedAmount float64
	FromCurrency    string
	ToCurrency      string
}

type BalanceEntry struct {
	Amount        float64
	PendingAmount float64
}

type SubscriptionRequest struct {
	PlanID            string
	SubPartnerID      string
	SubscriptionPlanID string
}

type SubscriptionResponse struct {
	SubscriptionID string
	Status         string
}

type DepositRequest struct {
	Crypto        string
	AmountUSD     float64
	SubPartnerID  string
	IPNCallbackURL string
}

type DepositResponse struct {
	PaymentID  string
	PayAddress string
	PayAmount  float64
	PayCurrency string
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
	CheckBalance(ctx context.Context, subPartnerID string) (map[string]BalanceEntry, error)
	CreateCustomer(ctx context.Context, name string) (string, error)
	CreateSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResponse, error)
	CreateDeposit(ctx context.Context, req DepositRequest) (*DepositResponse, error)
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error)
}
