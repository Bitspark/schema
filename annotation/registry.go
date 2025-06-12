package annotation

import (
	"fmt"
	"sync"

	"defs.dev/schema/api/core"
)

// registryImpl implements AnnotationRegistry with thread-safe type management
// and instance creation capabilities.
type registryImpl struct {
	types      map[string]AnnotationType
	strictMode bool
	mu         sync.RWMutex
}

// NewRegistry creates a new annotation registry with default configuration.
func NewRegistry() AnnotationRegistry {
	return &registryImpl{
		types:      make(map[string]AnnotationType),
		strictMode: false,
	}
}

// Type management implementation

func (r *registryImpl) RegisterType(name string, schema core.Schema, opts ...AnnotationTypeOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if type already exists
	if _, exists := r.types[name]; exists {
		return fmt.Errorf("annotation type '%s' already registered", name)
	}

	// Apply options to build configuration
	config := &AnnotationTypeConfig{
		Metadata: AnnotationMetadata{
			Name: name,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	// Create and register the annotation type
	annotationType := &annotationTypeImpl{
		name:     name,
		schema:   schema,
		metadata: config.Metadata,
		registry: r,
	}

	r.types[name] = annotationType
	return nil
}

func (r *registryImpl) GetType(name string) (AnnotationType, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	annotationType, exists := r.types[name]
	return annotationType, exists
}

func (r *registryImpl) ListTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.types))
	for name := range r.types {
		types = append(types, name)
	}
	return types
}

func (r *registryImpl) HasType(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.types[name]
	return exists
}

// Instance management implementation

func (r *registryImpl) Create(name string, value any) (Annotation, error) {
	return r.CreateWithMetadata(name, value, AnnotationMetadata{})
}

func (r *registryImpl) CreateWithMetadata(name string, value any, metadata AnnotationMetadata) (Annotation, error) {
	annotationType, exists := r.GetType(name)
	if !exists {
		if r.strictMode {
			return nil, fmt.Errorf("unknown annotation type: %s", name)
		}
		// In non-strict mode, create a flexible annotation
		return r.createFlexibleAnnotation(name, value, metadata), nil
	}

	return annotationType.CreateWithMetadata(value, metadata)
}

func (r *registryImpl) Validate(annotation Annotation) AnnotationValidationResult {
	// Validate the annotation against its type schema
	return annotation.Validate()
}

func (r *registryImpl) CreateMany(annotations map[string]any) ([]Annotation, error) {
	result := make([]Annotation, 0, len(annotations))

	for name, value := range annotations {
		annotation, err := r.Create(name, value)
		if err != nil {
			return nil, fmt.Errorf("failed to create annotation '%s': %w", name, err)
		}
		result = append(result, annotation)
	}

	return result, nil
}

func (r *registryImpl) ValidateMany(annotations []Annotation) AnnotationValidationResult {
	var allErrors []AnnotationValidationError
	var allWarnings []AnnotationValidationWarning

	for _, annotation := range annotations {
		result := r.Validate(annotation)
		if !result.Valid {
			allErrors = append(allErrors, result.Errors...)
		}
		allWarnings = append(allWarnings, result.Warnings...)
	}

	return AnnotationValidationResult{
		Valid:    len(allErrors) == 0,
		Errors:   allErrors,
		Warnings: allWarnings,
		Metadata: map[string]any{
			"total_annotations": len(annotations),
			"validation_passed": len(allErrors) == 0,
		},
	}
}

// Configuration

func (r *registryImpl) SetStrictMode(strict bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strictMode = strict
}

func (r *registryImpl) IsStrictMode() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.strictMode
}

// Helper methods

func (r *registryImpl) createFlexibleAnnotation(name string, value any, metadata AnnotationMetadata) Annotation {
	// Create a flexible annotation for unknown types in non-strict mode
	return &flexibleAnnotationImpl{
		name:     name,
		value:    value,
		metadata: metadata,
	}
}

// annotationTypeImpl implements AnnotationType
type annotationTypeImpl struct {
	name     string
	schema   core.Schema
	metadata AnnotationMetadata
	registry *registryImpl
}

func (at *annotationTypeImpl) Name() string {
	return at.name
}

func (at *annotationTypeImpl) Schema() core.Schema {
	return at.schema
}

func (at *annotationTypeImpl) Metadata() AnnotationMetadata {
	return at.metadata
}

func (at *annotationTypeImpl) Create(value any) (Annotation, error) {
	return at.CreateWithMetadata(value, AnnotationMetadata{})
}

func (at *annotationTypeImpl) CreateWithMetadata(value any, metadata AnnotationMetadata) (Annotation, error) {
	// Validate the value against the schema
	validationResult := at.ValidateValue(value)
	if !validationResult.Valid {
		return nil, fmt.Errorf("invalid value for annotation '%s': %v", at.name, validationResult.Errors)
	}

	// Merge metadata with type metadata
	mergedMetadata := at.metadata
	if metadata.Description != "" {
		mergedMetadata.Description = metadata.Description
	}
	if len(metadata.Tags) > 0 {
		mergedMetadata.Tags = append(mergedMetadata.Tags, metadata.Tags...)
	}
	if len(metadata.Properties) > 0 {
		if mergedMetadata.Properties == nil {
			mergedMetadata.Properties = make(map[string]string)
		}
		for k, v := range metadata.Properties {
			mergedMetadata.Properties[k] = v
		}
	}

	return &typedAnnotationImpl{
		name:           at.name,
		value:          value,
		schema:         at.schema,
		metadata:       mergedMetadata,
		annotationType: at,
	}, nil
}

func (at *annotationTypeImpl) ValidateValue(value any) AnnotationValidationResult {
	coreResult := at.schema.Validate(value)

	// Convert core.ValidationResult to annotation.AnnotationValidationResult
	errors := make([]AnnotationValidationError, len(coreResult.Errors))
	for i, err := range coreResult.Errors {
		errors[i] = AnnotationValidationError{
			Path:     err.Path,
			Message:  err.Message,
			Code:     err.Code,
			Value:    err.Value,
			Expected: err.Expected,
			Context:  err.Context,
		}
	}

	return AnnotationValidationResult{
		Valid:    coreResult.Valid,
		Errors:   errors,
		Metadata: coreResult.Metadata,
	}
}

// typedAnnotationImpl implements Annotation for typed annotations
type typedAnnotationImpl struct {
	name           string
	value          any
	schema         core.Schema
	metadata       AnnotationMetadata
	annotationType *annotationTypeImpl
}

func (a *typedAnnotationImpl) Name() string {
	return a.name
}

func (a *typedAnnotationImpl) Value() any {
	return a.value
}

func (a *typedAnnotationImpl) Schema() core.Schema {
	return a.schema
}

func (a *typedAnnotationImpl) Validators() []string {
	return a.metadata.Validators
}

func (a *typedAnnotationImpl) Metadata() AnnotationMetadata {
	return a.metadata
}

func (a *typedAnnotationImpl) Validate() AnnotationValidationResult {
	return a.annotationType.ValidateValue(a.value)
}

func (a *typedAnnotationImpl) ToMap() map[string]any {
	return map[string]any{
		"name":       a.name,
		"value":      a.value,
		"metadata":   a.metadata,
		"validators": a.metadata.Validators,
	}
}

// flexibleAnnotationImpl implements Annotation for untyped annotations in non-strict mode
type flexibleAnnotationImpl struct {
	name     string
	value    any
	metadata AnnotationMetadata
}

func (a *flexibleAnnotationImpl) Name() string {
	return a.name
}

func (a *flexibleAnnotationImpl) Value() any {
	return a.value
}

func (a *flexibleAnnotationImpl) Schema() core.Schema {
	// Return a flexible schema that accepts any value
	return nil // TODO: Create a flexible "any" schema
}

func (a *flexibleAnnotationImpl) Validators() []string {
	return a.metadata.Validators
}

func (a *flexibleAnnotationImpl) Metadata() AnnotationMetadata {
	return a.metadata
}

func (a *flexibleAnnotationImpl) Validate() AnnotationValidationResult {
	// Flexible annotations are always valid
	return ValidResult()
}

func (a *flexibleAnnotationImpl) ToMap() map[string]any {
	return map[string]any{
		"name":       a.name,
		"value":      a.value,
		"metadata":   a.metadata,
		"flexible":   true,
		"validators": a.metadata.Validators,
	}
}
