# Schema Concepts

This document provides a comprehensive guide to the core concepts and architectural principles of the Schema library.

## Table of Contents

1. [Core Schema Interface](#core-schema-interface)
2. [Schema Types](#schema-types)
3. [Function Schemas](#function-schemas)
4. [Builder Pattern](#builder-pattern)
5. [Validation System](#validation-system)
6. [Schema Registry](#schema-registry)
7. [Reflection and Struct Generation](#reflection-and-struct-generation)
8. [Metadata and Documentation](#metadata-and-documentation)
9. [Example Generation](#example-generation)
10. [Design Philosophy](#design-philosophy)

## Core Schema Interface

At the heart of the library is the `Schema` interface, which defines the contract for all schema types:

```go
type Schema interface {
    // Validation
    Validate(value any) ValidationResult

    // JSON Schema generation
    ToJSONSchema() map[string]any

    // Metadata
    Type() SchemaType
    Metadata() SchemaMetadata
    WithMetadata(metadata SchemaMetadata) Schema

    // Example generation
    GenerateExample() any

    // Utilities
    Clone() Schema
}
```

This interface ensures that all schemas provide consistent functionality for validation, serialization, metadata management, and example generation.

## Schema Types

### Primitive Types

The library supports all JSON primitive types with rich validation options:

#### String Schema
```go
schema.String().
    MinLength(1).
    MaxLength(100).
    Pattern("^[a-zA-Z]+$").
    Email().                    // Built-in format validation
    Enum("active", "inactive").
    Default("active").
    Build()
```

#### Numeric Schemas
```go
// Integer with constraints
schema.Integer().
    Min(0).
    Max(1000).
    MultipleOf(5).
    Build()

// Floating-point numbers
schema.Number().
    Min(0.0).
    Max(100.0).
    ExclusiveMax(true).
    Build()
```

#### Boolean Schema
```go
schema.Boolean().
    Default(false).
    Build()
```

### Collection Types

#### Object Schema
Objects represent structured data with properties:

```go
userSchema := schema.Object().
    Property("id", schema.Integer().Min(1).Build()).
    Property("name", schema.String().MinLength(1).Build()).
    Property("email", schema.String().Email().Build()).
    Property("profile", profileSchema).
    Required("id", "name", "email").
    AdditionalProperties(false).
    Build()
```

**Object Features:**
- **Properties**: Named fields with their own schemas
- **Required Fields**: Specify which properties must be present  
- **Additional Properties**: Control whether extra properties are allowed
- **Nested Objects**: Properties can be other object schemas
- **Introspection**: Access properties, required fields, and settings

#### Array Schema
Arrays represent ordered collections of items:

```go
numbersSchema := schema.Array().
    Items(schema.Integer().Min(0).Build()).
    MinItems(1).
    MaxItems(100).
    UniqueItems(true).
    Build()

// Tuple-like arrays with different item types
mixedSchema := schema.Array().
    Items(schema.Union().
        Option(schema.String().Build()).
        Option(schema.Integer().Build()).
        Build()).
    Build()
```

### Advanced Types

#### Union Schema
Represents a value that can be one of several types:

```go
stringOrNumber := schema.Union().
    Option(schema.String().Build()).
    Option(schema.Integer().Build()).
    Build()
```

#### Optional Schema
Wraps another schema to make it optional:

```go
optionalEmail := schema.Optional(schema.String().Email().Build())
```

#### Map Schema
For key-value mappings with homogeneous value types:

```go
stringMap := schema.Map(schema.String().Build())
```

## Function Schemas

A unique feature of this library is treating function signatures as first-class schema types. This enables powerful API contract definitions and validation.

### Basic Function Schema

```go
calculateTax := schema.NewFunctionSchema().
    Input("amount", schema.Number().Min(0).Build()).
    Input("rate", schema.Number().Min(0).Max(1).Build()).
    Input("jurisdiction", schema.String().Build()).
    Output(schema.Object().
        Property("tax", schema.Number().Build()).
        Property("total", schema.Number().Build()).
        Required("tax", "total").
        Build()).
    Error(schema.Object().
        Property("code", schema.String().Build()).
        Property("message", schema.String().Build()).
        Build()).
    Description("Calculate tax for a given amount and rate").
    Build()
```

### Function Schema Features

- **Input Parameters**: Define parameter names, types, and validation rules
- **Output Schema**: Specify the structure of successful responses
- **Error Schema**: Define the structure of error responses
- **Required Parameters**: Mark which inputs are required
- **Metadata**: Add descriptions, names, and other documentation

### Function Schema Use Cases

1. **API Contract Definition**: Define REST endpoint schemas
2. **Service Interface Validation**: Validate service method calls
3. **Code Generation**: Generate client SDKs from function schemas
4. **Documentation**: Auto-generate API documentation
5. **Testing**: Generate test data for function inputs/outputs

## Builder Pattern

All schemas use a fluent builder pattern that provides:

### Type Safety
Builders are strongly typed and prevent invalid configurations at compile time:

```go
// This won't compile - MinLength doesn't exist on IntegerBuilder
// schema.Integer().MinLength(5) // ❌ Compilation error

// Correct usage
schema.String().MinLength(5).Build() // ✅
```

### Method Chaining
Build complex schemas with readable, chainable methods:

```go
complexString := schema.String().
    MinLength(8).
    MaxLength(128).
    Pattern("^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).+$").
    Description("Strong password").
    Example("MyPassword123").
    Build()
```

### Immutability
Builders create immutable schema instances. Modifications return new instances:

```go
baseString := schema.String().MinLength(1)
email := baseString.Email().Build()     // New instance
username := baseString.Pattern("^[a-zA-Z0-9_]+$").Build() // Different instance
```

## Validation System

The validation system provides comprehensive data validation with detailed error reporting.

### Validation Result Structure

```go
type ValidationResult struct {
    Valid    bool              `json:"valid"`
    Errors   []ValidationError `json:"errors,omitempty"`
    Metadata map[string]any    `json:"metadata,omitempty"`
}

type ValidationError struct {
    Path       string `json:"path"`
    Message    string `json:"message"`
    Code       string `json:"code"`
    Value      any    `json:"value,omitempty"`
    Expected   string `json:"expected,omitempty"`
    Suggestion string `json:"suggestion,omitempty"`
    Context    string `json:"context,omitempty"`
}
```

### Validation Features

- **Path Tracking**: Errors include the exact path to invalid data
- **Error Codes**: Structured error codes for programmatic handling
- **Suggestions**: Helpful suggestions for fixing validation errors
- **Context**: Additional context about validation failures
- **Multiple Errors**: Reports all validation errors, not just the first

### Example Validation

```go
result := userSchema.Validate(map[string]any{
    "id": "invalid",  // Should be integer
    "name": "",       // Too short
    "email": "not-an-email", // Invalid format
})

if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Error at %s: %s\n", err.Path, err.Message)
        if err.Suggestion != "" {
            fmt.Printf("  Suggestion: %s\n", err.Suggestion)
        }
    }
}
```

## Schema Registry

The registry system enables reusable, parameterized schema definitions.

### Basic Registry Usage

```go
registry := registry.New()

// Define reusable schemas
registry.Define("User", userSchema)
registry.Define("Product", productSchema)

// Reference schemas
userRef := registry.Ref("User")
```

### Parameterized Schemas

Create generic schemas that can be specialized:

```go
// Define a generic List schema
registry.Define("List", schema.Array().
    Items(registry.Param("T")).
    Build(), "T")

// Define a generic Result schema  
registry.Define("Result", schema.Object().
    Property("success", schema.Boolean().Build()).
    Property("data", registry.Param("T")).
    Property("error", registry.Param("E")).
    Required("success").
    Build(), "T", "E")

// Create concrete instances
userList, _ := registry.Build("List").WithParam("T", registry.Ref("User")).Build()
userResult, _ := registry.Build("Result").
    WithParam("T", registry.Ref("User")).
    WithParam("E", schema.String().Build()).
    Build()
```

### Registry Features

- **Named Schemas**: Store schemas with string identifiers
- **Parameterization**: Create generic schemas with type parameters
- **Reference Resolution**: Automatically resolve schema references
- **Circular Reference Detection**: Prevents infinite recursion
- **Schema Merging**: Combine multiple registries
- **Cloning**: Create independent registry copies

## Reflection and Struct Generation

Generate schemas automatically from Go struct types using reflection and struct tags.

### Basic Struct Schema Generation

```go
type User struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    IsActive bool   `json:"is_active"`
}

userSchema := schema.FromStruct[User]()
```

### Struct Tag Support

Control schema generation with struct tags:

```go
type Product struct {
    ID          int64   `json:"id" schema:"min=1"`
    Name        string  `json:"name" schema:"minlen=1,maxlen=200"`
    Price       float64 `json:"price" schema:"min=0"`
    Category    string  `json:"category" schema:"enum=electronics|books|clothing"`
    Description *string `json:"description,omitempty" schema:"maxlen=1000"`
    Tags        []string `json:"tags,omitempty" schema:"maxitems=10"`
}
```

### Supported Struct Tags

- **Validation**: `min`, `max`, `minlen`, `maxlen`, `pattern`, `email`, `url`
- **Constraints**: `enum`, `maxitems`, `minitems`, `uniqueitems`
- **Documentation**: `desc`, `example`
- **Behavior**: `required`, `omitempty`

### Embedded Structs

Handle embedded struct fields:

```go
type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
}

type User struct {
    Name    string `json:"name"`
    Address        // Embedded - properties are merged into parent
}
```

## Metadata and Documentation

All schemas support rich metadata for documentation and tooling:

```go
schema := schema.String().
    Name("username").
    Description("Unique identifier for user account").
    Example("john_doe").
    Tags("user", "auth", "required").
    Build()

metadata := schema.Metadata()
fmt.Println(metadata.Description) // "Unique identifier for user account"
```

### Metadata Structure

```go
type SchemaMetadata struct {
    Name        string            `json:"name,omitempty"`
    Description string            `json:"description,omitempty"`
    Examples    []any             `json:"examples,omitempty"`
    Tags        []string          `json:"tags,omitempty"`
    Properties  map[string]string `json:"properties,omitempty"`
}
```

## Example Generation

Automatically generate valid example data for any schema:

```go
userSchema := schema.Object().
    Property("id", schema.Integer().Min(1).Max(1000).Build()).
    Property("name", schema.String().MinLength(3).Build()).
    Property("email", schema.String().Email().Build()).
    Build()

example := userSchema.GenerateExample()
// Might generate: {"id": 42, "name": "john", "email": "user@example.com"}
```

### Example Generation Features

- **Constraint Aware**: Generated examples respect all validation constraints
- **Realistic Data**: Uses realistic example data when possible
- **Deterministic**: Same schema generates consistent examples
- **Nested Support**: Handles complex nested object structures
- **Format Support**: Generates proper examples for email, URL, date formats

## Design Philosophy

### Functions as First-Class Citizens

Unlike traditional validation libraries, this package treats function signatures as schema types. This enables:

- **API-First Development**: Define contracts before implementation
- **Service Interface Validation**: Validate service method calls
- **Auto-Generated Documentation**: Generate API docs from schemas
- **Client SDK Generation**: Generate type-safe clients

### Type Safety Without Code Generation

The library provides type safety through Go's type system rather than code generation:

- **Compile-Time Safety**: Invalid schema configurations won't compile
- **No Magic**: All behavior is explicit and discoverable
- **Standard Go**: Uses standard Go patterns and idioms

### Comprehensive Validation

Validation goes beyond simple type checking:

- **Detailed Error Reporting**: Precise error locations and suggestions
- **Multiple Error Collection**: Report all errors, not just the first
- **Custom Error Messages**: Contextual, helpful error messages
- **Format Validation**: Built-in support for common formats

### Performance and Memory Efficiency

- **Immutable Schemas**: Safe for concurrent use
- **Efficient Validation**: Optimized validation paths
- **Minimal Allocations**: Careful memory management
- **Lazy Evaluation**: Expensive operations are deferred when possible

### Integration Ready

Designed for real-world integration scenarios:

- **JSON Schema Compatibility**: Export to standard JSON Schema
- **HTTP Integration**: Direct integration with HTTP handlers
- **WebSocket Support**: Real-time validation and communication
- **JavaScript Bridge**: Client-side validation capabilities
- **Registry System**: Manage schemas across large applications