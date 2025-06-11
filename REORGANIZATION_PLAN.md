# Schema Package Reorganization Plan

This document outlines the specific steps to reorganize the schema package for better structure and maintainability.

## 🎯 Goals

- Move reflection-related files to `/schema/reflection`
- Move generation-related files to `/schema/generator`  
- Move visitor-related files to `/schema/visitor`
- Maintain backward compatibility
- Improve package organization and discoverability

## 📋 Current State Analysis

### Files to Reorganize

#### Reflection Files → `/schema/reflection/`
```
reflection.go                → reflection/reflection.go
reflection_funcs.go          → reflection/funcs.go
reflect_service.go           → reflection/service.go
reflection_test.go           → reflection/reflection_test.go
reflection_funcs_test.go     → reflection/funcs_test.go
reflection_advanced_test.go  → reflection/advanced_test.go
```

#### Generation Files → `/schema/generator/`
```
generator.go                 → generator/generator.go
schema_generator.go          → generator/schema.go
generator_test.go            → generator/generator_test.go
schema_generator_test.go     → generator/schema_test.go
default_generator_test.go    → generator/default_test.go
```

#### Visitor Files → `/schema/visitor/`
```
visitor.go                   → visitor/visitor.go
visitor_test.go              → visitor/visitor_test.go
```

#### Files to Keep in Root
```
types.go                     # Core interfaces
function.go                  # Function schemas
function_types.go            # Function types (merge into function.go)
basic.go                     # Basic schema implementations
builder.go                   # Builder patterns
generics.go                  # Generic patterns
introspection_test.go        # Keep for core functionality
convenience_test.go          # Keep for core functionality
integration_test.go          # Keep for integration tests
```

## 🔄 Migration Steps

### Step 1: Create New Directory Structure

```bash
mkdir -p schema/reflection
mkdir -p schema/generator
mkdir -p schema/visitor
```

### Step 2: Move and Rename Files

#### Reflection Subsystem
```bash
# Move reflection files
mv schema/reflection.go schema/reflection/reflection.go
mv schema/reflection_funcs.go schema/reflection/funcs.go
mv schema/reflect_service.go schema/reflection/service.go

# Move reflection tests
mv schema/reflection_test.go schema/reflection/reflection_test.go
mv schema/reflection_funcs_test.go schema/reflection/funcs_test.go
mv schema/reflection_advanced_test.go schema/reflection/advanced_test.go
```

#### Generator Subsystem
```bash
# Move generator files
mv schema/generator.go schema/generator/generator.go
mv schema/schema_generator.go schema/generator/schema.go

# Move generator tests
mv schema/generator_test.go schema/generator/generator_test.go
mv schema/schema_generator_test.go schema/generator/schema_test.go
mv schema/default_generator_test.go schema/generator/default_test.go
```

#### Visitor Subsystem
```bash
# Move visitor files
mv schema/visitor.go schema/visitor/visitor.go
mv schema/visitor_test.go schema/visitor/visitor_test.go
```

### Step 3: Update Package Declarations

#### In `/schema/reflection/` files:
```go
// Change from:
package schema

// To:
package reflection
```

#### In `/schema/generator/` files:
```go
// Change from:
package schema

// To:
package generator
```

#### In `/schema/visitor/` files:
```go
// Change from:
package schema

// To:
package visitor
```

### Step 4: Update Imports and Dependencies

#### Update imports in moved files:
```go
// In reflection/reflection.go
import (
    "defs.dev/schema"  // Import parent package for core types
    // ... other imports
)

// In generator/generator.go  
import (
    "defs.dev/schema"
    "defs.dev/schema/reflection"  // Import reflection if needed
    // ... other imports
)
```

#### Update type references:
```go
// Change from:
func FromStruct[T any]() Schema { ... }

// To:
func FromStruct[T any]() schema.Schema { ... }
```

### Step 5: Create Backward Compatibility Layer

Create compatibility wrappers in the root package:

#### `/schema/reflection_compat.go`:
```go
package schema

import "defs.dev/schema/reflection"

// Deprecated: Use reflection.FromStruct instead.
// This function will be removed in a future version.
func FromStruct[T any]() Schema {
    return reflection.FromStruct[T]()
}

// Deprecated: Use reflection.FromType instead.
func FromType(typ reflect.Type) Schema {
    return reflection.FromType(typ)
}
```

#### `/schema/generator_compat.go`:
```go
package schema

import "defs.dev/schema/generator"

// Deprecated: Use generator.New instead.
func NewGenerator(options ...GeneratorOption) *Generator {
    return generator.New(options...)
}
```

### Step 6: Update Public APIs

#### New Public APIs:

```go
// schema/reflection/api.go
package reflection

// Public API for reflection subsystem
func FromStruct[T any]() schema.Schema { ... }
func FromType(typ reflect.Type) schema.Schema { ... }
func RegisterTypeMapping(typ reflect.Type, factory func() schema.Schema) { ... }

// schema/generator/api.go  
package generator

// Public API for generator subsystem
func New(options ...Option) *Generator { ... }
func GenerateJavaScript(s schema.Schema) ([]byte, error) { ... }
func GenerateTypeScript(s schema.Schema) ([]byte, error) { ... }
func GenerateOpenAPI(s schema.Schema) (map[string]any, error) { ... }

// schema/visitor/api.go
package visitor

// Public API for visitor subsystem
type Visitor interface { ... }
func Walk(schema schema.Schema, visitor Visitor) error { ... }
```

### Step 7: Update Tests

#### Update test imports:
```go
// In reflection tests
package reflection

import (
    "testing"
    "defs.dev/schema"
)

// In generator tests
package generator

import (
    "testing"
    "defs.dev/schema"
    "defs.dev/schema/reflection"
)
```

### Step 8: Update Documentation

#### Update import examples in documentation:
```go
// Before:
import "defs.dev/schema"
userSchema := schema.FromStruct[User]()

// After (new way):
import (
    "defs.dev/schema"
    "defs.dev/schema/reflection"
)
userSchema := reflection.FromStruct[User]()

// After (backward compatible):
import "defs.dev/schema"
userSchema := schema.FromStruct[User]() // Still works but deprecated
```

## 🧪 Testing the Migration

### Validation Steps

1. **Build Test**: Ensure all packages build successfully
   ```bash
   go build ./...
   ```

2. **Unit Tests**: Run all tests to ensure functionality is preserved
   ```bash
   go test ./...
   ```

3. **Integration Tests**: Verify end-to-end functionality
   ```bash
   go test ./... -tags=integration
   ```

4. **Backward Compatibility**: Test deprecated APIs still work
   ```bash
   # Create test using old API
   # Verify it still compiles and runs
   ```

### Expected Test Results

- ✅ All existing tests pass
- ✅ New package structure builds successfully  
- ✅ Backward compatibility maintained
- ✅ New APIs work as expected

## 📦 Final Structure

After migration, the structure will be:

```
schema/
├── types.go                    # Core interfaces
├── function.go                 # Function schemas  
├── basic.go                    # Basic implementations
├── builder.go                  # Builder patterns
├── generics.go                 # Generic patterns
├── *_compat.go                 # Backward compatibility
├── *_test.go                   # Core tests
│
├── reflection/                 # Reflection subsystem
│   ├── reflection.go           # Main API
│   ├── funcs.go               # Utilities
│   ├── service.go             # Service layer
│   └── *_test.go              # Tests
│
├── generator/                  # Generation subsystem
│   ├── generator.go           # Core engine
│   ├── schema.go              # Schema generation
│   └── *_test.go              # Tests
│
├── visitor/                    # Visitor subsystem
│   ├── visitor.go             # Visitor pattern
│   └── *_test.go              # Tests
│
├── registry/                   # Registry (existing)
└── functions/                  # Functions (existing)
```

## 🚀 Benefits After Migration

1. **Better Organization**: Related files grouped together
2. **Cleaner Root**: Fewer files in main package directory
3. **Easier Navigation**: Clear subsystem boundaries
4. **Better Testing**: Isolated test suites per subsystem
5. **Scalability**: Room for growth within each subsystem
6. **Maintainability**: Easier to understand and modify code

## 📅 Migration Timeline

**Phase 1** (Week 1): File moves and package updates
**Phase 2** (Week 2): Import updates and compatibility layer
**Phase 3** (Week 3): Testing and validation
**Phase 4** (Week 4): Documentation updates

## 🔧 Implementation Commands

Here's a script to execute the migration:

```bash
#!/bin/bash
# reorganize.sh

echo "Creating directory structure..."
mkdir -p schema/reflection schema/generator schema/visitor

echo "Moving reflection files..."
mv schema/reflection.go schema/reflection/reflection.go
mv schema/reflection_funcs.go schema/reflection/funcs.go
mv schema/reflect_service.go schema/reflection/service.go
mv schema/reflection_*test.go schema/reflection/

echo "Moving generator files..."
mv schema/generator.go schema/generator/generator.go
mv schema/schema_generator.go schema/generator/schema.go
mv schema/*generator*test.go schema/generator/

echo "Moving visitor files..."
mv schema/visitor.go schema/visitor/visitor.go
mv schema/visitor_test.go schema/visitor/

echo "Migration complete! Next: update package declarations and imports."
```

This reorganization will significantly improve the codebase structure while maintaining full backward compatibility. 