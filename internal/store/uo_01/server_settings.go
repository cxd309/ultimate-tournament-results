package store

import (
	"context"
	"fmt"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// ServerSettings is the plain-Go-typed form of the server_settings table's one row
type ServerSettings struct {
	ShowSpiritComments *bool // nil when it can't be told, see the schema
}

func (s *Store) InsertServerSettings(ctx context.Context, st ServerSettings) error {
	var show *int64
	if st.ShowSpiritComments != nil {
		n := convert.IntBool(*st.ShowSpiritComments).Int64()
		show = &n
	}
	return s.q.InsertServerSettings(ctx, convert.NullInt64(show))
}

// GetServerSettings reads the server_settings row; ok is false for an archive with no
// such table, i.e. a plain Live! archive, or a uo_01 one from before it was added
func (s *Store) GetServerSettings(ctx context.Context) (st ServerSettings, ok bool, err error) {
	var n int64
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'server_settings'").Scan(&n); err != nil {
		return st, false, fmt.Errorf("check for server_settings table: %w", err)
	}
	if n == 0 {
		return st, false, nil
	}
	show, err := s.q.GetServerSettings(ctx)
	if err != nil {
		return st, false, err
	}
	if show.Valid {
		b := show.Int64 != 0
		st.ShowSpiritComments = &b
	}
	return st, true, nil
}

// SetSpiritComments sets the comments of an existing spirit_scores row
func (s *Store) SetSpiritComments(ctx context.Context, gameID, teamID int64, comments string) error {
	return s.q.SetSpiritComments(ctx, dbgen.SetSpiritCommentsParams{
		GameID:   gameID,
		TeamID:   teamID,
		Comments: convert.NullString(comments),
	})
}
