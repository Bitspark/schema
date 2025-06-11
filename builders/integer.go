package builders

import (
	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// IntegerBuilder provides a fluent interface for building IntegerSchema instances.
// It implements core.IntegerSchemaBuilder interface and returns core.IntegerSchema.
type IntegerBuilder struct {
	config schemas.IntegerSchemaConfig
}

// Ensure IntegerBuilder implements the API interface at compile time
var _ core.IntegerSchemaBuilder = (*IntegerBuilder)(nil)

// NewIntegerSchema creates a new IntegerBuilder for creating integer schemas.
func NewIntegerSchema() core.IntegerSchemaBuilder {
	return &IntegerBuilder{
		config: schemas.IntegerSchemaConfig{
			Metadata: core.SchemaMetadata{},
		},
	}
}

// Build returns the constructed IntegerSchema as an core.IntegerSchema.
func (b *IntegerBuilder) Build() core.IntegerSchema {
	return schemas.NewIntegerSchema(b.config)
}

// Description sets the description metadata.
func (b *IntegerBuilder) Description(desc string) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *IntegerBuilder) Name(name string) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *IntegerBuilder) Tag(tag string) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Min sets the minimum value constraint.
func (b *IntegerBuilder) Min(min int64) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Minimum = &min
	return clone
}

// Max sets the maximum value constraint.
func (b *IntegerBuilder) Max(max int64) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Maximum = &max
	return clone
}

// Range sets both minimum and maximum value constraints.
func (b *IntegerBuilder) Range(min, max int64) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Minimum = &min
	clone.config.Maximum = &max
	return clone
}

// Example adds an example value to the metadata.
func (b *IntegerBuilder) Example(example int64) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *IntegerBuilder) Default(value int64) core.IntegerSchemaBuilder {
	clone := b.clone()
	clone.config.DefaultVal = &value
	return clone
}

// Common integer type helpers

// Positive ensures the integer is positive (> 0).
func (b *IntegerBuilder) Positive() core.IntegerSchemaBuilder {
	return b.Min(1).
		Description("Positive integer").
		Example(1)
}

// NonNegative ensures the integer is non-negative (≥ 0).
func (b *IntegerBuilder) NonNegative() core.IntegerSchemaBuilder {
	return b.Min(0).
		Description("Non-negative integer").
		Example(0)
}

// Port creates a port number (1-65535).
func (b *IntegerBuilder) Port() core.IntegerSchemaBuilder {
	return b.Range(1, 65535).
		Description("Port number").
		Example(8080)
}

// Age creates an age value (0-150).
func (b *IntegerBuilder) Age() core.IntegerSchemaBuilder {
	return b.Range(0, 150).
		Description("Age in years").
		Example(25)
}

// ID creates a positive ID value.
func (b *IntegerBuilder) ID() core.IntegerSchemaBuilder {
	return b.Positive().
		Description("Unique identifier").
		Example(1)
}

// Count creates a count value (≥ 0).
func (b *IntegerBuilder) Count() core.IntegerSchemaBuilder {
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
