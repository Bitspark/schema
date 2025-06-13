package registry

import (
	"context"
	"defs.dev/schema/consumers/validation"
	"fmt"
	"sync"

	"defs.dev/schema/api"
	"defs.dev/schema/core"
)

// ServiceRegistry manages registered services and their methods.
// It provides service discovery, method registration, and service lifecycle management.
type ServiceRegistry struct {
	mu           sync.RWMutex
	services     map[string]*registeredServiceImpl
	funcRegistry *FunctionRegistry // Embedded function registry for methods
}

// Ensure ServiceRegistry implements the API interface at compile time
var _ api.ServiceRegistry = (*ServiceRegistry)(nil)

// registeredServiceImpl is the concrete implementation of api.RegisteredService
type registeredServiceImpl struct {
	schema       core.ServiceSchema
	instance     any // The actual service instance (if available)
	methods      map[string]api.Function
	metadata     api.ServiceMetadata
	registeredAt int64
}

// Implement api.RegisteredService interface
func (r *registeredServiceImpl) Schema() core.ServiceSchema {
	return r.schema
}

func (r *registeredServiceImpl) Instance() any {
	return r.instance
}

func (r *registeredServiceImpl) Methods() map[string]api.Function {
	return r.methods
}

func (r *registeredServiceImpl) Metadata() api.ServiceMetadata {
	return r.metadata
}

func (r *registeredServiceImpl) RegisteredAt() int64 {
	return r.registeredAt
}

// NewServiceRegistry creates a new service registry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:     make(map[string]*registeredServiceImpl),
		funcRegistry: NewFunctionRegistry(),
	}
}

// Service registration methods

// RegisterService registers a service with its schema.
func (r *ServiceRegistry) RegisterService(name string, schema core.ServiceSchema) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if schema == nil {
		return fmt.Errorf("service schema cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	service := &registeredServiceImpl{
		schema:       schema,
		methods:      make(map[string]api.Function),
		metadata:     api.ServiceMetadata{Version: "1.0.0", Tags: []string{}, Description: fmt.Sprintf("Service %s", name)},
		registeredAt: getCurrentTimestamp(),
	}

	r.services[name] = service

	// Register all service methods as functions
	for _, methodSchema := range schema.Methods() {
		methodName := fmt.Sprintf("%s.%s", name, methodSchema.Name())
		// Create a function wrapper for the method
		fn := &ServiceMethodFunction{
			serviceName: name,
			methodName:  methodSchema.Name(),
			schema:      methodSchema.Function(),
			registry:    r,
		}
		r.funcRegistry.Register(methodName, fn)
		service.methods[methodSchema.Name()] = fn
	}

	return nil
}

// RegisterServiceWithInstance registers a service with its schema and instance.
func (r *ServiceRegistry) RegisterServiceWithInstance(name string, schema core.ServiceSchema, instance any) error {
	if err := r.RegisterService(name, schema); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if service, exists := r.services[name]; exists {
		service.instance = instance
	}

	return nil
}

// Service discovery methods

// GetService retrieves a service by name.
func (r *ServiceRegistry) GetService(name string) (api.RegisteredService, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, false
	}
	return service, true
}

// GetServiceMethod retrieves a specific method from a service.
func (r *ServiceRegistry) GetServiceMethod(serviceName, methodName string) (api.Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceName]
	if !exists {
		return nil, false
	}

	method, exists := service.methods[methodName]
	return method, exists
}

// ListServices returns all registered service names.
func (r *ServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}

// ListServiceMethods returns all method names for a service.
func (r *ServiceRegistry) ListServiceMethods(serviceName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceName]
	if !exists {
		return []string{}
	}

	methods := make([]string, 0, len(service.methods))
	for methodName := range service.methods {
		methods = append(methods, methodName)
	}
	return methods
}

// ListAllMethods returns all method names across all services in "service.method" format.
func (r *ServiceRegistry) ListAllMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var methods []string
	for serviceName, service := range r.services {
		for methodName := range service.methods {
			methods = append(methods, fmt.Sprintf("%s.%s", serviceName, methodName))
		}
	}
	return methods
}

// Service management methods

// UnregisterService removes a service and all its methods.
func (r *ServiceRegistry) UnregisterService(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	// Unregister all service methods from function registry
	for methodName := range service.methods {
		fullMethodName := fmt.Sprintf("%s.%s", name, methodName)
		r.funcRegistry.Unregister(fullMethodName)
	}

	delete(r.services, name)
	return nil
}

// ClearServices removes all services.
func (r *ServiceRegistry) ClearServices() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Unregister all service methods from function registry
	for serviceName, service := range r.services {
		for methodName := range service.methods {
			fullMethodName := fmt.Sprintf("%s.%s", serviceName, methodName)
			r.funcRegistry.Unregister(fullMethodName)
		}
	}

	r.services = make(map[string]*registeredServiceImpl)
	return nil
}

// Base Registry interface implementation

// List returns all registered service names (same as ListServices).
func (r *ServiceRegistry) List() []string {
	return r.ListServices()
}

// Count returns the number of registered services.
func (r *ServiceRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.services)
}

// Exists checks if a service is registered.
func (r *ServiceRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.services[name]
	return exists
}

// Clear removes all services (same as ClearServices).
func (r *ServiceRegistry) Clear() error {
	return r.ClearServices()
}

// Service execution methods

// CallServiceMethod executes a service method (deprecated - use method directly).
func (r *ServiceRegistry) CallServiceMethod(ctx context.Context, serviceName, methodName string, params map[string]any) (any, error) {
	method, exists := r.GetServiceMethod(serviceName, methodName)
	if !exists {
		return nil, fmt.Errorf("method %s.%s not found", serviceName, methodName)
	}

	// Convert to FunctionData
	data := api.NewFunctionData(params)
	result, err := method.Call(ctx, data)
	if err != nil {
		return nil, err
	}

	return result.Value(), nil
}

// ValidateServiceMethod validates input for a service method.
func (r *ServiceRegistry) ValidateServiceMethod(serviceName, methodName string, input any) validation.ValidationResult {
	_, exists := r.GetServiceMethod(serviceName, methodName)
	if !exists {
		return validation.NewValidationError([]string{}, "method_not_found",
			fmt.Sprintf("method %s.%s not found", serviceName, methodName))
	}

	// Note: Method validation moved to consumer-driven architecture.
	// For now, return a valid result as validation is handled by consumers.
	return validation.NewValidationResult()
}

// Service metadata methods

// GetServiceMetadata returns metadata for a service.
func (r *ServiceRegistry) GetServiceMetadata(name string) (api.ServiceMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return api.ServiceMetadata{}, false
	}

	return service.metadata, true
}

// SetServiceMetadata updates metadata for a service.
func (r *ServiceRegistry) SetServiceMetadata(name string, metadata api.ServiceMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	service.metadata = metadata
	return nil
}

// ListServicesByTag returns services that have the specified tag.
func (r *ServiceRegistry) ListServicesByTag(tag string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []string
	for name, service := range r.services {
		for _, t := range service.metadata.Tags {
			if t == tag {
				services = append(services, name)
				break
			}
		}
	}
	return services
}

// Integration with function registry

// GetFunctionRegistry returns the underlying function registry.
func (r *ServiceRegistry) GetFunctionRegistry() api.FunctionRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.funcRegistry
}

// ServiceMethodFunction wraps a service method as a function.
type ServiceMethodFunction struct {
	serviceName string
	methodName  string
	schema      core.FunctionSchema
	registry    *ServiceRegistry
}

// Ensure ServiceMethodFunction implements api.Function
var _ api.Function = (*ServiceMethodFunction)(nil)

func (f *ServiceMethodFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	// Get the service instance
	f.registry.mu.RLock()
	service, exists := f.registry.services[f.serviceName]
	f.registry.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service %s not found", f.serviceName)
	}

	if service.instance == nil {
		return nil, fmt.Errorf("service %s has no registered instance", f.serviceName)
	}

	// This would need reflection to call the actual method on the service instance
	// For now, return a placeholder response
	return api.NewFunctionDataValue(map[string]any{
		"service": f.serviceName,
		"method":  f.methodName,
		"message": fmt.Sprintf("Called %s.%s", f.serviceName, f.methodName),
	}), nil
}

func (f *ServiceMethodFunction) Schema() core.FunctionSchema {
	return f.schema
}

func (f *ServiceMethodFunction) Name() string {
	return fmt.Sprintf("%s.%s", f.serviceName, f.methodName)
}

// Statistics and introspection methods

// MethodCount returns the total number of registered methods across all services.
func (r *ServiceRegistry) MethodCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, service := range r.services {
		count += len(service.methods)
	}
	return count
}
