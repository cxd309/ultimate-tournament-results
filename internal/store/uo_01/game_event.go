package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// GameEvent is the plain-Go-typed form of a single row in the game_events table
type GameEvent struct {
	GameID int64
	Num    int64
	Time   *int64
	Type   string
	Ishome convert.IntBool
}

func (s *Store) InsertGameEvent(ctx context.Context, e GameEvent) error {
	return s.q.InsertGameEvent(ctx, dbgen.InsertGameEventParams{
		GameID: e.GameID,
		Num:    e.Num,
		Time:   convert.NullInt64(e.Time),
		Type:   e.Type,
		Ishome: e.Ishome.Int64(),
	})
}
