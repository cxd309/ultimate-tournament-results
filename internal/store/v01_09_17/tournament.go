package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_17"
)

// Tournament is the plain-Go-typed form of a single row in the tournament table
type Tournament struct {
	SeasonID        string
	Name            string
	StartTime       string
	EndTime         string
	Iscurrent       convert.IntBool
	Type            string
	Isinternational convert.IntBool
	Isnationalteams convert.IntBool
	Timezone        string
	Host            string
	BasePath        string
	AppVersion      string
	ArchivedAt      string
}

func (s *Store) InsertTournament(ctx context.Context, t Tournament) error {
	return s.q.InsertTournament(ctx, dbgen.InsertTournamentParams{
		SeasonID:        t.SeasonID,
		Name:            convert.NullString(t.Name),
		Starttime:       convert.NullString(t.StartTime),
		Endtime:         convert.NullString(t.EndTime),
		Iscurrent:       t.Iscurrent.Int64(),
		Type:            convert.NullString(t.Type),
		Isinternational: t.Isinternational.NullInt64(),
		Isnationalteams: t.Isnationalteams.NullInt64(),
		Timezone:        convert.NullString(t.Timezone),
		Host:            t.Host,
		BasePath:        t.BasePath,
		AppVersion:      convert.NullString(t.AppVersion),
		ArchivedAt:      t.ArchivedAt,
	})
}

func (s *Store) GetTournament(ctx context.Context) (Tournament, error) {
	row, err := s.q.GetTournament(ctx)
	if err != nil {
		return Tournament{}, err
	}
	return Tournament{
		SeasonID:        row.SeasonID,
		Name:            convert.String(row.Name),
		StartTime:       convert.String(row.Starttime),
		EndTime:         convert.String(row.Endtime),
		Iscurrent:       convert.IntBoolFromInt64(row.Iscurrent),
		Type:            convert.String(row.Type),
		Isinternational: convert.IntBoolFromInt64(row.Isinternational.Int64),
		Isnationalteams: convert.IntBoolFromInt64(row.Isnationalteams.Int64),
		Timezone:        convert.String(row.Timezone),
		Host:            row.Host,
		BasePath:        row.BasePath,
		AppVersion:      convert.String(row.AppVersion),
		ArchivedAt:      row.ArchivedAt,
	}, nil
}
