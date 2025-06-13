package validation

import (
	"fmt"

	"defs.dev/schema/consumer"
	"defs.dev/schema/core"
)

// FunctionValidationConsumer validates function input/output values
type FunctionValidationConsumer struct{}

func (c *FunctionValidationConsumer) Name() string {
	return "function_validator"
}

func (c *FunctionValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *FunctionValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeFunction)
}

func (c *FunctionValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	functionSchema, ok := ctx.Schema.(core.FunctionSchema)
	if !ok {
		return nil, fmt.Errorf("expected function schema, got %T", ctx.Schema)
	}

	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual function input value (should be a map)
	actualValue := value.Value()
	inputMap, ok := actualValue.(map[string]any)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected function input map, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Validate inputs
	inputs := functionSchema.Inputs()
	for _, input := range inputs.Args() {
		inputName := input.Name()
		inputValue, exists := inputMap[inputName]

		// Check if required input is missing
		if !input.Optional() && !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    append(ctx.Path, inputName),
				Message: fmt.Sprintf("required input '%s' is missing", inputName),
				Code:    "missing_required_input",
			})
			continue
		}

		// If input exists, validate its value against the input schema
		if exists {
			inputSchema := input.Schema()
			inputPath := append(ctx.Path, inputName)

			// Use recursive validation for the input value
			inputResult := ValidateWithRegistry(inputSchema, inputValue)
			if !inputResult.Valid {
				result.Valid = false
				// Add path context to input errors
				for _, err := range inputResult.Errors {
					err.Path = append(inputPath, err.Path...)
					result.Errors = append(result.Errors, err)
				}
			}
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *FunctionValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "function_validator",
		Purpose:      "validation",
		Description:  "Validates function input/output values against function schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "function", "inputs", "outputs"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
