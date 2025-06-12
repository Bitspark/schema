# Consumer Selection Architecture for Purpose-Agnostic Annotations

## Overview

This document describes the **Hybrid Consumer Selection Architecture** - a solution that enables precise control over which annotation consumers execute while maintaining purpose-agnostic annotations. The architecture solves the fundamental problem of unwanted consumer execution (e.g., formatters running when you only want validation) through purpose-based filtering and explicit selection patterns.

## The Problem

### Current System Issues

The existing annotation system has a **consumer selection problem**:

```go
// ❌ PROBLEM: All consumers run automatically
registry.ValidateWithAnnotations(value, annotations)
// This executes validators, formatters, generators, documentation generators, etc.
// You can't choose "only validation" or "only formatting"
```

**Issues:**
- ❌ **No Purpose Filtering**: When you want validation, formatters also run
- ❌ **All-or-Nothing**: Can't selectively invoke consumer types  
- ❌ **Performance Waste**: Unnecessary processing of irrelevant consumers
- ❌ **Result Mixing**: Validation, formatting, and generation results get mixed
- ❌ **Poor Developer Experience**: Unpredictable behavior and resource usage

### Real-World Scenarios

**Scenario 1: API Validation**
```go
// You want to validate user input before saving to database
// But you DON'T want formatting, generation, or documentation
result := validateUserEmail("user@EXAMPLE.com", emailAnnotations)
// Current system: Runs validators + formatters + generators + documenters
// Desired: Only run validators
```

**Scenario 2: Data Formatting**
```go
// You want to normalize email format for storage
// But you DON'T want validation (already validated) or generation
formatted := formatUserEmail("USER@EXAMPLE.COM", emailAnnotations)  
// Current system: Runs validators + formatters + generators + documenters
// Desired: Only run formatters
```

**Scenario 3: Code Generation**
```go
// You want to generate TypeScript types from schema
// But you DON'T want validation or formatting of runtime values
types := generateTypes(userSchema, emailAnnotations)
// Current system: Runs validators + formatters + generators + documenters  
// Desired: Only run generators
```

## The Solution: Hybrid Consumer Selection Architecture

### Core Innovation: Purpose-Driven Consumer Selection

The hybrid approach combines **Consumer Type Filtering** with **Explicit Selection** to provide precise control over consumer execution.

#### Key Principles

1. **Purpose-Agnostic Annotations**: Annotations remain pure metadata containers
2. **Purpose-Aware Consumers**: Consumers declare their processing purpose
3. **Selective Execution**: Registry filters consumers by requested purpose
4. **Multiple Selection Patterns**: Support different usage styles and migration paths

### Architecture Components

#### 1. Consumer Purpose Declaration

```go
// Enhanced consumer interface with purpose identification
type AnnotationConsumer interface {
    // Core identity
    Name() string
    Purpose() ConsumerPurpose  // 🆕 NEW: Declares what this consumer does
    SupportedAnnotations() []string
    
    // Processing
    ProcessAnnotations(context any, annotations []core.Annotation) (any, error)
    
    // Metadata
    Metadata() ConsumerMetadata
}

// Purpose enumeration
type ConsumerPurpose string

const (
    PurposeValidation    ConsumerPurpose = "validation"
    PurposeFormatting    ConsumerPurpose = "formatting"
    PurposeGeneration    ConsumerPurpose = "generation"
    PurposeDocumentation ConsumerPurpose = "documentation"
    PurposeTransform     ConsumerPurpose = "transform"
    PurposeAnalysis      ConsumerPurpose = "analysis"
)
```

#### 2. Purpose-Based Registry

```go
// Enhanced registry with purpose-based selection
type ConsumerRegistry interface {
    // Consumer management
    Register(consumer AnnotationConsumer) error
    Get(name string) (AnnotationConsumer, bool)
    List() []string
    ListByPurpose(purpose ConsumerPurpose) []string
    
    // 🎯 PURPOSE-BASED PROCESSING - THE KEY INNOVATION
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

#### 3. Processing Context and Results

```go
// Context for complex processing scenarios
type ProcessingContext struct {
    Value       any                  `json:"value"`
    Schema      core.Schema          `json:"schema,omitempty"`
    Annotations []core.Annotation    `json:"annotations"`
    Purposes    []ConsumerPurpose    `json:"purposes"`
    Options     map[string]any       `json:"options,omitempty"`
    Metadata    map[string]any       `json:"metadata,omitempty"`
}

// Rich result structure with purpose-specific results
type ProcessingResult struct {
    Success       bool                           `json:"success"`
    Results       map[ConsumerPurpose]any        `json:"results"`
    Errors        map[ConsumerPurpose]error      `json:"errors,omitempty"`
    Metadata      map[string]any                 `json:"metadata,omitempty"`
    ExecutedAt    time.Time                      `json:"executed_at"`
    Duration      time.Duration                  `json:"duration"`
}
```

## Usage Patterns

### Pattern 1: Single Purpose Selection

**Use Case**: You need exactly one type of processing.

```go
// ✅ SOLUTION: Only run validators
validationResult, err := registry.ProcessWithPurpose(
    PurposeValidation, 
    "user@EXAMPLE.com", 
    []core.Annotation{CreateAnnotation("format", "email")},
)

if vr, ok := validationResult.(ValidationResult); ok && vr.Valid {
    fmt.Println("✅ Email is valid")
}

// ✅ SOLUTION: Only run formatters  
formattedValue, err := registry.ProcessWithPurpose(
    PurposeFormatting,
    "  USER@EXAMPLE.COM  ",
    []core.Annotation{CreateAnnotation("format", "email")},
)
fmt.Printf("📝 Formatted: %s\n", formattedValue) // "user@example.com"

// ✅ SOLUTION: Only run generators
generatedCode, err := registry.ProcessWithPurpose(
    PurposeGeneration,
    userSchema,
    []core.Annotation{
        CreateAnnotation("format", "email"),
        CreateAnnotation("description", "User email address"),
    },
)
fmt.Printf("🔧 Generated TypeScript:\n%s\n", generatedCode)
```

**Benefits:**
- 🎯 **Precise Control**: Get exactly what you ask for
- ⚡ **Performance**: No wasted processing
- 🔍 **Predictable**: Clear input/output contract

### Pattern 2: Multiple Purpose Selection

**Use Case**: You need several types of processing in one operation.

```go
// ✅ SOLUTION: Run both validation AND formatting
result, err := registry.ProcessWithPurposes(
    []ConsumerPurpose{PurposeValidation, PurposeFormatting},
    "  USER@EXAMPLE.COM  ",
    []core.Annotation{CreateAnnotation("format", "email")},
)

if err != nil {
    log.Printf("Processing error: %v", err)
    return
}

fmt.Printf("Processing success: %t\n", result.Success)
fmt.Printf("Processing duration: %v\n", result.Duration)

// Access results by purpose
if validationResult, ok := result.Results[PurposeValidation].(ValidationResult); ok {
    if validationResult.Valid {
        fmt.Println("✅ Validation passed")
    } else {
        fmt.Printf("❌ Validation failed: %v\n", validationResult.Errors)
    }
}

if formattedValue, ok := result.Results[PurposeFormatting].(string); ok {
    fmt.Printf("📝 Formatted value: %s\n", formattedValue)
}

// Check for purpose-specific errors
if validationError, exists := result.Errors[PurposeValidation]; exists {
    fmt.Printf("Validation error: %v\n", validationError)
}
```

**Benefits:**
- 🔄 **Efficient Batching**: Multiple operations in one call
- 📊 **Rich Results**: Purpose-specific results and errors
- ⏱️ **Performance Tracking**: Built-in timing and success metrics

### Pattern 3: Context-Driven Processing

**Use Case**: Complex scenarios with configuration and metadata.

```go
// ✅ SOLUTION: Rich context with options and metadata
ctx := ProcessingContext{
    Value: "  USER@EXAMPLE.COM  ",
    Schema: userEmailSchema,
    Annotations: []core.Annotation{
        CreateAnnotation("format", "email"),
        CreateAnnotation("description", "User email address"),
        CreateAnnotation("required", true),
    },
    Purposes: []ConsumerPurpose{
        PurposeValidation,
        PurposeFormatting,
        PurposeGeneration,
    },
    Options: map[string]any{
        "strict_validation": true,
        "normalize_case":    true,
        "generate_docs":     true,
    },
    Metadata: map[string]any{
        "request_id": "req-123",
        "user_id":    "user-456",
    },
}

result, err := registry.ProcessWithContext(ctx)
if err != nil {
    log.Printf("Context processing error: %v", err)
    return
}

// Process results with rich context
fmt.Printf("Request %s completed in %v\n", 
    ctx.Metadata["request_id"], result.Duration)

for purpose, res := range result.Results {
    fmt.Printf("Purpose %s: %v\n", purpose, res)
}
```

**Benefits:**
- 🎛️ **Rich Configuration**: Options and metadata support
- 📝 **Audit Trail**: Request tracking and logging
- 🔧 **Flexible**: Supports complex business logic

### Pattern 4: Explicit Registry Access (Backward Compatible)

**Use Case**: Teams prefer explicit interfaces or gradual migration.

```go
// ✅ SOLUTION: Explicit access to specialized registries
registry := NewConsumerRegistry()

// Register consumers
registry.Register(NewEmailValidator())
registry.Register(NewEmailFormatter())
registry.Register(NewTypeScriptGenerator())

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

// Explicit generator access
generatedCode, err := registry.Generators().GenerateWithAnnotations(schema, annotations)
if err == nil {
    fmt.Printf("🔧 Generated code: %s\n", generatedCode)
}
```

**Benefits:**
- 🔄 **Backward Compatible**: Existing code continues to work
- 📚 **Familiar**: Traditional registry pattern
- 🛡️ **Type Safe**: Specialized interfaces with specific return types

### Pattern 5: Builder Pattern Selection (Advanced)

**Use Case**: Fluent API for complex selection scenarios.

```go
// ✅ SOLUTION: Fluent builder pattern
selector := NewConsumerSelector(registry, value, annotations)

// Only validation
result := selector.Validate().Execute()

// Validation + formatting
result = selector.Validate().Format().Execute()

// All purposes
result = selector.Validate().Format().Generate().Document().Execute()

// Conditional selection
selector = NewConsumerSelector(registry, value, annotations)
if needsValidation {
    selector = selector.Validate()
}
if needsFormatting {
    selector = selector.Format()
}
result = selector.Execute()

fmt.Printf("Results: %+v\n", result)
```

**Implementation:**
```go
// ConsumerSelector for fluent API
type ConsumerSelector struct {
    registry    ConsumerRegistry
    value       any
    annotations []core.Annotation
    purposes    []ConsumerPurpose
}

func NewConsumerSelector(registry ConsumerRegistry, value any, annotations []core.Annotation) *ConsumerSelector {
    return &ConsumerSelector{
        registry:    registry,
        value:       value,
        annotations: annotations,
        purposes:    []ConsumerPurpose{},
    }
}

func (s *ConsumerSelector) Validate() *ConsumerSelector {
    s.purposes = append(s.purposes, PurposeValidation)
    return s
}

func (s *ConsumerSelector) Format() *ConsumerSelector {
    s.purposes = append(s.purposes, PurposeFormatting)
    return s
}

func (s *ConsumerSelector) Generate() *ConsumerSelector {
    s.purposes = append(s.purposes, PurposeGeneration)
    return s
}

func (s *ConsumerSelector) Document() *ConsumerSelector {
    s.purposes = append(s.purposes, PurposeDocumentation)
    return s
}

func (s *ConsumerSelector) Execute() ProcessingResult {
    result, _ := s.registry.ProcessWithPurposes(s.purposes, s.value, s.annotations)
    return result
}
```

**Benefits:**
- 🎨 **Fluent API**: Readable and expressive
- 🔧 **Flexible**: Dynamic purpose selection
- 🎯 **Discoverable**: IDE autocomplete shows available purposes

## Consumer Implementation Guide

### Basic Consumer Structure

```go
// Example: EmailValidator
type EmailValidator struct {
    name     string
    metadata ConsumerMetadata
}

func NewEmailValidator() *EmailValidator {
    return &EmailValidator{
        name: "email",
        metadata: ConsumerMetadata{
            Name:           "email",
            Purpose:        PurposeValidation,  // 🎯 Declares purpose
            Description:    "Validates email address format",
            Version:        "1.0.0",
            SupportedTypes: []string{"string"},
            Tags:           []string{"validation", "format", "email"},
        },
    }
}

// Required interface methods
func (v *EmailValidator) Name() string                        { return v.name }
func (v *EmailValidator) Purpose() ConsumerPurpose           { return PurposeValidation }
func (v *EmailValidator) SupportedAnnotations() []string     { return []string{"format"} }
func (v *EmailValidator) Metadata() ConsumerMetadata         { return v.metadata }

// Core processing logic
func (v *EmailValidator) ProcessAnnotations(context any, annotations []core.Annotation) (any, error) {
    for _, ann := range annotations {
        if ann.Name() == "format" && ann.Value() == "email" {
            return v.validateEmail(context)
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
```

### Consumer Types by Purpose

#### Validation Consumers
```go
// Purpose: PurposeValidation
// Input: Value to validate + annotations
// Output: ValidationResult
type EmailValidator struct { /* ... */ }
type URLValidator struct { /* ... */ }
type PatternValidator struct { /* ... */ }
type RangeValidator struct { /* ... */ }
```

#### Formatting Consumers
```go
// Purpose: PurposeFormatting  
// Input: Value to format + annotations
// Output: Formatted value (usually string)
type EmailFormatter struct { /* ... */ }
type PhoneFormatter struct { /* ... */ }
type DateFormatter struct { /* ... */ }
type CurrencyFormatter struct { /* ... */ }
```

#### Generation Consumers
```go
// Purpose: PurposeGeneration
// Input: Schema + annotations
// Output: Generated code/content
type TypeScriptGenerator struct { /* ... */ }
type PythonGenerator struct { /* ... */ }
type JSONSchemaGenerator struct { /* ... */ }
type OpenAPIGenerator struct { /* ... */ }
```

#### Documentation Consumers
```go
// Purpose: PurposeDocumentation
// Input: Schema + annotations  
// Output: Documentation content
type MarkdownDocumenter struct { /* ... */ }
type HTMLDocumenter struct { /* ... */ }
type APIDocumenter struct { /* ... */ }
```

## Registry Implementation

### Core Registry Structure

```go
// DefaultConsumerRegistry implements ConsumerRegistry
type DefaultConsumerRegistry struct {
    consumers   map[string]AnnotationConsumer
    byPurpose   map[ConsumerPurpose][]AnnotationConsumer  // 🎯 Purpose index
    validators  ValidatorRegistry
    formatters  FormatterRegistry
    generators  GeneratorRegistry
    documenters DocumentationRegistry
    mu          sync.RWMutex
}

func NewConsumerRegistry() *DefaultConsumerRegistry {
    return &DefaultConsumerRegistry{
        consumers: make(map[string]AnnotationConsumer),
        byPurpose: make(map[ConsumerPurpose][]AnnotationConsumer),
        validators:  NewValidatorRegistry(),
        formatters:  NewFormatterRegistry(),
        generators:  NewGeneratorRegistry(),
        documenters: NewDocumentationRegistry(),
    }
}
```

### Registration Process

```go
func (r *DefaultConsumerRegistry) Register(consumer AnnotationConsumer) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    name := consumer.Name()
    if _, exists := r.consumers[name]; exists {
        return fmt.Errorf("consumer %s already registered", name)
    }
    
    r.consumers[name] = consumer
    
    // 🎯 Index by purpose for fast lookup
    purpose := consumer.Purpose()
    r.byPurpose[purpose] = append(r.byPurpose[purpose], consumer)
    
    // 🔄 Register with specialized registries for backward compatibility
    switch purpose {
    case PurposeValidation:
        if validator, ok := consumer.(Validator); ok {
            r.validators.Register(name, validator)
        }
    case PurposeFormatting:
        if formatter, ok := consumer.(Formatter); ok {
            r.formatters.Register(name, formatter)
        }
    case PurposeGeneration:
        if generator, ok := consumer.(Generator); ok {
            r.generators.Register(name, generator)
        }
    case PurposeDocumentation:
        if documenter, ok := consumer.(Documenter); ok {
            r.documenters.Register(name, documenter)
        }
    }
    
    return nil
}
```

### Purpose-Based Processing

```go
func (r *DefaultConsumerRegistry) ProcessWithPurpose(purpose ConsumerPurpose, context any, annotations []core.Annotation) (any, error) {
    r.mu.RLock()
    consumers := r.byPurpose[purpose]  // 🎯 Get consumers for specific purpose
    r.mu.RUnlock()
    
    var results []any
    
    for _, consumer := range consumers {
        // Check if consumer supports any of the annotations
        supported := false
        supportedAnnotations := consumer.SupportedAnnotations()
        
        for _, ann := range annotations {
            for _, supportedAnn := range supportedAnnotations {
                if ann.Name() == supportedAnn {
                    supported = true
                    break
                }
            }
            if supported {
                break
            }
        }
        
        if supported {
            result, err := consumer.ProcessAnnotations(context, annotations)
            if err != nil {
                return nil, fmt.Errorf("consumer %s failed: %w", consumer.Name(), err)
            }
            results = append(results, result)
        }
    }
    
    // Return appropriate result based on purpose
    switch purpose {
    case PurposeValidation:
        return r.combineValidationResults(results), nil
    case PurposeFormatting:
        return r.combineFormattingResults(results), nil
    default:
        if len(results) == 1 {
            return results[0], nil
        }
        return results, nil
    }
}
```

## Migration Strategy

### Phase 1: Add Purpose to Existing Consumers

**Goal**: Update existing consumers to declare their purpose.

```go
// BEFORE: Existing validator
type EmailValidator struct {
    name string
}

func (v *EmailValidator) Name() string { return v.name }
func (v *EmailValidator) SupportedAnnotations() []string { return []string{"format"} }
func (v *EmailValidator) ValidateWithAnnotations(value any, annotations []Annotation) ValidationResult {
    // existing logic
}

// AFTER: Add purpose declaration
func (v *EmailValidator) Purpose() ConsumerPurpose {
    return PurposeValidation  // 🆕 NEW: Declare purpose
}

func (v *EmailValidator) ProcessAnnotations(context any, annotations []Annotation) (any, error) {
    return v.ValidateWithAnnotations(context, annotations), nil  // 🔄 Adapt existing method
}

func (v *EmailValidator) Metadata() ConsumerMetadata {
    return ConsumerMetadata{
        Name:           v.name,
        Purpose:        PurposeValidation,
        Description:    "Validates email address format",
        Version:        "1.0.0",
        SupportedTypes: []string{"string"},
        Tags:           []string{"validation", "format", "email"},
    }
}
```

### Phase 2: Update Registry to Support Purpose-Based Selection

**Goal**: Enhance registry with purpose-based methods while maintaining backward compatibility.

```go
// BEFORE: Basic registry
type ValidatorRegistry interface {
    Register(name string, validator Validator) error
    ValidateWithAnnotations(value any, annotations []Annotation) ValidationResult
}

// AFTER: Enhanced registry with purpose support
type ConsumerRegistry interface {
    // 🔄 Existing methods (backward compatible)
    Register(consumer AnnotationConsumer) error
    ValidateWithAnnotations(value any, annotations []Annotation) ValidationResult
    
    // 🆕 NEW: Purpose-based methods
    ProcessWithPurpose(purpose ConsumerPurpose, context any, annotations []Annotation) (any, error)
    ProcessWithPurposes(purposes []ConsumerPurpose, context any, annotations []Annotation) (ProcessingResult, error)
    
    // 🆕 NEW: Explicit registry access
    Validators() ValidatorRegistry
    Formatters() FormatterRegistry
}
```

### Phase 3: Gradually Migrate Calling Code

**Goal**: Update application code to use purpose-based selection.

```go
// PHASE 3A: Keep existing code working
func validateUserInput(email string) bool {
    // ✅ OLD WAY: Still works
    result := registry.ValidateWithAnnotations(email, emailAnnotations)
    return result.Valid
}

// PHASE 3B: Introduce purpose-based calls
func validateUserInputNew(email string) bool {
    // ✅ NEW WAY: More precise
    result, err := registry.ProcessWithPurpose(PurposeValidation, email, emailAnnotations)
    if err != nil {
        return false
    }
    if vr, ok := result.(ValidationResult); ok {
        return vr.Valid
    }
    return false
}

// PHASE 3C: Use multiple purposes when needed
func processUserInput(email string) (bool, string, error) {
    result, err := registry.ProcessWithPurposes(
        []ConsumerPurpose{PurposeValidation, PurposeFormatting},
        email,
        emailAnnotations,
    )
    if err != nil {
        return false, "", err
    }
    
    valid := false
    if vr, ok := result.Results[PurposeValidation].(ValidationResult); ok {
        valid = vr.Valid
    }
    
    formatted := email
    if fr, ok := result.Results[PurposeFormatting].(string); ok {
        formatted = fr
    }
    
    return valid, formatted, nil
}
```

### Phase 4: Deprecate Old Methods (Optional)

**Goal**: Clean up API surface by deprecating old methods.

```go
// Mark old methods as deprecated
// Deprecated: Use ProcessWithPurpose(PurposeValidation, ...) instead
func (r *Registry) ValidateWithAnnotations(value any, annotations []Annotation) ValidationResult {
    result, _ := r.ProcessWithPurpose(PurposeValidation, value, annotations)
    return result.(ValidationResult)
}
```

## Benefits and Trade-offs

### ✅ Benefits

#### 1. **Precise Control**
- Choose exactly which consumers run
- No unwanted side effects
- Predictable resource usage

#### 2. **Performance Efficiency**
- Only run necessary consumers
- Reduce CPU and memory usage
- Faster response times

#### 3. **Clear Separation of Concerns**
- Validation logic separate from formatting
- Generation separate from runtime processing
- Each consumer has single responsibility

#### 4. **Flexible Usage Patterns**
- Single purpose selection
- Multiple purpose batching
- Context-driven processing
- Explicit registry access
- Fluent builder pattern

#### 5. **Backward Compatibility**
- Existing code continues to work
- Gradual migration path
- No breaking changes required

#### 6. **Rich Introspection**
- Discover consumers by purpose
- List supported annotations
- Debug consumer selection
- Performance monitoring

#### 7. **Extensibility**
- Easy to add new purposes
- Plugin architecture for custom consumers
- Rich metadata system

### ⚠️ Trade-offs

#### 1. **Increased Complexity**
- More interfaces and concepts
- Purpose management overhead
- Learning curve for developers

**Mitigation**: Comprehensive documentation, examples, and gradual migration

#### 2. **Registration Overhead**
- Consumers must declare purpose
- Registry indexing by purpose
- Additional metadata management

**Mitigation**: Automated registration helpers, clear patterns

#### 3. **API Surface Growth**
- More methods on registry interface
- Multiple usage patterns to maintain
- Potential for confusion

**Mitigation**: Clear documentation, consistent naming, deprecation strategy

## Future Extensions

### 1. **Parallel Processing**
```go
// Process multiple purposes in parallel
result, err := registry.ProcessWithPurposesParallel(
    []ConsumerPurpose{PurposeValidation, PurposeFormatting, PurposeGeneration},
    context,
    annotations,
)
```

### 2. **Conditional Processing**
```go
// Process with conditions
result, err := registry.ProcessWithConditions(
    map[ConsumerPurpose]func() bool{
        PurposeValidation: func() bool { return needsValidation },
        PurposeFormatting: func() bool { return needsFormatting },
    },
    context,
    annotations,
)
```

### 3. **Pipeline Processing**
```go
// Process in sequence with data flow
result, err := registry.ProcessPipeline(
    []ConsumerPurpose{PurposeValidation, PurposeFormatting, PurposeGeneration},
    context,
    annotations,
)
```

### 4. **Custom Purposes**
```go
// Allow custom purposes for domain-specific consumers
const (
    PurposeAuditLogging  ConsumerPurpose = "audit_logging"
    PurposeMetrics       ConsumerPurpose = "metrics"
    PurposeNotification  ConsumerPurpose = "notification"
)
```

## Conclusion

The **Hybrid Consumer Selection Architecture** solves the fundamental problem of unwanted consumer execution in annotation processing systems. By combining purpose-based filtering with multiple selection patterns, it provides:

- **🎯 Precise Control**: Choose exactly what processing you want
- **⚡ Performance**: Only run necessary consumers
- **🔄 Compatibility**: Smooth migration path from existing systems
- **🎨 Flexibility**: Multiple usage patterns for different needs
- **📈 Extensibility**: Easy to add new purposes and consumers

This architecture enables **purpose-agnostic annotations** (annotations don't know about consumers) while providing **purpose-aware processing** (consumers declare their purpose), giving you the best of both worlds: clean separation of concerns and precise execution control.

The solution transforms annotation processing from an all-or-nothing operation into a precise, controllable, and efficient system that scales with your application's needs. 