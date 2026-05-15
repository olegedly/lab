package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"diary-rag/pkg/embed"
	"diary-rag/pkg/search"
)

type SearchInput struct {
	Query  string `json:"query" jsonschema:"natural language query describing the topic, event, or feeling to search for,required"`
	TopK   int    `json:"top_k" jsonschema:"number of results to return (default 5, max 50)"`
	Before string `json:"before,omitempty" jsonschema:"only return entries dated on or before this date (YYYY-MM-DD). When set, only 'daily' journal entries are searched."`
	After  string `json:"after,omitempty" jsonschema:"only return entries dated on or after this date (YYYY-MM-DD). When set, only 'daily' journal entries are searched."`
}

type SearchOutput struct {
	Results []search.SearchResult `json:"results"`
}

type ReindexInput struct{}

type ReindexOutput struct {
	ChunksCount int    `json:"chunks_count"`
	CorpusPath  string `json:"corpus_path"`
}

type ReadNoteInput struct {
	Path     string `json:"path" jsonschema:"directory path relative to the diary root, from search_diary results,required"`
	Filename string `json:"filename" jsonschema:"filename of the note, from search_diary results,required"`
}

type ReadNoteOutput struct {
	Content string `json:"content"`
}

func RegisterSearchTool(server *mcp.Server, app *App) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_diary",
		Description: "Search diary/journal entries by semantic similarity (not keyword matching). Use when you need to find entries about a specific topic, event, person, or feeling. Supports optional date filtering via 'before' and 'after' (YYYY-MM-DD). Returns text snippets with relevance scores (0-1), dates, journal types, and file identifiers. The returned 'path' and 'filename' fields map directly to read_note inputs.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
		if input.TopK <= 0 || input.TopK > 50 {
			input.TopK = 5
		}

		queryEmbeddings, err := embed.GetEmbeddings([]string{input.Query}, app.embedURL, app.embedModel)
		if err != nil {
			return errorResult[SearchOutput](fmt.Sprintf("embedding error: %v", err))
		}

		opts := search.SearchOptions{
			TopK:   input.TopK,
			Before: input.Before,
			After:  input.After,
		}
		results := app.searcher.Search(queryEmbeddings[0], opts)
		return successResult(SearchOutput{Results: results})
	})
	return nil
}

func RegisterReindexTool(server *mcp.Server, app *App) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "reindex_diary",
		Description: "Rebuild the full diary search index. Call this after adding new diary entries or editing existing ones so the index reflects the latest content. Rescans all markdown files, recomputes all embeddings via the API, and atomically replaces the in-memory search index. May take some time depending on how many files there are. Also available as a one-shot CLI command: diary-rag --reindex",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ ReindexInput) (*mcp.CallToolResult, ReindexOutput, error) {
		if err := BuildEmbeddingIndex(app.rootDir, app.embedURL, app.embedModel, app.outputDir); err != nil {
			return errorResult[ReindexOutput](fmt.Sprintf("reindex error: %v", err))
		}

		corpusPath := filepath.Join(app.outputDir, "chunksWithEmbeddings.json")
		s, err := search.LoadFromFile(corpusPath)
		if err != nil {
			return errorResult[ReindexOutput](fmt.Sprintf("reindex succeeded but failed to reload corpus: %v", err))
		}
		app.searcher = s

		return successResult(ReindexOutput{
			ChunksCount: s.ChunkCount(),
			CorpusPath:  corpusPath,
		})
	})
	return nil
}

func RegisterReadNoteTool(server *mcp.Server, app *App) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_note",
		Description: "Read the full contents of a diary note by its directory path and filename. Use the 'path' and 'filename' values returned by search_diary — they map directly to these inputs. The note is resolved relative to the diary root directory.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ReadNoteInput) (*mcp.CallToolResult, ReadNoteOutput, error) {
		target := filepath.Join(app.rootDir, input.Path, input.Filename)
		content, err := os.ReadFile(target)
		if err != nil {
			return errorResult[ReadNoteOutput](fmt.Sprintf("read error: %v", err))
		}
		return successResult(ReadNoteOutput{Content: string(content)})
	})
	return nil
}

func successResult[T any](output T) (*mcp.CallToolResult, T, error) {
	raw, _ := json.Marshal(output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(raw)},
		},
		StructuredContent: output,
	}, output, nil
}

func errorResult[T any](msg string) (*mcp.CallToolResult, T, error) {
	var zero T
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}, zero, nil
}
