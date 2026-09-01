package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Division is the plain-Go-typed form of a division row
type Division struct {
	ID       int64 // internal id, assigned on insert
	SeriesID int64 // external id
	Name     string
	Ordering string
}

func (s *Store) InsertDivision(ctx context.Context, d Division) (Division, error) {
	row, err := s.q.InsertDivision(ctx, dbgen.InsertDivisionParams{
		SeriesID: d.SeriesID,
		Name:     d.Name,
		Ordering: d.Ordering,
	})
	if err != nil {
		return Division{}, err
	}
	return Division{ID: row.ID, SeriesID: row.SeriesID, Name: row.Name, Ordering: row.Ordering}, nil
}

// Pool is the plain-Go-typed form of a pool row
// DivisionID must be the internal id returned by InsertDivision,
// not the external series_id
type Pool struct {
	ID         int64
	PoolID     int64 // external id
	DivisionID int64 // internal divisions.id
	Name       string
	Ordering   string
	PoolType   int64
}

func (s *Store) InsertPool(ctx context.Context, p Pool) (Pool, error) {
	row, err := s.q.InsertPool(ctx, dbgen.InsertPoolParams{
		PoolID:     p.PoolID,
		DivisionID: p.DivisionID,
		Name:       p.Name,
		Ordering:   p.Ordering,
		PoolType:   p.PoolType,
	})
	if err != nil {
		return Pool{}, err
	}
	return Pool{ID: row.ID, PoolID: row.PoolID, DivisionID: row.DivisionID, Name: row.Name, Ordering: row.Ordering, PoolType: row.PoolType}, nil
}

// Country is the plain-Go-typed form of a country row
type Country struct {
	ID           int64
	CountryExtID int64 // external id, stable across events per the spec
	Name         string
	Abbreviation string
	FlagFile     string
}

func (s *Store) InsertCountry(ctx context.Context, c Country) (Country, error) {
	row, err := s.q.InsertCountry(ctx, dbgen.InsertCountryParams{
		CountryExtID: c.CountryExtID,
		Name:         c.Name,
		Abbreviation: c.Abbreviation,
		FlagFile:     c.FlagFile,
	})
	if err != nil {
		return Country{}, err
	}
	return Country{ID: row.ID, CountryExtID: row.CountryExtID, Name: row.Name, Abbreviation: row.Abbreviation, FlagFile: row.FlagFile}, nil
}
