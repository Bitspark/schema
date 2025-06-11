# API Package Reorganization Plan

This document outlines the proposed reorganization of the `schema/api` package to improve maintainability while preserving the current structure's benefits.

## Current Structure Assessment

### Current API Package (12 files)
```
schema/api/
├── function.go      # Function, FunctionData interfaces
├── portal.go        # All portal interfaces (5.6KB, 190 lines)
├── registry.go      # Registry, Factory, Consumer interfaces
├── schemas.go       # All schema interfaces (2.5KB, 129 lines)
├── builder.go       # All builder interfaces (4.8KB, 158 lines)
├── types.go         # Core types, ValidationResult (2.3KB, 75 lines)
├── service.go       # Service interface
├── generics.go      # Generic type interfaces (1.8KB, 75 lines)
├── visitor.go       # SchemaVisitor, Accepter (673B, 22 lines)
├── compat.go        # Legacy compatibility (2.5KB, 120 lines)
├── component.go     # Component placeholders (126B, 11 lines)
├── topic.go         # Topic placeholders (13B, 2 lines)
└── doc.go           # Package documentation (2.6KB, 87 lines)
```

### Current Core Package (Well-Organized)
```
schema/core/
├── core.go          # Factory functions
├── portal/          # Portal implementations
├── registry/        # Registry implementations  
├── schemas/         # Schema implementations
├── builders/        # Builder implementations
├── examples/        # Usage examples
├── tests/           # Test suites
├── doc.go           # Package docs
├── README.md        # Documentation
└── TODOS.md         # Implementation status
```

## Analysis

### What's Working Well
1. **Core package structure** is excellent - domain-organized implementations
2. **Flat API structure** is familiar and easy to import
3. **Clear separation** between interfaces (api) and implementations (core)
4. **Good documentation** with README, TODOS, and examples

### Current Issues
1. **API directory crowding** - 12 files with mixed concerns
2. **Core schema files scattered** - types.go, schemas.go, builder.go, visitor.go are tightly related but separate
3. **Large individual files** - portal.go (190 lines), builder.go (158 lines)
4. **Logical grouping could be clearer** - core schema system vs. higher-level systems

## Proposed Minimal Reorganization

### New API Structure
```
schema/api/
├── core/                    # Core schema system (tightly coupled)
│   ├── types.go            # Schema, SchemaType, ValidationResult, SchemaMetadata
│   ├── schemas.go          # StringSchema, NumberSchema, ArraySchema, etc.
│   ├── builder.go          # StringSchemaBuilder, NumberSchemaBuilder, etc.
│   └── visitor.go          # SchemaVisitor, Accepter
├── function.go             # Function, FunctionData, FunctionSchema, FunctionSchemaBuilder
├── service.go              # Service, ServiceSchema, ServiceSchemaBuilder
├── portal.go               # All portal interfaces (Address, FunctionPortal, HTTPPortal, etc.)
├── registry.go             # Registry, Factory, Consumer, Middleware
├── generics.go             # Generic type interfaces (ListBuilder[T], OptionalSchema[T], etc.)
├── compat.go               # Legacy compatibility interfaces
├── component.go            # Component system (future expansion)
├── topic.go                # Topic system (future expansion)
├── doc.go                  # Package documentation
├── TYPES.md                # Type system design patterns
└── REORGANIZATION.md       # This document
```

## Benefits of This Approach

### 1. **Preserves Current Benefits**
- ✅ **Familiar flat structure** for main systems (function, service, portal, registry)
- ✅ **Easy imports** - `import "defs.dev/schema/api"` still works
- ✅ **Mirrors core organization** - logical grouping without complexity
- ✅ **Minimal disruption** to existing code

### 2. **Improves Organization**
- ✅ **Groups tightly related core schema interfaces** that always change together
- ✅ **Reduces API root directory crowding** (12 → 9 files)
- ✅ **Clearer conceptual boundaries** between core schemas and higher-level systems
- ✅ **Room for growth** - component and topic systems have clear homes

### 3. **Maintains Consistency**
- ✅ **Matches core package principles** - domain organization where it makes sense
- ✅ **Logical grouping** - things that change together, stay together
- ✅ **Progressive enhancement** - start simple, add structure as needed

## Import Impact

### Before
```go
import "defs.dev/schema/api"

// Usage
var schema api.StringSchema
var builder api.StringSchemaBuilder
var function api.Function
var portal api.HTTPPortal
```

### After
```go
import (
    "defs.dev/schema/api"
    "defs.dev/schema/api/core"
)

// Usage  
var schema core.StringSchema       // Moved to core subpackage
var builder core.StringSchemaBuilder
var function api.Function          // Stays in main package
var portal api.HTTPPortal          // Stays in main package
```

## Migration Strategy

### Phase 1: Reorganize Core Schema System
1. Create `api/core/` directory
2. Move `types.go`, `schemas.go`, `builder.go`, `visitor.go` → `api/core/`
3. Update imports in core package implementations
4. Update documentation

### Phase 2: Update Imports
1. Update all `core/` package imports
2. Update examples and tests
3. Update external documentation

### Phase 3: Validate
1. Ensure all builds pass
2. Run full test suite
3. Update any missed references

## Alternative Considered: Full Domain Organization

We considered a full domain-based structure mirroring the core package:

```
schema/api/
├── core/       # Core schema system
├── function/   # Function system  
├── service/    # Service system
├── portal/     # Portal system
└── registry/   # Registry system
```

**Rejected because:**
- Too much disruption for current benefits
- Import complexity increases significantly
- Core package already provides this organization
- Current flat structure works well for main systems

## Conclusion

The proposed minimal reorganization:
- **Addresses current pain points** (crowding, logical grouping)
- **Preserves existing benefits** (familiarity, simplicity)
- **Aligns with core package philosophy** (domain organization where beneficial)
- **Provides foundation for future growth** (component, topic systems)

This strikes the right balance between **improvement** and **stability**. 