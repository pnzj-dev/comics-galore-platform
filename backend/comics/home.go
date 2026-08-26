package comics

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"
)

// AdBanner is a reserved advertising slot. Real ad content is provided by an
// ad agency integration; for now it's a placeholder.
type AdBanner struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	CTAText  string `json:"cta_text"`
	CTAHref  string `json:"cta_href"`
}

// HomeResponse is the full homepage payload.
type HomeResponse struct {
	Ad                AdBanner `json:"ad"`
	Categories        []string `json:"categories"`
	PopularByCategory []Series `json:"popular_by_category"`
	NewlyReleased     []Series `json:"newly_released"`
	DailySeries       []Series `json:"daily_series"`
	IndieSeries       []Series `json:"indie_series"`
}

//encore:api public method=GET path=/home
func GetHome(ctx context.Context) (*HomeResponse, error) {
	categories, err := listCategories(ctx)
	if err != nil {
		return nil, err
	}
	popular, err := listSeriesOrdered(ctx, `views_count`, 10)
	if err != nil {
		return nil, err
	}
	newly, err := listSeriesOrdered(ctx, `created_at`, 10)
	if err != nil {
		return nil, err
	}
	daily, err := listSeriesWhere(ctx, `schedule_day IS NOT NULL AND schedule_day != ''`, 20)
	if err != nil {
		return nil, err
	}
	// Indie series = series by uploader (every series has an uploader), ranked
	// by aggregate views.
	indie, err := listSeriesOrdered(ctx, `views_count`, 20)
	if err != nil {
		return nil, err
	}

	return &HomeResponse{
		Ad: AdBanner{
			Title:    "Your story belongs here",
			Subtitle: "Advertisement",
			CTAText:  "Learn more",
			CTAHref:  "#",
		},
		Categories:        categories,
		PopularByCategory: popular,
		NewlyReleased:     newly,
		DailySeries:       daily,
		IndieSeries:       indie,
	}, nil
}

// listCategories returns the distinct non-empty category values.
func listCategories(ctx context.Context) ([]string, error) {
	rows, err := db.Query(ctx, `SELECT DISTINCT category FROM series WHERE category <> '' ORDER BY category ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []string{}
	}
	return cats, rows.Err()
}

// seriesColumns is the shared SELECT list for home series queries, matching
// the Series struct scan order (rank is computed, not selected).
const seriesColumns = `id, title, slug, description, uploader_id, cover_key, genre, category, overlay_title, views_count, hearts_count, schedule_day, created_at`

// listSeriesOrdered returns series ordered by the given column (descending).
func listSeriesOrdered(ctx context.Context, orderCol string, limit int) ([]Series, error) {
	rows, err := db.Query(ctx, `
		SELECT `+seriesColumns+`
		FROM series
		ORDER BY `+orderCol+` DESC, created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSeriesRows(rows, false)
}

// listSeriesWhere returns series matching an extra WHERE clause.
func listSeriesWhere(ctx context.Context, where string, limit int) ([]Series, error) {
	rows, err := db.Query(ctx, `
		SELECT `+seriesColumns+`
		FROM series
		WHERE `+where+`
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSeriesRows(rows, false)
}

// scanSeriesRows scans series rows into a slice, resolving cover URLs and
// optionally assigning computed ranks.
func scanSeriesRows(rows *sqldb.Rows, rank bool) ([]Series, error) {
	var result []Series
	i := 0
	for rows.Next() {
		var s Series
		if err := scanSeries(&s, rows, rank, i+1); err != nil {
			return nil, err
		}
		i++
		result = append(result, s)
	}
	if result == nil {
		result = []Series{}
	}
	return result, rows.Err()
}

// scanSeries scans a single series row (shared by home + ranked queries).
func scanSeries(s *Series, rows *sqldb.Rows, rank bool, rankVal int) error {
	var scheduleDay sql.NullString
	if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.UploaderID, &s.CoverKey, &s.Genre, &s.Category, &s.OverlayTitle, &s.ViewsCount, &s.HeartsCount, &scheduleDay, &s.CreatedAt); err != nil {
		return err
	}
	s.ScheduleDay = scheduleDay.String
	if rank {
		s.Rank = rankVal
	}
	if s.OverlayTitle == "" {
		s.OverlayTitle = s.Title
	}
	s.CoverURL = resolveCoverURL(s.CoverKey)
	return nil
}
