package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_17"
)

// Pool is the plain-Go-typed form of a single row in the pools table
// Visible/Continuingpool/Placementpool/Played are IntBool: always known even though
// placementpool is a nullable column, so no pointer needed.
//
// Teams..Follower are only known once a game in this pool has been imported -- they
// come from that game's poolinfo, not the reference endpoint's own pools[] -- so stay
// nil for a pool with no games, e.g. an unused placeholder bracket pool.
type Pool struct {
	PoolID           int64
	Name             string
	Ordering         string
	Visible          convert.IntBool
	Continuingpool   convert.IntBool
	Placementpool    convert.IntBool
	Played           convert.IntBool
	Series           *int64
	Type             int64
	Color            string
	Timeslot         *int64
	Teams            *int64
	Mvgames          *int64
	Timeoutlen       *int64
	Halftime         *int64
	Winningscore     *int64
	Timecap          *int64
	Scorecap         *int64
	Addscore         *int64
	Halftimescore    *int64
	Timeouts         *int64
	Timeoutsper      string
	Timeoutsovertime *int64
	Timeoutstimecap  string
	Betweenpointslen *int64
	Forfeitscore     *int64
	Forfeitagainst   *int64
	Follower         *int64
}

func (s *Store) InsertPool(ctx context.Context, p Pool) error {
	return s.q.InsertPool(ctx, dbgen.InsertPoolParams{
		PoolID:           p.PoolID,
		Name:             convert.NullString(p.Name),
		Ordering:         convert.NullString(p.Ordering),
		Visible:          p.Visible.Int64(),
		Continuingpool:   p.Continuingpool.Int64(),
		Placementpool:    p.Placementpool.NullInt64(),
		Played:           p.Played.Int64(),
		Series:           convert.NullInt64(p.Series),
		Type:             p.Type,
		Color:            convert.NullString(p.Color),
		Timeslot:         convert.NullInt64(p.Timeslot),
		Teams:            convert.NullInt64(p.Teams),
		Mvgames:          convert.NullInt64(p.Mvgames),
		Timeoutlen:       convert.NullInt64(p.Timeoutlen),
		Halftime:         convert.NullInt64(p.Halftime),
		Winningscore:     convert.NullInt64(p.Winningscore),
		Timecap:          convert.NullInt64(p.Timecap),
		Scorecap:         convert.NullInt64(p.Scorecap),
		Addscore:         convert.NullInt64(p.Addscore),
		Halftimescore:    convert.NullInt64(p.Halftimescore),
		Timeouts:         convert.NullInt64(p.Timeouts),
		Timeoutsper:      convert.NullString(p.Timeoutsper),
		Timeoutsovertime: convert.NullInt64(p.Timeoutsovertime),
		Timeoutstimecap:  convert.NullString(p.Timeoutstimecap),
		Betweenpointslen: convert.NullInt64(p.Betweenpointslen),
		Forfeitscore:     convert.NullInt64(p.Forfeitscore),
		Forfeitagainst:   convert.NullInt64(p.Forfeitagainst),
		Follower:         convert.NullInt64(p.Follower),
	})
}
