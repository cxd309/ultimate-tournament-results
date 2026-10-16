// Package uoarchive archives one season from a classic UltiOrganizer site (no Live! by
// BULA deployment) into the uo_01 schema: db/v01_09_14/schema.sql, unchanged, plus
// db/uo_01/schema.sql's extra tables. The shared tables are written exactly as a
// 1.9.14-1.9.17 archive's are, so livepublish/v01_09_14 publishes it unchanged
package uoarchive

import (
	"context"
	"database/sql"

	"github.com/cxd309/ultimate-tournament-results/internal/livearchive"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

// Archive fetches one season from the UltiOrganizer install at host/basePath and
// writes it into a fresh sqlite archive
// see livearchive.Run for the write-once/slug/dbPath rules
// opts.SeasonID picks the season, needed when the site hosts more than one
// opts.CacheDir keeps every fetched page so a re-run doesn't fetch them again
func Archive(ctx context.Context, host, basePath, slug, dbPath string, opts livearchive.ArchiveOptions) (livearchive.Summary, error) {
	return livearchive.Run(ctx, livearchive.Deps[*rows]{
		SchemaPaths: []string{"db/v01_09_14/schema.sql", "db/uo_01/schema.sql"},
		Gather: func(ctx context.Context, host, basePath string) (*rows, error) {
			c := uoclient.NewClient(host, basePath)
			c.CacheDir = opts.CacheDir
			snap, err := c.Gather(ctx, opts.SeasonID)
			if err != nil {
				return nil, err
			}
			return resolve(snap, host, basePath, false)
		},
		Summarize: func(r *rows) livearchive.Summary {
			return livearchive.Summary{
				SeasonID:  r.tournament.SeasonID,
				Name:      r.tournament.Name,
				Divisions: len(r.divisions),
				Pools:     len(r.pools),
				Countries: len(r.countries),
				Teams:     len(r.teams),
			}
		},
		Import: func(ctx context.Context, tx *sql.Tx, _, _ string, r *rows) error {
			return write(ctx, tx, r)
		},
		ReadBack: func(ctx context.Context, db *sql.DB) (name, seasonID, startTime string, err error) {
			tournament, err := store.New(db).GetTournament(ctx)
			if err != nil {
				return "", "", "", err
			}
			return tournament.Name, tournament.SeasonID, tournament.StartTime, nil
		},
	}, host, basePath, slug, dbPath)
}
