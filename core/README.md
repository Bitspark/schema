# Schema Core - Clean API-First Implementation

This directory contains a complete re-implementation of the schema system using the clean interfaces defined in `schema/api`. This provides a fresh start with better organization, cleaner code, and API-first design.

## 🎯 Goals

1. **API-First Design**: All implementations use `schema/api` interfaces from the start
2. **Clean Organization**: Logical separation of concerns with clear module boundaries
3. **Zero Legacy Debt**: No backward compatibility constraints - pure, clean implementation
4. **Better Testing**: Each component is independently testable
5. **Performance**: Optimized implementations without legacy cruft
6. **Extensibility**: Easy to add new schema types and behaviors

## 📁 Proposed Directory Structure

```
schema/core/
├── README.md                    # This file
├── doc.go                       # Package documentation
├── 
├── schemas/                     # Core schema implementations
│   ├── string.go                # StringSchema (API-first)
│   ├── number.go                # NumberSchema (API-first)
│   ├── integer.go               # IntegerSchema (API-first)
│   ├── boolean.go               # BooleanSchema (API-first)
│   ├── array.go                 # ArraySchema (API-first)
│   ├── object.go                # ObjectSchema (API-first)
│   ├── union.go                 # UnionSchema (API-first)
│   └── function.go              # FunctionSchema (API-first)
│
├── builders/                    # Fluent builder implementations
│   ├── string.go                # StringBuilder
│   ├── number.go                # NumberBuilder
│   ├── integer.go               # IntegerBuilder
│   ├── boolean.go               # BooleanBuilder
│   ├── array.go                 # ArrayBuilder
│   ├── object.go                # ObjectBuilder
│   ├── union.go                 # UnionBuilder
│   ├── function.go              # FunctionBuilder
│   └── factory.go               # Factory functions (NewString, NewObject, etc.)
│
├── visitors/                    # Visitor pattern implementations
│   ├── base.go                  # BaseVisitor with default implementations
│   ├── traversal.go             # Tree traversal visitor
│   ├── collector.go             # Schema collection visitors
│   ├── validator.go             # Validation visitors
│   └── transformer.go           # Schema transformation visitors
│
├── generics/                    # Generic schema patterns
│   ├── list.go                  # List[T] implementation
│   ├── optional.go              # Optional[T] implementation
│   ├── result.go                # Result[T, E] implementation
│   ├── map.go                   # Map[K, V] implementation
│   └── union.go                 # Union[T1, T2, ...] implementations
│
├── validation/                  # Enhanced validation system
│   ├── formats.go               # String format validators (email, UUID, etc.)
│   ├── rules.go                 # Custom validation rules
│   ├── context.go               # Validation context and path tracking
│   └── errors.go                # Enhanced error reporting
│
├── examples/                    # Usage examples and documentation
│   ├── basic.go                 # Basic schema creation examples
│   ├── advanced.go              # Advanced patterns and composition
│   ├── validation.go            # Validation examples
│   └── visitors.go              # Visitor pattern examples
│
└── tests/                       # Comprehensive test suite
    ├── schemas_test.go          # Schema implementation tests
    ├── builders_test.go         # Builder tests
    ├── visitors_test.go         # Visitor tests
    ├── generics_test.go         # Generic pattern tests
    └── integration_test.go      # End-to-end integration tests
```

## 🏗️ Implementation Strategy

### Phase 1: Core Schema Types
1. Implement basic schema types (`string`, `number`, `integer`, `boolean`)
2. Each uses `api.Schema` interface natively
3. Clean, focused implementations with comprehensive tests

### Phase 2: Complex Schema Types  
1. Implement `array`, `object`, `union` schemas
2. Add visitor pattern support
3. Ensure proper composition and introspection

### Phase 3: Builders and Factory Functions
1. Implement fluent builders using `api.*Builder` interfaces
2. Create factory functions (`NewString()`, `NewObject()`, etc.)
3. Ensure builders return `api.Schema` types

### Phase 4: Generic Patterns
1. Implement type-safe generic schemas (`List[T]`, `Optional[T]`, etc.)
2. Use Go 1.24 generics for true type safety
3. Provide ergonomic APIs for common patterns

### Phase 5: Enhanced Features
1. Advanced validation system with custom rules
2. Enhanced error reporting with suggestions
3. Performance optimizations

## 🔄 Migration Strategy

### Backward Compatibility
- The existing `schema/` package remains unchanged
- Users can migrate gradually from `schema` to `schema/core`
- Both packages can coexist during transition

### Migration Path
```go
// Old way (still works)
import "defs.dev/schema"
stringSchema := schema.NewString().MinLength(5).Build()

// New way (cleaner, API-first)
import "defs.dev/schema/core"
stringSchema := core.NewString().MinLength(5).Build() // returns api.StringSchema
```

### Interoperability
- Both implementations use the same `api.*` types
- Schemas from both packages work with the same visitors
- Validation results are compatible

## 🎨 Design Principles

### 1. Interface Segregation
```go
// Each schema type has focused interfaces
type StringSchema interface {
    api.Schema
    api.Accepter
    MinLength() *int
    MaxLength() *int
    Pattern() string
    // ... other string-specific methods
}
```

### 2. Composition Over Inheritance
```go
// Schemas compose behaviors rather than inherit
type ObjectSchema struct {
    *BaseSchema                    // Common functionality
    properties map[string]api.Schema
    required   []string
}
```

### 3. Immutability
```go
// All schemas are immutable - methods return new instances
func (s *StringSchema) WithMinLength(min int) api.StringSchema {
    clone := s.clone()
    clone.minLength = &min
    return clone
}
```

### 4. Type Safety
```go
// Generic patterns provide compile-time type safety
func List[T any]() api.ListBuilder[T] {
    return &listBuilder[T]{
        itemSchema: FromStruct[T](),
    }
}
```

## 🚀 Key Improvements

### 1. **Performance**
- Optimized validation algorithms
- Efficient memory usage patterns
- Minimal allocations in hot paths

### 2. **Usability**
- Better error messages with suggestions
- Intuitive API design
- Comprehensive documentation and examples

### 3. **Extensibility**
- Plugin system for custom validators
- Easy addition of new schema types
- Composable validation rules

### 4. **Testing**
- 100% test coverage
- Property-based testing for validation
- Performance benchmarks

## 📖 Usage Examples

### Basic Schema Creation
```go
import "defs.dev/schema/core"

// String schema with validation
userNameSchema := core.NewString().
    MinLength(3).
    MaxLength(50).
    Pattern(`^[a-zA-Z0-9_]+$`).
    Description("Username for the system").
    Build()

// Validate a value
result := userNameSchema.Validate("john_doe")
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Error: %s\n", err.Message)
    }
}
```

### Object Schema Composition
```go
userSchema := core.NewObject().
    Property("name", core.NewString().MinLength(1).Build()).
    Property("email", core.NewString().Email().Build()).
    Property("age", core.NewInteger().Min(0).Max(150).Build()).
    Required("name", "email").
    Build()
```

### Generic Type Safety
```go
// Type-safe list of users
userListSchema := core.List[User]().
    MinItems(1).
    MaxItems(100).
    Build()

// Type-safe optional field
optionalEmailSchema := core.Optional[string]().
    Description("Optional email address").
    Build()
```

### Visitor Pattern
```go
// Collect all string schemas in a complex schema
stringCollector := core.NewStringCollector()
userSchema.Accept(stringCollector)

for _, stringSchema := range stringCollector.Strings {
    fmt.Printf("Found string schema: %s\n", stringSchema.Metadata().Name)
}
```

## 🔧 Implementation Notes

### Error Handling
- Rich error context with path information
- Actionable error messages with suggestions
- Structured error data for programmatic handling

### Validation Performance
- Early validation termination on first error (optional)
- Optimized validation for common patterns
- Caching of compiled regular expressions

### Memory Efficiency
- Copy-on-write for large schemas
- Shared immutable metadata structures
- Efficient internal representations

## 🎯 Success Metrics

1. **API Compliance**: 100% compatibility with `schema/api` interfaces
2. **Performance**: 20%+ faster than legacy implementation
3. **Test Coverage**: 95%+ code coverage with comprehensive test suite
4. **Documentation**: Complete documentation with examples for all features
5. **Adoption**: Easy migration path from legacy `schema` package

## 🚦 Next Steps

1. **Create Core Package Structure**: Set up the directory structure and basic files
2. **Implement String Schema**: Start with a clean StringSchema implementation
3. **Add Builder Pattern**: Implement fluent builders for schema creation
4. **Test and Validate**: Comprehensive testing to ensure correctness
5. **Document and Example**: Create documentation and usage examples
6. **Gradual Migration**: Help users migrate from legacy to core implementation

This proposal provides a path to a cleaner, more maintainable, and more powerful schema system while maintaining backward compatibility with existing code. 