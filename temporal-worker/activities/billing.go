// Package activities implements the side-effecting operations that the
// subscription/checkout workflows orchestrate. Each activity calls back into
// the Encore backend over HTTP, keeping NowPayments and DB ownership in Encore.
package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"comics-galore/temporal-worker/workflows"
)

// Activities holds configuration shared by all activities.
type Activities struct {
	// BackendURL is the base URL of the Encore backend (e.g. http://localhost:4000).
	BackendURL string
}

// ActivateSubscription marks a subscription active and promotes the tier.
func (a *Activities) ActivateSubscription(ctx context.Context, input workflows.ActivateInput) error {
	return a.post(ctx, "/billing/subscriptions/"+input.SubscriptionID+"/activate")
}

// ExpireSubscription expires a subscription and downgrades the tier.
func (a *Activities) ExpireSubscription(ctx context.Context, input workflows.ExpireInput) error {
	return a.post(ctx, "/billing/subscriptions/"+input.SubscriptionID+"/expire")
}

// CompleteDeposit marks a deposit completed and grants its quota boost.
func (a *Activities) CompleteDeposit(ctx context.Context, input workflows.DepositInput) error {
	return a.post(ctx, "/billing/deposits/"+input.DepositID+"/complete")
}

// ExpireDeposit marks a deposit expired.
func (a *Activities) ExpireDeposit(ctx context.Context, input workflows.DepositInput) error {
	return a.post(ctx, "/billing/deposits/"+input.DepositID+"/expire")
}

// CheckBalance reports whether the user has a non-zero balance for the crypto.
func (a *Activities) CheckBalance(ctx context.Context, input workflows.CheckBalanceInput) (workflows.CheckBalanceResult, error) {
	var out workflows.CheckBalanceResult
	err := a.postJSON(ctx, "/billing/internal/check-balance", map[string]string{
		"user_id": input.UserID,
		"crypto":  input.Crypto,
	}, &out)
	return out, err
}

// CreateDeposit creates a NowPayments deposit and returns its id + timeout.
func (a *Activities) CreateDeposit(ctx context.Context, input workflows.CreateDepositInput) (workflows.CreateDepositResult, error) {
	var out workflows.CreateDepositResult
	err := a.postJSON(ctx, "/billing/internal/create-deposit", map[string]string{
		"user_id":     input.UserID,
		"plan_id":     input.PlanID,
		"crypto":      input.Crypto,
		"checkout_id": input.CheckoutID,
	}, &out)
	return out, err
}

// CreateSubscription creates a NowPayments subscription and returns its id.
func (a *Activities) CreateSubscription(ctx context.Context, input workflows.CreateSubscriptionInput) (workflows.CreateSubscriptionResult, error) {
	var out workflows.CreateSubscriptionResult
	err := a.postJSON(ctx, "/billing/internal/create-subscription", map[string]string{
		"user_id":     input.UserID,
		"plan_id":     input.PlanID,
		"checkout_id": input.CheckoutID,
	}, &out)
	return out, err
}

func (a *Activities) post(ctx context.Context, path string) error {
	return a.postJSON(ctx, path, nil, nil)
}

func (a *Activities) postJSON(ctx context.Context, path string, body any, out any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BackendURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var e struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("encore %s: %s (status %d)", path, e.Message, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
