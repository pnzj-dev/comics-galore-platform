package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

func TestMe(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	u := createTestUser(ctx, t, "me@example.com", "user", "free")

	authCtx := auth.WithContext(ctx, auth.UID(u.ID), &AuthData{
		UserID: u.ID,
		Email:  u.Email,
		Role:   u.Role,
		Tier:   u.Tier,
	})

	user, err := Me(authCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "me@example.com" {
		t.Errorf("expected email me@example.com, got %s", user.Email)
	}
	if user.ID != u.ID {
		t.Errorf("expected ID %s, got %s", u.ID, user.ID)
	}
}

func TestAdminBanUser(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	victim := createTestUser(ctx, t, "victim@example.com", "user", "free")

	adminCtx := auth.WithContext(ctx, auth.UID(testAdminID), &AuthData{
		UserID: testAdminID, Email: "admin@example.com", Role: "admin", Tier: "platinum",
	})

	if err := AdminBanUser(adminCtx, victim.ID, &BanUserParams{Reason: "spam"}); err != nil {
		t.Fatalf("ban error: %v", err)
	}

	var bannedAt *time.Time
	if err := db.QueryRow(ctx, `SELECT banned_at FROM users WHERE id = $1`, victim.ID).Scan(&bannedAt); err != nil {
		t.Fatalf("query error: %v", err)
	}
	if bannedAt == nil {
		t.Error("expected banned_at to be set")
	}
}

func TestAdminUnbanUser(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	victim := createTestUser(ctx, t, "pardoned@example.com", "user", "free")

	adminCtx := auth.WithContext(ctx, auth.UID(testAdminID), &AuthData{
		UserID: testAdminID, Email: "admin@example.com", Role: "admin", Tier: "platinum",
	})

	if err := AdminBanUser(adminCtx, victim.ID, &BanUserParams{Reason: "spam"}); err != nil {
		t.Fatalf("ban error: %v", err)
	}
	if err := AdminUnbanUser(adminCtx, victim.ID); err != nil {
		t.Fatalf("unban error: %v", err)
	}

	var bannedAt *time.Time
	if err := db.QueryRow(ctx, `SELECT banned_at FROM users WHERE id = $1`, victim.ID).Scan(&bannedAt); err != nil {
		t.Fatalf("query error: %v", err)
	}
	if bannedAt != nil {
		t.Error("expected banned_at to be NULL after unban")
	}
}

func TestAdminBanUser_NonAdmin(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	u := createTestUser(ctx, t, "user@example.com", "user", "free")

	userCtx := auth.WithContext(ctx, auth.UID(u.ID), &AuthData{
		UserID: u.ID, Email: u.Email, Role: "user", Tier: u.Tier,
	})

	err := AdminBanUser(userCtx, "some-other-id", &BanUserParams{Reason: "test"})
	if err == nil {
		t.Fatal("expected error for non-admin, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestAdminSuspendUser(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	victim := createTestUser(ctx, t, "suspendme@example.com", "user", "free")

	adminCtx := auth.WithContext(ctx, auth.UID(testAdminID), &AuthData{
		UserID: testAdminID, Email: "admin@example.com", Role: "admin", Tier: "platinum",
	})

	if err := AdminSuspendUser(adminCtx, victim.ID, &BanUserParams{Reason: "violation"}); err != nil {
		t.Fatalf("suspend error: %v", err)
	}

	var suspendedAt *time.Time
	if err := db.QueryRow(ctx, `SELECT suspended_at FROM users WHERE id = $1`, victim.ID).Scan(&suspendedAt); err != nil {
		t.Fatalf("query error: %v", err)
	}
	if suspendedAt == nil {
		t.Error("expected suspended_at to be set")
	}
}

func TestAdminUnsuspendUser(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	victim := createTestUser(ctx, t, "unsuspendme@example.com", "user", "free")

	adminCtx := auth.WithContext(ctx, auth.UID(testAdminID), &AuthData{
		UserID: testAdminID, Email: "admin@example.com", Role: "admin", Tier: "platinum",
	})

	if err := AdminSuspendUser(adminCtx, victim.ID, &BanUserParams{Reason: "violation"}); err != nil {
		t.Fatalf("suspend error: %v", err)
	}
	if err := AdminUnsuspendUser(adminCtx, victim.ID); err != nil {
		t.Fatalf("unsuspend error: %v", err)
	}

	var suspendedAt *time.Time
	if err := db.QueryRow(ctx, `SELECT suspended_at FROM users WHERE id = $1`, victim.ID).Scan(&suspendedAt); err != nil {
		t.Fatalf("query error: %v", err)
	}
	if suspendedAt != nil {
		t.Error("expected suspended_at to be NULL after unsuspend")
	}
}

func TestGetContentPolicy_ForbidMatureForFree(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()

	adminCtx := auth.WithContext(ctx, auth.UID(testAdminID), &AuthData{
		UserID: testAdminID, Email: "admin@example.com", Role: "admin", Tier: "platinum",
	})

	// Default policy: not forbidden.
	p, err := GetContentPolicy(ctx)
	if err != nil {
		t.Fatalf("policy error: %v", err)
	}
	if p.ForbidMatureForFree {
		t.Error("expected forbid_mature_for_free=false by default")
	}

	// Enable it.
	settings, err := GetAdminSettings(adminCtx)
	if err != nil {
		t.Fatalf("get settings error: %v", err)
	}
	settings.ForbidMatureForFree = true
	if _, err := SaveAdminSettings(adminCtx, settings); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	p2, err := GetContentPolicy(ctx)
	if err != nil {
		t.Fatalf("policy error: %v", err)
	}
	if !p2.ForbidMatureForFree {
		t.Error("expected forbid_mature_for_free=true after save")
	}
}
