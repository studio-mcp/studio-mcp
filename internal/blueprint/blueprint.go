package blueprint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

var (
	// Matches {{variable}} or {{variable#description}}
	templateRegex = regexp.MustCompile(`\{\{([^#}]+)(?:#([^}]+))?\}\}`)
	// Matches [variable] or [variable#description] or [variable...] or [variable...#description]
	optionalRegex = regexp.MustCompile(`^\[([^#.\]]+)(?:\.\.\.)?(?:#([^.\]]+))?(\.\.\.)?\]$`)
)

// Blueprint represents a parsed command template
type Blueprint struct {
	BaseCommand     string
	ToolName        string
	ToolDescription string
	InputSchema     *jsonschema.Schema
	args            []string
	templates       []template
}

type template struct {
	argIndex    int
	name        string
	description string
	isArray     bool
	isOptional  bool
}

// FromArgs creates a new Blueprint from command arguments
func FromArgs(args []string) *Blueprint {
	bp := &Blueprint{
		args:      args,
		templates: []template{},
		InputSchema: &jsonschema.Schema{
			Type:       "object",
			Properties: make(map[string]*jsonschema.Schema),
		},
	}

	if len(args) == 0 {
		return bp
	}

	bp.BaseCommand = args[0]
	bp.ToolName = strings.ReplaceAll(args[0], "-", "_")

	// Parse arguments for templates
	descriptionParts := []string{bp.BaseCommand}
	properties := make(map[string]*jsonschema.Schema)
	required := []string{}

	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check for optional pattern [variable] or [variable...] or [variable#description] or [variable...#description]
		if matches := optionalRegex.FindStringSubmatch(arg); matches != nil {
			varName := strings.ReplaceAll(matches[1], "-", "_")
			description := ""
			if len(matches) > 2 && matches[2] != "" {
				description = strings.TrimSpace(matches[2])
			}
			isArray := strings.Contains(arg, "...")

			tmpl := template{
				argIndex:    i,
				name:        varName,
				isArray:     isArray,
				isOptional:  true,
				description: description,
			}

			if isArray {
				if description == "" {
					description = "Additional command line arguments"
				}
				properties[varName] = &jsonschema.Schema{
					Type:        "array",
					Items:       &jsonschema.Schema{Type: "string"},
					Description: description,
				}
				required = append(required, varName)
				descriptionParts = append(descriptionParts, "["+varName+"...]")
			} else {
				prop := &jsonschema.Schema{
					Type: "string",
				}
				if description != "" {
					prop.Description = description
				}
				properties[varName] = prop
				descriptionParts = append(descriptionParts, "["+varName+"]")
			}

			bp.templates = append(bp.templates, tmpl)
			continue
		}

		// Check for template patterns in the argument
		processedArg := arg

		// Find all template matches in this argument
		matches := templateRegex.FindAllStringSubmatch(arg, -1)
		if len(matches) > 0 {
			for _, match := range matches {
				varName := strings.TrimSpace(match[1])
				varName = strings.ReplaceAll(varName, "-", "_")
				description := ""
				if len(match) > 2 && match[2] != "" {
					description = strings.TrimSpace(match[2])
				}

				// Only set description if this is the first occurrence or has a description
				if existingProp, exists := properties[varName]; !exists || description != "" {
					prop := &jsonschema.Schema{
						Type: "string",
					}
					if description != "" {
						prop.Description = description
					}
					properties[varName] = prop
				} else if exists && description != "" {
					// Update description if provided
					existingProp.Description = description
				}

				if !contains(required, varName) {
					required = append(required, varName)
				}

				tmpl := template{
					argIndex:    i,
					name:        varName,
					description: description,
					isOptional:  false,
				}
				bp.templates = append(bp.templates, tmpl)
			}

			// Replace template syntax in description
			processedArg = templateRegex.ReplaceAllString(arg, "{{$1}}")
		}

		descriptionParts = append(descriptionParts, processedArg)
	}

	// Build tool description
	bp.ToolDescription = "Run the shell command `" + strings.Join(descriptionParts, " ") + "`"

	// Update InputSchema
	if len(properties) > 0 {
		bp.InputSchema.Properties = properties
	}
	if len(required) > 0 {
		bp.InputSchema.Required = required
	}

	return bp
}

// BuildCommandArgs builds the actual command arguments from the template
func (bp *Blueprint) BuildCommandArgs(params map[string]interface{}) ([]string, error) {
	// Validate required parameters
	for _, required := range bp.InputSchema.Required {
		if _, exists := params[required]; !exists {
			return nil, fmt.Errorf("missing required parameter: %s", required)
		}
	}

	// Validate parameter types
	for name, param := range params {
		if schema, exists := bp.InputSchema.Properties[name]; exists {
			if schema.Type == "array" {
				// Check if it's an array type
				switch v := param.(type) {
				case []string:
					// Valid
				case []interface{}:
					// Valid (from JSON)
				default:
					return nil, fmt.Errorf("parameter '%s' must be an array, got %T", name, v)
				}
			}
		}
	}

	result := []string{bp.BaseCommand}

	// Track which args to skip (for array expansions)
	skipArgs := make(map[int]bool)

	for i := 1; i < len(bp.args); i++ {
		if skipArgs[i] {
			continue
		}

		arg := bp.args[i]

		// Check if this is an array placeholder
		if matches := optionalRegex.FindStringSubmatch(arg); matches != nil {
			varName := strings.ReplaceAll(matches[1], "-", "_")
			isArray := strings.Contains(arg, "...")

			if isArray {
				// Handle array expansion
				if values, ok := params[varName]; ok {
					// Handle both []string and []interface{} (from JSON)
					if arr, ok := values.([]string); ok && len(arr) > 0 {
						result = append(result, arr...)
					} else if arr, ok := values.([]interface{}); ok && len(arr) > 0 {
						for _, item := range arr {
							if str, ok := item.(string); ok {
								result = append(result, str)
							}
						}
					}
				}
			} else {
				// Handle optional string
				if value, ok := params[varName]; ok {
					if str, ok := value.(string); ok && str != "" {
						result = append(result, str)
					}
				}
			}
			continue
		}

		// Process template replacements in the argument
		processedArg := arg
		matches := templateRegex.FindAllStringSubmatch(arg, -1)
		if len(matches) > 0 {
			for _, match := range matches {
				varName := strings.TrimSpace(match[1])
				varName = strings.ReplaceAll(varName, "-", "_")

				if value, ok := params[varName]; ok {
					if str, ok := value.(string); ok {
						// Replace the full template pattern with the value
						fullPattern := match[0]
						processedArg = strings.ReplaceAll(processedArg, fullPattern, str)
					}
				}
			}
		}

		result = append(result, processedArg)
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
