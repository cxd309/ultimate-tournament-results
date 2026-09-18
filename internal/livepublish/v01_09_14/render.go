package livepublish

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	livedatamodel "github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v01_09_14"
	livepublish "github.com/cxd309/ultimate-tournament-results/internal/livepublish"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

// liveRoundMinutes is the fixed buffer PHP calls LIVE_ROUND_MINUTES
// added to a day's last game start to get gameTimesByDay's "end"
const liveRoundMinutes = 60

func renderHeartbeat(data *tournamentData) livedatamodel.HeartbeatResponse {
	return livedatamodel.HeartbeatResponse{
		AppVersion: data.tournament.AppVersion,
		// LastUpdatedUTC is the archive's own capture time
		// the original cache-freshness meaning doesn't apply to a static
		// republish, this is the closest honest equivalent
		LastUpdatedUTC: data.tournament.ArchivedAt,
		Config: livedatamodel.HeartbeatConfig{
			LiveSeasonID:   data.tournament.SeasonID,
			TournamentName: data.tournament.Name,
		},
	}
}

func renderReference(data *tournamentData) livedatamodel.ReferenceResponse {
	return livedatamodel.ReferenceResponse{
		Season:       renderSeason(data),
		Series:       renderSeries(data),
		Pools:        renderPools(data),
		Teams:        renderTeams(data),
		Countries:    renderCountries(data),
		Reservations: renderReservations(data),
	}
}

func renderSeason(data *tournamentData) livedatamodel.Season {
	t := data.tournament
	return livedatamodel.Season{
		Name:            t.Name,
		StartTime:       t.StartTime,
		EndTime:         t.EndTime,
		Iscurrent:       t.Iscurrent,
		Type:            t.Type,
		Isinternational: t.Isinternational,
		Isnationalteams: t.Isnationalteams,
		Timezone:        t.Timezone,
		// this archiver only ever archives a finished tournament
		// a republish is a static snapshot with no live clock to
		// recompute status against, so it's always "completed"
		Status:         "completed",
		Timeslots:      renderTimeslots(data),
		PlayerCount:    data.playerCount,
		UtcOffset:      renderUtcOffset(data),
		Spirit:         convert.IntBool(len(data.spiritScoresByGame) > 0),
		GameTimesByDay: renderGameTimesByDay(data),
	}
}

// renderTimeslots collects every distinct pool timeslot, sorted ascending
func renderTimeslots(data *tournamentData) []int64 {
	seen := make(map[int64]bool)
	timeslots := make([]int64, 0)
	for _, p := range data.pools {
		if p.Timeslot == nil || seen[*p.Timeslot] {
			continue
		}
		seen[*p.Timeslot] = true
		timeslots = append(timeslots, *p.Timeslot)
	}
	sort.Slice(timeslots, func(i, j int) bool { return timeslots[i] < timeslots[j] })
	return timeslots
}

// renderUtcOffset computes the tournament timezone's UTC offset
// evaluated at the archive's own capture instant, since a republish has
// no live "now" to evaluate it against
// an invalid or empty timezone renders as ""
func renderUtcOffset(data *tournamentData) string {
	loc, err := time.LoadLocation(data.tournament.Timezone)
	if err != nil {
		return ""
	}
	at, err := time.Parse(time.RFC3339, data.tournament.ArchivedAt)
	if err != nil {
		return ""
	}
	_, offsetSeconds := at.In(loc).Zone()
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	return fmt.Sprintf("%s%02d:%02d", sign, offsetSeconds/3600, (offsetSeconds%3600)/60)
}

// renderGameTimesByDay groups every scheduled game by its calendar day
// first/last are the earliest/latest start times that day, End is Last
// plus the fixed liveRoundMinutes buffer
func renderGameTimesByDay(data *tournamentData) map[string]livedatamodel.DayTimes {
	type span struct{ first, last time.Time }
	byDay := make(map[string]span)
	for _, g := range data.games {
		if g.Time == "" {
			continue
		}
		t, err := time.Parse("2006-01-02 15:04:05", g.Time)
		if err != nil {
			continue
		}
		day := t.Format("2006-01-02")
		s, ok := byDay[day]
		if !ok || t.Before(s.first) {
			s.first = t
		}
		if !ok || t.After(s.last) {
			s.last = t
		}
		byDay[day] = s
	}

	result := make(map[string]livedatamodel.DayTimes, len(byDay))
	for day, s := range byDay {
		result[day] = livedatamodel.DayTimes{
			First: s.first.Format("15:04"),
			Last:  s.last.Format("15:04"),
			End:   s.last.Add(liveRoundMinutes * time.Minute).Format("15:04"),
		}
	}
	return result
}

func renderSeries(data *tournamentData) []livedatamodel.Series {
	series := make([]livedatamodel.Series, len(data.divisions))
	for i, d := range data.divisions {
		series[i] = livedatamodel.Series{
			SeriesID: d.SeriesID,
			Name:     d.Name,
			Ordering: d.Ordering,
		}
	}
	return series
}

func renderPools(data *tournamentData) []livedatamodel.Pool {
	pools := make([]livedatamodel.Pool, len(data.pools))
	for i, p := range data.pools {
		pools[i] = livedatamodel.Pool{
			PoolID:        p.PoolID,
			PoolName:      p.Name,
			SeriesID:      convert.Int64OrZero(p.Series),
			Ordering:      p.Ordering,
			Type:          p.Type,
			Visible:       p.Visible,
			Played:        p.Played,
			Placementpool: p.Placementpool,
			Continuing:    p.Continuingpool,
			Color:         convert.NumericString(p.Color),
			Timeslot:      p.Timeslot,
		}
	}
	return pools
}

func renderTeams(data *tournamentData) []livedatamodel.ReferenceTeam {
	teams := make([]livedatamodel.ReferenceTeam, len(data.teams))
	for i, t := range data.teams {
		teams[i] = livedatamodel.ReferenceTeam{
			TeamID:                  t.TeamID,
			Name:                    t.Name,
			Abbreviation:            t.Abbreviation,
			Series:                  convert.Int64OrZero(t.Series),
			Country:                 convert.Int64OrZero(t.Country),
			Rank:                    convert.Int64OrZero(t.Rank),
			FinalStanding:           convert.Int64OrZero(t.FinalStanding),
			FinalStandingCalculated: convert.Int64OrZero(t.FinalStandingCalculated),
			Club:                    t.ClubName,
		}
	}
	return teams
}

func renderCountries(data *tournamentData) []livedatamodel.Country {
	countries := make([]livedatamodel.Country, len(data.countries))
	for i, c := range data.countries {
		countries[i] = livedatamodel.Country{
			CountryID:    c.CountryID,
			Name:         c.Name,
			Abbreviation: c.Abbreviation,
			FlagFile:     c.FlagFile,
		}
	}
	return countries
}

func renderReservations(data *tournamentData) []livedatamodel.Reservation {
	reservations := make([]livedatamodel.Reservation, len(data.reservations))
	for i, r := range data.reservations {
		reservations[i] = livedatamodel.Reservation{
			ID:               r.ID,
			FieldName:        r.FieldName,
			ReservationGroup: r.ReservationGroup,
			Location:         r.Location,
			LocationName:     data.locationByID[r.Location].Name,
		}
	}
	return reservations
}

func renderTeamDetail(data *tournamentData, team store.Team) livedatamodel.TeamDetailResponse {
	players := data.playersByTeam[team.TeamID]
	stats := make([]livedatamodel.PlayerStats, len(players))
	for i, p := range players {
		stats[i] = livedatamodel.PlayerStats{
			PlayerID:  p.PlayerID,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Num:       p.Num,
			Games:     p.GamesPlayed,
		}
	}
	detail := livedatamodel.TeamDetailResponse{
		TeamID:  team.TeamID,
		Pool:    convert.Int64OrZero(team.Pool),
		Valid:   team.Valid,
		Players: stats,
	}
	if len(data.spiritScoresByGame) > 0 {
		renderTeamSpirit(data, team, &detail)
	}
	return detail
}

// renderTeamSpirit fills a team's spiritgiven/spiritreceived and their totals
// only sent when the event publishes spirit points (season.spirit), same as live
//
// spirit_scores is keyed by recipient: the row for this team is what it
// received, the opponent's row in the same game is what it gave
// one entry per game with a score, in game time order
func renderTeamSpirit(data *tournamentData, team store.Team, detail *livedatamodel.TeamDetailResponse) {
	given := make([]livedatamodel.TeamSpiritGame, 0)
	received := make([]livedatamodel.TeamSpiritGame, 0)
	// spiritstats/spirittotal only count games where both teams submitted,
	// unlike the arrays, which list each direction on its own
	var games, total int64
	for _, g := range data.gamesByTime {
		opponent, ok := opponentOf(g, team.TeamID)
		if !ok {
			continue
		}
		opponentName := data.teamByID[opponent].Name
		var gave, got *livedatamodel.TeamSpiritGame
		for _, sc := range data.spiritScoresByGame[g.GameID] {
			entry := teamSpiritGame(sc)
			switch sc.TeamID {
			case team.TeamID:
				entry.Givenby = opponentName
				got = &entry
			case opponent:
				entry.Givento = opponentName
				gave = &entry
			}
		}
		if gave != nil {
			given = append(given, *gave)
		}
		if got != nil {
			received = append(received, *got)
		}
		if gave != nil && got != nil {
			games++
			total += got.Total
		}
	}
	detail.SpiritGiven = given
	detail.SpiritReceived = received
	detail.SpiritStats = &livedatamodel.TeamSpiritStats{Games: games}
	detail.SpiritTotal = &livedatamodel.TeamSpiritTotal{Total: total}
}

func teamSpiritGame(sc store.SpiritScore) livedatamodel.TeamSpiritGame {
	entry := livedatamodel.TeamSpiritGame{
		GameID: sc.GameID,
		Cat1:   sc.Cat1,
		Cat2:   sc.Cat2,
		Cat3:   sc.Cat3,
		Cat4:   sc.Cat4,
		Cat5:   sc.Cat5,
		Total:  sc.Cat1 + sc.Cat2 + sc.Cat3 + sc.Cat4 + sc.Cat5,
	}
	if sc.Comments != "" {
		entry.Comments = &sc.Comments
	}
	return entry
}

// opponentOf is the other team in a game teamID played in
// false when teamID didn't play in it, or the other slot is unresolved
func opponentOf(g store.Game, teamID int64) (int64, bool) {
	switch {
	case g.Hometeam != nil && *g.Hometeam == teamID && g.Visitorteam != nil:
		return *g.Visitorteam, true
	case g.Visitorteam != nil && *g.Visitorteam == teamID && g.Hometeam != nil:
		return *g.Hometeam, true
	default:
		return 0, false
	}
}

// renderGames rebuilds the games list, following the live endpoint's own rules:
// every falsy value stripped (see livedatamodel.Game), scores restored to 0 for
// any game that isn't scheduled, and time_utc only on 1.9.17+
func renderGames(data *tournamentData) livedatamodel.GamesListResponse {
	withTimeUTC := sendsTimeUTC(data.tournament.AppVersion)
	games := make([]livedatamodel.Game, len(data.games))
	for i, g := range data.games {
		status := gameStatus(g)
		entry := livedatamodel.Game{
			GameID:                g.GameID,
			Hometeam:              convert.Int64OrZero(g.Hometeam),
			Visitorteam:           convert.Int64OrZero(g.Visitorteam),
			Homescore:             listScore(g.Homescore, status),
			Visitorscore:          listScore(g.Visitorscore, status),
			Reservation:           convert.Int64OrZero(g.Reservation),
			Time:                  g.Time,
			Pool:                  convert.Int64OrZero(g.Pool),
			Valid:                 g.Valid,
			Halftime:              convert.Int64OrZero(g.Halftime),
			Visitorsotg:           convert.Int64OrZero(g.Visitorsotg),
			Isongoing:             g.Isongoing,
			SchedulingNameHome:    convert.Int64OrZero(g.SchedulingNameHome),
			SchedulingNameVisitor: convert.Int64OrZero(g.SchedulingNameVisitor),
			Timeslot:              convert.Int64OrZero(g.Timeslot),
			Homedefenses:          convert.Int64OrZero(g.Homedefenses),
			Visitordefenses:       convert.Int64OrZero(g.Visitordefenses),
			Islive:                convert.Int64OrZero(g.Islive),
			// the archive doesn't distinguish "" from NULL, so an empty liveurl
			// is stripped like a null one rather than guessed back to ""
			Liveurl:               g.Liveurl,
			Gamename:              data.schedulingName(g.Name),
			Gameschedulingname:    data.schedulingName(g.Name),
			Homeschedulingname:    data.schedulingName(g.SchedulingNameHome),
			Visitorschedulingname: data.schedulingName(g.SchedulingNameVisitor),
			Status:                status,
		}
		if g.Name != nil && *g.Name != 0 {
			entry.Name = strconv.FormatInt(*g.Name, 10)
		}
		if withTimeUTC {
			entry.TimeUTC = livepublish.TimeUTC(g.Time, data.tournament.Timezone)
		}
		games[i] = entry
	}
	return livedatamodel.GamesListResponse{Games: games}
}

// gameStatus derives a game's status the way 1.9 does:
// ongoing when flagged live, completed once either score is above zero,
// scheduled otherwise -- so a finished 0-0 stays scheduled
func gameStatus(g store.Game) string {
	switch {
	case bool(g.Isongoing):
		return "ongoing"
	case convert.Int64OrZero(g.Homescore) > 0 || convert.Int64OrZero(g.Visitorscore) > 0:
		return "completed"
	default:
		return "scheduled"
	}
}

// listScore is a games-list score: stripped when falsy, except restored to 0
// for any game that isn't scheduled
func listScore(score *int64, status string) *int64 {
	value := convert.Int64OrZero(score)
	if value == 0 && status == "scheduled" {
		return nil
	}
	return &value
}

// sendsTimeUTC reports whether this deployment's games list carries time_utc,
// new in 1.9.17
// app_version is whatever the heartbeat reported, which isn't always a
// release number (e.g. "dev", a build timestamp); anything unparseable is
// treated as pre-1.9.17 rather than guessed
func sendsTimeUTC(appVersion string) bool {
	parts := strings.Split(strings.TrimPrefix(appVersion, "v"), ".")
	if len(parts) != 3 {
		return false
	}
	version := [3]int{}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return false
		}
		version[i] = n
	}
	for i, want := range [3]int{1, 9, 17} {
		if version[i] != want {
			return version[i] > want
		}
	}
	return true
}

// schedulingName resolves a scheduling-name id to its display text
// "" for no id, or one this archive has no row for
func (data *tournamentData) schedulingName(id *int64) string {
	if id == nil {
		return ""
	}
	return data.schedulingNameByID[*id].Name
}

func renderGameDetail(data *tournamentData, game store.Game) livedatamodel.GameDetailResponse {
	gamename := ""
	if game.Name != nil {
		gamename = data.schedulingNameByID[*game.Name].Name
	}

	result := livedatamodel.GameResult{
		GameID:                game.GameID,
		Hometeam:              game.Hometeam,
		Visitorteam:           game.Visitorteam,
		Homescore:             game.Homescore,
		Visitorscore:          game.Visitorscore,
		Time:                  game.Time,
		Pool:                  game.Pool,
		Reservation:           game.Reservation,
		Valid:                 game.Valid,
		Isongoing:             game.Isongoing,
		Halftime:              game.Halftime,
		Homesotg:              game.Homesotg,
		Visitorsotg:           game.Visitorsotg,
		Official:              game.Official,
		Respteam:              game.Respteam,
		Resppers:              game.Resppers,
		Timeslot:              game.Timeslot,
		Homedefenses:          game.Homedefenses,
		Visitordefenses:       game.Visitordefenses,
		Islive:                game.Islive,
		Liveurl:               game.Liveurl,
		SchedulingNameHome:    game.SchedulingNameHome,
		SchedulingNameVisitor: game.SchedulingNameVisitor,
		Name:                  convert.OptionalStringFromInt64(game.Name),
		Gamename:              gamename,
	}

	return livedatamodel.GameDetailResponse{
		GameResult:            result,
		GameInfo:              renderGameInfo(data, game),
		PoolInfo:              renderPoolInfo(data, game),
		Goals:                 renderGoals(data, game),
		SpiritStats:           renderSpiritStats(data, game),
		HometeamScoreboard:    renderScoreboard(data, game.Hometeam, game.GameID),
		VisitorteamScoreboard: renderScoreboard(data, game.Visitorteam, game.GameID),
	}
}

func renderGameInfo(data *tournamentData, game store.Game) *livedatamodel.GameInfo {
	if game.SchedulingNameHome == nil && game.SchedulingNameVisitor == nil {
		return nil
	}
	info := &livedatamodel.GameInfo{}
	if game.SchedulingNameHome != nil {
		if sn, ok := data.schedulingNameByID[*game.SchedulingNameHome]; ok {
			info.Phometeamname = &sn.Name
		}
	}
	if game.SchedulingNameVisitor != nil {
		if sn, ok := data.schedulingNameByID[*game.SchedulingNameVisitor]; ok {
			info.Pvisitorteamname = &sn.Name
		}
	}
	return info
}

func renderPoolInfo(data *tournamentData, game store.Game) *livedatamodel.PoolInfo {
	if game.Pool == nil {
		return nil
	}
	p, ok := data.poolByID[*game.Pool]
	if !ok {
		return nil
	}
	return &livedatamodel.PoolInfo{
		PoolID:           p.PoolID,
		Teams:            p.Teams,
		Mvgames:          p.Mvgames,
		Timeoutlen:       p.Timeoutlen,
		Halftime:         p.Halftime,
		Winningscore:     p.Winningscore,
		Timecap:          p.Timecap,
		Scorecap:         p.Scorecap,
		Addscore:         p.Addscore,
		Halftimescore:    p.Halftimescore,
		Timeouts:         p.Timeouts,
		Timeoutsper:      p.Timeoutsper,
		Timeoutsovertime: p.Timeoutsovertime,
		Timeoutstimecap:  convert.OptionalInt64FromString(p.Timeoutstimecap),
		Betweenpointslen: p.Betweenpointslen,
		Forfeitscore:     p.Forfeitscore,
		Forfeitagainst:   p.Forfeitagainst,
		Follower:         p.Follower,
	}
}

func renderGoals(data *tournamentData, game store.Game) []livedatamodel.Goal {
	stored := data.goalsByGame[game.GameID]
	goals := make([]livedatamodel.Goal, len(stored))
	for i, g := range stored {
		goals[i] = livedatamodel.Goal{
			Num:          g.Num,
			Time:         g.Time,
			Scorer:       convert.ZeroIfNil(g.Scorer),
			Assist:       convert.ZeroIfNil(g.Assist),
			Homescore:    convert.Int64OrZero(g.Homescore),
			Visitorscore: convert.Int64OrZero(g.Visitorscore),
			Ishomegoal:   g.Ishomegoal,
			Iscallahan:   g.Iscallahan,
		}
	}
	return goals
}

// spiritStatsNote is a fixed literal string the live API sends unchanged
// on every game's spiritstats
const spiritStatsNote = "These are the scores FOR the specified team"

func renderSpiritStats(data *tournamentData, game store.Game) *livedatamodel.GameSpiritStats {
	scores := data.spiritScoresByGame[game.GameID]
	if len(scores) == 0 {
		return nil
	}
	stats := &livedatamodel.GameSpiritStats{Note: spiritStatsNote}
	for _, sc := range scores {
		score := &livedatamodel.GameSpiritScore{
			GameID:   sc.GameID,
			TeamID:   sc.TeamID,
			Cat1:     sc.Cat1,
			Cat2:     sc.Cat2,
			Cat3:     sc.Cat3,
			Cat4:     sc.Cat4,
			Cat5:     sc.Cat5,
			Comments: sc.Comments,
		}
		switch {
		case game.Hometeam != nil && sc.TeamID == *game.Hometeam:
			stats.Hometeam = score
		case game.Visitorteam != nil && sc.TeamID == *game.Visitorteam:
			stats.Visitorteam = score
		}
	}
	return stats
}

// renderScoreboard builds one team's full roster for this game
// each player's done/fedin/total counted from this game's own goals
// nil teamID means an unresolved bracket slot, no roster to show
func renderScoreboard(data *tournamentData, teamID *int64, gameID int64) []livedatamodel.GameScoreboardPlayer {
	if teamID == nil {
		return nil
	}
	goals := data.goalsByGame[gameID]
	roster := data.playersByTeam[*teamID]

	scoreboard := make([]livedatamodel.GameScoreboardPlayer, len(roster))
	for i, p := range roster {
		var done, fedin int64
		for _, g := range goals {
			if g.Scorer != nil && *g.Scorer == p.PlayerID {
				done++
			}
			if g.Assist != nil && *g.Assist == p.PlayerID {
				fedin++
			}
		}
		scoreboard[i] = livedatamodel.GameScoreboardPlayer{
			PlayerID:  p.PlayerID,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Num:       p.Num,
			Done:      done,
			Fedin:     fedin,
			Total:     done + fedin,
		}
	}
	// sorted by total descending, done descending as a tie-break
	// the tie-break isn't independently confirmed for this version, but
	// matches what a real 3.0.6 deployment does, and costs nothing if wrong
	sort.SliceStable(scoreboard, func(i, j int) bool {
		if scoreboard[i].Total != scoreboard[j].Total {
			return scoreboard[i].Total > scoreboard[j].Total
		}
		return scoreboard[i].Done > scoreboard[j].Done
	})
	return scoreboard
}
