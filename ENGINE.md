# Schema Engine Implementation Plan

## 🎯 Overview

The **Schema Engine** is the central coordination layer that manages all cross-cutting concerns in the schema system. It provides unified management for schema resolution, type extensions, and annotations while maintaining backward compatibility with existing functionality.

## 🏗️ Architecture Goals

### Core Principles
1. **Centralized Coordination** - Single point of control for schema system operations
2. **Extensibility** - Plugin system for new schema types and annotations
3. **Type Safety** - All extensions validated against their own schemas
4. **Performance** - Cached resolution and optimized validation
5. **Backward Compatibility** - Existing code continues to work unchanged

### System Architecture
```
                    ┌─────────────────────────┐
                    │     SCHEMA ENGINE       │
                    │    (Central Kernel)     │
                    │                         │
                    │ ┌─────┐ ┌─────┐ ┌─────┐ │
                    │ │Res- │ │Ext- │ │Ann- │ │
                    │ │olu- │ │ens- │ │ota- │ │
                    │ │tion │ │ions │ │tions│ │
                    │ └─────┘ └─────┘ └─────┘ │
                    └─────────────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
    ┌─────▼─────┐       ┌─────▼─────┐       ┌─────▼─────┐
    │  Schema   │       │ Builder   │       │ Portal    │
    │  Types    │       │  System   │       │  System   │
    └───────────┘       └───────────┘       └───────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
                    ┌─────────▼─────────┐
                    │   Validation &    │
                    │   Serialization   │
                    └───────────────────┘
```

## 📋 Implementation Phases

## Phase 1: Core Engine Infrastructure

### 1.1 Base Engine Interface
**Location**: `engine/engine.go`

```go
package engine

type SchemaEngine interface {
    // Schema Resolution
    RegisterSchema(name string, schema core.Schema) error
    ResolveSchema(name string) (core.Schema, error)
    ResolveReference(ref SchemaReference) (core.Schema, error)
    ListSchemas() []string
    
    // Extension Management
    RegisterSchemaType(typeName string, factory SchemaTypeFactory) error
    CreateSchema(typeName string, config any) (core.Schema, error)
    GetAvailableTypes() []string
    ValidateTypeConfig(typeName string, config any) error
    
    // Annotation System
    RegisterAnnotation(name string, schema AnnotationSchema) error
    ValidateAnnotation(name string, value any) error
    GetAnnotationSchema(name string) (AnnotationSchema, bool)
    ListAnnotations() []string
    
    // Engine Management
    Validate() error
    Reset() error
    Clone() SchemaEngine
}
```

### 1.2 Core Implementation
**Location**: `engine/impl.go`

```go
type SchemaEngineImpl struct {
    // Schema resolution
    schemas     map[string]core.Schema
    schemaMu    sync.RWMutex
    
    // Type extensions
    typeFactories map[string]SchemaTypeFactory
    typesMu       sync.RWMutex
    
    // Annotation registry
    annotations map[string]AnnotationSchema
    annotMu     sync.RWMutex
    
    // Dependency resolution
    resolutionCache map[string]core.Schema
    cacheMu         sync.RWMutex
    
    // Configuration
    config EngineConfig
}

type EngineConfig struct {
    EnableCache        bool
    MaxCacheSize       int
    CircularDepthLimit int
    StrictMode         bool
}
```

### 1.3 Schema Reference System
**Location**: `engine/references.go`

```go
type SchemaReference interface {
    Name() string
    Namespace() string
    Version() string
    FullName() string
}

type SimpleReference struct {
    name      string
    namespace string
    version   string
}

// Reference builder for use in schemas
func Ref(name string) SchemaReference
func RefNS(namespace, name string) SchemaReference  
func RefVer(namespace, name, version string) SchemaReference
```

## Phase 2: Extension Management

### 2.1 Schema Type Factory System
**Location**: `engine/extensions.go`

```go
type SchemaTypeFactory interface {
    // Create schema instance from configuration
    CreateSchema(config any) (core.Schema, error)
    
    // Validate configuration before creation
    ValidateConfig(config any) error
    
    // Get schema that describes the configuration format
    GetConfigSchema() core.Schema
    
    // Metadata about this schema type
    GetMetadata() SchemaTypeMetadata
}

type SchemaTypeMetadata struct {
    Name        string
    Description string
    Version     string
    Author      string
    Category    string
    Tags        []string
}
```

### 2.2 Built-in Schema Types
**Location**: `engine/builtin_types.go`

Pre-registered schema types that extend the core system:

```go
// Advanced string types
EmailSchemaFactory    // Email validation with domain restrictions
PhoneSchemaFactory    // Phone number validation with formats
URLSchemaFactory      // URL validation with scheme restrictions
UUIDSchemaFactory     // UUID validation with version constraints

// Date/Time types
DateTimeSchemaFactory // ISO8601 datetime with timezone support
DateSchemaFactory     // Date-only validation
TimeSchemaFactory     // Time-only validation
DurationSchemaFactory // Duration strings (ISO8601)

// Numeric types
CurrencySchemaFactory // Currency amounts with precision
PercentSchemaFactory  // Percentage values with range
RatioSchemaFactory    // Ratio values between 0-1

// Composite types
GeoPointSchemaFactory // Geographic coordinates
AddressSchemaFactory  // Structured address validation
```

### 2.3 Extension Registration
**Location**: `engine/registry.go`

```go
func (e *SchemaEngineImpl) RegisterSchemaType(name string, factory SchemaTypeFactory) error {
    // Validate factory
    if err := e.validateFactory(factory); err != nil {
        return fmt.Errorf("invalid factory for type %s: %w", name, err)
    }
    
    // Check for conflicts
    if e.hasType(name) {
        return fmt.Errorf("schema type %s already registered", name)
    }
    
    // Register
    e.typesMu.Lock()
    e.typeFactories[name] = factory
    e.typesMu.Unlock()
    
    return nil
}
```

## Phase 3: Annotation System

### 3.1 Annotation Schema Constraints
**Location**: `engine/annotations.go`

```go
type AnnotationSchema interface {
    core.Schema
    
    // Additional constraint validation
    ValidateAsAnnotation() error
}

// Constraint: Annotations can only use primitive compositions
func validateAnnotationSchema(schema core.Schema) error {
    visitor := &AnnotationValidator{}
    return schema.Accept(visitor)
}

type AnnotationValidator struct {
    allowedTypes map[core.SchemaType]bool
}

func (v *AnnotationValidator) VisitFunction(core.FunctionSchema) error {
    return errors.New("function schemas not allowed in annotations")
}

func (v *AnnotationValidator) VisitService(core.ServiceSchema) error {
    return errors.New("service schemas not allowed in annotations")
}
```

### 3.2 Built-in Annotations
**Location**: `engine/builtin_annotations.go`

```go
func (e *SchemaEngineImpl) registerBuiltinAnnotations() error {
    // Pattern annotations
    e.RegisterAnnotation("pattern", stringEnum(
        "service", "component", "entity", "value_object", 
        "aggregate", "repository", "factory",
    ))
    
    // Deployment annotations
    e.RegisterAnnotation("deployment", objectSchema(
        "strategy": stringEnum("rolling", "blue-green", "canary"),
        "replicas": integerRange(1, 100),
        "resources": objectSchema(
            "cpu": string(),
            "memory": string(),
            "storage": string(),
        ),
        "health_check": objectSchema(
            "endpoint": string(),
            "interval": duration(),
            "timeout": duration(),
        ),
    ))
    
    // Caching annotations
    e.RegisterAnnotation("caching", objectSchema(
        "strategy": stringEnum("redis", "memory", "disk", "none"),
        "ttl": integerMin(0),
        "key_pattern": string(),
        "invalidation": stringEnum("time", "event", "manual"),
    ))
    
    // Performance annotations
    e.RegisterAnnotation("performance", objectSchema(
        "timeout": duration(),
        "rate_limit": integerMin(1),
        "batch_size": integerRange(1, 1000),
        "async": boolean(),
    ))
    
    // Security annotations
    e.RegisterAnnotation("security", objectSchema(
        "authentication": stringEnum("required", "optional", "none"),
        "authorization": arrayOf(string()),
        "encryption": stringEnum("required", "optional", "none"),
        "audit": boolean(),
    ))
    
    return nil
}
```

### 3.3 Annotation Validation
**Location**: `engine/annotation_validation.go`

```go
func (e *SchemaEngineImpl) ValidateAnnotation(name string, value any) error {
    annotSchema, exists := e.getAnnotationSchema(name)
    if !exists {
        if e.config.StrictMode {
            return fmt.Errorf("unknown annotation: %s", name)
        }
        return nil // Allow unknown annotations in non-strict mode
    }
    
    // Validate value against annotation schema
    result := annotSchema.Validate(value)
    if !result.Valid {
        return fmt.Errorf("invalid annotation %s: %v", name, result.Errors)
    }
    
    return nil
}
```

## Phase 4: Schema Resolution

### 4.1 Named Schema Registry
**Location**: `engine/resolution.go`

```go
func (e *SchemaEngineImpl) RegisterSchema(name string, schema core.Schema) error {
    // Validate schema
    if err := e.validateSchema(schema); err != nil {
        return fmt.Errorf("invalid schema for %s: %w", name, err)
    }
    
    // Check for conflicts
    if e.hasSchema(name) {
        return fmt.Errorf("schema %s already registered", name)
    }
    
    // Register
    e.schemaMu.Lock()
    e.schemas[name] = schema
    e.schemaMu.Unlock()
    
    // Clear cache that might reference this schema
    e.clearRelatedCache(name)
    
    return nil
}
```

### 4.2 Reference Resolution
**Location**: `engine/resolver.go`

```go
func (e *SchemaEngineImpl) ResolveReference(ref SchemaReference) (core.Schema, error) {
    fullName := ref.FullName()
    
    // Check cache first
    if e.config.EnableCache {
        if cached, found := e.getCached(fullName); found {
            return cached, nil
        }
    }
    
    // Resolve schema
    schema, err := e.resolveUncached(ref)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    if e.config.EnableCache {
        e.setCached(fullName, schema)
    }
    
    return schema, nil
}
```

### 4.3 Circular Dependency Detection
**Location**: `engine/circular.go`

```go
type resolutionContext struct {
    visited map[string]bool
    stack   []string
    depth   int
    maxDepth int
}

func (e *SchemaEngineImpl) resolveWithContext(ref SchemaReference, ctx *resolutionContext) (core.Schema, error) {
    name := ref.FullName()
    
    // Check for cycles
    if ctx.visited[name] {
        return nil, fmt.Errorf("circular dependency detected: %v -> %s", ctx.stack, name)
    }
    
    // Check depth limit
    if ctx.depth >= ctx.maxDepth {
        return nil, fmt.Errorf("resolution depth limit exceeded: %d", ctx.maxDepth)
    }
    
    // Add to context
    ctx.visited[name] = true
    ctx.stack = append(ctx.stack, name)
    ctx.depth++
    
    defer func() {
        // Remove from context
        delete(ctx.visited, name)
        ctx.stack = ctx.stack[:len(ctx.stack)-1]
        ctx.depth--
    }()
    
    // Resolve schema
    return e.resolveUncached(ref)
}
```

## Phase 5: Integration with Existing System

### 5.1 Builder Integration
**Location**: `builders/engine_integration.go`

```go
// Add engine support to builders
type EngineAwareBuilder interface {
    WithEngine(engine engine.SchemaEngine) BuilderType
    Engine() engine.SchemaEngine
}

// Reference method for all builders
func (b *ObjectSchemaBuilder) Reference(name string) *ObjectSchemaBuilder {
    if b.engine == nil {
        panic("engine required for references - use WithEngine()")
    }
    ref := engine.Ref(name)
    return b.Property("$ref", ref)
}
```

### 5.2 Core Schema Updates
**Location**: `api/core/schemas.go` (additions)

```go
// Add reference support to base schema interface
type Schema interface {
    // ... existing methods ...
    
    // Engine integration
    Annotations() map[string]any
    HasAnnotation(key string) bool
    GetAnnotation(key string) (any, bool)
    
    // Reference resolution
    IsReference() bool
    GetReference() SchemaReference
}
```

### 5.3 Portal Integration
**Location**: `portal/engine_integration.go`

```go
// Portals use engine for schema resolution
type EngineAwarePortal interface {
    WithEngine(engine engine.SchemaEngine) Portal
    Engine() engine.SchemaEngine
}

func (p *HTTPPortal) WithEngine(eng engine.SchemaEngine) Portal {
    p.engine = eng
    return p
}

// Use engine for schema resolution in portals
func (p *HTTPPortal) resolveServiceSchema(address Address) (core.ServiceSchema, error) {
    if p.engine == nil {
        return p.resolveSchemaLegacy(address)
    }
    
    // Use engine for resolution
    schemaName := p.extractSchemaName(address)
    schema, err := p.engine.ResolveSchema(schemaName)
    if err != nil {
        return nil, err
    }
    
    return schema.(core.ServiceSchema), nil
}
```

## Phase 6: Advanced Features

### 6.1 Schema Versioning
**Location**: `engine/versioning.go`

```go
type VersionedReference struct {
    name      string
    namespace string
    version   string
    
    // Version constraints
    minVersion string
    maxVersion string
    compatible []string
}

func (e *SchemaEngineImpl) RegisterVersionedSchema(name, version string, schema core.Schema) error
func (e *SchemaEngineImpl) ResolveVersionedSchema(name, version string) (core.Schema, error)
func (e *SchemaEngineImpl) GetLatestVersion(name string) (string, error)
func (e *SchemaEngineImpl) ListVersions(name string) ([]string, error)
```

### 6.2 Schema Namespacing
**Location**: `engine/namespaces.go`

```go
type Namespace struct {
    name        string
    description string
    owner       string
    schemas     map[string]core.Schema
    imports     []string
}

func (e *SchemaEngineImpl) CreateNamespace(name string) error
func (e *SchemaEngineImpl) RegisterSchemaInNamespace(ns, name string, schema core.Schema) error
func (e *SchemaEngineImpl) ImportNamespace(target, source string) error
```

### 6.3 Schema Migration
**Location**: `engine/migration.go`

```go
type SchemaMigration interface {
    FromVersion() string
    ToVersion() string
    Migrate(oldSchema core.Schema) (core.Schema, error)
    Validate(oldSchema, newSchema core.Schema) error
}

func (e *SchemaEngineImpl) RegisterMigration(name string, migration SchemaMigration) error
func (e *SchemaEngineImpl) MigrateSchema(name, fromVersion, toVersion string) (core.Schema, error)
```

## 📁 Directory Structure

```
schema/
├── engine/
│   ├── engine.go              # Core engine interface
│   ├── impl.go                # Engine implementation
│   ├── references.go          # Reference system
│   ├── extensions.go          # Extension management
│   ├── annotations.go         # Annotation system
│   ├── resolution.go          # Schema resolution
│   ├── resolver.go            # Reference resolver
│   ├── circular.go            # Circular dependency detection
│   ├── builtin_types.go       # Built-in schema types
│   ├── builtin_annotations.go # Built-in annotations
│   ├── annotation_validation.go # Annotation validation
│   ├── versioning.go          # Schema versioning (Phase 6)
│   ├── namespaces.go          # Schema namespacing (Phase 6)
│   ├── migration.go           # Schema migration (Phase 6)
│   └── engine_test.go         # Comprehensive tests
├── builders/
│   └── engine_integration.go  # Builder-engine integration
├── portal/
│   └── engine_integration.go  # Portal-engine integration
└── examples/
    ├── engine_basic.go         # Basic engine usage
    ├── engine_extensions.go    # Custom type extensions
    └── engine_annotations.go   # Advanced annotations
```

## 🧪 Testing Strategy

### Unit Tests
- **Engine Core**: Registration, resolution, validation
- **Extensions**: Custom type factories and validation
- **Annotations**: Schema validation and constraints
- **References**: Resolution with caching and circular detection

### Integration Tests
- **Builder Integration**: References in complex schemas
- **Portal Integration**: Schema resolution across transports
- **Performance**: Large schema graphs, cache efficiency

### Example Tests
```go
func TestEngine_BasicResolution(t *testing.T)
func TestEngine_CircularDependencies(t *testing.T)
func TestEngine_CustomTypeExtensions(t *testing.T)
func TestEngine_AnnotationValidation(t *testing.T)
func TestEngine_PerformanceWithLargeGraphs(t *testing.T)
```

## 🚀 Migration Path

### Phase 1: Non-Breaking Introduction
1. Implement engine as optional component
2. Existing code continues to work unchanged
3. New features only available through engine

### Phase 2: Gradual Integration
1. Add engine support to builders (`WithEngine()` methods)
2. Add engine support to portals
3. Provide migration utilities for existing schemas

### Phase 3: Engine-First
1. Deprecate non-engine usage patterns
2. Make engine the default for new projects
3. Provide migration path for legacy code

## 📊 Success Metrics

1. **Functionality**: All planned features implemented and tested
2. **Performance**: No significant performance regression
3. **Compatibility**: 100% backward compatibility maintained
4. **Adoption**: Clean integration with existing builders and portals
5. **Extensibility**: Easy to add new schema types and annotations
6. **Documentation**: Comprehensive examples and guides

## 🔮 Future Extensions

### Phase 7: Advanced Tooling
- **Code Generation**: Use engine for generating code from schemas
- **Documentation**: Auto-generate docs from schema + annotations
- **Validation**: Runtime validation using engine-resolved schemas
- **Serialization**: Optimized serialization using schema metadata

### Phase 8: Distributed Engine
- **Remote Schemas**: Resolve schemas from remote registries
- **Schema Marketplace**: Share and discover schema extensions
- **Federated Namespaces**: Multi-organization schema sharing

This plan provides a comprehensive roadmap for implementing the Schema Engine while maintaining backward compatibility and providing a solid foundation for future extensions. 