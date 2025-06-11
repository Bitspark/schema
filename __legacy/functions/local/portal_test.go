package local

import (
	"context"
	"testing"

	"defs.dev/schema"
	"defs.dev/schema/functions"
)

func TestPortalSystem(t *testing.T) {
	// Create a function schema
	userSchema := schema.NewFunctionSchema().
		Name("getUser").
		Description("Retrieve user by ID").
		Input("id", schema.NewString().Build()).
		Output(schema.NewObject().
			Property("id", schema.NewString().Build()).
			Property("name", schema.NewString().Build()).
			Required("id", "name").
			Build()).
		Build().(*schema.FunctionSchema)

	// Create a handler implementation
	var getUserHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		userID := params["id"].(string)

		return schema.FromAny(map[string]any{
			"id":   userID,
			"name": "Test User " + userID,
		}), nil
	}

	// Test registry registration
	registry := NewRegistry()

	address, err := registry.Register("getUser", userSchema, getUserHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	if address == "" {
		t.Fatal("Expected non-empty address")
	}

	// Address should have local:// scheme
	if address[:8] != "local://" {
		t.Errorf("Expected local:// scheme, got: %s", address)
	}

	// Test registry retrieval by name
	retrievedFunc, err := registry.GetFunction("getUser")
	if err != nil {
		t.Fatalf("Failed to get function: %v", err)
	}

	// Test function call through registry
	result, err := retrievedFunc.Call(context.Background(), map[string]any{
		"id": "123",
	})
	if err != nil {
		t.Fatalf("Failed to call function: %v", err)
	}

	// Verify result
	resultMap := result.Value().(map[string]any)
	if resultMap["id"] != "123" {
		t.Errorf("Expected id=123, got: %v", resultMap["id"])
	}
	if resultMap["name"] != "Test User 123" {
		t.Errorf("Expected name='Test User 123', got: %v", resultMap["name"])
	}
}

func TestConsumerSystem(t *testing.T) {
	// Create schema and handler
	addSchema := schema.NewFunctionSchema().
		Name("add").
		Description("Add two numbers").
		Input("a", schema.NewNumber().Build()).
		Input("b", schema.NewNumber().Build()).
		Output(schema.NewObject().
			Property("result", schema.NewNumber().Build()).
			Required("result").
			Build()).
		Build().(*schema.FunctionSchema)

	var addHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		a := params["a"].(float64)
		b := params["b"].(float64)

		return schema.FromAny(map[string]any{
			"result": a + b,
		}), nil
	}

	// Create registry and register function
	registry := NewRegistry()
	address, err := registry.Register("add", addSchema, addHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Create consumer and register the same portal instance
	consumer := functions.NewConsumer()
	consumer.RegisterPortal(registry.Portal())

	// Test calling by address
	result, err := consumer.CallAt(context.Background(), address, map[string]any{
		"a": 5.0,
		"b": 3.0,
	})
	if err != nil {
		t.Fatalf("Failed to call function by address: %v", err)
	}

	// Verify result
	resultMap := result.ToAny().(map[string]any)
	if resultMap["result"] != 8.0 {
		t.Errorf("Expected result=8.0, got: %v", resultMap["result"])
	}
}

func TestRegistryOperations(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if len(registry.Names()) != 0 {
		t.Error("Expected empty registry")
	}

	if registry.Exists("nonexistent") {
		t.Error("Expected function to not exist")
	}

	// Create test function
	testSchema := schema.NewFunctionSchema().
		Name("test").
		Description("Test function").
		Output(schema.NewString().Build()).
		Build().(*schema.FunctionSchema)

	var testHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return schema.FromAny(map[string]any{
			"result": "test result",
		}), nil
	}

	// Register function
	address, err := registry.Register("test", testSchema, testHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Test registry operations
	if !registry.Exists("test") {
		t.Error("Expected function to exist")
	}

	names := registry.Names()
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("Expected names=[test], got: %v", names)
	}

	retrievedAddr, err := registry.GetAddress("test")
	if err != nil {
		t.Fatalf("Failed to get address: %v", err)
	}
	if retrievedAddr != address {
		t.Errorf("Expected address=%s, got: %s", address, retrievedAddr)
	}

	// Test duplicate registration
	_, err = registry.Register("test", testSchema, testHandler)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test anonymous registration
	anonName, anonAddr, err := registry.RegisterAnon(testSchema, testHandler)
	if err != nil {
		t.Fatalf("Failed to register anonymous function: %v", err)
	}

	if anonName == "" || anonAddr == "" {
		t.Error("Expected non-empty anonymous name and address")
	}

	// Test removal
	err = registry.Remove("test")
	if err != nil {
		t.Fatalf("Failed to remove function: %v", err)
	}

	if registry.Exists("test") {
		t.Error("Expected function to be removed")
	}
}

func TestErrorHandling(t *testing.T) {
	registry := NewRegistry()
	consumer := NewConsumer()
	consumer.RegisterPortal(registry.Portal())

	// Test getting non-existent function
	_, err := registry.GetFunction("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent function")
	}

	// Test getting non-existent address
	_, err = registry.GetAddress("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent address")
	}

	// Test calling invalid address
	_, err = consumer.CallAt(context.Background(), "invalid-address", nil)
	if err == nil {
		t.Error("Expected error for invalid address")
	}

	// Test calling with unregistered scheme
	_, err = consumer.CallAt(context.Background(), "unknown://test", nil)
	if err == nil {
		t.Error("Expected error for unknown scheme")
	}
}
