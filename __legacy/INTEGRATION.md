# Integration Guide

This document covers the various integration capabilities and usage patterns for the Schema library, including HTTP APIs, WebSocket services, JavaScript clients, and other integration scenarios.

## Table of Contents

1. [HTTP API Integration](#http-api-integration)
2. [WebSocket Integration](#websocket-integration)
3. [JavaScript Integration](#javascript-integration)
4. [JSON Schema Compatibility](#json-schema-compatibility)
5. [Reflection and Struct Integration](#reflection-and-struct-integration)
6. [Service Architecture Patterns](#service-architecture-patterns)
7. [Testing Integration](#testing-integration)
8. [Performance Considerations](#performance-considerations)
9. [Error Handling Patterns](#error-handling-patterns)
10. [Best Practices](#best-practices)

## HTTP API Integration

The library provides seamless integration with HTTP APIs through function schemas and handlers.

### Basic HTTP Endpoint

```go
package main

import (
    "context"
    "net/http"
    "defs.dev/schema"
    "defs.dev/schema/functions/http"
)

func main() {
    // Define the API endpoint schema
    userCreateSchema := schema.NewFunctionSchema().
        Input("name", schema.String().MinLength(1).MaxLength(100).Build()).
        Input("email", schema.String().Email().Build()).
        Input("age", schema.Integer().Min(0).Max(150).Build()).
        Output(schema.Object().
            Property("id", schema.Integer().Build()).
            Property("name", schema.String().Build()).
            Property("email", schema.String().Build()).
            Property("created_at", schema.String().Build()).
            Required("id", "name", "email", "created_at").
            Build()).
        Error(schema.Object().
            Property("error", schema.String().Build()).
            Property("code", schema.Integer().Build()).
            Build()).
        Description("Create a new user").
        Build()

    // Implement the business logic
    handler := func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
        name := params["name"].(string)
        email := params["email"].(string)
        age := params["age"].(int)
        
        // Business logic here
        user := map[string]any{
            "id":         123,
            "name":      name,
            "email":     email,
            "created_at": "2024-01-01T00:00:00Z",
        }
        
        return schema.FunctionOutput(user), nil
    }

    // Create HTTP handler with automatic validation
    httpHandler := http.NewHandler(userCreateSchema, handler)
    
    // Register with HTTP server
    mux := http.NewServeMux()
    mux.Handle("/users", httpHandler)
    
    http.ListenAndServe(":8080", mux)
}
```

### HTTP Integration Features

- **Automatic Request Validation**: Input parameters are validated against the schema
- **Response Validation**: Output is validated before sending to client
- **Error Handling**: Structured error responses with validation details
- **Content Negotiation**: Supports JSON request/response bodies
- **Middleware Support**: Integrates with standard HTTP middleware patterns

### Advanced HTTP Patterns

#### Request/Response Middleware
```go
// Custom middleware for logging and metrics
httpHandler := http.NewHandler(schema, handler).
    WithMiddleware(loggingMiddleware).
    WithMiddleware(metricsMiddleware)
```

#### Custom Error Handling
```go
httpHandler := http.NewHandler(schema, handler).
    WithErrorHandler(func(err error) (int, any) {
        if validationErr, ok := err.(*schema.ValidationError); ok {
            return 400, map[string]any{
                "error": "Validation failed",
                "details": validationErr.Errors,
            }
        }
        return 500, map[string]any{"error": "Internal server error"}
    })
```

## WebSocket Integration

Real-time communication with schema validation and function calls over WebSockets.

### WebSocket Portal Setup

```go
package main

import (
    "context"
    "net/http"
    "defs.dev/schema"
    "defs.dev/schema/functions/websocket"
)

func main() {
    // Create WebSocket portal
    portal := websocket.NewPortal()
    
    // Define real-time function schemas
    chatMessageSchema := schema.NewFunctionSchema().
        Input("room", schema.String().MinLength(1).Build()).
        Input("message", schema.String().MinLength(1).MaxLength(1000).Build()).
        Input("sender", schema.String().MinLength(1).Build()).
        Output(schema.Object().
            Property("id", schema.String().Build()).
            Property("timestamp", schema.String().Build()).
            Required("id", "timestamp").
            Build()).
        Description("Send chat message").
        Build()

    // Register function handler
    portal.RegisterFunction("sendMessage", chatMessageSchema, func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
        room := params["room"].(string)
        message := params["message"].(string)
        sender := params["sender"].(string)
        
        // Broadcast to room participants
        messageData := map[string]any{
            "id":        generateID(),
            "room":      room,
            "message":   message,
            "sender":    sender,
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        }
        
        portal.BroadcastToRoom(room, "messageReceived", messageData)
        
        return schema.FunctionOutput{
            "id":        messageData["id"],
            "timestamp": messageData["timestamp"],
        }, nil
    })

    // Start WebSocket server
    http.Handle("/ws", portal.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### WebSocket Features

- **Bidirectional Communication**: Full-duplex communication with validation
- **Function Call Routing**: Route WebSocket messages to function handlers
- **Room Management**: Built-in support for rooms and broadcasting
- **Connection Management**: Automatic connection lifecycle management
- **Error Propagation**: Structured error responses over WebSocket

### WebSocket Client Integration

```javascript
// JavaScript client example
const ws = new WebSocket('ws://localhost:8080/ws');

// Send validated function call
ws.send(JSON.stringify({
    type: 'call',
    id: 'msg_123',
    function: 'sendMessage',
    params: {
        room: 'general',
        message: 'Hello, world!',
        sender: 'john_doe'
    }
}));

// Handle responses
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'response') {
        console.log('Function response:', message.result);
    } else if (message.type === 'error') {
        console.error('Validation error:', message.error);
    }
};
```

## JavaScript Integration

Bridge schemas to client-side JavaScript for consistent validation.

### JavaScript Schema Generation

```go
// Generate JavaScript validation code
jsCode := schema.ToJavaScript(userSchema, schema.JavaScriptOptions{
    ModuleName: "UserValidator",
    ExportType: "ES6",
})

// Write to file for client-side use
os.WriteFile("user-validator.js", []byte(jsCode), 0644)
```

### Client-Side Validation

```javascript
// Generated JavaScript module
import { validateUser } from './user-validator.js';

// Validate data on client-side
const userData = {
    name: 'John Doe',
    email: 'john@example.com',
    age: 30
};

const result = validateUser(userData);
if (!result.valid) {
    console.error('Validation errors:', result.errors);
}
```

### JavaScript Integration Features

- **Client-Side Validation**: Same validation logic on client and server
- **Real-Time Feedback**: Immediate validation feedback in forms
- **Type Safety**: TypeScript definitions generated from schemas
- **Framework Agnostic**: Works with React, Vue, Angular, and vanilla JS

### Framework-Specific Integration

#### React Hook Example
```javascript
import { useSchemaValidation } from './hooks/useSchemaValidation';

function UserForm() {
    const { validate, errors, isValid } = useSchemaValidation(userSchema);
    
    const handleSubmit = (formData) => {
        if (validate(formData)) {
            // Submit valid data
            submitUser(formData);
        }
    };
    
    return (
        <form onSubmit={handleSubmit}>
            {/* Form fields with validation errors */}
            {errors.name && <div className="error">{errors.name}</div>}
        </form>
    );
}
```

## JSON Schema Compatibility

Export schemas to standard JSON Schema format for interoperability.

### JSON Schema Export

```go
// Convert to JSON Schema
jsonSchema := userSchema.ToJSONSchema()

// Serialize to JSON
jsonBytes, _ := json.MarshalIndent(jsonSchema, "", "  ")
fmt.Println(string(jsonBytes))
```

### Example JSON Schema Output

```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "integer",
      "minimum": 1
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "email": {
      "type": "string",
      "format": "email"
    },
    "age": {
      "type": "integer",
      "minimum": 0,
      "maximum": 150
    }
  },
  "required": ["id", "name", "email"],
  "additionalProperties": false
}
```

### JSON Schema Integration

- **OpenAPI Support**: Use with OpenAPI/Swagger documentation
- **Third-Party Tools**: Compatible with JSON Schema validators
- **Documentation Generation**: Generate API documentation
- **Contract Testing**: Use for contract testing between services

## Reflection and Struct Integration

Seamlessly integrate with existing Go structs and generate schemas automatically.

### Struct Tag Configuration

```go
type User struct {
    ID       int64  `json:"id" schema:"min=1,desc=Unique user identifier"`
    Name     string `json:"name" schema:"minlen=1,maxlen=100,desc=User display name"`
    Email    string `json:"email" schema:"email,desc=User email address"`
    Age      *int   `json:"age,omitempty" schema:"min=0,max=150,desc=User age in years"`
    IsActive bool   `json:"is_active" schema:"desc=Whether the user account is active"`
    Tags     []string `json:"tags,omitempty" schema:"maxitems=10,desc=User tags"`
    Profile  UserProfile `json:"profile" schema:"desc=User profile information"`
}

type UserProfile struct {
    Bio       string    `json:"bio,omitempty" schema:"maxlen=500,desc=User biography"`
    Website   *string   `json:"website,omitempty" schema:"url,desc=User website URL"`
    CreatedAt time.Time `json:"created_at" schema:"desc=Account creation time"`
}
```

### Advanced Struct Integration

#### Custom Type Mappings
```go
// Register custom type mappings
registry.RegisterType(reflect.TypeOf(time.Time{}), func() schema.Schema {
    return schema.String().
        Format("date-time").
        Description("ISO 8601 datetime").
        Build()
})

// Register custom validation
registry.RegisterType(reflect.TypeOf(MyCustomType{}), func() schema.Schema {
    return schema.String().
        Pattern("^[A-Z]{2,4}$").
        Description("Custom format code").
        Build()
})
```

#### Struct Validation Service
```go
// Service for validating structs with schemas
type ValidationService struct {
    cache map[reflect.Type]schema.Schema
}

func (vs *ValidationService) Validate(v interface{}) schema.ValidationResult {
    schema := vs.getOrCreateSchema(reflect.TypeOf(v))
    return schema.Validate(v)
}

func (vs *ValidationService) getOrCreateSchema(t reflect.Type) schema.Schema {
    if cached, ok := vs.cache[t]; ok {
        return cached
    }
    
    // Generate schema from type
    s := schema.FromType(t)
    vs.cache[t] = s
    return s
}
```

## Service Architecture Patterns

Common patterns for using schemas in service architectures.

### API Gateway Pattern

```go
// API Gateway with schema validation
type APIGateway struct {
    services map[string]Service
    schemas  map[string]schema.Schema
}

func (gw *APIGateway) HandleRequest(serviceName, method string, data any) (any, error) {
    // Get service schema
    schema, ok := gw.schemas[serviceName+"."+method]
    if !ok {
        return nil, errors.New("unknown service method")
    }
    
    // Validate request
    if result := schema.Validate(data); !result.Valid {
        return nil, &ValidationError{Errors: result.Errors}
    }
    
    // Route to service
    service := gw.services[serviceName]
    return service.Call(method, data)
}
```

### Microservice Communication

```go
// Schema-validated service client
type ServiceClient struct {
    baseURL string
    schemas map[string]schema.Schema
}

func (c *ServiceClient) Call(method string, params any) (any, error) {
    // Validate parameters against schema
    schema := c.schemas[method]
    if result := schema.Validate(params); !result.Valid {
        return nil, &ValidationError{Errors: result.Errors}
    }
    
    // Make HTTP request with validated data
    resp, err := c.httpClient.Post(
        c.baseURL+"/"+method,
        "application/json",
        encodeJSON(params),
    )
    
    // Validate response
    var response any
    json.NewDecoder(resp.Body).Decode(&response)
    
    outputSchema := c.schemas[method+".output"]
    if result := outputSchema.Validate(response); !result.Valid {
        return nil, &ResponseValidationError{Errors: result.Errors}
    }
    
    return response, nil
}
```

### Event-Driven Architecture

```go
// Schema-validated event system
type EventBus struct {
    schemas    map[string]schema.Schema
    handlers   map[string][]EventHandler
}

func (eb *EventBus) Publish(eventType string, data any) error {
    // Validate event data
    schema, ok := eb.schemas[eventType]
    if ok {
        if result := schema.Validate(data); !result.Valid {
            return &EventValidationError{
                EventType: eventType,
                Errors:    result.Errors,
            }
        }
    }
    
    // Notify handlers
    for _, handler := range eb.handlers[eventType] {
        go handler.Handle(eventType, data)
    }
    
    return nil
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}
```

## Testing Integration

Use schemas for comprehensive testing strategies.

### Schema-Based Test Generation

```go
func TestUserValidation(t *testing.T) {
    generator := schema.NewGenerator()
    
    // Generate valid test cases
    for i := 0; i < 100; i++ {
        validUser := generator.Generate(userSchema)
        
        // Test that generated data passes validation
        result := userSchema.Validate(validUser)
        assert.True(t, result.Valid, "Generated data should be valid")
        
        // Test business logic with valid data
        err := userService.CreateUser(validUser)
        assert.NoError(t, err)
    }
}
```

### Property-Based Testing

```go
func TestUserServiceProperties(t *testing.T) {
    property := gopter.NewProperties(nil)
    
    property.Property("valid users are always accepted", prop.ForAll(
        func(user any) bool {
            // Only test with schema-valid users
            if result := userSchema.Validate(user); !result.Valid {
                return true // Skip invalid users
            }
            
            // Test that valid users are processed successfully
            err := userService.CreateUser(user)
            return err == nil
        },
        schema.GeneratorFor(userSchema),
    ))
    
    property.TestingRun(t)
}
```

### Contract Testing

```go
func TestAPIContract(t *testing.T) {
    // Test that API responses match schema
    response := callAPI("/users/123")
    
    result := userSchema.Validate(response)
    if !result.Valid {
        t.Errorf("API response violates contract:")
        for _, err := range result.Errors {
            t.Errorf("  %s: %s", err.Path, err.Message)
        }
    }
}
```

## Performance Considerations

Optimize schema usage for production environments.

### Schema Caching

```go
type SchemaCache struct {
    cache map[string]schema.Schema
    mutex sync.RWMutex
}

func (sc *SchemaCache) Get(key string) (schema.Schema, bool) {
    sc.mutex.RLock()
    defer sc.mutex.RUnlock()
    
    schema, ok := sc.cache[key]
    return schema, ok
}

func (sc *SchemaCache) Set(key string, schema schema.Schema) {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()
    
    sc.cache[key] = schema
}
```

### Validation Optimization

```go
// Pre-compile expensive validations
type OptimizedValidator struct {
    compiledPatterns map[string]*regexp.Regexp
    schemas          map[string]schema.Schema
}

func (ov *OptimizedValidator) Validate(schemaName string, data any) schema.ValidationResult {
    // Use cached compiled regex patterns and schemas
    schema := ov.schemas[schemaName]
    return schema.Validate(data)
}
```

### Memory Management

```go
// Pool validation results to reduce allocations
var validationResultPool = sync.Pool{
    New: func() interface{} {
        return &schema.ValidationResult{
            Errors: make([]schema.ValidationError, 0, 10),
        }
    },
}

func validateWithPool(schema schema.Schema, data any) schema.ValidationResult {
    result := validationResultPool.Get().(*schema.ValidationResult)
    defer validationResultPool.Put(result)
    
    // Reset result
    result.Valid = false
    result.Errors = result.Errors[:0]
    
    // Perform validation
    return schema.Validate(data)
}
```

## Error Handling Patterns

Best practices for handling validation and schema errors.

### Structured Error Response

```go
type APIError struct {
    Code       string                    `json:"code"`
    Message    string                    `json:"message"`
    Details    []schema.ValidationError  `json:"details,omitempty"`
    Timestamp  time.Time                 `json:"timestamp"`
    RequestID  string                    `json:"request_id,omitempty"`
}

func handleValidationError(result schema.ValidationResult, requestID string) APIError {
    return APIError{
        Code:      "VALIDATION_ERROR",
        Message:   "Request validation failed",
        Details:   result.Errors,
        Timestamp: time.Now().UTC(),
        RequestID: requestID,
    }
}
```

### Error Recovery

```go
func processWithRecovery(schema schema.Schema, data any) (any, error) {
    result := schema.Validate(data)
    if !result.Valid {
        // Attempt automatic recovery for common issues
        fixedData := attemptAutoFix(data, result.Errors)
        
        // Re-validate fixed data
        if retryResult := schema.Validate(fixedData); retryResult.Valid {
            log.Printf("Auto-fixed validation errors: %d", len(result.Errors))
            return processData(fixedData), nil
        }
        
        // Return original errors if auto-fix failed
        return nil, &ValidationError{Errors: result.Errors}
    }
    
    return processData(data), nil
}
```

## Best Practices

### Schema Design

1. **Keep Schemas Simple**: Start with basic validation and add complexity as needed
2. **Use Descriptive Names**: Give schemas and properties meaningful names
3. **Document Extensively**: Use descriptions, examples, and tags
4. **Version Schemas**: Plan for schema evolution and backward compatibility
5. **Validate Early**: Validate at service boundaries, not deep in business logic

### Performance Optimization

1. **Cache Schemas**: Pre-compile and cache schemas for reuse
2. **Pool Objects**: Use object pools for frequently allocated validation results
3. **Optimize Patterns**: Pre-compile regular expressions
4. **Lazy Loading**: Generate schemas on-demand rather than upfront
5. **Benchmark**: Measure validation performance in production scenarios

### Error Handling

1. **Provide Context**: Include helpful error messages and suggestions
2. **Fail Fast**: Return validation errors immediately
3. **Log Appropriately**: Log validation failures for monitoring
4. **Handle Gracefully**: Provide fallback behavior for validation failures
5. **User-Friendly Messages**: Transform technical errors for end users

### Integration Architecture

1. **Centralize Schemas**: Use registries to manage schemas across services
2. **Generate Code**: Auto-generate client code from schemas
3. **Test Contracts**: Validate API contracts in tests
4. **Monitor Validation**: Track validation success rates and common errors
5. **Version APIs**: Use schemas to manage API versioning and compatibility