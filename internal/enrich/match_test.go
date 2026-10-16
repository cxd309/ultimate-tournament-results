package enrich

import (
	"maps"
	"strconv"
	"strings"
	"testing"

	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

func testArchive() archive {
	id := func(n int64) *int64 { return &n }
	return archive{
		divisions:    []store.Division{{SeriesID: 1, Name: "Open"}},
		pools:        []store.Pool{{PoolID: 10, Name: "Pool A", Series: id(1)}},
		reservations: []store.Reservation{{ID: 100, FieldName: "H01", ReservationGroup: "Sunday"}},
		teams:        []store.Team{{TeamID: 7, Name: "Red"}, {TeamID: 8, Name: "Blue"}},
		games: []store.Game{
			// game 1: both teams submitted, Live! served both scores, comment for Blue only
			{GameID: 1, Hometeam: id(7), Visitorteam: id(8), Reservation: id(100), Pool: id(10), Time: "2026-06-28 09:00:00", Homesotg: id(10), Visitorsotg: id(11)},
			// game 2: only Red's opponent submitted, Live! served its total but no score
			{GameID: 2, Hometeam: id(8), Visitorteam: id(7), Reservation: id(100), Pool: id(10), Time: "2026-06-28 11:00:00", Visitorsotg: id(12)},
		},
		spirit: []store.SpiritScore{
			{GameID: 1, TeamID: 7, Cat1: 2, Cat2: 2, Cat3: 2, Cat4: 2, Cat5: 2},
			{GameID: 1, TeamID: 8, Cat1: 2, Cat2: 2, Cat3: 2, Cat4: 2, Cat5: 3, Comments: "nice"},
		},
	}
}

func row(time, evaluated, by, cats, comments string) map[string]string {
	c := strings.Split(cats, ",")
	total := 0
	for _, n := range c {
		v, _ := strconv.Atoi(n)
		total += v
	}
	return map[string]string{
		"division": "Open", "day": "Sunday", "field": "H01", "time": time, "pool": "Pool A",
		"team_evaluated": evaluated, "by_team": by,
		"rules": c[0], "fouls": c[1], "fair": c[2], "positive": c[3], "com": c[4],
		"total": strconv.Itoa(total), "comments": comments,
	}
}

func testExport() []map[string]string {
	return []map[string]string{
		row("09:00", "Red", "Blue", "2,2,2,2,2", "well played"),
		row("09:00", "Blue", "Red", "2,2,2,2,3", "nice"),
		row("11:00", "Red", "Blue", "3,2,2,2,3", "great"),
	}
}

func TestMatch(t *testing.T) {
	changes, err := match(testArchive(), testExport())
	if err != nil {
		t.Fatal(err)
	}
	want := []change{
		{score: store.SpiritScore{GameID: 1, TeamID: 7, Cat1: 2, Cat2: 2, Cat3: 2, Cat4: 2, Cat5: 2, Comments: "well played"}, comment: true},
		{score: store.SpiritScore{GameID: 2, TeamID: 7, Cat1: 3, Cat2: 2, Cat3: 2, Cat4: 2, Cat5: 3, Comments: "great"}, added: true},
	}
	if len(changes) != len(want) {
		t.Fatalf("got %+v, want %+v", changes, want)
	}
	for i := range want {
		if changes[i] != want[i] {
			t.Errorf("change %d = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestMatchErrors(t *testing.T) {
	cases := map[string]func(a *archive, e []map[string]string) []map[string]string{
		"category disagreement": func(_ *archive, e []map[string]string) []map[string]string {
			e[1] = row("09:00", "Blue", "Red", "3,2,2,2,2", "nice")
			return e
		},
		"comment disagreement": func(_ *archive, e []map[string]string) []map[string]string {
			e[1]["comments"] = "something else"
			return e
		},
		"total disagrees with sotg": func(a *archive, e []map[string]string) []map[string]string {
			v := int64(13)
			a.games[1].Visitorsotg = &v
			return e
		},
		"total column wrong": func(_ *archive, e []map[string]string) []map[string]string {
			e[0]["total"] = "11"
			return e
		},
		"no game": func(_ *archive, e []map[string]string) []map[string]string {
			e[0]["field"] = "H02"
			return e
		},
		"two games": func(a *archive, e []map[string]string) []map[string]string {
			g := a.games[0]
			g.GameID = 3
			a.games = append(a.games, g)
			return e
		},
		"two rows for one team": func(_ *archive, e []map[string]string) []map[string]string {
			return append(e, maps.Clone(e[2]))
		},
		"served score missing from export": func(_ *archive, e []map[string]string) []map[string]string {
			return e[1:]
		},
	}
	for name, edit := range cases {
		a := testArchive()
		e := edit(&a, testExport())
		if _, err := match(a, e); err == nil {
			t.Errorf("%s: matched without an error", name)
		}
	}
}

func TestMatchWhitespace(t *testing.T) {
	a := testArchive()
	a.teams[0].Name = "Red  "
	e := testExport()
	e[0]["team_evaluated"] = " Red"
	if _, err := match(a, e); err != nil {
		t.Error(err)
	}
}

func TestSourceDefaults(t *testing.T) {
	cases := []struct {
		src                      Source
		host, livePath, seasonID string
		want                     Source
	}{
		{Source{}, "eucf.ultimatefederation.eu", "/live/data/", "26EUCFWRO", Source{"eucf.ultimatefederation.eu", "/", "26EUCFWRO"}},
		{Source{}, "results.wfdf.sport", "/aouc/live/data/", "AOUC2025", Source{"results.wfdf.sport", "/aouc/", "AOUC2025"}},
		{Source{}, "results.wfdf.sport/wucc-2026", "/live/data/", "WUCC2026", Source{"results.wfdf.sport", "/wucc-2026/", "WUCC2026"}},
		{Source{Host: "results.wfdf.sport", BasePath: "/wmucc-2026/"}, "wmucc.wfdf.sport", "/live/data/", "wmucc2026", Source{"results.wfdf.sport", "/wmucc-2026/", "wmucc2026"}},
	}
	for _, c := range cases {
		got, err := c.src.withDefaults(c.host, c.livePath, c.seasonID)
		if err != nil || got != c.want {
			t.Errorf("%+v.withDefaults(%q, %q, %q) = %+v, %v; want %+v", c.src, c.host, c.livePath, c.seasonID, got, err, c.want)
		}
	}
	if _, err := (Source{Host: "results.wfdf.sport"}).withDefaults("x", "/live/data/", "s"); err == nil {
		t.Error("-host without -base-path accepted")
	}
	if _, err := (Source{}).withDefaults("x", "/data/", "s"); err == nil {
		t.Error("base path not ending live/data/ accepted")
	}
}
