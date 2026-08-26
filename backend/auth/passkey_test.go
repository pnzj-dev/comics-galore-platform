package auth

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"encore.dev/beta/auth"
	"encore.dev/et"
)

// registerTestUser creates a real user and returns its auth context.
func registerTestUser(t *testing.T, ctx context.Context, email string) context.Context {
	t.Helper()
	local := strings.ToLower(strings.Split(email, "@")[0])
	local = strings.NewReplacer("-", "_", ".", "_", "@", "_").Replace(local)
	resp, err := Register(ctx, &RegisterParams{Email: email, Password: "password123", Username: local})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID, Email: resp.User.Email, Role: resp.User.Role, Tier: resp.User.Tier,
	})
}

func TestStoreAndConsumeChallenge(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	authCtx := registerTestUser(t, ctx, "passkey-a@example.com")

	// Begin a registration ceremony to obtain a real challenge.
	opts, err := PasskeyRegisterOptions(authCtx, &PasskeyRegisterOptionsParams{Name: "MacBook"})
	if err != nil {
		t.Fatalf("register options: %v", err)
	}

	var creation struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
		} `json:"publicKey"`
	}
	if err := json.Unmarshal(opts.Options, &creation); err != nil {
		t.Fatalf("parse options: %v", err)
	}
	if creation.PublicKey.Challenge == "" {
		t.Fatal("expected a challenge in options")
	}

	// Single-use: consuming once succeeds, second consume fails.
	if _, err := consumeChallenge(ctx, creation.PublicKey.Challenge, "register"); err != nil {
		t.Fatalf("first consume should succeed: %v", err)
	}
	if _, err := consumeChallenge(ctx, creation.PublicKey.Challenge, "register"); err == nil {
		t.Fatal("second consume should fail (single-use)")
	}
}

func TestConsumeChallenge_WrongPurpose(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	authCtx := registerTestUser(t, ctx, "passkey-b@example.com")
	opts, err := PasskeyRegisterOptions(authCtx, &PasskeyRegisterOptionsParams{Name: "MacBook"})
	if err != nil {
		t.Fatalf("register options: %v", err)
	}
	var creation struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
		} `json:"publicKey"`
	}
	_ = json.Unmarshal(opts.Options, &creation)

	if _, err := consumeChallenge(ctx, creation.PublicKey.Challenge, "login"); err == nil {
		t.Fatal("expected error consuming a register challenge as login")
	}
}

func TestConsumeChallenge_Unknown(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")
	if _, err := consumeChallenge(ctx, "does-not-exist", "login"); err == nil {
		t.Fatal("expected error for unknown challenge")
	}
}

func TestPasskeyLoginOptions_StoresChallenge(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	opts, err := PasskeyLoginOptions(ctx)
	if err != nil {
		t.Fatalf("login options: %v", err)
	}
	var assertion struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
		} `json:"publicKey"`
	}
	if err := json.Unmarshal(opts.Options, &assertion); err != nil {
		t.Fatalf("parse assertion: %v", err)
	}
	if assertion.PublicKey.Challenge == "" {
		t.Fatal("expected challenge")
	}
	if _, err := consumeChallenge(ctx, assertion.PublicKey.Challenge, "login"); err != nil {
		t.Fatalf("challenge should be consumable: %v", err)
	}
}

func TestListAndDeletePasskey(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	authCtx := registerTestUser(t, ctx, "passkey-c@example.com")

	resp, err := ListPasskeys(authCtx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(resp.Passkeys) != 0 {
		t.Fatalf("expected no passkeys, got %d", len(resp.Passkeys))
	}

	// Deleting a nonexistent id should not error (idempotent).
	if err := DeletePasskey(authCtx, "00000000-0000-0000-0000-000000000000"); err != nil {
		t.Fatalf("delete should be idempotent: %v", err)
	}
}
