# Validator Decoupling Proposal

## Executive Summary

This proposal outlines the complete removal of validation logic from the core schema system and its migration to a registry-based architecture with a "validate_value" purpose. This change will improve modularity, extensibility, and align with the new `Value[T]` interface system introduced in `values.go`.

## Current State Analysis

### Integration Points
- **Core Interface**: `Schema.Validate(value any) ValidationResult` (marked deprecated)
- **Schema Implementation**: Only `StringSchema` currently uses `ValidatorRegistry`
- **Legacy Systems**: Most schemas still use direct validation logic
- **Dual Architecture**: Coexistence of legacy and annotation-based validation

### Problems with Current Approach
1. **Tight Coupling**: Validation logic embedded in schema implementations
2. **Limited Extensibility**: Adding new validators requires schema modifications
3. **Inconsistent Architecture**: Mixed legacy and registry-based approaches
4. **Import Cycles**: Schemas need to duplicate `ValidatorRegistry` interface
5. **Value System Misalignment**: New `Value[T]` interfaces not integrated with validation

## Proposed Architecture

### Core Principle: Complete Decoupling
- **Schemas** define structure and metadata only
- **Validators** are registered with purpose "validate_value" 
- **Value Types** represent validated data structures
- **Registry** manages all validation logic centrally

### Key Components

#### 1. Enhanced Value Validator Registry

```go
// registry/value_validator.go
package registry

type ValueValidator interface {
    Name() string
    Purpose() string // Always "validate_value"
    
    // Core validation
    ValidateValue(value any) ValueValidationResult
    ValidateTypedValue(value core.Value[any]) ValueValidationResult
    
    // Annotation support
    SupportedAnnotations() []string
    ValidateWithAnnotations(value any, annotations []core.Annotation) ValueValidationResult
    
    // Metadata
    Metadata() ValidatorMetadata
    SupportedTypes() []core.SchemaType
}

type ValueValidationResult struct {
    Valid         bool                    `json:"valid"`
    ValidatedValue core.Value[any]       `json:"validated_value,omitempty"`
    Errors        []ValidationError       `json:"errors,omitempty"`
    Warnings      []ValidationWarning     `json:"warnings,omitempty"`
    Suggestions   []ValidationSuggestion  `json:"suggestions,omitempty"`
    Metadata      map[string]any          `json:"metadata,omitempty"`
    AppliedValidators []string            `json:"applied_validators,omitempty"`
}
```

#### 2. Updated Core Schema Interface

```go
// api/core/types.go
type Schema interface {
    // Structure and metadata only - NO validation
    Type() SchemaType
    Annotations() []Annotation
    Metadata() SchemaMetadata
    Clone() Schema
    
    // Value creation (replaces validation)
    CreateValue(raw any) (Value[any], error)
    AcceptedTypes() []reflect.Type
}

// Validation removed from core - handled by registry
// type Schema interface { Validate(value any) ValidationResult } // REMOVED
```

#### 3. Value-First Validation Flow

```go
// Example: String validation
type StringValueValidator struct {
    registry.BaseValidator
}

func (v *StringValueValidator) ValidateValue(raw any) ValueValidationResult {
    // 1. Type checking
    str, ok := raw.(string)
    if !ok {
        return registry.InvalidValueResult(registry.NewValidationError(
            v.Name(), "type_error", "Expected string",
        ))
    }
    
    // 2. Create typed value
    stringValue := core.NewStringValue(str)
    
    // 3. Return validated value
    return registry.ValidValueResult(stringValue)
}

func (v *StringValueValidator) ValidateWithAnnotations(raw any, annotations []core.Annotation) ValueValidationResult {
    baseResult := v.ValidateValue(raw)
    if !baseResult.Valid {
        return baseResult
    }
    
    // Apply annotation-based constraints
    for _, ann := range annotations {
        if constraint := v.getConstraintForAnnotation(ann); constraint != nil {
            if err := constraint.Validate(baseResult.ValidatedValue); err != nil {
                return registry.InvalidValueResult(err)
            }
        }
    }
    
    return baseResult
}
```

## Migration Strategy

### Phase 1: Infrastructure Setup (Week 1-2)

**1.1 Create Value Validator Registry**
- [ ] Implement `ValueValidator` interface
- [ ] Create `ValueValidationResult` with `ValidatedValue` field
- [ ] Update registry to support "validate_value" purpose

**1.2 Extend Value System**
- [ ] Add constructor functions for each `Value[T]` type
- [ ] Implement value factories in schemas
- [ ] Create value constraint system

**1.3 Create Migration Utilities**
- [ ] Legacy validator wrapper for backward compatibility
- [ ] Validation result converter (legacy ↔ value-based)
- [ ] Schema migration helpers

### Phase 2: Core Schema Updates (Week 3-4)

**2.1 Remove Validate() Method**
```go
// Before:
type Schema interface {
    Validate(value any) ValidationResult  // REMOVE THIS
    Type() SchemaType
    // ...
}

// After:
type Schema interface {
    Type() SchemaType
    CreateValue(raw any) (Value[any], error)  // ADD THIS
    // ...
}
```

**2.2 Update Schema Implementations**
- [ ] Remove all `Validate()` implementations
- [ ] Add `CreateValue()` implementations
- [ ] Remove embedded validation logic
- [ ] Update to use registry for validation

**2.3 Create Value Validators**
```go
// Register all current validation logic as value validators
registry.Register("string_validator", &StringValueValidator{
    Purpose: "validate_value",
    SupportedTypes: []core.SchemaType{core.TypeString},
})

registry.Register("number_validator", &NumberValueValidator{
    Purpose: "validate_value", 
    SupportedTypes: []core.SchemaType{core.TypeNumber, core.TypeInteger},
})
```

### Phase 3: Consumer Updates (Week 5-6)

**3.1 Update All Validation Calls**
```go
// Before:
result := schema.Validate(data)
if !result.Valid {
    return result.Errors
}

// After:
validatedValue, err := valueRegistry.ValidateWithSchema(data, schema)
if err != nil {
    return err
}
// Use validatedValue.Value() for the actual data
```

**3.2 Update Tests**
- [ ] Convert all schema validation tests
- [ ] Add value validator tests
- [ ] Update integration tests

### Phase 4: Legacy Cleanup (Week 7)

**4.1 Remove Legacy Code**
- [ ] Remove all legacy validation methods
- [ ] Remove duplicate type definitions
- [ ] Clean up import cycles
- [ ] Remove deprecated interfaces

**4.2 Documentation**
- [ ] Update API documentation
- [ ] Create migration guide
- [ ] Update examples

## Impact Assessment

### Breaking Changes
1. **Schema.Validate() Method Removal**: All direct schema validation calls
2. **ValidationResult Structure**: New result includes validated values
3. **Import Changes**: Validation logic moves to registry package
4. **Constructor Patterns**: Schemas now create values instead of validating

### Benefits
1. **Complete Decoupling**: Validation is entirely separate from schema definition
2. **Enhanced Extensibility**: New validators can be added without schema changes
3. **Value-First Design**: Aligns with new `Value[T]` interface system
4. **Import Cycle Resolution**: No more circular dependencies
5. **Consistent Architecture**: Single validation approach across all types
6. **Performance**: Validated values can be cached and reused

### Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Large codebase changes | High | Phased migration with compatibility layer |
| Performance regression | Medium | Benchmark and optimize value creation |
| Developer confusion | Medium | Comprehensive documentation and examples |
| Test breakage | High | Automated test migration tools |

## Implementation Details

### New Registry Methods

```go
type ValueValidatorRegistry interface {
    // Value-specific validation
    ValidateWithSchema(raw any, schema core.Schema) (core.Value[any], error)
    ValidateMany(values map[string]any, schemas map[string]core.Schema) (map[string]core.Value[any], error)
    
    // Purpose-based registration
    RegisterValueValidator(name string, validator ValueValidator) error
    GetValueValidators() []ValueValidator
    GetValidatorsByPurpose(purpose string) []ValueValidator
}
```

### Value Constraint System

```go
type ValueConstraint interface {
    Name() string
    Validate(value core.Value[any]) error
    AppliesTo(schemaType core.SchemaType) bool
}

// Example: Length constraint for strings and arrays
type LengthConstraint struct {
    Min, Max int
}

func (c *LengthConstraint) Validate(value core.Value[any]) error {
    switch v := value.(type) {
    case core.StringValue:
        if len(v.Value()) < c.Min || len(v.Value()) > c.Max {
            return fmt.Errorf("length constraint violated")
        }
    case core.ArrayValue[any]:
        if v.Length() < c.Min || v.Length() > c.Max {
            return fmt.Errorf("length constraint violated")
        }
    }
    return nil
}
```

### Backward Compatibility Layer

```go
// Temporary compatibility wrapper
func (s *StringSchema) Validate(value any) core.ValidationResult {
    // DEPRECATED: Use ValueValidatorRegistry instead
    validatedValue, err := defaultValueRegistry.ValidateWithSchema(value, s)
    if err != nil {
        return core.ValidationResult{
            Valid: false,
            Errors: convertErrorsFromValueValidation(err),
        }
    }
    
    return core.ValidationResult{
        Valid: true,
        Metadata: map[string]any{
            "validated_value": validatedValue,
        },
    }
}
```

## Success Criteria

1. **Zero Import Cycles**: No circular dependencies between packages
2. **Complete Decoupling**: Schemas contain no validation logic
3. **Value Integration**: All validation produces typed `Value[T]` instances
4. **Extensibility**: New validators can be added without core changes
5. **Performance**: No significant performance regression
6. **Compatibility**: Smooth migration path for existing code

## Timeline

| Phase | Duration | Deliverables |
|-------|----------|-------------|
| Phase 1 | 2 weeks | Value validator registry, migration utilities |
| Phase 2 | 2 weeks | Updated core schemas, value validators |
| Phase 3 | 2 weeks | Consumer updates, test migration |
| Phase 4 | 1 week | Legacy cleanup, documentation |
| **Total** | **7 weeks** | Fully decoupled validation system |

## Conclusion

This proposal provides a comprehensive path to completely decouple validation from the core schema system while integrating with the new `Value[T]` interfaces. The registry-based approach with "validate_value" purpose creates a clean, extensible architecture that resolves current architectural issues and positions the system for future growth.

The phased migration strategy minimizes risk while ensuring a smooth transition for all consumers of the schema system. 