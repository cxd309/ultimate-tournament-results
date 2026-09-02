package livearchive

import (
	"context"
	"database/sql"

	"github.com/cxd309/ultimate-tournament-results/internal/livearchive"
	liveclient "github.com/cxd309/ultimate-tournament-results/internal/liveclient/v01_09_14"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

// Archive fetches one tournament from a live 1.9.14-1.9.16 deployment and writes it
// into a fresh sqlite archive
// see livearchive.Run for the write-once/slug/dbPath rules
func Archive(ctx context.Context, host, basePath, slug, dbPath string) (livearchive.Summary, error) {
	return livearchive.Run(ctx, livearchive.Deps[*liveclient.Snapshot]{
		SchemaPath: "db/v01_09_14/schema.sql",
		Gather: func(ctx context.Context, host, basePath string) (*liveclient.Snapshot, error) {
			return liveclient.NewClient(host, basePath).Gather(ctx)
		},
		Summarize: func(snap *liveclient.Snapshot) livearchive.Summary {
			return livearchive.Summary{
				SeasonID:  snap.Heartbeat.Config.LiveSeasonID,
				Name:      snap.Reference.Season.Name,
				Divisions: len(snap.Reference.Series),
				Pools:     len(snap.Reference.Pools),
				Countries: len(snap.Reference.Countries),
				Teams:     len(snap.Reference.Teams),
			}
		},
		Import: func(ctx context.Context, tx *sql.Tx, host, basePath string, snap *liveclient.Snapshot) error {
			return Import(ctx, store.New(tx), host, basePath, snap)
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
