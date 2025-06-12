# Visitor-Consumer Integration: Unified Architecture

## 🎯 **Core Insight**

**Visitors ARE Consumers!** The export system's visitor pattern can be enhanced to implement the AnnotationConsumer interface, creating a unified architecture where:

- **Type Dimension**: Extended via annotations (`@format`, `@length`, `@pattern`)
- **Feature Dimension**: Extended via consumers (validators, formatters, generators)

## 📊 **Two-Dimensional Extension Matrix**

```
                    FEATURE DIMENSION (Consumers)
                    ↓
    ┌─────────────┬─────────────┬─────────────┬─────────────┐
    │ Schema Type │ Validation  │ Formatting  │ Generation  │
    ├─────────────┼─────────────┼─────────────┼─────────────┤
    │ String      │ @length     │ @format     │ @typescript │
    │             │ @pattern    │ @case       │ @golang     │
    │             │ @required   │ @trim       │ @python     │
    ├─────────────┼─────────────┼─────────────┼─────────────┤
    │ Integer     │ @range      │ @currency   │ @int64      │
    │             │ @positive   │ @thousands  │ @bigint     │
    ├─────────────┼─────────────┼─────────────┼─────────────┤
    │ Object      │ @required   │ @json       │ @interface  │
    │             │ @validate   │ @yaml       │ @struct     │
    └─────────────┴─────────────┴─────────────┴─────────────┘
```

## 🔄 **Enhanced Generator Interface**

### **Current Export Generator:**
```go
type Generator interface {
    core.SchemaVisitor  // Visitor pattern
    Generate(schema core.Schema) ([]byte, error)
    Name() string
    Format() string
}
```

### **Enhanced Generator as Consumer:**
```go
type Generator interface {
    core.SchemaVisitor      // Visitor pattern (type dispatch)
    api.AnnotationConsumer  // Consumer pattern (purpose + annotation awareness)
    
    Generate(schema core.Schema) ([]byte, error)
    Name() string
    Format() string
}

// AnnotationConsumer methods:
// - Purpose() ConsumerPurpose
// - SupportedAnnotations() []string  
// - ProcessAnnotations(context any, annotations []core.Annotation) (any, error)
// - Metadata() ConsumerMetadata
```

## 🏗️ **Implementation Examples**

### **1. TypeScript Generator as Consumer**

```go
type TypeScriptGenerator struct {
    base.BaseVisitor
    options TypeScriptOptions
    output  strings.Builder
}

// ✅ Implement AnnotationConsumer interface
func (g *TypeScriptGenerator) Purpose() ConsumerPurpose {
    return PurposeGeneration
}

func (g *TypeScriptGenerator) SupportedAnnotations() []string {
    return []string{"typescript", "format", "description", "example"}
}

func (g *TypeScriptGenerator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    schema, ok := context.(core.Schema)
    if !ok {
        return nil, fmt.Errorf("context must be a schema")
    }
    return g.Generate(schema)
}

func (g *TypeScriptGenerator) Metadata() ConsumerMetadata {
    return ConsumerMetadata{
        Name:        "TypeScript Generator",
        Purpose:     PurposeGeneration,
        Description: "Generates TypeScript interfaces and types",
        Version:     "1.0.0",
        SupportedTypes: []string{"string", "number", "boolean", "object", "array"},
    }
}

// ✅ Enhanced visitor methods with annotation awareness
func (g *TypeScriptGenerator) VisitString(schema core.StringSchema) error {
    annotations := schema.Annotations()
    
    // Process TypeScript-specific annotations
    for _, ann := range annotations {
        switch ann.Name() {
        case "typescript":
            return g.generateCustomTypeScript(schema, ann)
        case "format":
            return g.generateFormattedString(schema, ann)
        case "enum":
            return g.generateStringEnum(schema, ann)
        }
    }
    
    // Default string generation
    return g.generateBasicString(schema)
}

func (g *TypeScriptGenerator) generateFormattedString(schema core.StringSchema, ann core.Annotation) error {
    format := ann.Value().(string)
    typeName := g.getTypeName(schema)
    
    switch format {
    case "email":
        g.output.WriteString(fmt.Sprintf("type %s = string; // email format\n", typeName))
    case "url":
        g.output.WriteString(fmt.Sprintf("type %s = string; // URL format\n", typeName))
    case "date":
        g.output.WriteString(fmt.Sprintf("type %s = Date;\n", typeName))
    default:
        g.output.WriteString(fmt.Sprintf("type %s = string; // %s format\n", typeName, format))
    }
    
    return nil
}
```

### **2. Go Generator as Consumer**

```go
type GoGenerator struct {
    base.BaseVisitor
    options GoOptions
    output  strings.Builder
}

func (g *GoGenerator) Purpose() ConsumerPurpose {
    return PurposeGeneration
}

func (g *GoGenerator) SupportedAnnotations() []string {
    return []string{"golang", "format", "validation", "json", "yaml"}
}

func (g *GoGenerator) VisitString(schema core.StringSchema) error {
    annotations := schema.Annotations()
    
    for _, ann := range annotations {
        switch ann.Name() {
        case "golang":
            return g.generateCustomGo(schema, ann)
        case "format":
            return g.generateFormattedString(schema, ann)
        case "validation":
            return g.generateValidatedString(schema, ann)
        }
    }
    
    return g.generateBasicString(schema)
}

func (g *GoGenerator) generateFormattedString(schema core.StringSchema, ann core.Annotation) error {
    format := ann.Value().(string)
    typeName := g.getTypeName(schema)
    
    switch format {
    case "email":
        g.output.WriteString(fmt.Sprintf("type %s string // email format\n", typeName))
        g.generateEmailValidation(typeName)
    case "url":
        g.output.WriteString(fmt.Sprintf("type %s string // URL format\n", typeName))
        g.generateURLValidation(typeName)
    case "uuid":
        g.output.WriteString(fmt.Sprintf("type %s string // UUID format\n", typeName))
        g.generateUUIDValidation(typeName)
    }
    
    return nil
}
```

### **3. JSON Schema Generator as Consumer**

```go
type JSONSchemaGenerator struct {
    base.BaseVisitor
    options JSONOptions
    result  map[string]any
}

func (g *JSONSchemaGenerator) Purpose() ConsumerPurpose {
    return PurposeGeneration
}

func (g *JSONSchemaGenerator) SupportedAnnotations() []string {
    return []string{"json-schema", "format", "pattern", "length", "range"}
}

func (g *JSONSchemaGenerator) VisitString(schema core.StringSchema) error {
    annotations := schema.Annotations()
    
    // Build JSON Schema object
    jsonSchema := map[string]any{
        "type": "string",
    }
    
    // Process annotations
    for _, ann := range annotations {
        switch ann.Name() {
        case "format":
            jsonSchema["format"] = ann.Value()
        case "pattern":
            jsonSchema["pattern"] = ann.Value()
        case "length":
            if lengthConfig, ok := ann.Value().(map[string]any); ok {
                if min, exists := lengthConfig["min"]; exists {
                    jsonSchema["minLength"] = min
                }
                if max, exists := lengthConfig["max"]; exists {
                    jsonSchema["maxLength"] = max
                }
            }
        }
    }
    
    g.result = jsonSchema
    return nil
}
```

## 🎯 **Consumer Selection with Visitors**

### **Registry Integration:**

```go
// Register generators as consumers
registry := NewConsumerRegistry()

// Register with purpose declaration
registry.Register(NewTypeScriptGenerator())  // Purpose: PurposeGeneration
registry.Register(NewGoGenerator())          // Purpose: PurposeGeneration  
registry.Register(NewJSONSchemaGenerator())  // Purpose: PurposeGeneration
registry.Register(NewEmailValidator())       // Purpose: PurposeValidation
registry.Register(NewEmailFormatter())       // Purpose: PurposeFormatting

// ✅ SOLUTION: Selective generation
schema := String().Format("email").Build()

// Only generate TypeScript (no Go, JSON, validation, formatting)
tsCode, err := registry.ProcessWithPurpose(PurposeGeneration, schema, schema.Annotations())

// Only validate (no generation, formatting)
validationResult, err := registry.ProcessWithPurpose(PurposeValidation, "user@example.com", schema.Annotations())

// Generate all formats
generationResults, err := registry.ProcessWithPurpose(PurposeGeneration, schema, schema.Annotations())
```

### **Annotation-Driven Selection:**

```go
// Schema with multiple generation targets
schema := String().
    Format("email").
    AddAnnotation("typescript", map[string]any{"interface": true}).
    AddAnnotation("golang", map[string]any{"struct": true, "validation": true}).
    AddAnnotation("json-schema", map[string]any{"draft": "2020-12"}).
    Build()

// Get consumers that support specific annotations
tsGenerators := registry.GetConsumersForAnnotationAndPurpose("typescript", PurposeGeneration)
goGenerators := registry.GetConsumersForAnnotationAndPurpose("golang", PurposeGeneration)

// Generate only TypeScript
tsResult, err := registry.ProcessWithPurpose(PurposeGeneration, schema, 
    []core.Annotation{schema.GetAnnotation("typescript")})
```

## 🚀 **Migration Strategy**

### **Phase 1: Enhance Existing Generators**

1. **Add AnnotationConsumer interface** to existing generators
2. **Implement Purpose() method** for each generator
3. **Add SupportedAnnotations()** method
4. **Enhance visitor methods** to process annotations

### **Phase 2: Registry Integration**

1. **Register generators as consumers** in the consumer registry
2. **Update export registry** to use consumer selection
3. **Add purpose-based generation methods**

### **Phase 3: Annotation-Aware Generation**

1. **Update visitor methods** to process relevant annotations
2. **Add annotation-specific generation logic**
3. **Maintain backward compatibility** with annotation-unaware schemas

## ✅ **Benefits of Unified Architecture**

### **🎯 Precise Control**
```go
// Only generate Go code (no TypeScript, JSON, validation)
goCode := registry.ProcessWithPurpose(PurposeGeneration, schema, annotations)

// Only validate (no generation)
validation := registry.ProcessWithPurpose(PurposeValidation, value, annotations)
```

### **🔄 Reusable Annotations**
```go
// Same @format annotation works for validation, formatting, AND generation
schema := String().Format("email").Build()

// Validator uses @format for validation rules
// Formatter uses @format for display formatting  
// Generator uses @format for type generation
```

### **🏗️ Extensible Architecture**
```go
// Easy to add new generators as consumers
registry.Register(NewRustGenerator())     // Purpose: PurposeGeneration
registry.Register(NewSwiftGenerator())    // Purpose: PurposeGeneration
registry.Register(NewDocGenerator())      // Purpose: PurposeDocumentation
```

### **🧪 Better Testing**
```go
// Test only generation (no validation side effects)
result := registry.ProcessWithPurpose(PurposeGeneration, schema, annotations)

// Test only validation (no generation overhead)
validation := registry.ProcessWithPurpose(PurposeValidation, value, annotations)
```

## 🎉 **Conclusion**

The **visitor pattern + consumer architecture** creates a powerful unified system where:

- **Visitors handle type dispatch** (`VisitString`, `VisitObject`)
- **Consumers handle purpose selection** (`PurposeGeneration`, `PurposeValidation`)
- **Annotations provide metadata** for both type and purpose processing

This solves both the **consumer selection problem** and maintains the **clean visitor pattern** for type-based processing, while adding **annotation awareness** throughout the system. 