# Annotation Architecture: Engine vs Registry vs Separate Package

## 🔍 Current State Analysis

Looking at `engine/engine.go`, there's already a well-designed annotation system:

```go
// Engine has annotation management interfaces
RegisterAnnotation(name string, schema AnnotationSchema) error
ValidateAnnotation(name string, value any) error
GetAnnotationSchema(name string) (AnnotationSchema, bool)
ListAnnotations() []string
HasAnnotation(name string) bool

// AnnotationSchema is a constrained schema type
type AnnotationSchema interface {
    core.Schema
    ValidateAsAnnotation() error // Additional validation for annotations
}
```

## 🤔 Design Decision: Three Approaches

### Option 1: Keep Annotations in Engine (Current)
**Pros:**
- Already implemented and designed
- Engine as central coordinator makes sense
- Unified management of schemas, types, and annotations

**Cons:**
- Engine becomes heavy with multiple responsibilities
- Lower-level components (registry, native) can't use annotations directly
- Creates dependency issues for native package

### Option 2: Move to Registry Package
**Pros:**
- Registry becomes the central "metadata management" layer
- Native can use annotations without depending on engine
- Cleaner separation of concerns

**Cons:**
- Registry becomes complex with validators + annotations
- Engine loses control over annotation lifecycle
- Potential duplication with engine's schema management

### Option 3: Separate Annotation Package ⭐ **RECOMMENDED**
**Pros:**
- Clean separation of concerns
- Can be used by registry, native, schemas, and engine
- Dedicated focus on annotation-specific logic
- Flexible dependency hierarchy

**Cons:**
- Additional package to maintain
- Need coordination between annotation package and engine

## 🏗️ Recommended Architecture: Separate Annotation Package

### Revised Dependency Hierarchy

```
Level 1: api/core          # Core interfaces
Level 2: annotation        # Annotation system (NEW)
Level 3: registry          # Validator registry (uses annotation)  
Level 4: native            # Type reflection (uses annotation + registry)
Level 5: schemas           # Schema implementations 
Level 6: builders          # Fluent builders
Level 7: engine            # Coordination (delegates to annotation + registry)
```

### Package Responsibilities

```go
// annotation/ - Pure annotation system
- Annotation interfaces and types
- Annotation validation logic
- Annotation registry implementation
- No dependencies on engine/schemas/builders

// registry/ - Validator registry  
- Validator interfaces and implementations
- Uses annotation package for metadata
- Validator registration and lookup

// native/ - Type reflection
- Uses annotation package for field metadata
- Uses registry for validation
- No dependency on engine/builders

// engine/ - Coordination layer
- Delegates annotation management to annotation package
- Orchestrates registry + annotation + native
- High-level schema management
```

## 🔧 Implementation Design

### 1. New Annotation Package

```go
// annotation/types.go
package annotation

import "defs.dev/schema/api/core"

// Annotation represents a single annotation with name, value, and metadata
type Annotation interface {
    Name() string
    Value() any
    Schema() core.Schema
    Validators() []string
    Metadata() AnnotationMetadata
}

// AnnotationRegistry manages annotation type definitions and validation
type AnnotationRegistry interface {
    // Type management
    RegisterType(name string, schema core.Schema) error
    GetType(name string) (core.Schema, bool)
    ListTypes() []string
    
    // Instance management
    Create(name string, value any) (Annotation, error)
    Validate(annotation Annotation) ValidationResult
    
    // Bulk operations
    CreateMany(annotations map[string]any) ([]Annotation, error)
    ValidateMany(annotations []Annotation) ValidationResult
}

// AnnotationMetadata provides metadata about annotation types
type AnnotationMetadata struct {
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Examples     []any             `json:"examples,omitempty"`
    DefaultValue any               `json:"default_value,omitempty"`
    Required     bool              `json:"required,omitempty"`
    Validators   []string          `json:"validators,omitempty"`
    Tags         []string          `json:"tags,omitempty"`
    Properties   map[string]string `json:"properties,omitempty"`
}

// ValidationResult represents annotation validation results
type ValidationResult struct {
    Valid        bool                    `json:"valid"`
    Errors       []ValidationError       `json:"errors,omitempty"`
    Warnings     []ValidationWarning     `json:"warnings,omitempty"`
    Metadata     map[string]any          `json:"metadata,omitempty"`
}
```

### 2. Built-in Annotation Types

```go
// annotation/builtin.go
package annotation

// Built-in annotation type definitions
func RegisterBuiltinTypes(registry AnnotationRegistry) error {
    // String annotations
    registry.RegisterType("format", stringSchema("email", "url", "uuid", "phone"))
    registry.RegisterType("pattern", regexSchema())
    registry.RegisterType("minLength", positiveIntegerSchema())
    registry.RegisterType("maxLength", positiveIntegerSchema())
    
    // Numeric annotations
    registry.RegisterType("min", numberSchema())
    registry.RegisterType("max", numberSchema())
    registry.RegisterType("range", arraySchema(numberSchema(), 2, 2))
    
    // Array annotations
    registry.RegisterType("minItems", nonNegativeIntegerSchema())
    registry.RegisterType("maxItems", nonNegativeIntegerSchema())
    registry.RegisterType("uniqueItems", booleanSchema())
    
    // Validation annotations
    registry.RegisterType("required", booleanSchema())
    registry.RegisterType("validators", arraySchema(stringSchema()))
    
    // Metadata annotations
    registry.RegisterType("description", stringSchema())
    registry.RegisterType("examples", arraySchema(anySchema()))
    registry.RegisterType("default", anySchema())
    
    return nil
}
```

### 3. Integration with Registry Package

```go
// registry/integration.go
package registry

import (
    "defs.dev/schema/annotation"
    "defs.dev/schema/api/core"
)

type ValidatorRegistryImpl struct {
    validators map[string]core.Validator
    annotations annotation.AnnotationRegistry
    mu         sync.RWMutex
}

func NewValidatorRegistry() core.ValidatorRegistry {
    annotations := annotation.NewRegistry()
    annotation.RegisterBuiltinTypes(annotations) // Register built-in annotations
    
    registry := &ValidatorRegistryImpl{
        validators:  make(map[string]core.Validator),
        annotations: annotations,
    }
    
    registry.registerBuiltinValidators()
    return registry
}

// Validators can use annotations for configuration
func (r *ValidatorRegistryImpl) registerBuiltinValidators() {
    r.Register("email", &EmailValidator{
        FormatAnnotation: r.annotations.GetType("format"),
    })
    r.Register("minLength", &MinLengthValidator{
        LengthAnnotation: r.annotations.GetType("minLength"),
    })
    // ... other validators
}
```

### 4. Native Package Integration

```go
// native/annotations.go
package native

import (
    "defs.dev/schema/annotation"
    "defs.dev/schema/api/core"
)

// FieldAnnotations manages annotations discovered from struct tags
type FieldAnnotations struct {
    registry annotation.AnnotationRegistry
    annotations []annotation.Annotation
}

func (f *FieldAnnotations) ParseStructTags(field reflect.StructField) error {
    tags := parseStructTags(field)
    
    for name, value := range tags {
        if ann, err := f.registry.Create(name, value); err == nil {
            f.annotations = append(f.annotations, ann)
        }
    }
    
    return nil
}

func (f *FieldAnnotations) ApplyToSchema(schema core.Schema) core.Schema {
    metadata := schema.Metadata()
    
    // Convert annotations to schema metadata
    if metadata.Annotations == nil {
        metadata.Annotations = make(map[string]annotation.Annotation)
    }
    
    for _, ann := range f.annotations {
        metadata.Annotations[ann.Name()] = ann
    }
    
    return schema.WithMetadata(metadata)
}
```

### 5. Engine Delegation

```go
// engine/impl.go - Engine delegates to annotation package
package engine

import "defs.dev/schema/annotation"

type SchemaEngineImpl struct {
    // ... other fields
    annotations annotation.AnnotationRegistry
}

func newSchemaEngineImpl(config EngineConfig) SchemaEngine {
    annotations := annotation.NewRegistry()
    annotation.RegisterBuiltinTypes(annotations)
    
    return &SchemaEngineImpl{
        annotations: annotations,
        // ... other initialization
    }
}

// Delegate annotation operations to annotation package
func (e *SchemaEngineImpl) RegisterAnnotation(name string, schema AnnotationSchema) error {
    return e.annotations.RegisterType(name, schema)
}

func (e *SchemaEngineImpl) ValidateAnnotation(name string, value any) error {
    ann, err := e.annotations.Create(name, value)
    if err != nil {
        return err
    }
    
    result := e.annotations.Validate(ann)
    if !result.Valid {
        return fmt.Errorf("annotation validation failed: %v", result.Errors)
    }
    
    return nil
}

func (e *SchemaEngineImpl) GetAnnotationSchema(name string) (AnnotationSchema, bool) {
    schema, exists := e.annotations.GetType(name)
    if !exists {
        return nil, false
    }
    
    // Wrap core.Schema as AnnotationSchema if needed
    if annotationSchema, ok := schema.(AnnotationSchema); ok {
        return annotationSchema, true
    }
    
    return &wrappedAnnotationSchema{schema}, true
}
```

## 🚀 Migration Strategy

### Phase 1: Create Annotation Package (Week 1)
- [ ] Create `annotation/` package with core interfaces
- [ ] Implement `AnnotationRegistry` 
- [ ] Register built-in annotation types
- [ ] Create comprehensive tests

### Phase 2: Registry Integration (Week 2)  
- [ ] Update `registry/` to use annotation package
- [ ] Integrate validators with annotations
- [ ] Update validator implementations

### Phase 3: Native Integration (Week 3)
- [ ] Update `native/` to use annotations for struct tags
- [ ] Integrate with registry for validation
- [ ] Update schema creation to include annotations

### Phase 4: Engine Delegation (Week 4)
- [ ] Update `engine/` to delegate to annotation package
- [ ] Maintain backward compatibility with existing interfaces
- [ ] Update all dependent code

## 🎯 Benefits of Separate Annotation Package

### 1. **Clean Dependencies**
```
annotation -> api/core (no circular dependencies)
registry -> annotation (can use annotations)
native -> annotation + registry (can use both)
engine -> annotation + registry + native (coordinates all)
```

### 2. **Flexible Usage**
```go
// Native can use annotations directly
fieldAnnotations := native.ParseStructTags(field)
annotatedSchema := fieldAnnotations.ApplyToSchema(schema)

// Registry can validate using annotations  
validator := registry.GetValidator("email")
validation := validator.ValidateWithAnnotations(value, annotations)

// Engine coordinates everything
engine.RegisterAnnotation("custom", customSchema)
```

### 3. **Extensibility**
```go
// Custom annotation types
registry.RegisterType("currency", currencySchema())

// Custom validators using annotations
registry.Register("price", &PriceValidator{
    CurrencyAnnotation: registry.GetAnnotationType("currency"),
    MinAnnotation: registry.GetAnnotationType("min"),
    MaxAnnotation: registry.GetAnnotationType("max"),
})
```

### 4. **Type Safety**
```go
// Annotations are validated against their schemas
formatAnnotation := registry.Create("format", "email") // Validated
invalidAnnotation := registry.Create("format", 123)    // Error: invalid type
```

## 🏆 Final Recommendation

**Create a separate `annotation` package** that:

1. **Provides core annotation interfaces and registry**
2. **Can be used by registry, native, and engine packages**  
3. **Maintains clean dependency hierarchy**
4. **Enables flexible annotation-based validation**
5. **Keeps engine focused on coordination rather than implementation**

This approach gives us the best of all worlds: clean architecture, flexibility, and the ability to use annotations throughout the system without circular dependencies. 