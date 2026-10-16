package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// GamePool is the plain-Go-typed form of a single row in the game_pools table
type GamePool struct {
	GameID int64
	PoolID int64
}

func (s *Store) InsertGamePool(ctx context.Context, gp GamePool) error {
	return s.q.InsertGamePool(ctx, dbgen.InsertGamePoolParams{
		GameID: gp.GameID,
		PoolID: gp.PoolID,
	})
}
