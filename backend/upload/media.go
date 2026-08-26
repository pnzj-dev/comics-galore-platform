package upload

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"regexp"
	"strings"

	myauth "comics-galore/backend/auth"

	"encore.dev/storage/objects"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var seedImages []string
var proxyClient = &http.Client{}

func init() {
	seedImages = []string{
		"cf2739fd-7ec2-44c8-bc47-47b31d8fe000",
		"0d90dacb-3868-4c71-2885-086cf63bd300",
		"7845d02b-f5b1-43b6-ff07-0002a3416100",
		"8328c47e-b4ec-43f0-997b-8321e7b96100",
	}
}

//encore:api public raw path=/media/*key method=GET
func ServeMedia(w http.ResponseWriter, req *http.Request) {
	// Encore raw endpoints are registered via httprouter, which does not expose
	// path params through req.PathValue. Extract the key from the URL path
	// ourselves; the *key wildcard matches keys containing slashes (e.g. the
	// backend upload path "uploads/<uid>/<date>/<file>").
	key := strings.TrimPrefix(req.URL.Path, "/media/")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if strings.HasPrefix(key, "seed/") {
		if len(seedImages) > 0 && secrets.CloudflareImagesHash != "" {
			h := fnv.New32a()
			h.Write([]byte(key))
			idx := int(h.Sum32()) % len(seedImages)
			url := fmt.Sprintf("https://imagedelivery.net/%s/%s/public", secrets.CloudflareImagesHash, seedImages[idx])
			proxyImage(w, req, url)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(placeholderSVG())
		return
	}

	mode := "direct"
	var cfg *myauth.AppSettings
	if c, err := myauth.GetAppConfig(req.Context()); err == nil {
		cfg = c
		mode = c.ImageServingMode
	}

	if uuidRe.MatchString(key) && mode == "cloudflare_images" && secrets.CloudflareImagesHash != "" {
		url := fmt.Sprintf("https://imagedelivery.net/%s/%s/public", secrets.CloudflareImagesHash, key)
		proxyImage(w, req, url)
		return
	}

	url, err := ComicBucket.SignedDownloadURL(req.Context(), key, objects.WithTTL(3600))
	if err != nil {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(placeholderSVG())
		return
	}

	// imgproxy mode: redirect to a signed imgproxy URL so the browser fetches
	// the CDN directly (the source is the S3 signed URL).
	if mode == "imgproxy" && cfg != nil && cfg.ImgproxyBaseURL != "" {
		if u := buildImgproxyURL(cfg.ImgproxyBaseURL, cfg.ImgproxyKey, cfg.ImgproxySalt, url.URL); u != "" {
			http.Redirect(w, req, u, http.StatusFound)
			return
		}
	}

	proxyImage(w, req, url.URL)
}

func proxyImage(w http.ResponseWriter, req *http.Request, url string) {
	fetchReq, err := http.NewRequestWithContext(req.Context(), "GET", url, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := proxyClient.Do(fetchReq)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
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
