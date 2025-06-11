package registry

import (
	"context"
	"defs.dev/schema/schemas"
	"testing"

	"defs.dev/schema/api"
)

// TestFunctionRegistry tests the basic functionality of the function registry
func TestFunctionRegistry(t *testing.T) {
	registry := NewFunctionRegistry()

	// Test initial state
	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d functions", registry.Count())
	}

	// Create a simple function
	mockFunction := &MockFunction{
		name:   "testFunc",
		schema: createMockFunctionSchema(),
	}

	// Test registration
	err := registry.Register("testFunc", mockFunction)
	if err != nil {
		t.Errorf("Failed to register function: %v", err)
	}

	// Test retrieval
	fn, exists := registry.Get("testFunc")
	if !exists {
		t.Error("Function not found after registration")
	}
	if fn.Name() != "testFunc" {
		t.Errorf("Expected function name 'testFunc', got '%s'", fn.Name())
	}

	// Test listing
	names := registry.List()
	if len(names) != 1 || names[0] != "testFunc" {
		t.Errorf("Expected [testFunc], got %v", names)
	}

	// Test existence check
	if !registry.Exists("testFunc") {
		t.Error("Function should exist")
	}
	if registry.Exists("nonexistent") {
		t.Error("Non-existent function should not exist")
	}

	// Test count
	if registry.Count() != 1 {
		t.Errorf("Expected 1 function, got %d", registry.Count())
	}

	// Test unregistration
	err = registry.Unregister("testFunc")
	if err != nil {
		t.Errorf("Failed to unregister function: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry after unregistration, got %d functions", registry.Count())
	}
}

// TestFunctionRegistryValidation tests function validation
func TestFunctionRegistryValidation(t *testing.T) {
	registry := NewFunctionRegistry()

	// Test validation for non-existent function
	result := registry.Validate("nonexistent", map[string]any{"input": "value"})
	if result.Valid {
		t.Error("Expected validation to fail for non-existent function")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected validation errors for non-existent function")
	}

	// Register a function
	mockFunction := &MockFunction{
		name:   "testFunc",
		schema: createMockFunctionSchema(),
	}
	registry.Register("testFunc", mockFunction)

	// Test validation for existing function
	result = registry.Validate("testFunc", map[string]any{"input": "test"})
	// The result depends on the mock function schema implementation
	if len(result.Errors) > 0 {
		t.Logf("Validation errors (expected for mock): %v", result.Errors)
	}
}

// TestFunctionRegistryExecution tests function execution
func TestFunctionRegistryExecution(t *testing.T) {
	registry := NewFunctionRegistry()

	// Create and register a function
	mockFunction := &MockFunction{
		name:   "testFunc",
		schema: createMockFunctionSchema(),
	}
	registry.Register("testFunc", mockFunction)

	// Test function call
	ctx := context.Background()
	params := FunctionInputMap{"input": "test"}

	output, err := registry.Call(ctx, "testFunc", params)
	if err != nil {
		t.Errorf("Function call failed: %v", err)
	}
	if output == nil {
		t.Error("Expected output from function call")
	}

	// Test call to non-existent function
	_, err = registry.Call(ctx, "nonexistent", params)
	if err == nil {
		t.Error("Expected error when calling non-existent function")
	}
}

// TestFunctionRegistryMetadata tests metadata functionality
func TestFunctionRegistryMetadata(t *testing.T) {
	registry := NewFunctionRegistry()

	// Register a function
	mockFunction := &MockFunction{
		name:   "testFunc",
		schema: createMockFunctionSchema(),
	}
	registry.Register("testFunc", mockFunction)

	// Test getting metadata
	metadata, exists := registry.GetMetadata("testFunc")
	if !exists {
		t.Error("Expected metadata to exist for registered function")
	}
	if metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", metadata.Version)
	}

	// Test setting metadata
	newMetadata := FunctionMetadata{
		Version:     "2.0.0",
		Tags:        []string{"test", "updated"},
		Description: "Updated function",
	}
	err := registry.SetMetadata("testFunc", newMetadata)
	if err != nil {
		t.Errorf("Failed to set metadata: %v", err)
	}

	// Verify updated metadata
	updated, _ := registry.GetMetadata("testFunc")
	if updated.Version != "2.0.0" {
		t.Errorf("Expected updated version '2.0.0', got '%s'", updated.Version)
	}
	if len(updated.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(updated.Tags))
	}
}

// TestFunctionRegistryTypedFunctions tests typed function functionality
func TestFunctionRegistryTypedFunctions(t *testing.T) {
	registry := NewFunctionRegistry()

	// Create and register a typed function
	mockTypedFunction := &MockTypedFunction{
		MockFunction: MockFunction{
			name:   "typedFunc",
			schema: createMockFunctionSchema(),
		},
	}

	err := registry.RegisterTyped("typedFunc", mockTypedFunction)
	if err != nil {
		t.Errorf("Failed to register typed function: %v", err)
	}

	// Test retrieval as typed function
	typedFn, exists := registry.GetTyped("typedFunc")
	if !exists {
		t.Error("Typed function not found after registration")
	}
	if typedFn == nil {
		t.Error("Expected typed function, got nil")
	}

	// Test that it's also available as regular function
	fn, exists := registry.Get("typedFunc")
	if !exists {
		t.Error("Typed function should also be available as regular function")
	}
	if fn.Name() != "typedFunc" {
		t.Errorf("Expected function name 'typedFunc', got '%s'", fn.Name())
	}
}

// TestFunctionRegistryClone tests registry cloning
func TestFunctionRegistryClone(t *testing.T) {
	registry := NewFunctionRegistry()

	// Register a function
	mockFunction := &MockFunction{
		name:   "testFunc",
		schema: createMockFunctionSchema(),
	}
	registry.Register("testFunc", mockFunction)

	// Clone the registry
	clone := registry.Clone()

	// Test that clone has the same functions
	if clone.Count() != registry.Count() {
		t.Errorf("Expected clone to have %d functions, got %d", registry.Count(), clone.Count())
	}

	fn, exists := clone.Get("testFunc")
	if !exists {
		t.Error("Function not found in cloned registry")
	}
	if fn.Name() != "testFunc" {
		t.Errorf("Expected function name 'testFunc' in clone, got '%s'", fn.Name())
	}

	// Test that modifications to original don't affect clone
	mockFunction2 := &MockFunction{
		name:   "testFunc2",
		schema: createMockFunctionSchema(),
	}
	registry.Register("testFunc2", mockFunction2)

	if clone.Count() == registry.Count() {
		t.Error("Clone should not be affected by modifications to original")
	}
}

// Mock implementations for testing

type MockFunction struct {
	name   string
	schema api.FunctionSchema
}

func (f *MockFunction) Call(ctx context.Context, params api.FunctionInput) (api.FunctionOutput, error) {
	return &FunctionOutputValue{
		value: map[string]any{
			"function": f.name,
			"input":    params.ToMap(),
			"result":   "mock_result",
		},
	}, nil
}

func (f *MockFunction) Schema() api.FunctionSchema {
	return f.schema
}

func (f *MockFunction) Name() string {
	return f.name
}

type MockTypedFunction struct {
	MockFunction
}

func (f *MockTypedFunction) CallTyped(ctx context.Context, input any, output any) error {
	// Mock implementation - just log the call
	return nil
}

// Helper function to create a mock function schema
func createMockFunctionSchema() api.FunctionSchema {
	inputs := schemas.NewArgSchemas()
	outputs := schemas.NewArgSchemas()

	return schemas.NewFunctionSchema(inputs, outputs)
}
