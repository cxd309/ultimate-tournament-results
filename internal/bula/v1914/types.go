// Package v1914 models the Live! by BULA 1.9.14-1.9.16 API shape
// see live-by-bula-openapi's openapi-1.9.14.yaml
// it can fetch a response and import into a internal/db/v1914 schema
package v1914

import (
	"encoding/json"
	"fmt"
)

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
	Season    Season          `json:"season"`
	Series    []Series        `json:"series"`
	Pools     []Pool          `json:"pools"`
	Teams     []ReferenceTeam `json:"teams"`
	Countries []Country       `json:"countries"`
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
	PoolID   int64  `json:"pool_id"`
	PoolName string `json:"poolname"`
	SeriesID int64  `json:"series_id"`
	Ordering string `json:"ordering"`
	Type     int64  `json:"type"`
}

// Country is one entry in ReferenceResponse.Countries
type Country struct {
	CountryID    int64  `json:"country_id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	FlagFile     string `json:"flagfile"`
}

// ReferenceTeam is one entry in ReferenceResponse.Teams
// Only Abbreviation and Club are used from here
// all other fields will come from TeamStats which is more complete
type ReferenceTeam struct {
	TeamID       int64  `json:"team_id"`
	Abbreviation string `json:"abbreviation"`
	Club         string `json:"club"`
}

// TeamsResponse is the response of GET {basePath}{seasonId}_teams.json
// See openapi-1.9.14.yaml #/components/schemas/TeamsResponse
type TeamsResponse struct {
	Teams map[string]TeamStats `json:"teams"` // keyed by team_id as a string
}

// TeamStats is one entry in TeamsResponse.Teams
type TeamStats struct {
	TeamID                  int64       `json:"team_id"`
	Name                    string      `json:"name"`
	Series                  int64       `json:"series"`
	Country                 int64       `json:"country"`
	Seed                    int64       `json:"seed"`
	Games                   int64       `json:"games"`
	Wins                    int64       `json:"wins"`
	Losses                  int64       `json:"losses"`
	For                     int64       `json:"for"`
	Against                 int64       `json:"against"`
	FinalStanding           int64       `json:"final_standing"`
	FinalStandingCalculated int64       `json:"final_standing_calculated"`
	Spirit                  SpiritValue `json:"spirit"`    // only sent when the event publishes spirit
	SpiritAvg               SpiritValue `json:"spiritavg"` // same condition
}

// SpiritValue decodes a spirit total/average field
// This is normally a JSON number but is "N/A" for a team marked invalid
// Valid is false for "N/A" or where an event does not publish spirit
type SpiritValue struct {
	Value float64
	Valid bool
}

func (v *SpiritValue) UnmarshalJSON(data []byte) error {
	if string(data) == `"N/A"` {
		v.Value, v.Valid = 0, false
		return nil
	}
	if err := json.Unmarshal(data, &v.Value); err != nil {
		return fmt.Errorf("spirit value: %w", err)
	}
	v.Valid = true
	return nil
}

// TeamDetailResponse is the response of GET {basePath}{seasonId}_teams_{teamId}.json
// See openapi-1.9.14.yaml #/components/schemas/TeamDetailResponse
//
// Only the fields this archiver currently uses are modeled here
// spirit given/received belongs to a later slice, not this one.
type TeamDetailResponse struct {
	TeamID  int64         `json:"team_id"`
	Pool    int64         `json:"pool"` // external pool id; 0 means not currently in a pool
	Players []PlayerStats `json:"players"`
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
	Done      *int64 `json:"done"`  // goals
	Fedin     *int64 `json:"fedin"` // assists
	Callahan  *int64 `json:"callahan"`
}
