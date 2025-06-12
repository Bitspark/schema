// Package api defines the core interfaces for the schema system.
// This file contains annotation-related interfaces.
package api

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
	Validate() AnnotationValidationResult

	// Serialization
	ToMap() map[string]any
}

// AnnotationRegistry manages annotation type definitions and provides
// factory methods for creating and validating annotation instances.
type AnnotationRegistry interface {
	// Type management - register annotation type definitions
	RegisterType(name string, schema core.Schema, opts ...AnnotationTypeOption) error
	GetType(name string) (AnnotationType, bool)
	ListTypes() []string
	HasType(name string) bool

	// Instance management - create and validate annotation instances
	Create(name string, value any) (Annotation, error)
	CreateWithMetadata(name string, value any, metadata AnnotationMetadata) (Annotation, error)
	Validate(annotation Annotation) AnnotationValidationResult

	// Bulk operations
	CreateMany(annotations map[string]any) ([]Annotation, error)
	ValidateMany(annotations []Annotation) AnnotationValidationResult

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
	ValidateValue(value any) AnnotationValidationResult
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

// AnnotationValidationResult represents the result of validating an annotation.
type AnnotationValidationResult struct {
	Valid       bool                             `json:"valid"`
	Errors      []AnnotationValidationError      `json:"errors,omitempty"`
	Warnings    []AnnotationValidationWarning    `json:"warnings,omitempty"`
	Suggestions []AnnotationValidationSuggestion `json:"suggestions,omitempty"`
	Metadata    map[string]any                   `json:"metadata,omitempty"`
}

// AnnotationValidationError represents a specific validation error with context.
type AnnotationValidationError struct {
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	Value    any    `json:"value,omitempty"`
	Expected string `json:"expected,omitempty"`
	Context  string `json:"context,omitempty"`
}

// AnnotationValidationWarning represents a non-fatal validation issue.
type AnnotationValidationWarning struct {
	Path       string `json:"path,omitempty"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion,omitempty"`
}

// AnnotationValidationSuggestion provides actionable suggestions for fixing validation issues.
type AnnotationValidationSuggestion struct {
	Message    string         `json:"message"`
	Action     string         `json:"action"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// AnnotationTypeOption provides configuration options when registering annotation types.
type AnnotationTypeOption func(*AnnotationTypeConfig)

// AnnotationTypeConfig holds configuration for annotation type registration.
type AnnotationTypeConfig struct {
	Metadata   AnnotationMetadata
	Validators []string
	StrictMode bool
	AppliesTo  []string
	Conflicts  []string
	Requires   []string
}

// AnnotationTypeOption factory functions
func WithAnnotationMetadata(metadata AnnotationMetadata) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Metadata = metadata
	}
}

func WithAnnotationValidators(validators ...string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Validators = validators
	}
}

func WithAnnotationDescription(description string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Metadata.Description = description
	}
}

func WithAnnotationCategory(category string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Metadata.Category = category
	}
}

func WithAnnotationTags(tags ...string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Metadata.Tags = tags
	}
}

func WithAnnotationAppliesTo(schemaTypes ...string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.AppliesTo = schemaTypes
	}
}

func WithAnnotationConflicts(annotations ...string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Conflicts = annotations
	}
}

func WithAnnotationRequires(annotations ...string) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Requires = annotations
	}
}

func WithAnnotationExamples(examples ...any) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Metadata.Examples = examples
	}
}

func WithAnnotationDefaultValue(defaultValue any) AnnotationTypeOption {
	return func(c *AnnotationTypeConfig) {
		c.Metadata.DefaultValue = defaultValue
	}
}

// Helper functions for creating validation results
func NewAnnotationValidationError(code, message string) AnnotationValidationError {
	return AnnotationValidationError{
		Code:    code,
		Message: message,
	}
}

func NewAnnotationValidationWarning(code, message string) AnnotationValidationWarning {
	return AnnotationValidationWarning{
		Code:    code,
		Message: message,
	}
}

func ValidAnnotationResult() AnnotationValidationResult {
	return AnnotationValidationResult{Valid: true}
}

func InvalidAnnotationResult(errors ...AnnotationValidationError) AnnotationValidationResult {
	return AnnotationValidationResult{Valid: false, Errors: errors}
}

func AnnotationResultWithWarnings(warnings ...AnnotationValidationWarning) AnnotationValidationResult {
	return AnnotationValidationResult{Valid: true, Warnings: warnings}
}
