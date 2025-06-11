# Function and Service Integration Strategy for Schema Core

This document outlines the integration strategy for bringing comprehensive function and service schema support to the `schema/core` package, based on the existing implementations in the main `schema` package.

## 🎯 Executive Summary

The current `schema/core` roadmap **significantly underestimates** the function and service capabilities that exist in the main `schema` package. This document proposes a comprehensive integration strategy to bring these advanced features to the core package using the API-first approach.

**Current Gap**: The existing schema package has mature function registry, service reflection, portal systems, and multi-transport execution - none of which are adequately represented in the core package plans.

---

## 📊 Current State Analysis

### ✅ Existing Capabilities in Main Schema Package

#### 1. Function Schema System
- **`FunctionSchema`** - Complete function signature validation
  - Input parameter schemas with validation
  - Output schema validation  
  - Error schema handling
  - JSON Schema generation
  - Metadata and documentation support

- **`TypedFunction`** Interface - Universal callable function abstraction
- **`FunctionHandler`** - Local function execution capability
- **Function Reflection** - Automatic schema generation from Go functions

#### 2. Service Reflection System
- **`ServiceReflector`** - Comprehensive service introspection
  - Method discovery from struct instances
  - Automatic function schema generation for methods
  - Service metadata extraction (`ServiceInfo`, `MethodInfo`)
  - Service schema generation as `ObjectSchema`

- **Service Method Binding** - Convert struct methods to callable functions
- **Service Validation** - Validate service contracts and signatures

#### 3. Function Registry System (`schema/functions/`)
- **Registry Interface** - Named function storage and retrieval
- **Portal System** - Multi-transport function execution
  - HTTP Portal (`functions/http/`)  
  - WebSocket Portal (`functions/websocket/`)
  - JavaScript Portal (`functions/javascript/`)
  - Local Portal (`functions/local/`)
  - Testing Portal (`functions/testing/`)

- **Function Addressing** - Unique address generation for functions
- **Type-safe Registration** - Generic portal system with type safety

#### 4. Advanced Features
- **Function Input/Output Types** - `FunctionInput`, `FunctionOutput` abstractions
- **Error Handling** - Structured error reporting for function calls
- **Concurrent Function Calls** - Thread-safe registry and execution
- **Function Lifecycle** - Registration, execution, cleanup

---

## 🚨 Core Package Integration Gaps

### Current TODO Status
- **Functions**: Mentioned but severely underscoped (basic signature validation only)
- **Services**: **Completely missing** from roadmap
- **Registry System**: **Not mentioned** at all
- **Portal System**: **Not mentioned** at all
- **Service Reflection**: **Not mentioned** at all

### Impact of Gaps
1. **Incomplete Migration Path** - Users cannot migrate service-based applications
2. **Feature Regression** - Core package would have fewer capabilities than legacy
3. **API Inconsistency** - Function/service APIs wouldn't align with core principles
4. **Integration Challenges** - No clear path for function registry integration

---

## 📋 Proposed Integration Roadmap

### Phase 2: Core Function & Service Types ⚠️ **EXPANDED**

#### FunctionSchema Implementation ✅ **ENHANCED SCOPE**
- [ ] **`schemas/function.go`** - Enhanced FunctionSchema with API-first design
  - [ ] Full input/output parameter validation using core schema types
  - [ ] Error schema support with structured error handling
  - [ ] Function metadata with rich documentation support
  - [ ] JSON Schema generation aligned with OpenAPI 3.1
  - [ ] Example generation for function documentation

- [ ] **`builders/function.go`** - FunctionBuilder with fluent API
  - [ ] `.Input(name, schema)` - Add input parameter with schema
  - [ ] `.Output(schema)` - Set output schema
  - [ ] `.Error(schema)` - Set error schema  
  - [ ] `.Required(params...)` - Mark parameters as required
  - [ ] `.Description()`, `.Name()`, `.Tags()` - Metadata builders
  - [ ] `.Example(input, output)` - Add usage examples

#### ServiceSchema Implementation ⚠️ **NEW ADDITION**
- [ ] **`schemas/service.go`** - ServiceSchema for service contract validation
  - [ ] Service method discovery and validation
  - [ ] Method signature consistency checking
  - [ ] Service-level metadata and documentation  
  - [ ] Service versioning support
  - [ ] Integration with FunctionSchema for method validation

- [ ] **`builders/service.go`** - ServiceBuilder with fluent API
  - [ ] `.Method(name, functionSchema)` - Add service method
  - [ ] `.FromStruct(instance)` - Generate from struct with reflection
  - [ ] `.Version(version)` - Set service API version
  - [ ] `.Description()`, `.Tags()` - Service-level metadata
  - [ ] `.Endpoint(path, method)` - HTTP endpoint mapping

### Phase 3: Function Registry Integration ⚠️ **NEW PHASE**

#### Registry System
- [ ] **`registry/function_registry.go`** - Core function registry
  - [ ] `api.FunctionRegistry` interface definition
  - [ ] Named function storage with address-based access
  - [ ] Function lifecycle management (register, execute, remove)
  - [ ] Thread-safe concurrent access
  - [ ] Function discovery and listing

- [ ] **`registry/service_registry.go`** - Service registry system
  - [ ] Service discovery and registration  
  - [ ] Method-level function registration
  - [ ] Service instance lifecycle management
  - [ ] Service metadata aggregation

#### Portal System Integration
- [ ] **`portals/base.go`** - Portal interface and base implementation
  - [ ] `api.Portal[T]` interface aligned with API design
  - [ ] Address generation strategy
  - [ ] Function transformation and wrapping
  - [ ] Error handling and result transformation

- [ ] **`portals/local.go`** - Local in-process execution portal
- [ ] **`portals/http.go`** - HTTP-based function execution portal  
- [ ] **`portals/websocket.go`** - WebSocket function execution portal

### Phase 4: Service Reflection System ⚠️ **NEW PHASE**

#### Service Introspection
- [ ] **`reflection/service_reflector.go`** - API-first service reflection
  - [ ] Struct method discovery with filtering
  - [ ] Automatic function schema generation from methods
  - [ ] Service contract validation
  - [ ] Performance-optimized method binding

- [ ] **`reflection/method_binding.go`** - Method-to-function conversion
  - [ ] Type-safe method binding with proper receiver handling
  - [ ] Context propagation for service calls
  - [ ] Error handling and panic recovery
  - [ ] Input/output transformation

#### Service Analysis
- [ ] **`analysis/service_analyzer.go`** - Service contract analysis
  - [ ] Service consistency validation
  - [ ] Breaking change detection
  - [ ] API compatibility checking
  - [ ] Service dependency analysis

### Phase 5: Advanced Function Features ⚠️ **NEW PHASE**

#### Function Composition
- [ ] **`composition/pipeline.go`** - Function pipeline composition
  - [ ] Type-safe function chaining
  - [ ] Pipeline schema validation
  - [ ] Error propagation handling
  - [ ] Parallel execution support

#### Function Validation
- [ ] **`validation/function_validator.go`** - Function contract validation
  - [ ] Input parameter validation using core schemas
  - [ ] Output validation with type checking
  - [ ] Function signature compatibility checking
  - [ ] Runtime validation integration

#### Function Middleware
- [ ] **`middleware/function_middleware.go`** - Function execution middleware
  - [ ] Authentication and authorization
  - [ ] Logging and metrics collection
  - [ ] Rate limiting and throttling
  - [ ] Caching and memoization

---

## 🏗️ Implementation Strategy

### 1. API-First Design Principles

```go
// Define clean interfaces in schema/api
type FunctionRegistry interface {
    Register(name string, schema api.FunctionSchema, handler any) (string, error)
    GetFunction(name string) (api.Function, error)
    ListFunctions() []string
}

type ServiceRegistry interface {
    RegisterService(name string, service any) (*ServiceInfo, error)
    GetService(name string) (api.Service, error)
    CallMethod(serviceName, methodName string, ctx context.Context, input any) (any, error)
}

type Portal[T any] interface {
    Apply(address string, schema api.FunctionSchema, implementation T) api.Function
    GenerateAddress(name string, implementation T) string
}
```

### 2. Backward Compatibility Strategy

```go
// Adapter pattern for legacy compatibility
func AdaptLegacyFunction(legacyFunc schema.Function) api.Function {
    return &FunctionAdapter{legacy: legacyFunc}
}

func AdaptLegacyService(legacyService any) api.Service {
    return &ServiceAdapter{instance: legacyService}
}
```

### 3. Migration Path

#### Phase 1: Core Types (Current)
- String, Number, Integer, Boolean, Array schemas

#### Phase 2: Function & Service Integration
- Function and Service schemas with full feature parity

#### Phase 3: Registry & Portal System  
- Advanced function execution and management

#### Phase 4: Reflection & Analysis
- Automatic schema generation and validation

---

## 🎯 Success Metrics

### Feature Parity
- [ ] **100%** of existing function features available in core
- [ ] **100%** of existing service features available in core  
- [ ] **100%** of portal types supported (HTTP, WebSocket, Local, etc.)
- [ ] **Zero regression** in functionality during migration

### Performance Targets
- [ ] **Function call overhead** ≤ 10% vs direct function calls
- [ ] **Service method binding** ≤ 50μs per method
- [ ] **Registry lookup** ≤ 1μs per function
- [ ] **Schema validation** ≤ 100μs for typical function signatures

### API Quality
- [ ] **Type-safe** function and service registration
- [ ] **Composable** function pipeline support
- [ ] **Extensible** portal system for new transports
- [ ] **Observable** function execution with metrics

---

## 📝 Integration Examples

### Function Schema Definition
```go
// Create a function schema using core package
userCreateFunc := core.NewFunction().
    Name("CreateUser").
    Description("Creates a new user account").
    Input("name", core.NewString().MinLength(1).MaxLength(100).Build()).
    Input("email", core.NewString().Email().Build()).
    Input("age", core.NewInteger().Min(13).Max(120).Optional().Build()).
    Output(core.NewObject().
        Property("id", core.NewInteger().Positive().Build()).
        Property("name", core.NewString().Build()).
        Property("email", core.NewString().Build()).
        Build()).
    Error(core.NewObject().
        Property("code", core.NewString().Build()).
        Property("message", core.NewString().Build()).
        Build()).
    Build()
```

### Service Schema Definition
```go
// Define a service schema from struct
type UserService struct {
    db Database
}

func (s *UserService) CreateUser(ctx context.Context, name, email string, age *int) (*User, error) {
    // Implementation
}

// Generate service schema
userService := &UserService{db: database}
serviceSchema := core.NewService().
    FromStruct(userService).
    Name("UserService").
    Version("v1.0.0").
    Description("User management service").
    Build()
```

### Registry Integration
```go
// Register functions in registry
registry := core.NewFunctionRegistry()
address, err := registry.Register("user.create", userCreateFunc, userCreateHandler)

// Register service with automatic method registration  
serviceRegistry := core.NewServiceRegistry()
serviceInfo, err := serviceRegistry.RegisterService("user", userService)

// Call service method through registry
result, err := serviceRegistry.CallMethod("user", "CreateUser", ctx, params)
```

---

## 🚦 Priority Assessment

### High Priority (Immediate - Next Sprint)
1. **FunctionSchema Enhancement** - Bring function schemas to feature parity
2. **ServiceSchema Implementation** - Add missing service schema support
3. **Basic Registry** - Core function registration and lookup

### Medium Priority (Next Month)  
1. **Portal System** - Multi-transport function execution
2. **Service Reflection** - Automatic service schema generation
3. **Function Composition** - Pipeline and chaining support

### Low Priority (Future Releases)
1. **Advanced Middleware** - Authentication, caching, rate limiting
2. **Function Analytics** - Performance monitoring and optimization
3. **Distributed Execution** - Cross-service function calls

---

## 🔗 Related Documents

- [`TODOS.md`](./TODOS.md) - Main implementation roadmap (needs updating)
- [`README.md`](./README.md) - Package overview (needs service section)
- [`schema/api/`](../api/) - API interface definitions
- [`schema/functions/`](../functions/) - Current function system implementation
- [`schema/reflect_service.go`](../reflect_service.go) - Current service reflection

---

**Author**: Schema Core Team  
**Created**: January 2025  
**Status**: Proposal - Awaiting Implementation  
**Priority**: High - Critical for feature parity and migration success 