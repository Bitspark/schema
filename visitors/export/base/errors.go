package base

import (
	"fmt"
	"strings"

	"defs.dev/schema/core"
)

// GenerationError represents an error that occurred during schema generation.
type GenerationError struct {
	// SchemaType is the type of schema that caused the error
	SchemaType string

	// GeneratorName is the name of the generator that failed
	GeneratorName string

	// Message is the error message
	Message string

	// Cause is the underlying error that caused this generation error
	Cause error

	// Context provides additional context about where the error occurred
	Context map[string]any

	// Path is the schema path where the error occurred (for nested schemas)
	Path []string
}

// Error implements the error interface.
func (e *GenerationError) Error() string {
	var parts []string

	if e.GeneratorName != "" {
		parts = append(parts, fmt.Sprintf("generator=%s", e.GeneratorName))
	}

	if e.SchemaType != "" {
		parts = append(parts, fmt.Sprintf("schema=%s", e.SchemaType))
	}

	if len(e.Path) > 0 {
		parts = append(parts, fmt.Sprintf("path=%s", strings.Join(e.Path, ".")))
	}

	contextStr := ""
	if len(parts) > 0 {
		contextStr = fmt.Sprintf("[%s] ", strings.Join(parts, ", "))
	}

	message := e.Message
	if e.Cause != nil {
		message = fmt.Sprintf("%s: %v", message, e.Cause)
	}

	return fmt.Sprintf("%s%s", contextStr, message)
}

// Unwrap returns the underlying cause error, supporting Go 1.13+ error unwrapping.
func (e *GenerationError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error.
func (e *GenerationError) WithContext(key string, value any) *GenerationError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithPath sets the schema path for the error.
func (e *GenerationError) WithPath(path ...string) *GenerationError {
	e.Path = path
	return e
}

// AppendPath adds elements to the schema path.
func (e *GenerationError) AppendPath(elements ...string) *GenerationError {
	e.Path = append(e.Path, elements...)
	return e
}

// NewGenerationError creates a new GenerationError.
func NewGenerationError(generatorName, schemaType, message string) *GenerationError {
	return &GenerationError{
		GeneratorName: generatorName,
		SchemaType:    schemaType,
		Message:       message,
		Context:       make(map[string]any),
	}
}

// NewGenerationErrorWithCause creates a new GenerationError with an underlying cause.
func NewGenerationErrorWithCause(generatorName, schemaType, message string, cause error) *GenerationError {
	return &GenerationError{
		GeneratorName: generatorName,
		SchemaType:    schemaType,
		Message:       message,
		Cause:         cause,
		Context:       make(map[string]any),
	}
}

// WrapGenerationError wraps an existing error as a GenerationError.
func WrapGenerationError(generatorName, schemaType string, cause error) *GenerationError {
	return &GenerationError{
		GeneratorName: generatorName,
		SchemaType:    schemaType,
		Message:       "generation failed",
		Cause:         cause,
		Context:       make(map[string]any),
	}
}

// ValidationError represents an error during output validation.
type ValidationError struct {
	// Format is the output format that failed validation
	Format string

	// Message is the validation error message
	Message string

	// Output is the output that failed validation (may be truncated for large outputs)
	Output string

	// Cause is the underlying error
	Cause error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	message := fmt.Sprintf("validation failed for format %s: %s", e.Format, e.Message)
	if e.Cause != nil {
		message = fmt.Sprintf("%s: %v", message, e.Cause)
	}
	return message
}

// Unwrap returns the underlying cause error.
func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new ValidationError.
func NewValidationError(format, message string, output []byte) *ValidationError {
	outputStr := string(output)
	if len(outputStr) > 500 {
		outputStr = outputStr[:500] + "... (truncated)"
	}

	return &ValidationError{
		Format:  format,
		Message: message,
		Output:  outputStr,
	}
}

// NewValidationErrorWithCause creates a new ValidationError with a cause.
func NewValidationErrorWithCause(format, message string, output []byte, cause error) *ValidationError {
	err := NewValidationError(format, message, output)
	err.Cause = cause
	return err
}

// TransformationError represents an error during output transformation.
type TransformationError struct {
	// TransformerName is the name of the transformer that failed
	TransformerName string

	// Format is the format being transformed
	Format string

	// Message is the error message
	Message string

	// Cause is the underlying error
	Cause error
}

// Error implements the error interface.
func (e *TransformationError) Error() string {
	message := fmt.Sprintf("transformation failed [%s, format=%s]: %s", e.TransformerName, e.Format, e.Message)
	if e.Cause != nil {
		message = fmt.Sprintf("%s: %v", message, e.Cause)
	}
	return message
}

// Unwrap returns the underlying cause error.
func (e *TransformationError) Unwrap() error {
	return e.Cause
}

// NewTransformationError creates a new TransformationError.
func NewTransformationError(transformerName, format, message string) *TransformationError {
	return &TransformationError{
		TransformerName: transformerName,
		Format:          format,
		Message:         message,
	}
}

// NewTransformationErrorWithCause creates a new TransformationError with a cause.
func NewTransformationErrorWithCause(transformerName, format, message string, cause error) *TransformationError {
	return &TransformationError{
		TransformerName: transformerName,
		Format:          format,
		Message:         message,
		Cause:           cause,
	}
}

// UnsupportedSchemaError represents an error when a generator encounters an unsupported schema type.
type UnsupportedSchemaError struct {
	// GeneratorName is the name of the generator
	GeneratorName string

	// SchemaType is the unsupported schema type
	SchemaType core.SchemaType

	// Message provides additional context
	Message string
}

// Error implements the error interface.
func (e *UnsupportedSchemaError) Error() string {
	message := fmt.Sprintf("generator %s does not support schema type %s", e.GeneratorName, e.SchemaType)
	if e.Message != "" {
		message = fmt.Sprintf("%s: %s", message, e.Message)
	}
	return message
}

// NewUnsupportedSchemaError creates a new UnsupportedSchemaError.
func NewUnsupportedSchemaError(generatorName string, schemaType core.SchemaType, message string) *UnsupportedSchemaError {
	return &UnsupportedSchemaError{
		GeneratorName: generatorName,
		SchemaType:    schemaType,
		Message:       message,
	}
}

// ErrorCollector collects multiple errors during generation.
type ErrorCollector struct {
	errors []error
}

// NewErrorCollector creates a new ErrorCollector.
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Add adds an error to the collection.
func (c *ErrorCollector) Add(err error) {
	if err != nil {
		c.errors = append(c.errors, err)
	}
}

// Addf adds a formatted error message to the collection.
func (c *ErrorCollector) Addf(format string, args ...any) {
	c.errors = append(c.errors, fmt.Errorf(format, args...))
}

// HasErrors returns true if the collector has any errors.
func (c *ErrorCollector) HasErrors() bool {
	return len(c.errors) > 0
}

// Errors returns all collected errors.
func (c *ErrorCollector) Errors() []error {
	return c.errors
}

// Error returns a combined error message, or nil if no errors.
func (c *ErrorCollector) Error() error {
	if len(c.errors) == 0 {
		return nil
	}
	if len(c.errors) == 1 {
		return c.errors[0]
	}

	var messages []string
	for _, err := range c.errors {
		messages = append(messages, err.Error())
	}

	return fmt.Errorf("multiple errors occurred: %s", strings.Join(messages, "; "))
}
