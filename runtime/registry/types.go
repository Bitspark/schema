// Package registry provides validator registration and management with
// deep integration with the annotation system. Validators can be configured
// using type-safe annotations and applied to schema validation.
package registry

import (
	"defs.dev/schema/core"
	"defs.dev/schema/core/annotation"
)

// Validator validates values against specific criteria with annotation support.
type Validator interface {
	// Basic properties
	Name() string
	Validate(value any) ValidationResult
	Metadata() ValidatorMetadata

	// Annotation integration
	SupportedAnnotations() []string
	ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult
	ConfigureFromAnnotations(annotations []annotation.Annotation) error
}

// ValidatorRegistry manages validator registration, lookup, and annotation integration.
type ValidatorRegistry interface {
	// Validator management
	Register(name string, validator Validator) error
	Get(name string) (Validator, bool)
	List() []string
	HasValidator(name string) bool

	// Annotation-based validation
	ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult
	GetValidatorsForAnnotations(annotations []annotation.Annotation) []Validator

	// Bulk operations
	ValidateMany(values map[string]any, annotationSets map[string][]annotation.Annotation) ValidationResult

	// Integration with annotation registry
	AnnotationRegistry() annotation.AnnotationRegistry
	SetAnnotationRegistry(registry annotation.AnnotationRegistry)

	// Configuration
	SetStrictMode(strict bool)
	IsStrictMode() bool
}

// ValidatorMetadata provides information about a validator.
type ValidatorMetadata struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Category            string   `json:"category,omitempty"`
	Tags                []string `json:"tags,omitempty"`
	SupportedTypes      []string `json:"supported_types,omitempty"`
	RequiredAnnotations []string `json:"required_annotations,omitempty"`
	OptionalAnnotations []string `json:"optional_annotations,omitempty"`
	Examples            []any    `json:"examples,omitempty"`
	Version             string   `json:"version,omitempty"`
}

// ValidationResult represents the result of validator-based validation.
type ValidationResult struct {
	Valid             bool                   `json:"valid"`
	Errors            []ValidationError      `json:"errors,omitempty"`
	Warnings          []ValidationWarning    `json:"warnings,omitempty"`
	Suggestions       []ValidationSuggestion `json:"suggestions,omitempty"`
	Metadata          map[string]any         `json:"metadata,omitempty"`
	AppliedValidators []string               `json:"applied_validators,omitempty"`
}

// ValidationError represents a specific validation error.
type ValidationError struct {
	ValidatorName string `json:"validator_name"`
	Path          string `json:"path,omitempty"`
	Message       string `json:"message"`
	Code          string `json:"code"`
	Value         any    `json:"value,omitempty"`
	Expected      string `json:"expected,omitempty"`
	Context       string `json:"context,omitempty"`
	Suggestion    string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a non-fatal validation issue.
type ValidationWarning struct {
	ValidatorName string `json:"validator_name"`
	Path          string `json:"path,omitempty"`
	Message       string `json:"message"`
	Code          string `json:"code"`
	Suggestion    string `json:"suggestion,omitempty"`
}

// ValidationSuggestion provides actionable suggestions for fixing validation issues.
type ValidationSuggestion struct {
	ValidatorName string         `json:"validator_name"`
	Message       string         `json:"message"`
	Action        string         `json:"action"`
	Parameters    map[string]any `json:"parameters,omitempty"`
}

// ValidatorFactory creates validators from configuration.
type ValidatorFactory interface {
	Name() string
	CreateValidator(config any) (Validator, error)
	GetConfigSchema() core.Schema
	GetMetadata() ValidatorMetadata
}

// Common validation error codes
const (
	ErrorCodeValidationFailed = "validation_failed"
	ErrorCodeInvalidFormat    = "invalid_format"
	ErrorCodeOutOfRange       = "out_of_range"
	ErrorCodeLengthConstraint = "length_constraint"
	ErrorCodePatternMismatch  = "pattern_mismatch"
	ErrorCodeRequiredMissing  = "required_missing"
	ErrorCodeTypeError        = "type_error"
	ErrorCodeCustomValidation = "custom_validation"
)

// Helper functions for creating validation results
func NewValidationError(validatorName, code, message string) ValidationError {
	return ValidationError{
		ValidatorName: validatorName,
		Code:          code,
		Message:       message,
	}
}

func NewValidationWarning(validatorName, code, message string) ValidationWarning {
	return ValidationWarning{
		ValidatorName: validatorName,
		Code:          code,
		Message:       message,
	}
}

func ValidResult() ValidationResult {
	return ValidationResult{Valid: true}
}

func InvalidResult(errors ...ValidationError) ValidationResult {
	return ValidationResult{
		Valid:  false,
		Errors: errors,
	}
}

func ResultWithWarnings(warnings ...ValidationWarning) ValidationResult {
	return ValidationResult{
		Valid:    true,
		Warnings: warnings,
	}
}
