package comics

import (
	"context"

	"encore.dev/beta/errs"
)

// ----- Staff Picks -----

type StaffPick struct {
	ComicID string `json:"comic_id"`
	Title   string `json:"title"`
	Cover   string `json:"cover_url"`
	Slug    string `json:"slug"`
	Position int   `json:"position"`
}

type ListStaffPicksResponse struct {
	Picks []StaffPick `json:"picks"`
}

//encore:api public method=GET path=/staff-picks
func ListStaffPicks(ctx context.Context) (*ListStaffPicksResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT sp.comic_id, c.title, c.slug, c.cover_key, sp.position
		FROM staff_picks sp
		JOIN comics c ON c.id = sp.comic_id
		WHERE c.status = 'published'
		ORDER BY sp.position ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var picks []StaffPick
	for rows.Next() {
		var p StaffPick
		var coverKey string
		if err := rows.Scan(&p.ComicID, &p.Title, &p.Slug, &coverKey, &p.Position); err != nil {
			return nil, err
		}
		p.Cover = resolveCoverURL(coverKey)
		picks = append(picks, p)
	}
	if picks == nil {
		picks = []StaffPick{}
	}
	return &ListStaffPicksResponse{Picks: picks}, rows.Err()
}

type AddStaffPickParams struct {
	ComicID string `json:"comic_id"`
}

//encore:api auth method=POST path=/admin/staff-picks
func AddStaffPick(ctx context.Context, p *AddStaffPickParams) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	var maxPos int
	db.QueryRow(ctx, `SELECT COALESCE(MAX(position), -1) FROM staff_picks`).Scan(&maxPos)
	_, err := db.Exec(ctx, `INSERT INTO staff_picks (comic_id, position) VALUES ($1, $2) ON CONFLICT (comic_id) DO NOTHING`, p.ComicID, maxPos+1)
	return err
}

//encore:api auth method=DELETE path=/admin/staff-picks/:comicId
func RemoveStaffPick(ctx context.Context, comicId string) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `DELETE FROM staff_picks WHERE comic_id = $1`, comicId)
	return err
}

// ----- Saved datalist views -----

type SavedView struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Name     string `json:"name"`
	Filters  string `json:"filters"`
}

type ListSavedViewsResponse struct {
	Views []SavedView `json:"views"`
}

type SaveViewParams struct {
	Resource string `json:"resource"`
	Name     string `json:"name"`
	Filters  string `json:"filters"`
}

//encore:api auth method=GET path=/admin/views
func ListSavedViews(ctx context.Context) (*ListSavedViewsResponse, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	rows, err := db.Query(ctx, `SELECT id, resource, name, COALESCE(filters::text, '') FROM saved_views WHERE admin_id = $1 ORDER BY created_at DESC`, ad.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []SavedView
	for rows.Next() {
		var v SavedView
		if err := rows.Scan(&v.ID, &v.Resource, &v.Name, &v.Filters); err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	if views == nil {
		views = []SavedView{}
	}
	return &ListSavedViewsResponse{Views: views}, rows.Err()
}

//encore:api auth method=POST path=/admin/views
func SaveView(ctx context.Context, p *SaveViewParams) (*SavedView, error) {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	if p.Name == "" || p.Resource == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "name and resource are required"}
	}

	var v SavedView
	err := db.QueryRow(ctx, `
		INSERT INTO saved_views (admin_id, resource, name, filters)
		VALUES ($1, $2, $3, $4::jsonb)
		RETURNING id, resource, name, COALESCE(filters::text, '')
	`, ad.UserID, p.Resource, p.Name, p.Filters).Scan(&v.ID, &v.Resource, &v.Name, &v.Filters)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

//encore:api auth method=DELETE path=/admin/views/:id
func DeleteSavedView(ctx context.Context, id string) error {
	ad, hasAuth := getAuthData(ctx)
	if !hasAuth || ad.Role != "admin" {
		return &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	_, err := db.Exec(ctx, `DELETE FROM saved_views WHERE id = $1 AND admin_id = $2`, id, ad.UserID)
	return err
}
