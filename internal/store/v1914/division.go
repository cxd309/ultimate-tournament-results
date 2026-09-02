package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Division is the plain-Go-typed form of a single row in the divisions table
type Division struct {
	SeriesID int64
	Name     string
	Ordering string
}

func (s *Store) InsertDivision(ctx context.Context, d Division) error {
	return s.q.InsertDivision(ctx, dbgen.InsertDivisionParams{
		SeriesID: d.SeriesID,
		Name:     convert.NullString(d.Name),
		Ordering: convert.NullString(d.Ordering),
	})
}
