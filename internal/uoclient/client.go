// Package uoclient gathers one tournament from a classic UltiOrganizer site, one with
// no Live! by BULA deployment, by scraping its public HTML pages and CSV exports
//
// UltiOrganizer's own ext/restful.php API is broken on current deployments (every
// collection 500s), so the HTML is the only route to the database ids. The CSV exports
// carry no ids, only the few columns the HTML doesn't show
//
// Every page is fetched in English (locale=en_GB.utf8) so labels, and date formats,
// are the same whatever the site's own default locale is
package uoclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/html"

	"github.com/cxd309/ultimate-tournament-results/internal/liveclient"
)

// DefaultMinRequestInterval is the minimum time Client leaves between requests,
// 10 a second: every request here is a full PHP page render, not a cached static
// file, and these are small servers often running other events at the same time
const DefaultMinRequestInterval = 100 * time.Millisecond

// locale pins every page to English, see the package doc
const locale = "en_GB.utf8"

// Client is the transport for one UltiOrganizer install on one host
// BasePath is the install's own path, e.g. "/wmuc/", with leading and trailing slash
// SeasonID (from the embedded liveclient.Client) is the UltiOrganizer season being
// archived, sent with every request as selseason so the left menu shows that season
//
// CacheDir, when set, keeps every response on disk keyed by its URL and serves
// repeats from there, so a re-run (e.g. after a resolve error) doesn't fetch the
// whole site again
type Client struct {
	liveclient.Client
	CacheDir string
}

// NewClient builds a Client for the UltiOrganizer install at host/basePath
func NewClient(host, basePath string) *Client {
	c := liveclient.New(host, basePath)
	c.MinRequestInterval = DefaultMinRequestInterval
	return &Client{Client: c}
}

// get fetches {BasePath}{file}?{query} and returns the body
// query always gains locale, and selseason once SeasonID is known
func (c *Client) get(ctx context.Context, file string, query url.Values) ([]byte, error) {
	q := url.Values{}
	for k, v := range query {
		q[k] = v
	}
	q.Set("locale", locale)
	if c.SeasonID != "" {
		q.Set("selseason", c.SeasonID)
	}
	u := fmt.Sprintf("https://%s%s%s?%s", c.Host, c.BasePath, file, q.Encode())

	var cachePath string
	if c.CacheDir != "" {
		sum := sha256.Sum256([]byte(u))
		cachePath = filepath.Join(c.CacheDir, hex.EncodeToString(sum[:]))
		if body, err := os.ReadFile(cachePath); err == nil {
			return body, nil
		}
	}

	c.Throttle()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request for %s: %w", u, err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", u, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %s", u, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", u, err)
	}
	if cachePath != "" {
		if err := os.MkdirAll(c.CacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("create cache %s: %w", c.CacheDir, err)
		}
		if err := os.WriteFile(cachePath, body, 0o644); err != nil {
			return nil, fmt.Errorf("cache %s: %w", u, err)
		}
	}
	return body, nil
}

// getView fetches index.php?view={view}&{params} and parses it as HTML
func (c *Client) getView(ctx context.Context, view string, params url.Values) (*html.Node, error) {
	q := url.Values{"view": {view}}
	for k, v := range params {
		q[k] = v
	}
	body, err := c.get(ctx, "index.php", q)
	if err != nil {
		return nil, err
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse %s page: %w", view, err)
	}
	return doc, nil
}

// getCSV fetches ext/{name}.php?season={SeasonID} and returns each data row keyed by
// its header, e.g. ext/gamescsv.php's "Time", "HomeTeam", ...
// see parseExportCSV for the format
func (c *Client) getCSV(ctx context.Context, name string) ([]map[string]string, error) {
	body, err := c.get(ctx, "ext/"+name+".php", url.Values{"season": {c.SeasonID}})
	if err != nil {
		return nil, err
	}
	records, err := parseExportCSV(string(body))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	if len(records) == 0 || !looksLikeHeader(records[0]) {
		return nil, fmt.Errorf("%s: %w", name, ErrNotCSV)
	}
	header := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for i, record := range records[1:] {
		if len(record) != len(header) {
			return nil, fmt.Errorf("parse %s: row %d has %d columns, header has %d", name, i+1, len(record), len(header))
		}
		row := make(map[string]string, len(header))
		for j, col := range header {
			row[col] = record[j]
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// ErrNotCSV is getCSV's error for a response that isn't an export at all, e.g. the
// "You don't have permission to access this page" text a site serves for an export
// it has locked down
var ErrNotCSV = errors.New("response is not a CSV export")

// looksLikeHeader is whether record is an export header: more than one column, all
// non-empty column names
func looksLikeHeader(record []string) bool {
	if len(record) < 2 {
		return false
	}
	for _, col := range record {
		if col == "" {
			return false
		}
	}
	return true
}

// getXML fetches ext/{name}.php?{params} and returns the raw body
func (c *Client) getXML(ctx context.Context, name string, params url.Values) ([]byte, error) {
	return c.get(ctx, "ext/"+name+".php", params)
}

// SpiritCSV fetches ext/spiritcsv.php for SeasonID: every submitted spirit score with
// its comments, keyed by the export's own header ("division", "day", "team_evaluated",
// ...). ErrNotCSV when the site has locked the export down
func (c *Client) SpiritCSV(ctx context.Context) ([]map[string]string, error) {
	return c.getCSV(ctx, "spiritcsv")
}
