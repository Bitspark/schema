package schemas

import (
	"fmt"
	"regexp"

	"defs.dev/schema/api"
)

// StringSchemaConfig holds the configuration for building a StringSchema.
type StringSchemaConfig struct {
	Metadata   api.SchemaMetadata
	MinLength  *int
	MaxLength  *int
	Pattern    *regexp.Regexp
	Format     string
	EnumValues []string
	DefaultVal *string
}

// StringSchema is a clean, API-first implementation of string schema validation.
// It implements api.StringSchema interface and provides immutable operations.
type StringSchema struct {
	config StringSchemaConfig
}

// Ensure StringSchema implements the API interfaces at compile time
var _ api.Schema = (*StringSchema)(nil)
var _ api.StringSchema = (*StringSchema)(nil)
var _ api.Accepter = (*StringSchema)(nil)

// NewStringSchema creates a new StringSchema with the given configuration.
func NewStringSchema(config StringSchemaConfig) *StringSchema {
	return &StringSchema{config: config}
}

// Type returns the schema type constant.
func (s *StringSchema) Type() api.SchemaType {
	return api.TypeString
}

// Metadata returns the schema metadata.
func (s *StringSchema) Metadata() api.SchemaMetadata {
	return s.config.Metadata
}

// Clone returns a deep copy of the StringSchema.
func (s *StringSchema) Clone() api.Schema {
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

// Validate validates a value against the string schema.
func (s *StringSchema) Validate(value any) api.ValidationResult {
	str, ok := value.(string)
	if !ok {
		return api.ValidationResult{
			Valid: false,
			Errors: []api.ValidationError{{
				Path:       "",
				Message:    "Expected string",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "string",
				Suggestion: "Provide a string value",
			}},
		}
	}

	var errors []api.ValidationError

	// Length validation
	if s.config.MinLength != nil && len(str) < *s.config.MinLength {
		errors = append(errors, api.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("String too short (minimum %d characters)", *s.config.MinLength),
			Code:       "min_length",
			Value:      str,
			Suggestion: fmt.Sprintf("Provide at least %d characters", *s.config.MinLength),
		})
	}

	if s.config.MaxLength != nil && len(str) > *s.config.MaxLength {
		errors = append(errors, api.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("String too long (maximum %d characters)", *s.config.MaxLength),
			Code:       "max_length",
			Value:      str,
			Suggestion: fmt.Sprintf("Limit to %d characters", *s.config.MaxLength),
		})
	}

	// Enum validation
	if len(s.config.EnumValues) > 0 {
		valid := false
		for _, enum := range s.config.EnumValues {
			if str == enum {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, api.ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("Value must be one of: %v", s.config.EnumValues),
				Code:       "enum_mismatch",
				Value:      str,
				Expected:   fmt.Sprintf("One of: %v", s.config.EnumValues),
				Suggestion: fmt.Sprintf("Use one of these values: %v", s.config.EnumValues),
			})
		}
	}

	// Pattern validation (optimized with pre-compiled regex)
	if s.config.Pattern != nil {
		if !s.config.Pattern.MatchString(str) {
			errors = append(errors, api.ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("String does not match pattern: %s", s.config.Pattern.String()),
				Code:       "pattern_mismatch",
				Value:      str,
				Expected:   fmt.Sprintf("Pattern: %s", s.config.Pattern.String()),
				Suggestion: "Provide a string that matches the required pattern",
			})
		}
	}

	// Format validation
	if s.config.Format != "" {
		if err := validateFormat(str, s.config.Format); err != nil {
			errors = append(errors, api.ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("Invalid %s format", s.config.Format),
				Code:       "format_invalid",
				Value:      str,
				Expected:   fmt.Sprintf("Valid %s", s.config.Format),
				Suggestion: getFormatSuggestion(s.config.Format),
			})
		}
	}

	return api.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ToJSONSchema converts the schema to JSON Schema format.
func (s *StringSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "string",
	}

	if s.config.MinLength != nil {
		schema["minLength"] = *s.config.MinLength
	}
	if s.config.MaxLength != nil {
		schema["maxLength"] = *s.config.MaxLength
	}
	if s.config.Pattern != nil {
		schema["pattern"] = s.config.Pattern.String()
	}
	if s.config.Format != "" {
		schema["format"] = s.config.Format
	}
	if len(s.config.EnumValues) > 0 {
		schema["enum"] = s.config.EnumValues
	}
	if s.config.Metadata.Description != "" {
		schema["description"] = s.config.Metadata.Description
	}
	if len(s.config.Metadata.Examples) > 0 {
		schema["examples"] = s.config.Metadata.Examples
	}

	return schema
}

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
		return generateFormatExample(s.config.Format)
	}

	// Generate based on constraints
	if s.config.MinLength != nil {
		minLen := *s.config.MinLength
		if minLen > 0 {
			return generateStringOfLength(minLen)
		}
	}

	// Default example
	return "string"
}

// Accept implements the visitor pattern.
func (s *StringSchema) Accept(visitor api.SchemaVisitor) error {
	return visitor.VisitString(s)
}

// Helper functions (these would be moved to validation package in full implementation)

func validateFormat(value, format string) error {
	// Simplified format validation - in full implementation this would be comprehensive
	switch format {
	case "email":
		if !isValidEmail(value) {
			return fmt.Errorf("invalid email format")
		}
	case "uuid":
		if !isValidUUID(value) {
			return fmt.Errorf("invalid UUID format")
		}
	case "url":
		if !isValidURL(value) {
			return fmt.Errorf("invalid URL format")
		}
	}
	return nil
}

func getFormatSuggestion(format string) string {
	switch format {
	case "email":
		return "Provide a valid email address (e.g., user@example.com)"
	case "uuid":
		return "Provide a valid UUID (e.g., 123e4567-e89b-12d3-a456-426614174000)"
	case "url":
		return "Provide a valid URL (e.g., https://example.com)"
	default:
		return fmt.Sprintf("Provide a valid %s", format)
	}
}

func generateFormatExample(format string) string {
	switch format {
	case "email":
		return "user@example.com"
	case "uuid":
		return "123e4567-e89b-12d3-a456-426614174000"
	case "url":
		return "https://example.com"
	default:
		return "string"
	}
}

func generateStringOfLength(length int) string {
	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

// Simplified validation functions - in full implementation these would be robust
func isValidEmail(email string) bool {
	// Simplified email validation
	return len(email) > 0 && containsChar(email, '@') && containsChar(email, '.')
}

func isValidUUID(uuid string) bool {
	// Simplified UUID validation
	return len(uuid) == 36 && containsChar(uuid, '-')
}

func isValidURL(url string) bool {
	// Simplified URL validation
	return len(url) > 0 && (startsWith(url, "http://") || startsWith(url, "https://"))
}

func containsChar(s string, c rune) bool {
	for _, char := range s {
		if char == c {
			return true
		}
	}
	return false
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
