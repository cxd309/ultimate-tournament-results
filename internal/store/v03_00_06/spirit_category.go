package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// SpiritCategory is the plain-Go-typed form of a single row in the spirit_categories
// table
type SpiritCategory struct {
	CategoryID    int64
	Mode          int64
	CategoryGroup int64
	Ordering      int64
	Min           int64
	Max           int64
	Factor        int64
	Label         string
}

func (s *Store) InsertSpiritCategory(ctx context.Context, c SpiritCategory) error {
	return s.q.InsertSpiritCategory(ctx, dbgen.InsertSpiritCategoryParams{
		CategoryID:    c.CategoryID,
		Mode:          c.Mode,
		CategoryGroup: c.CategoryGroup,
		Ordering:      c.Ordering,
		Min:           c.Min,
		Max:           c.Max,
		Factor:        c.Factor,
		Label:         c.Label,
	})
}
