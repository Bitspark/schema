package consumer

import (
	"reflect"

	"defs.dev/schema/api/core"
)

// ConsumerPurpose classifies what a consumer does. Purposes are *not* hard-coded – users can define their own.
type ConsumerPurpose string

// ----------------------------------------------------------------------------
//  Generic Consumer Result (purpose-agnostic)
// ----------------------------------------------------------------------------

// ConsumerResult is the uniform envelope returned from every consumer.
// It is intentionally minimal to stay purpose-agnostic while still providing
// enough introspection for tooling.
type ConsumerResult interface {
	// Kind is an arbitrary semantic identifier – e.g. "validation", "generation".
	Kind() string
	// Value returns the underlying strongly-typed result.
	Value() any
	// GoType returns the concrete Go type of Value() for reflection.
	GoType() reflect.Type
}

// resultImpl is a simple implementation of ConsumerResult.
type resultImpl struct {
	kind  string
	value any
}

func (r resultImpl) Kind() string { return r.kind }
func (r resultImpl) Value() any   { return r.value }
func (r resultImpl) GoType() reflect.Type {
	if r.value == nil {
		return nil
	}
	return reflect.TypeOf(r.value)
}

// NewResult creates a new generic ConsumerResult.
func NewResult(kind string, value any) ConsumerResult {
	return resultImpl{kind: kind, value: value}
}

// ----------------------------------------------------------------------------
//  Processing Context
// ----------------------------------------------------------------------------

// ProcessingContext carries auxiliary information to consumers.
type ProcessingContext struct {
	// Schema currently being processed.
	Schema core.Schema
	// JSON-pointer-like path (e.g. ["preferences", "theme"].
	Path []string
	// Immediate parent schema (nil for root).
	Parent core.Schema
	// Value being processed (for value consumers).
	Value core.Value[any]
	// Arbitrary caller-supplied options.
	Options map[string]any
}

// WithPath executes a function with a path segment added, then removes it.
// This simplifies path tracking in recursive processing.
func (c *ProcessingContext) WithPath(segment string, fn func() error) error {
	c.Path = append(c.Path, segment)
	defer func() { c.Path = c.Path[:len(c.Path)-1] }()
	return fn()
}

// ----------------------------------------------------------------------------
//  Consumer Interfaces
// ----------------------------------------------------------------------------

// AnnotationConsumer processes schemas that match ApplicableSchemas().
// Implementations should be stateless or concurrency-safe.
type AnnotationConsumer interface {
	Name() string
	Purpose() ConsumerPurpose

	// Declarative applicability filter.
	ApplicableSchemas() SchemaCondition

	// Processing entry point.
	ProcessSchema(ctx ProcessingContext) (ConsumerResult, error)

	// Optional metadata for discovery / docs.
	Metadata() ConsumerMetadata
}

// ValueConsumer processes values with strong typing via Go generics.
// Implementations should be stateless or concurrency-safe.
type ValueConsumer interface {
	Name() string
	Purpose() ConsumerPurpose

	// Declarative applicability filter (based on schema).
	ApplicableSchemas() SchemaCondition

	// Type-erased processing entry point.
	ProcessValue(ctx ProcessingContext, value core.Value[any]) (ConsumerResult, error)

	// Optional metadata for discovery / docs.
	Metadata() ConsumerMetadata
}

// ConsumerMetadata provides machine- and human-readable information about a consumer.
type ConsumerMetadata struct {
	Name         string          `json:"name"`
	Purpose      ConsumerPurpose `json:"purpose"`
	Description  string          `json:"description,omitempty"`
	Version      string          `json:"version,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	ResultKind   string          `json:"result_kind,omitempty"` // matches ConsumerResult.Kind()
	ResultGoType string          `json:"result_go_type,omitempty"`
	Extras       map[string]any  `json:"extras,omitempty"`
}
