package builders

import (
	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// NumberBuilder provides a fluent interface for building NumberSchema instances.
// It implements core.NumberSchemaBuilder interface and returns core.NumberSchema.
type NumberBuilder struct {
	config schemas.NumberSchemaConfig
}

// Ensure NumberBuilder implements the API interface at compile time
var _ core.NumberSchemaBuilder = (*NumberBuilder)(nil)

// NewNumberSchema creates a new NumberBuilder for creating number schemas.
func NewNumberSchema() core.NumberSchemaBuilder {
	return &NumberBuilder{
		config: schemas.NumberSchemaConfig{
			Metadata: core.SchemaMetadata{},
		},
	}
}

// Build returns the constructed NumberSchema as an core.NumberSchema.
func (b *NumberBuilder) Build() core.NumberSchema {
	return schemas.NewNumberSchema(b.config)
}

// Description sets the description metadata.
func (b *NumberBuilder) Description(desc string) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *NumberBuilder) Name(name string) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *NumberBuilder) Tag(tag string) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Min sets the minimum value constraint.
func (b *NumberBuilder) Min(min float64) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Minimum = &min
	return clone
}

// Max sets the maximum value constraint.
func (b *NumberBuilder) Max(max float64) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Maximum = &max
	return clone
}

// Range sets both minimum and maximum value constraints.
func (b *NumberBuilder) Range(min, max float64) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Minimum = &min
	clone.config.Maximum = &max
	return clone
}

// Example adds an example value to the metadata.
func (b *NumberBuilder) Example(example float64) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *NumberBuilder) Default(value float64) core.NumberSchemaBuilder {
	clone := b.clone()
	clone.config.DefaultVal = &value
	return clone
}

// Common number type helpers

// Positive ensures the number is positive (> 0).
func (b *NumberBuilder) Positive() core.NumberSchemaBuilder {
	return b.Min(0.0000001).
		Description("Positive number").
		Example(1.0)
}

// NonNegative ensures the number is non-negative (≥ 0).
func (b *NumberBuilder) NonNegative() core.NumberSchemaBuilder {
	return b.Min(0.0).
		Description("Non-negative number").
		Example(0.0)
}

// Percentage creates a percentage value (0-100).
func (b *NumberBuilder) Percentage() core.NumberSchemaBuilder {
	return b.Range(0, 100).
		Description("Percentage value").
		Example(50.0)
}

// Ratio creates a ratio value (0-1).
func (b *NumberBuilder) Ratio() core.NumberSchemaBuilder {
	return b.Range(0, 1).
		Description("Ratio value").
		Example(0.5)
}

// clone creates a deep copy of the builder to ensure immutability.
func (b *NumberBuilder) clone() *NumberBuilder {
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

	return &NumberBuilder{config: newConfig}
}
