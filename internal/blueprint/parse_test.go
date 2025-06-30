package blueprint

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlueprint_ParseSimpleCommand(t *testing.T) {
	t.Run("parses simple command without args", func(t *testing.T) {
		bp, err := FromArgs([]string{"git", "status"})
		require.NoError(t, err)

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, "git status", bp.GetCommandFormat())
		assert.Equal(t, &jsonschema.Schema{
			Type:       "object",
			Properties: map[string]*jsonschema.Schema{},
		}, bp.InputSchema)
	})

	t.Run("parses simple command with explicit args", func(t *testing.T) {
		bp, err := FromArgs([]string{"git", "status", "[args...]"})
		require.NoError(t, err)

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
		bp, err := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page#A valid wikipedia page}}"})
		require.NoError(t, err)

		assert.Equal(t, "curl", bp.BaseCommand)
		assert.Equal(t, "curl https://en.m.wikipedia.org/wiki/{{page}}", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"echo", "{{text}}"})
		require.NoError(t, err)

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
		bp, err := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page # A valid wikipedia page}}"})
		require.NoError(t, err)

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
		bp, err := FromArgs([]string{"command", "{{arg1#Custom description}}", "[arg2]"})
		require.NoError(t, err)

		assert.Equal(t, "command", bp.BaseCommand)
		assert.Equal(t, "command {{arg1}} [arg2]", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"echo", "{{text#Explicit description}}", "{{text}}"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo {{text}} {{text}}", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"echo", "[files...]"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo [files...]", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"ls", "[paths...]"})
		require.NoError(t, err)

		assert.Equal(t, "ls", bp.BaseCommand)
		assert.Equal(t, "ls [paths...]", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"echo", "[optional]"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo [optional]", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"echo", "[has-dashes]"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo [has_dashes]", bp.GetCommandFormat())
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
		bp, err := FromArgs([]string{"command", "{{flag#Command flag}}", "[files...]"})
		require.NoError(t, err)

		assert.Equal(t, "command", bp.BaseCommand)
		assert.Equal(t, "command {{flag}} [files...]", bp.GetCommandFormat())
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

func TestBlueprint_GetCommandFormat(t *testing.T) {
	t.Run("generates description for simple command without args", func(t *testing.T) {
		bp, err := FromArgs([]string{"git"})
		require.NoError(t, err)
		assert.Equal(t, "git", bp.GetCommandFormat())
	})

	t.Run("generates description for simple command with explicit args", func(t *testing.T) {
		bp, err := FromArgs([]string{"git", "[args...]"})
		require.NoError(t, err)
		assert.Equal(t, "git [args...]", bp.GetCommandFormat())
	})

	t.Run("generates description for blueprinted command", func(t *testing.T) {
		bp, err := FromArgs([]string{"rails", "generate", "{{generator#A rails generator}}"})
		require.NoError(t, err)
		assert.Equal(t, "rails generate {{generator}}", bp.GetCommandFormat())
	})
}

func TestBlueprint_EnhancedOptionalParsing(t *testing.T) {
	t.Run("parses optional argument with custom description", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "[name#Person's name]"})
		require.NoError(t, err)

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
		bp, err := FromArgs([]string{"ls", "[files...#Files to list]"})
		require.NoError(t, err)

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
		bp, err := FromArgs([]string{"cmd", "[required]", "[optional#Custom desc]"})
		require.NoError(t, err)

		assert.Equal(t, "cmd", bp.BaseCommand)

		// Check properties
		assert.Equal(t, "string", bp.InputSchema.Properties["required"].Type)
		assert.Equal(t, "", bp.InputSchema.Properties["required"].Description)

		assert.Equal(t, "string", bp.InputSchema.Properties["optional"].Type)
		assert.Equal(t, "Custom desc", bp.InputSchema.Properties["optional"].Description)
	})
}

func TestBlueprint_ParseBooleanFlags(t *testing.T) {
	t.Run("parses short boolean flag without description", func(t *testing.T) {
		bp, err := FromArgs([]string{"ls", "[-f]"})
		require.NoError(t, err)

		assert.Equal(t, "ls", bp.BaseCommand)
		assert.Equal(t, "ls [-f]", bp.GetCommandFormat())
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"f": {
					Type:        "boolean",
					Description: "Enable -f flag",
				},
			},
		}, bp.InputSchema)
	})

	t.Run("parses long boolean flag without description", func(t *testing.T) {
		bp, err := FromArgs([]string{"ls", "[--force]"})
		require.NoError(t, err)

		assert.Equal(t, "ls", bp.BaseCommand)
		assert.Equal(t, "ls [--force]", bp.GetCommandFormat())
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"force": {
					Type:        "boolean",
					Description: "Enable --force flag",
				},
			},
		}, bp.InputSchema)
	})

	t.Run("parses boolean flag with description", func(t *testing.T) {
		bp, err := FromArgs([]string{"rm", "[-f#force removal]"})
		require.NoError(t, err)

		assert.Equal(t, "rm", bp.BaseCommand)
		assert.Equal(t, "rm [-f]", bp.GetCommandFormat())
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"f": {
					Type:        "boolean",
					Description: "force removal",
				},
			},
		}, bp.InputSchema)
	})

	t.Run("parses mixed boolean flags and other arguments", func(t *testing.T) {
		bp, err := FromArgs([]string{"cp", "[-r#recursive]", "{{source}}", "{{dest}}"})
		require.NoError(t, err)

		assert.Equal(t, "cp", bp.BaseCommand)
		assert.Equal(t, "cp [-r] {{source}} {{dest}}", bp.GetCommandFormat())
		assert.Equal(t, &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"r": {
					Type:        "boolean",
					Description: "recursive",
				},
				"source": {
					Type: "string",
				},
				"dest": {
					Type: "string",
				},
			},
			Required: []string{"source", "dest"},
		}, bp.InputSchema)
	})
}

func TestBlueprint_FromArgs(t *testing.T) {
	t.Run("parses simple command", func(t *testing.T) {
		bp, err := FromArgs([]string{"git", "status"})
		require.NoError(t, err)

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, "git status", bp.GetCommandFormat())
		assert.Empty(t, bp.InputSchema.Properties)
	})

	t.Run("parses command with array arguments", func(t *testing.T) {
		bp, err := FromArgs([]string{"git", "status", "[args...]"})
		require.NoError(t, err)

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, "git status [args...]", bp.GetCommandFormat())

		// Check that args parameter exists and is required
		assert.Contains(t, bp.InputSchema.Properties, "args")
		assert.Equal(t, "array", bp.InputSchema.Properties["args"].Type)
		assert.Contains(t, bp.InputSchema.Required, "args")
	})

	t.Run("parses command with template argument", func(t *testing.T) {
		bp, err := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page#A valid wikipedia page}}"})
		require.NoError(t, err)

		assert.Equal(t, "curl", bp.BaseCommand)
		assert.Equal(t, "curl https://en.m.wikipedia.org/wiki/{{page}}", bp.GetCommandFormat())

		// Check that page parameter exists and is required
		assert.Contains(t, bp.InputSchema.Properties, "page")
		assert.Equal(t, "string", bp.InputSchema.Properties["page"].Type)
		assert.Equal(t, "A valid wikipedia page", bp.InputSchema.Properties["page"].Description)
		assert.Contains(t, bp.InputSchema.Required, "page")
	})

	t.Run("parses command with template argument without description", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "{{text}}"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo {{text}}", bp.GetCommandFormat())

		// Check that text parameter exists and is required
		assert.Contains(t, bp.InputSchema.Properties, "text")
		assert.Equal(t, "string", bp.InputSchema.Properties["text"].Type)
		assert.Contains(t, bp.InputSchema.Required, "text")
	})

	t.Run("parses command with template argument with space before description", func(t *testing.T) {
		bp, err := FromArgs([]string{"curl", "https://en.m.wikipedia.org/wiki/{{page # A valid wikipedia page}}"})
		require.NoError(t, err)

		assert.Equal(t, "curl", bp.BaseCommand)
		assert.Equal(t, "curl https://en.m.wikipedia.org/wiki/{{page }}", bp.GetCommandFormat())

		// Check that page parameter exists and is required
		assert.Contains(t, bp.InputSchema.Properties, "page")
		assert.Equal(t, "string", bp.InputSchema.Properties["page"].Type)
		assert.Equal(t, "A valid wikipedia page", bp.InputSchema.Properties["page"].Description)
		assert.Contains(t, bp.InputSchema.Required, "page")
	})

	t.Run("parses command with mixed template and optional arguments", func(t *testing.T) {
		bp, err := FromArgs([]string{"command", "{{arg1#Custom description}}", "[arg2]"})
		require.NoError(t, err)

		assert.Equal(t, "command", bp.BaseCommand)
		assert.Equal(t, "command {{arg1}} [arg2]", bp.GetCommandFormat())

		// Check that arg1 parameter exists and is required
		assert.Contains(t, bp.InputSchema.Properties, "arg1")
		assert.Equal(t, "string", bp.InputSchema.Properties["arg1"].Type)
		assert.Equal(t, "Custom description", bp.InputSchema.Properties["arg1"].Description)
		assert.Contains(t, bp.InputSchema.Required, "arg1")

		// Check that arg2 parameter exists and is optional
		assert.Contains(t, bp.InputSchema.Properties, "arg2")
		assert.Equal(t, "string", bp.InputSchema.Properties["arg2"].Type)
		assert.NotContains(t, bp.InputSchema.Required, "arg2")
	})

	t.Run("handles duplicate template variables with descriptions", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "{{text#Explicit description}}", "{{text}}"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)

		// Check that text parameter exists with the explicit description
		assert.Contains(t, bp.InputSchema.Properties, "text")
		assert.Equal(t, "string", bp.InputSchema.Properties["text"].Type)
		assert.Equal(t, "Explicit description", bp.InputSchema.Properties["text"].Description)
		assert.Contains(t, bp.InputSchema.Required, "text")
	})

	t.Run("parses command with array arguments and description", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "[files...]"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo [files...]", bp.GetCommandFormat())

		// Check that files parameter exists and is required array
		assert.Contains(t, bp.InputSchema.Properties, "files")
		assert.Equal(t, "array", bp.InputSchema.Properties["files"].Type)
		assert.Equal(t, "Additional command line arguments", bp.InputSchema.Properties["files"].Description)
		assert.Contains(t, bp.InputSchema.Required, "files")
	})

	t.Run("parses command with array arguments and custom description", func(t *testing.T) {
		bp, err := FromArgs([]string{"ls", "[paths...]"})
		require.NoError(t, err)

		assert.Equal(t, "ls", bp.BaseCommand)
		assert.Equal(t, "ls [paths...]", bp.GetCommandFormat())

		// Check that paths parameter exists and is required array
		assert.Contains(t, bp.InputSchema.Properties, "paths")
		assert.Equal(t, "array", bp.InputSchema.Properties["paths"].Type)
		assert.Equal(t, "Additional command line arguments", bp.InputSchema.Properties["paths"].Description)
		assert.Contains(t, bp.InputSchema.Required, "paths")
	})

	t.Run("parses command with optional string argument", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "[optional]"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)
		assert.Equal(t, "echo [optional]", bp.GetCommandFormat())

		// Check that optional parameter exists and is not required
		assert.Contains(t, bp.InputSchema.Properties, "optional")
		assert.Equal(t, "string", bp.InputSchema.Properties["optional"].Type)
		assert.NotContains(t, bp.InputSchema.Required, "optional")
	})

	t.Run("handles dashes in optional argument names", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "[has-dashes]"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)

		// Check that has_dashes parameter exists (dashes converted to underscores)
		assert.Contains(t, bp.InputSchema.Properties, "has_dashes")
		assert.Equal(t, "string", bp.InputSchema.Properties["has_dashes"].Type)
		assert.NotContains(t, bp.InputSchema.Required, "has_dashes")
	})

	t.Run("parses command with mixed required and array arguments", func(t *testing.T) {
		bp, err := FromArgs([]string{"command", "{{flag#Command flag}}", "[files...]"})
		require.NoError(t, err)

		assert.Equal(t, "command", bp.BaseCommand)
		assert.Equal(t, "command {{flag}} [files...]", bp.GetCommandFormat())

		// Check that flag parameter exists and is required
		assert.Contains(t, bp.InputSchema.Properties, "flag")
		assert.Equal(t, "string", bp.InputSchema.Properties["flag"].Type)
		assert.Equal(t, "Command flag", bp.InputSchema.Properties["flag"].Description)
		assert.Contains(t, bp.InputSchema.Required, "flag")

		// Check that files parameter exists and is required array
		assert.Contains(t, bp.InputSchema.Properties, "files")
		assert.Equal(t, "array", bp.InputSchema.Properties["files"].Type)
		assert.Equal(t, "Additional command line arguments", bp.InputSchema.Properties["files"].Description)
		assert.Contains(t, bp.InputSchema.Required, "files")
	})

	t.Run("creates tool name from command with dashes", func(t *testing.T) {
		bp, err := FromArgs([]string{"git-flow"})
		require.NoError(t, err)

		assert.Equal(t, "git-flow", bp.BaseCommand)
	})

	t.Run("handles command with no additional arguments", func(t *testing.T) {
		bp, err := FromArgs([]string{"git"})
		require.NoError(t, err)

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Equal(t, "git", bp.GetCommandFormat())
	})

	t.Run("handles command with optional array arguments", func(t *testing.T) {
		bp, err := FromArgs([]string{"git", "[args...]"})
		require.NoError(t, err)

		assert.Equal(t, "git", bp.BaseCommand)
		assert.Contains(t, bp.InputSchema.Required, "args")
	})

	t.Run("handles command with template generator", func(t *testing.T) {
		bp, err := FromArgs([]string{"rails", "generate", "{{generator#A rails generator}}"})
		require.NoError(t, err)

		assert.Equal(t, "rails", bp.BaseCommand)
		assert.Contains(t, bp.InputSchema.Required, "generator")
	})
}

func TestBlueprint_FromArgsErrors(t *testing.T) {
	t.Run("returns error for empty args", func(t *testing.T) {
		_, err := FromArgs([]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no command provided")
	})

	t.Run("returns error for empty command", func(t *testing.T) {
		_, err := FromArgs([]string{""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty command provided")
	})

	t.Run("returns error for whitespace-only command", func(t *testing.T) {
		_, err := FromArgs([]string{"   "})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty command provided")
	})
}

func TestBlueprint_FromArgsTokenization(t *testing.T) {
	t.Run("tokenizes simple command without templates", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "hello"})
		require.NoError(t, err)

		assert.Equal(t, "echo", bp.BaseCommand)

		expected := [][]Token{
			{TextToken{Value: "echo"}},
			{TextToken{Value: "hello"}},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})

	t.Run("tokenizes command with simple template", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "{{text}}"})
		require.NoError(t, err)

		expected := [][]Token{
			{TextToken{Value: "echo"}},
			{FieldToken{Name: "text", Description: "", Required: true, OriginalFlag: "", OriginalName: "text"}},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})

	t.Run("tokenizes command with template and description", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "{{text#message to echo}}"})
		require.NoError(t, err)

		expected := [][]Token{
			{TextToken{Value: "echo"}},
			{FieldToken{Name: "text", Description: "message to echo", Required: true, OriginalFlag: "", OriginalName: "text"}},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})

	t.Run("tokenizes command with mixed text and template", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "prefix{{text#desc}}suffix"})
		require.NoError(t, err)

		expected := [][]Token{
			{TextToken{Value: "echo"}},
			{
				TextToken{Value: "prefix"},
				FieldToken{Name: "text", Description: "desc", Required: true, OriginalFlag: "", OriginalName: "text"},
				TextToken{Value: "suffix"},
			},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})

	t.Run("tokenizes command with optional field", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "[optional]"})
		require.NoError(t, err)

		expected := [][]Token{
			{TextToken{Value: "echo"}},
			{FieldToken{Name: "optional", Description: "", Required: false, OriginalFlag: ""}},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})

	t.Run("tokenizes command with optional field and description", func(t *testing.T) {
		bp, err := FromArgs([]string{"echo", "[optional#optional text]"})
		require.NoError(t, err)

		expected := [][]Token{
			{TextToken{Value: "echo"}},
			{FieldToken{Name: "optional", Description: "optional text", Required: false, OriginalFlag: ""}},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})

	t.Run("tokenizes complex mixed command", func(t *testing.T) {
		bp, err := FromArgs([]string{"curl", "https://api.com/{{endpoint#API endpoint}}", "[--verbose]"})
		require.NoError(t, err)

		expected := [][]Token{
			{TextToken{Value: "curl"}},
			{
				TextToken{Value: "https://api.com/"},
				FieldToken{Name: "endpoint", Description: "API endpoint", Required: true, OriginalFlag: "", OriginalName: "endpoint"},
			},
			{FieldToken{Name: "verbose", Description: "Enable --verbose flag", Required: false, OriginalFlag: "--verbose", OriginalName: ""}},
		}
		assert.Equal(t, expected, bp.ShellWords)
	})
}
