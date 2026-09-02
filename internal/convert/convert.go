// Package convert converts between sql.* types and plain Go types
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

// NullFloat64 converts a Go float64 pointer to sql.NullFloat64, nil meaning not set
// See NullInt64 for pointer rationale
func NullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// Float64 converts a sql.NullFloat64 back to a Go float64 pointer; NULL becomes nil
func Float64(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	return &n.Float64
}
