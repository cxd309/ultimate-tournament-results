package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// TeamProfile is the plain-Go-typed form of a single row in the team_profiles table
type TeamProfile struct {
	TeamID       int64
	Coach        string
	Captain      string
	Story        string
	Achievements string
	ProfileImage string
}

func (s *Store) InsertTeamProfile(ctx context.Context, p TeamProfile) error {
	return s.q.InsertTeamProfile(ctx, dbgen.InsertTeamProfileParams{
		TeamID:       p.TeamID,
		Coach:        convert.NullString(p.Coach),
		Captain:      convert.NullString(p.Captain),
		Story:        convert.NullString(p.Story),
		Achievements: convert.NullString(p.Achievements),
		ProfileImage: convert.NullString(p.ProfileImage),
	})
}
