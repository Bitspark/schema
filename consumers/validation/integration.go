// Package validation provides validation functionality using the consumer framework.
package validation

import (
	"defs.dev/schema/consumer"
	"defs.dev/schema/core"
)

// ValidateWithRegistry validates a value against a schema using validation consumers.
func ValidateWithRegistry(schema core.Schema, value any) ValidationResult {
	// Create a consumer registry with validation consumers
	registry := NewValidationRegistry()

	// Create a simple value wrapper
	valueWrapper := &simpleValue{value: value}

	// Process with validation purpose
	results, err := registry.ProcessValueAllWithPurpose("validation", schema, valueWrapper)
	if err != nil {
		// If processing failed, return an error result
		return ValidationResult{
			Valid: false,
			Errors: []ValidationIssue{{
				Path:    []string{},
				Code:    "validation_error",
				Message: err.Error(),
			}},
		}
	}

	// Aggregate all validation results
	aggregated := ValidationResult{Valid: true}

	for _, result := range results {
		if validationResult, ok := result.Value().(ValidationResult); ok {
			aggregated.Merge(validationResult)
		}
	}

	return aggregated
}

// NewValidationRegistry creates a new consumer registry with validation consumers registered.
func NewValidationRegistry() consumer.Registry {
	registry := consumer.NewRegistry()

	// Register validation consumers
	RegisterValidationConsumers(registry)

	return registry
}

// RegisterValidationConsumers registers all validation consumers with a registry.
func RegisterValidationConsumers(registry consumer.Registry) {
	// Register all validation consumers
	registry.RegisterValueConsumer(&StringValidationConsumer{})
	registry.RegisterValueConsumer(&BooleanValidationConsumer{})
	registry.RegisterValueConsumer(&NumberValidationConsumer{})
	registry.RegisterValueConsumer(&FunctionValidationConsumer{})
	registry.RegisterValueConsumer(&ArrayValidationConsumer{})
	registry.RegisterValueConsumer(&ObjectValidationConsumer{})
	registry.RegisterValueConsumer(&ServiceValidationConsumer{})
}
