package blueprint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlueprint_ParseSimpleCommand(t *testing.T) {
	t.Run("parses simple command without args", func(t *testing.T) {
		bp := FromArgs([]string{"git", "status"})

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, "git", bp.ToolName)
		assert.Equal(t, "Run the shell command `git status`", bp.ToolDescription)
		assert.Equal(t, map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}, bp.InputSchema)
	})

	t.Run("parses simple command with explicit args", func(t *testing.T) {
		bp := FromArgs([]string{"git", "status", "[args...]"})

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"args": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Additional command line arguments",
				},
			},
			"required": []string{"args"},
		}, bp.InputSchema)
	})
}

func TestBlueprint_ParseBlueprintedCommand(t *testing.T) {
	t.Run("parses blueprinted command with description", func(t *testing.T) {
		bp := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page#A valid wikipedia page}}"})

		assert.Equal(t, "curl", bp.BaseCommand)
		assert.Equal(t, "curl", bp.ToolName)
		assert.Equal(t, "Run the shell command `curl https://en.m.wikipedia.org/wiki/{{page}}`", bp.ToolDescription)
		assert.Equal(t, map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type":        "string",
					"description": "A valid wikipedia page",
				},
			},
			"required": []string{"page"},
		}, bp.InputSchema)
	})

	t.Run("parses blueprinted command without description", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{text}}"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"text"},
		}, bp.InputSchema)
	})

	t.Run("parses blueprinted command with spaces in description", func(t *testing.T) {
		bp := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page # A valid wikipedia page}}"})

		assert.Equal(t, map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type":        "string",
					"description": "A valid wikipedia page",
				},
			},
			"required": []string{"page"},
		}, bp.InputSchema)
	})
}

func TestBlueprint_BuildCommandArgs(t *testing.T) {
	t.Run("builds simple command without templates", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "hello", "world"})
		args := bp.BuildCommandArgs(map[string]interface{}{})

		assert.Equal(t, []string{"echo", "hello", "world"}, args)
	})

	t.Run("builds command with required template", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{message}}"})
		args := bp.BuildCommandArgs(map[string]interface{}{
			"message": "Hello World",
		})

		assert.Equal(t, []string{"echo", "Hello World"}, args)
	})

	t.Run("builds command with template in middle of arg", func(t *testing.T) {
		bp := FromArgs([]string{"curl", "https://api.example.com/{{endpoint}}"})
		args := bp.BuildCommandArgs(map[string]interface{}{
			"endpoint": "users/123",
		})

		assert.Equal(t, []string{"curl", "https://api.example.com/users/123"}, args)
	})

	t.Run("builds command with multiple templates in one arg", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{greeting}} {{name}}!"})
		args := bp.BuildCommandArgs(map[string]interface{}{
			"greeting": "Hello",
			"name":     "World",
		})

		assert.Equal(t, []string{"echo", "Hello World!"}, args)
	})

	t.Run("builds command with array argument", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[files...]"})
		args := bp.BuildCommandArgs(map[string]interface{}{
			"files": []string{"file1.txt", "file2.txt", "file3.txt"},
		})

		assert.Equal(t, []string{"echo", "file1.txt", "file2.txt", "file3.txt"}, args)
	})

	t.Run("builds command with empty array argument", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "prefix", "[files...]"})
		args := bp.BuildCommandArgs(map[string]interface{}{
			"files": []string{},
		})

		assert.Equal(t, []string{"echo", "prefix"}, args)
	})

	t.Run("builds command with optional string argument provided", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "hello", "[name]"})
		args := bp.BuildCommandArgs(map[string]interface{}{
			"name": "world",
		})

		assert.Equal(t, []string{"echo", "hello", "world"}, args)
	})

	t.Run("builds command with optional string argument omitted", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "hello", "[name]"})
		args := bp.BuildCommandArgs(map[string]interface{}{})

		assert.Equal(t, []string{"echo", "hello"}, args)
	})
}
