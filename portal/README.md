# Portal System

The **Portal System** provides transport abstraction for function execution across different communication channels and protocols. It enables the same function to be called whether it's local, over HTTP, WebSockets, database connections, or any other transport mechanism.

## 🎯 Core Concept

**Portals** solve the fundamental problem: *"How do I execute the same function with the same interface regardless of where or how it's deployed?"*

They act as **transport adapters** that bridge the gap between function schemas (what to do) and transport mechanisms (how to communicate).

### Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Function      │    │     Portal      │    │  Transport      │
│   Schema        │◄──►│   (Adapter)     │◄──►│  Mechanism      │
│  (What to do)   │    │  (How to do)    │    │ (Where to do)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔧 Portal Responsibilities

### 1. **Address Generation**
Creates unique, addressable identifiers for functions:
```
Local:     "local://calculator/add"
HTTP:      "https://api.service.com/v1/calculator/add/abc123"
WebSocket: "ws://realtime.service.com/functions/add"
Database:  "postgres://db.service.com/functions.calculate_sum"
```

### 2. **Function Transformation**
Converts between different function representations:
- **Local functions** → **Network endpoints**
- **Network addresses** → **Callable function objects**
- **Function schemas** → **Transport-specific contracts**

### 3. **Protocol Bridging**
Handles transport-specific details transparently:
- HTTP: Request/response handling, headers, status codes
- WebSocket: Bidirectional messaging, connection management
- Database: SQL generation, connection pooling
- Message Queue: Async messaging, routing

### 4. **Schema Preservation**
Maintains function schema validation across all transports:
- Input parameter validation
- Output format validation
- Error handling consistency
- Metadata preservation

## 🌐 Portal Types

### Local Portal
**Purpose**: In-process function execution  
**Use Case**: Development, testing, high-performance local calls
```go
localPortal := portal.NewLocalPortal()
addr := localPortal.Apply("add", addSchema, addFunction)
// addr: "local://add/unique-id"
```

### HTTP Portal
**Purpose**: REST API-style function calls  
**Use Case**: Web APIs, microservices, public endpoints
```go
httpPortal := portal.NewHTTPPortal(":8080")
addr := httpPortal.Apply("add", addSchema, addHandler)
// addr: "http://localhost:8080/api/add/abc123"
```

### WebSocket Portal
**Purpose**: Real-time bidirectional communication  
**Use Case**: Live updates, streaming, interactive applications
```go
wsPortal := portal.NewWebSocketPortal(":8081")
addr := wsPortal.Apply("add", addSchema, addHandler)
// addr: "ws://localhost:8081/functions/add"
```

### Testing Portal
**Purpose**: Mock/stub functions for testing  
**Use Case**: Unit tests, integration tests, development
```go
testPortal := portal.NewTestingPortal()
addr := testPortal.Apply("add", addSchema, mockAddFunction)
// addr: "test://add/mock-123"
```

## 🔄 Usage Patterns

### Server Side (Function Publisher)

```go
package main

import (
    "defs.dev/schema/core"
    "defs.dev/schema/core/portal"
)

func main() {
    // Create function schema
    addSchema := core.NewFunction().
        Name("add").
        Description("Add two numbers").
        RequiredInput("a", core.NewNumber().Build()).
        RequiredInput("b", core.NewNumber().Build()).
        RequiredOutput("result", core.NewNumber().Build()).
        Build()

    // Create local function implementation
    addFunction := func(ctx context.Context, params api.FunctionInput) (api.FunctionOutput, error) {
        a, _ := params.Get("a")
        b, _ := params.Get("b")
        result := a.(float64) + b.(float64)
        return &portal.FunctionOutputValue{Value: result}, nil
    }

    // Publish via multiple portals
    httpPortal := portal.NewHTTPPortal(":8080")
    wsPortal := portal.NewWebSocketPortal(":8081") 
    
    // Same function, multiple transports
    httpAddr, _ := httpPortal.Apply("add", addSchema, addFunction)
    wsAddr, _ := wsPortal.Apply("add", addSchema, addFunction)
    
    fmt.Printf("HTTP Address: %s\n", httpAddr)      // http://localhost:8080/api/add/abc123
    fmt.Printf("WebSocket Address: %s\n", wsAddr)   // ws://localhost:8081/functions/add
}
```

### Client Side (Function Consumer)

```go
func callRemoteFunction() {
    ctx := context.Background()
    
    // Get portal for the target address
    httpPortal := portal.NewHTTPPortal()
    
    // Resolve remote function by address
    remoteAdd, err := httpPortal.ResolveFunction(ctx, "http://api.service.com/add/abc123")
    if err != nil {
        log.Fatal(err)
    }
    
    // Call remote function with same interface as local
    params := portal.FunctionInputMap{
        "a": 10.0,
        "b": 5.0,
    }
    
    result, err := remoteAdd.Call(ctx, params)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %v\n", result.ToAny()) // Result: 15.0
}
```

## 🏗️ Integration with Schema Core

### Registry Integration
Portals work seamlessly with the Function Registry:

```go
// Create registry and portal
registry := core.NewFunctionRegistry()
httpPortal := portal.NewHTTPPortal(":8080")

// Register function locally
registry.Register("add", addFunction)

// Expose via portal
addr, err := httpPortal.Apply("add", addFunction.Schema(), addFunction)
if err != nil {
    log.Fatal(err)
}

// Function is now available both locally and via HTTP
// Local call:
localResult, _ := registry.Call(ctx, "add", params)

// Remote call:  
remoteFunc, _ := httpPortal.ResolveFunction(ctx, addr)
remoteResult, _ := remoteFunc.Call(ctx, params)
```

### Service Integration
Portals automatically expose service methods:

```go
// Create service schema
userService := core.NewService().
    Name("UserService").
    Method("getUser", getUserSchema).
    Method("createUser", createUserSchema).
    Build()

// Register service
serviceRegistry := core.NewServiceRegistry()
serviceRegistry.RegisterService("UserService", userService)

// Expose entire service via portal
httpPortal := portal.NewHTTPPortal(":8080")
serviceAddr, _ := httpPortal.ApplyService("UserService", userService, serviceInstance)

// Individual method addresses:
// http://localhost:8080/api/UserService/getUser/abc123
// http://localhost:8080/api/UserService/createUser/def456
```

### Schema Validation
Portals enforce schema validation automatically:

```go
// Invalid input is rejected before transport
invalidParams := portal.FunctionInputMap{
    "a": "not a number",  // Schema expects number
    "b": 5.0,
}

result, err := remoteAdd.Call(ctx, invalidParams)
// err: Input validation failed: parameter 'a' must be a number
```

## 🔀 Multi-Transport Scenarios

### Load Balancing
```go
// Same function on multiple HTTP instances
addresses := []string{
    "http://server1:8080/api/add/abc123",
    "http://server2:8080/api/add/abc123", 
    "http://server3:8080/api/add/abc123",
}

// Round-robin calls
for i, addr := range addresses {
    fn, _ := httpPortal.ResolveFunction(ctx, addr)
    result, _ := fn.Call(ctx, params)
    fmt.Printf("Server %d result: %v\n", i+1, result.ToAny())
}
```

### Fallback Mechanisms
```go
// Try WebSocket first, fallback to HTTP
wsAddr := "ws://service.com/add"
httpAddr := "http://service.com/api/add/backup"

// Attempt WebSocket
wsFunc, err := wsPortal.ResolveFunction(ctx, wsAddr)
if err == nil {
    result, err := wsFunc.Call(ctx, params)
    if err == nil {
        return result
    }
}

// Fallback to HTTP
httpFunc, _ := httpPortal.ResolveFunction(ctx, httpAddr)
return httpFunc.Call(ctx, params)
```

### Cross-Protocol Communication
```go
// HTTP client calls WebSocket server
httpClient := portal.NewHTTPPortal()
wsFunction, _ := httpClient.ResolveFunction(ctx, "ws://realtime.service.com/add")

// Portal handles protocol conversion automatically
result, _ := wsFunction.Call(ctx, params)
```

## 📋 Portal Implementation Checklist

### Core Interface Implementation
- [ ] `Apply(address, schema, data) Function`
- [ ] `GenerateAddress(name, data) string`
- [ ] `Scheme() []string`
- [ ] `ResolveFunction(ctx, address) Function`

### Transport-Specific Features
- [ ] **Connection Management**: Pooling, keepalive, reconnection
- [ ] **Error Handling**: Transport errors, timeouts, retries
- [ ] **Middleware Support**: Authentication, logging, metrics
- [ ] **Configuration**: Transport-specific settings
- [ ] **Health Checks**: Connection and service health monitoring

### Schema Integration
- [ ] **Input Validation**: Enforce function schema on inputs
- [ ] **Output Validation**: Validate function outputs  
- [ ] **Error Schema**: Handle structured error responses
- [ ] **Metadata Preservation**: Maintain function metadata across transport

### Development Features
- [ ] **Address Parsing**: Parse and validate addresses
- [ ] **Discovery**: List available functions
- [ ] **Debugging**: Request/response logging
- [ ] **Testing**: Mock/stub functionality

## 🎭 Analogy: Portal as Phone System

Think of portals like **telephone systems**:

- **Function** = The person you want to talk to
- **Portal** = The phone system (landline, mobile, VoIP, video call)
- **Address** = The phone number  
- **Schema** = The language/protocol you both understand

You can call the same person via different "portals" (landline, mobile, Skype), but the conversation (function call) remains the same. The portal handles all the technical details of establishing the connection and routing the call.

## 🚀 Benefits

### 1. **Transport Agnostic Development**
Write functions once, deploy anywhere:
```go
// Same function works everywhere
addFunction := createAddFunction()

localPortal.Apply("add", schema, addFunction)    // local://add/123
httpPortal.Apply("add", schema, addFunction)     // http://api.com/add/123  
wsPortal.Apply("add", schema, addFunction)       // ws://api.com/add
dbPortal.Apply("add", schema, addFunction)       // postgres://db/add
```

### 2. **Consistent Interface**
Function calls look the same regardless of transport:
```go
// All of these have identical interfaces
result1, _ := localFunc.Call(ctx, params)
result2, _ := httpFunc.Call(ctx, params)
result3, _ := wsFunc.Call(ctx, params)
result4, _ := dbFunc.Call(ctx, params)
```

### 3. **Hot Swapping**
Change transports without changing code:
```go
// Development: local calls
func init() { 
    setCalculator(localPortal.ResolveFunction(ctx, "local://add/123"))
}

// Production: HTTP calls  
func init() {
    setCalculator(httpPortal.ResolveFunction(ctx, "https://api.prod.com/add/123"))
}
```

### 4. **Easy Testing**
Swap real functions with mocks:
```go
// Production code
realFunc, _ := httpPortal.ResolveFunction(ctx, "https://api.com/add/123")

// Test code  
mockFunc, _ := testPortal.ResolveFunction(ctx, "test://add/mock")
```

### 5. **Service Discovery**
Functions have addressable, shareable identifiers:
```go
// Share function addresses between services
config := map[string]string{
    "calculator.add": "https://math.service.com/add/abc123",
    "user.lookup":    "grpc://user.service.com/lookup",
    "email.send":     "queue://email.service.com/send",
}
```

## 🛠️ Implementation Roadmap

### Phase 1: Core Portal Infrastructure
1. **Base Portal Interface** - Define core portal contract
2. **Address System** - URL-like addressing for functions
3. **Local Portal** - In-process function execution
4. **Testing Portal** - Mock/stub functionality

### Phase 2: Network Portals  
1. **HTTP Portal** - REST API-style function calls
2. **WebSocket Portal** - Real-time bidirectional communication
3. **gRPC Portal** - High-performance RPC calls

### Phase 3: Advanced Portals
1. **Database Portal** - SQL function calls  
2. **Message Queue Portal** - Async function execution
3. **Cloud Portal** - AWS Lambda, Azure Functions, etc.

### Phase 4: Production Features
1. **Load Balancing** - Distribute calls across instances
2. **Circuit Breakers** - Handle service failures gracefully
3. **Metrics & Monitoring** - Track portal performance
4. **Security** - Authentication, authorization, encryption

The Portal System transforms the schema core from a validation library into a **distributed function execution platform**! 🌐 