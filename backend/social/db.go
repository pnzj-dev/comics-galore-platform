package social

import (
	"errors"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("socialdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

func isNoRows(err error) bool {
	return errors.Is(err, sqldb.ErrNoRows)
}
