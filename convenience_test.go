package schema

import (
	"testing"
)

func TestConvenienceDefaultFunctions(t *testing.T) {
	// Define a test schema
	schema := Object().
		Property("name", String().MinLength(3).Build()).
		Property("age", Integer().Range(1, 100).Build()).
		Property("active", Boolean().Build()).
		Property("tags", Array().Items(String().Build()).MinItems(0).Build()).
		Required("name").
		Build()

	t.Run("GenerateDefaults", func(t *testing.T) {
		result := GenerateDefaults(schema)
		resultObj := result.(map[string]any)

		// Should only have required fields with default values
		if len(resultObj) != 1 {
			t.Errorf("Expected 1 property, got %d: %v", len(resultObj), resultObj)
		}

		if name, exists := resultObj["name"]; !exists {
			t.Errorf("Expected 'name' property to exist")
		} else if name != "aaa" {
			t.Errorf("Expected name to be 'aaa', got: %v", name)
		}

		// Optional fields should not be present
		if _, exists := resultObj["age"]; exists {
			t.Errorf("Expected 'age' to be absent in default generation")
		}
		if _, exists := resultObj["active"]; exists {
			t.Errorf("Expected 'active' to be absent in default generation")
		}
		if _, exists := resultObj["tags"]; exists {
			t.Errorf("Expected 'tags' to be absent in default generation")
		}
	})

	t.Run("GenerateMinimal", func(t *testing.T) {
		result := GenerateMinimal(schema)
		resultObj := result.(map[string]any)

		// Should only have required fields
		if len(resultObj) != 1 {
			t.Errorf("Expected 1 property, got %d: %v", len(resultObj), resultObj)
		}

		if _, exists := resultObj["name"]; !exists {
			t.Errorf("Expected 'name' property to exist")
		}

		// Optional fields should not be present
		if _, exists := resultObj["age"]; exists {
			t.Errorf("Expected 'age' to be absent in minimal generation")
		}
	})

	t.Run("GenerateCustomDefaults", func(t *testing.T) {
		result := GenerateCustomDefaults(schema, "custom", 42.5, 123, true)
		resultObj := result.(map[string]any)

		// Should only have required fields with custom default values
		if len(resultObj) != 1 {
			t.Errorf("Expected 1 property, got %d: %v", len(resultObj), resultObj)
		}

		if name, exists := resultObj["name"]; !exists {
			t.Errorf("Expected 'name' property to exist")
		} else if name != "custom" {
			t.Errorf("Expected name to be 'custom', got: %v", name)
		}
	})
}

func TestConvenienceDefaultsWithDifferentTypes(t *testing.T) {
	t.Run("StringDefaults", func(t *testing.T) {
		schema := String().MinLength(5).Build()

		// Default generation
		result := GenerateDefaults(schema)
		if result != "aaaaa" {
			t.Errorf("Expected 'aaaaa', got: %v", result)
		}

		// Custom defaults
		customResult := GenerateCustomDefaults(schema, "hello", 0, 0, false)
		if customResult != "hello" {
			t.Errorf("Expected 'hello', got: %v", customResult)
		}
	})

	t.Run("NumberDefaults", func(t *testing.T) {
		schema := Number().Range(10.0, 100.0).Build()

		// Default generation (should use minimum when default is below range)
		result := GenerateDefaults(schema)
		if result != 10.0 {
			t.Errorf("Expected 10.0 (minimum), got: %v", result)
		}

		// Custom defaults
		customResult := GenerateCustomDefaults(schema, "", 50.0, 0, false)
		if customResult != 50.0 {
			t.Errorf("Expected 50.0, got: %v", customResult)
		}
	})

	t.Run("IntegerDefaults", func(t *testing.T) {
		schema := Integer().Range(5, 20).Build()

		// Default generation (should use minimum when default is below range)
		result := GenerateDefaults(schema)
		if result != int64(5) {
			t.Errorf("Expected 5 (minimum), got: %v", result)
		}

		// Custom defaults
		customResult := GenerateCustomDefaults(schema, "", 0, 15, false)
		if customResult != int64(15) {
			t.Errorf("Expected 15, got: %v", customResult)
		}
	})

	t.Run("BooleanDefaults", func(t *testing.T) {
		schema := Boolean().Build()

		// Default generation
		result := GenerateDefaults(schema)
		if result != false {
			t.Errorf("Expected false, got: %v", result)
		}

		// Custom defaults
		customResult := GenerateCustomDefaults(schema, "", 0, 0, true)
		if customResult != true {
			t.Errorf("Expected true, got: %v", customResult)
		}
	})

	t.Run("ArrayDefaults", func(t *testing.T) {
		schema := Array().Items(String().Build()).MinItems(0).Build()

		// Default generation (should be empty array)
		result := GenerateDefaults(schema)
		resultArray := result.([]any)
		if len(resultArray) != 0 {
			t.Errorf("Expected empty array, got: %v", resultArray)
		}

		// Minimal generation (should also be empty array)
		minimalResult := GenerateMinimal(schema)
		minimalArray := minimalResult.([]any)
		if len(minimalArray) != 0 {
			t.Errorf("Expected empty array, got: %v", minimalArray)
		}
	})
}

func TestDefaultsVsRandomComparison(t *testing.T) {
	schema := Object().
		Property("id", Integer().Range(1, 1000).Build()).
		Property("name", String().MinLength(3).MaxLength(20).Build()).
		Property("score", Number().Range(0.0, 100.0).Build()).
		Required("id", "name", "score").
		Build()

	// Generate random values
	random1 := Generate(schema)
	random2 := Generate(schema)

	// Generate default values
	default1 := GenerateDefaults(schema)
	default2 := GenerateDefaults(schema)

	t.Logf("Random 1: %+v", random1)
	t.Logf("Random 2: %+v", random2)
	t.Logf("Default 1: %+v", default1)
	t.Logf("Default 2: %+v", default2)

	// Random values should be different (high probability)
	// Note: We can't directly compare maps, so we'll just log them

	// Default values should be identical - compare individual fields
	default1Obj := default1.(map[string]any)
	default2Obj := default2.(map[string]any)

	if default1Obj["id"] != default2Obj["id"] {
		t.Errorf("Expected default id values to be identical, got %v vs %v", default1Obj["id"], default2Obj["id"])
	}
	if default1Obj["name"] != default2Obj["name"] {
		t.Errorf("Expected default name values to be identical, got %v vs %v", default1Obj["name"], default2Obj["name"])
	}
	if default1Obj["score"] != default2Obj["score"] {
		t.Errorf("Expected default score values to be identical, got %v vs %v", default1Obj["score"], default2Obj["score"])
	}

	// Default values should be predictable
	if default1Obj["id"] != int64(1) {
		t.Errorf("Expected default id to be 1 (minimum), got: %v", default1Obj["id"])
	}
	if default1Obj["name"] != "aaa" {
		t.Errorf("Expected default name to be 'aaa', got: %v", default1Obj["name"])
	}
	if default1Obj["score"] != 0.0 {
		t.Errorf("Expected default score to be 0.0 (minimum), got: %v", default1Obj["score"])
	}
}
