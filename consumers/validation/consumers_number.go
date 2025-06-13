// Package validation provides validation consumers that implement the consumer framework.
package validation

import (
	"fmt"

	"defs.dev/schema/consumer"
	"defs.dev/schema/core"
)

// NumberValidationConsumer validates numeric values
type NumberValidationConsumer struct{}

func (c *NumberValidationConsumer) Name() string {
	return "number_validator"
}

func (c *NumberValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *NumberValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Or(consumer.Type(core.TypeNumber), consumer.Type(core.TypeInteger))
}

func (c *NumberValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual numeric value
	actualValue := value.Value()

	// Check if this is an integer schema first, and handle large integers specially
	if integerSchema, ok := ctx.Schema.(core.IntegerSchema); ok {
		return c.validateIntegerValue(integerSchema, actualValue, ctx.Path)
	}

	// For number schemas, convert to float64
	var numValue float64
	var ok bool

	switch v := actualValue.(type) {
	case int:
		numValue = float64(v)
		ok = true
	case int8:
		numValue = float64(v)
		ok = true
	case int16:
		numValue = float64(v)
		ok = true
	case int32:
		numValue = float64(v)
		ok = true
	case int64:
		numValue = float64(v)
		ok = true
	case uint:
		numValue = float64(v)
		ok = true
	case uint8:
		numValue = float64(v)
		ok = true
	case uint16:
		numValue = float64(v)
		ok = true
	case uint32:
		numValue = float64(v)
		ok = true
	case uint64:
		numValue = float64(v)
		ok = true
	case float64:
		numValue = v
		ok = true
	case float32:
		numValue = float64(v)
		ok = true
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "Expected number",
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "invalid numeric value",
			Code:    "invalid_number",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Check for special float values (NaN, Inf)
	if numValue != numValue { // NaN check
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "NaN is not allowed",
			Code:    "invalid_number",
		})
		return consumer.NewResult("validation", result), nil
	}

	if numValue == numValue+1 { // Infinity check
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "Infinity is not allowed",
			Code:    "invalid_number",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Try to cast to NumberSchema
	if numberSchema, ok := ctx.Schema.(core.NumberSchema); ok {
		if min := numberSchema.Minimum(); min != nil {
			if err := c.validateMin(numValue, *min, ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}

		if max := numberSchema.Maximum(); max != nil {
			if err := c.validateMax(numValue, *max, ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}
		return consumer.NewResult("validation", result), nil
	}

	// Fallback to annotation-based validation
	for _, annotation := range ctx.Schema.Annotations() {
		switch annotation.Name() {
		case "min":
			if err := c.validateMin(numValue, annotation.Value(), ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		case "max":
			if err := c.validateMax(numValue, annotation.Value(), ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *NumberValidationConsumer) validateIntegerValue(integerSchema core.IntegerSchema, actualValue any, path []string) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Check if the value is a valid integer type
	var intValue int64
	var uintValue uint64
	var isSignedInt bool
	var isUnsignedInt bool
	var isValidInteger bool

	switch v := actualValue.(type) {
	case int:
		intValue = int64(v)
		isSignedInt = true
		isValidInteger = true
	case int8:
		intValue = int64(v)
		isSignedInt = true
		isValidInteger = true
	case int16:
		intValue = int64(v)
		isSignedInt = true
		isValidInteger = true
	case int32:
		intValue = int64(v)
		isSignedInt = true
		isValidInteger = true
	case int64:
		intValue = v
		isSignedInt = true
		isValidInteger = true
	case uint:
		uintValue = uint64(v)
		isUnsignedInt = true
		isValidInteger = true
	case uint8:
		uintValue = uint64(v)
		isUnsignedInt = true
		isValidInteger = true
	case uint16:
		uintValue = uint64(v)
		isUnsignedInt = true
		isValidInteger = true
	case uint32:
		uintValue = uint64(v)
		isUnsignedInt = true
		isValidInteger = true
	case uint64:
		uintValue = v
		isUnsignedInt = true
		isValidInteger = true
	case float32:
		if v == float32(int64(v)) { // Check if it's a whole number
			intValue = int64(v)
			isSignedInt = true
			isValidInteger = true
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    path,
				Message: "expected integer value",
				Code:    "not_integer",
			})
			return consumer.NewResult("validation", result), nil
		}
	case float64:
		if v == float64(int64(v)) { // Check if it's a whole number
			intValue = int64(v)
			isSignedInt = true
			isValidInteger = true
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    path,
				Message: "expected integer value",
				Code:    "not_integer",
			})
			return consumer.NewResult("validation", result), nil
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    path,
			Message: fmt.Sprintf("expected integer, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	if !isValidInteger {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    path,
			Message: "invalid integer value",
			Code:    "invalid_integer",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Validate integer constraints
	if min := integerSchema.Minimum(); min != nil {
		if isSignedInt {
			if intValue < *min {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    path,
					Code:    "number_too_small",
					Message: fmt.Sprintf("value %d is less than minimum %d", intValue, *min),
				})
			}
		} else if isUnsignedInt {
			// For unsigned integers, check if the minimum is negative (which would always pass)
			// or if the uint value is less than the minimum
			if *min > 0 && uintValue < uint64(*min) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    path,
					Code:    "number_too_small",
					Message: fmt.Sprintf("value %d is less than minimum %d", uintValue, *min),
				})
			}
		}
	}

	if max := integerSchema.Maximum(); max != nil {
		if isSignedInt {
			if intValue > *max {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    path,
					Code:    "number_too_large",
					Message: fmt.Sprintf("value %d exceeds maximum %d", intValue, *max),
				})
			}
		} else if isUnsignedInt {
			// For unsigned integers, check if the maximum is negative (which would always fail)
			// or if the uint value is greater than the maximum
			if *max < 0 {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    path,
					Code:    "number_too_large",
					Message: fmt.Sprintf("value %d exceeds maximum %d", uintValue, *max),
				})
			} else if uintValue > uint64(*max) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    path,
					Code:    "number_too_large",
					Message: fmt.Sprintf("value %d exceeds maximum %d", uintValue, *max),
				})
			}
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *NumberValidationConsumer) validateMin(value float64, min any, path []string) *ValidationIssue {
	minVal, ok := min.(float64)
	if !ok {
		if minInt, ok := min.(int); ok {
			minVal = float64(minInt)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_min_annotation",
				Message: "min annotation must be a number",
			}
		}
	}

	if value < minVal {
		return &ValidationIssue{
			Path:    path,
			Code:    "number_too_small",
			Message: fmt.Sprintf("value %g is less than minimum %g", value, minVal),
		}
	}
	return nil
}

func (c *NumberValidationConsumer) validateMax(value float64, max any, path []string) *ValidationIssue {
	maxVal, ok := max.(float64)
	if !ok {
		if maxInt, ok := max.(int); ok {
			maxVal = float64(maxInt)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_max_annotation",
				Message: "max annotation must be a number",
			}
		}
	}

	if value > maxVal {
		return &ValidationIssue{
			Path:    path,
			Code:    "number_too_large",
			Message: fmt.Sprintf("value %g exceeds maximum %g", value, maxVal),
		}
	}
	return nil
}

func (c *NumberValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "number_validator",
		Purpose:      "validation",
		Description:  "Validates numeric values against numeric schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "number", "constraints"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
