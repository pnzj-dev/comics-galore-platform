package billing

import (
	"database/sql"
	"strconv"
	"time"

	"encore.dev/storage/sqldb"
)

func isNoRows(err error) bool {
	return err == sql.ErrNoRows || err == sqldb.ErrNoRows ||
		err.Error() == "no rows in result set"
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
