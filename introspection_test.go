package schema

import (
	"testing"
)

func TestSchemaIntrospection(t *testing.T) {
	t.Run("ObjectSchema introspection", func(t *testing.T) {
		// Create an object schema
		userSchema := NewObject().
			Property("name", NewString().MinLength(1).Build()).
			Property("age", NewInteger().Min(0).Build()).
			Property("email", NewString().Email().Build()).
			Required("name", "email").
			AdditionalProperties(true).
			Build()

		obj := userSchema.(*ObjectSchema)

		// Test Properties() introspection
		properties := obj.Properties()
		if len(properties) != 3 {
			t.Errorf("Expected 3 properties, got %d", len(properties))
		}

		// Verify we can access individual properties
		nameSchema := properties["name"]
		if nameSchema == nil {
			t.Error("Expected name property to exist")
		}
		if nameSchema.Type() != TypeString {
			t.Errorf("Expected name to be string type, got %v", nameSchema.Type())
		}

		// Test Required() introspection
		required := obj.Required()
		if len(required) != 2 {
			t.Errorf("Expected 2 required fields, got %d", len(required))
		}

		// Check required fields
		hasName, hasEmail := false, false
		for _, field := range required {
			if field == "name" {
				hasName = true
			}
			if field == "email" {
				hasEmail = true
			}
		}
		if !hasName || !hasEmail {
			t.Errorf("Expected 'name' and 'email' to be required, got %v", required)
		}

		// Test AdditionalProperties() introspection
		if !obj.AdditionalProperties() {
			t.Error("Expected additional properties to be true")
		}

		// Verify mutation safety - modifying returned copy shouldn't affect original
		properties["hacker"] = NewString().Build()
		if len(obj.Properties()) != 3 {
			t.Error("Original schema was mutated by external modification")
		}
	})

	t.Run("ArraySchema introspection", func(t *testing.T) {
		// Create an array schema
		listSchema := NewArray().
			Items(NewString().MinLength(1).Build()).
			MinItems(1).
			MaxItems(10).
			UniqueItems().
			Build()

		arr := listSchema.(*ArraySchema)

		// Test ItemSchema() introspection
		itemSchema := arr.ItemSchema()
		if itemSchema == nil {
			t.Error("Expected item schema to exist")
		}
		if itemSchema.Type() != TypeString {
			t.Errorf("Expected item schema to be string type, got %v", itemSchema.Type())
		}

		// Test MinItems() introspection
		minItems := arr.MinItems()
		if minItems == nil || *minItems != 1 {
			t.Errorf("Expected min items to be 1, got %v", minItems)
		}

		// Test MaxItems() introspection
		maxItems := arr.MaxItems()
		if maxItems == nil || *maxItems != 10 {
			t.Errorf("Expected max items to be 10, got %v", maxItems)
		}

		// Test UniqueItemsRequired() introspection
		if !arr.UniqueItemsRequired() {
			t.Error("Expected unique items to be required")
		}
	})

	t.Run("FunctionSchema introspection", func(t *testing.T) {
		// Create a function schema
		calcSchema := NewFunctionSchema().
			Name("calculate").
			Description("Performs calculation").
			Input("x", NewNumber().Build()).
			Input("y", NewNumber().Build()).
			Output(NewNumber().Build()).
			Required("x", "y").
			Build()

		fn := calcSchema.(*FunctionSchema)

		// Test Inputs() introspection
		inputs := fn.Inputs()
		if len(inputs) != 2 {
			t.Errorf("Expected 2 inputs, got %d", len(inputs))
		}

		xSchema := inputs["x"]
		ySchema := inputs["y"]
		if xSchema == nil || ySchema == nil {
			t.Error("Expected x and y input schemas to exist")
		}
		if xSchema.Type() != TypeNumber || ySchema.Type() != TypeNumber {
			t.Error("Expected x and y to be number types")
		}

		// Test Outputs() introspection
		outputs := fn.Outputs()
		if outputs == nil {
			t.Error("Expected output schema to exist")
		}
		if outputs.Type() != TypeNumber {
			t.Errorf("Expected output to be number type, got %v", outputs.Type())
		}

		// Test Required() introspection
		required := fn.Required()
		if len(required) != 2 {
			t.Errorf("Expected 2 required inputs, got %d", len(required))
		}

		// Test Errors() introspection (should be nil in this case)
		errors := fn.Errors()
		if errors != nil {
			t.Error("Expected no error schema to be defined")
		}

		// Verify mutation safety
		inputs["z"] = NewBoolean().Build()
		if len(fn.Inputs()) != 2 {
			t.Error("Original function schema was mutated by external modification")
		}
	})

	t.Run("UnionSchema introspection", func(t *testing.T) {
		// Create a union schema
		unionSchema := Union2[string, int]().
			Name("StringOrInt").
			Description("Either a string or integer").
			Build()

		union := unionSchema.(*UnionSchema)

		// Test Schemas() introspection
		schemas := union.Schemas()
		if len(schemas) != 2 {
			t.Errorf("Expected 2 union schemas, got %d", len(schemas))
		}

		// Check that we have string and number types
		hasString, hasNumber := false, false
		for _, schema := range schemas {
			if schema.Type() == TypeString {
				hasString = true
			}
			if schema.Type() == TypeInteger {
				hasNumber = true
			}
		}
		if !hasString || !hasNumber {
			t.Error("Expected union to contain string and integer schemas")
		}

		// Verify mutation safety
		schemas[0] = NewBoolean().Build()
		if len(union.Schemas()) != 2 {
			t.Error("Original union schema was mutated by external modification")
		}
	})
}
