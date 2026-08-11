package comics

import (
	"context"
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
	CoverKey        string    `json:"cover_key"`
	FileKey         string    `json:"file_key"`
	PageKeys        []string  `json:"page_keys"`
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
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Resolved URLs (populated server-side, not from DB)
	CoverURL  string   `json:"cover_url,omitempty"`
	PageURLs  []string `json:"page_urls,omitempty"`
}

type CreateComicParams struct {
	Title           string   `json:"title"`
	Author          string   `json:"author"`
	Description     string   `json:"description"`
	ContentLanguage string   `json:"content_language"`
	CoverKey        string   `json:"cover_key"`
	FileKey         string   `json:"file_key"`
	PageKeys        []string `json:"page_keys"`
	FileSizeBytes   int64    `json:"file_size_bytes"`
	MinTierID       string   `json:"min_tier_id"`
	AgeRating       string   `json:"age_rating"`
	IsPremium       bool     `json:"is_premium"`
	Tags            []string `json:"tags"`
	UploadSessionID string   `json:"upload_session_id"`
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

	if p.Title == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "title is required"}
	}
	if p.CoverKey == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "cover_key is required"}
	}
	if p.FileKey == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "file_key is required"}
	}

	lang := p.ContentLanguage
	if lang == "" {
		lang = "en"
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

	err := db.QueryRow(ctx, `
		INSERT INTO comics (uploader_id, title, author, slug, description, content_language,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, uploader_id, title, author, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, view_count, download_count, like_count, fav_count, dislike_count,
			created_at, updated_at
	`, ad.UserID, p.Title, p.Author, slug, p.Description, lang,
		p.CoverKey, p.FileKey, pageKeys, p.FileSizeBytes, minTierID, ageRating, p.IsPremium, tags).Scan(
		&comic.ID, &comic.UploaderID, &comic.Title, &comic.Author, &comic.Slug, &comic.Description,
		&comic.ContentLanguage, &comic.Status, &comic.CoverKey, &comic.FileKey,
		scanStringSlice(&comic.PageKeys), &comic.FileSizeBytes, nulString(&comic.MinTierID),
		&comic.AgeRating, &comic.IsPremium, scanStringSlice(&comic.Tags), nulString(&comic.RejectionReason),
		&comic.ViewCount, &comic.DownloadCount, &comic.LikeCount, &comic.FavCount, &comic.DislikeCount,
		&comic.CreatedAt, &comic.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if p.UploadSessionID != "" {
		db.Exec(ctx, `UPDATE upload_sessions SET status = 'completed' WHERE id = $1 AND user_id = $2`,
			p.UploadSessionID, ad.UserID)
	}

	return &comic, nil
}

type ListComicsParams struct {
	Page          int    `query:"page"`
	Limit         int    `query:"limit"`
	Language      string `query:"language"`
	Search        string `query:"search"`
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
		limit = 20
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

	if p.Tag != "" {
		where += " AND tags @> $" + nextIdx(len(args)+1) + "::jsonb"
		args = append(args, `["`+p.Tag+`"]`)
	}

	if p.ExcludeMature == "true" {
		where += " AND age_rating NOT IN (" + nextIdx(len(args)+1) + ", " + nextIdx(len(args)+2) + ")"
		args = append(args, "mature", "explicit")
	}

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
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
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

//encore:api public method=GET path=/comics/:slug
func GetComic(ctx context.Context, slug string) (*Comic, error) {
	ad, hasAuth := getAuthData(ctx)

	var comic Comic
	err := db.QueryRow(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, view_count, download_count, like_count, fav_count, dislike_count,
			created_at, updated_at
		FROM comics WHERE slug = $1
	`, slug).Scan(
		&comic.ID, &comic.UploaderID, &comic.Title, &comic.Author, &comic.Slug, &comic.Description,
		&comic.ContentLanguage, &comic.Status, &comic.CoverKey, &comic.FileKey,
		scanStringSlice(&comic.PageKeys), &comic.FileSizeBytes, nulString(&comic.MinTierID),
		&comic.AgeRating, &comic.IsPremium, scanStringSlice(&comic.Tags), nulString(&comic.RejectionReason),
		&comic.ViewCount, &comic.DownloadCount, &comic.LikeCount, &comic.FavCount, &comic.DislikeCount,
		&comic.CreatedAt, &comic.UpdatedAt,
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

	if hasAuth {
		enrichReactions(ctx, []Comic{comic}, string(ad.UserID))
	}

	resolveComicURLs(&comic)
	return &comic, nil
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
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			created_at, updated_at
		FROM comics WHERE id = ANY($1) AND status = 'published'
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

//encore:api auth method=GET path=/uploader/comics
func MyComics(ctx context.Context) (*ListComicsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, view_count, download_count, like_count, fav_count, dislike_count,
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
	for _, k := range keys {
		if k == "" {
			continue
		}
		// For pages, use the same resolution as cover
	}
	urls := make([]string, len(keys))
	for i, k := range keys {
		urls[i] = resolveCoverURL(k)
	}
	return urls
}

func resolveComicURLs(c *Comic) {
	c.CoverURL = resolveCoverURL(c.CoverKey)
	if len(c.PageKeys) > 0 {
		c.PageURLs = resolvePageURLs(c.PageKeys)
	}
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
		var favCount int
		db.QueryRow(ctx, `SELECT fav_count FROM comics WHERE id = $1`, id).Scan(&favCount)
		return &ToggleFavResponse{Favorited: false, FavCount: favCount}, nil
	}

	_, err := db.Exec(ctx, `INSERT INTO favorites (user_id, comic_id) VALUES ($1, $2)`, ad.UserID, id)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "comic not found"}
	}
	db.Exec(ctx, `UPDATE comics SET fav_count = fav_count + 1 WHERE id = $1`, id)

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
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
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

	roots := buildThread(all)
	if roots == nil {
		roots = []CommentData{}
	}
	return &ListCommentsResponse{Comments: roots}, nil
}

type CreateCommentParams struct {
	BodyText string `json:"body_text"`
	ParentID string `json:"parent_id"`
}

//encore:api auth method=POST path=/comics/:id/comments
func CreateComment(ctx context.Context, id string, p *CreateCommentParams) (*CommentData, error) {
	ad := auth.Data().(*myauth.AuthData)
	if p.BodyText == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "body_text is required"}
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

// ----- Series -----

type Series struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	UploaderID  string    `json:"uploader_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateSeriesParams struct {
	Title       string `json:"title"`
	Description string `json:"description"`
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
	var s Series
	err := db.QueryRow(ctx, `
		INSERT INTO series (title, author, slug, description, uploader_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, author, slug, description, uploader_id, created_at
	`, p.Title, slug, p.Description, ad.UserID).Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type ListSeriesResponse struct {
	Series []Series `json:"series"`
}

//encore:api public method=GET path=/series
func ListSeries(ctx context.Context) (*ListSeriesResponse, error) {
	rows, err := db.Query(ctx, `SELECT id, title, author, slug, description, uploader_id, created_at FROM series ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Series
	for rows.Next() {
		var s Series
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return &ListSeriesResponse{Series: result}, rows.Err()
}

//encore:api public method=GET path=/series/:id
func GetSeries(ctx context.Context, id string) (*Series, error) {
	var s Series
	err := db.QueryRow(ctx, `SELECT id, title, author, slug, description, uploader_id, created_at FROM series WHERE id = $1`, id).Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CreatedAt)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "series not found"}
	}
	return &s, nil
}

//encore:api public method=GET path=/series/:id/comics
func SeriesComics(ctx context.Context, id string) (*ListComicsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, author, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, is_premium,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count, dislike_count,
			created_at, updated_at
		FROM comics WHERE series_id = $1 AND status = 'published'
		ORDER BY series_order ASC, published_at ASC
	`, id)
	if err != nil { return nil, err }
	defer rows.Close()
	comics, err := scanComics(rows)
	if err != nil { return nil, err }
	return &ListComicsResponse{Comics: comics, Total: len(comics)}, nil
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
		SELECT c.id, c.uploader_id, c.title, c.slug, c.description, c.content_language, c.status,
			c.cover_key, c.file_key, c.page_keys, c.file_size_bytes, c.min_tier_id, c.age_rating,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count,
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

func defaultValue(v int, def int) int {
	if v <= 0 { return def }
	return v
}

func sanitizeSortCol(sort string, allowed ...string) string {
	if sort == "" { return allowed[0] }
	for _, a := range allowed {
		if sort == a { return a }
	}
	return allowed[0]
}
