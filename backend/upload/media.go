package upload

import (
	"net/http"
	"strings"

	"encore.dev/storage/objects"
)

//encore:api public raw path=/media/:key method=GET
func ServeMedia(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Handle seed/placeholder data — return a visible placeholder SVG
	if strings.HasPrefix(key, "seed/") {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(placeholderSVG())
		return
	}

	// Try S3 presigned download URL for non-seed keys
	url, err := ComicBucket.SignedDownloadURL(req.Context(), key, objects.WithTTL(3600))
	if err != nil {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(placeholderSVG())
		return
	}
	http.Redirect(w, req, url.URL, http.StatusTemporaryRedirect)
}

func placeholderSVG() []byte {
	return []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="400" height="600" viewBox="0 0 400 600">
  <rect width="400" height="600" fill="#f3f4f6"/>
  <rect x="80" y="120" width="240" height="320" rx="8" fill="#e5e7eb" stroke="#d1d5db" stroke-width="2"/>
  <path d="M120 160 L280 160" stroke="#9ca3af" stroke-width="3" stroke-linecap="round"/>
  <path d="M120 200 L280 200" stroke="#9ca3af" stroke-width="3" stroke-linecap="round"/>
  <path d="M120 240 L240 240" stroke="#9ca3af" stroke-width="3" stroke-linecap="round"/>
  <text x="200" y="480" text-anchor="middle" fill="#9ca3af" font-size="16" font-family="system-ui,sans-serif">Cover Image</text>
</svg>`)
}
