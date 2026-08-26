package comics

import (
	"context"
	"errors"
	"testing"

	myauth "comics-galore/backend/auth"
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

// enableComments flips the global enable_comments setting on so the comment
// tests can run (commenting is disabled by default until moderation ships).
func enableComments(t *testing.T) {
	t.Helper()
	settings, err := myauth.GetAdminSettings(adminCtx)
	if err != nil {
		t.Fatalf("get settings error: %v", err)
	}
	settings.EnableComments = true
	if _, err := myauth.SaveAdminSettings(adminCtx, settings); err != nil {
		t.Fatalf("save settings error: %v", err)
	}
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

func TestCreateComic_MissingTitle(t *testing.T) {	ctx := uploaderCtx

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
	enableComments(t)

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
	enableComments(t)

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

func TestStaffPicks_AndSavedViews(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:    "Staff Pick Comic",
		CoverKey: "covers/sp.jpg",
		FileKey:  "files/sp.cbz",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, comic.ID)

	if err := AddStaffPick(adminCtx, &AddStaffPickParams{ComicID: comic.ID}); err != nil {
		t.Fatalf("add staff pick error: %v", err)
	}

	picks, err := ListStaffPicks(context.Background())
	if err != nil {
		t.Fatalf("list staff picks error: %v", err)
	}
	found := false
	for _, p := range picks.Picks {
		if p.ComicID == comic.ID {
			found = true
		}
	}
	if !found {
		t.Error("expected staff pick to be listed")
	}

	// Saved views
	sv, err := SaveView(adminCtx, &SaveViewParams{Resource: "users", Name: "My view", Filters: `{"tier":"gold"}`})
	if err != nil {
		t.Fatalf("save view error: %v", err)
	}
	views, err := ListSavedViews(adminCtx)
	if err != nil {
		t.Fatalf("list views error: %v", err)
	}
	if len(views.Views) < 1 {
		t.Error("expected at least 1 saved view")
	}
	_ = sv
}

func TestReadingLists_CRUD(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	userCtx2 := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400f1", "user", "free")
	list, err := CreateReadingList(userCtx2, &CreateReadingListParams{Name: "My Shelf", IsPublic: true})
	if err != nil {
		t.Fatalf("create list error: %v", err)
	}
	if list.ID == "" {
		t.Fatal("expected list id")
	}

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:    "List Comic",
		CoverKey: "covers/list.jpg",
		FileKey:  "files/list.cbz",
	})
	if err != nil {
		t.Fatalf("create comic error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, comic.ID)

	if err := AddToReadingList(userCtx2, list.ID, &AddToListParams{ComicID: comic.ID}); err != nil {
		t.Fatalf("add to list error: %v", err)
	}

	// Public fetch includes the comic.
	pub, err := GetReadingList(context.Background(), list.ID, &GetReadingListParams{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("get public list error: %v", err)
	}
	if len(pub.Comics) != 1 {
		t.Errorf("expected 1 comic in public list, got %d", len(pub.Comics))
	}

	if err := RemoveFromReadingList(userCtx2, list.ID, comic.ID); err != nil {
		t.Fatalf("remove from list error: %v", err)
	}
	pub2, _ := GetReadingList(context.Background(), list.ID, &GetReadingListParams{Page: 1, Limit: 20})
	if len(pub2.Comics) != 0 {
		t.Errorf("expected 0 comics after remove, got %d", len(pub2.Comics))
	}
}

func TestRelatedComics_ReturnsPublished(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	resp, err := RelatedComics(context.Background(), "550e8400-e29b-41d4-a716-446655440099")
	if err != nil {
		t.Fatalf("related error: %v", err)
	}
	if resp.Comics == nil {
		t.Error("expected non-nil comics slice")
	}
}

func TestListFavorites_Paginated(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	user := fixtures.TierGatedCtx("550e8400-e29b-41d4-a716-4466554400f2", "user", "free")

	// Create + approve 2 comics, favorite both.
	for i := 0; i < 2; i++ {
		c, err := CreateComic(uploaderCtx, &CreateComicParams{
			Title:    "Fav Comic",
			CoverKey: "covers/fav.jpg",
			FileKey:  "files/fav.cbz",
		})
		if err != nil {
			t.Fatalf("create error: %v", err)
		}
		_ = ApproveComic(moderatorCtx, c.ID)
		if _, err := ToggleFavorite(user, c.ID); err != nil {
			t.Fatalf("favorite error: %v", err)
		}
	}

	resp, err := ListFavorites(user, &ListFavoritesParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list favorites error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Comics) != 2 {
		t.Errorf("expected 2 comics, got %d", len(resp.Comics))
	}

	// Page 2 should be empty (limit 1 → 2 pages).
	page1, _ := ListFavorites(user, &ListFavoritesParams{Page: 1, Limit: 1})
	page2, _ := ListFavorites(user, &ListFavoritesParams{Page: 2, Limit: 1})
	if len(page1.Comics) != 1 || len(page2.Comics) != 1 {
		t.Errorf("expected 1 per page, got %d and %d", len(page1.Comics), len(page2.Comics))
	}
}

func TestListComics_SearchFilters(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	mk := func(title, author string) string {
		c, err := CreateComic(uploaderCtx, &CreateComicParams{
			Title:    title,
			Author:   author,
			CoverKey: "covers/search.jpg",
			FileKey:  "files/search.cbz",
		})
		if err != nil {
			t.Fatalf("create error: %v", err)
		}
		_ = ApproveComic(moderatorCtx, c.ID)
		return c.ID
	}

	mk("Cosmic Odyssey", "Jane Doe")
	mk("Detective Noir", "John Smith")

	// Search by title
	byTitle, _ := ListComics(context.Background(), &ListComicsParams{Search: "Cosmic", SearchField: "title"})
	if byTitle.Total < 1 {
		t.Errorf("expected >=1 title match, got %d", byTitle.Total)
	}
	// Search by author
	byAuthor, _ := ListComics(context.Background(), &ListComicsParams{Search: "Smith", SearchField: "author"})
	if byAuthor.Total < 1 {
		t.Errorf("expected >=1 author match, got %d", byAuthor.Total)
	}
	// Search all fields
	allFields, _ := ListComics(context.Background(), &ListComicsParams{Search: "Cosmic"})
	if allFields.Total < 1 {
		t.Errorf("expected >=1 all-fields match, got %d", allFields.Total)
	}
}

func TestPopularTags_Counts(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	c, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:    "Tagged Comic",
		CoverKey: "covers/tagged.jpg",
		FileKey:  "files/tagged.cbz",
		Tags:     []string{"sci-fi", "action"},
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, c.ID)

	resp, err := PopularTags(context.Background())
	if err != nil {
		t.Fatalf("popular tags error: %v", err)
	}
	counts := map[string]int{}
	for _, t := range resp.Tags {
		counts[t.Tag] = t.Count
	}
	if counts["sci-fi"] < 1 || counts["action"] < 1 {
		t.Errorf("expected sci-fi and action tags present, got %v", counts)
	}
}

func TestGetComic_MatureLockedForFree(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	// Enable forbid-mature-for-free globally via the auth service.
	admSettings, err := myauth.GetAdminSettings(adminCtx)
	if err != nil {
		t.Fatalf("get settings error: %v", err)
	}
	admSettings.ForbidMatureForFree = true
	if _, err := myauth.SaveAdminSettings(adminCtx, admSettings); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:     "Mature Locked",
		CoverKey:  "covers/mature.jpg",
		FileKey:   "files/mature.cbz",
		AgeRating: "mature",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, comic.ID)

	// Anonymous (free-equivalent) fetch → locked, pages withheld, cover kept.
	fetched, err := GetComic(context.Background(), comic.Slug)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if !fetched.MatureLocked {
		t.Error("expected mature_locked=true for anonymous")
	}
	if fetched.PageCount != 0 {
		t.Errorf("expected page_count=0, got %d", fetched.PageCount)
	}
	if fetched.CoverURL == "" {
		t.Error("expected cover_url retained for blurred teaser")
	}

	// Anonymous pages endpoint refuses to serve the mature comic.
	if _, err := GetComicPages(context.Background(), comic.Slug, &GetComicPagesParams{}); err == nil {
		t.Error("expected pages endpoint to refuse mature comic for anonymous")
	}

	// Staff still sees pages.
	staff, err := GetComic(moderatorCtx, comic.Slug)
	if err != nil {
		t.Fatalf("staff get error: %v", err)
	}
	if staff.MatureLocked {
		t.Error("expected staff not locked")
	}

	// Anonymous list excludes the mature comic.
	list, _ := ListComics(context.Background(), &ListComicsParams{Limit: 50})
	for _, c := range list.Comics {
		if c.ID == comic.ID {
			t.Error("expected mature comic excluded from anonymous list")
		}
	}
}

func TestGetComic_MatureNotLockedWhenPolicyOff(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	// Ensure the policy is off (authdb state may persist across tests).
	admSettings, err := myauth.GetAdminSettings(adminCtx)
	if err != nil {
		t.Fatalf("get settings error: %v", err)
	}
	admSettings.ForbidMatureForFree = false
	if _, err := myauth.SaveAdminSettings(adminCtx, admSettings); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:     "Mature Allowed",
		CoverKey:  "covers/allowed.jpg",
		FileKey:   "files/allowed.cbz",
		AgeRating: "mature",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, comic.ID)

	fetched, err := GetComic(context.Background(), comic.Slug)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if fetched.MatureLocked {
		t.Error("expected mature_locked=false when policy off")
	}
}

func TestCreateComic_ReaderFieldsRoundTrip(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:            "Reader Fields",
		CoverKey:         "covers/reader.jpg",
		FileKey:          "files/reader.cbz",
		PageKeys:         []string{"pages/1.jpg", "pages/2.jpg", "pages/3.jpg"},
		PageDimensions:   []PageDimension{{Width: 800, Height: 1200}, {Width: 800, Height: 1200}, {Width: 600, Height: 900}},
		ReadingDirection: "rtl",
		ArchiveMimetype:  "application/vnd.comicbook+zip",
		Isbn:             "978-3-16-148410-0",
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	if comic.PageCount != 3 {
		t.Errorf("expected page_count 3, got %d", comic.PageCount)
	}
	if comic.ReadingDirection != "rtl" {
		t.Errorf("expected reading_direction rtl, got %s", comic.ReadingDirection)
	}
	if comic.ArchiveMimetype != "application/vnd.comicbook+zip" {
		t.Errorf("expected archive_mimetype, got %q", comic.ArchiveMimetype)
	}
	if comic.Isbn != "978-3-16-148410-0" {
		t.Errorf("expected isbn, got %q", comic.Isbn)
	}
	if comic.Upc != "" || comic.Issn != "" {
		t.Errorf("expected empty upc/issn, got upc=%q issn=%q", comic.Upc, comic.Issn)
	}
}

func TestGetComicPages_PaginationAndDimensions(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:          "Paged",
		CoverKey:       "covers/paged.jpg",
		FileKey:        "files/paged.cbz",
		PageKeys:       []string{"p/1.jpg", "p/2.jpg", "p/3.jpg", "p/4.jpg", "p/5.jpg"},
		PageDimensions: []PageDimension{{Width: 100, Height: 200}, {Width: 101, Height: 201}, {Width: 102, Height: 202}, {Width: 103, Height: 203}, {Width: 104, Height: 204}},
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, comic.ID)

	// Reader (non-preview) access requires a paid tier; use a gold caller.
	gold := fixtures.TierGatedCtx("paged-gold", "user", "gold")

	res, err := GetComicPages(gold, comic.Slug, &GetComicPagesParams{Offset: 0, Limit: 2})
	if err != nil {
		t.Fatalf("pages error: %v", err)
	}
	if res.Total != 5 {
		t.Errorf("expected total 5, got %d", res.Total)
	}
	if len(res.Pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(res.Pages))
	}
	if res.Pages[0].Index != 0 || res.Pages[1].Index != 1 {
		t.Errorf("expected indices 0,1 got %d,%d", res.Pages[0].Index, res.Pages[1].Index)
	}
	if res.Pages[0].Width != 100 || res.Pages[0].Height != 200 {
		t.Errorf("expected page 0 dims 100x200, got %dx%d", res.Pages[0].Width, res.Pages[0].Height)
	}
	if res.Pages[0].URL == "" {
		t.Error("expected resolved page url")
	}

	// Second chunk.
	res2, err := GetComicPages(gold, comic.Slug, &GetComicPagesParams{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("pages error: %v", err)
	}
	if len(res2.Pages) != 2 || res2.Pages[0].Index != 2 {
		t.Errorf("expected page 2 first in chunk 2, got %v", res2.Pages)
	}

	// Offset beyond end → empty.
	res3, err := GetComicPages(gold, comic.Slug, &GetComicPagesParams{Offset: 10, Limit: 20})
	if err != nil {
		t.Fatalf("pages error: %v", err)
	}
	if len(res3.Pages) != 0 {
		t.Errorf("expected empty pages, got %d", len(res3.Pages))
	}
}

func TestGetComicPages_PreviewGating(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")

	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:    "PreviewGate",
		CoverKey: "covers/pg.jpg",
		FileKey:  "files/pg.cbz",
		PageKeys: []string{"p/1.jpg", "p/2.jpg", "p/3.jpg", "p/4.jpg", "p/5.jpg"},
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	_ = ApproveComic(moderatorCtx, comic.ID)

	// Anonymous preview → indices 0..2 sharp, 3..4 locked (no URL/key).
	res, err := GetComicPages(context.Background(), comic.Slug, &GetComicPagesParams{Offset: 0, Limit: 10, Preview: true})
	if err != nil {
		t.Fatalf("pages error: %v", err)
	}
	if len(res.Pages) != 5 {
		t.Fatalf("expected 5 pages, got %d", len(res.Pages))
	}
	for _, p := range res.Pages {
		if p.Index >= 3 {
			if !p.Locked {
				t.Errorf("page %d should be locked", p.Index)
			}
			if p.URL != "" || p.Key != "" {
				t.Errorf("page %d should not expose url/key", p.Index)
			}
		} else {
			if p.Locked {
				t.Errorf("page %d should not be locked", p.Index)
			}
			if p.URL == "" {
				t.Errorf("page %d should have a url", p.Index)
			}
		}
	}

	// Paid tier preview → everything sharp.
	gold := fixtures.TierGatedCtx("preview-gold", "user", "gold")
	resGold, err := GetComicPages(gold, comic.Slug, &GetComicPagesParams{Offset: 0, Limit: 10, Preview: true})
	if err != nil {
		t.Fatalf("gold pages error: %v", err)
	}
	for _, p := range resGold.Pages {
		if p.Locked || p.URL == "" {
			t.Errorf("gold page %d should be sharp, got locked=%v", p.Index, p.Locked)
		}
	}

	// Reader (no Preview flag) is gated for free/anonymous callers.
	if _, err := GetComicPages(context.Background(), comic.Slug, &GetComicPagesParams{Offset: 0, Limit: 10}); err == nil {
		t.Error("expected reader access denied for anonymous caller")
	}

	// Reader is available to paid tiers.
	resReader, err := GetComicPages(gold, comic.Slug, &GetComicPagesParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("gold reader pages error: %v", err)
	}
	for _, p := range resReader.Pages {
		if p.Locked {
			t.Errorf("gold reader page %d should not be locked", p.Index)
		}
	}
}

func TestListTrendingPopular_Ranking(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "comicsdb")

	// Create two series with different engagement sums.
	mkSeries := func(id, title, genre string) {
		_, err := db.Exec(ctx, `
			INSERT INTO series (id, title, slug, description, genre, category, uploader_id)
			VALUES ($1, $2, $3, '', $4, 'Comic', $5)
		`, id, title, title, genre, fixtures.TestUploaderID)
		if err != nil {
			t.Fatalf("insert series: %v", err)
		}
	}
	mkSeries("30000000-0000-0000-0000-0000000000a1", "Series A", "Action")
	mkSeries("30000000-0000-0000-0000-0000000000a2", "Series B", "Drama")

	// Series A: 1 published comic with high views, low likes.
	// Series B: 1 published comic with low views, high likes.
	insertComic := func(id, slug, seriesID string, views, likes int) {
		_, err := db.Exec(ctx, `
			INSERT INTO comics (id, uploader_id, title, author, slug, description, status,
				cover_key, file_key, series_id, series_order, view_count, like_count)
			VALUES ($1, $2, $3, '', $4, '', 'published', 'c.jpg', 'f.cbz', $5, 1, $6, $7)
		`, id, fixtures.TestUploaderID, slug, slug, seriesID, views, likes)
		if err != nil {
			t.Fatalf("insert comic: %v", err)
		}
	}
	insertComic("20000000-0000-0000-0000-0000000000b1", "a1", "30000000-0000-0000-0000-0000000000a1", 1000, 5)
	insertComic("20000000-0000-0000-0000-0000000000b2", "b1", "30000000-0000-0000-0000-0000000000a2", 100, 50)

	res, err := ListTrendingPopular(ctx)
	if err != nil {
		t.Fatalf("list trending popular: %v", err)
	}

	if len(res.Trending) < 2 {
		t.Fatalf("expected >=2 trending, got %d", len(res.Trending))
	}
	if res.Trending[0].Title != "Series A" {
		t.Errorf("expected Series A top trending (highest views), got %s", res.Trending[0].Title)
	}
	if res.Trending[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", res.Trending[0].Rank)
	}

	if len(res.Popular) < 2 {
		t.Fatalf("expected >=2 popular, got %d", len(res.Popular))
	}
	if res.Popular[0].Title != "Series B" {
		t.Errorf("expected Series B top popular (highest likes), got %s", res.Popular[0].Title)
	}
}

func TestGetHome_ReturnsSections(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "comicsdb")

	const aID = "30000000-0000-0000-0000-0000000000e1"
	const bID = "30000000-0000-0000-0000-0000000000e2"
	const aTitle = "Home Series Alpha"
	const bTitle = "Home Series Beta"

	// Two series with different categories, one scheduled.
	for _, s := range []struct {
		id, title, slug, category, schedule string
	}{
		{aID, aTitle, "home-series-alpha", "Manga", "fri"},
		{bID, bTitle, "home-series-beta", "Comic", ""},
	} {
		_, err := db.Exec(ctx, `
			INSERT INTO series (id, title, slug, description, genre, category, schedule_day, uploader_id)
			VALUES ($1, $2, $3, '', 'Action', $4, $5, $6)
		`, s.id, s.title, s.slug, s.category, nulOrValue(s.schedule), fixtures.TestUploaderID)
		if err != nil {
			t.Fatalf("insert series: %v", err)
		}
	}

	res, err := GetHome(ctx)
	if err != nil {
		t.Fatalf("get home: %v", err)
	}

	// Categories include ours.
	hasCategory := func(c string) bool {
		for _, x := range res.Categories {
			if x == c {
				return true
			}
		}
		return false
	}
	if !hasCategory("Manga") || !hasCategory("Comic") {
		t.Errorf("expected categories to include Manga and Comic, got %v", res.Categories)
	}

	// Daily includes the scheduled series and not the unscheduled one.
	var dailyHasA, dailyHasB bool
	for _, s := range res.DailySeries {
		if s.ID == aID {
			dailyHasA = true
		}
		if s.ID == bID {
			dailyHasB = true
		}
	}
	if !dailyHasA {
		t.Error("expected scheduled series (Alpha) in daily")
	}
	if dailyHasB {
		t.Error("unscheduled series (Beta) should not be in daily")
	}

	// Indie includes our series (by uploader).
	var indieHasA bool
	for _, s := range res.IndieSeries {
		if s.ID == aID {
			indieHasA = true
		}
	}
	if !indieHasA {
		t.Error("expected our series in indie list")
	}
}

func TestSeriesCounterIncrements(t *testing.T) {
	ctx := context.Background()
	_, _ = et.NewTestDatabase(ctx, "comicsdb")

	// Series + comic.
	_, err := db.Exec(ctx, `
		INSERT INTO series (id, title, slug, description, uploader_id)
		VALUES ('30000000-0000-0000-0000-0000000000d1', 'Counter Series', 'counter-series', '', $1)
	`, fixtures.TestUploaderID)
	if err != nil {
		t.Fatalf("insert series: %v", err)
	}
	comic, err := CreateComic(uploaderCtx, &CreateComicParams{
		Title:    "Counter Comic",
		CoverKey: "covers/counter.jpg",
		FileKey:  "files/counter.cbz",
	})
	if err != nil {
		t.Fatalf("create comic: %v", err)
	}
	_, _ = db.Exec(ctx, `UPDATE comics SET series_id = '30000000-0000-0000-0000-0000000000d1' WHERE id = $1`, comic.ID)
	_ = ApproveComic(moderatorCtx, comic.ID)

	// Toggle favorite on → hearts_count + 1.
	userCtx := fixtures.UserCtx()
	if _, err := ToggleFavorite(userCtx, comic.ID); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	var hearts int64
	db.QueryRow(ctx, `SELECT hearts_count FROM series WHERE id = '30000000-0000-0000-0000-0000000000d1'`).Scan(&hearts)
	if hearts != 1 {
		t.Errorf("expected hearts_count 1 after favorite, got %d", hearts)
	}

	// Toggle off → back to 0.
	if _, err := ToggleFavorite(userCtx, comic.ID); err != nil {
		t.Fatalf("unfavorite: %v", err)
	}
	db.QueryRow(ctx, `SELECT hearts_count FROM series WHERE id = '30000000-0000-0000-0000-0000000000d1'`).Scan(&hearts)
	if hearts != 0 {
		t.Errorf("expected hearts_count 0 after unfavorite, got %d", hearts)
	}
}
