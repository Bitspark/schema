# Consumer-Driven Architecture with Schema Condition Filtering

## 🎯 **Core Insight**

**Consumers should drive the architecture**, not visitors. Consumers use the visitor pattern internally for type dispatch, but the primary interface is consumer-based with sophisticated schema filtering.

## 📋 **Schema Condition System**

### **Condition Interface**
```go
type SchemaCondition interface {
    Matches(schema core.Schema) bool
    String() string // for debugging/logging
}
```

### **Condition Types**

#### **1. Logical Conditions (Nesting)**
```go
type AndCondition struct {
    Conditions []SchemaCondition
}

func (c AndCondition) Matches(schema core.Schema) bool {
    for _, condition := range c.Conditions {
        if !condition.Matches(schema) {
            return false
        }
    }
    return true
}

type OrCondition struct {
    Conditions []SchemaCondition
}

func (c OrCondition) Matches(schema core.Schema) bool {
    for _, condition := range c.Conditions {
        if condition.Matches(schema) {
            return true
        }
    }
    return false
}

type NotCondition struct {
    Condition SchemaCondition
}

func (c NotCondition) Matches(schema core.Schema) bool {
    return !c.Condition.Matches(schema)
}
```

#### **2. Type Conditions**
```go
type TypeCondition struct {
    Type core.SchemaType
}

func (c TypeCondition) Matches(schema core.Schema) bool {
    return schema.Type() == c.Type
}

type AnyTypeCondition struct {
    Types []core.SchemaType
}

func (c AnyTypeCondition) Matches(schema core.Schema) bool {
    schemaType := schema.Type()
    for _, t := range c.Types {
        if schemaType == t {
            return true
        }
    }
    return false
}
```

#### **3. Annotation Conditions**
```go
type AnnotationCondition struct {
    AnnotationName string
    Value          any    // optional - if nil, just checks presence
    Operator       string // "equals", "contains", "matches", etc.
}

func (c AnnotationCondition) Matches(schema core.Schema) bool {
    annotations := schema.Annotations()
    
    for _, ann := range annotations {
        if ann.Name() == c.AnnotationName {
            if c.Value == nil {
                return true // just checking presence
            }
            
            switch c.Operator {
            case "equals", "":
                return ann.Value() == c.Value
            case "contains":
                if str, ok := ann.Value().(string); ok {
                    if substr, ok := c.Value.(string); ok {
                        return strings.Contains(str, substr)
                    }
                }
            case "matches":
                if pattern, ok := c.Value.(string); ok {
                    if str, ok := ann.Value().(string); ok {
                        matched, _ := regexp.MatchString(pattern, str)
                        return matched
                    }
                }
            }
        }
    }
    return false
}

type HasAnnotationCondition struct {
    AnnotationName string
}

func (c HasAnnotationCondition) Matches(schema core.Schema) bool {
    annotations := schema.Annotations()
    for _, ann := range annotations {
        if ann.Name() == c.AnnotationName {
            return true
        }
    }
    return false
}
```

#### **4. Complex Conditions**
```go
type PropertyCondition struct {
    PropertyName string
    Condition    SchemaCondition
}

func (c PropertyCondition) Matches(schema core.Schema) bool {
    if objSchema, ok := schema.(core.ObjectSchema); ok {
        properties := objSchema.Properties()
        if propSchema, exists := properties[c.PropertyName]; exists {
            return c.Condition.Matches(propSchema)
        }
    }
    return false
}

type ArrayItemCondition struct {
    ItemCondition SchemaCondition
}

func (c ArrayItemCondition) Matches(schema core.Schema) bool {
    if arraySchema, ok := schema.(core.ArraySchema); ok {
        itemSchema := arraySchema.ItemSchema()
        return c.ItemCondition.Matches(itemSchema)
    }
    return false
}
```

## 🏗️ **Consumer-Driven Interface**

### **Enhanced Consumer Interface**
```go
type AnnotationConsumer interface {
    // Identity
    Name() string
    Purpose() ConsumerPurpose
    
    // 🎯 SCHEMA FILTERING - Your brilliant insight!
    ApplicableSchemas() SchemaCondition
    
    // Processing
    ProcessSchema(schema core.Schema) (any, error)
    
    // Metadata
    Metadata() ConsumerMetadata
}

// ConsumerMetadata provides rich information about the consumer
type ConsumerMetadata struct {
    Name           string            `json:"name"`
    Purpose        ConsumerPurpose   `json:"purpose"`
    Description    string            `json:"description"`
    Version        string            `json:"version"`
    SupportedTypes []core.SchemaType `json:"supported_types"`
    Tags           []string          `json:"tags"`
    Examples       []string          `json:"examples"`
}
```

### **Consumer Registry with Filtering**
```go
type ConsumerRegistry interface {
    // Registration
    Register(consumer AnnotationConsumer) error
    
    // 🎯 FILTERED DISCOVERY
    GetApplicableConsumers(schema core.Schema) []AnnotationConsumer
    GetApplicableConsumersByPurpose(schema core.Schema, purpose ConsumerPurpose) []AnnotationConsumer
    
    // Processing with automatic filtering
    ProcessWithPurpose(purpose ConsumerPurpose, schema core.Schema) (any, error)
    ProcessWithPurposes(purposes []ConsumerPurpose, schema core.Schema) (ProcessingResult, error)
    
    // Manual consumer selection
    ProcessWithConsumer(consumerName string, schema core.Schema) (any, error)
    ProcessWithConsumers(consumerNames []string, schema core.Schema) (map[string]any, error)
}
```

## 💡 **Implementation Examples**

### **1. Email Validator Consumer**
```go
type EmailValidator struct {
    name string
}

func (v *EmailValidator) Name() string {
    return "email-validator"
}

func (v *EmailValidator) Purpose() ConsumerPurpose {
    return PurposeValidation
}

// 🎯 SOPHISTICATED FILTERING
func (v *EmailValidator) ApplicableSchemas() SchemaCondition {
    return AndCondition{
        Conditions: []SchemaCondition{
            TypeCondition{Type: core.TypeString},
            AnnotationCondition{
                AnnotationName: "format",
                Value:          "email",
                Operator:       "equals",
            },
        },
    }
}

func (v *EmailValidator) ProcessSchema(schema core.Schema) (any, error) {
    // Consumer uses visitor pattern internally
    if accepter, ok := schema.(core.Accepter); ok {
        return nil, accepter.Accept(v)
    }
    return nil, fmt.Errorf("schema does not support visitor pattern")
}

// Internal visitor implementation
func (v *EmailValidator) VisitString(schema core.StringSchema) error {
    // Email validation logic
    return v.validateEmailSchema(schema)
}

// Implement other visitor methods as no-ops or errors
func (v *EmailValidator) VisitInteger(schema core.IntegerSchema) error {
    return fmt.Errorf("email validator does not support integer schemas")
}
// ... other visitor methods
```

### **2. TypeScript Generator Consumer**
```go
type TypeScriptGenerator struct {
    options TypeScriptOptions
    output  strings.Builder
}

func (g *TypeScriptGenerator) Purpose() ConsumerPurpose {
    return PurposeGeneration
}

func (g *TypeScriptGenerator) ApplicableSchemas() SchemaCondition {
    return OrCondition{
        Conditions: []SchemaCondition{
            // Generate for any schema with @typescript annotation
            HasAnnotationCondition{AnnotationName: "typescript"},
            
            // Or any schema with @format annotation (for type mapping)
            HasAnnotationCondition{AnnotationName: "format"},
            
            // Or any object/array schema (structural types)
            AnyTypeCondition{
                Types: []core.SchemaType{
                    core.TypeStructure,
                    core.TypeArray,
                },
            },
        },
    }
}

func (g *TypeScriptGenerator) ProcessSchema(schema core.Schema) (any, error) {
    g.output.Reset()
    
    if accepter, ok := schema.(core.Accepter); ok {
        if err := accepter.Accept(g); err != nil {
            return nil, err
        }
        return g.output.String(), nil
    }
    
    return nil, fmt.Errorf("schema does not support visitor pattern")
}

// Visitor methods with annotation awareness
func (g *TypeScriptGenerator) VisitString(schema core.StringSchema) error {
    annotations := schema.Annotations()
    
    for _, ann := range annotations {
        switch ann.Name() {
        case "typescript":
            return g.generateCustomTypeScript(schema, ann)
        case "format":
            return g.generateFormattedString(schema, ann)
        }
    }
    
    return g.generateBasicString(schema)
}
```

### **3. Complex Condition Example**
```go
type ObjectValidatorConsumer struct {
    name string
}

func (v *ObjectValidatorConsumer) ApplicableSchemas() SchemaCondition {
    return AndCondition{
        Conditions: []SchemaCondition{
            TypeCondition{Type: core.TypeStructure},
            OrCondition{
                Conditions: []SchemaCondition{
                    // Objects with @validate annotation
                    HasAnnotationCondition{AnnotationName: "validate"},
                    
                    // Objects with required properties
                    PropertyCondition{
                        PropertyName: "email",
                        Condition: AndCondition{
                            Conditions: []SchemaCondition{
                                TypeCondition{Type: core.TypeString},
                                AnnotationCondition{AnnotationName: "format", Value: "email"},
                            },
                        },
                    },
                    
                    // Objects with array properties that need validation
                    PropertyCondition{
                        PropertyName: "items",
                        Condition: ArrayItemCondition{
                            ItemCondition: HasAnnotationCondition{AnnotationName: "validate"},
                        },
                    },
                },
            },
        },
    }
}
```

## 🎯 **Registry Implementation with Filtering**

```go
type RegistryImpl struct {
    consumers []AnnotationConsumer
    mu        sync.RWMutex
}

func (r *RegistryImpl) GetApplicableConsumers(schema core.Schema) []AnnotationConsumer {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var applicable []AnnotationConsumer
    
    for _, consumer := range r.consumers {
        if consumer.ApplicableSchemas().Matches(schema) {
            applicable = append(applicable, consumer)
        }
    }
    
    return applicable
}

func (r *RegistryImpl) ProcessWithPurpose(purpose ConsumerPurpose, schema core.Schema) (any, error) {
    applicable := r.GetApplicableConsumersByPurpose(schema, purpose)
    
    if len(applicable) == 0 {
        return nil, fmt.Errorf("no applicable consumers found for purpose %s", purpose)
    }
    
    // Process with first applicable consumer
    return applicable[0].ProcessSchema(schema)
}
```

## ✅ **Benefits of Consumer-Driven Architecture**

### **🎯 Precise Filtering**
```go
// Only consumers that match the schema condition will be considered
emailSchema := String().Format("email").Build()
validators := registry.GetApplicableConsumersByPurpose(emailSchema, PurposeValidation)
// Only returns EmailValidator, not IntegerValidator or URLValidator
```

### **🔄 Complex Conditions**
```go
// Complex nested conditions
condition := AndCondition{
    Conditions: []SchemaCondition{
        OrCondition{
            Conditions: []SchemaCondition{
                TypeCondition{Type: core.TypeString},
                TypeCondition{Type: core.TypeInteger},
            },
        },
        HasAnnotationCondition{AnnotationName: "format"},
        NotCondition{
            Condition: AnnotationCondition{AnnotationName: "skip-validation", Value: true},
        },
    },
}
```

### **🏗️ Clean Separation**
- **Consumers**: Define what they can process and how
- **Visitors**: Internal implementation detail for type dispatch
- **Registry**: Handles filtering and routing
- **Conditions**: Declarative schema matching

## 🎉 **Conclusion**

Your insight about **consumers driving the architecture** with **schema condition filtering** is much more elegant than making visitors into consumers. This approach:

1. **Keeps concerns separated**: Consumers define applicability, visitors handle type dispatch
2. **Enables sophisticated filtering**: Complex nested conditions for precise matching
3. **Maintains clean interfaces**: No dual inheritance or interface pollution
4. **Provides better control**: Registry can intelligently route based on conditions

The **consumer-driven architecture** with **schema condition filtering** is the right approach! 🚀 

### **⚙️ Condition Builder DSL (Ergonomic Helpers)**

```go
// helpers.go
func And(conds ...SchemaCondition) SchemaCondition { return AndCondition{Conditions: conds} }
func Or(conds ...SchemaCondition) SchemaCondition  { return OrCondition{Conditions: conds} }
func Not(cond SchemaCondition) SchemaCondition     { return NotCondition{Condition: cond} }

func Type(t core.SchemaType) SchemaCondition {
    return TypeCondition{Type: t}
}

// Flexible: value optional, if omitted we only check presence
func HasAnnotation(name string, value ...any) SchemaCondition {
    if len(value) == 0 {
        return HasAnnotationCondition{AnnotationName: name}
    }
    return AnnotationCondition{
        AnnotationName: name,
        Value:          value[0],
    }
}
```

*Example usage*
```go
// Old (verbose)
AndCondition{
    Conditions: []SchemaCondition{
        TypeCondition{Type: core.TypeString},
        AnnotationCondition{AnnotationName: "format", Value: "email"},
    },
}

// New (concise)
And(Type(core.TypeString), HasAnnotation("format", "email"))
```

---

### **📦 Schema Annotation Accessors**

Add convenience methods on `core.Schema` (implemented via embedding or utility functions):
```go
type Schema interface {
    // ... existing methods ...

    // NEW helpers (sugar only – no breaking change)
    GetAnnotation(name string) (Annotation, bool)
    HasAnnotation(name string) bool
    AnnotationsByName(name string) []Annotation
}
```

Default implementation can live in a helper that introspects the `Annotations()` slice.

---

### **📑 Generic Consumer Result Interface (Purpose-Agnostic)**

Instead of returning raw `any`, provide minimal introspection **without binding to specific purposes**:
```go
type ConsumerResult interface {
    // Human/semantic identifier – e.g. "validation", "generation", "analysis", "custom" …
    Kind() string

    // Underlying strongly-typed value
    Value() any

    // Optional machine-readable type info for reflection / tooling
    GoType() reflect.Type
}
```

Consumers implement:
```go
func (v *EmailValidator) ProcessSchema(schema core.Schema) (ConsumerResult, error)
```

Helpers to build results:
```go
func NewResult(kind string, value any) ConsumerResult { /* … */ }
```

> No `AsValidation()` / `AsGeneration()` helpers – we stay **purpose-agnostic**.

---

### **🗃️ Registry Aggregation Helpers**

Allow aggregation across **all matching consumers**:
```go
// For a single purpose
func (r *RegistryImpl) ProcessAllWithPurpose(purpose ConsumerPurpose, schema core.Schema) ([]ConsumerResult, error)

// For multiple purposes
func (r *RegistryImpl) ProcessAllWithPurposes(purposes []ConsumerPurpose, schema core.Schema) (map[ConsumerPurpose][]ConsumerResult, error)
```

Existing `ProcessWithPurpose` keeps returning the *first* match (fast path).

---

### **🛠️ Structured Error & Processing Context**

```go
type ConsumerError struct {
    Consumer string
    Purpose  ConsumerPurpose
    Path     []string        // e.g. ["preferences", "theme"]
    Cause    error
}

func (e ConsumerError) Error() string { /* formatted message */ }

// Context passed into ProcessSchema for richer info
// (keeps backwards-compat by providing a zero-value default)
type ProcessingContext struct {
    Schema      core.Schema
    Path        []string
    Parent      core.Schema
    Options     map[string]any
}

// New signature (old kept as helper)
ProcessSchema(ctx ProcessingContext) (ConsumerResult, error)
```

Consumers that don't care about context can accept the zero-value `ProcessingContext{Schema: s}`.

---

### **🔗 Updated Consumer Interface (excerpt)**
```go
type AnnotationConsumer interface {
    Name() string
    Purpose() ConsumerPurpose
    ApplicableSchemas() SchemaCondition

    // Context-aware processing
    ProcessSchema(ctx ProcessingContext) (ConsumerResult, error)

    Metadata() ConsumerMetadata
}
```

These additions preserve the existing architecture **while greatly improving ergonomics, introspection, and composability**, all without hard-coding any particular purpose categories. 