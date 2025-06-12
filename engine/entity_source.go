package engine

import (
	"context"
	"fmt"
	"time"

	"defs.dev/core"
	schemacore "defs.dev/schema/api/core"
)

// SchemaEntitySource implements core.EntitySource to make schemas available
// through the universal entity resolution system
type SchemaEntitySource struct {
	engine   SchemaEngine
	config   CoreIntegrationConfig
	priority int
}

// NewSchemaEntitySource creates a new schema entity source
func NewSchemaEntitySource(engine SchemaEngine, config CoreIntegrationConfig) *SchemaEntitySource {
	return &SchemaEntitySource{
		engine:   engine,
		config:   config,
		priority: config.SourcePriority,
	}
}

// Name returns the unique name of this source
func (s *SchemaEntitySource) Name() string {
	return s.config.SourceName
}

// Priority returns the priority of this source (higher = more priority)
func (s *SchemaEntitySource) Priority() int {
	return s.priority
}

// Resolve attempts to resolve a schema entity from this source
func (s *SchemaEntitySource) Resolve(ctx context.Context, entityType core.EntityType, name string, resCtx core.ResolutionContext) (core.Entity, error) {
	// Only handle schema entities
	if entityType != core.EntityTypeSchema {
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	// Create a schema reference from the name
	ref, err := ParseReference(name)
	if err != nil {
		return nil, fmt.Errorf("invalid schema reference: %w", err)
	}

	// Resolve the schema using our engine
	schema, err := s.engine.ResolveReference(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve schema: %w", err)
	}

	// Convert schema to entity
	entity := s.convertSchemaToEntity(schema, name, resCtx)
	return entity, nil
}

// List returns available schema entities from this source
func (s *SchemaEntitySource) List(ctx context.Context, entityType core.EntityType, resCtx core.ResolutionContext) ([]core.EntityInfo, error) {
	// Only handle schema entities
	if entityType != core.EntityTypeSchema {
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	// Get all registered schemas
	schemaNames := s.engine.ListSchemas()

	var entities []core.EntityInfo
	for _, name := range schemaNames {
		// Try to resolve each schema to get its metadata
		schema, err := s.engine.ResolveSchema(name)
		if err != nil {
			continue // Skip schemas that can't be resolved
		}

		info := core.EntityInfo{
			ID:          fmt.Sprintf("defs.dev/%s/schema/%s/latest", s.config.DefaultScope, name),
			Type:        core.EntityTypeSchema,
			Name:        name,
			Version:     "latest", // Default version
			Scope:       s.config.DefaultScope,
			Description: schema.Metadata().Description,
			Authors:     s.config.DefaultAuthors,
			Tags:        schema.Metadata().Tags,
			TrustLevel:  core.TrustLevelCommunity,
			UpdatedAt:   time.Now(),
		}

		entities = append(entities, info)
	}

	return entities, nil
}

// Search searches for schema entities in this source
func (s *SchemaEntitySource) Search(ctx context.Context, query core.SearchQuery) ([]core.EntityInfo, error) {
	// For now, implement basic search by listing all and filtering
	// In a full implementation, this would use more sophisticated search
	entities, err := s.List(ctx, core.EntityTypeSchema, core.DefaultResolutionContext())
	if err != nil {
		return nil, err
	}

	// Simple filtering based on query
	var filtered []core.EntityInfo
	for _, entity := range entities {
		if s.matchesQuery(entity, query) {
			filtered = append(filtered, entity)
		}
	}

	// Apply limit
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
	}

	return filtered, nil
}

// HealthCheck returns the health status of this source
func (s *SchemaEntitySource) HealthCheck(ctx context.Context) error {
	// Check if the engine is functional by trying to validate it
	return s.engine.Validate()
}

// Helper methods

func (s *SchemaEntitySource) convertSchemaToEntity(schema schemacore.Schema, name string, resCtx core.ResolutionContext) *SchemaEntity {
	metadata := schema.Metadata()

	return &SchemaEntity{
		id:           fmt.Sprintf("defs.dev/%s/schema/%s/latest", s.config.DefaultScope, name),
		scope:        s.config.DefaultScope,
		kind:         "schema",
		name:         name,
		version:      "latest",
		description:  metadata.Description,
		authors:      s.config.DefaultAuthors,
		license:      "", // Could be extracted from metadata
		sources:      []string{s.Name()},
		tags:         metadata.Tags,
		spec:         schema,
		examples:     s.convertExamples(metadata.Examples),
		tests:        []core.Test{},      // Could be populated from schema tests
		dependencies: []core.EntityRef{}, // Could be extracted from schema references
		signature:    nil,                // Not implemented yet
		verified:     false,
		trustLevel:   core.TrustLevelCommunity,
		location:     &schemaLocation{source: s.Name(), path: name},
		resolvedAt:   time.Now(),
		cachedUntil:  time.Now().Add(s.config.CacheTTL),
	}
}

func (s *SchemaEntitySource) convertExamples(examples []any) []core.Example {
	var coreExamples []core.Example
	for i, example := range examples {
		coreExamples = append(coreExamples, core.Example{
			Name:        fmt.Sprintf("Example %d", i+1),
			Description: "Generated example from schema",
			Input:       example,
			Expected:    nil, // Not available from schema metadata
		})
	}
	return coreExamples
}

func (s *SchemaEntitySource) matchesQuery(entity core.EntityInfo, query core.SearchQuery) bool {
	// Simple text matching - in practice, this would be more sophisticated
	if query.Query != "" {
		queryLower := query.Query
		if entity.Name == queryLower || entity.Description == queryLower {
			return true
		}
		for _, tag := range entity.Tags {
			if tag == queryLower {
				return true
			}
		}
	}

	// Filter by scope
	if query.Scope != "" && entity.Scope != query.Scope {
		return false
	}

	// Filter by trust level
	if query.TrustLevel > 0 && entity.TrustLevel < query.TrustLevel {
		return false
	}

	// Filter by tags
	if len(query.Tags) > 0 {
		hasTag := false
		for _, queryTag := range query.Tags {
			for _, entityTag := range entity.Tags {
				if entityTag == queryTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// schemaLocation implements core.EntityLocation
type schemaLocation struct {
	source string
	path   string
}

func (l *schemaLocation) Protocol() string {
	return "schema-engine"
}

func (l *schemaLocation) Address() string {
	return fmt.Sprintf("%s:%s", l.source, l.path)
}

func (l *schemaLocation) String() string {
	return l.Address()
}
