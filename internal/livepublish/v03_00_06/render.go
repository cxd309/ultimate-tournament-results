package livepublish

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	livedatamodel "github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v03_00_06"
	livepublish "github.com/cxd309/ultimate-tournament-results/internal/livepublish"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v03_00_06"
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
		Season:         renderSeason(data),
		Series:         renderSeries(data),
		Pools:          renderPools(data),
		Teams:          renderTeams(data),
		Countries:      renderCountries(data),
		Reservations:   renderReservations(data),
		PoolPlacements: renderPoolPlacements(data),
	}
}

func renderSeason(data *tournamentData) livedatamodel.Season {
	t := data.tournament
	return livedatamodel.Season{
		Name:                           t.Name,
		StartTime:                      t.StartTime,
		EndTime:                        t.EndTime,
		Iscurrent:                      t.Iscurrent,
		Type:                           t.Type,
		Isinternational:                t.Isinternational,
		Isnationalteams:                t.Isnationalteams,
		Showspiritpointsonlyoncomplete: t.Showspiritpointsonlyoncomplete,
		Lockteamspiritonsubmit:         t.Lockteamspiritonsubmit,
		UseSeasonPoints:                t.UseSeasonPoints,
		HideTimeOnScoresheet:           t.HideTimeOnScoresheet,
		Hometeammode:                   t.Hometeammode,
		EventReadonly:                  t.EventReadonly,
		MaintenanceMode:                t.MaintenanceMode,
		PublicEvent:                    t.PublicEvent,
		ApiPublic:                      t.ApiPublic,
		Timezone:                       t.Timezone,
		// this archiver only ever archives a finished tournament
		// a republish is a static snapshot with no live clock to
		// recompute status against, so it's always "completed"
		Status:           "completed",
		SpiritMode:       t.Spiritmode,
		SpiritCategories: renderSpiritCategories(data),
		Timeslots:        renderTimeslots(data),
		PlayerCount:      data.playerCount,
		UtcOffset:        renderUtcOffset(data),
		Spirit:           convert.IntBool(len(data.spiritScoresByGame) > 0),
		GameTimesByDay:   renderGameTimesByDay(data),
	}
}

func renderSpiritCategories(data *tournamentData) []livedatamodel.SpiritCategory {
	categories := make([]livedatamodel.SpiritCategory, len(data.spiritCategories))
	for i, c := range data.spiritCategories {
		categories[i] = livedatamodel.SpiritCategory{
			CategoryID: c.CategoryID,
			Key:        data.spiritCategoryKeyByID[c.CategoryID],
			Index:      c.Ordering,
			Group:      c.CategoryGroup,
			Min:        c.Min,
			Max:        c.Max,
			Factor:     c.Factor,
			Label:      c.Label,
		}
	}
	return categories
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
			Isfollower:    p.Isfollower,
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
			Club:                    t.Club,
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
		var locationName string
		if r.Location != nil {
			locationName = data.locationByID[*r.Location].Name
		}
		reservations[i] = livedatamodel.Reservation{
			ID:               r.ID,
			FieldName:        r.FieldName,
			ReservationGroup: r.ReservationGroup,
			Location:         r.Location,
			LocationName:     locationName,
		}
	}
	return reservations
}

func renderPoolPlacements(data *tournamentData) []livedatamodel.PoolPlacement {
	placements := make([]livedatamodel.PoolPlacement, len(data.poolPlacements))
	for i, p := range data.poolPlacements {
		placements[i] = livedatamodel.PoolPlacement{
			PoolID:    p.PoolID,
			TeamID:    p.TeamID,
			Placement: p.Placement,
		}
	}
	return placements
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
	var clubname *string
	if team.ClubName != "" {
		clubname = &team.ClubName
	}
	detail := livedatamodel.TeamDetailResponse{
		TeamID:   team.TeamID,
		Pool:     convert.Int64OrZero(team.Pool),
		Valid:    team.Valid,
		Players:  stats,
		Clubname: clubname,
	}
	if len(data.spiritScoresByGame) > 0 {
		renderTeamSpirit(data, team, &detail)
	}
	return detail
}

// renderTeamSpirit fills a team's spiritgiven/spiritreceived and their totals
// only sent when the event publishes spirit points (season.spirit), same as live
//
// 3.0 only lists a game once both teams have submitted and it's cleared for
// display (show_spirit), so a game missing either side's score is skipped in
// both arrays, not just the one it's missing from
//
// spirit_scores is keyed by recipient: the rows for this team are what it
// received, the opponent's rows in the same game are what it gave
// one entry per game, in game time order
func renderTeamSpirit(data *tournamentData, team store.Team, detail *livedatamodel.TeamDetailResponse) {
	given := make([]livedatamodel.TeamSpiritGame, 0)
	received := make([]livedatamodel.TeamSpiritGame, 0)
	var total float64
	for _, g := range data.gamesByTime {
		opponent, ok := opponentOf(g, team.TeamID)
		if !ok || !bool(g.ShowSpirit) {
			continue
		}
		receivedScore, ok := teamSpiritGame(data, g, team.TeamID)
		if !ok {
			continue
		}
		givenScore, ok := teamSpiritGame(data, g, opponent)
		if !ok {
			continue
		}
		opponentName := data.teamByID[opponent].Name
		for _, entry := range []*livedatamodel.TeamSpiritGame{&receivedScore, &givenScore} {
			entry.Givenby = opponentName
			entry.Givento = opponentName
		}
		received = append(received, receivedScore)
		given = append(given, givenScore)
		total += float64(receivedScore.Total)
	}
	detail.SpiritGiven = given
	detail.SpiritReceived = received
	detail.SpiritStats = &livedatamodel.TeamSpiritStats{Games: int64(len(received))}
	detail.SpiritTotal = &livedatamodel.TeamSpiritTotal{Total: total}
}

// teamSpiritGame builds the spirit score recipient received in game g
// false unless every category has a visible (non-null) score
func teamSpiritGame(data *tournamentData, g store.Game, recipient int64) (livedatamodel.TeamSpiritGame, bool) {
	entry := livedatamodel.TeamSpiritGame{
		GameID:     g.GameID,
		Time:       g.Time,
		Spiritmode: data.tournament.Spiritmode,
		IsComplete: true,
		IsVisible:  true,
		Categories: make(map[string]int64),
	}
	for _, sc := range data.spiritScoresByGame[g.GameID] {
		if sc.TeamID != recipient {
			continue
		}
		if sc.Value == nil {
			return entry, false
		}
		entry.Categories[data.spiritCategoryKeyByID[sc.CategoryID]] = *sc.Value
		entry.Total += *sc.Value
	}
	if len(entry.Categories) == 0 {
		return entry, false
	}
	if comment, ok := data.spiritCommentsByGame[g.GameID][recipient]; ok {
		entry.Comments = &comment
	}
	return entry, true
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
// every falsy value stripped (see livedatamodel.Game) and scores restored to 0
// for any game that isn't scheduled
func renderGames(data *tournamentData) livedatamodel.GamesListResponse {
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
			Valid:                 g.Valid,
			Halftime:              convert.Int64OrZero(g.Halftime),
			Isongoing:             g.Isongoing,
			SchedulingNameHome:    convert.Int64OrZero(g.SchedulingNameHome),
			SchedulingNameVisitor: convert.Int64OrZero(g.SchedulingNameVisitor),
			Timeslot:              convert.Int64OrZero(g.Timeslot),
			Homedefenses:          convert.Int64OrZero(g.Homedefenses),
			Visitordefenses:       convert.Int64OrZero(g.Visitordefenses),
			Hasstarted:            convert.Int64OrZero(g.Hasstarted),
			Islive:                g.Islive,
			// the archive doesn't distinguish "" from NULL, so an empty liveurl
			// is stripped like a null one rather than guessed back to ""
			Liveurl:                   g.Liveurl,
			ShowSpirit:                g.ShowSpirit,
			TimerStart:                convert.Int64OrZero(g.TimerStart),
			TimerPauseStart:           convert.Int64OrZero(g.TimerPauseStart),
			TimerPausedDuration:       convert.Int64OrZero(g.TimerPausedDuration),
			Forfeit:                   g.Forfeit,
			Pool:                      convert.Int64OrZero(owningPool(data, g.GameID)),
			Gamename:                  data.schedulingName(g.Name).Name,
			Gameschedulingname:        data.schedulingName(g.Name).Name,
			Homeschedulingname:        data.schedulingName(g.SchedulingNameHome).Name,
			Visitorschedulingname:     data.schedulingName(g.SchedulingNameVisitor).Name,
			HomeSchedulingFrompool:    convert.Int64OrZero(data.schedulingName(g.SchedulingNameHome).Frompool),
			VisitorSchedulingFrompool: convert.Int64OrZero(data.schedulingName(g.SchedulingNameVisitor).Frompool),
			Pools:                     gamePools(data, g.GameID),
			Status:                    status,
			TimeUTC:                   livepublish.TimeUTC(g.Time, data.tournament.Timezone),
		}
		if g.Name != nil && *g.Name != 0 {
			entry.Name = strconv.FormatInt(*g.Name, 10)
		}
		games[i] = entry
	}
	return livedatamodel.GamesListResponse{Games: games}
}

// gameStatus derives a game's status the way 3.0 does, from hasstarted:
// ongoing while flagged live, completed once started, scheduled otherwise
// so 0-0 results and forfeits are completed, unlike 1.9
func gameStatus(g store.Game) string {
	switch {
	case bool(g.Isongoing):
		return "ongoing"
	case convert.Int64OrZero(g.Hasstarted) > 0:
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

// gamePools is every pool a game belongs to, as the list's comma-separated,
// de-duplicated, sorted string
func gamePools(data *tournamentData, gameID int64) string {
	ids := make([]int64, 0, len(data.poolsByGame[gameID]))
	for _, gp := range data.poolsByGame[gameID] {
		ids = append(ids, gp.PoolID)
	}
	slices.Sort(ids)
	ids = slices.Compact(ids)
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}

// schedulingName looks up a scheduling-name row by id
// the zero value for no id, or one this archive has no row for
func (data *tournamentData) schedulingName(id *int64) store.SchedulingName {
	if id == nil {
		return store.SchedulingName{}
	}
	return data.schedulingNameByID[*id]
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
		Pool:                  owningPool(data, game.GameID),
		Reservation:           game.Reservation,
		Valid:                 game.Valid,
		Isongoing:             game.Isongoing,
		Halftime:              game.Halftime,
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
		Hasstarted:            game.Hasstarted,
		ShowSpirit:            game.ShowSpirit,
		Forfeit:               game.Forfeit,
		TimerStart:            game.TimerStart,
		TimerPauseStart:       game.TimerPauseStart,
		TimerPausedDuration:   game.TimerPausedDuration,
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

// owningPool is the pool that "owns" (schedules) this game, i.e. the one
// with timetable=1 on game_pools -- uo_game itself has no pool column in
// UltiOrganizer 4, games join pools only through game_pools
func owningPool(data *tournamentData, gameID int64) *int64 {
	for _, gp := range data.poolsByGame[gameID] {
		if gp.Timetable {
			poolID := gp.PoolID
			return &poolID
		}
	}
	return nil
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

func renderPoolInfo(data *tournamentData, game store.Game) livedatamodel.PoolInfo {
	pool := owningPool(data, game.GameID)
	if pool == nil {
		return livedatamodel.PoolInfo{}
	}
	p, ok := data.poolByID[*pool]
	if !ok {
		return livedatamodel.PoolInfo{}
	}
	var playoffTemplate *string
	if p.PlayoffTemplate != "" {
		playoffTemplate = &p.PlayoffTemplate
	}
	return livedatamodel.PoolInfo{
		PoolID:           p.PoolID,
		Drawsallowed:     convert.IntBoolFromInt64(convert.Int64OrZero(p.Drawsallowed)),
		PlayoffTemplate:  playoffTemplate,
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
			Timestamp:    g.Timestamp,
		}
	}
	return goals
}

func renderSpiritStats(data *tournamentData, game store.Game) *livedatamodel.GameSpiritStats {
	scores := data.spiritScoresByGame[game.GameID]
	if len(scores) == 0 {
		return nil
	}
	comments := data.spiritCommentsByGame[game.GameID]

	byTeam := make(map[int64]*livedatamodel.GameSpiritScore)
	for _, sc := range scores {
		score, ok := byTeam[sc.TeamID]
		if !ok {
			score = &livedatamodel.GameSpiritScore{Categories: make(map[string]*int64)}
			if comment, ok := comments[sc.TeamID]; ok {
				score.Comments = &comment
			}
			byTeam[sc.TeamID] = score
		}
		key := data.spiritCategoryKeyByID[sc.CategoryID]
		score.Categories[key] = sc.Value
	}

	stats := &livedatamodel.GameSpiritStats{Note: spiritStatsNote}
	if game.Hometeam != nil {
		stats.Hometeam = byTeam[*game.Hometeam]
	}
	if game.Visitorteam != nil {
		stats.Visitorteam = byTeam[*game.Visitorteam]
	}
	return stats
}

// spiritStatsNote is a fixed literal string the live API sends unchanged
// on every game's spiritstats
const spiritStatsNote = "These are the scores FOR the specified team"

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
		var done, fedin, callahan int64
		for _, g := range goals {
			if g.Scorer != nil && *g.Scorer == p.PlayerID {
				done++
				if g.Iscallahan {
					callahan++
				}
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
			Callahan:  callahan,
			Total:     done + fedin,
		}
	}
	// sorted by total descending, done descending as a tie-break
	// matches the ordering observed from a real deployment
	sort.SliceStable(scoreboard, func(i, j int) bool {
		if scoreboard[i].Total != scoreboard[j].Total {
			return scoreboard[i].Total > scoreboard[j].Total
		}
		return scoreboard[i].Done > scoreboard[j].Done
	})
	return scoreboard
}
