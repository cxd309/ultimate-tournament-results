package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v1914"
)

// Tournament is the plain-Go-typed form of the tournament table's single row. Optional
// fields are "" for NULL -- see internal/convert.NullString.
type Tournament struct {
	EventName  string
	Host       string
	SeasonID   string
	BasePath   string
	AppVersion string
	StartDate  string
	EndDate    string
	Timezone   string
	Status     string
	ArchivedAt string
}

func (s *Store) InsertTournament(ctx context.Context, t Tournament) error {
	return s.q.InsertTournament(ctx, dbgen.InsertTournamentParams{
		EventName:  t.EventName,
		Host:       t.Host,
		SeasonID:   t.SeasonID,
		BasePath:   t.BasePath,
		AppVersion: convert.NullString(t.AppVersion),
		StartDate:  convert.NullString(t.StartDate),
		EndDate:    convert.NullString(t.EndDate),
		Timezone:   convert.NullString(t.Timezone),
		Status:     convert.NullString(t.Status),
		ArchivedAt: t.ArchivedAt,
	})
}

func (s *Store) GetTournament(ctx context.Context) (Tournament, error) {
	row, err := s.q.GetTournament(ctx)
	if err != nil {
		return Tournament{}, err
	}
	return Tournament{
		EventName:  row.EventName,
		Host:       row.Host,
		SeasonID:   row.SeasonID,
		BasePath:   row.BasePath,
		AppVersion: convert.String(row.AppVersion),
		StartDate:  convert.String(row.StartDate),
		EndDate:    convert.String(row.EndDate),
		Timezone:   convert.String(row.Timezone),
		Status:     convert.String(row.Status),
		ArchivedAt: row.ArchivedAt,
	}, nil
}
