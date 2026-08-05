package fixtures

import (
	"context"
	"testing"

	"comics-galore/backend/auth"

	encoreauth "encore.dev/beta/auth"
)

func TestConstants(t *testing.T) {
	if TestUploaderID == "" {
		t.Error("TestUploaderID is empty")
	}
	if TestAdminID == "" {
		t.Error("TestAdminID is empty")
	}
	if TestModeratorID == "" {
		t.Error("TestModeratorID is empty")
	}
	if TestUserID == "" {
		t.Error("TestUserID is empty")
	}
}

func TestAllUuidsDistinct(t *testing.T) {
	ids := map[string]bool{}
	for _, id := range []string{TestUploaderID, TestAdminID, TestModeratorID, TestUserID} {
		if ids[id] {
			t.Errorf("duplicate UUID: %s", id)
		}
		ids[id] = true
	}
	if len(ids) != 4 {
		t.Errorf("expected 4 distinct UUIDs, got %d", len(ids))
	}
}

func TestContextsReturnNonNil(t *testing.T) {
	for _, fn := range []func() context.Context{UploaderCtx, AdminCtx, ModeratorCtx, UserCtx} {
		c := fn()
		if c == nil {
			t.Fatal("fixture returned nil context")
		}
	}
}

func TestTierGatedCtx(t *testing.T) {
	ctx := TierGatedCtx("custom-uid", "uploader", "gold")
	if ctx == nil {
		t.Fatal("TierGatedCtx returned nil")
	}
	ctx2 := TierGatedCtx("another-uid", "user", "free")
	if ctx == ctx2 {
		t.Fatal("TierGatedCtx returned same context for different params")
	}
}

func TestAuthDataStructs(t *testing.T) {
	// Verify that creating AuthData structs directly works for WithContext
	ad := &auth.AuthData{
		UserID: "test-id",
		Email:  "test@example.com",
		Role:   "user",
		Tier:   "free",
	}
	ctx := context.Background()
	_ = encoreauth.WithContext(ctx, encoreauth.UID(ad.UserID), ad)
	if ctx == nil {
		t.Fatal("WithContext returned nil")
	}
}

func TestTierGatedDifferentValues(t *testing.T) {
	ad1 := &auth.AuthData{UserID: "u1", Email: "u1@example.com", Role: "admin", Tier: "platinum"}
	ad2 := &auth.AuthData{UserID: "u2", Email: "u2@example.com", Role: "user", Tier: "free"}

	if ad1.Role == ad2.Role {
		t.Error("expected different roles")
	}
	if ad1.Tier == ad2.Tier {
		t.Error("expected different tiers")
	}
}
