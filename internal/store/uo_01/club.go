package store

import (
	"context"

	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// Club is the plain-Go-typed form of a single row in the clubs table
type Club struct {
	ClubID int64
	Name   string
}

func (s *Store) InsertClub(ctx context.Context, c Club) error {
	return s.q.InsertClub(ctx, dbgen.InsertClubParams{
		ClubID: c.ClubID,
		Name:   c.Name,
	})
}

// TeamClub is the plain-Go-typed form of a single row in the team_clubs table
type TeamClub struct {
	TeamID int64
	ClubID int64
}

func (s *Store) InsertTeamClub(ctx context.Context, tc TeamClub) error {
	return s.q.InsertTeamClub(ctx, dbgen.InsertTeamClubParams{
		TeamID: tc.TeamID,
		ClubID: tc.ClubID,
	})
}
