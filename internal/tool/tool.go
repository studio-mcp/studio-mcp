package tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of executing a command
type Result struct {
	Output  string
	Success bool
}

// ToolResult represents the result returned to MCP
type ToolResult struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"isError"`
}

// Blueprint interface defines what we need from a blueprint
type Blueprint interface {
	BuildCommandArgs(args map[string]interface{}) ([]string, error)
}

// ToolFunction is the function signature for MCP tools
type ToolFunction func(args map[string]interface{}) ToolResult

var debugMode bool

// SetDebugMode enables or disables debug mode
func SetDebugMode(enabled bool) {
	debugMode = enabled
}

// IsDebugMode returns whether debug mode is enabled
func IsDebugMode() bool {
	return debugMode && os.Getenv("NODE_ENV") != "test"
}

// debug logs a message to stderr if debug mode is enabled
func debug(format string, args ...interface{}) {
	if IsDebugMode() {
		fmt.Fprintf(os.Stderr, "[Studio MCP] "+format+"\n", args...)
	}
}

// Execute runs a command and returns the result
func Execute(command string, args ...string) (*Result, error) {
	debug("Executing command: %s %s", command, strings.Join(args, " "))

	// Handle empty command
	if command == "" || strings.TrimSpace(command) == "" {
		errorMsg := "Studio error: Empty command provided"
		debug("Error: %s", errorMsg)
		return &Result{
			Output:  errorMsg,
			Success: false,
		}, nil
	}

	// Create the command
	cmd := exec.Command(command, args...)

	// Capture both stdout and stderr
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Combine stdout and stderr
	output := stdout.String()
	if stderr.String() != "" {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	// Check for execution errors
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Command executed but returned non-zero exit code
			debug("Command completed with exit code: %d", cmd.ProcessState.ExitCode())
			debug("Final output length: %d chars", len(output))
			return &Result{
				Output:  strings.TrimSpace(output),
				Success: false,
			}, nil
		}
		// Command failed to execute (e.g., command not found)
		errorMsg := fmt.Sprintf("Studio error: %s", err.Error())
		debug("Spawn error: %s", errorMsg)
		return &Result{
			Output:  errorMsg,
			Success: false,
		}, nil
	}

	debug("Command completed with exit code: 0")
	debug("Final output length: %d chars", len(output))

	return &Result{
		Output:  strings.TrimSpace(output),
		Success: true,
	}, nil
}

// CreateToolFunction creates a tool function for the given blueprint
func CreateToolFunction(blueprint Blueprint) ToolFunction {
	return func(args map[string]interface{}) ToolResult {
		debug("Tool called with args: %v", args)

		fullCommand, err := blueprint.BuildCommandArgs(args)
		if err != nil {
			// Validation error - return immediately
			return ToolResult{
				Content: []map[string]interface{}{
					{"type": "text", "text": fmt.Sprintf("Validation error: %s", err.Error())},
				},
				IsError: true,
			}
		}

		debug("Built command: %s", strings.Join(fullCommand, " "))

		// Execute the command
		result, err := Execute(fullCommand[0], fullCommand[1:]...)
		if err != nil {
			// Should not happen as Execute returns errors in Result
			return ToolResult{
				Content: []map[string]interface{}{
					{"type": "text", "text": err.Error()},
				},
				IsError: true,
			}
		}

		debug("Tool result - success: %v, output length: %d", result.Success, len(result.Output))

		return ToolResult{
			Content: []map[string]interface{}{
				{"type": "text", "text": result.Output},
			},
			IsError: !result.Success,
		}
	}
}
