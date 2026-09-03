package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_14"
)

// Game is the plain-Go-typed form of a single row in the games table
type Game struct {
	GameID                int64
	Hometeam              *int64 // nil for an unresolved bracket slot
	Visitorteam           *int64
	Homescore             *int64
	Visitorscore          *int64
	Reservation           *int64
	Time                  string
	Pool                  *int64
	Valid                 convert.IntBool
	Halftime              *int64
	Official              string
	Respteam              *int64
	Resppers              *int64
	Homesotg              *int64
	Visitorsotg           *int64
	Isongoing             convert.IntBool
	SchedulingNameHome    *int64
	SchedulingNameVisitor *int64
	Name                  *int64 // scheduling-name id; sent by the API as a numeric string
	Timeslot              *int64
	Homedefenses          *int64
	Visitordefenses       *int64
	Islive                convert.IntBool
	Liveurl               string
}

func (s *Store) InsertGame(ctx context.Context, g Game) error {
	return s.q.InsertGame(ctx, dbgen.InsertGameParams{
		GameID:                g.GameID,
		Hometeam:              convert.NullInt64(g.Hometeam),
		Visitorteam:           convert.NullInt64(g.Visitorteam),
		Homescore:             convert.NullInt64(g.Homescore),
		Visitorscore:          convert.NullInt64(g.Visitorscore),
		Reservation:           convert.NullInt64(g.Reservation),
		Time:                  convert.NullString(g.Time),
		Pool:                  convert.NullInt64(g.Pool),
		Valid:                 g.Valid.Int64(),
		Halftime:              convert.NullInt64(g.Halftime),
		Official:              convert.NullString(g.Official),
		Respteam:              convert.NullInt64(g.Respteam),
		Resppers:              convert.NullInt64(g.Resppers),
		Homesotg:              convert.NullInt64(g.Homesotg),
		Visitorsotg:           convert.NullInt64(g.Visitorsotg),
		Isongoing:             g.Isongoing.NullInt64(),
		SchedulingNameHome:    convert.NullInt64(g.SchedulingNameHome),
		SchedulingNameVisitor: convert.NullInt64(g.SchedulingNameVisitor),
		Name:                  convert.NullInt64(g.Name),
		Timeslot:              convert.NullInt64(g.Timeslot),
		Homedefenses:          convert.NullInt64(g.Homedefenses),
		Visitordefenses:       convert.NullInt64(g.Visitordefenses),
		Islive:                g.Islive.NullInt64(),
		Liveurl:               convert.NullString(g.Liveurl),
	})
}

func (s *Store) ListGames(ctx context.Context) ([]Game, error) {
	rows, err := s.q.ListGames(ctx)
	if err != nil {
		return nil, err
	}
	games := make([]Game, len(rows))
	for i, row := range rows {
		games[i] = Game{
			GameID:                row.GameID,
			Hometeam:              convert.Int64(row.Hometeam),
			Visitorteam:           convert.Int64(row.Visitorteam),
			Homescore:             convert.Int64(row.Homescore),
			Visitorscore:          convert.Int64(row.Visitorscore),
			Reservation:           convert.Int64(row.Reservation),
			Time:                  convert.String(row.Time),
			Pool:                  convert.Int64(row.Pool),
			Valid:                 convert.IntBoolFromInt64(row.Valid),
			Halftime:              convert.Int64(row.Halftime),
			Official:              convert.String(row.Official),
			Respteam:              convert.Int64(row.Respteam),
			Resppers:              convert.Int64(row.Resppers),
			Homesotg:              convert.Int64(row.Homesotg),
			Visitorsotg:           convert.Int64(row.Visitorsotg),
			Isongoing:             convert.IntBoolFromInt64(row.Isongoing.Int64),
			SchedulingNameHome:    convert.Int64(row.SchedulingNameHome),
			SchedulingNameVisitor: convert.Int64(row.SchedulingNameVisitor),
			Name:                  convert.Int64(row.Name),
			Timeslot:              convert.Int64(row.Timeslot),
			Homedefenses:          convert.Int64(row.Homedefenses),
			Visitordefenses:       convert.Int64(row.Visitordefenses),
			Islive:                convert.IntBoolFromInt64(row.Islive.Int64),
			Liveurl:               convert.String(row.Liveurl),
		}
	}
	return games, nil
}
