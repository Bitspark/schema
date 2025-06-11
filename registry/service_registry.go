package registry

import (
	"context"
	"fmt"
	"sync"

	"defs.dev/schema/portal"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
)

// ServiceRegistry manages registered services and their methods.
// It provides service discovery, method registration, and service lifecycle management.
type ServiceRegistry struct {
	mu           sync.RWMutex
	services     map[string]*RegisteredService
	funcRegistry *FunctionRegistry // Embedded function registry for methods
}

// RegisteredService represents a service with its methods and metadata.
type RegisteredService struct {
	Schema       core.ServiceSchema
	Instance     any // The actual service instance (if available)
	Methods      map[string]api.Function
	Metadata     ServiceMetadata
	RegisteredAt int64
}

// ServiceMetadata holds service-level metadata.
type ServiceMetadata struct {
	Version     string
	Tags        []string
	Description string
	Owner       string
}

// NewServiceRegistry creates a new service registry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:     make(map[string]*RegisteredService),
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

	service := &RegisteredService{
		Schema:       schema,
		Methods:      make(map[string]api.Function),
		Metadata:     ServiceMetadata{Version: "1.0.0", Tags: []string{}, Description: fmt.Sprintf("Service %s", name)},
		RegisteredAt: getCurrentTimestamp(),
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
		service.Methods[methodSchema.Name()] = fn
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
		service.Instance = instance
	}

	return nil
}

// Service discovery methods

// GetService retrieves a service by name.
func (r *ServiceRegistry) GetService(name string) (*RegisteredService, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	return service, exists
}

// GetServiceMethod retrieves a specific method from a service.
func (r *ServiceRegistry) GetServiceMethod(serviceName, methodName string) (api.Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceName]
	if !exists {
		return nil, false
	}

	method, exists := service.Methods[methodName]
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

	methods := make([]string, 0, len(service.Methods))
	for methodName := range service.Methods {
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
		for methodName := range service.Methods {
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
	for methodName := range service.Methods {
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
		for methodName := range service.Methods {
			fullMethodName := fmt.Sprintf("%s.%s", serviceName, methodName)
			r.funcRegistry.Unregister(fullMethodName)
		}
	}

	r.services = make(map[string]*RegisteredService)
	return nil
}

// Service execution methods

// CallServiceMethod executes a service method (deprecated - use method directly).
func (r *ServiceRegistry) CallServiceMethod(ctx context.Context, serviceName, methodName string, params map[string]any) (any, error) {
	method, exists := r.GetServiceMethod(serviceName, methodName)
	if !exists {
		return nil, fmt.Errorf("method %s.%s not found", serviceName, methodName)
	}

	// Convert to FunctionData
	data := portal.NewFunctionData(params)
	result, err := method.Call(ctx, data)
	if err != nil {
		return nil, err
	}

	return result.Value(), nil
}

// ValidateServiceMethod validates input for a service method.
func (r *ServiceRegistry) ValidateServiceMethod(serviceName, methodName string, input any) core.ValidationResult {
	method, exists := r.GetServiceMethod(serviceName, methodName)
	if !exists {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{
				{
					Path:       "",
					Message:    fmt.Sprintf("method %s.%s not found", serviceName, methodName),
					Code:       "method_not_found",
					Value:      fmt.Sprintf("%s.%s", serviceName, methodName),
					Expected:   "registered service method",
					Suggestion: "register the service method first or check the name",
					Context:    "service_registry_validation",
				},
			},
		}
	}

	return method.Schema().Validate(input)
}

// Service metadata methods

// GetServiceMetadata returns metadata for a service.
func (r *ServiceRegistry) GetServiceMetadata(name string) (ServiceMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return ServiceMetadata{}, false
	}

	return service.Metadata, true
}

// SetServiceMetadata updates metadata for a service.
func (r *ServiceRegistry) SetServiceMetadata(name string, metadata ServiceMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	service.Metadata = metadata
	return nil
}

// ListServicesByTag returns services that have the specified tag.
func (r *ServiceRegistry) ListServicesByTag(tag string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []string
	for name, service := range r.services {
		for _, t := range service.Metadata.Tags {
			if t == tag {
				services = append(services, name)
				break
			}
		}
	}
	return services
}

// Integration with function registry

// GetFunctionRegistry returns the embedded function registry.
func (r *ServiceRegistry) GetFunctionRegistry() *FunctionRegistry {
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

	if service.Instance == nil {
		return nil, fmt.Errorf("service %s has no registered instance", f.serviceName)
	}

	// This would need reflection to call the actual method on the service instance
	// For now, return a placeholder response
	return portal.NewFunctionDataValue(map[string]any{
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

// ServiceCount returns the total number of registered services.
func (r *ServiceRegistry) ServiceCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.services)
}

// MethodCount returns the total number of registered methods across all services.
func (r *ServiceRegistry) MethodCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, service := range r.services {
		count += len(service.Methods)
	}
	return count
}

// ServiceExists checks if a service is registered.
func (r *ServiceRegistry) ServiceExists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.services[name]
	return exists
}

// MethodExists checks if a specific method exists on a service.
func (r *ServiceRegistry) MethodExists(serviceName, methodName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceName]
	if !exists {
		return false
	}

	_, exists = service.Methods[methodName]
	return exists
}
