package omnisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
)

// Item is a raw search result from the Omnisearch HTTP API.
type Item struct {
	Basename string  `json:"basename"`
	Excerpt  string  `json:"excerpt"`
	Score    float64 `json:"score"`
	Path     string  `json:"path"`
}

// Result is a processed search result with an absolute filesystem path.
type Result struct {
	Basename string  `json:"basename"`
	Excerpt  string  `json:"excerpt"`
	Score    float64 `json:"score"`
	AbsPath  string  `json:"absPath"`
}

// Client searches Obsidian notes via the local Omnisearch HTTP API.
type Client interface {
	Search(ctx context.Context, query string) ([]Result, error)
}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// HTTPClient is an adapter that calls the Omnisearch HTTP endpoint.
type HTTPClient struct {
	baseURL   string
	vaultPath string
	httpDoer  httpDoer
}

// NewClient creates an HTTPClient adapter.
func NewClient(baseURL, vaultPath string, httpDoer httpDoer) *HTTPClient {
	return &HTTPClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		vaultPath: vaultPath,
		httpDoer:  httpDoer,
	}
}

// Search sends a query to the Omnisearch API and returns processed results.
func (c *HTTPClient) Search(ctx context.Context, query string) ([]Result, error) {
	searchURL := c.baseURL + "/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpDoer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	results := make([]Result, 0, len(items))
	for _, item := range items {
		results = append(results, Result{
			Basename: item.Basename,
			Excerpt:  item.Excerpt,
			Score:    item.Score,
			AbsPath:  filepath.Join(c.vaultPath, trimLeadingSlashes(item.Path)),
		})
	}

	return results, nil
}

func trimLeadingSlashes(s string) string {
	for len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	return s
}
