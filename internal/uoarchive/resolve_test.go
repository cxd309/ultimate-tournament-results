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
