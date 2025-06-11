package tests

import (
	"math"
	"testing"

	"defs.dev/schema/core"
)

func TestNumberSchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := core.NewNumber().Build()

		// Valid numbers
		validNumbers := []any{
			42.0, 3.14, float32(2.5), int(10), int64(100),
			uint(5), uint32(7), uint64(123),
		}

		for _, num := range validNumbers {
			result := schema.Validate(num)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", num, result.Errors)
			}
		}

		// Invalid values
		result := schema.Validate("not a number")
		if result.Valid {
			t.Error("Expected string to be invalid for number schema")
		}
	})

	t.Run("Min/Max constraints", func(t *testing.T) {
		schema := core.NewNumber().Range(0.0, 100.0).Build()

		// Valid range
		result := schema.Validate(50.0)
		if !result.Valid {
			t.Errorf("Expected 50.0 to be valid, got errors: %v", result.Errors)
		}

		// Below minimum
		result = schema.Validate(-10.0)
		if result.Valid {
			t.Error("Expected -10.0 to be invalid (below minimum)")
		}

		// Above maximum
		result = schema.Validate(150.0)
		if result.Valid {
			t.Error("Expected 150.0 to be invalid (above maximum)")
		}
	})

	t.Run("Special float values", func(t *testing.T) {
		schema := core.NewNumber().Build()

		// NaN should be invalid
		result := schema.Validate(math.NaN())
		if result.Valid {
			t.Error("Expected NaN to be invalid")
		}

		// Infinity should be invalid
		result = schema.Validate(math.Inf(1))
		if result.Valid {
			t.Error("Expected +Inf to be invalid")
		}

		result = schema.Validate(math.Inf(-1))
		if result.Valid {
			t.Error("Expected -Inf to be invalid")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := core.NewNumber().
			Range(0.0, 100.0).
			Description("Test number").
			Example(42.0).
			Build()

		jsonSchema := schema.ToJSONSchema()

		if jsonSchema["type"] != "number" {
			t.Errorf("Expected type 'number', got %v", jsonSchema["type"])
		}
		if jsonSchema["minimum"] != 0.0 {
			t.Errorf("Expected minimum 0.0, got %v", jsonSchema["minimum"])
		}
		if jsonSchema["maximum"] != 100.0 {
			t.Errorf("Expected maximum 100.0, got %v", jsonSchema["maximum"])
		}
		if jsonSchema["description"] != "Test number" {
			t.Errorf("Expected description 'Test number', got %v", jsonSchema["description"])
		}
	})
}

func TestIntegerSchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := core.NewInteger().Build()

		// Valid integers
		validIntegers := []any{
			42, int64(100), int32(50), int16(25), int8(12),
			uint(5), uint32(7), uint16(3), uint8(1),
		}

		for _, num := range validIntegers {
			result := schema.Validate(num)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", num, result.Errors)
			}
		}

		// Valid floats that are whole numbers
		result := schema.Validate(42.0)
		if !result.Valid {
			t.Errorf("Expected 42.0 to be valid, got errors: %v", result.Errors)
		}

		// Invalid decimal numbers
		result = schema.Validate(42.5)
		if result.Valid {
			t.Error("Expected 42.5 to be invalid (not a whole number)")
		}

		// Invalid types
		result = schema.Validate("not a number")
		if result.Valid {
			t.Error("Expected string to be invalid for integer schema")
		}
	})

	t.Run("Min/Max constraints", func(t *testing.T) {
		schema := core.NewInteger().Range(0, 100).Build()

		// Valid range
		result := schema.Validate(50)
		if !result.Valid {
			t.Errorf("Expected 50 to be valid, got errors: %v", result.Errors)
		}

		// Below minimum
		result = schema.Validate(-10)
		if result.Valid {
			t.Error("Expected -10 to be invalid (below minimum)")
		}

		// Above maximum
		result = schema.Validate(150)
		if result.Valid {
			t.Error("Expected 150 to be invalid (above maximum)")
		}
	})

	t.Run("Overflow handling", func(t *testing.T) {
		schema := core.NewInteger().Build()

		// Very large uint64 should cause overflow error
		result := schema.Validate(uint64(math.MaxUint64))
		if result.Valid {
			t.Error("Expected MaxUint64 to cause overflow error")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := core.NewInteger().
			Range(1, 100).
			Description("Test integer").
			Example(int64(42)).
			Build()

		jsonSchema := schema.ToJSONSchema()

		if jsonSchema["type"] != "integer" {
			t.Errorf("Expected type 'integer', got %v", jsonSchema["type"])
		}
		if jsonSchema["minimum"] != int64(1) {
			t.Errorf("Expected minimum 1, got %v", jsonSchema["minimum"])
		}
		if jsonSchema["maximum"] != int64(100) {
			t.Errorf("Expected maximum 100, got %v", jsonSchema["maximum"])
		}
	})
}

func TestBooleanSchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := core.NewBoolean().Build()

		// Valid booleans
		result := schema.Validate(true)
		if !result.Valid {
			t.Errorf("Expected true to be valid, got errors: %v", result.Errors)
		}

		result = schema.Validate(false)
		if !result.Valid {
			t.Errorf("Expected false to be valid, got errors: %v", result.Errors)
		}

		// Invalid types (without string conversion)
		result = schema.Validate("true")
		if result.Valid {
			t.Error("Expected string 'true' to be invalid without string conversion")
		}

		result = schema.Validate(1)
		if result.Valid {
			t.Error("Expected integer 1 to be invalid")
		}
	})

	t.Run("String conversion", func(t *testing.T) {
		schema := core.NewBoolean().AllowStringConversion().Build()

		// Valid string conversions
		validStrings := []string{"true", "false", "1", "0"}
		for _, str := range validStrings {
			result := schema.Validate(str)
			if !result.Valid {
				t.Errorf("Expected '%s' to be valid with string conversion, got errors: %v", str, result.Errors)
			}
		}

		// Invalid string
		result := schema.Validate("maybe")
		if result.Valid {
			t.Error("Expected 'maybe' to be invalid")
		}
	})

	t.Run("Case insensitive conversion", func(t *testing.T) {
		schema := core.NewBoolean().CaseInsensitive().Build()

		// Case variations should work
		validStrings := []string{"TRUE", "False", "YES", "no", "ON", "off"}
		for _, str := range validStrings {
			result := schema.Validate(str)
			if !result.Valid {
				t.Errorf("Expected '%s' to be valid with case insensitive conversion, got errors: %v", str, result.Errors)
			}
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := core.NewBoolean().
			Description("Test boolean").
			Example(true).
			AllowStringConversion().
			Build()

		jsonSchema := schema.ToJSONSchema()

		if jsonSchema["type"] != "boolean" {
			t.Errorf("Expected type 'boolean', got %v", jsonSchema["type"])
		}
		if jsonSchema["description"] != "Test boolean" {
			t.Errorf("Expected description 'Test boolean', got %v", jsonSchema["description"])
		}
		if jsonSchema["x-allow-string-conversion"] != true {
			t.Error("Expected x-allow-string-conversion to be true")
		}
	})
}

func TestBuilderHelpers(t *testing.T) {
	t.Run("Number helpers", func(t *testing.T) {
		// Test Positive helper
		schema := core.NewNumber().Positive().Build()
		result := schema.Validate(-1.0)
		if result.Valid {
			t.Error("Expected negative number to be invalid for Positive() schema")
		}

		// Test Percentage helper
		schema = core.NewNumber().Percentage().Build()
		result = schema.Validate(150.0)
		if result.Valid {
			t.Error("Expected 150 to be invalid for Percentage() schema (max 100)")
		}
	})

	t.Run("Integer helpers", func(t *testing.T) {
		// Test Port helper
		schema := core.NewInteger().Port().Build()
		result := schema.Validate(0)
		if result.Valid {
			t.Error("Expected 0 to be invalid for Port() schema (min 1)")
		}

		result = schema.Validate(70000)
		if result.Valid {
			t.Error("Expected 70000 to be invalid for Port() schema (max 65535)")
		}

		result = schema.Validate(8080)
		if !result.Valid {
			t.Errorf("Expected 8080 to be valid for Port() schema, got errors: %v", result.Errors)
		}
	})

	t.Run("Boolean helpers", func(t *testing.T) {
		// Test Switch helper
		schema := core.NewBoolean().Switch().Build()
		result := schema.Validate("TRUE")
		if !result.Valid {
			t.Errorf("Expected 'TRUE' to be valid for Switch() schema, got errors: %v", result.Errors)
		}
	})
}
