// Package fixtures provides shared test fixtures and auth context helpers
// for Encore service tests.
package fixtures

import (
	"context"

	"comics-galore/backend/auth"

	encoreauth "encore.dev/beta/auth"
)

// Fixed UUIDs for test users (consistent across all test files).
const (
	TestUploaderID = "550e8400-e29b-41d4-a716-446655440001"
	TestAdminID    = "550e8400-e29b-41d4-a716-446655440002"
	TestModeratorID = "550e8400-e29b-41d4-a716-446655440003"
	TestUserID     = "550e8400-e29b-41d4-a716-446655440004"
)

// Auth context helpers.
func UploaderCtx() context.Context {
	ctx := context.Background()
	return encoreauth.WithContext(ctx, encoreauth.UID(TestUploaderID), &auth.AuthData{
		UserID: TestUploaderID,
		Email:  "uploader@example.com",
		Role:   "uploader",
		Tier:   "free",
	})
}

func AdminCtx() context.Context {
	ctx := context.Background()
	return encoreauth.WithContext(ctx, encoreauth.UID(TestAdminID), &auth.AuthData{
		UserID: TestAdminID,
		Email:  "admin@example.com",
		Role:   "admin",
		Tier:   "free",
	})
}

func ModeratorCtx() context.Context {
	ctx := context.Background()
	return encoreauth.WithContext(ctx, encoreauth.UID(TestModeratorID), &auth.AuthData{
		UserID: TestModeratorID,
		Email:  "mod@example.com",
		Role:   "moderator",
		Tier:   "free",
	})
}

func UserCtx() context.Context {
	ctx := context.Background()
	return encoreauth.WithContext(ctx, encoreauth.UID(TestUserID), &auth.AuthData{
		UserID: TestUserID,
		Email:  "user@example.com",
		Role:   "user",
		Tier:   "free",
	})
}

// TierGatedCtx creates an auth context with a specific tier.
func TierGatedCtx(uid, role, tier string) context.Context {
	ctx := context.Background()
	return encoreauth.WithContext(ctx, encoreauth.UID(uid), &auth.AuthData{
		UserID: uid,
		Email:  uid + "@example.com",
		Role:   role,
		Tier:   tier,
	})
}
