// Package store is the direct interface to a 1.9.14 tournament's SQLite archive. It wraps
// internal/db/v1914's generated queries and translates to and from plain Go types, so that
// nothing outside this package needs to know sql.NullString/sql.NullInt64 exist.
package store

import (
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Store is the archive for one tournament, backed by a single SQLite file.
type Store struct {
	q *dbgen.Queries
}

// New wraps an already-open, schema-applied database connection.
func New(db dbgen.DBTX) *Store {
	return &Store{q: dbgen.New(db)}
}
