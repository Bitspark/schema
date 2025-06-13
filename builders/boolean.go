package builders

import (
	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// BooleanBuilder provides a fluent interface for building BooleanSchema instances.
// It implements core.BooleanSchemaBuilder interface and returns core.BooleanSchema.
type BooleanBuilder struct {
	config schemas.BooleanSchemaConfig
}

// Ensure BooleanBuilder implements the API interface at compile time
var _ core.BooleanSchemaBuilder = (*BooleanBuilder)(nil)

// NewBooleanSchema creates a new BooleanBuilder for creating boolean schemas.
func NewBooleanSchema() core.BooleanSchemaBuilder {
	return &BooleanBuilder{
		config: schemas.BooleanSchemaConfig{
			Metadata:        core.SchemaMetadata{},
			CaseInsensitive: false,
		},
	}
}

// Build returns the constructed BooleanSchema as an core.BooleanSchema.
func (b *BooleanBuilder) Build() core.BooleanSchema {
	return schemas.NewBooleanSchema(b.config)
}

// Description sets the description metadata.
func (b *BooleanBuilder) Description(desc string) core.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *BooleanBuilder) Name(name string) core.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *BooleanBuilder) Tag(tag string) core.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Example adds an example value to the metadata.
func (b *BooleanBuilder) Example(example bool) core.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *BooleanBuilder) Default(value bool) core.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.DefaultVal = &value
	return clone
}

// CaseInsensitive enables case-insensitive string conversion.
// This automatically enables string conversion if not already enabled.
func (b *BooleanBuilder) CaseInsensitive() core.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.CaseInsensitive = true
	return clone
}

// Common boolean type helpers

// Required creates a required boolean field.
func (b *BooleanBuilder) Required() core.BooleanSchemaBuilder {
	return b.Description("Required boolean value").
		Example(true)
}

// Flag creates a flag-style boolean (defaults to false).
func (b *BooleanBuilder) Flag() core.BooleanSchemaBuilder {
	return b.Default(false).
		Description("Boolean flag").
		Example(false)
}

// Switch creates a switch-style boolean with string conversion.
func (b *BooleanBuilder) Switch() core.BooleanSchemaBuilder {
	return b.CaseInsensitive().
		Description("Boolean switch (accepts true/false/1/0)").
		Example(true)
}

// Enabled creates an "enabled" boolean field.
func (b *BooleanBuilder) Enabled() core.BooleanSchemaBuilder {
	return b.Default(false).
		Description("Whether this feature is enabled").
		Example(true)
}

// Active creates an "active" boolean field.
func (b *BooleanBuilder) Active() core.BooleanSchemaBuilder {
	return b.Default(true).
		Description("Whether this item is active").
		Example(true)
}

// clone creates a deep copy of the builder to ensure immutability.
func (b *BooleanBuilder) clone() *BooleanBuilder {
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

	return &BooleanBuilder{config: newConfig}
}
