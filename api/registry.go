package api

import "context"

// Registry defines the interface for function registries.
type Registry interface {
	// Registration
	Register(name string, fn Function) error
	RegisterTyped(name string, fn TypedFunction) error

	// Retrieval
	Get(name string) (Function, bool)
	GetTyped(name string) (TypedFunction, bool)

	// Listing
	List() []string
	ListWithSchemas() map[string]FunctionSchema

	// Management
	Unregister(name string) error
	Clear() error

	// Validation
	Validate(name string, input any) ValidationResult

	// Execution
	Call(ctx context.Context, name string, params FunctionInput) (FunctionOutput, error)
	CallTyped(ctx context.Context, name string, input any, output any) error
}

// Factory defines the interface for creating registries and other components.
type Factory interface {
	CreateRegistry() Registry
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
