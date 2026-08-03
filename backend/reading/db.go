package reading

import (
	"encoding/hex"
	"errors"

	"encore.dev/storage/sqldb"
)

func isNoRows(err error) bool {
	return errors.Is(err, sqldb.ErrNoRows)
}

func _randomSuffix(n int) string {
	b := make([]byte, n)
	// crypto/rand not needed for this helper
	return hex.EncodeToString(b)[:n]
}
