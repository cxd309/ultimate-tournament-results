// Package enrich fills an existing 1.9.14-1.9.17 archive's spirit_scores in from the
// UltiOrganizer site under its Live! deployment, whose ext/spiritcsv.php export has
// the one-sided scores (a game where only one team has submitted) and the comments
// the Live! API leaves out
//
// The archive stays the same file with the same tables, spirit_scores is just more
// complete. db/enrich/schema.sql's tables record which rows and comments were added,
// so livepublish/v01_09_14 leaves them out and the published Live! JSON stays what the
// API served. Enrich runs once per archive, and a locked or empty export writes nothing
package enrich

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	enrichstore "github.com/cxd309/ultimate-tournament-results/internal/store/enrich"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	store030006 "github.com/cxd309/ultimate-tournament-results/internal/store/v03_00_06"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

// schemaPath is applied to the archive in the same transaction as the rows it holds
const schemaPath = "db/enrich/schema.sql"

// Source is the UltiOrganizer site to read the export from, every field optional:
// Host and BasePath default to the archive's own deployment with its Live! base path
// (/live/data/) taken off, and SeasonID to the archive's season
type Source struct {
	Host     string
	BasePath string
	SeasonID string
}

// Enrich fills the archive at dbPath in from src's spirit export, see the package doc
func Enrich(ctx context.Context, dbPath string, src Source) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	s, es := store.New(db), enrichstore.New(db)
	if enriched, err := es.Enriched(ctx); err != nil {
		return err
	} else if enriched {
		return fmt.Errorf("%s is already enriched; enrich runs once per archive", dbPath)
	}
	t, err := s.GetTournament(ctx)
	if err != nil {
		return fmt.Errorf("tournament: %w", err)
	}
	src, err = src.withDefaults(t.Host, t.BasePath, t.SeasonID)
	if err != nil {
		return err
	}

	export, err := fetch(ctx, src)
	if err != nil || export == nil {
		return err
	}

	a, err := load(ctx, s)
	if err != nil {
		return err
	}
	changes, err := match(a, export)
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		fmt.Printf("export's %d scores are all already archived, nothing to add\n", len(export))
		return nil
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("apply schema %s: %w", schemaPath, err)
	}
	if err := write(ctx, store.New(tx), enrichstore.New(tx), src, changes); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Enrich030006 is enrich for a 3.0.6 archive. UltiOrganizer 3.x serves the export only
// to a season made public externally, which no 3.x deployment has done, so this only
// checks: a locked or empty export is a no-op, an open one an error until 3.x
// enrichment is written
func Enrich030006(ctx context.Context, dbPath string, src Source) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	t, err := store030006.New(db).GetTournament(ctx)
	if err != nil {
		return fmt.Errorf("tournament: %w", err)
	}
	src, err = src.withDefaults(t.Host, t.BasePath, t.SeasonID)
	if err != nil {
		return err
	}
	export, err := fetch(ctx, src)
	if err != nil || export == nil {
		return err
	}
	return fmt.Errorf("%s serves %d spirit scores, but enriching a 3.0.6 archive isn't supported yet", src.url(), len(export))
}

// fetch reads src's spirit export; nil, with a note, when there's nothing to enrich from
func fetch(ctx context.Context, src Source) ([]map[string]string, error) {
	c := uoclient.NewClient(src.Host, src.BasePath)
	c.SeasonID = src.SeasonID
	fmt.Printf("fetching %s...\n", src.url())
	export, err := c.SpiritCSV(ctx)
	switch {
	case errors.Is(err, uoclient.ErrNotCSV):
		fmt.Println("  spirit export is locked, nothing to enrich from")
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("spirit csv: %w", err)
	case len(export) == 0:
		fmt.Println("  spirit export is empty, nothing to enrich from")
		return nil, nil
	}
	return export, nil
}

func load(ctx context.Context, s *store.Store) (archive, error) {
	var a archive
	var err error
	if a.divisions, err = s.ListDivisions(ctx); err != nil {
		return a, fmt.Errorf("divisions: %w", err)
	}
	if a.pools, err = s.ListPools(ctx); err != nil {
		return a, fmt.Errorf("pools: %w", err)
	}
	if a.reservations, err = s.ListReservations(ctx); err != nil {
		return a, fmt.Errorf("reservations: %w", err)
	}
	if a.teams, err = s.ListTeams(ctx); err != nil {
		return a, fmt.Errorf("teams: %w", err)
	}
	if a.games, err = s.ListGames(ctx); err != nil {
		return a, fmt.Errorf("games: %w", err)
	}
	if a.spirit, err = s.ListSpiritScores(ctx); err != nil {
		return a, fmt.Errorf("spirit scores: %w", err)
	}
	return a, nil
}

func write(ctx context.Context, s *store.Store, es *enrichstore.Store, src Source, changes []change) error {
	var scores, comments int
	for _, c := range changes {
		sc := c.score
		if c.added {
			if err := s.InsertSpiritScore(ctx, sc); err != nil {
				return fmt.Errorf("insert spirit score game %d team %d: %w", sc.GameID, sc.TeamID, err)
			}
			scores++
		} else {
			if err := es.SetSpiritComments(ctx, sc.GameID, sc.TeamID, sc.Comments); err != nil {
				return fmt.Errorf("set comments game %d team %d: %w", sc.GameID, sc.TeamID, err)
			}
		}
		if sc.Comments != "" {
			comments++
		}
		if err := es.InsertEnrichedSpirit(ctx, enrichstore.EnrichedSpirit{GameID: sc.GameID, TeamID: sc.TeamID, Score: c.added, Comments: c.comment}); err != nil {
			return fmt.Errorf("insert enriched spirit game %d team %d: %w", sc.GameID, sc.TeamID, err)
		}
	}
	if err := es.InsertEnrichment(ctx, enrichstore.Enrichment{
		Host:       src.Host,
		BasePath:   src.BasePath,
		SeasonID:   src.SeasonID,
		EnrichedAt: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return fmt.Errorf("insert enrichment: %w", err)
	}
	fmt.Printf("added %d spirit scores and %d comments\n", scores, comments)
	return nil
}

// withDefaults fills src's empty fields from the archive's own host, Live! base path
// and season. A host can carry a path of its own (results.wfdf.sport/wucc-2026), which
// becomes part of the base path
func (src Source) withDefaults(host, livePath, seasonID string) (Source, error) {
	if src.SeasonID == "" {
		src.SeasonID = seasonID
	}
	if src.Host != "" && src.BasePath == "" {
		return src, fmt.Errorf("-base-path is required with -host")
	}
	if src.Host == "" {
		src.Host = host
		if i := strings.Index(host, "/"); i >= 0 {
			src.Host, livePath = host[:i], host[i:]+livePath
		}
		path, ok := strings.CutSuffix(livePath, "live/data/")
		if !ok {
			return src, fmt.Errorf("archive's base path %q doesn't end live/data/, pass -host and -base-path for its UltiOrganizer site", livePath)
		}
		src.BasePath = path
	}
	return src, nil
}

func (src Source) url() string {
	return fmt.Sprintf("https://%s%sext/spiritcsv.php?season=%s", src.Host, src.BasePath, src.SeasonID)
}
