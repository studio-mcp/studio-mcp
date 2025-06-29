package studio

import (
	"context"
	"fmt"
	"studio-mcp/internal/blueprint"
	"studio-mcp/internal/tool"

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

	// Add the tool to the server using NewServerTool with schema directly from blueprint
	serverTool := mcp.NewServerTool[map[string]any, map[string]any](
		s.Blueprint.ToolName,
		s.Blueprint.ToolDescription,
		handler,
		mcp.Input(mcp.Schema(s.Blueprint.InputSchema)),
	)

	server.AddTools(serverTool)

	// Run the server over stdio
	return server.Run(context.Background(), mcp.NewStdioTransport())
}
