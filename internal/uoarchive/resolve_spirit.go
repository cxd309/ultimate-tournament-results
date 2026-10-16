package uoarchive

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
	"github.com/cxd309/ultimate-tournament-results/internal/uoclient"
)

// spiritColumns are the spirit export's category columns, cat1 to cat5
var spiritColumns = [5]string{"rules", "fouls", "fair", "positive", "com"}

// resolveSpirit fills spirit_scores and each game's homesotg/visitorsotg
//
// The spirit export is the source when the site serves it: it has every submitted
// score with its comments, including a game where only one team has submitted,
// which the game page hides until both have so neither team sees its score first.
// Each row is matched to a game by division, reservation group, field, start time,
// pool and the two team names, and checked against the game page's scores where the
// page shows them
//
// When the site has locked the export down, the game pages are all there is: only
// games where both teams submitted, with comments only if the season shows them
func (r *rows) resolveSpirit(snap *uoclient.Snapshot) error {
	gameIdx := make(map[int64]int, len(r.games))
	for i, g := range r.games {
		gameIdx[g.GameID] = i
	}

	if snap.SpiritCSV == nil {
		for _, id := range snap.GameIDs() {
			page := snap.Games[id]
			if page.Spirit == nil {
				continue
			}
			if err := r.addSpirit(gameIdx, id, true, page.Spirit.Home, page.Spirit.HomeComments); err != nil {
				return err
			}
			if err := r.addSpirit(gameIdx, id, false, page.Spirit.Visitor, page.Spirit.VisitorComments); err != nil {
				return err
			}
		}
		return nil
	}

	type key struct{ division, group, field, time, pool, evaluated, by string }
	gameFor := map[key]int64{}
	homeName := map[int64]string{}
	for _, g := range snap.Schedule {
		if g.GameID == nil {
			continue
		}
		k := key{g.Division, g.ReservationGroup, strings.TrimPrefix(g.FieldLabel, "Field "), g.Time, g.PoolName, g.Home, g.Visitor}
		gameFor[k] = *g.GameID
		k.evaluated, k.by = g.Visitor, g.Home
		gameFor[k] = *g.GameID
		homeName[*g.GameID] = g.Home
	}

	for i, row := range snap.SpiritCSV {
		k := key{clean(row["division"]), clean(row["day"]), clean(row["field"]), clean(row["time"]), clean(row["pool"]), clean(row["team_evaluated"]), clean(row["by_team"])}
		id, ok := gameFor[k]
		if !ok {
			return fmt.Errorf("spirit export row %d (%s, %s %s field %s, %s by %s) matches no scheduled game", i+1, k.division, k.group, k.time, k.field, k.evaluated, k.by)
		}
		var cats [5]int64
		for c, col := range spiritColumns {
			n, err := strconv.ParseInt(strings.TrimSpace(row[col]), 10, 64)
			if err != nil {
				return fmt.Errorf("spirit export row %d: %s %q: %w", i+1, col, row[col], err)
			}
			cats[c] = n
		}
		forHome := k.evaluated == homeName[id]
		if page := snap.Games[id].Spirit; page != nil {
			shown := page.Visitor
			if forHome {
				shown = page.Home
			}
			if shown != cats {
				return fmt.Errorf("game %d: spirit export gives %v for %s, game page %v", id, cats, k.evaluated, shown)
			}
		}
		comments := row["comments"]
		if err := r.addSpirit(gameIdx, id, forHome, cats, &comments); err != nil {
			return err
		}
	}
	for _, id := range snap.GameIDs() {
		if snap.Games[id].Spirit != nil {
			if g := r.games[gameIdx[id]]; g.Homesotg == nil || g.Visitorsotg == nil {
				return fmt.Errorf("game %d: game page shows spirit the export doesn't have", id)
			}
		}
	}
	slices.SortFunc(r.spiritScores, func(a, b store.SpiritScore) int {
		return cmp.Or(cmp.Compare(a.GameID, b.GameID), cmp.Compare(a.TeamID, b.TeamID))
	})
	return nil
}

// addSpirit records the score one team of a game received, and its total as that
// side's homesotg/visitorsotg; a second score for the same team is an error
func (r *rows) addSpirit(gameIdx map[int64]int, id int64, forHome bool, cats [5]int64, comments *string) error {
	game := &r.games[gameIdx[id]]
	team, total := game.Visitorteam, &game.Visitorsotg
	if forHome {
		team, total = game.Hometeam, &game.Homesotg
	}
	if team == nil {
		return fmt.Errorf("game %d: spirit for an unresolved team slot", id)
	}
	if *total != nil {
		return fmt.Errorf("game %d: two spirit scores for team %d", id, *team)
	}
	t := sum(cats)
	*total = &t
	r.spiritScores = append(r.spiritScores, spiritScore(id, *team, cats, comments))
	return nil
}

// showSpiritComments reads the server's ShowSpiritComments from the game pages: one
// with both spirit scores has a "Spirit comments" row exactly when it's on. nil when
// no page shows spirit at all
func showSpiritComments(snap *uoclient.Snapshot) *bool {
	var show *bool
	for _, page := range snap.Games {
		if page.Spirit == nil {
			continue
		}
		shown := page.Spirit.HomeComments != nil
		if show != nil && *show != shown {
			return nil // can't happen with one server setting, so don't guess
		}
		show = &shown
	}
	return show
}
