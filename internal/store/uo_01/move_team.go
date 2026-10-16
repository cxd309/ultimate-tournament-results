package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// MoveTeam is the plain-Go-typed form of a single row in the move_teams table
type MoveTeam struct {
	FromPool    int64
	FromPlacing int64
	ToPool      int64
}

func (s *Store) InsertMoveTeam(ctx context.Context, m MoveTeam) error {
	return s.q.InsertMoveTeam(ctx, dbgen.InsertMoveTeamParams{
		Frompool:    m.FromPool,
		Fromplacing: m.FromPlacing,
		Topool:      m.ToPool,
	})
}
