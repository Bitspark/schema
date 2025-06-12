# Schema Native: Architectural Redesign

## 🚨 Critical Architecture Issues Identified

You've identified several fundamental architectural problems with the original DRAFT_1 approach that require a complete redesign:

### 1. **Circular Dependency Problem**
- If `builders` uses `native`, then `native` cannot use `builders`
- Native needs to work at a **lower level** in the dependency hierarchy
- Native should create schemas directly, not through builders

### 2. **Service/Function Integration Gap**
- Current service discovery uses basic reflection in `service.go:createSchemaFromType()`
- Functions have complex argument schemas (`ArgSchema`, `ArgSchemas`) that need native support
- Service methods need to be discoverable through reflection with tag support

### 3. **Hardcoded Format/Pattern System**
- Current formats (email, url, uuid) are hardcoded in `string.go:validateFormat()`
- Patterns are regex-only, not extensible
- Should become **annotations** with pluggable validators

### 4. **Missing Validator Registry**
- No flexible validator registration system
- Validation logic is scattered across schema implementations
- Need `api` interface with `registry` implementation

## 🏗️ Revised Architecture

### Dependency Hierarchy (Bottom to Top)

```
Level 1: api/core          # Core interfaces only
Level 2: registry          # Validator registry, annotation system  
Level 3: native            # Type reflection, struct analysis (NO builders)
Level 4: schemas           # Schema implementations using native + registry
Level 5: builders          # Fluent builders using schemas + native
Level 6: engine            # Coordination layer using all above
```

### Key Principle: Native Works Below Builders

```go
// ❌ WRONG (circular dependency)
native -> builders -> schemas
builders -> native

// ✅ CORRECT (linear dependency) 
api -> registry -> native -> schemas -> builders -> engine
```

## 🔧 Core Components Redesign

### 1. Validator Registry System

**New Files:**
- `api/core/validator.go` - Core validator interfaces
- `registry/validators.go` - Validator registry implementation
- `registry/annotations.go` - Annotation system

```go
// api/core/validator.go
package core

// Validator validates values against specific criteria
type Validator interface {
    Name() string
    Validate(value any) ValidationResult
    Metadata() ValidatorMetadata
}

// ValidatorRegistry manages validator registration and lookup
type ValidatorRegistry interface {
    Register(name string, validator Validator) error
    Get(name string) (Validator, bool)
    List() []string
    
    // Annotation system
    RegisterAnnotation(name string, schema Schema) error
    ValidateAnnotation(name string, value any) ValidationResult
    GetAnnotationSchema(name string) (Schema, bool)
}

type ValidatorMetadata struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Examples    []any    `json:"examples,omitempty"`
    Tags        []string `json:"tags,omitempty"`
}

// Annotation represents a schema annotation with validation
type Annotation interface {
    Name() string
    Value() any
    Validators() []string
}
```

**Registry Implementation:**
```go
// registry/validators.go
package registry

import "defs.dev/schema/api/core"

type ValidatorRegistryImpl struct {
    validators  map[string]core.Validator
    annotations map[string]core.Schema
    mu          sync.RWMutex
}

func NewValidatorRegistry() core.ValidatorRegistry {
    registry := &ValidatorRegistryImpl{
        validators:  make(map[string]core.Validator),
        annotations: make(map[string]core.Schema),
    }
    
    // Register built-in validators
    registry.registerBuiltinValidators()
    return registry
}

func (r *ValidatorRegistryImpl) registerBuiltinValidators() {
    // String format validators
    r.Register("email", &EmailValidator{})
    r.Register("url", &URLValidator{})
    r.Register("uuid", &UUIDValidator{})
    r.Register("phone", &PhoneValidator{})
    r.Register("regex", &RegexValidator{})
    
    // Numeric validators
    r.Register("min", &MinValidator{})
    r.Register("max", &MaxValidator{})
    r.Register("range", &RangeValidator{})
    
    // Array validators  
    r.Register("minItems", &MinItemsValidator{})
    r.Register("maxItems", &MaxItemsValidator{})
    r.Register("uniqueItems", &UniqueItemsValidator{})
}
```

### 2. Annotation System Integration

**Transform Current Formats into Annotations:**

```go
// BEFORE: Hardcoded in string.go
func (b *StringBuilder) Email() core.StringSchemaBuilder {
    return b.Format("email").Description("Valid email address")
}

// AFTER: Annotation-based
func (b *StringBuilder) Email() core.StringSchemaBuilder {
    return b.WithAnnotation("format", "email").
        WithAnnotation("description", "Valid email address")
}

// Or even more flexible:
func (b *StringBuilder) WithValidator(name string, config any) core.StringSchemaBuilder {
    return b.WithAnnotation("validator", map[string]any{
        "name": name,
        "config": config,
    })
}
```

**Schema Annotations Enhancement:**
```go
// Enhanced SchemaMetadata in api/core/types.go
type SchemaMetadata struct {
    Name        string                 `json:"name,omitempty"`
    Description string                 `json:"description,omitempty"`
    Examples    []any                  `json:"examples,omitempty"`
    Tags        []string               `json:"tags,omitempty"`
    Properties  map[string]string      `json:"properties,omitempty"`
    
    // NEW: Annotation support
    Annotations map[string]Annotation  `json:"annotations,omitempty"`
}

type Annotation struct {
    Name       string   `json:"name"`
    Value      any      `json:"value"`
    Validators []string `json:"validators,omitempty"`
}
```

### 3. Native Package Redesign (No Builder Dependencies)

**Core Principle**: Native creates schemas directly, not through builders

```go
// native/converter.go - Core conversion without builders
package native

import (
    "reflect"
    "defs.dev/schema/api/core"
    "defs.dev/schema/schemas"  // Can use schema implementations
    "defs.dev/schema/registry" // Can use validators
)

// TypeConverter converts Go types to schemas (no builders!)
type TypeConverter interface {
    CanConvert(t reflect.Type) bool
    Convert(t reflect.Type, opts *Options) core.Schema
}

// Primary API - creates schemas directly
func FromType(t reflect.Type, opts ...*Options) core.Schema {
    options := mergeOptions(opts...)
    return convertType(t, options)
}

func FromValue[T any](opts ...*Options) core.Schema {
    var zero T
    return FromType(reflect.TypeOf(zero), opts...)
}

// Core conversion function - creates schemas directly
func convertType(t reflect.Type, opts *Options) core.Schema {
    // Check cache
    if cached := opts.cache.Get(t); cached != nil {
        return cached
    }
    
    // Try custom converters
    for _, converter := range opts.converters {
        if converter.CanConvert(t) {
            schema := converter.Convert(t, opts)
            opts.cache.Set(t, schema)
            return schema
        }
    }
    
    // Built-in type conversion
    switch t.Kind() {
    case reflect.String:
        return createStringSchema(t, opts)
    case reflect.Struct:
        return createObjectSchema(t, opts)
    // ... other types
    }
}

// Create schemas directly using schema constructors
func createStringSchema(t reflect.Type, opts *Options) core.Schema {
    config := schemas.StringSchemaConfig{
        Metadata: core.SchemaMetadata{},
    }
    
    // Apply any struct field tags if this came from a field
    if fieldInfo := opts.GetFieldInfo(t); fieldInfo != nil {
        applyStringTags(&config, fieldInfo.Tags)
    }
    
    return schemas.NewStringSchema(config)
}

func applyStringTags(config *schemas.StringSchemaConfig, tags map[string]string) {
    if minLen, ok := tags["minLength"]; ok {
        if val := parseInt(minLen); val != nil {
            config.MinLength = val
        }
    }
    
    if format, ok := tags["format"]; ok {
        // Instead of hardcoded format, use annotation
        if config.Metadata.Annotations == nil {
            config.Metadata.Annotations = make(map[string]core.Annotation)
        }
        config.Metadata.Annotations["format"] = core.Annotation{
            Name:       "format",
            Value:      format,
            Validators: []string{format}, // Validator name matches format
        }
    }
}
```

### 4. Service/Function Integration with Native

**Enhanced Service Discovery:**
```go
// native/service.go - Service-specific reflection
package native

// ServiceInfo represents discovered service information
type ServiceInfo struct {
    Name        string
    Methods     map[string]*MethodInfo
    Metadata    core.SchemaMetadata
}

type MethodInfo struct {
    Name        string
    Function    core.FunctionSchema
    Inputs      map[string]core.Schema
    Outputs     map[string]core.Schema
    Metadata    core.SchemaMetadata
}

// AnalyzeService discovers service methods using reflection + tags
func AnalyzeService(instance any, opts *Options) *ServiceInfo {
    t := reflect.TypeOf(instance)
    v := reflect.ValueOf(instance)
    
    if t.Kind() == reflect.Ptr {
        t = t.Elem() 
        v = v.Elem()
    }
    
    info := &ServiceInfo{
        Name:    getServiceName(t),
        Methods: make(map[string]*MethodInfo),
    }
    
    // Analyze methods
    for i := 0; i < t.NumMethod(); i++ {
        method := t.Method(i)
        if isServiceMethod(method) {
            methodInfo := analyzeMethod(method, opts)
            info.Methods[method.Name] = methodInfo
        }
    }
    
    return info
}

func analyzeMethod(method reflect.Method, opts *Options) *MethodInfo {
    methodType := method.Type
    
    // Create ArgSchemas for inputs (skip receiver)
    inputs := make(map[string]core.Schema)
    for i := 1; i < methodType.NumIn(); i++ {
        paramType := methodType.In(i)
        paramName := getParameterName(method, i-1) // From tags or reflection
        inputs[paramName] = FromType(paramType, opts)
    }
    
    // Create ArgSchemas for outputs  
    outputs := make(map[string]core.Schema)
    for i := 0; i < methodType.NumOut(); i++ {
        outputType := methodType.Out(i)
        if isErrorType(outputType) {
            continue // Handle separately
        }
        outputName := getOutputName(method, i)
        outputs[outputName] = FromType(outputType, opts) 
    }
    
    // Create function schema directly (no builders!)
    functionSchema := createFunctionSchemaFromMethod(method, inputs, outputs)
    
    return &MethodInfo{
        Name:     method.Name,
        Function: functionSchema,
        Inputs:   inputs,
        Outputs:  outputs,
    }
}
```

### 5. Builder Enhancement (Uses Native)

**Builders can now use native without circular dependency:**

```go
// builders/object.go - Enhanced with native
func (b *ObjectBuilder) FromValue[T any]() core.ObjectSchemaBuilder {
    // Use native to convert type to schema
    nativeSchema := native.FromValue[T]()
    
    // If it's already an object schema, use its properties
    if objSchema, ok := nativeSchema.(core.ObjectSchema); ok {
        // Copy properties from native schema
        for name, propSchema := range objSchema.Properties() {
            b.Property(name, propSchema)
        }
        // Copy required fields
        b.Required(objSchema.Required()...)
    }
    
    return b
}

// builders/service.go - Enhanced FromStruct  
func (b *ServiceBuilder) FromStruct(instance any) core.ServiceSchemaBuilder {
    // Use native for sophisticated analysis
    serviceInfo := native.AnalyzeService(instance, &native.Options{
        EnableTagProcessing: true,
        ValidatorRegistry:   registry.Global(),
    })
    
    b.name = serviceInfo.Name
    
    // Add discovered methods
    for methodName, methodInfo := range serviceInfo.Methods {
        b.methods[methodName] = methodInfo.Function
    }
    
    return b
}
```

## 🔄 Migration Strategy

### Phase 1: Validator Registry (Week 1)
- [ ] Create `api/core/validator.go` interfaces
- [ ] Implement `registry/validators.go` 
- [ ] Create built-in validators (email, url, uuid, etc.)
- [ ] Add annotation support to `SchemaMetadata`

### Phase 2: Native Package (No Builders) (Week 2)
- [ ] Create `native/converter.go` with direct schema creation
- [ ] Create `native/service.go` for service discovery
- [ ] Implement tag processing system
- [ ] Create comprehensive tests

### Phase 3: Schema Integration (Week 3)
- [ ] Update `schemas/string.go` to use validator registry
- [ ] Update all schema types to support annotations
- [ ] Replace hardcoded validation with registry lookups
- [ ] Integrate native metadata

### Phase 4: Builder Enhancement (Week 4)
- [ ] Add `FromValue[T]()` methods to builders using native
- [ ] Enhance `ServiceBuilder.FromStruct` with native analysis
- [ ] Update all builders to support annotations
- [ ] Comprehensive integration testing

## 🎯 Key Benefits

### 1. **No Circular Dependencies**
```
api -> registry -> native -> schemas -> builders
```
Clean dependency hierarchy with no cycles.

### 2. **Flexible Validation**
```go
// Register custom validator
registry.Register("phone", &CustomPhoneValidator{})

// Use in schema
schema.WithAnnotation("format", "phone")
```

### 3. **Enhanced Service Discovery**
```go
type UserService struct{}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    // Method discovered with full type information
}

// Native discovers: inputs, outputs, error handling
serviceSchema := native.AnalyzeService(&UserService{})
```

### 4. **Annotation-Based Validation**
```go
// Instead of hardcoded formats
type User struct {
    Email string `json:"email" format:"email" description:"User email"`
    Phone string `json:"phone" validator:"phone" pattern:"custom-phone-regex"`  
}

// Multiple validators per field
type Product struct {
    Price float64 `json:"price" min:"0" max:"10000" currency:"USD"`
}
```

## 📊 Architecture Comparison

### Before (Problematic)
```
native -> builders -> schemas  (circular dependency)
hardcoded formats in string.go
basic reflection in service.go
no validator extensibility
```

### After (Clean)
```
api -> registry -> native -> schemas -> builders -> engine
pluggable validators via registry  
sophisticated service discovery via native
annotation-based validation
zero circular dependencies
```

This redesign solves all the architectural issues while providing a much more flexible and extensible system. 