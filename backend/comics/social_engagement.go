package comics

import (
	"context"
	"database/sql"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"
	"encore.dev/beta/errs"
)

// ----- Reading lists (public shelves) -----

type ReadingList struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReadingListParams struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type ListReadingListsResponse struct {
	Lists []ReadingList `json:"lists"`
}

type AddToListParams struct {
	ComicID string `json:"comic_id"`
}

//encore:api auth method=POST path=/reading-lists
func CreateReadingList(ctx context.Context, p *CreateReadingListParams) (*ReadingList, error) {
	ad := auth.Data().(*myauth.AuthData)
	if p.Name == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "name is required"}
	}

	var l ReadingList
	err := db.QueryRow(ctx, `
		INSERT INTO reading_lists (user_id, name, is_public)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, is_public, created_at
	`, ad.UserID, p.Name, p.IsPublic).Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

//encore:api auth method=GET path=/reading-lists
func ListReadingLists(ctx context.Context) (*ListReadingListsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	rows, err := db.Query(ctx, `SELECT id, user_id, name, is_public, created_at FROM reading_lists WHERE user_id = $1 ORDER BY created_at DESC`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []ReadingList
	for rows.Next() {
		var l ReadingList
		if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	if lists == nil {
		lists = []ReadingList{}
	}
	return &ListReadingListsResponse{Lists: lists}, rows.Err()
}

type PublicReadingListResponse struct {
	List   ReadingList `json:"list"`
	Comics []Comic     `json:"comics"`
}

//encore:api public method=GET path=/reading-lists/:id
func GetReadingList(ctx context.Context, id string) (*PublicReadingListResponse, error) {
	var l ReadingList
	err := db.QueryRow(ctx, `SELECT id, user_id, name, is_public, created_at FROM reading_lists WHERE id = $1 AND is_public = true`, id).Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "list not found"}
		}
		return nil, err
	}

	rows, err := db.Query(ctx, `
		SELECT c.id, c.uploader_id, c.title, c.author, c.slug, c.description, c.content_language, c.status,
			c.cover_key, c.file_key, c.page_keys, c.file_size_bytes, c.min_tier_id, c.age_rating, c.is_premium,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count, c.dislike_count,
			COALESCE(c.series_order, 1), c.created_at, c.updated_at
		FROM reading_list_items rli
		JOIN comics c ON c.id = rli.comic_id
		WHERE rli.list_id = $1 AND c.status = 'published'
		ORDER BY rli.position ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comics []Comic
	for rows.Next() {
		var c Comic
		var pubAt sql.NullTime
		if err := rows.Scan(
			&c.ID, &c.UploaderID, &c.Title, &c.Author, &c.Slug, &c.Description,
			&c.ContentLanguage, &c.Status, &c.CoverKey, &c.FileKey,
			scanStringSlice(&c.PageKeys), &c.FileSizeBytes, nulString(&c.MinTierID),
			&c.AgeRating, &c.IsPremium, scanStringSlice(&c.Tags), nulString(&c.RejectionReason),
			&pubAt, &c.ViewCount, &c.DownloadCount, &c.LikeCount, &c.FavCount, &c.DislikeCount,
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
	if comics == nil {
		comics = []Comic{}
	}
	return &PublicReadingListResponse{List: l, Comics: comics}, rows.Err()
}

//encore:api auth method=POST path=/reading-lists/:id/items
func AddToReadingList(ctx context.Context, id string, p *AddToListParams) error {
	ad := auth.Data().(*myauth.AuthData)
	var owner string
	if err := db.QueryRow(ctx, `SELECT user_id FROM reading_lists WHERE id = $1`, id).Scan(&owner); err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "list not found"}
		}
		return err
	}
	if owner != ad.UserID {
		return &errs.Error{Code: errs.PermissionDenied, Message: "not your list"}
	}

	var maxPos int
	db.QueryRow(ctx, `SELECT COALESCE(MAX(position), -1) FROM reading_list_items WHERE list_id = $1`, id).Scan(&maxPos)
	_, err := db.Exec(ctx, `INSERT INTO reading_list_items (list_id, comic_id, position) VALUES ($1, $2, $3) ON CONFLICT (list_id, comic_id) DO NOTHING`, id, p.ComicID, maxPos+1)
	return err
}

//encore:api auth method=DELETE path=/reading-lists/:id/items/:comicId
func RemoveFromReadingList(ctx context.Context, id, comicId string) error {
	ad := auth.Data().(*myauth.AuthData)
	var owner string
	if err := db.QueryRow(ctx, `SELECT user_id FROM reading_lists WHERE id = $1`, id).Scan(&owner); err != nil {
		if isNoRows(err) {
			return &errs.Error{Code: errs.NotFound, Message: "list not found"}
		}
		return err
	}
	if owner != ad.UserID {
		return &errs.Error{Code: errs.PermissionDenied, Message: "not your list"}
	}

	_, err := db.Exec(ctx, `DELETE FROM reading_list_items WHERE list_id = $1 AND comic_id = $2`, id, comicId)
	return err
}

// ----- "People also liked" -----

//encore:api public method=GET path=/comics/:id/related
func RelatedComics(ctx context.Context, id string) (*ListComicsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT c.id, c.uploader_id, c.title, c.author, c.slug, c.description, c.content_language, c.status,
			c.cover_key, c.file_key, c.page_keys, c.file_size_bytes, c.min_tier_id, c.age_rating, c.is_premium,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count, c.dislike_count,
			created_at, updated_at
		FROM comics c
		WHERE c.id != $1 AND c.status = 'published'
		  AND c.id IN (
			SELECT l2.comic_id FROM likes l2
			WHERE l2.user_id IN (SELECT user_id FROM likes WHERE comic_id = $1)
		  )
		ORDER BY c.like_count DESC
		LIMIT 6
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics, err := scanComics(rows)
	if err != nil {
		return nil, err
	}
	if comics == nil {
		comics = []Comic{}
	}
	return &ListComicsResponse{Comics: comics, Total: len(comics)}, nil
}
