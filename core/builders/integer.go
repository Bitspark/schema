package builders

import (
	"defs.dev/schema/api"
	"defs.dev/schema/core/schemas"
)

// IntegerBuilder provides a fluent interface for building IntegerSchema instances.
// It implements api.IntegerSchemaBuilder interface and returns api.IntegerSchema.
type IntegerBuilder struct {
	config schemas.IntegerSchemaConfig
}

// Ensure IntegerBuilder implements the API interface at compile time
var _ api.IntegerSchemaBuilder = (*IntegerBuilder)(nil)

// NewInteger creates a new IntegerBuilder for creating integer schemas.
func NewInteger() api.IntegerSchemaBuilder {
	return &IntegerBuilder{
		config: schemas.IntegerSchemaConfig{
			Metadata: api.SchemaMetadata{},
		},
	}
}

// Build returns the constructed IntegerSchema as an api.IntegerSchema.
func (b *IntegerBuilder) Build() api.IntegerSchema {
	return schemas.NewIntegerSchema(b.config)
}

// Description sets the description metadata.
func (b *IntegerBuilder) Description(desc string) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *IntegerBuilder) Name(name string) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *IntegerBuilder) Tag(tag string) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Min sets the minimum value constraint.
func (b *IntegerBuilder) Min(min int64) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Minimum = &min
	return clone
}

// Max sets the maximum value constraint.
func (b *IntegerBuilder) Max(max int64) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Maximum = &max
	return clone
}

// Range sets both minimum and maximum value constraints.
func (b *IntegerBuilder) Range(min, max int64) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Minimum = &min
	clone.config.Maximum = &max
	return clone
}

// Example adds an example value to the metadata.
func (b *IntegerBuilder) Example(example int64) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *IntegerBuilder) Default(value int64) api.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.DefaultVal = &value
	return clone
}

// Common integer type helpers

// Positive ensures the integer is positive (> 0).
func (b *IntegerBuilder) Positive() api.IntegerSchemaBuilder {
	return b.Min(1).
		Description("Positive integer").
		Example(1)
}

// NonNegative ensures the integer is non-negative (≥ 0).
func (b *IntegerBuilder) NonNegative() api.IntegerSchemaBuilder {
	return b.Min(0).
		Description("Non-negative integer").
		Example(0)
}

// Port creates a port number (1-65535).
func (b *IntegerBuilder) Port() api.IntegerSchemaBuilder {
	return b.Range(1, 65535).
		Description("Port number").
		Example(8080)
}

// Age creates an age value (0-150).
func (b *IntegerBuilder) Age() api.IntegerSchemaBuilder {
	return b.Range(0, 150).
		Description("Age in years").
		Example(25)
}

// ID creates a positive ID value.
func (b *IntegerBuilder) ID() api.IntegerSchemaBuilder {
	return b.Positive().
		Description("Unique identifier").
		Example(1)
}

// Count creates a count value (≥ 0).
func (b *IntegerBuilder) Count() api.IntegerSchemaBuilder {
	return b.NonNegative().
		Description("Count of items").
		Example(5)
}

// clone creates a deep copy of the builder to ensure immutability.
func (b *IntegerBuilder) clone() *IntegerBuilder {
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

	return &IntegerBuilder{config: newConfig}
}
