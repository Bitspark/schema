package registry

import (
	"testing"

	"defs.dev/schema"
)

func TestRegistry_BasicOperations(t *testing.T) {
	reg := New()

	// Test Define and Get
	userSchema := schema.NewObject().
		Property("id", schema.NewInteger().Build()).
		Property("name", schema.NewString().MinLength(1).Build()).
		Required("id", "name").
		Build()

	err := reg.Define("User", userSchema)
	if err != nil {
		t.Fatalf("Failed to define schema: %v", err)
	}

	// Test Get
	retrieved, err := reg.Get("User")
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}

	if retrieved.Type() != schema.TypeObject {
		t.Errorf("Expected object schema, got %s", retrieved.Type())
	}

	// Test Exists
	if !reg.Exists("User") {
		t.Error("Expected User schema to exist")
	}

	if reg.Exists("NonExistent") {
		t.Error("Expected NonExistent schema to not exist")
	}

	// Test List
	names := reg.List()
	if len(names) != 1 || names[0] != "User" {
		t.Errorf("Expected ['User'], got %v", names)
	}
}

func TestRegistry_ParameterizedSchemas(t *testing.T) {
	reg := New()

	// Define a parameterized list schema
	listSchema := schema.NewArray().Items(Param("T")).Build()
	err := reg.Define("List", listSchema, "T")
	if err != nil {
		t.Fatalf("Failed to define parameterized schema: %v", err)
	}

	// Test Parameters method
	params := reg.Parameters("List")
	if len(params) != 1 || params[0] != "T" {
		t.Errorf("Expected ['T'], got %v", params)
	}

	// Define User schema for parameter
	userSchema := schema.NewObject().
		Property("name", schema.NewString().Build()).
		Required("name").
		Build()
	reg.Define("User", userSchema)

	// Test applying parameters
	userList, err := reg.Apply("List", map[string]schema.Schema{
		"T": reg.Ref("User"),
	})
	if err != nil {
		t.Fatalf("Failed to apply parameters: %v", err)
	}

	if userList.Type() != schema.TypeArray {
		t.Errorf("Expected array schema, got %s", userList.Type())
	}

	// Test fluent builder API
	userList2, err := reg.Build("List").WithParam("T", reg.Ref("User")).Build()
	if err != nil {
		t.Fatalf("Failed to build with fluent API: %v", err)
	}

	if userList2.Type() != schema.TypeArray {
		t.Errorf("Expected array schema, got %s", userList2.Type())
	}
}

func TestRegistry_SchemaRef(t *testing.T) {
	reg := New()

	// Define schemas
	userSchema := schema.NewObject().
		Property("name", schema.NewString().Build()).
		Required("name").
		Build()
	reg.Define("User", userSchema)

	// Create schema reference
	userRef := reg.Ref("User")
	if userRef.Type() != schema.TypeRef {
		t.Errorf("Expected ref schema, got %s", userRef.Type())
	}

	// Test validation through reference
	validUser := map[string]any{
		"name": "John Doe",
	}

	result := userRef.Validate(validUser)
	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}

	// Test invalid data
	invalidUser := map[string]any{
		"age": 30, // missing name
	}

	result = userRef.Validate(invalidUser)
	if result.Valid {
		t.Error("Expected invalid result for user without name")
	}
}

func TestRegistry_ErrorHandling(t *testing.T) {
	reg := New()

	// Test schema not found
	_, err := reg.Get("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent schema")
	}

	registryErr, ok := err.(*RegistryError)
	if !ok {
		t.Errorf("Expected RegistryError, got %T", err)
	}

	if registryErr.Type != "not_found" {
		t.Errorf("Expected 'not_found' error type, got %s", registryErr.Type)
	}

	// Test invalid parameters
	listSchema := schema.NewArray().Items(Param("T")).Build()
	reg.Define("List", listSchema, "T")

	// Missing parameter
	_, err = reg.Apply("List", map[string]schema.Schema{})
	if err == nil {
		t.Error("Expected error for missing parameter")
	}

	// Extra parameter
	userSchema := schema.NewObject().Property("name", schema.NewString().Build()).Build()
	_, err = reg.Apply("List", map[string]schema.Schema{
		"T":     userSchema,
		"Extra": schema.NewString().Build(),
	})
	if err == nil {
		t.Error("Expected error for extra parameter")
	}
}

func TestRegistry_Clone(t *testing.T) {
	reg := New()

	userSchema := schema.NewObject().
		Property("name", schema.NewString().Build()).
		Build()
	reg.Define("User", userSchema)

	// Clone registry
	cloned := reg.Clone()

	// Verify clone has same schemas
	if !cloned.Exists("User") {
		t.Error("Expected cloned registry to have User schema")
	}

	// Verify independence
	cloned.Define("Product", schema.NewObject().Build())

	if reg.Exists("Product") {
		t.Error("Original registry should not have Product schema")
	}

	if !cloned.Exists("Product") {
		t.Error("Cloned registry should have Product schema")
	}
}

func TestRegistry_Merge(t *testing.T) {
	reg1 := New()
	reg2 := New()

	// Define schemas in different registries
	reg1.Define("User", schema.NewObject().Property("name", schema.NewString().Build()).Build())
	reg2.Define("Product", schema.NewObject().Property("price", schema.NewNumber().Build()).Build())

	// Merge reg2 into reg1
	err := reg1.Merge(reg2)
	if err != nil {
		t.Fatalf("Failed to merge registries: %v", err)
	}

	// Verify both schemas exist in reg1
	if !reg1.Exists("User") {
		t.Error("Expected merged registry to have User schema")
	}

	if !reg1.Exists("Product") {
		t.Error("Expected merged registry to have Product schema")
	}

	// Verify reg2 is unchanged
	if !reg2.Exists("Product") {
		t.Error("Original registry should still have Product schema")
	}

	if reg2.Exists("User") {
		t.Error("Original registry should not have User schema")
	}
}
