package store

import (
	"context"
	"database/sql"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Pool is the plain-Go-typed form of a single row in the pools table
// Visible/Continuingpool/Placementpool/Played are IntBool
// the API omits them entirely rather than sending false
// so there's nothing to distinguish from a genuine 0/false
// hence plain int64 rather than a pointer dispite placementpool being nullable
type Pool struct {
	PoolID         int64
	Name           string
	Ordering       string
	Visible        int64
	Continuingpool int64
	Placementpool  int64
	Played         int64
	Series         *int64
	Type           int64
}

func (s *Store) InsertPool(ctx context.Context, p Pool) error {
	return s.q.InsertPool(ctx, dbgen.InsertPoolParams{
		PoolID:         p.PoolID,
		Name:           convert.NullString(p.Name),
		Ordering:       convert.NullString(p.Ordering),
		Visible:        p.Visible,
		Continuingpool: p.Continuingpool,
		Placementpool:  sql.NullInt64{Int64: p.Placementpool, Valid: true},
		Played:         p.Played,
		Series:         convert.NullInt64(p.Series),
		Type:           p.Type,
	})
}
