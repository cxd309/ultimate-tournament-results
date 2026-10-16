package uoclient

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// TeamCard is view=teamcard for one team
type TeamCard struct {
	TeamID      int64
	Heading     string // "{team} ({division})"
	SeriesID    *int64 // from the "Division:" link
	CountryID   *int64
	CountryName string
	FlagFile    string
	ClubID      *int64
	ClubName    string
	Profile     *TeamProfile // nil when the card shows no profile at all
	URLs        []URL
	Roster      []RosterPlayer
	// FromScoreboard is a card read from view=scorestatus because the card itself
	// errors: the roster without shirt numbers, and no profile, club or links
	FromScoreboard bool
}

// TeamProfile is the profile block a team card can show under its heading
type TeamProfile struct {
	Coach        string
	Captain      string
	Story        string
	Achievements string
	ProfileImage string // filename only
}

// URL is one link row from a "Team pages" or "Photos and Videos" list
type URL struct {
	Type        string // link icon name, images/linkicons/{type}.png
	Name        string // "" when the link text is just the url
	URL         string
	IsMediaLink bool
	MediaOwner  string // the "from ..." credit on a media link
}

// parseTeamScoreboard reads view=scorestatus for one team in place of its card: one row
// per player on the team, with games played. Division and country are the teams
// list's, the series id the left menu's division of that name
func parseTeamScoreboard(doc *html.Node, menu Menu, t TeamListEntry) (TeamCard, error) {
	card := TeamCard{TeamID: t.TeamID, CountryID: t.CountryID, CountryName: t.CountryName, FlagFile: t.FlagFile, FromScoreboard: true}
	for _, d := range menu.Divisions {
		if d.Name == t.Division {
			if card.SeriesID != nil {
				return card, fmt.Errorf("team %d: left menu has two %q divisions", t.TeamID, t.Division)
			}
			card.SeriesID = &d.SeriesID
		}
	}
	if card.SeriesID == nil {
		return card, fmt.Errorf("team %d: no %q division in the left menu", t.TeamID, t.Division)
	}
	c := content(doc)
	if c == nil {
		return card, fmt.Errorf("team %d scoreboard: no content", t.TeamID)
	}
	table := find(c, tag("table"))
	if table == nil {
		return card, fmt.Errorf("team %d scoreboard: no table", t.TeamID)
	}
	for _, tr := range rows(table) {
		// #, player, team, games, ...
		tds := children(tr, "td")
		if len(tds) < 4 {
			continue
		}
		a := first(links(tds[1], "playercard"))
		if a == nil {
			return card, fmt.Errorf("team %d scoreboard: row %q has no player link", t.TeamID, text(tr))
		}
		id, ok := linkID(a, "player")
		if !ok {
			return card, fmt.Errorf("team %d scoreboard: player link without an id: %q", t.TeamID, text(a))
		}
		if team := text(tds[2]); team != t.Name {
			return card, fmt.Errorf("team %d scoreboard: player %d listed for %q", t.TeamID, id, team)
		}
		card.Roster = append(card.Roster, RosterPlayer{PlayerID: id, Name: text(a), Games: optionalInt(text(tds[3]))})
	}
	return card, nil
}

// RosterPlayer is one row of a team card's season roster
type RosterPlayer struct {
	PlayerID int64
	Num      *int64
	Name     string // "{first} {last}", a single string on this page
	Games    *int64
}

func parseTeamCard(doc *html.Node, id int64) (TeamCard, error) {
	t := TeamCard{TeamID: id}
	c := content(doc)
	if c == nil {
		return t, fmt.Errorf("team %d: no content", id)
	}
	if h1 := find(c, tag("h1")); h1 != nil {
		t.Heading = text(h1)
	}

	for _, a := range findAll(c, tagClass("a", "headerlink")) {
		params := linkParams(attr(a, "href"))
		switch params.Get("view") {
		case "countrycard":
			if cid, ok := linkID(a, "country"); ok {
				t.CountryID = &cid
				t.CountryName = text(a)
				t.FlagFile = flagFile(a.Parent)
			}
		case "clubcard":
			if cid, ok := linkID(a, "club"); ok {
				t.ClubID = &cid
				t.ClubName = text(a)
			}
		}
	}

	for _, p := range children(c, "p") {
		if strings.HasPrefix(text(p), "Division:") {
			if sid, ok := firstLinkID(p, "poolstatus", "series"); ok {
				t.SeriesID = &sid
			}
		}
	}

	// the profile and link tables all come before the roster heading, the
	// tables after it are the roster itself and the team's games
	for _, table := range tablesBeforeRoster(c) {
		switch header := profileHeader(table); header {
		case "Team pages:":
			t.URLs = append(t.URLs, parseURLRows(table, false)...)
		case "Photos and Videos:":
			t.URLs = append(t.URLs, parseURLRows(table, true)...)
		default:
			if profile := parseTeamProfile(table, id); profile != nil {
				t.Profile = profile
			}
		}
	}

	roster, err := parseRoster(c)
	if err != nil {
		return t, fmt.Errorf("team %d: %w", id, err)
	}
	t.Roster = roster
	return t, nil
}

// tablesBeforeRoster is content's direct table children up to the roster heading
func tablesBeforeRoster(c *html.Node) []*html.Node {
	var tables []*html.Node
	for n := c.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode {
			continue
		}
		if n.Data == "p" && find(n, func(s *html.Node) bool {
			return s.Data == "span" && hasClass(s, "profileheader") && strings.HasSuffix(text(s), "roster:")
		}) != nil {
			break
		}
		if n.Data == "table" {
			tables = append(tables, n)
		}
	}
	return tables
}

// profileHeader is the text of a table's first td.profileheader, "" if it has none
func profileHeader(table *html.Node) string {
	if td := find(table, tagClass("td", "profileheader")); td != nil {
		return text(td)
	}
	return ""
}

// parseTeamProfile reads the profile table: "Coach:"/"Captain:" label rows, the story
// as a full-width row before any "Achievements:" heading, achievements as the
// full-width row after it, and an optional profile image row
// nil for a table that isn't the profile (no labels and no image)
func parseTeamProfile(table *html.Node, teamID int64) *TeamProfile {
	p := &TeamProfile{}
	found := false
	afterAchievements := false
	imagePrefix := fmt.Sprintf("teams/%d/", teamID)
	for _, tr := range rows(table) {
		tds := children(tr, "td")
		if len(tds) == 0 {
			continue
		}
		label := ""
		if hasClass(tds[0], "profileheader") {
			label = text(tds[0])
		}
		switch {
		case label == "Coach:" && len(tds) > 1:
			p.Coach, found = text(tds[1]), true
		case label == "Captain:" && len(tds) > 1:
			p.Captain, found = text(tds[1]), true
		case label == "Achievements:":
			afterAchievements, found = true, true
		case label == "":
			if a := find(tds[0], tag("a")); a != nil {
				if i := strings.Index(attr(a, "href"), imagePrefix); i >= 0 {
					p.ProfileImage, found = attr(a, "href")[i+len(imagePrefix):], true
					continue
				}
			}
			body := multilineText(tds[0])
			if body == "" {
				continue
			}
			if afterAchievements {
				p.Achievements = body
			} else {
				p.Story = body
			}
			found = true
		}
	}
	if !found {
		return nil
	}
	return p
}

// multilineText is n's text with each <br/> as a line break, every line cleaned
func multilineText(n *html.Node) string {
	var b strings.Builder
	walk(n, func(c *html.Node) bool {
		switch {
		case c.Type == html.TextNode:
			b.WriteString(c.Data)
		case c.Type == html.ElementNode && c.Data == "br":
			b.WriteString("\n")
		}
		return true
	})
	lines := strings.Split(b.String(), "\n")
	for i, line := range lines {
		lines[i] = clean(line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// parseURLRows reads a links table: an images/linkicons/{type}.png icon then the
// link, named by its text, followed on a media link by " from {owner}"
func parseURLRows(table *html.Node, media bool) []URL {
	var urls []URL
	for _, tr := range rows(table) {
		a := find(tr, func(n *html.Node) bool { return n.Data == "a" && attr(n, "href") != "" })
		if a == nil {
			continue
		}
		u := URL{URL: attr(a, "href"), IsMediaLink: media}
		if name := text(a); name != u.URL {
			u.Name = name
		}
		if img := find(tr, tag("img")); img != nil {
			u.Type = attr(img, "alt")
		}
		if media {
			if owner, ok := strings.CutPrefix(clean(textAfter(a)), "from "); ok {
				u.MediaOwner = owner
			}
		}
		urls = append(urls, u)
	}
	return urls
}

// textAfter is the text of n's following siblings within its parent
func textAfter(n *html.Node) string {
	var b strings.Builder
	for s := n.NextSibling; s != nil; s = s.NextSibling {
		b.WriteString(rawText(s))
	}
	return b.String()
}

// parseRoster reads the table that follows the "{season} roster:" heading
func parseRoster(c *html.Node) ([]RosterPlayer, error) {
	heading := find(c, func(n *html.Node) bool {
		return n.Data == "span" && hasClass(n, "profileheader") && strings.HasSuffix(text(n), "roster:")
	})
	if heading == nil {
		// a team with no players entered has no roster heading at all
		return nil, nil
	}
	var table *html.Node
	for n := heading.Parent.NextSibling; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == "table" {
			table = n
			break
		}
	}
	if table == nil {
		return nil, fmt.Errorf("roster heading with no table")
	}
	var roster []RosterPlayer
	for _, tr := range rows(table) {
		tds := children(tr, "td")
		if len(tds) < 3 {
			continue
		}
		a := first(links(tds[1], "playercard"))
		if a == nil {
			continue
		}
		id, ok := linkID(a, "player")
		if !ok {
			continue
		}
		roster = append(roster, RosterPlayer{
			PlayerID: id,
			Num:      optionalInt(text(tds[0])),
			Name:     text(a),
			Games:    optionalInt(text(tds[2])),
		})
	}
	return roster, nil
}
