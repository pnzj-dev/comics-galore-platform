package upload

import (
	"context"
	"sort"
	"strings"

	myauth "comics-galore/backend/auth"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/images"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
)

// StorageBucket is an aggregate of object count + bytes for one key-prefix
// category (archives, extracted pages, covers/previews, seed data, other).
type StorageBucket struct {
	Prefix      string `json:"prefix"`
	ObjectCount int64  `json:"object_count"`
	TotalBytes  int64  `json:"total_bytes"`
}

type StorageUsageResponse struct {
	S3TotalBytes  int64           `json:"s3_total_bytes"`
	S3ObjectCount int64           `json:"s3_object_count"`
	Breakdown     []StorageBucket `json:"breakdown"`
	CFConfigured  bool            `json:"cf_configured"`
	CFImagesCount int             `json:"cf_images_count"`
}

// GetStorageUsage enumerates the comic-files bucket (object count + bytes,
// broken down by key prefix) and reports the Cloudflare Images count.
//
//encore:api auth method=GET path=/admin/storage
func GetStorageUsage(ctx context.Context) (*StorageUsageResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	resp := &StorageUsageResponse{Breakdown: []StorageBucket{}}

	type agg struct {
		count int64
		bytes int64
	}
	buckets := map[string]*agg{}

	for entry, err := range ComicBucket.List(ctx, &objects.Query{}) {
		if err != nil {
			return nil, err
		}
		if entry == nil {
			continue
		}
		resp.S3TotalBytes += entry.Size
		resp.S3ObjectCount++
		p := classifyPrefix(entry.Name)
		b := buckets[p]
		if b == nil {
			b = &agg{}
			buckets[p] = b
		}
		b.count++
		b.bytes += entry.Size
	}

	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		resp.Breakdown = append(resp.Breakdown, StorageBucket{
			Prefix:      k,
			ObjectCount: buckets[k].count,
			TotalBytes:  buckets[k].bytes,
		})
	}

	if cfClient != nil && secrets.CloudflareAccountID != "" {
		resp.CFConfigured = true
		if r, err := cfClient.Images.V2.List(ctx, images.V2ListParams{
			AccountID: cloudflare.F(secrets.CloudflareAccountID),
			PerPage:   cloudflare.F(float64(10000)),
		}); err == nil && r != nil {
			resp.CFImagesCount = len(r.Images)
		}
	}

	return resp, nil
}

func classifyPrefix(name string) string {
	switch {
	case strings.HasPrefix(name, "uploads/"):
		return "uploads (archives)"
	case strings.HasPrefix(name, "extracted/"):
		return "extracted (pages)"
	case strings.HasPrefix(name, "image-"):
		return "image (covers/previews)"
	case strings.HasPrefix(name, "cover-"):
		return "cover"
	case strings.HasPrefix(name, "seed/"):
		return "seed"
	default:
		return "other"
	}
}
