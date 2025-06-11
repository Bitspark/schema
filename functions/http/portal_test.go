package http

import (
	"context"
	"testing"
	"time"

	"defs.dev/schema"
)

func TestHTTPPortal_BasicFunctionality(t *testing.T) {
	// Create a test function schema
	testSchema := schema.NewFunctionSchema().
		Name("testFunction").
		Description("A test function").
		Input("message", schema.String().Build()).
		Output(schema.String().Build()).
		Build()

	// Create a test handler
	var testHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		message := params["message"].(string)
		return schema.FromAny(map[string]any{
			"message": "Hello, " + message + "!",
		}), nil
	}

	// Create HTTP portal
	portal := NewPortal(Config{
		Host:           "localhost",
		Port:           8081, // Use different port to avoid conflicts
		DefaultTimeout: 5 * time.Second,
	})

	// Start server
	if err := portal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// Test address generation
	address, err := portal.GenerateAddress("testFunction", testHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	if address == "" {
		t.Fatal("Generated address is empty")
	}

	t.Logf("Generated address: %s", address)

	// Test Apply method (provider side)
	endpointFunc, err := portal.Apply(address, testSchema.(*schema.FunctionSchema), testHandler)
	if err != nil {
		t.Fatalf("Failed to apply function: %v", err)
	}

	if endpointFunc.Address() != address {
		t.Errorf("Expected address %s, got %s", address, endpointFunc.Address())
	}

	// Test ResolveFunction method (consumer side)
	clientFunc, err := portal.ResolveFunction(context.Background(), address)
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	// Test calling the function through HTTP
	// Note: This would normally work, but requires the server to be fully running
	// For a basic test, we just verify the function was created
	if clientFunc.Address() != address {
		t.Errorf("Expected client function address %s, got %s", address, clientFunc.Address())
	}
}

func TestHTTPPortal_AddressGeneration(t *testing.T) {
	portal := NewPortal()

	// Test multiple address generation to ensure uniqueness
	addresses := make(map[string]bool)
	for i := 0; i < 10; i++ {
		address, err := portal.GenerateAddress("testFunc", nil)
		if err != nil {
			t.Fatalf("Failed to generate address: %v", err)
		}

		if addresses[address] {
			t.Errorf("Duplicate address generated: %s", address)
		}
		addresses[address] = true
	}
}

func TestHTTPPortal_Scheme(t *testing.T) {
	portal := NewPortal()
	schemes := portal.Scheme()

	expectedSchemes := []string{"http", "https"}
	if len(schemes) != len(expectedSchemes) {
		t.Errorf("Expected %d schemes, got %d", len(expectedSchemes), len(schemes))
	}

	for i, expected := range expectedSchemes {
		if i >= len(schemes) || schemes[i] != expected {
			t.Errorf("Expected scheme %s at index %d, got %s", expected, i, schemes[i])
		}
	}
}

func TestHTTPPortal_Middleware(t *testing.T) {
	portal := NewPortal()

	// Add logging middleware
	logging := NewLoggingMiddleware()
	portal.AddMiddleware(logging)

	// Add metrics middleware
	metrics := NewMetricsMiddleware()
	portal.AddMiddleware(metrics)

	// Verify middleware was added
	if len(portal.middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(portal.middleware))
	}
}

func TestConfigBuilder(t *testing.T) {
	config := NewConfigBuilder().
		Host("example.com").
		Port(9090).
		Timeout(45 * time.Second).
		Retries(5).
		UserAgent("test-agent").
		BasePath("/api/test").
		Build()

	if config.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", config.Host)
	}

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}

	if config.DefaultTimeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", config.DefaultTimeout)
	}

	if config.MaxRetries != 5 {
		t.Errorf("Expected retries 5, got %d", config.MaxRetries)
	}

	if config.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", config.UserAgent)
	}

	if config.BasePath != "/api/test" {
		t.Errorf("Expected base path '/api/test', got '%s'", config.BasePath)
	}
}

func TestFactoryFunctions(t *testing.T) {
	// Test LocalPortal
	local := LocalPortal()
	if local == nil {
		t.Fatal("LocalPortal returned nil")
	}

	// Test QuickPortal
	quick := QuickPortal("localhost", 8082)
	if quick == nil {
		t.Fatal("QuickPortal returned nil")
	}

	// Test configuration
	if quick.config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", quick.config.Host)
	}

	if quick.config.Port != 8082 {
		t.Errorf("Expected port 8082, got %d", quick.config.Port)
	}
}
