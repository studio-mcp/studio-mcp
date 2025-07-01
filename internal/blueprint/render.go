package blueprint

import (
	"fmt"
	"strings"
)

// normalizeFieldName converts field names to use underscores instead of dashes
func normalizeFieldName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// findParamValue finds a parameter value by name, handling dash-underscore equivalence
func findParamValue(params map[string]interface{}, fieldName string) (interface{}, bool) {
	// Try exact match first
	if value, exists := params[fieldName]; exists {
		return value, true
	}

	// Try normalized version (dashes to underscores)
	normalized := normalizeFieldName(fieldName)
	if value, exists := params[normalized]; exists {
		return value, true
	}

	// Try reverse (underscores to dashes) if original had underscores
	if strings.Contains(fieldName, "_") {
		dashed := strings.ReplaceAll(fieldName, "_", "-")
		if value, exists := params[dashed]; exists {
			return value, true
		}
	}

	return nil, false
}

// buildCommandArgsTokenized builds the actual command arguments using the tokenized approach
func (bp *Blueprint) buildCommandArgsTokenized(params map[string]interface{}) ([]string, error) {
	inputSchema := bp.GenerateInputSchema()

	// Validate required parameters
	for _, required := range inputSchema.Required {
		if _, exists := findParamValue(params, required); !exists {
			return nil, fmt.Errorf("missing required parameter: %s", required)
		}
	}

	// Validate parameter types
	for name, param := range params {
		if schema, exists := inputSchema.Properties[normalizeFieldName(name)]; exists {
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

	result := []string{}

	for _, shellWord := range bp.ShellWords {
		// Check if this shell word should be included
		shouldInclude, wordResult := bp.renderShellWord(shellWord, params)
		if shouldInclude {
			if len(wordResult) == 0 {
				// Empty result means skip this word
				continue
			}
			result = append(result, wordResult...)
		}
	}

	return result, nil
}

// renderShellWord renders a single shell word from its tokens
func (bp *Blueprint) renderShellWord(tokens []Token, params map[string]interface{}) (bool, []string) {
	// Check if this word contains only optional fields that are not provided
	hasRequiredContent := false
	allOptionalFieldsEmpty := true

	// First pass: check if we should include this word at all
	for _, token := range tokens {
		switch t := token.(type) {
		case TextToken:
			hasRequiredContent = true
		case FieldToken:
			if value, exists := findParamValue(params, t.Name); exists {
				if t.Required {
					hasRequiredContent = true
					allOptionalFieldsEmpty = false
				} else {
					// Optional field - check if it has a meaningful value
					if bp.hasValue(value) {
						allOptionalFieldsEmpty = false
					}
				}
			} else if t.Required {
				hasRequiredContent = true
			}
		}
	}

	// If this word has only optional fields and they're all empty, skip it
	if !hasRequiredContent && allOptionalFieldsEmpty {
		return false, nil
	}

	// Handle special cases for single field tokens
	if len(tokens) == 1 {
		if fieldToken, ok := tokens[0].(FieldToken); ok {
			inputSchema := bp.GenerateInputSchema()
			// Check if this is an array field first (arrays take precedence)
			if schema, exists := inputSchema.Properties[normalizeFieldName(fieldToken.Name)]; exists && schema.Type == "array" {
				return bp.renderArrayField(fieldToken, params)
			}

			// Then check if it's an optional field
			if !fieldToken.Required {
				return bp.renderSingleOptionalField(fieldToken, params)
			}
		}
	}

	// Render as a single concatenated string
	var parts []string
	for _, token := range tokens {
		switch t := token.(type) {
		case TextToken:
			parts = append(parts, t.Value)
		case FieldToken:
			if value, exists := findParamValue(params, t.Name); exists {
				if strValue := bp.valueToString(value); strValue != "" {
					parts = append(parts, strValue)
				}
			}
		}
	}

	if len(parts) > 0 {
		return true, []string{strings.Join(parts, "")}
	}

	return false, nil
}

// renderSingleOptionalField handles rendering of a single optional field token
func (bp *Blueprint) renderSingleOptionalField(fieldToken FieldToken, params map[string]interface{}) (bool, []string) {
	value, exists := findParamValue(params, fieldToken.Name)
	if !exists {
		return false, nil
	}

	inputSchema := bp.GenerateInputSchema()
	// Check if this is a boolean flag
	if schema, schemaExists := inputSchema.Properties[normalizeFieldName(fieldToken.Name)]; schemaExists && schema.Type == "boolean" {
		if boolValue, ok := value.(bool); ok {
			if boolValue {
				// Use the original flag format if available, otherwise construct it
				if fieldToken.OriginalFlag != "" {
					return true, []string{fieldToken.OriginalFlag}
				} else {
					return true, []string{"-" + fieldToken.Name}
				}
			}
		}
		return false, nil
	}

	// Regular optional field
	if strValue := bp.valueToString(value); strValue != "" {
		return true, []string{strValue}
	}

	return false, nil
}

// renderArrayField handles rendering of array fields
func (bp *Blueprint) renderArrayField(fieldToken FieldToken, params map[string]interface{}) (bool, []string) {
	value, exists := findParamValue(params, fieldToken.Name)
	if !exists {
		return false, nil
	}

	if arr, ok := value.([]string); ok {
		return len(arr) > 0, arr
	} else if arr, ok := value.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return len(result) > 0, result
	}

	return false, nil
}

// hasValue checks if a value is meaningful (not empty)
func (bp *Blueprint) hasValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v != ""
	case bool:
		return v
	case []string:
		return len(v) > 0
	case []interface{}:
		return len(v) > 0
	default:
		return value != nil
	}
}

// valueToString converts a value to its string representation
func (bp *Blueprint) valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", value)
	}
}

// BuildCommandArgs builds the actual command arguments from the template
func (bp *Blueprint) BuildCommandArgs(params map[string]interface{}) ([]string, error) {
	// Use the tokenized approach directly
	return bp.buildCommandArgsTokenized(params)
}
