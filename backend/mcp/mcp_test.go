package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"comics-galore/backend/auth"
	"comics-galore/backend/comics"

	"encore.dev/beta/errs"
	"encore.dev/et"
)

const testUserID = "550e8400-e29b-41d4-a716-446655440001"

func isolateMcpDB(t *testing.T) (restore func()) {
	t.Helper()
	isolated, err := et.NewTestDatabase(context.Background(), "mcpdb")
	if err != nil {
		t.Fatalf("new test database: %v", err)
	}
	original := db
	db = isolated
	return func() { db = original }
}

func hashKey(k string) string {
	sum := sha256.Sum256([]byte(k))
	return hex.EncodeToString(sum[:])
}

func insertKey(t *testing.T, ctx context.Context, key, userID string) {
	t.Helper()
	if _, err := db.Exec(ctx, `INSERT INTO mcp_keys (key_hash, user_id, label) VALUES ($1, $2, 'test')`, hashKey(key), userID); err != nil {
		t.Fatalf("insert key: %v", err)
	}
}

func mockRole(t *testing.T, role string) {
	t.Helper()
	og := getUserRole
	getUserRole = func(ctx context.Context, p *auth.GetUserRoleParams) (*auth.GetUserRoleResponse, error) {
		return &auth.GetUserRoleResponse{Role: role}, nil
	}
	t.Cleanup(func() { getUserRole = og })
}

func TestResolveMcpKey_Valid(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "moderator")

	insertKey(t, ctx, "secret-key-123", testUserID)

	userID, role, err := resolveMcpKey(ctx, "secret-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != testUserID {
		t.Errorf("expected %s, got %s", testUserID, userID)
	}
	if role != "moderator" {
		t.Errorf("expected moderator, got %s", role)
	}
}

func TestResolveMcpKey_Invalid(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()

	_, _, err := resolveMcpKey(ctx, "wrong-key")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestResolveMcpKey_Revoked(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "moderator")

	insertKey(t, ctx, "revoked-key", testUserID)
	if _, err := db.Exec(ctx, `UPDATE mcp_keys SET revoked_at = now() WHERE key_hash = $1`, hashKey("revoked-key")); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	if _, _, err := resolveMcpKey(ctx, "revoked-key"); err == nil {
		t.Fatal("expected error for revoked key")
	}
}

func TestModerateComic_NonModeratorDenied(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "user")

	insertKey(t, ctx, "user-key", testUserID)

	err := doModerateComic(ctx, "user-key", &ModerateComicParams{Action: "approve", ComicID: "c1"})
	if err == nil {
		t.Fatal("expected error for non-moderator")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestModerateComic_Approve(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "moderator")

	var called bool
	og := approveComic
	approveComic = func(ctx context.Context, p *comics.InternalModerateParams) error {
		called = true
		if p.ActorID != testUserID {
			t.Errorf("expected actor %s, got %s", testUserID, p.ActorID)
		}
		return nil
	}
	t.Cleanup(func() { approveComic = og })

	insertKey(t, ctx, "mod-key", testUserID)

	if err := doModerateComic(ctx, "mod-key", &ModerateComicParams{Action: "approve", ComicID: "c1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected approveComic to be called")
	}
}
