package blueprint

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// parseResult holds the result of parsing a single argument
type parseResult struct {
	fields           []field
	properties       map[string]*jsonschema.Schema
	required         []string
	descriptionParts []string
}

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
		result := parseArgument(args[i], i)

		// Merge results
		bp.fields = append(bp.fields, result.fields...)
		for k, v := range result.properties {
			// If property already exists, only overwrite if new one has description and old one doesn't
			if existingProp, exists := properties[k]; exists {
				if v.Description != "" && existingProp.Description == "" {
					properties[k] = v
				}
			} else {
				properties[k] = v
			}
		}
		for _, req := range result.required {
			if !contains(required, req) {
				required = append(required, req)
			}
		}
		descriptionParts = append(descriptionParts, result.descriptionParts...)
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

// parseArgument parses a single argument and returns the parsing result
func parseArgument(arg string, argIndex int) parseResult {
	result := parseResult{
		fields:           []field{},
		properties:       make(map[string]*jsonschema.Schema),
		required:         []string{},
		descriptionParts: []string{},
	}

	// Try parsing as boolean flag
	if parseBooleanFlag(arg, argIndex, &result) {
		return result
	}

	// Try parsing as optional pattern
	if parseOptionalPattern(arg, argIndex, &result) {
		return result
	}

	// Parse as template pattern or literal
	parseTemplatePattern(arg, argIndex, &result)
	return result
}

// parseBooleanFlag attempts to parse a boolean flag pattern
func parseBooleanFlag(arg string, argIndex int, result *parseResult) bool {
	matches := booleanFlagRegex.FindStringSubmatch(arg)
	if matches == nil {
		return false
	}

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
		argIndex:    argIndex,
		name:        propName,
		isArray:     false,
		isOptional:  true,
		description: description,
		formatter:   getFormatter(false, true, flag),
	}

	result.properties[propName] = &jsonschema.Schema{
		Type:        "boolean",
		Description: description,
	}

	result.descriptionParts = append(result.descriptionParts, "["+flag+"]")
	result.fields = append(result.fields, fld)
	return true
}

// parseOptionalPattern attempts to parse an optional pattern
func parseOptionalPattern(arg string, argIndex int, result *parseResult) bool {
	matches := optionalRegex.FindStringSubmatch(arg)
	if matches == nil {
		return false
	}

	varName := strings.ReplaceAll(matches[1], "-", "_")
	description := ""
	if len(matches) > 2 && matches[2] != "" {
		description = strings.TrimSpace(matches[2])
	}
	isArray := strings.Contains(arg, "...")

	fld := field{
		argIndex:    argIndex,
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
		result.properties[varName] = &jsonschema.Schema{
			Type:        "array",
			Items:       &jsonschema.Schema{Type: "string"},
			Description: description,
		}
		result.required = append(result.required, varName)
		result.descriptionParts = append(result.descriptionParts, "["+varName+"...]")
	} else {
		prop := &jsonschema.Schema{
			Type: "string",
		}
		if description != "" {
			prop.Description = description
		}
		result.properties[varName] = prop
		result.descriptionParts = append(result.descriptionParts, "["+varName+"]")
	}

	result.fields = append(result.fields, fld)
	return true
}

// parseTemplatePattern parses template patterns or treats as literal
func parseTemplatePattern(arg string, argIndex int, result *parseResult) {
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
			if existingProp, exists := result.properties[varName]; !exists || description != "" {
				prop := &jsonschema.Schema{
					Type: "string",
				}
				if description != "" {
					prop.Description = description
				}
				result.properties[varName] = prop
			} else if exists && description != "" {
				// Update description if provided
				existingProp.Description = description
			}

			if !contains(result.required, varName) {
				result.required = append(result.required, varName)
			}

			fld := field{
				argIndex:    argIndex,
				name:        varName,
				description: description,
				isOptional:  false,
				formatter:   getFormatter(false, false),
			}
			result.fields = append(result.fields, fld)
		}

		// Replace template syntax in description
		processedArg = templateRegex.ReplaceAllString(arg, "{{$1}}")
	}

	result.descriptionParts = append(result.descriptionParts, processedArg)
}
