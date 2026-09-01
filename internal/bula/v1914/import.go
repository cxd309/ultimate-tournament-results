package v1914

import (
	"context"
	"time"

	store "github.com/cxd309/ultimate-tournament-results/internal/store/v1914"
)

// ImportHeartbeat write the single row in tournament table from fetched heartbeat
//
// event_name is taken from config.TOURNAMENT_NAME
// spec notes "may be an empty string even on a live event;
// prefer season.name from the reference endpoint"
// once the reference-endpoint slice is added, main's fetch order changes
// so ImportHeartbeat is only called after reference is in hand,
// and this can pass the real event name through instead.
func ImportHeartbeat(ctx context.Context, s *store.Store, host, basePath string, hb *Heartbeat) error {
	return s.InsertTournament(ctx, store.Tournament{
		EventName:  hb.Config.TournamentName,
		Host:       host,
		SeasonID:   hb.Config.LiveSeasonID,
		BasePath:   basePath,
		AppVersion: hb.AppVersion,
		ArchivedAt: time.Now().UTC().Format(time.RFC3339),
	})
}
