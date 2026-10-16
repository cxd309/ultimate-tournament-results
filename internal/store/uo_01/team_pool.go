package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// TeamPool is the plain-Go-typed form of a single row in the team_pools table
type TeamPool struct {
	TeamID     int64
	PoolID     int64
	Activerank *int64
}

func (s *Store) InsertTeamPool(ctx context.Context, tp TeamPool) error {
	return s.q.InsertTeamPool(ctx, dbgen.InsertTeamPoolParams{
		TeamID:     tp.TeamID,
		PoolID:     tp.PoolID,
		Activerank: convert.NullInt64(tp.Activerank),
	})
}
