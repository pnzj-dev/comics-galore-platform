package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"comics-galore/backend/nowpayments"
	"comics-galore/backend/turnstile"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var secrets struct {
	JWTSecret           string
	NowPaymentsAPIKey   string
	NowPaymentsIPNKey   string
	NowPaymentsEmail    string
	NowPaymentsPassword string

	// WebAuthn / Passkey
	WebAuthnRPID    string
	WebAuthnOrigins string // comma-separated list of allowed origins

	// OAuth providers
	FrontendURL           string
	GoogleClientID        string
	GoogleClientSecret    string
	FacebookClientID      string
	FacebookClientSecret  string
	TwitterClientID       string
	TwitterClientSecret   string
	AppleClientID         string
	AppleTeamID           string
	AppleKeyID            string
	ApplePrivateKey       string

	// Email (Resend)
	ResendAPIKey string

	// BootstrapSecret gates the one-time first-admin provisioning endpoint
	// (/auth/bootstrap). Empty disables bootstrap entirely.
	BootstrapSecret string
}

var npProvider *nowpayments.Provider

func init() {
	npProvider = nowpayments.NewProvider(secrets.NowPaymentsAPIKey, secrets.NowPaymentsIPNKey,
		secrets.NowPaymentsEmail, secrets.NowPaymentsPassword)
}

var db = sqldb.NewDatabase("authdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var AvatarBucket = objects.NewBucket("avatars", objects.BucketConfig{})

type AuthParams struct {
	Authorization string `header:"Authorization"`
}

type AuthData struct {
	UserID         string
	Email          string
	Role           string
	Tier           string
	ImpersonatedBy string
}

//encore:authhandler
func AuthHandler(ctx context.Context, p *AuthParams) (auth.UID, *AuthData, error) {
	if p.Authorization == "" {
		return "", nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "missing authorization header",
		}
	}

	token := strings.TrimPrefix(p.Authorization, "Bearer ")
	if token == p.Authorization {
		return "", nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "invalid authorization format",
		}
	}

	sess, err := validateSession(ctx, token)
	if err != nil {
		return "", nil, err
	}
	touchSession(ctx, token)

	var email sql.NullString
	var role, tier string
	if err := db.QueryRow(ctx, `SELECT email, role, tier FROM users WHERE id = $1`, sess.UserID).Scan(&email, &role, &tier); err != nil {
		if isNoRows(err) {
			return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid session"}
		}
		return "", nil, err
	}

	var maintenance bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'maintenance_mode')::boolean, false) FROM app_settings WHERE key = 'defaults'`).Scan(&maintenance); e == nil && maintenance && role != "admin" {
		return "", nil, &errs.Error{Code: errs.Unavailable, Message: "the platform is under maintenance, please try again later"}
	}

	var requireVerify bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'require_email_verify')::boolean, false) FROM app_settings WHERE key = 'defaults'`).Scan(&requireVerify); e == nil && requireVerify {
		var emailVerified sql.NullTime
		if e2 := db.QueryRow(ctx, `SELECT email_verified_at FROM users WHERE id = $1`, sess.UserID).Scan(&emailVerified); e2 == nil && !emailVerified.Valid {
			return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "email verification required"}
		}
	}

	var bannedAt, suspendedAt sql.NullTime
	if err := db.QueryRow(ctx, `SELECT banned_at, suspended_at FROM users WHERE id = $1`, sess.UserID).Scan(&bannedAt, &suspendedAt); err == nil {
		if bannedAt.Valid {
			return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "account is banned"}
		}
		if suspendedAt.Valid {
			return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "account is suspended"}
		}
	}

	return auth.UID(sess.UserID), &AuthData{
		UserID:         sess.UserID,
		Email:          email.String,
		Role:           role,
		Tier:           tier,
		ImpersonatedBy: sess.ImpersonatedBy,
	}, nil
}

type RegisterParams struct {
	Email          string `json:"email"`
	Password       string `json:"password" encore:"sensitive"`
	Username       string `json:"username"`
	TurnstileToken string `json:"turnstile_token"`
}

// usernameRe validates a public handle's characters/structure: starts and ends
// with a lowercase alphanumeric, with single `_`/`-` only between alphanumerics
// (no leading/trailing/consecutive). Length is checked separately (3-20).
var usernameRe = regexp.MustCompile(`^[a-z0-9](?:[_-]?[a-z0-9])*$`)

func validUsername(username string) bool {
	return len(username) >= 3 && len(username) <= 20 && usernameRe.MatchString(username)
}

type AuthResponse struct {
	Token        string `json:"token"`
	User         User   `json:"user"`
	RequiresTOTP bool   `json:"requires_totp,omitempty"`
	MFAToken     string `json:"mfa_token,omitempty"`
}

//encore:api public method=POST path=/auth/register
func Register(ctx context.Context, p *RegisterParams) (*AuthResponse, error) {
	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "register"}); err != nil {
		return nil, err
	}

	if p.Email == "" || p.Password == "" {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "email and password are required",
		}
	}

	if len(p.Password) < 8 {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "password must be at least 8 characters",
		}
	}

	username := strings.ToLower(strings.TrimSpace(p.Username))
	if username == "" || !validUsername(username) {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "username must be 3-20 characters, lowercase letters, numbers, and single - or _ in between",
		}
	}

	var usernameTaken bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&usernameTaken); err == nil && usernameTaken {
		return nil, &errs.Error{Code: errs.AlreadyExists, Message: "this username is already taken"}
	}

	var regOpen bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'registrations_open')::boolean, true) FROM app_settings WHERE key = 'defaults'`).Scan(&regOpen); e == nil && !regOpen {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "registration is currently closed"}
	}

	var maintenance bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'maintenance_mode')::boolean, false) FROM app_settings WHERE key = 'defaults'`).Scan(&maintenance); e == nil && maintenance {
		return nil, &errs.Error{Code: errs.Unavailable, Message: "the platform is under maintenance, please try again later"}
	}

	existing, err := getUserByEmail(ctx, p.Email)
	if err != nil && !isNoRows(err) {
		return nil, err
	}
	if existing != nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "a user with this email already exists",
		}
	}

	hash, err := hashPassword(p.Password)
	if err != nil {
		return nil, err
	}

	role := "user"

	var user User
	err = db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, username, terms_accepted_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, email, role, tier, COALESCE(username, ''), created_at
	`, p.Email, hash, role, username).Scan(&user.ID, &user.Email, &user.Role, &user.Tier, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	token, err := createSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	verifyToken := randomToken(32)
	db.Exec(ctx, `UPDATE users SET verify_token = $1, verify_token_expires_at = now() + interval '24 hours' WHERE id = $2`, verifyToken, user.ID)
	go sendVerificationEmail(user.Email, verifyToken)

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

type UsernameAvailableParams struct {
	Username string `query:"username"`
}

type UsernameAvailableResponse struct {
	Available bool   `json:"available"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message,omitempty"`
}

//encore:api public method=GET path=/auth/username-available
func UsernameAvailable(ctx context.Context, p *UsernameAvailableParams) (*UsernameAvailableResponse, error) {
	username := strings.ToLower(strings.TrimSpace(p.Username))
	if username == "" || !validUsername(username) {
		return &UsernameAvailableResponse{Available: false, Valid: false, Message: "username must be 3-20 characters, lowercase letters, numbers, and single - or _ in between"}, nil
	}

	var taken bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&taken); err != nil {
		return nil, err
	}
	if taken {
		return &UsernameAvailableResponse{Available: false, Valid: true, Message: "this username is already taken"}, nil
	}
	return &UsernameAvailableResponse{Available: true, Valid: true}, nil
}

type LoginParams struct {
	Email          string `json:"email"`
	Password       string `json:"password" encore:"sensitive"`
	TurnstileToken string `json:"turnstile_token"`
}

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, p *LoginParams) (*AuthResponse, error) {
	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "login"}); err != nil {
		return nil, err
	}

	if p.Email == "" || p.Password == "" {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "email and password are required",
		}
	}

	user, err := getUserByEmail(ctx, p.Email)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "invalid email or password",
			}
		}
		return nil, err
	}

	if !user.PasswordHash.Valid || !checkPassword(user.PasswordHash.String, p.Password) {
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "invalid email or password",
		}
	}

	if user.BannedAt.Valid {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "account is banned",
		}
	}
	if user.SuspendedAt.Valid {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "account is suspended",
		}
	}

	_, err = db.Exec(ctx, `UPDATE users SET last_seen_at = now() WHERE id = $1`, user.ID)
	if err != nil {
		return nil, err
	}

	// Two-factor authentication: when TOTP is enabled, don't issue a session yet.
	// Return a short-lived MFA challenge that the client exchanges for a session
	// after the user submits their authenticator-app code.
	if _, enabled, err := userTOTPSecret(ctx, user.ID); err == nil && enabled {
		mfaToken, err := storeMFAChallenge(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		return &AuthResponse{RequiresTOTP: true, MFAToken: mfaToken}, nil
	}

	token, err := createSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: User{
			ID:        user.ID,
			Email:     user.Email.String,
			Role:      user.Role,
			Tier:      user.Tier,
			Username:  user.Username.String,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

//encore:api auth method=GET path=/auth/me
func Me(ctx context.Context) (*User, error) {
	data := auth.Data().(*AuthData)
	user, err := getUserByID(ctx, data.UserID)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "user not found",
			}
		}
		return nil, err
	}
	return user, nil
}

//encore:api auth method=GET path=/auth/renew
func RenewToken(ctx context.Context) (*AuthResponse, error) {
	data := auth.Data().(*AuthData)
	user, err := getUserByID(ctx, data.UserID)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "user not found",
			}
		}
		return nil, err
	}

	token, err := createSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// LogoutParams carries the session token being terminated. The frontend sends
// the same bearer token it uses for API calls.
type LogoutParams struct {
	Token string `json:"token"`
}

//encore:api public method=POST path=/auth/logout
func Logout(ctx context.Context, p *LogoutParams) error {
	if p.Token == "" {
		return nil
	}
	_ = deleteSession(ctx, p.Token)
	return nil
}

//encore:api auth method=POST path=/auth/logout-all
func LogoutAll(ctx context.Context) error {
	data := auth.Data().(*AuthData)
	return revokeAllSessions(ctx, data.UserID)
}

//encore:api auth method=GET path=/auth/sessions
func ListSessions(ctx context.Context) (*SessionsResponse, error) {
	data := auth.Data().(*AuthData)
	rows, err := db.Query(ctx, `
		SELECT id, created_at, last_seen_at, expires_at
		FROM sessions WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
	`, data.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out SessionsResponse
	for rows.Next() {
		var s SessionInfo
		if err := rows.Scan(&s.ID, &s.CreatedAt, &s.LastSeenAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		out.Sessions = append(out.Sessions, s)
	}
	return &out, rows.Err()
}

type SessionsResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

type SessionInfo struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// RevokeSessionParams revokes one session by id (logout of another device).
type RevokeSessionParams struct {
	SessionID string `json:"session_id"`
}

//encore:api auth method=POST path=/auth/sessions/revoke
func RevokeSession(ctx context.Context, p *RevokeSessionParams) error {
	data := auth.Data().(*AuthData)
	if p.SessionID == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "session_id is required"}
	}
	_, err := db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, p.SessionID, data.UserID)
	return err
}

// ----- Email Verification -----

type VerifyEmailParams struct {
	Token string `json:"token"`
}

//encore:api public method=POST path=/auth/verify-email
func VerifyEmail(ctx context.Context, p *VerifyEmailParams) error {
	if p.Token == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "token is required"}
	}

	var userID string
	err := db.QueryRow(ctx, `
		UPDATE users SET email_verified_at = now(), verify_token = NULL, verify_token_expires_at = NULL
		WHERE verify_token = $1 AND verify_token_expires_at > now() AND email_verified_at IS NULL
		RETURNING id
	`, p.Token).Scan(&userID)
	if err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.InvalidArgument, Message: "invalid or expired verification token"}
		}
		return err
	}

	// Eagerly provision the NowPayments customer (synchronous). Failure is
	// non-fatal: subscription creation retries lazily via EnsureSubPartnerID.
	if _, err := ensureSubPartnerID(ctx, userID); err != nil {
		log.Printf("[auth] ensure sub_partner_id for %s: %v", userID, err)
	}

	return nil
}

// nowpaymentsSubPartnerName builds a unique, non-email, ≤30-character name for
// a NowPayments sub-partner. NowPayments rejects emails and names longer than
// 30 characters for the `name` field of POST /sub-partner/balance.
func nowpaymentsSubPartnerName(userID string) string {
	sum := sha256.Sum256([]byte(userID))
	return "cg-" + hex.EncodeToString(sum[:8]) // 19 characters
}

// ensureSubPartnerID returns the user's NowPayments sub-partner id, creating
// the customer on NowPayments and saving it atomically when missing.
func ensureSubPartnerID(ctx context.Context, userID string) (string, error) {
	var existing string
	err := db.QueryRow(ctx, `SELECT COALESCE(sub_partner_id, '') FROM users WHERE id = $1`, userID).Scan(&existing)
	if err != nil {
		return "", err
	}
	if existing != "" {
		return existing, nil
	}

	subID, err := npProvider.CreateCustomer(ctx, nowpaymentsSubPartnerName(userID))
	if err != nil {
		return "", err
	}
	if subID == "" {
		return "", fmt.Errorf("nowpayments created sub-partner with empty id for user %s", userID)
	}

	// Atomic claim: only set if still empty, guarding against concurrent creation.
	res, err := db.Exec(ctx, `
		UPDATE users SET sub_partner_id = $1
		WHERE id = $2 AND sub_partner_id IS NULL
	`, subID, userID)
	if err != nil {
		return "", err
	}
	if res.RowsAffected() == 0 {
		// Lost the race; return whatever won.
		_ = db.QueryRow(ctx, `SELECT COALESCE(sub_partner_id, '') FROM users WHERE id = $1`, userID).Scan(&existing)
		return existing, nil
	}
	return subID, nil
}

type EnsureSubPartnerIDParams struct {
	UserID string `json:"user_id"`
}

type SubPartnerIDResponse struct {
	SubPartnerID string `json:"sub_partner_id"`
}

//encore:api private method=POST path=/auth/ensure-sub-partner-id
func EnsureSubPartnerID(ctx context.Context, p *EnsureSubPartnerIDParams) (*SubPartnerIDResponse, error) {
	subID, err := ensureSubPartnerID(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	return &SubPartnerIDResponse{SubPartnerID: subID}, nil
}

type SetUserTierParams struct {
	UserID string `json:"user_id"`
	Tier   string `json:"tier"`
}

//encore:api private method=POST path=/auth/set-user-tier
func SetUserTier(ctx context.Context, p *SetUserTierParams) error {
	_, err := db.Exec(ctx, `UPDATE users SET tier = $1 WHERE id = $2`, p.Tier, p.UserID)
	return err
}

type NotifyFollowersNewComicParams struct {
	UserIDs    []string `json:"user_ids"`
	ComicTitle string   `json:"comic_title"`
}

//encore:api private method=POST path=/auth/notify-followers-new-comic
func NotifyFollowersNewComic(ctx context.Context, p *NotifyFollowersNewComicParams) error {
	if len(p.UserIDs) == 0 {
		return nil
	}

	rows, err := db.Query(ctx, `
		SELECT u.email
		FROM users u
		LEFT JOIN notification_preferences np ON np.user_id = u.id
		WHERE u.id = ANY($1)
		  AND u.email_verified_at IS NOT NULL
		  AND COALESCE(np.email_new_from_following, true) = true
	`, p.UserIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if rows.Scan(&email) == nil && email != "" {
			go sendNewComicFromFollowingEmail(email, p.ComicTitle)
		}
	}
	return rows.Err()
}

type NotifySupportReplyParams struct {
	UserID   string `json:"user_id"`
	Subject  string `json:"subject"`
}

//encore:api private method=POST path=/auth/notify-support-reply
func NotifySupportReply(ctx context.Context, p *NotifySupportReplyParams) error {
	if p.UserID == "" {
		return nil
	}

	var email string
	err := db.QueryRow(ctx, `
		SELECT u.email
		FROM users u
		LEFT JOIN notification_preferences np ON np.user_id = u.id
		WHERE u.id = $1
		  AND u.email_verified_at IS NOT NULL
		  AND COALESCE(np.email_support_replies, true) = true
	`, p.UserID).Scan(&email)
	if err != nil || email == "" {
		return nil
	}

	go sendSupportReplyEmail(email, p.Subject)
	return nil
}

// AIModerationConfig is the moderation configuration exposed to the comics
// service (which owns the AI decision flow). The API key stays a secret in the
// comics service; this endpoint returns non-secret configuration only.
type AIModerationConfig struct {
	Enabled             bool    `json:"enabled"`
	Model               string  `json:"model"`
	Endpoint            string  `json:"endpoint"`
	Prompt              string  `json:"prompt"`
	AutoApproveThreshold float64 `json:"auto_approve_threshold"`
	AutoRejectThreshold  float64 `json:"auto_reject_threshold"`
}

//encore:api private method=GET path=/auth/ai-moderation-config
func GetAIModerationConfig(ctx context.Context) (*AIModerationConfig, error) {
	var raw []byte
	err := db.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = 'defaults'`).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return &AIModerationConfig{}, nil
	}

	var settings AppSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return &AIModerationConfig{}, nil
	}

	cfg := &AIModerationConfig{
		Enabled:              settings.AIModerationEnabled,
		Model:                settings.AIModel,
		Endpoint:             settings.AIEndpoint,
		Prompt:               settings.AIPrompt,
		AutoApproveThreshold: settings.AIAutoApproveThreshold,
		AutoRejectThreshold:  settings.AIAutoRejectThreshold,
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.openai.com/v1/chat/completions"
	}
	if cfg.AutoApproveThreshold <= 0 {
		cfg.AutoApproveThreshold = 0.85
	}
	if cfg.AutoRejectThreshold >= cfg.AutoApproveThreshold {
		cfg.AutoRejectThreshold = 0.15
	}
	return cfg, nil
}

// ContentPolicy exposes the global content-access policy to other services
// (which may not read the auth database directly — ADR 0016).
type ContentPolicy struct {
	ForbidMatureForFree bool `json:"forbid_mature_for_free"`
	HideMatureDefault   bool `json:"hide_mature_default"`
	EnableComments      bool `json:"enable_comments"`
}

//encore:api private method=GET path=/auth/content-policy
func GetContentPolicy(ctx context.Context) (*ContentPolicy, error) {
	settings := loadSettings(ctx)
	return &ContentPolicy{
		ForbidMatureForFree: settings.ForbidMatureForFree,
		HideMatureDefault:   settings.HideMatureDefault,
		EnableComments:      settings.EnableComments,
	}, nil
}

// BillingConfig exposes the subscription-expiry job settings to the billing
// service (which may not read the auth database directly — ADR 0016).
type BillingConfig struct {
	WaitingPayJobEnabled  bool `json:"waiting_pay_job_enabled"`
	WaitingPayExpiryHours int  `json:"waiting_pay_expiry_hours"`
}

//encore:api private method=GET path=/auth/billing-config
func GetBillingConfig(ctx context.Context) (*BillingConfig, error) {
	var raw []byte
	err := db.QueryRow(ctx, `SELECT value FROM app_settings WHERE key = 'defaults'`).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return &BillingConfig{WaitingPayJobEnabled: true, WaitingPayExpiryHours: 24}, nil
	}
	// Merge onto defaults so settings written before these keys existed still
	// resolve to the intended defaults (enabled, 24h).
	settings := *defaultAppSettings()
	if err := json.Unmarshal(raw, &settings); err != nil {
		return &BillingConfig{WaitingPayJobEnabled: true, WaitingPayExpiryHours: 24}, nil
	}
	hours := settings.WaitingPayExpiryHours
	if hours <= 0 {
		hours = 24
	}
	return &BillingConfig{
		WaitingPayJobEnabled:  settings.WaitingPayJobEnabled,
		WaitingPayExpiryHours: hours,
	}, nil
}

// BoostConfig exposes the quota-boost tiers to the billing service (which may
// not read the auth database directly — ADR 0016). Per-tier download quotas
// now live on the tiers table (see tiers.GetTierQuotas).
type BoostConfig struct {
	Boost1Downloads int     `json:"boost_1_downloads"`
	Boost1Price     float64 `json:"boost_1_price"`
	Boost2Downloads int     `json:"boost_2_downloads"`
	Boost2Price     float64 `json:"boost_2_price"`
	Boost3Downloads int     `json:"boost_3_downloads"`
	Boost3Price     float64 `json:"boost_3_price"`
}

//encore:api private method=GET path=/auth/boost-config
func GetBoostConfig(ctx context.Context) (*BoostConfig, error) {
	settings := loadSettings(ctx)
	return &BoostConfig{
		Boost1Downloads: settings.Boost1Downloads,
		Boost1Price:     settings.Boost1Price,
		Boost2Downloads: settings.Boost2Downloads,
		Boost2Price:     settings.Boost2Price,
		Boost3Downloads: settings.Boost3Downloads,
		Boost3Price:     settings.Boost3Price,
	}, nil
}

// UserPublicInfo is the minimal public identity another service needs to
// display/address a user (e.g. comment authors, uploaders) without reading the
// auth database directly.
type UserPublicInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarKey string `json:"avatar_key"`
}

type GetUsersInfoParams struct {
	IDs []string `json:"ids"`
}

type GetUsersInfoResponse struct {
	Users []UserPublicInfo `json:"users"`
}

// GetUsersInfo returns public identity for a batch of user IDs.
//encore:api private method=POST path=/auth/users-info
func GetUsersInfo(ctx context.Context, p *GetUsersInfoParams) (*GetUsersInfoResponse, error) {
	if len(p.IDs) == 0 {
		return &GetUsersInfoResponse{Users: []UserPublicInfo{}}, nil
	}

	rows, err := db.Query(ctx, `
		SELECT id, COALESCE(username, ''), COALESCE(avatar_key::text, '')
		FROM users WHERE id = ANY($1)
	`, p.IDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]UserPublicInfo, 0, len(p.IDs))
	for rows.Next() {
		var u UserPublicInfo
		if err := rows.Scan(&u.ID, &u.Username, &u.AvatarKey); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return &GetUsersInfoResponse{Users: users}, rows.Err()
}

//encore:api auth method=POST path=/auth/resend-verification
func ResendVerification(ctx context.Context) error {
	data := auth.Data().(*AuthData)

	var verified bool
		db.QueryRow(ctx, `SELECT email_verified_at IS NOT NULL FROM users WHERE id = $1`, data.UserID).Scan(&verified)
	if verified {
		return &errs.Error{Code: errs.InvalidArgument, Message: "email already verified"}
	}

	token := randomToken(32)
	_, err := db.Exec(ctx, `
		UPDATE users SET verify_token = $1, verify_token_expires_at = now() + interval '24 hours'
		WHERE id = $2 AND email_verified_at IS NULL
	`, token, data.UserID)
	if err != nil {
		return err
	}

	var email string
	db.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, data.UserID).Scan(&email)
	if email != "" {
		go sendVerificationEmail(email, token)
	}
	return nil
}

// ----- Password Reset -----

type PasswordResetRequest struct {
	Email          string `json:"email"`
	TurnstileToken string `json:"turnstile_token"`
}

//encore:api public method=POST path=/auth/password-reset/request
func RequestPasswordReset(ctx context.Context, p *PasswordResetRequest) error {
	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "password_reset"}); err != nil {
		return err
	}

	if p.Email == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "email is required"}
	}

	token := randomToken(32)
	result, err := db.Exec(ctx, `
		UPDATE users SET reset_token = $1, reset_token_expires_at = now() + interval '1 hour'
		WHERE email = $2
	`, token, p.Email)
	if err != nil {
		return err
	}
	n := result.RowsAffected()
	if n > 0 {
		go sendPasswordResetEmail(p.Email, token)
	}
	return nil
}

type PasswordResetConfirm struct {
	Token    string `json:"token"`
	Password string `json:"password" encore:"sensitive"`
}

//encore:api public method=POST path=/auth/password-reset/confirm
func ConfirmPasswordReset(ctx context.Context, p *PasswordResetConfirm) error {
	if p.Token == "" || p.Password == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "token and password are required"}
	}
	if len(p.Password) < 8 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "password must be at least 8 characters"}
	}

	hash, err := hashPassword(p.Password)
	if err != nil {
		return err
	}

	result, err := db.Exec(ctx, `
		UPDATE users SET password_hash = $1, reset_token = NULL, reset_token_expires_at = NULL
		WHERE reset_token = $2 AND reset_token_expires_at > now()
	`, hash, p.Token)
	if err != nil {
		return err
	}
	n := result.RowsAffected()
	if n == 0 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid or expired reset token"}
	}
	return nil
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Tier      string    `json:"tier"`
	Username  string    `json:"username,omitempty"`
	AvatarKey string    `json:"avatar_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type userRow struct {
	ID           string
	Email        sql.NullString
	PasswordHash sql.NullString
	Role         string
	Tier         string
	Username     sql.NullString
	AvatarKey    sql.NullString
	BannedAt     sql.NullTime
	SuspendedAt  sql.NullTime
	CreatedAt    time.Time
}

func getUserByEmail(ctx context.Context, email string) (*userRow, error) {
	var u userRow
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, tier, COALESCE(username, ''), COALESCE(avatar_key::text, ''), banned_at, suspended_at, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Tier, &u.Username, &u.AvatarKey, &u.BannedAt, &u.SuspendedAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	var email sql.NullString
	err := db.QueryRow(ctx, `
		SELECT id, email, role, tier, COALESCE(username, ''), COALESCE(avatar_key::text, ''), created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &email, &u.Role, &u.Tier, &u.Username, &u.AvatarKey, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.Email = email.String
	return &u, nil
}

// ----- Extended Profile -----

type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Tier      string `json:"tier"`
	Username  string `json:"username,omitempty"`
	AvatarKey string `json:"avatar_key,omitempty"`
	CreatedAt string `json:"created_at"`
}

//encore:api auth method=GET path=/me/profile
func GetProfile(ctx context.Context) (*UserProfile, error) {
	data := auth.Data().(*AuthData)
	user, err := getUserByID(ctx, data.UserID)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "user not found"}
	}
	return &UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		Tier:      user.Tier,
		Username:  user.Username,
		AvatarKey: user.AvatarKey,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

type UpdateUsernameParams struct {
	Username string `json:"username"`
}

//encore:api auth method=POST path=/me/username
func UpdateUsername(ctx context.Context, p *UpdateUsernameParams) (*UserProfile, error) {
	data := auth.Data().(*AuthData)

	username := strings.ToLower(strings.TrimSpace(p.Username))
	if username == "" || !validUsername(username) {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "username must be 3-20 characters, lowercase letters, numbers, and single - or _ in between"}
	}

	var taken bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)`, username, data.UserID).Scan(&taken); err != nil {
		return nil, err
	}
	if taken {
		return nil, &errs.Error{Code: errs.AlreadyExists, Message: "this username is already taken"}
	}

	if _, err := db.Exec(ctx, `UPDATE users SET username = $1 WHERE id = $2`, username, data.UserID); err != nil {
		return nil, err
	}

	user, err := getUserByID(ctx, data.UserID)
	if err != nil {
		return nil, err
	}
	return &UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		Tier:      user.Tier,
		Username:  user.Username,
		AvatarKey: user.AvatarKey,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

type UpdateAvatarParams struct {
	AvatarData string `json:"avatar_data"`
}

type UpdateAvatarResponse struct {
	AvatarKey string `json:"avatar_key"`
}

//encore:api auth method=POST path=/me/avatar
func UpdateAvatar(ctx context.Context, p *UpdateAvatarParams) (*UpdateAvatarResponse, error) {
	data := auth.Data().(*AuthData)

	if !strings.HasPrefix(p.AvatarData, "data:image/") {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid image format, must be data:image/... base64"}
	}

	// Parse the base64 portion after the comma
	idx := strings.IndexByte(p.AvatarData, ',')
	if idx < 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid data URI"}
	}
	b64 := p.AvatarData[idx+1:]
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid base64 encoding"}
	}

	if len(decoded) > 500*1024 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "avatar image too large (max 500KB)"}
	}

	key := "avatar-" + data.UserID + ".png"
	uploadURL, err := AvatarBucket.SignedUploadURL(ctx, key, objects.WithTTL(7200*time.Second))
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "failed to generate upload URL"}
	}

	body := strings.NewReader(string(decoded))
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL.URL, body)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "upload request failed"}
	}
	req.Header.Set("Content-Type", "image/png")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "upload failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, &errs.Error{Code: errs.Internal, Message: "upload rejected by storage"}
	}

	_, err = db.Exec(ctx, `UPDATE users SET avatar_key = $1 WHERE id = $2`, key, data.UserID)
	if err != nil {
		return nil, err
	}

	return &UpdateAvatarResponse{AvatarKey: key}, nil
}

//encore:api auth method=GET path=/me/avatar
func GetAvatar(ctx context.Context) (*UpdateAvatarResponse, error) {
	data := auth.Data().(*AuthData)
	var key string
	db.QueryRow(ctx, `SELECT COALESCE(avatar_key, '') FROM users WHERE id = $1`, data.UserID).Scan(&key)
	return &UpdateAvatarResponse{AvatarKey: key}, nil
}

// ----- Admin endpoints -----

type AdminUser struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Tier        string     `json:"tier"`
	CreatedAt   time.Time  `json:"created_at"`
	BannedAt    *time.Time `json:"banned_at,omitempty"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty"`
}

type AdminUserListResponse struct {
	Users []AdminUser `json:"users"`
	Total int         `json:"total"`
}

type AdminListUsersParams struct {
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
	Search     string `query:"search"`
	Sort       string `query:"sort"`
	SortDir    string `query:"sort_dir"`
	FilterRole string `query:"filter_role"`
	FilterTier string `query:"filter_tier"`
	FilterEmail string `query:"filter_email"`
}

//encore:api auth method=GET path=/admin/users
func AdminListUsers(ctx context.Context, p *AdminListUsersParams) (*AdminUserListResponse, error) {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := p.Page
	if page <= 0 { page = 1 }
	limit := p.Limit
	if limit <= 0 { limit = 20 }
	if limit > 100 { limit = 100 }
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := "created_at"
	sortDir := "DESC"
	switch p.Sort {
	case "email": sortCol = "email"
	case "role": sortCol = "role"
	case "tier": sortCol = "tier"
	case "created_at": sortCol = "created_at"
	}
	if strings.ToLower(p.SortDir) == "asc" { sortDir = "ASC" }

	where := "WHERE (email ILIKE $1)"
	args := []interface{}{search}
	argIdx := 2

	if p.FilterRole != "" {
		where += fmt.Sprintf(" AND role = $%d", argIdx)
		args = append(args, p.FilterRole)
		argIdx++
	}
	if p.FilterTier != "" {
		where += fmt.Sprintf(" AND tier = $%d", argIdx)
		args = append(args, p.FilterTier)
		argIdx++
	}
	if p.FilterEmail != "" {
		where += fmt.Sprintf(" AND email ILIKE $%d", argIdx)
		args = append(args, "%"+p.FilterEmail+"%")
		argIdx++
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users `+where, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT id, email, role, tier, created_at, banned_at, suspended_at
		FROM users %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.CreatedAt, &u.BannedAt, &u.SuspendedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return &AdminUserListResponse{Users: users, Total: total}, rows.Err()
}

type UpdateUserRoleParams struct {
	Role string `json:"role"`
}

//encore:api auth method=POST path=/admin/users/:id/role
func AdminUpdateUserRole(ctx context.Context, id string, p *UpdateUserRoleParams) error {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	validRoles := map[string]bool{"user": true, "uploader": true, "moderator": true, "admin": true}
	if !validRoles[p.Role] {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid role"}
	}

	_, err := db.Exec(ctx, `UPDATE users SET role = $1 WHERE id = $2 AND id != $3`,
		p.Role, id, data.UserID)
	if err != nil {
		return err
	}

	details, _ := json.Marshal(map[string]string{"new_role": p.Role})
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'change_role', 'user', $2, $3)`,
		data.UserID, id, string(details))

	return nil
}

type BanUserParams struct {
	Reason string `json:"reason"`
}

//encore:api auth method=POST path=/admin/users/:id/ban
func AdminBanUser(ctx context.Context, id string, p *BanUserParams) error {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `UPDATE users SET banned_at = now() WHERE id = $1 AND id != $2`, id, data.UserID)
	return err
}

//encore:api auth method=POST path=/admin/users/:id/unban
func AdminUnbanUser(ctx context.Context, id string) error {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `UPDATE users SET banned_at = NULL WHERE id = $1`, id)
	return err
}

//encore:api auth method=POST path=/admin/users/:id/suspend
func AdminSuspendUser(ctx context.Context, id string, p *BanUserParams) error {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `UPDATE users SET suspended_at = now() WHERE id = $1 AND id != $2`, id, data.UserID)
	return err
}

//encore:api auth method=POST path=/admin/users/:id/unsuspend
func AdminUnsuspendUser(ctx context.Context, id string) error {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `UPDATE users SET suspended_at = NULL WHERE id = $1`, id)
	return err
}

type ImpersonateResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

//encore:api auth method=POST path=/admin/users/:id/impersonate
func AdminImpersonateUser(ctx context.Context, id string) (*ImpersonateResponse, error) {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	user, err := getUserByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "user not found"}
		}
		return nil, err
	}

	token, err := createImpersonationSession(ctx, user.ID, data.UserID)
	if err != nil {
		return nil, err
	}

	details, _ := json.Marshal(map[string]string{"impersonated": id})
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'impersonate', 'user', $2, $3)`,
		data.UserID, id, string(details))

	return &ImpersonateResponse{Token: token, User: *user}, nil
}

// ----- Notification Preferences -----

type NotificationPrefs struct {
	EmailFromFollowing bool `json:"email_new_from_following"`
	EmailSupportReplies bool `json:"email_support_replies"`
	EmailMarketing     bool `json:"email_marketing"`
	InAppEnabled       bool `json:"in_app_enabled"`
}

//encore:api auth method=GET path=/me/notification-preferences
func GetNotificationPrefs(ctx context.Context) (*NotificationPrefs, error) {
	data := auth.Data().(*AuthData)
	var p NotificationPrefs
	err := db.QueryRow(ctx, `
		SELECT COALESCE(np.email_new_from_following, true),
			COALESCE(np.email_support_replies, true),
			COALESCE(np.email_marketing, false),
			COALESCE(np.in_app_enabled, true)
		FROM notification_preferences np WHERE np.user_id = $1
	`, data.UserID).Scan(&p.EmailFromFollowing, &p.EmailSupportReplies, &p.EmailMarketing, &p.InAppEnabled)
	if err != nil {
		if isNoRows(err) {
			return &NotificationPrefs{EmailFromFollowing: true, EmailSupportReplies: true, InAppEnabled: true}, nil
		}
		return nil, err
	}
	return &p, nil
}

//encore:api auth method=PATCH path=/me/notification-preferences
func UpdateNotificationPrefs(ctx context.Context, p *NotificationPrefs) (*NotificationPrefs, error) {
	data := auth.Data().(*AuthData)
	_, err := db.Exec(ctx, `
		INSERT INTO notification_preferences (user_id, email_new_from_following, email_support_replies, email_marketing, in_app_enabled)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			email_new_from_following = $2, email_support_replies = $3,
			email_marketing = $4, in_app_enabled = $5, updated_at = now()
	`, data.UserID, p.EmailFromFollowing, p.EmailSupportReplies, p.EmailMarketing, p.InAppEnabled)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ----- Admin Dashboard Stats -----

type DashboardStats struct {
	TotalUsers      int `json:"total_users"`
	NewUsersThisMonth int `json:"new_users_this_month"`
}

//encore:api private
func AdminDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= date_trunc('month', now())`).Scan(&stats.NewUsersThisMonth)

	return &stats, nil
}

type SignupTrendPoint struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type SignupTrendResponse struct {
	Points []SignupTrendPoint `json:"points"`
}

//encore:api private
func GetSignupTrend(ctx context.Context) (*SignupTrendResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT to_char(date_trunc('day', created_at), 'YYYY-MM-DD') AS day, COUNT(*)
		FROM users
		WHERE created_at >= now() - interval '30 days'
		GROUP BY 1
		ORDER BY 1 ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := []SignupTrendPoint{}
	for rows.Next() {
		var p SignupTrendPoint
		if err := rows.Scan(&p.Day, &p.Count); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if points == nil {
		points = []SignupTrendPoint{}
	}
	return &SignupTrendResponse{Points: points}, rows.Err()
}

// ----- CSV Export -----

//encore:api auth raw method=GET path=/admin/export/:resource
func ExportCSV(w http.ResponseWriter, req *http.Request) {
	data, ok := auth.Data().(*AuthData)
	if !ok || data.Role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}

	resource := req.PathValue("resource")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="`+resource+`.csv"`)

	switch resource {
	case "users":
		rows, err := db.Query(req.Context(), `SELECT id, email, role, tier, created_at FROM users ORDER BY created_at DESC`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		fmt.Fprintln(w, "id,email,role,tier,created_at")
		for rows.Next() {
			var id, email, role, tier string
			var createdAt time.Time
			if rows.Scan(&id, &email, &role, &tier, &createdAt) == nil {
				fmt.Fprintf(w, "%s,%s,%s,%s,%s\n", id, email, role, tier, createdAt.Format(time.RFC3339))
			}
		}
	default:
		http.Error(w, "unsupported resource", http.StatusNotFound)
	}
}
