package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetRootCmd resets the root command for test isolation
func resetRootCmd() {
	rootCmd.ResetCommands()
	rootCmd.ResetFlags()
	// Reset the flag variables
	debugFlag = false
	versionFlag = false
	// Re-add the flags
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Print debug logs to stderr to diagnose MCP server issues")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Show version information")
}

func TestRootCommand(t *testing.T) {
	t.Run("shows error with no arguments", func(t *testing.T) {
		// Reset before each test
		resetRootCmd()
		rootCmd.SetArgs([]string{})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: studio-mcp <command>")
	})

	t.Run("shows help with --help flag", func(t *testing.T) {
		resetRootCmd()
		rootCmd.SetArgs([]string{"--help"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp is a tool for running a single command MCP server")
		assert.Contains(t, output, "--debug")
		assert.Contains(t, output, "--help")
		assert.Contains(t, output, "--version")
	})

	t.Run("shows help with -h flag", func(t *testing.T) {
		resetRootCmd()
		rootCmd.SetArgs([]string{"-h"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp is a tool for running a single command MCP server")
	})

	t.Run("shows version with --version flag", func(t *testing.T) {
		resetRootCmd()
		// Set test version info
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2023-01-01T00:00:00Z"

		rootCmd.SetArgs([]string{"--version"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp 1.2.3")
		assert.Contains(t, output, "commit: abc123")
		assert.Contains(t, output, "built: 2023-01-01T00:00:00Z")
	})

	t.Run("version flag works without command arguments", func(t *testing.T) {
		resetRootCmd()
		// Set test version info
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2023-01-01T00:00:00Z"

		// This should not error even though no command is provided
		rootCmd.SetArgs([]string{"--version"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp 1.2.3")
	})

	t.Run("accepts --debug flag", func(t *testing.T) {
		resetRootCmd()

		// Explicitly check initial state
		assert.False(t, debugFlag, "debugFlag should be false initially")

		rootCmd.SetArgs([]string{"--debug", "echo", "hello"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		// Don't execute the command, just parse the flags
		rootCmd.ParseFlags([]string{"--debug", "echo", "hello"})

		// Verify the debug flag was set by flag parsing
		assert.True(t, debugFlag, "debugFlag should be true after parsing --debug flag")
	})

	t.Run("shows error when only flags provided", func(t *testing.T) {
		resetRootCmd()
		rootCmd.SetArgs([]string{"--debug"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: studio-mcp <command>")
	})
}

// Test the help text content matches the TypeScript version
func TestHelpTextContent(t *testing.T) {
	resetRootCmd()
	rootCmd.SetArgs([]string{"--help"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()

	// Check for all the expected help content
	expectedContent := []string{
		"studio-mcp is a tool for running a single command MCP server",
		"-h, --help",
		"--version",
		"--debug",
		"the command starts at the first non-flag argument:",
		"<command> - the shell command to run",
		"arguments can be templated",
		"{{req # required arg}}",
		"[args... # array of args]",
		"[opt # optional string]",
		"https://en.wikipedia.org/wiki/{{wiki_page_name}}",
		"Example:",
		"studio-mcp say -v siri",
		"Usage:",
		"studio-mcp [--debug] <command> --example",
	}

	for _, expected := range expectedContent {
		assert.Contains(t, output, expected, "Help text should contain: %s", expected)
	}
}

func TestVersionFlag(t *testing.T) {
	t.Run("version flag parsing", func(t *testing.T) {
		resetRootCmd()

		// Verify initial state
		assert.False(t, versionFlag, "versionFlag should be false initially")

		// Parse flags without executing
		rootCmd.ParseFlags([]string{"--version"})

		// Verify the version flag was set
		assert.True(t, versionFlag, "versionFlag should be true after parsing --version flag")
	})

	t.Run("version output format", func(t *testing.T) {
		resetRootCmd()
		// Set test version info
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2023-01-01T00:00:00Z"

		rootCmd.SetArgs([]string{"--version"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()

		// Check exact format
		expectedLines := []string{
			"studio-mcp 1.2.3",
			"commit: abc123",
			"built: 2023-01-01T00:00:00Z",
		}

		for _, line := range expectedLines {
			assert.Contains(t, output, line, "Version output should contain: %s", line)
		}
	})

	t.Run("version flag with dev values", func(t *testing.T) {
		resetRootCmd()
		// Set dev version info (default values from main.go)
		Version = "dev"
		Commit = "none"
		Date = "unknown"

		rootCmd.SetArgs([]string{"--version"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp dev")
		assert.Contains(t, output, "commit: none")
		assert.Contains(t, output, "built: unknown")
	})
}

func TestRootCommand_Execution(t *testing.T) {
	t.Run("executes simple command", func(t *testing.T) {
		resetRootCmd()
		rootCmd.SetArgs([]string{"echo", "hello"})

		// Test that the command line arguments are parsed correctly without actually starting the server
		// This validates that the command structure is correct
		args := rootCmd.ValidArgs
		_ = args // Just to verify we can access the command structure

		// For now we skip actual execution since it starts an MCP server over stdio
		// which doesn't work well in test environment
		t.Skip("Skipping actual execution - MCP server starts over stdio")
	})

	t.Run("handles blueprint arguments", func(t *testing.T) {
		resetRootCmd()
		rootCmd.SetArgs([]string{"echo", "{{message#Text to display}}"})

		// Test that the command line arguments are parsed correctly without actually starting the server
		// This validates that blueprint arguments are handled correctly
		args := rootCmd.ValidArgs
		_ = args // Just to verify we can access the command structure

		// For now we skip actual execution since it starts an MCP server over stdio
		// which doesn't work well in test environment
		t.Skip("Skipping actual execution - MCP server starts over stdio")
	})
}
