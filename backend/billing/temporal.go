package billing

import (
	"context"
	"os"
	"sync"
	"time"

	myauth "comics-galore/backend/auth"

	"go.temporal.io/sdk/client"
)

// Contract shared with the temporal-worker module (see temporal-worker/):
// workflow "SubscriptionWorkflow", task queue "subscription", signal
// "payment_received". The struct shapes below must match the worker's
// definitions field-for-field (JSON tags are the wire format).
const (
	temporalTaskQueue        = "subscription"
	subscriptionWorkflowName = "SubscriptionWorkflow"
	paymentSignalName        = "payment_received"
)

type subscriptionWorkflowInput struct {
	SubscriptionID string        `json:"subscription_id"`
	PaymentTimeout time.Duration `json:"payment_timeout"`
}

type paymentSignal struct {
	Status string `json:"status"`
}

// startSubscriptionWorkflow and signalSubscriptionWorkflow are overridable in
// tests. The defaults are best-effort: callers log errors rather than fail.
var (
	startSubscriptionWorkflow  = defaultStartSubscriptionWorkflow
	signalSubscriptionWorkflow = defaultSignalSubscriptionWorkflow
)

var (
	temporalOnce   sync.Once
	temporalClient client.Client
	temporalErr    error
)

func getTemporalClient() (client.Client, error) {
	temporalOnce.Do(func() {
		address := os.Getenv("TEMPORAL_ADDRESS")
		if address == "" {
			address = "localhost:7233"
		}
		namespace := os.Getenv("TEMPORAL_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}
		temporalClient, temporalErr = client.Dial(client.Options{
			HostPort:  address,
			Namespace: namespace,
		})
	})
	return temporalClient, temporalErr
}

func defaultStartSubscriptionWorkflow(ctx context.Context, subscriptionID string, timeout time.Duration) error {
	c, err := getTemporalClient()
	if err != nil {
		return err
	}
	_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        subscriptionID,
		TaskQueue: temporalTaskQueue,
	}, subscriptionWorkflowName, subscriptionWorkflowInput{
		SubscriptionID: subscriptionID,
		PaymentTimeout: timeout,
	})
	return err
}

func defaultSignalSubscriptionWorkflow(ctx context.Context, subscriptionID, status string) error {
	c, err := getTemporalClient()
	if err != nil {
		return err
	}
	return c.SignalWorkflow(ctx, subscriptionID, "", paymentSignalName, paymentSignal{Status: status})
}

// waitingPayTimeout returns the configured WAITING_PAY expiry (default 24h),
// used as the durable payment timeout in the subscription workflow.
func waitingPayTimeout(ctx context.Context) time.Duration {
	hours := 24
	if cfg, err := myauth.GetBillingConfig(ctx); err == nil && cfg.WaitingPayExpiryHours > 0 {
		hours = cfg.WaitingPayExpiryHours
	}
	return time.Duration(hours) * time.Hour
}
