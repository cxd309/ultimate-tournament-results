package uoarchive

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

// resolveGames fills games, scheduling names, goals, played, game events and
// game urls; spirit is resolveSpirit's
//
// A game is only archivable with the id its schedule row links to, and UltiOrganizer
// only links games that have a scoresheet, so a game without one is an error here
// rather than silently left out
func (r *rows) resolveGames(snap *uoclient.Snapshot) error {
	teamByName := map[string]int64{}
	for _, t := range snap.Teams {
		teamByName[t.Name] = t.TeamID
	}
	playerTeam := map[int64]int64{}
	for _, p := range r.players {
		playerTeam[p.PlayerID] = p.Team
	}

	var missing []string
	scheduled := map[int64]bool{}
	gameNames := map[int64]string{}
	slotIDs := map[string]int64{}

	for _, g := range snap.Schedule {
		if g.GameID == nil {
			missing = append(missing, fmt.Sprintf("%s %s %s - %s (%s %s)", g.ReservationGroup, g.Time, g.Home, g.Visitor, g.Division, g.PoolName))
			continue
		}
		id := *g.GameID
		if scheduled[id] {
			return fmt.Errorf("game %d scheduled twice", id)
		}
		scheduled[id] = true
		page, ok := snap.Games[id]
		if !ok {
			return fmt.Errorf("game %d: page not gathered", id)
		}
		row := csvRowFor(snap, g)
		if row == nil {
			return fmt.Errorf("game %d: no games CSV row for %s %s %s - %s", id, g.ReservationGroup, g.Time, g.Home, g.Visitor)
		}

		reservation, pool := g.ReservationID, g.PoolID
		game := store.Game{
			GameID:       id,
			Hometeam:     cmp.Or(page.HomeTeamID, teamID(teamByName, g.Home)),
			Visitorteam:  cmp.Or(page.VisitorTeamID, teamID(teamByName, g.Visitor)),
			Homescore:    g.HomeScore,
			Visitorscore: g.VisitorScore,
			Reservation:  &reservation,
			Time:         row["Time"],
			Pool:         &pool,
			Valid:        true,
			Halftime:     page.Halftime,
			Official:     page.Official,
		}

		// the CSV's GameName is uo_game.name, the scheduling-name id itself; its text
		// is the schedule row's game-name column
		if gameNameID, ok := parseID(row["GameName"]); ok && g.GameName != "" {
			if prev, seen := gameNames[gameNameID]; seen && prev != g.GameName {
				return fmt.Errorf("game %d: scheduling name %d is both %q and %q", id, gameNameID, prev, g.GameName)
			}
			gameNames[gameNameID] = g.GameName
			game.Name = &gameNameID
		}
		// the home/visitor slot names have no id shown anywhere, see schema
		game.SchedulingNameHome = slotID(slotIDs, row["HomeSchedulingName"])
		game.SchedulingNameVisitor = slotID(slotIDs, row["AwaySchedulingName"])

		r.games = append(r.games, game)

		if err := r.resolveGameDetail(id, game, page, playerTeam); err != nil {
			return fmt.Errorf("game %d: %w", id, err)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%d scheduled games have no game id (no scoresheet entered), first: %s", len(missing), missing[0])
	}
	for _, id := range snap.GameIDs() {
		if !scheduled[id] {
			return fmt.Errorf("game %d is listed on a pool page but not on the schedule", id)
		}
	}

	for _, id := range slices.Sorted(maps.Keys(gameNames)) {
		r.schedulingNames = append(r.schedulingNames, store.SchedulingName{SchedulingID: id, Name: gameNames[id]})
	}
	for name, id := range slotIDs {
		r.schedulingNames = append(r.schedulingNames, store.SchedulingName{SchedulingID: id, Name: name})
	}
	slices.SortFunc(r.schedulingNames, func(a, b store.SchedulingName) int { return cmp.Compare(a.SchedulingID, b.SchedulingID) })
	slices.SortFunc(r.games, func(a, b store.Game) int { return cmp.Compare(a.GameID, b.GameID) })
	return nil
}

// resolveGameDetail fills one game's goals, played, events and urls
// each scoresheet goal names its assist and scorer by number and name only, so each is
// resolved against the scoring team's scoreboard for this game; the result is then
// checked against the scoreboard's own per-player goal and assist counts
func (r *rows) resolveGameDetail(id int64, game store.Game, page uoclient.GamePage, playerTeam map[int64]int64) error {
	for side, board := range map[string][]uoclient.ScoreboardPlayer{"home": page.Home, "visitor": page.Visitor} {
		team := game.Hometeam
		if side == "visitor" {
			team = game.Visitorteam
		}
		for _, p := range board {
			if team == nil || playerTeam[p.PlayerID] != *team {
				return fmt.Errorf("%s scoreboard player %d isn't on that team's roster", side, p.PlayerID)
			}
			r.played = append(r.played, uostore.Played{GameID: id, PlayerID: p.PlayerID, Num: p.Num, Captain: convert.IntBool(p.Captain)})
		}
	}

	goals := map[int64]int64{}
	assists := map[int64]int64{}
	for i, g := range page.Goals {
		board := page.Visitor
		if g.IsHome {
			board = page.Home
		}
		scorer, err := resolveRef(board, g.Scorer)
		if err != nil {
			return fmt.Errorf("goal %d scorer: %w", i+1, err)
		}
		var assist *int64
		if !g.IsCallahan {
			if assist, err = resolveRef(board, g.Assist); err != nil {
				return fmt.Errorf("goal %d assist: %w", i+1, err)
			}
		}
		if scorer != nil {
			goals[*scorer]++
		}
		if assist != nil {
			assists[*assist]++
		}
		home, visitor := g.HomeScore, g.VisitorScore
		r.goals = append(r.goals, store.Goal{
			GameID:       id,
			Num:          int64(i + 1),
			Assist:       assist,
			Scorer:       scorer,
			Time:         g.Time,
			Homescore:    &home,
			Visitorscore: &visitor,
			Ishomegoal:   convert.IntBool(g.IsHome),
			Iscallahan:   convert.IntBool(g.IsCallahan),
		})
	}
	for _, p := range slices.Concat(page.Home, page.Visitor) {
		if goals[p.PlayerID] != p.Goals || assists[p.PlayerID] != p.Assists {
			return fmt.Errorf("player %d: scoresheet gives %d goals %d assists, scoreboard %d goals %d assists",
				p.PlayerID, goals[p.PlayerID], assists[p.PlayerID], p.Goals, p.Assists)
		}
	}

	for i, e := range page.Events {
		at := e.Time
		r.gameEvents = append(r.gameEvents, uostore.GameEvent{GameID: id, Num: int64(i + 1), Time: &at, Type: e.Type, Ishome: convert.IntBool(e.IsHome)})
	}
	for _, u := range page.URLs {
		r.urls = append(r.urls, urlRow("game", id, u))
	}
	return nil
}

// resolveRef finds the scoreboard player a scoresheet reference names
// matched on number and name, then on name alone when the number doesn't match (a
// player can wear a different number in one game); nil for an empty reference
func resolveRef(board []uoclient.ScoreboardPlayer, ref uoclient.PlayerRef) (*int64, error) {
	if ref.Name == "" {
		return nil, nil
	}
	var byName []int64
	for _, p := range board {
		if clean(p.FirstName+" "+p.LastName) != ref.Name {
			continue
		}
		if ref.Num != nil && p.Num != nil && *ref.Num == *p.Num {
			id := p.PlayerID
			return &id, nil
		}
		byName = append(byName, p.PlayerID)
	}
	if len(byName) != 1 {
		return nil, fmt.Errorf("%d scoreboard players match %q", len(byName), ref.Name)
	}
	return &byName[0], nil
}

// resolveTeamPools fills team_pools: every team in a pool's standings, with its
// "#" rank, plus every team scheduled in one of the pool's games
func (r *rows) resolveTeamPools(snap *uoclient.Snapshot) {
	type key struct{ team, pool int64 }
	activerank := map[key]*int64{}
	for id, p := range snap.Pools {
		for _, s := range p.Standings {
			activerank[key{s.TeamID, id}] = s.Activerank
		}
	}
	for _, g := range r.games {
		for _, team := range []*int64{g.Hometeam, g.Visitorteam} {
			if team != nil {
				if _, ok := activerank[key{*team, *g.Pool}]; !ok {
					activerank[key{*team, *g.Pool}] = nil
				}
			}
		}
	}
	for k, rank := range activerank {
		r.teamPools = append(r.teamPools, uostore.TeamPool{TeamID: k.team, PoolID: k.pool, Activerank: rank})
	}
	slices.SortFunc(r.teamPools, func(a, b uostore.TeamPool) int {
		return cmp.Or(cmp.Compare(a.TeamID, b.TeamID), cmp.Compare(a.PoolID, b.PoolID))
	})
}

// resolveGamePools fills game_pools: each game's own pool, plus every round-robin
// pool page listing it
func (r *rows) resolveGamePools(snap *uoclient.Snapshot) {
	type key struct{ game, pool int64 }
	seen := map[key]bool{}
	for _, g := range r.games {
		seen[key{g.GameID, *g.Pool}] = true
	}
	for id, p := range snap.Pools {
		for _, g := range p.GameIDs {
			seen[key{g, id}] = true
		}
	}
	for k := range seen {
		r.gamePools = append(r.gamePools, uostore.GamePool{GameID: k.game, PoolID: k.pool})
	}
	slices.SortFunc(r.gamePools, func(a, b uostore.GamePool) int {
		return cmp.Or(cmp.Compare(a.GameID, b.GameID), cmp.Compare(a.PoolID, b.PoolID))
	})
}

// resolveMoveTeams fills move_teams from each standings row's colour, matched to the
// continuation pool shown in that colour; a colour shared by two continuation pools
// can't be told apart, so those placings are left out
func (r *rows) resolveMoveTeams(snap *uoclient.Snapshot) {
	for _, id := range slices.Sorted(maps.Keys(snap.Pools)) {
		p := snap.Pools[id]
		byColor := map[string][]int64{}
		for _, c := range p.ContinuesTo {
			byColor[c.Color] = append(byColor[c.Color], c.PoolID)
		}
		for _, s := range p.Standings {
			if to := byColor[s.MoveColor]; s.MoveColor != "" && len(to) == 1 {
				r.moveTeams = append(r.moveTeams, uostore.MoveTeam{FromPool: id, FromPlacing: s.Placing, ToPool: to[0]})
			}
		}
	}
}

func (r *rows) resolveClubsAndProfiles(snap *uoclient.Snapshot) {
	clubs := map[int64]string{}
	for _, id := range slices.Sorted(maps.Keys(snap.TeamCards)) {
		card := snap.TeamCards[id]
		if card.ClubID != nil {
			clubs[*card.ClubID] = card.ClubName
			r.teamClubs = append(r.teamClubs, uostore.TeamClub{TeamID: id, ClubID: *card.ClubID})
		}
		if p := card.Profile; p != nil {
			r.teamProfiles = append(r.teamProfiles, uostore.TeamProfile{
				TeamID: id, Coach: p.Coach, Captain: p.Captain, Story: p.Story, Achievements: p.Achievements, ProfileImage: p.ProfileImage,
			})
		}
		for _, u := range card.URLs {
			r.urls = append(r.urls, urlRow("team", id, u))
		}
	}
	for _, id := range slices.Sorted(maps.Keys(clubs)) {
		r.clubs = append(r.clubs, uostore.Club{ClubID: id, Name: clubs[id]})
	}
}

func urlRow(owner string, id int64, u uoclient.URL) uostore.URL {
	return uostore.URL{
		Owner: owner, OwnerID: id, Type: u.Type, Name: u.Name, URL: u.URL,
		Ismedialink: convert.IntBool(u.IsMediaLink), Mediaowner: u.MediaOwner,
	}
}

func spiritScore(game, team int64, cats [5]int64, comments *string) store.SpiritScore {
	sc := store.SpiritScore{GameID: game, TeamID: team, Cat1: cats[0], Cat2: cats[1], Cat3: cats[2], Cat4: cats[3], Cat5: cats[4]}
	if comments != nil {
		sc.Comments = *comments
	}
	return sc
}

func sum(cats [5]int64) int64 {
	return cats[0] + cats[1] + cats[2] + cats[3] + cats[4]
}

// teamID is the team listed under name, nil for a name that isn't a team (an
// unresolved slot's scheduling name)
func teamID(byName map[string]int64, name string) *int64 {
	if id, ok := byName[name]; ok {
		return &id
	}
	return nil
}

// slotID allocates a synthetic id for a home/visitor scheduling name, negative so it
// can never be mistaken for one of UltiOrganizer's own; nil for no name
func slotID(ids map[string]int64, name string) *int64 {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	id, ok := ids[name]
	if !ok {
		id = -int64(len(ids) + 1)
		ids[name] = id
	}
	return &id
}

func parseID(s string) (int64, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n, err == nil && n > 0
}
