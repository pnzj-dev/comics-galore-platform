package reading

import (
	"context"
	"testing"

	"comics-galore/backend/fixtures"
	comics "comics-galore/backend/comics"

	"encore.dev/beta/errs"
	"encore.dev/et"
)

const testComicA = "550e8400-e29b-41d4-a716-446655440010"
const testComicB = "550e8400-e29b-41d4-a716-446655440011"
const testComicC = "550e8400-e29b-41d4-a716-446655440012"

func TestSaveAndGetProgress(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")
	ctx := fixtures.UserCtx()

	p, err := SaveProgress(ctx, testComicA, &SaveProgressParams{CurrentPage: 5, TotalPages: 32, Completed: false})
	if err != nil {
		t.Fatalf("save error: %v", err)
	}
	if p.CurrentPage != 5 {
		t.Errorf("expected page 5, got %d", p.CurrentPage)
	}

	got, err := GetProgress(ctx, testComicA)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if got.CurrentPage != 5 {
		t.Errorf("expected page 5, got %d", got.CurrentPage)
	}
}

func TestProgressNotFound(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")
	ctx := fixtures.UserCtx()

	_, err := GetProgress(ctx, "550e8400-e29b-41d4-a716-446655440099")
	if err == nil {
		t.Fatal("expected error for nonexistent progress")
	}
}

func TestContinueReading(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")
	ctx := fixtures.UserCtx()

	SaveProgress(ctx, testComicA, &SaveProgressParams{CurrentPage: 3, TotalPages: 20, Completed: false})
	SaveProgress(ctx, testComicB, &SaveProgressParams{CurrentPage: 10, TotalPages: 50, Completed: false})
	SaveProgress(ctx, testComicC, &SaveProgressParams{CurrentPage: 30, TotalPages: 30, Completed: true})

	resp, err := ContinueReading(ctx)
	if err != nil {
		t.Fatalf("continue reading error: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 in-progress items, got %d", len(resp.Items))
	}
}

func TestRecordDownload(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")
	ctx := fixtures.TierGatedCtx(fixtures.TestUserID, "user", "free")

	resp, err := RecordDownload(ctx, testComicA)
	if err != nil {
		t.Fatalf("download error: %v", err)
	}
	if !resp.Allowed {
		t.Fatal("expected download allowed for first download")
	}
	if resp.Limit != 5 {
		t.Errorf("expected free limit 5, got %d", resp.Limit)
	}
}

func TestDownloadQuotaExceeded(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")
	ctx := fixtures.TierGatedCtx(fixtures.TestUserID, "user", "free")

	for i := 0; i < 5; i++ {
		RecordDownload(ctx, testComicA)
	}

	resp, err := RecordDownload(ctx, testComicA)
	if err != nil {
		t.Fatalf("download error: %v", err)
	}
	if resp.Allowed {
		t.Fatal("expected download not allowed after 5 downloads on free tier")
	}
}

func TestRecordDownload_IncrementsCounter(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "comicsdb")
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")

	comic, err := comics.CreateComic(fixtures.UploaderCtx(), &comics.CreateComicParams{
		Title:    "Download Count Comic",
		CoverKey: "covers/dlc.jpg",
		FileKey:  "files/dlc.cbz",
	})
	if err != nil {
		t.Fatalf("create comic error: %v", err)
	}

	dlCtx := fixtures.TierGatedCtx(fixtures.TestUserID, "user", "gold")
	_, err = RecordDownload(dlCtx, comic.ID)
	if err != nil {
		t.Fatalf("download error: %v", err)
	}

	fetched, err := comics.GetComic(context.Background(), comic.ID)
	if err != nil {
		t.Fatalf("get comic error: %v", err)
	}
	if fetched.DownloadCount < 1 {
		t.Errorf("expected download_count >= 1, got %d", fetched.DownloadCount)
	}
}

func TestContinueReading_ExcludesCompleted(t *testing.T) {
	_, _ = et.NewTestDatabase(context.Background(), "readingdb")
	ctx := fixtures.UserCtx()

	SaveProgress(ctx, testComicA, &SaveProgressParams{CurrentPage: 5, TotalPages: 10, Completed: true})
	SaveProgress(ctx, testComicB, &SaveProgressParams{CurrentPage: 3, TotalPages: 20, Completed: false})

	resp, err := ContinueReading(ctx)
	if err != nil {
		t.Fatalf("continue reading error: %v", err)
	}

	for _, item := range resp.Items {
		if item.ComicID == testComicA {
			t.Error("completed comic should not appear in continue reading")
		}
	}

	found := false
	for _, item := range resp.Items {
		if item.ComicID == testComicB {
			found = true
			break
		}
	}
	if !found {
		t.Error("in-progress comic should appear in continue reading")
	}
}

var _ = errs.NotFound
