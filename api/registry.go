package api

import (
	"context"
	"defs.dev/schema/consumers/validation"

	"defs.dev/schema/core"
)

// Registry defines the base interface for all registry types with common operations.
type Registry interface {
	// Common listing operations
	List() []string
	Count() int
	Exists(name string) bool

	// Common management operations
	Clear() error
}

// FunctionRegistry defines the interface for function registries.
type FunctionRegistry interface {
	Registry

	// Function registration
	Register(name string, fn Function) error
	RegisterTyped(name string, fn Function) error

	// Function retrieval
	Get(name string) (Function, bool)
	GetTyped(name string) (Function, bool)
	ListWithSchemas() map[string]core.FunctionSchema

	// Function management
	Unregister(name string) error

	// Function validation
	Validate(name string, input any) validation.ValidationResult

	// Function execution
	Call(ctx context.Context, name string, params FunctionData) (FunctionData, error)
	CallTyped(ctx context.Context, name string, input any, output any) error
}

// ServiceRegistry defines the interface for service registries.
type ServiceRegistry interface {
	Registry

	// Service registration
	RegisterService(name string, schema core.ServiceSchema) error
	RegisterServiceWithInstance(name string, schema core.ServiceSchema, instance any) error

	// Service retrieval
	GetService(name string) (RegisteredService, bool)
	GetServiceMethod(serviceName, methodName string) (Function, bool)
	ListServices() []string
	ListServiceMethods(serviceName string) []string
	ListAllMethods() []string

	// Service management
	UnregisterService(name string) error

	// Service execution
	CallServiceMethod(ctx context.Context, serviceName, methodName string, params map[string]any) (any, error)
	ValidateServiceMethod(serviceName, methodName string, input any) validation.ValidationResult

	// Access to underlying function registry for service methods
	GetFunctionRegistry() FunctionRegistry
}

// RegisteredService represents a service with its methods and metadata.
type RegisteredService interface {
	Schema() core.ServiceSchema
	Instance() any
	Methods() map[string]Function
	Metadata() ServiceMetadata
	RegisteredAt() int64
}

// ServiceMetadata holds service-level metadata.
type ServiceMetadata struct {
	Version     string
	Tags        []string
	Description string
	Owner       string
}

// Factory defines the interface for creating registries and other components.
type Factory interface {
	CreateFunctionRegistry() FunctionRegistry
	CreateServiceRegistry() ServiceRegistry
	CreateConsumer() Consumer
}

// Middleware defines common middleware interface for different function contexts.
type Middleware interface {
	Process(ctx context.Context, next func(context.Context) error) error
}

// HTTPMiddleware defines middleware specific to HTTP function contexts.
type HTTPMiddleware interface {
	Middleware
	ProcessHTTP(ctx context.Context, req any, res any, next func() error) error
}

// WebSocketMiddleware defines middleware specific to WebSocket function contexts.
type WebSocketMiddleware interface {
	Middleware
	ProcessWebSocket(ctx context.Context, conn any, next func() error) error
}
