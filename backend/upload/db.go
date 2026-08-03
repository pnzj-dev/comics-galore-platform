package upload

import (
	"encoding/json"
	"errors"

	"encore.dev/storage/sqldb"
)

func isNoRows(err error) bool {
	return errors.Is(err, sqldb.ErrNoRows)
}

func scanParts(src interface{}, parts *[]Part) error {
	if src == nil {
		*parts = []Part{}
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
		*parts = []Part{}
		return nil
	}
	return json.Unmarshal(b, parts)
}

func marshalParts(parts []Part) ([]byte, error) {
	if parts == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(parts)
}
