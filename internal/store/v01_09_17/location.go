package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_17"
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

func (s *Store) ListLocations(ctx context.Context) ([]Location, error) {
	rows, err := s.q.ListLocations(ctx)
	if err != nil {
		return nil, err
	}
	locations := make([]Location, len(rows))
	for i, row := range rows {
		locations[i] = Location{
			ID:   row.ID,
			Name: row.Name,
		}
	}
	return locations, nil
}
