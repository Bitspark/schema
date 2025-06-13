package tests

import (
	"testing"

	builders2 "defs.dev/schema/builders"
	"defs.dev/schema/consumers/validation"
	"defs.dev/schema/core"
)

// Helper function to generate JSON Schema using a simple stub
func toJSONSchema(schema core.Schema) map[string]any {
	// Simple stub implementation for testing
	result := map[string]any{}

	// Map schema types to JSON Schema types
	switch schema.Type() {
	case core.TypeStructure:
		result["type"] = "object"
	default:
		result["type"] = string(schema.Type())
	}

	if desc := schema.Metadata().Description; desc != "" {
		result["description"] = desc
	}

	// Handle array schemas
	if schema.Type() == core.TypeArray {
		if arraySchema, ok := schema.(core.ArraySchema); ok {
			if minItems := arraySchema.MinItems(); minItems != nil {
				result["minItems"] = float64(*minItems)
			}
			if maxItems := arraySchema.MaxItems(); maxItems != nil {
				result["maxItems"] = float64(*maxItems)
			}
			if arraySchema.UniqueItemsRequired() {
				result["uniqueItems"] = true
			}
			if itemSchema := arraySchema.ItemSchema(); itemSchema != nil {
				result["items"] = toJSONSchema(itemSchema)
			}
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

	return result
}

func TestArraySchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		s := builders2.NewArraySchema().Build()

		// Valid arrays
		validArrays := []any{
			[]any{"a", "b", "c"},
			[]string{"x", "y", "z"},
			[]int{1, 2, 3},
			[]float64{1.1, 2.2, 3.3},
			[]bool{true, false, true},
			[]any{}, // Empty array
		}

		for _, arr := range validArrays {
			result := validation.ValidateValue(s, arr)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", arr, result.Errors)
			}
		}

		// Invalid values
		invalidValues := []any{
			"not an array",
			42,
			true,
			map[string]any{"key": "value"},
		}

		for _, val := range invalidValues {
			result := validation.ValidateValue(s, val)
			if result.Valid {
				t.Errorf("Expected %v to be invalid for array schema", val)
			}
		}
	})

	t.Run("Min/Max items constraints", func(t *testing.T) {
		schema := builders2.NewArraySchema().Range(2, 4).Build()

		// Valid lengths
		validArrays := []any{
			[]any{"a", "b"},           // min length
			[]any{"a", "b", "c"},      // middle
			[]any{"a", "b", "c", "d"}, // max length
		}

		for _, arr := range validArrays {
			result := validation.ValidateValue(schema, arr)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", arr, result.Errors)
			}
		}

		// Invalid lengths
		result := validation.ValidateValue(schema, []any{"a"}) // too short
		if result.Valid {
			t.Error("Expected single-item array to be invalid (below minimum)")
		}

		result = validation.ValidateValue(schema, []any{"a", "b", "c", "d", "e"}) // too long
		if result.Valid {
			t.Error("Expected five-item array to be invalid (above maximum)")
		}
	})

	t.Run("Item schema validation", func(t *testing.T) {
		// Array of strings
		stringSchema := builders2.NewStringSchema().MinLength(2).Build()
		arraySchema := builders2.NewArraySchema().Items(stringSchema).Build()

		// Valid array - all strings meet criteria
		result := validation.ValidateValue(arraySchema, []any{"hello", "world", "test"})
		if !result.Valid {
			t.Errorf("Expected valid string array, got errors: %v", result.Errors)
		}

		// Invalid array - contains short string
		result = validation.ValidateValue(arraySchema, []any{"hello", "a", "test"})
		if result.Valid {
			t.Error("Expected array with short string to be invalid")
		}

		// Check that we got validation errors
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for array with invalid item")
		}

		// Invalid array - contains non-string
		result = validation.ValidateValue(arraySchema, []any{"hello", 123, "test"})
		if result.Valid {
			t.Error("Expected array with non-string to be invalid")
		}
	})

	t.Run("Unique items constraint", func(t *testing.T) {
		schema := builders2.NewArraySchema().UniqueItems().Build()

		// Valid - all unique
		result := validation.ValidateValue(schema, []any{"a", "b", "c"})
		if !result.Valid {
			t.Errorf("Expected unique array to be valid, got errors: %v", result.Errors)
		}

		// Valid - different types but unique
		result = validation.ValidateValue(schema, []any{"a", 1, true})
		if !result.Valid {
			t.Errorf("Expected mixed unique array to be valid, got errors: %v", result.Errors)
		}

		// Invalid - duplicates
		result = validation.ValidateValue(schema, []any{"a", "b", "a"})
		if result.Valid {
			t.Error("Expected array with duplicates to be invalid")
		}

		// Invalid - duplicate numbers
		result = validation.ValidateValue(schema, []any{1, 2, 1})
		if result.Valid {
			t.Error("Expected array with duplicate numbers to be invalid")
		}
	})

	t.Run("Contains schema validation", func(t *testing.T) {
		// Array must contain at least one string starting with "test"
		containsSchema := builders2.NewStringSchema().Pattern("^test").Build()
		arraySchema := builders2.NewArraySchema().Contains(containsSchema).Build()

		// Valid - contains matching string
		result := validation.ValidateValue(arraySchema, []any{"hello", "testing", "world"})
		if !result.Valid {
			t.Errorf("Expected array with matching string to be valid, got errors: %v", result.Errors)
		}

		// Invalid - no matching strings
		result = validation.ValidateValue(arraySchema, []any{"hello", "world", "foo"})
		if result.Valid {
			t.Error("Expected array without matching string to be invalid")
		}

		// Valid - multiple matching strings
		result = validation.ValidateValue(arraySchema, []any{"test1", "hello", "test2"})
		if !result.Valid {
			t.Errorf("Expected array with multiple matches to be valid, got errors: %v", result.Errors)
		}
	})

	t.Run("Complex nested validation", func(t *testing.T) {
		// Array of objects (using string schemas as proxy for complexity)
		itemSchema := builders2.NewStringSchema().MinLength(3).Build()
		arraySchema := builders2.NewArraySchema().
			Items(itemSchema).
			Range(1, 3).
			UniqueItems().
			Build()

		// Valid array
		result := validation.ValidateValue(arraySchema, []any{"abc", "def", "ghi"})
		if !result.Valid {
			t.Errorf("Expected valid complex array, got errors: %v", result.Errors)
		}

		// Multiple violations
		result = validation.ValidateValue(arraySchema, []any{"ab", "def", "ab"}) // short string + duplicate
		if result.Valid {
			t.Error("Expected array with multiple violations to be invalid")
		}

		// Should have multiple errors
		if len(result.Errors) < 2 {
			t.Errorf("Expected multiple errors, got %d", len(result.Errors))
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		itemSchema := builders2.NewStringSchema().Build()
		schema := builders2.NewArraySchema().
			Items(itemSchema).
			Range(1, 10).
			UniqueItems().
			Description("Test array").
			Example([]any{"item1", "item2"}).
			Build()

		jsonSchema := toJSONSchema(schema)

		if jsonSchema["type"] != "array" {
			t.Errorf("Expected type 'array', got %v", jsonSchema["type"])
		}

		if jsonSchema["minItems"] != float64(1) {
			t.Errorf("Expected minItems 1, got %v", jsonSchema["minItems"])
		}
		if jsonSchema["maxItems"] != float64(10) {
			t.Errorf("Expected maxItems 10, got %v", jsonSchema["maxItems"])
		}
		if jsonSchema["uniqueItems"] != true {
			t.Errorf("Expected uniqueItems true, got %v", jsonSchema["uniqueItems"])
		}
		if jsonSchema["description"] != "Test array" {
			t.Errorf("Expected description 'Test array', got %v", jsonSchema["description"])
		}

		// Check items schema
		items, ok := jsonSchema["items"].(map[string]any)
		if !ok {
			t.Fatal("Expected items to be a map")
		}
		if items["type"] != "string" {
			t.Errorf("Expected items type 'string', got %v", items["type"])
		}
	})
}

func TestArrayBuilder(t *testing.T) {
	t.Run("Fluent API", func(t *testing.T) {
		schema := builders2.NewArraySchema().
			MinItems(2).
			MaxItems(5).
			UniqueItems().
			Description("Test array").
			Build()

		// Test min items constraint
		result := validation.ValidateValue(schema, []any{"a"})
		if result.Valid {
			t.Error("Expected single item array to be invalid (below min)")
		}

		// Test max items constraint
		result = validation.ValidateValue(schema, []any{"a", "b", "c", "d", "e", "f"})
		if result.Valid {
			t.Error("Expected six item array to be invalid (above max)")
		}

		// Test unique items constraint
		result = validation.ValidateValue(schema, []any{"a", "b", "a"})
		if result.Valid {
			t.Error("Expected duplicate array to be invalid")
		}

		// Test valid array
		result = validation.ValidateValue(schema, []any{"a", "b", "c"})
		if !result.Valid {
			t.Errorf("Expected valid array, got errors: %v", result.Errors)
		}
	})

	t.Run("Immutability", func(t *testing.T) {
		builder1 := builders2.NewArraySchema().MinItems(2)
		builder2 := builder1.MaxItems(5)

		schema1 := builder1.Build()
		schema2 := builder2.Build()

		// Verify they're different instances
		if schema1 == schema2 {
			t.Error("Expected schemas to be different instances")
		}

		// Verify first schema doesn't have MaxItems
		if maxItems := schema1.MaxItems(); maxItems != nil {
			t.Error("Expected first schema to not have MaxItems constraint")
		}

		// Verify second schema has both constraints
		if minItems := schema2.MinItems(); minItems == nil || *minItems != 2 {
			t.Error("Expected second schema to have MinItems constraint")
		}
		if maxItems := schema2.MaxItems(); maxItems == nil || *maxItems != 5 {
			t.Error("Expected second schema to have MaxItems constraint")
		}
	})

	t.Run("Helper methods", func(t *testing.T) {
		// Test NonEmpty helper
		schema := builders2.NewArraySchema().NonEmpty().Build()
		result := validation.ValidateValue(schema, []any{})
		if result.Valid {
			t.Error("Expected empty array to be invalid for NonEmpty() schema")
		}

		result = validation.ValidateValue(schema, []any{"item"})
		if !result.Valid {
			t.Errorf("Expected non-empty array to be valid, got errors: %v", result.Errors)
		}

		// Test StringArray helper
		stringArraySchema := builders2.NewArraySchema().StringArray().Build()
		result = validation.ValidateValue(stringArraySchema, []any{"a", "b", "c"})
		if !result.Valid {
			t.Errorf("Expected string array to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(stringArraySchema, []any{"a", 123, "c"})
		if result.Valid {
			t.Error("Expected mixed array to be invalid for StringArray() schema")
		}
	})
}

func TestArraySchemaEdgeCases(t *testing.T) {
	t.Run("Empty array validation", func(t *testing.T) {
		// No constraints - empty array should be valid
		schema := builders2.NewArraySchema().Build()
		result := validation.ValidateValue(schema, []any{})
		if !result.Valid {
			t.Errorf("Expected empty array to be valid with no constraints, got errors: %v", result.Errors)
		}

		// With minItems constraint - empty array should be invalid
		schema = builders2.NewArraySchema().MinItems(1).Build()
		result = validation.ValidateValue(schema, []any{})
		if result.Valid {
			t.Error("Expected empty array to be invalid with minItems constraint")
		}
	})

	t.Run("Nil vs empty array", func(t *testing.T) {
		schema := builders2.NewArraySchema().Build()

		// nil should be invalid
		result := validation.ValidateValue(schema, nil)
		if result.Valid {
			t.Error("Expected nil to be invalid for array schema")
		}

		// Empty slice should be valid
		result = validation.ValidateValue(schema, []any{})
		if !result.Valid {
			t.Errorf("Expected empty slice to be valid, got errors: %v", result.Errors)
		}
	})

	t.Run("Type coercion limits", func(t *testing.T) {
		schema := builders2.NewArraySchema().Build()

		// Various slice types should work
		sliceTypes := []any{
			[]string{"a", "b"},
			[]int{1, 2},
			[]bool{true, false},
			[]any{"mixed", 123, true},
		}

		for _, slice := range sliceTypes {
			result := validation.ValidateValue(schema, slice)
			if !result.Valid {
				t.Errorf("Expected %T to be valid for array schema, got errors: %v", slice, result.Errors)
			}
		}
	})
}
