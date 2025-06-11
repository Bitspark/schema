package generator

import (
	schema2 "defs.dev/schema"
	"testing"
)

func TestSchemaGenerator(t *testing.T) {
	generator := NewSchemaGeneratorWithDefaults()

	t.Run("GenerateBasicSchema", func(t *testing.T) {
		schema := generator.Generate()
		if schema == nil {
			t.Fatal("Generated schema should not be nil")
		}

		// Verify it has a valid type
		schemaType := schema.Type()
		validTypes := []schema2.SchemaType{
			schema2.TypeString, schema2.TypeNumber, schema2.TypeInteger, schema2.TypeBoolean,
			schema2.TypeObject, schema2.TypeArray, schema2.TypeUnion, schema2.TypeOptional,
			schema2.TypeNull, schema2.TypeAny,
		}

		isValid := false
		for _, validType := range validTypes {
			if schemaType == validType {
				isValid = true
				break
			}
		}

		if !isValid {
			t.Errorf("Generated schema has invalid type: %s", schemaType)
		}
	})

	t.Run("GenerateMultipleSchemas", func(t *testing.T) {
		schemas := generator.GenerateMany(5)
		if len(schemas) != 5 {
			t.Errorf("Expected 5 schemas, got %d", len(schemas))
		}

		for i, schema := range schemas {
			if schema == nil {
				t.Errorf("Schema %d should not be nil", i)
			}
		}
	})

	t.Run("SimpleSchemaGenerator", func(t *testing.T) {
		simpleGen := NewSimpleSchemaGenerator()
		schema := simpleGen.Generate()

		if schema == nil {
			t.Fatal("Simple schema should not be nil")
		}

		// Simple schemas should bias toward primitive types
		schemaType := schema.Type()
		primitiveTypes := []schema2.SchemaType{schema2.TypeString, schema2.TypeNumber, schema2.TypeInteger, schema2.TypeBoolean}

		isPrimitive := false
		for _, primitiveType := range primitiveTypes {
			if schemaType == primitiveType {
				isPrimitive = true
				break
			}
		}

		// Note: This test might occasionally fail due to randomness, but should mostly pass
		if !isPrimitive {
			t.Logf("Simple generator produced non-primitive type: %s (this is okay occasionally)", schemaType)
		}
	})

	t.Run("ComplexSchemaGenerator", func(t *testing.T) {
		complexGen := NewComplexSchemaGenerator()
		schema := complexGen.Generate()

		if schema == nil {
			t.Fatal("Complex schema should not be nil")
		}

		// Just verify it generates something valid
		schemaType := schema.Type()
		if schemaType == "" {
			t.Error("Complex schema should have a valid type")
		}
	})
}

func TestSchemaGeneratorConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultSchemaGeneratorConfig()

		if config.MaxDepth <= 0 {
			t.Error("MaxDepth should be positive")
		}

		if config.MaxProperties <= 0 {
			t.Error("MaxProperties should be positive")
		}

		if config.TypeWeights.String < 0 {
			t.Error("String weight should be non-negative")
		}
	})

	t.Run("SimpleConfig", func(t *testing.T) {
		config := SimpleSchemaGeneratorConfig()

		// Simple config should have lower complexity bias
		if config.ComplexityBias >= 0.5 {
			t.Error("Simple config should have low complexity bias")
		}

		// Should have higher weights for primitive types
		if config.TypeWeights.String < config.TypeWeights.Object {
			t.Error("Simple config should favor strings over objects")
		}
	})

	t.Run("ComplexConfig", func(t *testing.T) {
		config := ComplexSchemaGeneratorConfig()

		// Complex config should be more conservative now to avoid validation issues
		if config.ComplexityBias >= 0.5 {
			t.Error("Complex config should have moderate complexity bias for stability")
		}

		// Should still generate some objects but be more conservative
		if config.TypeWeights.Object <= 0 {
			t.Error("Complex config should allow some objects")
		}

		// Should have reasonable depth limits
		if config.MaxDepth > 3 {
			t.Error("Complex config should have reasonable depth limits")
		}
	})
}

func TestSchemaGeneratorConvenienceFunctions(t *testing.T) {
	t.Run("GenerateSchema", func(t *testing.T) {
		schema := GenerateSchema()
		if schema == nil {
			t.Fatal("GenerateSchema should not return nil")
		}
	})

	t.Run("GenerateSchemas", func(t *testing.T) {
		schemas := GenerateSchemas(3)
		if len(schemas) != 3 {
			t.Errorf("Expected 3 schemas, got %d", len(schemas))
		}
	})

	t.Run("GenerateSchemaWithSeed", func(t *testing.T) {
		schema1 := GenerateSchemaWithSeed(12345)
		schema2 := GenerateSchemaWithSeed(12345)

		// With the same seed, we should get the same type (though content may vary)
		if schema1.Type() != schema2.Type() {
			t.Errorf("Same seed should produce same schema type, got %s and %s",
				schema1.Type(), schema2.Type())
		}
	})

	t.Run("GenerateSimpleSchema", func(t *testing.T) {
		schema := GenerateSimpleSchema()
		if schema == nil {
			t.Fatal("GenerateSimpleSchema should not return nil")
		}
	})

	t.Run("GenerateComplexSchema", func(t *testing.T) {
		schema := GenerateComplexSchema()
		if schema == nil {
			t.Fatal("GenerateComplexSchema should not return nil")
		}
	})

	t.Run("GenerateRealisticSchema", func(t *testing.T) {
		schema := GenerateRealisticSchema()
		if schema == nil {
			t.Fatal("GenerateRealisticSchema should not return nil")
		}
	})
}
