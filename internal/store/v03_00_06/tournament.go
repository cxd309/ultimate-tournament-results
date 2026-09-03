package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v03_00_06"
)

// Tournament is the plain-Go-typed form of a single row in the tournament table
type Tournament struct {
	SeasonID                       string
	Name                           string
	StartTime                      string
	EndTime                        string
	Iscurrent                      convert.IntBool
	Type                           string
	Isinternational                convert.IntBool
	Isnationalteams                convert.IntBool
	Showspiritpointsonlyoncomplete convert.IntBool
	Lockteamspiritonsubmit         convert.IntBool
	UseSeasonPoints                convert.IntBool
	HideTimeOnScoresheet           convert.IntBool
	Hometeammode                   convert.IntBool
	EventReadonly                  convert.IntBool
	MaintenanceMode                convert.IntBool
	PublicEvent                    convert.IntBool
	ApiPublic                      convert.IntBool
	Timezone                       string
	Spiritmode                     *int64 // which spirit score mode is used for the tournament; selects spirit_categories.mode
	Host                           string
	BasePath                       string
	AppVersion                     string
	ArchivedAt                     string
}

func (s *Store) InsertTournament(ctx context.Context, t Tournament) error {
	return s.q.InsertTournament(ctx, dbgen.InsertTournamentParams{
		SeasonID:                       t.SeasonID,
		Name:                           convert.NullString(t.Name),
		Starttime:                      convert.NullString(t.StartTime),
		Endtime:                        convert.NullString(t.EndTime),
		Iscurrent:                      t.Iscurrent.Int64(),
		Type:                           convert.NullString(t.Type),
		Isinternational:                t.Isinternational.NullInt64(),
		Isnationalteams:                t.Isnationalteams.NullInt64(),
		Showspiritpointsonlyoncomplete: t.Showspiritpointsonlyoncomplete.NullInt64(),
		Lockteamspiritonsubmit:         t.Lockteamspiritonsubmit.NullInt64(),
		UseSeasonPoints:                t.UseSeasonPoints.NullInt64(),
		HideTimeOnScoresheet:           t.HideTimeOnScoresheet.NullInt64(),
		Hometeammode:                   t.Hometeammode.NullInt64(),
		EventReadonly:                  t.EventReadonly.NullInt64(),
		MaintenanceMode:                t.MaintenanceMode.NullInt64(),
		PublicEvent:                    t.PublicEvent.Int64(),
		ApiPublic:                      t.ApiPublic.NullInt64(),
		Timezone:                       convert.NullString(t.Timezone),
		Spiritmode:                     convert.NullInt64(t.Spiritmode),
		Host:                           t.Host,
		BasePath:                       t.BasePath,
		AppVersion:                     convert.NullString(t.AppVersion),
		ArchivedAt:                     t.ArchivedAt,
	})
}

func (s *Store) GetTournament(ctx context.Context) (Tournament, error) {
	row, err := s.q.GetTournament(ctx)
	if err != nil {
		return Tournament{}, err
	}
	return Tournament{
		SeasonID:                       row.SeasonID,
		Name:                           convert.String(row.Name),
		StartTime:                      convert.String(row.Starttime),
		EndTime:                        convert.String(row.Endtime),
		Iscurrent:                      convert.IntBoolFromInt64(row.Iscurrent),
		Type:                           convert.String(row.Type),
		Isinternational:                convert.IntBoolFromInt64(row.Isinternational.Int64),
		Isnationalteams:                convert.IntBoolFromInt64(row.Isnationalteams.Int64),
		Showspiritpointsonlyoncomplete: convert.IntBoolFromInt64(row.Showspiritpointsonlyoncomplete.Int64),
		Lockteamspiritonsubmit:         convert.IntBoolFromInt64(row.Lockteamspiritonsubmit.Int64),
		UseSeasonPoints:                convert.IntBoolFromInt64(row.UseSeasonPoints.Int64),
		HideTimeOnScoresheet:           convert.IntBoolFromInt64(row.HideTimeOnScoresheet.Int64),
		Hometeammode:                   convert.IntBoolFromInt64(row.Hometeammode.Int64),
		EventReadonly:                  convert.IntBoolFromInt64(row.EventReadonly.Int64),
		MaintenanceMode:                convert.IntBoolFromInt64(row.MaintenanceMode.Int64),
		PublicEvent:                    convert.IntBoolFromInt64(row.PublicEvent),
		ApiPublic:                      convert.IntBoolFromInt64(row.ApiPublic.Int64),
		Timezone:                       convert.String(row.Timezone),
		Spiritmode:                     convert.Int64(row.Spiritmode),
		Host:                           row.Host,
		BasePath:                       row.BasePath,
		AppVersion:                     convert.String(row.AppVersion),
		ArchivedAt:                     row.ArchivedAt,
	}, nil
}
