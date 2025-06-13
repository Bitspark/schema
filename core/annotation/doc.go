// Package annotation provides a flexible, type-safe annotation system for schema metadata
// and validation. It replaces hardcoded format validation with pluggable annotations
// that can be registered, validated, and applied to schemas throughout the system.
//
// # Core Concepts
//
// The annotation package is built around several key concepts:
//
//   - Annotation: A named piece of metadata with a typed value and validation
//   - AnnotationType: A registered definition that constrains annotation values
//   - AnnotationRegistry: A central registry for managing annotation types and instances
//
// # Architecture
//
// The annotation package sits low in the dependency hierarchy:
//
//	api/core -> annotation -> registry -> native -> schemas -> builders -> engine
//
// This allows lower-level components (registry, native) to use annotations without
// depending on higher-level coordination (engine).
//
// # Basic Usage
//
//	// Create a registry and register built-in types
//	registry := annotation.NewRegistry()
//	annotation.RegisterBuiltinTypes(registry)
//
//	// Create typed annotations
//	formatAnnotation, err := registry.Create("format", "email")
//	minLengthAnnotation, err := registry.Create("minLength", 5)
//
//	// Validate annotations
//	result := formatAnnotation.Validate()
//	if !result.Valid {
//	    // Handle validation errors
//	}
//
// # Custom Annotation Types
//
// You can register custom annotation types with schemas and metadata:
//
//	// Create a schema for the annotation value
//	currencySchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
//	    EnumValues: []string{"USD", "EUR", "GBP"},
//	})
//
//	// Register the annotation type
//	err := registry.RegisterType("currency", currencySchema,
//	    annotation.WithDescription("Currency code for monetary values"),
//	    annotation.WithCategory("financial"),
//	    annotation.WithTags("money", "currency"),
//	    annotation.WithAppliesTo("number", "string"),
//	)
//
//	// Use the custom annotation
//	currencyAnnotation, err := registry.Create("currency", "USD")
//
// # Built-in Annotation Types
//
// The package provides many built-in annotation types that replace hardcoded
// validation in string schemas:
//
//   - String annotations: format, pattern, minLength, maxLength
//   - Numeric annotations: min, max, range
//   - Array annotations: minItems, maxItems, uniqueItems
//   - Validation annotations: required, validators
//   - Metadata annotations: description, examples, default, enum
//
// # Integration with Schema System
//
// Annotations integrate seamlessly with the broader schema system:
//
//	// Schemas can embed annotation metadata
//	schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
//	    Metadata: core.SchemaMetadata{
//	        Annotations: map[string]annotation.Annotation{
//	            "format": formatAnnotation,
//	            "minLength": minLengthAnnotation,
//	        },
//	    },
//	})
//
//	// Native package uses annotations for struct tag parsing
//	type User struct {
//	    Email string `json:"email" format:"email" minLength:"5"`
//	}
//
//	// Annotations are parsed from struct tags and applied to schemas
//	userSchema := native.FromValue[User]()
//
// # Strict vs Non-Strict Mode
//
// The registry can operate in two modes:
//
//   - Non-strict mode (default): Unknown annotation types create flexible annotations
//
//   - Strict mode: Unknown annotation types result in errors
//
//     registry.SetStrictMode(true)  // Enable strict validation
//     annotation, err := registry.Create("unknown", "value")  // Returns error
//
// # Thread Safety
//
// All operations on the annotation registry are thread-safe and can be used
// concurrently from multiple goroutines.
package annotation
