package tests

import (
	"defs.dev/schema"
	"testing"

	"defs.dev/schema/api"
)

func TestStringSchemaBasic(t *testing.T) {
	// Create a string schema
	schema := schema.NewString().
		MinLength(3).
		MaxLength(10).
		Description("Test string").
		Build()

	// Verify it implements the correct interfaces
	if _, ok := schema.(api.Schema); !ok {
		t.Error("StringSchema should implement api.Schema")
	}

	if _, ok := schema.(api.StringSchema); !ok {
		t.Error("StringSchema should implement api.StringSchema")
	}

	// Test basic properties
	if schema.Type() != api.TypeString {
		t.Errorf("Expected type %s, got %s", api.TypeString, schema.Type())
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
	schema := schema.NewString().
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
	schema := schema.NewString().
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
	schema := schema.NewString().Email().Build()

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
	schema := schema.NewString().
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
	builder1 := schema.NewString().MinLength(3)
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
	schema := schema.NewString().
		MinLength(3).
		MaxLength(10).
		Pattern("^[a-z]+$").
		Description("Test string").
		Example("hello").
		Build()

	jsonSchema := schema.ToJSONSchema()

	// Check basic properties
	if jsonSchema["type"] != "string" {
		t.Errorf("Expected type 'string', got %v", jsonSchema["type"])
	}

	if jsonSchema["minLength"] != 3 {
		t.Errorf("Expected minLength 3, got %v", jsonSchema["minLength"])
	}

	if jsonSchema["maxLength"] != 10 {
		t.Errorf("Expected maxLength 10, got %v", jsonSchema["maxLength"])
	}

	if jsonSchema["pattern"] != "^[a-z]+$" {
		t.Errorf("Expected pattern '^[a-z]+$', got %v", jsonSchema["pattern"])
	}

	if jsonSchema["description"] != "Test string" {
		t.Errorf("Expected description 'Test string', got %v", jsonSchema["description"])
	}
}
