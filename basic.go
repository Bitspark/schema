package schema

import (
	"fmt"
	"regexp"
)

// String schema
type StringSchema struct {
	metadata   SchemaMetadata
	minLength  *int
	maxLength  *int
	pattern    string
	format     string
	enumValues []string
	defaultVal *string
}

// Getter methods for StringSchema properties
func (s *StringSchema) MinLength() *int {
	return s.minLength
}

func (s *StringSchema) MaxLength() *int {
	return s.maxLength
}

func (s *StringSchema) Pattern() string {
	return s.pattern
}

func (s *StringSchema) Format() string {
	return s.format
}

func (s *StringSchema) EnumValues() []string {
	if s.enumValues == nil {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]string, len(s.enumValues))
	copy(result, s.enumValues)
	return result
}

func (s *StringSchema) DefaultValue() *string {
	return s.defaultVal
}

func (s *StringSchema) Type() SchemaType {
	return TypeString
}

func (s *StringSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *StringSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *StringSchema) Clone() Schema {
	clone := *s
	return &clone
}

func (s *StringSchema) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Path:       "",
				Message:    "Expected string",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "string",
				Suggestion: "Provide a string value",
			}},
		}
	}

	var errors []ValidationError

	// Length validation
	if s.minLength != nil && len(str) < *s.minLength {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("String too short (minimum %d characters)", *s.minLength),
			Code:       "min_length",
			Value:      str,
			Suggestion: fmt.Sprintf("Provide at least %d characters", *s.minLength),
		})
	}

	if s.maxLength != nil && len(str) > *s.maxLength {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("String too long (maximum %d characters)", *s.maxLength),
			Code:       "max_length",
			Value:      str,
			Suggestion: fmt.Sprintf("Limit to %d characters", *s.maxLength),
		})
	}

	// Enum validation
	if len(s.enumValues) > 0 {
		valid := false
		for _, enum := range s.enumValues {
			if str == enum {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("Value must be one of: %v", s.enumValues),
				Code:       "enum_mismatch",
				Value:      str,
				Expected:   fmt.Sprintf("One of: %v", s.enumValues),
				Suggestion: fmt.Sprintf("Use one of these values: %v", s.enumValues),
			})
		}
	}

	// Pattern validation
	if s.pattern != "" {
		matched, err := regexp.MatchString(s.pattern, str)
		if err != nil {
			errors = append(errors, ValidationError{
				Path:    "",
				Message: "Invalid regex pattern in schema",
				Code:    "pattern_error",
				Value:   str,
				Context: "Schema validation",
			})
		} else if !matched {
			errors = append(errors, ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("String does not match pattern: %s", s.pattern),
				Code:       "pattern_mismatch",
				Value:      str,
				Expected:   fmt.Sprintf("Pattern: %s", s.pattern),
				Suggestion: "Provide a string that matches the required pattern",
			})
		}
	}

	// Format validation (email, uuid, etc.)
	if s.format != "" {
		if err := validateFormat(str, s.format); err != nil {
			errors = append(errors, ValidationError{
				Path:       "",
				Message:    fmt.Sprintf("Invalid %s format", s.format),
				Code:       "format_invalid",
				Value:      str,
				Expected:   fmt.Sprintf("Valid %s", s.format),
				Suggestion: getFormatSuggestion(s.format),
			})
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (s *StringSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "string",
	}

	if s.minLength != nil {
		schema["minLength"] = *s.minLength
	}
	if s.maxLength != nil {
		schema["maxLength"] = *s.maxLength
	}
	if s.pattern != "" {
		schema["pattern"] = s.pattern
	}
	if s.format != "" {
		schema["format"] = s.format
	}
	if len(s.enumValues) > 0 {
		schema["enum"] = s.enumValues
	}
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}
	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	return schema
}

func (s *StringSchema) GenerateExample() any {
	// Use provided examples first
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}

	// Use enum values if available
	if len(s.enumValues) > 0 {
		return s.enumValues[0]
	}

	// Generate based on format
	if s.format != "" {
		return generateFormatExample(s.format)
	}

	// Generate pattern-based example (simplified)
	if s.pattern != "" {
		return "example"
	}

	// Default example
	return "string"
}

// Object schema
type ObjectSchema struct {
	metadata        SchemaMetadata
	properties      map[string]Schema
	required        []string
	additionalProps bool
}

// Introspection methods for ObjectSchema
func (s *ObjectSchema) Properties() map[string]Schema {
	// Return a copy to prevent external mutation
	props := make(map[string]Schema)
	for k, v := range s.properties {
		props[k] = v
	}
	return props
}

func (s *ObjectSchema) Required() []string {
	return s.required
}

func (s *ObjectSchema) AdditionalProperties() bool {
	return s.additionalProps
}

func (s *ObjectSchema) Type() SchemaType {
	return TypeObject
}

func (s *ObjectSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *ObjectSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *ObjectSchema) Clone() Schema {
	clone := *s
	clone.properties = make(map[string]Schema)
	for k, v := range s.properties {
		clone.properties[k] = v.Clone()
	}
	clone.required = append([]string(nil), s.required...)
	return &clone
}

func (s *ObjectSchema) Validate(value any) ValidationResult {
	obj, ok := value.(map[string]any)
	if !ok {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Path:       "",
				Message:    "Expected object",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "object",
				Suggestion: "Provide an object value",
			}},
		}
	}

	var errors []ValidationError

	// Check required properties
	for _, reqProp := range s.required {
		if _, exists := obj[reqProp]; !exists {
			errors = append(errors, ValidationError{
				Path:       reqProp,
				Message:    fmt.Sprintf("Required property '%s' is missing", reqProp),
				Code:       "required_missing",
				Expected:   fmt.Sprintf("Property '%s'", reqProp),
				Suggestion: fmt.Sprintf("Add the required property '%s'", reqProp),
			})
		}
	}

	// Validate each property
	for propName, propValue := range obj {
		propSchema, exists := s.properties[propName]
		if !exists {
			if !s.additionalProps {
				errors = append(errors, ValidationError{
					Path:       propName,
					Message:    fmt.Sprintf("Additional property '%s' not allowed", propName),
					Code:       "additional_property",
					Value:      propValue,
					Suggestion: "Remove this property or allow additional properties",
				})
			}
			continue
		}

		result := propSchema.Validate(propValue)
		if !result.Valid {
			for _, err := range result.Errors {
				err.Path = propName + "." + err.Path
				errors = append(errors, err)
			}
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (s *ObjectSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "object",
	}

	if len(s.properties) > 0 {
		props := make(map[string]any)
		for name, propSchema := range s.properties {
			props[name] = propSchema.ToJSONSchema()
		}
		schema["properties"] = props
	}

	if len(s.required) > 0 {
		schema["required"] = s.required
	}

	schema["additionalProperties"] = s.additionalProps

	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}

	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	return schema
}

func (s *ObjectSchema) GenerateExample() any {
	// Use provided examples first
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}

	example := make(map[string]any)

	// Generate examples for all properties
	for name, propSchema := range s.properties {
		example[name] = propSchema.GenerateExample()
	}

	return example
}

// Helper functions for format validation
func validateFormat(value, format string) error {
	switch format {
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("invalid email format")
		}
	case "uuid":
		uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
		if !uuidRegex.MatchString(value) {
			return fmt.Errorf("invalid UUID format")
		}
	case "url":
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(value) {
			return fmt.Errorf("invalid URL format")
		}
	}
	return nil
}

func getFormatSuggestion(format string) string {
	switch format {
	case "email":
		return "Provide a valid email address like 'user@example.com'"
	case "uuid":
		return "Provide a valid UUID like '123e4567-e89b-12d3-a456-426614174000'"
	case "url":
		return "Provide a valid URL like 'https://example.com'"
	default:
		return fmt.Sprintf("Provide a valid %s format", format)
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
		return "example"
	}
}
