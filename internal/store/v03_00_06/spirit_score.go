package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// SpiritScore is the plain-Go-typed form of a single row in the spirit_scores table:
// one category's score for one team in one game
// TeamID is who the score is for, not who gave it; see schema
type SpiritScore struct {
	GameID     int64
	TeamID     int64
	CategoryID int64
	Value      *int64 // 0 is a real submitted score; nil means not yet visible
}

func (s *Store) InsertSpiritScore(ctx context.Context, sc SpiritScore) error {
	return s.q.InsertSpiritScore(ctx, dbgen.InsertSpiritScoreParams{
		GameID:     sc.GameID,
		TeamID:     sc.TeamID,
		CategoryID: sc.CategoryID,
		Value:      convert.NullInt64(sc.Value),
	})
}
