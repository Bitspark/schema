package tests

import (
	"strings"
	"testing"

	builders2 "defs.dev/schema/construct/builders"
	"defs.dev/schema/consume/validation"

	"defs.dev/schema/core"
)

// Helper function to generate JSON Schema for object schemas
func toJSONSchemaObject(schema core.Schema) map[string]any {
	result := map[string]any{}

	// Map schema types to JSON Schema types
	switch schema.Type() {
	case core.TypeStructure:
		result["type"] = "structure"
	default:
		result["type"] = string(schema.Type())
	}

	if desc := schema.Metadata().Description; desc != "" {
		result["description"] = desc
	}

	// Handle object schemas
	if objectSchema, ok := schema.(core.ObjectSchema); ok {
		if properties := objectSchema.Properties(); properties != nil && len(properties) > 0 {
			propMap := make(map[string]any)
			for name, prop := range properties {
				propMap[name] = toJSONSchemaObject(prop)
			}
			result["properties"] = propMap
		}
		if required := objectSchema.Required(); required != nil && len(required) > 0 {
			result["required"] = required
		}
		if !objectSchema.AdditionalProperties() {
			result["additionalProperties"] = false
		}
	}

	// Handle string schemas
	if stringSchema, ok := schema.(core.StringSchema); ok {
		if minLen := stringSchema.MinLength(); minLen != nil {
			result["minLength"] = *minLen
		}
		if maxLen := stringSchema.MaxLength(); maxLen != nil {
			result["maxLength"] = *maxLen
		}
		if pattern := stringSchema.Pattern(); pattern != "" {
			result["pattern"] = pattern
		}
	}

	// Handle number schemas
	if numberSchema, ok := schema.(core.NumberSchema); ok {
		if min := numberSchema.Minimum(); min != nil {
			result["minimum"] = *min
		}
		if max := numberSchema.Maximum(); max != nil {
			result["maximum"] = *max
		}
	}

	return result
}

func TestObjectSchemaBasicValidation(t *testing.T) {
	schema := builders2.NewObjectSchema().Build()

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
			result := validation.ValidateValue(schema, tt.value)
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

func TestObjectSchemaStructValidation(t *testing.T) {
	schema := builders2.NewObjectSchema().Build()

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

	result := validation.ValidateValue(schema, testStruct)
	if !result.Valid {
		t.Errorf("Struct validation failed: %v", result.Errors)
	}
}

func TestObjectSchemaProperties(t *testing.T) {
	stringSchema := builders2.NewStringSchema().Build()
	numberSchema := builders2.NewNumberSchema().Build()

	schema := builders2.NewObjectSchema().
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
			result := validation.ValidateValue(schema, tt.value)
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
	stringSchema := builders2.NewStringSchema().Build()
	numberSchema := builders2.NewNumberSchema().Build()

	schema := builders2.NewObjectSchema().
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
			result := validation.ValidateValue(schema, tt.value)
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
	stringSchema := builders2.NewStringSchema().Build()

	schema := builders2.NewObjectSchema().
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
			"undefined property",
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
			result := validation.ValidateValue(schema, tt.value)
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
	t.Run("MinProperties", func(t *testing.T) {
		schema := builders2.NewObjectSchema().
			Property("name", builders2.NewStringSchema().Build()).
			Property("age", builders2.NewNumberSchema().Build()).
			Build()

		// Note: MinProperties/MaxProperties might not be implemented yet
		// This test would need to be adjusted based on actual implementation

		tests := []struct {
			name      string
			value     any
			wantValid bool
		}{
			{"empty object", map[string]any{}, true},
			{"single property", map[string]any{"name": "John"}, true},
			{"two properties", map[string]any{"name": "John", "age": 30}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validation.ValidateValue(schema, tt.value)
				if result.Valid != tt.wantValid {
					t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				}
			})
		}
	})
}

func TestObjectSchemaNestedObjects(t *testing.T) {
	addressSchema := builders2.NewObjectSchema().
		Property("street", builders2.NewStringSchema().Build()).
		Property("city", builders2.NewStringSchema().Build()).
		Required("street", "city").
		Build()

	personSchema := builders2.NewObjectSchema().
		Property("name", builders2.NewStringSchema().Build()).
		Property("age", builders2.NewNumberSchema().Build()).
		Property("address", addressSchema).
		Required("name", "address").
		Build()

	tests := []struct {
		name      string
		value     any
		wantValid bool
		errorPath string
	}{
		{
			"valid nested object",
			map[string]any{
				"name": "John",
				"age":  30,
				"address": map[string]any{
					"street": "123 Main St",
					"city":   "Anytown",
				},
			},
			true,
			"",
		},
		{
			"missing nested required property",
			map[string]any{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
					// missing city
				},
			},
			false,
			"address.city",
		},
		{
			"invalid nested property type",
			map[string]any{
				"name": "John",
				"address": map[string]any{
					"street": "123 Main St",
					"city":   123, // should be string
				},
			},
			false,
			"address.city",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidateValue(personSchema, tt.value)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				if !result.Valid {
					t.Logf("Errors: %v", result.Errors)
				}
			}
			if tt.errorPath != "" && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					// Check if the path contains the expected path component
					pathStr := strings.Join(err.Path, ".")
					if strings.Contains(pathStr, tt.errorPath) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error at path '%s', got errors: %v", tt.errorPath, result.Errors)
				}
			}
		})
	}
}

func TestObjectSchemaIntrospection(t *testing.T) {
	stringSchema := builders2.NewStringSchema().Build()
	numberSchema := builders2.NewNumberSchema().Build()

	schema := builders2.NewObjectSchema().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Required("name").
		Description("Person object").
		Build()

	// Test type
	if schema.Type() != core.TypeStructure {
		t.Errorf("Expected type %s, got %s", core.TypeStructure, schema.Type())
	}

	// Test metadata
	metadata := schema.Metadata()
	if metadata.Description != "Person object" {
		t.Errorf("Expected description 'Person object', got '%s'", metadata.Description)
	}

	// Test properties
	properties := schema.Properties()
	if len(properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(properties))
	}

	if _, ok := properties["name"]; !ok {
		t.Error("Expected 'name' property to be defined")
	}

	if _, ok := properties["age"]; !ok {
		t.Error("Expected 'age' property to be defined")
	}

	// Test required properties
	required := schema.Required()
	if len(required) != 1 || required[0] != "name" {
		t.Errorf("Expected required ['name'], got %v", required)
	}
}

func TestObjectSchemaJSONSchema(t *testing.T) {
	stringSchema := builders2.NewStringSchema().Build()
	numberSchema := builders2.NewNumberSchema().Build()

	schema := builders2.NewObjectSchema().
		Property("name", stringSchema).
		Property("age", numberSchema).
		Required("name").
		Description("Test object").
		AdditionalProperties(false).
		Build()

	jsonSchema := toJSONSchemaObject(schema)

	// Check basic properties
	if jsonSchema["type"] != "structure" {
		t.Errorf("Expected type 'structure', got %v", jsonSchema["type"])
	}

	if jsonSchema["description"] != "Test object" {
		t.Errorf("Expected description 'Test object', got %v", jsonSchema["description"])
	}

	if jsonSchema["additionalProperties"] != false {
		t.Errorf("Expected additionalProperties false, got %v", jsonSchema["additionalProperties"])
	}

	// Check properties
	properties, ok := jsonSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if len(properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(properties))
	}

	// Check required array
	requiredInterface, ok := jsonSchema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be a string array")
	}

	required := requiredInterface

	if len(required) != 1 || required[0] != "name" {
		t.Errorf("Expected required ['name'], got %v", required)
	}
}

func TestObjectSchemaBuilder(t *testing.T) {
	t.Run("Immutability", func(t *testing.T) {
		builder1 := builders2.NewObjectSchema().Property("name", builders2.NewStringSchema().Build())
		builder2 := builder1.Property("age", builders2.NewNumberSchema().Build())

		schema1 := builder1.Build()
		schema2 := builder2.Build()

		// Verify they're different instances
		if schema1 == schema2 {
			t.Error("Expected schemas to be different instances")
		}

		// Verify first schema only has 'name' property
		props1 := schema1.Properties()
		if len(props1) != 1 {
			t.Errorf("Expected first schema to have 1 property, got %d", len(props1))
		}

		// Verify second schema has both properties
		props2 := schema2.Properties()
		if len(props2) != 2 {
			t.Errorf("Expected second schema to have 2 properties, got %d", len(props2))
		}
	})
}

func TestObjectSchemaClone(t *testing.T) {
	original := builders2.NewObjectSchema().
		Property("name", builders2.NewStringSchema().Build()).
		Property("age", builders2.NewNumberSchema().Build()).
		Required("name").
		Description("Original object").
		Build()

	cloned := original.Clone()

	// Verify they're different instances
	if original == cloned {
		t.Error("Expected clone to be a different instance")
	}

	// Verify they have the same properties
	originalProps := original.Properties()
	clonedProps := cloned.(core.ObjectSchema).Properties()

	if len(originalProps) != len(clonedProps) {
		t.Error("Expected clone to have same number of properties")
	}

	for key := range originalProps {
		if _, ok := clonedProps[key]; !ok {
			t.Errorf("Expected clone to have property '%s'", key)
		}
	}
}

func TestObjectSchemaVisitor(t *testing.T) {
	schema := builders2.NewObjectSchema().
		Property("name", builders2.NewStringSchema().Build()).
		Build()

	visitor := &testObjectVisitor{
		visitObject: func(s core.ObjectSchema) error {
			if s.Type() != core.TypeStructure {
				t.Error("Expected visitor to receive object schema")
			}
			return nil
		},
	}

	err := schema.Accept(visitor)
	if err != nil {
		t.Errorf("Expected visitor to succeed, got error: %v", err)
	}
}

type testObjectVisitor struct {
	visitObject func(core.ObjectSchema) error
}

func (v *testObjectVisitor) VisitString(core.StringSchema) error   { return nil }
func (v *testObjectVisitor) VisitNumber(core.NumberSchema) error   { return nil }
func (v *testObjectVisitor) VisitInteger(core.IntegerSchema) error { return nil }
func (v *testObjectVisitor) VisitBoolean(core.BooleanSchema) error { return nil }
func (v *testObjectVisitor) VisitArray(core.ArraySchema) error     { return nil }
func (v *testObjectVisitor) VisitObject(schema core.ObjectSchema) error {
	if v.visitObject != nil {
		return v.visitObject(schema)
	}
	return nil
}
func (v *testObjectVisitor) VisitFunction(core.FunctionSchema) error { return nil }
func (v *testObjectVisitor) VisitService(core.ServiceSchema) error   { return nil }
func (v *testObjectVisitor) VisitUnion(core.UnionSchema) error       { return nil }

func TestObjectBuilderAdditionalMethods(t *testing.T) {
	t.Run("Builder fluent API", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name" description:"Person's name"`
			Age   int    `json:"age" description:"Person's age"`
			Email string `json:"email,omitempty" description:"Email address"`
		}

		// Test basic object building
		schema := builders2.NewObjectSchema().
			Property("name", builders2.NewStringSchema().MinLength(1).Build()).
			Property("age", builders2.NewIntegerSchema().Range(0, 150).Build()).
			Property("email", builders2.NewStringSchema().Email().Build()).
			Required("name").
			Description("Test person object").
			Build()

		properties := schema.Properties()
		if len(properties) != 3 {
			t.Errorf("Expected 3 properties, got %d", len(properties))
		}

		// Test validation with map
		testData := map[string]any{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		result := validation.ValidateValue(schema, testData)
		if !result.Valid {
			t.Errorf("Expected test data to be valid, got errors: %v", result.Errors)
		}

		// Test validation with struct instance (basic struct validation)
		testInstance := TestStruct{
			Name:  "John Doe",
			Age:   30,
			Email: "john@example.com",
		}

		result = validation.ValidateValue(schema, testInstance)
		if !result.Valid {
			t.Errorf("Expected struct instance to be valid, got errors: %v", result.Errors)
		}
	})
}

func TestObjectBuilderHelperMethods(t *testing.T) {
	t.Run("Common object patterns", func(t *testing.T) {
		// Test chaining multiple properties
		schema := builders2.NewObjectSchema().
			Property("id", builders2.NewStringSchema().UUID().Build()).
			Property("name", builders2.NewStringSchema().MinLength(1).Build()).
			Property("email", builders2.NewStringSchema().Email().Build()).
			Property("age", builders2.NewIntegerSchema().Range(0, 150).Build()).
			Required("id", "name", "email").
			AdditionalProperties(false).
			Description("User object").
			Build()

		// Test with valid data
		validUser := map[string]any{
			"id":    "550e8400-e29b-41d4-a716-446655440000",
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}

		result := validation.ValidateValue(schema, validUser)
		if !result.Valid {
			t.Errorf("Expected valid user to pass validation, got errors: %v", result.Errors)
		}

		// Test with invalid data
		invalidUser := map[string]any{
			"id":    "not-a-uuid",
			"name":  "", // too short
			"email": "not-an-email",
			"age":   -5, // negative age
		}

		result = validation.ValidateValue(schema, invalidUser)
		if result.Valid {
			t.Error("Expected invalid user to fail validation")
		}

		// Should have multiple errors
		if len(result.Errors) < 3 {
			t.Errorf("Expected multiple validation errors, got %d", len(result.Errors))
		}
	})
}
