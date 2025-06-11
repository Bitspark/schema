package schema

// Core interfaces
type Schema interface {
	// Validation
	Validate(value any) ValidationResult

	// JSON Schema generation
	ToJSONSchema() map[string]any

	// Metadata
	Type() SchemaType
	Metadata() SchemaMetadata
	WithMetadata(metadata SchemaMetadata) Schema

	// Example generation
	GenerateExample() any

	// Utilities
	Clone() Schema
}

type SchemaType string

const (
	TypeObject    SchemaType = "object"
	TypeArray     SchemaType = "array"
	TypeString    SchemaType = "string"
	TypeNumber    SchemaType = "number"
	TypeInteger   SchemaType = "integer"
	TypeBoolean   SchemaType = "boolean"
	TypeNull      SchemaType = "null"
	TypeAny       SchemaType = "any"
	TypeOptional  SchemaType = "optional"
	TypeResult    SchemaType = "result"
	TypeMap       SchemaType = "map"
	TypeUnion     SchemaType = "union"
	TypeRef       SchemaType = "ref"
	TypeParameter SchemaType = "parameter"
	TypeFunction  SchemaType = "function"
)

type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

type ValidationError struct {
	Path       string `json:"path"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Value      any    `json:"value,omitempty"`
	Expected   string `json:"expected,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Context    string `json:"context,omitempty"`
}

type SchemaMetadata struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Examples    []any             `json:"examples,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}
