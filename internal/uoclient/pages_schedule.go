package uoclient

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// ScheduledGame is one game row of the season's games page sorted by field
// (view=games&filter=places), the only listing that gives each game's reservation
// and pool ids together
type ScheduledGame struct {
	GameID           *int64 // nil when the row has no "Game play" link (no scoresheet)
	ReservationID    int64
	PlaceName        string // the reservation's location name, "" when it has none
	FieldLabel       string // the header after the location name, e.g. "Field 10"
	ReservationGroup string // the <h1> grouping heading, uo_reservation.reservationgroup
	Time             string // "HH:MM" start, the date is only in the <h3> heading
	Home             string // team name, or the scheduling name for an unresolved slot
	Visitor          string
	HomeScore        *int64 // nil when the game has no result
	VisitorScore     *int64
	Division         string
	PoolID           int64
	PoolName         string
	GameName         string // e.g. "Gold", "" when the game has none
}

// parseSchedule reads view=games&filter=places&group=all, plus the season timezone
// printed under it
func parseSchedule(doc *html.Node) ([]ScheduledGame, string, error) {
	c := content(doc)
	if c == nil {
		return nil, "", fmt.Errorf("no content")
	}
	var games []ScheduledGame
	var group, placeName, fieldLabel string
	var reservation int64
	var parseErr error

	walk(c, func(n *html.Node) bool {
		if n.Type != html.ElementNode || parseErr != nil {
			return parseErr == nil
		}
		switch n.Data {
		case "h1":
			group = text(n)
			return false
		case "th":
			a := first(links(n, "reservationinfo"))
			if a == nil {
				return false
			}
			id, ok := linkID(a, "reservation")
			if !ok {
				parseErr = fmt.Errorf("reservation header without an id: %q", text(n))
				return false
			}
			reservation = id
			placeName, fieldLabel = placeHeader(text(a), text(n))
			return false
		case "tr":
			tds := children(n, "td")
			if len(tds) < 11 || first(links(n, "poolstatus")) == nil {
				return true
			}
			g, err := parseScheduleRow(tds)
			if err != nil {
				parseErr = err
				return false
			}
			g.ReservationID = reservation
			g.PlaceName = placeName
			g.FieldLabel = fieldLabel
			g.ReservationGroup = group
			games = append(games, g)
			return false
		}
		return true
	})
	if parseErr != nil {
		return nil, "", parseErr
	}
	if len(games) == 0 {
		return nil, "", fmt.Errorf("no scheduled games listed")
	}

	timezone := ""
	if p := find(c, tagClass("p", "timezone")); p != nil {
		if m := timezonePattern.FindStringSubmatch(text(p)); m != nil {
			timezone = m[1]
		}
	}
	return games, timezone, nil
}

var timezonePattern = regexp.MustCompile(`Timezone: (\S+?)\.(\s|$)`)

// placeHeader splits a reservation header into location name and "Field N" label.
// 1.x links the location name and prints the label after the link; 3.x puts
// ReservationPlaceText ("{location} Field {field}", either part may be missing)
// inside the link and nothing after it
func placeHeader(link, header string) (string, string) {
	if label := strings.TrimSpace(strings.TrimPrefix(header, link)); label != "" {
		return link, label
	}
	if strings.HasPrefix(link, "Field ") {
		return "", link
	}
	if i := strings.LastIndex(link, " Field "); i >= 0 {
		return link[:i], link[i+1:]
	}
	return link, ""
}

// parseScheduleRow reads one GameRow rendered with time, series and pool columns.
// 1.x renders time, home, home score, "-", visitor score, visitor, division, pool,
// game name, media, info (holding the "Game play" link); 3.x renders time, home, "-",
// visitor, division, pool, home score, "-", visitor score, forfeit mark, game name,
// media, info. The pool link's column tells the two apart
func parseScheduleRow(tds []*html.Node) (ScheduledGame, error) {
	var g ScheduledGame
	var poolTD, infoTD *html.Node
	switch {
	case first(links(tds[7], "poolstatus")) != nil:
		g = ScheduledGame{Time: text(tds[0]), Home: text(tds[1]), Visitor: text(tds[5]), Division: text(tds[6]), PoolName: text(tds[7]), GameName: text(tds[8])}
		g.HomeScore, g.VisitorScore = scoreCell(tds[2]), scoreCell(tds[4])
		poolTD, infoTD = tds[7], tds[10]
	case len(tds) >= 13 && first(links(tds[5], "poolstatus")) != nil:
		g = ScheduledGame{Time: text(tds[0]), Home: text(tds[1]), Visitor: text(tds[3]), Division: text(tds[4]), PoolName: text(tds[5]), GameName: text(tds[10])}
		g.HomeScore, g.VisitorScore = scoreCell(tds[6]), scoreCell(tds[8])
		poolTD, infoTD = tds[5], tds[12]
	default:
		return g, fmt.Errorf("game row without a pool link: %s", text(tds[0]))
	}
	pool, ok := firstLinkID(poolTD, "poolstatus", "pool")
	if !ok {
		return g, fmt.Errorf("game row without a pool id: %s %s - %s", g.Time, g.Home, g.Visitor)
	}
	g.PoolID = pool
	if id, ok := firstLinkID(infoTD, "gameplay", "game"); ok {
		g.GameID = &id
	}
	return g, nil
}

// scoreCell is a GameRow score cell's score, the first <span>; on 1.x the second
// holds the spirit total in brackets. nil for a game with no result, which renders
// &nbsp; on 1.x and "?" on 3.x
func scoreCell(td *html.Node) *int64 {
	span := find(td, tag("span"))
	if span == nil {
		return nil
	}
	return optionalInt(text(span))
}

// ReservationPage is view=reservationinfo for one reservation
type ReservationPage struct {
	ReservationID int64
	LocationID    int64  // 0 when the reservation has no location
	Heading       string // "{location name} Field {fieldname}"
	Times         string // "DD.MM.YYYY HH:MM - HH:MM"
	Address       string
	Info          string
}

var locationIDPattern = regexp.MustCompile(`locationxml\.php\?id=(\d+)`)

// parseReservation reads view=reservationinfo. 1.x always loads the location from
// ext/locationxml.php in a page script and follows the heading with times, address and
// info paragraphs; 3.x has no such script, heads the page "Reservation: {place}" and
// prints only the paragraphs that aren't empty. 3.x never gives the location id, so a
// 3.x reservation is only read when it has no location: no name in the heading and
// neither address nor info, just the times
func parseReservation(doc *html.Node, id int64) (ReservationPage, error) {
	r := ReservationPage{ReservationID: id}
	hasScript := false
	for _, script := range findAll(doc, tag("script")) {
		if m := locationIDPattern.FindStringSubmatch(scriptText(script)); m != nil {
			r.LocationID, _ = strconv.ParseInt(m[1], 10, 64)
			hasScript = true
		}
	}
	c := content(doc)
	if c == nil {
		return r, fmt.Errorf("reservation %d: no content", id)
	}
	h1 := find(c, tag("h1"))
	if h1 == nil {
		return r, fmt.Errorf("reservation %d: no heading", id)
	}
	r.Heading = text(h1)
	var ps []string
	for n := h1.NextSibling; n != nil && len(ps) < 3; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == "p" {
			ps = append(ps, text(n))
		}
	}
	if !hasScript {
		r.Heading = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(r.Heading, "Reservation"), ":"))
		if place, _ := placeHeader(r.Heading, r.Heading); place != "" || len(ps) != 1 {
			return r, fmt.Errorf("reservation %d: has a location, which 3.x pages don't give the id of", id)
		}
		r.Times = ps[0]
		return r, nil
	}
	if len(ps) < 3 {
		return r, fmt.Errorf("reservation %d: expected times, address and info after the heading", id)
	}
	r.Times, r.Address, r.Info = ps[0], ps[1], ps[2]
	return r, nil
}

// Location is one marker of ext/locationxml.php
type Location struct {
	ID      int64   `xml:"id,attr"`
	Name    string  `xml:"name,attr"`
	Address string  `xml:"address,attr"`
	Lat     *string `xml:"lat,attr"`
	Lng     *string `xml:"lng,attr"`
	Info    string  `xml:"info,attr"`
	Indoor  string  `xml:"indoor,attr"`
	Fields  string  `xml:"fields,attr"`
}

func parseLocations(body []byte) ([]Location, error) {
	var doc struct {
		Markers []Location `xml:"marker"`
	}
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("parse locationxml: %w", err)
	}
	return doc.Markers, nil
}
