package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_17"
)

// Pool is the plain-Go-typed form of a single row in the pools table
// Visible/Continuingpool/Placementpool/Played are IntBool: always known even though
// placementpool is a nullable column, so no pointer needed.
type Pool struct {
	PoolID         int64
	Name           string
	Ordering       string
	Visible        convert.IntBool
	Continuingpool convert.IntBool
	Placementpool  convert.IntBool
	Played         convert.IntBool
	Series         *int64
	Type           int64
}

func (s *Store) InsertPool(ctx context.Context, p Pool) error {
	return s.q.InsertPool(ctx, dbgen.InsertPoolParams{
		PoolID:         p.PoolID,
		Name:           convert.NullString(p.Name),
		Ordering:       convert.NullString(p.Ordering),
		Visible:        p.Visible.Int64(),
		Continuingpool: p.Continuingpool.Int64(),
		Placementpool:  p.Placementpool.NullInt64(),
		Played:         p.Played.Int64(),
		Series:         convert.NullInt64(p.Series),
		Type:           p.Type,
	})
}
