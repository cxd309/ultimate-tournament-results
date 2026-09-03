package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_17"
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

func (s *Store) ListDivisions(ctx context.Context) ([]Division, error) {
	rows, err := s.q.ListDivisions(ctx)
	if err != nil {
		return nil, err
	}
	divisions := make([]Division, len(rows))
	for i, row := range rows {
		divisions[i] = Division{
			SeriesID: row.SeriesID,
			Name:     convert.String(row.Name),
			Ordering: convert.String(row.Ordering),
		}
	}
	return divisions, nil
}
