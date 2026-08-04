package reading

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	myauth "comics-galore/backend/auth"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("readingdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type Progress struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ComicID     string    `json:"comic_id"`
	CurrentPage int       `json:"current_page"`
	TotalPages  int       `json:"total_pages"`
	Completed   bool      `json:"completed"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SaveProgressParams struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	Completed   bool `json:"completed"`
}

//encore:api auth method=POST path=/reading/:comicId
func SaveProgress(ctx context.Context, comicId string, p *SaveProgressParams) (*Progress, error) {
	ad := auth.Data().(*myauth.AuthData)

	var progress Progress
	err := db.QueryRow(ctx, `
		INSERT INTO reading_progress (user_id, comic_id, current_page, total_pages, completed)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, comic_id)
		DO UPDATE SET current_page = $3, total_pages = $4, completed = $5, updated_at = now()
		RETURNING id, user_id, comic_id, current_page, total_pages, completed, updated_at
	`, ad.UserID, comicId, p.CurrentPage, p.TotalPages, p.Completed).Scan(
		&progress.ID, &progress.UserID, &progress.ComicID,
		&progress.CurrentPage, &progress.TotalPages, &progress.Completed, &progress.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

//encore:api auth method=GET path=/reading/:comicId
func GetProgress(ctx context.Context, comicId string) (*Progress, error) {
	ad := auth.Data().(*myauth.AuthData)

	var progress Progress
	err := db.QueryRow(ctx, `
		SELECT id, user_id, comic_id, current_page, total_pages, completed, updated_at
		FROM reading_progress WHERE user_id = $1 AND comic_id = $2
	`, ad.UserID, comicId).Scan(
		&progress.ID, &progress.UserID, &progress.ComicID,
		&progress.CurrentPage, &progress.TotalPages, &progress.Completed, &progress.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "no progress found"}
		}
		return nil, err
	}
	return &progress, nil
}

type ContinueReadingItem struct {
	ComicID     string    `json:"comic_id"`
	CurrentPage int       `json:"current_page"`
	TotalPages  int       `json:"total_pages"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ContinueReadingResponse struct {
	Items []ContinueReadingItem `json:"items"`
}

//encore:api auth method=GET path=/reading-continue
func ContinueReading(ctx context.Context) (*ContinueReadingResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	rows, err := db.Query(ctx, `
		SELECT comic_id, current_page, total_pages, updated_at
		FROM reading_progress WHERE user_id = $1 AND completed = false
		ORDER BY updated_at DESC LIMIT 20
	`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ContinueReadingItem
	for rows.Next() {
		var item ContinueReadingItem
		if err := rows.Scan(&item.ComicID, &item.CurrentPage, &item.TotalPages, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return &ContinueReadingResponse{Items: items}, rows.Err()
}

type DownloadResponse struct {
	Allowed    bool   `json:"allowed"`
	Used       int    `json:"used"`
	Limit      int    `json:"limit"`
	Message    string `json:"message,omitempty"`
}

//encore:api auth method=POST path=/comics/:comicId/download
func RecordDownload(ctx context.Context, comicId string) (*DownloadResponse, error) {
	ad := auth.Data().(*myauth.AuthData)

	limit := quotaLimit(ad.Tier)
	used, err := countDownloadsThisMonth(ctx, ad.UserID)
	if err != nil {
		used = 0
	}

	if used >= limit {
		return &DownloadResponse{
			Allowed: false,
			Used:    used,
			Limit:   limit,
			Message: "Download quota exceeded. Upgrade your plan to download more.",
		}, nil
	}

	_, err = db.Exec(ctx, `INSERT INTO download_logs (user_id, comic_id) VALUES ($1, $2)`, ad.UserID, comicId)
	if err != nil {
		return nil, err
	}

	return &DownloadResponse{
		Allowed: true,
		Used:    used + 1,
		Limit:   limit,
	}, nil
}

func quotaLimit(tier string) int {
	switch tier {
	case "free":
		return 5
	case "bronze":
		return 50
	case "silver":
		return 200
	case "gold":
		return 999999
	case "platinum":
		return 999999
	default:
		return 5
	}
}

func countDownloadsThisMonth(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM download_logs
		WHERE user_id = $1 AND created_at >= date_trunc('month', now())
	`, userID).Scan(&count)
	return count, err
}
