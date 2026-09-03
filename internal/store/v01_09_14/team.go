package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_14"
)

// Team is the plain-Go-typed form of a single row in the teams table
// Valid is IntBool, sourced from team-detail when known, false (not-yet-known) otherwise
type Team struct {
	TeamID                  int64
	Name                    string
	Pool                    *int64 // nil until the roster/team-detail slice sets it
	Rank                    *int64
	Valid                   convert.IntBool
	Series                  *int64
	Country                 *int64
	Abbreviation            string
	FinalStanding           *int64
	FinalStandingCalculated *int64
	ClubName                string // bare name only; there's no club id to join on, see schema
}

func (s *Store) InsertTeam(ctx context.Context, t Team) error {
	return s.q.InsertTeam(ctx, dbgen.InsertTeamParams{
		TeamID:                  t.TeamID,
		Name:                    convert.NullString(t.Name),
		Pool:                    convert.NullInt64(t.Pool),
		Rank:                    convert.NullInt64(t.Rank),
		Valid:                   t.Valid.Int64(),
		Series:                  convert.NullInt64(t.Series),
		Country:                 convert.NullInt64(t.Country),
		Abbreviation:            convert.NullString(t.Abbreviation),
		FinalStanding:           convert.NullInt64(t.FinalStanding),
		FinalStandingCalculated: convert.NullInt64(t.FinalStandingCalculated),
		ClubName:                convert.NullString(t.ClubName),
	})
}

func (s *Store) ListTeams(ctx context.Context) ([]Team, error) {
	rows, err := s.q.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	teams := make([]Team, len(rows))
	for i, row := range rows {
		teams[i] = Team{
			TeamID:                  row.TeamID,
			Name:                    convert.String(row.Name),
			Pool:                    convert.Int64(row.Pool),
			Rank:                    convert.Int64(row.Rank),
			Valid:                   convert.IntBoolFromInt64(row.Valid),
			Series:                  convert.Int64(row.Series),
			Country:                 convert.Int64(row.Country),
			Abbreviation:            convert.String(row.Abbreviation),
			FinalStanding:           convert.Int64(row.FinalStanding),
			FinalStandingCalculated: convert.Int64(row.FinalStandingCalculated),
			ClubName:                convert.String(row.ClubName),
		}
	}
	return teams, nil
}
