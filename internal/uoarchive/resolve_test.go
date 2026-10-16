package uoarchive

import (
	"testing"

	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

func TestReservationTimes(t *testing.T) {
	start, end, err := reservationTimes("10.11.2024 11:00 - 19:00")
	if err != nil || start != "2024-11-10 11:00:00" || end != "2024-11-10 19:00:00" {
		t.Errorf("got %q %q %v", start, end, err)
	}
	start, end, err = reservationTimes("28.6.2026 09:00 - 18:30")
	if err != nil || start != "2026-06-28 09:00:00" || end != "2026-06-28 18:30:00" {
		t.Errorf("unpadded: got %q %q %v", start, end, err)
	}
	if _, _, err := reservationTimes("2024-11-10 11:00"); err == nil {
		t.Error("accepted an unexpected format")
	}
}

func TestResolveRef(t *testing.T) {
	num := func(n int64) *int64 { return &n }
	board := []uoclient.ScoreboardPlayer{
		{PlayerID: 1, Num: num(3), FirstName: "Juan ", LastName: "Arias"},
		{PlayerID: 2, Num: num(23), FirstName: "Mario ", LastName: "Plaza "},
		{PlayerID: 3, Num: num(7), FirstName: "Sam", LastName: "Lee"},
		{PlayerID: 4, Num: num(8), FirstName: "Sam", LastName: "Lee"},
	}
	cases := []struct {
		ref  uoclient.PlayerRef
		want int64
	}{
		{uoclient.PlayerRef{Num: num(23), Name: "Mario Plaza"}, 2},
		{uoclient.PlayerRef{Num: num(99), Name: "Juan Arias"}, 1}, // number differs this game
		{uoclient.PlayerRef{Name: "Juan Arias"}, 1},
		{uoclient.PlayerRef{Num: num(8), Name: "Sam Lee"}, 4}, // same name, told apart by number
	}
	for _, c := range cases {
		got, err := resolveRef(board, c.ref)
		if err != nil || got == nil || *got != c.want {
			t.Errorf("resolveRef(%+v) = %v, %v; want %d", c.ref, got, err, c.want)
		}
	}
	if got, err := resolveRef(board, uoclient.PlayerRef{}); got != nil || err != nil {
		t.Errorf("empty ref = %v, %v", got, err)
	}
	if _, err := resolveRef(board, uoclient.PlayerRef{Name: "Sam Lee"}); err == nil {
		t.Error("ambiguous name without a number resolved")
	}
	if _, err := resolveRef(board, uoclient.PlayerRef{Name: "Nobody"}); err == nil {
		t.Error("unknown name resolved")
	}

	roster := append(board[:0:0], board...)
	roster = append(roster, uoclient.ScoreboardPlayer{PlayerID: 5, Num: num(88), FirstName: "Kelly", LastName: "Hyland"})
	if got, err := resolveScorer(board, roster, uoclient.PlayerRef{Num: num(88), Name: "Kelly Hyland"}); err != nil || got == nil || *got != 5 {
		t.Errorf("player not on the scoreboard = %v, %v; want 5 from the roster", got, err)
	}
	if _, err := resolveScorer(board, roster, uoclient.PlayerRef{Name: "Sam Lee"}); err == nil {
		t.Error("ambiguous scoreboard name fell back to the roster")
	}
	if _, err := resolveScorer(board, roster, uoclient.PlayerRef{Name: "Nobody"}); err == nil {
		t.Error("unknown name resolved from the roster")
	}
}

func TestResolveSpirit(t *testing.T) {
	id := func(n int64) *int64 { return &n }
	snap := &uoclient.Snapshot{
		Schedule: []uoclient.ScheduledGame{
			{GameID: id(1), Division: "Open", ReservationGroup: "Sunday", FieldLabel: "Field 0", Time: "09:00", PoolName: "Pool A", Home: "Red", Visitor: "Blue"},
			{GameID: id(2), Division: "Open", ReservationGroup: "Sunday", FieldLabel: "Field 0", Time: "11:00", PoolName: "Pool A", Home: "Blue", Visitor: "Red"},
		},
		Games: map[int64]uoclient.GamePage{
			1: {GameID: 1, Spirit: &uoclient.GameSpirit{Home: [5]int64{2, 2, 2, 2, 2}, Visitor: [5]int64{1, 2, 3, 2, 1}}},
			2: {GameID: 2},
		},
		SpiritCSV: []map[string]string{
			{"division": "Open", "day": "Sunday", "field": "0 ", "time": "09:00", "pool": "Pool A", "team_evaluated": "Red", "by_team": "Blue", "rules": "2", "fouls": "2", "fair": "2", "positive": "2", "com": "2", "comments": ""},
			{"division": "Open", "day": "Sunday", "field": "0 ", "time": "09:00", "pool": "Pool A", "team_evaluated": "Blue", "by_team": "Red", "rules": "1", "fouls": "2", "fair": "3", "positive": "2", "com": "1", "comments": "thanks"},
			// game 2: only Blue's opponent has submitted, hidden on the game page
			{"division": "Open", "day": "Sunday", "field": "0 ", "time": "11:00", "pool": "Pool A", "team_evaluated": "Blue", "by_team": "Red", "rules": "3", "fouls": "3", "fair": "3", "positive": "3", "com": "3", "comments": "great"},
		},
	}
	newRows := func() *rows {
		return &rows{games: []store.Game{
			{GameID: 1, Hometeam: id(10), Visitorteam: id(20)},
			{GameID: 2, Hometeam: id(20), Visitorteam: id(10)},
		}}
	}

	r := newRows()
	if err := r.resolveSpirit(snap); err != nil {
		t.Fatal(err)
	}
	if len(r.spiritScores) != 3 {
		t.Fatalf("%d spirit scores, want 3 (one-sided game 2 included)", len(r.spiritScores))
	}
	if g2 := r.games[1]; g2.Homesotg == nil || *g2.Homesotg != 15 || g2.Visitorsotg != nil {
		t.Errorf("game 2 sotg %v %v, want home 15 and visitor unset", g2.Homesotg, g2.Visitorsotg)
	}
	if c := r.spiritScores[1]; c.TeamID != 20 || c.Comments != "thanks" {
		t.Errorf("comments not kept: %+v", c)
	}

	snap.SpiritCSV[0]["rules"] = "4"
	if err := newRows().resolveSpirit(snap); err == nil {
		t.Error("an export score disagreeing with the game page was accepted")
	}
}

func TestShowSpiritComments(t *testing.T) {
	text := ""
	shown := uoclient.GamePage{Spirit: &uoclient.GameSpirit{HomeComments: &text, VisitorComments: &text}}
	hidden := uoclient.GamePage{Spirit: &uoclient.GameSpirit{}}
	none := uoclient.GamePage{}
	cases := []struct {
		name  string
		games map[int64]uoclient.GamePage
		want  *bool
	}{
		{"shown", map[int64]uoclient.GamePage{1: shown, 2: none}, new(true)},
		{"hidden", map[int64]uoclient.GamePage{1: hidden, 2: none}, new(false)},
		{"no spirit shown", map[int64]uoclient.GamePage{1: none}, nil},
		{"mixed", map[int64]uoclient.GamePage{1: shown, 2: hidden}, nil},
	}
	for _, c := range cases {
		got := showSpiritComments(&uoclient.Snapshot{Games: c.games})
		if (got == nil) != (c.want == nil) || (got != nil && *got != *c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestResolvePlayersFromScoreboard(t *testing.T) {
	games := func(n int64) *int64 { return &n }
	snap := &uoclient.Snapshot{
		Teams: []uoclient.TeamListEntry{{TeamID: 30, Name: "AUS Mixed"}},
		TeamCards: map[int64]uoclient.TeamCard{30: {TeamID: 30, FromScoreboard: true, Roster: []uoclient.RosterPlayer{
			{PlayerID: 42, Name: "Jake Upton", Games: games(10)},
			{PlayerID: 41, Name: "Jacob Ellis", Games: games(0)},
		}}},
		PlayersCSV: []map[string]string{
			{"FirstName": "Jake", "LastName": "Upton", "Jersey": "7", "TeamName": "AUS Mixed"},
			{"FirstName": "Jacob", "LastName": "Ellis", "Jersey": "-1", "TeamName": "AUS Mixed"},
		},
	}
	r := &rows{}
	if err := r.resolvePlayers(snap); err != nil {
		t.Fatal(err)
	}
	if len(r.players) != 2 {
		t.Fatalf("players %+v", r.players)
	}
	ellis, upton := r.players[0], r.players[1]
	if upton.Num == nil || *upton.Num != 7 || upton.FirstName != "Jake" || *upton.GamesPlayed != 10 {
		t.Errorf("upton %+v", upton)
	}
	if ellis.Num != nil || ellis.LastName != "Ellis" {
		t.Errorf("negative jersey kept: %+v", ellis)
	}

	snap.PlayersCSV = append(snap.PlayersCSV, map[string]string{"FirstName": "Jake", "LastName": "Upton", "Jersey": "8", "TeamName": "AUS Mixed"})
	if err := (&rows{}).resolvePlayers(snap); err == nil {
		t.Error("name on the team twice in the CSV resolved")
	}
}

func TestResolveGamesWithoutScoresheet(t *testing.T) {
	id := func(n int64) *int64 { return &n }
	newSnap := func() *uoclient.Snapshot {
		return &uoclient.Snapshot{
			Teams: []uoclient.TeamListEntry{{TeamID: 10, Name: "Red"}, {TeamID: 20, Name: "Blue"}},
			Schedule: []uoclient.ScheduledGame{
				{GameID: id(5), ReservationID: 1, ReservationGroup: "Saturday", FieldLabel: "Field 1", Time: "09:00", Home: "Red", Visitor: "Blue", HomeScore: id(15), VisitorScore: id(8), PoolID: 3},
				{ReservationID: 1, ReservationGroup: "Saturday", FieldLabel: "Field 1", Time: "11:00", Home: "Blue", Visitor: "Red", HomeScore: id(13), VisitorScore: id(11), PoolID: 3},
				{ReservationID: 1, ReservationGroup: "Saturday", FieldLabel: "Field 1", Time: "13:00", Home: "Red", Visitor: "Blue", PoolID: 3},
			},
			Games: map[int64]uoclient.GamePage{5: {GameID: 5, HasScoresheet: true}},
			GamesCSV: []map[string]string{
				{"Time": "2026-07-25 09:00:00", "HomeTeam": "Red", "AwayTeam": "Blue", "Field": "1", "ReservationGroup": "Saturday"},
				{"Time": "2026-07-25 11:00:00", "HomeTeam": "Blue", "AwayTeam": "Red", "Field": "1", "ReservationGroup": "Saturday"},
				{"Time": "2026-07-25 13:00:00", "HomeTeam": "Red", "AwayTeam": "Blue", "Field": "1", "ReservationGroup": "Saturday"},
			},
		}
	}

	snap := newSnap()
	r := &rows{}
	if err := r.resolveGames(snap, false); err != nil {
		t.Fatal(err)
	}
	if len(r.games) != 3 || r.games[0].GameID != -2 || r.games[1].GameID != -1 || r.games[2].GameID != 5 {
		t.Fatalf("games %+v, want synthetic -2, -1 and linked 5", r.games)
	}
	played := r.games[1]
	if *played.Hometeam != 20 || *played.Visitorteam != 10 || *played.Homescore != 13 || *played.Visitorscore != 11 || played.Time != "2026-07-25 11:00:00" {
		t.Errorf("result-only game %+v", played)
	}
	if unplayed := r.games[0]; unplayed.Homescore != nil || unplayed.Visitorscore != nil {
		t.Errorf("unplayed game got a score: %+v", unplayed)
	}
	if g := snap.Schedule[1].GameID; g == nil || *g != -1 {
		t.Errorf("schedule row not given its synthetic id for spirit: %v", g)
	}

	r = &rows{}
	if err := r.resolveGames(newSnap(), true); err != nil {
		t.Fatal(err)
	}
	if len(r.games) != 1 || len(r.unlinked) != 2 {
		t.Errorf("allowUnlinked: %d games, %d unlinked; want 1 and 2", len(r.games), len(r.unlinked))
	}
}
