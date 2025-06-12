# Schema Export Refactoring Plan

## 🎯 Goal

Refactor schema generation logic from individual schema types into a dedicated `schema/export/` package using the visitor pattern. This will separate concerns, improve maintainability, and enable easy addition of new output formats.

## 📊 Current State Analysis

### Problems with Current Architecture
- **Violation of Single Responsibility**: Each schema type handles both validation AND generation
- **Tight Coupling**: Adding new output formats requires modifying every schema type
- **Code Duplication**: Similar generation patterns repeated across schema types
- **Poor Extensibility**: Hard to add new formats (TypeScript, Python, Go, OpenAPI, etc.)
- **Testing Complexity**: Generation logic mixed with validation logic

### Current Generation Methods
- ✅ `ToJSONSchema()` - Implemented in all schema types
- ❌ TypeScript generation - Not implemented
- ❌ Python type generation - Not implemented  
- ❌ Go type generation - Not implemented
- ❌ OpenAPI specification - Not implemented
- ❌ Protocol Buffers - Not implemented

### Existing Infrastructure
- ✅ `SchemaVisitor` interface exists in `api/core/visitor.go`
- ✅ All schemas implement `Accept(SchemaVisitor)` method
- ✅ `Accepter` interface defined and implemented
- ✅ Visitor pattern already used in tests

## 🏗️ Target Architecture

### Directory Structure
```
schema/
├── export/                      # New generation package
│   ├── PLAN.md                  # This file
│   ├── README.md                # Package documentation
│   ├── doc.go                   # Package documentation
│   │
│   ├── base/                    # Base visitor implementations
│   │   ├── visitor.go           # BaseVisitor with default implementations
│   │   ├── builder.go           # Common generation utilities
│   │   └── errors.go            # Generation error types
│   │
│   ├── json/                    # JSON Schema generation
│   │   ├── generator.go         # JSONSchemaGenerator
│   │   ├── options.go           # JSON Schema options
│   │   └── json_test.go         # JSON Schema tests
│   │
│   ├── typescript/              # TypeScript generation
│   │   ├── generator.go         # TypeScriptGenerator
│   │   ├── options.go           # TypeScript options
│   │   ├── templates.go         # TS type templates
│   │   └── typescript_test.go   # TypeScript tests
│   │
│   ├── python/                  # Python type generation
│   │   ├── generator.go         # PythonGenerator
│   │   ├── options.go           # Python options
│   │   ├── templates.go         # Python type templates
│   │   └── python_test.go       # Python tests
│   │
│   ├── golang/                  # Go type generation
│   │   ├── generator.go         # GoGenerator
│   │   ├── options.go           # Go generation options
│   │   ├── templates.go         # Go type templates
│   │   └── golang_test.go       # Go generation tests
│   │
│   ├── openapi/                 # OpenAPI specification
│   │   ├── generator.go         # OpenAPIGenerator
│   │   ├── options.go           # OpenAPI options
│   │   └── openapi_test.go      # OpenAPI tests
│   │
│   ├── factory.go               # Generator factory functions
│   ├── registry.go              # Generator registry
│   └── export_test.go           # Integration tests
```

### Core Interfaces
```go
// Generator is the main interface for all export generators
type Generator interface {
    core.SchemaVisitor
    Generate(schema core.Schema) ([]byte, error)
    Name() string
    Format() string
}

// GeneratorWithOptions supports configuration
type GeneratorWithOptions[T any] interface {
    Generator
    WithOptions(options T) Generator
}

// MultiFileGenerator generates multiple files
type MultiFileGenerator interface {
    Generator
    GenerateFiles(schema core.Schema) (map[string][]byte, error)
}
```

## 📋 Implementation Plan

### Phase 1: Foundation (Week 1)
**Goal**: Set up package structure and base infrastructure

#### 1.1 Package Setup
- [ ] Create `schema/export/` directory
- [ ] Create `export/doc.go` with package documentation
- [ ] Create `export/README.md` with usage examples
- [ ] Update `schema/go.mod` if needed

#### 1.2 Base Infrastructure
- [ ] **`export/base/visitor.go`** - Base visitor implementation
  ```go
  type BaseVisitor struct{}
  func (b *BaseVisitor) VisitString(core.StringSchema) error { return nil }
  func (b *BaseVisitor) VisitNumber(core.NumberSchema) error { return nil }
  // ... default implementations for all visit methods
  ```
- [ ] **`export/base/builder.go`** - Common generation utilities
  ```go
  type GenerationContext struct {
      IndentLevel int
      Options     map[string]any
      Metadata    map[string]any
  }
  func Indent(level int) string
  func EscapeString(s string) string
  func SanitizeIdentifier(s string) string
  ```
- [ ] **`export/base/errors.go`** - Generation error types
  ```go
  type GenerationError struct {
      SchemaType string
      Message    string
      Cause      error
  }
  ```

#### 1.3 Core Interfaces
- [ ] **`export/interfaces.go`** - Generator interfaces
- [ ] **`export/factory.go`** - Generator factory functions
- [ ] **`export/registry.go`** - Generator registry for dynamic dispatch

### Phase 2: JSON Schema Migration (Week 1-2)
**Goal**: Move existing JSON Schema generation to visitor pattern

#### 2.1 JSON Schema Generator
- [ ] **`export/json/generator.go`** - JSONSchemaGenerator implementation
  ```go
  type JSONSchemaGenerator struct {
      *base.BaseVisitor
      result map[string]any
      stack  []map[string]any
  }
  
  func (g *JSONSchemaGenerator) VisitString(s core.StringSchema) error {
      schema := map[string]any{"type": "string"}
      if s.MinLength() != nil {
          schema["minLength"] = *s.MinLength()
      }
      // ... move logic from StringSchema.ToJSONSchema()
      g.setResult(schema)
      return nil
  }
  ```

#### 2.2 Options Support
- [ ] **`export/json/options.go`** - JSON Schema configuration
  ```go
  type JSONSchemaOptions struct {
      Draft            string // "draft-07", "draft-2020-12"
      IncludeExamples  bool
      IncludeDefaults  bool
      StrictMode       bool
  }
  ```

#### 2.3 Migration Support
- [ ] Create compatibility layer in existing schemas
  ```go
  // Temporary compatibility method
  func (s *StringSchema) ToJSONSchema() map[string]any {
      generator := export.NewJSONSchemaGenerator()
      s.Accept(generator)
      result, _ := generator.Result()
      return result
  }
  ```

#### 2.4 Testing
- [ ] **`export/json/json_test.go`** - Comprehensive JSON Schema tests
- [ ] Move existing JSON Schema tests to new package
- [ ] Add regression tests to ensure compatibility

### Phase 3: TypeScript Generation (Week 2-3)
**Goal**: Implement TypeScript type generation

#### 3.1 TypeScript Generator
- [ ] **`export/typescript/generator.go`** - TypeScript type generator
  ```go
  type TypeScriptGenerator struct {
      *base.BaseVisitor
      result   strings.Builder
      options  TypeScriptOptions
      imports  map[string]bool
  }
  
  func (g *TypeScriptGenerator) VisitString(s core.StringSchema) error {
      if s.EnumValues() != nil {
          g.writeEnumType(s.EnumValues())
      } else {
          g.result.WriteString("string")
      }
      return nil
  }
  ```

#### 3.2 TypeScript Templates
- [ ] **`export/typescript/templates.go`** - TypeScript type templates
  ```go
  const (
      InterfaceTemplate = `export interface {{.Name}} {
          {{range .Properties}}{{.Name}}: {{.Type}};
          {{end}}
      }`
      
      EnumTemplate = `export enum {{.Name}} {
          {{range .Values}}{{.}} = "{{.}}",
          {{end}}
      }`
  )
  ```

#### 3.3 TypeScript Options
- [ ] **`export/typescript/options.go`** - TypeScript configuration
  ```go
  type TypeScriptOptions struct {
      ExportStyle      string // "interface", "type", "class"
      IncludeComments  bool
      StrictNullChecks bool
      OptionalStyle    string // "?" or "| undefined"
  }
  ```

### Phase 4: Python Generation (Week 3-4)
**Goal**: Implement Python type generation

#### 4.1 Python Generator
- [ ] **`export/python/generator.go`** - Python type generator
  ```go
  type PythonGenerator struct {
      *base.BaseVisitor
      result  strings.Builder
      options PythonOptions
      imports map[string]bool
  }
  
  func (g *PythonGenerator) VisitString(s core.StringSchema) error {
      if s.EnumValues() != nil {
          g.writeEnumType(s.EnumValues())
      } else {
          g.result.WriteString("str")
      }
      return nil
  }
  ```

#### 4.2 Python Templates
- [ ] **`export/python/templates.go`** - Python type templates
  ```go
  const (
      DataclassTemplate = `@dataclass
      class {{.Name}}:
          {{range .Properties}}{{.Name}}: {{.Type}}
          {{end}}`
      
      TypedDictTemplate = `class {{.Name}}(TypedDict):
          {{range .Properties}}{{.Name}}: {{.Type}}
          {{end}}`
  )
  ```

#### 4.3 Python Options
- [ ] **`export/python/options.go`** - Python configuration
  ```go
  type PythonOptions struct {
      Style           string // "dataclass", "typeddict", "pydantic"
      IncludeImports  bool
      StrictOptional  bool
      ValidationMode  string // "none", "basic", "pydantic"
  }
  ```

### Phase 5: Go Generation (Week 4-5)
**Goal**: Implement Go type generation

#### 5.1 Go Generator
- [ ] **`export/golang/generator.go`** - Go type generator
  ```go
  type GoGenerator struct {
      *base.BaseVisitor
      result  strings.Builder
      options GoOptions
      imports map[string]bool
  }
  
  func (g *GoGenerator) VisitString(s core.StringSchema) error {
      if s.EnumValues() != nil {
          g.writeEnumType(s.EnumValues())
      } else {
          g.result.WriteString("string")
      }
      return nil
  }
  ```

#### 5.2 Go Templates
- [ ] **`export/golang/templates.go`** - Go type templates
  ```go
  const (
      StructTemplate = `type {{.Name}} struct {
          {{range .Properties}}{{.Name}} {{.Type}} \`json:"{{.JSONName}}"\`
          {{end}}
      }`
      
      EnumTemplate = `type {{.Name}} string
      
      const (
          {{range .Values}}{{.Name}}{{.}} {{$.Name}} = "{{.}}"
          {{end}}
      )`
  )
  ```

### Phase 6: OpenAPI Integration (Week 5-6)
**Goal**: Implement OpenAPI specification generation

#### 6.1 OpenAPI Generator
- [ ] **`export/openapi/generator.go`** - OpenAPI spec generator
- [ ] Support for OpenAPI 3.0.x and 3.1.x
- [ ] Function schema → OpenAPI operation mapping
- [ ] Service schema → OpenAPI paths mapping

### Phase 7: Integration & Testing (Week 6-7)
**Goal**: Complete integration and comprehensive testing

#### 7.1 Factory & Registry
- [ ] **`export/factory.go`** - Generator factory
  ```go
  func NewJSONSchemaGenerator(opts ...JSONSchemaOption) Generator
  func NewTypeScriptGenerator(opts ...TypeScriptOption) Generator
  func NewPythonGenerator(opts ...PythonOption) Generator
  func NewGoGenerator(opts ...GoOption) Generator
  ```

- [ ] **`export/registry.go`** - Generator registry
  ```go
  func RegisterGenerator(name string, factory GeneratorFactory)
  func GetGenerator(name string) (Generator, error)
  func ListGenerators() []string
  ```

#### 7.2 Core Integration
- [ ] Update `schema/core/core.go` with export factory functions
- [ ] Add convenience methods for common generation patterns

#### 7.3 Comprehensive Testing
- [ ] **`export/export_test.go`** - Integration tests
- [ ] Performance benchmarks for each generator
- [ ] Memory usage profiling
- [ ] Compatibility tests with existing JSON Schema output

### Phase 8: Migration & Cleanup (Week 7-8)
**Goal**: Complete migration and remove old generation code

#### 8.1 Deprecation
- [ ] Mark `ToJSONSchema()` methods as deprecated
- [ ] Add deprecation warnings with migration guidance
- [ ] Update documentation to use new export package

#### 8.2 Cleanup
- [ ] Remove `ToJSONSchema()` from core Schema interface
- [ ] Remove implementation from all schema types
- [ ] Update all tests to use new export package

#### 8.3 Documentation
- [ ] Update README.md with new usage patterns
- [ ] Create migration guide
- [ ] Add comprehensive examples

## 🔄 Migration Strategy

### Backward Compatibility
1. **Phase 1-7**: Keep existing `ToJSONSchema()` methods
2. **Phase 7**: Add deprecation warnings
3. **Phase 8**: Remove deprecated methods (breaking change)

### Migration Path
```go
// Old way (deprecated)
jsonSchema := schema.ToJSONSchema()

// New way
generator := export.NewJSONSchemaGenerator()
jsonSchema, err := generator.Generate(schema)
```

### Compatibility Layer
```go
// Temporary compatibility (to be removed in Phase 8)
func (s *StringSchema) ToJSONSchema() map[string]any {
    generator := export.NewJSONSchemaGenerator()
    result, _ := generator.Generate(s)
    return result
}
```

## 🧪 Testing Strategy

### Unit Testing
- [ ] Test each generator independently
- [ ] Test all schema types with each generator
- [ ] Test error conditions and edge cases

### Integration Testing
- [ ] End-to-end generation workflows
- [ ] Complex nested schema generation
- [ ] Performance testing with large schemas

### Regression Testing
- [ ] Ensure JSON Schema output matches current implementation
- [ ] Test backward compatibility during migration
- [ ] Validate generated code compiles/runs correctly

### Property-Based Testing
- [ ] Generate random schemas and validate output
- [ ] Test round-trip compatibility where possible
- [ ] Fuzzing for edge cases

## 📈 Benefits

### Immediate Benefits
- **Separation of Concerns**: Validation vs. generation logic
- **Extensibility**: Easy to add new output formats
- **Maintainability**: Generation logic centralized
- **Testability**: Test generators independently

### Long-term Benefits
- **Plugin System**: Third-party generators
- **Performance**: Specialized generators for different use cases
- **Consistency**: Unified approach to code generation
- **Evolution**: Easy to update generation logic without touching schemas

## ⚠️ Risks & Mitigation

### Risks
1. **Breaking Changes**: Removing `ToJSONSchema()` methods
2. **Performance Impact**: Visitor pattern overhead
3. **Complexity**: More complex codebase initially
4. **Migration Effort**: Updating existing code

### Mitigation
1. **Gradual Migration**: Maintain compatibility during transition
2. **Performance Testing**: Benchmark and optimize
3. **Clear Documentation**: Comprehensive migration guides
4. **Tooling Support**: Automated migration tools if needed

## 🎯 Success Criteria

### Functional
- [ ] All existing JSON Schema output identical
- [ ] TypeScript generation produces valid TypeScript
- [ ] Python generation produces valid Python types
- [ ] Go generation produces valid Go code
- [ ] OpenAPI generation produces valid OpenAPI specs

### Non-Functional
- [ ] Performance within 10% of current implementation
- [ ] Memory usage reasonable for large schemas
- [ ] Clear, maintainable code architecture
- [ ] Comprehensive test coverage (>95%)

### Developer Experience
- [ ] Easy to add new generators
- [ ] Clear error messages for generation failures
- [ ] Good documentation and examples
- [ ] Smooth migration path from old API

## 📚 Documentation Plan

### Package Documentation
- [ ] `export/README.md` - Package overview and usage
- [ ] `export/doc.go` - Go package documentation
- [ ] Generator-specific documentation in each subpackage

### Migration Documentation
- [ ] Migration guide from old `ToJSONSchema()` API
- [ ] Examples of common usage patterns
- [ ] Performance considerations and best practices

### API Documentation
- [ ] Complete godoc documentation for all public APIs
- [ ] Usage examples for each generator
- [ ] Configuration options documentation

## 🚀 Getting Started

To begin implementation:

1. **Create base package structure** (Phase 1.1)
2. **Implement base visitor** (Phase 1.2)
3. **Migrate JSON Schema generation** (Phase 2.1)
4. **Add comprehensive tests** (Phase 2.4)
5. **Validate compatibility** with existing tests

This plan provides a structured approach to refactoring the schema generation system while maintaining backward compatibility and enabling future extensibility. 