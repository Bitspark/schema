package tests

import (
	"testing"

	"defs.dev/schema/api/core"
	"defs.dev/schema/builders"
	"defs.dev/schema/schemas"
	"defs.dev/schema/validation"
)

func TestFunctionSchemaBuilder(t *testing.T) {
	t.Run("Basic function schema creation", func(t *testing.T) {
		schema := builders.NewFunctionSchema().
			Name("testFunction").
			Description("A test function").
			Input("name", builders.NewStringSchema().Build()).
			Input("age", builders.NewIntegerSchema().Build()).
			Output("greeting", builders.NewStringSchema().Build()).
			RequiredInputs("name").
			RequiredOutputs("greeting").
			Build()

		if schema.Metadata().Name != "testFunction" {
			t.Errorf("Expected name 'testFunction', got %s", schema.Metadata().Name)
		}

		if schema.Metadata().Description != "A test function" {
			t.Errorf("Expected description 'A test function', got %s", schema.Metadata().Description)
		}

		// Test inputs
		inputs := schema.Inputs()
		inputArgs := inputs.Args()
		if len(inputArgs) != 2 {
			t.Errorf("Expected 2 inputs, got %d", len(inputArgs))
		}

		// Find name input
		var nameArg core.ArgSchema
		found := false
		for _, arg := range inputArgs {
			if arg.Name() == "name" {
				nameArg = arg
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'name' input to exist")
		} else if nameArg.Schema().Type() != core.TypeString {
			t.Errorf("Expected 'name' to be string type, got %s", nameArg.Schema().Type())
		}

		// Test outputs
		outputs := schema.Outputs()
		outputArgs := outputs.Args()
		if len(outputArgs) != 1 {
			t.Errorf("Expected 1 output, got %d", len(outputArgs))
		}

		// Find greeting output
		var greetingArg core.ArgSchema
		found = false
		for _, arg := range outputArgs {
			if arg.Name() == "greeting" {
				greetingArg = arg
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'greeting' output to exist")
		} else if greetingArg.Schema().Type() != core.TypeString {
			t.Errorf("Expected 'greeting' to be string type, got %s", greetingArg.Schema().Type())
		}

		// Test required inputs/outputs - by default all inputs are required
		requiredInputs := schema.RequiredInputs()
		if len(requiredInputs) != 2 {
			t.Errorf("Expected 2 required inputs by default, got %v", requiredInputs)
		}

		requiredOutputs := schema.RequiredOutputs()
		if len(requiredOutputs) != 1 {
			t.Errorf("Expected 1 required output by default, got %v", requiredOutputs)
		}
	})

	t.Run("Function schema with error handling", func(t *testing.T) {
		errorSchema := builders.NewStringSchema().Build()
		schema := builders.NewFunctionSchema().
			Name("errorFunction").
			Input("data", builders.NewStringSchema().Build()).
			Output("result", builders.NewStringSchema().Build()).
			Error(errorSchema).
			Build()

		if schema.Errors() == nil {
			t.Error("Expected error schema to be set")
		}

		if schema.Errors().Type() != core.TypeString {
			t.Errorf("Expected error schema to be string type, got %s", schema.Errors().Type())
		}
	})

	t.Run("Function schema with examples", func(t *testing.T) {
		example := map[string]any{
			"input":  "test",
			"output": "processed test",
		}

		schema := builders.NewFunctionSchema().
			Name("exampleFunction").
			Input("input", builders.NewStringSchema().Build()).
			Output("output", builders.NewStringSchema().Build()).
			Example(example).
			Build()

		// Check examples in the function schema itself
		if functionSchema, ok := schema.(*schemas.FunctionSchema); ok {
			examples := functionSchema.Examples()
			if len(examples) != 1 {
				t.Errorf("Expected 1 example, got %d", len(examples))
			}

			if examples[0]["input"] != "test" {
				t.Errorf("Expected example input 'test', got %v", examples[0]["input"])
			}
		} else {
			t.Error("Expected schema to be *schemas.FunctionSchema")
		}
	})

	t.Run("Function schema with tags", func(t *testing.T) {
		schema := builders.NewFunctionSchema().
			Name("taggedFunction").
			Tag("api").
			Tag("public").
			Input("data", builders.NewStringSchema().Build()).
			Build()

		tags := schema.Metadata().Tags
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}

		expectedTags := map[string]bool{"api": true, "public": true}
		for _, tag := range tags {
			if !expectedTags[tag] {
				t.Errorf("Unexpected tag: %s", tag)
			}
		}
	})

	t.Run("Complex function schema", func(t *testing.T) {
		userSchema := builders.NewObject().
			Property("id", builders.NewIntegerSchema().Build()).
			Property("name", builders.NewStringSchema().Build()).
			Property("email", builders.NewStringSchema().Build()).
			Required("id", "name").
			Build()

		schema := builders.NewFunctionSchema().
			Name("createUser").
			Description("Creates a new user in the system").
			Input("userData", userSchema).
			Input("options", builders.NewObject().AdditionalProperties(true).Build()).
			Output("user", userSchema).
			Output("success", builders.NewBooleanSchema().Build()).
			RequiredInputs("userData").
			RequiredOutputs("user", "success").
			Tag("user").
			Tag("create").
			Build()

		// Validate the complex schema
		inputArgs := schema.Inputs().Args()
		if len(inputArgs) != 2 {
			t.Errorf("Expected 2 inputs, got %d", len(inputArgs))
		}

		outputArgs := schema.Outputs().Args()
		if len(outputArgs) != 2 {
			t.Errorf("Expected 2 outputs, got %d", len(outputArgs))
		}

		// Find userData input
		var userDataArg core.ArgSchema
		found := false
		for _, arg := range inputArgs {
			if arg.Name() == "userData" {
				userDataArg = arg
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'userData' input to exist")
		} else if userDataArg.Schema().Type() != core.TypeStructure {
			t.Errorf("Expected 'userData' to be object type, got %s", userDataArg.Schema().Type())
		}
	})
}

func TestFunctionSchemaValidation(t *testing.T) {
	schema := builders.NewFunctionSchema().
		Name("validateFunction").
		Input("name", builders.NewStringSchema().Build()).
		Input("age", builders.NewIntegerSchema().Build()).
		Output("valid", builders.NewBooleanSchema().Build()).
		RequiredInputs("name").
		Build()

	t.Run("Valid function data", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}

		result := validation.ValidateValue(schema, data)
		if !result.Valid {
			t.Errorf("Expected validation to pass, got errors: %v", result.Errors)
		}
	})

	t.Run("Missing required input", func(t *testing.T) {
		data := map[string]any{
			"age": 30,
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
				if err.Code == "missing_required_input" && len(err.Path) > 0 && err.Path[len(err.Path)-1] == "name" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected missing_required_input error for 'name', got errors: %v", result.Errors)
			}
		}
	})

	t.Run("Invalid input type", func(t *testing.T) {
		data := map[string]any{
			"name": 123, // Should be string
			"age":  30,
		}

		result := validation.ValidateValue(schema, data)
		if result.Valid {
			t.Error("Expected validation to fail for invalid input type")
		}
	})
}

func TestFunctionSchemaCloning(t *testing.T) {
	original := builders.NewFunctionSchema().
		Name("originalFunction").
		Description("Original description").
		Input("input", builders.NewStringSchema().Build()).
		Output("output", builders.NewStringSchema().Build()).
		Tag("original").
		Build()

	cloned := original.Clone().(core.FunctionSchema)

	// Test that clone is independent
	if cloned.Metadata().Name != original.Metadata().Name {
		t.Error("Cloned schema should have same name as original")
	}

	if len(cloned.Inputs().Args()) != len(original.Inputs().Args()) {
		t.Error("Cloned schema should have same number of inputs")
	}

	if len(cloned.Outputs().Args()) != len(original.Outputs().Args()) {
		t.Error("Cloned schema should have same number of outputs")
	}

	// Verify it's a deep clone by checking that modifications don't affect original
	// (This would require modifying the clone, which current API doesn't support directly)
}

func TestFunctionSchemaJSONSchema(t *testing.T) {
	schema := builders.NewFunctionSchema().
		Name("jsonFunction").
		Description("Function for JSON schema testing").
		Input("name", builders.NewStringSchema().Build()).
		Input("age", builders.NewIntegerSchema().Build()).
		Output("greeting", builders.NewStringSchema().Build()).
		Build()

	jsonSchema := toJSONSchema(schema)

	// Verify basic structure
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", jsonSchema["type"])
	}

	if jsonSchema["description"] != "Function for JSON schema testing" {
		t.Errorf("Expected description to be preserved in JSON schema")
	}

	// Check for function-specific properties (based on actual implementation)
	if _, exists := jsonSchema["x-function"]; !exists {
		t.Error("Expected x-function property in JSON schema")
	}

	if _, exists := jsonSchema["x-returns"]; !exists {
		t.Error("Expected x-returns property in JSON schema")
	}
}

func TestFunctionSchemaBuilderChaining(t *testing.T) {
	// Test that all builder methods return the correct type for chaining
	builder := builders.NewFunctionSchema().
		Name("chainedFunction").
		Description("Testing method chaining").
		Tag("test").
		Input("a", builders.NewStringSchema().Build()).
		Input("b", builders.NewIntegerSchema().Build()).
		Output("result", builders.NewStringSchema().Build()).
		RequiredInputs("a").
		RequiredOutputs("result").
		Example(map[string]any{"a": "test", "b": 42})

	schema := builder.Build()

	if schema.Metadata().Name != "chainedFunction" {
		t.Error("Method chaining failed to preserve name")
	}

	if len(schema.Metadata().Tags) != 1 || schema.Metadata().Tags[0] != "test" {
		t.Error("Method chaining failed to preserve tags")
	}

	if len(schema.Inputs().Args()) != 2 {
		t.Error("Method chaining failed to preserve inputs")
	}

	if len(schema.Outputs().Args()) != 1 {
		t.Error("Method chaining failed to preserve outputs")
	}
}

func TestFunctionSchemaAdvancedFeatures(t *testing.T) {
	t.Run("Function with complex nested schemas", func(t *testing.T) {
		addressSchema := builders.NewObject().
			Property("street", builders.NewStringSchema().Build()).
			Property("city", builders.NewStringSchema().Build()).
			Property("zipCode", builders.NewStringSchema().Build()).
			Required("street", "city").
			Build()

		personSchema := builders.NewObject().
			Property("name", builders.NewStringSchema().Build()).
			Property("age", builders.NewIntegerSchema().Build()).
			Property("address", addressSchema).
			Property("hobbies", builders.NewArraySchema().Items(builders.NewStringSchema().Build()).Build()).
			Required("name").
			Build()

		schema := builders.NewFunctionSchema().
			Name("processPersonData").
			Input("person", personSchema).
			Output("processed", builders.NewBooleanSchema().Build()).
			Output("errors", builders.NewArraySchema().Items(builders.NewStringSchema().Build()).Build()).
			RequiredInputs("person").
			RequiredOutputs("processed").
			Build()

		// Test validation with nested data
		validData := map[string]any{
			"person": map[string]any{
				"name": "John Doe",
				"age":  30,
				"address": map[string]any{
					"street":  "123 Main St",
					"city":    "Anytown",
					"zipCode": "12345",
				},
				"hobbies": []any{"reading", "swimming"},
			},
		}

		result := validation.ValidateValue(schema, validData)
		if !result.Valid {
			t.Errorf("Expected validation to pass for valid nested data, got errors: %v", result.Errors)
		}
	})

	t.Run("Function with array inputs and outputs", func(t *testing.T) {
		schema := builders.NewFunctionSchema().
			Name("processArray").
			Input("numbers", builders.NewArraySchema().Items(builders.NewNumberSchema().Build()).Build()).
			Input("operation", builders.NewStringSchema().Build()).
			Output("results", builders.NewArraySchema().Items(builders.NewNumberSchema().Build()).Build()).
			Output("summary", builders.NewObject().
				Property("count", builders.NewIntegerSchema().Build()).
				Property("sum", builders.NewNumberSchema().Build()).
				Build()).
			RequiredInputs("numbers", "operation").
			RequiredOutputs("results").
			Build()

		data := map[string]any{
			"numbers":   []any{1.0, 2.0, 3.0, 4.0, 5.0},
			"operation": "square",
		}

		result := validation.ValidateValue(schema, data)
		if !result.Valid {
			t.Errorf("Expected validation to pass for array data, got errors: %v", result.Errors)
		}
	})
}
