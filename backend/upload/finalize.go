package upload

import (
	"context"
	"io"
	"sort"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

// FinalizeMultipart concatenates a session's uploaded parts (in part-number
// order) into a single object under the session prefix, deletes the parts and
// returns the final object key. Used to reassemble archives that were split
// into parts for upload.
//
//encore:api auth method=POST path=/upload-sessions/:id/finalize
func FinalizeMultipart(ctx context.Context, id string, p *FinalizeMultipartParams) (*FinalizeMultipartResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "only uploaders can finalize uploads"}
	}

	var userID, prefix string
	var partsBytes []byte
	err := db.QueryRow(ctx, `
		SELECT user_id, s3_prefix, parts FROM upload_sessions
		WHERE id = $1 AND status = 'active' AND expires_at > now()
	`, id).Scan(&userID, &prefix, &partsBytes)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "session not found or expired"}
		}
		return nil, err
	}
	if userID != ad.UserID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "not your session"}
	}

	var parts []Part
	if err := scanParts(partsBytes, &parts); err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "session has no parts"}
	}

	sorted := append([]Part(nil), parts...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Number < sorted[j].Number })

	outputKey := prefix + "/" + p.OutputName

	w := ComicBucket.Upload(ctx, outputKey)
	for _, part := range sorted {
		r := ComicBucket.Download(ctx, part.Key)
		if err := r.Err(); err != nil {
			r.Close()
			w.Abort(err)
			return nil, err
		}
		if _, err := io.Copy(w, r); err != nil {
			r.Close()
			w.Abort(err)
			return nil, err
		}
		r.Close()
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	// Clean up the individual parts now that they are merged.
	for _, part := range sorted {
		_ = ComicBucket.Remove(ctx, part.Key)
	}

	db.Exec(ctx, `UPDATE upload_sessions SET status = 'completed' WHERE id = $1`, id)

	return &FinalizeMultipartResponse{Key: outputKey}, nil
}

type FinalizeMultipartParams struct {
	OutputName string `json:"output_name"`
}

type FinalizeMultipartResponse struct {
	Key string `json:"key"`
}
