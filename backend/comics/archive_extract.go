package comics

import (
	"context"
	"strings"

	myupload "comics-galore/backend/upload"

	"encore.dev/pubsub"
)

// ArchiveExtractEvent is published after a comic is created with an archive
// format the browser can't extract client-side (CBR/RAR, PDF).
type ArchiveExtractEvent struct {
	ComicID  string `json:"comic_id"`
	FileKey  string `json:"file_key"`
	Mimetype string `json:"mimetype"`
}

var archiveExtractTopic = pubsub.NewTopic[ArchiveExtractEvent]("archive-extract", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var _ = pubsub.NewSubscription(archiveExtractTopic, "extract-pages", pubsub.SubscriptionConfig[ArchiveExtractEvent]{
	Handler:     handleArchiveExtract,
	RetryPolicy: &pubsub.RetryPolicy{MaxRetries: 5},
})

func handleArchiveExtract(ctx context.Context, ev ArchiveExtractEvent) error {
	return runTrackedJob(ctx, "archive-extract", ev.ComicID, func() error {
		return extractArchive(ctx, ev)
	})
}

func extractArchive(ctx context.Context, ev ArchiveExtractEvent) error {
	kind := detectArchiveKind(ev.Mimetype, ev.FileKey)
	if kind == "" {
		return nil
	}

	res, err := myupload.ExtractArchive(ctx, &myupload.ExtractArchiveParams{
		FileKey: ev.FileKey,
		Kind:    kind,
	})
	if err != nil {
		return err
	}

	if len(res.PageKeys) == 0 {
		db.Exec(ctx, `UPDATE comics SET extraction_status = 'failed' WHERE id = $1`, ev.ComicID)
		return nil
	}

	dims := make([]PageDimension, len(res.PageDimensions))
	for i, d := range res.PageDimensions {
		dims[i] = PageDimension{Width: d.Width, Height: d.Height}
	}

	keysJSON, _ := marshalStringSlice(res.PageKeys)
	dimsJSON, _ := marshalPageDimensions(dims)
	_, err = db.Exec(ctx, `
		UPDATE comics
		SET file_key = $1,
		    archive_mimetype = 'application/vnd.comicbook+zip',
		    page_keys = page_keys || $2::jsonb,
		    page_dimensions = page_dimensions || $3::jsonb,
		    page_count = jsonb_array_length(page_keys || $2::jsonb),
		    extraction_status = 'done'
		WHERE id = $4
	`, res.FileKey, string(keysJSON), string(dimsJSON), ev.ComicID)
	return err
}

// detectArchiveKind returns "rar" or "pdf" when the archive needs server-side
// extraction, otherwise "" (zip-family archives are handled client-side).
func detectArchiveKind(mimetype, fileKey string) string {
	m := strings.ToLower(mimetype)
	if strings.Contains(m, "pdf") {
		return "pdf"
	}
	if strings.Contains(m, "rar") || strings.Contains(m, "cbr") {
		return "rar"
	}
	lower := strings.ToLower(fileKey)
	if strings.HasSuffix(lower, ".pdf") {
		return "pdf"
	}
	if strings.HasSuffix(lower, ".cbr") || strings.HasSuffix(lower, ".rar") {
		return "rar"
	}
	return ""
}
