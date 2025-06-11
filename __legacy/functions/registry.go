package functions

import (
	"fmt"
	"sync"

	"defs.dev/schema"
)

// Registry provides named storage for functions with address-based access
type Registry interface {
	// Register with preferred name, get address back
	Register(name string, schema *schema.FunctionSchema, implementation any) (string, error)

	// Register with auto-generated name
	RegisterAnon(schema *schema.FunctionSchema, implementation any) (string, string, error)

	// Get address by name (for sharing)
	GetAddress(name string) (string, error)

	// Get callable function by name
	GetFunction(name string) (schema.Function, error)

	// Utility for naming conflicts
	AllocateFunctionName() string

	// Check if name exists
	Exists(name string) bool

	// List all registered names
	Names() []string

	// Remove function by name
	Remove(name string) error

	// Get the portal used by this registry
	Portal() Portal[any]
}

// BaseRegistry implements Registry using a portal for transformation
type BaseRegistry struct {
	portal      Portal[any]
	functions   map[string]schema.Function // name -> function
	addresses   map[string]string          // name -> address
	nameCounter int
	mu          sync.RWMutex
}

// NewRegistry creates a new registry with the given portal
func NewRegistry[D any](portal Portal[D]) Registry {
	return &BaseRegistry{
		portal:    &PortalWrapper[D]{portal},
		functions: make(map[string]schema.Function),
		addresses: make(map[string]string),
	}
}

// PortalWrapper wraps a typed portal to implement Portal[any]
type PortalWrapper[D any] struct {
	Portal[D]
}

// portalWrapper is an alias for backward compatibility
type portalWrapper[D any] = PortalWrapper[D]

func (w *PortalWrapper[D]) Apply(address string, schema *schema.FunctionSchema, data any) schema.Function {
	// Type assert the data to the expected type D
	typedData, ok := data.(D)
	if !ok {
		panic(fmt.Sprintf("PortalWrapper: expected type %T, got %T", *new(D), data))
	}
	return w.Portal.Apply(address, schema, typedData)
}

func (w *PortalWrapper[D]) GenerateAddress(name string, data any) string {
	// Type assert the data to the expected type D
	typedData, ok := data.(D)
	if !ok {
		panic(fmt.Sprintf("PortalWrapper: expected type %T, got %T", *new(D), data))
	}
	return w.Portal.GenerateAddress(name, typedData)
}

func (r *BaseRegistry) Register(name string, schema *schema.FunctionSchema, implementation any) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; exists {
		return "", &RegistryError{
			Name:    name,
			Type:    "conflict",
			Message: "function already registered with this name",
		}
	}

	// Generate address using portal
	address := r.portal.GenerateAddress(name, implementation)

	// Create function using portal
	function := r.portal.Apply(address, schema, implementation)

	// Store in registry
	r.functions[name] = function
	r.addresses[name] = address

	return address, nil
}

func (r *BaseRegistry) RegisterAnon(schema *schema.FunctionSchema, implementation any) (string, string, error) {
	name := r.AllocateFunctionName()
	address, err := r.Register(name, schema, implementation)
	return name, address, err
}

func (r *BaseRegistry) GetAddress(name string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	address, exists := r.addresses[name]
	if !exists {
		return "", &RegistryError{
			Name:    name,
			Type:    "not_found",
			Message: "function not found in registry",
		}
	}

	return address, nil
}

func (r *BaseRegistry) GetFunction(name string) (schema.Function, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	function, exists := r.functions[name]
	if !exists {
		return nil, &RegistryError{
			Name:    name,
			Type:    "not_found",
			Message: "function not found in registry",
		}
	}

	return function, nil
}

func (r *BaseRegistry) AllocateFunctionName() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nameCounter++
	return fmt.Sprintf("func_%d", r.nameCounter)
}

func (r *BaseRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.functions[name]
	return exists
}

func (r *BaseRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}

	return names
}

func (r *BaseRegistry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; !exists {
		return &RegistryError{
			Name:    name,
			Type:    "not_found",
			Message: "function not found in registry",
		}
	}

	delete(r.functions, name)
	delete(r.addresses, name)

	return nil
}

func (r *BaseRegistry) Portal() Portal[any] {
	return r.portal
}
