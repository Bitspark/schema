package engine

import (
	"fmt"
	"time"

	"defs.dev/core"
	schemacore "defs.dev/schema/api/core"
)

// SchemaEntity wraps a schema as a core.Entity for integration with the universal entity system
type SchemaEntity struct {
	// Core identity
	id      string
	scope   string
	kind    string
	name    string
	version string

	// Metadata
	description string
	authors     []string
	license     string
	sources     []string
	tags        []string

	// Content and specification
	spec         schemacore.Schema
	examples     []core.Example
	tests        []core.Test
	dependencies []core.EntityRef

	// Trust and verification
	signature  *core.Signature
	verified   bool
	trustLevel core.TrustLevel

	// Resolution metadata
	location    core.EntityLocation
	resolvedAt  time.Time
	cachedUntil time.Time
}

// Core identity methods

func (e *SchemaEntity) ID() string {
	return e.id
}

func (e *SchemaEntity) Scope() string {
	return e.scope
}

func (e *SchemaEntity) Kind() string {
	return e.kind
}

func (e *SchemaEntity) Name() string {
	return e.name
}

func (e *SchemaEntity) Version() string {
	return e.version
}

// Metadata methods

func (e *SchemaEntity) Description() string {
	return e.description
}

func (e *SchemaEntity) Authors() []string {
	return e.authors
}

func (e *SchemaEntity) License() string {
	return e.license
}

func (e *SchemaEntity) Sources() []string {
	return e.sources
}

func (e *SchemaEntity) Tags() []string {
	return e.tags
}

// Content and specification methods

func (e *SchemaEntity) Spec() any {
	return e.spec
}

func (e *SchemaEntity) Examples() []core.Example {
	return e.examples
}

func (e *SchemaEntity) Tests() []core.Test {
	return e.tests
}

func (e *SchemaEntity) Dependencies() []core.EntityRef {
	return e.dependencies
}

// Trust and verification methods

func (e *SchemaEntity) Signature() *core.Signature {
	return e.signature
}

func (e *SchemaEntity) Verified() bool {
	return e.verified
}

func (e *SchemaEntity) TrustLevel() core.TrustLevel {
	return e.trustLevel
}

// Resolution metadata methods

func (e *SchemaEntity) Location() core.EntityLocation {
	return e.location
}

func (e *SchemaEntity) ResolvedAt() time.Time {
	return e.resolvedAt
}

func (e *SchemaEntity) CachedUntil() time.Time {
	return e.cachedUntil
}

// Helper methods for working with the wrapped schema

// GetSchema returns the underlying schema
func (e *SchemaEntity) GetSchema() schemacore.Schema {
	return e.spec
}

// ValidateData validates data against the wrapped schema
func (e *SchemaEntity) ValidateData(data any) schemacore.ValidationResult {
	return e.spec.Validate(data)
}

// ToJSONSchema returns the JSON Schema representation
func (e *SchemaEntity) ToJSONSchema() map[string]any {
	return e.spec.ToJSONSchema()
}

// GenerateExample generates an example using the wrapped schema
func (e *SchemaEntity) GenerateExample() any {
	return e.spec.GenerateExample()
}

// Clone creates a copy of this entity
func (e *SchemaEntity) Clone() *SchemaEntity {
	return &SchemaEntity{
		id:           e.id,
		scope:        e.scope,
		kind:         e.kind,
		name:         e.name,
		version:      e.version,
		description:  e.description,
		authors:      append([]string{}, e.authors...),
		license:      e.license,
		sources:      append([]string{}, e.sources...),
		tags:         append([]string{}, e.tags...),
		spec:         e.spec.Clone(),
		examples:     append([]core.Example{}, e.examples...),
		tests:        append([]core.Test{}, e.tests...),
		dependencies: append([]core.EntityRef{}, e.dependencies...),
		signature:    e.signature, // Signatures are immutable
		verified:     e.verified,
		trustLevel:   e.trustLevel,
		location:     e.location,
		resolvedAt:   e.resolvedAt,
		cachedUntil:  e.cachedUntil,
	}
}

// NewSchemaEntity creates a new SchemaEntity from a schema and metadata
func NewSchemaEntity(schema schemacore.Schema, config SchemaEntityConfig) *SchemaEntity {
	metadata := schema.Metadata()

	return &SchemaEntity{
		id:           config.ID,
		scope:        config.Scope,
		kind:         "schema",
		name:         config.Name,
		version:      config.Version,
		description:  metadata.Description,
		authors:      config.Authors,
		license:      config.License,
		sources:      config.Sources,
		tags:         metadata.Tags,
		spec:         schema,
		examples:     convertToExamples(metadata.Examples),
		tests:        config.Tests,
		dependencies: config.Dependencies,
		signature:    config.Signature,
		verified:     config.Verified,
		trustLevel:   config.TrustLevel,
		location:     config.Location,
		resolvedAt:   time.Now(),
		cachedUntil:  time.Now().Add(config.CacheTTL),
	}
}

// SchemaEntityConfig provides configuration for creating SchemaEntity instances
type SchemaEntityConfig struct {
	ID           string
	Scope        string
	Name         string
	Version      string
	Authors      []string
	License      string
	Sources      []string
	Tests        []core.Test
	Dependencies []core.EntityRef
	Signature    *core.Signature
	Verified     bool
	TrustLevel   core.TrustLevel
	Location     core.EntityLocation
	CacheTTL     time.Duration
}

// DefaultSchemaEntityConfig returns sensible defaults for schema entity configuration
func DefaultSchemaEntityConfig(name string) SchemaEntityConfig {
	return SchemaEntityConfig{
		ID:           fmt.Sprintf("defs.dev/local/schema/%s/latest", name),
		Scope:        "local",
		Name:         name,
		Version:      "latest",
		Authors:      []string{},
		License:      "",
		Sources:      []string{"schema-engine"},
		Tests:        []core.Test{},
		Dependencies: []core.EntityRef{},
		Signature:    nil,
		Verified:     false,
		TrustLevel:   core.TrustLevelCommunity,
		Location:     &schemaLocation{source: "schema-engine", path: name},
		CacheTTL:     time.Hour,
	}
}

// Helper function to convert schema examples to core examples
func convertToExamples(examples []any) []core.Example {
	var coreExamples []core.Example
	for i, example := range examples {
		coreExamples = append(coreExamples, core.Example{
			Name:        fmt.Sprintf("Example %d", i+1),
			Description: "Schema example",
			Input:       example,
			Expected:    nil, // Not available from schema metadata
		})
	}
	return coreExamples
}
