package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

const totpIssuer = "Comics Galore"
const mfaChallengeTTL = 5 * time.Minute

// userTOTPSecret returns the user's TOTP secret and whether 2FA is enabled.
func userTOTPSecret(ctx context.Context, userID string) (secret string, enabled bool, err error) {
	var s sql.NullString
	if err := db.QueryRow(ctx, `SELECT totp_secret FROM users WHERE id = $1`, userID).Scan(&s); err != nil {
		if isNoRows(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return s.String, s.Valid && s.String != "", nil
}

// generateTOTP creates a fresh TOTP secret, its otpauth URL, and a QR code
// rendered as a PNG data URL (so the frontend needs no QR library and the
// secret never leaves the backend).
func generateTOTP(accountName string) (secret, otpauthURL, qrDataURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", "", err
	}

	img, err := key.Image(240, 240)
	if err != nil {
		return "", "", "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", "", err
	}

	return key.Secret(), key.URL(), "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func validateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

func storeMFAChallenge(ctx context.Context, userID string) (string, error) {
	var id string
	err := db.QueryRow(ctx, `
		INSERT INTO mfa_challenges (user_id, expires_at)
		VALUES ($1, now() + $2::interval)
		RETURNING id
	`, userID, mfaChallengeTTL.String()).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func consumeMFAChallenge(ctx context.Context, token string) (string, error) {
	var userID string
	var expiresAt time.Time
	err := db.QueryRow(ctx, `SELECT user_id, expires_at FROM mfa_challenges WHERE id = $1`, token).Scan(&userID, &expiresAt)
	if err != nil {
		if isNoRows(err) {
			return "", &errs.Error{Code: errs.Unauthenticated, Message: "invalid or expired code request"}
		}
		return "", err
	}
	// Single-use: delete before returning so a replay cannot reuse it.
	if _, err := db.Exec(ctx, `DELETE FROM mfa_challenges WHERE id = $1`, token); err != nil {
		return "", err
	}
	if time.Now().After(expiresAt) {
		return "", &errs.Error{Code: errs.Unauthenticated, Message: "code request expired"}
	}
	return userID, nil
}

// ----- TOTP management (Security modal) -----

type TOTPStatusResponse struct {
	Enabled bool `json:"enabled"`
}

//encore:api auth method=GET path=/me/totp
func TOTPStatus(ctx context.Context) (*TOTPStatusResponse, error) {
	ad := auth.Data().(*AuthData)
	_, enabled, err := userTOTPSecret(ctx, ad.UserID)
	if err != nil {
		return nil, err
	}
	return &TOTPStatusResponse{Enabled: enabled}, nil
}

type SetupTOTPResponse struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauth_url"`
	QRImage    string `json:"qr_image"`
}

//encore:api auth method=POST path=/me/totp/setup
func SetupTOTP(ctx context.Context) (*SetupTOTPResponse, error) {
	ad := auth.Data().(*AuthData)
	user, err := getUserByID(ctx, ad.UserID)
	if err != nil {
		return nil, err
	}
	secret, otpauthURL, qr, err := generateTOTP(user.Email)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "failed to generate TOTP secret"}
	}
	return &SetupTOTPResponse{Secret: secret, OtpauthURL: otpauthURL, QRImage: qr}, nil
}

type ConfirmTOTPParams struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

type ConfirmTOTPResponse struct {
	Enabled bool `json:"enabled"`
}

//encore:api auth method=POST path=/me/totp/confirm
func ConfirmTOTP(ctx context.Context, p *ConfirmTOTPParams) (*ConfirmTOTPResponse, error) {
	ad := auth.Data().(*AuthData)
	if p.Secret == "" || p.Code == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "secret and code are required"}
	}
	if !validateTOTP(p.Secret, p.Code) {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid code"}
	}
	if _, err := db.Exec(ctx, `UPDATE users SET totp_secret = $1 WHERE id = $2`, p.Secret, ad.UserID); err != nil {
		return nil, err
	}
	return &ConfirmTOTPResponse{Enabled: true}, nil
}

type DisableTOTPParams struct {
	Code string `json:"code"`
}

//encore:api auth method=POST path=/me/totp/disable
func DisableTOTP(ctx context.Context, p *DisableTOTPParams) error {
	ad := auth.Data().(*AuthData)
	secret, enabled, err := userTOTPSecret(ctx, ad.UserID)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	if !validateTOTP(secret, p.Code) {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid code"}
	}
	_, err = db.Exec(ctx, `UPDATE users SET totp_secret = NULL WHERE id = $1`, ad.UserID)
	return err
}

// ----- TOTP login step -----

type VerifyTOTPLoginParams struct {
	MFAToken string `json:"mfa_token"`
	Code     string `json:"code"`
}

//encore:api public method=POST path=/auth/login/totp
func VerifyTOTPLogin(ctx context.Context, p *VerifyTOTPLoginParams) (*AuthResponse, error) {
	userID, err := consumeMFAChallenge(ctx, p.MFAToken)
	if err != nil {
		return nil, err
	}

	secret, enabled, err := userTOTPSecret(ctx, userID)
	if err != nil || !enabled {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "two-factor authentication is not enabled"}
	}
	if !validateTOTP(secret, p.Code) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid code"}
	}

	user, err := getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	token, err := createSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: *user}, nil
}
