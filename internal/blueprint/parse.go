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

	// Check for boolean flag patterns first
	if matches := booleanFlagRegex.FindStringSubmatch(word); matches != nil {
		flag := matches[1]
		description := ""
		if len(matches) > 2 && matches[2] != "" {
			description = strings.TrimSpace(matches[2])
		}

		// Keep original flag name for the token
		flagName := strings.TrimLeft(flag, "-")

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

		return tokens
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
			IsArray:     isArray,
		}
		tokens = append(tokens, token)

		return tokens
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
			originalName := word[nameStart:nameEnd] // Keep original spacing
			description := ""
			if descStart > 0 && descEnd > descStart {
				description = strings.TrimSpace(word[descStart:descEnd])
			}

			// Add field token with original name
			token := FieldToken{
				Name:         varName,      // Keep trimmed name for processing
				OriginalName: originalName, // Keep original name with spacing for display
				Description:  description,
				Required:     true,
			}
			tokens = append(tokens, token)

			currentPos = end
		}

		// Add remaining text after the last template
		if currentPos < len(word) {
			textAfter := word[currentPos:]
			if textAfter != "" {
				tokens = append(tokens, TextToken{Value: textAfter})
			}
		}
	} else {
		// If no templates were found, treat the entire word as text
		tokens = append(tokens, TextToken{Value: word})
	}

	return tokens
}
