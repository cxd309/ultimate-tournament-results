package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Player is the plain-Go-typed form of one players row. JerseyNum/GamesPlayed/Goals/
// Assists/Callahans are nil for NULL -- PlayerStats doesn't list these as required, so
// they aren't assumed always present (see internal/convert.NullInt64).
type Player struct {
	ID          int64
	PlayerID    int64 // external id
	TeamID      int64 // internal teams.id
	FirstName   string
	LastName    string
	JerseyNum   *int64
	GamesPlayed *int64
	Goals       *int64
	Assists     *int64
	Callahans   *int64
}

func (s *Store) InsertPlayer(ctx context.Context, p Player) (Player, error) {
	row, err := s.q.InsertPlayer(ctx, dbgen.InsertPlayerParams{
		PlayerID:    p.PlayerID,
		TeamID:      p.TeamID,
		FirstName:   p.FirstName,
		LastName:    p.LastName,
		JerseyNum:   convert.NullInt64(p.JerseyNum),
		GamesPlayed: convert.NullInt64(p.GamesPlayed),
		Goals:       convert.NullInt64(p.Goals),
		Assists:     convert.NullInt64(p.Assists),
		Callahans:   convert.NullInt64(p.Callahans),
	})
	if err != nil {
		return Player{}, err
	}
	return Player{
		ID:          row.ID,
		PlayerID:    row.PlayerID,
		TeamID:      row.TeamID,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		JerseyNum:   convert.Int64(row.JerseyNum),
		GamesPlayed: convert.Int64(row.GamesPlayed),
		Goals:       convert.Int64(row.Goals),
		Assists:     convert.Int64(row.Assists),
		Callahans:   convert.Int64(row.Callahans),
	}, nil
}
