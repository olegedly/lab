# tryMCP

An experimental [Model Context Protocol](https://modelcontextprotocol.io) server built in Go. Provides MCP tools over stdio transport for use with MCP hosts (Claude Desktop, Claude Code, etc.).

## Tools

- **greet** — Say hello to someone. Minimal example of tool registration with JSON Schema input validation.
- **weather** — Current weather and up to 3-day forecast for any city. Uses the [Open-Meteo](https://open-meteo.com/) geocoding and forecast APIs. Returns both human-readable text and structured data.
- **country_summary** — Capital, currencies, and population for a specific country via the [REST Countries](https://restcountries.com/) API.
- **all_countries** — Capital, currencies, and population for every country at once.

## Usage

Run the server:

```bash
go run ./cmd/server
```

By default it listens on stdio. Configure it as a stdio MCP server in your MCP client by pointing to the compiled binary or `go run ./cmd/server`.

Build a binary:

```bash
go build -o server ./cmd/server
```

## Project structure

```
cmd/server/main.go          — entry point, server setup, tool registration
internal/tools/             — tool handlers (greet, weather, country)
internal/weather/client.go  — Open-Meteo API client (geocoding + forecast)
internal/restcountries/     — REST Countries API client
```

## Stack

- **Language:** Go 1.26
- **SDK:** [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)
