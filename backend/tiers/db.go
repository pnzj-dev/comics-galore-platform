package tiers

import (
	"errors"

	"encore.dev/storage/sqldb"
)

func isNoRows(err error) bool {
	return errors.Is(err, sqldb.ErrNoRows)
}
