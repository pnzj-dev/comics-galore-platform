package upload

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
)

// InternalPresignUpload is the privileged surface for the key-gated MCP
// service: it returns a signed upload URL (and the object key) so an agent can
// PUT bytes directly to object storage, then reference the key when creating a
// comic. This keeps large/binary payloads out of the API path.
type InternalPresignUploadParams struct {
	ActorID  string `json:"actor_id"`
	Kind     string `json:"kind"`     // "cover" | "preview" | "archive"
	Filename string `json:"filename"` // optional; used for the key extension
}

type InternalPresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
}

//encore:api private method=POST path=/internal/presign-upload
func InternalPresignUpload(ctx context.Context, p *InternalPresignUploadParams) (*InternalPresignUploadResponse, error) {
	if strings.TrimSpace(p.ActorID) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "actor_id is required"}
	}

	kind := strings.TrimSpace(p.Kind)
	if kind != "cover" && kind != "preview" && kind != "archive" {
		kind = "preview"
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(p.Filename)))
	if ext == "" {
		if kind == "archive" {
			ext = ".cbz"
		} else {
			ext = ".jpg"
		}
	}
	if len(ext) > 10 || strings.ContainsAny(ext, "/\\") {
		ext = ".jpg"
	}

	key := kind + "s/" + randomHex(16) + ext

	ttl := 7200 * time.Second
	if cfg, err := myauth.GetAppConfig(ctx); err == nil && cfg.S3PresignedTTLMin > 0 {
		ttl = time.Duration(cfg.S3PresignedTTLMin) * time.Minute
	}
	u, err := ComicBucket.SignedUploadURL(ctx, key, objects.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	return &InternalPresignUploadResponse{UploadURL: u.URL, Key: key}, nil
}
