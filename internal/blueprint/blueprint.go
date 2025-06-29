package blueprint

import (
	"regexp"
	"strings"
)

var (
	// Matches {{variable}} or {{variable#description}}
	templateRegex = regexp.MustCompile(`\{\{([^#}]+)(?:#([^}]+))?\}\}`)
	// Matches [variable] or [variable...]
	optionalRegex = regexp.MustCompile(`^\[([^.\]]+)(\.\.\.)?]$`)
)

// Blueprint represents a parsed command template
type Blueprint struct {
	BaseCommand     string
	ToolName        string
	ToolDescription string
	InputSchema     map[string]interface{}
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
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}

	if len(args) == 0 {
		return bp
	}

	bp.BaseCommand = args[0]
	bp.ToolName = strings.ReplaceAll(args[0], "-", "_")

	// Parse arguments for templates
	descriptionParts := []string{bp.BaseCommand}
	properties := make(map[string]interface{})
	required := []string{}

	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check for optional pattern [variable] or [variable...]
		if matches := optionalRegex.FindStringSubmatch(arg); matches != nil {
			varName := strings.ReplaceAll(matches[1], "-", "_")
			isArray := matches[2] == "..."

			tmpl := template{
				argIndex:   i,
				name:       varName,
				isArray:    isArray,
				isOptional: true,
			}

			if isArray {
				tmpl.description = "Additional command line arguments"
				properties[varName] = map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": tmpl.description,
				}
				required = append(required, varName)
				descriptionParts = append(descriptionParts, "["+varName+"...]")
			} else {
				properties[varName] = map[string]interface{}{
					"type": "string",
				}
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
					prop := map[string]interface{}{
						"type": "string",
					}
					if description != "" {
						prop["description"] = description
					}
					properties[varName] = prop
				} else if exists && description != "" {
					// Update description if provided
					if propMap, ok := existingProp.(map[string]interface{}); ok {
						propMap["description"] = description
					}
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
		bp.InputSchema["properties"] = properties
	}
	if len(required) > 0 {
		bp.InputSchema["required"] = required
	}

	return bp
}

// BuildCommandArgs builds the actual command arguments from the template
func (bp *Blueprint) BuildCommandArgs(params map[string]interface{}) []string {
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
			isArray := matches[2] == "..."

			if isArray {
				// Handle array expansion
				if values, ok := params[varName]; ok {
					if arr, ok := values.([]string); ok && len(arr) > 0 {
						result = append(result, arr...)
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

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
