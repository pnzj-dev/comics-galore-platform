package auth

import (
	"context"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var secrets struct {
	JWTSecret string
}

var db = sqldb.NewDatabase("authdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

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
	var adminCount int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount)
	if err != nil {
		return nil, err
	}
	if adminCount == 0 {
		role = "admin"
	}

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

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
}

type userRow struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	Tier         string
	CreatedAt    time.Time
}

func getUserByEmail(ctx context.Context, email string) (*userRow, error) {
	var u userRow
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, tier, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Tier, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := db.QueryRow(ctx, `
		SELECT id, email, role, tier, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ----- Admin endpoints -----

type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminUserListResponse struct {
	Users []AdminUser `json:"users"`
	Total int         `json:"total"`
}

//encore:api auth method=GET path=/admin/users
func AdminListUsers(ctx context.Context) (*AdminUserListResponse, error) {
	data := auth.Data().(*AuthData)
	if data.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)

	rows, err := db.Query(ctx, `
		SELECT id, email, role, tier, created_at
		FROM users ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.Tier, &u.CreatedAt); err != nil {
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

	_, err := db.Exec(ctx, `UPDATE users SET role = $1 WHERE id = $2 AND id != $3`,
		p.Role, id, data.UserID)
	return err
}
