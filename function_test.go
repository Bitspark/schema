package schema

import (
	"testing"
)

func TestFunctionSchema(t *testing.T) {
	// Test creating a function schema (pure interface description)
	paymentSchema := NewFunctionSchema().
		Input("amount", Number().Min(0).Build()).
		Input("method", String().Enum("card", "bank").Build()).
		Input("currency", String().Pattern("^[A-Z]{3}$").Build()).
		Output(Object().
			Property("id", String().Build()).
			Property("status", String().Build()).
			Required("id", "status").
			Build()).
		Error(Object().
			Property("code", String().Build()).
			Property("message", String().Build()).
			Build()).
		Description("Process a payment with validation").
		Build()

	// Verify it's a proper schema
	if paymentSchema.Type() != TypeFunction {
		t.Errorf("Expected TypeFunction, got %s", paymentSchema.Type())
	}

	// Test JSON Schema generation
	jsonSchema := paymentSchema.ToJSONSchema()
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected object type in JSON schema")
	}

	// Test example generation
	example := paymentSchema.GenerateExample()
	if example == nil {
		t.Errorf("Expected non-nil example")
	}
}

func TestFunctionSchemaInObjects(t *testing.T) {
	// Test using function schemas as object properties
	paymentFunctionSchema := NewFunctionSchema().
		Input("amount", Number().Min(0).Build()).
		Input("method", String().Build()).
		Output(Object().Property("receipt", String().Build()).Build()).
		Build()

	// Objects with method schemas = records with function-valued fields
	bankAccountSchema := Object().
		Property("balance", Number().Build()).
		Property("deposit", paymentFunctionSchema).
		Property("withdraw", paymentFunctionSchema).
		Build()

	// Verify it works
	if bankAccountSchema.Type() != TypeObject {
		t.Errorf("Expected TypeObject, got %s", bankAccountSchema.Type())
	}

	// Generate example should work
	example := bankAccountSchema.GenerateExample()
	if example == nil {
		t.Errorf("Expected non-nil example")
	}
}
