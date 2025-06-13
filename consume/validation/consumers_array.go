package validation

import (
	"defs.dev/schema/core/consumer"
	"fmt"

	"defs.dev/schema/core"
)

// ArrayValidationConsumer validates array values
type ArrayValidationConsumer struct{}

func (c *ArrayValidationConsumer) Name() string {
	return "array_validator"
}

func (c *ArrayValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *ArrayValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeArray)
}

func (c *ArrayValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual array value
	actualValue := value.Value()

	// Check if it's an array/slice
	var arrayItems []any
	switch v := actualValue.(type) {
	case []any:
		arrayItems = v
	case []string:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	case []int:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	case []float64:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	case []bool:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected array, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Cast to ArraySchema to access properties
	arraySchema, ok := ctx.Schema.(core.ArraySchema)
	if !ok {
		// Fallback validation - just check it's an array
		return consumer.NewResult("validation", result), nil
	}

	// Validate array constraints
	if minItems := arraySchema.MinItems(); minItems != nil {
		if len(arrayItems) < *minItems {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("array has %d items, minimum required is %d", len(arrayItems), *minItems),
				Code:    "min_items_violation",
			})
		}
	}

	if maxItems := arraySchema.MaxItems(); maxItems != nil {
		if len(arrayItems) > *maxItems {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("array has %d items, maximum allowed is %d", len(arrayItems), *maxItems),
				Code:    "max_items_violation",
			})
		}
	}

	// Validate unique items if required
	if arraySchema.UniqueItemsRequired() {
		seen := make(map[string]bool)
		for i, item := range arrayItems {
			itemStr := fmt.Sprintf("%v", item)
			if seen[itemStr] {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    append(ctx.Path, fmt.Sprintf("[%d]", i)),
					Message: fmt.Sprintf("duplicate item found: %v", item),
					Code:    "unique_items_violation",
				})
			}
			seen[itemStr] = true
		}
	}

	// Validate each item against the item schema
	if itemSchema := arraySchema.ItemSchema(); itemSchema != nil {
		for i, item := range arrayItems {
			itemPath := append(ctx.Path, fmt.Sprintf("[%d]", i))

			// Use recursive validation for the item
			itemResult := ValidateWithRegistry(itemSchema, item)
			if !itemResult.Valid {
				result.Valid = false
				// Add path context to item errors
				for _, err := range itemResult.Errors {
					err.Path = itemPath
					result.Errors = append(result.Errors, err)
				}
			}
		}
	}

	// Validate contains constraint
	if containsSchema := arraySchema.ContainsSchema(); containsSchema != nil {
		containsMatched := false
		for _, item := range arrayItems {
			itemResult := ValidateWithRegistry(containsSchema, item)
			if itemResult.Valid {
				containsMatched = true
				break
			}
		}
		if !containsMatched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: "array does not contain any item matching the contains schema",
				Code:    "contains_constraint_violation",
			})
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *ArrayValidationConsumer) validateItem(ctx consumer.ProcessingContext, value core.Value[any]) ValidationResult {
	// Use recursive validation instead of simplified validation
	actualValue := value.Value()
	return ValidateWithRegistry(ctx.Schema, actualValue)
}

func (c *ArrayValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "array_validator",
		Purpose:      "validation",
		Description:  "Validates array values against array schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "array", "constraints"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
