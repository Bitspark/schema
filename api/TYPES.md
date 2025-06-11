# Type System Design Patterns

This document outlines the comprehensive design patterns for types, schemas, and builders in the schema API.

## Overview

The schema system uses a consistent pattern for all constructs, distinguishing between **primitive types** (native Go values) and **interface types** (complex implementations).

## The Complete Pattern

### For Primitive Types
Primitive types work directly with native Go values and only need schema-level construction:

| Type | Schema Interface | Schema Builder | Value Type | Value Builder |
|------|-----------------|----------------|------------|---------------|
| String | `StringSchema` | `StringSchemaBuilder` | `string` | *(native)* |
| Number | `NumberSchema` | `NumberSchemaBuilder` | `float64` | *(native)* |
| Integer | `IntegerSchema` | `IntegerSchemaBuilder` | `int64` | *(native)* |
| Boolean | `BooleanSchema` | `BooleanSchemaBuilder` | `bool` | *(native)* |
| Array | `ArraySchema` | `ArraySchemaBuilder` | `[]any` | *(native)* |
| Object | `ObjectSchema` | `ObjectSchemaBuilder` | `map[string]any` | *(native)* |

### For Interface Types
Interface types require both schema construction (contract definition) and value construction (implementation building):

| Type | Schema Interface | Schema Builder | Value Interface | Value Builder |
|------|-----------------|----------------|-----------------|---------------|
| **Function** | `FunctionSchema` | `FunctionSchemaBuilder` | `Function` | `FunctionBuilder` ⚠️ |
| **Service** | `ServiceSchema` | `ServiceSchemaBuilder` | `Service` | `ServiceBuilder` ⚠️ |
| **Component** | `ComponentSchema` | `ComponentSchemaBuilder` | `Component` | `ComponentBuilder` ⚠️ |
| **Topic** | `TopicSchema` | `TopicSchemaBuilder` | `Topic` | `TopicBuilder` ⚠️ |
| **Union** | `UnionSchema` | `UnionSchemaBuilder` | *(handled by schema)* | *(native)* |

⚠️ = Currently missing and should be implemented

## Why This Distinction Matters

### Primitive Types
- **Values are native Go types** (string, int, bool, []any, map[string]any)
- **Validation happens against schemas** - no complex construction needed
- **Usage**: Validate user input, generate examples, create JSON schemas

```go
// Schema building (defines validation rules)
nameSchema := NewString().
    MinLength(1).
    MaxLength(50).
    Pattern(`^[a-zA-Z\s]+$`).
    Build()

// Direct value usage (native Go)
name := "John Doe"
result := nameSchema.Validate(name) // ValidationResult
```

### Interface Types
- **Values must implement interfaces** - complex internal structure
- **Require fluent construction** - multiple components to assemble
- **Need both contract definition AND implementation building**

```go
// Schema building (defines the contract)
functionSchema := NewFunctionSchema().
    Input("name", NewString().Required()).
    Output("greeting", NewString()).
    Build()

// Value building (creates the implementation) - MISSING!
function := NewFunction().
    Schema(functionSchema).
    Handler(func(ctx context.Context, params FunctionData) (FunctionData, error) {
        name := params.Get("name").(string)
        return NewFunctionDataValue(fmt.Sprintf("Hello, %s!", name)), nil
    }).
    Build()
```

## Generic Types

Generics provide enhanced type safety for both patterns:

### Type-Safe Builders for Existing Schemas
- `ListBuilder[T]` → builds `ArraySchema` with compile-time type safety
- `UnionBuilder2[T1, T2]` → builds `UnionSchema` for two specific types

### New Generic Schema Types
- `OptionalSchema[T]` + `OptionalBuilder[T]` - nullable/optional values
- `ResultSchema[T, E]` + `ResultBuilder[T, E]` - success/error patterns  
- `MapSchema[K, V]` + `MapBuilder[K, V]` - key-value mappings

## Current Implementation Status

### ✅ Fully Implemented
- All primitive type schemas and builders
- Generic type-safe builders
- Schema visitor pattern
- Portal and registry systems

### ⚠️ Partially Implemented
- Function/Service interfaces exist but lack value builders
- Component/Topic interfaces are minimal placeholders

### ❌ Missing Implementation
- `FunctionBuilder` for fluent Function construction
- `ServiceBuilder` for fluent Service construction  
- `ComponentBuilder` and full Component system
- `TopicBuilder` and full Topic system

## Design Principles

1. **Consistency**: Every construct follows the same pattern
2. **Type Safety**: Generics provide compile-time guarantees where useful
3. **Fluent APIs**: Builders enable readable, discoverable construction
4. **Separation of Concerns**: Schemas define contracts, builders create implementations
5. **Progressive Enhancement**: Start with schemas, add value builders as needed

## Example Usage Patterns

### Schema Definition
```go
// Define what a user registration function should look like
userRegSchema := NewFunctionSchema().
    Input("email", NewString().Email().Required()).
    Input("password", NewString().MinLength(8).Required()).
    Input("age", NewInteger().Min(13).Required()).
    Output("userId", NewString().UUID()).
    Output("success", NewBoolean()).
    Build()
```

### Value Construction (Future)
```go
// Build an actual implementation of that function
userRegFunction := NewFunction().
    Schema(userRegSchema).
    Handler(func(ctx context.Context, params FunctionData) (FunctionData, error) {
        // Implementation logic here
        return NewFunctionDataValue(map[string]any{
            "userId": generateUUID(),
            "success": true,
        }), nil
    }).
    Middleware(authMiddleware, validationMiddleware).
    Build()
```

### Service Construction (Future)
```go
// Build a service with multiple methods
userService := NewService().
    Schema(userServiceSchema).
    Method("register", userRegFunction).
    Method("login", userLoginFunction).
    Method("profile", userProfileFunction).
    Build()
```

This pattern provides a complete, consistent foundation for building complex distributed systems with type safety and fluent APIs. 