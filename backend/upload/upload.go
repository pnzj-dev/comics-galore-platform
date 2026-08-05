package upload

import (
	"context"
	"time"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var ComicBucket = objects.NewBucket("comic-files", objects.BucketConfig{})

var db = sqldb.NewDatabase("uploaddb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type CreateSessionParams struct {
	Mode string `json:"mode"`
}

type UploadSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Mode      string    `json:"mode"`
	Status    string    `json:"status"`
	S3Prefix  string    `json:"s3_prefix"`
	Parts     []Part    `json:"parts"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Part struct {
	Number int    `json:"number"`
	Key    string `json:"key"`
	ETag   string `json:"etag,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

func scanSession(rs *sqldb.Row, s *UploadSession) error {
	var partsBytes []byte
	err := rs.Scan(&s.ID, &s.UserID, &s.Mode, &s.Status, &s.S3Prefix,
		&partsBytes, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return err
	}
	return scanParts(partsBytes, &s.Parts)
}

//encore:api auth method=POST path=/upload-sessions
func CreateSession(ctx context.Context, p *CreateSessionParams) (*UploadSession, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "only uploaders can create sessions",
		}
	}

	mode := p.Mode
	if mode != "manual" && mode != "archive" {
		mode = "manual"
	}

	now := time.Now()
	prefix := "uploads/" + ad.UserID + "/" + now.Format("2006/01/02/150405")

	var s UploadSession
	row := db.QueryRow(ctx, `
		INSERT INTO upload_sessions (user_id, mode, s3_prefix, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, mode, status, s3_prefix, parts, expires_at, created_at
	`, ad.UserID, mode, prefix, now.Add(24*time.Hour))
	if err := scanSession(row, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

type PresignParams struct {
	Number int    `json:"number"`
	Key    string `json:"key"`
}

type PresignResponse struct {
	Number int    `json:"number"`
	Key    string `json:"key"`
	URL    string `json:"url"`
}

//encore:api auth method=POST path=/upload-sessions/:id/presign
func PresignUpload(ctx context.Context, id string, p *PresignParams) (*PresignResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var userID, prefix string
	err := db.QueryRow(ctx, `
		SELECT user_id, s3_prefix FROM upload_sessions
		WHERE id = $1 AND status = 'active' AND expires_at > now()
	`, id).Scan(&userID, &prefix)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "session not found or expired"}
		}
		return nil, err
	}
	if userID != ad.UserID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "not your session"}
	}

	objKey := prefix + "/" + p.Key
	url, err := ComicBucket.SignedUploadURL(ctx, objKey, objects.WithTTL(7200*time.Second))
	if err != nil {
		return nil, err
	}
	return &PresignResponse{Number: p.Number, Key: objKey, URL: url.URL}, nil
}

type ConfirmPartParams struct {
	Number int    `json:"number"`
	Key    string `json:"key"`
	Size   int64  `json:"size"`
	ETag   string `json:"etag"`
}

//encore:api auth method=POST path=/upload-sessions/:id/parts
func ConfirmPart(ctx context.Context, id string, p *ConfirmPartParams) (*UploadSession, error) {
	ad := auth.Data().(*myauth.AuthData)

	var userID string
	var partsBytes []byte
	err := db.QueryRow(ctx, `
		SELECT user_id, parts FROM upload_sessions
		WHERE id = $1 AND status = 'active' AND expires_at > now()
	`, id).Scan(&userID, &partsBytes)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "session not found or expired"}
		}
		return nil, err
	}
	if userID != ad.UserID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "not your session"}
	}

	var parts []Part
	scanParts(partsBytes, &parts)

	found := false
	for i, part := range parts {
		if part.Number == p.Number {
			parts[i] = Part{Number: p.Number, Key: p.Key, Size: p.Size, ETag: p.ETag}
			found = true
			break
		}
	}
	if !found {
		parts = append(parts, Part{Number: p.Number, Key: p.Key, Size: p.Size, ETag: p.ETag})
	}

	newBytes, _ := marshalParts(parts)
	_, err = db.Exec(ctx, `
		UPDATE upload_sessions SET parts = $1, updated_at = now()
		WHERE id = $2 AND user_id = $3
	`, newBytes, id, userID)
	if err != nil {
		return nil, err
	}

	var s UploadSession
	row := db.QueryRow(ctx, `
		SELECT id, user_id, mode, status, s3_prefix, parts, expires_at, created_at
		FROM upload_sessions WHERE id = $1
	`, id)
	if err := scanSession(row, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

//encore:api auth method=DELETE path=/upload-sessions/:id
func AbortSession(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)

	var userID string
	err := db.QueryRow(ctx, `SELECT user_id FROM upload_sessions WHERE id = $1`, id).Scan(&userID)
	if err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "session not found"}
		}
		return err
	}
	if userID != ad.UserID {
		return &errs.Error{Code: errs.PermissionDenied, Message: "not your session"}
	}

	_, err = db.Exec(ctx, `UPDATE upload_sessions SET status = 'failed' WHERE id = $1`, id)
	return err
}

//encore:api auth method=GET path=/upload-sessions/:id
func GetSession(ctx context.Context, id string) (*UploadSession, error) {
	ad := auth.Data().(*myauth.AuthData)

	var s UploadSession
	row := db.QueryRow(ctx, `
		SELECT id, user_id, mode, status, s3_prefix, parts, expires_at, created_at
		FROM upload_sessions WHERE id = $1
	`, id)
	if err := scanSession(row, &s); err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "session not found"}
		}
		return nil, err
	}
	if s.UserID != ad.UserID && ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "not your session"}
	}
	return &s, nil
}

type ListSessionsResponse struct {
	Sessions []UploadSession `json:"sessions"`
}

//encore:api auth method=GET path=/upload-sessions
func ListActiveSessions(ctx context.Context) (*ListSessionsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	rows, err := db.Query(ctx, `
		SELECT id, user_id, mode, status, s3_prefix, parts, expires_at, created_at
		FROM upload_sessions WHERE user_id = $1 AND status = 'active' AND expires_at > now()
		ORDER BY created_at DESC LIMIT 10
	`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []UploadSession
	for rows.Next() {
		var s UploadSession
		var partsBytes []byte
		if err := rows.Scan(&s.ID, &s.UserID, &s.Mode, &s.Status, &s.S3Prefix, &partsBytes, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		scanParts(partsBytes, &s.Parts)
		sessions = append(sessions, s)
	}

	return &ListSessionsResponse{Sessions: sessions}, rows.Err()
}
