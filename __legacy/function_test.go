package schema

import (
	"testing"
)

func TestFunctionSchema(t *testing.T) {
	// Test creating a function schema (pure interface description)
	paymentSchema := NewFunctionSchema().
		Input("amount", NewNumber().Min(0).Build()).
		Input("method", NewString().Enum("card", "bank").Build()).
		Input("currency", NewString().Pattern("^[A-Z]{3}$").Build()).
		Output(NewObject().
			Property("id", NewString().Build()).
			Property("status", NewString().Build()).
			Required("id", "status").
			Build()).
		Error(NewObject().
			Property("code", NewString().Build()).
			Property("message", NewString().Build()).
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
		Input("amount", NewNumber().Min(0).Build()).
		Input("method", NewString().Build()).
		Output(NewObject().Property("receipt", NewString().Build()).Build()).
		Build()

	// Objects with method schemas = records with function-valued fields
	bankAccountSchema := NewObject().
		Property("balance", NewNumber().Build()).
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
