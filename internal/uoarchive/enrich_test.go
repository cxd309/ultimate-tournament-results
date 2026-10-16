package uoarchive

import (
	"testing"

	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

// testShared is a scrape and the Live! archive it agrees with: game 2 is a forfeit
// UltiOrganizer schedules with no game id
func testShared() (*rows, live) {
	id := func(n int64) *int64 { return &n }
	game := store.Game{GameID: 1, Hometeam: id(7), Visitorteam: id(8), Homescore: id(1), Visitorscore: id(0), Pool: id(10), Reservation: id(100), Time: "2026-06-28 09:00:00"}
	forfeit := store.Game{GameID: 2, Hometeam: id(8), Visitorteam: id(7), Homescore: id(1), Visitorscore: id(0), Pool: id(10), Reservation: id(100), Time: "2026-06-28 11:00:00"}
	goal := store.Goal{GameID: 1, Scorer: id(70), Homescore: id(1), Visitorscore: id(0), Ishomegoal: true}
	r := &rows{
		pools:        []store.Pool{{PoolID: 10}},
		locations:    []uostore.Location{{ID: 1}},
		reservations: []uostore.Reservation{{ID: 100, Location: 1}},
		teams:        []store.Team{{TeamID: 7, Name: "Red"}, {TeamID: 8, Name: "Blue"}},
		players:      []store.Player{{PlayerID: 70, Team: 7}},
		games:        []store.Game{game},
		goals:        []store.Goal{func() store.Goal { g := goal; g.Num = 1; return g }()},
		unlinked:     []string{"Sunday 11:00 Blue - Red"},
	}
	l := live{
		liveArchive: liveArchive{
			pools:        []store.Pool{{PoolID: 10}},
			reservations: []store.Reservation{{ID: 100, Location: 1}},
			teams:        []store.Team{{TeamID: 7, Name: "Red"}, {TeamID: 8, Name: "Blue"}},
			games:        []store.Game{game, forfeit},
		},
		locations: []store.Location{{ID: 1}},
		players:   []store.Player{{PlayerID: 70, Team: 7}, {PlayerID: 71, Team: 8}},
		goals:     []store.Goal{goal},
	}
	return r, l
}

func TestCheckShared(t *testing.T) {
	r, l := testShared()
	if err := checkShared(r, l); err != nil {
		t.Fatal(err)
	}

	cases := map[string]func(r *rows, l *live){
		"team name":             func(r *rows, _ *live) { r.teams[0].Name = "Green" },
		"player team":           func(r *rows, _ *live) { r.players[0].Team = 8 },
		"score":                 func(r *rows, _ *live) { v := int64(2); r.games[0].Homescore = &v },
		"time":                  func(r *rows, _ *live) { r.games[0].Time = "2026-06-28 09:30:00" },
		"goal scorer":           func(r *rows, _ *live) { v := int64(71); r.goals[0].Scorer = &v },
		"reservation location":  func(r *rows, _ *live) { r.reservations[0].Location = 0 },
		"unknown pool":          func(r *rows, _ *live) { r.pools = append(r.pools, store.Pool{PoolID: 11}) },
		"unknown game":          func(r *rows, _ *live) { g := r.games[0]; g.GameID = 3; r.games = append(r.games, g) },
		"forfeit with goals":    func(_ *rows, l *live) { l.goals = append(l.goals, store.Goal{GameID: 2}) },
		"extra unlinked game":   func(_ *rows, l *live) { l.games = l.games[:1] },
		"missing unlinked game": func(r *rows, _ *live) { r.unlinked = nil },
	}
	for name, edit := range cases {
		r, l := testShared()
		edit(r, &l)
		if err := checkShared(r, l); err == nil {
			t.Errorf("%s: checked without an error", name)
		}
	}
}

func TestPageSpirit(t *testing.T) {
	live := []store.SpiritScore{{GameID: 1, TeamID: 7, Cat1: 2}, {GameID: 1, TeamID: 8, Cat1: 3, Comments: "nice"}}
	changes, err := pageSpirit([]store.SpiritScore{
		{GameID: 1, TeamID: 7, Cat1: 2, Comments: "well played"},
		{GameID: 1, TeamID: 8, Cat1: 3, Comments: "nice"},
		{GameID: 2, TeamID: 7, Cat1: 4},
	}, live)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || !changes[0].comment || changes[0].score.Comments != "well played" || !changes[1].added || changes[1].score.GameID != 2 {
		t.Errorf("changes %+v", changes)
	}
	if _, err := pageSpirit([]store.SpiritScore{{GameID: 1, TeamID: 7, Cat1: 3}}, live); err == nil {
		t.Error("category disagreement accepted")
	}
	if _, err := pageSpirit([]store.SpiritScore{{GameID: 1, TeamID: 8, Cat1: 3, Comments: "other"}}, live); err == nil {
		t.Error("comment disagreement accepted")
	}
}
