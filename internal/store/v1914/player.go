package store

import (
	"context"
	"database/sql"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Player is the plain-Go-typed form of one players row (uo_player). PlayerID is used
// directly as the primary key, Team is the external teams.team_id. GamesPlayed is kept
// as reported -- see schema.sql for why. Goals/assists/callahans are deliberately not
// stored here: they're exactly what the goals table already records per player.
type Player struct {
	PlayerID    int64
	FirstName   string
	LastName    string
	Team        int64
	Num         *int64
	GamesPlayed *int64
}

func (s *Store) InsertPlayer(ctx context.Context, p Player) error {
	return s.q.InsertPlayer(ctx, dbgen.InsertPlayerParams{
		PlayerID:    p.PlayerID,
		Firstname:   convert.NullString(p.FirstName),
		Lastname:    convert.NullString(p.LastName),
		Team:        sql.NullInt64{Int64: p.Team, Valid: true},
		Num:         convert.NullInt64(p.Num),
		GamesPlayed: convert.NullInt64(p.GamesPlayed),
	})
}