package registry

import (
	"defs.dev/schema/core/annotation"
	"fmt"
	"sync"
)

// DefaultValidatorRegistry is a thread-safe implementation of ValidatorRegistry.
type DefaultValidatorRegistry struct {
	validators    map[string]Validator
	factories     map[string]ValidatorFactory
	annotationReg annotation.AnnotationRegistry
	strictMode    bool
	mu            sync.RWMutex
}

// NewDefaultValidatorRegistry creates a new validator registry with built-in validators.
func NewDefaultValidatorRegistry(annotationReg annotation.AnnotationRegistry) *DefaultValidatorRegistry {
	r := &DefaultValidatorRegistry{
		validators:    make(map[string]Validator),
		factories:     make(map[string]ValidatorFactory),
		annotationReg: annotationReg,
		strictMode:    false,
	}

	// Register built-in validators
	r.registerBuiltinValidators()

	return r
}

// Register implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) Register(name string, validator Validator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.validators[name]; exists {
		return fmt.Errorf("validator %s already registered", name)
	}

	r.validators[name] = validator
	return nil
}

// Get implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) Get(name string) (Validator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validator, exists := r.validators[name]
	return validator, exists
}

// List implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.validators))
	for name := range r.validators {
		names = append(names, name)
	}
	return names
}

// HasValidator implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) HasValidator(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.validators[name]
	return exists
}

// ValidateWithAnnotations implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allErrors []ValidationError
	var allWarnings []ValidationWarning
	var appliedValidators []string

	for _, ann := range annotations {
		// Find validators that support this annotation type
		for _, validator := range r.validators {
			supportedTypes := validator.SupportedAnnotations()
			if containsString(supportedTypes, ann.Name()) {
				result := validator.ValidateWithAnnotations(value, []annotation.Annotation{ann})
				allErrors = append(allErrors, result.Errors...)
				allWarnings = append(allWarnings, result.Warnings...)
				appliedValidators = append(appliedValidators, validator.Name())
			}
		}
	}

	return ValidationResult{
		Valid:             len(allErrors) == 0,
		Errors:            allErrors,
		Warnings:          allWarnings,
		AppliedValidators: appliedValidators,
	}
}

// GetValidatorsForAnnotations implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) GetValidatorsForAnnotations(annotations []annotation.Annotation) []Validator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Validator
	seenValidators := make(map[string]bool)

	for _, ann := range annotations {
		for _, validator := range r.validators {
			if seenValidators[validator.Name()] {
				continue
			}

			supportedTypes := validator.SupportedAnnotations()
			if containsString(supportedTypes, ann.Name()) {
				result = append(result, validator)
				seenValidators[validator.Name()] = true
			}
		}
	}

	return result
}

// ValidateMany implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) ValidateMany(values map[string]any, annotationSets map[string][]annotation.Annotation) ValidationResult {
	var allErrors []ValidationError
	var allWarnings []ValidationWarning
	var appliedValidators []string

	for key, value := range values {
		annotations, exists := annotationSets[key]
		if !exists {
			continue
		}

		result := r.ValidateWithAnnotations(value, annotations)

		// Add path context to errors
		for _, err := range result.Errors {
			err.Path = key
			allErrors = append(allErrors, err)
		}

		// Add path context to warnings
		for _, warn := range result.Warnings {
			warn.Path = key
			allWarnings = append(allWarnings, warn)
		}

		appliedValidators = append(appliedValidators, result.AppliedValidators...)
	}

	return ValidationResult{
		Valid:             len(allErrors) == 0,
		Errors:            allErrors,
		Warnings:          allWarnings,
		AppliedValidators: appliedValidators,
	}
}

// AnnotationRegistry implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) AnnotationRegistry() annotation.AnnotationRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.annotationReg
}

// SetAnnotationRegistry implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) SetAnnotationRegistry(registry annotation.AnnotationRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.annotationReg = registry
}

// SetStrictMode implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) SetStrictMode(strict bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.strictMode = strict
}

// IsStrictMode implements ValidatorRegistry.
func (r *DefaultValidatorRegistry) IsStrictMode() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.strictMode
}

// RegisterFactory registers a validator factory.
func (r *DefaultValidatorRegistry) RegisterFactory(name string, factory ValidatorFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("factory %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// CreateValidator creates a new validator using a registered factory.
func (r *DefaultValidatorRegistry) CreateValidator(factoryName string, config any) (Validator, error) {
	r.mu.RLock()
	factory, exists := r.factories[factoryName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("factory %s not found", factoryName)
	}

	return factory.CreateValidator(config)
}

// registerBuiltinValidators registers all built-in validators.
func (r *DefaultValidatorRegistry) registerBuiltinValidators() {
	// Format validators
	r.Register("email", NewEmailValidator())
	r.Register("url", NewURLValidator())
	r.Register("uuid", NewUUIDValidator())
	r.Register("date", NewDateValidator())
	r.Register("time", NewTimeValidator())
	r.Register("datetime", NewDateTimeValidator())

	// Pattern validators
	r.Register("pattern", NewPatternValidator())

	// Length validators
	r.Register("minLength", NewMinLengthValidator())
	r.Register("maxLength", NewMaxLengthValidator())

	// Numeric validators
	r.Register("min", NewMinValidator())
	r.Register("max", NewMaxValidator())
	r.Register("range", NewRangeValidator())

	// Array validators
	r.Register("minItems", NewMinItemsValidator())
	r.Register("maxItems", NewMaxItemsValidator())
	r.Register("uniqueItems", NewUniqueItemsValidator())

	// Required validator
	r.Register("required", NewRequiredValidator())
}

// Helper function to check if a slice contains a string.
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
