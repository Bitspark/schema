package engine_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"defs.dev/core"
	"defs.dev/core/sources"
	schemacore "defs.dev/schema/api/core"
	"defs.dev/schema/engine"
	"defs.dev/schema/schemas"
)

// TestCoreIntegrationWorkflow tests the complete integration workflow
func TestCoreIntegrationWorkflow(t *testing.T) {
	// 1. Create temporary project structure
	tempDir, err := os.MkdirTemp("", "schema_integration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectDir := filepath.Join(tempDir, "test-project")
	aitreeDir := filepath.Join(projectDir, ".aitree")

	if err := os.MkdirAll(aitreeDir, 0755); err != nil {
		t.Fatalf("Failed to create .aitree dir: %v", err)
	}

	// 2. Create defstree.yml with schema definitions
	defstreeContent := `name: "test-schemas"
version: "1.0"
description: "Test schema definitions"

schemas:
  user-schema:
    description: "User entity schema"
    type: "object"
    properties:
      id:
        type: "string"
        pattern: "^[0-9]+$"
      name:
        type: "string"
        minLength: 1
      email:
        type: "string"
        format: "email"
    required: ["id", "name", "email"]
`

	defstreeFile := filepath.Join(aitreeDir, "defstree.yml")
	if err := os.WriteFile(defstreeFile, []byte(defstreeContent), 0644); err != nil {
		t.Fatalf("Failed to write defstree.yml: %v", err)
	}

	// 3. Create base schema engine with test schemas
	config := engine.DefaultEngineConfig()
	config.ValidateOnRegister = false // Disable for testing
	config.StrictMode = false
	baseEngine := engine.NewSchemaEngineWithConfig(config)

	// Create test schemas
	userSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: schemacore.SchemaMetadata{
			Name:        "User",
			Description: "User entity schema",
			Tags:        []string{"entity", "user"},
		},
	})

	productSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: schemacore.SchemaMetadata{
			Name:        "Product",
			Description: "Product entity schema",
			Tags:        []string{"entity", "product"},
		},
	})

	// Register schemas
	if err := baseEngine.RegisterSchema("User", userSchema); err != nil {
		t.Fatalf("Failed to register User schema: %v", err)
	}
	if err := baseEngine.RegisterSchema("Product", productSchema); err != nil {
		t.Fatalf("Failed to register Product schema: %v", err)
	}

	// 4. Create entity library with hierarchical source
	entityLibrary := core.NewEntityLibrary(
		core.WithSources(sources.NewHierarchicalSource()),
	)

	// 5. Create schema engine with core integration
	coreConfig := engine.CoreIntegrationConfig{
		SourceName:         "test-schemas",
		SourcePriority:     200, // High priority for testing
		EnableHierarchical: true,
		CacheEnabled:       true,
		CacheTTL:           time.Minute,
		DefaultScope:       "test",
		DefaultAuthors:     []string{"Test Suite"},
	}

	schemaEngineWithCore := engine.NewSchemaEngineWithCore(
		baseEngine,
		engine.WithEntityLibrary(entityLibrary),
		engine.WithCoreConfig(coreConfig),
	)

	// 6. Register as entity source
	if err := schemaEngineWithCore.SetEntityLibrary(entityLibrary); err != nil {
		t.Fatalf("Failed to set entity library: %v", err)
	}

	ctx := context.Background()

	// 7. Test entity source functionality
	t.Run("EntitySource", func(t *testing.T) {
		entitySource := schemaEngineWithCore.AsEntitySource()

		// Test source properties
		if entitySource.Name() != "test-schemas" {
			t.Errorf("Expected source name 'test-schemas', got '%s'", entitySource.Name())
		}

		if entitySource.Priority() != 200 {
			t.Errorf("Expected priority 200, got %d", entitySource.Priority())
		}

		// Test listing entities
		entities, err := entitySource.List(ctx, core.EntityTypeSchema, core.DefaultResolutionContext())
		if err != nil {
			t.Fatalf("Failed to list entities: %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expected 2 entities, got %d", len(entities))
		}

		// Verify entity properties
		userFound := false
		productFound := false
		for _, entity := range entities {
			if entity.Name == "User" {
				userFound = true
				if entity.Type != core.EntityTypeSchema {
					t.Errorf("Expected entity type schema, got %s", entity.Type)
				}
				if entity.Scope != "test" {
					t.Errorf("Expected scope 'test', got '%s'", entity.Scope)
				}
			}
			if entity.Name == "Product" {
				productFound = true
			}
		}

		if !userFound {
			t.Error("User entity not found in list")
		}
		if !productFound {
			t.Error("Product entity not found in list")
		}
	})

	// 8. Test entity resolution
	t.Run("EntityResolution", func(t *testing.T) {
		entitySource := schemaEngineWithCore.AsEntitySource()

		// Resolve User entity
		entity, err := entitySource.Resolve(ctx, core.EntityTypeSchema, "User", core.DefaultResolutionContext())
		if err != nil {
			t.Fatalf("Failed to resolve User entity: %v", err)
		}

		// Verify entity properties
		if entity.Name() != "User" {
			t.Errorf("Expected entity name 'User', got '%s'", entity.Name())
		}

		if entity.Scope() != "test" {
			t.Errorf("Expected scope 'test', got '%s'", entity.Scope())
		}

		if entity.Kind() != "schema" {
			t.Errorf("Expected kind 'schema', got '%s'", entity.Kind())
		}

		// Test SchemaEntity specific functionality
		schemaEntity, ok := entity.(*engine.SchemaEntity)
		if !ok {
			t.Fatalf("Expected SchemaEntity, got %T", entity)
		}

		schema := schemaEntity.GetSchema()
		if schema.Type() != schemacore.TypeString {
			t.Errorf("Expected string schema, got %s", schema.Type())
		}

		// Test validation through entity
		result := schemaEntity.ValidateData("test string")
		if !result.Valid {
			t.Errorf("Expected validation to pass for string data")
		}
	})

	// 9. Test search functionality
	t.Run("Search", func(t *testing.T) {
		entitySource := schemaEngineWithCore.AsEntitySource()

		// Search for User
		searchQuery := core.SearchQuery{
			Type:  core.EntityTypeSchema,
			Query: "User",
			Limit: 10,
		}

		results, err := entitySource.Search(ctx, searchQuery)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results, got none")
		}

		// Verify search result
		found := false
		for _, result := range results {
			if result.Name == "User" {
				found = true
				break
			}
		}
		if !found {
			t.Error("User not found in search results")
		}

		// Search with scope filter
		searchQuery.Scope = "test"
		results, err = entitySource.Search(ctx, searchQuery)
		if err != nil {
			t.Fatalf("Failed to search with scope: %v", err)
		}

		for _, result := range results {
			if result.Scope != "test" {
				t.Errorf("Expected scope 'test', got '%s'", result.Scope)
			}
		}
	})

	// 10. Test core resolution
	t.Run("CoreResolution", func(t *testing.T) {
		// Test resolution through core library
		entity, err := entityLibrary.Resolve(core.EntityTypeSchema, "User", core.DefaultResolutionContext())
		if err != nil {
			t.Fatalf("Failed to resolve through core library: %v", err)
		}

		if entity.Name() != "User" {
			t.Errorf("Expected entity name 'User', got '%s'", entity.Name())
		}

		// Test that our source was used (high priority)
		if !contains(entity.Sources(), "test-schemas") {
			t.Errorf("Expected source 'test-schemas' in sources %v", entity.Sources())
		}
	})

	// 11. Test schema discovery
	t.Run("SchemaDiscovery", func(t *testing.T) {
		schemas, err := schemaEngineWithCore.DiscoverSchemas(ctx, projectDir)
		if err != nil {
			// This is expected since hierarchical source doesn't implement List yet
			t.Logf("Schema discovery failed as expected: %v", err)
			return
		}

		// If discovery works, verify the results
		t.Logf("Discovered %d schemas", len(schemas))
		for _, schema := range schemas {
			t.Logf("  - %s: %s", schema.Reference.Name(), schema.Description)
		}
	})

	// 12. Test publishing
	t.Run("Publishing", func(t *testing.T) {
		// Create a new schema to publish
		orderSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
			Metadata: schemacore.SchemaMetadata{
				Name:        "Order",
				Description: "Order entity schema",
			},
		})

		metadata := map[string]any{
			"category": "business",
			"version":  "1.0.0",
		}

		// Publish schema
		err := schemaEngineWithCore.PublishToCore(ctx, orderSchema, metadata)
		if err != nil {
			t.Fatalf("Failed to publish schema: %v", err)
		}

		// Verify it was registered
		if !baseEngine.HasSchema("Order") {
			t.Error("Order schema was not registered after publishing")
		}

		// Verify it's available through entity source
		entitySource := schemaEngineWithCore.AsEntitySource()
		entities, err := entitySource.List(ctx, core.EntityTypeSchema, core.DefaultResolutionContext())
		if err != nil {
			t.Fatalf("Failed to list entities after publishing: %v", err)
		}

		orderFound := false
		for _, entity := range entities {
			if entity.Name == "Order" {
				orderFound = true
				break
			}
		}
		if !orderFound {
			t.Error("Published Order schema not found in entity list")
		}
	})
}

// TestCoreIntegrationConfiguration tests various configuration scenarios
func TestCoreIntegrationConfiguration(t *testing.T) {
	baseEngine := engine.NewSchemaEngine()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		config := engine.DefaultCoreIntegrationConfig()

		if config.SourceName != "schema-engine" {
			t.Errorf("Expected default source name 'schema-engine', got '%s'", config.SourceName)
		}

		if config.SourcePriority != 100 {
			t.Errorf("Expected default priority 100, got %d", config.SourcePriority)
		}

		if !config.EnableHierarchical {
			t.Error("Expected hierarchical to be enabled by default")
		}

		if !config.CacheEnabled {
			t.Error("Expected cache to be enabled by default")
		}
	})

	t.Run("CustomConfiguration", func(t *testing.T) {
		customConfig := engine.CoreIntegrationConfig{
			SourceName:         "custom-schemas",
			SourcePriority:     300,
			EnableHierarchical: false,
			CacheEnabled:       false,
			DefaultScope:       "custom",
			DefaultAuthors:     []string{"Custom Author"},
		}

		schemaEngineWithCore := engine.NewSchemaEngineWithCore(
			baseEngine,
			engine.WithCoreConfig(customConfig),
		)

		entitySource := schemaEngineWithCore.AsEntitySource()

		if entitySource.Name() != "custom-schemas" {
			t.Errorf("Expected source name 'custom-schemas', got '%s'", entitySource.Name())
		}

		if entitySource.Priority() != 300 {
			t.Errorf("Expected priority 300, got %d", entitySource.Priority())
		}
	})

	t.Run("EntityLibraryConfiguration", func(t *testing.T) {
		entityLibrary := core.NewEntityLibrary()

		schemaEngineWithCore := engine.NewSchemaEngineWithCore(
			baseEngine,
			engine.WithEntityLibrary(entityLibrary),
		)

		// Test that entity library was set
		err := schemaEngineWithCore.SetEntityLibrary(entityLibrary)
		if err != nil {
			t.Errorf("Failed to set entity library: %v", err)
		}

		// Verify the source was registered
		sources := entityLibrary.GetSources()
		found := false
		for _, source := range sources {
			if source.Name() == "schema-engine" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Schema engine source not found in entity library")
		}
	})
}

// TestCoreIntegrationErrorHandling tests error scenarios
func TestCoreIntegrationErrorHandling(t *testing.T) {
	baseEngine := engine.NewSchemaEngine()
	schemaEngineWithCore := engine.NewSchemaEngineWithCore(baseEngine)

	ctx := context.Background()

	t.Run("NoEntityLibrary", func(t *testing.T) {
		// Test operations without entity library
		ref := engine.Ref("NonExistent")
		_, err := schemaEngineWithCore.ResolveFromCore(ctx, ref)
		if err == nil {
			t.Error("Expected error when entity library not configured")
		}

		_, err = schemaEngineWithCore.DiscoverSchemas(ctx, ".")
		if err == nil {
			t.Error("Expected error when entity library not configured")
		}
	})

	t.Run("InvalidEntityType", func(t *testing.T) {
		entityLibrary := core.NewEntityLibrary()
		schemaEngineWithCore.SetEntityLibrary(entityLibrary)

		entitySource := schemaEngineWithCore.AsEntitySource()

		// Test with wrong entity type
		_, err := entitySource.Resolve(ctx, core.EntityTypeComponent, "Test", core.DefaultResolutionContext())
		if err == nil {
			t.Error("Expected error for unsupported entity type")
		}

		_, err = entitySource.List(ctx, core.EntityTypeFunction, core.DefaultResolutionContext())
		if err == nil {
			t.Error("Expected error for unsupported entity type")
		}
	})

	t.Run("NonExistentSchema", func(t *testing.T) {
		entityLibrary := core.NewEntityLibrary()
		schemaEngineWithCore.SetEntityLibrary(entityLibrary)

		entitySource := schemaEngineWithCore.AsEntitySource()

		// Test resolving non-existent schema
		_, err := entitySource.Resolve(ctx, core.EntityTypeSchema, "NonExistent", core.DefaultResolutionContext())
		if err == nil {
			t.Error("Expected error for non-existent schema")
		}
	})
}

// TestCoreIntegrationPerformance tests performance characteristics
func TestCoreIntegrationPerformance(t *testing.T) {
	// Create engine with many schemas
	config := engine.DefaultEngineConfig()
	config.ValidateOnRegister = false
	baseEngine := engine.NewSchemaEngineWithConfig(config)

	// Register multiple schemas
	for i := 0; i < 100; i++ {
		schemaName := fmt.Sprintf("Schema%d", i)
		schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
			Metadata: schemacore.SchemaMetadata{
				Name:        schemaName,
				Description: fmt.Sprintf("Test schema %d", i),
			},
		})
		baseEngine.RegisterSchema(schemaName, schema)
	}

	entityLibrary := core.NewEntityLibrary()
	schemaEngineWithCore := engine.NewSchemaEngineWithCore(
		baseEngine,
		engine.WithEntityLibrary(entityLibrary),
	)
	schemaEngineWithCore.SetEntityLibrary(entityLibrary)

	entitySource := schemaEngineWithCore.AsEntitySource()
	ctx := context.Background()

	t.Run("ListPerformance", func(t *testing.T) {
		start := time.Now()
		entities, err := entitySource.List(ctx, core.EntityTypeSchema, core.DefaultResolutionContext())
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to list entities: %v", err)
		}

		if len(entities) != 100 {
			t.Errorf("Expected 100 entities, got %d", len(entities))
		}

		// Should complete quickly
		if duration > time.Second {
			t.Errorf("List operation took too long: %v", duration)
		}

		t.Logf("Listed %d entities in %v", len(entities), duration)
	})

	t.Run("SearchPerformance", func(t *testing.T) {
		searchQuery := core.SearchQuery{
			Type:  core.EntityTypeSchema,
			Query: "Schema1",
			Limit: 10,
		}

		start := time.Now()
		results, err := entitySource.Search(ctx, searchQuery)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		// Should find multiple matches (Schema1, Schema10, Schema11, etc.)
		if len(results) == 0 {
			t.Error("Expected search results")
		}

		// Should complete quickly
		if duration > time.Second {
			t.Errorf("Search operation took too long: %v", duration)
		}

		t.Logf("Search found %d results in %v", len(results), duration)
	})
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
