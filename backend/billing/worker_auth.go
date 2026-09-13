package billing

import (
	"crypto/subtle"
	"strings"

	"encore.dev"
	"encore.dev/beta/errs"
)

// requireWorker verifies the X-Worker-Token header against WorkerSecret.
// It is inert when WorkerSecret is empty (local development).
func requireWorker() error {
	secret := strings.TrimSpace(secrets.WorkerSecret)
	if secret == "" {
		return nil
	}
	token := strings.TrimSpace(encore.CurrentRequest().Headers.Get("X-Worker-Token"))
	if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
		return &errs.Error{Code: errs.PermissionDenied, Message: "unauthorized"}
	}
	return nil
}
