package comics

import (
	"context"
	"hash/fnv"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

// These private endpoints are the privileged surface for the key-gated MCP
// service. The mcp service verifies the actor's key and role before calling;
// here the explicit actorID is used only for audit-trail attribution.

type InternalModerateParams struct {
	ActorID string `json:"actor_id"`
	ComicID string `json:"comic_id"`
	Reason  string `json:"reason"`
}

//encore:api private method=POST path=/internal/comics/approve
func InternalApproveComic(ctx context.Context, p *InternalModerateParams) error {
	var status string
	if err := db.QueryRow(ctx, `SELECT status FROM comics WHERE id = $1`, p.ComicID).Scan(&status); err != nil {
		return &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	if status != "pending_review" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "comic is not pending review"}
	}
	if _, err := db.Exec(ctx, `UPDATE comics SET status = 'published', published_at = now(), updated_at = now() WHERE id = $1`, p.ComicID); err != nil {
		return err
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id) VALUES ($1, 'approve_comic', 'comic', $2)`, p.ActorID, p.ComicID)
	go notifyUploaderFollowers(context.Background(), p.ComicID)
	return nil
}

//encore:api private method=POST path=/internal/comics/reject
func InternalRejectComic(ctx context.Context, p *InternalModerateParams) error {
	_, err := db.Exec(ctx, `UPDATE comics SET status = 'rejected', rejection_reason = $1, updated_at = now() WHERE id = $2 AND status = 'pending_review'`, p.Reason, p.ComicID)
	if err != nil {
		return err
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'reject_comic', 'comic', $2, $3)`, p.ActorID, p.ComicID, `{"reason":"`+p.Reason+`"}`)
	return nil
}

type InternalResolveFlagParams struct {
	ActorID string `json:"actor_id"`
	FlagID  string `json:"flag_id"`
}

//encore:api private method=POST path=/internal/comics/resolve-flag
func InternalResolveCommentFlag(ctx context.Context, p *InternalResolveFlagParams) error {
	_, err := db.Exec(ctx, `UPDATE comment_flags SET status = 'resolved', resolved_at = now(), resolved_by = $1 WHERE id = $2 AND status = 'open'`, p.ActorID, p.FlagID)
	return err
}

type InternalCreateCommentParams struct {
	ActorID string `json:"actor_id"`
	ComicID string `json:"comic_id"`
	Body    string `json:"body"`
}

//encore:api private method=POST path=/internal/comics/create-comment
func InternalCreateComment(ctx context.Context, p *InternalCreateCommentParams) (*CommentData, error) {
	if p.Body == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body is required"}
	}
	var c CommentData
	err := db.QueryRow(ctx, `
		INSERT INTO comments (comic_id, user_id, body_text)
		VALUES ($1, $2, $3)
		RETURNING id, comic_id, user_id, COALESCE(parent_id::text, ''), body_text, created_at
	`, p.ComicID, p.ActorID, p.Body).Scan(&c.ID, &c.ComicID, &c.UserID, &c.ParentID, &c.BodyText, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	publishComment(p.ComicID, c)
	moderationTopic.Publish(ctx, ModerationEvent{TargetType: "comment", TargetID: c.ID})
	return &c, nil
}

// ----- Agentic creation / moderation (MCP-facing) -----

type InternalCreateComicParams struct {
	ActorID          string   `json:"actor_id"`
	Title            string   `json:"title"`
	Author           string   `json:"author"`
	Description      string   `json:"description"`
	ContentLanguage  string   `json:"content_language"`
	Category         string   `json:"category"`
	Genre            string   `json:"genre"`
	AgeRating        string   `json:"age_rating"`
	IsPremium        bool     `json:"is_premium"`
	Tags             []string `json:"tags"`
	ReadingDirection string   `json:"reading_direction"`
	CoverKey         string   `json:"cover_key"`
	PageKeys         []string `json:"page_keys"`
	Publish          bool     `json:"publish"`
	SeriesTitle      string   `json:"series_title"`
}

type InternalCreateComicResponse struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

// placeholderIndex derives a stable placeholder-art index from a string.
func placeholderIndex(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

//encore:api private method=POST path=/internal/comics/create
func InternalCreateComic(ctx context.Context, p *InternalCreateComicParams) (*InternalCreateComicResponse, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "title is required"}
	}
	if strings.TrimSpace(p.ActorID) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "actor_id is required"}
	}

	idx := placeholderIndex(title)
	coverKey := strings.TrimSpace(p.CoverKey)
	if coverKey == "" {
		coverKey = seedCoverKey(idx)
	}
	pageKeys := p.PageKeys
	if len(pageKeys) == 0 {
		pageKeys = seedPageKeys(idx)
	}

	lang := p.ContentLanguage
	if lang == "" {
		lang = "en"
	}
	ageRating := p.AgeRating
	if ageRating == "" {
		ageRating = "all_ages"
	}
	switch ageRating {
	case "all_ages", "teen", "mature", "explicit":
	default:
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid age_rating"}
	}
	readingDirection := p.ReadingDirection
	if readingDirection != "rtl" {
		readingDirection = "ltr"
	}
	genre := p.Genre
	if genre == "" {
		genre = "Comic"
	}
	category := p.Category
	if category == "" {
		category = "Comic"
	}

	slug := generateSlug(title)
	status := "pending_review"
	var publishedAt interface{}
	if p.Publish {
		status = "published"
		publishedAt = time.Now()
	}

	pageKeysJSON, _ := marshalStringSlice(pageKeys)
	pageDimsJSON, _ := marshalPageDimensions(seedPageDims(len(pageKeys)))
	tagsJSON, _ := marshalStringSlice(p.Tags)

	var out InternalCreateComicResponse
	err := db.QueryRow(ctx, `
		INSERT INTO comics (uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, page_keys, page_dimensions, page_count, reading_direction,
			file_size_bytes, age_rating, is_premium, tags, archive_mimetype, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, now(), now())
		RETURNING id, slug, status, title
	`, p.ActorID, title, p.Author, slug, p.Description, lang, status,
		category, genre, coverKey, "seed/"+slug+"/archive",
		pageKeysJSON, pageDimsJSON, len(pageKeys), readingDirection,
		1234567, ageRating, p.IsPremium, tagsJSON, "application/vnd.comicbook+zip", publishedAt).Scan(
		&out.ID, &out.Slug, &out.Status, &out.Title)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(p.SeriesTitle) != "" {
		if err := attachComicToSeries(ctx, p.ActorID, strings.TrimSpace(p.SeriesTitle), out.ID, genre, category, coverKey); err != nil {
			return nil, err
		}
	}

	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id) VALUES ($1, 'create_comic', 'comic', $2)`, p.ActorID, out.ID)
	moderationTopic.Publish(ctx, ModerationEvent{TargetType: "comic", TargetID: out.ID})
	return &out, nil
}

// attachComicToSeries find-or-creates a series by title and attaches the comic.
func attachComicToSeries(ctx context.Context, uploaderID, seriesTitle, comicID, genre, category, coverKey string) error {
	var seriesID string
	if err := db.QueryRow(ctx, `SELECT id FROM series WHERE LOWER(title) = LOWER($1) LIMIT 1`, seriesTitle).Scan(&seriesID); err != nil || seriesID == "" {
		slug := generateSlug(seriesTitle)
		if err := db.QueryRow(ctx, `
			INSERT INTO series (title, slug, description, genre, category, cover_key, uploader_id)
			VALUES ($1, $2, '', $3, $4, $5, $6)
			RETURNING id
		`, seriesTitle, slug, genre, category, coverKey, uploaderID).Scan(&seriesID); err != nil {
			return err
		}
	}

	var nextOrder int
	db.QueryRow(ctx, `SELECT COALESCE(MAX(series_order), 0) + 1 FROM comics WHERE series_id = $1`, seriesID).Scan(&nextOrder)
	db.Exec(ctx, `UPDATE comics SET series_id = $1, series_order = $2 WHERE id = $3`, seriesID, nextOrder, comicID)
	db.Exec(ctx, `
		UPDATE series s SET
			views_count  = COALESCE((SELECT SUM(c.view_count) FROM comics c WHERE c.series_id = s.id), 0),
			hearts_count = COALESCE((SELECT SUM(c.fav_count)  FROM comics c WHERE c.series_id = s.id), 0)
		WHERE s.id = $1
	`, seriesID)
	return nil
}

type InternalListFlagsParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

//encore:api private method=POST path=/internal/comics/list-flags
func InternalListFlaggedComments(ctx context.Context, p *InternalListFlagsParams) (*ListFlaggedCommentsResponse, error) {
	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comment_flags WHERE status = 'open'`).Scan(&total)

	rows, err := db.Query(ctx, `
		SELECT f.id, c.id, c.comic_id, co.title, c.user_id, c.body_text, COALESCE(f.reason, ''),
			(SELECT COUNT(*) FROM comment_flags fc WHERE fc.comment_id = c.id AND fc.status = 'open'),
			f.created_at
		FROM comment_flags f
		JOIN comments c ON c.id = f.comment_id
		JOIN comics co ON co.id = c.comic_id
		WHERE f.status = 'open'
		ORDER BY f.created_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []FlaggedComment
	for rows.Next() {
		var f FlaggedComment
		if err := rows.Scan(&f.FlagID, &f.CommentID, &f.ComicID, &f.ComicTitle, &f.UserID, &f.BodyText, &f.Reason, &f.FlagCount, &f.CreatedAt); err != nil {
			return nil, err
		}
		flags = append(flags, f)
	}
	if flags == nil {
		flags = []FlaggedComment{}
	}
	return &ListFlaggedCommentsResponse{Flags: flags, Total: total}, rows.Err()
}

type InternalDeleteCommentParams struct {
	ActorID   string `json:"actor_id"`
	CommentID string `json:"comment_id"`
}

//encore:api private method=POST path=/internal/comics/delete-comment
func InternalDeleteComment(ctx context.Context, p *InternalDeleteCommentParams) error {
	res, err := db.Exec(ctx, `DELETE FROM comments WHERE id = $1`, p.CommentID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return &errs.Error{Code: errs.NotFound, Message: "comment not found"}
	}
	db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id) VALUES ($1, 'delete_comment', 'comment', $2)`, p.ActorID, p.CommentID)
	return nil
}
