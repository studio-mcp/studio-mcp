package tool

import (
	"fmt"
	"strings"
	"testing"

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
			name:           "handles non-existent command",
			args:           []string{"this-command-does-not-exist-12345"},
			expectError:    true,
			containsOutput: "Studio error:",
		},
		{
			name:         "handles empty command",
			args:         []string{""},
			expectError:  true,
			expectOutput: "Studio error: Empty command provided",
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
		args           map[string]interface{}
		expectText     string
		expectIsError  bool
		expectContains string
	}{
		{
			name:       "creates function for simple command",
			blueprint:  &MockBlueprint{commandArgs: []string{"echo", "hello"}},
			args:       map[string]interface{}{},
			expectText: "hello",
		},
		{
			name:       "creates function with template arguments",
			blueprint:  &MockBlueprint{commandArgs: []string{"echo", "Hello World"}},
			args:       map[string]interface{}{"message": "Hello World"},
			expectText: "Hello World",
		},
		{
			name:          "handles command errors",
			blueprint:     &MockBlueprint{commandArgs: []string{"false"}},
			args:          map[string]interface{}{},
			expectText:    "",
			expectIsError: true,
		},
		{
			name:           "handles blueprint validation errors",
			blueprint:      &MockBlueprintWithError{err: fmt.Errorf("missing required parameter: name")},
			args:           map[string]interface{}{},
			expectIsError:  true,
			expectContains: "Validation error: missing required parameter: name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := CreateToolFunction(tt.blueprint)
			result := fn(tt.args)

			assert.Len(t, result.Content, 1)
			assert.Equal(t, "text", result.Content[0]["type"])

			text := result.Content[0]["text"].(string)

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

// MockBlueprintWithError is a test helper that returns an error
type MockBlueprintWithError struct {
	err error
}

func (m *MockBlueprintWithError) BuildCommandArgs(args map[string]interface{}) ([]string, error) {
	return nil, m.err
}
