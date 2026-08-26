package upload

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// buildImgproxyURL returns a signed imgproxy URL for the given source image.
// key/salt are hex-encoded (imgproxy IMGPROXY_KEY / IMGPROXY_SALT convention);
// the signature is the URL-safe base64 HMAC-SHA256 of (salt || path).
//
// Processing uses rs:fit:2000:2000 (fit within 2000x2000, no upscale) so comic
// pages are CDN-cached without destroying zoom detail; tune later if needed.
func buildImgproxyURL(baseURL, keyHex, saltHex, sourceURL string) string {
	if baseURL == "" || keyHex == "" || saltHex == "" {
		return ""
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return ""
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return ""
	}

	path := "/rs:fit:2000:2000/" + base64.RawURLEncoding.EncodeToString([]byte(sourceURL))

	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	mac.Write([]byte(path))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return strings.TrimRight(baseURL, "/") + "/" + sig + path
}
