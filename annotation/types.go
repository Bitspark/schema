// Package annotation provides a flexible annotation system for schema metadata
// and validation. It supports type-safe annotation definitions, validation,
// and integration with the broader schema system.
package annotation

import (
	"defs.dev/schema/api/core"
)

// Annotation represents a single annotation instance with name, value, and metadata.
// Annotations are used to attach typed metadata to schemas, fields, and other
// schema components for validation and documentation purposes.
type Annotation interface {
	// Basic properties
	Name() string
	Value() any

	// Schema and validation
	Schema() core.Schema
	Validators() []string

	// Metadata and documentation
	Metadata() AnnotationMetadata

	// Validation
	Validate() ValidationResult

	// Serialization
	ToMap() map[string]any
}

// AnnotationRegistry manages annotation type definitions and provides
// factory methods for creating and validating annotation instances.
type AnnotationRegistry interface {
	// Type management - register annotation type definitions
	RegisterType(name string, schema core.Schema, opts ...TypeOption) error
	GetType(name string) (AnnotationType, bool)
	ListTypes() []string
	HasType(name string) bool

	// Instance management - create and validate annotation instances
	Create(name string, value any) (Annotation, error)
	CreateWithMetadata(name string, value any, metadata AnnotationMetadata) (Annotation, error)
	Validate(annotation Annotation) ValidationResult

	// Bulk operations
	CreateMany(annotations map[string]any) ([]Annotation, error)
	ValidateMany(annotations []Annotation) ValidationResult

	// Configuration
	SetStrictMode(strict bool)
	IsStrictMode() bool
}

// AnnotationType represents a registered annotation type with its schema and metadata.
type AnnotationType interface {
	// Basic properties
	Name() string
	Schema() core.Schema
	Metadata() AnnotationMetadata

	// Factory methods
	Create(value any) (Annotation, error)
	CreateWithMetadata(value any, metadata AnnotationMetadata) (Annotation, error)

	// Validation
	ValidateValue(value any) ValidationResult
}

// AnnotationMetadata provides rich metadata about annotation types and instances.
type AnnotationMetadata struct {
	// Basic information
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Examples and defaults
	Examples     []any `json:"examples,omitempty"`
	DefaultValue any   `json:"default_value,omitempty"`

	// Validation configuration
	Required   bool     `json:"required,omitempty"`
	Validators []string `json:"validators,omitempty"`

	// Categorization
	Tags       []string          `json:"tags,omitempty"`
	Category   string            `json:"category,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`

	// Versioning and ownership
	Version string `json:"version,omitempty"`
	Author  string `json:"author,omitempty"`

	// Usage constraints
	AppliesTo []string `json:"applies_to,omitempty"` // Schema types this applies to
	Conflicts []string `json:"conflicts,omitempty"`  // Annotations this conflicts with
	Requires  []string `json:"requires,omitempty"`   // Required co-annotations
}

// ValidationResult represents the result of validating an annotation.
type ValidationResult struct {
	Valid       bool                   `json:"valid"`
	Errors      []ValidationError      `json:"errors,omitempty"`
	Warnings    []ValidationWarning    `json:"warnings,omitempty"`
	Suggestions []ValidationSuggestion `json:"suggestions,omitempty"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

// ValidationError represents a specific validation error with context.
type ValidationError struct {
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	Value    any    `json:"value,omitempty"`
	Expected string `json:"expected,omitempty"`
	Context  string `json:"context,omitempty"`
}

// ValidationWarning represents a non-fatal validation issue.
type ValidationWarning struct {
	Path       string `json:"path,omitempty"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationSuggestion provides actionable suggestions for fixing validation issues.
type ValidationSuggestion struct {
	Message    string         `json:"message"`
	Action     string         `json:"action"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// TypeOption provides configuration options when registering annotation types.
type TypeOption func(*typeConfig)

type typeConfig struct {
	metadata   AnnotationMetadata
	validators []string
	strictMode bool
	appliesTo  []string
	conflicts  []string
	requires   []string
}

// TypeOption factory functions
func WithMetadata(metadata AnnotationMetadata) TypeOption {
	return func(c *typeConfig) {
		c.metadata = metadata
	}
}

func WithValidators(validators ...string) TypeOption {
	return func(c *typeConfig) {
		c.validators = validators
	}
}

func WithDescription(description string) TypeOption {
	return func(c *typeConfig) {
		c.metadata.Description = description
	}
}

func WithCategory(category string) TypeOption {
	return func(c *typeConfig) {
		c.metadata.Category = category
	}
}

func WithTags(tags ...string) TypeOption {
	return func(c *typeConfig) {
		c.metadata.Tags = tags
	}
}

func WithAppliesTo(schemaTypes ...string) TypeOption {
	return func(c *typeConfig) {
		c.appliesTo = schemaTypes
	}
}

func WithConflicts(annotations ...string) TypeOption {
	return func(c *typeConfig) {
		c.conflicts = annotations
	}
}

func WithRequires(annotations ...string) TypeOption {
	return func(c *typeConfig) {
		c.requires = annotations
	}
}

func WithExamples(examples ...any) TypeOption {
	return func(c *typeConfig) {
		c.metadata.Examples = examples
	}
}

func WithDefaultValue(defaultValue any) TypeOption {
	return func(c *typeConfig) {
		c.metadata.DefaultValue = defaultValue
	}
}

// Common validation error codes
const (
	ErrorCodeInvalidType        = "invalid_type"
	ErrorCodeInvalidValue       = "invalid_value"
	ErrorCodeMissingRequired    = "missing_required"
	ErrorCodeConflictingAnnot   = "conflicting_annotation"
	ErrorCodeUnsupportedType    = "unsupported_type"
	ErrorCodeConstraintViolated = "constraint_violated"
	ErrorCodeUnknownAnnotation  = "unknown_annotation"
)

// Helper functions for creating common validation results
func NewValidationError(code, message string) ValidationError {
	return ValidationError{
		Code:    code,
		Message: message,
	}
}

func NewValidationWarning(code, message string) ValidationWarning {
	return ValidationWarning{
		Code:    code,
		Message: message,
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
