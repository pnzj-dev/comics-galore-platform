package upload

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/images"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
)

// UploadImage streams an uploaded cover/preview image to Cloudflare Images and
// returns its ID. Falls back to S3 when Cloudflare is not configured or the
// upload fails.
//encore:api auth raw method=POST path=/upload/image
func UploadImage(w http.ResponseWriter, req *http.Request) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		errs.HTTPError(w, &errs.Error{Code: errs.PermissionDenied, Message: "only uploaders can upload"})
		return
	}

	// Cover/preview images are small; buffer once so we can retry Cloudflare and
	// fall back to S3 without re-reading the body.
	body, err := io.ReadAll(req.Body)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	if cfClient != nil && secrets.CloudflareAccountID != "" {
		resp, err := cfClient.Images.V1.New(req.Context(), images.V1NewParams{
			AccountID: cloudflare.F(secrets.CloudflareAccountID),
			File:      cloudflare.F(io.Reader(bytes.NewReader(body))),
		})
		if err == nil && resp != nil && resp.ID != "" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"key": resp.ID})
			return
		}
	}

	key := "image-" + randomHex(16)
	writer := ComicBucket.Upload(req.Context(), key)
	if _, err := io.Copy(writer, bytes.NewReader(body)); err != nil {
		writer.Abort(err)
		errs.HTTPError(w, err)
		return
	}
	if err := writer.Close(); err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
}

// UploadFile streams an uploaded archive/page file to S3 via multipart and
// returns the object key (preserving the uploaded filename).
//encore:api auth raw method=POST path=/upload/file
func UploadFile(w http.ResponseWriter, req *http.Request) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		errs.HTTPError(w, &errs.Error{Code: errs.PermissionDenied, Message: "only uploaders can upload"})
		return
	}

	reader, err := req.MultipartReader()
	if err != nil {
		errs.HTTPError(w, err)
		return
	}
	part, err := reader.NextPart()
	if err != nil {
		errs.HTTPError(w, &errs.Error{Code: errs.InvalidArgument, Message: "no file part found"})
		return
	}

	filename := part.FileName()
	if filename == "" {
		filename = "upload-" + randomHex(8)
	}
	key := "uploads/" + ad.UserID + "/" + time.Now().Format("2006/01/02/150405") + "/" + filename

	writer := ComicBucket.Upload(req.Context(), key)
	if _, err := io.Copy(writer, part); err != nil {
		writer.Abort(err)
		errs.HTTPError(w, err)
		return
	}
	if err := writer.Close(); err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
}
