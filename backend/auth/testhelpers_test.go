package auth

import (
	"context"
	"testing"

	"encore.dev/et"
)

const testAdminID = "550e8400-e29b-41d4-a716-446655449999"

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

// createTestUser inserts a user row directly. Credentials now live in Logto,
// so tests seed the app-level record (role/tier) directly.
func createTestUser(ctx context.Context, t *testing.T, email, role, tier string) *User {
	t.Helper()
	var u User
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, role, tier, email_verified_at)
		VALUES ($1, $2, $3, now())
		RETURNING id, email, role, tier, COALESCE(username, ''), created_at
	`, email, role, tier).Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.Username, &u.CreatedAt)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return &u
}
