package blueprint

import (
	"fmt"
	"strings"
)

// normalizeFieldName converts dashes to underscores for field name matching
func normalizeFieldName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// findParamValue finds a parameter value, handling dash-underscore equivalence
func findParamValue(params map[string]interface{}, fieldName string) (interface{}, bool) {
	// Try exact match first
	if value, exists := params[fieldName]; exists {
		return value, true
	}

	// Try normalized version (dash to underscore)
	normalizedField := normalizeFieldName(fieldName)
	if value, exists := params[normalizedField]; exists {
		return value, true
	}

	// Try the reverse (underscore to dash)
	dashedField := strings.ReplaceAll(fieldName, "_", "-")
	if value, exists := params[dashedField]; exists {
		return value, true
	}

	return nil, false
}

// BuildCommandArgs builds the actual command arguments from the tokenized template
func (tbp *TokenizedBlueprint) BuildCommandArgs(params map[string]interface{}) ([]string, error) {
	// Validate required parameters
	for _, required := range tbp.InputSchema.Required {
		if _, exists := findParamValue(params, required); !exists {
			return nil, fmt.Errorf("missing required parameter: %s", required)
		}
	}

	// Validate parameter types
	for name, param := range params {
		if schema, exists := tbp.InputSchema.Properties[normalizeFieldName(name)]; exists {
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

	for _, shellWord := range tbp.ShellWords {
		// Check if this shell word should be included
		shouldInclude, wordResult := tbp.renderShellWord(shellWord, params)
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
func (tbp *TokenizedBlueprint) renderShellWord(tokens []Token, params map[string]interface{}) (bool, []string) {
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
					if tbp.hasValue(value) {
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
			// Check if this is an array field first (arrays take precedence)
			if schema, exists := tbp.InputSchema.Properties[normalizeFieldName(fieldToken.Name)]; exists && schema.Type == "array" {
				return tbp.renderArrayField(fieldToken, params)
			}

			// Then check if it's an optional field
			if !fieldToken.Required {
				return tbp.renderSingleOptionalField(fieldToken, params)
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
				if strValue := tbp.valueToString(value); strValue != "" {
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
func (tbp *TokenizedBlueprint) renderSingleOptionalField(fieldToken FieldToken, params map[string]interface{}) (bool, []string) {
	value, exists := findParamValue(params, fieldToken.Name)
	if !exists {
		return false, nil
	}

	// Check if this is a boolean flag
	if schema, schemaExists := tbp.InputSchema.Properties[normalizeFieldName(fieldToken.Name)]; schemaExists && schema.Type == "boolean" {
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
	if strValue := tbp.valueToString(value); strValue != "" {
		return true, []string{strValue}
	}

	return false, nil
}

// renderArrayField handles rendering of array fields
func (tbp *TokenizedBlueprint) renderArrayField(fieldToken FieldToken, params map[string]interface{}) (bool, []string) {
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
func (tbp *TokenizedBlueprint) hasValue(value interface{}) bool {
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
func (tbp *TokenizedBlueprint) valueToString(value interface{}) string {
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
	// Use the tokenized approach by creating a temporary TokenizedBlueprint
	tbp := &TokenizedBlueprint{
		BaseCommand:     bp.BaseCommand,
		ToolName:        bp.ToolName,
		ToolDescription: bp.ToolDescription,
		InputSchema:     bp.InputSchema,
		ShellWords:      bp.ShellWords,
	}

	return tbp.BuildCommandArgs(params)
}
