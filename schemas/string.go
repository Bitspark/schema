package schemas

import (
	"regexp"

	"defs.dev/schema/core"
)

// StringSchemaConfig holds the configuration for building a StringSchema.
type StringSchemaConfig struct {
	Metadata    core.SchemaMetadata
	MinLength   *int
	MaxLength   *int
	Pattern     *regexp.Regexp
	Format      string
	EnumValues  []string
	DefaultVal  *string
	Annotations []core.Annotation
}

// StringSchema is a clean, API-first implementation of string schema validation.
// It implements core.StringSchema interface and provides immutable operations.
type StringSchema struct {
	config StringSchemaConfig
}

// Ensure StringSchema implements the API interfaces at compile time
var _ core.Schema = (*StringSchema)(nil)
var _ core.StringSchema = (*StringSchema)(nil)
var _ core.Accepter = (*StringSchema)(nil)

// NewStringSchema creates a new StringSchema with the given configuration.
func NewStringSchema(config StringSchemaConfig) *StringSchema {
	return &StringSchema{config: config}
}

// Type returns the schema type constant.
func (s *StringSchema) Type() core.SchemaType {
	return core.TypeString
}

// Metadata returns the schema metadata.
func (s *StringSchema) Metadata() core.SchemaMetadata {
	return s.config.Metadata
}

// Clone returns a deep copy of the StringSchema.
func (s *StringSchema) Clone() core.Schema {
	newConfig := s.config

	// Deep copy enumValues
	if s.config.EnumValues != nil {
		newConfig.EnumValues = make([]string, len(s.config.EnumValues))
		copy(newConfig.EnumValues, s.config.EnumValues)
	}

	// Deep copy metadata examples and tags
	if s.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(s.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, s.config.Metadata.Examples)
	}

	if s.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(s.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, s.config.Metadata.Tags)
	}

	// Deep copy annotations
	if s.config.Annotations != nil {
		newConfig.Annotations = make([]core.Annotation, len(s.config.Annotations))
		copy(newConfig.Annotations, s.config.Annotations)
	}

	return NewStringSchema(newConfig)
}

// MinLength returns the minimum length constraint.
func (s *StringSchema) MinLength() *int {
	return s.config.MinLength
}

// MaxLength returns the maximum length constraint.
func (s *StringSchema) MaxLength() *int {
	return s.config.MaxLength
}

// Pattern returns the regex pattern as a string.
func (s *StringSchema) Pattern() string {
	if s.config.Pattern == nil {
		return ""
	}
	return s.config.Pattern.String()
}

// Format returns the format constraint.
func (s *StringSchema) Format() string {
	return s.config.Format
}

// EnumValues returns a copy of the enum values.
func (s *StringSchema) EnumValues() []string {
	if s.config.EnumValues == nil {
		return nil
	}
	result := make([]string, len(s.config.EnumValues))
	copy(result, s.config.EnumValues)
	return result
}

// DefaultValue returns the default value.
func (s *StringSchema) DefaultValue() *string {
	return s.config.DefaultVal
}

// Annotations returns the annotations for this schema.
func (s *StringSchema) Annotations() []core.Annotation {
	if s.config.Annotations == nil {
		return nil
	}
	result := make([]core.Annotation, len(s.config.Annotations))
	copy(result, s.config.Annotations)
	return result
}

// Note: Validation moved to consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) instead.

// GenerateExample generates an example value for the schema.
func (s *StringSchema) GenerateExample() any {
	// Use provided examples first
	if len(s.config.Metadata.Examples) > 0 {
		return s.config.Metadata.Examples[0]
	}

	// Use enum values if available
	if len(s.config.EnumValues) > 0 {
		return s.config.EnumValues[0]
	}

	// Use default value if available
	if s.config.DefaultVal != nil {
		return *s.config.DefaultVal
	}

	// Generate based on format
	if s.config.Format != "" {
		return s.generateFormatExampleLegacy(s.config.Format)
	}

	// Generate based on constraints
	if s.config.MinLength != nil {
		minLen := *s.config.MinLength
		if minLen > 0 {
			return s.generateStringOfLength(minLen)
		}
	}

	// Default example
	return "string"
}

// Accept implements the visitor pattern.
func (s *StringSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitString(s)
}

func (s *StringSchema) generateStringOfLength(length int) string {
	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

func (s *StringSchema) generateFormatExampleLegacy(format string) string {
	switch format {
	case "email":
		return "user@example.com"
	case "url":
		return "https://example.com"
	case "uuid":
		return "550e8400-e29b-41d4-a716-446655440000"
	case "date":
		return "2023-01-01"
	case "time":
		return "12:00:00"
	case "date-time":
		return "2023-01-01T12:00:00Z"
	default:
		return "string"
	}
}

func (s *StringSchema) isValidEmailLegacy(email string) bool {
	return len(email) > 0 && s.containsChar(email, '@') && s.containsChar(email, '.')
}

func (s *StringSchema) isValidUUIDLegacy(uuid string) bool {
	return len(uuid) == 36 && s.containsChar(uuid, '-')
}

func (s *StringSchema) isValidURLLegacy(url string) bool {
	return len(url) > 0 && (s.startsWith(url, "http://") || s.startsWith(url, "https://"))
}

func (s *StringSchema) containsChar(str string, c rune) bool {
	for _, char := range str {
		if char == c {
			return true
		}
	}
	return false
}

func (s *StringSchema) startsWith(str, prefix string) bool {
	return len(str) >= len(prefix) && str[:len(prefix)] == prefix
}
