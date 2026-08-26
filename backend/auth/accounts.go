package auth

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

// AccountInfo describes one linked authentication method.
type AccountInfo struct {
	ID         string `json:"id"`
	Provider   string `json:"provider"`
	Email      string `json:"email,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AccountsResponse struct {
	Accounts []AccountInfo `json:"accounts"`
}

//encore:api auth method=GET path=/auth/accounts
func ListAccounts(ctx context.Context) (*AccountsResponse, error) {
	ad := auth.Data().(*AuthData)

	rows, err := db.Query(ctx, `
		SELECT id, provider, COALESCE(email, ''), created_at
		FROM auth_accounts WHERE user_id = $1 ORDER BY created_at ASC
	`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out AccountsResponse
	for rows.Next() {
		var a AccountInfo
		if err := rows.Scan(&a.ID, &a.Provider, &a.Email, &a.CreatedAt); err != nil {
			return nil, err
		}
		out.Accounts = append(out.Accounts, a)
	}
	return &out, rows.Err()
}

// countAuthMethods returns the number of usable authentication methods a user
// has: password (1 if set), linked OAuth accounts, and passkeys.
func countAuthMethods(ctx context.Context, userID string) (int, error) {
	var hasPassword bool
	if err := db.QueryRow(ctx, `SELECT password_hash IS NOT NULL FROM users WHERE id = $1`, userID).Scan(&hasPassword); err != nil {
		if isNoRows(err) {
			return 0, &errs.Error{Code: errs.NotFound, Message: "user not found"}
		}
		return 0, err
	}
	count := 0
	if hasPassword {
		count++
	}

	var accounts int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM auth_accounts WHERE user_id = $1`, userID).Scan(&accounts); err != nil {
		return 0, err
	}
	count += accounts

	var passkeys int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM passkeys WHERE user_id = $1`, userID).Scan(&passkeys); err != nil {
		return 0, err
	}
	count += passkeys

	return count, nil
}

//encore:api auth method=DELETE path=/auth/accounts/:id
func UnlinkAccount(ctx context.Context, id string) error {
	ad := auth.Data().(*AuthData)
	if id == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "id is required"}
	}

	// Verify the account belongs to this user.
	var exists bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth_accounts WHERE id = $1 AND user_id = $2)`, id, ad.UserID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return &errs.Error{Code: errs.NotFound, Message: "account not found"}
	}

	// Recovery guard: a user must retain at least one authentication method.
	count, err := countAuthMethods(ctx, ad.UserID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "you must keep at least one sign-in method"}
	}

	_, err = db.Exec(ctx, `DELETE FROM auth_accounts WHERE id = $1 AND user_id = $2`, id, ad.UserID)
	return err
}

// guardRemoveAuthMethod is the recovery guard shared by account unlink and
// passkey deletion: a user must not remove their final authentication method.
func guardRemoveAuthMethod(ctx context.Context, userID string) error {
	count, err := countAuthMethods(ctx, userID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "you must keep at least one sign-in method"}
	}
	return nil
}
