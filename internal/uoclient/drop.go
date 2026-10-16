package uoclient

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// DropDivisions removes whole divisions from a gathered season, as if the site had
// never had them: their menu entries, schedule rows, pools, teams and team cards,
// games, CSV rows, and any reservation or location only they used. For a division
// that isn't archivable or wanted, e.g. WUC's Guts divisions, which were scored
// without scoresheets. Each name must be a division in the left menu
func (s *Snapshot) DropDivisions(names []string) error {
	if len(names) == 0 {
		return nil
	}
	drop := map[string]bool{}
	for _, name := range names {
		if !slices.ContainsFunc(s.Menu.Divisions, func(d MenuDivision) bool { return d.Name == name }) {
			return fmt.Errorf("no %q division to drop", name)
		}
		drop[name] = true
	}

	s.Menu.Divisions = slices.DeleteFunc(s.Menu.Divisions, func(d MenuDivision) bool { return drop[d.Name] })

	var droppedGames []int64
	s.Schedule = slices.DeleteFunc(s.Schedule, func(g ScheduledGame) bool {
		if drop[g.Division] && g.GameID != nil {
			droppedGames = append(droppedGames, *g.GameID)
		}
		return drop[g.Division]
	})
	for _, id := range droppedGames {
		delete(s.Games, id)
	}

	for id, p := range s.Pools {
		division, _, _ := strings.Cut(p.Heading, ", ")
		if drop[division] {
			delete(s.Pools, id)
		}
	}

	s.Teams = slices.DeleteFunc(s.Teams, func(t TeamListEntry) bool {
		if drop[t.Division] {
			delete(s.TeamCards, t.TeamID)
			delete(s.Seeds, t.TeamID)
			delete(s.Standings, t.TeamID)
		}
		return drop[t.Division]
	})

	byDivision := func(column string) func(map[string]string) bool {
		return func(row map[string]string) bool { return drop[row[column]] }
	}
	s.GamesCSV = slices.DeleteFunc(s.GamesCSV, byDivision("Division"))
	s.TeamsCSV = slices.DeleteFunc(s.TeamsCSV, byDivision("Division"))
	s.PlayersCSV = slices.DeleteFunc(s.PlayersCSV, byDivision("Division"))
	if s.SpiritCSV != nil {
		s.SpiritCSV = slices.DeleteFunc(s.SpiritCSV, byDivision("division"))
	}

	used := map[int64]bool{}
	for _, g := range s.Schedule {
		used[g.ReservationID] = true
	}
	usedLocations := map[int64]bool{}
	for _, id := range slices.Collect(maps.Keys(s.Reservations)) {
		if !used[id] {
			delete(s.Reservations, id)
			continue
		}
		usedLocations[s.Reservations[id].LocationID] = true
	}
	for id := range s.Locations {
		if !usedLocations[id] {
			delete(s.Locations, id)
		}
	}
	return nil
}
