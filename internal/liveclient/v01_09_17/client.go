// Package liveclient fetches Live! by BULA 1.9.17 shaped responses from one
// deployment
package liveclient

import (
	"context"
	"fmt"

	"github.com/cxd309/ultimate-tournament-results/internal/liveclient"
	"github.com/cxd309/ultimate-tournament-results/internal/livedatamodel/v01_09_17"
)

// Client fetches Live! by BULA 1.9.17 shaped responses from one deployment
type Client struct {
	liveclient.Client
}

// NewClient builds a Client struct from a host and basePath
func NewClient(host, basePath string) *Client {
	return &Client{Client: liveclient.New(host, basePath)}
}

// FetchHeartbeat fetches {BasePath}_heartbeat.json and records the season id it reports
// for use by every other Fetch* method
func (c *Client) FetchHeartbeat(ctx context.Context) (*livedatamodel.HeartbeatResponse, error) {
	var hb livedatamodel.HeartbeatResponse
	if err := c.GetJSON(ctx, "_heartbeat", &hb); err != nil {
		return nil, err
	}
	c.SeasonID = hb.Config.LiveSeasonID
	return &hb, nil
}

// FetchReference fetches {BasePath}{SeasonID}_reference.json
func (c *Client) FetchReference(ctx context.Context) (*livedatamodel.ReferenceResponse, error) {
	var ref livedatamodel.ReferenceResponse
	if err := c.GetSeasonJSON(ctx, "_reference", &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// FetchTeamDetail fetches {BasePath}{SeasonID}_teams_{teamID}.json
func (c *Client) FetchTeamDetail(ctx context.Context, teamID int64) (*livedatamodel.TeamDetailResponse, error) {
	var detail livedatamodel.TeamDetailResponse
	if err := c.GetSeasonJSON(ctx, fmt.Sprintf("_teams_%d", teamID), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// FetchGames fetches {BasePath}{SeasonID}_games.json
func (c *Client) FetchGames(ctx context.Context) (*livedatamodel.GamesResponse, error) {
	var games livedatamodel.GamesResponse
	if err := c.GetSeasonJSON(ctx, "_games", &games); err != nil {
		return nil, err
	}
	return &games, nil
}

// FetchGameDetail fetches {BasePath}{SeasonID}_games_{gameID}.json
func (c *Client) FetchGameDetail(ctx context.Context, gameID int64) (*livedatamodel.GameDetailResponse, error) {
	var detail livedatamodel.GameDetailResponse
	if err := c.GetSeasonJSON(ctx, fmt.Sprintf("_games_%d", gameID), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}
