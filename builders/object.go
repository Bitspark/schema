package builders

import (
	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// ObjectBuilder provides a fluent API for building ObjectSchemas.
type ObjectBuilder struct {
	config schemas.ObjectSchemaConfig
}

// Ensure ObjectBuilder implements the API interface at compile time
var _ core.ObjectSchemaBuilder = (*ObjectBuilder)(nil)

// NewObjectSchema creates a new ObjectBuilder with default configuration.
func NewObjectSchema() *ObjectBuilder {
	return &ObjectBuilder{
		config: schemas.ObjectSchemaConfig{
			Metadata:             core.SchemaMetadata{},
			Properties:           make(map[string]core.Schema),
			Required:             []string{},
			AdditionalProperties: true, // Default to allowing additional properties
		},
	}
}

// NewObject creates a new ObjectBuilder - this is the public API entry point.
func NewObject() core.ObjectSchemaBuilder {
	return NewObjectSchema()
}

// Build returns the constructed ObjectSchema.
func (b *ObjectBuilder) Build() core.ObjectSchema {
	return schemas.NewObjectSchema(b.config)
}

// Description sets the schema description.
func (b *ObjectBuilder) Description(desc string) core.ObjectSchemaBuilder {
	b.config.Metadata.Description = desc
	return b
}

// Name sets the schema name.
func (b *ObjectBuilder) Name(name string) core.ObjectSchemaBuilder {
	b.config.Metadata.Name = name
	return b
}

// Tag adds a tag to the schema.
func (b *ObjectBuilder) Tag(tag string) core.ObjectSchemaBuilder {
	b.config.Metadata.Tags = append(b.config.Metadata.Tags, tag)
	return b
}

// Property adds a property with its schema to the object.
func (b *ObjectBuilder) Property(name string, schema core.Schema) core.ObjectSchemaBuilder {
	if b.config.Properties == nil {
		b.config.Properties = make(map[string]core.Schema)
	}
	b.config.Properties[name] = schema
	return b
}

// Required marks properties as required.
func (b *ObjectBuilder) Required(names ...string) core.ObjectSchemaBuilder {
	b.config.Required = append(b.config.Required, names...)
	return b
}

// AdditionalProperties sets whether additional properties are allowed.
func (b *ObjectBuilder) AdditionalProperties(allowed bool) core.ObjectSchemaBuilder {
	b.config.AdditionalProperties = allowed
	return b
}

// Example adds an example value.
func (b *ObjectBuilder) Example(example map[string]any) core.ObjectSchemaBuilder {
	b.config.Metadata.Examples = append(b.config.Metadata.Examples, example)
	return b
}

// Default sets the default value.
func (b *ObjectBuilder) Default(value map[string]any) *ObjectBuilder {
	b.config.DefaultVal = value
	return b
}

// MinProperties sets the minimum number of properties.
func (b *ObjectBuilder) MinProperties(min int) *ObjectBuilder {
	b.config.MinProperties = &min
	return b
}

// MaxProperties sets the maximum number of properties.
func (b *ObjectBuilder) MaxProperties(max int) *ObjectBuilder {
	b.config.MaxProperties = &max
	return b
}

// PropertyCount sets both min and max properties to the same value.
func (b *ObjectBuilder) PropertyCount(count int) *ObjectBuilder {
	b.config.MinProperties = &count
	b.config.MaxProperties = &count
	return b
}

// PropertyRange sets the range of allowed properties.
func (b *ObjectBuilder) PropertyRange(min, max int) *ObjectBuilder {
	b.config.MinProperties = &min
	b.config.MaxProperties = &max
	return b
}

// PatternProperty adds a pattern-based property validation.
func (b *ObjectBuilder) PatternProperty(pattern string, schema core.Schema) *ObjectBuilder {
	if b.config.PatternProperties == nil {
		b.config.PatternProperties = make(map[string]core.Schema)
	}
	b.config.PatternProperties[pattern] = schema
	return b
}

// PropertyDependency adds a property dependency (if propName exists, dependencies are required).
func (b *ObjectBuilder) PropertyDependency(propName string, dependencies ...string) *ObjectBuilder {
	if b.config.PropertyDependencies == nil {
		b.config.PropertyDependencies = make(map[string][]string)
	}
	b.config.PropertyDependencies[propName] = append(b.config.PropertyDependencies[propName], dependencies...)
	return b
}

// Strict disallows additional properties and enforces strict validation.
func (b *ObjectBuilder) Strict() *ObjectBuilder {
	b.config.AdditionalProperties = false
	return b
}

// Flexible allows additional properties and relaxed validation.
func (b *ObjectBuilder) Flexible() *ObjectBuilder {
	b.config.AdditionalProperties = true
	return b
}

// RequiredProperty adds a property and marks it as required in one call.
func (b *ObjectBuilder) RequiredProperty(name string, schema core.Schema) *ObjectBuilder {
	b.Property(name, schema)
	b.Required(name)
	return b
}

// OptionalProperty adds a property without marking it as required.
func (b *ObjectBuilder) OptionalProperty(name string, schema core.Schema) *ObjectBuilder {
	b.Property(name, schema)
	return b
}

// NonEmpty ensures the object has at least one property.
func (b *ObjectBuilder) NonEmpty() *ObjectBuilder {
	min := 1
	b.config.MinProperties = &min
	return b
}

// Empty allows empty objects.
func (b *ObjectBuilder) Empty() *ObjectBuilder {
	min := 0
	b.config.MinProperties = &min
	return b
}

// FixedSize ensures the object has exactly the specified number of properties.
func (b *ObjectBuilder) FixedSize(size int) *ObjectBuilder {
	b.config.MinProperties = &size
	b.config.MaxProperties = &size
	return b
}

// Bounded sets both min and max property limits.
func (b *ObjectBuilder) Bounded(min, max int) *ObjectBuilder {
	b.config.MinProperties = &min
	b.config.MaxProperties = &max
	return b
}

// Dict creates a dictionary-like object with string keys and uniform value schema.
func (b *ObjectBuilder) Dict(valueSchema core.Schema) *ObjectBuilder {
	b.PatternProperty("*", valueSchema) // Universal pattern
	b.config.AdditionalProperties = true
	return b
}

// Record creates a record-like object with specified properties and no additional properties.
func (b *ObjectBuilder) Record() *ObjectBuilder {
	b.config.AdditionalProperties = false
	return b
}

// Partial makes all properties optional (removes required constraints).
func (b *ObjectBuilder) Partial() *ObjectBuilder {
	b.config.Required = []string{}
	return b
}

// DeepPartial makes the object and all nested objects partial.
// Note: This is a builder hint - actual deep partial logic would need schema traversal.
func (b *ObjectBuilder) DeepPartial() *ObjectBuilder {
	b.config.Required = []string{}
	b.Tag("deep_partial")
	return b
}

// Common domain-specific helper methods

// PersonExample creates a typical person object structure example.
// This is a simple example - for full domain schemas, create them in application code.
func (b *ObjectBuilder) PersonExample() *ObjectBuilder {
	b.Description("Person information")
	b.Example(map[string]any{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	})
	b.Tag("person")
	return b
}

// ConfigExample creates a configuration object structure example.
func (b *ObjectBuilder) ConfigExample() *ObjectBuilder {
	b.Description("Configuration object")
	b.AdditionalProperties(true) // Allow additional config properties
	b.Tag("config")
	return b
}

// APIResponseExample creates a typical API response structure example.
func (b *ObjectBuilder) APIResponseExample() *ObjectBuilder {
	b.Description("API response structure")
	b.Example(map[string]any{
		"success": true,
		"data":    map[string]any{"id": 1, "name": "example"},
		"message": "Operation successful",
	})
	b.Tag("api_response")
	return b
}
