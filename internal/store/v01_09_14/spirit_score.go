package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_14"
)

// SpiritScore is the plain-Go-typed form of a single row in the spirit_scores table.
// TeamID is who the score is for, not who gave it -- see the schema's own comment.
type SpiritScore struct {
	GameID   int64
	TeamID   int64
	Cat1     int64
	Cat2     int64
	Cat3     int64
	Cat4     int64
	Cat5     int64
	Comments string
}

func (s *Store) InsertSpiritScore(ctx context.Context, sc SpiritScore) error {
	return s.q.InsertSpiritScore(ctx, dbgen.InsertSpiritScoreParams{
		GameID:   sc.GameID,
		TeamID:   sc.TeamID,
		Cat1:     sc.Cat1,
		Cat2:     sc.Cat2,
		Cat3:     sc.Cat3,
		Cat4:     sc.Cat4,
		Cat5:     sc.Cat5,
		Comments: convert.NullString(sc.Comments),
	})
}
