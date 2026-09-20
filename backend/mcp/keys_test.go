package mcp

import (
	"context"
	"errors"
	"testing"

	myauth "comics-galore/backend/auth"

	encoreauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

func adminCtx(userID string) context.Context {
	return encoreauth.WithContext(context.Background(), encoreauth.UID(userID), &myauth.AuthData{
		UserID: userID, Email: "admin@example.com", Role: "admin", Tier: "platinum",
	})
}

func mockUsersInfo(t *testing.T) {
	t.Helper()
	og := getUsersInfo
	getUsersInfo = func(ctx context.Context, p *myauth.GetUsersInfoParams) (*myauth.GetUsersInfoResponse, error) {
		return &myauth.GetUsersInfoResponse{Users: []myauth.UserPublicInfo{}}, nil
	}
	t.Cleanup(func() { getUsersInfo = og })
}

func TestAdminCreateMcpKey(t *testing.T) {
	ctx := adminCtx(testUserID)
	restore := isolateMcpDB(t)
	defer restore()

	resp, err := AdminCreateMcpKey(ctx, &CreateMcpKeyParams{Label: "claude"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(resp.Key) == 0 {
		t.Error("expected full key")
	}
	if resp.Info.KeySuffix != resp.Key[len(resp.Key)-4:] {
		t.Errorf("suffix %q != last-4 of key", resp.Info.KeySuffix)
	}
	if resp.Info.UserID != testUserID {
		t.Errorf("expected bound user %s, got %s", testUserID, resp.Info.UserID)
	}
}

func TestAdminCreateMcpKey_BoundToChosenUser(t *testing.T) {
	ctx := adminCtx(testUserID)
	restore := isolateMcpDB(t)
	defer restore()

	other := "550e8400-e29b-41d4-a716-446655440002"
	resp, err := AdminCreateMcpKey(ctx, &CreateMcpKeyParams{Label: "bot", UserID: other})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.Info.UserID != other {
		t.Errorf("expected %s, got %s", other, resp.Info.UserID)
	}
}

func TestAdminListMcpKeys(t *testing.T) {
	ctx := adminCtx(testUserID)
	restore := isolateMcpDB(t)
	defer restore()
	mockUsersInfo(t)

	resp, err := AdminCreateMcpKey(ctx, &CreateMcpKeyParams{Label: "claude"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	list, err := AdminListMcpKeys(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(list.Keys))
	}
	if list.Keys[0].KeySuffix != resp.Key[len(resp.Key)-4:] {
		t.Errorf("listed suffix mismatch")
	}
}

func TestAdminRevokeMcpKey(t *testing.T) {
	ctx := adminCtx(testUserID)
	restore := isolateMcpDB(t)
	defer restore()

	resp, err := AdminCreateMcpKey(ctx, &CreateMcpKeyParams{Label: "x"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := AdminRevokeMcpKey(ctx, resp.Info.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	// The raw key must no longer resolve after revocation.
	if _, _, err := resolveMcpKey(ctx, resp.Key); err == nil {
		t.Error("expected revoked key to be rejected")
	}
}

func TestAdminCreateMcpKey_NonAdmin(t *testing.T) {
	ctx := encoreauth.WithContext(context.Background(), encoreauth.UID(testUserID), &myauth.AuthData{
		UserID: testUserID, Role: "user",
	})
	restore := isolateMcpDB(t)
	defer restore()

	_, err := AdminCreateMcpKey(ctx, &CreateMcpKeyParams{Label: "x"})
	if err == nil {
		t.Fatal("expected error for non-admin")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}
