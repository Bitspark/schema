package builders

import (
	"defs.dev/schema/api"
	"defs.dev/schema/schemas"
)

// ArrayBuilder provides a fluent interface for building ArraySchema instances.
// It implements api.ArraySchemaBuilder interface and returns api.ArraySchema.
type ArrayBuilder struct {
	config schemas.ArraySchemaConfig
}

// Ensure ArrayBuilder implements the API interface at compile time
var _ api.ArraySchemaBuilder = (*ArrayBuilder)(nil)

// NewArraySchema creates a new ArrayBuilder for creating array schemas.
func NewArraySchema() api.ArraySchemaBuilder {
	return &ArrayBuilder{
		config: schemas.ArraySchemaConfig{
			Metadata: api.SchemaMetadata{},
		},
	}
}

// Build returns the constructed ArraySchema as an api.ArraySchema.
func (b *ArrayBuilder) Build() api.ArraySchema {
	return schemas.NewArraySchema(b.config)
}

// Description sets the description metadata.
func (b *ArrayBuilder) Description(desc string) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Description = desc
	return clone
}

// Name sets the name metadata.
func (b *ArrayBuilder) Name(name string) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Name = name
	return clone
}

// Tag adds a tag to the metadata.
func (b *ArrayBuilder) Tag(tag string) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Tags = append(clone.config.Metadata.Tags, tag)
	return clone
}

// Items sets the schema for array items.
func (b *ArrayBuilder) Items(itemSchema api.Schema) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.ItemSchema = itemSchema
	return clone
}

// MinItems sets the minimum number of items constraint.
func (b *ArrayBuilder) MinItems(min int) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.MinItems = &min
	return clone
}

// MaxItems sets the maximum number of items constraint.
func (b *ArrayBuilder) MaxItems(max int) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.MaxItems = &max
	return clone
}

// UniqueItems requires that all items in the array be unique.
func (b *ArrayBuilder) UniqueItems() api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.UniqueItems = true
	return clone
}

// Example adds an example value to the metadata.
func (b *ArrayBuilder) Example(example []any) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.Metadata.Examples = append(clone.config.Metadata.Examples, example)
	return clone
}

// Default sets the default value.
func (b *ArrayBuilder) Default(value []any) api.ArraySchemaBuilder {
	clone := b.clone()
	// Deep copy the default value
	if value != nil {
		clone.config.DefaultVal = make([]any, len(value))
		copy(clone.config.DefaultVal, value)
	} else {
		clone.config.DefaultVal = nil
	}
	return clone
}

// Contains sets a schema that at least one item in the array must match.
func (b *ArrayBuilder) Contains(schema api.Schema) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.ContainsSchema = schema
	return clone
}

// Length sets both minimum and maximum items to the same value (fixed length).
func (b *ArrayBuilder) Length(length int) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.MinItems = &length
	clone.config.MaxItems = &length
	return clone
}

// Range sets both minimum and maximum items constraints.
func (b *ArrayBuilder) Range(min, max int) api.ArraySchemaBuilder {
	clone := b.clone()
	clone.config.MinItems = &min
	clone.config.MaxItems = &max
	return clone
}

// NonEmpty ensures the array has at least one item.
func (b *ArrayBuilder) NonEmpty() api.ArraySchemaBuilder {
	return b.MinItems(1).
		Description("Non-empty array")
}

// Common array type helpers

// StringArray creates an array of strings.
func (b *ArrayBuilder) StringArray() api.ArraySchemaBuilder {
	// We need to create a string schema, but we need to import core to avoid circular imports
	// For now, we'll just set description and let the user set Items manually
	return b.Description("Array of strings").
		Example([]any{"item1", "item2", "item3"})
}

// NumberArray creates an array of numbers.
func (b *ArrayBuilder) NumberArray() api.ArraySchemaBuilder {
	return b.Description("Array of numbers").
		Example([]any{1.0, 2.0, 3.0})
}

// IntegerArray creates an array of integers.
func (b *ArrayBuilder) IntegerArray() api.ArraySchemaBuilder {
	return b.Description("Array of integers").
		Example([]any{int64(1), int64(2), int64(3)})
}

// BooleanArray creates an array of booleans.
func (b *ArrayBuilder) BooleanArray() api.ArraySchemaBuilder {
	return b.Description("Array of booleans").
		Example([]any{true, false, true})
}

// List creates a simple list with basic configuration.
func (b *ArrayBuilder) List() api.ArraySchemaBuilder {
	return b.Description("List of items").
		Example([]any{"item1", "item2", "item3"})
}

// Set creates a unique items array (set-like).
func (b *ArrayBuilder) Set() api.ArraySchemaBuilder {
	return b.UniqueItems().
		Description("Set of unique items").
		Example([]any{"unique1", "unique2", "unique3"})
}

// Tuple creates a fixed-length array.
func (b *ArrayBuilder) Tuple(length int) api.ArraySchemaBuilder {
	return b.Length(length).
		Description("Fixed-length tuple")
}

// LimitedList creates a list with reasonable size constraints.
func (b *ArrayBuilder) LimitedList(maxItems int) api.ArraySchemaBuilder {
	return b.Range(0, maxItems).
		Description("Limited size list")
}

// clone creates a deep copy of the builder to ensure immutability.
func (b *ArrayBuilder) clone() *ArrayBuilder {
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

	if b.config.DefaultVal != nil {
		newConfig.DefaultVal = make([]any, len(b.config.DefaultVal))
		copy(newConfig.DefaultVal, b.config.DefaultVal)
	}

	// Note: ItemSchema and ContainsSchema are not deeply cloned as they should be immutable
	return &ArrayBuilder{config: newConfig}
}
