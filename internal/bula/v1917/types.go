// Package v1917 models the Live! by BULA 1.9.17 API shape
// see live-by-bula-openapi's openapi-1.9.17.yaml
// it can fetch a response and import into a internal/db/v1917 schema
package v1917

import "github.com/cxd309/ultimate-tournament-results/internal/convert"

// HeartbeatResponse is the response of GET {basePath}_heartbeat.json
// See openapi-1.9.17.yaml #/components/schemas/Heartbeat
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
// See openapi-1.9.17.yaml #/components/schemas/ReferenceResponse
type ReferenceResponse struct {
	Season       Season          `json:"season"`
	Series       []Series        `json:"series"`
	Pools        []Pool          `json:"pools"`
	Teams        []ReferenceTeam `json:"teams"`
	Countries    []Country       `json:"countries"`
	Reservations []Reservation   `json:"reservations"`
}

// Season is ReferenceResponse.Season
// Only the fields this archiver currently uses are modeled here
type Season struct {
	Name      string `json:"name"`
	StartTime string `json:"starttime"`
	EndTime   string `json:"endtime"`
	Timezone  string `json:"timezone"`
	Status    string `json:"status"`
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
	Club                    string `json:"club"`
}

// Reservation is one entry in ReferenceResponse.Reservations
// it represents a playing field on a day
// Unlike uo_reservation itself, the API resolves the venue name (LocationName)
// straight into this object rather than leaving it as a bare id to join.
type Reservation struct {
	ID               int64  `json:"id"`
	FieldName        string `json:"fieldname"`
	ReservationGroup string `json:"reservationgroup"`
	Location         int64  `json:"location"`
	LocationName     string `json:"name"`
}

// TeamDetailResponse is the response of GET {basePath}{seasonId}_teams_{teamId}.json
// See openapi-1.9.17.yaml #/components/schemas/TeamDetailResponse
//
// Only the fields this archiver currently uses are modeled here
// spirit given/received belongs to a later slice, not this one.
type TeamDetailResponse struct {
	TeamID  int64           `json:"team_id"`
	Pool    int64           `json:"pool"` // external pool id; 0 means not currently in a pool
	Valid   convert.IntBool `json:"valid"`
	Players []PlayerStats   `json:"players"`
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
// only game_id is modeled: see GamesResponse
type GameListEntry struct {
	GameID int64 `json:"game_id"`
}

// GameDetailResponse is the response of GET {basePath}{seasonId}_games_{gameId}.json
// See openapi-1.9.17.yaml #/components/schemas/GameDetailResponse
//
// Only the fields this archiver currently uses are modeled here: game_info, seasoninfo,
// poolinfo, teams, the two scoreboards, gameevents/mediaevents and the captains are all
// either derivable from data already stored elsewhere or out of scope for this schema.
type GameDetailResponse struct {
	GameResult  GameResult       `json:"game_result"`
	Goals       []Goal           `json:"goals"`
	SpiritStats *GameSpiritStats `json:"spiritstats"` // absent when the event doesn't publish spirit points
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
	Homesotg              *int64          `json:"homesotg"`
	Visitorsotg           *int64          `json:"visitorsotg"`
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
}

// GameSpiritStats is GameDetailResponse.SpiritStats: both teams' spirit scores for one
// game, each keyed by the RECIPIENT -- the score is FOR that team, not given by it. See
// the spirit_scores table's own comment for how this was confirmed against the source.
type GameSpiritStats struct {
	Hometeam    *GameSpiritScore `json:"hometeam"`
	Visitorteam *GameSpiritScore `json:"visitorteam"`
}

// GameSpiritScore is one team's spirit score for one game, awarded by its opponent.
type GameSpiritScore struct {
	GameID   int64  `json:"game_id"`
	TeamID   int64  `json:"team_id"` // the team this score is for, not the team that gave it
	Cat1     int64  `json:"cat1"`
	Cat2     int64  `json:"cat2"`
	Cat3     int64  `json:"cat3"`
	Cat4     int64  `json:"cat4"`
	Cat5     int64  `json:"cat5"`
	Comments string `json:"comments"`
}
