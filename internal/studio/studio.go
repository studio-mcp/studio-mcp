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
	Version   string
}

// New creates a new Studio instance from command arguments
func New(args []string, debugMode bool, version string) (*Studio, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	bp, err := blueprint.FromArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to create blueprint: %w", err)
	}

	// Set debug mode on tool
	tool.SetDebugMode(debugMode)

	return &Studio{
		Blueprint: bp,
		DebugMode: debugMode,
		Version:   version,
	}, nil
}

// Serve starts the MCP server over stdio
func (s *Studio) Serve() error {
	// Create server with version from build
	server := mcp.NewServer("studio-mcp", s.Version, nil)

	// Add the tool to the server using NewServerTool with schema directly from blueprint
	serverTool := mcp.NewServerTool(
		s.Blueprint.ToolName,
		s.Blueprint.ToolDescription,
		tool.CreateToolFunction(s.Blueprint),
		mcp.Input(mcp.Schema(s.Blueprint.InputSchema)),
	)

	server.AddTools(serverTool)

	// Run the server over stdio
	return server.Run(context.Background(), mcp.NewStdioTransport())
}
