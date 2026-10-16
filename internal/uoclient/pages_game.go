package uoclient

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// GamePage is view=gameplay for one game
type GamePage struct {
	GameID        int64
	HomeTeamID    *int64 // from the "Game history" link, only rendered for a game with a result
	VisitorTeamID *int64
	HomeScore     *int64 // from the heading, nil for a game with no result
	VisitorScore  *int64
	// HasScoresheet is false for a game with a result but no goals entered, which
	// the page renders as "Not fed in" with no scoreboards or scoresheet
	HasScoresheet bool
	Home          []ScoreboardPlayer
	Visitor       []ScoreboardPlayer
	Goals         []GoalRow
	Halftime      *int64 // seconds, from the timeline's halftime marker
	Events        []EventRow
	Official      string
	Spirit        *GameSpirit
	URLs          []URL
}

// ScoreboardPlayer is one row of a team's scoreboard on the game page: a player
// entered as having played this game
type ScoreboardPlayer struct {
	PlayerID  int64
	Num       *int64
	FirstName string // the link text is "{first}&nbsp;{last}", so the split is exact
	LastName  string
	Captain   bool
	Assists   int64
	Goals     int64
}

// GoalRow is one row of the scoresheet
type GoalRow struct {
	HomeScore    int64
	VisitorScore int64
	IsHome       bool
	IsCallahan   bool
	Assist       PlayerRef // empty on a Callahan goal
	Scorer       PlayerRef
	Time         *int64 // seconds; nil when the season hides goal times
}

// PlayerRef is how the scoresheet names a player: "#{num} {first} {last}", the number
// left out when it isn't known. It carries no player id, the id is resolved against
// the scoring team's scoreboard
type PlayerRef struct {
	Num  *int64
	Name string // "{first} {last}", "" when no player is recorded
}

// EventRow is a timeout, turnover or starting offence listed in the scoresheet
type EventRow struct {
	Type   string // timeout, turnover or offence
	Time   int64
	IsHome bool
}

// GameSpirit is the "Spirit points" row: each team's received category scores, and
// comments when the season shows them
type GameSpirit struct {
	Home            [5]int64
	Visitor         [5]int64
	HomeComments    *string
	VisitorComments *string
}

var (
	refPattern       = regexp.MustCompile(`^#(-?\d+) (.*)$`)
	scorePattern     = regexp.MustCompile(`^(\d+) - (\d+)$`)
	spiritPattern    = regexp.MustCompile(`\((\d+)\+(\d+)\+(\d+)\+(\d+)\+(\d+)\)`)
	halftimePattern  = regexp.MustCompile(`^(\d+\.\d{2}) halftime$`)
	eventPattern     = regexp.MustCompile(`^(Timeout|Turnover|Offence) (\d+\.\d{2})$`)
	headScorePattern = regexp.MustCompile(`(\d+) - (\d+)\s*(\(ongoing\))?$`)
)

func parseGame(doc *html.Node, id int64) (GamePage, error) {
	g := GamePage{GameID: id}
	c := content(doc)
	if c == nil {
		return g, fmt.Errorf("game %d: no content", id)
	}

	if a := first(links(c, "gamecard")); a != nil {
		if home, ok := linkID(a, "team1"); ok {
			g.HomeTeamID = &home
		}
		if visitor, ok := linkID(a, "team2"); ok {
			g.VisitorTeamID = &visitor
		}
	}
	if h1 := find(c, tag("h1")); h1 != nil {
		if m := headScorePattern.FindStringSubmatch(text(h1)); m != nil {
			g.HomeScore, g.VisitorScore = optionalInt(m[1]), optionalInt(m[2])
		}
	}

	for _, table := range findAll(c, tag("table")) {
		trs := rows(table)
		if len(trs) == 0 {
			continue
		}
		ths := children(trs[0], "th")
		switch {
		case len(ths) > 0 && text(ths[0]) == "#" && hasClass(ths[0], "home"):
			g.Home = parseScoreboard(trs[1:])
			g.HasScoresheet = true
		case len(ths) > 0 && text(ths[0]) == "#" && hasClass(ths[0], "guest"):
			g.Visitor = parseScoreboard(trs[1:])
		case len(ths) >= 3 && text(ths[0]) == "Scores":
			goals, events, err := parseScoresheet(trs[1:], ths)
			if err != nil {
				return g, fmt.Errorf("game %d: %w", id, err)
			}
			g.Goals, g.Events = goals, events
		}
	}

	// the timeline above the scoresheet carries the halftime as a tooltip
	for _, td := range findAll(c, tagClass("td", "halftime")) {
		if m := halftimePattern.FindStringSubmatch(attr(td, "title")); m != nil {
			if s, ok := secondsFromMinutes(m[1]); ok {
				g.Halftime = &s
			}
		}
	}

	for _, p := range findAll(c, tag("p")) {
		if official, ok := strings.CutPrefix(text(p), "Game official: "); ok {
			g.Official = official
		}
	}

	spirit, err := parseGameSpirit(c)
	if err != nil {
		return g, fmt.Errorf("game %d: %w", id, err)
	}
	g.Spirit = spirit

	if h2 := find(c, func(n *html.Node) bool { return n.Data == "h2" && text(n) == "Photos and Videos" }); h2 != nil {
		for n := h2.NextSibling; n != nil; n = n.NextSibling {
			if n.Type == html.ElementNode && n.Data == "table" {
				g.URLs = parseURLRows(n, true)
				break
			}
		}
	}
	return g, nil
}

func parseScoreboard(trs []*html.Node) []ScoreboardPlayer {
	var players []ScoreboardPlayer
	for _, tr := range trs {
		tds := children(tr, "td")
		if len(tds) < 4 {
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
		firstName, lastName, _ := strings.Cut(rawText(a), " ")
		assists, _ := parseInt(text(tds[2]))
		goals, _ := parseInt(text(tds[3]))
		players = append(players, ScoreboardPlayer{
			PlayerID:  id,
			Num:       optionalInt(text(tds[0])),
			FirstName: firstName,
			LastName:  lastName,
			Captain:   strings.HasSuffix(text(tds[1]), "(C)"),
			Assists:   assists,
			Goals:     goals,
		})
	}
	return players
}

// parseScoresheet reads the goal rows (Scores, Assist, Goal, then Time and Dur. unless
// the season hides times, then Game events when there are any) and the events listed
// alongside each goal. A row with class halftime marks where halftime fell and holds
// nothing else
func parseScoresheet(trs []*html.Node, header []*html.Node) ([]GoalRow, []EventRow, error) {
	hasTime := len(header) >= 5 && text(header[3]) == "Time"
	eventsCol := -1
	if last := header[len(header)-1]; text(last) == "Game events" {
		eventsCol = len(header) - 1
	}

	var goals []GoalRow
	var events []EventRow
	for _, tr := range trs {
		tds := children(tr, "td")
		if len(tds) == 0 || hasClass(tds[0], "halftime") {
			continue
		}
		if len(tds) < 3 {
			return nil, nil, fmt.Errorf("short scoresheet row %q", text(tr))
		}
		m := scorePattern.FindStringSubmatch(text(tds[0]))
		if m == nil {
			return nil, nil, fmt.Errorf("unreadable running score %q", text(tds[0]))
		}
		goal := GoalRow{IsHome: hasClass(tds[0], "home")}
		goal.HomeScore, _ = parseInt(m[1])
		goal.VisitorScore, _ = parseInt(m[2])
		if hasClass(tds[1], "callahan") {
			goal.IsCallahan = true
		} else {
			goal.Assist = parsePlayerRef(tds[1])
		}
		goal.Scorer = parsePlayerRef(tds[2])
		if hasTime && len(tds) > 3 {
			if s, ok := secondsFromMinutes(text(tds[3])); ok {
				goal.Time = &s
			}
		}
		goals = append(goals, goal)

		if eventsCol >= 0 && eventsCol < len(tds) {
			for _, div := range children(tds[eventsCol], "div") {
				em := eventPattern.FindStringSubmatch(text(div))
				if em == nil {
					continue // a media-event icon, not a game event
				}
				s, _ := secondsFromMinutes(em[2])
				events = append(events, EventRow{Type: strings.ToLower(em[1]), Time: s, IsHome: hasClass(div, "home")})
			}
		}
	}
	return goals, events, nil
}

func parsePlayerRef(td *html.Node) PlayerRef {
	s := text(td)
	if m := refPattern.FindStringSubmatch(s); m != nil {
		return PlayerRef{Num: optionalInt(m[1]), Name: m[2]}
	}
	return PlayerRef{Name: s}
}

// parseGameSpirit reads the game statistics table's "Spirit points" row, e.g.
// "12 (2+2+2+3+3)", and its "Spirit comments" row when shown
// nil when the season doesn't show spirit or a team hasn't submitted yet
func parseGameSpirit(c *html.Node) (*GameSpirit, error) {
	var spirit *GameSpirit
	for _, tr := range findAll(c, tag("tr")) {
		tds := children(tr, "td")
		if len(tds) < 3 {
			continue
		}
		switch text(tds[0]) {
		case "Spirit points:":
			home := spiritPattern.FindStringSubmatch(text(tds[1]))
			visitor := spiritPattern.FindStringSubmatch(text(tds[2]))
			if home == nil || visitor == nil {
				return nil, fmt.Errorf("unreadable spirit points %q / %q", text(tds[1]), text(tds[2]))
			}
			spirit = &GameSpirit{}
			for i := range 5 {
				spirit.Home[i], _ = parseInt(home[i+1])
				spirit.Visitor[i], _ = parseInt(visitor[i+1])
			}
		case "Spirit comments":
			if spirit != nil {
				home, visitor := multilineText(tds[1]), multilineText(tds[2])
				spirit.HomeComments, spirit.VisitorComments = &home, &visitor
			}
		}
	}
	return spirit, nil
}
