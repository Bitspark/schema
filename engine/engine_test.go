package engine

import (
	builders2 "defs.dev/schema/builders"
	"testing"
)

func TestSchemaEngine_BasicFunctionality(t *testing.T) {
	// Create engine with validation disabled for testing
	config := DefaultEngineConfig()
	config.ValidateOnRegister = false
	engine := NewSchemaEngineWithConfig(config)

	// Test initial state
	if len(engine.ListSchemas()) != 0 {
		t.Errorf("Expected empty schema list, got %d schemas", len(engine.ListSchemas()))
	}

	// Create a simple schema
	userSchema := builders2.NewObjectSchema().
		Name("User").
		Property("id", builders2.NewIntegerSchema().Build()).
		Property("name", builders2.NewStringSchema().Build()).
		Property("email", builders2.NewStringSchema().Build()).
		Required("id", "name", "email").
		Build()

	// Register schema
	err := engine.RegisterSchema("User", userSchema)
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Verify registration
	if !engine.HasSchema("User") {
		t.Error("Schema should be registered")
	}

	schemas := engine.ListSchemas()
	if len(schemas) != 1 {
		t.Errorf("Expected 1 schema, got %d", len(schemas))
	}

	if schemas[0] != "User" {
		t.Errorf("Expected schema name 'User', got %s", schemas[0])
	}

	// Resolve schema
	resolved, err := engine.ResolveSchema("User")
	if err != nil {
		t.Fatalf("Failed to resolve schema: %v", err)
	}

	if resolved == nil {
		t.Error("Resolved schema should not be nil")
	}

	// Verify resolved schema properties
	if resolved.Metadata().Name != "User" {
		t.Errorf("Expected schema name 'User', got %s", resolved.Metadata().Name)
	}
}

func TestSchemaEngine_References(t *testing.T) {
	// Test reference creation and parsing

	// Simple reference
	ref := Ref("User")
	if ref.Name() != "User" {
		t.Errorf("Expected name 'User', got %s", ref.Name())
	}
	if ref.Namespace() != "" {
		t.Errorf("Expected empty namespace, got %s", ref.Namespace())
	}
	if ref.FullName() != "User" {
		t.Errorf("Expected full name 'User', got %s", ref.FullName())
	}

	// Namespaced reference
	nsRef := RefNS("auth", "User")
	if nsRef.Name() != "User" {
		t.Errorf("Expected name 'User', got %s", nsRef.Name())
	}
	if nsRef.Namespace() != "auth" {
		t.Errorf("Expected namespace 'auth', got %s", nsRef.Namespace())
	}
	if nsRef.FullName() != "auth:User" {
		t.Errorf("Expected full name 'auth:User', got %s", nsRef.FullName())
	}

	// Versioned reference
	verRef := RefVer("auth", "User", "v1.0")
	if verRef.Version() != "v1.0" {
		t.Errorf("Expected version 'v1.0', got %s", verRef.Version())
	}
	if verRef.FullName() != "auth:User@v1.0" {
		t.Errorf("Expected full name 'auth:User@v1.0', got %s", verRef.FullName())
	}
}

func TestSchemaEngine_ParseReference(t *testing.T) {
	tests := []struct {
		input     string
		name      string
		namespace string
		version   string
		shouldErr bool
	}{
		{"User", "User", "", "", false},
		{"auth:User", "User", "auth", "", false},
		{"User@v1.0", "User", "", "v1.0", false},
		{"auth:User@v1.0", "User", "auth", "v1.0", false},
		{"", "", "", "", true},
		{":", "", "", "", true},
		{"@", "", "", "", true},
		{"auth:", "", "", "", true},
		{":User", "User", "", "", true}, // Invalid: empty namespace
	}

	for _, test := range tests {
		ref, err := ParseReference(test.input)

		if test.shouldErr {
			if err == nil {
				t.Errorf("Expected error for input %s", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for input %s: %v", test.input, err)
			continue
		}

		if ref.Name() != test.name {
			t.Errorf("Input %s: expected name %s, got %s", test.input, test.name, ref.Name())
		}

		if ref.Namespace() != test.namespace {
			t.Errorf("Input %s: expected namespace %s, got %s", test.input, test.namespace, ref.Namespace())
		}

		if ref.Version() != test.version {
			t.Errorf("Input %s: expected version %s, got %s", test.input, test.version, ref.Version())
		}
	}
}

func TestSchemaEngine_Configuration(t *testing.T) {
	// Test configuration
	config := DefaultEngineConfig()

	if !config.EnableCache {
		t.Error("Expected cache to be enabled by default")
	}

	if config.MaxCacheSize <= 0 {
		t.Error("Expected positive cache size")
	}

	if config.CircularDepthLimit <= 0 {
		t.Error("Expected positive circular depth limit")
	}

	// Create engine with custom config
	customConfig := EngineConfig{
		EnableCache:        false,
		MaxCacheSize:       100,
		CircularDepthLimit: 10,
		StrictMode:         true,
		ValidateOnRegister: false,
		EnableConcurrency:  false,
	}

	engine := NewSchemaEngineWithConfig(customConfig)
	actualConfig := engine.Config()

	if actualConfig.EnableCache != false {
		t.Error("Expected cache to be disabled")
	}

	if actualConfig.StrictMode != true {
		t.Error("Expected strict mode to be enabled")
	}
}

func TestSchemaEngine_ErrorHandling(t *testing.T) {
	engine := NewSchemaEngine()

	// Test registering with empty name
	userSchema := builders2.NewStringSchema().Build()
	err := engine.RegisterSchema("", userSchema)
	if err == nil {
		t.Error("Expected error for empty schema name")
	}

	// Test registering nil schema
	err = engine.RegisterSchema("Test", nil)
	if err == nil {
		t.Error("Expected error for nil schema")
	}

	// Test resolving non-existent schema
	_, err = engine.ResolveSchema("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent schema")
	}

	// Test resolving with empty name
	_, err = engine.ResolveSchema("")
	if err == nil {
		t.Error("Expected error for empty schema name")
	}
}

func TestSchemaEngine_Clone(t *testing.T) {
	config := DefaultEngineConfig()
	config.ValidateOnRegister = false
	engine := NewSchemaEngineWithConfig(config)

	// Register a schema
	userSchema := builders2.NewStringSchema().Build()
	err := engine.RegisterSchema("User", userSchema)
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Clone the engine
	clone := engine.Clone()

	// Verify clone has the schema
	if !clone.HasSchema("User") {
		t.Error("Clone should have the registered schema")
	}

	// Verify independence - register new schema in original
	err = engine.RegisterSchema("Order", builders2.NewStringSchema().Build())
	if err != nil {
		t.Fatalf("Failed to register schema in original: %v", err)
	}

	// Clone should not have the new schema
	if clone.HasSchema("Order") {
		t.Error("Clone should not have schema registered after cloning")
	}
}

func TestSchemaEngine_Reset(t *testing.T) {
	config := DefaultEngineConfig()
	config.ValidateOnRegister = false
	engine := NewSchemaEngineWithConfig(config)

	// Register some schemas
	err := engine.RegisterSchema("User", builders2.NewStringSchema().Build())
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	err = engine.RegisterSchema("Order", builders2.NewStringSchema().Build())
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Verify schemas are registered
	if len(engine.ListSchemas()) != 2 {
		t.Errorf("Expected 2 schemas before reset, got %d", len(engine.ListSchemas()))
	}

	// Reset the engine
	err = engine.Reset()
	if err != nil {
		t.Fatalf("Failed to reset engine: %v", err)
	}

	// Verify schemas are cleared
	if len(engine.ListSchemas()) != 0 {
		t.Errorf("Expected 0 schemas after reset, got %d", len(engine.ListSchemas()))
	}
}
