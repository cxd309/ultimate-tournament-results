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

// ReferenceIDs maps external ids from the reference endpoint to the internal ids
// InsertDivision/InsertCountry assigned, so later imports (teams, games, ...) can
// resolve their foreign keys without re-querying the store.
type ReferenceIDs struct {
	DivisionIDBySeriesID map[int64]int64
	CountryIDByCountryID map[int64]int64
	PoolIDByPoolID       map[int64]int64
}

// ImportReferenceData writes divisions, pools and countries from the reference endpoint
//
// Divisions are written first so pools can resolve their division's internal id from the
// external series_id.
func ImportReferenceData(ctx context.Context, s *store.Store, ref *ReferenceResponse) (ReferenceIDs, error) {
	ids := ReferenceIDs{
		DivisionIDBySeriesID: make(map[int64]int64, len(ref.Series)),
		CountryIDByCountryID: make(map[int64]int64, len(ref.Countries)),
		PoolIDByPoolID:       make(map[int64]int64, len(ref.Pools)),
	}

	for _, series := range ref.Series {
		division, err := s.InsertDivision(ctx, store.Division{
			SeriesID: series.SeriesID,
			Name:     series.Name,
			Ordering: series.Ordering,
		})
		if err != nil {
			return ReferenceIDs{}, fmt.Errorf("insert division %d: %w", series.SeriesID, err)
		}
		ids.DivisionIDBySeriesID[series.SeriesID] = division.ID
	}

	for _, pool := range ref.Pools {
		divisionID, ok := ids.DivisionIDBySeriesID[pool.SeriesID]
		if !ok {
			return ReferenceIDs{}, fmt.Errorf("insert pool %d: division %d not found in this reference response", pool.PoolID, pool.SeriesID)
		}
		row, err := s.InsertPool(ctx, store.Pool{
			PoolID:     pool.PoolID,
			DivisionID: divisionID,
			Name:       pool.PoolName,
			Ordering:   pool.Ordering,
			PoolType:   pool.Type,
		})
		if err != nil {
			return ReferenceIDs{}, fmt.Errorf("insert pool %d: %w", pool.PoolID, err)
		}
		ids.PoolIDByPoolID[pool.PoolID] = row.ID
	}

	for _, country := range ref.Countries {
		row, err := s.InsertCountry(ctx, store.Country{
			CountryExtID: country.CountryID,
			Name:         country.Name,
			Abbreviation: country.Abbreviation,
			FlagFile:     country.FlagFile,
		})
		if err != nil {
			return ReferenceIDs{}, fmt.Errorf("insert country %d: %w", country.CountryID, err)
		}
		ids.CountryIDByCountryID[country.CountryID] = row.ID
	}

	return ids, nil
}

// ImportTeams writes a teams row per team then that team's player roster.
//
// A team row merges three sources:
// reference endpoint: identity (abbreviation, club)
// teams endpoint: stats
// team detail endpoint: pool assignment
func ImportTeams(ctx context.Context, s *store.Store, ids ReferenceIDs, ref *ReferenceResponse, teams *TeamsResponse, detailByTeamID map[int64]*TeamDetailResponse) error {
	identityByTeamID := make(map[int64]ReferenceTeam, len(ref.Teams))
	for _, rt := range ref.Teams {
		identityByTeamID[rt.TeamID] = rt
	}

	for _, stats := range teams.Teams {
		identity := identityByTeamID[stats.TeamID] // zero value if absent: "", ""
		detail := detailByTeamID[stats.TeamID]     // nil if absent

		divisionID, ok := ids.DivisionIDBySeriesID[stats.Series]
		if !ok {
			return fmt.Errorf("insert team %d: division %d not found in this reference response", stats.TeamID, stats.Series)
		}
		countryID, ok := ids.CountryIDByCountryID[stats.Country]
		if !ok {
			return fmt.Errorf("insert team %d: country %d not found in this reference response", stats.TeamID, stats.Country)
		}

		var poolID *int64
		if detail != nil && detail.Pool != 0 {
			id, ok := ids.PoolIDByPoolID[detail.Pool]
			if !ok {
				return fmt.Errorf("insert team %d: pool %d not found in this reference response", stats.TeamID, detail.Pool)
			}
			poolID = &id
		}

		var spiritTotal *int64
		if stats.Spirit.Valid {
			v := int64(stats.Spirit.Value)
			spiritTotal = &v
		}
		var spiritAvg *float64
		if stats.SpiritAvg.Valid {
			v := stats.SpiritAvg.Value
			spiritAvg = &v
		}

		team, err := s.InsertTeam(ctx, store.Team{
			TeamID:                  stats.TeamID,
			DivisionID:              divisionID,
			PoolID:                  poolID,
			CountryID:               countryID,
			Name:                    stats.Name,
			Abbreviation:            identity.Abbreviation,
			Club:                    identity.Club,
			Seed:                    stats.Seed,
			GamesPlayed:             stats.Games,
			Wins:                    stats.Wins,
			Losses:                  stats.Losses,
			PointsFor:               stats.For,
			PointsAgainst:           stats.Against,
			SpiritTotal:             spiritTotal,
			SpiritAvg:               spiritAvg,
			FinalStanding:           stats.FinalStanding,
			FinalStandingCalculated: stats.FinalStandingCalculated,
		})
		if err != nil {
			return fmt.Errorf("insert team %d: %w", stats.TeamID, err)
		}

		if detail != nil {
			if err := importPlayers(ctx, s, team.ID, detail.Players); err != nil {
				return fmt.Errorf("insert players for team %d: %w", stats.TeamID, err)
			}
		}
	}

	return nil
}

// importPlayers writes a player row per squad member from a team's roster.
func importPlayers(ctx context.Context, s *store.Store, teamID int64, players []PlayerStats) error {
	for _, p := range players {
		if _, err := s.InsertPlayer(ctx, store.Player{
			PlayerID:    p.PlayerID,
			TeamID:      teamID,
			FirstName:   p.FirstName,
			LastName:    p.LastName,
			JerseyNum:   p.Num,
			GamesPlayed: p.Games,
			Goals:       p.Done,
			Assists:     p.Fedin,
			Callahans:   p.Callahan,
		}); err != nil {
			return fmt.Errorf("insert player %d: %w", p.PlayerID, err)
		}
	}
	return nil
}
