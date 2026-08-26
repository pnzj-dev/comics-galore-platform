package comics

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"
	"encore.dev/beta/errs"
)

// ----- Reading lists (public shelves) -----

type ReadingList struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Name       string    `json:"name"`
	IsPublic   bool      `json:"is_public"`
	CreatedAt  time.Time `json:"created_at"`
	ComicCount int       `json:"comic_count"`
	HasComic   bool      `json:"has_comic"`
}

type CreateReadingListParams struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type UpdateReadingListParams struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type ListReadingListsParams struct {
	ComicID string `query:"comic_id"`
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
func ListReadingLists(ctx context.Context, p *ListReadingListsParams) (*ListReadingListsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	rows, err := db.Query(ctx, `
		SELECT rl.id, rl.user_id, rl.name, rl.is_public, rl.created_at,
			COUNT(rli.comic_id) AS comic_count,
			COALESCE(BOOL_OR(rli.comic_id = $2), false) AS has_comic
		FROM reading_lists rl
		LEFT JOIN reading_list_items rli ON rli.list_id = rl.id
		WHERE rl.user_id = $1
		GROUP BY rl.id
		ORDER BY rl.created_at DESC
	`, ad.UserID, nulOrValue(p.ComicID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []ReadingList
	for rows.Next() {
		var l ReadingList
		if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt, &l.ComicCount, &l.HasComic); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	if lists == nil {
		lists = []ReadingList{}
	}
	return &ListReadingListsResponse{Lists: lists}, rows.Err()
}

//encore:api auth method=PATCH path=/reading-lists/:id
func UpdateReadingList(ctx context.Context, id string, p *UpdateReadingListParams) (*ReadingList, error) {
	ad := auth.Data().(*myauth.AuthData)
	var l ReadingList
	if err := db.QueryRow(ctx, `SELECT id, user_id, name, is_public, created_at FROM reading_lists WHERE id = $1`, id).Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "list not found"}
		}
		return nil, err
	}
	if l.UserID != ad.UserID {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "not your list"}
	}

	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "name is required"}
	}

	if _, err := db.Exec(ctx, `UPDATE reading_lists SET name = $1, is_public = $2 WHERE id = $3`, name, p.IsPublic, id); err != nil {
		return nil, err
	}
	l.Name = name
	l.IsPublic = p.IsPublic
	return &l, nil
}

//encore:api auth method=DELETE path=/reading-lists/:id
func DeleteReadingList(ctx context.Context, id string) error {
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
	_, err := db.Exec(ctx, `DELETE FROM reading_lists WHERE id = $1`, id)
	return err
}

type PublicReadingListResponse struct {
	List   ReadingList `json:"list"`
	Comics []Comic     `json:"comics"`
	Total  int         `json:"total"`
}

type GetReadingListParams struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

// listReadingListComics returns the published comics of a list, paginated by
// position, with resolved URLs.
func listReadingListComics(ctx context.Context, listID string, page, limit int) ([]Comic, int, error) {
	offset := (page - 1) * limit

	var total int
	db.QueryRow(ctx, `
		SELECT COUNT(*) FROM reading_list_items rli
		JOIN comics c ON c.id = rli.comic_id
		WHERE rli.list_id = $1 AND c.status = 'published'`+matureWhereClause(ctx)+`
	`, listID).Scan(&total)

	rows, err := db.Query(ctx, `
		SELECT c.id, c.uploader_id, c.title, c.author, c.slug, c.description, c.content_language, c.status,
			c.category, c.genre, c.cover_key, c.file_key, c.file_size_bytes, c.min_tier_id, c.age_rating, c.is_premium,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count, c.dislike_count,
			c.reading_direction, c.page_count, c.archive_mimetype, c.isbn, c.upc, c.issn, c.volume, c.issue_number,
			COALESCE(c.series_order, 1), c.created_at, c.updated_at
		FROM reading_list_items rli
		JOIN comics c ON c.id = rli.comic_id
		WHERE rli.list_id = $1 AND c.status = 'published'`+matureWhereClause(ctx)+`
		ORDER BY rli.position ASC
		LIMIT $2 OFFSET $3
	`, listID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
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
			return nil, 0, err
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
	return comics, total, rows.Err()
}

//encore:api public method=GET path=/reading-lists/:id
func GetReadingList(ctx context.Context, id string, p *GetReadingListParams) (*PublicReadingListResponse, error) {
	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 50 {
		limit = 50
	}

	var l ReadingList
	err := db.QueryRow(ctx, `SELECT id, user_id, name, is_public, created_at FROM reading_lists WHERE id = $1 AND is_public = true`, id).Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "list not found"}
		}
		return nil, err
	}

	comics, total, err := listReadingListComics(ctx, id, page, limit)
	if err != nil {
		return nil, err
	}
	return &PublicReadingListResponse{List: l, Comics: comics, Total: total}, nil
}

//encore:api auth method=GET path=/reading-lists/:id/mine
func GetMyReadingList(ctx context.Context, id string, p *GetReadingListParams) (*PublicReadingListResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	page := defaultValue(p.Page, 1)
	limit := defaultValue(p.Limit, 20)
	if limit > 50 {
		limit = 50
	}

	var l ReadingList
	if err := db.QueryRow(ctx, `SELECT id, user_id, name, is_public, created_at FROM reading_lists WHERE id = $1`, id).Scan(&l.ID, &l.UserID, &l.Name, &l.IsPublic, &l.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "list not found"}
		}
		return nil, err
	}
	if l.UserID != ad.UserID {
		return nil, &errs.Error{Code: errs.NotFound, Message: "list not found"}
	}

	comics, total, err := listReadingListComics(ctx, id, page, limit)
	if err != nil {
		return nil, err
	}
	return &PublicReadingListResponse{List: l, Comics: comics, Total: total}, nil
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
			c.category, c.genre, c.cover_key, c.file_key, c.file_size_bytes, c.min_tier_id, c.age_rating, c.is_premium,
			c.tags, c.rejection_reason, c.published_at, c.view_count, c.download_count, c.like_count, c.fav_count, c.dislike_count,
			c.reading_direction, c.page_count, c.archive_mimetype, c.isbn, c.upc, c.issn, c.volume, c.issue_number,
			created_at, updated_at
		FROM comics c
		WHERE c.id != $1 AND c.status = 'published'`+matureWhereClause(ctx)+`
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
