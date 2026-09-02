package v1914

import (
	"context"
	"fmt"
	"time"

	store "github.com/cxd309/ultimate-tournament-results/internal/store/v1914"
)

// Import writes an entire Snapshot (see Gather) to the store, in FK-dependency order:
// divisions before pools, locations before reservations, everything before teams. This
// is the package's only exported write entrypoint -- the per-table functions below are
// internal building blocks, not meant to be called individually from outside.
func Import(ctx context.Context, s *store.Store, host, basePath string, snap *Snapshot) error {
	if err := importTournament(ctx, s, host, basePath, snap.Heartbeat, snap.Reference); err != nil {
		return fmt.Errorf("import tournament: %w", err)
	}
	if err := importDivisions(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import divisions: %w", err)
	}
	if err := importCountries(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import countries: %w", err)
	}
	if err := importLocations(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import locations: %w", err)
	}
	if err := importReservations(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import reservations: %w", err)
	}
	if err := importPools(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import pools: %w", err)
	}
	if err := importTeams(ctx, s, snap.Reference, snap.TeamDetailByID); err != nil {
		return fmt.Errorf("import teams: %w", err)
	}
	return nil
}

// importTournament write the single row in the tournament table
//
// sourced from:
// heartbeat endpoint (host, base path, app version)
// reference endpoint season[]
//
// season.name is the reliable event name
// heartbeat's config.TOURNAMENT_NAME "may be an empty string even on a live event."
func importTournament(ctx context.Context, s *store.Store, host, basePath string, hb *HeartbeatResponse, ref *ReferenceResponse) error {
	return s.InsertTournament(ctx, store.Tournament{
		SeasonID:   hb.Config.LiveSeasonID,
		Name:       ref.Season.Name,
		StartTime:  ref.Season.StartTime,
		EndTime:    ref.Season.EndTime,
		Timezone:   ref.Season.Timezone,
		Host:       host,
		BasePath:   basePath,
		AppVersion: hb.AppVersion,
		ArchivedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

// importDivisions write the divisions table
//
// sourced from:
// reference endpoint series[]
func importDivisions(ctx context.Context, s *store.Store, ref *ReferenceResponse) error {
	for _, series := range ref.Series {
		if err := s.InsertDivision(ctx, store.Division{
			SeriesID: series.SeriesID,
			Name:     series.Name,
			Ordering: series.Ordering,
		}); err != nil {
			return fmt.Errorf("insert division %d: %w", series.SeriesID, err)
		}
	}
	return nil
}

// importCountries write the countries table
//
// sourced from:
// reference endpoint countries[]
func importCountries(ctx context.Context, s *store.Store, ref *ReferenceResponse) error {
	for _, country := range ref.Countries {
		if err := s.InsertCountry(ctx, store.Country{
			CountryID:    country.CountryID,
			Name:         country.Name,
			Abbreviation: country.Abbreviation,
			FlagFile:     country.FlagFile,
		}); err != nil {
			return fmt.Errorf("insert country %d: %w", country.CountryID, err)
		}
	}
	return nil
}

// importLocations write the locations table
// Deduped as there may be multiple reservations per location
// Must be run before importReservations
//
// sourced from:
// reference endpoint reservations[]
func importLocations(ctx context.Context, s *store.Store, ref *ReferenceResponse) error {
	seen := make(map[int64]bool, len(ref.Reservations))
	for _, res := range ref.Reservations {
		if seen[res.Location] {
			continue
		}
		if err := s.InsertLocation(ctx, store.Location{
			ID:   res.Location,
			Name: res.LocationName,
		}); err != nil {
			return fmt.Errorf("insert location %d: %w", res.Location, err)
		}
		seen[res.Location] = true
	}
	return nil
}

// importReservations write the reservations table
// Must be run after importLocations
//
// sourced from:
// reference endpoint reservations[]
func importReservations(ctx context.Context, s *store.Store, ref *ReferenceResponse) error {
	for _, res := range ref.Reservations {
		if err := s.InsertReservation(ctx, store.Reservation{
			ID:               res.ID,
			Location:         res.Location,
			FieldName:        res.FieldName,
			ReservationGroup: res.ReservationGroup,
		}); err != nil {
			return fmt.Errorf("insert reservation %d: %w", res.ID, err)
		}
	}
	return nil
}

// importPools write the pools table
// Must be run after importDivisions
//
// sourced from:
// reference endpoint pools[]
func importPools(ctx context.Context, s *store.Store, ref *ReferenceResponse) error {
	for _, pool := range ref.Pools {
		seriesID := pool.SeriesID
		if err := s.InsertPool(ctx, store.Pool{
			PoolID:         pool.PoolID,
			Name:           pool.PoolName,
			Ordering:       pool.Ordering,
			Visible:        pool.Visible.Int64(),
			Continuingpool: pool.Continuing.Int64(),
			Placementpool:  pool.Placementpool.Int64(),
			Played:         pool.Played.Int64(),
			Series:         &seriesID,
			Type:           pool.Type,
		}); err != nil {
			return fmt.Errorf("insert pool %d: %w", pool.PoolID, err)
		}
	}
	return nil
}

// importTeams write the teams table
//
// sourced from:
// reference endpoint teams[]
// teams endpoint (a request per team id)
func importTeams(ctx context.Context, s *store.Store, ref *ReferenceResponse, detailByTeamID map[int64]*TeamDetailResponse) error {
	for _, team := range ref.Teams {
		detail := detailByTeamID[team.TeamID] // nil if absent

		var pool *int64
		if detail != nil && detail.Pool != 0 {
			pool = &detail.Pool
		}
		var valid int64
		if detail != nil {
			valid = detail.Valid.Int64()
		}

		series := team.Series
		country := team.Country
		rank := team.Rank
		finalStanding := team.FinalStanding
		finalStandingCalculated := team.FinalStandingCalculated

		if err := s.InsertTeam(ctx, store.Team{
			TeamID:                  team.TeamID,
			Name:                    team.Name,
			Pool:                    pool,
			Rank:                    &rank,
			Valid:                   valid,
			Series:                  &series,
			Country:                 &country,
			Abbreviation:            team.Abbreviation,
			FinalStanding:           &finalStanding,
			FinalStandingCalculated: &finalStandingCalculated,
		}); err != nil {
			return fmt.Errorf("insert team %d: %w", team.TeamID, err)
		}

		if detail != nil {
			if err := importPlayers(ctx, s, team.TeamID, detail.Players); err != nil {
				return fmt.Errorf("insert players for team %d: %w", team.TeamID, err)
			}
		}
	}

	return nil
}

// importPlayers write a row in the players table
//
// sourced from:
// teams endpoint
func importPlayers(ctx context.Context, s *store.Store, teamID int64, players []PlayerStats) error {
	for _, p := range players {
		if err := s.InsertPlayer(ctx, store.Player{
			PlayerID:    p.PlayerID,
			FirstName:   p.FirstName,
			LastName:    p.LastName,
			Team:        teamID,
			Num:         p.Num,
			GamesPlayed: p.Games,
		}); err != nil {
			return fmt.Errorf("insert player %d: %w", p.PlayerID, err)
		}
	}
	return nil
}