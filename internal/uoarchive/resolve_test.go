package uoarchive

import (
	"testing"

	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

func TestReservationTimes(t *testing.T) {
	start, end, err := reservationTimes("10.11.2024 11:00 - 19:00")
	if err != nil || start != "2024-11-10 11:00:00" || end != "2024-11-10 19:00:00" {
		t.Errorf("got %q %q %v", start, end, err)
	}
	if _, _, err := reservationTimes("2024-11-10 11:00"); err == nil {
		t.Error("accepted an unexpected format")
	}
}

func TestResolveRef(t *testing.T) {
	num := func(n int64) *int64 { return &n }
	board := []uoclient.ScoreboardPlayer{
		{PlayerID: 1, Num: num(3), FirstName: "Juan ", LastName: "Arias"},
		{PlayerID: 2, Num: num(23), FirstName: "Mario ", LastName: "Plaza "},
		{PlayerID: 3, Num: num(7), FirstName: "Sam", LastName: "Lee"},
		{PlayerID: 4, Num: num(8), FirstName: "Sam", LastName: "Lee"},
	}
	cases := []struct {
		ref  uoclient.PlayerRef
		want int64
	}{
		{uoclient.PlayerRef{Num: num(23), Name: "Mario Plaza"}, 2},
		{uoclient.PlayerRef{Num: num(99), Name: "Juan Arias"}, 1}, // number differs this game
		{uoclient.PlayerRef{Name: "Juan Arias"}, 1},
		{uoclient.PlayerRef{Num: num(8), Name: "Sam Lee"}, 4}, // same name, told apart by number
	}
	for _, c := range cases {
		got, err := resolveRef(board, c.ref)
		if err != nil || got == nil || *got != c.want {
			t.Errorf("resolveRef(%+v) = %v, %v; want %d", c.ref, got, err, c.want)
		}
	}
	if got, err := resolveRef(board, uoclient.PlayerRef{}); got != nil || err != nil {
		t.Errorf("empty ref = %v, %v", got, err)
	}
	if _, err := resolveRef(board, uoclient.PlayerRef{Name: "Sam Lee"}); err == nil {
		t.Error("ambiguous name without a number resolved")
	}
	if _, err := resolveRef(board, uoclient.PlayerRef{Name: "Nobody"}); err == nil {
		t.Error("unknown name resolved")
	}
}
