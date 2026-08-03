package comics

import (
	"context"
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

	err := db.QueryRow(ctx, `
		INSERT INTO comics (uploader_id, title, slug, description, content_language,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, nullif($10,''), $11, $12)
		RETURNING id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, view_count, download_count, like_count, fav_count,
			created_at, updated_at
	`, ad.UserID, p.Title, slug, p.Description, lang,
		p.CoverKey, p.FileKey, pageKeys, p.FileSizeBytes, p.MinTierID, ageRating, tags).Scan(
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

	var total int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM comics WHERE status = 'published'`).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, `
		SELECT id, uploader_id, title, slug, description, content_language, status,
			cover_key, file_key, page_keys, file_size_bytes, min_tier_id, age_rating,
			tags, rejection_reason, view_count, download_count, like_count, fav_count,
			created_at, updated_at
		FROM comics WHERE status = 'published'
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
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
	return err
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
	return err
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
		return &ToggleLikeResponse{Liked: false, LikeCount: 0}, nil
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
		return &ToggleFavResponse{Favorited: false, FavCount: 0}, nil
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
