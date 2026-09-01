package v1914

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client fetches Live! by BULA 1.9.14-1.9.16 shaped responses from one deployment.
type Client struct {
	HTTP     *http.Client
	Host     string // e.g. "wbuc.wfdf.sport"
	BasePath string // e.g. "/live/data/"; leading and trailing slash, every endpoint hangs off this
	SeasonID string // discovered by FetchHeartbeat; required by every other Fetch* method
}

// NewClient build a Client struct from a host and basePath
func NewClient(host, basePath string) *Client {
	return &Client{HTTP: http.DefaultClient, Host: host, BasePath: basePath}
}

// getJSON fetches {BasePath}{name}.json and decodes it into out.
func (c *Client) getJSON(ctx context.Context, name string, out any) error {
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

// FetchReference fetches {BasePath}{SeasonID}_reference.json. FetchHeartbeat must be
// called first to discover SeasonID.
func (c *Client) FetchReference(ctx context.Context) (*ReferenceResponse, error) {
	if c.SeasonID == "" {
		return nil, fmt.Errorf("fetch reference: season id not known yet -- call FetchHeartbeat first")
	}
	var ref ReferenceResponse
	if err := c.getJSON(ctx, c.SeasonID+"_reference", &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}
