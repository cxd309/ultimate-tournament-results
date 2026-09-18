// Package livepublish holds what's shared across every version's publish step:
// writing one rendered response to disk
//
// Reading the archive back out and reconstructing each endpoint's JSON is
// version-specific: that depends on each version's own store and livedatamodel
// types, so it lives in livepublish/vXX
package livepublish

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteJSON marshals v and writes it to dir/filename, creating dir if needed
//
// filename matches the live API's own relative naming
// (_heartbeat.json, {seasonId}_reference.json, etc)
// so a tool pointed at dir as its basePath sees the same files a live
// deployment would have served
func WriteJSON(dir, filename string, v any) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filename, err)
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// liveTimeLayout is the tournament-local "YYYY-MM-DD HH:MM:SS" format
// every Live! game time (and time_utc) uses
const liveTimeLayout = "2006-01-02 15:04:05"

// TimeUTC converts a tournament-local game time to UTC in the same format,
// evaluated against the season's IANA timezone, so a daylight-saving change
// mid-event is handled per game
// "" when there's no time, or the timezone is missing or unrecognised,
// matching the live API omitting time_utc in both cases
func TimeUTC(localTime, timezone string) string {
	if localTime == "" || timezone == "" {
		return ""
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return ""
	}
	t, err := time.ParseInLocation(liveTimeLayout, localTime, loc)
	if err != nil {
		return ""
	}
	return t.UTC().Format(liveTimeLayout)
}
