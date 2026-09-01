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
}

// NewClient build a Client struct from a host and basePath
func NewClient(host, basePath string) *Client {
	return &Client{HTTP: http.DefaultClient, Host: host, BasePath: basePath}
}

// FetchHeartbeat fetches {BasePath}_heartbeat.json.
func (c *Client) FetchHeartbeat(ctx context.Context) (*Heartbeat, error) {
	url := fmt.Sprintf("https://%s%s_heartbeat.json", c.Host, c.BasePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build heartbeat request: %w", err)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch heartbeat: unexpected status %s from %s", resp.Status, url)
	}

	var hb Heartbeat
	if err := json.NewDecoder(resp.Body).Decode(&hb); err != nil {
		return nil, fmt.Errorf("decode heartbeat: %w", err)
	}

	return &hb, nil
}
