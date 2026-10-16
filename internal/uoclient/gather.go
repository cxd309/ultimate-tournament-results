package uoclient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

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
	if err := c.gatherReservations(ctx, snap); err != nil {
		return nil, err
	}

	fmt.Printf("fetching %d team cards...\n", len(snap.Teams))
	snap.TeamCards = make(map[int64]TeamCard, len(snap.Teams))
	for i, t := range snap.Teams {
		doc, err := c.getView(ctx, "teamcard", url.Values{"team": {strconv.FormatInt(t.TeamID, 10)}})
		if err != nil {
			return nil, fmt.Errorf("team %d: %w", t.TeamID, err)
		}
		card, err := parseTeamCard(doc, t.TeamID)
		if err != nil {
			return nil, err
		}
		snap.TeamCards[t.TeamID] = card
		liveclient.PrintProgress(i+1, len(snap.Teams), 10)
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
		pool, err := parsePool(doc, id)
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
