// Package workflows contains Temporal workflow definitions for Comics Galore.
//
// The workflow bodies are deterministic (no side effects): all DB and API
// mutations happen inside activities, which call back into the Encore backend.
package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SubscriptionWorkflowInput is the input to SubscriptionWorkflow.
type SubscriptionWorkflowInput struct {
	SubscriptionID string        `json:"subscription_id"`
	PaymentTimeout time.Duration `json:"payment_timeout"`
}

// PaymentSignal is the payload of the "payment_received" / "deposit_paid" /
// "subscription_paid" signals.
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

// SubscriptionWorkflow orchestrates a single subscription's initial payment
// lifecycle: wait for a payment confirmation (signal) or a timeout, then
// activate or expire the subscription.
//
// Renewal is intentionally out of scope for this first pass: after the first
// activation the workflow completes, and later "finished" IPNs for the same
// subscription are no-ops (the subscription stays active).
func SubscriptionWorkflow(ctx workflow.Context, input SubscriptionWorkflowInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var signal PaymentSignal
	gotPayment := false
	paymentCh := workflow.GetSignalChannel(ctx, PaymentSignalName)
	timeoutFuture := workflow.NewTimer(ctx, input.PaymentTimeout)

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(paymentCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signal)
		gotPayment = true
	})
	selector.AddFuture(timeoutFuture, func(workflow.Future) {})
	selector.Select(ctx)

	if gotPayment && signal.Status == "finished" {
		return workflow.ExecuteActivity(ctx, ActivateActivityName, ActivateInput{
			SubscriptionID: input.SubscriptionID,
		}).Get(ctx, nil)
	}

	return workflow.ExecuteActivity(ctx, ExpireSubscriptionActivityName, ExpireInput{
		SubscriptionID: input.SubscriptionID,
	}).Get(ctx, nil)
}
