// Package validation provides validation consumers that implement the consumer framework.
package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"defs.dev/schema/core"
	"defs.dev/schema/core/consumer"
)

// StringValidationConsumer validates string values against string schema constraints
type StringValidationConsumer struct{}

func (c *StringValidationConsumer) Name() string {
	return "string_validator"
}

func (c *StringValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *StringValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeString)
}

func (c *StringValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual string value
	actualValue := value.Value()
	str, ok := actualValue.(string)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected string, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Cast to StringSchema to access properties
	stringSchema, ok := ctx.Schema.(core.StringSchema)
	if !ok {
		// Fallback to annotation-based validation
		for _, annotation := range ctx.Schema.Annotations() {
			switch annotation.Name() {
			case "format":
				if err := c.validateFormat(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			case "pattern":
				if err := c.validatePattern(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			case "minLength":
				if err := c.validateMinLength(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			case "maxLength":
				if err := c.validateMaxLength(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			}
		}
		return consumer.NewResult("validation", result), nil
	}

	// Validate using schema properties
	if minLen := stringSchema.MinLength(); minLen != nil {
		if err := c.validateMinLength(str, *minLen, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	if maxLen := stringSchema.MaxLength(); maxLen != nil {
		if err := c.validateMaxLength(str, *maxLen, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	if pattern := stringSchema.Pattern(); pattern != "" {
		if err := c.validatePattern(str, pattern, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	if format := stringSchema.Format(); format != "" {
		if err := c.validateFormat(str, format, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	// Validate enum values
	if enumValues := stringSchema.EnumValues(); len(enumValues) > 0 {
		found := false
		for _, enumValue := range enumValues {
			if str == enumValue {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Code:    "enum_mismatch",
				Message: fmt.Sprintf("value '%s' is not one of the allowed values", str),
			})
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *StringValidationConsumer) validateFormat(value string, format any, path []string) *ValidationIssue {
	formatStr, ok := format.(string)
	if !ok {
		return &ValidationIssue{
			Path:    path,
			Code:    "invalid_format_annotation",
			Message: "format annotation must be a string",
		}
	}

	switch formatStr {
	case "email":
		if _, err := mail.ParseAddress(value); err != nil {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_email",
				Message: "invalid email format",
			}
		}
	case "url":
		if _, err := url.Parse(value); err != nil {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_url",
				Message: "invalid URL format",
			}
		}
	case "uuid":
		uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
		if matched, _ := regexp.MatchString(uuidPattern, strings.ToLower(value)); !matched {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_uuid",
				Message: "invalid UUID format",
			}
		}
	}
	return nil
}

func (c *StringValidationConsumer) validatePattern(value string, pattern any, path []string) *ValidationIssue {
	patternStr, ok := pattern.(string)
	if !ok {
		return &ValidationIssue{
			Path:    path,
			Code:    "invalid_pattern_annotation",
			Message: "pattern annotation must be a string",
		}
	}

	if matched, err := regexp.MatchString(patternStr, value); err != nil {
		return &ValidationIssue{
			Path:    path,
			Code:    "invalid_regex",
			Message: "invalid regular expression: " + err.Error(),
		}
	} else if !matched {
		return &ValidationIssue{
			Path:    path,
			Code:    "pattern_mismatch",
			Message: fmt.Sprintf("value does not match pattern: %s", patternStr),
		}
	}
	return nil
}

func (c *StringValidationConsumer) validateMinLength(value string, minLen any, path []string) *ValidationIssue {
	min, ok := minLen.(int)
	if !ok {
		if minFloat, ok := minLen.(float64); ok {
			min = int(minFloat)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_minlength_annotation",
				Message: "minLength annotation must be a number",
			}
		}
	}

	if len(value) < min {
		return &ValidationIssue{
			Path:    path,
			Code:    "string_too_short",
			Message: fmt.Sprintf("string length %d is less than minimum %d", len(value), min),
		}
	}
	return nil
}

func (c *StringValidationConsumer) validateMaxLength(value string, maxLen any, path []string) *ValidationIssue {
	max, ok := maxLen.(int)
	if !ok {
		if maxFloat, ok := maxLen.(float64); ok {
			max = int(maxFloat)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_maxlength_annotation",
				Message: "maxLength annotation must be a number",
			}
		}
	}

	if len(value) > max {
		return &ValidationIssue{
			Path:    path,
			Code:    "string_too_long",
			Message: fmt.Sprintf("string length %d exceeds maximum %d", len(value), max),
		}
	}
	return nil
}

func (c *StringValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "string_validator",
		Purpose:      "validation",
		Description:  "Validates string values against string schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "string", "constraints"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
