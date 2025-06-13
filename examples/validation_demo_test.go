package examples

import (
	builders2 "defs.dev/schema/builders"
	validation2 "defs.dev/schema/consumers/validation"
	"testing"
)

func TestConsumerDrivenValidationDemo(t *testing.T) {
	t.Run("String validation", func(t *testing.T) {
		// Create a string schema
		stringSchema := builders2.NewStringSchema().Build()

		// Valid string
		result := validation2.ValidateValue(stringSchema, "hello")
		if !result.Valid {
			t.Errorf("Expected valid string to pass validation, got: %v", result.Errors)
		}

		// Invalid type
		result = validation2.ValidateValue(stringSchema, 123)
		if result.Valid {
			t.Error("Expected integer to fail string validation")
		}
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}
		if result.Errors[0].Code != "type_mismatch" {
			t.Errorf("Expected type_mismatch error, got: %s", result.Errors[0].Code)
		}
	})

	t.Run("Boolean validation", func(t *testing.T) {
		// Create a boolean schema
		boolSchema := builders2.NewBooleanSchema().Build()

		// Valid boolean
		result := validation2.ValidateValue(boolSchema, true)
		if !result.Valid {
			t.Errorf("Expected valid boolean to pass validation, got: %v", result.Errors)
		}

		// Invalid type
		result = validation2.ValidateValue(boolSchema, "not a boolean")
		if result.Valid {
			t.Error("Expected string to fail boolean validation")
		}
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}
		if result.Errors[0].Code != "type_mismatch" {
			t.Errorf("Expected type_mismatch error, got: %s", result.Errors[0].Code)
		}
	})

	t.Run("Function validation with required inputs", func(t *testing.T) {
		// Create a function schema with required inputs
		functionSchema := builders2.NewFunctionSchema().
			RequiredInput("name", builders2.NewStringSchema().Build()).
			RequiredInput("age", builders2.NewIntegerSchema().Build()).
			OptionalInput("email", builders2.NewStringSchema().Build()).
			Build()

		// Valid input with all required fields
		validInput := map[string]any{
			"name":  "John",
			"age":   30,
			"email": "john@example.com",
		}
		result := validation2.ValidateValue(functionSchema, validInput)
		if !result.Valid {
			t.Errorf("Expected valid input to pass validation, got: %v", result.Errors)
		}

		// Valid input without optional field
		validInputNoEmail := map[string]any{
			"name": "John",
			"age":  30,
		}
		result = validation2.ValidateValue(functionSchema, validInputNoEmail)
		if !result.Valid {
			t.Errorf("Expected valid input without optional field to pass validation, got: %v", result.Errors)
		}

		// Invalid input missing required field
		invalidInput := map[string]any{
			"age": 30,
			// "name" is missing
		}
		result = validation2.ValidateValue(functionSchema, invalidInput)
		if result.Valid {
			t.Error("Expected input missing required field to fail validation")
		}

		// Check for specific error
		found := false
		for _, err := range result.Errors {
			if err.Code == "missing_required_input" && len(err.Path) > 0 && err.Path[len(err.Path)-1] == "name" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected missing_required_input error for 'name', got: %v", result.Errors)
		}
	})

	t.Run("Consumer registry functionality", func(t *testing.T) {
		// Create a registry and register validators
		registry := validation2.NewValidationRegistry()

		// Check that validators are registered
		_, valueConsumers := registry.ListByPurpose("validation")
		if len(valueConsumers) == 0 {
			t.Error("Expected validation consumers to be registered")
		}

		// Check specific validators
		stringValidator, exists := registry.GetValueConsumer("string_validator")
		if !exists {
			t.Error("Expected string_validator to be registered")
		}
		if stringValidator.Purpose() != "validation" {
			t.Errorf("Expected string_validator purpose to be 'validation', got: %s", stringValidator.Purpose())
		}

		boolValidator, exists := registry.GetValueConsumer("boolean_validator")
		if !exists {
			t.Error("Expected boolean_validator to be registered")
		}
		if boolValidator.Purpose() != "validation" {
			t.Errorf("Expected boolean_validator purpose to be 'validation', got: %s", boolValidator.Purpose())
		}

		functionValidator, exists := registry.GetValueConsumer("function_validator")
		if !exists {
			t.Error("Expected function_validator to be registered")
		}
		if functionValidator.Purpose() != "validation" {
			t.Errorf("Expected function_validator purpose to be 'validation', got: %s", functionValidator.Purpose())
		}
	})
}
