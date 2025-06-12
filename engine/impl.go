package engine

import (
	"fmt"
	"sync"

	"defs.dev/schema/api/core"
)

// schemaEngineImpl is the concrete implementation of SchemaEngine
type schemaEngineImpl struct {
	// Configuration
	config EngineConfig

	// Schema resolution
	schemas  map[string]core.Schema
	schemaMu sync.RWMutex

	// Type extensions
	typeFactories map[string]SchemaTypeFactory
	typesMu       sync.RWMutex

	// Annotation registry
	annotations map[string]AnnotationSchema
	annotMu     sync.RWMutex

	// Dependency resolution cache
	resolutionCache map[string]core.Schema
	cacheMu         sync.RWMutex

	// Global mutex for operations that need to coordinate across systems
	globalMu sync.RWMutex
}

// newSchemaEngineImpl creates a new schema engine implementation
func newSchemaEngineImpl(config EngineConfig) SchemaEngine {
	engine := &schemaEngineImpl{
		config:          config,
		schemas:         make(map[string]core.Schema),
		typeFactories:   make(map[string]SchemaTypeFactory),
		annotations:     make(map[string]AnnotationSchema),
		resolutionCache: make(map[string]core.Schema),
	}

	// Register built-in annotations
	engine.registerBuiltinAnnotations()

	return engine
}

// Schema Resolution Methods

func (e *schemaEngineImpl) RegisterSchema(name string, schema core.Schema) error {
	if name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	e.schemaMu.Lock()
	defer e.schemaMu.Unlock()

	// Check if schema already exists
	if _, exists := e.schemas[name]; exists {
		return NewSchemaExistsError(name)
	}

	// Validate schema if configured to do so
	if e.config.ValidateOnRegister {
		if result := schema.Validate(nil); !result.Valid {
			return EngineError{
				Type:    ErrorTypeValidationFailed,
				Message: fmt.Sprintf("schema validation failed for %s", name),
				Details: map[string]any{"validation_errors": result.Errors},
			}
		}
	}

	// Register the schema
	e.schemas[name] = schema

	// Clear related cache entries
	e.clearRelatedCache(name)

	return nil
}

func (e *schemaEngineImpl) ResolveSchema(name string) (core.Schema, error) {
	if name == "" {
		return nil, fmt.Errorf("schema name cannot be empty")
	}

	e.schemaMu.RLock()
	schema, exists := e.schemas[name]
	e.schemaMu.RUnlock()

	if !exists {
		return nil, NewSchemaNotFoundError(name)
	}

	return schema, nil
}

func (e *schemaEngineImpl) ResolveReference(ref SchemaReference) (core.Schema, error) {
	if ref == nil {
		return nil, fmt.Errorf("reference cannot be nil")
	}

	if err := ref.Validate(); err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	fullName := ref.FullName()

	// Check cache first
	if e.config.EnableCache {
		if cached := e.getCached(fullName); cached != nil {
			return cached, nil
		}
	}

	// Create resolution context to detect circular dependencies
	ctx := &resolutionContext{
		visited:  make(map[string]bool),
		stack:    []string{},
		depth:    0,
		maxDepth: e.config.CircularDepthLimit,
	}

	// Resolve with context
	schema, err := e.resolveWithContext(ref, ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if e.config.EnableCache {
		e.setCached(fullName, schema)
	}

	return schema, nil
}

func (e *schemaEngineImpl) ListSchemas() []string {
	e.schemaMu.RLock()
	defer e.schemaMu.RUnlock()

	names := make([]string, 0, len(e.schemas))
	for name := range e.schemas {
		names = append(names, name)
	}

	return names
}

func (e *schemaEngineImpl) HasSchema(name string) bool {
	e.schemaMu.RLock()
	defer e.schemaMu.RUnlock()

	_, exists := e.schemas[name]
	return exists
}

// Extension Management Methods

func (e *schemaEngineImpl) RegisterSchemaType(typeName string, factory SchemaTypeFactory) error {
	if typeName == "" {
		return fmt.Errorf("type name cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	e.typesMu.Lock()
	defer e.typesMu.Unlock()

	// Check if type already exists
	if _, exists := e.typeFactories[typeName]; exists {
		return EngineError{
			Type:    ErrorTypeTypeExists,
			Message: "schema type already exists: " + typeName,
			Details: map[string]any{"type_name": typeName},
		}
	}

	// Validate the factory
	if err := e.validateFactory(factory); err != nil {
		return fmt.Errorf("invalid factory for type %s: %w", typeName, err)
	}

	// Register the factory
	e.typeFactories[typeName] = factory

	return nil
}

func (e *schemaEngineImpl) CreateSchema(typeName string, config any) (core.Schema, error) {
	if typeName == "" {
		return nil, fmt.Errorf("type name cannot be empty")
	}

	e.typesMu.RLock()
	factory, exists := e.typeFactories[typeName]
	e.typesMu.RUnlock()

	if !exists {
		return nil, NewTypeNotFoundError(typeName)
	}

	// Validate configuration
	if err := factory.ValidateConfig(config); err != nil {
		return nil, EngineError{
			Type:    ErrorTypeInvalidConfig,
			Message: fmt.Sprintf("invalid configuration for type %s: %v", typeName, err),
			Details: map[string]any{"type_name": typeName, "config": config},
		}
	}

	// Create schema
	schema, err := factory.CreateSchema(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema of type %s: %w", typeName, err)
	}

	return schema, nil
}

func (e *schemaEngineImpl) GetAvailableTypes() []string {
	e.typesMu.RLock()
	defer e.typesMu.RUnlock()

	types := make([]string, 0, len(e.typeFactories))
	for typeName := range e.typeFactories {
		types = append(types, typeName)
	}

	return types
}

func (e *schemaEngineImpl) ValidateTypeConfig(typeName string, config any) error {
	if typeName == "" {
		return fmt.Errorf("type name cannot be empty")
	}

	e.typesMu.RLock()
	factory, exists := e.typeFactories[typeName]
	e.typesMu.RUnlock()

	if !exists {
		return NewTypeNotFoundError(typeName)
	}

	return factory.ValidateConfig(config)
}

func (e *schemaEngineImpl) HasSchemaType(typeName string) bool {
	e.typesMu.RLock()
	defer e.typesMu.RUnlock()

	_, exists := e.typeFactories[typeName]
	return exists
}

// Annotation System Methods

func (e *schemaEngineImpl) RegisterAnnotation(name string, schema AnnotationSchema) error {
	if name == "" {
		return fmt.Errorf("annotation name cannot be empty")
	}

	if schema == nil {
		return fmt.Errorf("annotation schema cannot be nil")
	}

	e.annotMu.Lock()
	defer e.annotMu.Unlock()

	// Check if annotation already exists
	if _, exists := e.annotations[name]; exists {
		return EngineError{
			Type:    ErrorTypeAnnotationExists,
			Message: "annotation already exists: " + name,
			Details: map[string]any{"annotation_name": name},
		}
	}

	// Validate as annotation schema (primitives only)
	if err := schema.ValidateAsAnnotation(); err != nil {
		return EngineError{
			Type:    ErrorTypeInvalidAnnotation,
			Message: fmt.Sprintf("invalid annotation schema for %s: %v", name, err),
			Details: map[string]any{"annotation_name": name},
		}
	}

	// Register the annotation
	e.annotations[name] = schema

	return nil
}

func (e *schemaEngineImpl) ValidateAnnotation(name string, value any) error {
	if name == "" {
		return fmt.Errorf("annotation name cannot be empty")
	}

	e.annotMu.RLock()
	annotSchema, exists := e.annotations[name]
	e.annotMu.RUnlock()

	if !exists {
		if e.config.StrictMode {
			return EngineError{
				Type:    ErrorTypeAnnotationNotFound,
				Message: "unknown annotation: " + name,
				Details: map[string]any{"annotation_name": name},
			}
		}
		return nil // Allow unknown annotations in non-strict mode
	}

	// Validate value against annotation schema
	result := annotSchema.Validate(value)
	if !result.Valid {
		return EngineError{
			Type:    ErrorTypeValidationFailed,
			Message: fmt.Sprintf("invalid annotation %s", name),
			Details: map[string]any{
				"annotation_name":   name,
				"validation_errors": result.Errors,
				"value":             value,
			},
		}
	}

	return nil
}

func (e *schemaEngineImpl) GetAnnotationSchema(name string) (AnnotationSchema, bool) {
	e.annotMu.RLock()
	defer e.annotMu.RUnlock()

	schema, exists := e.annotations[name]
	return schema, exists
}

func (e *schemaEngineImpl) ListAnnotations() []string {
	e.annotMu.RLock()
	defer e.annotMu.RUnlock()

	names := make([]string, 0, len(e.annotations))
	for name := range e.annotations {
		names = append(names, name)
	}

	return names
}

func (e *schemaEngineImpl) HasAnnotation(name string) bool {
	e.annotMu.RLock()
	defer e.annotMu.RUnlock()

	_, exists := e.annotations[name]
	return exists
}

// Engine Management Methods

func (e *schemaEngineImpl) Validate() error {
	// Validate all registered schemas
	e.schemaMu.RLock()
	schemas := make(map[string]core.Schema)
	for name, schema := range e.schemas {
		schemas[name] = schema
	}
	e.schemaMu.RUnlock()

	for name, schema := range schemas {
		if result := schema.Validate(nil); !result.Valid {
			return fmt.Errorf("schema %s is invalid: %v", name, result.Errors)
		}
	}

	// Validate all annotation schemas
	e.annotMu.RLock()
	annotations := make(map[string]AnnotationSchema)
	for name, schema := range e.annotations {
		annotations[name] = schema
	}
	e.annotMu.RUnlock()

	for name, schema := range annotations {
		if err := schema.ValidateAsAnnotation(); err != nil {
			return fmt.Errorf("annotation schema %s is invalid: %w", name, err)
		}
	}

	return nil
}

func (e *schemaEngineImpl) Reset() error {
	e.globalMu.Lock()
	defer e.globalMu.Unlock()

	// Clear all registries
	e.schemaMu.Lock()
	e.schemas = make(map[string]core.Schema)
	e.schemaMu.Unlock()

	e.typesMu.Lock()
	e.typeFactories = make(map[string]SchemaTypeFactory)
	e.typesMu.Unlock()

	e.annotMu.Lock()
	e.annotations = make(map[string]AnnotationSchema)
	e.annotMu.Unlock()

	// Clear cache
	e.cacheMu.Lock()
	e.resolutionCache = make(map[string]core.Schema)
	e.cacheMu.Unlock()

	// Re-register built-in annotations
	e.registerBuiltinAnnotations()

	return nil
}

func (e *schemaEngineImpl) Clone() SchemaEngine {
	e.globalMu.RLock()
	defer e.globalMu.RUnlock()

	clone := &schemaEngineImpl{
		config:          e.config,
		schemas:         make(map[string]core.Schema),
		typeFactories:   make(map[string]SchemaTypeFactory),
		annotations:     make(map[string]AnnotationSchema),
		resolutionCache: make(map[string]core.Schema),
	}

	// Copy schemas
	e.schemaMu.RLock()
	for name, schema := range e.schemas {
		clone.schemas[name] = schema.Clone()
	}
	e.schemaMu.RUnlock()

	// Copy type factories
	e.typesMu.RLock()
	for name, factory := range e.typeFactories {
		clone.typeFactories[name] = factory
	}
	e.typesMu.RUnlock()

	// Copy annotations
	e.annotMu.RLock()
	for name, schema := range e.annotations {
		clone.annotations[name] = schema.(AnnotationSchema)
	}
	e.annotMu.RUnlock()

	return clone
}

// Configuration Methods

func (e *schemaEngineImpl) Config() EngineConfig {
	return e.config
}

func (e *schemaEngineImpl) WithConfig(config EngineConfig) SchemaEngine {
	clone := e.Clone().(*schemaEngineImpl)
	clone.config = config
	return clone
}

// Helper Methods

// Resolution context for circular dependency detection
type resolutionContext struct {
	visited  map[string]bool
	stack    []string
	depth    int
	maxDepth int
}

func (e *schemaEngineImpl) resolveWithContext(ref SchemaReference, ctx *resolutionContext) (core.Schema, error) {
	fullName := ref.FullName()

	// Check for cycles
	if ctx.visited[fullName] {
		return nil, NewCircularDependencyError(ctx.stack)
	}

	// Check depth limit
	if ctx.depth >= ctx.maxDepth {
		return nil, fmt.Errorf("resolution depth limit exceeded: %d", ctx.maxDepth)
	}

	// Add to context
	ctx.visited[fullName] = true
	ctx.stack = append(ctx.stack, fullName)
	ctx.depth++

	defer func() {
		// Remove from context
		delete(ctx.visited, fullName)
		ctx.stack = ctx.stack[:len(ctx.stack)-1]
		ctx.depth--
	}()

	// For now, resolve using simple name lookup
	// In future phases, this will handle namespaces and versions
	return e.ResolveSchema(ref.Name())
}

// Cache management
func (e *schemaEngineImpl) getCached(key string) core.Schema {
	e.cacheMu.RLock()
	defer e.cacheMu.RUnlock()

	return e.resolutionCache[key]
}

func (e *schemaEngineImpl) setCached(key string, schema core.Schema) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// Check cache size limit
	if len(e.resolutionCache) >= e.config.MaxCacheSize {
		// Simple eviction: clear the cache
		e.resolutionCache = make(map[string]core.Schema)
	}

	e.resolutionCache[key] = schema
}

func (e *schemaEngineImpl) clearRelatedCache(schemaName string) {
	if !e.config.EnableCache {
		return
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// For now, clear entire cache when any schema changes
	// In the future, we could be more intelligent about this
	e.resolutionCache = make(map[string]core.Schema)
}

// Factory validation
func (e *schemaEngineImpl) validateFactory(factory SchemaTypeFactory) error {
	// Check if factory can provide metadata
	metadata := factory.GetMetadata()
	if metadata.Name == "" {
		return fmt.Errorf("factory metadata must have a name")
	}

	// Check if factory can provide config schema
	configSchema := factory.GetConfigSchema()
	if configSchema == nil {
		return fmt.Errorf("factory must provide a config schema")
	}

	// Validate the config schema itself
	if result := configSchema.Validate(nil); !result.Valid {
		return fmt.Errorf("factory config schema is invalid: %v", result.Errors)
	}

	return nil
}

// registerBuiltinAnnotations is implemented in annotations.go
