package uoclient

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Menu is the left menu every season page carries: the season, its divisions, and
// each division's visible pools, in the order the site lists them
type Menu struct {
	SeasonID   string
	SeasonName string
	Divisions  []MenuDivision
}

type MenuDivision struct {
	SeriesID int64
	Name     string
	Pools    []MenuPool
}

type MenuPool struct {
	PoolID int64
	Name   string
}

func parseMenu(doc *html.Node) (Menu, error) {
	var menu Menu
	// id on most sites, class on some custom templates (EBUCC 2023's cust/ebucc2022)
	left := find(doc, func(n *html.Node) bool { return attr(n, "id") == "menu_left" || hasClass(n, "menu_left") })
	if left == nil {
		return menu, fmt.Errorf("no left menu")
	}

	if a := find(left, tagClass("a", "seasonnav")); a != nil {
		menu.SeasonID = linkParams(attr(a, "href")).Get("season")
		menu.SeasonName = text(a)
	}
	if menu.SeasonID == "" {
		return menu, fmt.Errorf("left menu has no season link")
	}

	walk(left, func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "td" {
			return true
		}
		switch {
		default:
			return n == left
		case hasClass(n, "menuserieslevel"):
			for _, a := range links(n, "seriesstatus") {
				if id, ok := linkID(a, "series"); ok {
					menu.Divisions = append(menu.Divisions, MenuDivision{SeriesID: id, Name: text(a)})
				}
			}
		case hasClass(n, "menupoollevel") && len(menu.Divisions) > 0:
			d := &menu.Divisions[len(menu.Divisions)-1]
			for _, a := range links(n, "poolstatus") {
				if id, ok := linkID(a, "pool"); ok {
					d.Pools = append(d.Pools, MenuPool{PoolID: id, Name: strings.TrimSpace(strings.TrimPrefix(text(a), "»"))})
				}
			}
		}
		return false
	})
	return menu, nil
}

// parseSeasonList returns every season id the season list page links to, in page order
func parseSeasonList(doc *html.Node) []string {
	var ids []string
	seen := map[string]bool{}
	for _, a := range links(content(doc), "teams") {
		id := linkParams(attr(a, "href")).Get("season")
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// TeamListEntry is one row of the season's teams page (list=allteams)
type TeamListEntry struct {
	TeamID      int64
	Name        string
	Division    string // the division heading the row is listed under
	CountryID   *int64
	CountryName string
	FlagFile    string // images/flags/tiny/ filename, "" when none is shown
}

func parseAllTeams(doc *html.Node) ([]TeamListEntry, error) {
	c := content(doc)
	if c == nil {
		return nil, fmt.Errorf("no content")
	}
	var teams []TeamListEntry
	division := ""
	for _, table := range findAll(c, tag("table")) {
		for _, tr := range rows(table) {
			if th := find(tr, tag("th")); th != nil {
				division = text(th)
				continue
			}
			a := first(links(tr, "teamcard"))
			if a == nil {
				continue
			}
			id, ok := linkID(a, "team")
			if !ok {
				continue
			}
			entry := TeamListEntry{TeamID: id, Name: text(a), Division: division}
			if ca := first(links(tr, "countrycard")); ca != nil {
				if cid, ok := linkID(ca, "country"); ok {
					entry.CountryID = &cid
					entry.CountryName = text(ca)
				}
			}
			entry.FlagFile = flagFile(tr)
			teams = append(teams, entry)
		}
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams listed")
	}
	return teams, nil
}

// parseSeeds reads list=byseeding: each team's "N." seed within its division
func parseSeeds(doc *html.Node) map[int64]int64 {
	seeds := map[int64]int64{}
	for _, tr := range findAll(content(doc), tag("tr")) {
		tds := children(tr, "td")
		if len(tds) < 2 {
			continue
		}
		id, ok := firstLinkID(tds[1], "teamcard", "team")
		if !ok {
			continue
		}
		if seed, ok := parseInt(strings.TrimSuffix(text(tds[0]), ".")); ok {
			seeds[id] = seed
		}
	}
	return seeds
}

// parseStandings reads list=bystandings, the placement table with one column per
// division: the team in row N of a division's column finished Nth
// a tie lists several teams in one cell, they all get that row's placing
// false when the page isn't the computed table (a season can replace it with a
// hand-written CSV, which carries no team ids)
func parseStandings(doc *html.Node) (map[int64]int64, bool) {
	c := content(doc)
	if c == nil {
		return nil, false
	}
	for _, table := range findAll(c, tag("table")) {
		trs := rows(table)
		if len(trs) == 0 {
			continue
		}
		ths := children(trs[0], "th")
		if len(ths) < 2 || text(ths[0]) != "Placement" {
			continue
		}
		standings := map[int64]int64{}
		for i, tr := range trs[1:] {
			tds := children(tr, "td")
			if len(tds) < 2 {
				continue
			}
			for _, td := range tds[1:] {
				for _, a := range links(td, "teamcard") {
					if id, ok := linkID(a, "team"); ok {
						standings[id] = int64(i + 1)
					}
				}
			}
		}
		return standings, true
	}
	return nil, false
}

// flagFile is the filename of the first images/flags/tiny/ flag under n
func flagFile(n *html.Node) string {
	for _, img := range findAll(n, tag("img")) {
		if f, ok := strings.CutPrefix(attr(img, "src"), "images/flags/tiny/"); ok {
			return f
		}
	}
	return ""
}

func first(nodes []*html.Node) *html.Node {
	if len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}
