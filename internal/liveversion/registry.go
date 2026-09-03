// Package liveversion is the plain, non-generic registry cmd/utr dispatches through --
// it maps a -version flag value to that version's Archive/Publish entrypoints without
// needing to know anything about the divergent types underneath them
package liveversion

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cxd309/ultimate-tournament-results/internal/livearchive"
	livearchive010914 "github.com/cxd309/ultimate-tournament-results/internal/livearchive/v01_09_14"
	livearchive030006 "github.com/cxd309/ultimate-tournament-results/internal/livearchive/v03_00_06"
	livepublish010914 "github.com/cxd309/ultimate-tournament-results/internal/livepublish/v01_09_14"
	livepublish030006 "github.com/cxd309/ultimate-tournament-results/internal/livepublish/v03_00_06"
)

// Version is one Live! by BULA spec implementation, keyed by its binary-version string
// (e.g. "v01_09_14")
type Version struct {
	Key     string
	Archive func(ctx context.Context, host, basePath, slug, dbPath string) (livearchive.Summary, error)
	Publish func(ctx context.Context, dbPath, outDir string) error
}

// v01_09_14 covers 1.9.14 through 1.9.17
// see the README's compatibility table
var registry = map[string]Version{
	"v01_09_14": {Key: "v01_09_14", Archive: livearchive010914.Archive, Publish: livepublish010914.Publish},
	"v03_00_06": {Key: "v03_00_06", Archive: livearchive030006.Archive, Publish: livepublish030006.Publish},
}

// Get looks up a Version by its key, accepting both the canonical zero-padded form
// ("v03_00_06") and a natural dotted app version ("v3.0.6" or "3.0.6")
func Get(key string) (Version, bool) {
	if v, ok := registry[key]; ok {
		return v, true
	}
	v, ok := registry[normalize(key)]
	return v, ok
}

// normalize converts a dotted app version like "3.0.6" or "v3.0.6" into the canonical
// registry key "v03_00_06": strip a leading v, split on "." or "_", zero-pad each
// segment to two digits, and rejoin with "_". Anything that isn't three numeric
// segments is returned unchanged, so it simply won't match the registry
func normalize(key string) string {
	key = strings.TrimPrefix(strings.TrimPrefix(key, "v"), "V")
	parts := strings.FieldsFunc(key, func(r rune) bool { return r == '.' || r == '_' })
	if len(parts) != 3 {
		return key
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return key
		}
		parts[i] = fmt.Sprintf("%02d", n)
	}
	return "v" + strings.Join(parts, "_")
}

// Keys returns every registered version key, sorted, for building flag help and error
// text
func Keys() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
