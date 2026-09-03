package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
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
	Timestamp    string // scorekeeper data-entry time, not the goal's -- see schema
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
		Timestamp:    convert.NullString(g.Timestamp),
	})
}

func (s *Store) ListGoals(ctx context.Context) ([]Goal, error) {
	rows, err := s.q.ListGoals(ctx)
	if err != nil {
		return nil, err
	}
	goals := make([]Goal, len(rows))
	for i, row := range rows {
		goals[i] = Goal{
			GameID:       row.GameID,
			Num:          row.Num,
			Assist:       convert.Int64(row.Assist),
			Scorer:       convert.Int64(row.Scorer),
			Time:         convert.Int64(row.Time),
			Homescore:    convert.Int64(row.Homescore),
			Visitorscore: convert.Int64(row.Visitorscore),
			Ishomegoal:   convert.IntBoolFromInt64(row.Ishomegoal),
			Iscallahan:   convert.IntBoolFromInt64(row.Iscallahan),
			Timestamp:    convert.String(row.Timestamp),
		}
	}
	return goals, nil
}
