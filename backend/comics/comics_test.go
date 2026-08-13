package comics

import (
	"context"
	"errors"
	"testing"

	"comics-galore/backend/fixtures"

	"encore.dev/beta/errs"
	"encore.dev/et"
)

var uploaderCtx, adminCtx, moderatorCtx, userCtx context.Context

func init() {
	uploaderCtx = fixtures.UploaderCtx()
	adminCtx = fixtures.AdminCtx()
	moderatorCtx = fixtures.ModeratorCtx()
	userCtx = fixtures.UserCtx()
}

func TestCreateComic_Valid(t *testing.T) {
	ctx := uploaderCtx

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
	ctx := uploaderCtx

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
	ctx := uploaderCtx

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
	ctx := userCtx

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
	ctx := uploaderCtx

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

	ctx := uploaderCtx
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

	modCtx := moderatorCtx
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

	ctx := uploaderCtx
	comic, err := CreateComic(ctx, &CreateComicParams{
		Title:    "View Comic",
		CoverKey: "covers/view.jpg",
		FileKey:  "files/view.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx
	if err := ApproveComic(modCtx, comic.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	pubCtx := context.Background()
	fetched, err := GetComic(pubCtx, comic.Slug)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	fetchedView, err := GetComic(pubCtx, comic.Slug)
	if err != nil {
		t.Fatalf("get again error: %v", err)
	}
	if fetchedView.ViewCount < 1 {
		t.Errorf("expected view_count >= 1 on second fetch, got %d", fetchedView.ViewCount)
	}
	_ = fetched

	fetched2, err := GetComic(pubCtx, comic.Slug)
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

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Like Comic",
		CoverKey: "covers/like.jpg",
		FileKey:  "files/like.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx

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

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Fav Comic",
		CoverKey: "covers/fav.jpg",
		FileKey:  "files/fav.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx

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

func TestToggleDislike(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Dislike Comic",
		CoverKey: "covers/dislike.jpg",
		FileKey:  "files/dislike.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx

	resp1, err := ToggleDislike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("toggle dislike error: %v", err)
	}
	if !resp1.Disliked {
		t.Error("expected disliked=true after first toggle")
	}
	if resp1.DislikeCount != 1 {
		t.Errorf("expected dislike_count 1, got %d", resp1.DislikeCount)
	}

	resp2, err := ToggleDislike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("toggle undislike error: %v", err)
	}
	if resp2.Disliked {
		t.Error("expected disliked=false after second toggle")
	}
	if resp2.DislikeCount != 0 {
		t.Errorf("expected dislike_count 0, got %d", resp2.DislikeCount)
	}
}

func TestToggleLike_RemovesDislike(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")
	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title: "MutualEx Comic", CoverKey: "covers/mutual.jpg", FileKey: "files/mutual.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	uCtx := userCtx

	dResp, err := ToggleDislike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("dislike error: %v", err)
	}
	if dResp.DislikeCount != 1 {
		t.Errorf("expected dislike_count 1, got %d", dResp.DislikeCount)
	}

	lResp, err := ToggleLike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("like error: %v", err)
	}
	if !lResp.Liked {
		t.Error("expected liked=true")
	}
	if lResp.LikeCount != 1 {
		t.Errorf("expected like_count 1, got %d", lResp.LikeCount)
	}

	status, err := GetLikeStatus(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if status.Disliked {
		t.Error("expected disliked=false after like")
	}
	if !status.Liked {
		t.Error("expected liked=true")
	}
}

func TestToggleDislike_RemovesLike(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")
	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title: "MutualEx Comic2", CoverKey: "covers/mutual2.jpg", FileKey: "files/mutual2.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	uCtx := userCtx

	ToggleLike(uCtx, comic.ID)

	dResp, err := ToggleDislike(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("dislike error: %v", err)
	}
	if !dResp.Disliked {
		t.Error("expected disliked=true")
	}
	if dResp.DislikeCount != 1 {
		t.Errorf("expected dislike_count 1, got %d", dResp.DislikeCount)
	}

	status, err := GetLikeStatus(uCtx, comic.ID)
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if status.Liked {
		t.Error("expected liked=false after dislike")
	}
	if !status.Disliked {
		t.Error("expected disliked=true")
	}
}

func TestApproveComic_ChangesStatusToPublished(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
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

	modCtx := moderatorCtx
	if err := ApproveComic(modCtx, comic.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	pubCtx := context.Background()
	fetched, err := GetComic(pubCtx, comic.Slug)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if fetched.Status != "published" {
		t.Errorf("expected published, got %s", fetched.Status)
	}
}

func TestApproveComic_RequiresModeratorOrAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "NoPerm Comic",
		CoverKey: "covers/noperm.jpg",
		FileKey:  "files/noperm.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx
	err = ApproveComic(uCtx, comic.ID)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}

func TestRejectComic_ChangesStatusToRejected(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Reject Comic",
		CoverKey: "covers/reject.jpg",
		FileKey:  "files/reject.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx
	if err := RejectComic(modCtx, comic.ID, &RejectParams{Reason: "low quality"}); err != nil {
		t.Fatalf("reject error: %v", err)
	}

	fetched, err := GetComic(uploaderCtx, comic.Slug)
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

	upCtx := uploaderCtx
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

	modCtx := moderatorCtx
	_ = ApproveComic(modCtx, "550e8400-e29b-41d4-a716-446655440000")

	admCtx := adminCtx
	resp, err := AdminListComics(admCtx, &AdminListComicsParams{})
	if err != nil {
		t.Fatalf("admin list error: %v", err)
	}
	if resp.Total < 2 {
		t.Errorf("expected at least 2 comics in admin list, got %d", resp.Total)
	}
}

func TestAdminListComics_RequiresAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	uCtx := userCtx
	_, err := AdminListComics(uCtx, &AdminListComicsParams{})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}

func TestGetLikeStatus(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Status Comic",
		CoverKey: "covers/status.jpg",
		FileKey:  "files/status.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx
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

func TestArchiveComic(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Archive Comic",
		CoverKey: "covers/archive.jpg",
		FileKey:  "files/archive.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	admCtx := adminCtx
	if err := ArchiveComic(admCtx, comic.ID); err != nil {
		t.Fatalf("archive error: %v", err)
	}

	resp, err := RecycleBin(admCtx, &RecycleBinParams{})
	if err != nil {
		t.Fatalf("recycle bin error: %v", err)
	}
	found := false
	for _, c := range resp.Comics {
		if c.ID == comic.ID {
			found = true; break
		}
	}
	if !found {
		t.Error("archived comic should appear in recycle bin")
	}
}

func TestArchiveComic_RequiresAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Archive NoPerm",
		CoverKey: "covers/ar-noperm.jpg",
		FileKey:  "files/ar-noperm.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx
	err = ArchiveComic(uCtx, comic.ID)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}

func TestRestoreComic(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Restore Comic",
		CoverKey: "covers/restore.jpg",
		FileKey:  "files/restore.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	admCtx := adminCtx
	if err := ArchiveComic(admCtx, comic.ID); err != nil {
		t.Fatalf("archive error: %v", err)
	}
	if err := RestoreComic(admCtx, comic.ID); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	resp, err := RecycleBin(admCtx, &RecycleBinParams{})
	if err != nil {
		t.Fatalf("recycle bin error: %v", err)
	}
	for _, c := range resp.Comics {
		if c.ID == comic.ID {
			t.Error("restored comic should NOT appear in recycle bin")
		}
	}
}

func TestRecycleBin_ListsDeletedComics(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Bin Comic",
		CoverKey: "covers/bin.jpg",
		FileKey:  "files/bin.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	admCtx := adminCtx
	if err := ArchiveComic(admCtx, comic.ID); err != nil {
		t.Fatalf("archive error: %v", err)
	}

	resp, err := RecycleBin(admCtx, &RecycleBinParams{})
	if err != nil {
		t.Fatalf("recycle bin error: %v", err)
	}
	if resp.Total < 1 {
		t.Error("expected at least 1 comic in recycle bin")
	}
}

func TestRecycleBin_RequiresModeratorOrAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	uCtx := userCtx
	_, err := RecycleBin(uCtx, &RecycleBinParams{})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}

func TestListComicsRandomSort(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	pubCtx := context.Background()
	resp, err := ListComics(pubCtx, &ListComicsParams{Sort: "random", Limit: 1})
	if err != nil {
		t.Fatalf("list random error: %v", err)
	}
	if len(resp.Comics) > 1 {
		t.Errorf("expected at most 1 comic, got %d", len(resp.Comics))
	}
}

func TestDeleteComic_Valid(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Delete Me",
		CoverKey: "covers/delete.jpg",
		FileKey:  "files/delete.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	admCtx := adminCtx
	if err := DeleteComic(admCtx, comic.ID); err != nil {
		t.Fatalf("delete error: %v", err)
	}

	_, err = GetComic(context.Background(), comic.ID)
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.NotFound {
		t.Errorf("expected NotFound, got %v", e.Code)
	}
}

func TestDeleteComic_RequiresAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Delete NoPerm",
		CoverKey: "covers/del-noperm.jpg",
		FileKey:  "files/del-noperm.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	uCtx := userCtx
	err = DeleteComic(uCtx, comic.ID)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestBatchGetComics(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	c1, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Batch Comic 1",
		CoverKey: "covers/batch1.jpg",
		FileKey:  "files/batch1.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	c2, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Batch Comic 2",
		CoverKey: "covers/batch2.jpg",
		FileKey:  "files/batch2.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	c3, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Batch Comic 3",
		CoverKey: "covers/batch3.jpg",
		FileKey:  "files/batch3.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx
	if err := ApproveComic(modCtx, c1.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}
	if err := ApproveComic(modCtx, c2.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}
	if err := ApproveComic(modCtx, c3.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	pubCtx := context.Background()
	resp, err := BatchGetComics(pubCtx, &BatchComicsParams{IDs: []string{c1.ID, c2.ID, c3.ID}})
	if err != nil {
		t.Fatalf("batch get error: %v", err)
	}
	if len(resp.Comics) != 3 {
		t.Errorf("expected 3 comics, got %d", len(resp.Comics))
	}

	titles := map[string]bool{c1.Title: true, c2.Title: true, c3.Title: true}
	for _, c := range resp.Comics {
		if !titles[c.Title] {
			t.Errorf("unexpected comic title: %s", c.Title)
		}
	}
}

func TestBatchGetComics_EmptyIds(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	pubCtx := context.Background()
	resp, err := BatchGetComics(pubCtx, &BatchComicsParams{IDs: []string{}})
	if err != nil {
		t.Fatalf("batch get error: %v", err)
	}
	if len(resp.Comics) != 0 {
		t.Errorf("expected empty slice, got %d comics", len(resp.Comics))
	}
}

func TestBatchGetComics_LimitExceeded(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	ids := make([]string, 51)
	for i := range ids {
		ids[i] = "550e8400-e29b-41d4-a716-446655440099"
	}

	pubCtx := context.Background()
	_, err := BatchGetComics(pubCtx, &BatchComicsParams{IDs: ids})
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

func TestAdminAuditLogs(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Audit Comic",
		CoverKey: "covers/audit.jpg",
		FileKey:  "files/audit.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	modCtx := moderatorCtx
	if err := ApproveComic(modCtx, comic.ID); err != nil {
		t.Fatalf("approve error: %v", err)
	}

	admCtx := adminCtx
	resp, err := AdminAuditLogs(admCtx)
	if err != nil {
		t.Fatalf("audit logs error: %v", err)
	}
	if len(resp.Entries) < 1 {
		t.Fatal("expected at least 1 audit log entry")
	}

	found := false
	for _, entry := range resp.Entries {
		if entry.Action == "approve_comic" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected approve_comic audit log entry")
	}
}

func TestAdminAuditLogs_RequiresAdmin(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	uCtx := userCtx
	_, err := AdminAuditLogs(uCtx)
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestFlagComment_AndList(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Flag Comic",
		CoverKey: "covers/flag.jpg",
		FileKey:  "files/flag.cbz",
	})
	if err != nil {
		t.Fatalf("create comic error: %v", err)
	}

	comment, err := CreateComment(userCtx, comic.ID, &CreateCommentParams{BodyText: "spam comment"})
	if err != nil {
		t.Fatalf("create comment error: %v", err)
	}

	// A different user flags it.
	flagCtx := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-446655440010", "user", "free")
	if err := FlagComment(flagCtx, comment.ID, &FlagCommentParams{Reason: "spam"}); err != nil {
		t.Fatalf("flag error: %v", err)
	}

	// Idempotent: same user flags again → no error, no duplicate.
	if err := FlagComment(flagCtx, comment.ID, &FlagCommentParams{Reason: "spam"}); err != nil {
		t.Fatalf("duplicate flag error: %v", err)
	}

	// Moderator lists flags.
	modCtx := moderatorCtx
	resp, err := ListFlaggedComments(modCtx, &ListFlaggedCommentsParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list flags error: %v", err)
	}
	if resp.Total < 1 {
		t.Fatal("expected at least 1 flagged comment")
	}
	var found bool
	for _, f := range resp.Flags {
		if f.CommentID == comment.ID && f.FlagCount >= 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected the flagged comment in list")
	}
}

func TestFlagComment_NotFound(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	err := FlagComment(userCtx, "550e8400-e29b-41d4-a716-446655440099", &FlagCommentParams{})
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestListFlaggedComments_RequiresModerator(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	_, err := ListFlaggedComments(userCtx, &ListFlaggedCommentsParams{})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errs.Error, got %T", err)
	}
	if e.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", e.Code)
	}
}

func TestResolveFlag_MarksResolved(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	comic, err := CreateComic(upCtx, &CreateComicParams{
		Title:    "Resolve Comic",
		CoverKey: "covers/resolve.jpg",
		FileKey:  "files/resolve.cbz",
	})
	if err != nil {
		t.Fatalf("create comic error: %v", err)
	}

	comment, err := CreateComment(userCtx, comic.ID, &CreateCommentParams{BodyText: "flagged once"})
	if err != nil {
		t.Fatalf("create comment error: %v", err)
	}

	flagCtx := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-446655440011", "user", "free")
	if err := FlagComment(flagCtx, comment.ID, &FlagCommentParams{}); err != nil {
		t.Fatalf("flag error: %v", err)
	}

	modCtx := moderatorCtx
	before, _ := ListFlaggedComments(modCtx, &ListFlaggedCommentsParams{Page: 1, Limit: 20})
	if before.Total < 1 {
		t.Fatal("expected open flag before resolve")
	}

	if err := ResolveFlag(modCtx, before.Flags[0].FlagID); err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	after, _ := ListFlaggedComments(modCtx, &ListFlaggedCommentsParams{Page: 1, Limit: 20})
	if after.Total >= before.Total {
		t.Errorf("expected fewer open flags after resolve, got %d", after.Total)
	}
}

func TestFollowUploader_AndStatus(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	uploaderID := "550e8400-e29b-41d4-a716-446655440001" // fixtures.TestUploaderID
	followerCtx := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-446655440099", "user", "free")

	if err := FollowUploader(followerCtx, uploaderID); err != nil {
		t.Fatalf("follow error: %v", err)
	}

	status, err := GetUploaderFollowStatus(followerCtx, uploaderID)
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if !status.Following {
		t.Error("expected following=true after follow")
	}

	if err := UnfollowUploader(followerCtx, uploaderID); err != nil {
		t.Fatalf("unfollow error: %v", err)
	}
	status, _ = GetUploaderFollowStatus(followerCtx, uploaderID)
	if status.Following {
		t.Error("expected following=false after unfollow")
	}
}

func TestFollowUploader_CannotFollowSelf(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	err := FollowUploader(upCtx, fixtures.TestUploaderID)
	if err == nil {
		t.Fatal("expected error for self-follow, got nil")
	}
}

func TestLanguageFacets_CountsPublishedByLanguage(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	upCtx := uploaderCtx
	// Create + publish two comics in different languages.
	mk := func(title, lang string) string {
		c, err := CreateComic(upCtx, &CreateComicParams{
			Title:           title,
			CoverKey:        "covers/" + title + ".jpg",
			FileKey:         "files/" + title + ".cbz",
			ContentLanguage: lang,
		})
		if err != nil {
			t.Fatalf("create error: %v", err)
		}
		if err := ApproveComic(moderatorCtx, c.ID); err != nil {
			t.Fatalf("approve error: %v", err)
		}
		return c.ID
	}

	mk("English One", "en")
	mk("Japanese One", "ja")
	mk("Japanese Two", "ja")

	resp, err := LanguageFacets(context.Background())
	if err != nil {
		t.Fatalf("facets error: %v", err)
	}

	counts := map[string]int{}
	for _, f := range resp.Facets {
		counts[f.Language] = f.Count
	}
	if counts["en"] < 1 {
		t.Errorf("expected en count >= 1, got %d", counts["en"])
	}
	if counts["ja"] < 2 {
		t.Errorf("expected ja count >= 2, got %d", counts["ja"])
	}
}
