package studio

import (
	"fmt"
	"studio-mcp/internal/blueprint"
	"studio-mcp/internal/tool"
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

// Serve would start the MCP server (placeholder for now)
func (s *Studio) Serve() error {
	// This is where the MCP server logic would go
	// For now, just return nil to indicate success
	return nil
}
