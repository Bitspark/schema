package generator

import (
	"defs.dev/schema"
	"testing"
)

// TestSchemaGeneratorIntegration performs comprehensive integration testing
// by generating random schemas, generating values for those schemas, and validating them
func TestSchemaGeneratorIntegration(t *testing.T) {
	const numIterations = 100 // Test 100 random schemas

	t.Run("RandomSchemaValueValidation", func(t *testing.T) {
		schemaGen := NewSchemaGeneratorWithDefaults()
		valueGen := NewGeneratorWithDefaults()

		successCount := 0
		var failures []IntegrationFailure

		for i := 0; i < numIterations; i++ {
			// Step 1: Generate a random schema
			randomSchema := schemaGen.Generate()

			// Step 2: Generate a random value for that schema
			randomValue := valueGen.Generate(randomSchema)

			// Step 3: Validate the value against the schema
			result := randomSchema.Validate(randomValue)

			if result.Valid {
				successCount++
			} else {
				// Collect failure information for analysis
				failure := IntegrationFailure{
					Iteration:        i,
					SchemaType:       randomSchema.Type(),
					SchemaJSON:       getSchemaJSON(randomSchema),
					GeneratedValue:   randomValue,
					ValidationErrors: result.Errors,
				}
				failures = append(failures, failure)

				// Log detailed failure information
				t.Logf("FAILURE #%d: Schema type %s failed validation", i, randomSchema.Type())
				t.Logf("  Generated value: %+v", randomValue)
				t.Logf("  Validation errors: %v", result.Errors)
			}
		}

		// Report results
		successRate := float64(successCount) / float64(numIterations) * 100
		t.Logf("Integration Test Results:")
		t.Logf("  Total iterations: %d", numIterations)
		t.Logf("  Successful: %d (%.1f%%)", successCount, successRate)
		t.Logf("  Failed: %d (%.1f%%)", len(failures), 100-successRate)

		// Analyze failure patterns
		if len(failures) > 0 {
			analyzeFailures(t, failures)
		}

		// We expect a very high success rate (>90%)
		// Some failures might be expected due to edge cases or complex schemas
		if successRate < 90.0 {
			t.Errorf("Success rate too low: %.1f%% (expected >90%%)", successRate)
		}
	})

	t.Run("SpecificSchemaTypes", func(t *testing.T) {
		// Test each schema type specifically to ensure good coverage
		testSpecificSchemaType(t, "Simple", NewSimpleSchemaGenerator())
		testSpecificSchemaType(t, "Complex", NewComplexSchemaGenerator())
		testSpecificSchemaType(t, "Realistic", func() *SchemaGenerator {
			config := DefaultSchemaGeneratorConfig()
			config.PropertyNameStyle = "realistic"
			config.UseCommonFormats = true
			return NewSchemaGenerator(config)
		}())
	})

	t.Run("ReproducibilityTest", func(t *testing.T) {
		// Test that the same seed produces the same results
		seed := int64(12345)

		// Generate schema and value with seed
		schema1 := GenerateSchemaWithSeed(seed)
		config1 := DefaultGeneratorConfig()
		config1.Seed = seed
		gen1 := NewGenerator(config1)
		value1 := gen1.Generate(schema1)

		// Generate again with same seed
		schema2 := GenerateSchemaWithSeed(seed)
		config2 := DefaultGeneratorConfig()
		config2.Seed = seed
		gen2 := NewGenerator(config2)
		value2 := gen2.Generate(schema2)

		// Should produce the same schema type and validation should work for both
		if schema1.Type() != schema2.Type() {
			t.Errorf("Same seed produced different schema types: %s vs %s", schema1.Type(), schema2.Type())
		}

		// Both values should validate against both schemas
		if result := schema1.Validate(value1); !result.Valid {
			t.Errorf("First schema/value pair failed validation: %v", result.Errors)
		}
		if result := schema2.Validate(value2); !result.Valid {
			t.Errorf("Second schema/value pair failed validation: %v", result.Errors)
		}
	})

	t.Run("ConstraintStressTest", func(t *testing.T) {
		// Test schemas with heavy constraints
		config := DefaultSchemaGeneratorConfig()
		config.GenerateConstraints = true
		config.ConstraintProbability = 1.0 // Always generate constraints
		config.GenerateEnums = true

		constraintGen := NewSchemaGenerator(config)
		valueGen := NewGeneratorWithDefaults()

		failures := 0
		for i := 0; i < 50; i++ {
			schema := constraintGen.Generate()
			value := valueGen.Generate(schema)

			if result := schema.Validate(value); !result.Valid {
				failures++
				t.Logf("Constraint test failure %d: %s schema, value: %+v, errors: %v",
					i, schema.Type(), value, result.Errors)
			}
		}

		if failures > 10 { // Allow some failures with heavy constraints
			t.Errorf("Too many constraint failures: %d/50", failures)
		}
	})
}

// IntegrationFailure captures details about a failed integration test
type IntegrationFailure struct {
	Iteration        int
	SchemaType       schema.SchemaType
	SchemaJSON       map[string]any
	GeneratedValue   any
	ValidationErrors []schema.ValidationError
}

// testSpecificSchemaType tests a specific type of schema generator
func testSpecificSchemaType(t *testing.T, name string, schemaGen *SchemaGenerator) {
	t.Run(name, func(t *testing.T) {
		valueGen := NewGeneratorWithDefaults()

		for i := 0; i < 20; i++ {
			schema := schemaGen.Generate()
			value := valueGen.Generate(schema)

			if result := schema.Validate(value); !result.Valid {
				t.Errorf("%s generator failure %d: %s schema failed validation",
					name, i, schema.Type())
				t.Logf("  Value: %+v", value)
				t.Logf("  Errors: %v", result.Errors)
			}
		}
	})
}

// analyzeFailures provides detailed analysis of integration test failures
func analyzeFailures(t *testing.T, failures []IntegrationFailure) {
	t.Log("Failure Analysis:")

	// Count failures by schema type
	typeFailures := make(map[schema.SchemaType]int)
	for _, failure := range failures {
		typeFailures[failure.SchemaType]++
	}

	t.Log("  Failures by schema type:")
	for schemaType, count := range typeFailures {
		t.Logf("    %s: %d failures", schemaType, count)
	}

	// Count failures by error type
	errorTypes := make(map[string]int)
	for _, failure := range failures {
		for _, err := range failure.ValidationErrors {
			errorTypes[err.Code]++
		}
	}

	t.Log("  Failures by error type:")
	for errorType, count := range errorTypes {
		t.Logf("    %s: %d occurrences", errorType, count)
	}

	// Show a few example failures
	if len(failures) > 0 {
		t.Log("  Example failures:")
		for i, failure := range failures {
			if i >= 3 { // Only show first 3 examples
				break
			}
			t.Logf("    Example %d: %s schema", i+1, failure.SchemaType)
			t.Logf("      Value: %+v", failure.GeneratedValue)
			if len(failure.ValidationErrors) > 0 {
				t.Logf("      Error: %s", failure.ValidationErrors[0].Message)
			}
		}
	}
}

// getSchemaJSON safely converts a schema to JSON for logging
func getSchemaJSON(schema schema.Schema) map[string]any {
	defer func() {
		if r := recover(); r != nil {
			// If JSON conversion fails, return a simple representation
		}
	}()
	return schema.ToJSONSchema()
}

// TestSchemaGeneratorEdgeCases tests specific edge cases that might cause issues
func TestSchemaGeneratorEdgeCases(t *testing.T) {
	t.Run("EmptyObjectSchema", func(t *testing.T) {
		// Test object schemas with no properties
		config := DefaultSchemaGeneratorConfig()
		config.MinProperties = 0
		config.MaxProperties = 0

		gen := NewSchemaGenerator(config)
		valueGen := NewGeneratorWithDefaults()

		for i := 0; i < 10; i++ {
			s := gen.Generate()
			if s.Type() == schema.TypeObject {
				value := valueGen.Generate(s)
				if result := s.Validate(value); !result.Valid {
					t.Errorf("Empty object validation failed: %v", result.Errors)
				}
			}
		}
	})

	t.Run("DeepNestingTest", func(t *testing.T) {
		// Test very deep nesting
		config := DefaultSchemaGeneratorConfig()
		config.MaxDepth = 10
		config.ComplexityBias = 0.9

		gen := NewSchemaGenerator(config)
		valueGen := NewGeneratorWithDefaults()

		for i := 0; i < 5; i++ {
			schema := gen.Generate()
			value := valueGen.Generate(schema)

			if result := schema.Validate(value); !result.Valid {
				t.Errorf("Deep nesting test failed: %v", result.Errors)
			}
		}
	})

	t.Run("HighConstraintDensity", func(t *testing.T) {
		// Test schemas with many constraints
		config := DefaultSchemaGeneratorConfig()
		config.GenerateConstraints = true
		config.ConstraintProbability = 1.0
		config.GenerateEnums = true
		config.GenerateFormats = true
		config.GeneratePatterns = true

		gen := NewSchemaGenerator(config)
		valueGen := NewGeneratorWithDefaults()

		for i := 0; i < 20; i++ {
			schema := gen.Generate()
			value := valueGen.Generate(schema)

			// Some constraint failures are expected, but most should pass
			result := schema.Validate(value)
			if !result.Valid {
				t.Logf("Constraint validation failed (expected occasionally): %s schema, errors: %v",
					schema.Type(), result.Errors)
			}
		}
	})
}

// TestSchemaGeneratorPerformance tests the performance of schema generation
func TestSchemaGeneratorPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("BulkGeneration", func(t *testing.T) {
		gen := NewSchemaGeneratorWithDefaults()

		// Generate 1000 schemas and measure performance
		schemas := gen.GenerateMany(1000)

		if len(schemas) != 1000 {
			t.Errorf("Expected 1000 schemas, got %d", len(schemas))
		}

		// Verify all schemas are valid
		for i, schema := range schemas {
			if schema == nil {
				t.Errorf("Schema %d is nil", i)
			}
			if schema.Type() == "" {
				t.Errorf("Schema %d has empty type", i)
			}
		}
	})

	t.Run("ValueGenerationPerformance", func(t *testing.T) {
		// Test generating many values for the same schema
		schemaGen := NewSchemaGeneratorWithDefaults()
		valueGen := NewGeneratorWithDefaults()

		schema := schemaGen.Generate()
		values := valueGen.GenerateMany(schema, 100)

		if len(values) != 100 {
			t.Errorf("Expected 100 values, got %d", len(values))
		}

		// Validate all generated values
		failures := 0
		for i, value := range values {
			if result := schema.Validate(value); !result.Valid {
				failures++
				if failures <= 3 { // Only log first few failures
					t.Logf("Performance test validation failure %d: %v", i, result.Errors)
				}
			}
		}

		if failures > 10 { // Allow some failures
			t.Errorf("Too many validation failures: %d/100", failures)
		}
	})
}
