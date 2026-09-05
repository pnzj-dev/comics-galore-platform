// Package turnstile provides a shared Cloudflare Turnstile verification step
// that the auth, comics and social services call before processing user input.
//
// The widget is rendered in the SvelteKit frontend; the token is sent with the
// request and verified server-side here (single source of truth for the bot
// check, matching the rest of the Encore backend).
package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encore.dev"
	"encore.dev/beta/errs"
)

var secrets struct {
	TurnstileSecret    string
	TurnstileHostnames string // comma-separated list of allowed frontend hostnames
}

// VerifyParams is the request body for the private Verify endpoint.
type VerifyParams struct {
	Token  string `json:"token"`
	Action string `json:"action"`
}

// Verify checks a Turnstile token via the Cloudflare siteverify endpoint.
//
// When running as a unit test, or when TurnstileSecret is unset, the check is
// inert (returns nil). Otherwise it fails closed: the token must be present,
// redeem successfully, match the expected action, and come from an allowed
// hostname.
//
//encore:api private method=POST path=/turnstile/verify
func Verify(ctx context.Context, p *VerifyParams) error {
	if encore.Meta().Environment.Type == encore.EnvTest {
		return nil
	}
	if strings.TrimSpace(secrets.TurnstileSecret) == "" {
		return nil
	}

	token := strings.TrimSpace(p.Token)
	if token == "" || len(token) > 2048 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "turnstile token is required"}
	}

	ok, err := siteverify(ctx, secrets.TurnstileSecret, token, p.Action)
	if err != nil {
		return &errs.Error{Code: errs.Internal, Message: "turnstile verification failed"}
	}
	if !ok {
		return &errs.Error{Code: errs.PermissionDenied, Message: "turnstile verification failed"}
	}
	return nil
}

// siteverify performs the canonical server-side validation and returns whether
// the token is valid for the given action and an allowed hostname.
func siteverify(ctx context.Context, secret, token, action string) (bool, error) {
	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result struct {
		Success    bool     `json:"success"`
		Action     string   `json:"action"`
		Hostname   string   `json:"hostname"`
		ErrorCodes []string `json:"error-codes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("siteverify parse: %w", err)
	}

	if !result.Success {
		return false, nil
	}
	if result.Action != action {
		return false, nil
	}
	if !hostnameAllowed(result.Hostname, secrets.TurnstileHostnames) {
		return false, nil
	}
	return true, nil
}

// hostnameAllowed reports whether the widget's hostname is in the allowlist.
// An empty allowlist accepts any hostname.
func hostnameAllowed(hostname, allowlist string) bool {
	hostname = strings.TrimSpace(hostname)
	if strings.TrimSpace(allowlist) == "" {
		return true
	}
	for _, h := range strings.Split(allowlist, ",") {
		if strings.TrimSpace(h) == hostname {
			return true
		}
	}
	return false
}
