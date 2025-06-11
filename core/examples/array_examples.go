package examples

import (
	"fmt"
	"log"

	"defs.dev/schema/core"
)

// ArrayExamples demonstrates various ArraySchema usage patterns
func ArrayExamples() {
	fmt.Println("=== Array Schema Examples ===")

	// Basic array validation
	basicArrayExample()

	// Array with item schema
	itemSchemaExample()

	// Array constraints
	constraintsExample()

	// Unique items arrays
	uniqueItemsExample()

	// Contains validation
	containsExample()

	// Nested arrays
	nestedArraysExample()

	// Common array patterns
	commonPatternsExample()

	// Complex compositions
	complexCompositionExample()
}

func basicArrayExample() {
	fmt.Println("--- Basic Array Validation ---")

	// Simple array schema that accepts any array
	schema := core.NewArray().
		Description("Simple array").
		Build()

	// Test various array types
	testArrays := []any{
		[]any{"a", "b", "c"},
		[]string{"hello", "world"},
		[]int{1, 2, 3},
		[]float64{1.1, 2.2, 3.3},
		[]bool{true, false},
		[]any{},                        // Empty array
		[]any{"mixed", 42, true, 3.14}, // Mixed types
	}

	for _, arr := range testArrays {
		result := schema.Validate(arr)
		fmt.Printf("Array %v: %s\n", arr, validationStatus(result))
	}

	// Test invalid values
	invalidValues := []any{"not array", 42, true, map[string]any{"key": "value"}}
	for _, val := range invalidValues {
		result := schema.Validate(val)
		fmt.Printf("Value %v: %s\n", val, validationStatus(result))
	}

	fmt.Println()
}

func itemSchemaExample() {
	fmt.Println("--- Array with Item Schema ---")

	// Array of strings with minimum length
	stringItemSchema := core.NewString().
		MinLength(3).
		Pattern("^[a-zA-Z]+$").
		Build()

	arraySchema := core.NewArray().
		Items(stringItemSchema).
		Description("Array of valid words (min 3 letters, letters only)").
		Build()

	testArrays := []any{
		[]any{"hello", "world", "test"},    // Valid
		[]any{"hello", "world", "ok"},      // Invalid - "ok" too short
		[]any{"hello", "world123", "test"}, // Invalid - contains numbers
		[]any{"hello", 123, "test"},        // Invalid - contains non-string
	}

	for _, arr := range testArrays {
		result := arraySchema.Validate(arr)
		fmt.Printf("Array %v: %s\n", arr, validationStatus(result))
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error at %s: %s\n", err.Path, err.Message)
			}
		}
	}

	fmt.Println()
}

func constraintsExample() {
	fmt.Println("--- Array Constraints ---")

	// Array with size constraints
	schema := core.NewArray().
		Range(2, 5). // Between 2 and 5 items
		Description("Array with size constraints (2-5 items)").
		Build()

	testArrays := []any{
		[]any{"a"},                          // Too short
		[]any{"a", "b"},                     // Valid (minimum)
		[]any{"a", "b", "c"},                // Valid (middle)
		[]any{"a", "b", "c", "d", "e"},      // Valid (maximum)
		[]any{"a", "b", "c", "d", "e", "f"}, // Too long
	}

	for _, arr := range testArrays {
		result := schema.Validate(arr)
		fmt.Printf("Array %v: %s\n", arr, validationStatus(result))
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}

	fmt.Println()
}

func uniqueItemsExample() {
	fmt.Println("--- Unique Items Arrays ---")

	// Array requiring unique items
	schema := core.NewArray().
		UniqueItems().
		Description("Array with unique items only").
		Build()

	testArrays := []any{
		[]any{"a", "b", "c"},      // Valid - all unique
		[]any{1, 2, 3, 4},         // Valid - all unique numbers
		[]any{"a", 1, true, 3.14}, // Valid - different types
		[]any{"a", "b", "a"},      // Invalid - duplicate string
		[]any{1, 2, 1},            // Invalid - duplicate number
		[]any{true, false, true},  // Invalid - duplicate boolean
	}

	for _, arr := range testArrays {
		result := schema.Validate(arr)
		fmt.Printf("Array %v: %s\n", arr, validationStatus(result))
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}

	fmt.Println()
}

func containsExample() {
	fmt.Println("--- Contains Validation ---")

	// Array must contain at least one number > 10
	containsSchema := core.NewNumber().Min(10).Build()
	arraySchema := core.NewArray().
		Contains(containsSchema).
		Description("Array must contain at least one number >= 10").
		Build()

	testArrays := []any{
		[]any{1, 2, 15},         // Valid - contains 15
		[]any{5, 20, 3},         // Valid - contains 20
		[]any{100},              // Valid - single qualifying number
		[]any{1, 2, 3},          // Invalid - no number >= 10
		[]any{"hello", "world"}, // Invalid - no numbers at all
		[]any{1, 2, 9.9},        // Invalid - 9.9 < 10
	}

	for _, arr := range testArrays {
		result := arraySchema.Validate(arr)
		fmt.Printf("Array %v: %s\n", arr, validationStatus(result))
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}

	fmt.Println()
}

func nestedArraysExample() {
	fmt.Println("--- Nested Arrays ---")

	// Array of arrays of strings
	innerArraySchema := core.NewArray().
		Items(core.NewString().Build()).
		Range(1, 3). // Each inner array has 1-3 strings
		Build()

	outerArraySchema := core.NewArray().
		Items(innerArraySchema).
		Range(2, 4). // 2-4 inner arrays
		Description("Array of string arrays").
		Build()

	testArrays := []any{
		[]any{
			[]any{"a", "b"},
			[]any{"c", "d", "e"},
		}, // Valid
		[]any{
			[]any{"a"},
			[]any{"b", "c"},
			[]any{"d", "e", "f"},
			[]any{"g"},
		}, // Valid - maximum outer size
		[]any{
			[]any{"a", "b", "c", "d"}, // Invalid - inner array too long
			[]any{"e", "f"},
		},
		[]any{
			[]any{"a", "b"},
			[]any{123, "c"}, // Invalid - contains non-string
		},
	}

	for i, arr := range testArrays {
		result := outerArraySchema.Validate(arr)
		fmt.Printf("Nested array %d: %s\n", i+1, validationStatus(result))
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error at %s: %s\n", err.Path, err.Message)
			}
		}
	}

	fmt.Println()
}

func commonPatternsExample() {
	fmt.Println("--- Common Array Patterns ---")

	// String list
	stringListSchema := core.NewArray().
		StringArray().
		NonEmpty().
		Build()

	// Set of unique IDs
	idSetSchema := core.NewArray().
		Items(core.NewInteger().ID().Build()).
		Set(). // Unique items
		Build()

	// Fixed-size tuple
	tupleSchema := core.NewArray().
		Tuple(3).
		Description("3-element tuple").
		Build()

	// Limited list
	limitedListSchema := core.NewArray().
		LimitedList(10).
		Description("List with max 10 items").
		Build()

	fmt.Println("Limited list validation:")
	limitedTestData := []any{
		[]any{"a", "b", "c"}, // Valid
		make([]any, 15),      // Invalid - too long
	}
	for _, data := range limitedTestData {
		result := limitedListSchema.Validate(data)
		fmt.Printf("  %v (len=%d): %s\n", data, len(data.([]any)), validationStatus(result))
	}

	fmt.Println("String list validation:")
	testData := []any{
		[]any{"hello", "world"}, // Valid
		[]any{},                 // Invalid - empty (NonEmpty requirement)
		[]any{"hello", 123},     // Invalid without item schema enforcement
	}
	for _, data := range testData {
		result := stringListSchema.Validate(data)
		fmt.Printf("  %v: %s\n", data, validationStatus(result))
	}

	fmt.Println("ID set validation:")
	testData = []any{
		[]any{int64(1), int64(2), int64(3)},  // Valid unique IDs
		[]any{int64(1), int64(2), int64(1)},  // Invalid - duplicate
		[]any{int64(1), "not-int", int64(3)}, // Invalid - non-integer
	}
	for _, data := range testData {
		result := idSetSchema.Validate(data)
		fmt.Printf("  %v: %s\n", data, validationStatus(result))
	}

	fmt.Println("Tuple validation:")
	testData = []any{
		[]any{"a", "b", "c"},      // Valid - exactly 3 items
		[]any{"a", "b"},           // Invalid - too few
		[]any{"a", "b", "c", "d"}, // Invalid - too many
	}
	for _, data := range testData {
		result := tupleSchema.Validate(data)
		fmt.Printf("  %v: %s\n", data, validationStatus(result))
	}

	fmt.Println()
}

func complexCompositionExample() {
	fmt.Println("--- Complex Array Composition ---")

	// Array of user objects (simulated with strings for now)
	// Each user must have a name and email pattern
	userSchema := core.NewString().
		Pattern("^[^:]+:[^@]+@[^@]+$"). // Simple name:email pattern
		Description("User in format 'name:email'").
		Build()

	usersArraySchema := core.NewArray().
		Items(userSchema).
		UniqueItems(). // No duplicate users
		Range(1, 5).   // 1-5 users
		Description("Array of unique users").
		Example([]any{"john:john@example.com", "jane:jane@example.com"}).
		Build()

	testArrays := []any{
		[]any{"john:john@example.com", "jane:jane@example.com"}, // Valid
		[]any{"john:john@example.com"},                          // Valid - single user
		[]any{},                                                 // Invalid - empty (below minimum)
		[]any{"john:john@example.com", "jane:jane@example.com", "bob:bob@example.com",
			"alice:alice@example.com", "charlie:charlie@example.com", "david:david@example.com"}, // Invalid - too many
		[]any{"john:john@example.com", "invalid-format"},        // Invalid - bad format
		[]any{"john:john@example.com", "john:john@example.com"}, // Invalid - duplicate
	}

	for i, arr := range testArrays {
		result := usersArraySchema.Validate(arr)
		fmt.Printf("Users array %d: %s\n", i+1, validationStatus(result))
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error at %s: %s\n", err.Path, err.Message)
			}
		}
	}

	// Show JSON Schema output
	fmt.Println("JSON Schema representation:")
	jsonSchema := usersArraySchema.ToJSONSchema()
	fmt.Printf("%+v\n", jsonSchema)

	// Show generated example
	fmt.Println("Generated example:")
	example := usersArraySchema.GenerateExample()
	fmt.Printf("%v\n", example)

	fmt.Println()
}

func validationStatus(result any) string {
	switch r := result.(type) {
	case interface{ Valid() bool }:
		if r.Valid() {
			return "✅ Valid"
		}
		return "❌ Invalid"
	default:
		// Handle ValidationResult struct
		if result, ok := result.(struct {
			Valid  bool
			Errors []interface{}
		}); ok {
			if result.Valid {
				return "✅ Valid"
			}
			return "❌ Invalid"
		}
		return "❌ Invalid"
	}
}

// Usage function that can be called from main or tests
func RunArrayExamples() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Array examples panicked: %v", r)
		}
	}()

	ArrayExamples()
}
