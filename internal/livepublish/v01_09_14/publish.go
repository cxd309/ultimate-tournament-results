// Package livepublish renders a 1.9.14-1.9.17 archive as static JSON files in docs/
package livepublish

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/cxd309/ultimate-tournament-results/internal/liveclient"
	livepublish "github.com/cxd309/ultimate-tournament-results/internal/livepublish"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

// Publish reads back a 1.9.14-1.9.17 archive at dbPath
// writes every endpoint's JSON under outDir/<seasonId>/
// using the live API's own relative filenames
//
// prints its own progress: one file per team and per game, so a big
// tournament writes hundreds of files with nothing else to show for it
func Publish(ctx context.Context, dbPath, outDir string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer db.Close()

	fmt.Printf("loading %s...\n", dbPath)
	data, err := load(ctx, store.New(db))
	if err != nil {
		return fmt.Errorf("load %s: %w", dbPath, err)
	}

	dir := filepath.Join(outDir, data.tournament.SeasonID)
	prefix := data.tournament.SeasonID
	fmt.Printf("publishing %q (%d teams, %d games) to %s\n", data.tournament.Name, len(data.teams), len(data.games), dir)

	if err := livepublish.WriteJSON(dir, "_heartbeat.json", renderHeartbeat(data)); err != nil {
		return err
	}
	if err := livepublish.WriteJSON(dir, prefix+"_reference.json", renderReference(data)); err != nil {
		return err
	}

	fmt.Printf("writing %d team files...\n", len(data.teams))
	for i, team := range data.teams {
		filename := fmt.Sprintf("%s_teams_%d.json", prefix, team.TeamID)
		if err := livepublish.WriteJSON(dir, filename, renderTeamDetail(data, team)); err != nil {
			return err
		}
		liveclient.PrintProgress(i+1, len(data.teams), 10)
	}

	if err := livepublish.WriteJSON(dir, prefix+"_games.json", renderGames(data)); err != nil {
		return err
	}

	fmt.Printf("writing %d game files...\n", len(data.games))
	for i, game := range data.games {
		filename := fmt.Sprintf("%s_games_%d.json", prefix, game.GameID)
		if err := livepublish.WriteJSON(dir, filename, renderGameDetail(data, game)); err != nil {
			return err
		}
		liveclient.PrintProgress(i+1, len(data.games), 25)
	}

	fmt.Printf("published %s to %s\n", dbPath, dir)
	return nil
}

// tournamentData is everything read back from one archive
// plus the lookup indices the renderers need
type tournamentData struct {
	tournament   store.Tournament
	divisions    []store.Division
	countries    []store.Country
	pools        []store.Pool
	teams        []store.Team
	reservations []store.Reservation
	games        []store.Game
	playerCount  int64

	locationByID       map[int64]store.Location
	poolByID           map[int64]store.Pool
	playersByTeam      map[int64][]store.Player
	schedulingNameByID map[int64]store.SchedulingName
	goalsByGame        map[int64][]store.Goal
	spiritScoresByGame map[int64][]store.SpiritScore
}

func load(ctx context.Context, s *store.Store) (*tournamentData, error) {
	tournament, err := s.GetTournament(ctx)
	if err != nil {
		return nil, fmt.Errorf("tournament: %w", err)
	}
	divisions, err := s.ListDivisions(ctx)
	if err != nil {
		return nil, fmt.Errorf("divisions: %w", err)
	}
	countries, err := s.ListCountries(ctx)
	if err != nil {
		return nil, fmt.Errorf("countries: %w", err)
	}
	locations, err := s.ListLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("locations: %w", err)
	}
	reservations, err := s.ListReservations(ctx)
	if err != nil {
		return nil, fmt.Errorf("reservations: %w", err)
	}
	pools, err := s.ListPools(ctx)
	if err != nil {
		return nil, fmt.Errorf("pools: %w", err)
	}
	teams, err := s.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("teams: %w", err)
	}
	players, err := s.ListPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("players: %w", err)
	}
	schedulingNames, err := s.ListSchedulingNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("scheduling names: %w", err)
	}
	games, err := s.ListGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("games: %w", err)
	}
	goals, err := s.ListGoals(ctx)
	if err != nil {
		return nil, fmt.Errorf("goals: %w", err)
	}
	spiritScores, err := s.ListSpiritScores(ctx)
	if err != nil {
		return nil, fmt.Errorf("spirit scores: %w", err)
	}

	data := &tournamentData{
		tournament:   tournament,
		divisions:    divisions,
		countries:    countries,
		pools:        pools,
		teams:        teams,
		reservations: reservations,
		games:        games,
		playerCount:  int64(len(players)),

		locationByID:       make(map[int64]store.Location, len(locations)),
		poolByID:           make(map[int64]store.Pool, len(pools)),
		playersByTeam:      make(map[int64][]store.Player, len(teams)),
		schedulingNameByID: make(map[int64]store.SchedulingName, len(schedulingNames)),
		goalsByGame:        make(map[int64][]store.Goal, len(games)),
		spiritScoresByGame: make(map[int64][]store.SpiritScore, len(games)),
	}
	for _, l := range locations {
		data.locationByID[l.ID] = l
	}
	for _, p := range pools {
		data.poolByID[p.PoolID] = p
	}
	for _, p := range players {
		data.playersByTeam[p.Team] = append(data.playersByTeam[p.Team], p)
	}
	for _, sn := range schedulingNames {
		data.schedulingNameByID[sn.SchedulingID] = sn
	}
	for _, g := range goals {
		data.goalsByGame[g.GameID] = append(data.goalsByGame[g.GameID], g)
	}
	for _, sc := range spiritScores {
		data.spiritScoresByGame[sc.GameID] = append(data.spiritScoresByGame[sc.GameID], sc)
	}

	return data, nil
}
