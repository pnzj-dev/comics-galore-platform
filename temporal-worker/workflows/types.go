package workflows

// PaymentSignal is the payload of the "deposit_paid" / "subscription_paid"
// signals.
type PaymentSignal struct {
	Status string `json:"status"`
}

// ActivateInput is the input to the ActivateSubscription activity.
type ActivateInput struct {
	SubscriptionID string `json:"subscription_id"`
}

// ExpireInput is the input to the ExpireSubscription activity.
type ExpireInput struct {
	SubscriptionID string `json:"subscription_id"`
}

// DepositInput is the input to the CompleteDeposit / ExpireDeposit activities.
type DepositInput struct {
	DepositID string `json:"deposit_id"`
}
