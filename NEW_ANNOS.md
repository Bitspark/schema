# Annotation System Refactoring Plan

## 🎯 **Objective**

Refactor the annotation system to be **purpose-agnostic** and **consumer-driven**, removing validation-specific coupling and enabling any system (validators, formatters, generators, documentation) to consume annotations for their specific purposes.

## 🚨 **Current Problems**

### **Tight Coupling Issues**
- ❌ `Annotation.Validators() []string` - Annotations shouldn't know about validators
- ❌ `Annotation.Validate() AnnotationValidationResult` - Validation logic embedded in annotations
- ❌ `AnnotationMetadata.Validators []string` - Purpose-specific metadata in generic structure
- ❌ `AnnotationMetadata.Required bool` - Validation-specific field in generic metadata

### **Architectural Issues**
- ❌ Annotations are validation-aware instead of purpose-agnostic
- ❌ Hard to extend for new purposes (formatters, generators, documentation)
- ❌ Violates separation of concerns
- ❌ Difficult to test annotations independently of validation

### **Consumer Selection Problem**
- ❌ **No Purpose Filtering**: When you want validation, formatters also run
- ❌ **All-or-Nothing**: Can't selectively invoke consumer types
- ❌ **Performance Waste**: Unnecessary processing of irrelevant consumers
- ❌ **Result Mixing**: Validation, formatting, and generation results get mixed
- ❌ **Poor Developer Experience**: Unpredictable behavior and resource usage

**Example of the Problem:**
```go
// ❌ CURRENT PROBLEM: All consumers run automatically
registry.ValidateWithAnnotations(value, annotations)
// This runs validators + formatters + generators + documenters
// You can't choose "only validation" or "only formatting"
```

## ✅ **Target Architecture**

### **Core Principles**
1. **Purpose-Agnostic**: Annotations are pure metadata containers
2. **Consumer-Driven**: Validators, formatters, generators discover and process annotations
3. **Extensible**: Easy to add new annotation consumers
4. **Testable**: Annotations and consumers can be tested independently
5. **Reusable**: Same annotation can serve multiple purposes

### **Architecture Layers**
```
┌─────────────────────────────────────────────────────────────┐
│                    ANNOTATION CONSUMERS                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ Validators  │ │ Formatters  │ │ Generators  │ │  Docs  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   DISCOVERY MECHANISM                      │
│           SupportedAnnotations() + Name Matching           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  CLEAN ANNOTATION SYSTEM                   │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ Annotation: Name() + Value() + Schema() + Metadata()   │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 📋 **Implementation Plan**

### **Phase 1: Clean Core Annotation Interfaces**

#### **1.1 Update `schema/api/core/annotation.go`**

**BEFORE:**
```go
type Annotation interface {
    Name() string
    Value() any
    Schema() Schema
    Validators() []string        // ❌ REMOVE
    Metadata() AnnotationMetadata
    Validate() AnnotationValidationResult  // ❌ REMOVE
    ToMap() map[string]any
}
```

**AFTER:**
```go
type Annotation interface {
    // Core identity
    Name() string
    Value() any
    
    // Type information
    Schema() Schema
    
    // Generic metadata
    Metadata() AnnotationMetadata
    
    // Serialization
    ToMap() map[string]any
}
```

#### **1.2 Clean AnnotationMetadata**

**REMOVE from AnnotationMetadata:**
- `Required bool` - Validation-specific
- `Validators []string` - Purpose-specific

**KEEP in AnnotationMetadata:**
- Basic information (Name, Description)
- Examples and defaults
- Categorization (Tags, Category, Properties)
- Versioning and ownership
- Usage constraints (AppliesTo, Conflicts, Requires)

### **Phase 2: Create Consumer Interface System with Purpose-Based Selection**

#### **2.1 Create `schema/api/consumer.go`**

```go
// ConsumerPurpose identifies the type of processing a consumer performs
type ConsumerPurpose string

const (
    PurposeValidation    ConsumerPurpose = "validation"
    PurposeFormatting    ConsumerPurpose = "formatting"
    PurposeGeneration    ConsumerPurpose = "generation"
    PurposeDocumentation ConsumerPurpose = "documentation"
    PurposeTransform     ConsumerPurpose = "transform"
    PurposeAnalysis      ConsumerPurpose = "analysis"
)

// AnnotationConsumer defines the interface for any system that processes annotations.
type AnnotationConsumer interface {
    Name() string
    Purpose() ConsumerPurpose  // 🆕 NEW: Declares what this consumer does
    SupportedAnnotations() []string
    ProcessAnnotations(context any, annotations []core.Annotation) (any, error)
    Metadata() ConsumerMetadata
}

// ConsumerMetadata provides information about a consumer
type ConsumerMetadata struct {
    Name           string          `json:"name"`
    Purpose        ConsumerPurpose `json:"purpose"`
    Description    string          `json:"description"`
    Version        string          `json:"version"`
    SupportedTypes []string        `json:"supported_types"`
    Tags           []string        `json:"tags"`
}

// ProcessingContext provides context for annotation processing
type ProcessingContext struct {
    Value       any                 `json:"value"`
    Schema      core.Schema         `json:"schema,omitempty"`
    Annotations []core.Annotation   `json:"annotations"`
    Purposes    []ConsumerPurpose   `json:"purposes"`
    Options     map[string]any      `json:"options,omitempty"`
}

// ProcessingResult contains results from multiple consumer types
type ProcessingResult struct {
    Success    bool                          `json:"success"`
    Results    map[ConsumerPurpose]any       `json:"results"`
    Errors     map[ConsumerPurpose]error     `json:"errors,omitempty"`
    ExecutedAt time.Time                     `json:"executed_at"`
    Duration   time.Duration                 `json:"duration"`
}

// ConsumerRegistry manages annotation consumers with purpose-based selection
type ConsumerRegistry interface {
    // Consumer management
    Register(consumer AnnotationConsumer) error
    Get(name string) (AnnotationConsumer, bool)
    List() []string
    ListByPurpose(purpose ConsumerPurpose) []string
    
    // 🎯 PURPOSE-BASED PROCESSING - SOLVES CONSUMER SELECTION PROBLEM
    ProcessWithPurpose(purpose ConsumerPurpose, context any, annotations []core.Annotation) (any, error)
    ProcessWithPurposes(purposes []ConsumerPurpose, context any, annotations []core.Annotation) (ProcessingResult, error)
    ProcessWithContext(ctx ProcessingContext) (ProcessingResult, error)
    
    // Discovery
    GetConsumersForAnnotation(annotationName string) []AnnotationConsumer
    GetConsumersForPurpose(purpose ConsumerPurpose) []AnnotationConsumer
    GetConsumersForAnnotationAndPurpose(annotationName string, purpose ConsumerPurpose) []AnnotationConsumer
    
    // 🔄 EXPLICIT REGISTRY ACCESS - BACKWARD COMPATIBILITY
    Validators() ValidatorRegistry
    Formatters() FormatterRegistry
    Generators() GeneratorRegistry
    Documenters() DocumentationRegistry
}
```

#### **2.2 Update Validator Interface**

```go
// Validator now implements AnnotationConsumer with purpose declaration
type Validator interface {
    api.AnnotationConsumer  // Embeds the consumer interface (includes Purpose())
    
    // Validation-specific methods
    Validate(value any) ValidationResult
    ValidateWithAnnotations(value any, annotations []core.Annotation) ValidationResult
    ConfigureFromAnnotations(annotations []core.Annotation) error
    
    // Metadata
    Metadata() ValidatorMetadata
}

// Example implementation
type EmailValidator struct {
    name string
}

func (v *EmailValidator) Purpose() ConsumerPurpose {
    return PurposeValidation  // 🎯 Declares validation purpose
}

func (v *EmailValidator) SupportedAnnotations() []string {
    return []string{"format"}
}

func (v *EmailValidator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    return v.ValidateWithAnnotations(context, annotations), nil
}
```

### **Phase 3: Create Purpose-Specific Consumer Types**

#### **3.1 Formatter Interface**

```go
type Formatter interface {
    AnnotationConsumer  // Includes Purpose() method
    Format(value any) (string, error)
    FormatWithAnnotations(value any, annotations []core.Annotation) (string, error)
    ConfigureFromAnnotations(annotations []core.Annotation) error
    Metadata() FormatterMetadata
}

// Example implementation
type EmailFormatter struct {
    name string
}

func (f *EmailFormatter) Purpose() ConsumerPurpose {
    return PurposeFormatting  // 🎯 Declares formatting purpose
}

func (f *EmailFormatter) SupportedAnnotations() []string {
    return []string{"format"}
}

func (f *EmailFormatter) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    return f.FormatWithAnnotations(context, annotations)
}
```

#### **3.2 Code Generator Interface**

```go
type CodeGenerator interface {
    AnnotationConsumer  // Includes Purpose() method
    Generate(schema core.Schema) ([]byte, error)
    GenerateWithAnnotations(schema core.Schema, annotations []core.Annotation) ([]byte, error)
    ConfigureFromAnnotations(annotations []core.Annotation) error
    Metadata() GeneratorMetadata
}

// Example implementation
type TypeScriptGenerator struct {
    name string
}

func (g *TypeScriptGenerator) Purpose() ConsumerPurpose {
    return PurposeGeneration  // 🎯 Declares generation purpose
}

func (g *TypeScriptGenerator) SupportedAnnotations() []string {
    return []string{"format", "description", "pattern"}
}

func (g *TypeScriptGenerator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    schema, ok := context.(core.Schema)
    if !ok {
        return nil, fmt.Errorf("context must be a schema for code generation")
    }
    return g.GenerateWithAnnotations(schema, annotations)
}
```

#### **3.3 Documentation Generator Interface**

```go
type DocumentationGenerator interface {
    AnnotationConsumer
    GenerateDocumentation(schema core.Schema) (string, error)
    GenerateDocumentationWithAnnotations(schema core.Schema, annotations []core.Annotation) (string, error)
    ConfigureFromAnnotations(annotations []core.Annotation) error
    Metadata() DocumentationMetadata
}
```

### **Phase 4: Update Schema System**

#### **4.1 Remove Embedded Validation from Schemas**

**BEFORE:**
```go
func (s *StringSchema) Validate(value any) core.ValidationResult {
    // Type check
    str, ok := value.(string)
    if !ok { return typeError }
    
    // ❌ REMOVE: Embedded constraint validation
    if s.config.MinLength != nil && len(str) < *s.config.MinLength { ... }
    if s.config.MaxLength != nil && len(str) > *s.config.MaxLength { ... }
    if s.config.Pattern != nil && !s.config.Pattern.MatchString(str) { ... }
    
    // ✅ KEEP: Annotation-based validation
    if s.config.ValidatorRegistry != nil {
        return s.config.ValidatorRegistry.ValidateWithAnnotations(str, s.config.Annotations)
    }
}
```

**AFTER:**
```go
func (s *StringSchema) Validate(value any) core.ValidationResult {
    // Type check only
    str, ok := value.(string)
    if !ok { return typeError }
    
    // Pure annotation-based validation
    if s.config.ValidatorRegistry != nil {
        return s.config.ValidatorRegistry.ValidateWithAnnotations(str, s.config.Annotations)
    }
    
    return core.ValidationResult{Valid: true}
}
```

#### **4.2 Clean Schema Configs**

**REMOVE from StringSchemaConfig:**
- `MinLength *int`
- `MaxLength *int`
- `Pattern *regexp.Regexp`
- `Format string`
- `EnumValues []string`

**KEEP in StringSchemaConfig:**
- `Metadata core.SchemaMetadata`
- `DefaultVal *string`
- `Annotations []core.Annotation` (contains @length, @pattern, @format, @enum)
- `ValidatorRegistry ValidatorRegistry`

### **Phase 5: Update Builder System**

#### **5.1 Convert Builder Methods to Annotation-Based**

**BEFORE:**
```go
func (b *StringBuilder) MinLength(min int) *StringBuilder {
    b.config.MinLength = &min  // ❌ Direct config
    return b
}
```

**AFTER:**
```go
func (b *StringBuilder) MinLength(min int) *StringBuilder {
    return b.AddAnnotation("length", map[string]any{"min": min})
}

func (b *StringBuilder) Length(min, max int) *StringBuilder {
    return b.AddAnnotation("length", map[string]any{"min": min, "max": max})
}

func (b *StringBuilder) Pattern(pattern string) *StringBuilder {
    return b.AddAnnotation("pattern", pattern)
}

func (b *StringBuilder) Format(format string) *StringBuilder {
    return b.AddAnnotation("format", format)
}

// Convenience methods
func (b *StringBuilder) Email() *StringBuilder {
    return b.Format("email")
}

func (b *StringBuilder) URL() *StringBuilder {
    return b.Format("url")
}
```

### **Phase 6: Create Standard Annotations**

#### **6.1 Create `schema/annotation/standard.go`**

```go
// Standard annotation schemas for common use cases

// LengthAnnotationSchema defines the schema for @length annotations
var LengthAnnotationSchema = Object().
    Property("min", Integer().Minimum(0).Optional()).
    Property("max", Integer().Minimum(0).Optional()).
    Build()

// FormatAnnotationSchema defines the schema for @format annotations
var FormatAnnotationSchema = String().
    Enum("email", "url", "uuid", "date", "time", "datetime").
    Build()

// RegisterStandardAnnotations registers all standard annotation types
func RegisterStandardAnnotations(registry core.AnnotationRegistry) error {
    annotations := map[string]core.Schema{
        "length":  LengthAnnotationSchema,
        "range":   RangeAnnotationSchema,
        "pattern": PatternAnnotationSchema,
        "format":  FormatAnnotationSchema,
        "enum":    EnumAnnotationSchema,
    }
    
    for name, schema := range annotations {
        if err := registry.RegisterType(name, schema); err != nil {
            return fmt.Errorf("failed to register %s annotation: %w", name, err)
        }
    }
    
    return nil
}
```

### **Phase 7: Update Consumer Implementations**

#### **7.1 Update Validators**

```go
// EmailValidator now implements AnnotationConsumer
type EmailValidator struct {
    BaseValidator
}

func (v *EmailValidator) SupportedAnnotations() []string {
    return []string{"format"}
}

func (v *EmailValidator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    value := context
    return v.ValidateWithAnnotations(value, annotations), nil
}

func (v *EmailValidator) ValidateWithAnnotations(value any, annotations []core.Annotation) ValidationResult {
    for _, ann := range annotations {
        if ann.Name() == "format" && ann.Value() == "email" {
            return v.Validate(value)
        }
    }
    return ValidResult()
}
```

#### **7.2 Create Formatters**

```go
// EmailFormatter formats email addresses
type EmailFormatter struct {
    name string
}

func (f *EmailFormatter) SupportedAnnotations() []string {
    return []string{"format"}
}

func (f *EmailFormatter) FormatWithAnnotations(value any, annotations []core.Annotation) (string, error) {
    for _, ann := range annotations {
        if ann.Name() == "format" && ann.Value() == "email" {
            return f.formatAsEmail(value)
        }
    }
    return fmt.Sprintf("%v", value), nil
}

func (f *EmailFormatter) formatAsEmail(value any) (string, error) {
    str := fmt.Sprintf("%v", value)
    return strings.ToLower(strings.TrimSpace(str)), nil
}
```

### **Phase 8: Consumer Selection Usage Patterns**

#### **8.1 Single Purpose Selection - SOLVES THE CORE PROBLEM**

```go
// ✅ SOLUTION: Only run validators (no formatters, generators, etc.)
validationResult, err := registry.ProcessWithPurpose(
    PurposeValidation, 
    "user@EXAMPLE.com", 
    []core.Annotation{CreateAnnotation("format", "email")},
)

// ✅ SOLUTION: Only run formatters (no validators, generators, etc.)
formattedValue, err := registry.ProcessWithPurpose(
    PurposeFormatting,
    "  USER@EXAMPLE.COM  ",
    []core.Annotation{CreateAnnotation("format", "email")},
)

// ✅ SOLUTION: Only run generators (no validators, formatters, etc.)
generatedCode, err := registry.ProcessWithPurpose(
    PurposeGeneration,
    userSchema,
    []core.Annotation{CreateAnnotation("format", "email")},
)
```

#### **8.2 Multiple Purpose Selection**

```go
// ✅ SOLUTION: Run both validation AND formatting (but not generation)
result, err := registry.ProcessWithPurposes(
    []ConsumerPurpose{PurposeValidation, PurposeFormatting},
    "  USER@EXAMPLE.COM  ",
    []core.Annotation{CreateAnnotation("format", "email")},
)

// Access results by purpose
if validationResult, ok := result.Results[PurposeValidation].(ValidationResult); ok {
    fmt.Printf("Valid: %t\n", validationResult.Valid)
}

if formattedValue, ok := result.Results[PurposeFormatting].(string); ok {
    fmt.Printf("Formatted: %s\n", formattedValue)
}
```

#### **8.3 Context-Driven Processing**

```go
// ✅ SOLUTION: Rich context with purpose selection
ctx := ProcessingContext{
    Value: "  USER@EXAMPLE.COM  ",
    Annotations: []core.Annotation{
        CreateAnnotation("format", "email"),
        CreateAnnotation("description", "User email address"),
    },
    Purposes: []ConsumerPurpose{PurposeValidation, PurposeFormatting},
    Options: map[string]any{
        "strict_validation": true,
        "normalize_case":    true,
    },
}

result, err := registry.ProcessWithContext(ctx)
```

#### **8.4 Explicit Registry Access (Backward Compatible)**

```go
// ✅ SOLUTION: Explicit access to specialized registries
validationResult := registry.Validators().ValidateWithAnnotations(value, annotations)
formattedValue := registry.Formatters().FormatWithAnnotations(value, annotations)
generatedCode := registry.Generators().GenerateWithAnnotations(schema, annotations)
```

### **Phase 9: Testing Strategy**

#### **9.1 Test Annotations Independently**

```go
func TestAnnotationCreation(t *testing.T) {
    ann := CreateAnnotation("format", "email")
    assert.Equal(t, "format", ann.Name())
    assert.Equal(t, "email", ann.Value())
}
```

#### **9.2 Test Consumers Independently**

```go
func TestEmailValidator(t *testing.T) {
    validator := NewEmailValidator()
    annotations := []core.Annotation{
        CreateAnnotation("format", "email"),
    }
    
    result := validator.ValidateWithAnnotations("test@example.com", annotations)
    assert.True(t, result.Valid)
}
```

#### **9.3 Test Consumer Selection**

```go
func TestConsumerSelection(t *testing.T) {
    registry := NewConsumerRegistry()
    registry.Register(NewEmailValidator())
    registry.Register(NewEmailFormatter())
    
    annotations := []core.Annotation{CreateAnnotation("format", "email")}
    
    // Test single purpose selection
    validationResult, err := registry.ProcessWithPurpose(PurposeValidation, "test@example.com", annotations)
    assert.NoError(t, err)
    assert.True(t, validationResult.(ValidationResult).Valid)
    
    // Test multiple purpose selection
    result, err := registry.ProcessWithPurposes(
        []ConsumerPurpose{PurposeValidation, PurposeFormatting},
        "  TEST@EXAMPLE.COM  ",
        annotations,
    )
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Contains(t, result.Results, PurposeValidation)
    assert.Contains(t, result.Results, PurposeFormatting)
}
```

## 🚀 **Migration Strategy**

### **Phase-by-Phase Rollout**

1. **Phase 1-2**: Core interface changes (breaking changes)
2. **Phase 3**: Add new consumer interfaces with purpose declaration (additive)
3. **Phase 4**: Update schema implementations (breaking changes)
4. **Phase 5**: Update builders (breaking changes for some methods)
5. **Phase 6**: Add standard annotations (additive)
6. **Phase 7**: Update existing consumers with purpose support (breaking changes)
7. **Phase 8**: Add consumer selection usage patterns (additive)
8. **Phase 9**: Add comprehensive tests including consumer selection

### **Backward Compatibility**

- **Deprecation Period**: Mark old methods as deprecated with migration guidance
- **Adapter Pattern**: Provide adapters for old validator interfaces
- **Documentation**: Clear migration guide with examples
- **Gradual Migration**: Allow both old and new systems to coexist temporarily

## 📊 **Success Metrics**

### **Code Quality**
- ✅ Reduced coupling between annotations and validation
- ✅ Increased testability of individual components
- ✅ Cleaner, more focused interfaces

### **Extensibility**
- ✅ Easy to add new annotation consumers (formatters, generators, etc.)
- ✅ Same annotation can serve multiple purposes
- ✅ New annotation types can be added without changing existing code

### **Developer Experience**
- ✅ Clearer separation of concerns
- ✅ Better error messages and debugging
- ✅ More intuitive API for adding annotations

## 🎉 **Expected Benefits**

### **For Developers**
- **Cleaner Code**: Purpose-agnostic annotations with clear consumer interfaces
- **Better Testing**: Independent testing of annotations and consumers
- **Easier Extension**: Simple to add new annotation consumers
- **Clear Architecture**: Well-defined separation between metadata and processing

### **For Users**
- **More Features**: Same annotations work for validation, formatting, generation, docs
- **Better Documentation**: Annotations drive automatic documentation generation
- **Consistent Experience**: Unified annotation system across all features
- **Future-Proof**: Easy to add new capabilities without breaking changes

### **For the Ecosystem**
- **Plugin Architecture**: Third-party consumers can easily integrate
- **Standardization**: Common annotation format across different tools
- **Interoperability**: Annotations work across different parts of the system
- **Innovation**: Easy to experiment with new annotation-driven features

---

This refactoring will transform the annotation system from a **validation-specific** implementation to a **truly generic, extensible metadata system** that can power validation, formatting, code generation, documentation, and any future annotation-driven features. 