package blueprint

import (
	"fmt"
	"strings"
)

// FromArgs creates a new Blueprint from command arguments using tokenization
func FromArgs(args []string) (*Blueprint, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("cannot create blueprint: no command provided")
	}

	if strings.TrimSpace(args[0]) == "" {
		return nil, fmt.Errorf("cannot create blueprint: empty command provided")
	}

	bp := &Blueprint{
		BaseCommand: args[0],
		ShellWords:  make([][]Token, len(args)),
	}

	// Tokenize each shell word
	for i, arg := range args {
		tokens := tokenizeShellWord(arg)
		bp.ShellWords[i] = tokens
	}

	return bp, nil
}

// tokenizeShellWord tokenizes a single shell word into tokens
func tokenizeShellWord(word string) []Token {
	tokens := []Token{}

	// Check if entire word is a single field (starts with {{ or [)
	if (strings.HasPrefix(word, "{{") && strings.HasSuffix(word, "}}")) ||
		(strings.HasPrefix(word, "[") && strings.HasSuffix(word, "]")) {

		token := parseField(word)
		if token != nil {
			tokens = append(tokens, token)
			return tokens
		}
		// If parsing failed, treat as literal text
		tokens = append(tokens, TextToken{Value: word})
		return tokens
	}

	// Parse mixed content with templates
	currentPos := 0
	for currentPos < len(word) {
		// Look for next template start (either {{ or [)
		nextTemplateStart := -1
		templateType := ""

		// Find the closest template start
		requiredStart := strings.Index(word[currentPos:], "{{")
		optionalStart := strings.Index(word[currentPos:], "[")

		if requiredStart != -1 && (optionalStart == -1 || requiredStart < optionalStart) {
			nextTemplateStart = requiredStart + currentPos
			templateType = "required"
		} else if optionalStart != -1 {
			nextTemplateStart = optionalStart + currentPos
			templateType = "optional"
		}

		if nextTemplateStart == -1 {
			// No more templates, add remaining text
			if currentPos < len(word) {
				tokens = append(tokens, TextToken{Value: word[currentPos:]})
			}
			break
		}

		// Add text before template
		if nextTemplateStart > currentPos {
			tokens = append(tokens, TextToken{Value: word[currentPos:nextTemplateStart]})
		}

		// Find template end based on type
		var templateEnd int
		var templateText string

		if templateType == "required" {
			endIndex := strings.Index(word[nextTemplateStart:], "}}")
			if endIndex == -1 {
				// Malformed template, treat rest as text
				tokens = append(tokens, TextToken{Value: word[nextTemplateStart:]})
				break
			}
			templateEnd = nextTemplateStart + endIndex + 2
			templateText = word[nextTemplateStart:templateEnd]
		} else { // optional
			endIndex := strings.Index(word[nextTemplateStart:], "]")
			if endIndex == -1 {
				// Malformed template, treat rest as text
				tokens = append(tokens, TextToken{Value: word[nextTemplateStart:]})
				break
			}
			templateEnd = nextTemplateStart + endIndex + 1
			templateText = word[nextTemplateStart:templateEnd]
		}

		// Parse the template
		token := parseField(templateText)
		if token != nil {
			tokens = append(tokens, token)
		} else {
			// If parsing failed, treat as literal text
			tokens = append(tokens, TextToken{Value: templateText})
		}

		currentPos = templateEnd
	}

	// If no tokens were created, treat entire word as text
	if len(tokens) == 0 {
		tokens = append(tokens, TextToken{Value: word})
	}

	return tokens
}

// parseField parses a field enclosed in {{ }} or [ ]
func parseField(field string) Token {
	var content string
	var required bool

	// Determine field type and extract content
	if strings.HasPrefix(field, "{{") && strings.HasSuffix(field, "}}") {
		content = field[2 : len(field)-2] // Remove {{ }}
		required = true
	} else if strings.HasPrefix(field, "[") && strings.HasSuffix(field, "]") {
		content = field[1 : len(field)-1] // Remove [ ]
		required = false
	} else {
		return nil // Not a valid field
	}

	// Parse content for name, description, and modifiers
	var name, description string
	isArray := false
	var originalFlag string

	// Check for description (split on #)
	parts := strings.SplitN(content, "#", 2)
	name = strings.TrimSpace(parts[0])

	// If name is empty, this is not a valid field (e.g., {{}})
	if name == "" {
		return nil
	}

	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	// Check for array notation (...)
	if strings.HasSuffix(name, "...") {
		isArray = true
		name = strings.TrimSuffix(name, "...")
		name = strings.TrimSpace(name)
	}

	// Check for boolean flag (starts with - or --)
	if !required && (strings.HasPrefix(name, "-") || strings.HasPrefix(name, "--")) {
		originalFlag = name
		name = strings.TrimLeft(name, "-")
		if description == "" {
			description = fmt.Sprintf("Enable %s flag", originalFlag)
		}
	}

	return FieldToken{
		Name:         name,
		Description:  description,
		Required:     required,
		IsArray:      isArray,
		OriginalFlag: originalFlag,
	}
}
