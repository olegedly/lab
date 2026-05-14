package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"trymcp/internal/tools"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "oleg-mcp"}, nil)

	if err := tools.RegisterGreetTool(server); err != nil {
		log.Fatalf("Failed to register greet tool: %v", err)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}
