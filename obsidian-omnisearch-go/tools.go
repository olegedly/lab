package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ObsidianSearchInput struct {
	Query string `json:"query" jsonschema:"search query for Obsidian notes,required"`
}

type ObsidianSearchItem struct {
	Basename string  `json:"basename"`
	Excerpt  string  `json:"excerpt"`
	Score    float64 `json:"score"`
	Path     string  `json:"path"`
}

type ObsidianSearchOutput struct {
	Results []string `json:"results"`
}

type ReadNoteInput struct {
	Filepath string `json:"filepath" jsonschema:"absolute path to the Obsidian note file,required"`
}

type ReadNoteOutput struct {
	Content string `json:"content"`
}

func RegisterObsidianNotesSearchTool(server *mcp.Server, vaultPath string) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "obsidian_notes_search",
		Description: "Search Obsidian notes and return absolute paths to the matching notes. The returned paths can be used with the read_note tool to view the note contents.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ObsidianSearchInput) (*mcp.CallToolResult, ObsidianSearchOutput, error) {
		results := obsidianNotesSearch(ctx, vaultPath, input.Query)
		output := ObsidianSearchOutput{Results: results}

		raw, _ := json.Marshal(output)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(raw)},
			},
			StructuredContent: output,
			IsError:           false,
		}, output, nil
	})

	return nil
}
func RegisterReadNoteTool(server *mcp.Server) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_note",
		Description: "Read and return the contents of an Obsidian note file.",
	}, HandleReadNote)

	return nil
}

func HandleReadNote(ctx context.Context, _ *mcp.CallToolRequest, input ReadNoteInput) (*mcp.CallToolResult, ReadNoteOutput, error) {
	data, err := os.ReadFile(input.Filepath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("read_note error: %v", err)},
			},
			IsError: true,
		}, ReadNoteOutput{}, nil
	}

	output := ReadNoteOutput{Content: string(data)}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output.Content},
		},
		StructuredContent: output,
		IsError:           false,
	}, output, nil
}

func obsidianNotesSearch(ctx context.Context, vaultPath, query string) []string {
	searchURL := "http://localhost:51361/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return []string{}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []string{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}
	}

	var items []ObsidianSearchItem
	if err := json.Unmarshal(body, &items); err != nil {
		return []string{}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	results := make([]string, 0, len(items))
	for _, item := range items {
		absPath := filepath.Join(vaultPath, trimLeadingSlashes(item.Path))
		results = append(results,
			fmt.Sprintf(
				"<title>%s</title>\n<excerpt>%s</excerpt>\n<score>%v</score>\n<filepath>%s</filepath>",
				item.Basename,
				item.Excerpt,
				item.Score,
				absPath,
			),
		)
	}

	return results
}

func trimLeadingSlashes(s string) string {
	for len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	return s
}
