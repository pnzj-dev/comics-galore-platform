package comics

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("comicsdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Comic is the public representation
type Comic struct {
	ID              string    `json:"id"`
	UploaderID      string    `json:"uploader_id"`
	Title           string    `json:"title"`
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
	Tags            []string  `json:"tags"`
	RejectionReason string    `json:"rejection_reason,omitempty"`
	PublishedAt     time.Time `json:"published_at,omitempty"`
	ViewCount       int64     `json:"view_count"`
	DownloadCount   int64     `json:"download_count"`
	LikeCount       int       `json:"like_count"`
	FavCount        int       `json:"fav_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateComicParams struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	ContentLanguage string   `json:"content_language"`
	CoverKey        string   `json:"cover_key"`
	FileKey         string   `json:"file_key"`
	PageKeys        []string `json:"page_keys"`
	FileSizeBytes   int64    `json:"file_size_bytes"`
	MinTierID       string   `json:"min_tier_id"`
	AgeRating       string   `json:"age_rating"`
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
		INSERT INTO comics (uploader_id, title, slug, description, content_language,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, view_count, download_count, like_count, fav_count,
			created_at, updated_at
	`, ad.UserID, p.Title, slug, p.Description, lang,
		p.CoverKey, p.FileKey, pageKeys, p.FileSizeBytes, minTierID, ageRating, tags).Scan(
		&comic.ID, &comic.UploaderID, &comic.Title, &comic.Slug, &comic.Description,
		&comic.ContentLanguage, &comic.Status, &comic.CoverKey, &comic.FileKey,
		scanStringSlice(&comic.PageKeys), &comic.FileSizeBytes, nulString(&comic.MinTierID),
		&comic.AgeRating, scanStringSlice(&comic.Tags), nulString(&comic.RejectionReason),
		&comic.ViewCount, &comic.DownloadCount, &comic.LikeCount, &comic.FavCount,
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
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	Language string `query:"language"`
	Search   string `query:"search"`
	Tag      string `query:"tag"`
	Sort     string `query:"sort"`
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
	default:
		orderBy = "published_at DESC"
	}

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count,
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

	return &ListComicsResponse{Comics: comics, Total: total}, nil
}

//encore:api public method=GET path=/comics/:id
func GetComic(ctx context.Context, id string) (*Comic, error) {
	ad, hasAuth := getAuthData(ctx)

	var comic Comic
	err := db.QueryRow(ctx, `
		SELECT id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, view_count, download_count, like_count, fav_count,
			created_at, updated_at
		FROM comics WHERE id = $1
	`, id).Scan(
		&comic.ID, &comic.UploaderID, &comic.Title, &comic.Slug, &comic.Description,
		&comic.ContentLanguage, &comic.Status, &comic.CoverKey, &comic.FileKey,
		scanStringSlice(&comic.PageKeys), &comic.FileSizeBytes, nulString(&comic.MinTierID),
		&comic.AgeRating, scanStringSlice(&comic.Tags), nulString(&comic.RejectionReason),
		&comic.ViewCount, &comic.DownloadCount, &comic.LikeCount, &comic.FavCount,
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

	db.Exec(ctx, `UPDATE comics SET view_count = view_count + 1 WHERE id = $1`, id)

	return &comic, nil
}

//encore:api auth method=GET path=/uploader/comics
func MyComics(ctx context.Context) (*ListComicsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, view_count, download_count, like_count, fav_count,
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

	return &ListComicsResponse{Comics: comics, Total: len(comics)}, nil
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

//encore:api auth method=GET path=/moderation/comics
func PendingComics(ctx context.Context) (*ListPendingResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || (ad.Role != "moderator" && ad.Role != "admin") {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "requires moderator or admin"}
	}

	var total int
	db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE status = 'pending_review'`).Scan(&total)

	rows, err := db.Query(ctx, `
		SELECT id, title, uploader_id, status, created_at
		FROM comics WHERE status = 'pending_review'
		ORDER BY created_at ASC
	`)
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
}

//encore:api auth method=GET path=/comics/:id/like-status
func GetLikeStatus(ctx context.Context, id string) (*LikeStatus, error) {
	ad := auth.Data().(*myauth.AuthData)

	var liked, favorited bool
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&liked)
	db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND comic_id = $2)`, ad.UserID, id).Scan(&favorited)

	return &LikeStatus{Liked: liked, Favorited: favorited}, nil
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

// ----- Admin comic list -----

//encore:api auth method=GET path=/admin/comics
func AdminListComics(ctx context.Context) (*ListComicsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count,
			created_at, updated_at
		FROM comics ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil {
		return nil, err
	}

	return &ListComicsResponse{Comics: comics, Total: len(comics)}, nil
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
	return &c, nil
}

func buildThread(comments []CommentData) []CommentData {
	children := make(map[string][]CommentData)
	for _, c := range comments {
		if c.ParentID != "" {
			children[c.ParentID] = append(children[c.ParentID], c)
		}
	}
	var result []CommentData
	for _, c := range comments {
		if c.ParentID == "" {
			c.Replies = children[c.ID]
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
		INSERT INTO series (title, slug, description, uploader_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, slug, description, uploader_id, created_at
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
	rows, err := db.Query(ctx, `SELECT id, title, slug, description, uploader_id, created_at FROM series ORDER BY created_at DESC`)
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
	err := db.QueryRow(ctx, `SELECT id, title, slug, description, uploader_id, created_at FROM series WHERE id = $1`, id).Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CreatedAt)
	if err != nil {
		return nil, &errs.Error{Code: errs.NotFound, Message: "series not found"}
	}
	return &s, nil
}

//encore:api public method=GET path=/series/:id/comics
func SeriesComics(ctx context.Context, id string) (*ListComicsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, published_at, view_count, download_count, like_count, fav_count,
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
