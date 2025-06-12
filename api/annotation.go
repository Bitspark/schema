// Package api defines the core interfaces for the schema system.
// This file contains annotation-related interfaces.
package api

import (
	"defs.dev/schema/api/core"
)

// AnnotationRegistry manages annotation type definitions and provides
// factory methods for creating and validating annotation instances.
type AnnotationRegistry interface {
	// Type management - register annotation type definitions
	RegisterType(name string, schema core.Schema, opts ...AnnotationTypeOption) error
	GetType(name string) (core.AnnotationType, bool)
	ListTypes() []string
	HasType(name string) bool

	// Instance management - create and validate annotation instances
	Create(name string, value any) (core.Annotation, error)
	CreateWithMetadata(name string, value any, metadata core.AnnotationMetadata) (core.Annotation, error)
	Validate(annotation core.Annotation) core.AnnotationValidationResult

	// Bulk operations
	CreateMany(annotations map[string]any) ([]core.Annotation, error)
	ValidateMany(annotations []core.Annotation) core.AnnotationValidationResult

	// Configuration
	SetStrictMode(strict bool)
	IsStrictMode() bool
}

// AnnotationTypeOption provides configuration options when registering annotation types.
type AnnotationTypeOption func(*AnnotationTypeConfig)

// AnnotationTypeConfig holds configuration for annotation type registration.
type AnnotationTypeConfig struct {
	Metadata   core.AnnotationMetadata
	Validators []string
	StrictMode bool
	AppliesTo  []string
	Conflicts  []string
	Requires   []string
}

// AnnotationTypeOption factory functions
func WithAnnotationMetadata(metadata core.AnnotationMetadata) AnnotationTypeOption {
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
func NewAnnotationValidationError(code, message string) core.AnnotationValidationError {
	return core.AnnotationValidationError{
		Code:    code,
		Message: message,
	}
}

func NewAnnotationValidationWarning(code, message string) core.AnnotationValidationWarning {
	return core.AnnotationValidationWarning{
		Code:    code,
		Message: message,
	}
}

func ValidAnnotationResult() core.AnnotationValidationResult {
	return core.AnnotationValidationResult{Valid: true}
}

func InvalidAnnotationResult(errors ...core.AnnotationValidationError) core.AnnotationValidationResult {
	return core.AnnotationValidationResult{Valid: false, Errors: errors}
}

func AnnotationResultWithWarnings(warnings ...core.AnnotationValidationWarning) core.AnnotationValidationResult {
	return core.AnnotationValidationResult{Valid: true, Warnings: warnings}
}
