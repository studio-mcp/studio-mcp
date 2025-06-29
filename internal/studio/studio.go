package studio

import (
	"context"
	"fmt"
	"studio-mcp/internal/blueprint"
	"studio-mcp/internal/tool"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Studio represents the main application logic
type Studio struct {
	Blueprint *blueprint.Blueprint
	DebugMode bool
}

// New creates a new Studio instance from command arguments
func New(args []string, debugMode bool) (*Studio, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	bp := blueprint.FromArgs(args)

	// Set debug mode on tool
	tool.SetDebugMode(debugMode)

	return &Studio{
		Blueprint: bp,
		DebugMode: debugMode,
	}, nil
}

// Serve starts the MCP server over stdio
func (s *Studio) Serve() error {
	// Create server with basic capabilities
	server := mcp.NewServer("studio-mcp", "1.0.0", nil)

	// Create tool function from blueprint
	toolFunc := tool.CreateToolFunction(s.Blueprint)

	// Create handler that wraps our tool function
	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[map[string]any], error) {
		// Call our tool function
		result := toolFunc(params.Arguments)

		// Convert to MCP result format
		var content []mcp.Content
		for _, contentItem := range result.Content {
			if textType, ok := contentItem["type"].(string); ok && textType == "text" {
				if text, ok := contentItem["text"].(string); ok {
					content = append(content, &mcp.TextContent{Text: text})
				}
			}
		}

		return &mcp.CallToolResultFor[map[string]any]{
			Content: content,
			IsError: result.IsError,
		}, nil
	}

	// Convert blueprint schema to jsonschema.Schema
	schema := &jsonschema.Schema{
		Type: "object",
	}

	if props, ok := s.Blueprint.InputSchema["properties"].(map[string]interface{}); ok && len(props) > 0 {
		schema.Properties = make(map[string]*jsonschema.Schema)

		for name, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				propSchema := &jsonschema.Schema{}

				if propType, ok := propMap["type"].(string); ok {
					propSchema.Type = propType
				}

				if desc, ok := propMap["description"].(string); ok {
					propSchema.Description = desc
				}

				// Handle array type with items
				if propType, ok := propMap["type"].(string); ok && propType == "array" {
					if items, ok := propMap["items"].(map[string]interface{}); ok {
						if itemType, ok := items["type"].(string); ok {
							propSchema.Items = &jsonschema.Schema{Type: itemType}
						}
					}
				}

				schema.Properties[name] = propSchema
			}
		}

		// Add required fields
		if required, ok := s.Blueprint.InputSchema["required"].([]string); ok && len(required) > 0 {
			schema.Required = required
		}
	}

	// Add the tool to the server using NewServerTool with schema
	serverTool := mcp.NewServerTool[map[string]any, map[string]any](
		s.Blueprint.ToolName,
		s.Blueprint.ToolDescription,
		handler,
		mcp.Input(mcp.Schema(schema)),
	)

	server.AddTools(serverTool)

	// Run the server over stdio
	return server.Run(context.Background(), mcp.NewStdioTransport())
}
