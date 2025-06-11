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

### Current Schema Package (Flattened Structure)
```
schema/
├── api/             # All interfaces
├── portal/          # Portal implementations
│   ├── http.go, websocket.go, local.go, testing.go
│   ├── address.go, function.go, registry.go
│   └── README.md, TODOS.md
├── registry/        # Registry implementations  
│   ├── function_registry.go, service_registry.go
│   └── factory.go
├── schemas/         # Schema implementations
│   ├── string.go, number.go, integer.go, boolean.go
│   ├── array.go, object.go, function.go, service.go
├── builders/        # Builder implementations
│   ├── string.go, number.go, integer.go, boolean.go
│   ├── array.go, object.go, function.go, service.go
├── examples/        # Usage examples
├── tests/           # Test suites
├── core/            # Empty (moved up)
├── core.go          # Factory functions
├── doc.go           # Package docs
├── README.md        # Documentation
├── TODOS.md         # Implementation status
└── go.mod           # Module definition
```

## Analysis

### What's Working Well
1. **Flattened structure** eliminates unnecessary nesting
2. **Domain-organized implementations** - portal/, registry/, schemas/, builders/
3. **Clear separation** between interfaces (api/) and implementations (top-level domains)
4. **Excellent documentation** with domain-specific README/TODOS files
5. **Direct access** to implementations without core/ indirection

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

### Unchanged Implementation Structure
```
schema/
├── api/             # Reorganized interfaces (above)
├── portal/          # Portal implementations (unchanged)
├── registry/        # Registry implementations (unchanged)
├── schemas/         # Schema implementations (unchanged)
├── builders/        # Builder implementations (unchanged)
├── examples/        # Usage examples (unchanged)
├── tests/           # Test suites (unchanged)
├── core.go          # Factory functions (unchanged)
├── doc.go           # Package docs (unchanged)
├── README.md        # Documentation (unchanged)
├── TODOS.md         # Implementation status (unchanged)
└── go.mod           # Module definition (unchanged)
```

## Benefits of This Approach

### 1. **Preserves New Flattened Structure**
- ✅ **Excellent domain organization** in implementations remains untouched
- ✅ **Direct access** to portal/, registry/, schemas/, builders/ continues
- ✅ **No disruption** to the well-organized implementation structure
- ✅ **Clear separation** between interfaces and implementations maintained

### 2. **Improves API Organization**
- ✅ **Groups tightly related core schema interfaces** that always change together
- ✅ **Reduces API root directory crowding** (12 → 9 files)
- ✅ **Clearer conceptual boundaries** between core schemas and higher-level systems
- ✅ **Room for growth** - component and topic systems have clear homes

### 3. **Maintains Consistency**
- ✅ **Matches implementation organization** - logical grouping where beneficial
- ✅ **Familiar flat structure** for main systems (function, service, portal, registry)
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

### Implementation Imports (Unchanged)
```go
import (
    "defs.dev/schema/portal"        # Direct access to implementations
    "defs.dev/schema/registry"      # No core/ indirection needed
    "defs.dev/schema/schemas"       # Clean, flat structure
    "defs.dev/schema/builders"      # Domain-organized
)
```

## Migration Strategy

### Phase 1: Reorganize Core Schema System
1. Create `api/core/` directory
2. Move `types.go`, `schemas.go`, `builder.go`, `visitor.go` → `api/core/`
3. Update imports in implementation packages (portal/, registry/, schemas/, builders/)
4. Update documentation

### Phase 2: Update Imports
1. Update all implementation package imports
2. Update examples and tests
3. Update external documentation

### Phase 3: Validate
1. Ensure all builds pass
2. Run full test suite
3. Update any missed references

## Why This Works Well With New Structure

### Perfect Complement to Flattened Implementation
The new flattened structure already solved the implementation organization:
- ✅ **portal/** - Complete portal system with excellent docs
- ✅ **registry/** - Clean registry implementations
- ✅ **schemas/** - All schema implementations in one place  
- ✅ **builders/** - All builder implementations in one place

### API Reorganization Completes the Picture
By organizing just the API interfaces, we get:
- ✅ **Complete system organization** - both interfaces and implementations well-structured
- ✅ **Minimal disruption** - only touches interface definitions
- ✅ **Maximum benefit** - addresses crowding without touching working implementations

## Alternative Considered: Leave API As-Is

We considered leaving the API package unchanged since implementations are now well-organized.

**Rejected because:**
- API crowding still exists (12 files)
- Core schema interfaces still scattered despite being tightly coupled
- Missed opportunity to complete the organizational improvement
- Logical grouping benefits are significant

## Conclusion

The proposed minimal API reorganization perfectly complements the excellent new flattened implementation structure:
- **Addresses remaining pain points** (API crowding, logical grouping)
- **Preserves all benefits** of the new flattened structure
- **Completes the organizational vision** - well-structured interfaces AND implementations
- **Minimal disruption** - only touches interface definitions

This provides the **best of both worlds**: clean domain organization where beneficial, with flat access where it makes sense. 