package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"comics-galore/backend/auth"
	"comics-galore/backend/comics"
	"comics-galore/backend/upload"

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

func assertCode(t *testing.T, err error, code errs.ErrCode) {
	t.Helper()
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != code {
		t.Errorf("expected code %v, got %v", code, e.Code)
	}
}

func TestCreateComic_NonUploaderDenied(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "user")

	insertKey(t, ctx, "user-key", testUserID)

	_, err := doCreateComic(ctx, "user-key", &CreateComicParams{Title: "Test"})
	if err == nil {
		t.Fatal("expected error for non-uploader")
	}
	assertCode(t, err, errs.PermissionDenied)
}

func TestCreateComic_Uploader(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "uploader")

	var called bool
	og := createComic
	createComic = func(ctx context.Context, p *comics.InternalCreateComicParams) (*comics.InternalCreateComicResponse, error) {
		called = true
		if p.ActorID != testUserID {
			t.Errorf("expected actor %s, got %s", testUserID, p.ActorID)
		}
		if p.Title != "My Comic" || p.SeriesTitle != "S1" || !p.Publish {
			t.Errorf("unexpected params: %+v", p)
		}
		return &comics.InternalCreateComicResponse{ID: "c1"}, nil
	}
	t.Cleanup(func() { createComic = og })

	insertKey(t, ctx, "up-key", testUserID)

	if _, err := doCreateComic(ctx, "up-key", &CreateComicParams{Title: "My Comic", Publish: true, SeriesTitle: "S1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected createComic to be called")
	}
}

func TestDeleteComment_NonModeratorDenied(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "user")

	insertKey(t, ctx, "user-key", testUserID)

	if err := doDeleteComment(ctx, "user-key", &DeleteCommentParams{CommentID: "c1"}); err == nil {
		t.Fatal("expected error for non-moderator")
	} else {
		assertCode(t, err, errs.PermissionDenied)
	}
}

func TestBanUser_ModeratorDenied(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "moderator")

	insertKey(t, ctx, "mod-key", testUserID)

	if err := doBanUser(ctx, "mod-key", &UserActionParams{UserID: "u2"}); err == nil {
		t.Fatal("expected error for non-admin")
	} else {
		assertCode(t, err, errs.PermissionDenied)
	}
}

func TestBanUser_Admin(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "admin")

	var called bool
	og := banUser
	banUser = func(ctx context.Context, p *auth.InternalUserActionParams) error {
		called = true
		if p.ActorID != testUserID || p.UserID != "u2" || p.Reason != "spam" {
			t.Errorf("unexpected params: %+v", p)
		}
		return nil
	}
	t.Cleanup(func() { banUser = og })

	insertKey(t, ctx, "admin-key", testUserID)

	if err := doBanUser(ctx, "admin-key", &UserActionParams{UserID: "u2", Reason: "spam"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected banUser to be called")
	}
}

func TestPresignUpload_NonUploaderDenied(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "user")

	insertKey(t, ctx, "user-key", testUserID)

	if _, err := doPresignUpload(ctx, "user-key", &PresignUploadParams{Kind: "cover"}); err == nil {
		t.Fatal("expected error for non-uploader")
	} else {
		assertCode(t, err, errs.PermissionDenied)
	}
}

func TestPresignUpload_Uploader(t *testing.T) {
	ctx := context.Background()
	restore := isolateMcpDB(t)
	defer restore()
	mockRole(t, "uploader")

	var called bool
	og := presignUpload
	presignUpload = func(ctx context.Context, p *upload.InternalPresignUploadParams) (*upload.InternalPresignUploadResponse, error) {
		called = true
		if p.ActorID != testUserID || p.Kind != "cover" {
			t.Errorf("unexpected params: %+v", p)
		}
		return &upload.InternalPresignUploadResponse{Key: "covers/x.jpg", UploadURL: "https://example.com"}, nil
	}
	t.Cleanup(func() { presignUpload = og })

	insertKey(t, ctx, "up-key", testUserID)

	if _, err := doPresignUpload(ctx, "up-key", &PresignUploadParams{Kind: "cover", Filename: "cover.jpg"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected presignUpload to be called")
	}
}
