package schemas

import (
	"fmt"
	"regexp"

	"defs.dev/schema/api/core"
)

// ValidatorRegistry interface to avoid import cycle
type ValidatorRegistry interface {
	ValidateWithAnnotations(value any, annotations []core.Annotation) ValidationResult
}

// ValidationResult represents the result of validator-based validation.
type ValidationResult struct {
	Valid             bool                   `json:"valid"`
	Errors            []ValidationError      `json:"errors,omitempty"`
	Warnings          []ValidationWarning    `json:"warnings,omitempty"`
	Suggestions       []ValidationSuggestion `json:"suggestions,omitempty"`
	Metadata          map[string]any         `json:"metadata,omitempty"`
	AppliedValidators []string               `json:"applied_validators,omitempty"`
}

// ValidationError represents a specific validation error.
type ValidationError struct {
	ValidatorName string `json:"validator_name"`
	Path          string `json:"path,omitempty"`
	Message       string `json:"message"`
	Code          string `json:"code"`
	Value         any    `json:"value,omitempty"`
	Expected      string `json:"expected,omitempty"`
	Context       string `json:"context,omitempty"`
	Suggestion    string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a non-fatal validation issue.
type ValidationWarning struct {
	ValidatorName string `json:"validator_name"`
	Path          string `json:"path,omitempty"`
	Message       string `json:"message"`
	Code          string `json:"code"`
	Suggestion    string `json:"suggestion,omitempty"`
}

// ValidationSuggestion provides actionable suggestions for fixing validation issues.
type ValidationSuggestion struct {
	ValidatorName string         `json:"validator_name"`
	Message       string         `json:"message"`
	Action        string         `json:"action"`
	Parameters    map[string]any `json:"parameters,omitempty"`
}

// StringSchemaConfig holds the configuration for building a StringSchema.
type StringSchemaConfig struct {
	Metadata          core.SchemaMetadata
	MinLength         *int
	MaxLength         *int
	Pattern           *regexp.Regexp
	Format            string
	EnumValues        []string
	DefaultVal        *string
	Annotations       []core.Annotation
	ValidatorRegistry ValidatorRegistry
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

// Validate validates a value against the string schema using annotation-based validation.
func (s *StringSchema) Validate(value any) core.ValidationResult {
	str, ok := value.(string)
	if !ok {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{{
				Path:       "",
				Message:    "Expected string",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "string",
				Suggestion: "Provide a string value",
			}},
		}
	}

	var errors []core.ValidationError

	// Legacy constraint validation (for backward compatibility during transition)
	if s.config.MinLength != nil && len(str) < *s.config.MinLength {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("String too short (minimum %d characters)", *s.config.MinLength),
			Code:       "min_length",
			Value:      str,
			Suggestion: fmt.Sprintf("Provide at least %d characters", *s.config.MinLength),
		})
	}

	if s.config.MaxLength != nil && len(str) > *s.config.MaxLength {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("String too long (maximum %d characters)", *s.config.MaxLength),
			Code:       "max_length",
			Value:      str,
			Suggestion: fmt.Sprintf("Limit to %d characters", *s.config.MaxLength),
		})
	}

	if len(s.config.EnumValues) > 0 {
		valid := false
		for _, enum := range s.config.EnumValues {
			if str == enum {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, core.ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("Value must be one of: %v", s.config.EnumValues),
				Code:       "enum_mismatch",
				Value:      str,
				Expected:   fmt.Sprintf("One of: %v", s.config.EnumValues),
				Suggestion: fmt.Sprintf("Use one of these values: %v", s.config.EnumValues),
			})
		}
	}

	if s.config.Pattern != nil {
		if !s.config.Pattern.MatchString(str) {
			errors = append(errors, core.ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("String does not match pattern: %s", s.config.Pattern.String()),
				Code:       "pattern_mismatch",
				Value:      str,
				Expected:   fmt.Sprintf("Pattern: %s", s.config.Pattern.String()),
				Suggestion: "Provide a string that matches the required pattern",
			})
		}
	}

	// NEW: Annotation-based validation using the registry system
	if s.config.ValidatorRegistry != nil && len(s.config.Annotations) > 0 {
		validationResult := s.config.ValidatorRegistry.ValidateWithAnnotations(str, s.config.Annotations)

		// Convert registry validation errors to core validation errors
		for _, regError := range validationResult.Errors {
			errors = append(errors, core.ValidationError{
				Path:       regError.Path,
				Message:    regError.Message,
				Code:       regError.Code,
				Value:      regError.Value,
				Expected:   regError.Expected,
				Suggestion: regError.Suggestion,
				Context:    regError.Context,
			})
		}
	} else if s.config.Format != "" {
		// Fallback to legacy format validation only if no annotation-based validation
		if err := s.validateFormatLegacy(str, s.config.Format); err != nil {
			errors = append(errors, core.ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("Invalid %s format", s.config.Format),
				Code:       "format_invalid",
				Value:      str,
				Expected:   fmt.Sprintf("Valid %s", s.config.Format),
				Suggestion: s.getFormatSuggestionLegacy(s.config.Format),
			})
		}
	}

	return core.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
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

// Legacy validation functions (kept temporarily for fallback)
func (s *StringSchema) validateFormatLegacy(value, format string) error {
	switch format {
	case "email":
		if !s.isValidEmailLegacy(value) {
			return fmt.Errorf("invalid email format")
		}
	case "uuid":
		if !s.isValidUUIDLegacy(value) {
			return fmt.Errorf("invalid UUID format")
		}
	case "url":
		if !s.isValidURLLegacy(value) {
			return fmt.Errorf("invalid URL format")
		}
	}
	return nil
}

func (s *StringSchema) getFormatSuggestionLegacy(format string) string {
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

func (s *StringSchema) generateFormatExampleLegacy(format string) string {
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
