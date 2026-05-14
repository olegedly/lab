package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const cacheTTL = 5 * time.Minute

type cacheEntry struct {
	Data string
	TS   time.Time
}

type ttlCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
}

func newTTLCache() *ttlCache {
	return &ttlCache{
		items: make(map[string]cacheEntry),
	}
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])[:16]
}

func (c *ttlCache) Get(raw string) (string, bool) {
	key := hashKey(raw)

	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return "", false
	}
	if time.Since(entry.TS) >= cacheTTL {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return "", false
	}
	return entry.Data, true
}

func (c *ttlCache) Set(raw, data string) {
	key := hashKey(raw)

	c.mu.Lock()
	c.items[key] = cacheEntry{
		Data: data,
		TS:   time.Now(),
	}
	c.mu.Unlock()
}

type SearchInput struct {
	Query string  `json:"query"`
	Count float64 `json:"count"` // Decoding unmarshals numbers to float64 by default
}

type FetchInput struct {
	URL string `json:"url"`
}

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

var (
	cache      = newTTLCache()
	wsRe       = regexp.MustCompile(`\s+`)
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "web-search-mcp",
			Version: "1.0.0",
		},
		nil,
	)

	// Explicitly structure the JSON schema to bypass the strict validation
	searchSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
			"count": map[string]any{
				"type":        "number",
				"description": "Number of results (1-10)",
				"default":     5,
			},
		},
		"required": []string{"query"},
	}

	fetchSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to fetch",
			},
		},
		"required": []string{"url"},
	}

	server.AddTool(&mcp.Tool{
		Name:        "web_search",
		Description: "Search the web using DuckDuckGo. Returns up to 10 results with titles, URLs, and snippets.",
		InputSchema: searchSchema,
	}, handleWebSearch)

	server.AddTool(&mcp.Tool{
		Name:        "web_fetch",
		Description: "Fetch a URL and extract its readable content. Best for documentation, articles, and reference pages.",
		InputSchema: fetchSchema,
	}, handleWebFetch)

	// Note: You can remove the jsonschema dependency from go.mod entirely now
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func handleWebSearch(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in SearchInput
	if err := decodeArgs(req, &in); err != nil {
		return textResult("Search error: invalid arguments: " + err.Error()), nil
	}

	in.Query = strings.TrimSpace(in.Query)
	if in.Query == "" {
		return textResult("Search error: query is required"), nil
	}

	count := int(in.Count)
	if count <= 0 {
		count = 5
	}
	if count > 10 {
		count = 10
	}

	cacheKey := fmt.Sprintf("search:%s:%d", in.Query, count)
	if cached, ok := cache.Get(cacheKey); ok {
		return textResult(cached), nil
	}

	results, err := duckDuckGoSearch(ctx, in.Query, count)
	if err != nil {
		return textResult("Search error: " + err.Error()), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Search results for: %q\n\n", in.Query)
	for i, r := range results {
		fmt.Fprintf(&b, "%d. %s\n   URL: %s\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
	}
	fmt.Fprintf(&b, "---\nResults: %d", len(results))

	out := b.String()
	cache.Set(cacheKey, out)
	return textResult(out), nil
}

func handleWebFetch(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in FetchInput
	if err := decodeArgs(req, &in); err != nil {
		return textResult("Fetch error: invalid arguments: " + err.Error()), nil
	}

	in.URL = strings.TrimSpace(in.URL)
	if in.URL == "" {
		return textResult("Fetch error: url is required"), nil
	}

	cacheKey := "fetch:" + in.URL
	if cached, ok := cache.Get(cacheKey); ok {
		return textResult(cached), nil
	}

	content, err := fetchAndExtract(ctx, in.URL)
	if err != nil {
		return textResult("Fetch error: " + err.Error()), nil
	}

	cache.Set(cacheKey, content)
	return textResult(content), nil
}

func duckDuckGoSearch(ctx context.Context, query string, count int) ([]SearchResult, error) {
	endpoint := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MCP-Web-Search/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, count)

	doc.Find(".result").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if len(results) >= count {
			return false
		}

		title := cleanText(s.Find(".result__title .result__a").First().Text())
		link, _ := s.Find(".result__title .result__a").First().Attr("href")

		if link == "" {
			link, _ = s.Find(".result__url").First().Attr("href")
		}
		if link == "" {
			link = cleanText(s.Find(".result__url").First().Text())
		}

		snippet := cleanText(s.Find(".result__snippet").First().Text())

		if title == "" && link == "" && snippet == "" {
			return true
		}

		results = append(results, SearchResult{
			Title:   title,
			URL:     link,
			Snippet: snippet,
		})
		return true
	})

	return results, nil
}

func fetchAndExtract(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MCP-Web-Search/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return extractContent(string(body), rawURL)
}

func extractContent(html, pageURL string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	doc.Find("script, style, nav, footer, header, iframe, noscript").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	title := cleanText(doc.Find("title").First().Text())
	body := cleanText(doc.Find("body").First().Text())

	if len(body) > 5000 {
		body = body[:5000]
	}

	return fmt.Sprintf("URL: %s\nTitle: %s\n\n%s", pageURL, title, body), nil
}

func cleanText(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

func decodeArgs(req *mcp.CallToolRequest, dst any) error {
	b, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: text,
			},
		},
	}
}
