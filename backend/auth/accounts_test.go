package auth

import (
	"context"
	"errors"
	"testing"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
)

func TestCountAuthMethods(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	// Fresh password user = 1 method (password).
	resp, err := Register(ctx, &RegisterParams{Email: "count@example.com", Username: "count", Password: "password123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	count, err := countAuthMethods(ctx, resp.User.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 method, got %d", count)
	}

	// Add an OAuth account → 2.
	if _, err := resolveOAuthUser(ctx, "google", &oauthIdentity{ProviderID: "g-1"}, resp.User.ID); err != nil {
		t.Fatalf("link: %v", err)
	}
	count, _ = countAuthMethods(ctx, resp.User.ID)
	if count != 2 {
		t.Errorf("expected 2 methods, got %d", count)
	}
}

func TestUnlinkAccount_LastMethodGuard(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	// OAuth-only user (no password) → 1 method (the OAuth account).
	userID, err := resolveOAuthUser(ctx, "google", &oauthIdentity{ProviderID: "solo-1"}, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	authCtx := auth.WithContext(ctx, auth.UID(userID), &AuthData{UserID: userID, Role: "user", Tier: "free"})
	accts, err := ListAccounts(authCtx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(accts.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accts.Accounts))
	}

	// Removing the only method must fail.
	err = UnlinkAccount(authCtx, accts.Accounts[0].ID)
	if err == nil {
		t.Fatal("expected error removing last auth method")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestUnlinkAccount_SucceedsWithBackup(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	// Password user with an added OAuth account → 2 methods.
	resp, err := Register(ctx, &RegisterParams{Email: "backup@example.com", Username: "backup", Password: "password123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := resolveOAuthUser(ctx, "twitter", &oauthIdentity{ProviderID: "tw-1"}, resp.User.ID); err != nil {
		t.Fatalf("link: %v", err)
	}

	authCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{UserID: resp.User.ID, Role: "user", Tier: "free"})
	accts, err := ListAccounts(authCtx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(accts.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accts.Accounts))
	}

	if err := UnlinkAccount(authCtx, accts.Accounts[0].ID); err != nil {
		t.Fatalf("unlink should succeed with password backup: %v", err)
	}
}

func TestUnlinkAccount_NotOwner(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	u1, _ := Register(ctx, &RegisterParams{Email: "owner@example.com", Username: "owner", Password: "password123"})
	if _, err := resolveOAuthUser(ctx, "google", &oauthIdentity{ProviderID: "own-1"}, u1.User.ID); err != nil {
		t.Fatalf("link: %v", err)
	}

	u2, _ := Register(ctx, &RegisterParams{Email: "other@example.com", Username: "other", Password: "password123"})
	authCtx2 := auth.WithContext(ctx, auth.UID(u2.User.ID), &AuthData{UserID: u2.User.ID, Role: "user", Tier: "free"})

	// Find u1's account id.
	var accountID string
	_ = db.QueryRow(ctx, `SELECT id FROM auth_accounts WHERE user_id = $1`, u1.User.ID).Scan(&accountID)

	err := UnlinkAccount(authCtx2, accountID)
	if err == nil {
		t.Fatal("expected NotFound for another user's account")
	}
}
