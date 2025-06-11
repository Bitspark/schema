package builders

import (
	"defs.dev/schema/api"
	"defs.dev/schema/schemas"
)

// BooleanBuilder provides a fluent interface for building BooleanSchema instances.
// It implements api.BooleanSchemaBuilder interface and returns api.BooleanSchema.
type BooleanBuilder struct {
	config schemas.BooleanSchemaConfig
}

// Ensure BooleanBuilder implements the API interface at compile time
var _ api.BooleanSchemaBuilder = (*BooleanBuilder)(nil)

// NewBooleanSchema creates a new BooleanBuilder for creating boolean schemas.
func NewBooleanSchema() api.BooleanSchemaBuilder {
	return &BooleanBuilder{
		config: schemas.BooleanSchemaConfig{
			Metadata:        api.SchemaMetadata{},
			AllowStringConv: false,
			CaseInsensitive: false,
		},
	}
}

// Build returns the constructed BooleanSchema as an api.BooleanSchema.
func (b *BooleanBuilder) Build() api.BooleanSchema {
	return schemas.NewBooleanSchema(b.config)
}

// Description sets the description metadata.
func (b *BooleanBuilder) Description(desc string) api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *BooleanBuilder) Name(name string) api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *BooleanBuilder) Tag(tag string) api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Example adds an example value to the metadata.
func (b *BooleanBuilder) Example(example bool) api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *BooleanBuilder) Default(value bool) api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.DefaultVal = &value
	return clone
}

// AllowStringConversion enables conversion from string values ("true", "false", "1", "0").
func (b *BooleanBuilder) AllowStringConversion() api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.AllowStringConv = true
	return clone
}

// CaseInsensitive enables case-insensitive string conversion.
// This automatically enables string conversion if not already enabled.
func (b *BooleanBuilder) CaseInsensitive() api.BooleanSchemaBuilder {
	clone := b.clone()
	clone.config.AllowStringConv = true
	clone.config.CaseInsensitive = true
	return clone
}

// Common boolean type helpers

// Required creates a required boolean field.
func (b *BooleanBuilder) Required() api.BooleanSchemaBuilder {
	return b.Description("Required boolean value").
		Example(true)
}

// Flag creates a flag-style boolean (defaults to false).
func (b *BooleanBuilder) Flag() api.BooleanSchemaBuilder {
	return b.Default(false).
		Description("Boolean flag").
		Example(false)
}

// Switch creates a switch-style boolean with string conversion.
func (b *BooleanBuilder) Switch() api.BooleanSchemaBuilder {
	return b.CaseInsensitive().
		Description("Boolean switch (accepts true/false/1/0)").
		Example(true)
}

// Enabled creates an "enabled" boolean field.
func (b *BooleanBuilder) Enabled() api.BooleanSchemaBuilder {
	return b.Default(false).
		Description("Whether this feature is enabled").
		Example(true)
}

// Active creates an "active" boolean field.
func (b *BooleanBuilder) Active() api.BooleanSchemaBuilder {
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
