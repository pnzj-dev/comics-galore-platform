package auth

import (
	"context"
	"errors"
	"testing"

	"encore.dev/beta/errs"
	"encore.dev/et"
)

const testBootstrapSecret = "test-bootstrap-secret"

// isolateAuthDB points the package-level db at a fresh, isolated test database
// for the current test and returns a function that restores the previous value.
func isolateAuthDB(t *testing.T) (restore func()) {
	t.Helper()
	isolated, err := et.NewTestDatabase(context.Background(), "authdb")
	if err != nil {
		t.Fatalf("new test database: %v", err)
	}
	original := db
	db = isolated
	return func() { db = original }
}

func TestBootstrapAdmin_Disabled(t *testing.T) {
	ctx := context.Background()
	secrets.BootstrapSecret = ""

	_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token:    "whatever",
		Email:    "admin@example.com",
		Password: "password123",
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
		Token:    "wrong-token",
		Email:    "admin@example.com",
		Password: "password123",
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

func TestBootstrapAdmin_Validation(t *testing.T) {
	ctx := context.Background()
	secrets.BootstrapSecret = testBootstrapSecret

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"empty email", "", "password123"},
		{"short password", "admin@example.com", "short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
				Token:    testBootstrapSecret,
				Email:    tt.email,
				Password: tt.password,
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
		})
	}
}

func TestBootstrapAdmin_CreatesAdmin(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.BootstrapSecret = testBootstrapSecret

	resp, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token:    testBootstrapSecret,
		Email:    "Admin@Example.com",
		Password: "password123",
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
		Token:    testBootstrapSecret,
		Email:    "admin@example.com",
		Password: "password123",
	}); err != nil {
		t.Fatalf("first bootstrap should succeed: %v", err)
	}

	_, err := BootstrapAdmin(ctx, &BootstrapAdminParams{
		Token:    testBootstrapSecret,
		Email:    "admin2@example.com",
		Password: "password123",
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
