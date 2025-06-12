// Package core defines the core interfaces and types for the schema system.
// This package contains only interfaces and types, no implementations.
package core

// Schema is the core interface that all schema types must implement.
// It provides metadata handling and structural information.
type Schema interface {
	// Type is the schema type.
	Type() SchemaType

	// Annotations return the annotations of the schema.
	Annotations() []Annotation

	// Metadata returns the schema metadata.
	Metadata() SchemaMetadata

	// Clone clones the schema.
	Clone() Schema
}

// SchemaType represents the type of a schema (string, number, structure, etc.).
type SchemaType string

const (
	TypeStructure SchemaType = "structure"
	TypeArray     SchemaType = "array"
	TypeString    SchemaType = "string"
	TypeNumber    SchemaType = "number"
	TypeInteger   SchemaType = "integer"
	TypeBoolean   SchemaType = "boolean"
	TypeNull      SchemaType = "null"
	TypeAny       SchemaType = "any"
	TypeOptional  SchemaType = "optional"
	TypeMap       SchemaType = "map"
	TypeUnion     SchemaType = "union"
	TypeRef       SchemaType = "ref"
	TypeParameter SchemaType = "parameter"
	TypeFunction  SchemaType = "function"
	TypeService   SchemaType = "service"

	// Validation schema types for file system validation
	TypeFileValidation      SchemaType = "file-validation"
	TypeDirectoryValidation SchemaType = "directory-validation"
	TypeNodeValidation      SchemaType = "node-validation"
)

// SchemaMetadata contains descriptive information about a schema.
type SchemaMetadata struct {
	Name        string            `json:"name,omitempty"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	Examples    []any             `json:"examples,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

func (m SchemaMetadata) ToMap() map[string]any {
	return map[string]any{
		"name":        m.Name,
		"version":     m.Version,
		"description": m.Description,
		"examples":    m.Examples,
		"tags":        m.Tags,
		"properties":  m.Properties,
	}
}
