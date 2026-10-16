// Package store is the direct interface to the uo_01-only tables of an UltiOrganizer
// tournament's SQLite archive, plus the two shared tables (locations, reservations)
// UltiOrganizer exposes more of than the Live! API does
// every other shared table is written through store/v01_09_14 unchanged
// Wraps internal/db/uo_01 generated queries and translates from plain Go types
package store

import (
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// Store is the uo_01 side of the archive for one tournament, backed by a single SQLite file.
type Store struct {
	db dbgen.DBTX
	q  *dbgen.Queries
}

// New wraps an already-open, schema-applied database connection.
func New(db dbgen.DBTX) *Store {
	return &Store{db: db, q: dbgen.New(db)}
}
