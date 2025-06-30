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
	// Use tokenization internally
	tbp, err := TokenizeFromArgs(args)
	if err != nil {
		return nil, err
	}

	// Create Blueprint with tokenized data
	bp := &Blueprint{
		BaseCommand:     tbp.BaseCommand,
		ToolName:        tbp.ToolName,
		ToolDescription: tbp.ToolDescription,
		InputSchema:     tbp.InputSchema,
		ShellWords:      tbp.ShellWords,
		args:            args,      // Keep for backward compatibility
		fields:          []field{}, // Keep empty for backward compatibility
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

// TokenizeFromArgs creates a new TokenizedBlueprint from command arguments
func TokenizeFromArgs(args []string) (*TokenizedBlueprint, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("cannot create tokenized blueprint: no command provided")
	}

	if strings.TrimSpace(args[0]) == "" {
		return nil, fmt.Errorf("cannot create tokenized blueprint: empty command provided")
	}

	tbp := &TokenizedBlueprint{
		BaseCommand: args[0],
		ToolName:    strings.ReplaceAll(args[0], "-", "_"),
		ShellWords:  make([][]Token, len(args)),
		InputSchema: &jsonschema.Schema{
			Type:       "object",
			Properties: make(map[string]*jsonschema.Schema),
		},
	}

	// Tokenize each shell word
	properties := make(map[string]*jsonschema.Schema)
	required := []string{}
	descriptionParts := []string{}

	for i, arg := range args {
		tokens, argProperties, argRequired, argDescription := tokenizeShellWord(arg)
		tbp.ShellWords[i] = tokens

		// Merge properties
		for k, v := range argProperties {
			if existingProp, exists := properties[k]; exists {
				if v.Description != "" && existingProp.Description == "" {
					properties[k] = v
				}
			} else {
				properties[k] = v
			}
		}

		// Merge required fields
		for _, req := range argRequired {
			if !contains(required, req) {
				required = append(required, req)
			}
		}

		descriptionParts = append(descriptionParts, argDescription)
	}

	// Build tool description
	tbp.ToolDescription = "Run the shell command `" + strings.Join(descriptionParts, " ") + "`"

	// Update InputSchema
	if len(properties) > 0 {
		tbp.InputSchema.Properties = properties
	}
	if len(required) > 0 {
		tbp.InputSchema.Required = required
	}

	return tbp, nil
}

// tokenizeShellWord tokenizes a single shell word into tokens
func tokenizeShellWord(word string) ([]Token, map[string]*jsonschema.Schema, []string, string) {
	tokens := []Token{}
	properties := make(map[string]*jsonschema.Schema)
	required := []string{}
	descriptionPart := word

	// Check for boolean flag patterns first
	if matches := booleanFlagRegex.FindStringSubmatch(word); matches != nil {
		flag := matches[1]
		description := ""
		if len(matches) > 2 && matches[2] != "" {
			description = strings.TrimSpace(matches[2])
		}

		// Keep original flag name for the token
		flagName := strings.TrimLeft(flag, "-")

		// Use normalized name for schema properties (dashes to underscores)
		propName := strings.ReplaceAll(flagName, "-", "_")

		if description == "" {
			description = fmt.Sprintf("Enable %s flag", flag)
		}

		token := FieldToken{
			Name:         flagName, // Store original name
			Description:  description,
			Required:     false,
			OriginalFlag: flag, // Store original flag format
		}
		tokens = append(tokens, token)

		properties[propName] = &jsonschema.Schema{
			Type:        "boolean",
			Description: description,
		}

		descriptionPart = "[" + flag + "]"
		return tokens, properties, required, descriptionPart
	}

	// Check for optional patterns (they match the entire word)
	if matches := optionalRegex.FindStringSubmatch(word); matches != nil {
		varName := matches[1] // Keep original name without dash conversion
		description := ""
		if len(matches) > 2 && matches[2] != "" {
			description = strings.TrimSpace(matches[2])
		}

		isArray := strings.Contains(word, "...")

		token := FieldToken{
			Name:        varName,
			Description: description,
			Required:    false,
		}
		tokens = append(tokens, token)

		// Use normalized name for schema properties (dashes to underscores)
		normalizedName := strings.ReplaceAll(varName, "-", "_")

		if isArray {
			if description == "" {
				description = "Additional command line arguments"
			}
			properties[normalizedName] = &jsonschema.Schema{
				Type:        "array",
				Items:       &jsonschema.Schema{Type: "string"},
				Description: description,
			}
			required = append(required, normalizedName)
			descriptionPart = "[" + varName + "...]"
		} else {
			prop := &jsonschema.Schema{Type: "string"}
			if description != "" {
				prop.Description = description
			}
			properties[normalizedName] = prop
			descriptionPart = "[" + normalizedName + "]"
		}

		return tokens, properties, required, descriptionPart
	}

	// Parse template patterns within the word
	currentPos := 0
	matches := templateRegex.FindAllStringSubmatchIndex(word, -1)

	if len(matches) > 0 {
		for _, match := range matches {
			start, end := match[0], match[1]
			nameStart, nameEnd := match[2], match[3]
			var descStart, descEnd int
			if len(match) > 4 {
				descStart, descEnd = match[4], match[5]
			}

			// Add text before the template
			if start > currentPos {
				textBefore := word[currentPos:start]
				if textBefore != "" {
					tokens = append(tokens, TextToken{Value: textBefore})
				}
			}

			// Extract variable name and description
			varName := strings.TrimSpace(word[nameStart:nameEnd])
			description := ""
			if descStart > 0 && descEnd > descStart {
				description = strings.TrimSpace(word[descStart:descEnd])
			}

			// Add field token with original name
			token := FieldToken{
				Name:        varName, // Keep original name
				Description: description,
				Required:    true,
			}
			tokens = append(tokens, token)

			// Use normalized name for schema properties (dashes to underscores)
			normalizedName := strings.ReplaceAll(varName, "-", "_")

			// Update properties
			prop := &jsonschema.Schema{Type: "string"}
			if description != "" {
				prop.Description = description
			}
			properties[normalizedName] = prop
			required = append(required, normalizedName)

			currentPos = end
		}

		// Add remaining text after the last template
		if currentPos < len(word) {
			textAfter := word[currentPos:]
			if textAfter != "" {
				tokens = append(tokens, TextToken{Value: textAfter})
			}
		}

		// Update description part to show template syntax
		descriptionPart = templateRegex.ReplaceAllString(word, "{{$1}}")
	} else {
		// If no templates were found, treat the entire word as text
		tokens = append(tokens, TextToken{Value: word})
	}

	return tokens, properties, required, descriptionPart
}
