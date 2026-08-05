package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/images"
	"github.com/cloudflare/cloudflare-go/v4/option"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
)

var cfSecrets struct {
	CloudflareAccountID  string
	CloudflareAPIToken   string
	CloudflareImagesHash string
}

var cfClient *cloudflare.Client

func init() {
	if cfSecrets.CloudflareAPIToken != "" {
		cfClient = cloudflare.NewClient(option.WithAPIToken(cfSecrets.CloudflareAPIToken))
	}
}

type CloudflareUploadResponse struct {
	UploadURL string `json:"uploadURL"`
	ImageID   string `json:"imageID"`
}

//encore:api auth method=POST path=/media/cloudflare/upload-url
func CloudflarePresignedUpload(ctx context.Context) (*CloudflareUploadResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "only uploaders can upload"}
	}

	// If Cloudflare is not configured, fall back to S3 presigned URLs
	if cfClient == nil {
		return s3PresignedFallback(ctx)
	}

	resp, err := cfClient.Images.V2.DirectUploads.New(ctx, images.V2DirectUploadNewParams{
		AccountID: cloudflare.F(cfSecrets.CloudflareAccountID),
	})
	if err != nil {
		return s3PresignedFallback(ctx)
	}

	return &CloudflareUploadResponse{
		UploadURL: resp.UploadURL,
		ImageID:   resp.ID,
	}, nil
}

func s3PresignedFallback(ctx context.Context) (*CloudflareUploadResponse, error) {
	// Generate a unique S3 key and presigned upload URL
	key := "cover-" + randomHex(16)
	url, err := ComicBucket.SignedUploadURL(ctx, key, objects.WithTTL(7200))
	if err != nil {
		return nil, err
	}
	return &CloudflareUploadResponse{
		UploadURL: url.URL,
		ImageID:   key,
	}, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
