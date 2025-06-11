# Structure

This document provides a comprehensive guide to the code organization, architectural design, and structural patterns of the Schema library. It complements the conceptual overview in [CONCEPTS.md](CONCEPTS.md) and integration examples in [INTEGRATION.md](INTEGRATION.md).

## Table of Contents

1. [Package Architecture](#package-architecture)
2. [File Organization](#file-organization)
3. [Proposed Reorganization](#proposed-reorganization)
4. [Component Relationships](#component-relationships)
5. [Design Patterns](#design-patterns)
6. [Module Boundaries](#module-boundaries)
7. [Extension Points](#extension-points)
8. [Code Organization Principles](#code-organization-principles)
9. [Development Workflow](#development-workflow)

## Package Architecture

### 🏗️ Layered Architecture

The schema package follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────┐
│              Integration Layer              │
│  (/functions/http, /websocket, /javascript) │
├─────────────────────────────────────────────┤
│             Application Layer               │
│     (/functions/registry, /functions)       │
├─────────────────────────────────────────────┤
│              Feature Layer                  │
│  (/reflection, /generator, /visitor)        │
├─────────────────────────────────────────────┤
│             Schema Type Layer               │
│      (builder.go, basic.go)                 │
├─────────────────────────────────────────────┤
│             Foundation Layer                │
│         (types.go, function.go)             │
└─────────────────────────────────────────────┘
```

### Component Hierarchy

```
schema/
├── Core Types & Interfaces      (types.go, function.go)
├── Schema Implementations       (basic.go, builder.go)
├── Advanced Features           (/reflection, /generator, /visitor)
├── Schema Registry             (/registry)
├── Function System             (/functions)
└── Integration Modules         (/functions/*)
```

## File Organization

### 📁 Current Structure

The current flat structure has grown organically but could benefit from better organization:

```
schema/
├── types.go                    # Core interfaces
├── function.go                 # Function schemas
├── function_types.go           # Function types
├── basic.go                    # Basic schema implementations
├── builder.go                  # Builder patterns
├── generics.go                 # Generic patterns
├── reflection.go               # Reflection main
├── reflection_funcs.go         # Reflection utilities
├── reflect_service.go          # Reflection service
├── visitor.go                  # Visitor pattern
├── generator.go                # Generation engine
├── schema_generator.go         # Schema generation
├── convenience_test.go         # Convenience tests
├── *_test.go                   # Various test files
├── registry/                   # Registry subsystem
└── functions/                  # Functions subsystem
```

## Proposed Reorganization

### 🎯 Improved Structure

Here's a proposed reorganization that groups related functionality:

```
schema/
├── types.go                    # Core interfaces and types
├── function.go                 # Function schema core
├── basic.go                    # Basic schema implementations
├── builder.go                  # Builder pattern implementations
├── generics.go                 # Generic type patterns
├── 
├── reflection/                 # Reflection subsystem
│   ├── reflection.go           # Main reflection API
│   ├── funcs.go               # Reflection utilities
│   ├── service.go             # Reflection service
│   ├── analyzer.go            # Struct analysis
│   ├── tags.go                # Tag parsing
│   └── *_test.go              # Reflection tests
│
├── generator/                  # Generation subsystem
│   ├── generator.go           # Core generation engine
│   ├── schema.go              # Schema-specific generation
│   ├── javascript.go          # JavaScript generation
│   ├── typescript.go          # TypeScript generation
│   ├── openapi.go             # OpenAPI generation
│   └── *_test.go              # Generator tests
│
├── visitor/                    # Visitor pattern subsystem
│   ├── visitor.go             # Visitor interfaces
│   ├── transformer.go         # Schema transformers
│   ├── analyzer.go            # Schema analyzers
│   └── *_test.go              # Visitor tests
│
├── registry/                   # Schema registry (existing)
│   ├── registry.go
│   ├── builder.go
│   ├── param.go
│   ├── ref.go
│   ├── resolver.go
│   └── errors.go
│
└── functions/                  # Function system (existing)
    ├── registry.go
    ├── factory.go
    ├── consumer.go
    ├── portal.go
    ├── http/
    ├── websocket/
    ├── javascript/
    ├── local/
    └── testing/
```

### 🔄 Migration Benefits

**Improved Organization:**
- Related functionality grouped together
- Cleaner root directory
- Better discoverability
- Easier navigation

**Clearer Boundaries:**
- Each subsystem has its own namespace
- Reduced import pollution
- Better encapsulation
- Easier testing

**Scalability:**
- Room for growth within each subsystem
- Independent development
- Modular architecture
- Plugin-friendly design

### 📦 Subsystem Details

#### `/reflection` Subsystem
```go
// schema/reflection/reflection.go
package reflection

// Main API for struct-to-schema generation
func FromStruct[T any]() schema.Schema { ... }
func FromType(typ reflect.Type) schema.Schema { ... }

// schema/reflection/analyzer.go
type StructAnalyzer struct { ... }
func (a *StructAnalyzer) Analyze(typ reflect.Type) SchemaInfo { ... }

// schema/reflection/tags.go
type TagParser struct { ... }
func ParseSchemaTag(tag string) TagInfo { ... }
```

#### `/generator` Subsystem
```go
// schema/generator/generator.go
package generator

type Generator struct { ... }
func New(options Options) *Generator { ... }

// schema/generator/javascript.go
func GenerateJavaScript(schema schema.Schema) ([]byte, error) { ... }

// schema/generator/openapi.go
func GenerateOpenAPI(schema schema.Schema) (map[string]any, error) { ... }
```

#### `/visitor` Subsystem
```go
// schema/visitor/visitor.go
package visitor

type Visitor interface {
    VisitString(*schema.StringSchema) error
    VisitObject(*schema.ObjectSchema) error
    // ... other visit methods
}

// schema/visitor/transformer.go
type Transformer struct { ... }
func (t *Transformer) Transform(schema schema.Schema) (schema.Schema, error) { ... }
```

### 🔧 Import Changes

**Before:**
```go
import "defs.dev/schema"

// Everything accessed through schema package
userSchema := schema.FromStruct[User]()
generator := schema.NewGenerator()
```

**After:**
```go
import (
    "defs.dev/schema"
    "defs.dev/schema/reflection"
    "defs.dev/schema/generator"
)

// Cleaner, more explicit imports
userSchema := reflection.FromStruct[User]()
gen := generator.New()
```

### 📋 Backward Compatibility

To maintain backward compatibility during transition:

```go
// schema/reflection.go (compatibility layer)
package schema

import "defs.dev/schema/reflection"

// Deprecated: Use reflection.FromStruct instead
func FromStruct[T any]() Schema {
    return reflection.FromStruct[T]()
}
```

## Component Relationships

### 🔗 Improved Dependency Graph

```mermaid
graph TD
    A[types.go] --> B[basic.go]
    A --> C[builder.go]
    A --> D[function.go]
    
    A --> E[/reflection]
    E --> F[reflection/analyzer.go]
    E --> G[reflection/tags.go]
    
    A --> H[/generator]
    H --> I[generator/javascript.go]
    H --> J[generator/openapi.go]
    
    A --> K[/visitor]
    K --> L[visitor/transformer.go]
    
    A --> M[/registry]
    A --> N[/functions]
    
    E --> H
    K --> H
```

### Subsystem Relationships

#### Core Dependencies
```go
// Foundation layer - no dependencies
types.go
function.go

// Schema layer - depends on foundation
basic.go    ──> types.go
builder.go  ──> types.go, basic.go

// Feature layer - depends on core
/reflection ──> types.go, basic.go
/generator  ──> types.go, /reflection
/visitor    ──> types.go
```

#### Integration Dependencies
```go
// Application layer
/registry   ──> types.go, /reflection
/functions  ──> types.go, function.go, /registry

// Integration layer
/functions/http       ──> /functions
/functions/websocket  ──> /functions
/functions/javascript ──> /functions, /generator
```

## Design Patterns

### 🏗️ Modular Architecture Pattern

**Implementation:**
```go
// Each subsystem has its own package with clear interfaces
package reflection

// Public API
type API interface {
    FromStruct[T any]() schema.Schema
    FromType(reflect.Type) schema.Schema
}

// Internal implementation
type analyzer struct { ... }
type tagParser struct { ... }
```

**Benefits:**
- Clear subsystem boundaries
- Independent development
- Testable components
- Plugin architecture support

### 📦 Package Organization Pattern

**Subsystem Structure:**
```
subsystem/
├── api.go          # Public interfaces
├── impl.go         # Main implementation
├── internal.go     # Internal utilities
├── types.go        # Subsystem-specific types
└── *_test.go       # Tests
```

**Benefits:**
- Consistent organization
- Clear public/private boundaries
- Easy to understand
- Scalable structure

## Module Boundaries

### 🎯 Updated Module Structure

#### Core Module (`/schema`)
**Exports:**
- Core `Schema` interface
- Basic schema types and builders
- Function schema support

**Files:**
- `types.go`, `function.go`, `basic.go`, `builder.go`, `generics.go`

#### Reflection Module (`/schema/reflection`)
**Exports:**
- `FromStruct[T]()` function
- Struct analysis utilities
- Tag parsing functions

**Dependencies:**
- Core schema module only

#### Generator Module (`/schema/generator`)
**Exports:**
- Code generation for multiple targets
- Template processing utilities

**Dependencies:**
- Core schema module
- Reflection module

#### Registry Module (`/schema/registry`)
**Exports:**
- Schema registry and references
- Parameterized schemas

**Dependencies:**
- Core schema module
- Reflection module (for auto-registration)

#### Functions Module (`/schema/functions`)
**Exports:**
- Function registry and execution
- Integration frameworks

**Dependencies:**
- Core schema module
- Registry module
- Generator module (for client generation)

## Extension Points

### 🔧 Subsystem Extension

**Adding New Subsystems:**
```go
// Create new subsystem directory
mkdir schema/newsystem

// Define public API
// schema/newsystem/api.go
package newsystem

type API interface {
    NewFeature() error
}

// Implement functionality
// schema/newsystem/impl.go
type implementation struct { ... }
```

### 🎯 Feature Extension

**Within Existing Subsystems:**
```go
// Add new generator target
// schema/generator/rust.go
package generator

func GenerateRust(schema schema.Schema) ([]byte, error) {
    // Rust code generation
}

// Register with main generator
func init() {
    RegisterTarget("rust", GenerateRust)
}
```

## Code Organization Principles

### 🎯 Subsystem Principles

1. **Single Responsibility**: Each subsystem has one clear purpose
2. **Interface Segregation**: Small, focused public APIs
3. **Dependency Inversion**: Depend on interfaces, not implementations
4. **Open/Closed**: Easy to extend, hard to break

### 📦 File Organization Principles

1. **Consistent Naming**: Similar file names across subsystems
2. **Clear Boundaries**: Public vs internal interfaces
3. **Logical Grouping**: Related functionality together
4. **Minimal Dependencies**: Reduce coupling between subsystems

### 🔄 Migration Strategy

#### Phase 1: Create New Structure
```bash
# Create new directories
mkdir -p schema/reflection
mkdir -p schema/generator  
mkdir -p schema/visitor

# Move files to new locations
mv reflection*.go schema/reflection/
mv generator*.go schema/generator/
mv visitor*.go schema/visitor/
```

#### Phase 2: Update Package Declarations
```go
// Update package declarations in moved files
package reflection  // was: package schema
package generator   // was: package schema
package visitor     // was: package schema
```

#### Phase 3: Add Compatibility Layer
```go
// schema/compatibility.go
package schema

import (
    "defs.dev/schema/reflection"
    "defs.dev/schema/generator"
)

// Deprecated APIs for backward compatibility
func FromStruct[T any]() Schema {
    return reflection.FromStruct[T]()
}
```

#### Phase 4: Update Documentation
- Update import examples
- Add migration guide
- Update README and other docs

## Development Workflow

### 🧪 Improved Testing Structure

```
schema/
├── *_test.go                   # Core functionality tests
├── integration_test.go         # End-to-end tests
├── reflection/
│   └── *_test.go              # Reflection-specific tests
├── generator/
│   └── *_test.go              # Generator tests
├── visitor/
│   └── *_test.go              # Visitor tests
└── functions/
    └── testing/               # Testing utilities
```

### 🔧 Build Commands

```bash
# Test specific subsystems
go test ./reflection/...
go test ./generator/...
go test ./visitor/...

# Test everything
go test ./...

# Build specific subsystems
go build ./reflection
go build ./generator
```

---

This reorganization would significantly improve the codebase structure while maintaining backward compatibility. The modular approach makes the codebase more maintainable, testable, and extensible. Combined with [CONCEPTS.md](CONCEPTS.md) for conceptual understanding and [INTEGRATION.md](INTEGRATION.md) for practical usage, developers would have complete visibility into a well-organized system. 