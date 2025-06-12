package annotation

import (
	"testing"

	"defs.dev/schema/schemas"
)

func TestAnnotationRegistry_BasicOperations(t *testing.T) {
	registry := NewRegistry()

	// Test registration
	stringSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{})
	err := registry.RegisterType("test", stringSchema,
		WithDescription("Test annotation"),
		WithCategory("test"),
		WithTags("test", "example"),
	)
	if err != nil {
		t.Fatalf("Failed to register annotation type: %v", err)
	}

	// Test type retrieval
	annotationType, exists := registry.GetType("test")
	if !exists {
		t.Fatal("Annotation type not found after registration")
	}

	if annotationType.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", annotationType.Name())
	}

	// Test listing types
	types := registry.ListTypes()
	if len(types) != 1 || types[0] != "test" {
		t.Errorf("Expected types ['test'], got %v", types)
	}

	// Test has type
	if !registry.HasType("test") {
		t.Error("HasType should return true for registered type")
	}

	if registry.HasType("nonexistent") {
		t.Error("HasType should return false for unregistered type")
	}
}

func TestAnnotationRegistry_CreateAnnotation(t *testing.T) {
	registry := NewRegistry()

	// Register a string annotation type
	stringSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{})
	err := registry.RegisterType("format", stringSchema,
		WithDescription("String format"),
		WithCategory("string"),
	)
	if err != nil {
		t.Fatalf("Failed to register annotation type: %v", err)
	}

	// Test creating valid annotation
	annotation, err := registry.Create("format", "email")
	if err != nil {
		t.Fatalf("Failed to create annotation: %v", err)
	}

	if annotation.Name() != "format" {
		t.Errorf("Expected name 'format', got '%s'", annotation.Name())
	}

	if annotation.Value() != "email" {
		t.Errorf("Expected value 'email', got '%v'", annotation.Value())
	}

	// Test validation
	result := annotation.Validate()
	if !result.Valid {
		t.Errorf("Valid annotation should pass validation: %v", result.Errors)
	}
}

func TestAnnotationRegistry_StrictMode(t *testing.T) {
	registry := NewRegistry()

	// Test non-strict mode (default)
	if registry.IsStrictMode() {
		t.Error("Registry should start in non-strict mode")
	}

	// Should allow unknown annotation types in non-strict mode
	annotation, err := registry.Create("unknown", "value")
	if err != nil {
		t.Errorf("Non-strict mode should allow unknown types: %v", err)
	}

	if annotation == nil {
		t.Error("Should create flexible annotation for unknown type")
	}

	// Enable strict mode
	registry.SetStrictMode(true)
	if !registry.IsStrictMode() {
		t.Error("Strict mode should be enabled")
	}

	// Should reject unknown annotation types in strict mode
	_, err = registry.Create("another_unknown", "value")
	if err == nil {
		t.Error("Strict mode should reject unknown annotation types")
	}
}

func TestAnnotationRegistry_BulkOperations(t *testing.T) {
	registry := NewRegistry()

	// Register annotation types
	stringSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{})
	intSchema := schemas.NewIntegerSchema(schemas.IntegerSchemaConfig{})

	registry.RegisterType("format", stringSchema)
	registry.RegisterType("minLength", intSchema)

	// Test creating many annotations
	annotations := map[string]any{
		"format":    "email",
		"minLength": 5,
	}

	result, err := registry.CreateMany(annotations)
	if err != nil {
		t.Fatalf("Failed to create many annotations: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(result))
	}

	// Test validating many annotations
	validationResult := registry.ValidateMany(result)
	if !validationResult.Valid {
		t.Errorf("All annotations should be valid: %v", validationResult.Errors)
	}
}

func TestBuiltinAnnotations(t *testing.T) {
	registry := NewRegistry()

	// Register built-in annotation types
	err := RegisterBuiltinTypes(registry)
	if err != nil {
		t.Fatalf("Failed to register built-in types: %v", err)
	}

	// Test that all expected built-in types are registered
	expectedTypes := []string{
		"format", "pattern", "minLength", "maxLength",
		"min", "max", "range",
		"minItems", "maxItems", "uniqueItems",
		"required", "validators",
		"description", "examples", "default", "enum",
	}

	registeredTypes := registry.ListTypes()
	typeSet := make(map[string]bool)
	for _, t := range registeredTypes {
		typeSet[t] = true
	}

	for _, expectedType := range expectedTypes {
		if !typeSet[expectedType] {
			t.Errorf("Expected built-in type '%s' not found in registry", expectedType)
		}
	}

	// Test creating annotations with built-in types
	testCases := []struct {
		name  string
		value any
	}{
		{"format", "email"},
		{"minLength", 5},
		{"pattern", "^[a-zA-Z0-9]+$"},
		{"required", true},
		{"description", "Test field"},
	}

	for _, tc := range testCases {
		annotation, err := registry.Create(tc.name, tc.value)
		if err != nil {
			t.Errorf("Failed to create annotation '%s': %v", tc.name, err)
			continue
		}

		result := annotation.Validate()
		if !result.Valid {
			t.Errorf("Built-in annotation '%s' should validate: %v", tc.name, result.Errors)
		}
	}
}

func TestAnnotationTypeOptions(t *testing.T) {
	registry := NewRegistry()

	stringSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{})
	err := registry.RegisterType("test", stringSchema,
		WithDescription("Test annotation with options"),
		WithCategory("test"),
		WithTags("tag1", "tag2"),
		WithAppliesTo("string", "object"),
		WithExamples("example1", "example2"),
	)
	if err != nil {
		t.Fatalf("Failed to register annotation type with options: %v", err)
	}

	annotationType, exists := registry.GetType("test")
	if !exists {
		t.Fatal("Annotation type not found")
	}

	metadata := annotationType.Metadata()
	if metadata.Description != "Test annotation with options" {
		t.Errorf("Expected description 'Test annotation with options', got '%s'", metadata.Description)
	}

	if metadata.Category != "test" {
		t.Errorf("Expected category 'test', got '%s'", metadata.Category)
	}

	if len(metadata.Tags) != 2 || metadata.Tags[0] != "tag1" || metadata.Tags[1] != "tag2" {
		t.Errorf("Expected tags ['tag1', 'tag2'], got %v", metadata.Tags)
	}

	if len(metadata.Examples) != 2 {
		t.Errorf("Expected 2 examples, got %d", len(metadata.Examples))
	}
}
