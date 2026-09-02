package livearchive

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	liveclient "github.com/cxd309/ultimate-tournament-results/internal/liveclient/v03_00_06"
	"github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v03_00_06"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v03_00_06"
)

// Import writes an entire snapshot (see liveclient.Client.Gather) to the store, in FK-dependency order:
// divisions before pools, locations before reservations, everything before teams. This
// is the package's only exported write entrypoint -- the per-table functions below are
// internal building blocks, not meant to be called individually from outside.
func Import(ctx context.Context, s *store.Store, host, basePath string, snap *liveclient.Snapshot) error {
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
	if err := importPools(ctx, s, snap.Reference, poolInfoByPoolID(snap.GameDetailByID)); err != nil {
		return fmt.Errorf("import pools: %w", err)
	}
	if err := importTeams(ctx, s, snap.Reference, snap.TeamDetailByID); err != nil {
		return fmt.Errorf("import teams: %w", err)
	}
	if err := importPoolPlacements(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import pool placements: %w", err)
	}
	if err := importGames(ctx, s, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import games: %w", err)
	}
	if err := importGamePools(ctx, s, snap.GameDetailByID, snap.GamePoolsByGameID); err != nil {
		return fmt.Errorf("import game pools: %w", err)
	}
	if err := importGoals(ctx, s, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import goals: %w", err)
	}
	if err := importSpiritCategories(ctx, s, snap.Reference); err != nil {
		return fmt.Errorf("import spirit categories: %w", err)
	}
	if err := importSpiritScores(ctx, s, snap.Reference, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import spirit scores: %w", err)
	}
	if err := importSpiritComments(ctx, s, snap.GameDetailByID); err != nil {
		return fmt.Errorf("import spirit comments: %w", err)
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
func importTournament(ctx context.Context, s *store.Store, host, basePath string, hb *livedatamodel.HeartbeatResponse, ref *livedatamodel.ReferenceResponse) error {
	return s.InsertTournament(ctx, store.Tournament{
		SeasonID:   hb.Config.LiveSeasonID,
		Name:       ref.Season.Name,
		StartTime:  ref.Season.StartTime,
		EndTime:    ref.Season.EndTime,
		Timezone:   ref.Season.Timezone,
		Spiritmode: ref.Season.SpiritMode,
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
func importDivisions(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse) error {
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
func importCountries(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse) error {
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
//
// A nil or 0 Location means "the event uses a single unnamed site" -- there's no real
// location to insert, so those reservations are skipped here.
func importLocations(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse) error {
	seen := make(map[int64]bool, len(ref.Reservations))
	for _, res := range ref.Reservations {
		locationID := zeroToNil(res.Location)
		if locationID == nil || seen[*locationID] {
			continue
		}
		if err := s.InsertLocation(ctx, store.Location{
			ID:   *locationID,
			Name: res.LocationName,
		}); err != nil {
			return fmt.Errorf("insert location %d: %w", *locationID, err)
		}
		seen[*locationID] = true
	}
	return nil
}

// importReservations write the reservations table
// Must be run after importLocations
//
// sourced from:
// reference endpoint reservations[]
func importReservations(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse) error {
	for _, res := range ref.Reservations {
		if err := s.InsertReservation(ctx, store.Reservation{
			ID:               res.ID,
			Location:         zeroToNil(res.Location),
			FieldName:        res.FieldName,
			ReservationGroup: res.ReservationGroup,
		}); err != nil {
			return fmt.Errorf("insert reservation %d: %w", res.ID, err)
		}
	}
	return nil
}

// poolInfoByPoolID indexes every game's poolinfo by pool_id, for importPools to pull
// drawsallowed/playoff_template from -- the reference endpoint's own Pool objects don't
// carry those two fields. A pool with no games in it never appears here.
func poolInfoByPoolID(detailByGameID map[int64]*livedatamodel.GameDetailResponse) map[int64]livedatamodel.PoolInfo {
	byPoolID := make(map[int64]livedatamodel.PoolInfo, len(detailByGameID))
	for _, detail := range detailByGameID {
		byPoolID[detail.PoolInfo.PoolID] = detail.PoolInfo
	}
	return byPoolID
}

// importPools write the pools table
// Must be run after importDivisions
//
// sourced from:
// reference endpoint pools[]
// game detail endpoint poolinfo (drawsallowed/playoff_template only, via poolInfoByID)
func importPools(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse, poolInfoByID map[int64]livedatamodel.PoolInfo) error {
	for _, pool := range ref.Pools {
		seriesID := pool.SeriesID
		info, hasInfo := poolInfoByID[pool.PoolID]

		var drawsallowed *int64
		var playoffTemplate string
		if hasInfo {
			d := info.Drawsallowed.Int64()
			drawsallowed = &d
			if info.PlayoffTemplate != nil {
				playoffTemplate = *info.PlayoffTemplate
			}
		}

		if err := s.InsertPool(ctx, store.Pool{
			PoolID:          pool.PoolID,
			Name:            pool.PoolName,
			Ordering:        pool.Ordering,
			Visible:         pool.Visible,
			Continuingpool:  pool.Continuing,
			Placementpool:   pool.Placementpool,
			Played:          pool.Played,
			Series:          &seriesID,
			Type:            pool.Type,
			Drawsallowed:    drawsallowed,
			PlayoffTemplate: playoffTemplate,
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
func importTeams(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse, detailByTeamID map[int64]*livedatamodel.TeamDetailResponse) error {
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
func importPlayers(ctx context.Context, s *store.Store, teamID int64, players []livedatamodel.PlayerStats) error {
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

// importPoolPlacements write the pool_placements table
// Must be run after importPools and importTeams
//
// sourced from:
// reference endpoint pool_placements[]
func importPoolPlacements(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse) error {
	for _, placement := range ref.PoolPlacements {
		if err := s.InsertPoolPlacement(ctx, store.PoolPlacement{
			PoolID:    placement.PoolID,
			TeamID:    placement.TeamID,
			Placement: placement.Placement,
		}); err != nil {
			return fmt.Errorf("insert pool placement %d/%d: %w", placement.PoolID, placement.TeamID, err)
		}
	}
	return nil
}

// importGames writes the games table
// Must be run after importPools, importTeams and importReservations
//
// sourced from:
// games endpoint (game ids to enumerate)
// game detail endpoint game_result -- the only endpoint where nothing is falsy-stripped
//
// No pool: not a games column on this line, see game_pools. No homesotg/visitorsotg:
// derivable from spirit_scores, see the games table's own comment.
func importGames(ctx context.Context, s *store.Store, detailByGameID map[int64]*livedatamodel.GameDetailResponse) error {
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
			Valid:                 gr.Valid,
			Halftime:              gr.Halftime,
			Official:              gr.Official,
			Respteam:              gr.Respteam,
			Resppers:              gr.Resppers,
			Isongoing:             gr.Isongoing,
			SchedulingNameHome:    gr.SchedulingNameHome,
			SchedulingNameVisitor: gr.SchedulingNameVisitor,
			Name:                  name,
			Timeslot:              gr.Timeslot,
			Homedefenses:          gr.Homedefenses,
			Visitordefenses:       gr.Visitordefenses,
			Islive:                gr.Islive,
			Liveurl:               gr.Liveurl,
			Hasstarted:            gr.Hasstarted,
			ShowSpirit:            gr.ShowSpirit,
			TimerStart:            gr.TimerStart,
			TimerPauseStart:       gr.TimerPauseStart,
			TimerPausedDuration:   gr.TimerPausedDuration,
			Forfeit:               gr.Forfeit,
		}); err != nil {
			return fmt.Errorf("insert game %d: %w", gr.GameID, err)
		}
	}
	return nil
}

// importGamePools writes the game_pools table
// Must be run after importGames and importPools
//
// sourced from:
// games endpoint (every pool a game belongs to, via GamePoolsByGameID)
// game detail endpoint game_result.pool (the owning pool, to set timetable)
func importGamePools(ctx context.Context, s *store.Store, detailByGameID map[int64]*livedatamodel.GameDetailResponse, poolsByGameID map[int64][]int64) error {
	for gameID, detail := range detailByGameID {
		owning := detail.GameResult.Pool
		for _, poolID := range poolsByGameID[gameID] {
			timetable := convert.IntBool(owning != nil && *owning == poolID)
			if err := s.InsertGamePool(ctx, store.GamePool{
				GameID:    gameID,
				PoolID:    poolID,
				Timetable: timetable,
			}); err != nil {
				return fmt.Errorf("insert game pool %d/%d: %w", gameID, poolID, err)
			}
		}
	}
	return nil
}

// importGoals writes the goals table
// Must be run after importGames and importPlayers
//
// sourced from:
// game detail endpoint goals[]
func importGoals(ctx context.Context, s *store.Store, detailByGameID map[int64]*livedatamodel.GameDetailResponse) error {
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
				Timestamp:    g.Timestamp,
			}); err != nil {
				return fmt.Errorf("insert goal %d/%d: %w", gameID, g.Num, err)
			}
		}
	}
	return nil
}

// importSpiritCategories writes the spirit_categories table
//
// sourced from:
// reference endpoint season.spiritCategories[]
func importSpiritCategories(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse) error {
	for _, cat := range ref.Season.SpiritCategories {
		if err := s.InsertSpiritCategory(ctx, store.SpiritCategory{
			CategoryID:    cat.CategoryID,
			Mode:          modeOrZero(ref.Season.SpiritMode),
			CategoryGroup: cat.Group,
			Ordering:      cat.Index,
			Min:           cat.Min,
			Max:           cat.Max,
			Factor:        cat.Factor,
			Label:         cat.Label,
		}); err != nil {
			return fmt.Errorf("insert spirit category %d: %w", cat.CategoryID, err)
		}
	}
	return nil
}

// spiritSide pairs one side of GameSpiritStats with the team id it belongs to,
// resolved from game_result
// livedatamodel.GameSpiritScore itself no longer carries team_id on this line, the
// parent object's key (hometeam/visitorteam) is the only thing that says which team a
// score is for
type spiritSide struct {
	score  *livedatamodel.GameSpiritScore
	teamID *int64
}

// spiritSides pairs both sides of one game's SpiritStats with their resolved team ids,
// skipping a side with no score or an unresolved team
// (an unresolved bracket slot has no team id to attach a spirit score/comment to)
func spiritSides(detail *livedatamodel.GameDetailResponse) []spiritSide {
	all := []spiritSide{
		{detail.SpiritStats.Hometeam, detail.GameResult.Hometeam},
		{detail.SpiritStats.Visitorteam, detail.GameResult.Visitorteam},
	}
	sides := make([]spiritSide, 0, len(all))
	for _, side := range all {
		if side.score != nil && side.teamID != nil {
			sides = append(sides, side)
		}
	}
	return sides
}

// importSpiritScores writes the spirit_scores table
// Must be run after importGames, importTeams and importSpiritCategories
//
// sourced from:
// game detail endpoint spiritstats
// reference endpoint season.spiritCategories[] (to resolve a "catN" key to a category_id)
func importSpiritScores(ctx context.Context, s *store.Store, ref *livedatamodel.ReferenceResponse, detailByGameID map[int64]*livedatamodel.GameDetailResponse) error {
	categoryIDByKey := make(map[string]int64, len(ref.Season.SpiritCategories))
	for _, cat := range ref.Season.SpiritCategories {
		categoryIDByKey[cat.Key] = cat.CategoryID
	}

	for gameID, detail := range detailByGameID {
		if detail.SpiritStats == nil {
			continue // event doesn't publish spirit points
		}
		for _, side := range spiritSides(detail) {
			teamID := *side.teamID
			for key, value := range side.score.Categories {
				categoryID, ok := categoryIDByKey[key]
				if !ok {
					return fmt.Errorf("game %d: spirit category %q not in season.spiritCategories", gameID, key)
				}
				if err := s.InsertSpiritScore(ctx, store.SpiritScore{
					GameID:     gameID,
					TeamID:     teamID,
					CategoryID: categoryID,
					Value:      value,
				}); err != nil {
					return fmt.Errorf("insert spirit score game %d team %d category %s: %w", gameID, teamID, key, err)
				}
			}
		}
	}
	return nil
}

// importSpiritComments writes the spirit_comments table
// Must be run after importGames and importTeams
//
// sourced from:
// game detail endpoint spiritstats
func importSpiritComments(ctx context.Context, s *store.Store, detailByGameID map[int64]*livedatamodel.GameDetailResponse) error {
	for gameID, detail := range detailByGameID {
		if detail.SpiritStats == nil {
			continue
		}
		for _, side := range spiritSides(detail) {
			if side.score.Comments == nil {
				continue
			}
			teamID := *side.teamID
			if err := s.InsertSpiritComment(ctx, store.SpiritComment{
				GameID:  gameID,
				TeamID:  teamID,
				Comment: *side.score.Comments,
			}); err != nil {
				return fmt.Errorf("insert spirit comment game %d team %d: %w", gameID, teamID, err)
			}
		}
	}
	return nil
}

// modeOrZero unwraps season.spiritmode, defaulting to 0 when the season has none.
func modeOrZero(mode *int64) int64 {
	if mode == nil {
		return 0
	}
	return *mode
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

// zeroToNil normalizes an API "not recorded" sentinel to nil: goal scorer/assist and
// reservation.location all use non-positive values (0, or -1 for goals) to mean "unset,"
// none of which are ever a real id and all of which would violate an FK.
func zeroToNil(id *int64) *int64 {
	if id == nil || *id <= 0 {
		return nil
	}
	return id
}
