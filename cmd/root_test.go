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
	// Re-add the debug flag
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Print debug logs to stderr to diagnose MCP server issues")
}

func TestRootCommand(t *testing.T) {
	// Reset before each test group
	resetRootCmd()

	t.Run("shows error with no arguments", func(t *testing.T) {
		rootCmd.SetArgs([]string{})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: studio-mcp <command>")
	})

	t.Run("shows help with --help flag", func(t *testing.T) {
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
	})

	t.Run("shows help with -h flag", func(t *testing.T) {
		rootCmd.SetArgs([]string{"-h"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp is a tool for running a single command MCP server")
	})

	t.Run("accepts --debug flag", func(t *testing.T) {
		// Reset command to ensure clean state
		resetRootCmd()

		rootCmd.SetArgs([]string{"--debug", "echo", "hello"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)

		// Execute normally - the command should accept these args
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Verify the debug flag was set
		assert.True(t, debugFlag)
	})

	t.Run("shows error when only flags provided", func(t *testing.T) {
		// Reset command to ensure clean state
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
	// Reset before test
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
