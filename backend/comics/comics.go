package comics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"
	"comics-galore/backend/turnstile"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("comicsdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var (
	commentMu      sync.Mutex
	commentChannels = make(map[string][]chan CommentData)
)

func subscribeComments(comicID string) chan CommentData {
	commentMu.Lock()
	defer commentMu.Unlock()
	ch := make(chan CommentData, 8)
	commentChannels[comicID] = append(commentChannels[comicID], ch)
	return ch
}

func unsubscribeComments(comicID string, ch chan CommentData) {
	commentMu.Lock()
	defer commentMu.Unlock()
	list := commentChannels[comicID]
	for i, c := range list {
		if c == ch {
			commentChannels[comicID] = append(list[:i], list[i+1:]...)
			break
		}
	}
}

func publishComment(comicID string, c CommentData) {
	commentMu.Lock()
	defer commentMu.Unlock()
	for _, ch := range commentChannels[comicID] {
		select {
		case ch <- c:
		default:
		}
	}
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Comic is the public representation
type Comic struct {
	ID              string    `json:"id"`
	UploaderID      string    `json:"uploader_id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	ContentLanguage string    `json:"content_language"`
	Status          string    `json:"status"`
	Category        string    `json:"category,omitempty"`
	Genre           string    `json:"genre,omitempty"`
	CoverKey        string    `json:"cover_key"`
	FileKey         string    `json:"file_key"`
	FileSizeBytes   int64     `json:"file_size_bytes"`
	MinTierID       string    `json:"min_tier_id,omitempty"`
	AgeRating       string    `json:"age_rating"`
	IsPremium       bool      `json:"is_premium"`
	Tags            []string  `json:"tags"`
	RejectionReason string    `json:"rejection_reason,omitempty"`
	PublishedAt     time.Time `json:"published_at,omitempty"`
	ViewCount       int64     `json:"view_count"`
	DownloadCount   int64     `json:"download_count"`
	LikeCount       int       `json:"like_count"`
	FavCount        int       `json:"fav_count"`
	DislikeCount    int       `json:"dislike_count"`
	IsLiked         bool      `json:"is_liked"`
	IsFavorited     bool      `json:"is_favorited"`
	IsDisliked      bool      `json:"is_disliked"`
	SeriesOrder     int       `json:"series_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Identifiers (single-issue vs graphic-novel vs periodical; at most one set)
	Isbn string `json:"isbn,omitempty"`
	Upc  string `json:"upc,omitempty"`
	Issn string `json:"issn,omitempty"`

	// Optional volume (collected edition / chapter) and issue number.
	Volume      string `json:"volume,omitempty"`
	IssueNumber string `json:"issue_number,omitempty"`

	// Reader support
	ReadingDirection string `json:"reading_direction"`
	PageCount        int    `json:"page_count"`
	ArchiveMimetype  string `json:"archive_mimetype,omitempty"`

	// ExtractionStatus tracks server-side archive extraction (CBR/PDF):
	// none | processing | done | failed.
	ExtractionStatus string `json:"extraction_status,omitempty"`

	// Resolved cover URL (populated server-side, not from DB)
	CoverURL string `json:"cover_url,omitempty"`

	// MatureLocked marks a comic whose pages are withheld from the caller
	// (free users when `forbid_mature_for_free` is enabled). The cover stays
	// present so the frontend can render a blurred teaser.
	MatureLocked bool `json:"mature_locked,omitempty"`

	// CommentsEnabled reflects the global `enable_comments` setting, so the
	// frontend can hide the comment form when commenting is disabled.
	CommentsEnabled bool `json:"comments_enabled,omitempty"`
}

type CreateComicParams struct {
	Title           string   `json:"title"`
	Author          string   `json:"author"`
	Description     string   `json:"description"`
	ContentLanguage string   `json:"content_language"`
	Category        string   `json:"category"`
	Genre           string   `json:"genre"`
	CoverKey        string   `json:"cover_key"`
	FileKey         string   `json:"file_key"`
	PageKeys        []string `json:"page_keys"`
	FileSizeBytes   int64    `json:"file_size_bytes"`
	MinTierID       string   `json:"min_tier_id"`
	AgeRating       string   `json:"age_rating"`
	IsPremium       bool     `json:"is_premium"`
	Tags            []string `json:"tags"`
	UploadSessionID string   `json:"upload_session_id"`

	// Identifiers (at most one set per comic)
	Isbn string `json:"isbn"`
	Upc  string `json:"upc"`
	Issn string `json:"issn"`

	// Optional volume (collected edition / chapter) and issue number.
	Volume      string `json:"volume"`
	IssueNumber string `json:"issue_number"`

	// Reader support
	ReadingDirection string          `json:"reading_direction"`
	PageDimensions   []PageDimension `json:"page_dimensions"`
	ArchiveMimetype  string          `json:"archive_mimetype"`

	// Series association: either attach to an existing series or create a new
	// one (title). SeriesGenre/Category/ScheduleDay are used only when a new
	// series is created.
	SeriesID         string `json:"series_id"`
	SeriesTitle      string `json:"series_title"`
	SeriesGenre      string `json:"series_genre"`
	SeriesCategory   string `json:"series_category"`
	SeriesScheduleDay string `json:"series_schedule_day"`

	TurnstileToken string `json:"turnstile_token"`
}

// PageDimension is one page's pixel dimensions, aligned by index with
// page_keys, used for layout-shift-free rendering.
type PageDimension struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

//encore:api auth method=POST path=/comics
func CreateComic(ctx context.Context, p *CreateComicParams) (*Comic, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "only uploaders can create comics",
		}
	}

	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "comic_upload"}); err != nil {
		return nil, err
	}

	if p.Title == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "title is required"}
	}
	if p.CoverKey == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "cover_key is required"}
	}
	if p.FileKey == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "file_key is required"}
	}

	// Enforce the global maximum archive size when configured.
	if cfg, err := myauth.GetAppConfig(ctx); err == nil && cfg.MaxUploadSizeMB > 0 {
		maxBytes := int64(cfg.MaxUploadSizeMB) * 1024 * 1024
		if p.FileSizeBytes > maxBytes {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: fmt.Sprintf("archive exceeds the %d MB upload limit", cfg.MaxUploadSizeMB),
			}
		}
	}

	// Validate age_rating against the allowed enum (avoids a DB CHECK 500).
	switch p.AgeRating {
	case "", "all_ages", "teen", "mature", "explicit":
	default:
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid age_rating"}
	}

	lang := p.ContentLanguage
	if lang == "" {
		lang = "en"
		if cfg, err := myauth.GetAppConfig(ctx); err == nil && cfg.DefaultContentLang != "" {
			lang = cfg.DefaultContentLang
		}
	}

	ageRating := p.AgeRating
	if ageRating == "" {
		ageRating = "all_ages"
	}

	slug := generateSlug(p.Title)

	var comic Comic
	var pageKeys []byte
	if p.PageKeys == nil {
		pageKeys = []byte("[]")
	} else {
		pageKeys, _ = marshalStringSlice(p.PageKeys)
	}

	pageDims, err := marshalPageDimensions(p.PageDimensions)
	if err != nil {
		return nil, err
	}

	readingDirection := p.ReadingDirection
	if readingDirection != "rtl" {
		readingDirection = "ltr"
	}

	pageCount := len(p.PageKeys)

	var tags []byte
	if p.Tags == nil {
		tags = []byte("[]")
	} else {
		tags, _ = marshalStringSlice(p.Tags)
	}

	var minTierID interface{}
	if p.MinTierID == "" {
		minTierID = nil
	} else {
		minTierID = p.MinTierID
	}

	err = db.QueryRow(ctx, `
		INSERT INTO comics (uploader_id, title, author, slug, description, content_language,
			category, genre, cover_key, file_key, page_keys, page_dimensions, page_count, reading_direction,
			archive_mimetype, isbn, upc, issn, volume, issue_number,
			file_size_bytes, min_tier_id, age_rating, is_premium, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		RETURNING id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			extraction_status, created_at, updated_at
	`, ad.UserID, p.Title, p.Author, slug, p.Description, lang,
		p.Category, p.Genre, p.CoverKey, p.FileKey, pageKeys, pageDims, pageCount, readingDirection,
		p.ArchiveMimetype, nulOrValue(p.Isbn), nulOrValue(p.Upc), nulOrValue(p.Issn), nulOrValue(p.Volume), nulOrValue(p.IssueNumber),
		p.FileSizeBytes, minTierID, ageRating, p.IsPremium, tags).Scan(
		&comic.ID, &comic.UploaderID, &comic.Title, &comic.Author, &comic.Slug, &comic.Description,
		&comic.ContentLanguage, &comic.Status, &comic.Category, &comic.Genre, &comic.CoverKey, &comic.FileKey,
		&comic.FileSizeBytes, nulString(&comic.MinTierID),
		&comic.AgeRating, &comic.IsPremium, scanStringSlice(&comic.Tags), nulString(&comic.RejectionReason),
		&comic.ViewCount, &comic.DownloadCount, &comic.LikeCount, &comic.FavCount, &comic.DislikeCount,
		&comic.ReadingDirection, &comic.PageCount, nulString(&comic.ArchiveMimetype),
		nulString(&comic.Isbn), nulString(&comic.Upc), nulString(&comic.Issn), nulString(&comic.Volume), nulString(&comic.IssueNumber),
		&comic.ExtractionStatus, &comic.CreatedAt, &comic.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Series association: attach to an existing series or create a new one.
	seriesID, err := resolveSeriesForComic(ctx, ad.UserID, p, &comic)
	if err != nil {
		return nil, err
	}
	if seriesID != "" {
		var nextOrder int
		db.QueryRow(ctx, `SELECT COALESCE(MAX(series_order), 0) + 1 FROM comics WHERE series_id = $1`, seriesID).Scan(&nextOrder)
		db.Exec(ctx, `UPDATE comics SET series_id = $1, series_order = $2 WHERE id = $3`, seriesID, nextOrder, comic.ID)
		db.Exec(ctx, `
			UPDATE series s SET
				views_count  = COALESCE((SELECT SUM(c.view_count) FROM comics c WHERE c.series_id = s.id), 0),
				hearts_count = COALESCE((SELECT SUM(c.fav_count)  FROM comics c WHERE c.series_id = s.id), 0)
			WHERE s.id = $1
		`, seriesID)
	}

	if p.UploadSessionID != "" {
		db.Exec(ctx, `UPDATE upload_sessions SET status = 'completed' WHERE id = $1 AND user_id = $2`,
			p.UploadSessionID, ad.UserID)
	}

	moderationTopic.Publish(ctx, ModerationEvent{TargetType: "comic", TargetID: comic.ID})

	// CBR/RAR and PDF archives can't be extracted client-side; kick off
	// server-side page extraction.
	if detectArchiveKind(p.ArchiveMimetype, p.FileKey) != "" {
		db.Exec(ctx, `UPDATE comics SET extraction_status = 'processing' WHERE id = $1`, comic.ID)
		archiveExtractTopic.Publish(ctx, ArchiveExtractEvent{
			ComicID:  comic.ID,
			FileKey:  p.FileKey,
			Mimetype: p.ArchiveMimetype,
		})
	}

	return &comic, nil
}

type ListComicsParams struct {
	Page          int    `query:"page"`
	Limit         int    `query:"limit"`
	Language      string `query:"language"`
	Search        string `query:"search"`
	SearchField   string `query:"search_field"`
	Tag           string `query:"tag"`
	Sort          string `query:"sort"`
	ExcludeMature string `query:"exclude_mature"`
}

type ListComicsResponse struct {
	Comics []Comic `json:"comics"`
	Total  int     `json:"total"`
}

//encore:api public method=GET path=/comics
func ListComics(ctx context.Context, p *ListComicsParams) (*ListComicsResponse, error) {
	page := p.Page
	if page < 1 {
		page = 1
	}
	limit := p.Limit
	if limit < 1 || limit > 50 {
		limit = defaultPageSize(ctx, 20)
	}
	offset := (page - 1) * limit

	var where string
	var args []interface{}
	args = append(args, "published")

	if p.Language != "" {
		where = "WHERE status = $1 AND content_language = $" + nextIdx(len(args)+1)
		args = append(args, p.Language)
	} else {
		where = "WHERE status = $1"
	}

	if p.Search != "" {
		pattern := "%" + p.Search + "%"
		switch p.SearchField {
		case "title":
			where += " AND title ILIKE $" + nextIdx(len(args)+1)
			args = append(args, pattern)
		case "description":
			where += " AND description ILIKE $" + nextIdx(len(args)+1)
			args = append(args, pattern)
		case "author":
			where += " AND author ILIKE $" + nextIdx(len(args)+1)
			args = append(args, pattern)
		default:
			where += " AND (title ILIKE $" + nextIdx(len(args)+1) + " OR author ILIKE $" + nextIdx(len(args)+2) + " OR description ILIKE $" + nextIdx(len(args)+3) + ")"
			args = append(args, pattern, pattern, pattern)
		}
	}

	if p.Tag != "" {
		where += " AND tags @> $" + nextIdx(len(args)+1) + "::jsonb"
		args = append(args, `["`+p.Tag+`"]`)
	}

	if p.ExcludeMature == "true" {
		where += " AND age_rating NOT IN (" + nextIdx(len(args)+1) + ", " + nextIdx(len(args)+2) + ")"
		args = append(args, "mature", "explicit")
	}

	// Free/anonymous users never see mature content when the policy forbids it.
	where += matureWhereClause(ctx)

	var total int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM comics `+where, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	queryArgs := append(args, limit, offset)
	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	var orderBy string
	switch p.Sort {
	case "popular":
		orderBy = "view_count DESC"
	case "newest":
		orderBy = "published_at DESC"
	case "random":
		orderBy = "RANDOM()"
	default:
		orderBy = "published_at DESC"
	}

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			created_at, updated_at
		FROM comics `+where+`
		ORDER BY `+orderBy+`
		LIMIT $`+nextIdx(limitIdx)+` OFFSET $`+nextIdx(offsetIdx), queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil {
		return nil, err
	}

	if ad, ok := getAuthData(ctx); ok {
		enrichReactions(ctx, comics, string(ad.UserID))
	}

	return &ListComicsResponse{Comics: comics, Total: total}, nil
}

// ----- Language facets -----

type LanguageFacet struct {
	Language string `json:"language"`
	Count    int    `json:"count"`
}

type LanguageFacetsResponse struct {
	Facets []LanguageFacet `json:"facets"`
}

//encore:api public method=GET path=/comics-language-facets
func LanguageFacets(ctx context.Context) (*LanguageFacetsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT COALESCE(NULLIF(content_language, ''), 'en'), COUNT(*)
		FROM comics WHERE status = 'published'
		GROUP BY content_language
		ORDER BY COUNT(*) DESC, content_language ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facets []LanguageFacet
	for rows.Next() {
		var f LanguageFacet
		if err := rows.Scan(&f.Language, &f.Count); err != nil {
			return nil, err
		}
		facets = append(facets, f)
	}
	if facets == nil {
		facets = []LanguageFacet{}
	}
	return &LanguageFacetsResponse{Facets: facets}, rows.Err()
}

// ----- Popular tags -----

type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type TagCountsResponse struct {
	Tags []TagCount `json:"tags"`
}

//encore:api public method=GET path=/tags
func PopularTags(ctx context.Context) (*TagCountsResponse, error) {
	limit := 20
	if ad, ok := getAuthData(ctx); ok {
		if prefs, err := myauth.GetUserPreferences(ctx, ad.UserID); err == nil && prefs.PopularTagsLimit > 0 {
			limit = prefs.PopularTagsLimit
		}
	}
	if limit == 20 {
		if cfg, err := myauth.GetAppConfig(ctx); err == nil && cfg.PopularTagsLimit > 0 {
			limit = cfg.PopularTagsLimit
		}
	}
	rows, err := db.Query(ctx, `
		SELECT tag, COUNT(*) AS count
		FROM comics, jsonb_array_elements_text(COALESCE(tags, '[]'::jsonb)) AS tag
		WHERE status = 'published'
		GROUP BY tag
		ORDER BY count DESC, tag ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagCount
	for rows.Next() {
		var t TagCount
		if err := rows.Scan(&t.Tag, &t.Count); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []TagCount{}
	}
	return &TagCountsResponse{Tags: tags}, rows.Err()
}

//encore:api public method=GET path=/comics/:slug
func GetComic(ctx context.Context, slug string) (*Comic, error) {
	ad, hasAuth := getAuthData(ctx)

	var comic Comic
	err := db.QueryRow(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			extraction_status, created_at, updated_at
		FROM comics WHERE slug = $1
	`, slug).Scan(
		&comic.ID, &comic.UploaderID, &comic.Title, &comic.Author, &comic.Slug, &comic.Description,
		&comic.ContentLanguage, &comic.Status, &comic.Category, &comic.Genre, &comic.CoverKey, &comic.FileKey,
		&comic.FileSizeBytes, nulString(&comic.MinTierID),
		&comic.AgeRating, &comic.IsPremium, scanStringSlice(&comic.Tags), nulString(&comic.RejectionReason),
		&comic.ViewCount, &comic.DownloadCount, &comic.LikeCount, &comic.FavCount, &comic.DislikeCount,
		&comic.ReadingDirection, &comic.PageCount, nulString(&comic.ArchiveMimetype),
		nulString(&comic.Isbn), nulString(&comic.Upc), nulString(&comic.Issn), nulString(&comic.Volume), nulString(&comic.IssueNumber),
		&comic.ExtractionStatus, &comic.CreatedAt, &comic.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
		}
		return nil, err
	}

	if comic.Status != "published" {
		if !hasAuth || (ad.UserID != comic.UploaderID && ad.Role != "admin" && ad.Role != "moderator") {
			return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
		}
	}

	db.Exec(ctx, `UPDATE comics SET view_count = view_count + 1 WHERE id = $1`, comic.ID)
	db.Exec(ctx, `UPDATE series SET views_count = views_count + 1 WHERE id = (SELECT series_id FROM comics WHERE id = $1)`, comic.ID)

	if hasAuth {
		enrichReactions(ctx, []Comic{comic}, string(ad.UserID))
	}

	resolveComicURLs(&comic)

	// Free/anonymous users: flag the comic so the frontend shows a blurred
	// teaser; the pages endpoint refuses to serve them.
	if isMatureRating(comic.AgeRating) && matureBlocked(ctx) {
		comic.MatureLocked = true
	}

	if policy, err := myauth.GetContentPolicy(ctx); err == nil {
		comic.CommentsEnabled = policy.EnableComments
	}

	return &comic, nil
}

// ----- Pages (paginated for the reader) -----

// ComicPage is one page of a comic with its resolved URL and dimensions.
type ComicPage struct {
	Index  int    `json:"index"`
	Key    string `json:"key"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	// Locked marks a page withheld from the caller (free-tier previews past the
	// preview limit). Locked pages have no URL or key.
	Locked bool `json:"locked,omitempty"`
}

type GetComicPagesParams struct {
	Offset  int  `query:"offset"`
	Limit   int  `query:"limit"`
	Preview bool `query:"preview"`
}

type GetComicPagesResponse struct {
	Total  int         `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
	Pages  []ComicPage `json:"pages"`
}

//encore:api public method=GET path=/comics/:slug/pages
func GetComicPages(ctx context.Context, slug string, p *GetComicPagesParams) (*GetComicPagesResponse, error) {
	ad, hasAuth := getAuthData(ctx)

	limit := p.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	var (
		status     string
		uploaderID string
		ageRating  string
		keys       []string
		dims       []PageDimension
	)
	err := db.QueryRow(ctx, `
		SELECT status, uploader_id, age_rating, page_keys, page_dimensions
		FROM comics WHERE slug = $1
	`, slug).Scan(&status, &uploaderID, &ageRating, scanStringSlice(&keys), scanPageDimensions(&dims))
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
		}
		return nil, err
	}

	if status != "published" {
		if !hasAuth || (ad.UserID != uploaderID && ad.Role != "admin" && ad.Role != "moderator") {
			return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
		}
	}

	if isMatureRating(ageRating) && matureBlocked(ctx) {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "mature content requires a paid subscription"}
	}

	// Determine whether the caller is on the free tier (or anonymous) for
	// gating. Staff (admin/moderator/uploader) and paid tiers are exempt.
	isFree := true
	if hasAuth {
		switch ad.Role {
		case "admin", "moderator", "uploader":
			isFree = false
		}
		if ad.Tier != "" && ad.Tier != "free" {
			isFree = false
		}
	}

	// The web reader (non-preview) requires a paid tier (Bronze or above);
	// free and anonymous callers are denied.
	if !p.Preview && isFree {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "the web reader requires a paid subscription (Bronze or above)"}
	}

	total := len(keys)
	if offset >= total {
		return &GetComicPagesResponse{Total: total, Offset: offset, Limit: limit, Pages: []ComicPage{}}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	// Free/anonymous preview gating: when a caller requests a preview (not the
	// reader's windowed fetch) and is on the free tier, withhold full page URLs
	// beyond a small limit and mark them locked (no sharp URL leaks). Paid tiers
	// and staff are never gated.
	const freePreviewLimit = 3
	gated := p.Preview && isFree

	urls := resolvePageURLs(keys[offset:end])
	pages := make([]ComicPage, 0, len(urls))
	for i, u := range urls {
		idx := offset + i
		page := ComicPage{Index: idx, Key: keys[idx], URL: u}
		if idx < len(dims) {
			page.Width = dims[idx].Width
			page.Height = dims[idx].Height
		}
		if gated && idx >= freePreviewLimit {
			page.Locked = true
			page.URL = ""
			page.Key = ""
		}
		pages = append(pages, page)
	}

	return &GetComicPagesResponse{Total: total, Offset: offset, Limit: limit, Pages: pages}, nil
}

// ComicMaturity is the minimal view the reading service needs to gate downloads.
type ComicMaturity struct {
	AgeRating string `json:"age_rating"`
}

//encore:api private method=GET path=/comics-maturity/:id
func GetComicMaturity(ctx context.Context, id string) (*ComicMaturity, error) {
	var m ComicMaturity
	err := db.QueryRow(ctx, `SELECT age_rating FROM comics WHERE id = $1`, id).Scan(&m.AgeRating)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
		}
		return nil, err
	}
	return &m, nil
}

type BatchComicsParams struct {
	IDs []string `json:"ids"`
}

//encore:api public method=POST path=/comics-batch
func BatchGetComics(ctx context.Context, p *BatchComicsParams) (*ListComicsResponse, error) {
	if len(p.IDs) == 0 {
		return &ListComicsResponse{Comics: []Comic{}, Total: 0}, nil
	}
	if len(p.IDs) > 50 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "max 50 ids"}
	}

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			created_at, updated_at
		FROM comics WHERE id = ANY($1) AND status = 'published'`+matureWhereClause(ctx)+`
		ORDER BY created_at DESC
	`, p.IDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil {
		return nil, err
	}

	if ad, ok := getAuthData(ctx); ok {
		enrichReactions(ctx, comics, string(ad.UserID))
	}

	return &ListComicsResponse{Comics: comics, Total: len(comics)}, nil
}

// ----- Favorites -----

type ListFavoritesParams struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

//encore:api auth method=GET path=/favorites
func ListFavorites(ctx context.Context, p *ListFavoritesParams) (*ListComicsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 50 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM favorites f
		JOIN comics c ON c.id = f.comic_id
		WHERE f.user_id = $1 AND c.status = 'published'`+matureWhereClause(ctx)+`
	`, ad.UserID).Scan(&total)

	rows, err := db.Query(ctx, `
		SELECT c.id, c.uploader_id, c.title, c.author, c.slug, c.description, c.content_language, c.status,
			c.category, c.genre, c.cover_key, c.file_key, c.file_size_bytes, c.min_tier_id, c.age_rating, c.is_premium,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count, c.dislike_count,
			c.reading_direction, c.page_count, c.archive_mimetype, c.isbn, c.upc, c.issn, c.volume, c.issue_number,
			c.created_at, c.updated_at
		FROM favorites f
		JOIN comics c ON c.id = f.comic_id
		WHERE f.user_id = $1 AND c.status = 'published'`+matureWhereClause(ctx)+`
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`, ad.UserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil {
		return nil, err
	}

	enrichReactions(ctx, comics, string(ad.UserID))

	return &ListComicsResponse{Comics: comics, Total: total}, nil
}

//encore:api auth method=GET path=/uploader/comics
func MyComics(ctx context.Context) (*ListComicsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			created_at, updated_at
		FROM comics WHERE uploader_id = $1
		ORDER BY created_at DESC
	`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil {
		return nil, err
	}

	enrichReactions(ctx, comics, string(ad.UserID))

	return &ListComicsResponse{Comics: comics, Total: len(comics)}, nil
}

// ----- Image URL Resolution -----

var seedCoverImages = []string{
	"cf2739fd-7ec2-44c8-bc47-47b31d8fe000",
	"0d90dacb-3868-4c71-2885-086cf63bd300",
	"7845d02b-f5b1-43b6-ff07-0002a3416100",
	"8328c47e-b4ec-43f0-997b-8321e7b96100",
}

var uuidLike = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var cfDeliveryHash = "HSI9oWWzl51z3qob-AZdpA"

func resolveCoverURL(key string) string {
	if key == "" {
		return ""
	}
	if strings.HasPrefix(key, "seed/") && len(seedCoverImages) > 0 {
		h := fnv.New32a()
		h.Write([]byte(key))
		idx := int(h.Sum32()) % len(seedCoverImages)
		return fmt.Sprintf("https://imagedelivery.net/%s/%s/public", cfDeliveryHash, seedCoverImages[idx])
	}
	if uuidLike.MatchString(key) {
		return fmt.Sprintf("https://imagedelivery.net/%s/%s/public", cfDeliveryHash, key)
	}
	return fmt.Sprintf("http://localhost:4000/media/%s", key)
}

func resolvePageURLs(keys []string) []string {
	urls := make([]string, len(keys))
	for i, k := range keys {
		urls[i] = resolveCoverURL(k)
	}
	return urls
}

func resolveComicURLs(c *Comic) {
	c.CoverURL = resolveCoverURL(c.CoverKey)
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugRe.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "comic"
	}
	return slug + "-" + randomSuffix(6)
}

func enrichReactions(ctx context.Context, comics []Comic, userID string) {
	if userID == "" || len(comics) == 0 {
		return
	}
	ids := make([]string, len(comics))
	for i, c := range comics {
		ids[i] = c.ID
	}

	rows, err := db.Query(ctx, `SELECT comic_id FROM likes WHERE user_id = $1 AND comic_id = ANY($2)`, userID, ids)
	if err == nil {
		defer rows.Close()
		liked := make(map[string]bool)
		for rows.Next() {
			var cid string
			rows.Scan(&cid)
			liked[cid] = true
		}
		for i := range comics {
			comics[i].IsLiked = liked[comics[i].ID]
		}
	}

	rows, err = db.Query(ctx, `SELECT comic_id FROM favorites WHERE user_id = $1 AND comic_id = ANY($2)`, userID, ids)
	if err == nil {
		defer rows.Close()
		favd := make(map[string]bool)
		for rows.Next() {
			var cid string
			rows.Scan(&cid)
			favd[cid] = true
		}
		for i := range comics {
			comics[i].IsFavorited = favd[comics[i].ID]
		}
	}

	rows, err = db.Query(ctx, `SELECT comic_id FROM dislikes WHERE user_id = $1 AND comic_id = ANY($2)`, userID, ids)
	if err == nil {
		defer rows.Close()
		disliked := make(map[string]bool)
		for rows.Next() {
			var cid string
			rows.Scan(&cid)
			disliked[cid] = true
		}
		for i := range comics {
			comics[i].IsDisliked = disliked[comics[i].ID]
		}
	}
}

func getAuthData(ctx context.Context) (*myauth.AuthData, bool) {
	data, ok := auth.Data().(*myauth.AuthData)
	return data, ok
}

// isMatureRating reports whether an age rating is mature/explicit.
func isMatureRating(rating string) bool {
	return rating == "mature" || rating == "explicit"
}

// matureBlocked reports whether mature/explicit content should be withheld from
// the caller. Blocked when: the caller is anonymous and the global
// hide-mature-default/forbid-for-free policy applies; a free user under
// forbid-mature-for-free; or any authenticated user who has opted into their
// per-user hide-mature preference. Staff (admin/moderator) are never blocked.
func matureBlocked(ctx context.Context) bool {
	ad, ok := getAuthData(ctx)
	if ok {
		if ad.Role == "admin" || ad.Role == "moderator" {
			return false
		}
		// Per-user hide-mature preference applies at any tier.
		if userHidesMature(ctx, ad.UserID) {
			return true
		}
		if ad.Tier != "" && ad.Tier != "free" {
			return false
		}
	}
	// Anonymous or free tier → consult the global policy.
	policy, err := myauth.GetContentPolicy(ctx)
	if err != nil {
		return false
	}
	if !ok {
		// Anonymous visitors hide mature content by default when configured.
		return policy.ForbidMatureForFree || policy.HideMatureDefault
	}
	return policy.ForbidMatureForFree
}

// userHidesMature reports whether an authenticated user has the hide-mature
// preference enabled.
func userHidesMature(ctx context.Context, userID string) bool {
	prefs, err := myauth.GetUserPreferences(ctx, userID)
	if err != nil {
		return false
	}
	return prefs.HideMature
}

// matureWhereClause returns a SQL fragment that excludes mature content, to be
// appended to a WHERE clause. The values are hardcoded enum literals (no user
// input), so they are safe to inline without placeholders.
func matureWhereClause(ctx context.Context) string {
	if !matureBlocked(ctx) {
		return ""
	}
	return " AND age_rating NOT IN ('mature', 'explicit')"
}

// ----- Moderation endpoints -----

type PendingComic struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	UploaderID string    `json:"uploader_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListPendingResponse struct {
	Comics []PendingComic `json:"comics"`
	Total  int            `json:"total"`
}

type PendingComicsParams struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	Search  string `query:"search"`
	Sort    string `query:"sort"`
	SortDir string `query:"sort_dir"`
}

//encore:api auth method=GET path=/moderation/comics
func PendingComics(ctx context.Context, p *PendingComicsParams) (*ListPendingResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 100 { limit = 100 }
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := sanitizeSortCol(p.Sort, "created_at", "title", "status", "uploader_id")
	sortDir := "ASC"
	if strings.ToLower(p.SortDir) == "desc" { sortDir = "DESC" }

	var total int
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM comics WHERE status = 'pending_review'
		AND (title ILIKE $1 OR uploader_id ILIKE $1)
	`, search).Scan(&total)

	rows, err := db.Query(ctx, fmt.Sprintf(`
		SELECT id, title, uploader_id, status, created_at
		FROM comics WHERE status = 'pending_review'
		AND (title ILIKE $1 OR uploader_id ILIKE $1)
		ORDER BY %s %s LIMIT $2 OFFSET $3
	`, sortCol, sortDir), search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comics []PendingComic
	for rows.Next() {
		var c PendingComic
		if err := rows.Scan(&c.ID, &c.Title, &c.UploaderID, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		comics = append(comics, c)
	}

	return &ListPendingResponse{Comics: comics, Total: total}, rows.Err()
}

type RejectParams struct {
	Reason string `json:"reason"`
}

//encore:api auth method=POST path=/moderation/comics/:id/approve
func ApproveComic(ctx context.Context, id string) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	var status string
	err := db.QueryRow(ctx, `SELECT status FROM comics WHERE id = $1`, id).Scan(&status)
	if err != nil {
		return &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	if status != "pending_review" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "comic is not pending review"}
	}

	_, err = db.Exec(ctx, `
		UPDATE comics SET status = 'published', published_at = now(), updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	db.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, target_type, target_id)
		VALUES ($1, 'approve_comic', 'comic', $2)
	`, ad.UserID, id)
	go notifyUploaderFollowers(context.Background(), id)
	return nil
}

//encore:api auth method=POST path=/moderation/comics/:id/reject
func RejectComic(ctx context.Context, id string, p *RejectParams) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	_, err := db.Exec(ctx, `
		UPDATE comics SET status = 'rejected', rejection_reason = $1, updated_at = now()
		WHERE id = $2 AND status = 'pending_review'
	`, p.Reason, id)
	if err != nil {
		return err
	}

	db.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, target_type, target_id, details)
		VALUES ($1, 'reject_comic', 'comic', $2, $3)
	`, ad.UserID, id, `{"reason":"`+p.Reason+`"}`)
	return nil
}

type BulkActionParams struct {
	IDs    []string `json:"ids"`
	Action string   `json:"action"`
}

//encore:api auth method=POST path=/moderation/bulk
func BulkModerate(ctx context.Context, p *BulkActionParams) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	if p.Action != "approve" && p.Action != "reject" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "action must be 'approve' or 'reject'"}
	}

	for _, id := range p.IDs {
		if p.Action == "approve" {
			db.Exec(ctx, `UPDATE comics SET status = 'published', published_at = now(), updated_at = now() WHERE id = $1 AND status = 'pending_review'`, id)
			db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id) VALUES ($1, 'approve_comic', 'comic', $2)`, ad.UserID, id)
			go notifyUploaderFollowers(context.Background(), id)
		} else {
			db.Exec(ctx, `UPDATE comics SET status = 'rejected', updated_at = now() WHERE id = $1 AND status = 'pending_review'`, id)
			db.Exec(ctx, `INSERT INTO audit_logs (actor_id, action, target_type, target_id, details) VALUES ($1, 'reject_comic', 'comic', $2, '{"bulk":true}')`, ad.UserID, id)
		}
	}

	return nil
}

type LikeStatus struct {
	Liked     bool `json:"liked"`
	Favorited bool `json:"favorited"`
	Disliked  bool `json:"disliked"`
}

//encore:api auth method=GET path=/comics/:id/like-status
func GetLikeStatus(ctx context.Context, id string) (*LikeStatus, error) {
	ad := auth.Data().(*myauth.AuthData)

	var liked, favorited, disliked bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&liked)
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&favorited)
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM dislikes WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&disliked)

	return &LikeStatus{Liked: liked, Favorited: favorited, Disliked: disliked}, nil
}

type ToggleLikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

//encore:api auth method=POST path=/comics/:id/like
func ToggleLike(ctx context.Context, id string) (*ToggleLikeResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var exists bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&exists)

	if exists {
		db.Exec(ctx, `DELETE FROM likes WHERE user_id = $1 AND comic_id = $2`, ad.UserID, id)
		db.Exec(ctx, `UPDATE comics SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, id)
		var likeCount int
		db.QueryRow(ctx, `SELECT like_count FROM comics WHERE id = $1`, id).Scan(&likeCount)
		return &ToggleLikeResponse{Liked: false, LikeCount: likeCount}, nil
	}

	db.Exec(ctx, `DELETE FROM dislikes WHERE user_id = $1 AND comic_id = $2`, ad.UserID, id)
	db.Exec(ctx, `UPDATE comics SET dislike_count = GREATEST(dislike_count - 1, 0) WHERE id = $1`, id)

	_, err := db.Exec(ctx, `INSERT INTO likes (user_id, comic_id) VALUES ($1, $2)`, ad.UserID, id)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	db.Exec(ctx, `UPDATE comics SET like_count = like_count + 1 WHERE id = $1`, id)

	var likeCount int
	db.QueryRow(ctx, `SELECT like_count FROM comics WHERE id = $1`, id).Scan(&likeCount)
	return &ToggleLikeResponse{Liked: true, LikeCount: likeCount}, nil
}

type ToggleFavResponse struct {
	Favorited bool `json:"favorited"`
	FavCount  int  `json:"fav_count"`
}

//encore:api auth method=POST path=/comics/:id/favorite
func ToggleFavorite(ctx context.Context, id string) (*ToggleFavResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var exists bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&exists)

	if exists {
		db.Exec(ctx, `DELETE FROM favorites WHERE user_id = $1 AND comic_id = $2`, ad.UserID, id)
		db.Exec(ctx, `UPDATE comics SET fav_count = GREATEST(fav_count - 1, 0) WHERE id = $1`, id)
		db.Exec(ctx, `UPDATE series SET hearts_count = GREATEST(hearts_count - 1, 0) WHERE id = (SELECT series_id FROM comics WHERE id = $1)`, id)
		var favCount int
		db.QueryRow(ctx, `SELECT fav_count FROM comics WHERE id = $1`, id).Scan(&favCount)
		return &ToggleFavResponse{Favorited: false, FavCount: favCount}, nil
	}

	_, err := db.Exec(ctx, `INSERT INTO favorites (user_id, comic_id) VALUES ($1, $2)`, ad.UserID, id)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	db.Exec(ctx, `UPDATE comics SET fav_count = fav_count + 1 WHERE id = $1`, id)
	db.Exec(ctx, `UPDATE series SET hearts_count = hearts_count + 1 WHERE id = (SELECT series_id FROM comics WHERE id = $1)`, id)

	var favCount int
	db.QueryRow(ctx, `SELECT fav_count FROM comics WHERE id = $1`, id).Scan(&favCount)
	return &ToggleFavResponse{Favorited: true, FavCount: favCount}, nil
}

type ToggleDislikeResponse struct {
	Disliked     bool `json:"disliked"`
	DislikeCount int  `json:"dislike_count"`
}

//encore:api auth method=POST path=/comics/:id/dislike
func ToggleDislike(ctx context.Context, id string) (*ToggleDislikeResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	var exists bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM dislikes WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&exists)

	if exists {
		db.Exec(ctx, `DELETE FROM dislikes WHERE user_id = $1 AND comic_id = $2`, ad.UserID, id)
		db.Exec(ctx, `UPDATE comics SET dislike_count = GREATEST(dislike_count - 1, 0) WHERE id = $1`, id)
		var dislikeCount int
		db.QueryRow(ctx, `SELECT dislike_count FROM comics WHERE id = $1`, id).Scan(&dislikeCount)
		return &ToggleDislikeResponse{Disliked: false, DislikeCount: dislikeCount}, nil
	}

	db.Exec(ctx, `DELETE FROM likes WHERE user_id = $1 AND comic_id = $2`, ad.UserID, id)
	db.Exec(ctx, `UPDATE comics SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, id)

	_, err := db.Exec(ctx, `INSERT INTO dislikes (user_id, comic_id) VALUES ($1, $2)`, ad.UserID, id)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	db.Exec(ctx, `UPDATE comics SET dislike_count = dislike_count + 1 WHERE id = $1`, id)

	var dislikeCount int
	db.QueryRow(ctx, `SELECT dislike_count FROM comics WHERE id = $1`, id).Scan(&dislikeCount)
	return &ToggleDislikeResponse{Disliked: true, DislikeCount: dislikeCount}, nil
}

// ----- Admin comic list -----

type AdminListComicsParams struct {
	Page          int    `query:"page"`
	Limit         int    `query:"limit"`
	Search        string `query:"search"`
	Sort          string `query:"sort"`
	SortDir       string `query:"sort_dir"`
	FilterStatus  string `query:"filter_status"`
	FilterAuthor  string `query:"filter_author"`
	FilterTitle   string `query:"filter_title"`
}

//encore:api auth method=GET path=/admin/comics
func AdminListComics(ctx context.Context, p *AdminListComicsParams) (*ListComicsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 100 { limit = 100 }
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := sanitizeSortCol(p.Sort, "created_at", "title", "author", "status", "published_at", "view_count", "download_count")
	sortDir := "DESC"
	if strings.ToLower(p.SortDir) == "asc" { sortDir = "ASC" }

	where := "WHERE (title ILIKE $1 OR author ILIKE $1)"
	args := []interface{}{search}
	argIdx := 2

	if p.FilterStatus != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, p.FilterStatus)
		argIdx++
	}
	if p.FilterAuthor != "" {
		where += fmt.Sprintf(" AND author ILIKE $%d", argIdx)
		args = append(args, "%"+p.FilterAuthor+"%")
		argIdx++
	}
	if p.FilterTitle != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+p.FilterTitle+"%")
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			created_at, updated_at
		FROM comics %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil { return nil, err }

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics `+where, args[:len(args)-2]...).Scan(&total)

	enrichReactions(ctx, comics, string(ad.UserID))

	return &ListComicsResponse{Comics: comics, Total: total}, nil
}

// ----- Comments -----

type CommentData struct {
	ID        string        `json:"id"`
	ComicID   string        `json:"comic_id"`
	UserID    string        `json:"user_id"`
	Username  string        `json:"username,omitempty"`
	AvatarKey string        `json:"avatar_key,omitempty"`
	ParentID  string        `json:"parent_id,omitempty"`
	BodyText  string        `json:"body_text"`
	CreatedAt time.Time     `json:"created_at"`
	Replies   []CommentData `json:"replies,omitempty"`
}

type ListCommentsResponse struct {
	Comments []CommentData `json:"comments"`
}

//encore:api public method=GET path=/comics/:id/comments
func ListComments(ctx context.Context, id string) (*ListCommentsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT c.id, c.comic_id, c.user_id, COALESCE(c.parent_id::text, ''), c.body_text, c.created_at
		FROM comments c
		WHERE c.comic_id = $1
		ORDER BY c.created_at ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []CommentData
	for rows.Next() {
		var c CommentData
		if err := rows.Scan(&c.ID, &c.ComicID, &c.UserID, &c.ParentID, &c.BodyText, &c.CreatedAt); err != nil {
			return nil, err
		}
		all = append(all, c)
	}

	// Enrich comments with the author's public identity (username/avatar) for
	// display and messaging. Users live in the auth service (ADR 0016).
	enrichCommentAuthors(ctx, all)

	roots := buildThread(all)
	if roots == nil {
		roots = []CommentData{}
	}
	return &ListCommentsResponse{Comments: roots}, nil
}

type CreateCommentParams struct {
	BodyText       string `json:"body_text"`
	ParentID       string `json:"parent_id"`
	TurnstileToken string `json:"turnstile_token"`
}

//encore:api auth method=POST path=/comics/:id/comments
func CreateComment(ctx context.Context, id string, p *CreateCommentParams) (*CommentData, error) {
	ad := auth.Data().(*myauth.AuthData)

	if err := turnstile.Verify(ctx, &turnstile.VerifyParams{Token: p.TurnstileToken, Action: "comment"}); err != nil {
		return nil, err
	}

	if p.BodyText == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body_text is required"}
	}

	if policy, err := myauth.GetContentPolicy(ctx); err == nil && !policy.EnableComments {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "comments are currently disabled"}
	}

	var parentID interface{}
	if p.ParentID != "" {
		parentID = p.ParentID
	}

	var c CommentData
	err := db.QueryRow(ctx, `
		INSERT INTO comments (comic_id, user_id, parent_id, body_text)
		VALUES ($1, $2, $3, $4)
		RETURNING id, comic_id, user_id, COALESCE(parent_id::text, ''), body_text, created_at
	`, id, ad.UserID, parentID, p.BodyText).Scan(&c.ID, &c.ComicID, &c.UserID, &c.ParentID, &c.BodyText, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	publishComment(id, c)
	moderationTopic.Publish(ctx, ModerationEvent{TargetType: "comment", TargetID: c.ID})
	return &c, nil
}

//encore:api public raw method=GET path=/comments-stream/:id
func CommentStream(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	comicID := req.PathValue("id")
	ch := subscribeComments(comicID)
	defer unsubscribeComments(comicID, ch)

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case c := <-ch:
			data, _ := json.Marshal(c)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, ":\n\n")
			flusher.Flush()
		case <-req.Context().Done():
			return
		}
	}
}

// enrichCommentAuthors attaches each comment author's public identity
// (username/avatar) from the auth service (ADR 0016).
func enrichCommentAuthors(ctx context.Context, comments []CommentData) {
	seen := make(map[string]bool)
	var ids []string
	for _, c := range comments {
		if c.UserID != "" && !seen[c.UserID] {
			seen[c.UserID] = true
			ids = append(ids, c.UserID)
		}
	}
	if len(ids) == 0 {
		return
	}

	res, err := myauth.GetUsersInfo(ctx, &myauth.GetUsersInfoParams{IDs: ids})
	if err != nil {
		return
	}
	m := make(map[string]myauth.UserPublicInfo, len(res.Users))
	for _, u := range res.Users {
		m[u.ID] = u
	}
	for i := range comments {
		if u, ok := m[comments[i].UserID]; ok {
			comments[i].Username = u.Username
			comments[i].AvatarKey = u.AvatarKey
		}
	}
}

func buildThread(comments []CommentData) []CommentData {
	children := make(map[string][]CommentData)
	for _, c := range comments {
		if c.ParentID != "" {
			children[c.ParentID] = append(children[c.ParentID], c)
		}
	}

	var thread func(parentID string) []CommentData
	thread = func(parentID string) []CommentData {
		kids := children[parentID]
		for i := range kids {
			kids[i].Replies = thread(kids[i].ID)
		}
		return kids
	}

	var result []CommentData
	for _, c := range comments {
		if c.ParentID == "" {
			c.Replies = thread(c.ID)
			result = append(result, c)
		}
	}
	return result
}

//encore:api auth method=DELETE path=/comments/:id
func DeleteComment(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	var userID string
	err := db.QueryRow(ctx, `SELECT user_id FROM comments WHERE id = $1`, id).Scan(&userID)
	if err != nil {
		return &errs.Error{Code: errs.NotFound, Message: "comment not found"}
	}
	if userID != ad.UserID && ad.Role != "admin" && ad.Role != "moderator" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "not your comment"}
	}
	_, err = db.Exec(ctx, `DELETE FROM comments WHERE id = $1`, id)
	return err
}

// ----- Comment Flagging -----

type FlagCommentParams struct {
	Reason string `json:"reason"`
}

//encore:api auth method=POST path=/comments/:id/flag
func FlagComment(ctx context.Context, id string, p *FlagCommentParams) error {
	ad := auth.Data().(*myauth.AuthData)

	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return &errs.Error{Code: errs.NotFound, Message: "comment not found"}
	}
	if !exists {
		return &errs.Error{Code: errs.NotFound, Message: "comment not found"}
	}

	// Idempotent: one flag per user per comment (unique index).
	_, err = db.Exec(ctx, `
		INSERT INTO comment_flags (comment_id, user_id, reason)
		VALUES ($1, $2, $3)
		ON CONFLICT (comment_id, user_id) DO NOTHING
	`, id, ad.UserID, p.Reason)
	return err
}

type FlaggedComment struct {
	FlagID      string    `json:"flag_id"`
	CommentID   string    `json:"comment_id"`
	ComicID     string    `json:"comic_id"`
	ComicTitle  string    `json:"comic_title"`
	UserID      string    `json:"user_id"`
	BodyText    string    `json:"body_text"`
	Reason      string    `json:"reason"`
	FlagCount   int       `json:"flag_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListFlaggedCommentsResponse struct {
	Flags []FlaggedComment `json:"flags"`
	Total int              `json:"total"`
}

type ListFlaggedCommentsParams struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
}

//encore:api auth method=GET path=/moderation/flags
func ListFlaggedComments(ctx context.Context, p *ListFlaggedCommentsParams) (*ListFlaggedCommentsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

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

//encore:api auth method=POST path=/moderation/flags/:id/resolve
func ResolveFlag(ctx context.Context, id string) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	_, err := db.Exec(ctx, `
		UPDATE comment_flags SET status = 'resolved', resolved_at = now(), resolved_by = $1
		WHERE id = $2 AND status = 'open'
	`, ad.UserID, id)
	return err
}

// ----- Series -----

type Series struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	UploaderID   string    `json:"uploader_id"`
	CoverKey     string    `json:"cover_key"`
	Genre        string    `json:"genre,omitempty"`
	Category     string    `json:"category,omitempty"`
	OverlayTitle string    `json:"overlay_title,omitempty"`
	ViewsCount   int64     `json:"views_count"`
	HeartsCount  int64     `json:"hearts_count"`
	ScheduleDay  string    `json:"schedule_day,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// Computed fields (not stored).
	Rank     int    `json:"rank,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
}

type CreateSeriesParams struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
	Category    string `json:"category"`
	ScheduleDay string `json:"schedule_day"`
	CoverKey    string `json:"cover_key"`
}

//encore:api auth method=POST path=/series
func CreateSeries(ctx context.Context, p *CreateSeriesParams) (*Series, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "uploader" && ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "only uploaders can create series"}
	}
	if p.Title == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "title is required"}
	}

	slug := generateSlug(p.Title)
	scheduleDayInput := p.ScheduleDay
	var s Series
	var scheduleDay sql.NullString
	err := db.QueryRow(ctx, `
		INSERT INTO series (title, slug, description, genre, category, schedule_day, cover_key, uploader_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, slug, description, uploader_id, cover_key, genre, category, overlay_title, views_count, hearts_count, schedule_day, created_at
	`, p.Title, slug, p.Description, p.Genre, p.Category, nulOrValue(scheduleDayInput), p.CoverKey, ad.UserID).Scan(
		&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.ScheduleDay = scheduleDay.String
	return &s, nil
}

// resolveSeriesForComic returns the series ID a comic should be attached to,
// either an existing series (SeriesID) or a newly-created series (SeriesTitle).
// Returns "" when no series association was requested.
func resolveSeriesForComic(ctx context.Context, uploaderID string, p *CreateComicParams, comic *Comic) (string, error) {
	if strings.TrimSpace(p.SeriesID) != "" {
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM series WHERE id = $1)`, p.SeriesID).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return "", &errs.Error{Code: errs.NotFound, Message: "series not found"}
		}
		return p.SeriesID, nil
	}

	title := strings.TrimSpace(p.SeriesTitle)
	if title == "" {
		return "", nil
	}

	// Reuse an existing series with the same title (case-insensitive) if present.
	var existingID string
	if err := db.QueryRow(ctx, `SELECT id FROM series WHERE LOWER(title) = LOWER($1) LIMIT 1`, title).Scan(&existingID); err == nil && existingID != "" {
		return existingID, nil
	}

	genre := p.SeriesGenre
	if genre == "" {
		genre = comic.Genre
	}
	category := p.SeriesCategory
	if category == "" {
		category = comic.Category
	}

	slug := generateSlug(title)
	var seriesID string
	err := db.QueryRow(ctx, `
		INSERT INTO series (title, slug, description, genre, category, schedule_day, cover_key, uploader_id)
		VALUES ($1, $2, '', $3, $4, $5, $6, $7)
		RETURNING id
	`, title, slug, genre, category, nulOrValue(p.SeriesScheduleDay), comic.CoverKey, uploaderID).Scan(&seriesID)
	if err != nil {
		return "", err
	}
	return seriesID, nil
}

type ListSeriesResponse struct {
	Series []Series `json:"series"`
	Total  int      `json:"total"`
}

//encore:api public method=GET path=/series
func ListSeries(ctx context.Context) (*ListSeriesResponse, error) {
	rows, err := db.Query(ctx, `SELECT id, title, slug, description, uploader_id, cover_key, genre, category, overlay_title, views_count, hearts_count, schedule_day, created_at FROM series ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Series
	for rows.Next() {
		var s Series
		var scheduleDay sql.NullString
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.ScheduleDay = scheduleDay.String
		s.CoverURL = resolveCoverURL(s.CoverKey)
		result = append(result, s)
	}
	return &ListSeriesResponse{Series: result}, rows.Err()
}

type SearchSeriesParams struct {
	Search   string `query:"search"`
	Category string `query:"category"`
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
}

//encore:api public method=GET path=/series-search
func SearchSeries(ctx context.Context, p *SearchSeriesParams) (*ListSeriesResponse, error) {
	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1
	if p.Search != "" {
		where += fmt.Sprintf(" AND (title ILIKE $%d OR slug ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+p.Search+"%")
		argIdx++
	}
	if p.Category != "" {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, p.Category)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, title, slug, description, uploader_id, cover_key, genre, category, overlay_title, views_count, hearts_count, schedule_day, created_at
		FROM series %s ORDER BY title ASC LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Series
	for rows.Next() {
		var s Series
		var scheduleDay sql.NullString
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.ScheduleDay = scheduleDay.String
		s.CoverURL = resolveCoverURL(s.CoverKey)
		result = append(result, s)
	}
	if result == nil {
		result = []Series{}
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM series `+where, args[:len(args)-2]...).Scan(&total)

	return &ListSeriesResponse{Series: result, Total: total}, nil
}

type SeriesCategoriesResponse struct {
	Categories []string `json:"categories"`
}

//encore:api public method=GET path=/series-categories
func ListSeriesCategories(ctx context.Context) (*SeriesCategoriesResponse, error) {
	cats, err := listCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &SeriesCategoriesResponse{Categories: cats}, nil
}

//encore:api public method=GET path=/series/:id
func GetSeries(ctx context.Context, id string) (*Series, error) {
	var s Series
	var scheduleDay sql.NullString
	err := db.QueryRow(ctx, `SELECT id, title, slug, description, uploader_id, cover_key, genre, category, overlay_title, views_count, hearts_count, schedule_day, created_at FROM series WHERE id = $1`, id).Scan(
		&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt,
	)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "series not found"}
	}
	s.ScheduleDay = scheduleDay.String
	s.CoverURL = resolveCoverURL(s.CoverKey)
	return &s, nil
}

type AdminListSeriesParams struct {
	Page          int    `query:"page"`
	Limit         int    `query:"limit"`
	Search        string `query:"search"`
	Sort          string `query:"sort"`
	SortDir       string `query:"sort_dir"`
	FilterGenre   string `query:"filter_genre"`
	FilterCategory string `query:"filter_category"`
}

//encore:api auth method=GET path=/admin/series
func AdminListSeries(ctx context.Context, p *AdminListSeriesParams) (*ListSeriesResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 100 { limit = 100 }
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := sanitizeSortCol(p.Sort, "created_at", "title", "genre", "category", "schedule_day", "views_count", "hearts_count")
	sortDir := "DESC"
	if strings.ToLower(p.SortDir) == "asc" { sortDir = "ASC" }

	where := "WHERE (title ILIKE $1 OR slug ILIKE $1)"
	args := []interface{}{search}
	argIdx := 2

	if p.FilterGenre != "" {
		where += fmt.Sprintf(" AND genre = $%d", argIdx)
		args = append(args, p.FilterGenre)
		argIdx++
	}
	if p.FilterCategory != "" {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, p.FilterCategory)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, title, slug, description, uploader_id, cover_key, genre, category, overlay_title, views_count, hearts_count, schedule_day, created_at
		FROM series %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var result []Series
	for rows.Next() {
		var s Series
		var scheduleDay sql.NullString
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.ScheduleDay = scheduleDay.String
		s.CoverURL = resolveCoverURL(s.CoverKey)
		result = append(result, s)
	}
	if result == nil {
		result = []Series{}
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM series `+where, args[:len(args)-2]...).Scan(&total)

	return &ListSeriesResponse{Series: result, Total: total}, nil
}

// ----- Trending & Popular series -----

type TrendingPopularResponse struct {
	Trending []Series `json:"trending"`
	Popular  []Series `json:"popular"`
}

//encore:api public method=GET path=/series-trending-popular
func ListTrendingPopular(ctx context.Context) (*TrendingPopularResponse, error) {
	trending, err := rankedSeries(ctx, `SUM(c.view_count)`)
	if err != nil {
		return nil, err
	}
	popular, err := rankedSeries(ctx, `SUM(c.like_count)`)
	if err != nil {
		return nil, err
	}
	return &TrendingPopularResponse{Trending: trending, Popular: popular}, nil
}

// rankedSeries returns series ranked 1..N by the given aggregate expression
// over their published comics. Only series with at least one published comic
// are included.
func rankedSeries(ctx context.Context, orderExpr string) ([]Series, error) {
	rows, err := db.Query(ctx, `
		SELECT s.id, s.title, s.slug, s.description, s.uploader_id, s.cover_key, s.genre, s.category, s.overlay_title, s.views_count, s.hearts_count, s.schedule_day, s.created_at
		FROM series s
		JOIN comics c ON c.series_id = s.id AND c.status = 'published'
		GROUP BY s.id
		ORDER BY `+orderExpr+` DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Series
	rank := 0
	for rows.Next() {
		var s Series
		var scheduleDay sql.NullString
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt); err != nil {
			return nil, err
		}
		rank++
		s.Rank = rank
		s.ScheduleDay = scheduleDay.String
		if s.OverlayTitle == "" {
			s.OverlayTitle = s.Title
		}
		s.CoverURL = resolveCoverURL(s.CoverKey)
		result = append(result, s)
	}
	if result == nil {
		result = []Series{}
	}
	return result, rows.Err()
}

type SeriesComicsParams struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

//encore:api public method=GET path=/series/:id/comics
func SeriesComics(ctx context.Context, id string, p *SeriesComicsParams) (*ListComicsResponse, error) {
	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 50 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE series_id = $1 AND status = 'published'`+matureWhereClause(ctx), id).Scan(&total)

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			category, genre, cover_key, file_key, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			reading_direction, page_count, archive_mimetype, isbn, upc, issn, volume, issue_number,
			COALESCE(series_order, 1), created_at, updated_at
		FROM comics WHERE series_id = $1 AND status = 'published'`+matureWhereClause(ctx)+`
		ORDER BY series_order ASC, published_at ASC
		LIMIT $2 OFFSET $3
	`, id, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()

	var comics []Comic
	for rows.Next() {
		var c Comic
		var pubAt sql.NullTime
		if err := rows.Scan(
			&c.ID, &c.UploaderID, &c.Title, &c.Author, &c.Slug, &c.Description,
			&c.ContentLanguage, &c.Status, &c.Category, &c.Genre, &c.CoverKey, &c.FileKey,
			&c.FileSizeBytes, nulString(&c.MinTierID),
			&c.AgeRating, &c.IsPremium, scanStringSlice(&c.Tags), nulString(&c.RejectionReason),
			&pubAt, &c.ViewCount, &c.DownloadCount, &c.LikeCount, &c.FavCount, &c.DislikeCount,
			&c.ReadingDirection, &c.PageCount, nulString(&c.ArchiveMimetype),
			nulString(&c.Isbn), nulString(&c.Upc), nulString(&c.Issn), nulString(&c.Volume), nulString(&c.IssueNumber),
			&c.SeriesOrder, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if pubAt.Valid {
			c.PublishedAt = pubAt.Time
		}
		resolveComicURLs(&c)
		comics = append(comics, c)
	}
	if err := rows.Err(); err != nil { return nil, err }
	if comics == nil {
		comics = []Comic{}
	}
	return &ListComicsResponse{Comics: comics, Total: total}, nil
}

//encore:api auth method=POST path=/series/:id/follow
func FollowSeries(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	_, err := db.Exec(ctx, `INSERT INTO series_follows (user_id, series_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, ad.UserID, id)
	return err
}

//encore:api auth method=DELETE path=/series/:id/follow
func UnfollowSeries(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	_, err := db.Exec(ctx, `DELETE FROM series_follows WHERE user_id = $1 AND series_id = $2`, ad.UserID, id)
	return err
}

// ----- Uploader follow -----

//encore:api auth method=POST path=/uploaders/:id/follow
func FollowUploader(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.UserID == id {
		return &errs.Error{Code: errs.InvalidArgument, Message: "cannot follow yourself"}
	}
	_, err := db.Exec(ctx, `INSERT INTO uploader_follows (user_id, uploader_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, ad.UserID, id)
	return err
}

//encore:api auth method=DELETE path=/uploaders/:id/follow
func UnfollowUploader(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	_, err := db.Exec(ctx, `DELETE FROM uploader_follows WHERE user_id = $1 AND uploader_id = $2`, ad.UserID, id)
	return err
}

type UploaderFollowStatus struct {
	Following bool `json:"following"`
}

//encore:api auth method=GET path=/uploaders/:id/follow-status
func GetUploaderFollowStatus(ctx context.Context, id string) (*UploaderFollowStatus, error) {
	ad := auth.Data().(*myauth.AuthData)
	var following bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM uploader_follows WHERE user_id = $1 AND uploader_id = $2)`, ad.UserID, id).Scan(&following)
	return &UploaderFollowStatus{Following: following}, nil
}

// notifyUploaderFollowers emails followers of a newly published comic. Called
// after a comic transitions to published. Best-effort; errors are logged, not
// propagated (publishing must not fail because email delivery failed).
func notifyUploaderFollowers(ctx context.Context, comicID string) {
	var uploaderID, title string
	if err := db.QueryRow(ctx, `SELECT uploader_id, title FROM comics WHERE id = $1`, comicID).Scan(&uploaderID, &title); err != nil {
		return
	}

	rows, err := db.Query(ctx, `
		SELECT user_id FROM uploader_follows WHERE uploader_id = $1
	`, uploaderID)
	if err != nil {
		return
	}
	defer rows.Close()

	var followerIDs []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			followerIDs = append(followerIDs, id)
		}
	}
	if len(followerIDs) == 0 {
		return
	}

	// Delegate email delivery to the auth service (owns users + prefs).
	myauth.NotifyFollowersNewComic(ctx, &myauth.NotifyFollowersNewComicParams{
		UserIDs:    followerIDs,
		ComicTitle: title,
	})
}

// ----- RSS Feed -----

//encore:api public raw method=GET path=/rss
func RSSFeed(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	rows, err := db.Query(ctx, `
		SELECT title, slug, description, published_at FROM comics
		WHERE status = 'published' ORDER BY published_at DESC LIMIT 20
	`)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>Comics Galore</title><link>https://comicsgalore.com</link><description>Discover and read digital comics</description>`))

	for rows.Next() {
		var title, slug, desc string
		var pubAt *time.Time
		rows.Scan(&title, &slug, &desc, &pubAt)
		pubDate := ""
		if pubAt != nil {
			pubDate = pubAt.Format(time.RFC1123Z)
		}
		w.Write([]byte(`<item><title>` + escapeXML(title) + `</title><link>https://comicsgalore.com/comics/` + slug + `</link><description>` + escapeXML(desc) + `</description><pubDate>` + pubDate + `</pubDate></item>`))
	}

	w.Write([]byte(`</channel></rss>`))
}

func escapeXML(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '&': result += "&amp;"
		case '<': result += "&lt;"
		case '>': result += "&gt;"
		case '"': result += "&quot;"
		default: result += string(c)
		}
	}
	return result
}

type ArchiveComicParams struct {
	ID string `path:"id"`
}

//encore:api auth method=POST path=/admin/comics/:id/archive
func ArchiveComic(ctx context.Context, id string) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "admin" && ad.Role != "moderator") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin or moderator only"}
	}
	if id == "" { return &errs.Error{Code: errs.InvalidArgument, Message: "id required"} }
	_, err := db.Exec(ctx, `UPDATE comics SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

//encore:api auth method=POST path=/admin/comics/:id/restore
func RestoreComic(ctx context.Context, id string) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "admin" && ad.Role != "moderator") {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin or moderator only"}
	}
	if id == "" { return &errs.Error{Code: errs.InvalidArgument, Message: "id required"} }
	_, err := db.Exec(ctx, `UPDATE comics SET deleted_at = NULL WHERE id = $1`, id)
	return err
}

type RecycleBinParams struct {
	Page          int    `query:"page"`
	Limit         int    `query:"limit"`
	Search        string `query:"search"`
	Sort          string `query:"sort"`
	SortDir       string `query:"sort_dir"`
	FilterStatus  string `query:"filter_status"`
	FilterAuthor  string `query:"filter_author"`
	FilterTitle   string `query:"filter_title"`
}

//encore:api auth method=GET path=/admin/recycle-bin
func RecycleBin(ctx context.Context, p *RecycleBinParams) (*ListComicsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "admin" && ad.Role != "moderator") {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin or moderator only"}
	}

	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 100 { limit = 100 }
	offset := (page - 1) * limit

	search := "%" + p.Search + "%"
	sortCol := sanitizeSortCol(p.Sort, "deleted_at", "title", "author", "status")
	sortDir := "DESC"
	if strings.ToLower(p.SortDir) == "asc" { sortDir = "ASC" }

	where := "WHERE c.deleted_at IS NOT NULL AND (c.title ILIKE $1 OR c.author ILIKE $1)"
	args := []interface{}{search}
	argIdx := 2

	if p.FilterStatus != "" {
		where += fmt.Sprintf(" AND c.status = $%d", argIdx)
		args = append(args, p.FilterStatus)
		argIdx++
	}
	if p.FilterAuthor != "" {
		where += fmt.Sprintf(" AND c.author ILIKE $%d", argIdx)
		args = append(args, "%"+p.FilterAuthor+"%")
		argIdx++
	}
	if p.FilterTitle != "" {
		where += fmt.Sprintf(" AND c.title ILIKE $%d", argIdx)
		args = append(args, "%"+p.FilterTitle+"%")
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.uploader_id, c.title, c.author, c.slug, c.description, c.content_language, c.status,
			c.category, c.genre, c.cover_key, c.file_key, c.file_size_bytes, c.min_tier_id, c.age_rating, c.is_premium,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count, c.dislike_count,
			c.reading_direction, c.page_count, c.archive_mimetype, c.isbn, c.upc, c.issn, c.volume, c.issue_number,
			c.created_at, c.updated_at
		FROM comics c %s ORDER BY %s %s LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	comics, err := scanComics(rows)
	if err != nil { return nil, err }

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics c `+where, args[:len(args)-2]...).Scan(&total)

	return &ListComicsResponse{Comics: comics, Total: total}, nil
}

//encore:api auth method=DELETE path=/comics/:id
func DeleteComic(ctx context.Context, id string) error {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" && ad.Role != "moderator" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin or moderator only"}
	}

	_, err := db.Exec(ctx, `DELETE FROM comics WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return nil
}

type AuditLogEntry struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Details    string    `json:"details,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLogsResponse struct {
	Entries []AuditLogEntry `json:"entries"`
}

//encore:api auth method=GET path=/admin/audit-logs
func AdminAuditLogs(ctx context.Context) (*AuditLogsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `SELECT id, actor_id, action, target_type, target_id, COALESCE(details::text, ''), created_at FROM audit_logs ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditLogEntry
	for rows.Next() {
		var e AuditLogEntry
		var rawDetails string
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.TargetType, &e.TargetID, &rawDetails, &e.CreatedAt); err != nil {
			return nil, err
		}
		if rawDetails != "" && rawDetails != "{}" {
			var parsed interface{}
			if json.Unmarshal([]byte(rawDetails), &parsed) == nil {
				e.Details = rawDetails
			}
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AuditLogsResponse{Entries: entries}, nil
}

// ----- Admin Stats -----

type TopLikedComic struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	CoverKey  string `json:"cover_key"`
	LikeCount int    `json:"like_count"`
}

type TopViewedComic struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	CoverKey  string `json:"cover_key"`
	ViewCount int64  `json:"view_count"`
}

type ComicsStats struct {
	TotalComics    int             `json:"total_comics"`
	PublishedComics int            `json:"published_comics"`
	PendingComics  int             `json:"pending_comics"`
	RejectedComics int             `json:"rejected_comics"`
	TotalViews     int64           `json:"total_views"`
	StorageBytes   int64           `json:"storage_bytes"`
	TopLiked       []TopLikedComic `json:"top_liked"`
	TopViewed      []TopViewedComic `json:"top_viewed"`
}

//encore:api private
func GetComicsStats(ctx context.Context) (*ComicsStats, error) {
	var stats ComicsStats
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE deleted_at IS NULL`).Scan(&stats.TotalComics)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE status = 'published' AND deleted_at IS NULL`).Scan(&stats.PublishedComics)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE status = 'pending_review' AND deleted_at IS NULL`).Scan(&stats.PendingComics)
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE status = 'rejected' AND deleted_at IS NULL`).Scan(&stats.RejectedComics)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(view_count), 0) FROM comics WHERE deleted_at IS NULL`).Scan(&stats.TotalViews)
	db.QueryRow(ctx, `SELECT COALESCE(SUM(file_size_bytes), 0) FROM comics WHERE deleted_at IS NULL`).Scan(&stats.StorageBytes)

	rows, err := db.Query(ctx, `
		SELECT id, title, slug, COALESCE(cover_key, ''), COALESCE(like_count, 0)
		FROM comics
		WHERE deleted_at IS NULL
		ORDER BY like_count DESC NULLS LAST
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c TopLikedComic
			if rows.Scan(&c.ID, &c.Title, &c.Slug, &c.CoverKey, &c.LikeCount) == nil {
				stats.TopLiked = append(stats.TopLiked, c)
			}
		}
	}

	vrows, err := db.Query(ctx, `
		SELECT id, title, slug, COALESCE(cover_key, ''), COALESCE(view_count, 0)
		FROM comics
		WHERE deleted_at IS NULL
		ORDER BY view_count DESC NULLS LAST
		LIMIT 5
	`)
	if err == nil {
		defer vrows.Close()
		for vrows.Next() {
			var c TopViewedComic
			if vrows.Scan(&c.ID, &c.Title, &c.Slug, &c.CoverKey, &c.ViewCount) == nil {
				stats.TopViewed = append(stats.TopViewed, c)
			}
		}
	}

	return &stats, nil
}

func defaultValue(v int, def int) int {
	if v <= 0 { return def }
	return v
}

func defaultPageSize(ctx context.Context, fallback int) int {
	if ad, ok := getAuthData(ctx); ok {
		if prefs, err := myauth.GetUserPreferences(ctx, ad.UserID); err == nil && prefs.ItemsPerPage > 0 {
			return prefs.ItemsPerPage
		}
	}
	if cfg, err := myauth.GetAppConfig(ctx); err == nil && cfg.ItemsPerPage > 0 {
		return cfg.ItemsPerPage
	}
	return fallback
}

func sanitizeSortCol(sort string, allowed ...string) string {
	if sort == "" { return allowed[0] }
	for _, a := range allowed {
		if sort == a { return a }
	}
	return allowed[0]
}
