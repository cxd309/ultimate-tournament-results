package convert

import "database/sql"

// NullString converts a Go string to sql.NullString
// An empty string is treated as "not set" (NULL)
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// String converts a sql.NullString back to a plain Go string; NULL becomes ""
func String(n sql.NullString) string {
	return n.String
}
