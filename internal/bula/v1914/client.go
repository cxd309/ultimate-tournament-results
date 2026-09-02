package v1914

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DefaultMinRequestInterval is the minimum time Client leaves between requests
// To avoid overloading the source, throttling is a Client-wide property
const DefaultMinRequestInterval = 100 * time.Millisecond

// Client fetches Live! by BULA 1.9.14-1.9.16 shaped responses from one deployment
type Client struct {
	HTTP               *http.Client
	Host               string        // e.g. "wbuc.wfdf.sport"
	BasePath           string        // e.g. "/live/data/"; leading and trailing slash, every endpoint hangs off this
	SeasonID           string        // discovered by FetchHeartbeat; required by every other Fetch* method
	MinRequestInterval time.Duration // minimum gap between requests; <= 0 disables throttling

	lastRequestAt time.Time
}

// NewClient build a Client struct from a host and basePath
func NewClient(host, basePath string) *Client {
	return &Client{
		HTTP:               http.DefaultClient,
		Host:               host,
		BasePath:           basePath,
		MinRequestInterval: DefaultMinRequestInterval,
	}
}

// throttle blocks until at least MinRequestInterval has passed since the last request
// this Client made.
func (c *Client) throttle() {
	if c.MinRequestInterval <= 0 {
		return
	}
	if wait := c.MinRequestInterval - time.Since(c.lastRequestAt); wait > 0 {
		time.Sleep(wait)
	}
	c.lastRequestAt = time.Now()
}

// getJSON fetches {BasePath}{name}.json and decodes it into out
func (c *Client) getJSON(ctx context.Context, name string, out any) error {
	c.throttle()

	url := fmt.Sprintf("https://%s%s%s.json", c.Host, c.BasePath, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request for %s: %w", name, err)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s: unexpected status %s from %s", name, resp.Status, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}

	return nil
}

// getSeasonJSON fetches {BasePath}{SeasonID}{suffix}.json and decodes it into out
// FetchHeartbeat must be called first to discover SeasonID
func (c *Client) getSeasonJSON(ctx context.Context, suffix string, out any) error {
	if c.SeasonID == "" {
		return fmt.Errorf("season id not known yet -- call FetchHeartbeat first")
	}
	return c.getJSON(ctx, c.SeasonID+suffix, out)
}

// FetchHeartbeat fetches {BasePath}_heartbeat.json and records the season id it reports
// for use by every other Fetch* method.
func (c *Client) FetchHeartbeat(ctx context.Context) (*HeartbeatResponse, error) {
	var hb HeartbeatResponse
	if err := c.getJSON(ctx, "_heartbeat", &hb); err != nil {
		return nil, err
	}
	c.SeasonID = hb.Config.LiveSeasonID
	return &hb, nil
}

// FetchReference fetches {BasePath}{SeasonID}_reference.json
func (c *Client) FetchReference(ctx context.Context) (*ReferenceResponse, error) {
	var ref ReferenceResponse
	if err := c.getSeasonJSON(ctx, "_reference", &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// FetchTeams fetches {BasePath}{SeasonID}_teams.json
func (c *Client) FetchTeams(ctx context.Context) (*TeamsResponse, error) {
	var resp TeamsResponse
	if err := c.getSeasonJSON(ctx, "_teams", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FetchTeamDetail fetches {BasePath}{SeasonID}_teams_{teamID}.json.
func (c *Client) FetchTeamDetail(ctx context.Context, teamID int64) (*TeamDetailResponse, error) {
	var detail TeamDetailResponse
	if err := c.getSeasonJSON(ctx, fmt.Sprintf("_teams_%d", teamID), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}
