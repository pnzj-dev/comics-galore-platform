package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var secrets struct {
	JWTSecret string
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

	claims, err := validateToken(token)
	if err != nil {
		return "", nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "invalid or expired token",
		}
	}

	var maintenance bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'maintenance_mode')::boolean, false) FROM app_settings WHERE key = 'defaults'`).Scan(&maintenance); e == nil && maintenance && claims.Role != "admin" {
		return "", nil, &errs.Error{Code: errs.Unavailable, Message: "the platform is under maintenance, please try again later"}
	}

	var requireVerify bool
	if e := db.QueryRow(ctx, `SELECT COALESCE((value::jsonb->>'require_email_verify')::boolean, false) FROM app_settings WHERE key = 'defaults'`).Scan(&requireVerify); e == nil && requireVerify {
		var emailVerified sql.NullTime
		if e2 := db.QueryRow(ctx, `SELECT email_verified_at FROM users WHERE id = $1`, claims.UserID).Scan(&emailVerified); e2 == nil && !emailVerified.Valid {
			return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "email verification required"}
		}
	}

	var bannedAt, suspendedAt sql.NullTime
	if err := db.QueryRow(ctx, `SELECT banned_at, suspended_at FROM users WHERE id = $1`, claims.UserID).Scan(&bannedAt, &suspendedAt); err == nil {
		if bannedAt.Valid {
			return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "account is banned"}
		}
		if suspendedAt.Valid {
			return "", nil, &errs.Error{Code: errs.PermissionDenied, Message: "account is suspended"}
		}
	}

	return auth.UID(claims.UserID), &AuthData{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Tier:   claims.Tier,
	}, nil
}

type RegisterParams struct {
	Email    string `json:"email"`
	Password string `json:"password" encore:"sensitive"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

//encore:api public method=POST path=/auth/register
func Register(ctx context.Context, p *RegisterParams) (*AuthResponse, error) {
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
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, role, tier, created_at
	`, p.Email, hash, role).Scan(&user.ID, &user.Email, &user.Role, &user.Tier, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	token, err := generateToken(&Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Tier:   user.Tier,
	})
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

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password" encore:"sensitive"`
}

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, p *LoginParams) (*AuthResponse, error) {
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

	if !checkPassword(user.PasswordHash, p.Password) {
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

	token, err := generateToken(&Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Tier:   user.Tier,
	})
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			Tier:      user.Tier,
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

	token, err := generateToken(&Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Tier:   user.Tier,
	})
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
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
	result, err := db.Exec(ctx, `
		UPDATE users SET email_verified_at = now(), verify_token = NULL, verify_token_expires_at = NULL
		WHERE verify_token = $1 AND verify_token_expires_at > now() AND email_verified_at IS NULL
	`, p.Token)
	if err != nil {
		return err
	}
	n := result.RowsAffected()
	if n == 0 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid or expired verification token"}
	}
	return nil
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
	Email string `json:"email"`
}

//encore:api public method=POST path=/auth/password-reset/request
func RequestPasswordReset(ctx context.Context, p *PasswordResetRequest) error {
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
	AvatarKey string    `json:"avatar_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type userRow struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	Tier         string
	AvatarKey    sql.NullString
	BannedAt     sql.NullTime
	SuspendedAt  sql.NullTime
	CreatedAt    time.Time
}

func getUserByEmail(ctx context.Context, email string) (*userRow, error) {
	var u userRow
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, tier, COALESCE(avatar_key::text, ''), banned_at, suspended_at, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Tier, &u.AvatarKey, &u.BannedAt, &u.SuspendedAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := db.QueryRow(ctx, `
		SELECT id, email, role, tier, COALESCE(avatar_key::text, ''), created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.AvatarKey, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ----- Extended Profile -----

type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Tier      string `json:"tier"`
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
	TotalUsers        int `json:"total_users"`
	TotalComics       int `json:"total_comics"`
	PendingComics     int `json:"pending_comics"`
	ActiveSubs        int `json:"active_subscriptions"`
	TotalDownloads    int `json:"total_downloads"`
	TotalViews        int `json:"total_views"`
}

//encore:api auth method=GET path=/admin/stats
func AdminDashboardStats(ctx context.Context) (*DashboardStats, error) {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var stats DashboardStats
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics`).Scan(&stats.TotalComics)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE status = 'pending_review'`).Scan(&stats.PendingComics)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE active = true`).Scan(&stats.ActiveSubs)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM download_logs`).Scan(&stats.TotalDownloads)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(view_count), 0) FROM comics`).Scan(&stats.TotalViews)

	return &stats, nil
}
