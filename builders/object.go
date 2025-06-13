package builders

import (
	"defs.dev/schema/core"
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
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the schema name.
func (b *ObjectBuilder) Name(name string) core.ObjectSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the schema.
func (b *ObjectBuilder) Tag(tag string) core.ObjectSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Property adds a property with its schema to the object.
func (b *ObjectBuilder) Property(name string, schema core.Schema) core.ObjectSchemaBuilder {
	clone := b.clone()
	if clone.config.Properties == nil {
		clone.config.Properties = make(map[string]core.Schema)
	}
	clone.config.Properties[name] = schema
	return clone
}

// Required marks properties as required.
func (b *ObjectBuilder) Required(names ...string) core.ObjectSchemaBuilder {
	clone := b.clone()
	clone.config.Required = append(clone.config.Required, names...)
	return clone
}

// AdditionalProperties sets whether additional properties are allowed.
func (b *ObjectBuilder) AdditionalProperties(allowed bool) core.ObjectSchemaBuilder {
	clone := b.clone()
	clone.config.AdditionalProperties = allowed
	return clone
}

// Example adds an example value.
func (b *ObjectBuilder) Example(example map[string]any) core.ObjectSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *ObjectBuilder) Default(value map[string]any) *ObjectBuilder {
	clone := b.clone()
	clone.config.DefaultVal = value
	return clone
}

// MinProperties sets the minimum number of properties.
func (b *ObjectBuilder) MinProperties(min int) *ObjectBuilder {
	clone := b.clone()
	clone.config.MinProperties = &min
	return clone
}

// MaxProperties sets the maximum number of properties.
func (b *ObjectBuilder) MaxProperties(max int) *ObjectBuilder {
	clone := b.clone()
	clone.config.MaxProperties = &max
	return clone
}

// PropertyCount sets both min and max properties to the same value.
func (b *ObjectBuilder) PropertyCount(count int) *ObjectBuilder {
	clone := b.clone()
	clone.config.MinProperties = &count
	clone.config.MaxProperties = &count
	return clone
}

// PropertyRange sets the range of allowed properties.
func (b *ObjectBuilder) PropertyRange(min, max int) *ObjectBuilder {
	clone := b.clone()
	clone.config.MinProperties = &min
	clone.config.MaxProperties = &max
	return clone
}

// PatternProperty adds a pattern-based property validation.
func (b *ObjectBuilder) PatternProperty(pattern string, schema core.Schema) *ObjectBuilder {
	clone := b.clone()
	if clone.config.PatternProperties == nil {
		clone.config.PatternProperties = make(map[string]core.Schema)
	}
	clone.config.PatternProperties[pattern] = schema
	return clone
}

// PropertyDependency adds a property dependency (if propName exists, dependencies are required).
func (b *ObjectBuilder) PropertyDependency(propName string, dependencies ...string) *ObjectBuilder {
	clone := b.clone()
	if clone.config.PropertyDependencies == nil {
		clone.config.PropertyDependencies = make(map[string][]string)
	}
	clone.config.PropertyDependencies[propName] = append(clone.config.PropertyDependencies[propName], dependencies...)
	return clone
}

// Strict disallows additional properties and enforces strict validation.
func (b *ObjectBuilder) Strict() *ObjectBuilder {
	clone := b.clone()
	clone.config.AdditionalProperties = false
	return clone
}

// Flexible allows additional properties and relaxed validation.
func (b *ObjectBuilder) Flexible() *ObjectBuilder {
	clone := b.clone()
	clone.config.AdditionalProperties = true
	return clone
}

// RequiredProperty adds a property and marks it as required in one call.
func (b *ObjectBuilder) RequiredProperty(name string, schema core.Schema) *ObjectBuilder {
	return b.Property(name, schema).(*ObjectBuilder).Required(name).(*ObjectBuilder)
}

// OptionalProperty adds a property without marking it as required.
func (b *ObjectBuilder) OptionalProperty(name string, schema core.Schema) *ObjectBuilder {
	return b.Property(name, schema).(*ObjectBuilder)
}

// NonEmpty ensures the object has at least one property.
func (b *ObjectBuilder) NonEmpty() *ObjectBuilder {
	min := 1
	clone := b.clone()
	clone.config.MinProperties = &min
	return clone
}

// Empty allows empty objects.
func (b *ObjectBuilder) Empty() *ObjectBuilder {
	min := 0
	clone := b.clone()
	clone.config.MinProperties = &min
	return clone
}

// FixedSize ensures the object has exactly the specified number of properties.
func (b *ObjectBuilder) FixedSize(size int) *ObjectBuilder {
	clone := b.clone()
	clone.config.MinProperties = &size
	clone.config.MaxProperties = &size
	return clone
}

// Bounded sets both min and max property limits.
func (b *ObjectBuilder) Bounded(min, max int) *ObjectBuilder {
	clone := b.clone()
	clone.config.MinProperties = &min
	clone.config.MaxProperties = &max
	return clone
}

// Dict creates a dictionary-like object with string keys and uniform value schema.
func (b *ObjectBuilder) Dict(valueSchema core.Schema) *ObjectBuilder {
	clone := b.clone()
	if clone.config.PatternProperties == nil {
		clone.config.PatternProperties = make(map[string]core.Schema)
	}
	clone.config.PatternProperties["*"] = valueSchema // Universal pattern
	clone.config.AdditionalProperties = true
	return clone
}

// Record creates a record-like object with specified properties and no additional properties.
func (b *ObjectBuilder) Record() *ObjectBuilder {
	clone := b.clone()
	clone.config.AdditionalProperties = false
	return clone
}

// Partial makes all properties optional (removes required constraints).
func (b *ObjectBuilder) Partial() *ObjectBuilder {
	clone := b.clone()
	clone.config.Required = []string{}
	return clone
}

// DeepPartial makes the object and all nested objects partial.
// Note: This is a builder hint - actual deep partial logic would need schema traversal.
func (b *ObjectBuilder) DeepPartial() *ObjectBuilder {
	clone := b.clone()
	clone.config.Required = []string{}
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, "deep_partial")
	return clone
}

// Common domain-specific helper methods

// PersonExample creates a typical person object structure example.
// This is a simple example - for full domain schemas, create them in application code.
func (b *ObjectBuilder) PersonExample() *ObjectBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = "Person information"
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, map[string]any{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	})
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, "person")
	return clone
}

// ConfigExample creates a configuration object structure example.
func (b *ObjectBuilder) ConfigExample() *ObjectBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = "Configuration object"
	clone.config.AdditionalProperties = true // Allow additional config properties
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, "config")
	return clone
}

// APIResponseExample creates a typical API response structure example.
func (b *ObjectBuilder) APIResponseExample() *ObjectBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = "API response structure"
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, map[string]any{
		"success": true,
		"data":    map[string]any{"id": 1, "name": "example"},
		"message": "Operation successful",
	})
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, "api_response")
	return clone
}

// clone creates a deep copy of the builder to ensure immutability.
func (b *ObjectBuilder) clone() *ObjectBuilder {
	newConfig := b.config

	// Deep copy slices
	if b.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(b.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, b.config.Metadata.Examples)
	}

	if b.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(b.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, b.config.Metadata.Tags)
	}

	if b.config.Required != nil {
		newConfig.Required = make([]string, len(b.config.Required))
		copy(newConfig.Required, b.config.Required)
	}

	// Deep copy maps
	if b.config.Properties != nil {
		newConfig.Properties = make(map[string]core.Schema)
		for k, v := range b.config.Properties {
			newConfig.Properties[k] = v
		}
	}

	if b.config.PatternProperties != nil {
		newConfig.PatternProperties = make(map[string]core.Schema)
		for k, v := range b.config.PatternProperties {
			newConfig.PatternProperties[k] = v
		}
	}

	if b.config.PropertyDependencies != nil {
		newConfig.PropertyDependencies = make(map[string][]string)
		for k, v := range b.config.PropertyDependencies {
			newConfig.PropertyDependencies[k] = make([]string, len(v))
			copy(newConfig.PropertyDependencies[k], v)
		}
	}

	// Deep copy default value map
	if b.config.DefaultVal != nil {
		newConfig.DefaultVal = make(map[string]any)
		for k, v := range b.config.DefaultVal {
			newConfig.DefaultVal[k] = v
		}
	}

	return &ObjectBuilder{config: newConfig}
}
