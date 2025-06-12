# Schema Native: Go Struct to Schema Conversion

**Version:** Draft 1  
**Date:** 2024  
**Status:** 🚧 Draft Proposal  

## 📋 Overview

This document proposes the implementation of a `schema/native` subpackage that provides convenient conversion from Go types to schema objects, enabling developers to generate schemas directly from Go structs with support for struct tags and type inference.

## 🎯 Goals

### Primary Goals
1. **Type-Safe Schema Generation**: Provide `FromValue[T]()` for compile-time type safety
2. **Struct Tag Support**: Parse and respect Go struct tags for schema customization
3. **Zero Configuration**: Work out-of-the-box with minimal setup
4. **Full Integration**: Seamlessly integrate with existing schema system
5. **Performance**: Efficient reflection-based conversion with caching

### Secondary Goals
1. **Extensible**: Allow custom tag processors and type converters
2. **Comprehensive**: Support all existing schema types and constraints
3. **Developer Experience**: Provide clear error messages and documentation

## 🏗️ Architecture Analysis

### Current Schema System Understanding

Based on thorough analysis of the existing codebase:

1. **Interface-First Design**: All schemas implement `core.Schema` interface
2. **Builder Pattern**: Fluent API using `*Builder` types that implement `core.*Builder` interfaces  
3. **Immutable Schemas**: Schemas are immutable after creation
4. **Visitor Pattern**: Support for schema traversal via `core.SchemaVisitor`
5. **Type System**: Comprehensive type system with validation, JSON schema export, example generation
6. **Reflection Exists**: Basic reflection exists in `builders/service.go:createSchemaFromType()`

### Existing Patterns
- `ServiceBuilder.FromValue(instance)` - Service-specific struct processing
- `createSchemaFromType(reflect.Type)` - Basic type-to-schema conversion
- JSON tag parsing in `schemas/object.go` - Basic JSON tag support
- Format support in `builders/string.go` - Format constraints like `Email()`

## 🚀 Proposed Implementation

### Package Structure

```
schema/native/
├── doc.go                    # Package documentation
├── converter.go              # Core conversion logic
├── tags.go                   # Struct tag parsing
├── types.go                  # Type mapping registry
├── cache.go                  # Reflection result caching
├── visitors.go               # Schema visitors for analysis
├── options.go                # Configuration options
├── examples/
│   ├── basic.go              # Basic usage examples
│   ├── advanced.go           # Advanced tag usage
│   └── custom.go             # Custom converters
└── tests/
    ├── converter_test.go     # Core conversion tests
    ├── tags_test.go          # Tag parsing tests
    ├── types_test.go         # Type mapping tests
    ├── cache_test.go         # Caching tests
    └── integration_test.go   # End-to-end tests
```

### Core API Design

```go
package native

import (
    "defs.dev/schema/api/core"
)

// FromValue generates a schema from a Go type with compile-time type safety
func FromValue[T any]() core.Schema {
    return FromValueWithOptions[T](DefaultOptions())
}

// FromValueWithOptions provides configurable schema generation
func FromValueWithOptions[T any](opts *Options) core.Schema {
    var zero T
    return convertValue(reflect.TypeOf(zero), opts)
}

// FromInstance generates a schema from a Go instance (runtime)
func FromInstance(instance any) core.Schema {
    return FromInstanceWithOptions(instance, DefaultOptions())
}

// FromInstanceWithOptions provides configurable schema generation from instance
func FromInstanceWithOptions(instance any, opts *Options) core.Schema {
    return convertValue(reflect.TypeOf(instance), opts)
}
```

### Supported Struct Tags

```go
type User struct {
    // String constraints
    Name     string `json:"name" validate:"required" description:"User's full name"`
    Email    string `json:"email" format:"email" validate:"required" description:"Email address"`
    Username string `json:"username" minLength:"3" maxLength:"50" pattern:"^[a-zA-Z0-9_]+$"`
    
    // Number constraints  
    Age      int     `json:"age" min:"0" max:"150" description:"User's age"`
    Height   float64 `json:"height" min:"0.0" max:"3.0" description:"Height in meters"`
    
    // Array constraints
    Tags     []string `json:"tags" minItems:"0" maxItems:"10" description:"User tags"`
    
    // Object constraints
    Profile  Profile  `json:"profile" description:"User profile information"`
    
    // Optional fields
    Bio      *string  `json:"bio,omitempty" maxLength:"500" description:"User biography"`
    
    // Ignored fields
    Internal string   `json:"-"`
    
    // Custom formats
    Phone    string   `json:"phone" format:"phone" description:"Phone number"`
    
    // Enums
    Role     string   `json:"role" enum:"admin,user,guest" description:"User role"`
    
    // Default values
    Active   bool     `json:"active" default:"true" description:"Whether user is active"`
}

type Profile struct {
    Avatar    string    `json:"avatar" format:"url" description:"Avatar image URL"`
    CreatedAt time.Time `json:"created_at" format:"date-time" description:"Profile creation date"`
}
```

### Tag Processing System

```go
// TagProcessor defines interface for processing struct tags
type TagProcessor interface {
    ProcessTag(field reflect.StructField, builder any) error
    SupportedTags() []string
}

// Built-in processors
var defaultProcessors = []TagProcessor{
    &JSONTagProcessor{},      // json:"name,omitempty"
    &ValidateTagProcessor{},  // validate:"required,email"
    &ConstraintTagProcessor{}, // minLength:"3" maxLength:"50"
    &FormatTagProcessor{},    // format:"email"
    &DescriptionTagProcessor{}, // description:"Field description"
    &EnumTagProcessor{},      // enum:"value1,value2,value3"
    &DefaultTagProcessor{},   // default:"defaultValue"
}
```

### Type Mapping System

```go
// TypeConverter defines interface for custom type conversion
type TypeConverter interface {
    CanConvert(t reflect.Type) bool
    Convert(t reflect.Type, opts *Options) core.Schema
}

// Built-in converters with priority order
var defaultConverters = []TypeConverter{
    &TimeConverter{},         // time.Time -> string with date-time format
    &UUIDConverter{},         // uuid.UUID -> string with uuid format
    &URLConverter{},          // url.URL -> string with url format
    &DurationConverter{},     // time.Duration -> string
    &BigIntConverter{},       // big.Int -> string
    &BigFloatConverter{},     // big.Float -> string
    &EnumConverter{},         // Custom enum types
    &BasicTypeConverter{},    // Built-in Go types
}
```

### Options System

```go
type Options struct {
    // Tag processing
    TagProcessors    []TagProcessor
    CustomTags       map[string]TagProcessor
    
    // Type conversion
    TypeConverters   []TypeConverter
    CustomConverters map[reflect.Type]TypeConverter
    
    // Behavior options
    RequireJSONTags  bool              // Only process fields with json tags
    IgnoreUnexported bool              // Skip unexported fields (default: true)
    AllowRecursion   bool              // Allow recursive struct references
    MaxDepth         int               // Maximum recursion depth (default: 10)
    DefaultRequired  bool              // Make all fields required by default
    
    // Naming options
    NamingStrategy   NamingStrategy    // Field naming strategy
    
    // Validation options
    StrictValidation bool              // Strict tag validation
    FailOnError      bool              // Fail conversion on any error
    
    // Caching
    EnableCache      bool              // Enable reflection result caching
    CacheSize        int               // Maximum cache entries
}

type NamingStrategy interface {
    FieldName(field reflect.StructField) string
}
```

### Caching System

```go
type SchemaCache struct {
    cache map[reflect.Type]core.Schema
    mutex sync.RWMutex
    size  int
    maxSize int
}

func (c *SchemaCache) Get(t reflect.Type) (core.Schema, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    schema, exists := c.cache[t]
    return schema, exists
}

func (c *SchemaCache) Set(t reflect.Type, schema core.Schema) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if c.size >= c.maxSize {
        c.evictOldest()
    }
    
    c.cache[t] = schema
    c.size++
}
```

## 📚 Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "defs.dev/schema/native"
)

type User struct {
    Name  string `json:"name" validate:"required" description:"User's name"`
    Email string `json:"email" format:"email" validate:"required"`
    Age   int    `json:"age" min:"0" max:"150"`
}

func main() {
    // Generate schema from type
    userSchema := native.FromValue[User]()
    
    // Use the schema
    result := userSchema.Validate(map[string]any{
        "name":  "John Doe",
        "email": "john@example.com", 
        "age":   30,
    })
    
    if !result.Valid {
        fmt.Printf("Validation errors: %v\n", result.Errors)
    }
}
```

### Advanced Usage

```go
// Custom options
opts := &native.Options{
    RequireJSONTags:  true,
    DefaultRequired:  false,
    StrictValidation: true,
    EnableCache:      true,
}

userSchema := native.FromValueWithOptions[User](opts)

// Custom type converter
type PhoneNumber string

type PhoneConverter struct{}

func (p *PhoneConverter) CanConvert(t reflect.Type) bool {
    return t == reflect.TypeOf(PhoneNumber(""))
}

func (p *PhoneConverter) Convert(t reflect.Type, opts *native.Options) core.Schema {
    return builders.NewStringSchema().
        Format("phone").
        Pattern(`^\+?[1-9]\d{1,14}$`).
        Description("Phone number in E.164 format").
        Build()
}

// Register custom converter
opts.CustomConverters[reflect.TypeOf(PhoneNumber(""))] = &PhoneConverter{}
```

### Integration with Existing Builders

```go
// Mix native conversion with manual building
userSchema := native.FromValue[User]()

// Create a service using the generated schema
serviceSchema := builders.NewServiceSchema().
    Name("UserService").
    Method("CreateUser", builders.NewFunctionSchema().
        Input("user", userSchema).
        Output("result", builders.NewObject().
            Property("id", builders.NewStringSchema().UUID().Build()).
            Property("created", userSchema).
            Build()).
        Build()).
    Build()
```

## 🔧 Implementation Details

### Core Conversion Algorithm

```go
func convertValue(t reflect.Type, opts *Options) core.Schema {
    // Check cache first
    if opts.EnableCache {
        if cached, exists := globalCache.Get(t); exists {
            return cached
        }
    }
    
    // Handle pointers and interfaces
    t = resolveType(t)
    
    // Try custom converters first
    for _, converter := range opts.TypeConverters {
        if converter.CanConvert(t) {
            schema := converter.Convert(t, opts)
            if opts.EnableCache {
                globalCache.Set(t, schema)
            }
            return schema
        }
    }
    
    // Handle built-in types
    switch t.Kind() {
    case reflect.Struct:
        return convertStruct(t, opts)
    case reflect.Slice, reflect.Array:
        return convertArray(t, opts)
    case reflect.Map:
        return convertMap(t, opts)
    case reflect.String:
        return convertString(t, opts)
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return convertInteger(t, opts)
    case reflect.Float32, reflect.Float64:
        return convertNumber(t, opts)
    case reflect.Bool:
        return convertBoolean(t, opts)
    default:
        return builders.NewObject().AdditionalProperties(true).Build()
    }
}
```

### Struct Conversion with Tag Processing

```go
func convertStruct(t reflect.Type, opts *Options) core.Schema {
    builder := builders.NewObject()
    var required []string
    
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        
        // Skip unexported fields if configured
        if opts.IgnoreUnexported && !field.IsExported() {
            continue
        }
        
        // Process field tags
        fieldInfo := processFieldTags(field, opts)
        if fieldInfo.Skip {
            continue
        }
        
        // Convert field type to schema
        fieldSchema := convertValue(field.Type, opts)
        
        // Apply tag constraints to schema
        fieldSchema = applyConstraints(fieldSchema, fieldInfo.Constraints)
        
        // Add to object
        builder.Property(fieldInfo.Name, fieldSchema)
        
        if fieldInfo.Required {
            required = append(required, fieldInfo.Name)
        }
    }
    
    if len(required) > 0 {
        builder.Required(required...)
    }
    
    return builder.Build()
}
```

### Tag Processing Pipeline

```go
type FieldInfo struct {
    Name        string
    Required    bool
    Skip        bool
    Constraints map[string]string
    Description string
    Examples    []any
}

func processFieldTags(field reflect.StructField, opts *Options) FieldInfo {
    info := FieldInfo{
        Name:        field.Name,
        Required:    opts.DefaultRequired,
        Constraints: make(map[string]string),
    }
    
    // Process each tag processor
    for _, processor := range opts.TagProcessors {
        if err := processor.ProcessTag(field, &info); err != nil {
            if opts.FailOnError {
                panic(fmt.Sprintf("Tag processing error for field %s: %v", field.Name, err))
            }
            // Log error and continue
        }
    }
    
    return info
}
```

## 🧪 Testing Strategy

### Test Categories

1. **Unit Tests**
   - Tag parsing for each supported tag type
   - Type conversion for all Go built-in types
   - Custom converter registration and usage
   - Caching behavior and performance
   - Error handling and edge cases

2. **Integration Tests**
   - End-to-end struct conversion
   - Complex nested structures
   - Recursive type handling
   - Performance benchmarks
   - Memory usage profiling

3. **Compatibility Tests**
   - Integration with existing schema builders
   - JSON Schema output validation
   - Visitor pattern support
   - Clone and metadata preservation

### Performance Requirements

- **Conversion Speed**: < 1ms for simple structs, < 10ms for complex nested structures
- **Memory Usage**: Minimal allocations, efficient caching
- **Cache Hit Rate**: > 90% for repeated conversions of same types
- **Concurrent Safety**: Thread-safe operations with minimal lock contention

## 🚀 Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Basic package structure and documentation
- [ ] Core type conversion algorithm
- [ ] Basic tag processors (json, validate, description)
- [ ] Simple test suite

### Phase 2: Tag System (Week 3-4)
- [ ] Complete tag processor system
- [ ] All constraint tag processors (min, max, pattern, format, etc.)
- [ ] Custom tag processor registration
- [ ] Comprehensive tag parsing tests

### Phase 3: Advanced Features (Week 5-6)
- [ ] Custom type converter system
- [ ] Caching implementation
- [ ] Options and configuration system
- [ ] Performance optimization

### Phase 4: Integration & Polish (Week 7-8)
- [ ] Integration with existing schema system
- [ ] Comprehensive documentation and examples
- [ ] Performance benchmarks
- [ ] Production readiness

## 🔄 Migration and Compatibility

### Backward Compatibility
- Fully compatible with existing schema system
- No changes to existing APIs
- Uses same core interfaces and types

### Migration Path
```go
// Before: Manual schema creation
userSchema := builders.NewObject().
    Property("name", builders.NewStringSchema().Build()).
    Property("email", builders.NewStringSchema().Email().Build()).
    Required("name", "email").
    Build()

// After: Native generation
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" format:"email" validate:"required"`
}

userSchema := native.FromValue[User]()
```

## 📊 Success Metrics

### Functional Requirements
- [ ] **Type Safety**: 100% compile-time type safety for FromValue[T]()
- [ ] **Tag Coverage**: Support for all major struct tag conventions
- [ ] **Schema Compatibility**: Perfect integration with existing schema types
- [ ] **Error Handling**: Clear, actionable error messages

### Performance Requirements
- [ ] **Speed**: 5x faster than manual schema building for equivalent schemas
- [ ] **Memory**: 50% less memory usage through caching
- [ ] **Scalability**: Linear performance scaling with struct complexity

### Developer Experience
- [ ] **Zero Config**: Works out-of-the-box with sensible defaults
- [ ] **Documentation**: Complete examples and API documentation
- [ ] **Testing**: 95%+ test coverage with comprehensive edge case handling

## 🎯 Future Enhancements (Post-MVP)

### Advanced Features
- **Code Generation**: Generate Go structs from schemas
- **Schema Evolution**: Detect and handle schema changes
- **Validation Rules**: Custom validation rule DSL
- **Plugin System**: Extensible plugin architecture

### Integrations
- **gRPC/Protobuf**: Direct protobuf schema generation
- **OpenAPI**: Enhanced OpenAPI specification generation
- **Database**: ORM integration for database schema generation
- **GraphQL**: GraphQL schema generation from Go types

---

**Status**: 🚧 Draft Proposal - Ready for Review and Implementation  
**Next Steps**: Review, refine, and begin Phase 1 implementation 