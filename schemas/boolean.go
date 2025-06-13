package schemas

import (
	"strconv"
	"strings"

	"defs.dev/schema/core"
)

// BooleanSchemaConfig holds the configuration for building a BooleanSchema.
type BooleanSchemaConfig struct {
	Metadata        core.SchemaMetadata
	Annotations     []core.Annotation
	DefaultVal      *bool
	CaseInsensitive bool // Case-insensitive string conversion
}

// BooleanSchema is a clean, API-first implementation of boolean schema validation.
// It implements core.BooleanSchema interface and provides immutable operations.
type BooleanSchema struct {
	config BooleanSchemaConfig
}

// Ensure BooleanSchema implements the API interfaces at compile time
var _ core.Schema = (*BooleanSchema)(nil)
var _ core.BooleanSchema = (*BooleanSchema)(nil)
var _ core.Accepter = (*BooleanSchema)(nil)

// NewBooleanSchema creates a new BooleanSchema with the given configuration.
func NewBooleanSchema(config BooleanSchemaConfig) *BooleanSchema {
	return &BooleanSchema{config: config}
}

// Type returns the schema type constant.
func (b *BooleanSchema) Type() core.SchemaType {
	return core.TypeBoolean
}

// Metadata returns the schema metadata.
func (b *BooleanSchema) Metadata() core.SchemaMetadata {
	return b.config.Metadata
}

// Annotations returns the annotations of the schema.
func (b *BooleanSchema) Annotations() []core.Annotation {
	if b.config.Annotations == nil {
		return nil
	}
	result := make([]core.Annotation, len(b.config.Annotations))
	copy(result, b.config.Annotations)
	return result
}

// Clone returns a deep copy of the BooleanSchema.
func (b *BooleanSchema) Clone() core.Schema {
	newConfig := b.config

	// Deep copy metadata examples and tags
	if b.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(b.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, b.config.Metadata.Examples)
	}

	if b.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(b.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, b.config.Metadata.Tags)
	}

	return NewBooleanSchema(newConfig)
}

// DefaultValue returns the default value.
func (b *BooleanSchema) DefaultValue() *bool {
	return b.config.DefaultVal
}

// CaseInsensitive returns whether string conversion is case-insensitive.
func (b *BooleanSchema) CaseInsensitive() bool {
	return b.config.CaseInsensitive
}

// Note: Validation moved to consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) instead.

// convertStringToBool converts a string to boolean using various formats.
func (b *BooleanSchema) convertStringToBool(str string) (bool, error) {
	// Handle case insensitivity
	convertStr := str
	if b.config.CaseInsensitive {
		convertStr = strings.ToLower(str)
	}

	// Standard boolean strings
	switch convertStr {
	case "true", "1", "yes", "on", "y", "t":
		return true, nil
	case "false", "0", "no", "off", "n", "f":
		return false, nil
	}

	// Try Go's standard strconv.ParseBool as fallback
	return strconv.ParseBool(convertStr)
}

// GenerateExample generates an example value for the boolean schema.
func (b *BooleanSchema) GenerateExample() any {
	// Use provided examples if available
	if len(b.config.Metadata.Examples) > 0 {
		return b.config.Metadata.Examples[0]
	}

	// Use default value if set
	if b.config.DefaultVal != nil {
		return *b.config.DefaultVal
	}

	// Default example
	return true
}

// Accept implements the visitor pattern for schema traversal.
func (b *BooleanSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitBoolean(b)
}
