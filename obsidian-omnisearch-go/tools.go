package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"obsidian-omnisearch-go/internal/note"
	"obsidian-omnisearch-go/internal/omnisearch"
)

type ObsidianSearchInput struct {
	Query string `json:"query" jsonschema:"search query for Obsidian notes,required"`
}

type ObsidianSearchOutput struct {
	Results []omnisearch.Result `json:"results"`
}

type ReadNoteInput struct {
	Filepath string `json:"filepath" jsonschema:"absolute path to the Obsidian note file,required"`
}

type ReadNoteOutput struct {
	Content string `json:"content"`
}

func RegisterObsidianNotesSearchTool(server *mcp.Server, client omnisearch.Client) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "obsidian_notes_search",
		Description: "Search Obsidian notes and return absolute paths to the matching notes. The returned paths can be used with the read_note tool to view the note contents.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ObsidianSearchInput) (*mcp.CallToolResult, ObsidianSearchOutput, error) {
		results, err := client.Search(ctx, input.Query)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("search error: %v", err)},
				},
				IsError: true,
			}, ObsidianSearchOutput{}, nil
		}

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

func RegisterReadNoteTool(server *mcp.Server, reader note.Reader) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_note",
		Description: "Read and return the contents of an Obsidian note file.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ReadNoteInput) (*mcp.CallToolResult, ReadNoteOutput, error) {
		content, err := reader.Read(ctx, input.Filepath)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("read_note error: %v", err)},
				},
				IsError: true,
			}, ReadNoteOutput{}, nil
		}

		output := ReadNoteOutput{Content: content}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output.Content},
			},
			StructuredContent: output,
			IsError:           false,
		}, output, nil
	})

	return nil
}
