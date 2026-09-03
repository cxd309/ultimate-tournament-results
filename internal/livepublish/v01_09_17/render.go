package livepublish

import (
	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	livedatamodel "github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v01_09_17"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_17"
)

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
		Status: "completed",
	}
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
	return livedatamodel.TeamDetailResponse{
		TeamID:  team.TeamID,
		Pool:    convert.Int64OrZero(team.Pool),
		Valid:   team.Valid,
		Players: stats,
	}
}

func renderGames(data *tournamentData) livedatamodel.GamesResponse {
	games := make([]livedatamodel.GameListEntry, len(data.games))
	for i, g := range data.games {
		games[i] = livedatamodel.GameListEntry{GameID: g.GameID}
	}
	return livedatamodel.GamesResponse{Games: games}
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
		GameResult:  result,
		GameInfo:    renderGameInfo(data, game),
		PoolInfo:    renderPoolInfo(data, game),
		Goals:       renderGoals(data, game),
		SpiritStats: renderSpiritStats(data, game),
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

func renderSpiritStats(data *tournamentData, game store.Game) *livedatamodel.GameSpiritStats {
	scores := data.spiritScoresByGame[game.GameID]
	if len(scores) == 0 {
		return nil
	}
	stats := &livedatamodel.GameSpiritStats{}
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
