package app

import (
	"regexp"
)

// ApplicationIDRegex bounds the advisory application identifier to lowercase
// letters, digits, dashes and underscores, 1..66 chars — matching the DB
// column width (VARCHAR(66), see
// nitronode/config/migrations/postgres/20260420000000_add_application_id_to_writes.sql).
var ApplicationIDRegex = regexp.MustCompile(`^[a-z0-9_-]{1,66}$`)

// IsValidApplicationID reports whether id is a well-formed advisory
// application identifier (see ApplicationIDRegex).
func IsValidApplicationID(id string) bool {
	return ApplicationIDRegex.MatchString(id)
}
