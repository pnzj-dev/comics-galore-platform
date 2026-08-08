package billing

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"encore.dev/storage/sqldb"
)

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqldb.ErrNoRows)
}

func timePtr(t *time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

func fmtNumS(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}
