package portal

import (
	"context"
	"fmt"
	"sync"

	"defs.dev/schema/api"
)

// LocalPortalImpl implements api.LocalPortal for in-process function execution
type LocalPortalImpl struct {
	functions map[string]api.Function
	services  map[string]api.Service
	mutex     sync.RWMutex
	idCounter int64
}

// NewLocalPortal creates a new local portal
func NewLocalPortal() api.LocalPortal {
	return &LocalPortalImpl{
		functions: make(map[string]api.Function),
		services:  make(map[string]api.Service),
		idCounter: 0,
	}
}

// Apply registers a function with the portal and returns its address
func (p *LocalPortalImpl) Apply(ctx context.Context, function api.Function) (api.Address, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name := function.Name()

	for _, f := range p.functions {
		if f.Name() == name {
			return nil, fmt.Errorf("function %s already registered", name)
		}
	}

	// Generate unique address
	address := p.generateAddress(name)

	// Store function
	p.functions[address.String()] = function

	return address, nil
}

// ApplyService registers an entire service with the portal
func (p *LocalPortalImpl) ApplyService(ctx context.Context, service api.Service) (api.Address, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Generate unique address for the service
	address := p.generateServiceAddress(service.Name())

	// Store service
	p.services[address.String()] = service

	return address, nil
}

// ResolveFunction resolves an address to a callable function
func (p *LocalPortalImpl) ResolveFunction(ctx context.Context, address api.Address) (api.Function, error) {
	if !address.IsLocal() {
		return nil, fmt.Errorf("address is not local: %s", address.String())
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	function, exists := p.functions[address.String()]
	if !exists {
		return nil, fmt.Errorf("function not found: %s", address.String())
	}

	return function, nil
}

// ResolveService resolves an address to a service
func (p *LocalPortalImpl) ResolveService(ctx context.Context, address api.Address) (api.Service, error) {
	if !address.IsLocal() {
		return nil, fmt.Errorf("address is not local: %s", address.String())
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	service, exists := p.services[address.String()]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", address.String())
	}

	return service, nil
}

// GenerateAddress creates a new address for the given name and metadata
func (p *LocalPortalImpl) GenerateAddress(name string, metadata map[string]any) api.Address {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.generateAddress(name)
}

func (p *LocalPortalImpl) generateAddress(name string) api.Address {
	p.idCounter++
	builder := NewAddressBuilder().
		Scheme("local").
		Path(fmt.Sprintf("/%s", name))

	// Add unique ID as query parameter
	builder = builder.Query("id", fmt.Sprintf("%d", p.idCounter))

	return builder.Build()
}

func (p *LocalPortalImpl) generateServiceAddress(name string) api.Address {
	p.idCounter++
	builder := NewAddressBuilder().
		Scheme("local").
		Path(fmt.Sprintf("/service/%s", name))

	// Add unique ID as query parameter
	builder = builder.Query("id", fmt.Sprintf("%d", p.idCounter))

	return builder.Build()
}

// Schemes returns the schemes this portal handles
func (p *LocalPortalImpl) Schemes() []string {
	return []string{"local"}
}

// Close closes the portal and releases resources
func (p *LocalPortalImpl) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Clear all functions and services
	p.functions = make(map[string]api.Function)
	p.services = make(map[string]api.Service)

	return nil
}

// Health returns the current health status of the portal
func (p *LocalPortalImpl) Health(ctx context.Context) error {
	// Local portal is always healthy if it exists
	return nil
}

// LocalPortal-specific methods

// ListFunctions returns all registered local functions
func (p *LocalPortalImpl) ListFunctions() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	functions := make([]string, 0, len(p.functions))
	for address := range p.functions {
		functions = append(functions, address)
	}
	return functions
}

// GetFunction returns a local function by name
func (p *LocalPortalImpl) GetFunction(name string) (api.Function, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Search for function by name (not address)
	for _, function := range p.functions {
		if function.Name() == name {
			return function, true
		}
	}
	return nil, false
}

// RemoveFunction removes a function from the local registry
func (p *LocalPortalImpl) RemoveFunction(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Find and remove function by name
	var addressToRemove string
	for address, function := range p.functions {
		if function.Name() == name {
			addressToRemove = address
			break
		}
	}

	if addressToRemove == "" {
		return fmt.Errorf("function not found: %s", name)
	}

	delete(p.functions, addressToRemove)
	return nil
}

// GetFunctionByAddress returns a function by its exact address
func (p *LocalPortalImpl) GetFunctionByAddress(address api.Address) (api.Function, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	function, exists := p.functions[address.String()]
	return function, exists
}

// GetServiceByAddress returns a service by its exact address
func (p *LocalPortalImpl) GetServiceByAddress(address api.Address) (api.Service, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	service, exists := p.services[address.String()]
	return service, exists
}

// Stats returns statistics about the local portal
func (p *LocalPortalImpl) Stats() LocalPortalStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return LocalPortalStats{
		FunctionCount: len(p.functions),
		ServiceCount:  len(p.services),
		NextID:        p.idCounter + 1,
	}
}

// LocalPortalStats represents statistics for the local portal
type LocalPortalStats struct {
	FunctionCount int
	ServiceCount  int
	NextID        int64
}

// CallLocalFunction is a convenience method to call a function by name (deprecated)
func (p *LocalPortalImpl) CallLocalFunction(ctx context.Context, name string, params map[string]any) (any, error) {
	function, exists := p.GetFunction(name)
	if !exists {
		return nil, fmt.Errorf("function not found: %s", name)
	}

	// Convert to FunctionData
	data := NewFunctionData(params)
	result, err := function.Call(ctx, data)
	if err != nil {
		return nil, err
	}

	return result.Value(), nil
}
