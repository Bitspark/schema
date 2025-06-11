package javascript

import (
	"context"
	"testing"
	"time"

	"defs.dev/schema"
	"defs.dev/schema/functions"
)

func TestJavaScriptPortal_Basic(t *testing.T) {
	// Create portal with default config
	portal := NewPortalWithDefaults()

	// Verify scheme
	if portal.Scheme() != "js" {
		t.Errorf("Expected scheme 'js', got '%s'", portal.Scheme())
	}

	// Test address generation
	jsFunc := JSFunction{
		Code:         `function add(params) { return params.a + params.b; }`,
		FunctionName: "add",
	}

	address := portal.GenerateAddress("testAdd", jsFunc)
	if address == "" {
		t.Error("Expected non-empty address")
	}

	// Address should follow js://engine/name/id format
	expected := "js://goja/testAdd/"
	if !contains(address, expected) {
		t.Errorf("Address should start with '%s', got '%s'", expected, address)
	}
}

func TestJavaScriptFunction_Execution(t *testing.T) {
	// Create a complete JavaScript system
	_, registry, consumer := NewDefaultJavaScriptSystem()

	// Define output schema
	outputSchema := schema.NewNumber().Build()

	functionSchema := schema.NewFunctionSchema().
		Input("a", schema.NewNumber().Build()).
		Input("b", schema.NewNumber().Build()).
		Required("a", "b").
		Output(outputSchema).
		Build().(*schema.FunctionSchema)

	// Create JavaScript function
	jsFunc := JSFunction{
		Code: `
			function add(params) {
				if (typeof params.a !== 'number' || typeof params.b !== 'number') {
					throw new Error('Both a and b must be numbers');
				}
				return params.a + params.b;
			}
		`,
		FunctionName: "add",
	}

	// Register function
	address, err := registry.Register("testAdd", functionSchema, jsFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Test calling via consumer
	ctx := context.Background()
	params := map[string]any{
		"a": 5.0,
		"b": 3.0,
	}

	result, err := consumer.CallAt(ctx, address, params)
	if err != nil {
		t.Fatalf("Failed to call function: %v", err)
	}

	// Verify result (JavaScript returns integers as integers, not floats)
	expected := 8.0
	if resultFloat, ok := result.ToAny().(float64); ok {
		if resultFloat != expected {
			t.Errorf("Expected result %.1f, got %.1f", expected, resultFloat)
		}
	} else if resultInt, ok := result.ToAny().(int64); ok {
		if float64(resultInt) != expected {
			t.Errorf("Expected result %.1f, got %d", expected, resultInt)
		}
	} else {
		t.Errorf("Expected numeric result, got %T: %v", result, result)
	}
}

func TestJavaScriptFunction_ValidationError(t *testing.T) {
	// Create registry
	registry := NewRegistry()

	// Define schema requiring specific parameters
	functionSchema := schema.NewFunctionSchema().
		Input("username", schema.NewString().MinLength(3).Build()).
		Input("email", schema.NewString().Email().Build()).
		Required("username", "email").
		Build().(*schema.FunctionSchema)

	// Create and register function
	_, err := registry.Register("validateUser", functionSchema, JSFunction{
		Code: `
			function validateUser(params) {
				return {
					valid: params.username.length >= 3 && params.email.includes('@'),
					username: params.username,
					email: params.email
				};
			}
		`,
		FunctionName: "validateUser",
	})
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Get function for direct calling
	fn, err := registry.GetFunction("validateUser")
	if err != nil {
		t.Fatalf("Failed to get function: %v", err)
	}

	// Test with invalid parameters (should fail validation)
	ctx := context.Background()
	invalidParams := map[string]any{
		"username": "ab",            // Too short
		"email":    "invalid-email", // Not an email
	}

	_, err = fn.Call(ctx, invalidParams)
	if err == nil {
		t.Error("Expected validation error but got none")
	}

	// Check error type
	if jsErr, ok := err.(*JSFunctionError); ok {
		if jsErr.Stage != "parameter_validation" {
			t.Errorf("Expected parameter_validation error, got %s", jsErr.Stage)
		}
	} else {
		t.Errorf("Expected JSFunctionError, got %T", err)
	}
}

func TestJavaScriptFunction_Timeout(t *testing.T) {
	// Create portal with short timeout
	config := DefaultConfig()
	config.DefaultTimeout = 100 * time.Millisecond
	portal := NewPortal(config)
	registry := functions.NewRegistry(portal)

	// Create function schema
	functionSchema := schema.NewFunctionSchema().
		Input("delay", schema.NewNumber().Build()).
		Required("delay").
		Output(schema.NewString().Build()).
		Build().(*schema.FunctionSchema)

	// Create JavaScript function that takes time
	jsFunc := JSFunction{
		Code: `
			function slowFunction(params) {
				// Simulate slow operation
				var start = Date.now();
				while (Date.now() - start < params.delay) {
					// Busy wait
				}
				return "completed";
			}
		`,
		FunctionName: "slowFunction",
	}

	// Register function
	_, err := registry.Register("slowFunction", functionSchema, jsFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Get function
	fn, err := registry.GetFunction("slowFunction")
	if err != nil {
		t.Fatalf("Failed to get function: %v", err)
	}

	// Call with delay longer than timeout
	ctx := context.Background()
	params := map[string]any{
		"delay": 500.0, // 500ms, longer than 100ms timeout
	}

	_, err = fn.Call(ctx, params)
	if err == nil {
		t.Error("Expected timeout error but got none")
	}

	// Check error type
	if jsErr, ok := err.(*JSFunctionError); ok {
		if jsErr.Stage != "timeout" {
			t.Errorf("Expected timeout error, got %s", jsErr.Stage)
		}
	} else {
		t.Errorf("Expected JSFunctionError, got %T", err)
	}
}

func TestJavaScriptFunction_SyntaxError(t *testing.T) {
	registry := NewRegistry()

	// Create function schema
	functionSchema := schema.NewFunctionSchema().
		Input("test", schema.NewString().Build()).
		Required("test").
		Output(schema.NewString().Build()).
		Build().(*schema.FunctionSchema)

	// Create JavaScript function with syntax error
	jsFunc := JSFunction{
		Code: `
			function badSyntax(params) {
				return params.test +; // Syntax error
			}
		`,
		FunctionName: "badSyntax",
	}

	// Register function (this might succeed)
	_, err := registry.Register("badSyntax", functionSchema, jsFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Get function
	fn, err := registry.GetFunction("badSyntax")
	if err != nil {
		t.Fatalf("Failed to get function: %v", err)
	}

	// Call function (this should fail with syntax error)
	ctx := context.Background()
	params := map[string]any{
		"test": "hello",
	}

	_, err = fn.Call(ctx, params)
	if err == nil {
		t.Error("Expected syntax error but got none")
	}

	// Check error type
	if jsErr, ok := err.(*JSFunctionError); ok {
		if jsErr.Stage != "vm_creation" && jsErr.Stage != "syntax_error" {
			t.Errorf("Expected syntax or vm_creation error, got %s", jsErr.Stage)
		}
	} else {
		t.Errorf("Expected JSFunctionError, got %T", err)
	}
}

func TestPortalWrapper(t *testing.T) {
	portal := NewPortalWithDefaults()
	wrapper := &PortalWrapper{portal}

	// Test type assertion in wrapper
	jsFunc := JSFunction{
		Code:         `function test(params) { return "ok"; }`,
		FunctionName: "test",
	}

	// Should work with correct type
	address := wrapper.GenerateAddress("test", jsFunc)
	if address == "" {
		t.Error("Expected non-empty address from wrapper")
	}

	// Test with wrong type (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with wrong type, but didn't get one")
		}
	}()

	wrapper.GenerateAddress("test", "wrong type")
}

func TestConsumerIntegration(t *testing.T) {
	// Create consumer with JavaScript portal
	consumer := NewConsumer()

	// Verify portal is registered
	portals := consumer.Portals()
	found := false
	for _, scheme := range portals {
		if scheme == "js" {
			found = true
			break
		}
	}

	if !found {
		t.Error("JavaScript portal not registered with consumer")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
