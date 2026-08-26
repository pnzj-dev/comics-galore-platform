package comics

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"

	"encore.dev/storage/sqldb"
)

func nextIdx(n int) string {
	return strconv.Itoa(n)
}

func isNoRows(err error) bool {
	return errors.Is(err, sqldb.ErrNoRows)
}

func scanStringSlice(dst *[]string) interface{} {
	return &stringSliceScanner{dst: dst}
}

type stringSliceScanner struct{ dst *[]string }

func (s *stringSliceScanner) Scan(src interface{}) error {
	if src == nil {
		*s.dst = []string{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		*s.dst = []string{}
		return nil
	}
	if len(b) == 0 {
		*s.dst = []string{}
		return nil
	}
	return json.Unmarshal(b, s.dst)
}

func marshalStringSlice(ss []string) ([]byte, error) {
	if ss == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(ss)
}

// marshalPageDimensions serializes the parallel page-dimension array.
func marshalPageDimensions(dims []PageDimension) ([]byte, error) {
	if dims == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(dims)
}

// scanPageDimensions reads the page_dimensions JSONB column into a slice.
func scanPageDimensions(dst *[]PageDimension) interface{} {
	return &pageDimensionsScanner{dst: dst}
}

type pageDimensionsScanner struct{ dst *[]PageDimension }

func (s *pageDimensionsScanner) Scan(src interface{}) error {
	*s.dst = []PageDimension{}
	if src == nil {
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return nil
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, s.dst)
}

func nulString(s *string) *nulStr {
	return &nulStr{s: s}
}

// nulOrValue converts an empty string to NULL for nullable identifier columns.
func nulOrValue(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

type nulStr struct{ s *string }

func (n *nulStr) Scan(src interface{}) error {
	if src == nil {
		*n.s = ""
		return nil
	}
	switch v := src.(type) {
	case []byte:
		*n.s = string(v)
	case string:
		*n.s = v
	}
	return nil
}

func scanComics(rows *sqldb.Rows) ([]Comic, error) {
	var comics []Comic
	for rows.Next() {
		var c Comic
		var pubAt sql.NullTime
		err := rows.Scan(
			&c.ID, &c.UploaderID, &c.Title, &c.Author, &c.Slug, &c.Description,
			&c.ContentLanguage, &c.Status, &c.Category, &c.Genre, &c.CoverKey, &c.FileKey,
			&c.FileSizeBytes, nulString(&c.MinTierID),
			&c.AgeRating, &c.IsPremium, scanStringSlice(&c.Tags), nulString(&c.RejectionReason),
			&pubAt, &c.ViewCount, &c.DownloadCount, &c.LikeCount, &c.FavCount, &c.DislikeCount,
			&c.ReadingDirection, &c.PageCount, nulString(&c.ArchiveMimetype),
			nulString(&c.Isbn), nulString(&c.Upc), nulString(&c.Issn),
			nulString(&c.Volume), nulString(&c.IssueNumber),
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if pubAt.Valid {
			c.PublishedAt = pubAt.Time
		}
		c.CoverURL = resolveCoverURL(c.CoverKey)
		comics = append(comics, c)
	}
	return comics, rows.Err()
}

func IncrementDownloadCount(ctx context.Context, comicID string) error {
	_, err := db.Exec(ctx, `UPDATE comics SET download_count = download_count + 1 WHERE id = $1`, comicID)
	return err
}

func randomSuffix(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}
