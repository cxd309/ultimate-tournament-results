// Command utr archives and publishes Live! by BULA tournament results, and results
// from classic UltiOrganizer sites with no Live! deployment
// it multiplexes across every supported source version via -version
// see internal/liveversion for the registry of implementations
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/cxd309/ultimate-tournament-results/internal/livearchive"
	"github.com/cxd309/ultimate-tournament-results/internal/liveversion"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "utr:", err)
		os.Exit(1)
	}
}

func run() error {
	versionKey := flag.String("version", "", "source version to target, a Live! by BULA spec version or uo_01 for a classic UltiOrganizer site, one of: "+strings.Join(liveversion.Keys(), ", ")+" (required)")
	mode := flag.String("mode", "all", "what to run: archive, publish, or all")
	host := flag.String("host", "", "deployment host, e.g. wbuc.wfdf.sport (required for archive)")
	basePath := flag.String("base-path", "", "Live!: static cache base path (default /live/data/); uo_01: the UltiOrganizer install's path, e.g. /wmuc/ (required)")
	slug := flag.String("slug", "", "archive filename slug (default: the season id reported by the heartbeat, lowercased)")
	dbPath := flag.String("db", "", "sqlite archive path: written by archive (default data/<slug>.db; must not exist), read by publish (must exist)")
	outDir := flag.String("out", "docs/archive/", "directory to render published JSON into")
	seasonID := flag.String("season-id", "", "override the season id instead of discovering it from the heartbeat (legacy deployments only); uo_01: the season to archive, needed when the site hosts more than one")
	cacheDir := flag.String("cache", "", "uo_01 only: directory to keep fetched pages in and reuse on a re-run, e.g. after a failed archive")
	unprefixed := flag.Bool("unprefixed", false, "static JSON filenames carry no season-id prefix (legacy deployments only)")
	flag.Parse()

	v, ok := liveversion.Get(*versionKey)
	if !ok {
		return fmt.Errorf("-version is required, one of: %s", strings.Join(liveversion.Keys(), ", "))
	}

	if *basePath == "" {
		*basePath = v.DefaultBasePath
	}

	opts := livearchive.ArchiveOptions{SeasonID: *seasonID, Unprefixed: *unprefixed, CacheDir: *cacheDir}

	ctx := context.Background()

	switch *mode {
	case "archive":
		if err := requireSource(*host, *basePath); err != nil {
			return err
		}
		if _, err := v.Archive(ctx, *host, *basePath, *slug, *dbPath, opts); err != nil {
			return fmt.Errorf("archive: %w", err)
		}
		return nil

	case "publish":
		if err := requireExistingDB(*dbPath); err != nil {
			return err
		}
		if err := v.Publish(ctx, *dbPath, *outDir); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		return nil

	case "all":
		if err := requireSource(*host, *basePath); err != nil {
			return err
		}
		summary, err := v.Archive(ctx, *host, *basePath, *slug, *dbPath, opts)
		if err != nil {
			return fmt.Errorf("archive: %w", err)
		}
		if err := v.Publish(ctx, summary.DBPath, *outDir); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("-mode must be archive, publish, or all (got %q)", *mode)
	}
}

// requireSource checks host and basePath are set for a mode that archives
func requireSource(host, basePath string) error {
	if host == "" {
		return fmt.Errorf("-host is required to archive")
	}
	if basePath == "" {
		return fmt.Errorf("-base-path is required to archive with this -version")
	}
	return nil
}

// requireExistingDB checks dbPath is set and points at an archive that already exists
// -- the inverse of archive's write-once check
func requireExistingDB(dbPath string) error {
	if dbPath == "" {
		return fmt.Errorf("-db is required for -mode publish")
	}
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist; archive it first with -mode archive", dbPath)
		}
		return fmt.Errorf("check %s: %w", dbPath, err)
	}
	return nil
}
