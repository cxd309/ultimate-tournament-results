package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// Played is the plain-Go-typed form of a single row in the played table
type Played struct {
	GameID   int64
	PlayerID int64
	Num      *int64
	Captain  convert.IntBool
}

func (s *Store) InsertPlayed(ctx context.Context, p Played) error {
	return s.q.InsertPlayed(ctx, dbgen.InsertPlayedParams{
		GameID:   p.GameID,
		PlayerID: p.PlayerID,
		Num:      convert.NullInt64(p.Num),
		Captain:  p.Captain.Int64(),
	})
}
