package auth

import (
	"context"
	"errors"
	"testing"

	"encore.dev/beta/errs"
	"encore.dev/et"
)

func TestStoreAndConsumeOAuthState(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	if err := storeOAuthState(ctx, "state-1", "google", "verifier-1", ""); err != nil {
		t.Fatalf("store: %v", err)
	}
	verifier, link, err := consumeOAuthState(ctx, "state-1", "google")
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if verifier != "verifier-1" {
		t.Errorf("expected verifier-1, got %s", verifier)
	}
	if link != "" {
		t.Errorf("expected empty link user, got %s", link)
	}

	// Single-use.
	if _, _, err := consumeOAuthState(ctx, "state-1", "google"); err == nil {
		t.Fatal("second consume should fail (single-use)")
	}
}

func TestConsumeOAuthState_WrongProvider(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	if err := storeOAuthState(ctx, "state-2", "google", "v", ""); err != nil {
		t.Fatalf("store: %v", err)
	}
	if _, _, err := consumeOAuthState(ctx, "state-2", "apple"); err == nil {
		t.Fatal("expected error for wrong provider")
	}
}

func TestConsumeOAuthState_Unknown(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")
	if _, _, err := consumeOAuthState(ctx, "nope", "google"); err == nil {
		t.Fatal("expected error for unknown state")
	}
}

func TestExchangeCode_Roundtrip(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{Email: "exch@example.com", Username: "exch", Password: "password123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	code, err := issueExchangeCode(ctx, resp.User.ID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	userID, err := consumeExchangeCode(ctx, code)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if userID != resp.User.ID {
		t.Errorf("expected user %s, got %s", resp.User.ID, userID)
	}

	// Single-use.
	if _, err := consumeExchangeCode(ctx, code); err == nil {
		t.Fatal("expected error on reuse")
	}
}

func TestResolveOAuthUser_NewUser(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	userID, err := resolveOAuthUser(ctx, "google", &oauthIdentity{
		ProviderID: "google-123", Email: "g@example.com", Verified: true,
	}, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if userID == "" {
		t.Fatal("expected a user id")
	}

	// Linking the same provider identity again returns the same user.
	again, err := resolveOAuthUser(ctx, "google", &oauthIdentity{
		ProviderID: "google-123", Email: "g@example.com", Verified: true,
	}, "")
	if err != nil {
		t.Fatalf("resolve again: %v", err)
	}
	if again != userID {
		t.Errorf("expected same user %s, got %s", userID, again)
	}
}

func TestResolveOAuthUser_EmailCollisionDoesNotMerge(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	// Existing password user owns the email.
	if _, err := Register(ctx, &RegisterParams{Email: "collide@example.com", Username: "collide", Password: "password123"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	// A new provider identity with the same email must NOT merge into the
	// existing user; it creates a separate user with a NULL email.
	userID, err := resolveOAuthUser(ctx, "facebook", &oauthIdentity{
		ProviderID: "fb-999", Email: "collide@example.com", Verified: true,
	}, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var email string
	if err := db.QueryRow(ctx, `SELECT COALESCE(email, '') FROM users WHERE id = $1`, userID).Scan(&email); err != nil {
		t.Fatalf("query: %v", err)
	}
	if email != "" {
		t.Errorf("expected null email (collision avoided), got %s", email)
	}

	// And the account is linked to the new user, not the existing one.
	linked, err := findAccountUser(ctx, "facebook", "fb-999")
	if err != nil {
		t.Fatalf("find account: %v", err)
	}
	if linked != userID {
		t.Errorf("expected account linked to %s, got %s", userID, linked)
	}
}

func TestResolveOAuthUser_LinkToExistingUser(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{Email: "linkme@example.com", Username: "linkme", Password: "password123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	userID, err := resolveOAuthUser(ctx, "apple", &oauthIdentity{
		ProviderID: "apple-abc", Email: "linkme@example.com", Verified: true,
	}, resp.User.ID)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if userID != resp.User.ID {
		t.Errorf("expected link to existing user %s, got %s", resp.User.ID, userID)
	}
}

func TestResolveOAuthUser_LinkCollision(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	// First user owns the apple account.
	u1, err := Register(ctx, &RegisterParams{Email: "u1@example.com", Username: "user_u1", Password: "password123"})
	if err != nil {
		t.Fatalf("register u1: %v", err)
	}
	if _, err := resolveOAuthUser(ctx, "apple", &oauthIdentity{ProviderID: "apple-1"}, u1.User.ID); err != nil {
		t.Fatalf("resolve u1: %v", err)
	}

	// Second user tries to link the same apple account → conflict.
	u2, err := Register(ctx, &RegisterParams{Email: "u2@example.com", Username: "user_u2", Password: "password123"})
	if err != nil {
		t.Fatalf("register u2: %v", err)
	}
	_, err = resolveOAuthUser(ctx, "apple", &oauthIdentity{ProviderID: "apple-1"}, u2.User.ID)
	if err == nil {
		t.Fatal("expected AlreadyExists error")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.AlreadyExists {
		t.Errorf("expected AlreadyExists, got %v", err)
	}
}
