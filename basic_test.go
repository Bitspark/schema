package schema

import (
	"testing"
)

func TestStringSchemaWithMetadata(t *testing.T) {
	original := String().Build().(*StringSchema)
	
	metadata := SchemaMetadata{
		Name:        "test-string",
		Description: "Test string schema",
		Tags:        []string{"test", "string"},
	}
	
	result := original.WithMetadata(metadata)
	resultString := result.(*StringSchema)
	
	// Verify original is not modified
	if original.metadata.Name == "test-string" {
		t.Error("Original schema was modified, expected clone")
	}
	
	// Verify clone has the metadata
	if resultString.metadata.Name != "test-string" {
		t.Errorf("Expected name 'test-string', got %s", resultString.metadata.Name)
	}
	
	if resultString.metadata.Description != "Test string schema" {
		t.Errorf("Expected description 'Test string schema', got %s", resultString.metadata.Description)
	}
	
	if len(resultString.metadata.Tags) != 2 || resultString.metadata.Tags[0] != "test" {
		t.Errorf("Expected tags [test, string], got %v", resultString.metadata.Tags)
	}
}

func TestObjectSchemaWithMetadata(t *testing.T) {
	original := Object().Build().(*ObjectSchema)
	
	metadata := SchemaMetadata{
		Name:        "test-object",
		Description: "Test object schema",
	}
	
	result := original.WithMetadata(metadata)
	resultObject := result.(*ObjectSchema)
	
	// Verify original is not modified
	if original.metadata.Name == "test-object" {
		t.Error("Original schema was modified, expected clone")
	}
	
	// Verify clone has the metadata
	if resultObject.metadata.Name != "test-object" {
		t.Errorf("Expected name 'test-object', got %s", resultObject.metadata.Name)
	}
}

func TestNumberSchemaWithMetadata(t *testing.T) {
	original := Number().Build().(*NumberSchema)
	
	metadata := SchemaMetadata{
		Name: "test-number",
	}
	
	result := original.WithMetadata(metadata)
	resultNumber := result.(*NumberSchema)
	
	// Verify original is not modified
	if original.metadata.Name == "test-number" {
		t.Error("Original schema was modified, expected clone")
	}
	
	// Verify clone has the metadata
	if resultNumber.metadata.Name != "test-number" {
		t.Errorf("Expected name 'test-number', got %s", resultNumber.metadata.Name)
	}
}

func TestIntegerSchemaWithMetadata(t *testing.T) {
	original := Integer().Build().(*IntegerSchema)
	
	metadata := SchemaMetadata{
		Name: "test-integer",
	}
	
	result := original.WithMetadata(metadata)
	resultInteger := result.(*IntegerSchema)
	
	// Verify original is not modified
	if original.metadata.Name == "test-integer" {
		t.Error("Original schema was modified, expected clone")
	}
	
	// Verify clone has the metadata
	if resultInteger.metadata.Name != "test-integer" {
		t.Errorf("Expected name 'test-integer', got %s", resultInteger.metadata.Name)
	}
}

func TestGetFormatSuggestion(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"email", "Provide a valid email address like 'user@example.com'"},
		{"uuid", "Provide a valid UUID like '123e4567-e89b-12d3-a456-426614174000'"},
		{"url", "Provide a valid URL like 'https://example.com'"},
		{"custom", "Provide a valid custom format"},
		{"", "Provide a valid  format"},
	}
	
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			result := getFormatSuggestion(test.format)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestGenerateFormatExample(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"email", "user@example.com"},
		{"uuid", "123e4567-e89b-12d3-a456-426614174000"},
		{"url", "https://example.com"},
		{"custom", "example"},
		{"", "example"},
	}
	
	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			result := generateFormatExample(test.format)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}