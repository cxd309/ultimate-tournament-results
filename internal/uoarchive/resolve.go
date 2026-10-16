package uoarchive

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

// rows is a snapshot resolved into archive rows, ready to write in FK order
type rows struct {
	tournament      store.Tournament
	divisions       []store.Division
	countries       []store.Country
	locations       []uostore.Location
	reservations    []uostore.Reservation
	pools           []store.Pool
	teams           []store.Team
	players         []store.Player
	schedulingNames []store.SchedulingName
	games           []store.Game
	goals           []store.Goal
	spiritScores    []store.SpiritScore

	played       []uostore.Played
	gameEvents   []uostore.GameEvent
	teamPools    []uostore.TeamPool
	gamePools    []uostore.GamePool
	moveTeams    []uostore.MoveTeam
	clubs        []uostore.Club
	teamClubs    []uostore.TeamClub
	teamProfiles []uostore.TeamProfile
	urls         []uostore.URL

	serverSettings uostore.ServerSettings
}

// reservationTimeLayout is the reservation page's DefTimeFormat, "d.m.Y H:i", or
// "j.n.Y H:i" on sites that don't zero-pad the date (WMUCC 2026's "28.6.2026");
// Go's unpadded day and month read either
const reservationTimeLayout = "2.1.2006 15:04"

const archiveTimeLayout = "2006-01-02 15:04:05"

func resolve(snap *uoclient.Snapshot, host, basePath string) (*rows, error) {
	r := &rows{}
	if err := r.resolveReservations(snap); err != nil {
		return nil, fmt.Errorf("reservations: %w", err)
	}
	r.resolveTournament(snap, host, basePath)
	r.resolveDivisions(snap)
	r.resolveCountries(snap)
	if err := r.resolvePools(snap); err != nil {
		return nil, fmt.Errorf("pools: %w", err)
	}
	if err := r.resolveTeams(snap); err != nil {
		return nil, fmt.Errorf("teams: %w", err)
	}
	if err := r.resolvePlayers(snap); err != nil {
		return nil, fmt.Errorf("players: %w", err)
	}
	if err := r.resolveGames(snap); err != nil {
		return nil, fmt.Errorf("games: %w", err)
	}
	if err := r.resolveSpirit(snap); err != nil {
		return nil, fmt.Errorf("spirit: %w", err)
	}
	r.serverSettings.ShowSpiritComments = showSpiritComments(snap)
	r.resolveTeamPools(snap)
	r.resolveGamePools(snap)
	r.resolveMoveTeams(snap)
	r.resolveClubsAndProfiles(snap)
	return r, nil
}

// resolveTournament fills the single tournament row
// starttime/endtime are the first reservation's start and the last one's end:
// UltiOrganizer shows the season's own dates nowhere publicly
// type/isnationalteams aren't shown publicly either, left empty
func (r *rows) resolveTournament(snap *uoclient.Snapshot, host, basePath string) {
	var start, end string
	for _, res := range r.reservations {
		if start == "" || res.StartTime < start {
			start = res.StartTime
		}
		if res.EndTime > end {
			end = res.EndTime
		}
	}
	international := false
	for _, p := range snap.Pools {
		international = international || p.International
	}
	r.tournament = store.Tournament{
		SeasonID:        snap.SeasonID,
		Name:            snap.Menu.SeasonName,
		StartTime:       start,
		EndTime:         end,
		Isinternational: convert.IntBool(international),
		Timezone:        snap.Timezone,
		Host:            host,
		BasePath:        basePath,
		ArchivedAt:      time.Now().UTC().Format(time.RFC3339),
	}
}

func (r *rows) resolveDivisions(snap *uoclient.Snapshot) {
	for _, d := range snap.Menu.Divisions {
		r.divisions = append(r.divisions, store.Division{SeriesID: d.SeriesID, Name: d.Name})
	}
}

func (r *rows) resolveCountries(snap *uoclient.Snapshot) {
	seen := map[int64]bool{}
	for _, t := range snap.Teams {
		if t.CountryID == nil || seen[*t.CountryID] {
			continue
		}
		seen[*t.CountryID] = true
		r.countries = append(r.countries, store.Country{CountryID: *t.CountryID, Name: t.CountryName, FlagFile: t.FlagFile})
	}
	slices.SortFunc(r.countries, func(a, b store.Country) int { return cmp.Compare(a.CountryID, b.CountryID) })
}

// resolveReservations fills locations and reservations
// fieldname is the schedule header with its "Field " label taken off; the
// reservation's times come from its own page, its group from the schedule heading
// it's listed under. A reservation with no location gets location 0, written as a
// nameless row so the reservations FK holds, the same as UltiOrganizer's own id 0
func (r *rows) resolveReservations(snap *uoclient.Snapshot) error {
	type seen struct{ group, field string }
	byID := map[int64]seen{}
	for _, g := range snap.Schedule {
		field, ok := strings.CutPrefix(g.FieldLabel, "Field ")
		if !ok {
			return fmt.Errorf("reservation %d: header %q has no Field label", g.ReservationID, g.FieldLabel)
		}
		s := seen{g.ReservationGroup, field}
		if prev, ok := byID[g.ReservationID]; ok && prev != s {
			return fmt.Errorf("reservation %d listed as both %v and %v", g.ReservationID, prev, s)
		}
		byID[g.ReservationID] = s
	}

	needLocation0 := false
	for _, id := range slices.Sorted(maps.Keys(byID)) {
		page, ok := snap.Reservations[id]
		if !ok {
			return fmt.Errorf("reservation %d: page not gathered", id)
		}
		start, end, err := reservationTimes(page.Times)
		if err != nil {
			return fmt.Errorf("reservation %d: %w", id, err)
		}
		if page.LocationID == 0 {
			needLocation0 = true
		}
		r.reservations = append(r.reservations, uostore.Reservation{
			ID:               id,
			Location:         page.LocationID,
			FieldName:        byID[id].field,
			ReservationGroup: byID[id].group,
			StartTime:        start,
			EndTime:          end,
			Season:           snap.SeasonID,
			Date:             start[:10] + " 00:00:00",
		})
	}

	if needLocation0 {
		r.locations = append(r.locations, uostore.Location{ID: 0})
	}
	for _, id := range slices.Sorted(maps.Keys(snap.Locations)) {
		l := snap.Locations[id]
		fields, _ := strconv.ParseInt(l.Fields, 10, 64)
		r.locations = append(r.locations, uostore.Location{
			ID:      l.ID,
			Name:    l.Name,
			Fields:  fields,
			Indoor:  convert.IntBool(l.Indoor == "1"),
			Address: l.Address,
			Lat:     optionalFloat(l.Lat),
			Lng:     optionalFloat(l.Lng),
		})
	}
	return nil
}

// reservationTimes parses "DD.MM.YYYY HH:MM - HH:MM"; the end has no date of its own,
// it's the same day as the start
func reservationTimes(s string) (string, string, error) {
	startText, endHour, ok := strings.Cut(s, " - ")
	if !ok {
		return "", "", fmt.Errorf("unreadable times %q", s)
	}
	start, err := time.Parse(reservationTimeLayout, startText)
	if err != nil {
		return "", "", fmt.Errorf("unreadable start %q: %w", startText, err)
	}
	end, err := time.Parse(reservationTimeLayout, start.Format("02.01.2006")+" "+endHour)
	if err != nil {
		return "", "", fmt.Errorf("unreadable end %q: %w", endHour, err)
	}
	return start.Format(archiveTimeLayout), end.Format(archiveTimeLayout), nil
}

func optionalFloat(s *string) *float64 {
	if s == nil || *s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return nil
	}
	return &f
}

// resolvePools fills pools from each pool page
// name is the page heading with its "{division}, " prefix taken off; series is the
// menu division the pool is listed under, or for a pool the menu doesn't list (a
// hidden pool, e.g. a later playoff round) the division its heading names
// visible is whether the menu lists it. continuingpool is whether another pool
// continues to it, and color the colour that pool codes it with. played is whether
// every scheduled game in it has a result. type is as inferred by the pool page.
// Nothing else (ordering, rules, timeslot, follower) is shown publicly
func (r *rows) resolvePools(snap *uoclient.Snapshot) error {
	divisionOf := map[int64]uoclient.MenuDivision{}
	for _, d := range snap.Menu.Divisions {
		for _, p := range d.Pools {
			divisionOf[p.PoolID] = d
		}
	}
	color := map[int64]string{}
	for _, p := range snap.Pools {
		for _, c := range p.ContinuesTo {
			color[c.PoolID] = c.Color
		}
	}
	played := map[int64]bool{}
	for _, g := range snap.Schedule {
		hasResult := g.HomeScore != nil && g.VisitorScore != nil
		if prev, ok := played[g.PoolID]; ok {
			played[g.PoolID] = prev && hasResult
		} else {
			played[g.PoolID] = hasResult
		}
	}

	for _, id := range slices.Sorted(maps.Keys(snap.Pools)) {
		page := snap.Pools[id]
		division, visible := divisionOf[id]
		if !visible {
			for _, d := range snap.Menu.Divisions {
				if strings.HasPrefix(page.Heading, d.Name+", ") {
					division = d
				}
			}
			if division.SeriesID == 0 {
				return fmt.Errorf("pool %d: heading %q names no known division", id, page.Heading)
			}
		}
		name, ok := strings.CutPrefix(page.Heading, division.Name+", ")
		if !ok {
			return fmt.Errorf("pool %d: heading %q doesn't start with its division %q", id, page.Heading, division.Name)
		}
		_, continuing := color[id]
		series := division.SeriesID
		r.pools = append(r.pools, store.Pool{
			PoolID:         id,
			Name:           name,
			Visible:        convert.IntBool(visible),
			Continuingpool: convert.IntBool(continuing),
			Played:         convert.IntBool(played[id]),
			Series:         &series,
			Type:           page.Type,
			Color:          color[id],
		})
	}
	return nil
}

// resolveTeams fills teams from the teams list and each team card
// rank is the byseeding seed. pool is the pool of the team's last scheduled game,
// UltiOrganizer's own uo_team.pool (the pool a team currently sits in) isn't shown,
// falling back to the first pool listing it. final_standing and
// final_standing_calculated are both the bystandings placing: UltiOrganizer
// computes it from the placement pools, there's no separate override to report
// abbreviation and club come from the teams CSV, matched by division and name
func (r *rows) resolveTeams(snap *uoclient.Snapshot) error {
	abbreviation := map[string]string{}
	club := map[string]string{}
	for _, row := range snap.TeamsCSV {
		key := row["Division"] + "\x00" + row["Team"]
		abbreviation[key] = row["ShortName"]
		club[key] = row["Club"]
	}

	lastPool := map[int64]int64{}
	lastTime := map[int64]string{}
	for _, g := range snap.Schedule {
		if g.GameID == nil {
			continue
		}
		page := snap.Games[*g.GameID]
		for _, team := range []*int64{page.HomeTeamID, page.VisitorTeamID} {
			if team == nil {
				continue
			}
			when := gameTime(snap, g)
			if when >= lastTime[*team] {
				lastTime[*team] = when
				lastPool[*team] = g.PoolID
			}
		}
	}
	firstPool := map[int64]int64{}
	for _, id := range slices.Sorted(maps.Keys(snap.Pools)) {
		for _, s := range snap.Pools[id].Standings {
			if _, ok := firstPool[s.TeamID]; !ok {
				firstPool[s.TeamID] = id
			}
		}
	}

	for _, t := range snap.Teams {
		card, ok := snap.TeamCards[t.TeamID]
		if !ok {
			return fmt.Errorf("team %d: card not gathered", t.TeamID)
		}
		team := store.Team{
			TeamID:  t.TeamID,
			Name:    t.Name,
			Valid:   true,
			Series:  card.SeriesID,
			Country: t.CountryID,
		}
		if pool, ok := lastPool[t.TeamID]; ok {
			team.Pool = &pool
		} else if pool, ok := firstPool[t.TeamID]; ok {
			team.Pool = &pool
		}
		if seed, ok := snap.Seeds[t.TeamID]; ok {
			team.Rank = &seed
		}
		if standing, ok := snap.Standings[t.TeamID]; ok {
			team.FinalStanding = &standing
			team.FinalStandingCalculated = &standing
		}
		key := t.Division + "\x00" + t.Name
		team.Abbreviation = abbreviation[key]
		team.ClubName = cmp.Or(card.ClubName, club[key])
		r.teams = append(r.teams, team)
	}
	slices.SortFunc(r.teams, func(a, b store.Team) int { return cmp.Compare(a.TeamID, b.TeamID) })
	return nil
}

// resolvePlayers fills players from each team card's roster
// the card prints names as one string, so first/last are taken from the player's
// game scoreboards, which separate them exactly; a player with no games falls back to
// the players CSV, matched by team, number and name
func (r *rows) resolvePlayers(snap *uoclient.Snapshot) error {
	type name struct{ first, last string }
	names := map[int64]name{}
	for _, g := range snap.Games {
		for _, p := range slices.Concat(g.Home, g.Visitor) {
			names[p.PlayerID] = name{p.FirstName, p.LastName}
		}
	}
	csvNames := map[string]name{}
	for _, row := range snap.PlayersCSV {
		key := row["TeamName"] + "\x00" + row["Jersey"] + "\x00" + clean(row["FirstName"]+" "+row["LastName"])
		csvNames[key] = name{row["FirstName"], row["LastName"]}
	}

	seen := map[int64]int64{}
	for _, t := range snap.Teams {
		for _, p := range snap.TeamCards[t.TeamID].Roster {
			if team, dup := seen[p.PlayerID]; dup {
				return fmt.Errorf("player %d on both team %d and team %d", p.PlayerID, team, t.TeamID)
			}
			seen[p.PlayerID] = t.TeamID
			n, ok := names[p.PlayerID]
			if !ok {
				jersey := ""
				if p.Num != nil {
					jersey = strconv.FormatInt(*p.Num, 10)
				}
				n, ok = csvNames[t.Name+"\x00"+jersey+"\x00"+p.Name]
			}
			if !ok {
				return fmt.Errorf("player %d (%q, team %d): no first/last name split found", p.PlayerID, p.Name, t.TeamID)
			}
			r.players = append(r.players, store.Player{
				PlayerID:    p.PlayerID,
				FirstName:   n.first,
				LastName:    n.last,
				Team:        t.TeamID,
				Num:         p.Num,
				GamesPlayed: p.Games,
			})
		}
	}
	slices.SortFunc(r.players, func(a, b store.Player) int { return cmp.Compare(a.PlayerID, b.PlayerID) })
	return nil
}

func clean(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// gameTime is a scheduled game's full start time from the games CSV, the schedule
// page only shows HH:MM under a date heading
func gameTime(snap *uoclient.Snapshot, g uoclient.ScheduledGame) string {
	if row := csvRowFor(snap, g); row != nil {
		return row["Time"]
	}
	return ""
}

// csvRowFor is the games CSV row for a scheduled game, matched on reservation group,
// field, start time and the two listed names (team or scheduling name), which
// together can only be one game
func csvRowFor(snap *uoclient.Snapshot, g uoclient.ScheduledGame) map[string]string {
	field := strings.TrimPrefix(g.FieldLabel, "Field ")
	for _, row := range snap.GamesCSV {
		home := cmp.Or(row["HomeTeam"], row["HomeSchedulingName"])
		visitor := cmp.Or(row["AwayTeam"], row["AwaySchedulingName"])
		if row["ReservationGroup"] == g.ReservationGroup && clean(row["Field"]) == field &&
			len(row["Time"]) >= 16 && row["Time"][11:16] == g.Time &&
			clean(home) == g.Home && clean(visitor) == g.Visitor {
			return row
		}
	}
	return nil
}
