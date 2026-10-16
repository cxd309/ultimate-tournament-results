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
	"github.com/cxd309/ultimate-tournament-results/internal/uoarchive"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "utr:", err)
		os.Exit(1)
	}
}

func run() error {
	versionKey := flag.String("version", "", "source version to target, a Live! by BULA spec version or uo_01 for a classic UltiOrganizer site, one of: "+strings.Join(liveversion.Keys(), ", ")+" (required)")
	mode := flag.String("mode", "all", "what to run: archive, publish, all, or enrich (turn an existing Live! archive into a uo_01 one from its UltiOrganizer site)")
	host := flag.String("host", "", "deployment host, e.g. wbuc.wfdf.sport (required for archive); enrich: the UltiOrganizer site's host (default: the archive's own)")
	basePath := flag.String("base-path", "", "Live!: static cache base path (default /live/data/); uo_01: the UltiOrganizer install's path, e.g. /wmuc/ (required); enrich: the UltiOrganizer install's path (default: the archive's base path less live/data/)")
	slug := flag.String("slug", "", "archive filename slug (default: the season id reported by the heartbeat, lowercased)")
	dbPath := flag.String("db", "", "sqlite archive path: written by archive (default data/<slug>.db; must not exist), read by publish and updated by enrich (must exist)")
	outDir := flag.String("out", "docs/archive/", "directory to render published JSON into")
	seasonID := flag.String("season-id", "", "override the season id instead of discovering it from the heartbeat (legacy deployments only); uo_01: the season to archive, needed when the site hosts more than one")
	cacheDir := flag.String("cache", "", "uo_01 and enrich: directory to keep fetched pages in and reuse on a re-run, e.g. after a failed archive")
	unprefixed := flag.Bool("unprefixed", false, "static JSON filenames carry no season-id prefix (legacy deployments only)")
	skipDivisions := flag.String("skip-divisions", "", "uo_01 archive: comma-separated division names to leave out entirely, e.g. divisions scored without scoresheets")
	flag.Parse()

	v, ok := liveversion.Get(*versionKey)
	if !ok {
		return fmt.Errorf("-version is required, one of: %s", strings.Join(liveversion.Keys(), ", "))
	}

	// enrich defaults its source from the archive, not the version
	enrichOpts := uoarchive.EnrichOptions{Host: *host, BasePath: *basePath, SeasonID: *seasonID, CacheDir: *cacheDir}
	if *basePath == "" {
		*basePath = v.DefaultBasePath
	}

	opts := livearchive.ArchiveOptions{SeasonID: *seasonID, Unprefixed: *unprefixed, CacheDir: *cacheDir}
	if *skipDivisions != "" {
		if *versionKey != "uo_01" {
			return fmt.Errorf("-skip-divisions is only for -version uo_01")
		}
		opts.SkipDivisions = strings.Split(*skipDivisions, ",")
	}

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

	case "enrich":
		if v.Enrich == nil {
			return fmt.Errorf("-version %s has nothing to enrich, its archives already come from UltiOrganizer", v.Key)
		}
		if err := requireExistingDB(*dbPath); err != nil {
			return err
		}
		if err := v.Enrich(ctx, *dbPath, enrichOpts); err != nil {
			return fmt.Errorf("enrich: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("-mode must be archive, publish, all, or enrich (got %q)", *mode)
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
		return fmt.Errorf("-db is required for -mode publish and enrich")
	}
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist; archive it first with -mode archive", dbPath)
		}
		return fmt.Errorf("check %s: %w", dbPath, err)
	}
	return nil
}
