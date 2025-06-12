package annotation

import (
	"fmt"
	"strconv"

	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// RegisterBuiltinTypes registers all built-in annotation types with the given registry.
// This replaces hardcoded format validation with flexible, type-safe annotations.
func RegisterBuiltinTypes(registry AnnotationRegistry) error {
	// String format annotations
	if err := registerStringAnnotations(registry); err != nil {
		return fmt.Errorf("failed to register string annotations: %w", err)
	}

	// Numeric constraint annotations
	if err := registerNumericAnnotations(registry); err != nil {
		return fmt.Errorf("failed to register numeric annotations: %w", err)
	}

	// Array constraint annotations
	if err := registerArrayAnnotations(registry); err != nil {
		return fmt.Errorf("failed to register array annotations: %w", err)
	}

	// Validation annotations
	if err := registerValidationAnnotations(registry); err != nil {
		return fmt.Errorf("failed to register validation annotations: %w", err)
	}

	// Metadata annotations
	if err := registerMetadataAnnotations(registry); err != nil {
		return fmt.Errorf("failed to register metadata annotations: %w", err)
	}

	return nil
}

// String format annotations (replaces hardcoded formats)
func registerStringAnnotations(registry AnnotationRegistry) error {
	// Format annotation with enum of supported formats
	formatSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		EnumValues: []string{
			"email", "url", "uuid", "phone", "date", "time", "date-time",
			"ipv4", "ipv6", "hostname", "password", "binary", "base64",
		},
		Metadata: core.SchemaMetadata{
			Description: "String format specification",
			Examples:    []any{"email", "url", "uuid"},
		},
	})

	err := registry.RegisterType("format", formatSchema,
		WithDescription("Specifies the format for string validation"),
		WithCategory("string"),
		WithTags("validation", "format"),
		WithAppliesTo("string"),
		WithExamples("email", "url", "uuid", "phone"),
	)
	if err != nil {
		return err
	}

	// Pattern annotation for regex validation
	patternSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Description: "Regular expression pattern for string validation",
			Examples:    []any{"^[a-zA-Z0-9_]+$", "\\d{3}-\\d{3}-\\d{4}"},
		},
	})

	err = registry.RegisterType("pattern", patternSchema,
		WithDescription("Regular expression pattern for string validation"),
		WithCategory("string"),
		WithTags("validation", "regex"),
		WithAppliesTo("string"),
		WithExamples("^[a-zA-Z0-9_]+$", "\\d{3}-\\d{3}-\\d{4}"),
	)
	if err != nil {
		return err
	}

	// String length constraints
	lengthSchema := schemas.NewIntegerSchema(schemas.IntegerSchemaConfig{
		Minimum: func() *int64 { v := int64(0); return &v }(),
		Metadata: core.SchemaMetadata{
			Description: "String length constraint",
			Examples:    []any{5, 10, 100},
		},
	})

	err = registry.RegisterType("minLength", lengthSchema,
		WithDescription("Minimum string length"),
		WithCategory("string"),
		WithTags("validation", "length"),
		WithAppliesTo("string"),
		WithExamples(1, 5, 10),
	)
	if err != nil {
		return err
	}

	err = registry.RegisterType("maxLength", lengthSchema,
		WithDescription("Maximum string length"),
		WithCategory("string"),
		WithTags("validation", "length"),
		WithAppliesTo("string"),
		WithExamples(50, 100, 500),
	)
	if err != nil {
		return err
	}

	return nil
}

// Numeric constraint annotations
func registerNumericAnnotations(registry AnnotationRegistry) error {
	// Numeric minimum/maximum constraints
	numberSchema := schemas.NewNumberSchema(schemas.NumberSchemaConfig{
		Metadata: core.SchemaMetadata{
			Description: "Numeric constraint value",
			Examples:    []any{0, 1.5, 100, -10.5},
		},
	})

	err := registry.RegisterType("min", numberSchema,
		WithDescription("Minimum numeric value"),
		WithCategory("numeric"),
		WithTags("validation", "constraint"),
		WithAppliesTo("number", "integer"),
		WithExamples(0, 1, 10.5, -5),
	)
	if err != nil {
		return err
	}

	err = registry.RegisterType("max", numberSchema,
		WithDescription("Maximum numeric value"),
		WithCategory("numeric"),
		WithTags("validation", "constraint"),
		WithAppliesTo("number", "integer"),
		WithExamples(100, 1000, 99.99),
	)
	if err != nil {
		return err
	}

	// Numeric range (array of [min, max])
	rangeSchema := schemas.NewArraySchema(schemas.ArraySchemaConfig{
		ItemSchema: numberSchema,
		MinItems:   func() *int { v := 2; return &v }(),
		MaxItems:   func() *int { v := 2; return &v }(),
		Metadata: core.SchemaMetadata{
			Description: "Numeric range as [min, max] array",
			Examples:    []any{[]float64{0, 100}, []float64{-10.5, 50.5}},
		},
	})

	err = registry.RegisterType("range", rangeSchema,
		WithDescription("Numeric range constraint as [min, max]"),
		WithCategory("numeric"),
		WithTags("validation", "constraint", "range"),
		WithAppliesTo("number", "integer"),
		WithExamples([]float64{0, 100}, []float64{-10, 10}),
	)
	if err != nil {
		return err
	}

	return nil
}

// Array constraint annotations
func registerArrayAnnotations(registry AnnotationRegistry) error {
	// Array length constraints
	lengthSchema := schemas.NewIntegerSchema(schemas.IntegerSchemaConfig{
		Minimum: func() *int64 { v := int64(0); return &v }(),
		Metadata: core.SchemaMetadata{
			Description: "Array length constraint",
			Examples:    []any{0, 1, 10, 100},
		},
	})

	err := registry.RegisterType("minItems", lengthSchema,
		WithDescription("Minimum number of array items"),
		WithCategory("array"),
		WithTags("validation", "length"),
		WithAppliesTo("array"),
		WithExamples(0, 1, 5),
	)
	if err != nil {
		return err
	}

	err = registry.RegisterType("maxItems", lengthSchema,
		WithDescription("Maximum number of array items"),
		WithCategory("array"),
		WithTags("validation", "length"),
		WithAppliesTo("array"),
		WithExamples(10, 50, 100),
	)
	if err != nil {
		return err
	}

	// Unique items constraint
	uniqueSchema := schemas.NewBooleanSchema(schemas.BooleanSchemaConfig{
		Metadata: core.SchemaMetadata{
			Description: "Whether array items must be unique",
			Examples:    []any{true, false},
		},
	})

	err = registry.RegisterType("uniqueItems", uniqueSchema,
		WithDescription("Require unique items in array"),
		WithCategory("array"),
		WithTags("validation", "uniqueness"),
		WithAppliesTo("array"),
		WithExamples(true, false),
	)
	if err != nil {
		return err
	}

	return nil
}

// Validation annotations
func registerValidationAnnotations(registry AnnotationRegistry) error {
	// Required field annotation
	requiredSchema := schemas.NewBooleanSchema(schemas.BooleanSchemaConfig{
		Metadata: core.SchemaMetadata{
			Description: "Whether the field is required",
			Examples:    []any{true, false},
		},
	})

	err := registry.RegisterType("required", requiredSchema,
		WithDescription("Mark field as required or optional"),
		WithCategory("validation"),
		WithTags("validation", "required"),
		WithAppliesTo("string", "number", "integer", "boolean", "array", "object"),
		WithExamples(true, false),
	)
	if err != nil {
		return err
	}

	// Custom validators list
	validatorsSchema := schemas.NewArraySchema(schemas.ArraySchemaConfig{
		ItemSchema: schemas.NewStringSchema(schemas.StringSchemaConfig{}),
		Metadata: core.SchemaMetadata{
			Description: "List of custom validator names to apply",
			Examples:    []any{[]string{"email"}, []string{"phone", "required"}},
		},
	})

	err = registry.RegisterType("validators", validatorsSchema,
		WithDescription("List of custom validators to apply"),
		WithCategory("validation"),
		WithTags("validation", "custom"),
		WithAppliesTo("string", "number", "integer", "boolean", "array", "object"),
		WithExamples([]string{"email"}, []string{"phone", "custom"}),
	)
	if err != nil {
		return err
	}

	return nil
}

// Metadata annotations
func registerMetadataAnnotations(registry AnnotationRegistry) error {
	// Description annotation
	descriptionSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Description: "Human-readable description",
			Examples:    []any{"User's email address", "Product price in USD"},
		},
	})

	err := registry.RegisterType("description", descriptionSchema,
		WithDescription("Human-readable description of the field"),
		WithCategory("metadata"),
		WithTags("documentation", "description"),
		WithAppliesTo("string", "number", "integer", "boolean", "array", "object"),
		WithExamples("User's email address", "Product price"),
	)
	if err != nil {
		return err
	}

	// Examples annotation (array of any values)
	examplesSchema := schemas.NewArraySchema(schemas.ArraySchemaConfig{
		ItemSchema: createAnySchema(), // TODO: Need an "any" schema type
		Metadata: core.SchemaMetadata{
			Description: "Example values for documentation",
			Examples:    []any{[]any{"example1", "example2"}, []any{123, 456}},
		},
	})

	err = registry.RegisterType("examples", examplesSchema,
		WithDescription("Example values for documentation"),
		WithCategory("metadata"),
		WithTags("documentation", "examples"),
		WithAppliesTo("string", "number", "integer", "boolean", "array", "object"),
		WithExamples([]any{"user@example.com"}, []any{100, 200, 300}),
	)
	if err != nil {
		return err
	}

	// Default value annotation (any value)
	defaultSchema := createAnySchema() // TODO: Need an "any" schema type

	err = registry.RegisterType("default", defaultSchema,
		WithDescription("Default value when field is not provided"),
		WithCategory("metadata"),
		WithTags("default", "fallback"),
		WithAppliesTo("string", "number", "integer", "boolean", "array", "object"),
		WithExamples("", 0, false, []any{}, map[string]any{}),
	)
	if err != nil {
		return err
	}

	// Enum annotation (array of values)
	enumSchema := schemas.NewArraySchema(schemas.ArraySchemaConfig{
		ItemSchema: createAnySchema(), // TODO: Need an "any" schema type
		MinItems:   func() *int { v := 1; return &v }(),
		Metadata: core.SchemaMetadata{
			Description: "Allowed values for the field",
			Examples:    []any{[]string{"admin", "user", "guest"}, []int{1, 2, 3}},
		},
	})

	err = registry.RegisterType("enum", enumSchema,
		WithDescription("Enumeration of allowed values"),
		WithCategory("validation"),
		WithTags("validation", "enum", "choices"),
		WithAppliesTo("string", "number", "integer"),
		WithExamples([]string{"red", "green", "blue"}, []int{1, 2, 3}),
	)
	if err != nil {
		return err
	}

	return nil
}

// Helper function to create a flexible "any" schema
// TODO: This should be replaced with a proper "any" schema type
func createAnySchema() core.Schema {
	// For now, create a very permissive object schema
	return schemas.NewObjectSchema(schemas.ObjectSchemaConfig{
		AdditionalProperties: true,
		Metadata: core.SchemaMetadata{
			Description: "Accepts any value",
		},
	})
}

// Helper functions to convert annotation values to appropriate types
func ParseIntAnnotation(annotation Annotation) (int, error) {
	switch v := annotation.Value().(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

func ParseFloatAnnotation(annotation Annotation) (float64, error) {
	switch v := annotation.Value().(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func ParseStringAnnotation(annotation Annotation) (string, error) {
	switch v := annotation.Value().(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func ParseBoolAnnotation(annotation Annotation) (bool, error) {
	switch v := annotation.Value().(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}
