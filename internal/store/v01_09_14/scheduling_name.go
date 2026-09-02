package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_14"
)

// SchedulingName is the plain-Go-typed form of a single row in the scheduling_names
// table -- resolves a scheduling-name id to its display text
type SchedulingName struct {
	SchedulingID int64
	Name         string
}

func (s *Store) InsertSchedulingName(ctx context.Context, sn SchedulingName) error {
	return s.q.InsertSchedulingName(ctx, dbgen.InsertSchedulingNameParams{
		SchedulingID: sn.SchedulingID,
		Name:         sn.Name,
	})
}
