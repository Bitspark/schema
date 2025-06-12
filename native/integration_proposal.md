# Schema Native: Enhanced Integration Proposal

## 🎯 Better Integration with Existing Schema System

Based on analysis of the current schema implementation, the `schema/native` package can be integrated much more deeply and effectively with the existing architecture.

## 🔄 Key Integration Points

### 1. **Replace/Enhance Existing Reflection Code**

The current `builders/service.go:createSchemaFromType()` function is a basic implementation that the native package should replace:

```go
// Current basic implementation (207 lines in service.go)
func createSchemaFromType(t reflect.Type) core.Schema {
    switch t.Kind() {
    case reflect.String:
        return NewStringSchema().Build()
    // ... basic type mapping only
    }
}

// Enhanced native implementation
func createSchemaFromType(t reflect.Type) core.Schema {
    return native.FromTypeWithOptions(t, &native.Options{
        UseExistingCache: true,
        IntegrateWithEngine: true,
    })
}
```

### 2. **Integrate with Schema Engine Extension System**

Register native converters as first-class schema type extensions:

```go
// In engine initialization
func (e *SchemaEngineImpl) RegisterNativeTypes() error {
    // Register built-in Go type converters
    e.RegisterSchemaType("native_time", &TimeTypeFactory{})
    e.RegisterSchemaType("native_uuid", &UUIDTypeFactory{})
    e.RegisterSchemaType("native_duration", &DurationTypeFactory{})
    
    // Register struct converter
    e.RegisterSchemaType("native_struct", &StructTypeFactory{})
    
    return nil
}

// Usage through engine
userSchema := engine.CreateSchema("native_struct", reflect.TypeOf(User{}))
```

### 3. **Enhance ServiceBuilder.FromStruct Integration**

The existing `FromStruct` method should leverage native's advanced capabilities:

```go
// Enhanced FromStruct using native package
func (b *ServiceBuilder) FromStruct(instance any) core.ServiceSchemaBuilder {
    if instance == nil {
        return b
    }

    // Use native package for sophisticated struct analysis
    opts := &native.Options{
        TagProcessors: native.DefaultTagProcessors(),
        CustomConverters: map[reflect.Type]native.TypeConverter{
            reflect.TypeOf(time.Time{}): &native.TimeConverter{},
        },
        RequireJSONTags: false,
        EnableMethodDiscovery: true, // Native-specific for services
    }
    
    // Let native handle the complex type analysis
    serviceInfo := native.AnalyzeServiceStruct(instance, opts)
    
    // Apply the analysis results
    b.name = serviceInfo.Name
    for methodName, methodSchema := range serviceInfo.Methods {
        b.methods[methodName] = methodSchema
    }
    
    return b
}
```

### 4. **Extend Core Builder Interfaces**

Add native-specific methods to existing builders:

```go
// Extend ObjectSchemaBuilder in api/core/builder.go
type ObjectSchemaBuilder interface {
    Builder[ObjectSchema]
    MetadataBuilder[ObjectSchemaBuilder]
    
    // Existing methods...
    Property(name string, schema Schema) ObjectSchemaBuilder
    Required(names ...string) ObjectSchemaBuilder
    
    // NEW: Native integration methods
    FromStruct(structType reflect.Type) ObjectSchemaBuilder
    FromStructWithTags(structType reflect.Type, tagOpts *native.TagOptions) ObjectSchemaBuilder
    FromValue[T any]() ObjectSchemaBuilder
}
```

### 5. **Integrate with Visitor Pattern**

Native schemas should work seamlessly with the existing visitor system:

```go
// Native schemas implement core.Accepter
type NativeStructSchema struct {
    *schemas.ObjectSchema // Embed existing object schema
    sourceType reflect.Type
    fieldMappings map[string]native.FieldInfo
}

func (s *NativeStructSchema) Accept(visitor core.SchemaVisitor) error {
    // First visit as object schema
    if err := s.ObjectSchema.Accept(visitor); err != nil {
        return err
    }
    
    // Then provide native-specific context if visitor supports it
    if nativeVisitor, ok := visitor.(native.NativeSchemaVisitor); ok {
        return nativeVisitor.VisitNativeStruct(s)
    }
    
    return nil
}
```

### 6. **Enhance Metadata System**

Integrate native field information with existing metadata:

```go
// Enhanced metadata for native schemas
type NativeSchemaMetadata struct {
    core.SchemaMetadata
    
    // Native-specific metadata
    SourceType    reflect.Type     `json:"source_type,omitempty"`
    FieldMappings map[string]FieldMetadata `json:"field_mappings,omitempty"`
    TagSources    map[string]string `json:"tag_sources,omitempty"`
}

type FieldMetadata struct {
    GoFieldName string            `json:"go_field_name"`
    JsonName    string            `json:"json_name"`
    Tags        map[string]string `json:"tags"`
    IsPointer   bool              `json:"is_pointer"`
    IsOptional  bool              `json:"is_optional"`
}
```

### 7. **Unify Factory Functions**

Extend existing factory functions to support native conversion:

```go
// In builders/factory.go (or new native/factory.go)

// Enhanced factory functions
func NewObject() *ObjectBuilder {
    return &ObjectBuilder{}
}

// NEW: Native-aware factory functions
func NewObjectFromStruct[T any]() *ObjectBuilder {
    return NewObject().FromValue[T]()
}

func NewObjectFromType(t reflect.Type) *ObjectBuilder {
    return NewObject().FromStruct(t)
}

// Service-specific factories
func NewServiceFromStruct[T any]() *ServiceBuilder {
    return NewServiceSchema().FromValue[T]()
}
```

### 8. **Integrate with Validation System**

Native tag constraints should integrate with existing validation:

```go
// Native constraints become first-class validation rules
type NativeValidationRule struct {
    FieldName string
    TagName   string
    TagValue  string
    GoType    reflect.Type
}

func (r *NativeValidationRule) Validate(value any) core.ValidationResult {
    // Convert native tag constraints to validation logic
    switch r.TagName {
    case "minLength":
        return validateMinLength(value, r.TagValue)
    case "format":
        return validateFormat(value, r.TagValue)
    // ... other native tag validations
    }
}
```

### 9. **Configuration Integration**

Native options should integrate with engine configuration:

```go
// Enhanced engine config
type EngineConfig struct {
    // Existing config...
    EnableCache        bool
    MaxCacheSize       int
    
    // NEW: Native integration config
    NativeOptions     *native.Options  `json:"native_options,omitempty"`
    EnableNativeTypes bool             `json:"enable_native_types"`
    NativeTagPrefix   string           `json:"native_tag_prefix"`
}

// Engine method to configure native integration
func (e *SchemaEngineImpl) ConfigureNative(opts *native.Options) error {
    e.config.NativeOptions = opts
    return e.reloadNativeConverters()
}
```

### 10. **Performance Integration**

Share caching between native package and existing schema system:

```go
// Shared cache key structure
type SchemaCacheKey struct {
    Type      reflect.Type
    Options   string // serialized options
    Tags      string // relevant tag hash
}

// Engine-level cache integration  
func (e *SchemaEngineImpl) getCachedSchema(key SchemaCacheKey) (core.Schema, bool) {
    e.cacheMu.RLock()
    defer e.cacheMu.RUnlock()
    
    schema, exists := e.resolutionCache[key.String()]
    return schema, exists
}
```

## 🚀 Implementation Strategy

### Phase 1: Core Integration (Week 1-2)
- [ ] Replace `createSchemaFromType` with native implementation
- [ ] Integrate with existing builder interfaces
- [ ] Add native factory functions

### Phase 2: Engine Integration (Week 3-4) 
- [ ] Register native converters as schema type extensions
- [ ] Integrate with engine configuration system
- [ ] Share caching infrastructure

### Phase 3: Advanced Features (Week 5-6)
- [ ] Enhanced visitor pattern support
- [ ] Advanced metadata integration
- [ ] Service discovery integration

### Phase 4: Optimization (Week 7-8)
- [ ] Performance optimization with shared caching
- [ ] Memory usage optimization
- [ ] Concurrent access optimization

## 🎯 Benefits of Enhanced Integration

1. **Single Source of Truth**: All reflection-based schema generation goes through native
2. **Consistent API**: Native features available through existing builder interfaces  
3. **Better Performance**: Shared caching and optimized reflection
4. **Richer Metadata**: Native field information available throughout system
5. **Extensibility**: Easy to add new native type converters
6. **Backward Compatibility**: Existing code continues to work, but with better capabilities

## 🔧 Migration Path

```go
// Before: Basic reflection
userSchema := builders.NewObject().
    Property("name", builders.NewStringSchema().Build()).
    Property("email", builders.NewStringSchema().Email().Build()).
    Build()

// After: Enhanced native integration (same API, better implementation)  
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" format:"email" validate:"required"`
}

// Option 1: Through existing builder (enhanced internally)
userSchema := builders.NewObject().FromValue[User]()

// Option 2: Through factory function
userSchema := builders.NewObjectFromStruct[User]()

// Option 3: Through engine (for dynamic use)
userSchema := engine.CreateSchema("native_struct", reflect.TypeOf(User{}))
``` 