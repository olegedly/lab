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
	Query string `json:"query" jsonschema:"search query,required"`
	TopK  int    `json:"top_k" jsonschema:"maximum number of results (default 5),maximum=50"`
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
	Filepath string `json:"filepath" jsonschema:"absolute path to the diary note file,required"`
}

type ReadNoteOutput struct {
	Content string `json:"content"`
}

func RegisterSearchTool(server *mcp.Server, app *App) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_diary",
		Description: "Search diary entries by semantic similarity. Returns matching text chunks with relevance scores and metadata.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
		if input.TopK <= 0 || input.TopK > 50 {
			input.TopK = 5
		}

		queryEmbeddings, err := embed.GetEmbeddings([]string{input.Query}, app.embedURL, app.embedModel)
		if err != nil {
			return errorResult[SearchOutput](fmt.Sprintf("embedding error: %v", err))
		}

		results := app.searcher.Search(queryEmbeddings[0], input.TopK)
		return successResult(SearchOutput{Results: results})
	})
	return nil
}

func RegisterReindexTool(server *mcp.Server, app *App) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "reindex_diary",
		Description: "Re-scan all diary markdown files, recompute embeddings, and rebuild the search index.",
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

func RegisterReadNoteTool(server *mcp.Server) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_note",
		Description: "Read the full content of a diary note by file path.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ReadNoteInput) (*mcp.CallToolResult, ReadNoteOutput, error) {
		content, err := os.ReadFile(input.Filepath)
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
