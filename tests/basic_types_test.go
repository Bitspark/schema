package tests

import (
	"math"
	"testing"

	json2 "encoding/json"

	"defs.dev/schema/builders"
	"defs.dev/schema/export/json"
)

func TestNumberSchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := builders.NewNumberSchema().Build()

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
		schema := builders.NewNumberSchema().Range(0.0, 100.0).Build()

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
		schema := builders.NewNumberSchema().Build()

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
		schema := builders.NewNumberSchema().
			Range(0.0, 100.0).
			Description("Test number").
			Example(42.0).
			Build()

		gen, err := json.NewJSONGenerator()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		jsonSchemaBytes, err := gen.Generate(schema)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		} else {
			t.Logf("JSON Schema: %v", jsonSchemaBytes)
		}

		var jsonSchema map[string]any
		err = json2.Unmarshal(jsonSchemaBytes, &jsonSchema)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

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
		schema := builders.NewIntegerSchema().Build()

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
		schema := builders.NewIntegerSchema().Range(0, 100).Build()

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
		schema := builders.NewIntegerSchema().Build()

		// Very large uint64 should cause overflow error
		result := schema.Validate(uint64(math.MaxUint64))
		if result.Valid {
			t.Error("Expected MaxUint64 to cause overflow error")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := builders.NewIntegerSchema().
			Range(1, 100).
			Description("Test integer").
			Example(int64(42)).
			Build()

		gen, err := json.NewJSONGenerator()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		jsonSchemaBytes, err := gen.Generate(schema)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		} else {
			t.Logf("JSON Schema: %v", jsonSchemaBytes)
		}

		var jsonSchema map[string]any
		err = json2.Unmarshal(jsonSchemaBytes, &jsonSchema)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if jsonSchema["type"] != "integer" {
			t.Errorf("Expected type 'integer', got %v", jsonSchema["type"])
		}
		if jsonSchema["minimum"] != float64(1) {
			t.Errorf("Expected minimum 1, got %v", jsonSchema["minimum"])
		}
		if jsonSchema["maximum"] != float64(100) {
			t.Errorf("Expected maximum 100, got %v", jsonSchema["maximum"])
		}
	})
}

func TestBooleanSchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := builders.NewBooleanSchema().Build()

		// Valid booleans
		result := schema.Validate(true)
		if !result.Valid {
			t.Errorf("Expected true to be valid, got errors: %v", result.Errors)
		}

		result = schema.Validate(false)
		if !result.Valid {
			t.Errorf("Expected false to be valid, got errors: %v", result.Errors)
		}

		// Invalid types
		result = schema.Validate("true")
		if result.Valid {
			t.Error("Expected string 'true' to be invalid for boolean schema")
		}

		result = schema.Validate(1)
		if result.Valid {
			t.Error("Expected integer 1 to be invalid for boolean schema")
		}
	})

	t.Run("String conversion", func(t *testing.T) {
		schema := builders.NewBooleanSchema().AllowStringConversion().Build()

		// Valid string representations
		validStrings := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"True", true},
			{"False", false},
			{"TRUE", true},
			{"FALSE", false},
		}

		for _, test := range validStrings {
			result := schema.Validate(test.input)
			if !result.Valid {
				t.Errorf("Expected '%s' to be valid with string conversion, got errors: %v", test.input, result.Errors)
			}
		}

		// Invalid string representations
		result := schema.Validate("maybe")
		if result.Valid {
			t.Error("Expected 'maybe' to be invalid even with string conversion")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := builders.NewBooleanSchema().
			Description("Test boolean").
			Example(true).
			Build()

		gen, err := json.NewJSONGenerator()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		jsonSchemaBytes, err := gen.Generate(schema)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		var jsonSchema map[string]any
		err = json2.Unmarshal(jsonSchemaBytes, &jsonSchema)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if jsonSchema["type"] != "boolean" {
			t.Errorf("Expected type 'boolean', got %v", jsonSchema["type"])
		}
		if jsonSchema["description"] != "Test boolean" {
			t.Errorf("Expected description 'Test boolean', got %v", jsonSchema["description"])
		}
	})
}

func TestBuilderHelpers(t *testing.T) {
	t.Run("Number helpers", func(t *testing.T) {
		// Test positive number
		schema := builders.NewNumberSchema().Positive().Build()
		result := schema.Validate(10.0)
		if !result.Valid {
			t.Errorf("Expected positive number to be valid, got errors: %v", result.Errors)
		}

		result = schema.Validate(-5.0)
		if result.Valid {
			t.Error("Expected negative number to be invalid for positive schema")
		}

		// Test percentage
		percentSchema := builders.NewNumberSchema().Percentage().Build()
		result = percentSchema.Validate(50.0)
		if !result.Valid {
			t.Errorf("Expected 50%% to be valid, got errors: %v", result.Errors)
		}

		result = percentSchema.Validate(150.0)
		if result.Valid {
			t.Error("Expected 150%% to be invalid for percentage schema")
		}
	})

	t.Run("Integer helpers", func(t *testing.T) {
		// Test positive integer
		schema := builders.NewIntegerSchema().Positive().Build()
		result := schema.Validate(10)
		if !result.Valid {
			t.Errorf("Expected positive integer to be valid, got errors: %v", result.Errors)
		}

		result = schema.Validate(-5)
		if result.Valid {
			t.Error("Expected negative integer to be invalid for positive schema")
		}
	})

	t.Run("String helpers", func(t *testing.T) {
		// Test email
		emailSchema := builders.NewStringSchema().Email().Build()
		result := emailSchema.Validate("test@example.com")
		if !result.Valid {
			t.Errorf("Expected valid email to be valid, got errors: %v", result.Errors)
		}

		result = emailSchema.Validate("not-an-email")
		if result.Valid {
			t.Error("Expected invalid email to be invalid")
		}

		// Test UUID
		uuidSchema := builders.NewStringSchema().UUID().Build()
		result = uuidSchema.Validate("550e8400-e29b-41d4-a716-446655440000")
		if !result.Valid {
			t.Errorf("Expected valid UUID to be valid, got errors: %v", result.Errors)
		}

		result = uuidSchema.Validate("not-a-uuid")
		if result.Valid {
			t.Error("Expected invalid UUID to be invalid")
		}
	})
}
