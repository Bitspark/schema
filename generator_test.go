package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratorBasicTypes(t *testing.T) {
	generator := NewGeneratorWithDefaults()

	tests := []struct {
		name   string
		schema Schema
		check  func(value any) bool
	}{
		{
			name:   "String",
			schema: String().MinLength(5).MaxLength(10).Build(),
			check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) >= 5 && len(str) <= 10
			},
		},
		{
			name:   "StringWithEnum",
			schema: String().Enum("red", "green", "blue").Build(),
			check: func(value any) bool {
				str, ok := value.(string)
				return ok && (str == "red" || str == "green" || str == "blue")
			},
		},
		{
			name:   "StringWithFormat",
			schema: String().Email().Build(),
			check: func(value any) bool {
				str, ok := value.(string)
				return ok && strings.Contains(str, "@")
			},
		},
		{
			name:   "Number",
			schema: Number().Range(10, 100).Build(),
			check: func(value any) bool {
				num, ok := value.(float64)
				return ok && num >= 10 && num <= 100
			},
		},
		{
			name:   "Integer",
			schema: Integer().Range(1, 50).Build(),
			check: func(value any) bool {
				num, ok := value.(int64)
				return ok && num >= 1 && num <= 50
			},
		},
		{
			name:   "Boolean",
			schema: Boolean().Build(),
			check: func(value any) bool {
				_, ok := value.(bool)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple values to test consistency
			for i := 0; i < 10; i++ {
				value := generator.Generate(tt.schema)
				if !tt.check(value) {
					t.Errorf("Generated value %v does not meet schema constraints", value)
				}

				// Validate that generated value passes schema validation
				result := tt.schema.Validate(value)
				if !result.Valid {
					t.Errorf("Generated value %v failed schema validation: %v", value, result.Errors)
				}
			}
		})
	}
}

func TestGeneratorComplexTypes(t *testing.T) {
	generator := NewGeneratorWithDefaults()

	t.Run("Array", func(t *testing.T) {
		schema := Array().
			Items(String().MinLength(3).Build()).
			MinItems(2).
			MaxItems(5).
			Build()

		value := generator.Generate(schema)
		arr, ok := value.([]any)
		if !ok {
			t.Fatalf("Expected array, got %T", value)
		}

		if len(arr) < 2 || len(arr) > 5 {
			t.Errorf("Array length %d not in range [2, 5]", len(arr))
		}

		for i, item := range arr {
			str, ok := item.(string)
			if !ok {
				t.Errorf("Array item %d is not string: %T", i, item)
				continue
			}
			if len(str) < 3 {
				t.Errorf("Array item %d length %d < 3", i, len(str))
			}
		}

		// Validate schema compliance
		result := schema.Validate(value)
		if !result.Valid {
			t.Errorf("Generated array failed validation: %v", result.Errors)
		}
	})

	t.Run("Object", func(t *testing.T) {
		schema := Object().
			Property("name", String().MinLength(2).Build()).
			Property("age", Integer().Range(18, 100).Build()).
			Property("email", String().Email().Build()).
			Required("name", "age").
			Build()

		value := generator.Generate(schema)
		obj, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("Expected object, got %T", value)
		}

		// Check required properties
		if _, exists := obj["name"]; !exists {
			t.Error("Required property 'name' missing")
		}
		if _, exists := obj["age"]; !exists {
			t.Error("Required property 'age' missing")
		}

		// Validate schema compliance
		result := schema.Validate(value)
		if !result.Valid {
			t.Errorf("Generated object failed validation: %v", result.Errors)
		}
	})

	t.Run("NestedObject", func(t *testing.T) {
		addressSchema := Object().
			Property("street", String().MinLength(5).Build()).
			Property("city", String().MinLength(2).Build()).
			Property("zipcode", String().Pattern("[0-9]{5}").Build()).
			Required("street", "city").
			Build()

		userSchema := Object().
			Property("name", String().MinLength(2).Build()).
			Property("address", addressSchema).
			Property("hobbies", Array().Items(String().Build()).MinItems(1).MaxItems(3).Build()).
			Required("name", "address").
			Build()

		value := generator.Generate(userSchema)

		// Validate schema compliance
		result := userSchema.Validate(value)
		if !result.Valid {
			t.Errorf("Generated nested object failed validation: %v", result.Errors)
		}

		// Pretty print for manual inspection
		jsonData, _ := json.MarshalIndent(value, "", "  ")
		t.Logf("Generated nested object:\n%s", jsonData)
	})
}

func TestGeneratorConfiguration(t *testing.T) {
	t.Run("MaxDepth", func(t *testing.T) {
		config := DefaultGeneratorConfig()
		config.MaxDepth = 2
		generator := NewGenerator(config)

		// Create a deeply nested schema
		deepSchema := Object().
			Property("level1", Object().
				Property("level2", Object().
					Property("level3", Object().
						Property("level4", String().Build()).
						Build()).
					Build()).
				Build()).
			Build()

		value := generator.Generate(deepSchema)

		// The generator should limit depth, so deep nesting should result in nil values
		obj, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("Expected object, got %T", value)
		}

		t.Logf("Generated with max depth 2: %+v", obj)
	})

	t.Run("PreferExamples", func(t *testing.T) {
		config := DefaultGeneratorConfig()
		config.PreferExamples = true
		generator := NewGenerator(config)

		schema := String().Example("test-example").Build()

		// Should always return the example when PreferExamples is true
		for i := 0; i < 5; i++ {
			value := generator.Generate(schema)
			if value != "test-example" {
				t.Errorf("Expected example 'test-example', got %v", value)
			}
		}
	})

	t.Run("OptionalProbability", func(t *testing.T) {
		config := DefaultGeneratorConfig()
		config.OptionalProbability = 0.0 // Never include optional properties
		generator := NewGenerator(config)

		schema := Object().
			Property("required", String().Build()).
			Property("optional", String().Build()).
			Required("required").
			Build()

		optionalCount := 0
		totalRuns := 20

		for i := 0; i < totalRuns; i++ {
			value := generator.Generate(schema)
			obj := value.(map[string]any)
			if _, exists := obj["optional"]; exists {
				optionalCount++
			}
		}

		if optionalCount > 2 { // Allow some variance due to randomness
			t.Errorf("Expected very few optional properties with probability 0.0, got %d/%d", optionalCount, totalRuns)
		}
	})

	t.Run("CustomGenerator", func(t *testing.T) {
		config := DefaultGeneratorConfig()
		config.CustomGenerators["special"] = func(schema Schema, config GeneratorConfig, depth int) any {
			return "custom-generated-value"
		}
		generator := NewGenerator(config)

		schema := String().Name("special").Build()

		value := generator.Generate(schema)
		if value != "custom-generated-value" {
			t.Errorf("Expected custom generator result, got %v", value)
		}
	})

	t.Run("Seed", func(t *testing.T) {
		seed := int64(12345)

		config1 := DefaultGeneratorConfig()
		config1.Seed = seed
		generator1 := NewGenerator(config1)

		config2 := DefaultGeneratorConfig()
		config2.Seed = seed
		generator2 := NewGenerator(config2)

		schema := String().MinLength(10).MaxLength(10).Build()

		// Same seed should produce same results
		value1 := generator1.Generate(schema)
		value2 := generator2.Generate(schema)

		if value1 != value2 {
			t.Errorf("Same seed should produce same results: %v != %v", value1, value2)
		}
	})
}

func TestGeneratorManyValues(t *testing.T) {
	generator := NewGeneratorWithDefaults()

	schema := Object().
		Property("id", Integer().Range(1, 1000).Build()).
		Property("name", String().MinLength(3).MaxLength(15).Build()).
		Property("active", Boolean().Build()).
		Required("id", "name", "active").
		Build()

	values := generator.GenerateMany(schema, 5)

	if len(values) != 5 {
		t.Errorf("Expected 5 values, got %d", len(values))
	}

	for i, value := range values {
		result := schema.Validate(value)
		if !result.Valid {
			t.Errorf("Generated value %d failed validation: %v", i, result.Errors)
		}
	}

	// Print for manual inspection
	for i, value := range values {
		jsonData, _ := json.MarshalIndent(value, "", "  ")
		t.Logf("Generated value %d:\n%s", i+1, jsonData)
	}
}

func TestGeneratorFormats(t *testing.T) {
	generator := NewGeneratorWithDefaults()

	formats := map[string]func(string) bool{
		"email": func(s string) bool {
			return strings.Contains(s, "@") && strings.Contains(s, ".")
		},
		"uuid": func(s string) bool {
			parts := strings.Split(s, "-")
			return len(parts) == 5 &&
				len(parts[0]) == 8 &&
				len(parts[1]) == 4 &&
				len(parts[2]) == 4 &&
				len(parts[3]) == 4 &&
				len(parts[4]) == 12
		},
		"url": func(s string) bool {
			return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
		},
		"date": func(s string) bool {
			parts := strings.Split(s, "-")
			return len(parts) == 3 && len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2
		},
		"time": func(s string) bool {
			parts := strings.Split(s, ":")
			return len(parts) == 3 && len(parts[0]) == 2 && len(parts[1]) == 2 && len(parts[2]) == 2
		},
		"date-time": func(s string) bool {
			return strings.Contains(s, "T") && strings.HasSuffix(s, "Z")
		},
	}

	for format, validator := range formats {
		t.Run(format, func(t *testing.T) {
			var schema Schema
			switch format {
			case "email":
				schema = String().Email().Build()
			case "uuid":
				schema = String().UUID().Build()
			case "url":
				schema = String().URL().Build()
			default:
				// Create a string schema with the format
				schema = &StringSchema{
					metadata: SchemaMetadata{},
					format:   format,
				}
			}

			for i := 0; i < 5; i++ {
				value := generator.Generate(schema)
				str, ok := value.(string)
				if !ok {
					t.Errorf("Expected string, got %T", value)
					continue
				}

				if !validator(str) {
					t.Errorf("Generated value %q does not match format %s", str, format)
				}

				t.Logf("Generated %s: %s", format, str)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	schema := Object().
		Property("name", String().MinLength(3).Build()).
		Property("count", Integer().Range(1, 100).Build()).
		Required("name", "count").
		Build()

	t.Run("Generate", func(t *testing.T) {
		value := Generate(schema)
		result := schema.Validate(value)
		if !result.Valid {
			t.Errorf("Generate() produced invalid value: %v", result.Errors)
		}
	})

	t.Run("GenerateMany", func(t *testing.T) {
		values := GenerateMany(schema, 3)
		if len(values) != 3 {
			t.Errorf("Expected 3 values, got %d", len(values))
		}

		for i, value := range values {
			result := schema.Validate(value)
			if !result.Valid {
				t.Errorf("GenerateMany() value %d is invalid: %v", i, result.Errors)
			}
		}
	})

	t.Run("GenerateWithSeed", func(t *testing.T) {
		value1 := GenerateWithSeed(schema, 42)
		value2 := GenerateWithSeed(schema, 42)

		// Same seed should produce same result
		if !reflect.DeepEqual(value1, value2) {
			t.Errorf("Same seed should produce same result: %v != %v", value1, value2)
		}

		// Different seed should produce different result (with high probability)
		value3 := GenerateWithSeed(schema, 43)
		if reflect.DeepEqual(value1, value3) {
			t.Logf("Different seeds produced same result (rare but possible): %v == %v", value1, value3)
		}
	})
}

func TestGeneratorWithComplexScenarios(t *testing.T) {
	generator := NewGeneratorWithDefaults()

	t.Run("APIResponse", func(t *testing.T) {
		// Simulate a complex API response schema
		userSchema := Object().
			Property("id", Integer().Range(1, 10000).Build()).
			Property("username", String().MinLength(3).MaxLength(20).Build()).
			Property("email", String().Email().Build()).
			Property("profile", Object().
				Property("firstName", String().MinLength(1).MaxLength(50).Build()).
				Property("lastName", String().MinLength(1).MaxLength(50).Build()).
				Property("avatar", String().URL().Build()).
				Property("preferences", Object().
					Property("theme", String().Enum("light", "dark").Build()).
					Property("notifications", Boolean().Build()).
					Build()).
				Required("firstName", "lastName").
				Build()).
			Property("roles", Array().
				Items(String().Enum("admin", "user", "moderator").Build()).
				MinItems(1).
				MaxItems(3).
				Build()).
			Required("id", "username", "email", "profile", "roles").
			Build()

		apiResponseSchema := Object().
			Property("success", Boolean().Build()).
			Property("data", userSchema).
			Property("metadata", Object().
				Property("timestamp", String().Build()).
				Property("version", String().Build()).
				Build()).
			Required("success", "data").
			Build()

		value := generator.Generate(apiResponseSchema)
		result := apiResponseSchema.Validate(value)
		if !result.Valid {
			t.Errorf("Generated API response failed validation: %v", result.Errors)
		}

		// Pretty print the result
		jsonData, _ := json.MarshalIndent(value, "", "  ")
		t.Logf("Generated API Response:\n%s", jsonData)
	})

	t.Run("ConfigurationFile", func(t *testing.T) {
		// Simulate a configuration file schema
		configSchema := Object().
			Property("server", Object().
				Property("host", String().Default("localhost").Build()).
				Property("port", Integer().Range(1000, 9999).Build()).
				Property("ssl", Boolean().Build()).
				Required("port").
				Build()).
			Property("database", Object().
				Property("url", String().URL().Build()).
				Property("maxConnections", Integer().Range(10, 100).Build()).
				Property("timeout", Integer().Range(5, 60).Build()).
				Required("url").
				Build()).
			Property("features", Object().
				Property("enableLogging", Boolean().Build()).
				Property("enableMetrics", Boolean().Build()).
				Property("enableTracing", Boolean().Build()).
				Build()).
			Property("environments", Array().
				Items(String().Enum("development", "staging", "production").Build()).
				MinItems(1).
				MaxItems(3).
				UniqueItems().
				Build()).
			Required("server", "database").
			Build()

		value := generator.Generate(configSchema)
		result := configSchema.Validate(value)
		if !result.Valid {
			t.Errorf("Generated config failed validation: %v", result.Errors)
		}

		// Pretty print the result
		jsonData, _ := json.MarshalIndent(value, "", "  ")
		t.Logf("Generated Configuration:\n%s", jsonData)
	})
}

// Benchmark the generator performance
func BenchmarkGenerator(b *testing.B) {
	generator := NewGeneratorWithDefaults()

	schema := Object().
		Property("users", Array().
			Items(Object().
				Property("id", Integer().Range(1, 1000000).Build()).
				Property("name", String().MinLength(5).MaxLength(25).Build()).
				Property("email", String().Email().Build()).
				Property("metadata", Object().
					Property("created", String().Build()).
					Property("updated", String().Build()).
					Build()).
				Required("id", "name", "email").
				Build()).
			MinItems(10).
			MaxItems(50).
			Build()).
		Property("total", Integer().Build()).
		Required("users", "total").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Generate(schema)
	}
}

func ExampleGenerator() {
	// Create a generator with custom configuration
	config := DefaultGeneratorConfig()
	config.MaxDepth = 3
	config.MaxItems = 5
	config.StringLength.Min = 5
	config.StringLength.Max = 15
	config.OptionalProbability = 0.8
	config.Seed = 12345 // For reproducible results

	generator := NewGenerator(config)

	// Define a user schema
	userSchema := Object().
		Property("id", Integer().Range(1, 1000).Build()).
		Property("name", String().MinLength(3).MaxLength(20).Build()).
		Property("email", String().Email().Build()).
		Property("age", Integer().Range(18, 80).Build()).
		Property("isActive", Boolean().Build()).
		Property("tags", Array().Items(String().Build()).MaxItems(3).Build()).
		Required("id", "name", "email").
		Build()

	// Generate a random user
	user := generator.Generate(userSchema)

	// The generated user will be a valid object conforming to the schema
	fmt.Printf("Generated user: %+v\n", user)

	// Generate multiple users
	users := generator.GenerateMany(userSchema, 3)
	for i, u := range users {
		fmt.Printf("User %d: %+v\n", i+1, u)
	}

	// Use convenience functions
	quickUser := Generate(userSchema)
	fmt.Printf("Quick generated user: %+v\n", quickUser)
}
