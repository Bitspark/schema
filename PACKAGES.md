# Schema System Package Overview

This document provides an overview of all packages in the schema system, their purposes, responsibilities, and dependencies.

## 📦 Package Descriptions

### `schema/api` - Core Interfaces
**Purpose**: Defines all core interfaces and types for the schema system  
**Responsibilities**:
- Core schema interfaces (`Schema`, `StringSchema`, `ObjectSchema`, etc.)
- Annotation system interfaces (`Annotation`, `AnnotationRegistry`, etc.)
- Function and service interfaces (`FunctionSchema`, `ServiceSchema`, etc.)
- Validation result types and error structures
- Builder interfaces for fluent API construction

**Key Files**:
- `api/core/types.go` - Base schema interfaces and types
- `api/core/schemas.go` - Specific schema type interfaces
- `api/annotation.go` - Annotation system interfaces
- `api/function.go` - Function-related interfaces
- `api/service.go` - Service-related interfaces

**Dependencies**: None (pure interfaces)

---

### `schema/schemas` - Schema Implementations
**Purpose**: Concrete implementations of all schema types  
**Responsibilities**:
- Implement all schema interfaces from `api/core`
- Provide validation logic for each schema type
- Support JSON Schema generation and example generation
- Handle schema cloning and metadata management
- Integrate with annotation system for enhanced validation

**Key Files**:
- `schemas/string.go` - String schema with format validation
- `schemas/object.go` - Object schema with property validation
- `schemas/array.go` - Array schema with item validation
- `schemas/function.go` - Function schema with argument validation
- `schemas/service.go` - Service schema with method validation

**Dependencies**: `api/core`, `api` (for annotations)

---

### `schema/builders` - Fluent Builders
**Purpose**: Fluent builder pattern for constructing schemas  
**Responsibilities**:
- Implement builder interfaces from `api/core`
- Provide immutable, chainable API for schema construction
- Offer domain-specific helper methods (`.Email()`, `.UUID()`, etc.)
- Support metadata and constraint configuration
- Create configured schema instances

**Key Files**:
- `builders/string.go` - String schema builder
- `builders/object.go` - Object schema builder
- `builders/array.go` - Array schema builder
- `builders/function.go` - Function schema builder
- `builders/service.go` - Service schema builder

**Dependencies**: `api/core`, `schemas`

---

### `schema/annotation` - Annotation System
**Purpose**: Implementation of the annotation system for schema metadata  
**Responsibilities**:
- Implement annotation interfaces from `api`
- Provide annotation registry for type management
- Support annotation validation and metadata
- Enable flexible, typed metadata attachment to schemas
- Re-export API interfaces for convenience

**Key Files**:
- `annotation/types.go` - Re-exports of API interfaces
- `annotation/registry.go` - Annotation registry implementation
- `annotation/annotation_test.go` - Comprehensive tests

**Dependencies**: `api`, `schemas` (for testing)

---

### `schema/registry` - Validator Registry
**Purpose**: Pluggable validation system with built-in validators  
**Responsibilities**:
- Implement validator registry for extensible validation
- Provide built-in validators (email, URL, UUID, pattern, etc.)
- Support annotation-based validation configuration
- Manage validator lifecycle and metadata
- Enable custom validator registration

**Key Files**:
- `registry/types.go` - Validator interfaces and types
- `registry/registry.go` - Main validator registry
- `registry/validators.go` - Built-in validator implementations
- `registry/function_registry.go` - Function registration and execution

**Dependencies**: `annotation`, `schemas`

---

### `schema/native` - Go Type Integration
**Purpose**: Bridge between Go types and schema system  
**Responsibilities**:
- Convert Go types to schemas with annotation support
- Parse struct tags into annotations
- Discover services and functions from Go types
- Support automatic schema generation from reflection
- Enable Go-first schema development

**Key Files**:
- `native/types.go` - Type conversion interfaces
- `native/converter.go` - Go type to schema converter
- `native/tag_parser.go` - Struct tag to annotation parser
- `native/converter_test.go` - Comprehensive tests

**Dependencies**: `annotation`, `schemas`, `registry`

---

### `schema/engine` - Schema Engine
**Purpose**: Central coordination layer for the schema system  
**Responsibilities**:
- Manage named schema registration and resolution
- Support schema references with namespacing and versioning
- Provide pluggable schema type system
- Coordinate annotation validation
- Enable schema caching and circular dependency detection

**Key Files**:
- `engine/engine.go` - Core engine interfaces
- `engine/impl.go` - Engine implementation
- `engine/references.go` - Schema reference system
- `engine/annotations.go` - Annotation integration

**Dependencies**: `api/core`

---

### `schema/portal` - Service Portal System
**Purpose**: HTTP/WebSocket service exposure and consumption  
**Responsibilities**:
- Expose functions and services over HTTP/WebSocket
- Provide local, HTTP, and WebSocket portal implementations
- Support automatic API generation from schemas
- Handle request/response validation and serialization
- Enable service discovery and registration

**Key Files**:
- `portal/factory.go` - Portal factory implementations
- `portal/http.go` - HTTP portal implementation
- `portal/websocket.go` - WebSocket portal implementation
- `portal/local.go` - Local portal implementation

**Dependencies**: `api`, `registry`

---

### `schema/export` - Code Generation
**Purpose**: Generate code and schemas from schema definitions  
**Responsibilities**:
- Export schemas to various formats (JSON Schema, OpenAPI, etc.)
- Generate type-safe code in multiple languages
- Support template-based code generation
- Enable schema-driven development workflows
- Provide extensible export system

**Key Subdirectories**:
- `export/json/` - JSON Schema export
- `export/golang/` - Go code generation
- `export/python/` - Python code generation
- `export/typescript/` - TypeScript code generation

**Dependencies**: `api/core`, `schemas`

---

### `schema/tests` - Integration Tests
**Purpose**: End-to-end testing of the schema system  
**Responsibilities**:
- Integration testing across all packages
- Performance benchmarking
- Real-world usage scenario testing
- Cross-package compatibility verification
- Regression testing

**Dependencies**: All schema packages

---

## 🔗 Dependency Graph

```
# Core API Layer (No Dependencies)
`schema/api` -> (none)

# Schema Implementations
`schema/schemas` -> `schema/api`

# Builder Layer
`schema/builders` -> `schema/api`, `schema/schemas`

# Annotation System
`schema/annotation` -> `schema/api`, `schema/schemas`

# Registry System
`schema/registry` -> `schema/annotation`, `schema/schemas`

# Native Go Integration
`schema/native` -> `schema/annotation`, `schema/schemas`, `schema/registry`

# Engine Coordination
`schema/engine` -> `schema/api`

# Portal System
`schema/portal` -> `schema/api`, `schema/registry`

# Code Generation
`schema/export` -> `schema/api`, `schema/schemas`
`schema/export/json` -> `schema/api`, `schema/schemas`
`schema/export/golang` -> `schema/api`, `schema/schemas`
`schema/export/python` -> `schema/api`, `schema/schemas`
`schema/export/typescript` -> `schema/api`, `schema/schemas`

# Integration Testing
`schema/tests` -> (all packages)
```

## 🏗️ Architecture Layers

The schema system is organized into clear architectural layers:

### **Layer 1: API Interfaces**
- `schema/api` - Pure interfaces, no implementations

### **Layer 2: Core Implementations**
- `schema/schemas` - Schema implementations
- `schema/engine` - Engine coordination

### **Layer 3: Builder & Annotation**
- `schema/builders` - Fluent construction API
- `schema/annotation` - Annotation system implementation

### **Layer 4: Advanced Features**
- `schema/registry` - Validator registry
- `schema/native` - Go type integration
- `schema/portal` - Service portal system
- `schema/export` - Code generation

### **Layer 5: Testing**
- `schema/tests` - Integration testing

## 🎯 Design Principles

### **Separation of Concerns**
Each package has a single, well-defined responsibility with minimal overlap.

### **Dependency Inversion**
Higher-level packages depend on abstractions (interfaces) rather than concrete implementations.

### **Extensibility**
Plugin architectures and interface-based design enable easy extension without modification.

### **Type Safety**
Compile-time interface verification ensures type safety across package boundaries.

### **Performance**
Efficient implementations with caching, pre-compilation, and minimal allocations.

### **Testability**
Clear interfaces and dependency injection enable comprehensive testing at all levels.

## 🚀 Integration Points

### **Builder → Schema**
Builders create configured schema instances using the schemas package.

### **Schema → Annotation**
Schemas support annotations for enhanced validation and metadata.

### **Registry → Annotation**
Validators are configured and discovered through the annotation system.

### **Native → All**
Go type conversion integrates with annotations, schemas, and registry.

### **Engine → API**
Engine coordinates schemas through clean API interfaces.

### **Portal → Registry**
Service portals use registries for function and service management.

### **Export → Schema**
Code generation reads schema definitions to produce output.

This architecture provides a solid foundation for schema-driven development with clear separation of concerns, strong type safety, and excellent extensibility. 