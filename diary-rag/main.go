package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"diary-rag/pkg/search"
)

// resolvePath makes a relative path absolute by resolving against the binary's
// own directory. This ensures the binary works regardless of the caller's
// working directory (important for MCP stdio transport).
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	exe, err := os.Executable()
	if err != nil {
		return path
	}
	return filepath.Join(filepath.Dir(exe), path)
}

// App holds the application state shared across MCP tool handlers.
type App struct {
	searcher   *search.Searcher
	rootDir    string
	embedURL   string
	embedModel string
	outputDir  string
}

func main() {
	rootDir := flag.String("dir", "/home/morket/0mni/1. Life Admin/12. Logging/11. Diary", "root directory of markdown diary files")
	embedURL := flag.String("embed-url", "http://192.168.1.5:5001/v1/embeddings", "embedding API URL")
	embedModel := flag.String("embed-model", "nomic-embed-text-v1.5", "embedding model name")
	outputDir := flag.String("output", "output", "output directory for JSON files")
	flag.Parse()

	app := &App{
		rootDir:    resolvePath(*rootDir),
		embedURL:   *embedURL,
		embedModel: *embedModel,
		outputDir:  resolvePath(*outputDir),
	}

	corpusPath := filepath.Join(app.outputDir, "chunksWithEmbeddings.json")
	searcher, err := search.LoadFromFile(corpusPath)
	if err != nil {
		log.Fatalf("Failed to load corpus: %v. Run the reindex_diary tool first.", err)
	}
	app.searcher = searcher

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "diary-rag",
		Version: "1.0.0",
	}, nil)

	RegisterSearchTool(server, app)
	RegisterReindexTool(server, app)
	RegisterReadNoteTool(server, app)

	fmt.Fprintf(os.Stderr, "diary-rag MCP server started (dir=%s, output=%s)\n", app.rootDir, app.outputDir)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
