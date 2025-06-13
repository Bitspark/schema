// Package validation provides validation consumers that implement the consumer framework.
package validation

import (
	"fmt"
	"strings"

	"defs.dev/schema/consumer"
	"defs.dev/schema/core"
)

// BooleanValidationConsumer validates boolean values
type BooleanValidationConsumer struct{}

func (c *BooleanValidationConsumer) Name() string {
	return "boolean_validator"
}

func (c *BooleanValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *BooleanValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeBoolean)
}

func (c *BooleanValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual value
	actualValue := value.Value()

	// Try to cast to BooleanSchema to check for string conversion
	if _, ok := ctx.Schema.(core.BooleanSchema); ok {
		// Check if it's a boolean first
		if _, ok := actualValue.(bool); ok {
			return consumer.NewResult("validation", result), nil
		}

		// If we get here, it's not a boolean
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected boolean, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Fallback: just check if it's a boolean
	_, ok := actualValue.(bool)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected boolean, got %T", actualValue),
			Code:    "type_mismatch",
		})
	}

	return consumer.NewResult("validation", result), nil
}

func (c *BooleanValidationConsumer) convertStringToBool(str string, caseInsensitive bool) (bool, error) {
	// Always check exact matches first
	switch str {
	case "true", "1", "yes", "on", "y", "t":
		return true, nil
	case "false", "0", "no", "off", "n", "f":
		return false, nil
	}

	// For string conversion, be reasonably permissive with common case variations
	// even if CaseInsensitive is not explicitly enabled
	lowerStr := strings.ToLower(str)
	switch lowerStr {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	// If case insensitive is enabled, also check additional patterns
	if caseInsensitive {
		switch lowerStr {
		case "yes", "on", "y", "t", "1":
			return true, nil
		case "no", "off", "n", "f", "0":
			return false, nil
		}
	}

	return false, fmt.Errorf("invalid boolean string: %s", str)
}

func (c *BooleanValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "boolean_validator",
		Purpose:      "validation",
		Description:  "Validates boolean values against boolean schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "boolean"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
