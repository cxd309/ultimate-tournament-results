// Command archive fetches one tournament's results from a Live! by BULA 1.9.14-1.9.16
// deployment and writes them into a fresh, per-tournament SQLite archive file. Find the
// host and base path for a known deployment in live-by-bula-openapi's README.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/cxd309/ultimate-tournament-results/internal/bula/v1914"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v1914"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "archive:", err)
		os.Exit(1)
	}
}

func run() error {
	host := flag.String("host", "", "deployment host, e.g. wbuc.wfdf.sport (required)")
	basePath := flag.String("base-path", "/live/data/", "static cache base path, e.g. /live/data/")
	slug := flag.String("slug", "", "archive filename slug (default: the season id reported by the heartbeat, lowercased)")
	dbPath := flag.String("db", "", "path to write the new sqlite archive file (default data/<slug>.db); must not already exist")
	flag.Parse()

	if *host == "" {
		return fmt.Errorf("-host is required")
	}

	client := v1914.NewClient(*host, *basePath)
	ctx := context.Background()

	hb, err := client.FetchHeartbeat(ctx)
	if err != nil {
		return fmt.Errorf("fetch heartbeat: %w", err)
	}
	fmt.Printf("heartbeat: app_version=%s season_id=%s\n", hb.AppVersion, hb.Config.LiveSeasonID)

	ref, err := client.FetchReference(ctx)
	if err != nil {
		return fmt.Errorf("fetch reference: %w", err)
	}
	fmt.Printf("reference: name=%q divisions=%d pools=%d countries=%d\n", ref.Season.Name, len(ref.Series), len(ref.Pools), len(ref.Countries))

	teams, err := client.FetchTeams(ctx)
	if err != nil {
		return fmt.Errorf("fetch teams: %w", err)
	}
	fmt.Printf("teams: %d\n", len(teams.Teams))

	detailByTeamID := make(map[int64]*v1914.TeamDetailResponse, len(teams.Teams))
	for _, ts := range teams.Teams {
		detail, err := client.FetchTeamDetail(ctx, ts.TeamID)
		if err != nil {
			return fmt.Errorf("fetch team detail %d: %w", ts.TeamID, err)
		}
		detailByTeamID[detail.TeamID] = detail
	}
	fmt.Printf("team details: %d\n", len(detailByTeamID))

	if *slug == "" {
		*slug = strings.ToLower(hb.Config.LiveSeasonID)
	}
	if *dbPath == "" {
		*dbPath = fmt.Sprintf("data/%s.db", *slug)
	}

	if _, err := os.Stat(*dbPath); err == nil {
		return fmt.Errorf("%s already exists; archives are write-once, remove it first if you want to rebuild", *dbPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check %s: %w", *dbPath, err)
	}

	schema, err := os.ReadFile("db/v1914/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", *dbPath, err)
	}
	defer sqlDB.Close()

	if _, err := sqlDB.Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	s := store.New(sqlDB)

	if err := v1914.ImportTournament(ctx, s, *host, *basePath, hb, ref); err != nil {
		return fmt.Errorf("import tournament: %w", err)
	}
	refIDs, err := v1914.ImportReferenceData(ctx, s, ref)
	if err != nil {
		return fmt.Errorf("import reference data: %w", err)
	}
	if err := v1914.ImportTeams(ctx, s, refIDs, ref, teams, detailByTeamID); err != nil {
		return fmt.Errorf("import teams: %w", err)
	}

	tournament, err := s.GetTournament(ctx)
	if err != nil {
		return fmt.Errorf("read back tournament: %w", err)
	}
	fmt.Printf("wrote %s: event_name=%q season_id=%s start_date=%s\n", *dbPath, tournament.EventName, tournament.SeasonID, tournament.StartDate)

	return nil
}
