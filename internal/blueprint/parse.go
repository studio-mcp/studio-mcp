package blueprint

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// FromArgs creates a new Blueprint from command arguments
func FromArgs(args []string) (*Blueprint, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("cannot create blueprint: no command provided")
	}

	if strings.TrimSpace(args[0]) == "" {
		return nil, fmt.Errorf("cannot create blueprint: empty command provided")
	}

	bp := &Blueprint{
		args:   args,
		fields: []field{},
		InputSchema: &jsonschema.Schema{
			Type:       "object",
			Properties: make(map[string]*jsonschema.Schema),
		},
	}

	bp.BaseCommand = args[0]
	bp.ToolName = strings.ReplaceAll(args[0], "-", "_")

	// Parse arguments for templates
	descriptionParts := []string{bp.BaseCommand}
	properties := make(map[string]*jsonschema.Schema)
	required := []string{}

	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check for boolean flag pattern [-f] or [--flag] or [-f#description] or [--flag#description]
		if matches := booleanFlagRegex.FindStringSubmatch(arg); matches != nil {
			flag := matches[1]
			description := ""
			if len(matches) > 2 && matches[2] != "" {
				description = strings.TrimSpace(matches[2])
			}

			// Extract property name from flag (remove dashes)
			propName := strings.TrimLeft(flag, "-")
			propName = strings.ReplaceAll(propName, "-", "_")

			if description == "" {
				description = fmt.Sprintf("Enable %s flag", flag)
			}

			fld := field{
				argIndex:    i,
				name:        propName,
				isArray:     false,
				isOptional:  true,
				description: description,
				formatter:   getFormatter(false, true, flag),
			}

			properties[propName] = &jsonschema.Schema{
				Type:        "boolean",
				Description: description,
			}

			descriptionParts = append(descriptionParts, "["+flag+"]")
			bp.fields = append(bp.fields, fld)
			continue
		}

		// Check for optional pattern [variable] or [variable...] or [variable#description] or [variable...#description]
		if matches := optionalRegex.FindStringSubmatch(arg); matches != nil {
			varName := strings.ReplaceAll(matches[1], "-", "_")
			description := ""
			if len(matches) > 2 && matches[2] != "" {
				description = strings.TrimSpace(matches[2])
			}
			isArray := strings.Contains(arg, "...")

			fld := field{
				argIndex:    i,
				name:        varName,
				isArray:     isArray,
				isOptional:  true,
				description: description,
				formatter:   getFormatter(isArray, false),
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

			bp.fields = append(bp.fields, fld)
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

				fld := field{
					argIndex:    i,
					name:        varName,
					description: description,
					isOptional:  false,
					formatter:   getFormatter(false, false),
				}
				bp.fields = append(bp.fields, fld)
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

	return bp, nil
}
