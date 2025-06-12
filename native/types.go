// Package native provides Go struct to schema conversion with annotation support.
// It bridges the gap between Go types and the schema system, enabling
// automatic schema generation from struct definitions with rich metadata.
package native

import (
	"reflect"

	"defs.dev/schema/annotation"
	"defs.dev/schema/api/core"
	"defs.dev/schema/registry"
)

// TypeConverter converts Go types to schemas with annotation support.
type TypeConverter interface {
	// Basic conversion
	FromType(t reflect.Type) (core.Schema, error)
	FromValue(v any) (core.Schema, error)
	FromTypeName(typeName string) (core.Schema, error)

	// Annotation integration
	FromTypeWithAnnotations(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error)
	FromValueWithAnnotations(v any, annotations []annotation.Annotation) (core.Schema, error)

	// Bulk operations
	FromTypes(types map[string]reflect.Type) (map[string]core.Schema, error)
	FromValues(values map[string]any) (map[string]core.Schema, error)

	// Configuration
	SetAnnotationRegistry(registry annotation.AnnotationRegistry)
	SetValidatorRegistry(registry registry.ValidatorRegistry)
	SetStrictMode(strict bool)

	// Metadata
	GetSupportedTags() []string
	GetConfiguration() ConverterConfig
}

// ServiceDiscovery analyzes Go types for service/function definitions.
type ServiceDiscovery interface {
	// Service analysis
	DiscoverServices(pkg any) ([]ServiceDefinition, error)
	DiscoverServiceFromType(t reflect.Type) (*ServiceDefinition, error)
	DiscoverServiceFromInterface(iface any) (*ServiceDefinition, error)

	// Function analysis
	DiscoverFunctions(obj any) ([]FunctionDefinition, error)
	DiscoverFunctionFromMethod(method reflect.Method) (*FunctionDefinition, error)
	DiscoverFunctionFromFunc(fn any) (*FunctionDefinition, error)

	// Type analysis
	AnalyzeType(t reflect.Type) (*TypeAnalysis, error)
	AnalyzeValue(v any) (*TypeAnalysis, error)

	// Configuration
	SetTagParser(parser TagParser)
	SetAnnotationRegistry(registry annotation.AnnotationRegistry)
}

// TagParser parses struct tags into annotations.
type TagParser interface {
	// Basic parsing
	ParseTags(tags reflect.StructTag) ([]annotation.Annotation, error)
	ParseTag(key, value string) (annotation.Annotation, error)

	// Supported tags
	GetSupportedTags() []string
	HasTag(key string) bool

	// Configuration
	SetAnnotationRegistry(registry annotation.AnnotationRegistry)
	SetStrictMode(strict bool)
}

// ConverterConfig provides configuration for type conversion.
type ConverterConfig struct {
	StrictMode          bool     `json:"strict_mode"`
	SupportedTags       []string `json:"supported_tags"`
	DefaultAnnotations  bool     `json:"default_annotations"`
	ValidateAnnotations bool     `json:"validate_annotations"`
	RecursiveConversion bool     `json:"recursive_conversion"`
	MaxDepth            int      `json:"max_depth"`
	CacheResults        bool     `json:"cache_results"`
	IgnoreUnknownTags   bool     `json:"ignore_unknown_tags"`
}

// ServiceDefinition represents a discovered service.
type ServiceDefinition struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Type        reflect.Type            `json:"-"`
	Functions   []FunctionDefinition    `json:"functions"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Metadata    ServiceMetadata         `json:"metadata"`
	Package     string                  `json:"package,omitempty"`
	Module      string                  `json:"module,omitempty"`
}

// FunctionDefinition represents a discovered function/method.
type FunctionDefinition struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Type        reflect.Type            `json:"-"`
	Method      *reflect.Method         `json:"-"`
	Signature   FunctionSignature       `json:"signature"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Metadata    FunctionMetadata        `json:"metadata"`
	Service     string                  `json:"service,omitempty"`
}

// FunctionSignature represents function input/output schema.
type FunctionSignature struct {
	Parameters []ParameterDefinition `json:"parameters"`
	Returns    []ReturnDefinition    `json:"returns"`
	Errors     []ErrorDefinition     `json:"errors,omitempty"`
}

// ParameterDefinition represents a function parameter.
type ParameterDefinition struct {
	Name        string                  `json:"name"`
	Type        reflect.Type            `json:"-"`
	Schema      core.Schema             `json:"schema"`
	Required    bool                    `json:"required"`
	Default     any                     `json:"default,omitempty"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Description string                  `json:"description,omitempty"`
	Position    int                     `json:"position"`
}

// ReturnDefinition represents a function return value.
type ReturnDefinition struct {
	Name        string                  `json:"name,omitempty"`
	Type        reflect.Type            `json:"-"`
	Schema      core.Schema             `json:"schema"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Description string                  `json:"description,omitempty"`
	Position    int                     `json:"position"`
}

// ErrorDefinition represents a function error type.
type ErrorDefinition struct {
	Type        reflect.Type            `json:"-"`
	Schema      core.Schema             `json:"schema"`
	Code        string                  `json:"code,omitempty"`
	Message     string                  `json:"message,omitempty"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
}

// TypeAnalysis provides detailed analysis of a Go type.
type TypeAnalysis struct {
	Type         reflect.Type            `json:"-"`
	Kind         reflect.Kind            `json:"kind"`
	Name         string                  `json:"name"`
	Package      string                  `json:"package"`
	Fields       []FieldAnalysis         `json:"fields,omitempty"`
	Methods      []MethodAnalysis        `json:"methods,omitempty"`
	Implements   []string                `json:"implements,omitempty"`
	Annotations  []annotation.Annotation `json:"annotations,omitempty"`
	Dependencies []string                `json:"dependencies,omitempty"`
	Schema       core.Schema             `json:"schema,omitempty"`
	Metadata     TypeMetadata            `json:"metadata"`
}

// FieldAnalysis represents analysis of a struct field.
type FieldAnalysis struct {
	Name        string                  `json:"name"`
	Type        reflect.Type            `json:"-"`
	Tag         reflect.StructTag       `json:"tag,omitempty"`
	Schema      core.Schema             `json:"schema,omitempty"`
	Required    bool                    `json:"required"`
	Default     any                     `json:"default,omitempty"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Position    int                     `json:"position"`
	Exported    bool                    `json:"exported"`
	Embedded    bool                    `json:"embedded"`
}

// MethodAnalysis represents analysis of a type method.
type MethodAnalysis struct {
	Name        string                  `json:"name"`
	Type        reflect.Type            `json:"-"`
	Method      reflect.Method          `json:"-"`
	Signature   FunctionSignature       `json:"signature"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Exported    bool                    `json:"exported"`
	Receiver    string                  `json:"receiver"`
}

// ServiceMetadata provides metadata about a service.
type ServiceMetadata struct {
	Version       string            `json:"version,omitempty"`
	Author        string            `json:"author,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Category      string            `json:"category,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
	Deprecated    bool              `json:"deprecated,omitempty"`
	Stability     string            `json:"stability,omitempty"`
	Documentation string            `json:"documentation,omitempty"`
}

// FunctionMetadata provides metadata about a function.
type FunctionMetadata struct {
	HTTPMethod string            `json:"http_method,omitempty"`
	HTTPPath   string            `json:"http_path,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Category   string            `json:"category,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Deprecated bool              `json:"deprecated,omitempty"`
	Async      bool              `json:"async,omitempty"`
	Idempotent bool              `json:"idempotent,omitempty"`
	RateLimit  *RateLimit        `json:"rate_limit,omitempty"`
	Timeout    string            `json:"timeout,omitempty"`
	Retry      *RetryConfig      `json:"retry,omitempty"`
}

// TypeMetadata provides metadata about a type.
type TypeMetadata struct {
	Serializable bool              `json:"serializable"`
	Comparable   bool              `json:"comparable"`
	Immutable    bool              `json:"immutable"`
	ThreadSafe   bool              `json:"thread_safe"`
	Tags         []string          `json:"tags,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
	Size         int               `json:"size,omitempty"`
	Alignment    int               `json:"alignment,omitempty"`
}

// RateLimit represents rate limiting configuration.
type RateLimit struct {
	Rate     int    `json:"rate"`
	Period   string `json:"period"`
	Burst    int    `json:"burst,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

// RetryConfig represents retry configuration.
type RetryConfig struct {
	MaxAttempts int      `json:"max_attempts"`
	Backoff     string   `json:"backoff"`
	Conditions  []string `json:"conditions,omitempty"`
}

// ConversionResult represents the result of type conversion.
type ConversionResult struct {
	Schema      core.Schema             `json:"schema"`
	Annotations []annotation.Annotation `json:"annotations,omitempty"`
	Metadata    map[string]any          `json:"metadata,omitempty"`
	Errors      []ConversionError       `json:"errors,omitempty"`
	Warnings    []ConversionWarning     `json:"warnings,omitempty"`
}

// ConversionError represents an error during conversion.
type ConversionError struct {
	Type    string `json:"type"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
}

// ConversionWarning represents a warning during conversion.
type ConversionWarning struct {
	Type       string `json:"type"`
	Field      string `json:"field,omitempty"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Common conversion error codes.
const (
	ErrorCodeUnsupportedType   = "unsupported_type"
	ErrorCodeInvalidAnnotation = "invalid_annotation"
	ErrorCodeCircularReference = "circular_reference"
	ErrorCodeDepthExceeded     = "depth_exceeded"
	ErrorCodeTagParsingFailed  = "tag_parsing_failed"
	ErrorCodeSchemaGeneration  = "schema_generation"
	ErrorCodeValidationFailed  = "validation_failed"
)

// Helper functions for creating conversion results.
func NewConversionError(code, message string) ConversionError {
	return ConversionError{
		Code:    code,
		Message: message,
	}
}

func NewConversionWarning(code, message string) ConversionWarning {
	return ConversionWarning{
		Code:    code,
		Message: message,
	}
}

func SuccessResult(schema core.Schema) ConversionResult {
	return ConversionResult{
		Schema: schema,
	}
}

func ErrorResult(errors ...ConversionError) ConversionResult {
	return ConversionResult{
		Errors: errors,
	}
}
