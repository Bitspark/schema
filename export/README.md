# Schema Export Package

The `schema/export` package provides a clean, extensible system for generating various output formats from schemas using the visitor pattern. This package separates generation concerns from schema validation, enabling easy addition of new output formats without modifying existing schema types.

## 🎯 Goals

- **Clean Architecture**: Separation of validation and generation concerns
- **Extensibility**: Easy addition of new output formats
- **Type Safety**: Full integration with the schema core API
- **Performance**: Efficient generation with optional parallel processing
- **Consistency**: Unified approach across all generators

## 🏗️ Architecture

The export package uses the **visitor pattern** where:

- **Schemas** implement `core.Accepter` and can accept visitors
- **Generators** implement `core.SchemaVisitor` and `export.Generator`
- **Each generator** produces output in its specific format

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Schema Types  │    │   Generators    │    │     Output      │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │StringSchema │ │──▶│ │JSONGenerator│ │───▶│ │JSON Schema  │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ObjectSchema │ │──▶│ │TSGenerator  │ │───▶│ │TypeScript   │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│      ...        │    │      ...        │    │      ...        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📦 Package Structure

```
schema/export/
├── doc.go                  # Package documentation
├── README.md               # This file
├── interfaces.go           # Core interfaces
├── factory.go              # Generator factories and options
├── registry.go             # Generator registry and batch processing
│
├── base/                   # Base implementations
│   ├── visitor.go          # BaseVisitor with default implementations
│   ├── builder.go          # Generation utilities and context
│   └── errors.go           # Error types and handling
│
├── json/                   # JSON Schema generation (Phase 2)
│   ├── generator.go        # JSONSchemaGenerator
│   ├── options.go          # JSON Schema options
│   └── json_test.go        # Tests
│
├── typescript/             # TypeScript generation (Phase 3)
│   ├── generator.go        # TypeScriptGenerator
│   ├── options.go          # TypeScript options
│   └── typescript_test.go  # Tests
│
└── ...                     # Additional generators
```

## 🚀 Quick Start

### Basic Usage

```go
import "defs.dev/schema/export"

// Create a generator
generator := export.NewJSONSchemaGenerator()

// Generate output from a schema
output, err := generator.Generate(schema)
if err != nil {
    log.Fatal(err)
}

fmt.Println(string(output))
```

### Using the Registry

```go
// Register generators
export.Register("json", export.NewJSONSchemaGenerator())
export.Register("typescript", export.NewTypeScriptGenerator())

// Generate with specific generator
output, err := export.Generate("json", schema)

// Generate with all registered generators
outputs, err := export.GenerateAll(schema)
for format, output := range outputs {
    fmt.Printf("%s: %s\n", format, output)
}
```

### Generator Configuration

```go
// Using functional options
generator := export.NewJSONSchemaGenerator(
    export.WithIndentSize(4),
    export.WithComments(true),
    export.WithExamples(true),
)

// Using builder pattern
generator := export.NewGeneratorBuilder("json").
    IndentSize(4).
    Comments(true).
    Examples(true).
    Build()
```

## 🔧 Core Interfaces

### Generator

The main interface that all generators must implement:

```go
type Generator interface {
    core.SchemaVisitor

    // Generate produces output from a schema
    Generate(schema core.Schema) ([]byte, error)

    // Name returns the human-readable name
    Name() string

    // Format returns the output format identifier
    Format() string
}
```

### Option

Functional options for configuring generators:

```go
type Option interface {
    Apply(target any) error
}

// Common options
export.WithIndentSize(4)
export.WithComments(true)
export.WithNamingStyle("camelCase")
export.WithStrictMode(true)
```

### GeneratorRegistry

Manages multiple generators and enables batch generation:

```go
type GeneratorRegistry interface {
    Register(name string, generator Generator) error
    Get(name string) (Generator, bool)
    Generate(generatorName string, schema core.Schema) ([]byte, error)
    GenerateAll(schema core.Schema) (map[string][]byte, error)
    // ... more methods
}
```

## 🎨 Creating Custom Generators

### Basic Generator

```go
type MyGenerator struct {
    *base.BaseVisitor
    output strings.Builder
    opts   MyOptions
}

func NewMyGenerator(options ...export.Option) export.Generator {
    gen := &MyGenerator{
        BaseVisitor: base.NewBaseVisitor("my-generator"),
        opts:        defaultMyOptions(),
    }
    
    // Apply options
    for _, opt := range options {
        opt.Apply(&gen.opts)
    }
    
    return gen
}

func (g *MyGenerator) VisitString(s core.StringSchema) error {
    g.output.WriteString("custom string representation")
    return nil
}

func (g *MyGenerator) Generate(schema core.Schema) ([]byte, error) {
    g.output.Reset()
    
    if err := schema.Accept(g); err != nil {
        return nil, err
    }
    
    return []byte(g.output.String()), nil
}

func (g *MyGenerator) Name() string { return "My Generator" }
func (g *MyGenerator) Format() string { return "my-format" }
```

### Generator with Options

```go
type MyOptions struct {
    IndentSize int
    UseColors  bool
}

func (o *MyOptions) SetOption(key string, value any) {
    switch key {
    case "indent_size":
        if v, ok := value.(int); ok {
            o.IndentSize = v
        }
    case "use_colors":
        if v, ok := value.(bool); ok {
            o.UseColors = v
        }
    }
}

// Register the generator
export.RegisterGenerator("my-generator", func(options ...export.Option) (export.Generator, error) {
    return NewMyGenerator(options...), nil
})
```

## 📊 Batch Generation

### Sequential Generation

```go
// Generate with all registered generators
outputs, err := export.GenerateAll(schema)
```

### Parallel Generation

```go
batch := export.NewBatchGenerator(
    export.DefaultRegistry,
    export.BatchOptions{
        Parallel:       true,
        MaxConcurrency: 4,
        ContinueOnError: true,
    },
)

result := batch.Generate(schema)

for name, genResult := range result.Results {
    fmt.Printf("%s: %d bytes\n", name, len(genResult.Output))
}

for name, err := range result.Errors {
    fmt.Printf("%s failed: %v\n", name, err)
}
```

## 🛠️ Utilities

### Generation Context

The `base.GenerationContext` provides utilities for generators:

```go
ctx := base.NewGenerationContext()

// Indentation
ctx.PushIndent()
indent := ctx.Indent() // "  "
ctx.PopIndent()

// Path tracking
ctx.PushPath("user")
ctx.PushPath("name")
path := ctx.CurrentPath() // "user.name"

// Unique identifiers
id1 := ctx.UniqueIdentifier("User") // "User"
id2 := ctx.UniqueIdentifier("User") // "User1"

// String utilities
camel := base.ToCamelCase("user_name")     // "userName"
pascal := base.ToPascalCase("user_name")   // "UserName"
snake := base.ToSnakeCase("userName")      // "user_name"
```

### Error Handling

```go
// Structured generation errors
err := base.NewGenerationError("json-generator", "string", "invalid format")
err.WithPath("user", "name").WithContext("format", "email")

// Error collection
collector := base.NewErrorCollector()
collector.Add(err1)
collector.Add(err2)
if collector.HasErrors() {
    return collector.Error()
}
```

## 🔮 Supported Formats (Roadmap)

| Format | Status | Phase | Description |
|--------|--------|-------|-------------|
| **JSON Schema** | ✅ Planned | Phase 2 | Standard JSON Schema Draft 7/2020-12 |
| **TypeScript** | 📋 Planned | Phase 3 | TypeScript interfaces and types |
| **Python** | 📋 Planned | Phase 4 | Python type hints (dataclass, TypedDict, Pydantic) |
| **Go** | 📋 Planned | Phase 5 | Go struct definitions with JSON tags |
| **OpenAPI** | 📋 Planned | Phase 6 | OpenAPI 3.x specification components |
| **Protocol Buffers** | 🚧 Future | TBD | Protocol Buffer schema definitions |
| **Avro** | 🚧 Future | TBD | Apache Avro schema definitions |

## 🎯 Common Patterns

### Conditional Generation

```go
func (g *MyGenerator) VisitObject(obj core.ObjectSchema) error {
    if g.opts.IncludeComments && obj.Metadata().Description != "" {
        g.writeComment(obj.Metadata().Description)
    }
    
    // Generate object...
    return nil
}
```

### Nested Schema Handling

```go
func (g *MyGenerator) VisitObject(obj core.ObjectSchema) error {
    for name, propSchema := range obj.Properties() {
        g.Context.WithPath(name, func() {
            if err := g.VisitNested(propSchema, name); err != nil {
                // Error will include path context
            }
        })
    }
    return nil
}
```

### Template-Based Generation

```go
const templateString = `
export interface {{.Name}} {
{{range .Properties}}  {{.Name}}: {{.Type}};
{{end}}}
`

func (g *TypeScriptGenerator) generateInterface(data TemplateData) error {
    tmpl := template.Must(template.New("interface").Parse(templateString))
    return tmpl.Execute(&g.output, data)
}
```

## 🚦 Implementation Status

### ✅ Phase 1 Complete (Foundation)
- [x] Core interfaces and architecture
- [x] Base visitor implementations
- [x] Generator factory and registry system
- [x] Error handling and utilities
- [x] Comprehensive documentation

### 📋 Next Steps
- **Phase 2**: JSON Schema generator migration
- **Phase 3**: TypeScript generator implementation
- **Phase 4**: Python generator implementation
- **Phase 5**: Go generator implementation

## 🤝 Contributing

When implementing new generators:

1. **Embed `base.BaseVisitor`** for default implementations
2. **Override only needed `Visit*` methods**
3. **Support functional options** for configuration
4. **Include comprehensive tests**
5. **Follow naming conventions** (`GeneratorName` + "Generator")
6. **Register with factory function** for dynamic creation

## 📚 References

- [Visitor Pattern](https://en.wikipedia.org/wiki/Visitor_pattern)
- [Functional Options in Go](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
- [JSON Schema Specification](https://json-schema.org/)
- [Schema Core API Documentation](../api/core/)

---

**Next**: Proceed to Phase 2 - [JSON Schema Generator Migration](./json/README.md) 