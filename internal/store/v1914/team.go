package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Team is the plain-Go-typed form of a single row in the teams table
type Team struct {
	TeamID                  int64
	Name                    string
	Pool                    *int64 // nil until the roster/team-detail slice sets it
	Rank                    *int64
	Valid                   int64
	Series                  *int64
	Country                 *int64
	Abbreviation            string
	FinalStanding           *int64
	FinalStandingCalculated *int64
}

func (s *Store) InsertTeam(ctx context.Context, t Team) error {
	return s.q.InsertTeam(ctx, dbgen.InsertTeamParams{
		TeamID:                  t.TeamID,
		Name:                    convert.NullString(t.Name),
		Pool:                    convert.NullInt64(t.Pool),
		Rank:                    convert.NullInt64(t.Rank),
		Valid:                   t.Valid,
		Series:                  convert.NullInt64(t.Series),
		Country:                 convert.NullInt64(t.Country),
		Abbreviation:            convert.NullString(t.Abbreviation),
		FinalStanding:           convert.NullInt64(t.FinalStanding),
		FinalStandingCalculated: convert.NullInt64(t.FinalStandingCalculated),
	})
}
