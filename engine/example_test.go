package engine

import (
	builders2 "defs.dev/schema/builders"
	"fmt"
	"testing"
)

// TestSchemaEngine_ComprehensiveExample demonstrates the full Schema Engine functionality
func TestSchemaEngine_ComprehensiveExample(t *testing.T) {
	// Create engine with custom configuration
	config := EngineConfig{
		EnableCache:        true,
		MaxCacheSize:       100,
		CircularDepthLimit: 20,
		StrictMode:         false,
		ValidateOnRegister: false, // Disabled for this example
		EnableConcurrency:  true,
	}
	engine := NewSchemaEngineWithConfig(config)

	fmt.Println("=== Schema Engine Comprehensive Example ===")

	// 1. Schema Registration
	fmt.Println("\n1. Registering Schemas:")

	// User schema
	userSchema := builders2.NewObjectSchema().
		Name("User").
		Property("id", builders2.NewIntegerSchema().Build()).
		Property("name", builders2.NewStringSchema().Build()).
		Property("email", builders2.NewStringSchema().Build()).
		Property("active", builders2.NewBooleanSchema().Build()).
		Required("id", "name", "email").
		Build()

	err := engine.RegisterSchema("User", userSchema)
	if err != nil {
		t.Fatalf("Failed to register User schema: %v", err)
	}
	fmt.Printf("✓ Registered schema: User\n")

	// Order schema
	orderSchema := builders2.NewObjectSchema().
		Name("Order").
		Property("id", builders2.NewIntegerSchema().Build()).
		Property("userId", builders2.NewIntegerSchema().Build()).
		Property("total", builders2.NewNumberSchema().Build()).
		Property("items", builders2.NewArraySchema().
			Items(builders2.NewStringSchema().Build()).
			Build()).
		Required("id", "userId", "total").
		Build()

	err = engine.RegisterSchema("Order", orderSchema)
	if err != nil {
		t.Fatalf("Failed to register Order schema: %v", err)
	}
	fmt.Printf("✓ Registered schema: Order\n")

	// 2. Schema Resolution
	fmt.Println("\n2. Schema Resolution:")

	resolved, err := engine.ResolveSchema("User")
	if err != nil {
		t.Fatalf("Failed to resolve User schema: %v", err)
	}
	fmt.Printf("✓ Resolved schema: %s (type: %s)\n", resolved.Metadata().Name, resolved.Type())

	// 3. Reference System
	fmt.Println("\n3. Reference System:")

	// Simple reference
	ref := Ref("User")
	fmt.Printf("✓ Simple reference: %s\n", ref.FullName())

	// Namespaced reference
	nsRef := RefNS("auth", "User")
	fmt.Printf("✓ Namespaced reference: %s\n", nsRef.FullName())

	// Versioned reference
	verRef := RefVer("auth", "User", "v1.0")
	fmt.Printf("✓ Versioned reference: %s\n", verRef.FullName())

	// Parse reference
	parsed, err := ParseReference("billing:Order@v2.1")
	if err != nil {
		t.Fatalf("Failed to parse reference: %v", err)
	}
	fmt.Printf("✓ Parsed reference: %s (namespace: %s, name: %s, version: %s)\n",
		parsed.FullName(), parsed.Namespace(), parsed.Name(), parsed.Version())

	// 4. Built-in Annotations
	fmt.Println("\n4. Built-in Annotations:")

	annotations := engine.ListAnnotations()
	fmt.Printf("✓ Available annotations: %v\n", annotations)

	// Test annotation validation
	err = engine.ValidateAnnotation("pattern", "service")
	if err != nil {
		t.Fatalf("Failed to validate pattern annotation: %v", err)
	}
	fmt.Printf("✓ Validated annotation: pattern=service\n")

	// Test complex annotation
	deploymentConfig := map[string]any{
		"strategy": "rolling",
		"replicas": 3,
		"resources": map[string]any{
			"cpu":    "100m",
			"memory": "256Mi",
		},
	}

	err = engine.ValidateAnnotation("deployment", deploymentConfig)
	if err != nil {
		t.Fatalf("Failed to validate deployment annotation: %v", err)
	}
	fmt.Printf("✓ Validated complex annotation: deployment configuration\n")

	// 5. Engine Management
	fmt.Println("\n5. Engine Management:")

	// List all schemas
	schemas := engine.ListSchemas()
	fmt.Printf("✓ Total schemas registered: %d (%v)\n", len(schemas), schemas)

	// Clone engine
	clone := engine.Clone()
	fmt.Printf("✓ Cloned engine with %d schemas\n", len(clone.ListSchemas()))

	// Verify independence
	err = engine.RegisterSchema("Product", builders2.NewStringSchema().Build())
	if err != nil {
		t.Fatalf("Failed to register Product schema: %v", err)
	}

	if clone.HasSchema("Product") {
		t.Error("Clone should not have Product schema")
	}
	fmt.Printf("✓ Clone independence verified\n")

	// 6. Custom Schema Types (placeholder)
	fmt.Println("\n6. Schema Type System:")

	availableTypes := engine.GetAvailableTypes()
	fmt.Printf("✓ Available schema types: %d (placeholder system)\n", len(availableTypes))

	// 7. Error Handling
	fmt.Println("\n7. Error Handling:")

	// Test non-existent schema
	_, err = engine.ResolveSchema("NonExistent")
	if err != nil {
		fmt.Printf("✓ Proper error for non-existent schema: %v\n", err)
	}

	// Test invalid annotation
	err = engine.ValidateAnnotation("pattern", "invalid_pattern")
	if err != nil {
		fmt.Printf("✓ Proper error for invalid annotation: %v\n", err)
	}

	// 8. Configuration
	fmt.Println("\n8. Engine Configuration:")

	currentConfig := engine.Config()
	fmt.Printf("✓ Cache enabled: %v\n", currentConfig.EnableCache)
	fmt.Printf("✓ Max cache size: %d\n", currentConfig.MaxCacheSize)
	fmt.Printf("✓ Strict mode: %v\n", currentConfig.StrictMode)

	fmt.Println("\n=== Example Complete ===")
}

// TestSchemaEngine_AnnotationSystem demonstrates the annotation system
func TestSchemaEngine_AnnotationSystem(t *testing.T) {
	engine := NewSchemaEngine()

	fmt.Println("\n=== Annotation System Demo ===")

	// Test built-in annotations
	builtins := []string{"pattern", "behavior", "deployment", "caching", "performance", "security"}

	for _, name := range builtins {
		if !engine.HasAnnotation(name) {
			t.Errorf("Built-in annotation %s not found", name)
			continue
		}

		schema, exists := engine.GetAnnotationSchema(name)
		if !exists {
			t.Errorf("Failed to get schema for annotation %s", name)
			continue
		}

		fmt.Printf("✓ Built-in annotation: %s (type: %s)\n", name, schema.Type())
	}

	// Test custom annotation registration
	customAnnotation := StringEnumAnnotation("development", "staging", "production")
	err := engine.RegisterAnnotation("environment", customAnnotation)
	if err != nil {
		t.Fatalf("Failed to register custom annotation: %v", err)
	}

	fmt.Printf("✓ Registered custom annotation: environment\n")

	// Test custom annotation validation
	err = engine.ValidateAnnotation("environment", "production")
	if err != nil {
		t.Fatalf("Failed to validate custom annotation: %v", err)
	}

	fmt.Printf("✓ Validated custom annotation: environment=production\n")

	// Test invalid custom annotation value
	err = engine.ValidateAnnotation("environment", "invalid")
	if err != nil {
		fmt.Printf("✓ Proper error for invalid custom annotation: %v\n", err)
	} else {
		t.Error("Expected error for invalid custom annotation value")
	}

	fmt.Println("=== Annotation System Demo Complete ===")
}

// TestSchemaEngine_ReferenceSystem demonstrates the reference system
func TestSchemaEngine_ReferenceSystem(t *testing.T) {
	fmt.Println("\n=== Reference System Demo ===")

	// Test reference parsing
	testCases := []string{
		"User",
		"auth:User",
		"User@v1.0",
		"auth:User@v1.0",
		"billing:Order@v2.1.3",
	}

	for _, refStr := range testCases {
		ref, err := ParseReference(refStr)
		if err != nil {
			t.Errorf("Failed to parse reference %s: %v", refStr, err)
			continue
		}

		fmt.Printf("✓ Parsed '%s' -> Name: %s, Namespace: %s, Version: %s, Full: %s\n",
			refStr, ref.Name(), ref.Namespace(), ref.Version(), ref.FullName())

		// Test validation
		if err := ref.Validate(); err != nil {
			t.Errorf("Reference validation failed for %s: %v", refStr, err)
		}
	}

	// Test reference set
	refSet := NewReferenceSet()

	refSet.Add(Ref("User"))
	refSet.Add(RefNS("auth", "User"))
	refSet.Add(RefVer("", "User", "v1.0"))

	fmt.Printf("✓ Reference set size: %d\n", refSet.Size())

	// Test filtering
	authRefs := refSet.FilterByNamespace("auth")
	fmt.Printf("✓ References in 'auth' namespace: %d\n", len(authRefs))

	versionedRefs := refSet.FilterByVersion("v1.0")
	fmt.Printf("✓ References with version 'v1.0': %d\n", len(versionedRefs))

	fmt.Println("=== Reference System Demo Complete ===")
}

// Benchmark tests for performance
func BenchmarkSchemaEngine_RegisterSchema(b *testing.B) {
	engine := NewSchemaEngine()
	schema := builders2.NewStringSchema().Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("Schema%d", i)
		engine.RegisterSchema(name, schema)
	}
}

func BenchmarkSchemaEngine_ResolveSchema(b *testing.B) {
	engine := NewSchemaEngine()
	schema := builders2.NewStringSchema().Build()

	// Pre-register schemas
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("Schema%d", i)
		engine.RegisterSchema(name, schema)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("Schema%d", i%100)
		engine.ResolveSchema(name)
	}
}

func BenchmarkSchemaEngine_ValidateAnnotation(b *testing.B) {
	engine := NewSchemaEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ValidateAnnotation("pattern", "service")
	}
}
