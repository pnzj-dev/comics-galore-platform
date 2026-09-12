// Package activities implements the side-effecting operations that the
// subscription workflow orchestrates. Each activity calls back into the Encore
// backend over HTTP, keeping NowPayments and DB ownership inside Encore.
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

// ActivateSubscription asks the Encore backend to mark a subscription active
// and promote the user's tier. It is idempotent.
func (a *Activities) ActivateSubscription(ctx context.Context, input workflows.ActivateInput) error {
	return a.post(ctx, "/billing/subscriptions/"+input.SubscriptionID+"/activate")
}

// ExpireSubscription asks the Encore backend to expire a subscription and
// downgrade the user to the free tier. It is idempotent.
func (a *Activities) ExpireSubscription(ctx context.Context, input workflows.ExpireInput) error {
	return a.post(ctx, "/billing/subscriptions/"+input.SubscriptionID+"/expire")
}

func (a *Activities) post(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BackendURL+path, bytes.NewReader([]byte{}))
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
	return nil
}
