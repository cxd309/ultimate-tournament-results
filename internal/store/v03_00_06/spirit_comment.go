package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// SpiritComment is the plain-Go-typed form of a single row in the spirit_comments table
// TeamID is who the comment is for, not who gave it -- same convention as spirit_scores
type SpiritComment struct {
	GameID  int64
	TeamID  int64
	Comment string
}

func (s *Store) InsertSpiritComment(ctx context.Context, c SpiritComment) error {
	return s.q.InsertSpiritComment(ctx, dbgen.InsertSpiritCommentParams{
		GameID:  c.GameID,
		TeamID:  c.TeamID,
		Comment: convert.NullString(c.Comment),
	})
}