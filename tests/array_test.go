package tests

import (
	"defs.dev/schema"
	"defs.dev/schema/schemas"
	"testing"
)

func TestArraySchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := schema.NewArray().Build()

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
			result := schema.Validate(arr)
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
			result := schema.Validate(val)
			if result.Valid {
				t.Errorf("Expected %v to be invalid for array schema", val)
			}
		}
	})

	t.Run("Min/Max items constraints", func(t *testing.T) {
		schema := schema.NewArray().Range(2, 4).Build()

		// Valid lengths
		validArrays := []any{
			[]any{"a", "b"},           // min length
			[]any{"a", "b", "c"},      // middle
			[]any{"a", "b", "c", "d"}, // max length
		}

		for _, arr := range validArrays {
			result := schema.Validate(arr)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", arr, result.Errors)
			}
		}

		// Invalid lengths
		result := schema.Validate([]any{"a"}) // too short
		if result.Valid {
			t.Error("Expected single-item array to be invalid (below minimum)")
		}

		result = schema.Validate([]any{"a", "b", "c", "d", "e"}) // too long
		if result.Valid {
			t.Error("Expected five-item array to be invalid (above maximum)")
		}
	})

	t.Run("Item schema validation", func(t *testing.T) {
		// Array of strings
		stringSchema := schema.NewString().MinLength(2).Build()
		arraySchema := schema.NewArray().Items(stringSchema).Build()

		// Valid array - all strings meet criteria
		result := arraySchema.Validate([]any{"hello", "world", "test"})
		if !result.Valid {
			t.Errorf("Expected valid string array, got errors: %v", result.Errors)
		}

		// Invalid array - contains short string
		result = arraySchema.Validate([]any{"hello", "a", "test"})
		if result.Valid {
			t.Error("Expected array with short string to be invalid")
		}

		// Check error path includes array index
		if len(result.Errors) > 0 {
			foundIndexedError := false
			for _, err := range result.Errors {
				if err.Path == "[1]" && err.Context == "Array item 1" {
					foundIndexedError = true
					break
				}
			}
			if !foundIndexedError {
				t.Error("Expected error with array index path")
			}
		}

		// Invalid array - contains non-string
		result = arraySchema.Validate([]any{"hello", 123, "test"})
		if result.Valid {
			t.Error("Expected array with non-string to be invalid")
		}
	})

	t.Run("Unique items constraint", func(t *testing.T) {
		schema := schema.NewArray().UniqueItems().Build()

		// Valid - all unique
		result := schema.Validate([]any{"a", "b", "c"})
		if !result.Valid {
			t.Errorf("Expected unique array to be valid, got errors: %v", result.Errors)
		}

		// Valid - different types but unique
		result = schema.Validate([]any{"a", 1, true})
		if !result.Valid {
			t.Errorf("Expected mixed unique array to be valid, got errors: %v", result.Errors)
		}

		// Invalid - duplicates
		result = schema.Validate([]any{"a", "b", "a"})
		if result.Valid {
			t.Error("Expected array with duplicates to be invalid")
		}

		// Invalid - duplicate numbers
		result = schema.Validate([]any{1, 2, 1})
		if result.Valid {
			t.Error("Expected array with duplicate numbers to be invalid")
		}
	})

	t.Run("Contains schema validation", func(t *testing.T) {
		// Array must contain at least one string starting with "test"
		containsSchema := schema.NewString().Pattern("^test").Build()
		arraySchema := schema.NewArray().Contains(containsSchema).Build()

		// Valid - contains matching string
		result := arraySchema.Validate([]any{"hello", "testing", "world"})
		if !result.Valid {
			t.Errorf("Expected array with matching string to be valid, got errors: %v", result.Errors)
		}

		// Invalid - no matching strings
		result = arraySchema.Validate([]any{"hello", "world", "foo"})
		if result.Valid {
			t.Error("Expected array without matching string to be invalid")
		}

		// Valid - multiple matching strings
		result = arraySchema.Validate([]any{"test1", "hello", "test2"})
		if !result.Valid {
			t.Errorf("Expected array with multiple matches to be valid, got errors: %v", result.Errors)
		}
	})

	t.Run("Complex nested validation", func(t *testing.T) {
		// Array of objects (using string schemas as proxy for complexity)
		itemSchema := schema.NewString().MinLength(3).Build()
		arraySchema := schema.NewArray().
			Items(itemSchema).
			Range(1, 3).
			UniqueItems().
			Build()

		// Valid array
		result := arraySchema.Validate([]any{"abc", "def", "ghi"})
		if !result.Valid {
			t.Errorf("Expected valid complex array, got errors: %v", result.Errors)
		}

		// Multiple violations
		result = arraySchema.Validate([]any{"ab", "def", "ab"}) // short string + duplicate
		if result.Valid {
			t.Error("Expected array with multiple violations to be invalid")
		}

		// Should have multiple errors
		if len(result.Errors) < 2 {
			t.Errorf("Expected multiple errors, got %d", len(result.Errors))
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		itemSchema := schema.NewString().Build()
		schema := schema.NewArray().
			Items(itemSchema).
			Range(1, 10).
			UniqueItems().
			Description("Test array").
			Example([]any{"item1", "item2"}).
			Build()

		jsonSchema := schema.ToJSONSchema()

		if jsonSchema["type"] != "array" {
			t.Errorf("Expected type 'array', got %v", jsonSchema["type"])
		}
		if jsonSchema["minItems"] != 1 {
			t.Errorf("Expected minItems 1, got %v", jsonSchema["minItems"])
		}
		if jsonSchema["maxItems"] != 10 {
			t.Errorf("Expected maxItems 10, got %v", jsonSchema["maxItems"])
		}
		if jsonSchema["uniqueItems"] != true {
			t.Error("Expected uniqueItems to be true")
		}
		if jsonSchema["description"] != "Test array" {
			t.Errorf("Expected description 'Test array', got %v", jsonSchema["description"])
		}

		// Check items schema
		items, ok := jsonSchema["items"].(map[string]any)
		if !ok {
			t.Error("Expected items to be a schema object")
		} else if items["type"] != "string" {
			t.Errorf("Expected item type 'string', got %v", items["type"])
		}
	})

	t.Run("Example generation", func(t *testing.T) {
		// Test with item schema
		itemSchema := schema.NewString().Build()
		schema := schema.NewArray().
			Items(itemSchema).
			Range(2, 4).
			Build()

		example := schema.GenerateExample()
		exampleArray, ok := example.([]any)
		if !ok {
			t.Errorf("Expected generated example to be array, got %T", example)
		}

		if len(exampleArray) < 2 || len(exampleArray) > 4 {
			t.Errorf("Expected example length between 2-4, got %d", len(exampleArray))
		}

		// Test unique items example generation
		uniqueSchema := schema.NewArray().
			Items(itemSchema).
			UniqueItems().
			Range(2, 3).
			Build()

		uniqueExample := uniqueSchema.GenerateExample()
		uniqueArray, ok := uniqueExample.([]any)
		if !ok {
			t.Errorf("Expected unique example to be array, got %T", uniqueExample)
		}

		// Check uniqueness
		seen := make(map[any]bool)
		for _, item := range uniqueArray {
			if seen[item] {
				t.Error("Generated example should have unique items")
				break
			}
			seen[item] = true
		}
	})
}

func TestArrayBuilder(t *testing.T) {
	t.Run("Basic builder methods", func(t *testing.T) {
		schema := schema.NewArray().
			Description("Test array").
			Name("test_array").
			Tag("testing").
			MinItems(1).
			MaxItems(5).
			Build()

		if schema.Type() != "array" {
			t.Errorf("Expected type 'array', got %v", schema.Type())
		}

		metadata := schema.Metadata()
		if metadata.Description != "Test array" {
			t.Errorf("Expected description 'Test array', got %v", metadata.Description)
		}
		if metadata.Name != "test_array" {
			t.Errorf("Expected name 'test_array', got %v", metadata.Name)
		}
		if len(metadata.Tags) != 1 || metadata.Tags[0] != "testing" {
			t.Errorf("Expected tag 'testing', got %v", metadata.Tags)
		}

		if *schema.MinItems() != 1 {
			t.Errorf("Expected minItems 1, got %v", schema.MinItems())
		}
		if *schema.MaxItems() != 5 {
			t.Errorf("Expected maxItems 5, got %v", schema.MaxItems())
		}
	})

	t.Run("Helper methods", func(t *testing.T) {
		// Test NonEmpty
		nonEmptySchema := schema.NewArray().NonEmpty().Build()
		if *nonEmptySchema.MinItems() != 1 {
			t.Error("Expected NonEmpty to set minItems to 1")
		}

		// Test Set (unique items)
		setSchema := schema.NewArray().Set().Build()
		if !setSchema.UniqueItemsRequired() {
			t.Error("Expected Set to require unique items")
		}

		// Test Tuple (fixed length)
		tupleSchema := schema.NewArray().Tuple(3).Build()
		if *tupleSchema.MinItems() != 3 || *tupleSchema.MaxItems() != 3 {
			t.Error("Expected Tuple to set fixed length")
		}

		// Test LimitedList
		limitedSchema := schema.NewArray().LimitedList(10).Build()
		if *limitedSchema.MinItems() != 0 || *limitedSchema.MaxItems() != 10 {
			t.Error("Expected LimitedList to set range 0-10")
		}
	})

	t.Run("Type-specific helpers", func(t *testing.T) {
		// Test StringArray helper
		stringArraySchema := schema.NewArray().StringArray().Build()
		metadata := stringArraySchema.Metadata()
		if metadata.Description != "Array of strings" {
			t.Error("Expected StringArray to set appropriate description")
		}

		// Test NumberArray helper
		numberArraySchema := schema.NewArray().NumberArray().Build()
		metadata = numberArraySchema.Metadata()
		if metadata.Description != "Array of numbers" {
			t.Error("Expected NumberArray to set appropriate description")
		}

		// Test IntegerArray helper
		intArraySchema := schema.NewArray().IntegerArray().Build()
		metadata = intArraySchema.Metadata()
		if metadata.Description != "Array of integers" {
			t.Error("Expected IntegerArray to set appropriate description")
		}

		// Test BooleanArray helper
		boolArraySchema := schema.NewArray().BooleanArray().Build()
		metadata = boolArraySchema.Metadata()
		if metadata.Description != "Array of booleans" {
			t.Error("Expected BooleanArray to set appropriate description")
		}
	})

	t.Run("Complex composition", func(t *testing.T) {
		// Test complex schema composition
		stringSchema := schema.NewString().MinLength(2).Build()
		schema := schema.NewArray().
			Items(stringSchema).
			UniqueItems().
			Range(1, 5).
			Description("Array of unique strings").
			Example([]any{"hello", "world"}).
			Default([]any{"default"}).
			Build()

		// Test validation works
		result := schema.Validate([]any{"ab", "cd"})
		if !result.Valid {
			t.Errorf("Expected valid array, got errors: %v", result.Errors)
		}

		// Test invalid case
		result = schema.Validate([]any{"a", "b"}) // strings too short
		if result.Valid {
			t.Error("Expected array with short strings to be invalid")
		}

		// Test default value (through the concrete implementation)
		if concreteSchema, ok := schema.(*schemas.ArraySchema); ok {
			defaultVal := concreteSchema.DefaultValue()
			if len(defaultVal) != 1 || defaultVal[0] != "default" {
				t.Errorf("Expected default value ['default'], got %v", defaultVal)
			}
		} else {
			t.Error("Could not cast to concrete ArraySchema to test default value")
		}
	})
}

func TestArraySchemaEdgeCases(t *testing.T) {
	t.Run("Empty array validation", func(t *testing.T) {
		schema := schema.NewArray().Build()

		// Empty array should be valid
		result := schema.Validate([]any{})
		if !result.Valid {
			t.Errorf("Expected empty array to be valid, got errors: %v", result.Errors)
		}

		// With minimum constraint
		minSchema := schema.NewArray().MinItems(1).Build()
		result = minSchema.Validate([]any{})
		if result.Valid {
			t.Error("Expected empty array to be invalid with min constraint")
		}
	})

	t.Run("Nil array handling", func(t *testing.T) {
		schema := schema.NewArray().Build()

		// Nil should be invalid
		result := schema.Validate(nil)
		if result.Valid {
			t.Error("Expected nil to be invalid for array schema")
		}
	})

	t.Run("Mixed type arrays", func(t *testing.T) {
		schema := schema.NewArray().Build()

		// Mixed types should be valid without item schema
		result := schema.Validate([]any{"string", 42, true, 3.14})
		if !result.Valid {
			t.Errorf("Expected mixed type array to be valid, got errors: %v", result.Errors)
		}
	})

	t.Run("Large array performance", func(t *testing.T) {
		schema := schema.NewArray().UniqueItems().Build()

		// Create large array with unique items
		largeArray := make([]any, 1000)
		for i := 0; i < 1000; i++ {
			largeArray[i] = i
		}

		result := schema.Validate(largeArray)
		if !result.Valid {
			t.Errorf("Expected large unique array to be valid, got errors: %v", result.Errors)
		}

		// Test with duplicate (should catch it)
		largeArray[999] = 0 // Create duplicate
		result = schema.Validate(largeArray)
		if result.Valid {
			t.Error("Expected large array with duplicate to be invalid")
		}
	})
}
