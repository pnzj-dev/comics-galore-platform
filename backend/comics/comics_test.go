package comics

import (
	"context"
	"errors"
	"testing"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
)

func uploaderCtx() context.Context {
	ctx := context.Background()
	return auth.WithContext(ctx, auth.UID("550e8400-e29b-41d4-a716-446655440001"), &myauth.AuthData{
		UserID: "550e8400-e29b-41d4-a716-446655440001",
		Email:  "uploader@example.com",
		Role:   "uploader",
		Tier:   "free",
	})
}

func adminCtx() context.Context {
	ctx := context.Background()
	return auth.WithContext(ctx, auth.UID("550e8400-e29b-41d4-a716-446655440002"), &myauth.AuthData{
		UserID: "550e8400-e29b-41d4-a716-446655440002",
		Email:  "admin@example.com",
		Role:   "admin",
		Tier:   "free",
	})
}

func moderatorCtx() context.Context {
	ctx := context.Background()
	return auth.WithContext(ctx, auth.UID("550e8400-e29b-41d4-a716-446655440003"), &myauth.AuthData{
		UserID: "550e8400-e29b-41d4-a716-446655440003",
		Email:  "mod@example.com",
		Role:   "moderator",
		Tier:   "free",
	})
}

func userCtx() context.Context {
	ctx := context.Background()
	return auth.WithContext(ctx, auth.UID("550e8400-e29b-41d4-a716-446655440004"), &myauth.AuthData{
		UserID: "550e8400-e29b-41d4-a716-446655440004",
		Email:  "user@example.com",
		Role:   "user",
		Tier:   "free",
	})
}

func TestCreateComic_Valid(t *testing.T) {
	ctx := uploaderCtx()

	comic, err := CreateComic(ctx, &CreateComicParams{
		Title:           "Test Comic",
		Description:     "A test comic",
		ContentLanguage: "en",
		CoverKey:        "covers/test-cover.jpg",
		FileKey:         "files/test-comic.cbz",
		FileSizeBytes:   1024,
		AgeRating:       "all_ages",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comic.ID == "" {
		t.Error("expected comic ID, got empty")
	}
	if comic.Status != "pending_review" {
		t.Errorf("expected status pending_review, got %s", comic.Status)
	}
	if comic.Title != "Test Comic" {
		t.Errorf("expected title Test Comic, got %s", comic.Title)
	}
	if comic.LikeCount != 0 {
		t.Errorf("expected like_count 0, got %d", comic.LikeCount)
	}
	if comic.FavCount != 0 {
		t.Errorf("expected fav_count 0, got %d", comic.FavCount)
	}
}

func TestCreateComic_MissingTitle(t *testing.T) {
	ctx := uploaderCtx()

	_, err := CreateComic(ctx, &CreateComicParams{
		Title:    "",
		CoverKey: "covers/test.jpg",
		FileKey:  "files/test.cbz",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", e.Code)
	}
}

func TestCreateComic_MissingCoverKey(t *testing.T) {
	ctx := uploaderCtx()

	_, err := CreateComic(ctx, &CreateComicParams{
		Title:   "No Cover",
		FileKey: "files/test.cbz",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", e.Code)
	}
}

func TestCreateComic_NotUploader(t *testing.T) {
	ctx := userCtx()

	_, err := CreateComic(ctx, &CreateComicParams{
		Title:    "User Comic",
		CoverKey: "covers/test.jpg",
		FileKey:  "files/test.cbz",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestCreateComic_DefaultsLanguageAndAge(t *testing.T) {
	ctx := uploaderCtx()

	comic, err := CreateComic(ctx, &CreateComicParams{
		Title:    "Defaults Comic",
		CoverKey: "covers/defaults.jpg",
		FileKey:  "files/defaults.cbz",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comic.ContentLanguage != "en" {
		t.Errorf("expected default language en, got %s", comic.ContentLanguage)
	}
	if comic.AgeRating != "all_ages" {
		t.Errorf("expected default age_rating all_ages, got %s", comic.AgeRating)
	}
}

func TestListComics_ReturnsOnlyPublished(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	ctx := uploaderCtx()
	c1, err := CreateComic(ctx, &CreateComicParams{
		Title:    "Published Comic",
		CoverKey: "covers/pub.jpg",
		FileKey:  "files/pub.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	c2, err := CreateComic(ctx, &CreateComicParams{
		Title:    "Pending Comic",
		CoverKey: "covers/pend.jpg",
		FileKey:  "files/pend.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx()
	if err := ApproveComic(modCtx, c1.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	pubCtx := context.Background()
	resp, err := ListComics(pubCtx, &ListComicsParams{})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}

	for _, c := range resp.Comics {
		if c.ID == c2.ID {
			t.Error("pending comic should not appear in public list")
		}
	}

	found := false
	for _, c := range resp.Comics {
		if c.ID == c1.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("published comic should appear in list")
	}
}

func TestGetComic_IncrementsViewCount(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	ctx := uploaderCtx()
	comic, err := CreateComic(ctx, &CreateComicParams{
		Title:    "View Comic",
		CoverKey: "covers/view.jpg",
		FileKey:  "files/view.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx()
	if err := ApproveComic(modCtx, comic.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	pubCtx := context.Background()
	fetched, err := GetComic(pubCtx, comic.ID)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	fetchedView, err := GetComic(pubCtx, comic.ID)
	if err != nil {
		t.Fatalf("get again error: %v", err)
	}
	if fetchedView.ViewCount < 1 {
		t.Errorf("expected view_count >= 1 on second fetch, got %d", fetchedView.ViewCount)
	}
	_ = fetched

	fetched2, err := GetComic(pubCtx, comic.ID)
	if err != nil {
		t.Fatalf("second get error: %v", err)
	}
	if fetched2.ViewCount < fetched.ViewCount {
		t.Errorf("view_count should increase: was %d, now %d", fetched.ViewCount, fetched2.ViewCount)
	}
}

func TestGetComic_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := GetComic(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestToggleLike(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Like Comic",
		CoverKey: "covers/like.jpg",
		FileKey:  "files/like.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx()

	resp1, err := ToggleLike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("toggle like error: %v", err)
	}
	if !resp1.Liked {
		t.Error("expected liked=true after first toggle")
	}
	if resp1.LikeCount != 1 {
		t.Errorf("expected like_count 1, got %d", resp1.LikeCount)
	}

	resp2, err := ToggleLike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("toggle unlike error: %v", err)
	}
	if resp2.Liked {
		t.Error("expected liked=false after second toggle")
	}
	if resp2.LikeCount != 0 {
		t.Errorf("expected like_count 0, got %d", resp2.LikeCount)
	}
}

func TestToggleFavorite(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Fav Comic",
		CoverKey: "covers/fav.jpg",
		FileKey:  "files/fav.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx()

	resp1, err := ToggleFavorite(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("toggle fav error: %v", err)
	}
	if !resp1.Favorited {
		t.Error("expected favorited=true after first toggle")
	}
	if resp1.FavCount != 1 {
		t.Errorf("expected fav_count 1, got %d", resp1.FavCount)
	}

	resp2, err := ToggleFavorite(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("toggle unfav error: %v", err)
	}
	if resp2.Favorited {
		t.Error("expected favorited=false after second toggle")
	}
	if resp2.FavCount != 0 {
		t.Errorf("expected fav_count 0, got %d", resp2.FavCount)
	}
}

func TestApproveComic_ChangesStatusToPublished(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Approve Comic",
		CoverKey: "covers/approve.jpg",
		FileKey:  "files/approve.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	if comic.Status != "pending_review" {
		t.Fatalf("expected pending_review, got %s", comic.Status)
	}

	modCtx := moderatorCtx()
	if err := ApproveComic(modCtx, comic.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	pubCtx := context.Background()
	fetched, err := GetComic(pubCtx, comic.ID)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if fetched.Status != "published" {
		t.Errorf("expected published, got %s", fetched.Status)
	}
}

func TestApproveComic_RequiresModeratorOrAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "NoPerm Comic",
		CoverKey: "covers/noperm.jpg",
		FileKey:  "files/noperm.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx()
	err = ApproveComic(uCtx, comic.ID)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}

func TestRejectComic_ChangesStatusToRejected(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Reject Comic",
		CoverKey: "covers/reject.jpg",
		FileKey:  "files/reject.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx()
	if err := RejectComic(modCtx, comic.ID, &RejectParams{Reason: "low quality"}); err != nil {
		t.Fatalf("reject error: %v", err)
	}

	fetched, err := GetComic(uploaderCtx(), comic.ID)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if fetched.Status != "rejected" {
		t.Errorf("expected rejected, got %s", fetched.Status)
	}
	if fetched.RejectionReason != "low quality" {
		t.Errorf("expected rejection_reason 'low quality', got '%s'", fetched.RejectionReason)
	}
}

func TestAdminListComics_ReturnsAllComics(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	_, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Admin Comic 1",
		CoverKey: "covers/admin1.jpg",
		FileKey:  "files/admin1.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_, err = CreateComic(upCtx, &CreateComicParams{
		Title:    "Admin Comic 2",
		CoverKey: "covers/admin2.jpg",
		FileKey:  "files/admin2.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx()
	_ = ApproveComic(modCtx, "550e8400-e29b-41d4-a716-446655440000")

	admCtx := adminCtx()
	resp, err := AdminListComics(admCtx)
	if err != nil {
		t.Fatalf("admin list error: %v", err)
	}
	if resp.Total < 2 {
		t.Errorf("expected at least 2 comics in admin list, got %d", resp.Total)
	}
}

func TestAdminListComics_RequiresAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	uCtx := userCtx()
	_, err := AdminListComics(uCtx)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}

func TestGetLikeStatus(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx()
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Status Comic",
		CoverKey: "covers/status.jpg",
		FileKey:  "files/status.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx()
	status, err := GetLikeStatus(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if status.Liked {
		t.Error("expected not liked initially")
	}
	if status.Favorited {
		t.Error("expected not favorited initially")
	}

	ToggleLike(uCtx, comic.ID)
	status, err = GetLikeStatus(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if !status.Liked {
		t.Error("expected liked after toggle")
	}
}
