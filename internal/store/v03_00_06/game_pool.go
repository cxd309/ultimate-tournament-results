package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// GamePool is the plain-Go-typed form of a single row in the game_pools table
// it manages many-many relations between games and pools
// games can now belong to multiple pools
// e.g. for power pools where previous results carry
// Timetable is a flag for if the pool "owns" and schedules a game
type GamePool struct {
	GameID    int64
	PoolID    int64
	Timetable convert.IntBool // true = this pool "owns" this game
}

func (s *Store) InsertGamePool(ctx context.Context, gp GamePool) error {
	return s.q.InsertGamePool(ctx, dbgen.InsertGamePoolParams{
		GameID:    gp.GameID,
		PoolID:    gp.PoolID,
		Timetable: gp.Timetable.Int64(),
	})
}

func (s *Store) ListGamePools(ctx context.Context) ([]GamePool, error) {
	rows, err := s.q.ListGamePools(ctx)
	if err != nil {
		return nil, err
	}
	gamePools := make([]GamePool, len(rows))
	for i, row := range rows {
		gamePools[i] = GamePool{
			GameID:    row.GameID,
			PoolID:    row.PoolID,
			Timetable: convert.IntBoolFromInt64(row.Timetable),
		}
	}
	return gamePools, nil
}
