package tests

import (
	"defs.dev/schema"
	"defs.dev/schema/builders"
	"testing"

	"defs.dev/schema/api"
)

func TestObjectSchemaBasicValidation(t *testing.T) {
	schema := schema.NewObject().Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantError string
	}{
		{"empty object", map[string]any{}, true, ""},
		{"simple object", map[string]any{"key": "value"}, true, ""},
		{"nested object", map[string]any{"nested": map[string]any{"key": "value"}}, true, ""},
		{"nil value", nil, false, "Expected object or map"},
		{"string value", "not an object", false, "Expected object or map"},
		{"number value", 42, false, "Expected object or map"},
		{"array value", []any{1, 2, 3}, false, "Expected object or map"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantError != "" && (len(result.Errors) == 0 || result.Errors[0].Message != tt.wantError) {
				t.Errorf("Validate() error = %v, want %v", result.Errors, tt.wantError)
			}
		})
	}
}

func TestObjectSchemaStructValidation(t *testing.T) {
	schema := schema.NewObject().Build()

	// Test struct validation
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	testStruct := TestStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	result := schema.Validate(testStruct)
	if !result.Valid {
		t.Errorf("Struct validation failed: %v", result.Errors)
	}
}

func TestObjectSchemaProperties(t *testing.T) {
	stringSchema := schema.NewString().Build()
	numberSchema := schema.NewNumber().Build()

	schema := schema.NewObject().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantError string
	}{
		{
			"valid properties",
			map[string]any{"name": "John", "age": 30.0},
			true,
			"",
		},
		{
			"invalid property type",
			map[string]any{"name": "John", "age": "thirty"},
			false,
			"Expected number",
		},
		{
			"missing properties (optional)",
			map[string]any{},
			true,
			"",
		},
		{
			"extra properties (allowed by default)",
			map[string]any{"name": "John", "age": 30.0, "extra": "value"},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Message == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s', got: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestObjectSchemaRequired(t *testing.T) {
	stringSchema := schema.NewString().Build()
	numberSchema := schema.NewNumber().Build()

	schema := schema.NewObject().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Required("name").
		Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantError string
	}{
		{
			"has required property",
			map[string]any{"name": "John"},
			true,
			"",
		},
		{
			"missing required property",
			map[string]any{"age": 30},
			false,
			"Missing required property 'name'",
		},
		{
			"has all properties",
			map[string]any{"name": "John", "age": 30},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Message == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s', got: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestObjectSchemaAdditionalProperties(t *testing.T) {
	stringSchema := schema.NewString().Build()

	schema := schema.NewObject().
		Property("name", stringSchema).
		AdditionalProperties(false).
		Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantError string
	}{
		{
			"defined property only",
			map[string]any{"name": "John"},
			true,
			"",
		},
		{
			"additional property not allowed",
			map[string]any{"name": "John", "extra": "value"},
			false,
			"Additional property 'extra' is not allowed",
		},
		{
			"empty object",
			map[string]any{},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Message == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s', got: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestObjectSchemaConstraints(t *testing.T) {
	// Create builder with additional methods that return *ObjectBuilder
	builder := builders.NewObjectSchema()

	schema := builder.
		MinProperties(1).
		MaxProperties(3).
		Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantError string
	}{
		{
			"within range",
			map[string]any{"key1": "value1", "key2": "value2"},
			true,
			"",
		},
		{
			"too few properties",
			map[string]any{},
			false,
			"Object has too few properties (minimum 1)",
		},
		{
			"too many properties",
			map[string]any{"key1": "v1", "key2": "v2", "key3": "v3", "key4": "v4"},
			false,
			"Object has too many properties (maximum 3)",
		},
		{
			"exactly minimum",
			map[string]any{"key1": "value1"},
			true,
			"",
		},
		{
			"exactly maximum",
			map[string]any{"key1": "v1", "key2": "v2", "key3": "v3"},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Message == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s', got: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestObjectSchemaNestedObjects(t *testing.T) {
	// Create nested object schema
	addressSchema := schema.NewObject().
		Property("street", schema.NewString().Build()).
		Property("city", schema.NewString().Build()).
		Required("street", "city").
		Build()

	personSchema := schema.NewObject().
		Property("name", schema.NewString().Build()).
		Property("address", addressSchema).
		Required("name").
		Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		wantError string
	}{
		{
			"valid nested object",
			map[string]any{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
					"city":   "Anytown",
				},
			},
			true,
			"",
		},
		{
			"missing nested required field",
			map[string]any{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
				},
			},
			false,
			"Missing required property 'city'",
		},
		{
			"invalid nested field type",
			map[string]any{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
					"city":   123, // Should be string
				},
			},
			false,
			"Expected string",
		},
		{
			"person without address",
			map[string]any{
				"name": "John",
			},
			true,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := personSchema.Validate(tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				for _, err := range result.Errors {
					t.Logf("Error: %s", err.Message)
				}
			}
			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Message == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s', got: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestObjectSchemaIntrospection(t *testing.T) {
	stringSchema := schema.NewString().Build()
	numberSchema := schema.NewNumber().Build()

	schema := schema.NewObject().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Required("name").
		AdditionalProperties(false).
		Build()

	// Test Properties
	properties := schema.Properties()
	if len(properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(properties))
	}
	if properties["name"] == nil {
		t.Errorf("Expected 'name' property to exist")
	}
	if properties["age"] == nil {
		t.Errorf("Expected 'age' property to exist")
	}

	// Test Required
	required := schema.Required()
	if len(required) != 1 {
		t.Errorf("Expected 1 required property, got %d", len(required))
	}
	if required[0] != "name" {
		t.Errorf("Expected required property 'name', got %s", required[0])
	}

	// Test AdditionalProperties
	if schema.AdditionalProperties() != false {
		t.Errorf("Expected AdditionalProperties to be false")
	}
}

func TestObjectSchemaJSONSchema(t *testing.T) {
	stringSchema := schema.NewString().Build()
	numberSchema := schema.NewNumber().Min(0).Build()

	schema := schema.NewObject().
		Description("A person object").
		Property("name", stringSchema).
		Property("age", numberSchema).
		Required("name").
		AdditionalProperties(false).
		Build()

	jsonSchema := schema.ToJSONSchema()

	// Check basic structure
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", jsonSchema["type"])
	}

	if jsonSchema["description"] != "A person object" {
		t.Errorf("Expected description 'A person object', got %v", jsonSchema["description"])
	}

	// Check properties
	properties, ok := jsonSchema["properties"].(map[string]any)
	if !ok {
		t.Errorf("Expected properties to be a map")
	} else {
		if len(properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(properties))
		}
	}

	// Check required
	required, ok := jsonSchema["required"].([]string)
	if !ok {
		t.Errorf("Expected required to be a string slice")
	} else {
		if len(required) != 1 || required[0] != "name" {
			t.Errorf("Expected required ['name'], got %v", required)
		}
	}

	// Check additionalProperties
	if jsonSchema["additionalProperties"] != false {
		t.Errorf("Expected additionalProperties false, got %v", jsonSchema["additionalProperties"])
	}
}

func TestObjectSchemaExampleGeneration(t *testing.T) {
	stringSchema := schema.NewString().Build()
	numberSchema := schema.NewNumber().Build()

	schema := schema.NewObject().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Required("name").
		Build()

	example := schema.GenerateExample()
	exampleMap, ok := example.(map[string]any)
	if !ok {
		t.Errorf("Expected example to be a map, got %T", example)
	}

	// Should have required property
	if _, exists := exampleMap["name"]; !exists {
		t.Errorf("Expected example to have required 'name' property")
	}

	// Test with explicit example
	explicitExample := map[string]any{
		"name": "John Doe",
		"age":  30,
	}

	schemaWithExample := schema.NewObject().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Example(explicitExample).
		Build()

	generatedExample := schemaWithExample.GenerateExample()
	generatedMap, ok := generatedExample.(map[string]any)
	if !ok {
		t.Errorf("Expected generated example to be a map")
	} else {
		// Check that the generated example has the expected values
		if generatedMap["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got %v", generatedMap["name"])
		}
		if generatedMap["age"] != 30 {
			t.Errorf("Expected age 30, got %v", generatedMap["age"])
		}
	}
}

func TestObjectSchemaBuilder(t *testing.T) {
	// Test fluent interface
	schema := schema.NewObject().
		Description("Test object").
		Property("name", schema.NewString().Build()).
		Property("age", schema.NewNumber().Build()).
		Required("name").
		AdditionalProperties(false).
		Example(map[string]any{"name": "John", "age": 30}).
		Build()

	// Test that schema works
	result := schema.Validate(map[string]any{"name": "John", "age": 30})
	if !result.Valid {
		t.Errorf("Schema validation failed: %v", result.Errors)
	}

	// Test metadata
	metadata := schema.Metadata()
	if metadata.Description != "Test object" {
		t.Errorf("Expected description 'Test object', got %s", metadata.Description)
	}

	if len(metadata.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(metadata.Examples))
	}
}

func TestObjectSchemaClone(t *testing.T) {
	original := schema.NewObject().
		Property("name", schema.NewString().Build()).
		Required("name").
		Build()

	cloned := original.Clone()

	// Verify they're different instances
	if original == cloned {
		t.Errorf("Expected cloned schema to be a different instance")
	}

	// Verify they have the same behavior
	testValue := map[string]any{"name": "John"}

	originalResult := original.Validate(testValue)
	clonedResult := cloned.Validate(testValue)

	if originalResult.Valid != clonedResult.Valid {
		t.Errorf("Original and cloned schemas have different validation results")
	}
}

func TestObjectSchemaVisitor(t *testing.T) {
	schema := schema.NewObject().Build()

	// Simple visitor implementation for testing
	var visited bool
	visitor := &testObjectVisitor{
		visitObject: func(api.ObjectSchema) error {
			visited = true
			return nil
		},
	}

	err := schema.Accept(visitor)
	if err != nil {
		t.Errorf("Accept() error = %v", err)
	}

	if !visited {
		t.Errorf("Expected visitor to be called")
	}
}

// testObjectVisitor implements api.SchemaVisitor for testing
type testObjectVisitor struct {
	visitObject func(api.ObjectSchema) error
}

func (v *testObjectVisitor) VisitString(api.StringSchema) error   { return nil }
func (v *testObjectVisitor) VisitNumber(api.NumberSchema) error   { return nil }
func (v *testObjectVisitor) VisitInteger(api.IntegerSchema) error { return nil }
func (v *testObjectVisitor) VisitBoolean(api.BooleanSchema) error { return nil }
func (v *testObjectVisitor) VisitArray(api.ArraySchema) error     { return nil }
func (v *testObjectVisitor) VisitObject(schema api.ObjectSchema) error {
	if v.visitObject != nil {
		return v.visitObject(schema)
	}
	return nil
}
func (v *testObjectVisitor) VisitFunction(api.FunctionSchema) error { return nil }
func (v *testObjectVisitor) VisitService(api.ServiceSchema) error   { return nil }
func (v *testObjectVisitor) VisitUnion(api.UnionSchema) error       { return nil }

func TestObjectBuilderAdditionalMethods(t *testing.T) {
	// Test additional methods that return *ObjectBuilder
	builder := builders.NewObjectSchema()

	schema := builder.
		RequiredProperty("name", schema.NewString().Build()).
		OptionalProperty("email", schema.NewString().Email().Build()).
		MinProperties(1).
		MaxProperties(10).
		Strict().
		Build()

	// Test that it works
	result := schema.Validate(map[string]any{"name": "John"})
	if !result.Valid {
		t.Errorf("Schema validation failed: %v", result.Errors)
	}

	// Test strict mode (no additional properties)
	result = schema.Validate(map[string]any{"name": "John", "extra": "not allowed"})
	if result.Valid {
		t.Errorf("Expected validation to fail with additional property")
	}
}

func TestObjectBuilderHelperMethods(t *testing.T) {
	builder := builders.NewObjectSchema()

	// Test helper methods
	schema := builder.
		PersonExample().
		Build()

	metadata := schema.Metadata()
	if metadata.Description != "Person information" {
		t.Errorf("Expected description 'Person information', got %s", metadata.Description)
	}

	if len(metadata.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(metadata.Examples))
	}

	// Test API response example
	apiSchema := builders.NewObjectSchema().
		APIResponseExample().
		Build()

	apiMetadata := apiSchema.Metadata()
	if apiMetadata.Description != "API response structure" {
		t.Errorf("Expected description 'API response structure', got %s", apiMetadata.Description)
	}
}
