package tools

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GreetInput is the input for the greet tool.
type GreetInput struct {
	Name string `json:"name" jsonschema:"the person to greet"`
}

// HandleGreet handles a greet tool call.
func HandleGreet(_ context.Context, _ *mcp.CallToolRequest, input GreetInput) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Hello, " + input.Name + "!"}},
	}, nil, nil
}

// RegisterGreetTool adds the greet tool to the server with schema validation.
func RegisterGreetTool(server *mcp.Server) error {
	schema, err := jsonschema.For[GreetInput](nil)
	if err != nil {
		return err
	}
	if prop, ok := schema.Properties["name"]; ok {
		prop.MinLength = jsonschema.Ptr(1)
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "greet",
		Description: "Say hello to someone",
		InputSchema: schema,
	}, HandleGreet)
	return nil
}
