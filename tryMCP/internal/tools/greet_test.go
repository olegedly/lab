package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"trymcp/internal/tools"
)

func TestHandleGreet(t *testing.T) {
	t.Run("returns greeting with the given name", func(t *testing.T) {
		result, _, err := tools.HandleGreet(context.Background(), nil, tools.GreetInput{Name: "MCP"})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		tc, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Equal(t, "Hello, MCP!", tc.Text)
	})

	t.Run("produces output for empty name", func(t *testing.T) {
		result, _, err := tools.HandleGreet(context.Background(), nil, tools.GreetInput{Name: ""})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		tc, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Equal(t, "Hello, !", tc.Text)
	})
}

func TestGreetInputValidation(t *testing.T) {
	schema, err := jsonschema.For[tools.GreetInput](nil)
	require.NoError(t, err)
	require.NotNil(t, schema.Properties["name"])

	schema.Properties["name"].MinLength = jsonschema.Ptr(1)

	resolved, err := schema.Resolve(nil)
	require.NoError(t, err)

	t.Run("accepts non-empty name", func(t *testing.T) {
		err := resolved.Validate(map[string]any{"name": "Alice"})
		assert.NoError(t, err)
	})

	t.Run("rejects empty name", func(t *testing.T) {
		err := resolved.Validate(map[string]any{"name": ""})
		assert.Error(t, err)
	})

	t.Run("rejects missing name", func(t *testing.T) {
		err := resolved.Validate(map[string]any{})
		assert.Error(t, err)
	})
}

func TestGreetSchemaInJSON(t *testing.T) {
	schema, err := jsonschema.For[tools.GreetInput](nil)
	require.NoError(t, err)

	_, err = json.Marshal(schema)
	require.NoError(t, err)
}
