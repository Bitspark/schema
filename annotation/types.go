// Package annotation provides implementations for the annotation system interfaces
// defined in the api package.
package annotation

import (
	"defs.dev/schema/api"
)

// Re-export the interfaces from api for convenience
type Annotation = api.Annotation
type AnnotationRegistry = api.AnnotationRegistry
type AnnotationType = api.AnnotationType
type AnnotationMetadata = api.AnnotationMetadata
type AnnotationValidationResult = api.AnnotationValidationResult
type AnnotationValidationError = api.AnnotationValidationError
type AnnotationValidationWarning = api.AnnotationValidationWarning
type AnnotationValidationSuggestion = api.AnnotationValidationSuggestion
type AnnotationTypeOption = api.AnnotationTypeOption
type AnnotationTypeConfig = api.AnnotationTypeConfig

// Re-export the option functions
var WithMetadata = api.WithAnnotationMetadata
var WithValidators = api.WithAnnotationValidators
var WithDescription = api.WithAnnotationDescription
var WithCategory = api.WithAnnotationCategory
var WithTags = api.WithAnnotationTags
var WithAppliesTo = api.WithAnnotationAppliesTo
var WithConflicts = api.WithAnnotationConflicts
var WithRequires = api.WithAnnotationRequires
var WithExamples = api.WithAnnotationExamples
var WithDefaultValue = api.WithAnnotationDefaultValue

// Re-export helper functions
var NewValidationError = api.NewAnnotationValidationError
var NewValidationWarning = api.NewAnnotationValidationWarning
var ValidResult = api.ValidAnnotationResult
var InvalidResult = api.InvalidAnnotationResult
var ResultWithWarnings = api.AnnotationResultWithWarnings
