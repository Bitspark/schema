package schema

import (
	"testing"
)

func TestDefaultValueGeneration(t *testing.T) {
	generator := NewDefaultValueGenerator()

	t.Run("DefaultString", func(t *testing.T) {
		schema := String().Build()
		result := generator.Generate(schema)

		if result != "" {
			t.Errorf("Expected empty string, got: %v", result)
		}
	})

	t.Run("DefaultStringWithMinLength", func(t *testing.T) {
		schema := String().MinLength(5).Build()
		result := generator.Generate(schema)
		resultStr := result.(string)

		if len(resultStr) != 5 {
			t.Errorf("Expected string of length 5, got: %q (length %d)", resultStr, len(resultStr))
		}
		if resultStr != "aaaaa" {
			t.Errorf("Expected 'aaaaa', got: %q", resultStr)
		}
	})

	t.Run("DefaultStringEnum", func(t *testing.T) {
		schema := String().Enum("option1", "option2", "option3").Build()
		result := generator.Generate(schema)

		if result != "option1" {
			t.Errorf("Expected first enum value 'option1', got: %v", result)
		}
	})

	t.Run("DefaultNumber", func(t *testing.T) {
		schema := Number().Build()
		result := generator.Generate(schema)

		if result != 0.0 {
			t.Errorf("Expected 0.0, got: %v", result)
		}
	})

	t.Run("DefaultNumberWithMinimum", func(t *testing.T) {
		schema := Number().Range(10.0, 100.0).Build()
		result := generator.Generate(schema)

		if result != 10.0 {
			t.Errorf("Expected 10.0 (minimum), got: %v", result)
		}
	})

	t.Run("DefaultInteger", func(t *testing.T) {
		schema := Integer().Build()
		result := generator.Generate(schema)

		if result != int64(0) {
			t.Errorf("Expected 0, got: %v", result)
		}
	})

	t.Run("DefaultIntegerWithMinimum", func(t *testing.T) {
		schema := Integer().Range(5, 20).Build()
		result := generator.Generate(schema)

		if result != int64(5) {
			t.Errorf("Expected 5 (minimum), got: %v", result)
		}
	})

	t.Run("DefaultBoolean", func(t *testing.T) {
		schema := Boolean().Build()
		result := generator.Generate(schema)

		if result != false {
			t.Errorf("Expected false, got: %v", result)
		}
	})

	t.Run("DefaultArray", func(t *testing.T) {
		schema := Array().Items(String().Build()).Build()
		result := generator.Generate(schema)
		resultArray := result.([]any)

		if len(resultArray) != 0 {
			t.Errorf("Expected empty array, got array with %d items: %v", len(resultArray), resultArray)
		}
	})

	t.Run("DefaultArrayWithMinItems", func(t *testing.T) {
		schema := Array().Items(String().Build()).MinItems(2).Build()
		result := generator.Generate(schema)
		resultArray := result.([]any)

		if len(resultArray) != 2 {
			t.Errorf("Expected array with 2 items, got %d items: %v", len(resultArray), resultArray)
		}

		// Check that all items are empty strings
		for i, item := range resultArray {
			if item != "" {
				t.Errorf("Expected empty string at index %d, got: %v", i, item)
			}
		}
	})

	t.Run("DefaultObject", func(t *testing.T) {
		schema := Object().
			Property("name", String().Build()).
			Property("age", Integer().Build()).
			Required("name").
			Build()

		result := generator.Generate(schema)
		resultObj := result.(map[string]any)

		// Should only have required properties
		if len(resultObj) != 1 {
			t.Errorf("Expected object with 1 property, got %d properties: %v", len(resultObj), resultObj)
		}

		// Check required property
		if name, exists := resultObj["name"]; !exists || name != "" {
			t.Errorf("Expected required property 'name' to be empty string, got: %v", name)
		}

		// Optional property should not be present
		if _, exists := resultObj["age"]; exists {
			t.Errorf("Expected optional property 'age' to be absent, but it was present: %v", resultObj)
		}
	})
}

func TestMinimalGeneration(t *testing.T) {
	generator := NewMinimalGenerator()

	t.Run("MinimalObject", func(t *testing.T) {
		schema := Object().
			Property("required1", String().Build()).
			Property("required2", Integer().Build()).
			Property("optional1", String().Build()).
			Property("optional2", Boolean().Build()).
			Required("required1", "required2").
			Build()

		result := generator.Generate(schema)
		resultObj := result.(map[string]any)

		// Should only have required properties
		if len(resultObj) != 2 {
			t.Errorf("Expected object with 2 properties, got %d properties: %v", len(resultObj), resultObj)
		}

		// Check required properties exist
		if _, exists := resultObj["required1"]; !exists {
			t.Errorf("Expected required property 'required1' to exist")
		}
		if _, exists := resultObj["required2"]; !exists {
			t.Errorf("Expected required property 'required2' to exist")
		}

		// Optional properties should not be present
		if _, exists := resultObj["optional1"]; exists {
			t.Errorf("Expected optional property 'optional1' to be absent")
		}
		if _, exists := resultObj["optional2"]; exists {
			t.Errorf("Expected optional property 'optional2' to be absent")
		}
	})

	t.Run("MinimalArray", func(t *testing.T) {
		schema := Array().Items(String().Build()).MinItems(0).MaxItems(10).Build()
		result := generator.Generate(schema)
		resultArray := result.([]any)

		// Should generate minimum number of items (0)
		if len(resultArray) != 0 {
			t.Errorf("Expected array with 0 items, got %d items: %v", len(resultArray), resultArray)
		}
	})

	t.Run("MinimalArrayWithRequiredItems", func(t *testing.T) {
		schema := Array().Items(String().Build()).MinItems(3).MaxItems(10).Build()
		result := generator.Generate(schema)
		resultArray := result.([]any)

		// Should generate minimum required number of items (3)
		if len(resultArray) != 3 {
			t.Errorf("Expected array with 3 items, got %d items: %v", len(resultArray), resultArray)
		}
	})
}

func TestCustomDefaultValues(t *testing.T) {
	config := DefaultGeneratorConfig()
	config.GenerateDefaults = true
	config.DefaultValues.String = "custom"
	config.DefaultValues.Number = 42.0
	config.DefaultValues.Integer = 123
	config.DefaultValues.Boolean = true

	generator := NewGenerator(config)

	t.Run("CustomStringDefault", func(t *testing.T) {
		schema := String().Build()
		result := generator.Generate(schema)

		if result != "custom" {
			t.Errorf("Expected 'custom', got: %v", result)
		}
	})

	t.Run("CustomNumberDefault", func(t *testing.T) {
		schema := Number().Build()
		result := generator.Generate(schema)

		if result != 42.0 {
			t.Errorf("Expected 42.0, got: %v", result)
		}
	})

	t.Run("CustomIntegerDefault", func(t *testing.T) {
		schema := Integer().Build()
		result := generator.Generate(schema)

		if result != int64(123) {
			t.Errorf("Expected 123, got: %v", result)
		}
	})

	t.Run("CustomBooleanDefault", func(t *testing.T) {
		schema := Boolean().Build()
		result := generator.Generate(schema)

		if result != true {
			t.Errorf("Expected true, got: %v", result)
		}
	})
}

func TestDefaultsWithTemplates(t *testing.T) {
	// This test shows how default generation can be used with templates
	generator := NewDefaultValueGenerator()

	schema := Object().
		Property("username", String().MinLength(3).MaxLength(20).Build()).
		Property("email", String().MinLength(5).Build()).
		Property("age", Integer().Range(18, 100).Build()).
		Property("active", Boolean().Build()).
		Property("tags", Array().Items(String().Build()).MinItems(0).Build()).
		Required("username", "email").
		Build()

	result := generator.Generate(schema)
	resultObj := result.(map[string]any)

	t.Logf("Generated default user object: %+v", resultObj)

	// Verify structure
	if username, exists := resultObj["username"]; !exists {
		t.Errorf("Expected 'username' to exist")
	} else if usernameStr := username.(string); len(usernameStr) < 3 {
		t.Errorf("Expected username to be at least 3 characters, got: %q", usernameStr)
	}

	if email, exists := resultObj["email"]; !exists {
		t.Errorf("Expected 'email' to exist")
	} else if emailStr := email.(string); len(emailStr) < 5 {
		t.Errorf("Expected email to be at least 5 characters, got: %q", emailStr)
	}

	// Optional fields should not be present in default generation
	if _, exists := resultObj["age"]; exists {
		t.Errorf("Expected optional 'age' field to be absent")
	}
	if _, exists := resultObj["active"]; exists {
		t.Errorf("Expected optional 'active' field to be absent")
	}
	if _, exists := resultObj["tags"]; exists {
		t.Errorf("Expected optional 'tags' field to be absent")
	}
}
