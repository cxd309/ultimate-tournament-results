package uoclient

import (
	"fmt"
	"slices"

	"golang.org/x/net/html"
)

// UltiOrganizer pool types, uo_pool.type
const (
	PoolTypeRoundRobin = 1
	PoolTypePlayoff    = 2
	PoolTypeSwissdraw  = 3
	PoolTypeCrossmatch = 4
)

// PoolPage is view=poolstatus for one pool
type PoolPage struct {
	PoolID int64
	// Heading is "{division}, {pool}"
	Heading string
	// Type is inferred from how the page renders the pool, there's no label for it:
	// round robin has the captioned PG/W/L table, Swiss draw an uncaptioned
	// border='2' table, crossmatch rows that open with two colour-coded
	// td.center cells, and a playoff is the bracket, none of those
	Type int64
	// Standings are the rows of a round-robin or Swiss-draw standings table
	Standings []PoolStanding
	// ContinuesTo are the visible pools listed under a round-robin table, each with
	// the colour its placings are coded with
	ContinuesTo []PoolContinuation
	// GameIDs are the games a round-robin pool lists under "Games", including games
	// moved in from a previous pool; nil for other types, whose pages show the whole
	// bracket across follower pools rather than this pool's games
	GameIDs []int64
	// LinkedGames is every game the page links to anywhere, any pool type
	LinkedGames []int64
	// International is whether team flags are shown, which UltiOrganizer only does
	// for an international season
	International bool
}

type PoolStanding struct {
	TeamID     int64
	Placing    int64  // 1-based row position
	Activerank *int64 // the "#" column; nil when it shows "-" (no games played yet)
	MoveColor  string // the placing's continuation colour, "" when it doesn't continue
}

type PoolContinuation struct {
	PoolID int64
	Color  string
}

func parsePool(doc *html.Node, id int64) (PoolPage, error) {
	p := PoolPage{PoolID: id}
	c := content(doc)
	if c == nil {
		return p, fmt.Errorf("pool %d: no content", id)
	}
	h2 := find(c, tag("h2"))
	if h2 == nil {
		return p, fmt.Errorf("pool %d: no heading", id)
	}
	p.Heading = text(h2)
	p.International = flagFile(c) != ""

	var table *html.Node
	for n := h2.NextSibling; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == "table" {
			table = n
			break
		}
	}
	if table == nil {
		return p, fmt.Errorf("pool %d: no table after the heading", id)
	}

	switch {
	case find(table, tag("caption")) != nil:
		p.Type = PoolTypeRoundRobin
	case attr(table, "border") == "2":
		p.Type = PoolTypeSwissdraw
	case isCrossmatch(table):
		p.Type = PoolTypeCrossmatch
	default:
		p.Type = PoolTypePlayoff
	}

	if p.Type == PoolTypeRoundRobin || p.Type == PoolTypeSwissdraw {
		for _, tr := range rows(table) {
			tds := children(tr, "td")
			if len(tds) < 2 {
				continue
			}
			team, ok := firstLinkID(tds[1], "teamcard", "team")
			if !ok {
				continue
			}
			s := PoolStanding{TeamID: team, Placing: int64(len(p.Standings) + 1), Activerank: optionalInt(text(tds[0]))}
			if div := find(tds[0], tag("div")); div != nil {
				s.MoveColor = backgroundColor(div)
			}
			p.Standings = append(p.Standings, s)
		}
	}

	for _, a := range links(c, "gameplay") {
		if id, ok := linkID(a, "game"); ok && !slices.Contains(p.LinkedGames, id) {
			p.LinkedGames = append(p.LinkedGames, id)
		}
	}
	if p.Type == PoolTypeRoundRobin {
		p.ContinuesTo = parseContinuesTo(c)
		p.GameIDs = parsePoolGames(c)
	}
	return p, nil
}

func isCrossmatch(table *html.Node) bool {
	trs := rows(table)
	if len(trs) == 0 {
		return false
	}
	tds := children(trs[0], "td")
	return len(tds) > 0 && hasClass(tds[0], "center")
}

// parseContinuesTo reads the "Continues to:" row under a round-robin table
// a continuation pool that isn't visible is named without a link, so has no id here
func parseContinuesTo(c *html.Node) []PoolContinuation {
	label := find(c, func(n *html.Node) bool { return n.Data == "td" && text(n) == "Continues to:" })
	if label == nil {
		return nil
	}
	var pools []PoolContinuation
	for td := label.NextSibling; td != nil; td = td.NextSibling {
		if td.Type != html.ElementNode || td.Data != "td" {
			continue
		}
		if id, ok := firstLinkID(td, "poolstatus", "pool"); ok {
			pools = append(pools, PoolContinuation{PoolID: id, Color: backgroundColor(td)})
		}
	}
	return pools
}

// parsePoolGames reads the game ids of the table under a round-robin pool's
// "Games" heading
func parsePoolGames(c *html.Node) []int64 {
	h2 := find(c, func(n *html.Node) bool { return n.Data == "h2" && text(n) == "Games" })
	if h2 == nil {
		return nil
	}
	ids := []int64{}
	for n := h2.NextSibling; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode || n.Data != "table" {
			continue
		}
		for _, a := range links(n, "gameplay") {
			if id, ok := linkID(a, "game"); ok && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		break
	}
	return ids
}
