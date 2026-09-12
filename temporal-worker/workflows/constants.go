// Package workflows contains Temporal workflow definitions for Comics Galore.
//
// The workflow bodies are deterministic (no side effects): all DB and API
// mutations happen inside activities, which call back into the Encore backend.
package workflows

// Signal and activity names form the contract between the Encore backend
// (Temporal client) and this worker. Keep them stable across deployments.
const (
	TaskQueue = "subscription"

	// Legacy (Phase-1) subscription workflow signal.
	PaymentSignalName = "payment_received"

	// Combined SubscribeWorkflow signals.
	DepositPaidSignalName      = "deposit_paid"
	SubscriptionPaidSignalName = "subscription_paid"

	// Activity names (shared by both workflows).
	ActivateActivityName           = "ActivateSubscription"
	ExpireSubscriptionActivityName = "ExpireSubscription"
	CheckBalanceActivityName       = "CheckBalance"
	CreateDepositActivityName      = "CreateDeposit"
	CreateSubscriptionActivityName = "CreateSubscription"
	CompleteDepositActivityName    = "CompleteDeposit"
	ExpireDepositActivityName      = "ExpireDeposit"
)
