package core

// Annotation represents a single annotation instance with name, value, and metadata.
// Annotations are used to attach typed metadata to schemas, fields, and other
// schema components for validation and documentation purposes.
type Annotation interface {
	// Basic properties
	Name() string
	Value() any

	// Schema and validation
	Schema() Schema
	Validators() []string

	// Metadata and documentation
	Metadata() AnnotationMetadata

	// Validation
	Validate() AnnotationValidationResult

	// Serialization
	ToMap() map[string]any
}

// AnnotationType represents a registered annotation type with its schema and metadata.
type AnnotationType interface {
	// Basic properties
	Name() string
	Schema() Schema
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
