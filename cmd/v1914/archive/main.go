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

	ctx := context.Background()
	client := v1914.NewClient(*host, *basePath)

	snap, err := v1914.Gather(ctx, client)
	if err != nil {
		return fmt.Errorf("gather: %w", err)
	}
	fmt.Printf("gathered: season_id=%s name=%q divisions=%d pools=%d countries=%d teams=%d\n",
		snap.Heartbeat.Config.LiveSeasonID, snap.Reference.Season.Name, len(snap.Reference.Series), len(snap.Reference.Pools), len(snap.Reference.Countries), len(snap.Reference.Teams))

	if *slug == "" {
		*slug = strings.ToLower(snap.Heartbeat.Config.LiveSeasonID)
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
	if err := v1914.Import(ctx, s, *host, *basePath, snap); err != nil {
		return fmt.Errorf("import: %w", err)
	}

	tournament, err := s.GetTournament(ctx)
	if err != nil {
		return fmt.Errorf("read back tournament: %w", err)
	}
	fmt.Printf("wrote %s: name=%q season_id=%s start=%s\n", *dbPath, tournament.Name, tournament.SeasonID, tournament.StartTime)

	return nil
}
