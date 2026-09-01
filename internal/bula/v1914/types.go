// Package v1914 models the Live! by BULA 1.9.14-1.9.16 API shape
// see live-by-bula-openapi's openapi-1.9.14.yaml
// it can fetch a response and import into a internal/db/v1914 schema
package v1914

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
//
// reference.teams isn't modeled here yet: team identity (from this response) and team
// stats (from _teams.json) are merged into one row in a later slice, so there's nothing
// to do with it until that's built -- an unmodeled JSON field is simply ignored on decode.
type ReferenceResponse struct {
	Season    Season    `json:"season"`
	Series    []Series  `json:"series"`
	Pools     []Pool    `json:"pools"`
	Countries []Country `json:"countries"`
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
