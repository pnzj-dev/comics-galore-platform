package upload

import (
	"net/http"
	"strings"

	"encore.dev/beta/auth"
	"encore.dev/storage/objects"
)

// DownloadArchive serves a comic archive for download via a presigned redirect,
// so the browser pulls bytes directly from object storage (R2) rather than
// proxying them through the backend.
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

	u, err := ComicBucket.SignedDownloadURL(req.Context(), key, objects.WithTTL(3600))
	if err != nil {
		http.Error(w, "object not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, req, u.URL, http.StatusFound)
}
