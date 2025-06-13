package schemas

import (
	"defs.dev/schema/core"
)

// IntegerSchemaConfig holds the configuration for building an IntegerSchema.
type IntegerSchemaConfig struct {
	Metadata    core.SchemaMetadata
	Annotations []core.Annotation
	Minimum     *int64
	Maximum     *int64
	DefaultVal  *int64
}

// IntegerSchema is a clean, API-first implementation of integer schema validation.
// It implements core.IntegerSchema interface and provides immutable operations.
type IntegerSchema struct {
	config IntegerSchemaConfig
}

func (i *IntegerSchema) Annotations() []core.Annotation {
	if i.config.Annotations == nil {
		return nil
	}
	result := make([]core.Annotation, len(i.config.Annotations))
	copy(result, i.config.Annotations)
	return result
}

// Ensure IntegerSchema implements the API interfaces at compile time
var _ core.Schema = (*IntegerSchema)(nil)
var _ core.IntegerSchema = (*IntegerSchema)(nil)
var _ core.Accepter = (*IntegerSchema)(nil)

// NewIntegerSchema creates a new IntegerSchema with the given configuration.
func NewIntegerSchema(config IntegerSchemaConfig) *IntegerSchema {
	return &IntegerSchema{config: config}
}

// Type returns the schema type constant.
func (i *IntegerSchema) Type() core.SchemaType {
	return core.TypeInteger
}

// Metadata returns the schema metadata.
func (i *IntegerSchema) Metadata() core.SchemaMetadata {
	return i.config.Metadata
}

// Clone returns a deep copy of the IntegerSchema.
func (i *IntegerSchema) Clone() core.Schema {
	newConfig := i.config

	// Deep copy metadata examples and tags
	if i.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(i.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, i.config.Metadata.Examples)
	}

	if i.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(i.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, i.config.Metadata.Tags)
	}

	return NewIntegerSchema(newConfig)
}

// Minimum returns the minimum value constraint.
func (i *IntegerSchema) Minimum() *int64 {
	return i.config.Minimum
}

// Maximum returns the maximum value constraint.
func (i *IntegerSchema) Maximum() *int64 {
	return i.config.Maximum
}

// DefaultValue returns the default value.
func (i *IntegerSchema) DefaultValue() *int64 {
	return i.config.DefaultVal
}

// Note: Validation moved to consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) instead.

// GenerateExample generates an example value for the integer schema.
func (i *IntegerSchema) GenerateExample() any {
	// Use provided examples if available
	if len(i.config.Metadata.Examples) > 0 {
		return i.config.Metadata.Examples[0]
	}

	// Use default value if set
	if i.config.DefaultVal != nil {
		return *i.config.DefaultVal
	}

	// Generate based on constraints
	if i.config.Minimum != nil && i.config.Maximum != nil {
		// Return midpoint between min and max
		return (*i.config.Minimum + *i.config.Maximum) / 2
	}

	if i.config.Minimum != nil {
		// Return minimum + 1 (or minimum if it's positive)
		if *i.config.Minimum >= 0 {
			return *i.config.Minimum + 1
		}
		return *i.config.Minimum
	}

	if i.config.Maximum != nil {
		// Return maximum - 1 (or maximum if it's negative)
		if *i.config.Maximum <= 0 {
			return *i.config.Maximum - 1
		}
		return *i.config.Maximum
	}

	// Default example
	return int64(42)
}

// Accept implements the visitor pattern for schema traversal.
func (i *IntegerSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitInteger(i)
}
