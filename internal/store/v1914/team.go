package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Team is the plain-Go-typed form of a teams row
// PoolID must be the internal id returned by InsertPool
// not the external pool_id, it's a pointer so nil means not yet known, not "no pool"
// Abbreviation/Club are "" for NULL (optional identity fields)
// SpiritTotal/SpiritAvg are nil for NULL (0 is a meaningul value)
type Team struct {
	ID                      int64
	TeamID                  int64 // external id
	DivisionID              int64 // internal divisions.id
	PoolID                  *int64
	CountryID               int64 // internal countries.id
	Name                    string
	Abbreviation            string
	Club                    string
	Seed                    int64
	GamesPlayed             int64
	Wins                    int64
	Losses                  int64
	PointsFor               int64
	PointsAgainst           int64
	SpiritTotal             *int64
	SpiritAvg               *float64
	FinalStanding           int64
	FinalStandingCalculated int64
}

func (s *Store) InsertTeam(ctx context.Context, t Team) (Team, error) {
	row, err := s.q.InsertTeam(ctx, dbgen.InsertTeamParams{
		TeamID:                  t.TeamID,
		DivisionID:              t.DivisionID,
		PoolID:                  convert.NullInt64(t.PoolID),
		CountryID:               t.CountryID,
		Name:                    t.Name,
		Abbreviation:            convert.NullString(t.Abbreviation),
		Club:                    convert.NullString(t.Club),
		Seed:                    t.Seed,
		GamesPlayed:             t.GamesPlayed,
		Wins:                    t.Wins,
		Losses:                  t.Losses,
		PointsFor:               t.PointsFor,
		PointsAgainst:           t.PointsAgainst,
		SpiritTotal:             convert.NullInt64(t.SpiritTotal),
		SpiritAvg:               convert.NullFloat64(t.SpiritAvg),
		FinalStanding:           t.FinalStanding,
		FinalStandingCalculated: t.FinalStandingCalculated,
	})
	if err != nil {
		return Team{}, err
	}
	return Team{
		ID:                      row.ID,
		TeamID:                  row.TeamID,
		DivisionID:              row.DivisionID,
		PoolID:                  convert.Int64(row.PoolID),
		CountryID:               row.CountryID,
		Name:                    row.Name,
		Abbreviation:            convert.String(row.Abbreviation),
		Club:                    convert.String(row.Club),
		Seed:                    row.Seed,
		GamesPlayed:             row.GamesPlayed,
		Wins:                    row.Wins,
		Losses:                  row.Losses,
		PointsFor:               row.PointsFor,
		PointsAgainst:           row.PointsAgainst,
		SpiritTotal:             convert.Int64(row.SpiritTotal),
		SpiritAvg:               convert.Float64(row.SpiritAvg),
		FinalStanding:           row.FinalStanding,
		FinalStandingCalculated: row.FinalStandingCalculated,
	}, nil
}
