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
	debugFlag = false
	versionFlag = false

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Print debug logs to stderr to diagnose MCP server issues")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Show version information")
}

// testSetup resets rootCmd and sets output buffers
func testSetup(args ...string) (*bytes.Buffer, error) {
	resetRootCmd()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return &buf, err
}

func TestHelpFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectLines []string
	}{
		{
			name: "shows help with --help",
			args: []string{"--help"},
			expectLines: []string{
				"studio-mcp is a tool for running a single command MCP server",
				"--debug", "--help", "--version",
			},
		},
		{
			name: "shows help with -h",
			args: []string{"-h"},
			expectLines: []string{
				"studio-mcp is a tool for running a single command MCP server",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := testSetup(tt.args...)
			assert.NoError(t, err)
			output := buf.String()
			for _, line := range tt.expectLines {
				assert.Contains(t, output, line)
			}
		})
	}
}

func TestVersionFlagOutput(t *testing.T) {
	t.Run("prints version info", func(t *testing.T) {
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2023-01-01T00:00:00Z"

		buf, err := testSetup("--version")
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp 1.2.3")
		assert.Contains(t, output, "commit: abc123")
		assert.Contains(t, output, "built: 2023-01-01T00:00:00Z")
	})

	t.Run("handles dev build values", func(t *testing.T) {
		Version, Commit, Date = "dev", "none", "unknown"

		buf, err := testSetup("--version")
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "studio-mcp dev")
		assert.Contains(t, output, "commit: none")
		assert.Contains(t, output, "built: unknown")
	})
}

func TestErrorCases(t *testing.T) {
	t.Run("shows error with no arguments", func(t *testing.T) {
		buf, err := testSetup()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: studio-mcp <command>")
	})

	t.Run("errors when only flag is --debug", func(t *testing.T) {
		buf, err := testSetup("--debug")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: studio-mcp <command>")
	})
}

func TestFlagParsing(t *testing.T) {
	t.Run("parses debug flag", func(t *testing.T) {
		resetRootCmd()
		assert.False(t, debugFlag)

		// Don't execute the command, just parse
		rootCmd.ParseFlags([]string{"--debug", "echo", "hello"})
		assert.True(t, debugFlag)
	})

	t.Run("parses version flag", func(t *testing.T) {
		resetRootCmd()
		assert.False(t, versionFlag)

		rootCmd.ParseFlags([]string{"--version"})
		assert.True(t, versionFlag)
	})
}

func TestHelpTextDetailedContent(t *testing.T) {
	t.Run("includes custom syntax and examples", func(t *testing.T) {
		buf, err := testSetup("--help")
		assert.NoError(t, err)

		output := buf.String()
		lines := []string{
			"the command starts at the first non-flag argument:",
			"<command> - the shell command to run",
			"{{req # required arg}}",
			"[args... # array of args]",
			"[opt # optional string]",
			"https://en.wikipedia.org/wiki/{{wiki_page_name}}",
			"studio-mcp say -v siri",
			"studio-mcp [--debug] <command> --example",
		}

		for _, line := range lines {
			assert.Contains(t, output, line)
		}
	})
}
