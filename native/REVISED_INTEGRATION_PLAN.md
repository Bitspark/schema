# Schema Native: Revised Integration Plan

## 📋 Executive Summary

Based on analysis of the existing schema implementation, the `schema/native` package should be redesigned for much deeper integration with the current architecture rather than being a standalone add-on package.

## 🔄 Key Changes from Original DRAFT_1.md

### Original Approach Issues
- **Standalone Package**: Created separate `native` package with its own APIs
- **Duplicate Reflection**: Reimplemented reflection logic already present in `service.go`
- **Separate Caching**: Proposed independent caching system
- **API Proliferation**: Added new APIs (`FromValue[T]()`) instead of enhancing existing ones

### Enhanced Integration Approach
- **Deep Integration**: Replace and enhance existing reflection capabilities
- **Unified API**: Extend existing builder interfaces rather than create new ones
- **Shared Infrastructure**: Leverage engine's extension system and caching
- **Backward Compatibility**: Enhance existing functionality without breaking changes

## 🏗️ Revised Architecture

### 1. Integration Points Rather Than Standalone Package

```go
// BEFORE (DRAFT_1 approach): Separate native package
import "defs.dev/schema/native"
userSchema := native.FromValue[User]()

// AFTER (Integrated approach): Enhance existing builders
import "defs.dev/schema/builders"
userSchema := builders.NewObject().FromValue[User]()
```

### 2. Replace Existing Reflection Code

Instead of new implementation, enhance existing `createSchemaFromType()`:

```go
// Location: schema/builders/reflection.go (new file extracting from service.go)

// Enhanced version of existing createSchemaFromType
func createSchemaFromType(t reflect.Type, opts ...*ReflectionOptions) core.Schema {
    options := mergeReflectionOptions(opts...)
    return createSchemaFromTypeWithOptions(t, options)
}

// New comprehensive implementation
func createSchemaFromTypeWithOptions(t reflect.Type, opts *ReflectionOptions) core.Schema {
    // Check cache first
    if cached := opts.Cache.Get(t, opts); cached != nil {
        return cached
    }
    
    // Handle custom converters
    if converter := opts.GetConverter(t); converter != nil {
        return converter.Convert(t, opts)
    }
    
    // Enhanced type mapping with tag support
    return buildSchemaForType(t, opts)
}
```

### 3. Extend Core Builder Interfaces

Add native methods to existing `api/core/builder.go` interfaces:

```go
// Enhanced ObjectSchemaBuilder
type ObjectSchemaBuilder interface {
    Builder[ObjectSchema]
    MetadataBuilder[ObjectSchemaBuilder]
    
    // Existing methods
    Property(name string, schema Schema) ObjectSchemaBuilder
    Required(names ...string) ObjectSchemaBuilder
    AdditionalProperties(allowed bool) ObjectSchemaBuilder
    
    // NEW: Native integration methods
    FromStruct(structType reflect.Type) ObjectSchemaBuilder
    FromValue[T any]() ObjectSchemaBuilder
    WithTagOptions(opts *TagOptions) ObjectSchemaBuilder
}

// Enhanced ServiceSchemaBuilder  
type ServiceSchemaBuilder interface {
    Builder[ServiceSchema]
    MetadataBuilder[ServiceSchemaBuilder]
    
    // Existing methods
    Method(name string, functionSchema FunctionSchema) ServiceSchemaBuilder
    FromStruct(instance any) ServiceSchemaBuilder
    
    // NEW: Enhanced native methods
    WithNativeOptions(opts *NativeOptions) ServiceSchemaBuilder
    WithTagProcessors(processors ...TagProcessor) ServiceSchemaBuilder
}
```

### 4. Engine Integration as First-Class Extension

Register native capabilities through the engine's extension system:

```go
// In engine initialization
func InitializeNativeExtensions(engine SchemaEngine) error {
    // Register built-in type factories
    typeFactories := map[string]SchemaTypeFactory{
        "struct":     &StructTypeFactory{},
        "time":       &TimeTypeFactory{},
        "uuid":       &UUIDTypeFactory{},
        "duration":   &DurationTypeFactory{},
        "url":        &URLTypeFactory{},
        "bigint":     &BigIntTypeFactory{},
    }
    
    for name, factory := range typeFactories {
        if err := engine.RegisterSchemaType(name, factory); err != nil {
            return fmt.Errorf("failed to register type %s: %w", name, err)
        }
    }
    
    return nil
}
```

### 5. Unified Configuration

Integrate native options with engine configuration:

```go
// Enhanced EngineConfig in engine/engine.go
type EngineConfig struct {
    // Existing fields
    EnableCache        bool `json:"enable_cache"`
    MaxCacheSize       int  `json:"max_cache_size"`
    
    // Native integration configuration
    Native NativeConfig `json:"native,omitempty"`
}

type NativeConfig struct {
    EnableStructConversion bool              `json:"enable_struct_conversion"`
    DefaultTagProcessors   []string          `json:"default_tag_processors"`
    CustomConverters       map[string]string `json:"custom_converters"`
    RequireJSONTags        bool              `json:"require_json_tags"`
    StrictTagValidation    bool              `json:"strict_tag_validation"`
}
```

## 🚀 Revised Implementation Plan

### Phase 1: Core Infrastructure Enhancement (Week 1-2)

**Files to Modify:**
- `schema/builders/reflection.go` (new, extract from service.go)
- `schema/api/core/builder.go` (extend interfaces)
- `schema/builders/object.go` (add native methods)
- `schema/builders/service.go` (enhance FromStruct)

**Tasks:**
- [ ] Extract and enhance `createSchemaFromType` function
- [ ] Add `FromValue[T]()` methods to existing builders
- [ ] Implement basic tag processing system
- [ ] Update ServiceBuilder.FromStruct to use enhanced reflection

### Phase 2: Tag Processing Integration (Week 3-4)

**Files to Create:**
- `schema/internal/tags/processor.go`
- `schema/internal/tags/json.go`
- `schema/internal/tags/validation.go`
- `schema/internal/tags/constraints.go`

**Tasks:**
- [ ] Implement tag processor interface and implementations
- [ ] Integrate tag processing with existing builders
- [ ] Add constraint validation support
- [ ] Create comprehensive tag processor tests

### Phase 3: Engine Integration (Week 5-6)

**Files to Modify:**
- `schema/engine/extensions.go` (add native type factories)
- `schema/engine/config.go` (add native configuration)
- `schema/engine/cache.go` (integrate native caching)

**Tasks:**
- [ ] Register native converters as schema type extensions
- [ ] Integrate with engine configuration system
- [ ] Implement shared caching for native types
- [ ] Add engine-level native type registration

### Phase 4: Advanced Features & Optimization (Week 7-8)

**Files to Create:**
- `schema/internal/native/converters.go`
- `schema/internal/native/metadata.go`
- `schema/internal/native/visitors.go`

**Tasks:**
- [ ] Implement advanced type converters (time.Time, UUID, etc.)
- [ ] Add native-specific metadata and visitor support
- [ ] Performance optimization and memory usage improvements
- [ ] Comprehensive integration testing

## 📁 Revised File Structure

```
schema/
├── api/core/
│   ├── builder.go              # Extended with native methods
│   └── ...
├── builders/
│   ├── object.go               # Enhanced with FromValue[T]()
│   ├── service.go              # Enhanced FromStruct implementation
│   ├── reflection.go           # NEW: Enhanced type reflection
│   └── ...
├── engine/
│   ├── extensions.go           # Native type factory registration
│   ├── config.go               # Native configuration integration
│   └── ...
├── internal/
│   ├── tags/                   # Tag processing system
│   │   ├── processor.go
│   │   ├── json.go
│   │   ├── validation.go
│   │   └── constraints.go
│   └── native/                 # Native-specific internals
│       ├── converters.go
│       ├── metadata.go
│       └── visitors.go
└── schemas/
    ├── object.go               # Enhanced with native metadata
    └── ...
```

## 🎯 API Design Comparison

### Original DRAFT_1 API
```go
// Separate package with new APIs
import "defs.dev/schema/native"

userSchema := native.FromValue[User]()
serviceSchema := native.FromInstance(userService)
```

### Revised Integrated API
```go
// Enhanced existing APIs
import "defs.dev/schema/builders"

// Object schemas
userSchema := builders.NewObject().FromValue[User]()

// Service schemas  
serviceSchema := builders.NewServiceSchema().FromStruct(userService)

// Engine-based (for dynamic use)
userSchema := engine.CreateSchema("struct", reflect.TypeOf(User{}))
```

## 🔄 Migration Benefits

### For Existing Code
- **Zero Breaking Changes**: All existing code continues to work
- **Enhanced Capabilities**: Existing `FromStruct` gets tag support automatically  
- **Better Performance**: Shared caching and optimized reflection
- **Richer Metadata**: Native field information available throughout system

### For New Code
- **Familiar APIs**: Use existing builder patterns with enhanced capabilities
- **Type Safety**: `FromValue[T]()` provides compile-time type safety
- **Consistent Experience**: Native features integrated into existing workflows
- **Engine Integration**: Native types work with all engine features

## 📊 Success Metrics (Revised)

### Integration Success
- [ ] Zero breaking changes to existing APIs
- [ ] 100% backward compatibility maintained
- [ ] Enhanced `FromStruct` performance (5x faster)
- [ ] Unified caching reduces memory usage by 50%

### Feature Completeness
- [ ] All tag types from DRAFT_1 supported
- [ ] Engine integration provides dynamic type creation
- [ ] Visitor pattern works with native schemas
- [ ] Metadata system includes native field information

### Developer Experience
- [ ] Existing builder APIs enhanced, not replaced
- [ ] Documentation shows migration path for each pattern
- [ ] Integration tests verify all existing functionality
- [ ] Performance benchmarks show improvements

---

This revised approach transforms the native package from a separate add-on into a deep enhancement of the existing schema system, providing better integration, performance, and developer experience while maintaining full backward compatibility. 