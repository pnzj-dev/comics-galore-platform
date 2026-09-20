package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	myauth "comics-galore/backend/auth"

	encoreauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

// getUsersInfo is a cross-service seam (overridable in tests) to enrich key
// listings with the bound user's public handle.
var getUsersInfo = func(ctx context.Context, p *myauth.GetUsersInfoParams) (*myauth.GetUsersInfoResponse, error) {
	return myauth.GetUsersInfo(ctx, p)
}

// requireAdmin gates the admin-only key management surface.
func requireAdmin(ctx context.Context) error {
	ad, ok := encoreauth.Data().(*myauth.AuthData)
	if !ok || ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	return nil
}

// McpKeyInfo is the admin-facing view of an MCP key (never the raw key).
type McpKeyInfo struct {
	ID        string     `json:"id"`
	Label     string     `json:"label"`
	KeySuffix string     `json:"key_suffix"`
	UserID    string     `json:"user_id"`
	Username  string     `json:"username,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type McpKeyListResponse struct {
	Keys []McpKeyInfo `json:"keys"`
}

//encore:api auth method=GET path=/admin/mcp/keys
func AdminListMcpKeys(ctx context.Context) (*McpKeyListResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx, `SELECT id, label, key_suffix, user_id, created_at, revoked_at FROM mcp_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []McpKeyInfo{}
	for rows.Next() {
		var k McpKeyInfo
		var revoked sql.NullTime
		if err := rows.Scan(&k.ID, &k.Label, &k.KeySuffix, &k.UserID, &k.CreatedAt, &revoked); err != nil {
			return nil, err
		}
		if revoked.Valid {
			k.RevokedAt = &revoked.Time
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []McpKeyInfo{}
	}
	enrichUsernames(ctx, keys)
	return &McpKeyListResponse{Keys: keys}, rows.Err()
}

func enrichUsernames(ctx context.Context, keys []McpKeyInfo) {
	ids := make([]string, 0, len(keys))
	seen := map[string]bool{}
	for _, k := range keys {
		if k.UserID != "" && !seen[k.UserID] {
			ids = append(ids, k.UserID)
			seen[k.UserID] = true
		}
	}
	if len(ids) == 0 {
		return
	}
	resp, err := getUsersInfo(ctx, &myauth.GetUsersInfoParams{IDs: ids})
	if err != nil {
		return
	}
	names := map[string]string{}
	for _, u := range resp.Users {
		names[u.ID] = u.Username
	}
	for i := range keys {
		keys[i].Username = names[keys[i].UserID]
	}
}

type CreateMcpKeyParams struct {
	Label  string `json:"label"`
	UserID string `json:"user_id"`
}

type CreateMcpKeyResponse struct {
	Key  string     `json:"key"`
	Info McpKeyInfo `json:"info"`
}

//encore:api auth method=POST path=/admin/mcp/keys
func AdminCreateMcpKey(ctx context.Context, p *CreateMcpKeyParams) (*CreateMcpKeyResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	ad := encoreauth.Data().(*myauth.AuthData)
	userID := strings.TrimSpace(p.UserID)
	if userID == "" {
		userID = ad.UserID
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	key := hex.EncodeToString(raw)
	suffix := key[len(key)-4:]
	hash := sha256.Sum256([]byte(key))

	var info McpKeyInfo
	var revoked sql.NullTime
	err := db.QueryRow(ctx, `
		INSERT INTO mcp_keys (key_hash, key_suffix, user_id, label)
		VALUES ($1, $2, $3, $4)
		RETURNING id, label, key_suffix, user_id, created_at, revoked_at
	`, hex.EncodeToString(hash[:]), suffix, userID, strings.TrimSpace(p.Label)).Scan(&info.ID, &info.Label, &info.KeySuffix, &info.UserID, &info.CreatedAt, &revoked)
	if err != nil {
		return nil, err
	}
	if revoked.Valid {
		info.RevokedAt = &revoked.Time
	}
	return &CreateMcpKeyResponse{Key: key, Info: info}, nil
}

//encore:api auth method=DELETE path=/admin/mcp/keys/:id
func AdminRevokeMcpKey(ctx context.Context, id string) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE mcp_keys SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	return err
}
