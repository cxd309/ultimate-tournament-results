package uoclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/cxd309/ultimate-tournament-results/internal/liveclient"
)

// Snapshot is everything gathered from one UltiOrganizer season, page by page
// resolving it into archive rows is uoarchive's job
type Snapshot struct {
	SeasonID     string
	Menu         Menu
	Timezone     string
	Schedule     []ScheduledGame
	Teams        []TeamListEntry
	Seeds        map[int64]int64
	Standings    map[int64]int64 // nil when the season has no computed standings page
	Pools        map[int64]PoolPage
	Reservations map[int64]ReservationPage
	Locations    map[int64]Location
	TeamCards    map[int64]TeamCard
	Games        map[int64]GamePage

	// CSV exports, one map per row keyed by the export's own header
	GamesCSV   []map[string]string
	TeamsCSV   []map[string]string
	PlayersCSV []map[string]string
	// SpiritCSV is every submitted spirit score with its comments, including a game
	// where only one team has submitted, which the game page hides until both have
	// nil when the site has locked the export down
	SpiritCSV []map[string]string
}

// Gather fetches one season. seasonID may be empty when the site hosts exactly one
// season, which is then discovered from its season list
func (c *Client) Gather(ctx context.Context, seasonID string) (*Snapshot, error) {
	if seasonID == "" {
		doc, err := c.getView(ctx, "seasonlist", nil)
		if err != nil {
			return nil, fmt.Errorf("season list: %w", err)
		}
		ids := parseSeasonList(doc)
		if len(ids) != 1 {
			return nil, fmt.Errorf("site lists %d seasons (%s), pass -season-id to pick one", len(ids), strings.Join(ids, ", "))
		}
		seasonID = ids[0]
	}
	c.SeasonID = seasonID
	snap := &Snapshot{SeasonID: seasonID}

	fmt.Printf("fetching schedule for %s...\n", seasonID)
	doc, err := c.getView(ctx, "games", url.Values{"season": {seasonID}, "filter": {"places"}, "group": {"all"}})
	if err != nil {
		return nil, fmt.Errorf("schedule: %w", err)
	}
	if snap.Menu, err = parseMenu(doc); err != nil {
		return nil, fmt.Errorf("schedule: %w", err)
	}
	if snap.Menu.SeasonID != seasonID {
		return nil, fmt.Errorf("schedule: site showed season %q, not %q", snap.Menu.SeasonID, seasonID)
	}
	if snap.Schedule, snap.Timezone, err = parseSchedule(doc); err != nil {
		return nil, fmt.Errorf("schedule: %w", err)
	}
	unlisted := addUnlisted(snap, c.Unlisted)

	fmt.Println("fetching team lists...")
	if doc, err = c.getView(ctx, "teams", url.Values{"season": {seasonID}, "list": {"allteams"}}); err != nil {
		return nil, fmt.Errorf("teams: %w", err)
	}
	if snap.Teams, err = parseAllTeams(doc); err != nil {
		return nil, fmt.Errorf("teams: %w", err)
	}
	if doc, err = c.getView(ctx, "teams", url.Values{"season": {seasonID}, "list": {"byseeding"}}); err != nil {
		return nil, fmt.Errorf("seeds: %w", err)
	}
	snap.Seeds = parseSeeds(doc)
	if doc, err = c.getView(ctx, "teams", url.Values{"season": {seasonID}, "list": {"bystandings"}}); err != nil {
		return nil, fmt.Errorf("standings: %w", err)
	}
	if standings, ok := parseStandings(doc); ok {
		snap.Standings = standings
	} else {
		fmt.Println("  no computed standings table, final standings left empty")
	}

	if err := c.gatherPools(ctx, snap); err != nil {
		return nil, err
	}
	if err := linkUnlisted(snap, unlisted); err != nil {
		return nil, err
	}
	if err := c.gatherReservations(ctx, snap); err != nil {
		return nil, err
	}

	fmt.Printf("fetching %d team cards...\n", len(snap.Teams))
	snap.TeamCards = make(map[int64]TeamCard, len(snap.Teams))
	var fromScoreboard []string
	for i, t := range snap.Teams {
		card, err := c.gatherTeamCard(ctx, snap.Menu, t)
		if err != nil {
			return nil, err
		}
		if card.FromScoreboard {
			fromScoreboard = append(fromScoreboard, strconv.FormatInt(t.TeamID, 10))
		}
		snap.TeamCards[t.TeamID] = card
		liveclient.PrintProgress(i+1, len(snap.Teams), 10)
	}
	if len(fromScoreboard) > 0 {
		fmt.Printf("  %d team cards error on this site, rosters read from the team scoreboard instead, without profiles, clubs or links: teams %s\n",
			len(fromScoreboard), strings.Join(fromScoreboard, ", "))
	}

	gameIDs := snap.GameIDs()
	fmt.Printf("fetching %d games...\n", len(gameIDs))
	snap.Games = make(map[int64]GamePage, len(gameIDs))
	for i, id := range gameIDs {
		doc, err := c.getView(ctx, "gameplay", url.Values{"game": {strconv.FormatInt(id, 10)}})
		if err != nil {
			return nil, fmt.Errorf("game %d: %w", id, err)
		}
		game, err := parseGame(doc, id)
		if err != nil {
			return nil, err
		}
		snap.Games[id] = game
		liveclient.PrintProgress(i+1, len(gameIDs), 25)
	}

	fmt.Println("fetching CSV exports...")
	if snap.GamesCSV, err = c.getCSV(ctx, "gamescsv"); err != nil {
		return nil, fmt.Errorf("games csv: %w", err)
	}
	if snap.TeamsCSV, err = c.getCSV(ctx, "teamscsv"); err != nil {
		return nil, fmt.Errorf("teams csv: %w", err)
	}
	if snap.PlayersCSV, err = c.getCSV(ctx, "playerscsv"); err != nil {
		return nil, fmt.Errorf("players csv: %w", err)
	}
	switch snap.SpiritCSV, err = c.SpiritCSV(ctx); {
	case errors.Is(err, ErrNotCSV):
		fmt.Println("  spirit export is locked, spirit taken from game pages (both-submitted only, comments only where shown)")
	case err != nil:
		return nil, fmt.Errorf("spirit csv: %w", err)
	}

	return snap, nil
}

// addUnlisted appends the rows of unlisted whose pool the schedule page doesn't list,
// returning their indexes in snap.Schedule
func addUnlisted(snap *Snapshot, unlisted []ScheduledGame) []int {
	listed := map[int64]bool{}
	for _, g := range snap.Schedule {
		listed[g.PoolID] = true
	}
	var added []int
	for _, g := range unlisted {
		if !listed[g.PoolID] {
			added = append(added, len(snap.Schedule))
			snap.Schedule = append(snap.Schedule, g)
		}
	}
	if len(added) > 0 {
		fmt.Printf("  %d games in pools the schedule doesn't list, taken from the archive\n", len(added))
	}
	return added
}

// linkUnlisted keeps an added row's game id only when its pool page links the game,
// as the schedule would have: GameRow links a game only once it has a scoresheet,
// and a playoff page with no bracket template lists its games with GameRow
func linkUnlisted(snap *Snapshot, added []int) error {
	for _, i := range added {
		g := &snap.Schedule[i]
		pool, ok := snap.Pools[g.PoolID]
		if !ok {
			return fmt.Errorf("pool %d: page not gathered", g.PoolID)
		}
		if g.GameID != nil && !slices.Contains(pool.LinkedGames, *g.GameID) {
			g.GameID = nil
		}
	}
	return nil
}

// gatherPools fetches every pool the menu or the schedule names, then any pool those
// pages continue to, until no new pool turns up
func (c *Client) gatherPools(ctx context.Context, snap *Snapshot) error {
	var queue []int64
	for _, d := range snap.Menu.Divisions {
		for _, p := range d.Pools {
			queue = append(queue, p.PoolID)
		}
	}
	for _, g := range snap.Schedule {
		queue = append(queue, g.PoolID)
	}

	fmt.Println("fetching pools...")
	snap.Pools = map[int64]PoolPage{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if _, done := snap.Pools[id]; done {
			continue
		}
		doc, err := c.getView(ctx, "poolstatus", url.Values{"pool": {strconv.FormatInt(id, 10)}})
		if err != nil {
			return fmt.Errorf("pool %d: %w", id, err)
		}
		var pool PoolPage
		if isGamePage(doc) {
			pool, err = singleGamePool(snap, id)
			pool.International = flagFile(content(doc)) != ""
		} else {
			pool, err = parsePool(doc, id)
		}
		if err != nil {
			return err
		}
		snap.Pools[id] = pool
		for _, cont := range pool.ContinuesTo {
			queue = append(queue, cont.PoolID)
		}
	}
	fmt.Printf("  %d pools\n", len(snap.Pools))
	return nil
}

// isGamePage is whether a page is a game's view=gameplay, headed by an <h1> of its
// teams and score; a pool page heads with an <h2>. Neither link tells them apart: a
// game with no scoresheet fed in has no gamecard link, and a pool's unplayed games
// link theirs
func isGamePage(doc *html.Node) bool {
	h := find(content(doc), func(n *html.Node) bool {
		return n.Type == html.ElementNode && (n.Data == "h1" || n.Data == "h2")
	})
	return h != nil && h.Data == "h1"
}

// singleGamePool is a round-robin pool with exactly one game, whose poolstatus
// redirects to that game's gameplay page, so the pool itself is never shown: its name
// is the left menu's and its game the one schedule row in it. It has no standings, and
// no games when that row isn't linked (a game with no scoresheet)
func singleGamePool(snap *Snapshot, id int64) (PoolPage, error) {
	p := PoolPage{PoolID: id, Type: PoolTypeRoundRobin}
	for _, d := range snap.Menu.Divisions {
		for _, mp := range d.Pools {
			if mp.PoolID == id {
				p.Heading = d.Name + ", " + mp.Name
			}
		}
	}
	if p.Heading == "" {
		return p, fmt.Errorf("pool %d: shows a game page, but isn't in the left menu to name it", id)
	}
	scheduled := 0
	for _, g := range snap.Schedule {
		if g.PoolID != id {
			continue
		}
		if scheduled++; scheduled > 1 {
			return p, fmt.Errorf("pool %d: shows a game page, but schedules more than one game", id)
		}
		if g.GameID != nil {
			p.GameIDs = []int64{*g.GameID}
		}
	}
	if scheduled == 0 {
		return p, fmt.Errorf("pool %d: shows a game page, but schedules no game", id)
	}
	p.LinkedGames = p.GameIDs
	return p, nil
}

// gatherTeamCard fetches t's team card. Some installs' cards 500 (2023 sites, every
// team), as does the roster view; the team scoreboard renders the same roster query
// (TeamScoreBoard), so it stands in for the card's roster then
func (c *Client) gatherTeamCard(ctx context.Context, menu Menu, t TeamListEntry) (TeamCard, error) {
	team := url.Values{"team": {strconv.FormatInt(t.TeamID, 10)}}
	doc, err := c.getView(ctx, "teamcard", team)
	if err == nil {
		return parseTeamCard(doc, t.TeamID)
	}
	var status *StatusError
	if !errors.As(err, &status) || status.StatusCode != http.StatusInternalServerError {
		return TeamCard{}, fmt.Errorf("team %d: %w", t.TeamID, err)
	}
	if doc, err = c.getView(ctx, "scorestatus", team); err != nil {
		return TeamCard{}, fmt.Errorf("team %d: card errors, and so does its scoreboard: %w", t.TeamID, err)
	}
	return parseTeamScoreboard(doc, menu, t)
}

// gatherReservations fetches every reservation the schedule uses, then every
// location those reservations are at
func (c *Client) gatherReservations(ctx context.Context, snap *Snapshot) error {
	var ids []int64
	for _, g := range snap.Schedule {
		if !slices.Contains(ids, g.ReservationID) {
			ids = append(ids, g.ReservationID)
		}
	}
	fmt.Printf("fetching %d reservations...\n", len(ids))
	snap.Reservations = make(map[int64]ReservationPage, len(ids))
	snap.Locations = map[int64]Location{}
	for i, id := range ids {
		doc, err := c.getView(ctx, "reservationinfo", url.Values{"reservation": {strconv.FormatInt(id, 10)}})
		if err != nil {
			return fmt.Errorf("reservation %d: %w", id, err)
		}
		r, err := parseReservation(doc, id)
		if err != nil {
			return err
		}
		snap.Reservations[id] = r
		liveclient.PrintProgress(i+1, len(ids), 25)

		if _, done := snap.Locations[r.LocationID]; done || r.LocationID == 0 {
			continue
		}
		body, err := c.getXML(ctx, "locationxml", url.Values{"id": {strconv.FormatInt(r.LocationID, 10)}})
		if err != nil {
			return fmt.Errorf("location %d: %w", r.LocationID, err)
		}
		locations, err := parseLocations(body)
		if err != nil {
			return fmt.Errorf("location %d: %w", r.LocationID, err)
		}
		for _, l := range locations {
			snap.Locations[l.ID] = l
		}
		if _, ok := snap.Locations[r.LocationID]; !ok {
			return fmt.Errorf("location %d: not in its own locationxml", r.LocationID)
		}
	}
	return nil
}

// GameIDs is every game id the schedule or a pool page links to, ascending
func (s *Snapshot) GameIDs() []int64 {
	var ids []int64
	for _, g := range s.Schedule {
		if g.GameID != nil && !slices.Contains(ids, *g.GameID) {
			ids = append(ids, *g.GameID)
		}
	}
	for _, p := range s.Pools {
		for _, id := range p.GameIDs {
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
	}
	slices.Sort(ids)
	return ids
}
