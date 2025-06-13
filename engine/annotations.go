package engine

import (
	builders2 "defs.dev/schema/builders"
	"fmt"

	"defs.dev/schema/core"
)

// SimpleAnnotationSchema wraps a core.Schema and adds annotation validation
type SimpleAnnotationSchema struct {
	core.Schema
}

// NewAnnotationSchema creates an annotation schema from a regular schema
func NewAnnotationSchema(schema core.Schema) AnnotationSchema {
	return &SimpleAnnotationSchema{Schema: schema}
}

// ValidateAsAnnotation ensures the schema only uses primitive types
// For now, we'll do a simple type check based on the schema type
func (a *SimpleAnnotationSchema) ValidateAsAnnotation() error {
	schemaType := a.Schema.Type()

	// Allow primitive types and their compositions
	switch schemaType {
	case core.TypeString, core.TypeNumber, core.TypeInteger, core.TypeBoolean:
		return nil // These are always allowed
	case core.TypeArray:
		// Arrays are allowed - we trust the builder to create valid item schemas
		return nil
	case core.TypeStructure:
		// Objects are allowed - we trust the builder to create valid property schemas
		return nil
	case core.TypeFunction, core.TypeService:
		return fmt.Errorf("%s schemas not allowed in annotations", schemaType)
	default:
		return nil // Allow other types for now
	}
}

// Helper functions for creating annotation schemas

// StringAnnotation creates a string-based annotation schema
func StringAnnotation() AnnotationSchema {
	schema := builders2.NewStringSchema().Build()
	return NewAnnotationSchema(schema)
}

// StringEnumAnnotation creates an enum string annotation
func StringEnumAnnotation(values ...string) AnnotationSchema {
	schema := builders2.NewStringSchema().Enum(values...).Build()
	return NewAnnotationSchema(schema)
}

// BooleanAnnotation creates a boolean annotation schema
func BooleanAnnotation() AnnotationSchema {
	schema := builders2.NewBooleanSchema().Build()
	return NewAnnotationSchema(schema)
}

// IntegerAnnotation creates an integer annotation schema
func IntegerAnnotation() AnnotationSchema {
	schema := builders2.NewIntegerSchema().Build()
	return NewAnnotationSchema(schema)
}

// IntegerRangeAnnotation creates an integer annotation with min/max
func IntegerRangeAnnotation(min, max int64) AnnotationSchema {
	schema := builders2.NewIntegerSchema().Min(min).Max(max).Build()
	return NewAnnotationSchema(schema)
}

// NumberAnnotation creates a number annotation schema
func NumberAnnotation() AnnotationSchema {
	schema := builders2.NewNumberSchema().Build()
	return NewAnnotationSchema(schema)
}

// ArrayAnnotation creates an array annotation schema
func ArrayAnnotation(itemSchema core.Schema) AnnotationSchema {
	schema := builders2.NewArraySchema().Items(itemSchema).Build()
	return NewAnnotationSchema(schema)
}

// ObjectAnnotation creates an object annotation builder
func ObjectAnnotation() *ObjectAnnotationBuilder {
	return &ObjectAnnotationBuilder{
		builder: builders2.NewObjectSchema(),
	}
}

// ObjectAnnotationBuilder provides a fluent interface for building object annotations
type ObjectAnnotationBuilder struct {
	builder core.ObjectSchemaBuilder
}

func (b *ObjectAnnotationBuilder) Property(name string, schema core.Schema) *ObjectAnnotationBuilder {
	b.builder = b.builder.Property(name, schema)
	return b
}

func (b *ObjectAnnotationBuilder) Required(names ...string) *ObjectAnnotationBuilder {
	b.builder = b.builder.Required(names...)
	return b
}

func (b *ObjectAnnotationBuilder) Build() AnnotationSchema {
	schema := b.builder.Build()
	return NewAnnotationSchema(schema)
}

// Common annotation patterns

// PatternAnnotation creates a pattern annotation for common design patterns
func PatternAnnotation() AnnotationSchema {
	return StringEnumAnnotation(
		"service",
		"component",
		"entity",
		"value_object",
		"aggregate",
		"repository",
		"factory",
		"controller",
		"middleware",
	)
}

// BehaviorAnnotation creates a behavior annotation for common behaviors
func BehaviorAnnotation() AnnotationSchema {
	return ArrayAnnotation(StringEnumAnnotation(
		"stateful",
		"stateless",
		"cached",
		"persistent",
		"transient",
		"singleton",
		"immutable",
		"async",
		"sync",
	))
}

// DeploymentAnnotation creates a deployment configuration annotation
func DeploymentAnnotation() AnnotationSchema {
	return ObjectAnnotation().
		Property("strategy", StringEnumAnnotation("rolling", "blue-green", "canary")).
		Property("replicas", IntegerRangeAnnotation(1, 100)).
		Property("resources", ObjectAnnotation().
			Property("cpu", StringAnnotation()).
			Property("memory", StringAnnotation()).
			Property("storage", StringAnnotation()).
			Build()).
		Property("health_check", ObjectAnnotation().
			Property("endpoint", StringAnnotation()).
			Property("interval", StringAnnotation()).
			Property("timeout", StringAnnotation()).
			Build()).
		Required("strategy").
		Build()
}

// CachingAnnotation creates a caching configuration annotation
func CachingAnnotation() AnnotationSchema {
	return ObjectAnnotation().
		Property("strategy", StringEnumAnnotation("redis", "memory", "disk", "none")).
		Property("ttl", IntegerRangeAnnotation(0, 86400)). // 0 to 24 hours in seconds
		Property("key_pattern", StringAnnotation()).
		Property("invalidation", StringEnumAnnotation("time", "event", "manual")).
		Required("strategy").
		Build()
}

// PerformanceAnnotation creates a performance configuration annotation
func PerformanceAnnotation() AnnotationSchema {
	return ObjectAnnotation().
		Property("timeout", StringAnnotation()).
		Property("rate_limit", IntegerRangeAnnotation(1, 10000)).
		Property("batch_size", IntegerRangeAnnotation(1, 1000)).
		Property("async", BooleanAnnotation()).
		Build()
}

// SecurityAnnotation creates a security configuration annotation
func SecurityAnnotation() AnnotationSchema {
	return ObjectAnnotation().
		Property("authentication", StringEnumAnnotation("required", "optional", "none")).
		Property("authorization", ArrayAnnotation(StringAnnotation())).
		Property("encryption", StringEnumAnnotation("required", "optional", "none")).
		Property("audit", BooleanAnnotation()).
		Build()
}

// Update the registerBuiltinAnnotations method
func (e *schemaEngineImpl) registerBuiltinAnnotations() {
	// Register common pattern annotations
	annotations := map[string]AnnotationSchema{
		"pattern":     PatternAnnotation(),
		"behavior":    BehaviorAnnotation(),
		"deployment":  DeploymentAnnotation(),
		"caching":     CachingAnnotation(),
		"performance": PerformanceAnnotation(),
		"security":    SecurityAnnotation(),
	}

	for name, schema := range annotations {
		// Register without validation since these are built-in
		e.annotMu.Lock()
		e.annotations[name] = schema
		e.annotMu.Unlock()
	}
}
