// Package store is the direct interface to a 3.0.6 tournament's SQLite archive.
// Wraps internal/db/ generated queries and translates to and from plain Go types
package store

import (
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// Store is the archive for one tournament, backed by a single SQLite file.
type Store struct {
	q *dbgen.Queries
}

// New wraps an already-open, schema-applied database connection.
func New(db dbgen.DBTX) *Store {
	return &Store{q: dbgen.New(db)}
}
