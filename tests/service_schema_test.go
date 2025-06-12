package tests

import (
	"testing"

	"defs.dev/schema/api/core"
	"defs.dev/schema/builders"
	"defs.dev/schema/schemas"
)

func TestServiceSchemaBuilder(t *testing.T) {
	t.Run("Basic service schema creation", func(t *testing.T) {
		greetFunction := builders.NewFunctionSchema().
			Name("greet").
			Description("Greets a user").
			Input("name", builders.NewStringSchema().Build()).
			Output("greeting", builders.NewStringSchema().Build()).
			RequiredInputs("name").
			RequiredOutputs("greeting").
			Build()

		schema := builders.NewServiceSchema().
			Name("GreetingService").
			Description("A service for greeting users").
			Method("greet", greetFunction).
			Tag("greeting").
			Tag("service").
			Build()

		if schema.Name() != "GreetingService" {
			t.Errorf("Expected name 'GreetingService', got %s", schema.Name())
		}

		if schema.Metadata().Description != "A service for greeting users" {
			t.Errorf("Expected description 'A service for greeting users', got %s", schema.Metadata().Description)
		}

		// Test methods
		methods := schema.Methods()
		if len(methods) != 1 {
			t.Errorf("Expected 1 method, got %d", len(methods))
		}

		greetMethod := methods[0]
		if greetMethod.Name() != "greet" {
			t.Errorf("Expected method name 'greet', got %s", greetMethod.Name())
		}

		if greetMethod.Function().Metadata().Name != "greet" {
			t.Errorf("Expected function name 'greet', got %s", greetMethod.Function().Metadata().Name)
		}

		// Test tags
		tags := schema.Metadata().Tags
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}

		expectedTags := map[string]bool{"greeting": true, "service": true}
		for _, tag := range tags {
			if !expectedTags[tag] {
				t.Errorf("Unexpected tag: %s", tag)
			}
		}
	})

	t.Run("Service schema with multiple methods", func(t *testing.T) {
		// Create function schemas for different methods
		createUserFunc := builders.NewFunctionSchema().
			Name("createUser").
			Input("userData", builders.NewObject().
				Property("name", builders.NewStringSchema().Build()).
				Property("email", builders.NewStringSchema().Build()).
				Required("name", "email").
				Build()).
			Output("user", builders.NewObject().
				Property("id", builders.NewIntegerSchema().Build()).
				Property("name", builders.NewStringSchema().Build()).
				Property("email", builders.NewStringSchema().Build()).
				Build()).
			Output("success", builders.NewBooleanSchema().Build()).
			RequiredInputs("userData").
			RequiredOutputs("user", "success").
			Build()

		getUserFunc := builders.NewFunctionSchema().
			Name("getUser").
			Input("id", builders.NewIntegerSchema().Build()).
			Output("user", builders.NewObject().
				Property("id", builders.NewIntegerSchema().Build()).
				Property("name", builders.NewStringSchema().Build()).
				Property("email", builders.NewStringSchema().Build()).
				Build()).
			RequiredInputs("id").
			RequiredOutputs("user").
			Build()

		deleteUserFunc := builders.NewFunctionSchema().
			Name("deleteUser").
			Input("id", builders.NewIntegerSchema().Build()).
			Output("success", builders.NewBooleanSchema().Build()).
			RequiredInputs("id").
			RequiredOutputs("success").
			Build()

		schema := builders.NewServiceSchema().
			Name("UserService").
			Description("A service for managing users").
			Method("createUser", createUserFunc).
			Method("getUser", getUserFunc).
			Method("deleteUser", deleteUserFunc).
			Tag("user").
			Tag("crud").
			Build()

		// Validate the service schema
		if schema.Name() != "UserService" {
			t.Errorf("Expected name 'UserService', got %s", schema.Name())
		}

		methods := schema.Methods()
		if len(methods) != 3 {
			t.Errorf("Expected 3 methods, got %d", len(methods))
		}

		// Check that all methods exist
		methodNames := make(map[string]bool)
		for _, method := range methods {
			methodNames[method.Name()] = true
		}

		expectedMethods := []string{"createUser", "getUser", "deleteUser"}
		for _, expectedMethod := range expectedMethods {
			if !methodNames[expectedMethod] {
				t.Errorf("Expected method '%s' not found", expectedMethod)
			}
		}
	})

	t.Run("Service schema with examples", func(t *testing.T) {
		echoFunc := builders.NewFunctionSchema().
			Name("echo").
			Input("message", builders.NewStringSchema().Build()).
			Output("response", builders.NewStringSchema().Build()).
			Example(map[string]any{
				"message":  "Hello",
				"response": "Echo: Hello",
			}).
			Build()

		schema := builders.NewServiceSchema().
			Name("EchoService").
			Method("echo", echoFunc).
			Example(map[string]any{
				"service": "EchoService",
				"version": "1.0",
			}).
			Build()

		// Check service-level examples
		examples := schema.Metadata().Examples
		if len(examples) != 1 {
			t.Errorf("Expected 1 service example, got %d", len(examples))
		}

		if examples[0].(map[string]any)["service"] != "EchoService" {
			t.Errorf("Expected service example to contain 'EchoService'")
		}
	})
}

func TestServiceSchemaIntrospection(t *testing.T) {
	// Create a comprehensive service for testing introspection
	calculateFunc := builders.NewFunctionSchema().
		Name("calculate").
		Input("a", builders.NewNumberSchema().Build()).
		Input("b", builders.NewNumberSchema().Build()).
		Input("operation", builders.NewStringSchema().Build()).
		Output("result", builders.NewNumberSchema().Build()).
		RequiredInputs("a", "b", "operation").
		RequiredOutputs("result").
		Build()

	validateFunc := builders.NewFunctionSchema().
		Name("validate").
		Input("expression", builders.NewStringSchema().Build()).
		Output("valid", builders.NewBooleanSchema().Build()).
		Output("errors", builders.NewArraySchema().Items(builders.NewStringSchema().Build()).Build()).
		RequiredInputs("expression").
		RequiredOutputs("valid").
		Build()

	schema := builders.NewServiceSchema().
		Name("CalculatorService").
		Description("A mathematical calculator service").
		Method("calculate", calculateFunc).
		Method("validate", validateFunc).
		Tag("math").
		Tag("calculator").
		Build()

	t.Run("Service name and metadata", func(t *testing.T) {
		if schema.Name() != "CalculatorService" {
			t.Errorf("Expected name 'CalculatorService', got %s", schema.Name())
		}

		if schema.Metadata().Description != "A mathematical calculator service" {
			t.Errorf("Expected description to match")
		}

		if schema.Type() != core.TypeService {
			t.Errorf("Expected type to be TypeService, got %s", schema.Type())
		}
	})

	t.Run("Method introspection", func(t *testing.T) {
		methods := schema.Methods()
		if len(methods) != 2 {
			t.Errorf("Expected 2 methods, got %d", len(methods))
		}

		// Test method lookup (cast to concrete type to access GetMethod)
		if serviceSchema, ok := schema.(*schemas.ServiceSchema); ok {
			calculateMethod, found := serviceSchema.GetMethod("calculate")
			if !found {
				t.Error("Expected to find 'calculate' method")
			} else {
				if calculateMethod.Name() != "calculate" {
					t.Errorf("Expected method name 'calculate', got %s", calculateMethod.Name())
				}

				funcSchema := calculateMethod.Function()
				if len(funcSchema.Inputs().Args()) != 3 {
					t.Errorf("Expected 3 inputs for calculate method, got %d", len(funcSchema.Inputs().Args()))
				}
			}

			// Test non-existent method
			_, found = serviceSchema.GetMethod("nonexistent")
			if found {
				t.Error("Expected not to find 'nonexistent' method")
			}
		} else {
			t.Error("Expected schema to be *schemas.ServiceSchema")
		}
	})

	t.Run("Method validation", func(t *testing.T) {
		if serviceSchema, ok := schema.(*schemas.ServiceSchema); ok {
			calculateMethod, _ := serviceSchema.GetMethod("calculate")
			funcSchema := calculateMethod.Function()

			// Test valid input
			validInput := map[string]any{
				"a":         10.0,
				"b":         5.0,
				"operation": "add",
			}

			result := funcSchema.Validate(validInput)
			if !result.Valid {
				t.Errorf("Expected validation to pass for valid input, got errors: %v", result.Errors)
			}

			// Test invalid input
			invalidInput := map[string]any{
				"a":         "not a number",
				"b":         5.0,
				"operation": "add",
			}

			result = funcSchema.Validate(invalidInput)
			if result.Valid {
				t.Error("Expected validation to fail for invalid input type")
			}
		} else {
			t.Error("Expected schema to be *schemas.ServiceSchema")
		}
	})
}

func TestServiceSchemaCloning(t *testing.T) {
	originalFunc := builders.NewFunctionSchema().
		Name("original").
		Input("input", builders.NewStringSchema().Build()).
		Output("output", builders.NewStringSchema().Build()).
		Build()

	original := builders.NewServiceSchema().
		Name("OriginalService").
		Description("Original service").
		Method("original", originalFunc).
		Tag("original").
		Build()

	cloned := original.Clone().(core.ServiceSchema)

	// Test that clone is independent
	if cloned.Name() != original.Name() {
		t.Error("Cloned schema should have same name as original")
	}

	if len(cloned.Methods()) != len(original.Methods()) {
		t.Error("Cloned schema should have same number of methods")
	}

	if cloned.Metadata().Description != original.Metadata().Description {
		t.Error("Cloned schema should have same description")
	}

	// Verify it's a deep clone by checking method details
	originalMethods := original.Methods()
	clonedMethods := cloned.Methods()

	if originalMethods[0].Name() != clonedMethods[0].Name() {
		t.Error("Cloned method should have same name as original")
	}
}

func TestServiceSchemaJSONSchema(t *testing.T) {
	simpleFunc := builders.NewFunctionSchema().
		Name("simple").
		Input("input", builders.NewStringSchema().Build()).
		Output("output", builders.NewStringSchema().Build()).
		Build()

	schema := builders.NewServiceSchema().
		Name("SimpleService").
		Description("A simple service for testing").
		Method("simple", simpleFunc).
		Build()

	jsonSchema := toJSONSchema(schema)

	// Verify basic structure
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", jsonSchema["type"])
	}

	if jsonSchema["description"] != "A simple service for testing" {
		t.Errorf("Expected description to be preserved in JSON schema")
	}

	// Check for service-specific properties
	if _, exists := jsonSchema["x-service"]; !exists {
		t.Error("Expected x-service property in JSON schema")
	}

	if _, exists := jsonSchema["x-methods"]; !exists {
		t.Error("Expected x-methods property in JSON schema")
	}
}

func TestServiceSchemaBuilderChaining(t *testing.T) {
	// Test that all builder methods return the correct type for chaining
	func1 := builders.NewFunctionSchema().
		Name("func1").
		Input("input", builders.NewStringSchema().Build()).
		Build()

	func2 := builders.NewFunctionSchema().
		Name("func2").
		Input("input", builders.NewIntegerSchema().Build()).
		Build()

	builder := builders.NewServiceSchema().
		Name("ChainedService").
		Description("Testing method chaining").
		Tag("test").
		Method("func1", func1).
		Method("func2", func2).
		Example(map[string]any{"test": "example"})

	schema := builder.Build()

	if schema.Name() != "ChainedService" {
		t.Error("Method chaining failed to preserve name")
	}

	if len(schema.Metadata().Tags) != 1 || schema.Metadata().Tags[0] != "test" {
		t.Error("Method chaining failed to preserve tags")
	}

	if len(schema.Methods()) != 2 {
		t.Error("Method chaining failed to preserve methods")
	}
}

func TestServiceSchemaAdvancedFeatures(t *testing.T) {
	t.Run("Service with complex method signatures", func(t *testing.T) {
		// Create a complex data processing service
		processDataFunc := builders.NewFunctionSchema().
			Name("processData").
			Input("data", builders.NewArraySchema().
				Items(builders.NewObject().
					Property("id", builders.NewIntegerSchema().Build()).
					Property("value", builders.NewNumberSchema().Build()).
					Property("metadata", builders.NewObject().AdditionalProperties(true).Build()).
					Required("id", "value").
					Build()).
				Build()).
			Input("options", builders.NewObject().
				Property("sortBy", builders.NewStringSchema().Build()).
				Property("filterBy", builders.NewObject().AdditionalProperties(true).Build()).
				Property("limit", builders.NewIntegerSchema().Build()).
				Build()).
			Output("processedData", builders.NewArraySchema().
				Items(builders.NewObject().
					Property("id", builders.NewIntegerSchema().Build()).
					Property("processedValue", builders.NewNumberSchema().Build()).
					Property("status", builders.NewStringSchema().Build()).
					Build()).
				Build()).
			Output("summary", builders.NewObject().
				Property("totalProcessed", builders.NewIntegerSchema().Build()).
				Property("errors", builders.NewArraySchema().Items(builders.NewStringSchema().Build()).Build()).
				Build()).
			RequiredInputs("data").
			RequiredOutputs("processedData", "summary").
			Build()

		schema := builders.NewServiceSchema().
			Name("DataProcessingService").
			Description("Advanced data processing service").
			Method("processData", processDataFunc).
			Tag("data").
			Tag("processing").
			Tag("advanced").
			Build()

		// Test the complex schema
		if len(schema.Methods()) != 1 {
			t.Errorf("Expected 1 method, got %d", len(schema.Methods()))
		}

		method := schema.Methods()[0]
		funcSchema := method.Function()

		// Test complex input validation
		validData := map[string]any{
			"data": []any{
				map[string]any{
					"id":    1,
					"value": 10.5,
					"metadata": map[string]any{
						"source": "sensor1",
					},
				},
				map[string]any{
					"id":    2,
					"value": 20.3,
				},
			},
			"options": map[string]any{
				"sortBy": "value",
				"limit":  100,
			},
		}

		result := funcSchema.Validate(validData)
		if !result.Valid {
			t.Errorf("Expected validation to pass for complex valid data, got errors: %v", result.Errors)
		}
	})

	t.Run("Service with error handling methods", func(t *testing.T) {
		errorSchema := builders.NewObject().
			Property("code", builders.NewStringSchema().Build()).
			Property("message", builders.NewStringSchema().Build()).
			Property("details", builders.NewObject().AdditionalProperties(true).Build()).
			Required("code", "message").
			Build()

		riskyFunc := builders.NewFunctionSchema().
			Name("riskyOperation").
			Input("data", builders.NewStringSchema().Build()).
			Output("result", builders.NewStringSchema().Build()).
			Error(errorSchema).
			RequiredInputs("data").
			RequiredOutputs("result").
			Build()

		schema := builders.NewServiceSchema().
			Name("RiskyService").
			Method("riskyOperation", riskyFunc).
			Build()

		method := schema.Methods()[0]
		funcSchema := method.Function()

		if funcSchema.Errors() == nil {
			t.Error("Expected error schema to be defined for risky operation")
		}

		if funcSchema.Errors().Type() != core.TypeObject {
			t.Errorf("Expected error schema to be object type, got %s", funcSchema.Errors().Type())
		}
	})
}
