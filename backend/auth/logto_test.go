package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"encore.dev/beta/errs"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const testIssuer = "https://test.logto.app/oidc"

type testKeyPair struct {
	set     jwk.Set
	signing jwk.Key
}

func newTestKeyPair(t *testing.T) testKeyPair {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pub, err := jwk.Import(&priv.PublicKey)
	if err != nil {
		t.Fatalf("fromraw pub: %v", err)
	}
	_ = pub.Set(jwk.AlgorithmKey, jwa.RS256())
	_ = pub.Set(jwk.KeyIDKey, "test-kid")

	signing, err := jwk.Import(priv)
	if err != nil {
		t.Fatalf("fromraw priv: %v", err)
	}
	_ = signing.Set(jwk.AlgorithmKey, jwa.RS256())
	_ = signing.Set(jwk.KeyIDKey, "test-kid")

	set := jwk.NewSet()
	_ = set.AddKey(pub)
	return testKeyPair{set: set, signing: signing}
}

func (kp testKeyPair) mint(t *testing.T, sub, email string) string {
	t.Helper()
	tok := jwt.New()
	_ = tok.Set(jwt.SubjectKey, sub)
	_ = tok.Set(jwt.IssuerKey, testIssuer)
	if email != "" {
		_ = tok.Set("email", email)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), kp.signing))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return string(signed)
}

func withJWKS(t *testing.T, set jwk.Set) (restore func()) {
	t.Helper()
	original := jwksProvider
	jwksProvider = func(ctx context.Context) (jwk.Set, error) { return set, nil }
	return func() { jwksProvider = original }
}

func TestAuthHandler_ProvisionsNewUser(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.LogtoIssuer = testIssuer
	kp := newTestKeyPair(t)
	restoreJWKS := withJWKS(t, kp.set)
	defer restoreJWKS()

	token := kp.mint(t, "logto-sub-1", "new@example.com")
	uid, ad, err := AuthHandler(ctx, &AuthParams{Authorization: "Bearer " + token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid == "" {
		t.Error("expected uid")
	}
	if ad.UserID == "" {
		t.Error("expected user id")
	}
	if ad.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", ad.Email)
	}
	if ad.Role != "user" {
		t.Errorf("expected role user, got %s", ad.Role)
	}
	if ad.Tier != "free" {
		t.Errorf("expected tier free, got %s", ad.Tier)
	}
}

func TestAuthHandler_BannedUser(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.LogtoIssuer = testIssuer
	kp := newTestKeyPair(t)
	restoreJWKS := withJWKS(t, kp.set)
	defer restoreJWKS()

	u := createTestUser(ctx, t, "banned@example.com", "user", "free")
	if _, err := db.Exec(ctx, `UPDATE users SET banned_at = now() WHERE id = $1`, u.ID); err != nil {
		t.Fatalf("ban: %v", err)
	}
	// Link the Logto identity to the existing row so resolution finds it.
	if _, err := db.Exec(ctx, `UPDATE users SET logto_id = $1 WHERE id = $2`, "logto-sub-banned", u.ID); err != nil {
		t.Fatalf("link: %v", err)
	}

	token := kp.mint(t, "logto-sub-banned", "banned@example.com")
	_, _, err := AuthHandler(ctx, &AuthParams{Authorization: "Bearer " + token})
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

func TestAuthHandler_InvalidToken(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.LogtoIssuer = testIssuer
	kp := newTestKeyPair(t)
	restoreJWKS := withJWKS(t, kp.set)
	defer restoreJWKS()

	_, _, err := AuthHandler(ctx, &AuthParams{Authorization: "Bearer not-a-jwt"})
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", e.Code)
	}
}

func TestAuthHandler_LinksExistingByEmail(t *testing.T) {
	ctx := context.Background()
	restore := isolateAuthDB(t)
	defer restore()
	secrets.LogtoIssuer = testIssuer
	kp := newTestKeyPair(t)
	restoreJWKS := withJWKS(t, kp.set)
	defer restoreJWKS()

	u := createTestUser(ctx, t, "existing@example.com", "uploader", "gold")

	token := kp.mint(t, "logto-sub-link", "existing@example.com")
	uid, ad, err := AuthHandler(ctx, &AuthParams{Authorization: "Bearer " + token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(uid) != u.ID {
		t.Errorf("expected uid %s, got %s", u.ID, string(uid))
	}
	if ad.Role != "uploader" {
		t.Errorf("expected existing role uploader, got %s", ad.Role)
	}

	var logtoID string
	if err := db.QueryRow(ctx, `SELECT COALESCE(logto_id, '') FROM users WHERE id = $1`, u.ID).Scan(&logtoID); err != nil {
		t.Fatalf("query logto_id: %v", err)
	}
	if logtoID != "logto-sub-link" {
		t.Errorf("expected logto_id linked, got %q", logtoID)
	}
}
