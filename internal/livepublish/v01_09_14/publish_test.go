package livepublish

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"

	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

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
