package upload

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"regexp"
	"strings"

	"encore.dev/storage/objects"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var seedImages []string
var proxyClient = &http.Client{}

func init() {
	seedImages = []string{
		"43fa19e1-5bbc-4865-3c5f-80dab3711200",
		"7ef3f0b7-c330-4302-96c3-3fe876cf0200",
		"7845d02b-f5b1-43b6-ff07-0002a3416100",
		"5504dd7d-2dbd-4e36-d337-0e2b27542600",
		"8cf41eb3-249e-4906-5824-cf31a866af00",
		"8328c47e-b4ec-43f0-997b-8321e7b96100",
		"cf2739fd-7ec2-44c8-bc47-47b31d8fe000",
		"fd535c8b-95fa-49e5-be59-02fd3be9f100",
		"0d90dacb-3868-4c71-2885-086cf63bd300",
	}
}

//encore:api public raw path=/media/:key method=GET
func ServeMedia(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if strings.HasPrefix(key, "seed/") {
		if len(seedImages) > 0 && cfSecrets.CloudflareImagesHash != "" {
			h := fnv.New32a()
			h.Write([]byte(key))
			idx := int(h.Sum32()) % len(seedImages)
			url := fmt.Sprintf("https://imagedelivery.net/%s/%s/public", cfSecrets.CloudflareImagesHash, seedImages[idx])
			proxyImage(w, req, url)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(placeholderSVG())
		return
	}

	if uuidRe.MatchString(key) && cfSecrets.CloudflareImagesHash != "" {
		url := fmt.Sprintf("https://imagedelivery.net/%s/%s/public", cfSecrets.CloudflareImagesHash, key)
		proxyImage(w, req, url)
		return
	}

	url, err := ComicBucket.SignedDownloadURL(req.Context(), key, objects.WithTTL(3600))
	if err != nil {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(placeholderSVG())
		return
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
