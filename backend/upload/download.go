package upload

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/storage/objects"
)

// DownloadArchive serves a comic archive for download. Small files are
// streamed directly through the backend; files larger than the configured
// threshold are served via a presigned redirect to object storage.
//
//encore:api auth raw method=GET path=/download/*key
func DownloadArchive(w http.ResponseWriter, req *http.Request) {
	// auth endpoint — the auth handler rejects unauthenticated requests before
	// we get here, but keep a defensive check for raw handler safety.
	if auth.Data() == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	key := strings.TrimPrefix(req.URL.Path, "/download/")
	if key == "" {
		http.Error(w, "missing object key", http.StatusBadRequest)
		return
	}

	name := req.URL.Query().Get("name")
	if name == "" {
		name = filepath.Base(key)
	}
	size, _ := strconv.ParseInt(req.URL.Query().Get("size"), 10, 64)

	threshold := int64(10) << 20
	if cfg, err := myauth.GetAppConfig(req.Context()); err == nil && cfg.DownloadStreamThresholdMB > 0 {
		threshold = int64(cfg.DownloadStreamThresholdMB) << 20
	}

	// Large files (archives): redirect to a presigned URL so the browser pulls
	// directly from object storage instead of proxying through the backend.
	if size > threshold {
		u, err := ComicBucket.SignedDownloadURL(req.Context(), key, objects.WithTTL(3600))
		if err != nil {
			http.Error(w, "object not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, req, u.URL, http.StatusFound)
		return
	}

	// Small files: stream directly.
	r := ComicBucket.Download(req.Context(), key)
	if err := r.Err(); err != nil {
		r.Close()
		http.Error(w, "object not found", http.StatusNotFound)
		return
	}
	defer r.Close()

	w.Header().Set("Content-Type", "application/vnd.comicbook+zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	if _, err := io.Copy(w, r); err != nil {
		return
	}
}
