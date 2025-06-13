package schemas

import (
	"defs.dev/schema/consumers/validation"
	"testing"

	"defs.dev/schema/core"
)

func TestFunctionSchema_RequiredInputValidation(t *testing.T) {
	// Create input schemas using proper constructors
	nameSchema := NewStringSchema(StringSchemaConfig{
		Metadata: core.SchemaMetadata{Name: "name"},
	})
	ageSchema := NewIntegerSchema(IntegerSchemaConfig{
		Metadata: core.SchemaMetadata{Name: "age"},
	})

	// Create function schema with required inputs
	inputs := NewArgSchemas()
	inputs.AddArg(NewArgSchema("name", nameSchema))
	inputs.AddArg(NewArgSchema("age", ageSchema))
	inputs.SetOptionalByName("age", true) // age is optional, name is required

	outputs := NewArgSchemas()

	schema := NewFunctionSchema(inputs, outputs)
	schema.metadata = core.SchemaMetadata{Name: "testFunction"}

	t.Run("Valid input with all required fields", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}

		result := validation.ValidateValue(schema, data)
		if !result.Valid {
			t.Errorf("Expected validation to pass, got errors: %v", result.Errors)
		}
	})

	t.Run("Missing required input should fail", func(t *testing.T) {
		data := map[string]any{
			"age": 30,
			// "name" is missing but required
		}

		result := validation.ValidateValue(schema, data)
		if result.Valid {
			t.Error("Expected validation to fail for missing required input 'name'")
		}

		// Check that we get the correct error
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for missing required input")
		} else {
			found := false
			for _, err := range result.Errors {
				if err.Code == "missing_required_input" {
					found = true
					if err.Message != "required input 'name' is missing" {
						t.Errorf("Expected error message 'required input 'name' is missing', got: %s", err.Message)
					}
					break
				}
			}
			if !found {
				t.Errorf("Expected missing_required_input error for 'name', got errors: %v", result.Errors)
			}
		}
	})

	t.Run("Multiple missing required inputs", func(t *testing.T) {
		// Create schema with multiple required inputs
		emailSchema := NewStringSchema(StringSchemaConfig{
			Metadata: core.SchemaMetadata{Name: "email"},
		})

		multiInputs := NewArgSchemas()
		multiInputs.AddArg(NewArgSchema("name", nameSchema))
		multiInputs.AddArg(NewArgSchema("email", emailSchema))
		multiInputs.AddArg(NewArgSchema("age", ageSchema))
		multiInputs.SetOptionalByName("age", true) // only age is optional

		multiRequiredSchema := NewFunctionSchema(multiInputs, outputs)
		multiRequiredSchema.metadata = core.SchemaMetadata{Name: "multiRequiredFunction"}

		data := map[string]any{
			"age": 30,
			// Both "name" and "email" are missing
		}

		result := validation.ValidateValue(multiRequiredSchema, data)
		if result.Valid {
			t.Error("Expected validation to fail for missing required inputs")
		}

		// Check that we get errors for both missing required inputs
		nameErrorFound := false
		emailErrorFound := false
		for _, err := range result.Errors {
			if err.Code == "missing_required_input" {
				if len(err.Path) > 0 && err.Path[len(err.Path)-1] == "name" {
					nameErrorFound = true
				}
				if len(err.Path) > 0 && err.Path[len(err.Path)-1] == "email" {
					emailErrorFound = true
				}
			}
		}

		if !nameErrorFound {
			t.Error("Expected missing_required_input error for 'name'")
		}
		if !emailErrorFound {
			t.Error("Expected missing_required_input error for 'email'")
		}
	})

	t.Run("Optional inputs can be missing", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			// "age" is optional and can be missing
		}

		result := validation.ValidateValue(schema, data)
		if !result.Valid {
			t.Errorf("Expected validation to pass when optional input is missing, got errors: %v", result.Errors)
		}
	})
}

func TestFunctionSchema_RequiredInputsMethod(t *testing.T) {
	nameSchema := NewStringSchema(StringSchemaConfig{
		Metadata: core.SchemaMetadata{Name: "name"},
	})
	ageSchema := NewIntegerSchema(IntegerSchemaConfig{
		Metadata: core.SchemaMetadata{Name: "age"},
	})
	emailSchema := NewStringSchema(StringSchemaConfig{
		Metadata: core.SchemaMetadata{Name: "email"},
	})

	inputs := NewArgSchemas()
	inputs.AddArg(NewArgSchema("name", nameSchema))
	inputs.AddArg(NewArgSchema("age", ageSchema))
	inputs.AddArg(NewArgSchema("email", emailSchema))
	inputs.SetOptionalByName("age", true) // only age is optional

	outputs := NewArgSchemas()

	schema := NewFunctionSchema(inputs, outputs)
	schema.metadata = core.SchemaMetadata{Name: "testFunction"}

	requiredInputs := schema.RequiredInputs()

	if len(requiredInputs) != 2 {
		t.Errorf("Expected 2 required inputs, got %d", len(requiredInputs))
	}

	expectedRequired := map[string]bool{"name": true, "email": true}
	for _, input := range requiredInputs {
		if !expectedRequired[input] {
			t.Errorf("Unexpected required input: %s", input)
		}
		delete(expectedRequired, input)
	}

	if len(expectedRequired) > 0 {
		t.Errorf("Missing required inputs in result: %v", expectedRequired)
	}
}
