package store

import (
	"context"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	dbgen "github.com/cxd309/ultimate-tournament-results/internal/db/v01_09_14"
)

// Reservation is the plain-Go-typed form of a single row in the reservations table
// Location must be an id already present in locations table
type Reservation struct {
	ID               int64
	Location         int64
	FieldName        string
	ReservationGroup string
}

func (s *Store) InsertReservation(ctx context.Context, r Reservation) error {
	return s.q.InsertReservation(ctx, dbgen.InsertReservationParams{
		ID:               r.ID,
		Location:         r.Location,
		Fieldname:        convert.NullString(r.FieldName),
		Reservationgroup: convert.NullString(r.ReservationGroup),
	})
}

func (s *Store) ListReservations(ctx context.Context) ([]Reservation, error) {
	rows, err := s.q.ListReservations(ctx)
	if err != nil {
		return nil, err
	}
	reservations := make([]Reservation, len(rows))
	for i, row := range rows {
		reservations[i] = Reservation{
			ID:               row.ID,
			Location:         row.Location,
			FieldName:        convert.String(row.Fieldname),
			ReservationGroup: convert.String(row.Reservationgroup),
		}
	}
	return reservations, nil
}
