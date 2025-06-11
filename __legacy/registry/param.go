package registry

import (
	"fmt"

	"defs.dev/schema"
)

// ParameterRef represents an unresolved parameter reference in a schema template
type ParameterRef struct {
	name string
}

// Param creates a new parameter reference with the given name
func Param(name string) schema.Schema {
	return &ParameterRef{name: name}
}

// Type returns the schema type for parameter references
func (p *ParameterRef) Type() schema.SchemaType {
	return schema.TypeParameter
}

// Validate always returns an error since unresolved parameters cannot be validated
func (p *ParameterRef) Validate(value any) schema.ValidationResult {
	return schema.ValidationResult{
		Valid: false,
		Errors: []schema.ValidationError{{
			Path:    "",
			Message: fmt.Sprintf("Cannot validate unresolved parameter '%s'", p.name),
			Code:    "unresolved_parameter",
		}},
	}
}

// ToJSONSchema returns a JSON Schema reference for the parameter
func (p *ParameterRef) ToJSONSchema() map[string]any {
	return map[string]any{
		"$ref": fmt.Sprintf("#/parameters/%s", p.name),
	}
}

// Metadata returns basic metadata for the parameter
func (p *ParameterRef) Metadata() schema.SchemaMetadata {
	return schema.SchemaMetadata{
		Name:        p.name,
		Description: fmt.Sprintf("Parameter reference: %s", p.name),
	}
}

// WithMetadata returns the same parameter since parameters are immutable references
func (p *ParameterRef) WithMetadata(metadata schema.SchemaMetadata) schema.Schema {
	return p // Parameters maintain their identity
}

// Clone creates a copy of the parameter reference
func (p *ParameterRef) Clone() schema.Schema {
	return &ParameterRef{name: p.name}
}

// GenerateExample returns nil since unresolved parameters cannot generate examples
func (p *ParameterRef) GenerateExample() any {
	return nil
}

// Name returns the parameter name (helper method)
func (p *ParameterRef) Name() string {
	return p.name
}
