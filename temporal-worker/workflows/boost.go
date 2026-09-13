package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// BoostWorkflowInput is the input to BoostWorkflow.
type BoostWorkflowInput struct {
	UserID     string `json:"user_id"`
	Downloads  int    `json:"downloads"`
	Crypto     string `json:"crypto"`
	CheckoutID string `json:"checkout_id"`
}

// CreateBoostDepositInput is the input to the CreateBoostDeposit activity.
type CreateBoostDepositInput struct {
	UserID     string `json:"user_id"`
	Downloads  int    `json:"downloads"`
	Crypto     string `json:"crypto"`
	CheckoutID string `json:"checkout_id"`
}

// CreateBoostDepositResult is the result of the CreateBoostDeposit activity.
type CreateBoostDepositResult struct {
	DepositID      string `json:"deposit_id"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// BoostWorkflow creates a quota-boost deposit and grants the boost once the
// deposit completes (or expires it on timeout / non-finished status).
func BoostWorkflow(ctx workflow.Context, input BoostWorkflowInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var dep CreateBoostDepositResult
	if err := workflow.ExecuteActivity(ctx, CreateBoostDepositActivityName, CreateBoostDepositInput{
		UserID:     input.UserID,
		Downloads:  input.Downloads,
		Crypto:     input.Crypto,
		CheckoutID: input.CheckoutID,
	}).Get(ctx, &dep); err != nil {
		return err
	}

	var depStatus PaymentSignal
	gotDep := false
	depCh := workflow.GetSignalChannel(ctx, DepositPaidSignalName)
	depTimer := workflow.NewTimer(ctx, time.Duration(dep.TimeoutSeconds)*time.Second)

	sel := workflow.NewSelector(ctx)
	sel.AddReceive(depCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &depStatus)
		gotDep = true
	})
	sel.AddFuture(depTimer, func(workflow.Future) {})
	sel.Select(ctx)

	if gotDep && depStatus.Status == "completed" {
		return workflow.ExecuteActivity(ctx, CompleteDepositActivityName, DepositInput{DepositID: dep.DepositID}).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, ExpireDepositActivityName, DepositInput{DepositID: dep.DepositID}).Get(ctx, nil)
}
