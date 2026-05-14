package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"obsidian-omnisearch-go/internal/note"
	"obsidian-omnisearch-go/internal/omnisearch"
)

const omnisearchBaseURL = "http://localhost:51361"

func main() {
	if err := run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) < 1 {
		return errors.New("usage: obsidian-omnisearch <obsidian_vault_path>")
	}

	vaultPath := args[0]

	searchClient := omnisearch.NewClient(omnisearchBaseURL, vaultPath, http.DefaultClient)
	noteReader := note.NewReader()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "obsidian-omnisearch",
		Version: "1.0.0",
	}, nil)

	if err := RegisterObsidianNotesSearchTool(server, searchClient); err != nil {
		return err
	}

	if err := RegisterReadNoteTool(server, noteReader); err != nil {
		return err
	}

	return server.Run(context.Background(), &mcp.StdioTransport{})
}
