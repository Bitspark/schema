package tests

import (
	"encoding/json"
	"testing"

	"defs.dev/schema/api/core"
	"defs.dev/schema/builders"
	jsonexport "defs.dev/schema/export/json"
)

// Helper function to generate JSON Schema using the export system
func toJSONSchema(schema core.Schema) map[string]any {
	generator := jsonexport.NewGenerator()
	jsonBytes, err := generator.Generate(schema)
	if err != nil {
		panic(err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		panic(err)
	}

	return result
}

func TestStringSchemaBasic(t *testing.T) {
	// Create a string schema
	schema := builders.NewStringSchema().
		MinLength(3).
		MaxLength(10).
		Description("Test string").
		Build()

	// Verify it implements the correct interfaces
	if _, ok := schema.(core.Schema); !ok {
		t.Error("StringSchema should implement core.Schema")
	}

	// Test basic properties
	if schema.Type() != core.TypeString {
		t.Errorf("Expected type %s, got %s", core.TypeString, schema.Type())
	}

	if schema.Metadata().Description != "Test string" {
		t.Errorf("Expected description 'Test string', got '%s'", schema.Metadata().Description)
	}

	// Test constraints
	if minLen := schema.MinLength(); minLen == nil || *minLen != 3 {
		t.Errorf("Expected MinLength 3, got %v", minLen)
	}

	if maxLen := schema.MaxLength(); maxLen == nil || *maxLen != 10 {
		t.Errorf("Expected MaxLength 10, got %v", maxLen)
	}
}

func TestStringSchemaValidation(t *testing.T) {
	schema := builders.NewStringSchema().
		MinLength(3).
		MaxLength(10).
		Build()

	// Test valid input
	result := schema.Validate("hello")
	if !result.Valid {
		t.Errorf("Expected 'hello' to be valid, got errors: %v", result.Errors)
	}

	// Test too short
	result = schema.Validate("hi")
	if result.Valid {
		t.Error("Expected 'hi' to be invalid (too short)")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected validation errors for 'hi'")
	}

	// Test too long
	result = schema.Validate("this_is_too_long")
	if result.Valid {
		t.Error("Expected 'this_is_too_long' to be invalid (too long)")
	}

	// Test wrong type
	result = schema.Validate(123)
	if result.Valid {
		t.Error("Expected 123 to be invalid (wrong type)")
	}
}

func TestStringSchemaPattern(t *testing.T) {
	schema := builders.NewStringSchema().
		Pattern(`^[a-z]+$`).
		Build()

	// Test valid pattern
	result := schema.Validate("hello")
	if !result.Valid {
		t.Errorf("Expected 'hello' to match pattern, got errors: %v", result.Errors)
	}

	// Test invalid pattern
	result = schema.Validate("Hello123")
	if result.Valid {
		t.Error("Expected 'Hello123' to not match pattern")
	}
}

func TestStringSchemaEmail(t *testing.T) {
	schema := builders.NewStringSchema().Email().Build()

	// Test valid email
	result := schema.Validate("user@example.com")
	if !result.Valid {
		t.Errorf("Expected 'user@example.com' to be valid, got errors: %v", result.Errors)
	}

	// Test invalid email
	result = schema.Validate("not-an-email")
	if result.Valid {
		t.Error("Expected 'not-an-email' to be invalid")
	}
}

func TestStringSchemaEnum(t *testing.T) {
	schema := builders.NewStringSchema().
		Enum("red", "green", "blue").
		Build()

	// Test valid enum value
	result := schema.Validate("red")
	if !result.Valid {
		t.Errorf("Expected 'red' to be valid, got errors: %v", result.Errors)
	}

	// Test invalid enum value
	result = schema.Validate("yellow")
	if result.Valid {
		t.Error("Expected 'yellow' to be invalid")
	}
}

func TestStringSchemaImmutability(t *testing.T) {
	builder1 := builders.NewStringSchema().MinLength(3)
	builder2 := builder1.MaxLength(10)

	schema1 := builder1.Build()
	schema2 := builder2.Build()

	// Verify they're different instances
	if schema1 == schema2 {
		t.Error("Expected schemas to be different instances")
	}

	// Verify first schema doesn't have MaxLength
	if maxLen := schema1.MaxLength(); maxLen != nil {
		t.Error("Expected first schema to not have MaxLength constraint")
	}

	// Verify second schema has both constraints
	if minLen := schema2.MinLength(); minLen == nil || *minLen != 3 {
		t.Error("Expected second schema to have MinLength constraint")
	}
	if maxLen := schema2.MaxLength(); maxLen == nil || *maxLen != 10 {
		t.Error("Expected second schema to have MaxLength constraint")
	}
}

func TestStringSchemaJSONSchema(t *testing.T) {
	schema := builders.NewStringSchema().
		MinLength(3).
		MaxLength(10).
		Pattern("^[a-z]+$").
		Description("Test string").
		Example("hello").
		Build()

	jsonSchema := toJSONSchema(schema)

	// Check basic properties
	if jsonSchema["type"] != "string" {
		t.Errorf("Expected type 'string', got %v", jsonSchema["type"])
	}

	if jsonSchema["minLength"] != float64(3) {
		t.Errorf("Expected minLength 3, got %v", jsonSchema["minLength"])
	}

	if jsonSchema["maxLength"] != float64(10) {
		t.Errorf("Expected maxLength 10, got %v", jsonSchema["maxLength"])
	}

	if jsonSchema["pattern"] != "^[a-z]+$" {
		t.Errorf("Expected pattern '^[a-z]+$', got %v", jsonSchema["pattern"])
	}

	if jsonSchema["description"] != "Test string" {
		t.Errorf("Expected description 'Test string', got %v", jsonSchema["description"])
	}
}

func TestGoModSchema(t *testing.T) {
	// Structural validation - syntax and basic constraints
	goModSchema := builders.NewObject().
		Property("module", builders.NewStringSchema().Pattern(`^[a-zA-Z0-9\-\.\/]+$`).Build()).
		Property("go", builders.NewStringSchema().Pattern(`^\d+\.\d+$`).Build()).
		Property("require", builders.NewArraySchema().Items(
			builders.NewObjectSchema().
				Property("module", builders.NewStringSchema().Build()).
				Property("version", builders.NewStringSchema().Build()).
				Build(),
		).Build()).
		Required("module", "go").
		Build()

	// Verify it implements the correct interfaces
	if _, ok := goModSchema.(core.Schema); !ok {
		t.Error("GoModSchema should implement core.Schema")
	}

	// Test basic properties
	if goModSchema.Type() != core.TypeObject {
		t.Errorf("Expected type %s, got %s", core.TypeObject, goModSchema.Type())
	}

	if goModSchema.Metadata().Description != "" {
		t.Errorf("Expected description '', got '%s'", goModSchema.Metadata().Description)
	}

	// Test required properties
	if required := goModSchema.Required(); len(required) != 2 || !contains(required, "module") || !contains(required, "go") {
		t.Errorf("Expected required properties 'module' and 'go', got %v", required)
	}

	// Test property types
	properties := goModSchema.Properties()
	if module, ok := properties["module"].(core.StringSchema); !ok {
		t.Error("Expected 'module' to be a StringSchema")
	} else {
		if module.Pattern() != `^[a-zA-Z0-9\-\.\/]+$` {
			t.Errorf("Expected 'module' pattern '^[a-zA-Z0-9\\-\\.\\/]+$', got %v", module.Pattern())
		}
	}
	if g, ok := properties["go"].(core.StringSchema); !ok {
		t.Error("Expected 'go' to be a StringSchema")
	} else {
		if g.Pattern() != `^\d+\.\d+$` {
			t.Errorf("Expected 'go' pattern '^\\d+\\.\\d+$', got %v", g.Pattern())
		}
	}

	// Test require property types
	require := goModSchema.Required()
	if len(require) != 2 || !contains(require, "module") || !contains(require, "go") {
		t.Errorf("Expected require properties 'module' and 'go', got %v", require)
	}
}

func contains(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}
