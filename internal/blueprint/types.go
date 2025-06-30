package blueprint

import (
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// Token represents a part of a shell word after parsing
type Token interface {
	String() string
}

// TextToken represents literal text in a shell word
type TextToken struct {
	Value string
}

func (t TextToken) String() string {
	return t.Value
}

// FieldToken represents a template field in a shell word
type FieldToken struct {
	Name         string
	Description  string
	Required     bool
	OriginalFlag string // For boolean flags, stores the original flag format (e.g., "-f", "--verbose")
}

func (t FieldToken) String() string {
	if t.Required {
		return "{{" + t.Name + "}}"
	}
	return "[" + t.Name + "]"
}

// TokenizedBlueprint represents the tokenized result of parsing
type TokenizedBlueprint struct {
	BaseCommand     string
	ToolName        string
	ToolDescription string
	InputSchema     *jsonschema.Schema
	ShellWords      [][]Token // Array of token arrays, one per shell word
}

// Blueprint represents a parsed command template
type Blueprint struct {
	BaseCommand     string
	ToolName        string
	ToolDescription string
	InputSchema     *jsonschema.Schema
	ShellWords      [][]Token // Tokenized shell words
}
