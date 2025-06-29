package blueprint

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestBlueprint_ParseSimpleCommand(t *testing.T) {
	t.Run("parses simple command without args", func(t *testing.T) {
		bp := FromArgs([]string{"git", "status"})

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, "git", bp.ToolName)
		assert.Equal(t, "Run the shell command `git status`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type:       "object",
			Properties: map[string]*jsonschema.Schema{},
		}, bp.InputSchema)
	})

	t.Run("parses simple command with explicit args", func(t *testing.T) {
		bp := FromArgs([]string{"git", "status", "[args...]"})

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"args": {
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: "Additional command line arguments",
				},
			},
			Required: []string{"args"},
		}, bp.InputSchema)
	})
}

func TestBlueprint_ParseBlueprintedCommand(t *testing.T) {
	t.Run("parses blueprinted command with description", func(t *testing.T) {
		bp := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page#A valid wikipedia page}}"})

		assert.Equal(t, "curl", bp.BaseCommand)
		assert.Equal(t, "curl", bp.ToolName)
		assert.Equal(t, "Run the shell command `curl https://en.m.wikipedia.org/wiki/{{page}}`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"page": {
					Type:        "string",
					Description: "A valid wikipedia page",
				},
			},
			Required: []string{"page"},
		}, bp.InputSchema)
	})

	t.Run("parses blueprinted command without description", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{text}}"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"text": {
					Type: "string",
				},
			},
			Required: []string{"text"},
		}, bp.InputSchema)
	})

	t.Run("parses blueprinted command with spaces in description", func(t *testing.T) {
		bp := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page # A valid wikipedia page}}"})

		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"page": {
					Type:        "string",
					Description: "A valid wikipedia page",
				},
			},
			Required: []string{"page"},
		}, bp.InputSchema)
	})

	t.Run("parses mixed blueprints with required and optional arguments", func(t *testing.T) {
		bp := FromArgs([]string{"command", "{{arg1#Custom description}}", "[arg2]"})

		assert.Equal(t, "command", bp.BaseCommand)
		assert.Equal(t, "command", bp.ToolName)
		assert.Equal(t, "Run the shell command `command {{arg1}} [arg2]`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"arg1": {
					Type:        "string",
					Description: "Custom description",
				},
				"arg2": {
					Type: "string",
				},
			},
			Required: []string{"arg1"},
		}, bp.InputSchema)
	})

	t.Run("prioritizes explicit description over default", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{text#Explicit description}}", "{{text}}"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo", bp.ToolName)
		assert.Equal(t, "Run the shell command `echo {{text}} {{text}}`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"text": {
					Type:        "string",
					Description: "Explicit description",
				},
			},
			Required: []string{"text"},
		}, bp.InputSchema)
	})

	t.Run("parses array arguments with description", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[files...]"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo", bp.ToolName)
		assert.Equal(t, "Run the shell command `echo [files...]`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"files": {
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: "Additional command line arguments",
				},
			},
			Required: []string{"files"},
		}, bp.InputSchema)
	})

	t.Run("parses array arguments without description", func(t *testing.T) {
		bp := FromArgs([]string{"ls", "[paths...]"})

		assert.Equal(t, "ls", bp.BaseCommand)
		assert.Equal(t, "ls", bp.ToolName)
		assert.Equal(t, "Run the shell command `ls [paths...]`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"paths": {
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: "Additional command line arguments",
				},
			},
			Required: []string{"paths"},
		}, bp.InputSchema)
	})

	t.Run("parses optional string field without ellipsis", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[optional]"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo", bp.ToolName)
		assert.Equal(t, "Run the shell command `echo [optional]`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"optional": {
					Type: "string",
				},
			},
		}, bp.InputSchema)
	})

	t.Run("converts dashes to underscores in argument names", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[has-dashes]"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo", bp.ToolName)
		assert.Equal(t, "Run the shell command `echo [has_dashes]`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"has_dashes": {
					Type: "string",
				},
			},
		}, bp.InputSchema)
	})

	t.Run("parses mixed string and array arguments", func(t *testing.T) {
		bp := FromArgs([]string{"command", "{{flag#Command flag}}", "[files...]"})

		assert.Equal(t, "command", bp.BaseCommand)
		assert.Equal(t, "command", bp.ToolName)
		assert.Equal(t, "Run the shell command `command {{flag}} [files...]`", bp.ToolDescription)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"flag": {
					Type:        "string",
					Description: "Command flag",
				},
				"files": {
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: "Additional command line arguments",
				},
			},
			Required: []string{"flag", "files"},
		}, bp.InputSchema)
	})
}

func TestBlueprint_ToolName(t *testing.T) {
	t.Run("converts command to valid tool name", func(t *testing.T) {
		bp := FromArgs([]string{"git-flow"})
		assert.Equal(t, "git_flow", bp.ToolName)
	})
}

func TestBlueprint_ToolDescription(t *testing.T) {
	t.Run("generates description for simple command without args", func(t *testing.T) {
		bp := FromArgs([]string{"git"})
		assert.Equal(t, "Run the shell command `git`", bp.ToolDescription)
	})

	t.Run("generates description for simple command with explicit args", func(t *testing.T) {
		bp := FromArgs([]string{"git", "[args...]"})
		assert.Equal(t, "Run the shell command `git [args...]`", bp.ToolDescription)
	})

	t.Run("generates description for blueprinted command", func(t *testing.T) {
		bp := FromArgs([]string{"rails", "generate", "{{generator#A rails generator}}"})
		assert.Equal(t, "Run the shell command `rails generate {{generator}}`", bp.ToolDescription)
	})
}

func TestBlueprint_BuildCommandArgs(t *testing.T) {
	t.Run("builds simple command without templates", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "hello", "world"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "hello", "world"}, args)
	})

	t.Run("builds command with required template", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{message}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"message": "Hello World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Hello World"}, args)
	})

	t.Run("builds command with template in middle of arg", func(t *testing.T) {
		bp := FromArgs([]string{"curl", "https://api.example.com/{{endpoint}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"endpoint": "users/123",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"curl", "https://api.example.com/users/123"}, args)
	})

	t.Run("builds command with multiple templates in one arg", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{greeting}} {{name}}!"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"greeting": "Hello",
			"name":     "World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Hello World!"}, args)
	})

	t.Run("builds command with array argument", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[files...]"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"files": []string{"file1.txt", "file2.txt", "file3.txt"},
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "file1.txt", "file2.txt", "file3.txt"}, args)
	})

	t.Run("builds command with empty array argument", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "prefix", "[files...]"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"files": []string{},
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "prefix"}, args)
	})

	t.Run("builds command with optional string argument provided", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "hello", "[name]"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"name": "world",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "hello", "world"}, args)
	})

	t.Run("builds command with optional string argument omitted", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "hello", "[name]"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "hello"}, args)
	})

	t.Run("builds command with blueprint arguments containing spaces", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{text#text to echo}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"text": "Hello World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Hello World"}, args)
	})

	t.Run("builds command with mixed blueprint with and without descriptions", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{greeting#The greeting}}", "{{name}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"greeting": "Hello",
			"name":     "World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Hello", "World"}, args)
	})

	t.Run("builds command with blueprint arguments in mixed content", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "simon says {{text#text for simon to say}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"text": "Hello World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "simon says Hello World"}, args)
	})

	t.Run("builds command with blueprint arguments containing special shell characters", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{text#text to echo}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"text": "Hello & World; echo pwned",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Hello & World; echo pwned"}, args)
	})

	t.Run("builds command with blueprint in middle of argument", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "--message={{text#message content}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"text": "Hello World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "--message=Hello World"}, args)
	})

	t.Run("builds command with blueprint with prefix and suffix", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "prefix-{{text#middle part}}-suffix"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"text": "Hello World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "prefix-Hello World-suffix"}, args)
	})

	t.Run("builds command with mixed blueprint and non-blueprint arguments", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "static", "{{dynamic#dynamic content}}", "more-static"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"dynamic": "Hello World",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "static", "Hello World", "more-static"}, args)
	})

	t.Run("preserves shell safety with complex blueprint values", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "Result: {{text#text content}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"text": "$(echo 'dangerous'); echo 'safe'",
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Result: $(echo 'dangerous'); echo 'safe'"}, args)
	})

	t.Run("builds command with mixed string and array arguments", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{prefix#Prefix text}}", "[files...]"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"prefix": "Files:",
			"files":  []string{"a.txt", "b.txt"},
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "Files:", "a.txt", "b.txt"}, args)
	})

	t.Run("builds command with mixed required and optional arguments", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{required#Required text}}", "[optional]"})

		// With both provided
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"required": "hello",
			"optional": "world",
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "hello", "world"}, args)

		// With only required provided
		args, err = bp.BuildCommandArgs(map[string]interface{}{
			"required": "hello",
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "hello"}, args)
	})
}

func TestBlueprint_EnhancedOptionalParsing(t *testing.T) {
	t.Run("parses optional argument with custom description", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[name#Person's name]"})

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"name": {
					Type:        "string",
					Description: "Person's name",
				},
			},
		}, bp.InputSchema)
	})

	t.Run("parses array argument with custom description", func(t *testing.T) {
		bp := FromArgs([]string{"ls", "[files...#Files to list]"})

		assert.Equal(t, "ls", bp.BaseCommand)
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"files": {
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: "Files to list",
				},
			},
			Required: []string{"files"},
		}, bp.InputSchema)
	})

	t.Run("parses mixed optional arguments with and without descriptions", func(t *testing.T) {
		bp := FromArgs([]string{"cmd", "[required]", "[optional#Custom desc]"})

		assert.Equal(t, "cmd", bp.BaseCommand)

		// Check properties
		assert.Equal(t, "string", bp.InputSchema.Properties["required"].Type)
		assert.Equal(t, "", bp.InputSchema.Properties["required"].Description)

		assert.Equal(t, "string", bp.InputSchema.Properties["optional"].Type)
		assert.Equal(t, "Custom desc", bp.InputSchema.Properties["optional"].Description)
	})
}

func TestBlueprint_TemplateValidation(t *testing.T) {
	t.Run("validates missing required parameters", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{required}}"})
		_, err := bp.BuildCommandArgs(map[string]interface{}{})
		assert.Error(t, err) // Should error on missing required param
	})

	t.Run("validates parameter type mismatches", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "[files...]"})
		_, err := bp.BuildCommandArgs(map[string]interface{}{
			"files": "not-an-array", // Should be []string
		})
		assert.Error(t, err)
	})
}

func TestBlueprint_EnhancedTemplateProcessing(t *testing.T) {
	t.Run("handles malformed template syntax", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{incomplete"})
		// Should ignore any bad template syntax and print as normal text (always fallback on parse error to literal text)
		args, err := bp.BuildCommandArgs(map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "{{incomplete"}, args)
	})

	t.Run("handles malformed template with only opening braces", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{no_closing_braces"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "{{no_closing_braces"}, args)
	})

	t.Run("handles malformed template with only closing braces", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "no_opening_braces}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "no_opening_braces}}"}, args)
	})

	t.Run("handles mixed valid and malformed templates", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{valid}}", "{{incomplete", "}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{
			"valid": "works",
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "works", "{{incomplete", "}}"}, args)
	})

	t.Run("handles empty template braces", func(t *testing.T) {
		bp := FromArgs([]string{"echo", "{{}}"})
		args, err := bp.BuildCommandArgs(map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, []string{"echo", "{{}}"}, args)
	})
}
