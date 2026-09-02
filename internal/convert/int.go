package convert

import (
	"database/sql"
	"strconv"
)

// NullInt64 converts a Go int64 pointer to sql.NullInt64, nil meaning not set
// Unlike NullString, 0 is a real, meaningful value for most integer columns
// so this needs to be properly handled by a pointer
func NullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// Int64 converts a sql.NullInt64 back to a Go int64 pointer; NULL becomes nil
func Int64(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

// StringFromOptionalInt64 renders an optional int64 as its decimal string, or "" when absent
func StringFromOptionalInt64(i *int64) string {
	if i == nil {
		return ""
	}
	return strconv.FormatInt(*i, 10)
}
