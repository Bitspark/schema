# Schema

A comprehensive, type-safe schema validation and generation library for Go that treats function signatures as first-class citizens.

## Overview

The `schema` package provides a powerful and flexible way to define, validate, and generate schemas in Go. Unlike traditional validation libraries, it treats function signatures as first-class schema types, making it ideal for API development, service interfaces, and data validation pipelines.

### Key Features

- **Type-Safe Schema Definitions**: Fluent builder API with compile-time safety
- **Function Schemas**: Define function signatures as schemas for API contracts
- **Comprehensive Validation**: Detailed error reporting with suggestions
- **Schema Registry**: Manage and parameterize reusable schema definitions
- **Reflection Support**: Generate schemas from Go structs with struct tags
- **Multiple Integrations**: HTTP endpoints, WebSocket, JavaScript, JSON Schema
- **Example Generation**: Automatically generate valid example data
- **Performance Focused**: Efficient validation and minimal memory overhead

## Installation

```bash
go get defs.dev/schema
```

## Quick Start

### Basic Schema Creation

```go
package main

import (
    "fmt"
    "defs.dev/schema"
)

func main() {
    // Create a user schema using the fluent builder API
    userSchema := schema.Object().
        Property("id", schema.Integer().Min(1).Build()).
        Property("name", schema.String().MinLength(1).MaxLength(100).Build()).
        Property("email", schema.String().Email().Build()).
        Property("age", schema.Integer().Min(0).Max(150).Build()).
        Required("id", "name", "email").
        Build()

    // Validate data
    user := map[string]any{
        "id":    123,
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }

    result := userSchema.Validate(user)
    if result.Valid {
        fmt.Println("User is valid!")
    } else {
        for _, err := range result.Errors {
            fmt.Printf("Validation error: %s\n", err.Message)
        }
    }
}
```

### Function Schemas

Define function signatures as schemas for API contracts:

```go
// Define a payment processing function schema
paymentSchema := schema.NewFunctionSchema().
    Input("amount", schema.Number().Min(0).Build()).
    Input("method", schema.String().Enum("card", "bank", "crypto").Build()).
    Input("currency", schema.String().Pattern("^[A-Z]{3}$").Build()).
    Output(schema.Object().
        Property("transactionId", schema.String().Build()).
        Property("status", schema.String().Enum("pending", "completed", "failed").Build()).
        Required("transactionId", "status").
        Build()).
    Error(schema.Object().
        Property("code", schema.String().Build()).
        Property("message", schema.String().Build()).
        Build()).
    Description("Process a payment transaction").
    Build()
```

### Schema Registry

Manage reusable schemas with parameterization:

```go
registry := registry.New()

// Define parameterized schemas
registry.Define("List", schema.Array().Items(registry.Param("T")).Build(), "T")

registry.Define("User", schema.Object().
    Property("id", schema.Integer().Build()).
    Property("name", schema.String().Build()).
    Build())

// Create concrete instances
userList, err := registry.Build("List").WithParam("T", registry.Ref("User")).Build()
if err != nil {
    log.Fatal(err)
}
```

### Reflection-Based Schemas

Generate schemas from Go structs:

```go
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name" schema:"minlen=1,maxlen=100"`
    Email string `json:"email" schema:"email"`
    Age   *int   `json:"age,omitempty" schema:"min=0,max=150"`
}

// Generate schema from struct
userSchema := schema.FromStruct[User]()
```

## Core Schema Types

The library supports all common data types and advanced constructs:

- **Primitives**: `String()`, `Integer()`, `Number()`, `Boolean()`
- **Collections**: `Array()`, `Object()`, `Map()`
- **Advanced**: `Union()`, `Optional()`, `Any()`, `Null()`
- **References**: `Ref()` for schema registry references
- **Functions**: `NewFunctionSchema()` for function signatures

## Validation Features

- **Type Validation**: Ensure data matches expected types
- **Constraint Validation**: Min/max values, string lengths, patterns
- **Format Validation**: Email, URL, date formats, and custom formats
- **Required Fields**: Mark fields as required or optional
- **Nested Validation**: Deep validation of object and array properties
- **Custom Error Messages**: Detailed error reporting with suggestions

## Builder Pattern

All schemas use a fluent builder pattern for intuitive construction:

```go
schema := schema.String().
    MinLength(5).
    MaxLength(50).
    Pattern("^[a-zA-Z0-9_]+$").
    Description("Username field").
    Example("john_doe123").
    Build()
```

## JSON Schema Compatibility

Convert to standard JSON Schema format:

```go
jsonSchema := userSchema.ToJSONSchema()
// Output can be used with any JSON Schema validator
```

## Example Generation

Generate valid example data automatically:

```go
example := userSchema.GenerateExample()
// Returns a valid example that passes schema validation
```

## Documentation

- [CONCEPTS.md](CONCEPTS.md) - Core concepts and detailed feature explanations
- [INTEGRATION.md](INTEGRATION.md) - Integration capabilities and usage patterns

## Examples

See the `/functions` directory for comprehensive examples including:
- HTTP API endpoints
- WebSocket real-time services  
- JavaScript client integration
- Registry usage patterns

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests for any improvements.

## License

This project is licensed under the MIT License - see the LICENSE file for details.