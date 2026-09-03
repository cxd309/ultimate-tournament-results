package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// SchedulingName is a plain-Go-typed single row in the scheduling_names table
// resolves a scheduling-name id to its display text, and optionally the pool
// a placeholder slot's team will be drawn from
type SchedulingName struct {
	SchedulingID int64
	Name         string
	Frompool     *int64
}

func (s *Store) InsertSchedulingName(ctx context.Context, sn SchedulingName) error {
	return s.q.InsertSchedulingName(ctx, dbgen.InsertSchedulingNameParams{
		SchedulingID: sn.SchedulingID,
		Name:         sn.Name,
		Frompool:     convert.NullInt64(sn.Frompool),
	})
}
