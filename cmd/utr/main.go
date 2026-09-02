// Command utr archives and publishes Live! by BULA tournament results
// it multiplexes across every supported spec version via -version
// see internal/liveversion for the registry of implementations
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/cxd309/ultimate-tournament-results/internal/liveversion"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "utr:", err)
		os.Exit(1)
	}
}

func run() error {
	versionKey := flag.String("version", "", "Live! by BULA spec version to target, one of: "+strings.Join(liveversion.Keys(), ", ")+" (required)")
	mode := flag.String("mode", "all", "what to run: archive, publish, or all")
	host := flag.String("host", "", "deployment host, e.g. wbuc.wfdf.sport (required for archive)")
	basePath := flag.String("base-path", "/live/data/", "static cache base path, e.g. /live/data/")
	slug := flag.String("slug", "", "archive filename slug (default: the season id reported by the heartbeat, lowercased)")
	dbPath := flag.String("db", "", "sqlite archive path: written by archive (default data/<slug>.db; must not exist), read by publish (must exist)")
	outDir := flag.String("out", "docs/", "directory to render published JSON into")
	flag.Parse()

	v, ok := liveversion.Get(*versionKey)
	if !ok {
		return fmt.Errorf("-version is required, one of: %s", strings.Join(liveversion.Keys(), ", "))
	}

	ctx := context.Background()

	switch *mode {
	case "archive":
		if *host == "" {
			return fmt.Errorf("-host is required for -mode archive")
		}
		if _, err := v.Archive(ctx, *host, *basePath, *slug, *dbPath); err != nil {
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
		if *host == "" {
			return fmt.Errorf("-host is required for -mode all")
		}
		summary, err := v.Archive(ctx, *host, *basePath, *slug, *dbPath)
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
