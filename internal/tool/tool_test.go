package tool

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTool_Execute(t *testing.T) {
	t.Run("executes simple command successfully", func(t *testing.T) {
		result, err := Execute("echo", "hello")

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "hello", strings.TrimSpace(result.Output))
	})

	t.Run("executes command with multiple arguments", func(t *testing.T) {
		result, err := Execute("echo", "hello", "world")

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "hello world", strings.TrimSpace(result.Output))
	})

	t.Run("captures stderr output", func(t *testing.T) {
		// Using sh -c to run a command that writes to stderr
		result, err := Execute("sh", "-c", "echo 'error message' >&2")

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output, "error message")
	})

	t.Run("handles command failure", func(t *testing.T) {
		result, err := Execute("false")

		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, "", strings.TrimSpace(result.Output))
	})

	t.Run("handles non-existent command", func(t *testing.T) {
		result, err := Execute("this-command-does-not-exist-12345")

		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Output, "Studio error:")
	})

	t.Run("handles empty command", func(t *testing.T) {
		result, err := Execute("")

		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, "Studio error: Empty command provided", result.Output)
	})

	t.Run("handles command with spaces", func(t *testing.T) {
		result, err := Execute("echo", "hello world with spaces")

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "hello world with spaces", strings.TrimSpace(result.Output))
	})
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
	t.Run("creates function for simple command", func(t *testing.T) {
		blueprint := &MockBlueprint{
			commandArgs: []string{"echo", "hello"},
		}

		fn := CreateToolFunction(blueprint)
		result := fn(map[string]interface{}{})

		assert.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0]["type"])
		assert.Equal(t, "hello", result.Content[0]["text"])
		assert.False(t, result.IsError)
	})

	t.Run("creates function with template arguments", func(t *testing.T) {
		blueprint := &MockBlueprint{
			commandArgs: []string{"echo", "Hello World"},
		}

		fn := CreateToolFunction(blueprint)
		result := fn(map[string]interface{}{
			"message": "Hello World",
		})

		assert.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0]["type"])
		assert.Equal(t, "Hello World", result.Content[0]["text"])
		assert.False(t, result.IsError)
	})

	t.Run("handles command errors", func(t *testing.T) {
		blueprint := &MockBlueprint{
			commandArgs: []string{"false"},
		}

		fn := CreateToolFunction(blueprint)
		result := fn(map[string]interface{}{})

		assert.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0]["type"])
		assert.Equal(t, "", result.Content[0]["text"])
		assert.True(t, result.IsError)
	})

	t.Run("handles blueprint validation errors", func(t *testing.T) {
		blueprint := &MockBlueprintWithError{
			err: fmt.Errorf("missing required parameter: name"),
		}

		fn := CreateToolFunction(blueprint)
		result := fn(map[string]interface{}{})

		assert.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0]["type"])
		assert.Contains(t, result.Content[0]["text"], "Validation error: missing required parameter: name")
		assert.True(t, result.IsError)
	})
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
