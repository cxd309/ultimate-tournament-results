package uoarchive

import (
	"context"
	"database/sql"
	"fmt"

	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

// write inserts every resolved row inside tx, in FK-dependency order
// shared tables go through store/v01_09_14, except locations and reservations whose
// fuller rows go through store/uo_01, as do the uo_01-only tables
func write(ctx context.Context, tx *sql.Tx, r *rows) error {
	s, uo := store.New(tx), uostore.New(tx)

	if err := s.InsertTournament(ctx, r.tournament); err != nil {
		return fmt.Errorf("tournament: %w", err)
	}
	if err := uo.InsertServerSettings(ctx, r.serverSettings); err != nil {
		return fmt.Errorf("server settings: %w", err)
	}
	steps := []step{
		{"divisions", each(r.divisions, func(v store.Division) error { return s.InsertDivision(ctx, v) })},
		{"countries", each(r.countries, func(v store.Country) error { return s.InsertCountry(ctx, v) })},
		{"locations", each(r.locations, func(v uostore.Location) error { return uo.InsertLocation(ctx, v) })},
		{"reservations", each(r.reservations, func(v uostore.Reservation) error { return uo.InsertReservation(ctx, v) })},
		{"pools", each(r.pools, func(v store.Pool) error { return s.InsertPool(ctx, v) })},
		{"teams", each(r.teams, func(v store.Team) error { return s.InsertTeam(ctx, v) })},
		{"players", each(r.players, func(v store.Player) error { return s.InsertPlayer(ctx, v) })},
		{"scheduling names", each(r.schedulingNames, func(v store.SchedulingName) error { return s.InsertSchedulingName(ctx, v) })},
		{"games", each(r.games, func(v store.Game) error { return s.InsertGame(ctx, v) })},
		{"goals", each(r.goals, func(v store.Goal) error { return s.InsertGoal(ctx, v) })},
		{"spirit scores", each(r.spiritScores, func(v store.SpiritScore) error { return s.InsertSpiritScore(ctx, v) })},
	}
	if err := runSteps(steps); err != nil {
		return err
	}
	return writeExtras(ctx, uo, r)
}

// writeExtras inserts the uo_01-only tables, the ones a Live! archive has no place for
func writeExtras(ctx context.Context, uo *uostore.Store, r *rows) error {
	return runSteps([]step{
		{"played", each(r.played, func(v uostore.Played) error { return uo.InsertPlayed(ctx, v) })},
		{"game events", each(r.gameEvents, func(v uostore.GameEvent) error { return uo.InsertGameEvent(ctx, v) })},
		{"team pools", each(r.teamPools, func(v uostore.TeamPool) error { return uo.InsertTeamPool(ctx, v) })},
		{"game pools", each(r.gamePools, func(v uostore.GamePool) error { return uo.InsertGamePool(ctx, v) })},
		{"move teams", each(r.moveTeams, func(v uostore.MoveTeam) error { return uo.InsertMoveTeam(ctx, v) })},
		{"clubs", each(r.clubs, func(v uostore.Club) error { return uo.InsertClub(ctx, v) })},
		{"team clubs", each(r.teamClubs, func(v uostore.TeamClub) error { return uo.InsertTeamClub(ctx, v) })},
		{"team profiles", each(r.teamProfiles, func(v uostore.TeamProfile) error { return uo.InsertTeamProfile(ctx, v) })},
		{"urls", each(r.urls, func(v uostore.URL) error { return uo.InsertURL(ctx, v) })},
	})
}

type step struct {
	name string
	run  func() error
}

func runSteps(steps []step) error {
	for _, step := range steps {
		if err := step.run(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	return nil
}

// each returns a step inserting every row with insert, stopping at the first error
func each[T any](rows []T, insert func(T) error) func() error {
	return func() error {
		for _, row := range rows {
			if err := insert(row); err != nil {
				return fmt.Errorf("%+v: %w", row, err)
			}
		}
		return nil
	}
}
