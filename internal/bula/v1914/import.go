package v1914

import (
	"context"
	"fmt"
	"time"

	store "github.com/cxd309/ultimate-tournament-results/internal/store/v1914"
)

// ImportTournament writes the single row in the tournament table
//
// sourced from responses including:
// the heartbeat (host, base path, app version)
// the reference season block (name, dates, timezone, status)
//
// season.name is the reliable event name
// heartbeat's config.TOURNAMENT_NAME "may be an empty string even on a live event."
func ImportTournament(ctx context.Context, s *store.Store, host, basePath string, hb *HeartbeatResponse, ref *ReferenceResponse) error {
	return s.InsertTournament(ctx, store.Tournament{
		EventName:  ref.Season.Name,
		Host:       host,
		SeasonID:   hb.Config.LiveSeasonID,
		BasePath:   basePath,
		AppVersion: hb.AppVersion,
		StartDate:  ref.Season.StartTime,
		EndDate:    ref.Season.EndTime,
		Timezone:   ref.Season.Timezone,
		Status:     ref.Season.Status,
		ArchivedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

// ImportReferenceData writes divisions, pools and countries from the reference endpoint
//
// Divisions are written first so pools can resolve their division's internal id from the
// external series_id.
func ImportReferenceData(ctx context.Context, s *store.Store, ref *ReferenceResponse) error {
	divisionIDBySeriesID := make(map[int64]int64, len(ref.Series))
	for _, series := range ref.Series {
		division, err := s.InsertDivision(ctx, store.Division{
			SeriesID: series.SeriesID,
			Name:     series.Name,
			Ordering: series.Ordering,
		})
		if err != nil {
			return fmt.Errorf("insert division %d: %w", series.SeriesID, err)
		}
		divisionIDBySeriesID[series.SeriesID] = division.ID
	}

	for _, pool := range ref.Pools {
		divisionID, ok := divisionIDBySeriesID[pool.SeriesID]
		if !ok {
			return fmt.Errorf("insert pool %d: division %d not found in this reference response", pool.PoolID, pool.SeriesID)
		}
		if _, err := s.InsertPool(ctx, store.Pool{
			PoolID:     pool.PoolID,
			DivisionID: divisionID,
			Name:       pool.PoolName,
			Ordering:   pool.Ordering,
			PoolType:   pool.Type,
		}); err != nil {
			return fmt.Errorf("insert pool %d: %w", pool.PoolID, err)
		}
	}

	for _, country := range ref.Countries {
		if _, err := s.InsertCountry(ctx, store.Country{
			CountryExtID: country.CountryID,
			Name:         country.Name,
			Abbreviation: country.Abbreviation,
			FlagFile:     country.FlagFile,
		}); err != nil {
			return fmt.Errorf("insert country %d: %w", country.CountryID, err)
		}
	}

	return nil
}
