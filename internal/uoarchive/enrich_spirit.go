package uoarchive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
	store "github.com/cxd309/ultimate-tournament-results/internal/store/v01_09_14"
)

// liveArchive is what enrich reads from the Live! archive it fills in
type liveArchive struct {
	divisions    []store.Division
	pools        []store.Pool
	reservations []store.Reservation
	teams        []store.Team
	games        []store.Game
	spirit       []store.SpiritScore
}

// spiritChange is one spirit_scores row to write: a whole new row (score), or the comment
// of a row the Live! API served without one
type spiritChange struct {
	score   store.SpiritScore
	added   bool // insert the whole row
	comment bool // only set the comment of the existing row
}

// exportKey is what an export row has to name its game: UltiOrganizer's SpiritToCsv selects
// division name, reservation group, field name, HH:MM, pool name and the team names,
// with no ids (lib/game.functions.php)
type exportKey struct{ division, group, field, time, pool, evaluated, by string }

// matchExport lines every export row up with an archived game and the spirit_scores row for
// the team it scores, and returns what the archive is missing: rows the Live! API
// didn't serve (one-sided submissions) and comments it served empty
//
// Both come from the same UltiOrganizer database, so anything that doesn't line up
// is an error rather than skipped: a row that names no game, or more than one, two
// rows for one team, category scores or a comment that differ from what the API
// served, a total that differs from the game's homesotg/visitorsotg, or a score the
// API served that the export doesn't have
func matchExport(a liveArchive, export []map[string]string) ([]spiritChange, error) {
	divisionName := map[int64]string{}
	for _, d := range a.divisions {
		divisionName[d.SeriesID] = clean(d.Name)
	}
	poolByID := map[int64]store.Pool{}
	for _, p := range a.pools {
		poolByID[p.PoolID] = p
	}
	resByID := map[int64]store.Reservation{}
	for _, r := range a.reservations {
		resByID[r.ID] = r
	}
	teamName := map[int64]string{}
	for _, t := range a.teams {
		teamName[t.TeamID] = clean(t.Name)
	}

	type target struct {
		game    store.Game
		forHome bool
	}
	targets := map[exportKey][]target{}
	for _, g := range a.games {
		if g.Hometeam == nil || g.Visitorteam == nil || g.Reservation == nil || g.Pool == nil || len(g.Time) < 16 {
			continue // the export joins on all of these, a game missing one can't be in it
		}
		pool, res := poolByID[*g.Pool], resByID[*g.Reservation]
		k := exportKey{divisionName[convert.Int64OrZero(pool.Series)], clean(res.ReservationGroup), clean(res.FieldName), g.Time[11:16], clean(pool.Name), teamName[*g.Hometeam], teamName[*g.Visitorteam]}
		targets[k] = append(targets[k], target{g, true})
		k.evaluated, k.by = k.by, k.evaluated
		targets[k] = append(targets[k], target{g, false})
	}

	type scoreKey struct{ game, team int64 }
	served := map[scoreKey]store.SpiritScore{}
	for _, sc := range a.spirit {
		served[scoreKey{sc.GameID, sc.TeamID}] = sc
	}
	seen := map[scoreKey]bool{}

	var changes []spiritChange
	for i, row := range export {
		k := exportKey{clean(row["division"]), clean(row["day"]), clean(row["field"]), clean(row["time"]), clean(row["pool"]), clean(row["team_evaluated"]), clean(row["by_team"])}
		where := fmt.Sprintf("spirit export row %d (%s, %s %s field %s, %s by %s)", i+1, k.division, k.group, k.time, k.field, k.evaluated, k.by)
		ts := targets[k]
		switch {
		case len(ts) == 0:
			return nil, fmt.Errorf("%s matches no archived game", where)
		case len(ts) > 1:
			return nil, fmt.Errorf("%s matches %d archived games", where, len(ts))
		}
		g, forHome := ts[0].game, ts[0].forHome

		var cats [5]int64
		for c, col := range spiritColumns {
			n, err := strconv.ParseInt(strings.TrimSpace(row[col]), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%s: %s %q: %w", where, col, row[col], err)
			}
			cats[c] = n
		}
		total := cats[0] + cats[1] + cats[2] + cats[3] + cats[4]
		if t := strings.TrimSpace(row["total"]); t != strconv.FormatInt(total, 10) {
			return nil, fmt.Errorf("%s: total %q isn't the sum of its categories, %d", where, t, total)
		}

		team, sotg := *g.Visitorteam, g.Visitorsotg
		if forHome {
			team, sotg = *g.Hometeam, g.Homesotg
		}
		sk := scoreKey{g.GameID, team}
		if seen[sk] {
			return nil, fmt.Errorf("%s: a second score for team %d in game %d", where, team, g.GameID)
		}
		seen[sk] = true
		if sotg != nil && *sotg != total {
			return nil, fmt.Errorf("%s: total %d, game %d's %s is %d", where, total, g.GameID, sotgColumn(forHome), *sotg)
		}

		got := store.SpiritScore{GameID: g.GameID, TeamID: team, Cat1: cats[0], Cat2: cats[1], Cat3: cats[2], Cat4: cats[3], Cat5: cats[4], Comments: row["comments"]}
		have, ok := served[sk]
		if !ok {
			changes = append(changes, spiritChange{score: got, added: true})
			continue
		}
		if [5]int64{have.Cat1, have.Cat2, have.Cat3, have.Cat4, have.Cat5} != cats {
			return nil, fmt.Errorf("%s: scores %v, archived %v for team %d in game %d", where, cats, [5]int64{have.Cat1, have.Cat2, have.Cat3, have.Cat4, have.Cat5}, team, g.GameID)
		}
		switch {
		case have.Comments == got.Comments || got.Comments == "":
		case have.Comments == "":
			changes = append(changes, spiritChange{score: got, comment: true})
		default:
			return nil, fmt.Errorf("%s: comment %q, archived %q for team %d in game %d", where, got.Comments, have.Comments, team, g.GameID)
		}
	}
	for _, sc := range a.spirit {
		if !seen[scoreKey{sc.GameID, sc.TeamID}] {
			return nil, fmt.Errorf("game %d: archive has a spirit score for team %d the export doesn't", sc.GameID, sc.TeamID)
		}
	}
	return changes, nil
}

func sotgColumn(forHome bool) string {
	if forHome {
		return "homesotg"
	}
	return "visitorsotg"
}
