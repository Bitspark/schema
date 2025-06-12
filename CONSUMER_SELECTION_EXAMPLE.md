# Consumer Selection Architecture - Hybrid Approach

This document demonstrates the hybrid consumer selection approach that combines **Consumer Type Filtering** with **Explicit Selection** for purpose-agnostic annotations.

## Core Interfaces

### 1. Enhanced Consumer Interface

```go
// schema/api/consumer.go
package api

import "defs.dev/schema/api/core"

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

// AnnotationConsumer defines the interface for any system that processes annotations
type AnnotationConsumer interface {
    // Identity and capabilities
    Name() string
    Purpose() ConsumerPurpose
    SupportedAnnotations() []string
    
    // Processing
    ProcessAnnotations(context any, annotations []core.Annotation) (any, error)
    
    // Metadata
    Metadata() ConsumerMetadata
}

// ConsumerMetadata provides information about a consumer
type ConsumerMetadata struct {
    Name           string            `json:"name"`
    Purpose        ConsumerPurpose   `json:"purpose"`
    Description    string            `json:"description"`
    Version        string            `json:"version"`
    SupportedTypes []string          `json:"supported_types"`
    Tags           []string          `json:"tags"`
    Properties     map[string]string `json:"properties,omitempty"`
}

// ProcessingContext provides context for annotation processing
type ProcessingContext struct {
    Value       any                  `json:"value"`
    Schema      core.Schema          `json:"schema,omitempty"`
    Annotations []core.Annotation    `json:"annotations"`
    Purposes    []ConsumerPurpose    `json:"purposes"`
    Options     map[string]any       `json:"options,omitempty"`
    Metadata    map[string]any       `json:"metadata,omitempty"`
}

// ProcessingResult contains results from multiple consumer types
type ProcessingResult struct {
    Success       bool                           `json:"success"`
    Results       map[ConsumerPurpose]any        `json:"results"`
    Errors        map[ConsumerPurpose]error      `json:"errors,omitempty"`
    Metadata      map[string]any                 `json:"metadata,omitempty"`
    ExecutedAt    time.Time                      `json:"executed_at"`
    Duration      time.Duration                  `json:"duration"`
}
```

### 2. Enhanced Registry Interface

```go
// ConsumerRegistry manages annotation consumers with purpose-based selection
type ConsumerRegistry interface {
    // Consumer management
    Register(consumer AnnotationConsumer) error
    Get(name string) (AnnotationConsumer, bool)
    List() []string
    ListByPurpose(purpose ConsumerPurpose) []string
    
    // Purpose-based processing
    ProcessWithPurpose(purpose ConsumerPurpose, context any, annotations []core.Annotation) (any, error)
    ProcessWithPurposes(purposes []ConsumerPurpose, context any, annotations []core.Annotation) (ProcessingResult, error)
    ProcessWithContext(ctx ProcessingContext) (ProcessingResult, error)
    
    // Discovery
    GetConsumersForAnnotation(annotationName string) []AnnotationConsumer
    GetConsumersForPurpose(purpose ConsumerPurpose) []AnnotationConsumer
    GetConsumersForAnnotationAndPurpose(annotationName string, purpose ConsumerPurpose) []AnnotationConsumer
    
    // Explicit registry access
    Validators() ValidatorRegistry
    Formatters() FormatterRegistry
    Generators() GeneratorRegistry
    Documenters() DocumentationRegistry
}

// Specialized registries for explicit selection
type ValidatorRegistry interface {
    ValidateWithAnnotations(value any, annotations []core.Annotation) ValidationResult
    GetValidatorsForAnnotations(annotations []core.Annotation) []Validator
}

type FormatterRegistry interface {
    FormatWithAnnotations(value any, annotations []core.Annotation) (string, error)
    GetFormattersForAnnotations(annotations []core.Annotation) []Formatter
}

type GeneratorRegistry interface {
    GenerateWithAnnotations(schema core.Schema, annotations []core.Annotation) ([]byte, error)
    GetGeneratorsForAnnotations(annotations []core.Annotation) []Generator
}

type DocumentationRegistry interface {
    GenerateDocumentationWithAnnotations(schema core.Schema, annotations []core.Annotation) (string, error)
    GetDocumentersForAnnotations(annotations []core.Annotation) []Documenter
}
```

## Consumer Implementations

### 1. Validator Implementation

```go
// registry/email_validator.go
package registry

import (
    "net/mail"
    "defs.dev/schema/api"
    "defs.dev/schema/api/core"
)

// EmailValidator validates email addresses
type EmailValidator struct {
    name     string
    metadata api.ConsumerMetadata
}

func NewEmailValidator() *EmailValidator {
    return &EmailValidator{
        name: "email",
        metadata: api.ConsumerMetadata{
            Name:           "email",
            Purpose:        api.PurposeValidation,
            Description:    "Validates email address format",
            Version:        "1.0.0",
            SupportedTypes: []string{"string"},
            Tags:           []string{"validation", "format", "email"},
        },
    }
}

func (v *EmailValidator) Name() string {
    return v.name
}

func (v *EmailValidator) Purpose() api.ConsumerPurpose {
    return api.PurposeValidation
}

func (v *EmailValidator) SupportedAnnotations() []string {
    return []string{"format"}
}

func (v *EmailValidator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    value := context
    
    for _, ann := range annotations {
        if ann.Name() == "format" && ann.Value() == "email" {
            return v.validateEmail(value)
        }
    }
    
    return ValidationResult{Valid: true}, nil
}

func (v *EmailValidator) validateEmail(value any) ValidationResult {
    str, ok := value.(string)
    if !ok {
        return ValidationResult{
            Valid: false,
            Errors: []ValidationError{{
                ValidatorName: v.name,
                Message:       "value must be a string",
                Code:          "type_error",
            }},
        }
    }
    
    _, err := mail.ParseAddress(str)
    if err != nil {
        return ValidationResult{
            Valid: false,
            Errors: []ValidationError{{
                ValidatorName: v.name,
                Message:       "invalid email format",
                Code:          "format_invalid",
                Value:         str,
            }},
        }
    }
    
    return ValidationResult{Valid: true}
}

func (v *EmailValidator) Metadata() api.ConsumerMetadata {
    return v.metadata
}
```

### 2. Formatter Implementation

```go
// formatters/email_formatter.go
package formatters

import (
    "strings"
    "defs.dev/schema/api"
    "defs.dev/schema/api/core"
)

// EmailFormatter formats email addresses
type EmailFormatter struct {
    name     string
    metadata api.ConsumerMetadata
}

func NewEmailFormatter() *EmailFormatter {
    return &EmailFormatter{
        name: "email",
        metadata: api.ConsumerMetadata{
            Name:           "email",
            Purpose:        api.PurposeFormatting,
            Description:    "Formats email addresses (lowercase, trim)",
            Version:        "1.0.0",
            SupportedTypes: []string{"string"},
            Tags:           []string{"formatting", "email", "normalization"},
        },
    }
}

func (f *EmailFormatter) Name() string {
    return f.name
}

func (f *EmailFormatter) Purpose() api.ConsumerPurpose {
    return api.PurposeFormatting
}

func (f *EmailFormatter) SupportedAnnotations() []string {
    return []string{"format"}
}

func (f *EmailFormatter) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    value := context
    
    for _, ann := range annotations {
        if ann.Name() == "format" && ann.Value() == "email" {
            return f.formatEmail(value)
        }
    }
    
    return value, nil
}

func (f *EmailFormatter) formatEmail(value any) (string, error) {
    str, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("value must be a string")
    }
    
    // Normalize email: lowercase and trim
    formatted := strings.ToLower(strings.TrimSpace(str))
    return formatted, nil
}

func (f *EmailFormatter) Metadata() api.ConsumerMetadata {
    return f.metadata
}
```

### 3. Generator Implementation

```go
// generators/typescript_generator.go
package generators

import (
    "fmt"
    "strings"
    "defs.dev/schema/api"
    "defs.dev/schema/api/core"
)

// TypeScriptGenerator generates TypeScript type definitions
type TypeScriptGenerator struct {
    name     string
    metadata api.ConsumerMetadata
}

func NewTypeScriptGenerator() *TypeScriptGenerator {
    return &TypeScriptGenerator{
        name: "typescript",
        metadata: api.ConsumerMetadata{
            Name:           "typescript",
            Purpose:        api.PurposeGeneration,
            Description:    "Generates TypeScript type definitions",
            Version:        "1.0.0",
            SupportedTypes: []string{"string", "number", "boolean", "object", "array"},
            Tags:           []string{"generation", "typescript", "types"},
        },
    }
}

func (g *TypeScriptGenerator) Name() string {
    return g.name
}

func (g *TypeScriptGenerator) Purpose() api.ConsumerPurpose {
    return api.PurposeGeneration
}

func (g *TypeScriptGenerator) SupportedAnnotations() []string {
    return []string{"format", "pattern", "minLength", "maxLength", "description"}
}

func (g *TypeScriptGenerator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    schema, ok := context.(core.Schema)
    if !ok {
        return nil, fmt.Errorf("context must be a schema for code generation")
    }
    
    return g.generateTypeScript(schema, annotations)
}

func (g *TypeScriptGenerator) generateTypeScript(schema core.Schema, annotations []core.Annotation) (string, error) {
    var tsType string
    var comments []string
    
    // Base type
    switch schema.Type() {
    case core.TypeString:
        tsType = "string"
    case core.TypeNumber:
        tsType = "number"
    case core.TypeBoolean:
        tsType = "boolean"
    default:
        tsType = "any"
    }
    
    // Process annotations for additional constraints
    for _, ann := range annotations {
        switch ann.Name() {
        case "format":
            if ann.Value() == "email" {
                comments = append(comments, "@format email")
            }
        case "description":
            if desc, ok := ann.Value().(string); ok {
                comments = append(comments, fmt.Sprintf("@description %s", desc))
            }
        case "pattern":
            if pattern, ok := ann.Value().(string); ok {
                comments = append(comments, fmt.Sprintf("@pattern %s", pattern))
            }
        }
    }
    
    // Build TypeScript definition
    var result strings.Builder
    if len(comments) > 0 {
        result.WriteString("/**\n")
        for _, comment := range comments {
            result.WriteString(fmt.Sprintf(" * %s\n", comment))
        }
        result.WriteString(" */\n")
    }
    result.WriteString(tsType)
    
    return result.String(), nil
}

func (g *TypeScriptGenerator) Metadata() api.ConsumerMetadata {
    return g.metadata
}
```

## Usage Examples

### 1. Purpose-Based Selection

```go
func ExamplePurposeBasedSelection() {
    registry := NewConsumerRegistry()
    
    // Register consumers
    registry.Register(NewEmailValidator())
    registry.Register(NewEmailFormatter())
    registry.Register(NewTypeScriptGenerator())
    
    // Create annotations
    annotations := []core.Annotation{
        CreateAnnotation("format", "email"),
        CreateAnnotation("description", "User email address"),
    }
    
    value := "  USER@EXAMPLE.COM  "
    
    // 1. Only validation
    validationResult, err := registry.ProcessWithPurpose(
        api.PurposeValidation, 
        value, 
        annotations,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if vr, ok := validationResult.(ValidationResult); ok && vr.Valid {
        fmt.Println("✅ Email is valid")
    }
    
    // 2. Only formatting
    formattedValue, err := registry.ProcessWithPurpose(
        api.PurposeFormatting, 
        value, 
        annotations,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("📝 Formatted: %s\n", formattedValue) // "user@example.com"
    
    // 3. Only code generation
    schema := builders.NewStringSchema().Build()
    generatedCode, err := registry.ProcessWithPurpose(
        api.PurposeGeneration, 
        schema, 
        annotations,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("🔧 Generated TypeScript:\n%s\n", generatedCode)
}
```

### 2. Multiple Purposes

```go
func ExampleMultiplePurposes() {
    registry := NewConsumerRegistry()
    
    // Register consumers
    registry.Register(NewEmailValidator())
    registry.Register(NewEmailFormatter())
    
    annotations := []core.Annotation{
        CreateAnnotation("format", "email"),
    }
    
    value := "  USER@EXAMPLE.COM  "
    
    // Process with multiple purposes
    result, err := registry.ProcessWithPurposes(
        []api.ConsumerPurpose{api.PurposeValidation, api.PurposeFormatting},
        value,
        annotations,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Check validation result
    if validationResult, ok := result.Results[api.PurposeValidation].(ValidationResult); ok {
        if validationResult.Valid {
            fmt.Println("✅ Validation passed")
        } else {
            fmt.Printf("❌ Validation failed: %v\n", validationResult.Errors)
        }
    }
    
    // Get formatted value
    if formattedValue, ok := result.Results[api.PurposeFormatting].(string); ok {
        fmt.Printf("📝 Formatted value: %s\n", formattedValue)
    }
    
    fmt.Printf("⏱️  Processing took: %v\n", result.Duration)
}
```

### 3. Context-Driven Processing

```go
func ExampleContextDrivenProcessing() {
    registry := NewConsumerRegistry()
    
    // Register consumers
    registry.Register(NewEmailValidator())
    registry.Register(NewEmailFormatter())
    registry.Register(NewTypeScriptGenerator())
    
    // Create processing context
    ctx := api.ProcessingContext{
        Value: "  USER@EXAMPLE.COM  ",
        Annotations: []core.Annotation{
            CreateAnnotation("format", "email"),
            CreateAnnotation("description", "User email address"),
        },
        Purposes: []api.ConsumerPurpose{
            api.PurposeValidation,
            api.PurposeFormatting,
        },
        Options: map[string]any{
            "strict_validation": true,
            "normalize_case":    true,
        },
    }
    
    result, err := registry.ProcessWithContext(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process results
    for purpose, res := range result.Results {
        fmt.Printf("Purpose %s: %v\n", purpose, res)
    }
}
```

### 4. Explicit Registry Access

```go
func ExampleExplicitRegistryAccess() {
    registry := NewConsumerRegistry()
    
    // Register consumers
    registry.Register(NewEmailValidator())
    registry.Register(NewEmailFormatter())
    
    annotations := []core.Annotation{
        CreateAnnotation("format", "email"),
    }
    
    value := "user@example.com"
    
    // Explicit validator access
    validationResult := registry.Validators().ValidateWithAnnotations(value, annotations)
    if validationResult.Valid {
        fmt.Println("✅ Email validation passed")
    }
    
    // Explicit formatter access
    formattedValue, err := registry.Formatters().FormatWithAnnotations(value, annotations)
    if err == nil {
        fmt.Printf("📝 Formatted email: %s\n", formattedValue)
    }
}
```

### 5. Builder Pattern Selection

```go
func ExampleBuilderPatternSelection() {
    registry := NewConsumerRegistry()
    
    // Register consumers
    registry.Register(NewEmailValidator())
    registry.Register(NewEmailFormatter())
    registry.Register(NewTypeScriptGenerator())
    
    annotations := []core.Annotation{
        CreateAnnotation("format", "email"),
    }
    
    value := "  USER@EXAMPLE.COM  "
    
    // Fluent builder pattern
    selector := NewConsumerSelector(registry, value, annotations)
    
    // Only validation
    result := selector.Validate().Execute()
    
    // Validation + formatting
    result = selector.Validate().Format().Execute()
    
    // All purposes
    result = selector.Validate().Format().Generate().Execute()
    
    fmt.Printf("Results: %+v\n", result)
}

// ConsumerSelector for fluent API
type ConsumerSelector struct {
    registry    api.ConsumerRegistry
    value       any
    annotations []core.Annotation
    purposes    []api.ConsumerPurpose
}

func NewConsumerSelector(registry api.ConsumerRegistry, value any, annotations []core.Annotation) *ConsumerSelector {
    return &ConsumerSelector{
        registry:    registry,
        value:       value,
        annotations: annotations,
        purposes:    []api.ConsumerPurpose{},
    }
}

func (s *ConsumerSelector) Validate() *ConsumerSelector {
    s.purposes = append(s.purposes, api.PurposeValidation)
    return s
}

func (s *ConsumerSelector) Format() *ConsumerSelector {
    s.purposes = append(s.purposes, api.PurposeFormatting)
    return s
}

func (s *ConsumerSelector) Generate() *ConsumerSelector {
    s.purposes = append(s.purposes, api.PurposeGeneration)
    return s
}

func (s *ConsumerSelector) Execute() api.ProcessingResult {
    result, _ := s.registry.ProcessWithPurposes(s.purposes, s.value, s.annotations)
    return result
}
```

## Benefits of This Approach

### ✅ **Clear Purpose Separation**
- Each consumer declares its purpose explicitly
- No accidental cross-purpose execution
- Easy to understand what each consumer does

### ✅ **Flexible Selection**
- Choose specific purposes: validation only, formatting only, etc.
- Combine multiple purposes as needed
- Context-driven processing for complex scenarios

### ✅ **Performance Efficient**
- Only runs requested consumer types
- No wasted processing on irrelevant consumers
- Parallel processing potential for independent purposes

### ✅ **Backward Compatible**
- Existing validator interfaces can be preserved
- Gradual migration path
- Legacy code continues to work

### ✅ **Extensible**
- Easy to add new consumer purposes
- Plugin architecture for custom consumers
- Rich metadata system for discovery

### ✅ **Developer Friendly**
- Multiple usage patterns (explicit, fluent, context-driven)
- Clear error handling and result types
- Rich debugging and introspection capabilities

This hybrid approach gives you the control you need: when you want validation, you get only validation. When you want formatting, you get only formatting. And when you want both, you can explicitly request both with clear, predictable results. 