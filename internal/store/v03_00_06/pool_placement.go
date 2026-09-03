package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// PoolPlacement is the plain-Go-typed form of a single row in the pool_placements table
type PoolPlacement struct {
	PoolID    int64
	TeamID    int64
	Placement *int64 // nil while the pool has no resolved rank for that team yet
}

func (s *Store) InsertPoolPlacement(ctx context.Context, p PoolPlacement) error {
	return s.q.InsertPoolPlacement(ctx, dbgen.InsertPoolPlacementParams{
		PoolID:    p.PoolID,
		TeamID:    p.TeamID,
		Placement: convert.NullInt64(p.Placement),
	})
}

func (s *Store) ListPoolPlacements(ctx context.Context) ([]PoolPlacement, error) {
	rows, err := s.q.ListPoolPlacements(ctx)
	if err != nil {
		return nil, err
	}
	placements := make([]PoolPlacement, len(rows))
	for i, row := range rows {
		placements[i] = PoolPlacement{
			PoolID:    row.PoolID,
			TeamID:    row.TeamID,
			Placement: convert.Int64(row.Placement),
		}
	}
	return placements, nil
}
