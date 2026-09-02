package v03_00_06

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v03_00_06"
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
	if err := importGames(ctx, s, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import games: %w", err)
	}
	if err := importGoals(ctx, s, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import goals: %w", err)
	}
	if err := importSpiritScores(ctx, s, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import spirit scores: %w", err)
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
			Visible:        pool.Visible,
			Continuingpool: pool.Continuing,
			Placementpool:  pool.Placementpool,
			Played:         pool.Played,
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
		var valid convert.IntBool
		if detail != nil {
			valid = detail.Valid
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
			ClubName:                team.Club,
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

// importGames writes the games table
// Must be run after importPools, importTeams and importReservations
//
// sourced from:
// games endpoint (game ids to enumerate)
// game detail endpoint game_result
func importGames(ctx context.Context, s *store.Store, detailByGameID map[int64]*GameDetailResponse) error {
	for _, detail := range detailByGameID {
		gr := detail.GameResult
		name, err := parseSchedulingNameID(gr.Name)
		if err != nil {
			return fmt.Errorf("game %d: %w", gr.GameID, err)
		}
		if err := s.InsertGame(ctx, store.Game{
			GameID:                gr.GameID,
			Hometeam:              gr.Hometeam,
			Visitorteam:           gr.Visitorteam,
			Homescore:             gr.Homescore,
			Visitorscore:          gr.Visitorscore,
			Reservation:           gr.Reservation,
			Time:                  gr.Time,
			Pool:                  gr.Pool,
			Valid:                 gr.Valid,
			Halftime:              gr.Halftime,
			Official:              gr.Official,
			Respteam:              gr.Respteam,
			Resppers:              gr.Resppers,
			Homesotg:              gr.Homesotg,
			Visitorsotg:           gr.Visitorsotg,
			Isongoing:             gr.Isongoing,
			SchedulingNameHome:    gr.SchedulingNameHome,
			SchedulingNameVisitor: gr.SchedulingNameVisitor,
			Name:                  name,
			Timeslot:              gr.Timeslot,
			Homedefenses:          gr.Homedefenses,
			Visitordefenses:       gr.Visitordefenses,
			Islive:                gr.Islive,
			Liveurl:               gr.Liveurl,
		}); err != nil {
			return fmt.Errorf("insert game %d: %w", gr.GameID, err)
		}
	}
	return nil
}

// importGoals writes the goals table
// Must be run after importGames and importPlayers
//
// sourced from:
// game detail endpoint goals[]
func importGoals(ctx context.Context, s *store.Store, detailByGameID map[int64]*GameDetailResponse) error {
	for gameID, detail := range detailByGameID {
		for _, g := range detail.Goals {
			if err := s.InsertGoal(ctx, store.Goal{
				GameID:       gameID,
				Num:          g.Num,
				Assist:       zeroToNil(g.Assist),
				Scorer:       zeroToNil(g.Scorer),
				Time:         g.Time,
				Homescore:    &g.Homescore,
				Visitorscore: &g.Visitorscore,
				Ishomegoal:   g.Ishomegoal,
				Iscallahan:   g.Iscallahan,
			}); err != nil {
				return fmt.Errorf("insert goal %d/%d: %w", gameID, g.Num, err)
			}
		}
	}
	return nil
}

// importSpiritScores writes the spirit_scores table
// Must be run after importGames and importTeams
//
// sourced from:
// game detail endpoint spiritstats
func importSpiritScores(ctx context.Context, s *store.Store, detailByGameID map[int64]*GameDetailResponse) error {
	for gameID, detail := range detailByGameID {
		if detail.SpiritStats == nil {
			continue // event doesn't publish spirit points
		}
		for _, score := range []*GameSpiritScore{detail.SpiritStats.Hometeam, detail.SpiritStats.Visitorteam} {
			if score == nil {
				continue
			}
			if err := s.InsertSpiritScore(ctx, store.SpiritScore{
				GameID:   gameID,
				TeamID:   score.TeamID,
				Cat1:     score.Cat1,
				Cat2:     score.Cat2,
				Cat3:     score.Cat3,
				Cat4:     score.Cat4,
				Cat5:     score.Cat5,
				Comments: score.Comments,
			}); err != nil {
				return fmt.Errorf("insert spirit score game %d team %d: %w", gameID, score.TeamID, err)
			}
		}
	}
	return nil
}

// parseSchedulingNameID parses GameResult.Name, sent as a numeric string because `name`
// is exempt from the API's usual number coercion, back into the int64 the games table
// stores it as.
func parseSchedulingNameID(s *string) (*int64, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse scheduling name id %q: %w", *s, err)
	}
	return &id, nil
}

// zeroToNil normalizes goal scorer/assist "not recorded" sentinels to nil, since neither
// is ever a real player_id and both would violate the FK. The spec only documents 0, but
// -1 shows up in practice too -- the same sentinel the API uses for homecaptain/
// awaycaptain -- so anything non-positive is treated as unrecorded.
func zeroToNil(id *int64) *int64 {
	if id == nil || *id <= 0 {
		return nil
	}
	return id
}
