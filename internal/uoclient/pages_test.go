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
