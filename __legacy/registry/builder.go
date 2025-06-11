package registry

import "defs.dev/schema"

// NamedSchemaBuilder provides a fluent interface for building parameterized schemas
type NamedSchemaBuilder struct {
	registry *Registry
	name     string
	params   map[string]schema.Schema
}

// WithParam adds a parameter to the builder
func (b *NamedSchemaBuilder) WithParam(name string, s schema.Schema) *NamedSchemaBuilder {
	if b.params == nil {
		b.params = make(map[string]schema.Schema)
	}
	b.params[name] = s
	return b
}

// Build applies the parameters and returns the resolved schema
func (b *NamedSchemaBuilder) Build() (schema.Schema, error) {
	return b.registry.Apply(b.name, b.params)
}
