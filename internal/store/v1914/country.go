package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Country is the plain-Go-typed form of a single row in the countries table
type Country struct {
	CountryID    int64
	Name         string
	Abbreviation string
	FlagFile     string
}

func (s *Store) InsertCountry(ctx context.Context, c Country) error {
	return s.q.InsertCountry(ctx, dbgen.InsertCountryParams{
		CountryID:    c.CountryID,
		Name:         c.Name,
		Abbreviation: convert.NullString(c.Abbreviation),
		FlagFile:     convert.NullString(c.FlagFile),
	})
}
