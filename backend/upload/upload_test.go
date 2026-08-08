package upload

import (
	"context"
	"testing"

	"comics-galore/backend/fixtures"

	"encore.dev/beta/errs"
	"encore.dev/et"
)

func TestCreateSession(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	ctx := fixtures.UploaderCtx()

	s, err := CreateSession(ctx, &CreateSessionParams{Mode: "manual"})
	if err != nil {
		t.Fatalf("create session error: %v", err)
	}
	if s.Status != "active" {
		t.Errorf("expected status active, got %s", s.Status)
	}
	if s.Mode != "manual" {
		t.Errorf("expected mode manual, got %s", s.Mode)
	}
	if s.S3Prefix == "" {
		t.Error("expected non-empty s3_prefix")
	}
	if len(s.Parts) != 0 {
		t.Errorf("expected empty parts, got %d", len(s.Parts))
	}
}

func TestCreateSessionDenied(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	ctx := fixtures.UserCtx()

	_, err := CreateSession(ctx, &CreateSessionParams{Mode: "manual"})
	if err == nil {
		t.Fatal("expected permission denied for non-uploader")
	}
}

func TestCreateSessionDefaultMode(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	ctx := fixtures.UploaderCtx()

	s, err := CreateSession(ctx, &CreateSessionParams{})
	if err != nil {
		t.Fatalf("create session error: %v", err)
	}
	if s.Mode != "manual" {
		t.Errorf("expected default mode manual, got %s", s.Mode)
	}
}

func TestGetSession(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	uploaderCtx := fixtures.UploaderCtx()

	s, _ := CreateSession(uploaderCtx, &CreateSessionParams{Mode: "manual"})

	got, err := GetSession(uploaderCtx, s.ID)
	if err != nil {
		t.Fatalf("get session error: %v", err)
	}
	if got.ID != s.ID {
		t.Errorf("expected ID %s, got %s", s.ID, got.ID)
	}

	// Other user should not be able to get it
	userCtx := fixtures.UserCtx()
	_, err = GetSession(userCtx, s.ID)
	if err == nil {
		t.Fatal("expected permission denied for other user")
	}

	// Admin can see any session
	adminCtx := fixtures.AdminCtx()
	_, err = GetSession(adminCtx, s.ID)
	if err != nil {
		t.Errorf("admin should be able to see any session: %v", err)
	}
}

func TestAbortSession(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	ctx := fixtures.UploaderCtx()

	s, _ := CreateSession(ctx, &CreateSessionParams{Mode: "manual"})

	err := AbortSession(ctx, s.ID)
	if err != nil {
		t.Fatalf("abort error: %v", err)
	}

	got, err := GetSession(ctx, s.ID)
	if err != nil {
		t.Fatalf("get session after abort error: %v", err)
	}
	if got.Status != "failed" {
		t.Errorf("expected status failed after abort, got %s", got.Status)
	}
}

func TestConfirmPart(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	ctx := fixtures.UploaderCtx()

	s, _ := CreateSession(ctx, &CreateSessionParams{Mode: "manual"})

	s, err := ConfirmPart(ctx, s.ID, &ConfirmPartParams{
		Number: 1,
		Key:    "test-key-1",
		Size:   12345,
		ETag:   "abc123",
	})
	if err != nil {
		t.Fatalf("confirm part error: %v", err)
	}
	if len(s.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(s.Parts))
	}
	if s.Parts[0].Key != "test-key-1" {
		t.Errorf("expected key test-key-1, got %s", s.Parts[0].Key)
	}
	if s.Parts[0].Size != 12345 {
		t.Errorf("expected size 12345, got %d", s.Parts[0].Size)
	}
}

var _ = errs.NotFound

func TestListActiveSessions(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	ctx := fixtures.UploaderCtx()

	s, _ := CreateSession(ctx, &CreateSessionParams{Mode: "manual"})

	resp, err := ListActiveSessions(ctx)
	if err != nil {
		t.Fatalf("list active sessions error: %v", err)
	}

	found := false
	for _, sess := range resp.Sessions {
		if sess.ID == s.ID {
			found = true
			if sess.Status != "active" {
				t.Errorf("expected status active, got %s", sess.Status)
			}
			break
		}
	}
	if !found {
		t.Error("created session should appear in active sessions list")
	}
}

func TestListActiveSessions_FiltersByUser(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "uploaddb")
	upCtx := fixtures.UploaderCtx()
	userCtx := fixtures.UserCtx()

	s, _ := CreateSession(upCtx, &CreateSessionParams{Mode: "manual"})

	usrResp, err := ListActiveSessions(userCtx)
	if err != nil {
		t.Fatalf("regular user should be able to list their sessions: %v", err)
	}

	found := false
	for _, sess := range usrResp.Sessions {
		if sess.ID == s.ID {
			found = true
			break
		}
	}
	if found {
		t.Error("regular user should not see uploader's sessions")
	}
}
