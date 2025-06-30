package blueprint

import (
	"fmt"
	"strings"
)

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

		// Check if this is a boolean flag placeholder
		if matches := booleanFlagRegex.FindStringSubmatch(arg); matches != nil {
			flag := matches[1]
			propName := strings.TrimLeft(flag, "-")
			propName = strings.ReplaceAll(propName, "-", "_")

			// Find the field for this variable
			var fld *field
			for _, f := range bp.fields {
				if f.name == propName && f.argIndex == i {
					fld = &f
					break
				}
			}

			if fld != nil && fld.formatter != nil {
				if value, ok := params[propName]; ok {
					formatted := fld.formatter(value)
					result = append(result, formatted...)
				}
			}
			continue
		}

		// Check if this is an optional placeholder
		if matches := optionalRegex.FindStringSubmatch(arg); matches != nil {
			varName := strings.ReplaceAll(matches[1], "-", "_")

			// Find the field for this variable
			var fld *field
			for _, f := range bp.fields {
				if f.name == varName && f.argIndex == i {
					fld = &f
					break
				}
			}

			if fld != nil && fld.formatter != nil {
				if value, ok := params[varName]; ok {
					formatted := fld.formatter(value)
					result = append(result, formatted...)
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

				// Find the field for this variable
				var fld *field
				for _, f := range bp.fields {
					if f.name == varName && f.argIndex == i {
						fld = &f
						break
					}
				}

				if fld != nil && fld.formatter != nil {
					if value, ok := params[varName]; ok {
						formatted := fld.formatter(value)
						if len(formatted) > 0 {
							// Replace the full template pattern with the first formatted value
							fullPattern := match[0]
							processedArg = strings.ReplaceAll(processedArg, fullPattern, formatted[0])
						}
					}
				}
			}
		}

		result = append(result, processedArg)
	}

	return result, nil
}
