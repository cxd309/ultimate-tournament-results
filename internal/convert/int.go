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

// OptionalStringFromInt64 renders an optional int64 as a pointer to its decimal string
// or nil when absent
// unlike StringFromOptionalInt64 this keeps the pointer
// for a field the live API itself sends as an optional JSON string
func OptionalStringFromInt64(i *int64) *string {
	if i == nil {
		return nil
	}
	s := strconv.FormatInt(*i, 10)
	return &s
}

// OptionalInt64FromString parses a numeric-string column back to an optional int64
// treating an empty string as absent
//
// a malformed value degrades to absent rather than returning an error
// this is meant for errors from round-tripped data, not live API input
func OptionalInt64FromString(s string) *int64 {
	if s == "" {
		return nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &i
}

// Int64OrZero dereferences an optional int64, treating nil as 0
func Int64OrZero(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// SentinelIfNil ensures a non-nil int64 pointer
// substitutes a pointer to the sentinel for nil
func SentinelIfNil(i *int64, sentinel int64) *int64 {
	if i != nil {
		return i
	}
	return &sentinel
}

// ZeroIfNil ensures a non-nil int64 pointer, substituting a pointer to 0 for nil
//
// Live! API convention for "not recorded" on a field
// e.g. an unknown goal scorer/assist is a literal 0, not a missing key
func ZeroIfNil(i *int64) *int64 {
	return SentinelIfNil(i, 0)
}

// NilIfNotPositive normalizes Live! API "not recorded" sentinel to nil
//
// goal scorer/assist and reservation.location all use a non-positive value
// (0, or -1 for goals) to mean unset, none of which are ever a real id
func NilIfNotPositive(i *int64) *int64 {
	if i == nil || *i <= 0 {
		return nil
	}
	return i
}
