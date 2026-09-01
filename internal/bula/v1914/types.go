// Package v1914 models the Live! by BULA 1.9.14-1.9.16 API shape
// see live-by-bula-openapi's openapi-1.9.14.yaml
// it can fetch a response and import into a internal/db/v1914 schema
package v1914

// Heartbeat is the response of GET {basePath}_heartbeat.json.
// See openapi-1.9.14.yaml #/components/schemas/Heartbeat.
type Heartbeat struct {
	AppVersion     string          `json:"app_version"`
	CacheVersion   string          `json:"cache_version"`
	LastUpdatedUTC string          `json:"last_updated_utc"`
	Config         HeartbeatConfig `json:"config"`
}

// HeartbeatConfig is Heartbeat.Config
// A real deployment sends ~32 keys
// only the ones this archiver currently uses are modeled here
type HeartbeatConfig struct {
	LiveSeasonID       string `json:"LIVE_SEASON_ID"`
	StaticCacheBaseURL string `json:"STATIC_CACHE_BASE_URL"`
	TournamentName     string `json:"TOURNAMENT_NAME"`
}
