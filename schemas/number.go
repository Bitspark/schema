package schemas

import (
	"defs.dev/schema/core"
)

// NumberSchemaConfig holds the configuration for building a NumberSchema.
type NumberSchemaConfig struct {
	Metadata    core.SchemaMetadata
	Annotations []core.Annotation
	Minimum     *float64
	Maximum     *float64
	DefaultVal  *float64
}

// NumberSchema is a clean, API-first implementation of number schema validation.
// It implements core.NumberSchema interface and provides immutable operations.
type NumberSchema struct {
	config NumberSchemaConfig
}

// Ensure NumberSchema implements the API interfaces at compile time
var _ core.Schema = (*NumberSchema)(nil)
var _ core.NumberSchema = (*NumberSchema)(nil)
var _ core.Accepter = (*NumberSchema)(nil)

// NewNumberSchema creates a new NumberSchema with the given configuration.
func NewNumberSchema(config NumberSchemaConfig) *NumberSchema {
	return &NumberSchema{config: config}
}

// Type returns the schema type constant.
func (n *NumberSchema) Type() core.SchemaType {
	return core.TypeNumber
}

// Metadata returns the schema metadata.
func (n *NumberSchema) Metadata() core.SchemaMetadata {
	return n.config.Metadata
}

// Annotations returns the annotations of the schema.
func (n *NumberSchema) Annotations() []core.Annotation {
	if n.config.Annotations == nil {
		return nil
	}
	result := make([]core.Annotation, len(n.config.Annotations))
	copy(result, n.config.Annotations)
	return result
}

// Clone returns a deep copy of the NumberSchema.
func (n *NumberSchema) Clone() core.Schema {
	newConfig := n.config

	// Deep copy metadata examples and tags
	if n.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(n.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, n.config.Metadata.Examples)
	}

	if n.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(n.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, n.config.Metadata.Tags)
	}

	return NewNumberSchema(newConfig)
}

// Minimum returns the minimum value constraint.
func (n *NumberSchema) Minimum() *float64 {
	return n.config.Minimum
}

// Maximum returns the maximum value constraint.
func (n *NumberSchema) Maximum() *float64 {
	return n.config.Maximum
}

// DefaultValue returns the default value.
func (n *NumberSchema) DefaultValue() *float64 {
	return n.config.DefaultVal
}

// Note: Validation moved to consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) instead.

// GenerateExample generates an example value for the number schema.
func (n *NumberSchema) GenerateExample() any {
	// Use provided examples if available
	if len(n.config.Metadata.Examples) > 0 {
		return n.config.Metadata.Examples[0]
	}

	// Use default value if set
	if n.config.DefaultVal != nil {
		return *n.config.DefaultVal
	}

	// Generate based on constraints
	if n.config.Minimum != nil && n.config.Maximum != nil {
		// Return midpoint between min and max
		return (*n.config.Minimum + *n.config.Maximum) / 2
	}

	if n.config.Minimum != nil {
		// Return minimum + 1 (or minimum if it's positive)
		if *n.config.Minimum >= 0 {
			return *n.config.Minimum + 1
		}
		return *n.config.Minimum
	}

	if n.config.Maximum != nil {
		// Return maximum - 1 (or maximum if it's negative)
		if *n.config.Maximum <= 0 {
			return *n.config.Maximum - 1
		}
		return *n.config.Maximum
	}

	// Default example
	return 42.0
}

// Accept implements the visitor pattern for schema traversal.
func (n *NumberSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitNumber(n)
}
