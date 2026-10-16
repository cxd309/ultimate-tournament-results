package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/uo_01"
)

// Reservation is the plain-Go-typed form of a single row in the reservations table
// with the times the reservation page shows beyond the Live! API's columns
type Reservation struct {
	ID               int64
	Location         int64
	FieldName        string
	ReservationGroup string
	StartTime        string // "YYYY-MM-DD HH:MM:SS", tournament-local
	EndTime          string
	Season           string
	Date             string // StartTime's day at 00:00:00, as uo_reservation.date stores it
}

func (s *Store) InsertReservation(ctx context.Context, r Reservation) error {
	return s.q.InsertReservation(ctx, dbgen.InsertReservationParams{
		ID:               r.ID,
		Location:         r.Location,
		Fieldname:        convert.NullString(r.FieldName),
		Reservationgroup: convert.NullString(r.ReservationGroup),
		Starttime:        convert.NullString(r.StartTime),
		Endtime:          convert.NullString(r.EndTime),
		Season:           convert.NullString(r.Season),
		Date:             convert.NullString(r.Date),
	})
}

// UpdateReservation fills in the times of an existing row, e.g. one a Live! archive
// wrote
func (s *Store) UpdateReservation(ctx context.Context, r Reservation) error {
	return s.q.UpdateReservation(ctx, dbgen.UpdateReservationParams{
		ID:        r.ID,
		Starttime: convert.NullString(r.StartTime),
		Endtime:   convert.NullString(r.EndTime),
		Season:    convert.NullString(r.Season),
		Date:      convert.NullString(r.Date),
	})
}
