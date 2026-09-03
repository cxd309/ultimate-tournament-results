// Package liveclient fetches Live! by BULA 1.9.14-1.9.17 shaped responses from one
// deployment
package liveclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/cxd309/ultimate-tournament-results/internal/liveclient"
	livedatamodel "github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v01_09_14"
)

// Client fetches Live! by BULA 1.9.14-1.9.17 shaped responses from one deployment
type Client struct {
	liveclient.Client

	// SeasonIDOverride and Unprefixed cover known-unsupported deployments
	// zero value for both means behave normally:
	// discover the season id from the heartbeat, and
	// prefix every filename with it

	// SeasonIDOverride replaces the heartbeat-discovered season id when set
	// also tolerates the heartbeat fetch itself failing, for deployments with no
	// heartbeat endpoint at all
	SeasonIDOverride string

	// Unprefixed means static filenames carry no season-id prefix
	// ({suffix}.json instead of {SeasonID}{suffix}.json)
	Unprefixed bool
}

// NewClient builds a Client struct from a host and basePath
func NewClient(host, basePath string) *Client {
	return &Client{Client: liveclient.New(host, basePath)}
}

// FetchHeartbeat fetches {BasePath}_heartbeat.json and records the season id it reports
// for use by every other Fetch* method
// a legacy deployment with no config.LIVE_SEASON_ID, or no heartbeat endpoint at
// all, still succeeds here as long as SeasonIDOverride is set
func (c *Client) FetchHeartbeat(ctx context.Context) (*livedatamodel.HeartbeatResponse, error) {
	var hb livedatamodel.HeartbeatResponse
	if err := c.GetJSON(ctx, "_heartbeat", &hb); err != nil {
		if c.SeasonIDOverride == "" {
			return nil, err
		}
		hb = livedatamodel.HeartbeatResponse{}
	}
	if c.SeasonIDOverride != "" {
		hb.Config.LiveSeasonID = c.SeasonIDOverride
	}
	c.SeasonID = hb.Config.LiveSeasonID
	if c.SeasonID == "" {
		return nil, fmt.Errorf("season id not discovered (heartbeat has no config.LIVE_SEASON_ID) and no -season-id override given")
	}
	return &hb, nil
}

// seasonFilename builds one static endpoint's filename (without the .json
// extension) from a canonical suffix like "_reference" or "_teams_5"
// Unprefixed deployments serve these with the leading season id stripped
func (c *Client) seasonFilename(suffix string) string {
	if c.Unprefixed {
		return strings.TrimPrefix(suffix, "_")
	}
	return c.SeasonID + suffix
}

// FetchReference fetches {BasePath}{SeasonID}_reference.json
func (c *Client) FetchReference(ctx context.Context) (*livedatamodel.ReferenceResponse, error) {
	var ref livedatamodel.ReferenceResponse
	if err := c.GetJSON(ctx, c.seasonFilename("_reference"), &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// FetchTeamDetail fetches {BasePath}{SeasonID}_teams_{teamID}.json
func (c *Client) FetchTeamDetail(ctx context.Context, teamID int64) (*livedatamodel.TeamDetailResponse, error) {
	var detail livedatamodel.TeamDetailResponse
	if err := c.GetJSON(ctx, c.seasonFilename(fmt.Sprintf("_teams_%d", teamID)), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// FetchGames fetches {BasePath}{SeasonID}_games.json
func (c *Client) FetchGames(ctx context.Context) (*livedatamodel.GamesResponse, error) {
	var games livedatamodel.GamesResponse
	if err := c.GetJSON(ctx, c.seasonFilename("_games"), &games); err != nil {
		return nil, err
	}
	return &games, nil
}

// FetchGameDetail fetches {BasePath}{SeasonID}_games_{gameID}.json
func (c *Client) FetchGameDetail(ctx context.Context, gameID int64) (*livedatamodel.GameDetailResponse, error) {
	var detail livedatamodel.GameDetailResponse
	if err := c.GetJSON(ctx, c.seasonFilename(fmt.Sprintf("_games_%d", gameID)), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}
