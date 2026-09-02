package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Location is the plain-Go-typed form of a single row in the loctions table
type Location struct {
	ID   int64
	Name string
}

func (s *Store) InsertLocation(ctx context.Context, l Location) error {
	return s.q.InsertLocation(ctx, dbgen.InsertLocationParams{
		ID:   l.ID,
		Name: l.Name,
	})
}
