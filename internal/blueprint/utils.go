package blueprint

import (
	"regexp"
)

var (
	// Matches {{variable}} or {{variable#description}}
	templateRegex = regexp.MustCompile(`\{\{([^#}]+)(?:#([^}]+))?\}\}`)
	// Matches [variable] or [variable#description] or [variable...] or [variable...#description]
	optionalRegex = regexp.MustCompile(`^\[([^#.\]]+)(?:\.\.\.)?(?:#([^.\]]+))?(\.\.\.)?\]$`)
	// Matches boolean flags like [-f] or [--flag] or [-f#description] or [--flag#description]
	booleanFlagRegex = regexp.MustCompile(`^\[(-{1,2}[a-zA-Z][a-zA-Z0-9-]*)(?:#([^\]]+))?\]$`)
)

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// formatString formats a string value
func formatString(value interface{}) []string {
	if str, ok := value.(string); ok && str != "" {
		return []string{str}
	}
	return []string{}
}

// formatArray formats an array value
func formatArray(value interface{}) []string {
	if arr, ok := value.([]string); ok {
		return arr
	} else if arr, ok := value.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// formatBoolean formats a boolean flag value
func formatBoolean(flag string) func(interface{}) []string {
	return func(value interface{}) []string {
		if b, ok := value.(bool); ok && b {
			return []string{flag}
		}
		return []string{}
	}
}

// getFormatter returns the appropriate formatter function
func getFormatter(isArray, isBoolean bool, flag ...string) func(interface{}) []string {
	if isArray {
		return formatArray
	}
	if isBoolean && len(flag) > 0 {
		return formatBoolean(flag[0])
	}
	return formatString
}
