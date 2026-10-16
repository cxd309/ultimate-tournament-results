package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// Location is the plain-Go-typed form of a single row in the locations table
// with the columns ext/locationxml.php exposes beyond the Live! API's id/name
type Location struct {
	ID      int64
	Name    string
	Fields  int64
	Indoor  convert.IntBool
	Address string
	Lat     *float64
	Lng     *float64
}

func (s *Store) InsertLocation(ctx context.Context, l Location) error {
	return s.q.InsertLocation(ctx, dbgen.InsertLocationParams{
		ID:      l.ID,
		Name:    l.Name,
		Fields:  l.Fields,
		Indoor:  l.Indoor.Int64(),
		Address: convert.NullString(l.Address),
		Lat:     convert.NullFloat64(l.Lat),
		Lng:     convert.NullFloat64(l.Lng),
	})
}
