package convert

import "database/sql"

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
