package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// URL is the plain-Go-typed form of a single row in the urls table
type URL struct {
	Owner       string // game or team
	OwnerID     int64
	Type        string
	Name        string
	URL         string
	Ismedialink convert.IntBool
	Mediaowner  string
}

func (s *Store) InsertURL(ctx context.Context, u URL) error {
	return s.q.InsertURL(ctx, dbgen.InsertURLParams{
		Owner:       u.Owner,
		OwnerID:     u.OwnerID,
		Type:        convert.NullString(u.Type),
		Name:        convert.NullString(u.Name),
		Url:         u.URL,
		Ismedialink: u.Ismedialink.Int64(),
		Mediaowner:  convert.NullString(u.Mediaowner),
	})
}
