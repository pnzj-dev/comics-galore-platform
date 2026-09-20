package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"comics-galore/backend/nowpayments"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var secrets struct {
	NowPaymentsAPIKey   string
	NowPaymentsIPNKey   string
	NowPaymentsEmail    string
	NowPaymentsPassword string

	// Email (Resend)
	ResendAPIKey string

	// BootstrapSecret gates the one-time first-admin provisioning endpoint
	// (/auth/bootstrap). Empty disables bootstrap entirely.
	BootstrapSecret string

	// Logto identity provider (OIDC).
	LogtoIssuer   string // e.g. https://<tenant>.logto.app/oidc
	LogtoJWKSURI  string // e.g. https://<tenant>.logto.app/oidc/jwks
	LogtoAudience string // optional; if set, the token aud claim is enforced
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
	UserID string
	Email  string
	Role   string
	Tier   string
}

//encore:authhandler
func AuthHandler(ctx context.Context, p *AuthParams) (auth.UID, *AuthData, error) {
	token := strings.TrimSpace(p.Authorization)
	if token == "" {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing authorization header"}
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid authorization format"}
	}

	claims, err := validateLogtoToken(ctx, token)
	if err != nil {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid or expired token"}
	}

	if claims.Sub == "" {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "token missing subject"}
	}

	u, err := resolveLogtoUser(ctx, claims.Sub, claims.Email)
	if err != nil {
		return "", nil, err
	}

	// Role lives on the internal users row (app-level; Logto only holds
	// identity + credentials).
	role := u.Role

	var maintenance bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'maintenance_mode')::boolean, false) FROM app_settings WHERE key = 'defaults'`).Scan(&maintenance); e == nil && maintenance && role != "admin" {
		return "", nil, &errs.Error{Code: errs.Unavailable, Message: "the platform is under maintenance, please try again later"}
	}

	if u.BannedAt.Valid {
		return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "account is banned"}
	}
	if u.SuspendedAt.Valid {
		return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "account is suspended"}
	}

	db.Exec(ctx, `UPDATE users SET last_seen_at = now() WHERE id = $1`, u.ID)

	return auth.UID(u.ID), &AuthData{
		UserID: u.ID,
		Email:  u.Email.String,
		Role:   role,
		Tier:   u.Tier,
	}, nil
}

// usernameRe validates a public handle's characters/structure: starts and ends
// with a lowercase alphanumeric, with single `_`/`-` only between alphanumerics
// (no leading/trailing/consecutive). Length is checked separately (3-20).
var usernameRe = regexp.MustCompile(`^[a-z0-9](?:[_-]?[a-z0-9])*$`)

func validUsername(username string) bool {
	return len(username) >= 3 && len(username) <= 20 && usernameRe.MatchString(username)
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

type GetUserRoleParams struct {
	UserID string `json:"user_id"`
}

type GetUserRoleResponse struct {
	Role string `json:"role"`
}

// GetUserRole returns a user's role to other services that must authorize an
// actor without reading the auth database directly (ADR 0016).
//encore:api private method=POST path=/auth/user-role
func GetUserRole(ctx context.Context, p *GetUserRoleParams) (*GetUserRoleResponse, error) {
	var role string
	if err := db.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, p.UserID).Scan(&role); err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "user not found"}
		}
		return nil, err
	}
	return &GetUserRoleResponse{Role: role}, nil
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
	UserID  string `json:"user_id"`
	Subject string `json:"subject"`
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
	Enabled              bool    `json:"enabled"`
	Model                string  `json:"model"`
	Endpoint             string  `json:"endpoint"`
	Prompt               string  `json:"prompt"`
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
	ID          string
	Email       sql.NullString
	Role        string
	Tier        string
	Username    sql.NullString
	AvatarKey   sql.NullString
	BannedAt    sql.NullTime
	SuspendedAt sql.NullTime
	CreatedAt   time.Time
}

func getUserByEmail(ctx context.Context, email string) (*userRow, error) {
	var u userRow
	err := db.QueryRow(ctx, `
		SELECT id, email, role, tier, COALESCE(username, ''), COALESCE(avatar_key::text, ''), banned_at, suspended_at, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.Username, &u.AvatarKey, &u.BannedAt, &u.SuspendedAt, &u.CreatedAt)
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
	Page        int    `query:"page"`
	Limit       int    `query:"limit"`
	Search      string `query:"search"`
	Sort        string `query:"sort"`
	SortDir     string `query:"sort_dir"`
	FilterRole  string `query:"filter_role"`
	FilterTier  string `query:"filter_tier"`
	FilterEmail string `query:"filter_email"`
}

//encore:api auth method=GET path=/admin/users
func AdminListUsers(ctx context.Context, p *AdminListUsersParams) (*AdminUserListResponse, error) {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := p.Page
	if page <= 0 {
		page = 1
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := "created_at"
	sortDir := "DESC"
	switch p.Sort {
	case "email":
		sortCol = "email"
	case "role":
		sortCol = "role"
	case "tier":
		sortCol = "tier"
	case "created_at":
		sortCol = "created_at"
	}
	if strings.ToLower(p.SortDir) == "asc" {
		sortDir = "ASC"
	}

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
	if err != nil {
		return nil, err
	}
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
