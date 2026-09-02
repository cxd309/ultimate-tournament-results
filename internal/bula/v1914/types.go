// Package v1914 models the Live! by BULA 1.9.14-1.9.16 API shape
// see live-by-bula-openapi's openapi-1.9.14.yaml
// it can fetch a response and import into a internal/db/v1914 schema
package v1914

import "github.com/cxd309/ultimate-tournament-results/internal/convert"

// HeartbeatResponse is the response of GET {basePath}_heartbeat.json
// See openapi-1.9.14.yaml #/components/schemas/Heartbeat
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
// See openapi-1.9.14.yaml #/components/schemas/ReferenceResponse
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
// See openapi-1.9.14.yaml #/components/schemas/TeamDetailResponse
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
