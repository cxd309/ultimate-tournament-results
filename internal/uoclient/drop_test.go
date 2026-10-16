package uoclient

import "testing"

func TestDropDivisions(t *testing.T) {
	id := func(n int64) *int64 { return &n }
	snap := &Snapshot{
		Menu: Menu{Divisions: []MenuDivision{{SeriesID: 1, Name: "Open"}, {SeriesID: 2, Name: "Guts Open"}}},
		Schedule: []ScheduledGame{
			{GameID: id(1), Division: "Open", PoolID: 10, ReservationID: 100},
			{Division: "Guts Open", PoolID: 20, ReservationID: 100},
			{GameID: id(2), Division: "Guts Open", PoolID: 20, ReservationID: 200},
		},
		Teams:        []TeamListEntry{{TeamID: 5, Division: "Open"}, {TeamID: 6, Division: "Guts Open"}},
		Seeds:        map[int64]int64{5: 1, 6: 1},
		Pools:        map[int64]PoolPage{10: {Heading: "Open, Pool A"}, 20: {Heading: "Guts Open, Pool A"}},
		Reservations: map[int64]ReservationPage{100: {LocationID: 1}, 200: {LocationID: 2}},
		Locations:    map[int64]Location{1: {}, 2: {}},
		TeamCards:    map[int64]TeamCard{5: {}, 6: {}},
		Games:        map[int64]GamePage{1: {}, 2: {}},
		GamesCSV:     []map[string]string{{"Division": "Open"}, {"Division": "Guts Open"}},
		SpiritCSV:    []map[string]string{{"division": "Guts Open"}},
	}
	if err := snap.DropDivisions([]string{"Guts Open"}); err != nil {
		t.Fatal(err)
	}
	if len(snap.Menu.Divisions) != 1 || len(snap.Schedule) != 1 || len(snap.Teams) != 1 || len(snap.Seeds) != 1 ||
		len(snap.Pools) != 1 || len(snap.TeamCards) != 1 || len(snap.Games) != 1 || len(snap.GamesCSV) != 1 || len(snap.SpiritCSV) != 0 {
		t.Errorf("left %+v", snap)
	}
	if _, ok := snap.Reservations[200]; ok || len(snap.Reservations) != 1 || len(snap.Locations) != 1 {
		t.Errorf("reservations %v, locations %v", snap.Reservations, snap.Locations)
	}
	if err := snap.DropDivisions([]string{"Beach"}); err == nil {
		t.Error("unknown division dropped")
	}
}
