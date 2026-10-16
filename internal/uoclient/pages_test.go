package uoclient

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parse(t *testing.T, body string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestSecondsFromMinutes(t *testing.T) {
	for in, want := range map[string]int64{"0.00": 0, "1.55": 115, "84.05": 5045, "45.00": 2700} {
		if got, ok := secondsFromMinutes(in); !ok || got != want {
			t.Errorf("secondsFromMinutes(%q) = %d, %v; want %d", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "1.5", "1.60", "x.00"} {
		if _, ok := secondsFromMinutes(in); ok {
			t.Errorf("secondsFromMinutes(%q) accepted", in)
		}
	}
}

// the shapes below are trimmed from a real results.wfdf.sport/wmuc/ page

func TestParseScheduleRow(t *testing.T) {
	doc := parse(t, `<div class='content'><h1>Sunday</h1><h3>Sun 10.11.2024</h3>
<table><tr><th align='left' colspan='13'><a class='thlink' href='?view=reservationinfo&amp;reservation=2'></a> Field 10</th></tr>
<tr style='width:100%'>
<td><span>11:00</span></td>
<td><span>USA GGMO Relics</span></td>
<td><span>15</span>&nbsp;<span style='color: #aaaaaa;'>(10)</span></td>
<td><span>-</span></td>
<td><span>5</span>&nbsp;<span style='color: #aaaaaa;'>(11)</span></td>
<td><span>Japan GGMO</span></td>
<td><span>Great Grand Master Open</span></td>
<td><span><a href='?view=poolstatus&pool=8'>Pool A</a></span></td>
<td><span>Gold</span></td>
<td></td>
<td class='right'><span>&nbsp;<a href='?view=gameplay&amp;game=356'>Game play</a></span></td>
</tr></table>
<p class='timezone'>Timezone: America/Los_Angeles. Local time: 16.10.2026 01:04</p></div>`)
	games, tz, err := parseSchedule(doc)
	if err != nil {
		t.Fatal(err)
	}
	if tz != "America/Los_Angeles" {
		t.Errorf("timezone %q", tz)
	}
	if len(games) != 1 {
		t.Fatalf("%d games", len(games))
	}
	g := games[0]
	if g.GameID == nil || *g.GameID != 356 || g.ReservationID != 2 || g.PoolID != 8 ||
		g.FieldLabel != "Field 10" || g.ReservationGroup != "Sunday" || g.Time != "11:00" ||
		g.Home != "USA GGMO Relics" || *g.HomeScore != 15 || *g.VisitorScore != 5 || g.GameName != "Gold" {
		t.Errorf("unexpected row %+v", g)
	}
}

// trimmed from results.wfdf.sport/wjuc-2026/, which runs UltiOrganizer 3.x templates
func TestParseScheduleRow3x(t *testing.T) {
	doc := parse(t, `<div class='content'><h1>Saturday</h1><h3>Sat 11.7.2026</h3>
<table class='games-table'><tr><th align='left' colspan='13'><a class='thlink' href='?view=reservationinfo&amp;reservation=2'>Field 1</a></th></tr>
<tr style='width:100%'>
<td style='width:40px'><span>09:30</span></td>
<td style='width:140px'><span class='game-loser'><img height='10' src='images/flags/tiny/United_Kingdom.png' alt=''/> GBR O</span></td>
<td style='width:5px'>-</td>
<td style='width:140px'><span class='game-winner'><img height='10' src='images/flags/tiny/Israel.png' alt=''/> ISR O</span></td>
<td style='width:80px'><span>Open</span></td>
<td style='width:130px'><span><a href='?view=poolstatus&pool=6'>Pool B</a></span></td>
<td style='width:15px'><span>6</span></td>
<td style='width:5px'><span>-</span></td>
<td style='width:15px'><span>13</span></td>
<td></td>
<td style='width:50px'><span>S6</span></td>
<td style='width:40px;white-space: nowrap;'></td>
<td class='right nowrap' style='width:80px'><span>&nbsp;<a href='?view=gameplay&amp;game=57'>Gameplay</a></span></td>
</tr>
<tr><th align='left' colspan='13'><a class='thlink' href='?view=reservationinfo&amp;reservation=3'>Highfields Field H01</a></th></tr>
<tr style='width:100%'>
<td><span>11:00</span></td><td><span class='schedulingname'>A1</span></td><td>-</td><td><span class='schedulingname'>B2</span></td>
<td><span>Open</span></td><td><span><a href='?view=poolstatus&pool=7'>Final</a></span></td>
<td><span>?</span></td><td><span>-</span></td><td><span>?</span></td><td></td><td></td><td></td><td class='left'></td>
</tr></table></div>`)
	games, _, err := parseSchedule(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 2 {
		t.Fatalf("%d games", len(games))
	}
	g := games[0]
	if g.GameID == nil || *g.GameID != 57 || g.ReservationID != 2 || g.PoolID != 6 || g.PlaceName != "" ||
		g.FieldLabel != "Field 1" || g.Time != "09:30" || g.Home != "GBR O" || g.Visitor != "ISR O" ||
		g.Division != "Open" || g.PoolName != "Pool B" || *g.HomeScore != 6 || *g.VisitorScore != 13 || g.GameName != "S6" {
		t.Errorf("unexpected row %+v", g)
	}
	g = games[1]
	if g.GameID != nil || g.HomeScore != nil || g.VisitorScore != nil || g.PlaceName != "Highfields" || g.FieldLabel != "Field H01" || g.Home != "A1" || g.PoolID != 7 {
		t.Errorf("unexpected unplayed row %+v", g)
	}
}

func TestParseGame(t *testing.T) {
	doc := parse(t, `<div class='content'>
<h1>Germany MO - Venezuela MO&nbsp;&nbsp;&nbsp;&nbsp;15 - 10</h1>
<table><tr><td>
<table><tr><th class='home'>#</th><th class='home'>Name</th><th class='home center'>Ast.</th><th class='home center'>Goals</th><th class='home center'>Tot.</th></tr>
<tr><td>8</td><td><a href='?view=playercard&amp;series=0&amp;player=1225'>Julius&nbsp;Wierichs</a>&nbsp;(C)</td><td class='center'>1</td><td class='center'>0</td><td class='center'>1</td></tr>
<tr><td>1</td><td><a href='?view=playercard&amp;series=0&amp;player=1226'>Lennart&nbsp;Vollmer</a></td><td class='center'>0</td><td class='center'>1</td><td class='center'>1</td></tr>
</table></td><td>
<table><tr><th class='guest'>#</th><th class='guest'>Name</th><th>Ast.</th><th>Goals</th><th>Tot.</th></tr>
<tr><td>23</td><td><a href='?view=playercard&amp;series=0&amp;player=1335'>Mario &nbsp;Plaza </a></td><td>0</td><td>1</td><td>1</td></tr>
</table></td></tr></table>
<table><tr><td class='guest' title='10.00 0-1 x'></td><td class='halftime' title='45.00 halftime'></td></tr></table>
<table border='1'><tr><th>Scores</th><th>Assist</th><th>Goal</th><th>Time</th><th>Dur.</th><th>Game events </th></tr>
<tr><td class='home'>1 - 0</td><td>#8 Julius Wierichs&nbsp;</td><td>#1 Lennart Vollmer&nbsp;</td><td>1.55</td><td>1.55</td><td><div class='guest'>Offence&nbsp;0.00</div></td></tr>
<tr><td colspan='6' class='halftime'>Half-time</td></tr>
<tr><td class='guest'>1 - 1</td><td class='callahan'>Callahan-goal&nbsp;</td><td>#23 Mario  Plaza &nbsp;</td><td>50.00</td><td>48.05</td><td><div class='home'>Timeout&nbsp;46.10</div></td></tr>
</table>
<table><tr><td>Spirit points:</td><td class='home'>12 (2+2+2+3+3)</td><td class='guest'>10 (1+2+2+2+3)</td></tr></table>
<p><a href='?view=gamecard&amp;team1=53&amp;team2=58'>Game history</a></p></div>`)
	g, err := parseGame(doc, 36)
	if err != nil {
		t.Fatal(err)
	}
	if *g.HomeTeamID != 53 || *g.VisitorTeamID != 58 || *g.HomeScore != 15 || *g.VisitorScore != 10 {
		t.Errorf("teams/score %+v", g)
	}
	if len(g.Home) != 2 || !g.Home[0].Captain || g.Home[0].FirstName != "Julius" || g.Home[1].Goals != 1 {
		t.Errorf("home scoreboard %+v", g.Home)
	}
	if v := g.Visitor[0]; v.FirstName != "Mario " || v.LastName != "Plaza " {
		t.Errorf("visitor names kept verbatim %q %q", v.FirstName, v.LastName)
	}
	if len(g.Goals) != 2 || !g.Goals[1].IsCallahan || g.Goals[1].Scorer.Name != "Mario Plaza" || *g.Goals[0].Assist.Num != 8 || *g.Goals[0].Time != 115 {
		t.Errorf("goals %+v", g.Goals)
	}
	if g.Halftime == nil || *g.Halftime != 2700 {
		t.Errorf("halftime %v", g.Halftime)
	}
	if len(g.Events) != 2 || g.Events[0] != (EventRow{Type: "offence", Time: 0}) || g.Events[1] != (EventRow{Type: "timeout", Time: 2770, IsHome: true}) {
		t.Errorf("events %+v", g.Events)
	}
	if g.Spirit == nil || g.Spirit.Home != [5]int64{2, 2, 2, 3, 3} || g.Spirit.Visitor != [5]int64{1, 2, 2, 2, 3} || g.Spirit.HomeComments != nil {
		t.Errorf("spirit %+v", g.Spirit)
	}
}

func TestParsePoolRoundRobin(t *testing.T) {
	doc := parse(t, `<div class='content'><h2>Master Open, Pool A</h2><table border='2'><caption>PG - Played Games</caption>
<tr><th>#</th><th>Team</th></tr>
<tr><td class='center'><div style='background-color:#006400;background-color:rgba(0,100,0,0.3)'>1</div></td><td><div><img src='images/flags/tiny/Canada.png'/><a href='?view=teamcard&amp;team=57'>USA MO</a></div></td></tr>
<tr><td class='center'><div style=''>2</div></td><td><div><a href='?view=teamcard&amp;team=51'>Canada MO</a></div></td></tr>
</table>
<table><tr><td>Continues to:</td><td style='background-color:#006400;'><a href='?view=poolstatus&amp;pool=23'>Playoff (1-8)</a></td></tr></table>
<h2>Games</h2><table><tr><td><a href='?view=gameplay&amp;game=36'>Game play</a></td></tr></table></div>`)
	p, err := parsePool(doc, 1)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != PoolTypeRoundRobin || !p.International || len(p.Standings) != 2 || p.Standings[0].MoveColor != "006400" || p.Standings[1].MoveColor != "" {
		t.Errorf("pool %+v", p)
	}
	if len(p.ContinuesTo) != 1 || p.ContinuesTo[0] != (PoolContinuation{PoolID: 23, Color: "006400"}) || len(p.GameIDs) != 1 {
		t.Errorf("continuations/games %+v %v", p.ContinuesTo, p.GameIDs)
	}
}

func TestParseReservation(t *testing.T) {
	doc := parse(t, `<html><head><script type="text/javascript">
		var searchUrl = 'ext/locationxml.php?id=1';
		searchLocations(searchUrl);
</script></head><body><div class='content'><h1>Highfields Field H01</h1>
<p>28.6.2026 09:00 - 18:30</p><p>Highfields Road</p><p></p></div></body></html>`)
	r, err := parseReservation(doc, 2)
	if err != nil {
		t.Fatal(err)
	}
	if r.LocationID != 1 || r.Heading != "Highfields Field H01" || r.Times != "28.6.2026 09:00 - 18:30" || r.Address != "Highfields Road" {
		t.Errorf("reservation %+v", r)
	}
}

func TestParseReservation3x(t *testing.T) {
	doc := parse(t, `<div class='content'>
<h1>Reservation: Field 1</h1>
<p>11.7.2026 09:30 - 17:00</p>
</div>`)
	r, err := parseReservation(doc, 2)
	if err != nil {
		t.Fatal(err)
	}
	if r.LocationID != 0 || r.Heading != "Field 1" || r.Times != "11.7.2026 09:30 - 17:00" || r.Address != "" {
		t.Errorf("reservation %+v", r)
	}
	// a 3.x page never gives the location id, so a reservation with one is refused
	doc = parse(t, `<div class='content'><h1>Reservation: Highfields Field H01</h1><p>28.6.2026 09:00 - 18:30</p><p>Highfields Road</p></div>`)
	if _, err := parseReservation(doc, 3); err == nil {
		t.Error("3.x reservation with a location accepted")
	}
}

func TestParsePoolPlayoffLinks(t *testing.T) {
	doc := parse(t, `<div class='content'><h2>Open, Playoff (1-8) Semifinals</h2><p>No playoff tree template found.</p>
<table><tr><td class='right nowrap'><span>&nbsp;<a href='?view=gameplay&amp;game=152'>Gameplay</a></span></td></tr>
<tr><td class='right nowrap'><span>&nbsp;<a href='?view=gameplay&amp;game=153'>Gameplay</a></span></td></tr></table></div>`)
	p, err := parsePool(doc, 14)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type == PoolTypeRoundRobin || p.GameIDs != nil || len(p.LinkedGames) != 2 || p.LinkedGames[0] != 152 || p.LinkedGames[1] != 153 {
		t.Errorf("pool %+v", p)
	}
}

func TestAddUnlisted(t *testing.T) {
	id := func(n int64) *int64 { return &n }
	snap := &Snapshot{Schedule: []ScheduledGame{{GameID: id(1), PoolID: 6}}}
	added := addUnlisted(snap, []ScheduledGame{
		{GameID: id(1), PoolID: 6}, // listed pool, left to the schedule
		{GameID: id(152), PoolID: 14},
		{GameID: id(160), PoolID: 14}, // a forfeit: the pool page doesn't link it
	})
	if len(snap.Schedule) != 3 || len(added) != 2 {
		t.Fatalf("schedule %+v, added %v", snap.Schedule, added)
	}
	snap.Pools = map[int64]PoolPage{6: {}, 14: {LinkedGames: []int64{152}}}
	if err := linkUnlisted(snap, added); err != nil {
		t.Fatal(err)
	}
	if g := snap.Schedule[1].GameID; g == nil || *g != 152 {
		t.Errorf("linked game lost its id: %v", g)
	}
	if g := snap.Schedule[2].GameID; g != nil {
		t.Errorf("unlinked game kept id %d", *g)
	}
}

func TestParseMenu(t *testing.T) {
	for _, menuAttr := range []string{`id='menu_left'`, `class='menu_left'`} {
		doc := parse(t, `<table><tr><td `+menuAttr+`><table>
<tr><td><a class='seasonnav' href='?view=teams&amp;season=EBUCC2023'>EBUCC 2023</a></td></tr>
<tr><td class='menuserieslevel'><a href='?view=seriesstatus&amp;series=3'>Open</a></td></tr>
<tr><td class='menupoollevel'><a href='?view=poolstatus&amp;pool=12'>» Pool A</a></td></tr>
</table></td></tr></table>`)
		m, err := parseMenu(doc)
		if err != nil {
			t.Fatalf("%s: %v", menuAttr, err)
		}
		if m.SeasonID != "EBUCC2023" || len(m.Divisions) != 1 || len(m.Divisions[0].Pools) != 1 || m.Divisions[0].Pools[0] != (MenuPool{PoolID: 12, Name: "Pool A"}) {
			t.Errorf("%s: menu %+v", menuAttr, m)
		}
	}
}
