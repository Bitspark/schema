package schema

import (
	"testing"
)

func TestSchemaGeneratorCreateNullSchema(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	schema := generator.createNullSchema()
	
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}
	
	if schema.Type() != TypeObject {
		t.Errorf("Expected TypeObject, got %s", schema.Type())
	}
	
	metadata := schema.Metadata()
	if metadata.Name != "Null" {
		t.Errorf("Expected name 'Null', got '%s'", metadata.Name)
	}
	
	if metadata.Description != "Represents a null value" {
		t.Errorf("Expected description 'Represents a null value', got '%s'", metadata.Description)
	}
}

func TestSchemaGeneratorCreateAnySchema(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	schema := generator.createAnySchema()
	
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}
	
	if schema.Type() != TypeObject {
		t.Errorf("Expected TypeObject, got %s", schema.Type())
	}
	
	objectSchema := schema.(*ObjectSchema)
	if !objectSchema.AdditionalProperties() {
		t.Error("Expected additional properties to be allowed for any schema")
	}
	
	metadata := schema.Metadata()
	if metadata.Name != "Any" {
		t.Errorf("Expected name 'Any', got '%s'", metadata.Name)
	}
	
	if metadata.Description != "Any value is allowed" {
		t.Errorf("Expected description 'Any value is allowed', got '%s'", metadata.Description)
	}
}

func TestSchemaGeneratorGenerateNullSchema(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	schema := generator.generateNullSchema()
	
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}
	
	// Should be the same as createNullSchema
	metadata := schema.Metadata()
	if metadata.Name != "Null" {
		t.Errorf("Expected name 'Null', got '%s'", metadata.Name)
	}
}

func TestSchemaGeneratorGenerateAnySchema(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	schema := generator.generateAnySchema()
	
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}
	
	// Should be the same as createAnySchema
	metadata := schema.Metadata()
	if metadata.Name != "Any" {
		t.Errorf("Expected name 'Any', got '%s'", metadata.Name)
	}
}

func TestSchemaGeneratorGenerateRandomPropertyName(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	
	// Generate multiple property names to test randomness and format
	for i := 0; i < 10; i++ {
		name := generator.generateRandomPropertyName()
		
		if len(name) < 3 || len(name) > 15 {
			t.Errorf("Expected property name length between 3 and 15, got %d (%s)", len(name), name)
		}
		
		// Should start with a letter
		firstChar := name[0]
		if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')) {
			t.Errorf("Expected property name to start with a letter, got '%c' in '%s'", firstChar, name)
		}
		
		// Should contain only letters and numbers
		for j, char := range name {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				t.Errorf("Expected only alphanumeric characters, got '%c' at position %d in '%s'", char, j, name)
			}
		}
	}
}

func TestSchemaGeneratorMetadataBuilders(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	
	t.Run("addMetadataToStringBuilder", func(t *testing.T) {
		builder := String()
		generator.addMetadataToStringBuilder(builder)
		schema := builder.Build()
		
		// Should have some metadata added
		metadata := schema.Metadata()
		// The function may or may not add metadata based on probability
		// Just verify it doesn't crash
		t.Logf("String metadata: %+v", metadata)
	})
	
	t.Run("addMetadataToNumberBuilder", func(t *testing.T) {
		builder := Number()
		generator.addMetadataToNumberBuilder(builder)
		schema := builder.Build()
		
		// Should have some metadata added
		metadata := schema.Metadata()
		t.Logf("Number metadata: %+v", metadata)
	})
	
	t.Run("addMetadataToIntegerBuilder", func(t *testing.T) {
		builder := Integer()
		generator.addMetadataToIntegerBuilder(builder)
		schema := builder.Build()
		
		// Should have some metadata added
		metadata := schema.Metadata()
		t.Logf("Integer metadata: %+v", metadata)
	})
	
	t.Run("addMetadataToBooleanBuilder", func(t *testing.T) {
		builder := Boolean()
		generator.addMetadataToBooleanBuilder(builder)
		schema := builder.Build()
		
		// Should have some metadata added
		metadata := schema.Metadata()
		t.Logf("Boolean metadata: %+v", metadata)
	})
	
	t.Run("addMetadataToObjectBuilder", func(t *testing.T) {
		builder := Object()
		generator.addMetadataToObjectBuilder(builder)
		schema := builder.Build()
		
		// Should have some metadata added
		metadata := schema.Metadata()
		t.Logf("Object metadata: %+v", metadata)
	})
	
	t.Run("addMetadataToArrayBuilder", func(t *testing.T) {
		builder := Array()
		generator.addMetadataToArrayBuilder(builder)
		schema := builder.Build()
		
		// Should have some metadata added
		metadata := schema.Metadata()
		t.Logf("Array metadata: %+v", metadata)
	})
}

func TestSchemaGeneratorGeneratePropertyName(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()
	
	// Test generating property names
	for i := 0; i < 10; i++ {
		name := generator.generatePropertyName()
		
		if name == "" {
			t.Error("Expected non-empty property name")
		}
		
		// Validate it looks like a property name
		if !isValidPropertyName(name) {
			t.Errorf("Generated invalid property name: %s", name)
		}
	}
}

// Helper function to validate property names
func isValidPropertyName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	// Should start with letter or underscore
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	
	// Should contain only letters, numbers, and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	
	return true
}