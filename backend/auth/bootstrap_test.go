package auth

import (
	"context"
	"errors"
	"testing"

	"encore.dev/beta/errs"
)

const testBootstrapSecret = "test-bootstrap-secret"

func TestBootstrapAdmin_Disabled(t *testing.T) {
	ctx := context.Background()
	secrets.BootstrapSecret = ""

	_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token: "whatever",
		Email: "admin@example.com",
	})
	if err == nil {
		t.Fatal("expected error when bootstrap is disabled, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.Unavailable {
		t.Errorf("expected Unavailable, got %v", e.Code)
	}
}

func TestBootstrapAdmin_WrongToken(t *testing.T) {
	ctx := context.Background()
	secrets.BootstrapSecret = testBootstrapSecret

	_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token: "wrong-token",
		Email: "admin@example.com",
	})
	if err == nil {
		t.Fatal("expected error for wrong token, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestBootstrapAdmin_MissingEmail(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.BootstrapSecret = testBootstrapSecret

	_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token: testBootstrapSecret,
		Email: "",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", e.Code)
	}
}

func TestBootstrapAdmin_CreatesAdmin(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.BootstrapSecret = testBootstrapSecret

	resp, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token: testBootstrapSecret,
		Email: "Admin@Example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Email != "admin@example.com" {
		t.Errorf("expected normalized email admin@example.com, got %s", resp.User.Email)
	}
	if resp.User.Role != "admin" {
		t.Errorf("expected role admin, got %s", resp.User.Role)
	}
	if resp.User.Tier != "platinum" {
		t.Errorf("expected tier platinum, got %s", resp.User.Tier)
	}
	if resp.User.ID == "" {
		t.Error("expected user ID, got empty")
	}
}

func TestBootstrapAdmin_OneTime(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.BootstrapSecret = testBootstrapSecret

	if _, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token: testBootstrapSecret,
		Email: "admin@example.com",
	}); err != nil {
		t.Fatalf("first bootstrap should succeed: %v", err)
	}

	_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token: testBootstrapSecret,
		Email: "admin2@example.com",
	})
	if err == nil {
		t.Fatal("expected error on second bootstrap, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}
