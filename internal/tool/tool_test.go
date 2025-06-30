package tool

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestTool_Execute(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectOutput   string
		containsOutput string
	}{
		{
			name:         "executes simple command successfully",
			args:         []string{"echo", "hello"},
			expectOutput: "hello",
		},
		{
			name:         "executes command with multiple arguments",
			args:         []string{"echo", "hello", "world"},
			expectOutput: "hello world",
		},
		{
			name:           "captures stderr output",
			args:           []string{"sh", "-c", "echo 'error message' >&2"},
			containsOutput: "error message",
		},
		{
			name:        "handles command failure",
			args:        []string{"false"},
			expectError: true,
		},
		{
			name:        "handles non-existent command",
			args:        []string{"this-command-does-not-exist-12345"},
			expectError: true,
		},
		{
			name:         "handles command with spaces",
			args:         []string{"echo", "hello world with spaces"},
			expectOutput: "hello world with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Execute(tt.args[0], tt.args[1:]...)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectOutput != "" {
				assert.Equal(t, tt.expectOutput, strings.TrimSpace(output))
			}

			if tt.containsOutput != "" {
				assert.Contains(t, output, tt.containsOutput)
			}
		})
	}
}

func TestTool_DebugMode(t *testing.T) {
	t.Run("debug mode is off by default", func(t *testing.T) {
		assert.False(t, IsDebugMode())
	})

	t.Run("can enable debug mode", func(t *testing.T) {
		SetDebugMode(true)
		assert.True(t, IsDebugMode())

		// Reset for other tests
		SetDebugMode(false)
		assert.False(t, IsDebugMode())
	})
}

func TestTool_CreateToolFunction(t *testing.T) {
	tests := []struct {
		name           string
		blueprint      Blueprint
		args           map[string]any
		expectText     string
		expectIsError  bool
		expectContains string
	}{
		{
			name:       "creates function for simple command",
			blueprint:  &MockBlueprint{commandArgs: []string{"echo", "hello"}},
			args:       map[string]any{},
			expectText: "hello",
		},
		{
			name:       "creates function with template arguments",
			blueprint:  &MockBlueprint{commandArgs: []string{"echo", "Hello World"}},
			args:       map[string]any{"message": "Hello World"},
			expectText: "Hello World",
		},
		{
			name:          "handles command errors",
			blueprint:     &MockBlueprint{commandArgs: []string{"false"}},
			args:          map[string]any{},
			expectText:    "",
			expectIsError: true,
		},
		{
			name:           "handles command errors with stderr output",
			blueprint:      &MockBlueprint{commandArgs: []string{"sh", "-c", "echo 'error message' >&2; exit 1"}},
			args:           map[string]any{},
			expectContains: "error message",
			expectIsError:  true,
		},
		{
			name:           "handles blueprint validation errors",
			blueprint:      &MockBlueprintWithError{err: fmt.Errorf("missing required parameter: name")},
			args:           map[string]any{},
			expectIsError:  true,
			expectContains: "Validation error: missing required parameter: name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CreateToolFunction(tt.blueprint)

			// Create MCP parameters
			params := &mcp.CallToolParamsFor[map[string]any]{
				Arguments: tt.args,
			}

			result, err := handler(context.Background(), nil, params)

			assert.NoError(t, err)
			assert.Len(t, result.Content, 1)

			// Cast to TextContent to access the text
			textContent, ok := result.Content[0].(*mcp.TextContent)
			assert.True(t, ok, "Expected content to be TextContent")

			text := textContent.Text

			if tt.expectContains != "" {
				assert.Contains(t, text, tt.expectContains)
			} else {
				assert.Equal(t, tt.expectText, text)
			}

			assert.Equal(t, tt.expectIsError, result.IsError)
		})
	}
}

// MockBlueprint is a test helper that implements the Blueprint interface
type MockBlueprint struct {
	commandArgs []string
}

func (m *MockBlueprint) BuildCommandArgs(args map[string]interface{}) ([]string, error) {
	return m.commandArgs, nil
}

func (m *MockBlueprint) GetToolName() string {
	return "mock_tool"
}

func (m *MockBlueprint) GetToolDescription() string {
	return "Mock tool for testing"
}

func (m *MockBlueprint) GetInputSchema() interface{} {
	return &struct{}{} // Empty schema for testing
}

// MockBlueprintWithError is a test helper that returns an error
type MockBlueprintWithError struct {
	err error
}

func (m *MockBlueprintWithError) BuildCommandArgs(args map[string]interface{}) ([]string, error) {
	return nil, m.err
}

func (m *MockBlueprintWithError) GetToolName() string {
	return "mock_error_tool"
}

func (m *MockBlueprintWithError) GetToolDescription() string {
	return "Mock error tool for testing"
}

func (m *MockBlueprintWithError) GetInputSchema() interface{} {
	return &struct{}{} // Empty schema for testing
}
