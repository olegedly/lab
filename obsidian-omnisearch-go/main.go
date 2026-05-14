package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func run() error {
	if len(getArgs()) < 2 {
		return errUsage()
	}

	vaultPath := getArgs()[1]

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "obsidian-omnisearch",
		Version: "1.0.0",
	}, nil)

	if err := RegisterObsidianNotesSearchTool(server, vaultPath); err != nil {
		return err
	}

	if err := RegisterReadNoteTool(server); err != nil {
		return err
	}

	return server.Run(context.Background(), &mcp.StdioTransport{})
}
