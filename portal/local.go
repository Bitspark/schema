package portal

import (
	"context"
	"fmt"
	"sync"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
	"defs.dev/schema/registry"
)

// LocalPortalImpl implements api.LocalPortal for in-process function execution
// It combines portal functionality with registry storage
type LocalPortalImpl struct {
	// Embed actual registries for storage
	funcRegistry    api.FunctionRegistry
	serviceRegistry api.ServiceRegistry

	// Portal-specific fields
	mutex     sync.RWMutex
	idCounter int64

	// Address mappings for portal functionality
	addressToFunction map[string]api.Function
	addressToService  map[string]api.Service
}

// Ensure LocalPortalImpl implements all required interfaces at compile time
var _ api.LocalPortal = (*LocalPortalImpl)(nil)
var _ api.FunctionRegistry = (*LocalPortalImpl)(nil)
var _ api.ServiceRegistry = (*LocalPortalImpl)(nil)

// NewLocalPortal creates a new local portal
func NewLocalPortal() api.LocalPortal {
	return &LocalPortalImpl{
		funcRegistry:      registry.NewFunctionRegistry(),
		serviceRegistry:   registry.NewServiceRegistry(),
		addressToFunction: make(map[string]api.Function),
		addressToService:  make(map[string]api.Service),
		idCounter:         0,
	}
}

// Portal-specific methods (FunctionPortal interface)

// Apply registers a function with the portal and returns its address
func (p *LocalPortalImpl) Apply(ctx context.Context, function api.Function) (api.Address, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Register with underlying function registry
	err := p.funcRegistry.Register(function.Name(), function)
	if err != nil {
		return nil, fmt.Errorf("failed to register function: %w", err)
	}

	// Generate unique address for portal access
	address := p.generateAddress(function.Name())
	p.addressToFunction[address.String()] = function

	return address, nil
}

// ApplyService registers an entire service with the portal
func (p *LocalPortalImpl) ApplyService(ctx context.Context, service api.Service) (api.Address, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Register with underlying service registry
	err := p.serviceRegistry.RegisterService(service.Schema().Name(), service.Schema())
	if err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	// Generate unique address for portal access
	address := p.generateServiceAddress(service.Schema().Name())
	p.addressToService[address.String()] = service

	return address, nil
}

// ResolveFunction resolves an address to a callable function
func (p *LocalPortalImpl) ResolveFunction(ctx context.Context, address api.Address) (api.Function, error) {
	if !address.IsLocal() {
		return nil, fmt.Errorf("address is not local: %s", address.String())
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	function, exists := p.addressToFunction[address.String()]
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

	service, exists := p.addressToService[address.String()]
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

// Schemes returns the schemes this portal handles
func (p *LocalPortalImpl) Schemes() []string {
	return []string{"local"}
}

// Close closes the portal and releases resources
func (p *LocalPortalImpl) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Clear registries
	p.funcRegistry.Clear()
	p.serviceRegistry.Clear()

	// Clear address mappings
	p.addressToFunction = make(map[string]api.Function)
	p.addressToService = make(map[string]api.Service)

	return nil
}

// Health returns the current health status of the portal
func (p *LocalPortalImpl) Health(ctx context.Context) error {
	// Local portal is always healthy if it exists
	return nil
}

// FunctionRegistry interface implementation (delegated to embedded registry)

func (p *LocalPortalImpl) Register(name string, fn api.Function) error {
	return p.funcRegistry.Register(name, fn)
}

func (p *LocalPortalImpl) RegisterTyped(name string, fn api.Function) error {
	return p.funcRegistry.RegisterTyped(name, fn)
}

func (p *LocalPortalImpl) Get(name string) (api.Function, bool) {
	return p.funcRegistry.Get(name)
}

func (p *LocalPortalImpl) GetTyped(name string) (api.Function, bool) {
	return p.funcRegistry.GetTyped(name)
}

func (p *LocalPortalImpl) ListWithSchemas() map[string]core.FunctionSchema {
	return p.funcRegistry.ListWithSchemas()
}

func (p *LocalPortalImpl) Unregister(name string) error {
	// Also remove from address mappings if it exists
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Remove from underlying registry
	err := p.funcRegistry.Unregister(name)
	if err != nil {
		return err
	}

	// Remove from address mappings
	for addr, fn := range p.addressToFunction {
		if fn.Name() == name {
			delete(p.addressToFunction, addr)
			break
		}
	}

	return nil
}

func (p *LocalPortalImpl) Validate(name string, input any) core.ValidationResult {
	return p.funcRegistry.Validate(name, input)
}

func (p *LocalPortalImpl) Call(ctx context.Context, name string, params api.FunctionData) (api.FunctionData, error) {
	return p.funcRegistry.Call(ctx, name, params)
}

func (p *LocalPortalImpl) CallTyped(ctx context.Context, name string, input any, output any) error {
	return p.funcRegistry.CallTyped(ctx, name, input, output)
}

// ServiceRegistry interface implementation (delegated to embedded registry)

func (p *LocalPortalImpl) RegisterService(name string, schema core.ServiceSchema) error {
	return p.serviceRegistry.RegisterService(name, schema)
}

func (p *LocalPortalImpl) RegisterServiceWithInstance(name string, schema core.ServiceSchema, instance any) error {
	return p.serviceRegistry.RegisterServiceWithInstance(name, schema, instance)
}

func (p *LocalPortalImpl) GetService(name string) (api.RegisteredService, bool) {
	return p.serviceRegistry.GetService(name)
}

func (p *LocalPortalImpl) GetServiceMethod(serviceName, methodName string) (api.Function, bool) {
	return p.serviceRegistry.GetServiceMethod(serviceName, methodName)
}

func (p *LocalPortalImpl) ListServices() []string {
	return p.serviceRegistry.ListServices()
}

func (p *LocalPortalImpl) ListServiceMethods(serviceName string) []string {
	return p.serviceRegistry.ListServiceMethods(serviceName)
}

func (p *LocalPortalImpl) ListAllMethods() []string {
	return p.serviceRegistry.ListAllMethods()
}

func (p *LocalPortalImpl) UnregisterService(name string) error {
	// Also remove from address mappings if it exists
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Remove from underlying registry
	err := p.serviceRegistry.UnregisterService(name)
	if err != nil {
		return err
	}

	// Remove from address mappings
	for addr, svc := range p.addressToService {
		if svc.Schema().Name() == name {
			delete(p.addressToService, addr)
			break
		}
	}

	return nil
}

func (p *LocalPortalImpl) CallServiceMethod(ctx context.Context, serviceName, methodName string, params map[string]any) (any, error) {
	return p.serviceRegistry.CallServiceMethod(ctx, serviceName, methodName, params)
}

func (p *LocalPortalImpl) ValidateServiceMethod(serviceName, methodName string, input any) core.ValidationResult {
	return p.serviceRegistry.ValidateServiceMethod(serviceName, methodName, input)
}

func (p *LocalPortalImpl) GetFunctionRegistry() api.FunctionRegistry {
	return p.serviceRegistry.GetFunctionRegistry()
}

// Base Registry interface implementation (common to both function and service registries)

func (p *LocalPortalImpl) List() []string {
	// Return combined list of functions and services
	functions := p.funcRegistry.List()
	services := p.serviceRegistry.List()

	combined := make([]string, 0, len(functions)+len(services))
	combined = append(combined, functions...)
	combined = append(combined, services...)

	return combined
}

func (p *LocalPortalImpl) Count() int {
	return p.funcRegistry.Count() + p.serviceRegistry.Count()
}

func (p *LocalPortalImpl) Exists(name string) bool {
	return p.funcRegistry.Exists(name) || p.serviceRegistry.Exists(name)
}

func (p *LocalPortalImpl) Clear() error {
	err1 := p.funcRegistry.Clear()
	err2 := p.serviceRegistry.Clear()

	p.mutex.Lock()
	p.addressToFunction = make(map[string]api.Function)
	p.addressToService = make(map[string]api.Service)
	p.mutex.Unlock()

	if err1 != nil {
		return err1
	}
	return err2
}

// Helper methods for address generation

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

// Legacy compatibility methods (if needed)

// ListFunctions returns all registered local functions (legacy compatibility)
func (p *LocalPortalImpl) ListFunctions() []string {
	return p.funcRegistry.List()
}

// GetFunction returns a local function by name (legacy compatibility)
func (p *LocalPortalImpl) GetFunction(name string) (api.Function, bool) {
	return p.funcRegistry.Get(name)
}

// RemoveFunction removes a function from the local registry (legacy compatibility)
func (p *LocalPortalImpl) RemoveFunction(name string) error {
	return p.Unregister(name)
}
