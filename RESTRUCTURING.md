# Schema Package Restructuring

## 🎯 Overview

This document outlines a proposed restructuring of the schema package to improve organization and reduce cognitive complexity. The current structure mixes different concerns, making it harder to understand the system's architecture at a glance.

## 🔍 Current Structure Analysis

The current schema package contains various components with different purposes:

- **Type-dispatch processors** (visitors): `export/`
- **Capability-based processors** (consumers): `consumer/`
- **Schema builders**: `builders/`
- **Go type conversion**: `native/`
- **Execution infrastructure**: `portal/`, `registry/`
- **Core data structures**: `api/`, `schemas/`, `annotation/`

## 💡 Proposed Structure

Reorganize around **fundamental schema operations**:

```
schema/
├── core/          # Fundamental data structures and types
├── visit/         # Type-dispatch processing (visitor pattern)
├── consume/       # Capability-based processing (consumer pattern)
├── construct/     # Schema creation and building
├── runtime/       # Execution infrastructure
└── api/          # Public interfaces and contracts
```

## 📁 Detailed Directory Mapping

### `core/` - Fundamental Data Structures
**Purpose**: Core schema types, annotations, and basic abstractions

**Current → New**:
- `schemas/` → `core/schemas/`
- `annotation/` → `core/annotation/`
- `api/core/` → `core/types/`

**Contents**:
- Schema type definitions (`StringSchema`, `ObjectSchema`, etc.)
- Annotation system
- Core interfaces and types
- Basic value types

### `visit/` - Type-Dispatch Processing
**Purpose**: Systematic processing of schemas using the visitor pattern

**Current → New**:
- `export/` → `visit/export/`

**Contents**:
- `export/` - Code generation (TypeScript, Go, JSON, Python generators)
- Future: Analysis visitors, transformation visitors, documentation visitors

**Pattern**: Implements `SchemaVisitor` interface with `VisitString()`, `VisitObject()`, etc.

### `consume/` - Capability-Based Processing  
**Purpose**: Selective processing based on schema characteristics and consumer capabilities

**Current → New**:
- `consumer/` → `consume/validation/`

**Contents**:
- `validation/` - Schema and value validation
- Future: Formatting consumers, documentation consumers, API consumers

**Pattern**: Implements `AnnotationConsumer` interface with `ApplicableSchemas()` filtering

### `construct/` - Schema Creation and Building
**Purpose**: Creating schemas from various input sources

**Current → New**:
- `builders/` → `construct/builders/`
- `native/` → `construct/native/`

**Contents**:
- `builders/` - Fluent schema builders (`NewStringSchema()`, etc.)
- `native/` - Go type reflection and struct tag parsing
- Future: Parser-based construction, API-based construction, import/export

**Pattern**: Various factory and builder patterns for schema creation

### `runtime/` - Execution Infrastructure
**Purpose**: Components needed for system execution and management

**Current → New**:
- `portal/` → `runtime/portals/`
- `registry/` → `runtime/registry/`

**Contents**:
- `portals/` - Function execution transport (HTTP, WebSocket, local)
- `registry/` - Component storage and discovery
- Future: Caching, monitoring, configuration management

**Pattern**: Infrastructure and plumbing for the schema system

### `api/` - Public Interfaces
**Purpose**: Clean public contracts and interfaces

**Current → New**:
- `api/` → `api/` (cleaned up)

**Contents**:
- Public interfaces for all major components
- Stable contracts for external consumers
- Version-stable APIs

## 🔄 Migration Strategy

### Phase 1: Create New Structure
1. Create new directory structure
2. Move files to new locations
3. Update import paths
4. Update documentation

### Phase 2: Consolidate Interfaces
1. Clean up `api/` package
2. Ensure consistent interfaces across operations
3. Remove duplicate or conflicting interfaces

### Phase 3: Validate and Test
1. Ensure all tests pass
2. Validate that examples still work
3. Update any external documentation

## 🎯 Benefits of New Structure

### **Conceptual Clarity**
Each directory represents a **fundamental operation** rather than mixed concerns:
- **`visit/`** = "I systematically process schema structures"
- **`consume/`** = "I selectively handle schemas I can work with"  
- **`construct/`** = "I create schemas from various inputs"
- **`runtime/`** = "I make the system work at execution time"

### **Reduced Cognitive Load**
- Related functionality is grouped together
- Clear separation of concerns
- Easier to find relevant code
- Simpler mental model

### **Better Extensibility**
- Clear place for new visitors (`visit/`)
- Clear place for new consumers (`consume/`)
- Clear place for new construction methods (`construct/`)
- Infrastructure concerns isolated (`runtime/`)

### **Improved Navigation**
- Developers can quickly find the right directory for their use case
- Less confusion about where functionality belongs
- Clearer dependency relationships

## 📋 Implementation Checklist

- [ ] Create new directory structure
- [ ] Move `export/` → `visit/export/`
- [ ] Move `consumer/` → `consume/validation/`
- [ ] Move `builders/` → `construct/builders/`
- [ ] Move `native/` → `construct/native/`
- [ ] Move `portal/` → `runtime/portals/`
- [ ] Move `registry/` → `runtime/registry/`
- [ ] Reorganize `core/` with schemas, annotations, types
- [ ] Clean up `api/` package
- [ ] Update all import statements
- [ ] Update documentation and examples
- [ ] Run full test suite
- [ ] Update README files in each directory

## 🚀 Future Additions

With this structure, future additions have clear homes:

### `visit/`
- Analysis visitors for schema introspection
- Transformation visitors for schema modification
- Documentation visitors for generating docs

### `consume/`
- Formatting consumers for data presentation
- API consumers for REST/GraphQL generation
- Documentation consumers for spec generation

### `construct/`
- Parser-based construction from JSON Schema, OpenAPI
- Database schema construction from SQL DDL
- API-based construction from external sources

### `runtime/`
- Caching systems for performance
- Monitoring and metrics collection
- Configuration management 