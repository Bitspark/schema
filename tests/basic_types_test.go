package tests

import (
	builders2 "defs.dev/schema/builders"
	"defs.dev/schema/consumers/validation"
	"math"
	"testing"

	json2 "encoding/json"

	"defs.dev/schema/visit/export/json"
)

func TestNumberSchema(t *testing.T) {
	t.Run("Basic validation", func(t *testing.T) {
		schema := builders2.NewNumberSchema().Build()

		// Valid numbers
		validNumbers := []any{
			42.0, 3.14, float32(2.5), int(10), int64(100),
			uint(5), uint32(7), uint64(123),
		}

		for _, num := range validNumbers {
			result := validation.ValidateValue(schema, num)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", num, result.Errors)
			}
		}

		// Invalid values
		result := validation.ValidateValue(schema, "not a number")
		if result.Valid {
			t.Error("Expected string to be invalid for number schema")
		}
	})

	t.Run("Min/Max constraints", func(t *testing.T) {
		schema := builders2.NewNumberSchema().Range(0.0, 100.0).Build()

		// Valid range
		result := validation.ValidateValue(schema, 50.0)
		if !result.Valid {
			t.Errorf("Expected 50.0 to be valid, got errors: %v", result.Errors)
		}

		// Below minimum
		result = validation.ValidateValue(schema, -10.0)
		if result.Valid {
			t.Error("Expected -10.0 to be invalid (below minimum)")
		}

		// Above maximum
		result = validation.ValidateValue(schema, 150.0)
		if result.Valid {
			t.Error("Expected 150.0 to be invalid (above maximum)")
		}
	})

	t.Run("Special float values", func(t *testing.T) {
		schema := builders2.NewNumberSchema().Build()

		// NaN should be invalid
		result := validation.ValidateValue(schema, math.NaN())
		if result.Valid {
			t.Error("Expected NaN to be invalid")
		}

		// Infinity should be invalid
		result = validation.ValidateValue(schema, math.Inf(1))
		if result.Valid {
			t.Error("Expected +Inf to be invalid")
		}

		result = validation.ValidateValue(schema, math.Inf(-1))
		if result.Valid {
			t.Error("Expected -Inf to be invalid")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := builders2.NewNumberSchema().
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
		schema := builders2.NewIntegerSchema().Build()

		// Valid integers
		validIntegers := []any{
			42, int64(100), int32(50), int16(25), int8(12),
			uint(5), uint32(7), uint16(3), uint8(1),
		}

		for _, num := range validIntegers {
			result := validation.ValidateValue(schema, num)
			if !result.Valid {
				t.Errorf("Expected %v to be valid, got errors: %v", num, result.Errors)
			}
		}

		// Valid floats that are whole numbers
		result := validation.ValidateValue(schema, 42.0)
		if !result.Valid {
			t.Errorf("Expected 42.0 to be valid, got errors: %v", result.Errors)
		}

		// Invalid decimal numbers
		result = validation.ValidateValue(schema, 42.5)
		if result.Valid {
			t.Error("Expected 42.5 to be invalid (not a whole number)")
		}

		// Invalid types
		result = validation.ValidateValue(schema, "not a number")
		if result.Valid {
			t.Error("Expected string to be invalid for integer schema")
		}
	})

	t.Run("Min/Max constraints", func(t *testing.T) {
		schema := builders2.NewIntegerSchema().Range(0, 100).Build()

		// Valid range
		result := validation.ValidateValue(schema, 50)
		if !result.Valid {
			t.Errorf("Expected 50 to be valid, got errors: %v", result.Errors)
		}

		// Below minimum
		result = validation.ValidateValue(schema, -10)
		if result.Valid {
			t.Error("Expected -10 to be invalid (below minimum)")
		}

		// Above maximum
		result = validation.ValidateValue(schema, 150)
		if result.Valid {
			t.Error("Expected 150 to be invalid (above maximum)")
		}
	})

	t.Run("Overflow handling", func(t *testing.T) {
		schema := builders2.NewIntegerSchema().Build()

		// Large uint64 values should be valid for integer schemas
		result := validation.ValidateValue(schema, uint64(math.MaxUint64))
		if !result.Valid {
			t.Errorf("Expected MaxUint64 to be valid, got errors: %v", result.Errors)
		}

		// Test that non-integer values are still rejected
		result = validation.ValidateValue(schema, 42.5)
		if result.Valid {
			t.Error("Expected non-integer float to be invalid")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := builders2.NewIntegerSchema().
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
		schema := builders2.NewBooleanSchema().Build()

		// Valid booleans
		result := validation.ValidateValue(schema, true)
		if !result.Valid {
			t.Errorf("Expected true to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(schema, false)
		if !result.Valid {
			t.Errorf("Expected false to be valid, got errors: %v", result.Errors)
		}

		// Invalid types
		result = validation.ValidateValue(schema, "true")
		if result.Valid {
			t.Error("Expected string 'true' to be invalid for boolean schema")
		}

		result = validation.ValidateValue(schema, 1)
		if result.Valid {
			t.Error("Expected integer 1 to be invalid for boolean schema")
		}
	})

	t.Run("JSON Schema generation", func(t *testing.T) {
		schema := builders2.NewBooleanSchema().
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
		schema := builders2.NewNumberSchema().Positive().Build()
		result := validation.ValidateValue(schema, 10.0)
		if !result.Valid {
			t.Errorf("Expected positive number to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(schema, -5.0)
		if result.Valid {
			t.Error("Expected negative number to be invalid for positive schema")
		}

		// Test percentage
		percentSchema := builders2.NewNumberSchema().Percentage().Build()
		result = validation.ValidateValue(percentSchema, 50.0)
		if !result.Valid {
			t.Errorf("Expected 50%% to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(percentSchema, 150.0)
		if result.Valid {
			t.Error("Expected 150%% to be invalid for percentage schema")
		}
	})

	t.Run("Integer helpers", func(t *testing.T) {
		// Test positive integer
		schema := builders2.NewIntegerSchema().Positive().Build()
		result := validation.ValidateValue(schema, 10)
		if !result.Valid {
			t.Errorf("Expected positive integer to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(schema, -5)
		if result.Valid {
			t.Error("Expected negative integer to be invalid for positive schema")
		}
	})

	t.Run("String helpers", func(t *testing.T) {
		// Test email
		emailSchema := builders2.NewStringSchema().Email().Build()
		result := validation.ValidateValue(emailSchema, "test@example.com")
		if !result.Valid {
			t.Errorf("Expected valid email to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(emailSchema, "not-an-email")
		if result.Valid {
			t.Error("Expected invalid email to be invalid")
		}

		// Test UUID
		uuidSchema := builders2.NewStringSchema().UUID().Build()
		result = validation.ValidateValue(uuidSchema, "550e8400-e29b-41d4-a716-446655440000")
		if !result.Valid {
			t.Errorf("Expected valid UUID to be valid, got errors: %v", result.Errors)
		}

		result = validation.ValidateValue(uuidSchema, "not-a-uuid")
		if result.Valid {
			t.Error("Expected invalid UUID to be invalid")
		}
	})
}

func TestNumberSchemaBasic(t *testing.T) {
	schema := builders2.NewNumberSchema().
		Min(0).
		Max(100).
		Build()

	// Test valid number
	result := validation.ValidateValue(schema, 50.0)
	if !result.Valid {
		t.Errorf("Expected 50.0 to be valid, got errors: %v", result.Errors)
	}

	// Test invalid number (too small)
	result = validation.ValidateValue(schema, -10.0)
	if result.Valid {
		t.Error("Expected -10.0 to be invalid (too small)")
	}

	// Test invalid number (too large)
	result = validation.ValidateValue(schema, 150.0)
	if result.Valid {
		t.Error("Expected 150.0 to be invalid (too large)")
	}

	// Test wrong type
	result = validation.ValidateValue(schema, "not a number")
	if result.Valid {
		t.Error("Expected string to be invalid (wrong type)")
	}
}

func TestNumberSchemaSpecialValues(t *testing.T) {
	schema := builders2.NewNumberSchema().Build()

	// Test NaN
	result := validation.ValidateValue(schema, math.NaN())
	if result.Valid {
		t.Error("Expected NaN to be invalid")
	}

	// Test positive infinity
	result = validation.ValidateValue(schema, math.Inf(1))
	if result.Valid {
		t.Error("Expected +Inf to be invalid")
	}

	// Test negative infinity
	result = validation.ValidateValue(schema, math.Inf(-1))
	if result.Valid {
		t.Error("Expected -Inf to be invalid")
	}
}

func TestNumberSchemaConstraints(t *testing.T) {
	tests := []struct {
		name     string
		min      *float64
		max      *float64
		value    float64
		expected bool
	}{
		{"valid in range", ptr(0.0), ptr(100.0), 50.0, true},
		{"at minimum", ptr(0.0), ptr(100.0), 0.0, true},
		{"at maximum", ptr(0.0), ptr(100.0), 100.0, true},
		{"below minimum", ptr(0.0), ptr(100.0), -1.0, false},
		{"above maximum", ptr(0.0), ptr(100.0), 101.0, false},
		{"no constraints", nil, nil, 999.0, true},
		{"only min constraint", ptr(10.0), nil, 15.0, true},
		{"only min constraint violated", ptr(10.0), nil, 5.0, false},
		{"only max constraint", nil, ptr(50.0), 25.0, true},
		{"only max constraint violated", nil, ptr(50.0), 75.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := builders2.NewNumberSchema()
			if tt.min != nil {
				builder = builder.Min(*tt.min)
			}
			if tt.max != nil {
				builder = builder.Max(*tt.max)
			}
			schema := builder.Build()

			result := validation.ValidateValue(schema, tt.value)
			if result.Valid != tt.expected {
				t.Errorf("Expected valid=%v for value %v, got %v", tt.expected, tt.value, result.Valid)
			}
		})
	}
}

func TestIntegerSchemaBasic(t *testing.T) {
	schema := builders2.NewIntegerSchema().
		Min(0).
		Max(100).
		Build()

	// Test valid integer
	result := validation.ValidateValue(schema, 50)
	if !result.Valid {
		t.Errorf("Expected 50 to be valid, got errors: %v", result.Errors)
	}

	// Test valid float that's an integer
	result = validation.ValidateValue(schema, 42.0)
	if !result.Valid {
		t.Errorf("Expected 42.0 to be valid, got errors: %v", result.Errors)
	}

	// Test invalid float (not an integer)
	result = validation.ValidateValue(schema, 42.5)
	if result.Valid {
		t.Error("Expected 42.5 to be invalid (not an integer)")
	}

	// Test wrong type
	result = validation.ValidateValue(schema, "not a number")
	if result.Valid {
		t.Error("Expected string to be invalid (wrong type)")
	}
}

func TestIntegerSchemaConstraints(t *testing.T) {
	schema := builders2.NewIntegerSchema().
		Min(0).
		Max(100).
		Build()

	// Test valid integer
	result := validation.ValidateValue(schema, 50)
	if !result.Valid {
		t.Errorf("Expected 50 to be valid, got errors: %v", result.Errors)
	}

	// Test invalid integer (too small)
	result = validation.ValidateValue(schema, -10)
	if result.Valid {
		t.Error("Expected -10 to be invalid (too small)")
	}

	// Test invalid integer (too large)
	result = validation.ValidateValue(schema, 150)
	if result.Valid {
		t.Error("Expected 150 to be invalid (too large)")
	}
}

func TestIntegerSchemaLargeValues(t *testing.T) {
	schema := builders2.NewIntegerSchema().Build()

	// Test large uint64 value
	result := validation.ValidateValue(schema, uint64(math.MaxUint64))
	if !result.Valid {
		t.Errorf("Expected large uint64 to be valid, got errors: %v", result.Errors)
	}
}

func TestBooleanSchemaBasic(t *testing.T) {
	schema := builders2.NewBooleanSchema().Build()

	// Test valid boolean values
	result := validation.ValidateValue(schema, true)
	if !result.Valid {
		t.Errorf("Expected true to be valid, got errors: %v", result.Errors)
	}

	result = validation.ValidateValue(schema, false)
	if !result.Valid {
		t.Errorf("Expected false to be valid, got errors: %v", result.Errors)
	}

	// Test invalid values
	result = validation.ValidateValue(schema, "true")
	if result.Valid {
		t.Error("Expected string 'true' to be invalid")
	}

	result = validation.ValidateValue(schema, 1)
	if result.Valid {
		t.Error("Expected integer 1 to be invalid")
	}
}

func TestBooleanSchemaValidation(t *testing.T) {
	schema := builders2.NewBooleanSchema().Build()

	tests := []struct {
		input    any
		expected bool
	}{
		{true, true},
		{false, true},
		{"true", false},
		{"false", false},
		{1, false},
		{0, false},
		{nil, false},
	}

	for _, test := range tests {
		result := validation.ValidateValue(schema, test.input)
		if result.Valid != test.expected {
			t.Errorf("For input %v, expected valid=%v, got %v", test.input, test.expected, result.Valid)
		}
	}

	// Test with a non-boolean value that should fail
	result := validation.ValidateValue(schema, "maybe")
	if result.Valid {
		t.Error("Expected 'maybe' to be invalid")
	}
}

func TestNumberSchemaWithPercentage(t *testing.T) {
	// Create a percentage schema (0-100)
	schema := builders2.NewNumberSchema().
		Min(0).
		Max(100).
		Description("Percentage value").
		Build()

	// Test valid percentage
	result := validation.ValidateValue(schema, 50.0)
	if !result.Valid {
		t.Errorf("Expected 50.0 to be valid, got errors: %v", result.Errors)
	}

	// Test invalid percentage (negative)
	result = validation.ValidateValue(schema, -5.0)
	if result.Valid {
		t.Error("Expected -5.0 to be invalid (negative percentage)")
	}

	// Test another percentage schema with different constraints
	percentSchema := builders2.NewNumberSchema().Min(0).Max(100).Build()

	result = validation.ValidateValue(percentSchema, 50.0)
	if !result.Valid {
		t.Errorf("Expected 50.0 to be valid for percent schema, got errors: %v", result.Errors)
	}

	result = validation.ValidateValue(percentSchema, 150.0)
	if result.Valid {
		t.Error("Expected 150.0 to be invalid (over 100%)")
	}
}

func TestIntegerSchemaWithAge(t *testing.T) {
	// Create an age schema (0-120)
	schema := builders2.NewIntegerSchema().
		Min(0).
		Max(120).
		Description("Age in years").
		Build()

	// Test valid age
	result := validation.ValidateValue(schema, 25)
	if !result.Valid {
		t.Errorf("Expected 25 to be valid, got errors: %v", result.Errors)
	}

	// Test invalid age (negative)
	result = validation.ValidateValue(schema, -5)
	if result.Valid {
		t.Error("Expected -5 to be invalid (negative age)")
	}
}

func TestStringSchemaWithFormats(t *testing.T) {
	// Test email format
	emailSchema := builders2.NewStringSchema().Email().Build()

	result := validation.ValidateValue(emailSchema, "test@example.com")
	if !result.Valid {
		t.Errorf("Expected 'test@example.com' to be valid, got errors: %v", result.Errors)
	}

	result = validation.ValidateValue(emailSchema, "not-an-email")
	if result.Valid {
		t.Error("Expected 'not-an-email' to be invalid")
	}

	// Test UUID format
	uuidSchema := builders2.NewStringSchema().UUID().Build()

	result = validation.ValidateValue(uuidSchema, "550e8400-e29b-41d4-a716-446655440000")
	if !result.Valid {
		t.Errorf("Expected UUID to be valid, got errors: %v", result.Errors)
	}

	result = validation.ValidateValue(uuidSchema, "not-a-uuid")
	if result.Valid {
		t.Error("Expected 'not-a-uuid' to be invalid")
	}
}

// Helper function to create a pointer to a float64
func ptr(f float64) *float64 {
	return &f
}
