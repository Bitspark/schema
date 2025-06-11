package portal

import (
	"context"
	"testing"

	"defs.dev/schema/api"
)

func TestAddressSystem(t *testing.T) {
	t.Run("Basic address parsing", func(t *testing.T) {
		// Test local address
		addr, err := NewAddress("local://add")
		if err != nil {
			t.Fatalf("Failed to parse local address: %v", err)
		}

		if addr.Scheme() != "local" {
			t.Errorf("Expected scheme 'local', got %s", addr.Scheme())
		}

		if addr.Path() != "/add" {
			t.Errorf("Expected path '/add', got %s", addr.Path())
		}

		if !addr.IsLocal() {
			t.Error("Expected address to be local")
		}

		if addr.IsNetwork() {
			t.Error("Expected address to not be network")
		}

		// Test HTTP address
		httpAddr, err := NewAddress("http://localhost:8080/api/add")
		if err != nil {
			t.Fatalf("Failed to parse HTTP address: %v", err)
		}

		if httpAddr.Scheme() != "http" {
			t.Errorf("Expected scheme 'http', got %s", httpAddr.Scheme())
		}

		if httpAddr.Authority() != "localhost:8080" {
			t.Errorf("Expected authority 'localhost:8080', got %s", httpAddr.Authority())
		}

		if httpAddr.IsLocal() {
			t.Error("Expected HTTP address to not be local")
		}

		if !httpAddr.IsNetwork() {
			t.Error("Expected HTTP address to be network")
		}
	})

	t.Run("Address builder", func(t *testing.T) {
		addr := NewAddressBuilder().
			Scheme("https").
			Host("api.example.com").
			Path("/api/v1/add").
			Query("version", "1").
			Build()

		if addr.Scheme() != "https" {
			t.Errorf("Expected scheme https, got %s", addr.Scheme())
		}

		if addr.Authority() != "api.example.com" {
			t.Errorf("Expected authority api.example.com, got %s", addr.Authority())
		}

		query := addr.Query()
		if query["version"] != "1" {
			t.Errorf("Expected version=1, got %s", query["version"])
		}
	})
}

func TestLocalPortal(t *testing.T) {
	portal := NewLocalPortal()
	ctx := context.Background()

	// Create a simple function handler
	addHandler := func(ctx context.Context, params api.FunctionInput) (api.FunctionOutput, error) {
		a, _ := params.Get("a")
		b, _ := params.Get("b")
		result := a.(float64) + b.(float64)
		return NewFunctionValue(result), nil
	}

	t.Run("Function registration and execution", func(t *testing.T) {
		// Apply the function to the portal (using nil schema for simplicity)
		addr, err := portal.Apply(ctx, "add", nil, addHandler)
		if err != nil {
			t.Fatalf("Failed to apply function: %v", err)
		}

		if !addr.IsLocal() {
			t.Errorf("Expected local address, got %s", addr.String())
		}

		// Resolve the function
		function, err := portal.ResolveFunction(ctx, addr)
		if err != nil {
			t.Fatalf("Failed to resolve function: %v", err)
		}

		if function.Name() != "add" {
			t.Errorf("Expected function name 'add', got %s", function.Name())
		}

		// Test function call
		params := FunctionInputMap{
			"a": 10.0,
			"b": 5.0,
		}

		result, err := function.Call(ctx, params)
		if err != nil {
			t.Fatalf("Failed to call function: %v", err)
		}

		if result.ToAny() != 15.0 {
			t.Errorf("Expected result 15.0, got %v", result.ToAny())
		}
	})

	t.Run("Portal management", func(t *testing.T) {
		// Test schemes
		schemes := portal.Schemes()
		if len(schemes) != 1 || schemes[0] != "local" {
			t.Errorf("Expected schemes [local], got %v", schemes)
		}

		// Test health
		err := portal.Health(ctx)
		if err != nil {
			t.Errorf("Expected healthy portal, got error: %v", err)
		}

		// Test GetFunction by name
		function, exists := portal.GetFunction("add")
		if !exists {
			t.Error("Expected to find function 'add'")
		}

		if function.Name() != "add" {
			t.Errorf("Expected function name 'add', got %s", function.Name())
		}
	})
}

func TestTestingPortal(t *testing.T) {
	portal := NewTestingPortal()
	ctx := context.Background()

	mockHandler := func(ctx context.Context, params api.FunctionInput) (api.FunctionOutput, error) {
		input, _ := params.Get("input")
		return NewFunctionValue("mocked: " + input.(string)), nil
	}

	t.Run("Mock functionality", func(t *testing.T) {
		// Register a mock (using nil schema for simplicity)
		addr := portal.Mock("mock", nil, mockHandler)
		if addr.Scheme() != "mock" {
			t.Errorf("Expected mock scheme, got %s", addr.Scheme())
		}

		// Resolve and call the mock
		function, err := portal.ResolveFunction(ctx, addr)
		if err != nil {
			t.Fatalf("Failed to resolve mock function: %v", err)
		}

		params := FunctionInputMap{
			"input": "test",
		}

		result, err := function.Call(ctx, params)
		if err != nil {
			t.Fatalf("Failed to call mock function: %v", err)
		}

		if result.ToAny() != "mocked: test" {
			t.Errorf("Expected 'mocked: test', got %v", result.ToAny())
		}

		// Check call history
		history := portal.CallHistory()
		if len(history) != 1 {
			t.Errorf("Expected 1 call in history, got %d", len(history))
		}

		call := history[0]
		if call.FunctionName != "mock" {
			t.Errorf("Expected function name 'mock', got %s", call.FunctionName)
		}
	})
}

func TestPortalRegistry(t *testing.T) {
	registry := NewPortalRegistry()
	localPortal := NewLocalPortal()
	testingPortal := NewTestingPortal()

	t.Run("Portal registration", func(t *testing.T) {
		// Register local portal
		err := registry.RegisterPortal([]string{"local"}, localPortal)
		if err != nil {
			t.Fatalf("Failed to register local portal: %v", err)
		}

		// Register testing portal
		err = registry.RegisterPortal([]string{"test", "mock"}, testingPortal)
		if err != nil {
			t.Fatalf("Failed to register testing portal: %v", err)
		}

		// Test portal retrieval
		portals := registry.ListPortals()
		if len(portals) != 3 { // local, test, mock
			t.Errorf("Expected 3 portal registrations, got %d", len(portals))
		}
	})

	t.Run("Portal resolution", func(t *testing.T) {
		localAddr := LocalAddress("test")
		portal, err := registry.GetPortal(localAddr)
		if err != nil {
			t.Fatalf("Failed to get local portal: %v", err)
		}

		if portal != localPortal {
			t.Error("Expected to get the same local portal instance")
		}
	})
}

func TestDefaultPortalRegistry(t *testing.T) {
	registry := NewDefaultPortalRegistry()

	t.Run("Pre-registered portals", func(t *testing.T) {
		portals := registry.ListPortals()

		// Should have local, test, and mock portals
		expectedSchemes := []string{"local", "test", "mock"}
		for _, scheme := range expectedSchemes {
			if _, exists := portals[scheme]; !exists {
				t.Errorf("Expected %s portal to be pre-registered", scheme)
			}
		}
	})
}

func TestFunctionInputOutput(t *testing.T) {
	t.Run("FunctionInputMap", func(t *testing.T) {
		input := FunctionInputMap{
			"a": 10.0,
			"b": "test",
			"c": true,
		}

		// Test Get
		if val, exists := input.Get("a"); !exists || val != 10.0 {
			t.Errorf("Expected a=10.0, got %v", val)
		}

		// Test Has
		if !input.Has("b") {
			t.Error("Expected to have key 'b'")
		}

		if input.Has("nonexistent") {
			t.Error("Expected to not have key 'nonexistent'")
		}

		// Test Set
		input.Set("d", 42)
		if val, exists := input.Get("d"); !exists || val != 42 {
			t.Errorf("Expected d=42, got %v", val)
		}

		// Test Keys
		keys := input.Keys()
		if len(keys) != 4 {
			t.Errorf("Expected 4 keys, got %d", len(keys))
		}

		// Test ToMap
		m := input.ToMap()
		if len(m) != 4 {
			t.Errorf("Expected map with 4 entries, got %d", len(m))
		}
	})

	t.Run("FunctionOutput", func(t *testing.T) {
		output := NewFunctionValue("test result")

		if output.Value() != "test result" {
			t.Errorf("Expected 'test result', got %v", output.Value())
		}

		if output.ToAny() != "test result" {
			t.Errorf("Expected 'test result', got %v", output.ToAny())
		}
	})
}
