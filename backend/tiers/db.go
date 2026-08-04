package tiers

import (
	"encoding/json"
	"errors"

	"encore.dev/storage/sqldb"
)

func isNoRows(err error) bool {
	return errors.Is(err, sqldb.ErrNoRows)
}

func parseFeatures(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var features []string
	if err := json.Unmarshal(raw, &features); err != nil {
		return []string{}
	}
	return features
}
