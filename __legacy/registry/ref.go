package registry

import (
	"fmt"

	"defs.dev/schema"
)

// SchemaRef represents a reference to a named schema with optional parameters
type SchemaRef struct {
	registry   *Registry
	name       string
	parameters map[string]schema.Schema
}

// Type returns the schema type for references
func (r *SchemaRef) Type() schema.SchemaType {
	return schema.TypeRef
}

// WithParam adds a parameter to the schema reference and returns a new reference
func (r *SchemaRef) WithParam(name string, s schema.Schema) *SchemaRef {
	newRef := *r
	if newRef.parameters == nil {
		newRef.parameters = make(map[string]schema.Schema)
	}
	newRef.parameters[name] = s
	return &newRef
}

// Resolve resolves the reference to a concrete schema
func (r *SchemaRef) Resolve() (schema.Schema, error) {
	if len(r.parameters) == 0 {
		return r.registry.Get(r.name)
	}
	return r.registry.Apply(r.name, r.parameters)
}

// Validate validates a value against the resolved schema
func (r *SchemaRef) Validate(value any) schema.ValidationResult {
	resolved, err := r.Resolve()
	if err != nil {
		return schema.ValidationResult{
			Valid: false,
			Errors: []schema.ValidationError{{
				Path:    "",
				Message: fmt.Sprintf("Failed to resolve schema reference '%s': %s", r.name, err.Error()),
				Code:    "resolution_error",
			}},
		}
	}
	return resolved.Validate(value)
}

// ToJSONSchema converts the reference to JSON Schema format
func (r *SchemaRef) ToJSONSchema() map[string]any {
	resolved, err := r.Resolve()
	if err != nil {
		// Return a JSON Schema reference if we can't resolve
		return map[string]any{
			"$ref": fmt.Sprintf("#/definitions/%s", r.name),
		}
	}
	return resolved.ToJSONSchema()
}

// Metadata returns metadata from the resolved schema
func (r *SchemaRef) Metadata() schema.SchemaMetadata {
	resolved, err := r.Resolve()
	if err != nil {
		return schema.SchemaMetadata{
			Name:        r.name,
			Description: fmt.Sprintf("Reference to schema '%s'", r.name),
		}
	}
	return resolved.Metadata()
}

// WithMetadata creates a new reference with additional metadata
func (r *SchemaRef) WithMetadata(metadata schema.SchemaMetadata) schema.Schema {
	// For references, we maintain the reference identity but could store metadata separately
	// This is a design decision - for now, we'll just return the same reference
	newRef := *r
	return &newRef
}

// Clone creates a deep copy of the schema reference
func (r *SchemaRef) Clone() schema.Schema {
	clone := *r
	if r.parameters != nil {
		clone.parameters = make(map[string]schema.Schema)
		for k, v := range r.parameters {
			clone.parameters[k] = v.Clone()
		}
	}
	return &clone
}

// GenerateExample generates an example value from the resolved schema
func (r *SchemaRef) GenerateExample() any {
	resolved, err := r.Resolve()
	if err != nil {
		return nil
	}
	return resolved.GenerateExample()
}
