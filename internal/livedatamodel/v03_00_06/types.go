// Package livedatamodel models the Live! by BULA 3.0.6 API response shapes
// see live-by-bula-openapi's openapi-3.0.6.yaml
package livedatamodel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cxd309/ultimate-tournament-results/internal/convert"
)

// HeartbeatResponse is the response of GET {basePath}_heartbeat.json
// See openapi-3.0.6.yaml #/components/schemas/Heartbeat
type HeartbeatResponse struct {
	AppVersion     string          `json:"app_version"`
	CacheVersion   string          `json:"cache_version"`
	LastUpdatedUTC string          `json:"last_updated_utc"`
	Config         HeartbeatConfig `json:"config"`
}

// HeartbeatConfig is HeartbeatResponse.Config
// A real deployment sends ~32 keys
// only the ones this archiver currently uses are modeled here
type HeartbeatConfig struct {
	LiveSeasonID       string `json:"LIVE_SEASON_ID"`
	StaticCacheBaseURL string `json:"STATIC_CACHE_BASE_URL"`
	TournamentName     string `json:"TOURNAMENT_NAME"`
}

// ReferenceResponse is the response of GET {basePath}{seasonId}_reference.json
// See openapi-3.0.6.yaml #/components/schemas/ReferenceResponse
type ReferenceResponse struct {
	Season         Season          `json:"season"`
	Series         []Series        `json:"series"`
	Pools          []Pool          `json:"pools"`
	Teams          []ReferenceTeam `json:"teams"`
	Countries      []Country       `json:"countries"`
	Reservations   []Reservation   `json:"reservations"`
	PoolPlacements []PoolPlacement `json:"pool_placements"`
}

// Season is ReferenceResponse.Season
// Only the fields this archiver currently uses are modeled here
// enrollopen/enroll_deadline/istournament/organizer/category/showspiritpoints/
// showspiritcomments are genuine uo_season columns too, but the source endpoint strips
// or only partially exposes them so they are unrecoverable
type Season struct {
	Name                           string           `json:"name"`
	StartTime                      string           `json:"starttime"`
	EndTime                        string           `json:"endtime"`
	Iscurrent                      convert.IntBool  `json:"iscurrent"`
	Type                           string           `json:"type"`
	Isinternational                convert.IntBool  `json:"isinternational"`
	Isnationalteams                convert.IntBool  `json:"isnationalteams"`
	Showspiritpointsonlyoncomplete convert.IntBool  `json:"showspiritpointsonlyoncomplete"`
	Lockteamspiritonsubmit         convert.IntBool  `json:"lockteamspiritonsubmit"`
	UseSeasonPoints                convert.IntBool  `json:"use_season_points"`
	HideTimeOnScoresheet           convert.IntBool  `json:"hide_time_on_scoresheet"`
	Hometeammode                   convert.IntBool  `json:"hometeammode"`
	EventReadonly                  convert.IntBool  `json:"event_readonly"`
	MaintenanceMode                convert.IntBool  `json:"maintenance_mode"`
	PublicEvent                    convert.IntBool  `json:"public_event"`
	ApiPublic                      convert.IntBool  `json:"api_public"`
	Timezone                       string           `json:"timezone"`
	Status                         string           `json:"status"`
	SpiritMode                     *int64           `json:"spiritmode"`       // which spirit mode does this tournament use
	SpiritCategories               []SpiritCategory `json:"spiritCategories"` // spirit categories per spirit mode
}

// SpiritCategory is one entry in Season.SpiritCategories
// the definition of a spirit category for a sprit mode
type SpiritCategory struct {
	CategoryID int64  `json:"category_id"`
	Key        string `json:"key"` // "cat1", "cat2", ...
	Index      int64  `json:"index"`
	Group      int64  `json:"group"`
	Min        int64  `json:"min"`
	Max        int64  `json:"max"`
	Factor     int64  `json:"factor"`
	Label      string `json:"label"`
}

// PoolPlacement is one entry in ReferenceResponse.PoolPlacements
// a team's resolved position in a pool
type PoolPlacement struct {
	PoolID    int64  `json:"pool_id"`
	TeamID    int64  `json:"team_id"`
	Placement *int64 `json:"placement"` // null while the pool has no resolved rank yet
}

// Series is one entry in ReferenceResponse.Series, a division
type Series struct {
	SeriesID int64  `json:"series_id"`
	Name     string `json:"name"`
	Ordering string `json:"ordering"`
}

// Pool is one entry in ReferenceResponse.Pools, a pool or bracket within a division
type Pool struct {
	PoolID        int64           `json:"pool_id"`
	PoolName      string          `json:"poolname"`
	SeriesID      int64           `json:"series_id"`
	Ordering      string          `json:"ordering"`
	Type          int64           `json:"type"`
	Visible       convert.IntBool `json:"visible"`
	Played        convert.IntBool `json:"played"`
	Placementpool convert.IntBool `json:"placementpool"`
	Continuing    convert.IntBool `json:"continuingpool"`
	Color         string          `json:"color"`
	Timeslot      *int64          `json:"timeslot"`
	// Isfollower is computed by the reference endpoint from the raw `follower` column
	// (the target pool id itself, only exposed via poolinfo, see PoolInfo.Follower)
	Isfollower convert.IntBool `json:"isfollower"`
}

// Country is one entry in ReferenceResponse.Countries
type Country struct {
	CountryID    int64  `json:"country_id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	FlagFile     string `json:"flagfile"`
}

// ReferenceTeam is one entry in ReferenceResponse.Teams
// Has everything the teams table needs except pool/valid,
// which only come from the team-detail endpoint.
type ReferenceTeam struct {
	TeamID                  int64  `json:"team_id"`
	Name                    string `json:"name"`
	Abbreviation            string `json:"abbreviation"`
	Series                  int64  `json:"series"`
	Country                 int64  `json:"country"`
	Rank                    int64  `json:"rank"`
	FinalStanding           int64  `json:"final_standing"`
	FinalStandingCalculated int64  `json:"final_standing_calculated"`
	Club                    *int64 `json:"club"` // the raw uo_team.club FK
}

// Reservation is one entry in ReferenceResponse.Reservations
// it represents a playing field on a day
// Unlike uo_reservation itself, the API resolves the venue name (LocationName)
// straight into this object rather than leaving it as a bare id to join.
//
// Location is nullable on this line
// 0 or null both mean "the event uses a single unnamed site"
type Reservation struct {
	ID               int64  `json:"id"`
	FieldName        string `json:"fieldname"`
	ReservationGroup string `json:"reservationgroup"`
	Location         *int64 `json:"location"`
	LocationName     string `json:"name"`
}

// TeamDetailResponse is the response of GET {basePath}{seasonId}_teams_{teamId}.json
// See openapi-3.0.6.yaml #/components/schemas/TeamDetailResponse
//
// Only the fields this archiver currently uses are modeled here
// spirit given/received belongs to a later slice, not this one.
type TeamDetailResponse struct {
	TeamID  int64           `json:"team_id"`
	Pool    int64           `json:"pool"` // external pool id; 0 means not currently in a pool
	Valid   convert.IntBool `json:"valid"`
	Players []PlayerStats   `json:"players"`
	// Clubname is the resolved club-name text
	// ReferenceResponce.Teams[].Club is the FK id
	Clubname *string `json:"clubname"`
}

// PlayerStats is one entry in TeamDetailResponse.Players
// a squad member with event totals already aggregated
// Only PlayerID/FirstName/LastName are in PlayerStats's required list
// the rest are pointers because the spec doesn't guarantee they're sent.
type PlayerStats struct {
	PlayerID  int64  `json:"player_id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Num       *int64 `json:"num"`
	Games     *int64 `json:"games"`
}

// GamesResponse is the response of GET {basePath}{seasonId}_games.json
// Every value that is falsy is stripped before this is sent, so it's only used here to
// enumerate every game_id -- the per-game detail endpoint (GameDetailResponse) is the
// source of truth for every other field, since it's the only endpoint where nothing is
// stripped.
type GamesResponse struct {
	Games []GameListEntry `json:"games"`
}

// GameListEntry is one entry in GamesResponse.Games
// limited fields are recorded here
// everything else is gathered from GameDetailResponse
type GameListEntry struct {
	GameID int64
	// Pools: every pool this game belongs to
	// Absent from GameResult (the detail endpoint), so this is the only source for game_pools
	Pools []int64
	// HomeSchedulingFrompool/VisitorSchedulingFrompool name the pool a placeholder slot's
	// team will be drawn from; Absent from GameResult/GameInfo
	HomeSchedulingFrompool    *int64
	VisitorSchedulingFrompool *int64
}

// GamesResponse.Games.Pools is sent as a comma-seperated, de-duplicted, sorted string
// Unmarshall into a list of ints
func (g *GameListEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		GameID                    int64  `json:"game_id"`
		Pools                     string `json:"pools"`
		HomeSchedulingFrompool    *int64 `json:"home_scheduling_frompool"`
		VisitorSchedulingFrompool *int64 `json:"visitor_scheduling_frompool"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	g.GameID = raw.GameID
	g.HomeSchedulingFrompool = raw.HomeSchedulingFrompool
	g.VisitorSchedulingFrompool = raw.VisitorSchedulingFrompool
	if raw.Pools == "" {
		return nil
	}
	for _, id := range strings.Split(raw.Pools, ",") {
		poolID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return fmt.Errorf("parse pool id %q: %w", id, err)
		}
		g.Pools = append(g.Pools, poolID)
	}
	return nil
}

// GameDetailResponse is the response of GET {basePath}{seasonId}_games_{gameId}.json
// See openapi-3.0.6.yaml #/components/schemas/GameDetailResponse
//
// seasoninfo, teams, the two scoreboards, gameevents/mediaevents and the captains are
// all either derivable from data already stored elsewhere or out of scope for this
// schema; game_info/poolinfo are only partially modeled, see GameInfo/PoolInfo
type GameDetailResponse struct {
	GameResult  GameResult       `json:"game_result"`
	GameInfo    *GameInfo        `json:"game_info"`
	PoolInfo    PoolInfo         `json:"poolinfo"`
	Goals       []Goal           `json:"goals"`
	SpiritStats *GameSpiritStats `json:"spiritstats"` // absent when the event doesn't publish spirit points
}

// GameInfo is GameDetailResponse.GameInfo: the display view of the game, resolving
// team/pool/division ids to names and adding the pool's scoring rules
//
// Only the scheduling-name placeholder text is modeled here
// everything else is redundant with data from GameResult
type GameInfo struct {
	Phometeamname    *string `json:"phometeamname"`
	Pvisitorteamname *string `json:"pvisitorteamname"`
}

// PoolInfo is GameDetailResponse.PoolInfo:
// the pool's full rule set, as seen from one of its games.
//
// PlayoffTemplate decodes itself: the API sends it as a string, an integer or null, and
// this archive's schema stores it as text regardless of which.
type PoolInfo struct {
	PoolID           int64
	Drawsallowed     convert.IntBool
	PlayoffTemplate  *string
	Teams            *int64
	Mvgames          *int64
	Timeoutlen       *int64
	Halftime         *int64
	Winningscore     *int64
	Timecap          *int64
	Scorecap         *int64
	Addscore         *int64
	Halftimescore    *int64
	Timeouts         *int64
	Timeoutsper      string
	Timeoutsovertime *int64
	// Timeoutstimecap mirrors uo_pool's own varchar(5) column
	// Live! API sends as an integer
	Timeoutstimecap  *int64
	Betweenpointslen *int64
	Forfeitscore     *int64
	Forfeitagainst   *int64
	// Follower is the raw target pool id
	// the reference endpoint only exposes the computed Pool.Isfollower
	// derived from this, not the id itself
	Follower *int64
}

func (p *PoolInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		PoolID           int64           `json:"pool_id"`
		Drawsallowed     convert.IntBool `json:"drawsallowed"`
		PlayoffTemplate  json.RawMessage `json:"playoff_template"`
		Teams            *int64          `json:"teams"`
		Mvgames          *int64          `json:"mvgames"`
		Timeoutlen       *int64          `json:"timeoutlen"`
		Halftime         *int64          `json:"halftime"`
		Winningscore     *int64          `json:"winningscore"`
		Timecap          *int64          `json:"timecap"`
		Scorecap         *int64          `json:"scorecap"`
		Addscore         *int64          `json:"addscore"`
		Halftimescore    *int64          `json:"halftimescore"`
		Timeouts         *int64          `json:"timeouts"`
		Timeoutsper      string          `json:"timeoutsper"`
		Timeoutsovertime *int64          `json:"timeoutsovertime"`
		Timeoutstimecap  *int64          `json:"timeoutstimecap"`
		Betweenpointslen *int64          `json:"betweenpointslen"`
		Forfeitscore     *int64          `json:"forfeitscore"`
		Forfeitagainst   *int64          `json:"forfeitagainst"`
		Follower         *int64          `json:"follower"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.PoolID = raw.PoolID
	p.Drawsallowed = raw.Drawsallowed
	p.Teams = raw.Teams
	p.Mvgames = raw.Mvgames
	p.Timeoutlen = raw.Timeoutlen
	p.Halftime = raw.Halftime
	p.Winningscore = raw.Winningscore
	p.Timecap = raw.Timecap
	p.Scorecap = raw.Scorecap
	p.Addscore = raw.Addscore
	p.Halftimescore = raw.Halftimescore
	p.Timeouts = raw.Timeouts
	p.Timeoutsper = raw.Timeoutsper
	p.Timeoutsovertime = raw.Timeoutsovertime
	p.Timeoutstimecap = raw.Timeoutstimecap
	p.Betweenpointslen = raw.Betweenpointslen
	p.Forfeitscore = raw.Forfeitscore
	p.Forfeitagainst = raw.Forfeitagainst
	p.Follower = raw.Follower

	switch string(raw.PlayoffTemplate) {
	case "", "null":
		p.PlayoffTemplate = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(raw.PlayoffTemplate, &s); err == nil {
		p.PlayoffTemplate = &s
		return nil
	}
	// not a JSON string (e.g. a bare number) -- store its literal text as-is
	s = string(raw.PlayoffTemplate)
	p.PlayoffTemplate = &s
	return nil
}

// GameResult is GameDetailResponse.GameResult: the game's stored row, closer to raw
// uo_game than game_info, and the source of truth for the games table since nothing is
// falsy-stripped on this endpoint (unlike the games list).
//
// Hometeam/Visitorteam/Pool are modeled as nullable here despite the spec's required
// list, matching genuine uo_game nullability -- see the games table's own comment.
type GameResult struct {
	GameID                int64           `json:"game_id"`
	Hometeam              *int64          `json:"hometeam"`
	Visitorteam           *int64          `json:"visitorteam"`
	Homescore             *int64          `json:"homescore"`
	Visitorscore          *int64          `json:"visitorscore"`
	Time                  string          `json:"time"`
	Pool                  *int64          `json:"pool"`
	Reservation           *int64          `json:"reservation"`
	Valid                 convert.IntBool `json:"valid"`
	Isongoing             convert.IntBool `json:"isongoing"`
	Halftime              *int64          `json:"halftime"`
	Official              string          `json:"official"`
	Respteam              *int64          `json:"respteam"`
	Resppers              *int64          `json:"resppers"`
	Timeslot              *int64          `json:"timeslot"`
	Homedefenses          *int64          `json:"homedefenses"`
	Visitordefenses       *int64          `json:"visitordefenses"`
	Islive                convert.IntBool `json:"islive"`
	Liveurl               string          `json:"liveurl"`
	SchedulingNameHome    *int64          `json:"scheduling_name_home"`
	SchedulingNameVisitor *int64          `json:"scheduling_name_visitor"`
	// Name is the scheduling-name id, sent as a numeric string because `name` is exempt
	// from the API's usual number coercion -- parsed back to int64 on import.
	Name *string `json:"name"`
	// Gamename resolves Name to its display text, same value as gameschedulingname
	// a duplicate from a second join on uo_scheduling_name
	Gamename   string          `json:"gamename"`
	Hasstarted *int64          `json:"hasstarted"`  // has been seen with value 2 so not IntBool
	ShowSpirit convert.IntBool `json:"show_spirit"` // Are spirit scores shown
	// Forfeit: whether the game was forfeited. A forfeit reports completed with a 0-0
	// scoreline, so this is the only way to distinguish it from a genuine 0-0 result.
	Forfeit             convert.IntBool `json:"forfeit"`
	TimerStart          *int64          `json:"timer_start"`
	TimerPauseStart     *int64          `json:"timer_pause_start"`
	TimerPausedDuration *int64          `json:"timer_paused_duration"`
}

// Goal is one entry in GameDetailResponse.Goals, one goal in scoring order.
//
// Scorer/Assist arrive as literal 0 (not null) when unrecorded, per the spec; import
// normalizes that to nil since 0 isn't a real player_id and would violate the FK.
// Scorer/assist name and jersey fields are skipped: redundant with the players table.
type Goal struct {
	Num          int64           `json:"num"`
	Time         *int64          `json:"time"`
	Scorer       *int64          `json:"scorer"`
	Assist       *int64          `json:"assist"`
	Homescore    int64           `json:"homescore"`
	Visitorscore int64           `json:"visitorscore"`
	Ishomegoal   convert.IntBool `json:"ishomegoal"`
	Iscallahan   convert.IntBool `json:"iscallahan"`
	Timestamp    string          `json:"timestamp"` // scorekeeper's entry time, not the goal's
}

// GameSpiritStats is GameDetailResponse.SpiritStats: both teams' spirit scores for one
// game, each keyed by the RECIPIENT -- the score is FOR that team, not given by it. See
// the spirit_scores table's own comment for how this was confirmed against the source.
type GameSpiritStats struct {
	Hometeam    *GameSpiritScore `json:"hometeam"`
	Visitorteam *GameSpiritScore `json:"visitorteam"`
}

// GameSpiritScore is a team's spirit score for a game, awarded by their opponent
// categories are variable cat1 to catN set from season.spiritCategories
type GameSpiritScore struct {
	// Categories holds each category's score, keyed by its "catN" name (SpiritCategory.Key)
	// A nil value means the category is present but not yet visible; 0 is a real score
	Categories map[string]*int64
	// Comments is present only when:
	// spirit comments are enabled for the season and
	// game is cleared for publication
	Comments *string
}

func (s *GameSpiritScore) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Categories = make(map[string]*int64, len(raw))
	for key, value := range raw {
		if key == "comments" {
			if err := json.Unmarshal(value, &s.Comments); err != nil {
				return fmt.Errorf("spirit score comments: %w", err)
			}
			continue
		}
		var score *int64
		if err := json.Unmarshal(value, &score); err != nil {
			return fmt.Errorf("spirit score category %s: %w", key, err)
		}
		s.Categories[key] = score
	}
	return nil
}
