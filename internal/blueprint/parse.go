package blueprint

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

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
	}

	return bp, nil
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
