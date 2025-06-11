package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"defs.dev/schema"
)

// namedSchema represents a schema definition with optional parameters
type namedSchema struct {
	schema     schema.Schema
	parameters []string
}

// Registry manages named schemas with optional parameterization
type Registry struct {
	mu           sync.RWMutex
	schemas      map[string]*namedSchema
	resolveCache map[cacheKey]schema.Schema
}

// cacheKey represents a cache key for resolved schemas
type cacheKey struct {
	name      string
	paramHash string
}

// New creates a new empty registry
func New() *Registry {
	return &Registry{
		schemas:      make(map[string]*namedSchema),
		resolveCache: make(map[cacheKey]schema.Schema),
	}
}

// Define registers a schema with optional parameters
func (r *Registry) Define(name string, s schema.Schema, params ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemas[name] = &namedSchema{
		schema:     s,
		parameters: params,
	}

	// Clear cache entries that might be affected
	r.clearCacheForSchema(name)

	return nil
}

// Get retrieves a schema that has no parameters
func (r *Registry) Get(name string) (schema.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	named, exists := r.schemas[name]
	if !exists {
		return nil, NewNotFoundError(name)
	}

	if len(named.parameters) > 0 {
		return nil, NewInvalidParamsError(name, named.parameters, nil)
	}

	return named.schema, nil
}

// Apply applies parameters to a schema and returns the resolved result
func (r *Registry) Apply(name string, params map[string]schema.Schema) (schema.Schema, error) {
	r.mu.RLock()
	named, exists := r.schemas[name]
	if !exists {
		r.mu.RUnlock()
		return nil, NewNotFoundError(name)
	}

	// Validate parameters
	if err := validateParameters(name, named.parameters, params); err != nil {
		r.mu.RUnlock()
		return nil, err
	}

	// Check cache
	key := r.buildCacheKey(name, params)
	if cached, exists := r.resolveCache[key]; exists {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	// Resolve parameters
	visiting := make(map[string]bool)
	visiting[name] = true

	resolved, err := resolveParameters(named.schema, params, visiting)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.mu.Lock()
	r.resolveCache[key] = resolved
	r.mu.Unlock()

	return resolved, nil
}

// Build creates a builder for applying parameters to a schema
func (r *Registry) Build(name string) *NamedSchemaBuilder {
	return &NamedSchemaBuilder{
		registry: r,
		name:     name,
		params:   make(map[string]schema.Schema),
	}
}

// Ref creates a reference to a named schema
func (r *Registry) Ref(name string) *SchemaRef {
	return &SchemaRef{
		registry:   r,
		name:       name,
		parameters: nil,
	}
}

// List returns all schema names in the registry
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}

// Parameters returns the parameter names for a schema
func (r *Registry) Parameters(name string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if named, exists := r.schemas[name]; exists {
		// Return a copy to prevent modification
		params := make([]string, len(named.parameters))
		copy(params, named.parameters)
		return params
	}
	return nil
}

// Exists checks if a schema exists in the registry
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.schemas[name]
	return exists
}

// Merge merges another registry into this one
func (r *Registry) Merge(other *Registry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	for name, named := range other.schemas {
		r.schemas[name] = &namedSchema{
			schema:     named.schema.Clone(),
			parameters: append([]string(nil), named.parameters...),
		}
	}

	// Clear cache after merge
	r.resolveCache = make(map[cacheKey]schema.Schema)

	return nil
}

// Clone creates a deep copy of the registry
func (r *Registry) Clone() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := New()
	for name, named := range r.schemas {
		clone.schemas[name] = &namedSchema{
			schema:     named.schema.Clone(),
			parameters: append([]string(nil), named.parameters...),
		}
	}

	return clone
}

// MarshalJSON serializes the registry to JSON
func (r *Registry) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type schemaEntry struct {
		Parameters []string        `json:"parameters"`
		Schema     json.RawMessage `json:"schema"`
	}

	entries := make(map[string]schemaEntry)
	for name, named := range r.schemas {
		schemaJSON, err := json.Marshal(named.schema.ToJSONSchema())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema '%s': %w", name, err)
		}

		entries[name] = schemaEntry{
			Parameters: named.parameters,
			Schema:     schemaJSON,
		}
	}

	return json.Marshal(map[string]any{
		"schemas": entries,
	})
}

// UnmarshalJSON deserializes the registry from JSON
func (r *Registry) UnmarshalJSON(data []byte) error {
	// This is a simplified implementation
	// A full implementation would need to reconstruct schema objects from JSON Schema
	return fmt.Errorf("UnmarshalJSON not yet implemented")
}

// SaveToFile saves the registry to a JSON file
func (r *Registry) SaveToFile(filename string) error {
	data, err := r.MarshalJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads the registry from a JSON file
func (r *Registry) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return r.UnmarshalJSON(data)
}

// Helper methods

// buildCacheKey creates a cache key for the given schema name and parameters
func (r *Registry) buildCacheKey(name string, params map[string]schema.Schema) cacheKey {
	// Simple hash based on parameter names - could be improved
	hash := ""
	for paramName := range params {
		hash += paramName + ","
	}
	return cacheKey{
		name:      name,
		paramHash: hash,
	}
}

// clearCacheForSchema removes cache entries related to a schema
func (r *Registry) clearCacheForSchema(name string) {
	for key := range r.resolveCache {
		if key.name == name {
			delete(r.resolveCache, key)
		}
	}
}
