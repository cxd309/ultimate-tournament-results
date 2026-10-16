package livepublish

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"

	enrichstore "github.com/cxd309/ultimate-tournament-results/internal/store/enrich"
	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

func TestDropEnriched(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	es := enrichstore.New(db)

	scores := func() map[int64][]store.SpiritScore {
		return map[int64][]store.SpiritScore{
			1: {{GameID: 1, TeamID: 7, Cat1: 2}, {GameID: 1, TeamID: 8, Cat1: 3, Comments: "nice"}},
			2: {{GameID: 2, TeamID: 7, Cat1: 4, Comments: "great"}},
		}
	}

	// never enriched: no enrich tables at all, nothing dropped
	data := &tournamentData{spiritScoresByGame: scores()}
	if err := dropEnriched(ctx, es, data); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data.spiritScoresByGame, scores()) {
		t.Errorf("never-enriched archive changed: %+v", data.spiritScoresByGame)
	}

	schema, err := os.ReadFile("../../../db/enrich/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	for _, e := range []enrichstore.EnrichedSpirit{
		{GameID: 1, TeamID: 8, Comments: true},
		{GameID: 2, TeamID: 7, Score: true},
	} {
		if err := es.InsertEnrichedSpirit(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	data = &tournamentData{spiritScoresByGame: scores()}
	if err := dropEnriched(ctx, es, data); err != nil {
		t.Fatal(err)
	}
	want := map[int64][]store.SpiritScore{
		1: {{GameID: 1, TeamID: 7, Cat1: 2}, {GameID: 1, TeamID: 8, Cat1: 3}},
	}
	if !reflect.DeepEqual(data.spiritScoresByGame, want) {
		t.Errorf("got %+v, want %+v", data.spiritScoresByGame, want)
	}
}

func TestHideSpiritComments(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	uo := uostore.New(db)

	comments := func(data *tournamentData) []string {
		var out []string
		for _, sc := range data.spiritScoresByGame[1] {
			out = append(out, sc.Comments)
		}
		return out
	}
	check := func(name string, want ...string) {
		t.Helper()
		data := &tournamentData{spiritScoresByGame: map[int64][]store.SpiritScore{
			1: {{GameID: 1, TeamID: 7, Comments: "nice"}, {GameID: 1, TeamID: 8}},
		}}
		if err := hideSpiritComments(ctx, uo, data); err != nil {
			t.Fatal(err)
		}
		if got := comments(data); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}

	check("plain Live! archive", "nice", "")

	schema, err := os.ReadFile("../../../db/uo_01/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		show *bool
		want []string
	}{
		{"setting unknown", nil, []string{"", ""}},
		{"setting off", new(false), []string{"", ""}},
		{"setting on", new(true), []string{"nice", ""}},
	} {
		if _, err := db.Exec("DELETE FROM server_settings"); err != nil {
			t.Fatal(err)
		}
		if err := uo.InsertServerSettings(ctx, uostore.ServerSettings{ShowSpiritComments: c.show}); err != nil {
			t.Fatal(err)
		}
		check(c.name, c.want...)
	}
}
