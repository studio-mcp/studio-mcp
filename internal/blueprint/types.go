package blueprint

import (
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// Blueprint represents a parsed command template
type Blueprint struct {
	BaseCommand     string
	ToolName        string
	ToolDescription string
	InputSchema     *jsonschema.Schema
	args            []string
	fields          []field
}

type field struct {
	argIndex    int
	name        string
	description string
	isArray     bool
	isOptional  bool
	formatter   func(interface{}) []string
}
