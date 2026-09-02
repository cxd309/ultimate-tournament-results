package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// Game is the plain-Go-typed form of a single row in the games table
// Valid/Isongoing/Islive/ShowSpirit/Forfeit are IntBool:
// always known even though most are db nullable columns, so no pointer needed
// Hasstarted is a plain int64 pointer, not IntBool: observed values go beyond 0/1
type Game struct {
	GameID                int64
	Hometeam              *int64 // nil for an unresolved bracket slot
	Visitorteam           *int64
	Homescore             *int64
	Visitorscore          *int64
	Reservation           *int64
	Time                  string
	Valid                 convert.IntBool
	Halftime              *int64
	Official              string
	Respteam              *int64
	Resppers              *int64
	Isongoing             convert.IntBool
	SchedulingNameHome    *int64
	SchedulingNameVisitor *int64
	Name                  *int64 // scheduling-name id; sent by the API as a numeric string
	Timeslot              *int64
	Homedefenses          *int64
	Visitordefenses       *int64
	Islive                convert.IntBool
	Liveurl               string
	Hasstarted            *int64
	ShowSpirit            convert.IntBool
	TimerStart            *int64
	TimerPauseStart       *int64
	TimerPausedDuration   *int64
	Forfeit               convert.IntBool
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
		Valid:                 g.Valid.Int64(),
		Halftime:              convert.NullInt64(g.Halftime),
		Official:              convert.NullString(g.Official),
		Respteam:              convert.NullInt64(g.Respteam),
		Resppers:              convert.NullInt64(g.Resppers),
		Isongoing:             g.Isongoing.NullInt64(),
		SchedulingNameHome:    convert.NullInt64(g.SchedulingNameHome),
		SchedulingNameVisitor: convert.NullInt64(g.SchedulingNameVisitor),
		Name:                  convert.NullInt64(g.Name),
		Timeslot:              convert.NullInt64(g.Timeslot),
		Homedefenses:          convert.NullInt64(g.Homedefenses),
		Visitordefenses:       convert.NullInt64(g.Visitordefenses),
		Islive:                g.Islive.NullInt64(),
		Liveurl:               convert.NullString(g.Liveurl),
		Hasstarted:            convert.NullInt64(g.Hasstarted),
		ShowSpirit:            g.ShowSpirit.NullInt64(),
		TimerStart:            convert.NullInt64(g.TimerStart),
		TimerPauseStart:       convert.NullInt64(g.TimerPauseStart),
		TimerPausedDuration:   convert.NullInt64(g.TimerPausedDuration),
		Forfeit:               g.Forfeit.NullInt64(),
	})
}
