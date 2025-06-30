package tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

// Execute runs a command and returns trimmed combined stdout+stderr or an error
func Execute(command string, args ...string) (string, error) {
	debug("Executing command: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Always combine outputs for visibility
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			debug("Command completed with non-zero exit code: %d", exitErr.ExitCode())
			debug("Final output length: %d chars", len(output))
			return output, fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		debug("Spawn error: %s", err.Error())
		return output, fmt.Errorf("Studio error: %w", err)
	}

	debug("Command completed successfully with exit code 0")
	debug("Final output length: %d chars", len(output))

	return output, nil
}

// CreateToolFunction creates a tool function for the given blueprint
func CreateToolFunction(blueprint Blueprint) ToolFunction {
	return func(args map[string]interface{}) ToolResult {
		debug("Tool called with args: %v", args)

		fullCommand, err := blueprint.BuildCommandArgs(args)
		if err != nil {
			return ToolResult{
				Content: []map[string]interface{}{
					{"type": "text", "text": fmt.Sprintf("Validation error: %s", err.Error())},
				},
				IsError: true,
			}
		}

		debug("Built command: %s", strings.Join(fullCommand, " "))

		output, err := Execute(fullCommand[0], fullCommand[1:]...)
		isError := err != nil

		if isError {
			debug("Execution error: %s", err)
		}

		return ToolResult{
			Content: []map[string]interface{}{
				{"type": "text", "text": strings.TrimSpace(output)},
			},
			IsError: isError,
		}
	}
}
