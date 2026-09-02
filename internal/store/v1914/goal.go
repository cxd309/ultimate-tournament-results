package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Goal is the plain-Go-typed form of a single row in the goals table
// Ishomegoal/Iscallahan are IntBool and required on every goal
type Goal struct {
	GameID       int64
	Num          int64
	Assist       *int64 // nil when unrecorded; the API sends player_id 0, normalized away
	Scorer       *int64
	Time         *int64
	Homescore    *int64
	Visitorscore *int64
	Ishomegoal   convert.IntBool
	Iscallahan   convert.IntBool
}

func (s *Store) InsertGoal(ctx context.Context, g Goal) error {
	return s.q.InsertGoal(ctx, dbgen.InsertGoalParams{
		GameID:       g.GameID,
		Num:          g.Num,
		Assist:       convert.NullInt64(g.Assist),
		Scorer:       convert.NullInt64(g.Scorer),
		Time:         convert.NullInt64(g.Time),
		Homescore:    convert.NullInt64(g.Homescore),
		Visitorscore: convert.NullInt64(g.Visitorscore),
		Ishomegoal:   g.Ishomegoal.Int64(),
		Iscallahan:   g.Iscallahan.Int64(),
	})
}
