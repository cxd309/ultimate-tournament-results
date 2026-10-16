package uoarchive

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strings"

	uostore "github.com/cxd309/ultimate-tournament-results/internal/store/uo_01"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

// EnrichOptions picks the UltiOrganizer site to enrich from, every field optional:
// Host and BasePath default to the archive's own deployment with its Live! base path
// (live/data/) taken off, SeasonID to the archive's season. CacheDir keeps fetched
// pages, as for Archive
type EnrichOptions struct {
	Host     string
	BasePath string
	SeasonID string
	CacheDir string
}

// Enrich turns the 1.9.14-1.9.17 Live! archive at dbPath into a uo_01 one by
// scraping the UltiOrganizer site under the deployment, as Archive does, and adding
// what Live! has no place for: db/uo_01/schema.sql's tables, location addresses and
// coordinates, reservation times, and the spirit the Live! API leaves out
//
// The Live! archive stays the base: it has every game, including a forfeit, which
// UltiOrganizer never links to a game id. Its rows are left as served, apart from
// spirit_scores gaining one-sided scores and comments, so the published JSON only
// changes there, and only shows comments when the server did (server_settings).
// Everything the scrape shares with the archive (games, teams, players, goals,
// spirit) is checked against it first, and any disagreement stops enrich before it
// writes anything
func Enrich(ctx context.Context, dbPath string, opts EnrichOptions) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	s, uo := store.New(db), uostore.New(db)
	if _, ok, err := uo.GetServerSettings(ctx); err != nil {
		return err
	} else if ok {
		return fmt.Errorf("%s is already a uo_01 archive, enriched or archived from UltiOrganizer", dbPath)
	}
	t, err := s.GetTournament(ctx)
	if err != nil {
		return fmt.Errorf("tournament: %w", err)
	}
	if opts, err = opts.withDefaults(t.Host, t.BasePath, t.SeasonID); err != nil {
		return err
	}

	c := uoclient.NewClient(opts.Host, opts.BasePath)
	c.CacheDir = opts.CacheDir
	snap, err := c.Gather(ctx, opts.SeasonID)
	if err != nil {
		return fmt.Errorf("gather: %w", err)
	}
	// the export is matched against the Live! archive below, which has the games it
	// names that resolve can't place (forfeits); resolve takes spirit from game pages
	export := snap.SpiritCSV
	snap.SpiritCSV = nil
	r, err := resolve(snap, opts.Host, opts.BasePath, true)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	live, err := loadLive(ctx, s)
	if err != nil {
		return err
	}
	if err := checkShared(r, live); err != nil {
		return err
	}
	var spirit []spiritChange
	if export != nil {
		spirit, err = matchExport(live.liveArchive, export)
	} else {
		spirit, err = pageSpirit(r.spiritScores, live.spirit)
	}
	if err != nil {
		return fmt.Errorf("spirit: %w", err)
	}
	settings := r.serverSettings
	for _, sc := range live.spirit {
		if sc.Comments != "" {
			// the Live! API only serves comments when the server shows them
			show := true
			settings.ShowSpiritComments = &show
			break
		}
	}

	schema, err := os.ReadFile("db/uo_01/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("apply schema db/uo_01/schema.sql: %w", err)
	}
	ts, tuo := store.New(tx), uostore.New(tx)
	if err := tuo.InsertServerSettings(ctx, settings); err != nil {
		return fmt.Errorf("server settings: %w", err)
	}
	if err := runSteps([]step{
		{"locations", each(r.locations, func(v uostore.Location) error { return tuo.UpdateLocation(ctx, v) })},
		{"reservations", each(r.reservations, func(v uostore.Reservation) error { return tuo.UpdateReservation(ctx, v) })},
		{"spirit scores", each(spirit, func(v spiritChange) error {
			if v.added {
				return ts.InsertSpiritScore(ctx, v.score)
			}
			return tuo.SetSpiritComments(ctx, v.score.GameID, v.score.TeamID, v.score.Comments)
		})},
	}); err != nil {
		return err
	}
	if err := writeExtras(ctx, tuo, r); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	var scores, comments int
	for _, v := range spirit {
		if v.added {
			scores++
		}
		if v.score.Comments != "" {
			comments++
		}
	}
	fmt.Printf("enriched %s: %d played, %d game events, %d team profiles, %d urls, %d spirit scores and %d comments added; %d games with no scoresheet kept from Live!\n",
		dbPath, len(r.played), len(r.gameEvents), len(r.teamProfiles), len(r.urls), scores, comments, len(r.unlinked))
	return nil
}

// EnrichUnsupported is Enrich for an archive version it can't fill in yet (3.0.6):
// a note, and nothing written
func EnrichUnsupported(_ context.Context, dbPath string, _ EnrichOptions) error {
	fmt.Printf("enriching this version's archives isn't supported yet, %s left as it is\n", dbPath)
	return nil
}

// live is the Live! archive Enrich fills in
type live struct {
	liveArchive
	locations []store.Location
	players   []store.Player
	goals     []store.Goal
}

func loadLive(ctx context.Context, s *store.Store) (live, error) {
	var l live
	var err error
	if l.divisions, err = s.ListDivisions(ctx); err != nil {
		return l, fmt.Errorf("divisions: %w", err)
	}
	if l.pools, err = s.ListPools(ctx); err != nil {
		return l, fmt.Errorf("pools: %w", err)
	}
	if l.locations, err = s.ListLocations(ctx); err != nil {
		return l, fmt.Errorf("locations: %w", err)
	}
	if l.reservations, err = s.ListReservations(ctx); err != nil {
		return l, fmt.Errorf("reservations: %w", err)
	}
	if l.teams, err = s.ListTeams(ctx); err != nil {
		return l, fmt.Errorf("teams: %w", err)
	}
	if l.players, err = s.ListPlayers(ctx); err != nil {
		return l, fmt.Errorf("players: %w", err)
	}
	if l.games, err = s.ListGames(ctx); err != nil {
		return l, fmt.Errorf("games: %w", err)
	}
	if l.goals, err = s.ListGoals(ctx); err != nil {
		return l, fmt.Errorf("goals: %w", err)
	}
	if l.spirit, err = s.ListSpiritScores(ctx); err != nil {
		return l, fmt.Errorf("spirit scores: %w", err)
	}
	return l, nil
}

// checkShared checks the scrape agrees with the Live! archive on everything they
// both have, and that every row enrich adds refers to something the archive has
//
// The archive can have games the scrape doesn't, only as many as it found scheduled
// with no game id, and only ones with no goals: a forfeit, or a result entered with
// no scoresheet
func checkShared(r *rows, l live) error {
	liveRes := map[int64]store.Reservation{}
	for _, res := range l.reservations {
		liveRes[res.ID] = res
	}
	for _, res := range r.reservations {
		if lr := liveRes[res.ID]; lr.Location != res.Location {
			return fmt.Errorf("reservation %d is at location %d on UltiOrganizer, %d in the Live! archive", res.ID, res.Location, lr.Location)
		}
	}

	poolID := func(p store.Pool) int64 { return p.PoolID }
	for _, c := range []struct {
		name          string
		live, scraped []int64
	}{
		{"pool", mapIDs(l.pools, poolID), mapIDs(r.pools, poolID)},
		{"location", mapIDs(l.locations, func(v store.Location) int64 { return v.ID }), mapIDs(r.locations, func(v uostore.Location) int64 { return v.ID })},
		{"reservation", mapIDs(l.reservations, func(v store.Reservation) int64 { return v.ID }), mapIDs(r.reservations, func(v uostore.Reservation) int64 { return v.ID })},
	} {
		for _, id := range c.scraped {
			if !slices.Contains(c.live, id) {
				return fmt.Errorf("UltiOrganizer %s %d isn't in the Live! archive", c.name, id)
			}
		}
	}

	teams := map[int64]store.Team{}
	for _, t := range l.teams {
		teams[t.TeamID] = t
	}
	for _, t := range r.teams {
		lt, ok := teams[t.TeamID]
		if !ok {
			return fmt.Errorf("UltiOrganizer team %d (%s) isn't in the Live! archive", t.TeamID, t.Name)
		}
		if clean(lt.Name) != clean(t.Name) {
			return fmt.Errorf("team %d is %q on UltiOrganizer, %q in the Live! archive", t.TeamID, t.Name, lt.Name)
		}
	}
	players := map[int64]store.Player{}
	for _, p := range l.players {
		players[p.PlayerID] = p
	}
	for _, p := range r.players {
		lp, ok := players[p.PlayerID]
		if !ok {
			return fmt.Errorf("UltiOrganizer player %d (%s %s) isn't in the Live! archive", p.PlayerID, p.FirstName, p.LastName)
		}
		if lp.Team != p.Team {
			return fmt.Errorf("player %d is on team %d on UltiOrganizer, %d in the Live! archive", p.PlayerID, p.Team, lp.Team)
		}
	}

	games := map[int64]store.Game{}
	for _, g := range l.games {
		games[g.GameID] = g
	}
	goals := func(gs []store.Goal) map[int64][]store.Goal {
		m := map[int64][]store.Goal{}
		for _, g := range gs {
			m[g.GameID] = append(m[g.GameID], g)
		}
		return m
	}
	liveGoals, uoGoals := goals(l.goals), goals(r.goals)
	scraped := map[int64]bool{}
	for _, g := range r.games {
		scraped[g.GameID] = true
		lg, ok := games[g.GameID]
		if !ok {
			return fmt.Errorf("UltiOrganizer game %d isn't in the Live! archive", g.GameID)
		}
		for _, f := range []struct {
			name     string
			uo, live *int64
		}{
			{"home team", g.Hometeam, lg.Hometeam},
			{"visitor team", g.Visitorteam, lg.Visitorteam},
			{"home score", g.Homescore, lg.Homescore},
			{"visitor score", g.Visitorscore, lg.Visitorscore},
			{"pool", g.Pool, lg.Pool},
			{"reservation", g.Reservation, lg.Reservation},
		} {
			if !equalPtr(f.uo, f.live) {
				return fmt.Errorf("game %d: %s is %s on UltiOrganizer, %s in the Live! archive", g.GameID, f.name, showPtr(f.uo), showPtr(f.live))
			}
		}
		if g.Time != lg.Time {
			return fmt.Errorf("game %d: time is %q on UltiOrganizer, %q in the Live! archive", g.GameID, g.Time, lg.Time)
		}
		if err := checkGoals(g.GameID, uoGoals[g.GameID], liveGoals[g.GameID]); err != nil {
			return err
		}
	}
	var kept []string
	for _, g := range l.games {
		if scraped[g.GameID] {
			continue
		}
		if len(liveGoals[g.GameID]) > 0 {
			return fmt.Errorf("game %d has goals in the Live! archive but UltiOrganizer doesn't link it", g.GameID)
		}
		kept = append(kept, fmt.Sprint(g.GameID))
	}
	if len(kept) != len(r.unlinked) {
		return fmt.Errorf("the Live! archive has %d games UltiOrganizer doesn't link (%s), UltiOrganizer schedules %d with no game id", len(kept), strings.Join(kept, ", "), len(r.unlinked))
	}
	return nil
}

// checkGoals checks one game's scoresheet matches, goal by goal in order: Live!
// numbers goals from 0, uo_01 from 1. Time is left out, the Live! API and the game
// page round it differently
func checkGoals(gameID int64, uo, live []store.Goal) error {
	if len(uo) != len(live) {
		return fmt.Errorf("game %d: %d goals on UltiOrganizer, %d in the Live! archive", gameID, len(uo), len(live))
	}
	show := func(g store.Goal) string {
		return fmt.Sprintf("%s-%s scored by %s assisted by %s (home %t, callahan %t)",
			showPtr(g.Homescore), showPtr(g.Visitorscore), showPtr(g.Scorer), showPtr(g.Assist), g.Ishomegoal, g.Iscallahan)
	}
	for i := range uo {
		if a, b := show(uo[i]), show(live[i]); a != b {
			return fmt.Errorf("game %d goal %d: %s on UltiOrganizer, %s in the Live! archive", gameID, i+1, a, b)
		}
	}
	return nil
}

// pageSpirit is spirit from game pages for a site whose export is locked: both teams'
// scores, with comments where the server shows them. The Live! archive has every one
// already, so only comments it served empty can be new
func pageSpirit(uo, live []store.SpiritScore) ([]spiritChange, error) {
	type k struct{ game, team int64 }
	have := map[k]store.SpiritScore{}
	for _, sc := range live {
		have[k{sc.GameID, sc.TeamID}] = sc
	}
	var changes []spiritChange
	for _, sc := range uo {
		lsc, ok := have[k{sc.GameID, sc.TeamID}]
		if !ok {
			changes = append(changes, spiritChange{score: sc, added: true})
			continue
		}
		if [5]int64{sc.Cat1, sc.Cat2, sc.Cat3, sc.Cat4, sc.Cat5} != [5]int64{lsc.Cat1, lsc.Cat2, lsc.Cat3, lsc.Cat4, lsc.Cat5} {
			return nil, fmt.Errorf("game %d team %d: game page scores differ from the Live! archive's", sc.GameID, sc.TeamID)
		}
		switch {
		case sc.Comments == lsc.Comments || sc.Comments == "":
		case lsc.Comments == "":
			changes = append(changes, spiritChange{score: sc, comment: true})
		default:
			return nil, fmt.Errorf("game %d team %d: game page comment %q, Live! archive %q", sc.GameID, sc.TeamID, sc.Comments, lsc.Comments)
		}
	}
	return changes, nil
}

// withDefaults fills empty fields from the archive's own host, Live! base path and
// season. A host can carry a path of its own (results.wfdf.sport/wucc-2026, or
// results.wfdf.sport/wwuc/), which becomes part of the base path
func (o EnrichOptions) withDefaults(host, livePath, seasonID string) (EnrichOptions, error) {
	if o.SeasonID == "" {
		o.SeasonID = seasonID
	}
	if o.Host != "" && o.BasePath == "" {
		return o, fmt.Errorf("-base-path is required with -host")
	}
	if o.Host == "" {
		o.Host = host
		if i := strings.Index(host, "/"); i >= 0 {
			o.Host, livePath = host[:i], strings.TrimSuffix(host[i:], "/")+livePath
		}
		path, ok := strings.CutSuffix(livePath, "live/data/")
		if !ok {
			return o, fmt.Errorf("archive's base path %q doesn't end live/data/, pass -host and -base-path for its UltiOrganizer site", livePath)
		}
		o.BasePath = path
	}
	return o, nil
}

func mapIDs[T any](rows []T, id func(T) int64) []int64 {
	out := make([]int64, len(rows))
	for i, row := range rows {
		out[i] = id(row)
	}
	return out
}

func equalPtr(a, b *int64) bool {
	return (a == nil) == (b == nil) && (a == nil || *a == *b)
}

func showPtr(p *int64) string {
	if p == nil {
		return "none"
	}
	return fmt.Sprint(*p)
}
