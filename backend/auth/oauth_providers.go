package auth

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"encore.dev/beta/errs"

	"golang.org/x/oauth2"
)

// newConfig returns a base oauth2 config for a provider with client id/secret
// and endpoints. RedirectURL is filled in per-request.
func newConfig(clientID, clientSecret, authURL, tokenURL string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{AuthURL: authURL, TokenURL: tokenURL},
		Scopes:       scopes,
	}
}

// ----- Google -----

type googleOAuth struct{}

func (googleOAuth) name() string { return "google" }

func (googleOAuth) authURL(state, redirectURL, challenge string) string {
	cfg := newConfig(secrets.GoogleClientID, secrets.GoogleClientSecret,
		"https://accounts.google.com/o/oauth2/v2/auth",
		"https://oauth2.googleapis.com/token",
		[]string{"openid", "email", "profile"})
	cfg.RedirectURL = redirectURL
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(challenge))
}

func (googleOAuth) exchange(ctx context.Context, code, verifier, redirectURL string) (*oauthIdentity, error) {
	cfg := newConfig(secrets.GoogleClientID, secrets.GoogleClientSecret,
		"https://accounts.google.com/o/oauth2/v2/auth",
		"https://oauth2.googleapis.com/token",
		[]string{"openid", "email", "profile"})
	cfg.RedirectURL = redirectURL
	tok, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}
	body, err := fetchJSON(ctx, "https://openidconnect.googleapis.com/v1/userinfo", tok.AccessToken)
	if err != nil {
		return nil, err
	}
	return parseUserInfo(body, "sub", "email", "email_verified")
}

// ----- Facebook -----

type facebookOAuth struct{}

func (facebookOAuth) name() string { return "facebook" }

func (facebookOAuth) authURL(state, redirectURL, challenge string) string {
	cfg := newConfig(secrets.FacebookClientID, secrets.FacebookClientSecret,
		"https://www.facebook.com/v19.0/dialog/oauth",
		"https://graph.facebook.com/v19.0/oauth/access_token",
		[]string{"email", "public_profile"})
	cfg.RedirectURL = redirectURL
	return cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(challenge))
}

func (facebookOAuth) exchange(ctx context.Context, code, verifier, redirectURL string) (*oauthIdentity, error) {
	cfg := newConfig(secrets.FacebookClientID, secrets.FacebookClientSecret,
		"https://www.facebook.com/v19.0/dialog/oauth",
		"https://graph.facebook.com/v19.0/oauth/access_token",
		[]string{"email", "public_profile"})
	cfg.RedirectURL = redirectURL
	tok, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}
	body, err := fetchJSON(ctx, "https://graph.facebook.com/me?fields=id,email,name", tok.AccessToken)
	if err != nil {
		return nil, err
	}
	id, err := parseUserInfo(body, "id", "email", "email")
	if err != nil {
		return nil, err
	}
	id.Verified = id.Email != ""
	return id, nil
}

// ----- Twitter / X -----

type twitterOAuth struct{}

func (twitterOAuth) name() string { return "twitter" }

func (twitterOAuth) authURL(state, redirectURL, challenge string) string {
	cfg := newConfig(secrets.TwitterClientID, secrets.TwitterClientSecret,
		"https://twitter.com/i/oauth2/authorize",
		"https://api.twitter.com/2/oauth2/token",
		[]string{"users.read", "tweet.read"})
	cfg.RedirectURL = redirectURL
	return cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(challenge))
}

func (twitterOAuth) exchange(ctx context.Context, code, verifier, redirectURL string) (*oauthIdentity, error) {
	cfg := newConfig(secrets.TwitterClientID, secrets.TwitterClientSecret,
		"https://twitter.com/i/oauth2/authorize",
		"https://api.twitter.com/2/oauth2/token",
		[]string{"users.read", "tweet.read"})
	cfg.RedirectURL = redirectURL
	tok, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}
	body, err := fetchJSON(ctx, "https://api.twitter.com/2/users/me", tok.AccessToken)
	if err != nil {
		return nil, err
	}
	var m struct {
		Data struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	if m.Data.ID == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "provider returned no account id"}
	}
	name := m.Data.Name
	if name == "" {
		name = m.Data.Username
	}
	return &oauthIdentity{ProviderID: m.Data.ID, Name: name, Verified: false}, nil
}

// ----- Apple -----

type appleOAuth struct{}

func (appleOAuth) name() string { return "apple" }

func (appleOAuth) authURL(state, redirectURL, challenge string) string {
	cfg := &oauth2.Config{
		ClientID: secrets.AppleClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
		Scopes:      []string{"name", "email"},
		RedirectURL: redirectURL,
	}
	// Apple does not support PKCE; it uses a client-secret JWT instead.
	return cfg.AuthCodeURL(state, oauth2.SetAuthURLParam("response_mode", "form_post"))
}

func (appleOAuth) clientSecret() (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": secrets.AppleTeamID,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": secrets.AppleClientID,
	}
	block, _ := pem.Decode([]byte(secrets.ApplePrivateKey))
	if block == nil {
		return "", &errs.Error{Code: errs.Internal, Message: "apple private key not configured"}
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = secrets.AppleKeyID
	return token.SignedString(key)
}

func (appleOAuth) exchange(ctx context.Context, code, verifier, redirectURL string) (*oauthIdentity, error) {
	secret, err := (appleOAuth{}).clientSecret()
	if err != nil {
		return nil, err
	}
	resp, err := httpPostForm(ctx, "https://appleid.apple.com/auth/token", map[string]string{
		"client_id":     secrets.AppleClientID,
		"client_secret": secret,
		"code":          code,
		"grant_type":    "authorization_code",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "apple token exchange failed"}
	}

	var tok struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(tok.IDToken, claims); err != nil {
		return nil, err
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "apple returned no account id"}
	}
	email, _ := claims["email"].(string)
	emailVerified, _ := claims["email_verified"].(bool)

	return &oauthIdentity{ProviderID: sub, Email: email, Verified: emailVerified}, nil
}
