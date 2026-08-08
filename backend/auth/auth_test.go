package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
)

func TestRegister_Valid(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token, got empty")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.Role != "user" {
		t.Errorf("expected role user, got %s", resp.User.Role)
	}
	if resp.User.Tier != "free" {
		t.Errorf("expected tier free, got %s", resp.User.Tier)
	}
	if resp.User.ID == "" {
		t.Error("expected user ID, got empty")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	_, err := Register(ctx, &RegisterParams{
		Email:    "dup@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("first register should succeed: %v", err)
	}

	_, err = Register(ctx, &RegisterParams{
		Email:    "dup@example.com",
		Password: "different123",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.AlreadyExists {
		t.Errorf("expected AlreadyExists, got %v", e.Code)
	}
}

func TestRegister_EmptyFields(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"empty email and password", "", ""},
		{"empty email", "", "password123"},
		{"empty password", "test@example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Register(ctx, &RegisterParams{
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

func TestRegister_ShortPassword(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	_, err := Register(ctx, &RegisterParams{
		Email:    "short@example.com",
		Password: "1234567",
	})
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", e.Code)
	}
}

func TestLogin_Valid(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	_, err := Register(ctx, &RegisterParams{
		Email:    "login@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	resp, err := Login(ctx, &LoginParams{
		Email:    "login@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token, got empty")
	}
	if resp.User.Email != "login@example.com" {
		t.Errorf("expected email login@example.com, got %s", resp.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	_, err := Register(ctx, &RegisterParams{
		Email:    "wrong@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	_, err = Login(ctx, &LoginParams{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", e.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	_, err := Login(ctx, &LoginParams{
		Email:    "nobody@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", e.Code)
	}
}

func TestMe(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "me@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	authCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   resp.User.Role,
		Tier:   resp.User.Tier,
	})

	user, err := Me(authCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "me@example.com" {
		t.Errorf("expected email me@example.com, got %s", user.Email)
	}
	if user.ID != resp.User.ID {
		t.Errorf("expected ID %s, got %s", resp.User.ID, user.ID)
	}
}

func TestRenewToken(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "renew@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	authCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   resp.User.Role,
		Tier:   resp.User.Tier,
	})

	renewed, err := RenewToken(authCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if renewed.Token == "" {
		t.Error("expected new token, got empty")
	}
	if renewed.User.Email != "renew@example.com" {
		t.Errorf("expected email renew@example.com, got %s", renewed.User.Email)
	}
}

func TestAdminBanUser(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "victim@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	adminCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "admin",
		Tier:   resp.User.Tier,
	})

	err = AdminBanUser(adminCtx, resp.User.ID, &BanUserParams{Reason: "spam"})
	if err != nil {
		t.Fatalf("ban error: %v", err)
	}

	var bannedAt *time.Time
	err = db.QueryRow(ctx, `SELECT banned_at FROM users WHERE id = $1`, resp.User.ID).Scan(&bannedAt)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if bannedAt == nil {
		t.Error("expected banned_at to be set")
	}
}

func TestAdminUnbanUser(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "pardoned@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	adminCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "admin",
		Tier:   resp.User.Tier,
	})

	err = AdminBanUser(adminCtx, resp.User.ID, &BanUserParams{Reason: "spam"})
	if err != nil {
		t.Fatalf("ban error: %v", err)
	}

	err = AdminUnbanUser(adminCtx, resp.User.ID)
	if err != nil {
		t.Fatalf("unban error: %v", err)
	}

	var bannedAt *time.Time
	err = db.QueryRow(ctx, `SELECT banned_at FROM users WHERE id = $1`, resp.User.ID).Scan(&bannedAt)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if bannedAt != nil {
		t.Error("expected banned_at to be NULL after unban")
	}
}

func TestAdminBanUser_NonAdmin(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	userCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "user",
		Tier:   resp.User.Tier,
	})

	err = AdminBanUser(userCtx, "some-other-id", &BanUserParams{Reason: "test"})
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
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "suspendme@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	adminCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "admin",
		Tier:   resp.User.Tier,
	})

	err = AdminSuspendUser(adminCtx, resp.User.ID, &BanUserParams{Reason: "violation"})
	if err != nil {
		t.Fatalf("suspend error: %v", err)
	}

	var suspendedAt *time.Time
	err = db.QueryRow(ctx, `SELECT suspended_at FROM users WHERE id = $1`, resp.User.ID).Scan(&suspendedAt)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if suspendedAt == nil {
		t.Error("expected suspended_at to be set")
	}
}

func TestAdminUnsuspendUser(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "unsuspendme@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	adminCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "admin",
		Tier:   resp.User.Tier,
	})

	err = AdminSuspendUser(adminCtx, resp.User.ID, &BanUserParams{Reason: "violation"})
	if err != nil {
		t.Fatalf("suspend error: %v", err)
	}

	err = AdminUnsuspendUser(adminCtx, resp.User.ID)
	if err != nil {
		t.Fatalf("unsuspend error: %v", err)
	}

	var suspendedAt *time.Time
	err = db.QueryRow(ctx, `SELECT suspended_at FROM users WHERE id = $1`, resp.User.ID).Scan(&suspendedAt)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if suspendedAt != nil {
		t.Error("expected suspended_at to be NULL after unsuspend")
	}
}

func TestBannedUser_CannotLogin(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "banned@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	adminCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "admin",
		Tier:   resp.User.Tier,
	})

	err = AdminBanUser(adminCtx, resp.User.ID, &BanUserParams{Reason: "spam"})
	if err != nil {
		t.Fatalf("ban error: %v", err)
	}

	_, err = Login(ctx, &LoginParams{
		Email:    "banned@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for banned user, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestSuspendedUser_CannotLogin(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "authdb")

	resp, err := Register(ctx, &RegisterParams{
		Email:    "suspended@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	adminCtx := auth.WithContext(ctx, auth.UID(resp.User.ID), &AuthData{
		UserID: resp.User.ID,
		Email:  resp.User.Email,
		Role:   "admin",
		Tier:   resp.User.Tier,
	})

	err = AdminSuspendUser(adminCtx, resp.User.ID, &BanUserParams{Reason: "violation"})
	if err != nil {
		t.Fatalf("suspend error: %v", err)
	}

	_, err = Login(ctx, &LoginParams{
		Email:    "suspended@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for suspended user, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}
