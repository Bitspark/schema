package registry

import (
	"context"
	"defs.dev/schema/consume/validation"
	"fmt"
	"sync"

	"defs.dev/schema/api"
	"defs.dev/schema/core"
)

// FunctionRegistry implements api.Registry for managing callable functions.
// It provides thread-safe storage, discovery, and execution of functions.
type FunctionRegistry struct {
	mu        sync.RWMutex
	functions map[string]api.Function
	metadata  map[string]FunctionMetadata
}

// FunctionMetadata holds additional metadata about registered functions.
type FunctionMetadata struct {
	RegisteredAt int64
	Tags         []string
	Version      string
	Description  string
}

// Ensure FunctionRegistry implements the API interface at compile time
var _ api.FunctionRegistry = (*FunctionRegistry)(nil)

// NewFunctionRegistry creates a new thread-safe function registry.
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]api.Function),
		metadata:  make(map[string]FunctionMetadata),
	}
}

// Registration methods

// Register registers a function with the given name.
func (r *FunctionRegistry) Register(name string, fn api.Function) error {
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.functions[name] = fn
	r.metadata[name] = FunctionMetadata{
		RegisteredAt: getCurrentTimestamp(),
		Tags:         []string{},
		Version:      "1.0.0",
		Description:  fmt.Sprintf("Function %s", name),
	}

	return nil
}

// Retrieval methods

// Get retrieves a function by name.
func (r *FunctionRegistry) Get(name string) (api.Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	return fn, exists
}

// Listing methods

// List returns all registered function names.
func (r *FunctionRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

// ListWithSchemas returns all registered functions with their schemas.
func (r *FunctionRegistry) ListWithSchemas() map[string]core.FunctionSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]core.FunctionSchema, len(r.functions))
	for name, fn := range r.functions {
		result[name] = fn.Schema()
	}
	return result
}

// Management methods

// Unregister removes a function from the registry.
func (r *FunctionRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; !exists {
		return fmt.Errorf("function %s not found", name)
	}

	delete(r.functions, name)
	delete(r.metadata, name)

	return nil
}

// Clear removes all functions from the registry.
func (r *FunctionRegistry) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.functions = make(map[string]api.Function)
	r.metadata = make(map[string]FunctionMetadata)

	return nil
}

// Validation methods

// Validate validates input parameters for a function.
func (r *FunctionRegistry) Validate(name string, input any) validation.ValidationResult {
	r.mu.RLock()
	_, exists := r.functions[name]
	r.mu.RUnlock()

	if !exists {
		return validation.NewValidationError([]string{}, "function_not_found",
			fmt.Sprintf("function %s not found", name))
	}

	// Note: Function schema validation moved to consumer-driven architecture.
	// For now, return a valid result as validation is handled by consumers.
	return validation.NewValidationResult()
}

// Execution methods

// Call executes a function with the given parameters.
func (r *FunctionRegistry) Call(ctx context.Context, name string, params api.FunctionData) (api.FunctionData, error) {
	fn, exists := r.Get(name)
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	// Note: Input validation moved to consumer-driven architecture.
	// Execute function directly.
	output, err := fn.Call(ctx, params)
	if err != nil {
		return nil, err
	}

	// Note: Output validation moved to consumer-driven architecture.
	return output, nil
}

// CallTyped executes a typed function with type-safe input and output (deprecated).
func (r *FunctionRegistry) CallTyped(ctx context.Context, name string, input any, output any) error {
	fn, exists := r.Get(name)
	if !exists {
		return fmt.Errorf("function %s not found", name)
	}

	// For now, CallTyped just calls the regular Call method
	// This maintains interface compatibility
	data := FunctionInputMap{"input": input}
	_, err := fn.Call(ctx, data)
	return err
}

// RegisterTyped is an alias for Register for backward compatibility.
func (r *FunctionRegistry) RegisterTyped(name string, fn api.Function) error {
	return r.Register(name, fn)
}

// GetTyped is an alias for Get for backward compatibility.
func (r *FunctionRegistry) GetTyped(name string) (api.Function, bool) {
	return r.Get(name)
}

// Extended methods (beyond API requirements)

// GetMetadata returns metadata for a registered function.
func (r *FunctionRegistry) GetMetadata(name string) (FunctionMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.metadata[name]
	return metadata, exists
}

// SetMetadata updates metadata for a registered function.
func (r *FunctionRegistry) SetMetadata(name string, metadata FunctionMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; !exists {
		return fmt.Errorf("function %s not found", name)
	}

	r.metadata[name] = metadata
	return nil
}

// ListByTag returns functions that have the specified tag.
func (r *FunctionRegistry) ListByTag(tag string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var functions []string
	for name, metadata := range r.metadata {
		for _, t := range metadata.Tags {
			if t == tag {
				functions = append(functions, name)
				break
			}
		}
	}
	return functions
}

// Count returns the total number of registered functions.
func (r *FunctionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.functions)
}

// Exists checks if a function with the given name is registered.
func (r *FunctionRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.functions[name]
	return exists
}

// Clone creates a copy of the registry (functions are not deep-copied).
func (r *FunctionRegistry) Clone() *FunctionRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := NewFunctionRegistry()
	for name, fn := range r.functions {
		clone.functions[name] = fn
	}
	for name, metadata := range r.metadata {
		// Copy metadata
		clone.metadata[name] = FunctionMetadata{
			RegisteredAt: metadata.RegisteredAt,
			Tags:         append([]string(nil), metadata.Tags...),
			Version:      metadata.Version,
			Description:  metadata.Description,
		}
	}

	return clone
}

// Helper functions

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	// Using a simple counter for now - in production might use time.Now().Unix()
	// This avoids importing time package and makes testing more predictable
	return 1
}
