package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type SubscribeWorkflowInput struct {
	UserID              string        `json:"user_id"`
	PlanID              string        `json:"plan_id"`
	Crypto              string        `json:"crypto"`
	CheckoutID          string        `json:"checkout_id"`
	SubscriptionTimeout time.Duration `json:"subscription_timeout"`
}

// Activity inputs/results for the combined workflow.
type CheckBalanceInput struct {
	UserID string `json:"user_id"`
	Crypto string `json:"crypto"`
}

type CheckBalanceResult struct {
	HasBalance bool `json:"has_balance"`
}

type CreateDepositInput struct {
	UserID     string `json:"user_id"`
	PlanID     string `json:"plan_id"`
	Crypto     string `json:"crypto"`
	CheckoutID string `json:"checkout_id"`
}

type CreateDepositResult struct {
	DepositID      string `json:"deposit_id"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type CreateSubscriptionInput struct {
	UserID     string `json:"user_id"`
	PlanID     string `json:"plan_id"`
	CheckoutID string `json:"checkout_id"`
}

type CreateSubscriptionResult struct {
	SubscriptionID string `json:"subscription_id"`
}

func SubscribeWorkflow(ctx workflow.Context, input SubscribeWorkflowInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: check balance.
	var balance CheckBalanceResult
	if err := workflow.ExecuteActivity(ctx, CheckBalanceActivityName, CheckBalanceInput{
		UserID: input.UserID,
		Crypto: input.Crypto,
	}).Get(ctx, &balance); err != nil {
		return err
	}

	// Step 2: deposit if the user has no funds.
	if !balance.HasBalance {
		var dep CreateDepositResult
		if err := workflow.ExecuteActivity(ctx, CreateDepositActivityName, CreateDepositInput{
			UserID:     input.UserID,
			PlanID:     input.PlanID,
			Crypto:     input.Crypto,
			CheckoutID: input.CheckoutID,
		}).Get(ctx, &dep); err != nil {
			return err
		}

		var depStatus PaymentSignal
		gotDep := false
		depCh := workflow.GetSignalChannel(ctx, DepositPaidSignalName)
		depTimer := workflow.NewTimer(ctx, time.Duration(dep.TimeoutSeconds)*time.Second)

		depSel := workflow.NewSelector(ctx)
		depSel.AddReceive(depCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &depStatus)
			gotDep = true
		})
		depSel.AddFuture(depTimer, func(workflow.Future) {})
		depSel.Select(ctx)

		if !gotDep || depStatus.Status != "completed" {
			_ = workflow.ExecuteActivity(ctx, ExpireDepositActivityName, DepositInput{DepositID: dep.DepositID}).Get(ctx, nil)
			return nil
		}
	}

	// Step 3: create the subscription.
	var sub CreateSubscriptionResult
	if err := workflow.ExecuteActivity(ctx, CreateSubscriptionActivityName, CreateSubscriptionInput{
		UserID:     input.UserID,
		PlanID:     input.PlanID,
		CheckoutID: input.CheckoutID,
	}).Get(ctx, &sub); err != nil {
		return err
	}

	// Step 4: wait for the subscription payment (or timeout).
	var subStatus PaymentSignal
	gotSub := false
	subCh := workflow.GetSignalChannel(ctx, SubscriptionPaidSignalName)
	subTimer := workflow.NewTimer(ctx, input.SubscriptionTimeout)

	subSel := workflow.NewSelector(ctx)
	subSel.AddReceive(subCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &subStatus)
		gotSub = true
	})
	subSel.AddFuture(subTimer, func(workflow.Future) {})
	subSel.Select(ctx)

	if gotSub && subStatus.Status == "finished" {
		return workflow.ExecuteActivity(ctx, ActivateActivityName, ActivateInput{SubscriptionID: sub.SubscriptionID}).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, ExpireSubscriptionActivityName, ExpireInput{SubscriptionID: sub.SubscriptionID}).Get(ctx, nil)
}
