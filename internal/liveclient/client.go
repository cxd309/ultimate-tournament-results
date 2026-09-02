// Package liveclient is the HTTP transport every Live! by BULA spec version's Client
// embeds
// fetching {BasePath}{name}.json from one deployment host, with throttling
// It has no knowledge of any version's response shapes
// each internal/liveclient/vXX package embeds Client and adds its own strongly-typed
// Fetch* methods on top
package liveclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DefaultMinRequestInterval is the minimum time Client leaves between requests
// avoids overloading the source, throttling is a Client-wide property
const DefaultMinRequestInterval = 100 * time.Millisecond

// Client is the shared transport for one deployment host.
type Client struct {
	HTTP               *http.Client
	Host               string        // e.g. "wbuc.wfdf.sport"
	BasePath           string        // e.g. "/live/data/"; leading and trailing slash, every endpoint hangs off this
	SeasonID           string        // discovered by FetchHeartbeat; required by every other Fetch* method
	MinRequestInterval time.Duration // minimum gap between requests; <= 0 disables throttling

	lastRequestAt time.Time
}

// New builds a Client from a host and basePath.
func New(host, basePath string) Client {
	return Client{
		HTTP:               http.DefaultClient,
		Host:               host,
		BasePath:           basePath,
		MinRequestInterval: DefaultMinRequestInterval,
	}
}

// Throttle blocks until at least MinRequestInterval has passed since the last request
// this Client made.
func (c *Client) Throttle() {
	if c.MinRequestInterval <= 0 {
		return
	}
	if wait := c.MinRequestInterval - time.Since(c.lastRequestAt); wait > 0 {
		time.Sleep(wait)
	}
	c.lastRequestAt = time.Now()
}

// GetJSON fetches {BasePath}{name}.json and decodes it into out
func (c *Client) GetJSON(ctx context.Context, name string, out any) error {
	c.Throttle()

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

// GetSeasonJSON fetches {BasePath}{SeasonID}{suffix}.json and decodes it into out
// FetchHeartbeat must be called first to discover SeasonID
func (c *Client) GetSeasonJSON(ctx context.Context, suffix string, out any) error {
	if c.SeasonID == "" {
		return fmt.Errorf("season id not known yet -- call FetchHeartbeat first")
	}
	return c.GetJSON(ctx, c.SeasonID+suffix, out)
}
