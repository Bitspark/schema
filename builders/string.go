package builders

import (
	"regexp"

	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// StringBuilder provides a fluent interface for building StringSchema instances.
// It implements core.StringSchemaBuilder interface and returns core.StringSchema.
type StringBuilder struct {
	config schemas.StringSchemaConfig
}

// Ensure StringBuilder implements the API interface at compile time
var _ core.StringSchemaBuilder = (*StringBuilder)(nil)

// NewStringSchema creates a new StringBuilder for creating string schemas.
func NewStringSchema() core.StringSchemaBuilder {
	return &StringBuilder{
		config: schemas.StringSchemaConfig{
			Metadata: core.SchemaMetadata{},
		},
	}
}

// Build returns the constructed StringSchema as an core.StringSchema.
func (b *StringBuilder) Build() core.StringSchema {
	return schemas.NewStringSchema(b.config)
}

// Description sets the description metadata.
func (b *StringBuilder) Description(desc string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *StringBuilder) Name(name string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *StringBuilder) Tag(tag string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// MinLength sets the minimum length constraint.
func (b *StringBuilder) MinLength(min int) core.StringSchemaBuilder {
	clone := b.clone()
	minVal := min
	clone.config.MinLength = &minVal
	return clone
}

// MaxLength sets the maximum length constraint.
func (b *StringBuilder) MaxLength(max int) core.StringSchemaBuilder {
	clone := b.clone()
	maxVal := max
	clone.config.MaxLength = &maxVal
	return clone
}

// Pattern sets the regex pattern constraint.
// The pattern is pre-compiled for performance during validation.
func (b *StringBuilder) Pattern(pattern string) core.StringSchemaBuilder {
	clone := b.clone()
	if pattern != "" {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			// In a production implementation, we might want to handle this more gracefully
			panic("invalid regex pattern: " + err.Error())
		}
		clone.config.Pattern = compiled
	}
	return clone
}

// Format sets the format constraint (e.g., "email", "uuid", "url").
func (b *StringBuilder) Format(format string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.Format = format
	return clone
}

// Enum sets the allowed values constraint.
func (b *StringBuilder) Enum(values ...string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.EnumValues = make([]string, len(values))
	copy(clone.config.EnumValues, values)
	return clone
}

// Default sets the default value.
func (b *StringBuilder) Default(value string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.DefaultVal = &value
	return clone
}

// Example adds an example value to the metadata.
func (b *StringBuilder) Example(example string) core.StringSchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Common format helpers

// Email sets the format to email and adds helpful metadata.
func (b *StringBuilder) Email() core.StringSchemaBuilder {
	return b.Format("email").
		Description("Valid email address").
		Example("user@example.com")
}

// UUID sets the format to UUID and adds helpful metadata.
func (b *StringBuilder) UUID() core.StringSchemaBuilder {
	return b.Format("uuid").
		Description("UUID identifier").
		Example("123e4567-e89b-12d3-a456-426614174000")
}

// URL sets the format to URL and adds helpful metadata.
func (b *StringBuilder) URL() core.StringSchemaBuilder {
	return b.Format("url").
		Description("Valid URL").
		Example("https://example.com")
}

// clone creates a deep copy of the builder to ensure immutability.
func (b *StringBuilder) clone() *StringBuilder {
	newConfig := b.config

	// Deep copy slices
	if b.config.EnumValues != nil {
		newConfig.EnumValues = make([]string, len(b.config.EnumValues))
		copy(newConfig.EnumValues, b.config.EnumValues)
	}

	if b.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(b.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, b.config.Metadata.Examples)
	}

	if b.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(b.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, b.config.Metadata.Tags)
	}

	return &StringBuilder{config: newConfig}
}
