package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"trymcp/internal/tools"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "oleg-mcp"}, nil)

	if err := tools.RegisterGreetTool(server); err != nil {
		log.Fatalf("Failed to register greet tool: %v", err)
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "weather",
		Description: "Get current weather and short forecast for a city",
	}, tools.HandleWeather)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "country_summary",
		Description: "Get capital, currencies, and population for a country",
	}, tools.HandleCountrySummary)

	server.AddTool(&mcp.Tool{
		Name:        "all_countries",
		Description: "List capitals, currencies, and population for all countries",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	}, tools.HandleAllCountries)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}
