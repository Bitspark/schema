package schema

import "defs.dev/schema/api"

// Type aliases for transitional compatibility
// These allow gradual migration from concrete types to API interfaces

// SchemaMetadata alias to the API version
type SchemaMetadata = api.SchemaMetadata

// SchemaType alias to the API version
type SchemaType = api.SchemaType

// ValidationResult alias to the API version
type ValidationResult = api.ValidationResult

// ValidationError alias to the API version
type ValidationError = api.ValidationError

// Schema type constants - these are now aliases to the API constants
const (
	TypeObject    = api.TypeObject
	TypeArray     = api.TypeArray
	TypeString    = api.TypeString
	TypeNumber    = api.TypeNumber
	TypeInteger   = api.TypeInteger
	TypeBoolean   = api.TypeBoolean
	TypeNull      = api.TypeNull
	TypeAny       = api.TypeAny
	TypeOptional  = api.TypeOptional
	TypeResult    = api.TypeResult
	TypeMap       = api.TypeMap
	TypeUnion     = api.TypeUnion
	TypeRef       = api.TypeRef
	TypeParameter = api.TypeParameter
	TypeFunction  = api.TypeFunction
)

// Legacy Schema interface for backward compatibility
// This will be gradually phased out in favor of api.Schema
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
