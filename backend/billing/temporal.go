package billing

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"sync"
	"time"

	myauth "comics-galore/backend/auth"

	"go.temporal.io/sdk/client"
)

// Contract shared with the temporal-worker module (see temporal-worker/):
// task queue "subscription", workflows "SubscribeWorkflow" and "BoostWorkflow",
// signals "deposit_paid" and "subscription_paid". The struct shapes below must
// match the worker's definitions field-for-field (JSON tags are the wire format).
const (
	temporalTaskQueue          = "subscription"
	subscribeWorkflowName      = "SubscribeWorkflow"
	boostWorkflowName          = "BoostWorkflow"
	depositPaidSignalName      = "deposit_paid"
	subscriptionPaidSignalName = "subscription_paid"
)

type paymentSignal struct {
	Status string `json:"status"`
}

var (
	temporalOnce   sync.Once
	temporalClient client.Client
	temporalErr    error
)

func getTemporalClient() (client.Client, error) {
	temporalOnce.Do(func() {
		address := firstNonEmpty(secrets.TemporalAddress, os.Getenv("TEMPORAL_ADDRESS"), "localhost:7233")
		namespace := firstNonEmpty(secrets.TemporalNamespace, os.Getenv("TEMPORAL_NAMESPACE"), "default")
		apiKey := firstNonEmpty(secrets.TemporalAPIKey, os.Getenv("TEMPORAL_API_KEY"))
		certPEM := firstNonEmpty(secrets.TemporalCert, os.Getenv("TEMPORAL_CERT"))
		keyPEM := firstNonEmpty(secrets.TemporalKey, os.Getenv("TEMPORAL_KEY"))

		opts := client.Options{HostPort: address, Namespace: namespace}
		switch {
		case apiKey != "":
			opts.Credentials = client.NewAPIKeyStaticCredentials(apiKey)
		case certPEM != "" && keyPEM != "":
			cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
			if err != nil {
				temporalErr = err
				return
			}
			opts.ConnectionOptions = client.ConnectionOptions{
				TLS: &tls.Config{Certificates: []tls.Certificate{cert}},
			}
		}
		temporalClient, temporalErr = client.Dial(opts)
	})
	return temporalClient, temporalErr
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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

// ----- SubscribeWorkflow (fund → subscribe → activate) -----

type subscribeWorkflowInput struct {
	UserID              string        `json:"user_id"`
	PlanID              string        `json:"plan_id"`
	Crypto              string        `json:"crypto"`
	CheckoutID          string        `json:"checkout_id"`
	SubscriptionTimeout time.Duration `json:"subscription_timeout"`
}

var (
	startSubscribeWorkflow      = defaultStartSubscribeWorkflow
	signalSubscribeDeposit      = defaultSignalSubscribeDeposit
	signalSubscribeSubscription = defaultSignalSubscribeSubscription
)

func defaultStartSubscribeWorkflow(ctx context.Context, checkoutID, userID, planID, crypto string, subscriptionTimeout time.Duration) error {
	c, err := getTemporalClient()
	if err != nil {
		return err
	}
	_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        checkoutID,
		TaskQueue: temporalTaskQueue,
	}, subscribeWorkflowName, subscribeWorkflowInput{
		UserID:              userID,
		PlanID:              planID,
		Crypto:              crypto,
		CheckoutID:          checkoutID,
		SubscriptionTimeout: subscriptionTimeout,
	})
	return err
}

func defaultSignalSubscribeDeposit(ctx context.Context, checkoutID, status string) error {
	c, err := getTemporalClient()
	if err != nil {
		return err
	}
	return c.SignalWorkflow(ctx, checkoutID, "", depositPaidSignalName, paymentSignal{Status: status})
}

func defaultSignalSubscribeSubscription(ctx context.Context, checkoutID, status string) error {
	c, err := getTemporalClient()
	if err != nil {
		return err
	}
	return c.SignalWorkflow(ctx, checkoutID, "", subscriptionPaidSignalName, paymentSignal{Status: status})
}

// ----- BoostWorkflow (create boost deposit → wait → grant boost) -----

type boostWorkflowInput struct {
	UserID     string `json:"user_id"`
	Downloads  int    `json:"downloads"`
	Crypto     string `json:"crypto"`
	CheckoutID string `json:"checkout_id"`
}

var startBoostWorkflow = defaultStartBoostWorkflow

func defaultStartBoostWorkflow(ctx context.Context, checkoutID, userID string, downloads int, crypto string) error {
	c, err := getTemporalClient()
	if err != nil {
		return err
	}
	_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        checkoutID,
		TaskQueue: temporalTaskQueue,
	}, boostWorkflowName, boostWorkflowInput{
		UserID:     userID,
		Downloads:  downloads,
		Crypto:     crypto,
		CheckoutID: checkoutID,
	})
	return err
}
