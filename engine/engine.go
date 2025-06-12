// Package engine provides the central coordination layer for the schema system.
// It manages schema resolution, type extensions, and annotations in a unified way.
package engine

import (
	"defs.dev/schema/api/core"
)

// SchemaEngine is the central coordination layer that manages all cross-cutting
// concerns in the schema system including schema resolution, type extensions,
// and annotations.
type SchemaEngine interface {
	// Schema Resolution - Named schema management
	RegisterSchema(name string, schema core.Schema) error
	ResolveSchema(name string) (core.Schema, error)
	ResolveReference(ref SchemaReference) (core.Schema, error)
	ListSchemas() []string
	HasSchema(name string) bool

	// Extension Management - Pluggable schema types
	RegisterSchemaType(typeName string, factory SchemaTypeFactory) error
	CreateSchema(typeName string, config any) (core.Schema, error)
	GetAvailableTypes() []string
	ValidateTypeConfig(typeName string, config any) error
	HasSchemaType(typeName string) bool

	// Annotation System - Type-safe metadata
	RegisterAnnotation(name string, schema AnnotationSchema) error
	ValidateAnnotation(name string, value any) error
	GetAnnotationSchema(name string) (AnnotationSchema, bool)
	ListAnnotations() []string
	HasAnnotation(name string) bool

	// Engine Management
	Validate() error
	Reset() error
	Clone() SchemaEngine

	// Configuration
	Config() EngineConfig
	WithConfig(config EngineConfig) SchemaEngine
}

// SchemaReference represents a reference to a named schema that can be resolved
// by the engine. It supports namespacing and versioning for complex scenarios.
type SchemaReference interface {
	// Basic identification
	Name() string
	Namespace() string
	Version() string

	// Computed properties
	FullName() string // namespace:name@version or just name
	IsVersioned() bool
	IsNamespaced() bool

	// Validation
	Validate() error
}

// SchemaTypeFactory creates schema instances from configuration data.
// This enables pluggable schema types that extend the core system.
type SchemaTypeFactory interface {
	// Create schema instance from configuration
	CreateSchema(config any) (core.Schema, error)

	// Validate configuration before creation
	ValidateConfig(config any) error

	// Get schema that describes the configuration format
	GetConfigSchema() core.Schema

	// Metadata about this schema type
	GetMetadata() SchemaTypeMetadata
}

// SchemaTypeMetadata provides information about a schema type factory.
type SchemaTypeMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// AnnotationSchema represents a schema that can be used for annotations.
// It must be composed only of primitive types and their compositions.
type AnnotationSchema interface {
	core.Schema

	// Additional constraint validation for annotations
	ValidateAsAnnotation() error
}

// EngineConfig controls the behavior of the schema engine.
type EngineConfig struct {
	// Caching configuration
	EnableCache  bool `json:"enableCache"`
	MaxCacheSize int  `json:"maxCacheSize"`

	// Resolution configuration
	CircularDepthLimit int `json:"circularDepthLimit"`

	// Validation configuration
	StrictMode         bool `json:"strictMode"`         // Reject unknown annotations
	ValidateOnRegister bool `json:"validateOnRegister"` // Validate schemas when registered

	// Performance configuration
	EnableConcurrency bool `json:"enableConcurrency"` // Allow concurrent operations
}

// DefaultEngineConfig returns a sensible default configuration for the engine.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		EnableCache:        true,
		MaxCacheSize:       1000,
		CircularDepthLimit: 50,
		StrictMode:         false,
		ValidateOnRegister: true,
		EnableConcurrency:  true,
	}
}

// NewSchemaEngine creates a new schema engine with default configuration.
func NewSchemaEngine() SchemaEngine {
	return NewSchemaEngineWithConfig(DefaultEngineConfig())
}

// NewSchemaEngineWithConfig creates a new schema engine with the specified configuration.
func NewSchemaEngineWithConfig(config EngineConfig) SchemaEngine {
	return newSchemaEngineImpl(config)
}

// Common error types for the engine
type EngineError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e EngineError) Error() string {
	return e.Message
}

// Predefined error types
const (
	ErrorTypeSchemaNotFound     = "schema_not_found"
	ErrorTypeSchemaExists       = "schema_exists"
	ErrorTypeTypeNotFound       = "type_not_found"
	ErrorTypeTypeExists         = "type_exists"
	ErrorTypeAnnotationNotFound = "annotation_not_found"
	ErrorTypeAnnotationExists   = "annotation_exists"
	ErrorTypeCircularDependency = "circular_dependency"
	ErrorTypeInvalidConfig      = "invalid_config"
	ErrorTypeValidationFailed   = "validation_failed"
	ErrorTypeInvalidAnnotation  = "invalid_annotation"
)

// Helper functions for creating common errors
func NewSchemaNotFoundError(name string) error {
	return EngineError{
		Type:    ErrorTypeSchemaNotFound,
		Message: "schema not found: " + name,
		Details: map[string]any{"schema_name": name},
	}
}

func NewSchemaExistsError(name string) error {
	return EngineError{
		Type:    ErrorTypeSchemaExists,
		Message: "schema already exists: " + name,
		Details: map[string]any{"schema_name": name},
	}
}

func NewTypeNotFoundError(typeName string) error {
	return EngineError{
		Type:    ErrorTypeTypeNotFound,
		Message: "schema type not found: " + typeName,
		Details: map[string]any{"type_name": typeName},
	}
}

func NewCircularDependencyError(path []string) error {
	return EngineError{
		Type:    ErrorTypeCircularDependency,
		Message: "circular dependency detected in schema resolution",
		Details: map[string]any{"resolution_path": path},
	}
}
