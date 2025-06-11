// Package api defines the core interfaces and types for the schema system.
// This package contains only interfaces and types, no implementations.
package api

// Schema is the core interface that all schema types must implement.
// It provides validation, JSON Schema generation, metadata handling, and example generation.
type Schema interface {
	// Validation
	Validate(value any) ValidationResult

	// JSON Schema generation
	ToJSONSchema() map[string]any

	// Type
	Type() SchemaType

	// Metadata
	Metadata() SchemaMetadata

	// Example generation
	GenerateExample() any

	// Utilities
	Clone() Schema
}

// SchemaType represents the type of a schema (string, number, object, etc.).
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
	TypeService   SchemaType = "service"
)

// ValidationResult represents the result of validating a value against a schema.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// ValidationError represents a single validation error with context and suggestions.
type ValidationError struct {
	Path       string `json:"path"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Value      any    `json:"value,omitempty"`
	Expected   string `json:"expected,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Context    string `json:"context,omitempty"`
}

// SchemaMetadata contains descriptive information about a schema.
type SchemaMetadata struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Examples    []any             `json:"examples,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}
