package local

import (
	"context"
	"testing"

	"defs.dev/schema"
)

func TestLocalFunction(t *testing.T) {
	// Test creating a local function (with actual implementation)
	userSchema := schema.Object().
		Property("id", schema.String().Build()).
		Property("name", schema.String().Build()).
		Required("id", "name").
		Build()

	getUserFunction := NewLocalFunction("getUser").
		Description("Retrieve user by ID").
		Parameters(schema.Object().
			Property("id", schema.String().Build()).
			Required("id").
			Build()).
		Returns(userSchema).
		Handler(func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
			// Mock implementation
			return schema.FromAny(map[string]any{
				"id":   "123",
				"name": "Test User",
			}), nil
		}).
		Build()

	// Test execution
	result, err := getUserFunction.Call(context.Background(), schema.NewFunctionInput(map[string]any{
		"id": "123",
	}))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Value() == nil {
		t.Errorf("Expected non-nil result")
	}

	// Test validation failure
	_, err = getUserFunction.Call(context.Background(), schema.NewFunctionInput(map[string]any{
		"invalidParam": "value",
	}))

	if err == nil {
		t.Errorf("Expected validation error")
	}
}
