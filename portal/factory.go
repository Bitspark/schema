package portal

import (
	"defs.dev/schema/api"
	"defs.dev/schema/runtime/registry"
)

// DefaultPortalFactory implements the api.PortalFactory interface for creating portal components.
type DefaultPortalFactory struct {
	*registry.DefaultFactory // Embed the registry factory
}

// Ensure DefaultPortalFactory implements the PortalFactory interface at compile time
var _ api.PortalFactory = (*DefaultPortalFactory)(nil)

// NewDefaultPortalFactory creates a new portal factory instance.
func NewDefaultPortalFactory() *DefaultPortalFactory {
	return &DefaultPortalFactory{
		DefaultFactory: registry.NewDefaultFactory(),
	}
}

// CreateLocalPortal creates a local portal with embedded registries
func (f *DefaultPortalFactory) CreateLocalPortal() api.LocalPortal {
	return NewLocalPortal()
}

// CreateHTTPPortal creates an HTTP portal
func (f *DefaultPortalFactory) CreateHTTPPortal(baseURL string) api.HTTPPortal {
	// Extract configuration from baseURL if needed
	config := DefaultHTTPConfig()
	// TODO: Parse baseURL to configure host/port
	return NewHTTPPortal(config)
}

// CreateWebSocketPortal creates a WebSocket portal
func (f *DefaultPortalFactory) CreateWebSocketPortal(baseURL string) api.WebSocketPortal {
	config := DefaultWebSocketConfig()
	// TODO: Parse baseURL to configure host/port
	return NewWebSocketPortal(config, nil, nil) // Uses default registries
}

// CreateTestingPortal creates a testing portal with mock capabilities
func (f *DefaultPortalFactory) CreateTestingPortal() api.TestingPortal {
	return NewTestingPortal()
}

// CreatePortalRegistry creates a portal registry
func (f *DefaultPortalFactory) CreatePortalRegistry() api.PortalRegistry {
	return NewPortalRegistry()
}

// Global portal factory instance for convenience
var DefaultFactory = NewDefaultPortalFactory()

// Convenience functions using the default portal factory
func CreateLocalPortal() api.LocalPortal {
	return DefaultFactory.CreateLocalPortal()
}

func CreateHTTPPortal(baseURL string) api.HTTPPortal {
	return DefaultFactory.CreateHTTPPortal(baseURL)
}

func CreateTestingPortal() api.TestingPortal {
	return DefaultFactory.CreateTestingPortal()
}

// Advanced convenience functions for dependency injection
func CreateSharedPortalSystem() (api.FunctionRegistry, api.ServiceRegistry, api.LocalPortal, api.HTTPPortal, api.WebSocketPortal) {
	// Create shared registries
	funcRegistry := registry.CreateFunctionRegistry()
	serviceRegistry := registry.CreateServiceRegistry()

	// Create portals that share the same registries
	localPortal := NewLocalPortalWithRegistries(funcRegistry, serviceRegistry)
	httpPortal := NewHTTPPortalWithRegistries(DefaultHTTPConfig(), funcRegistry, serviceRegistry)
	wsPortal := NewWebSocketPortal(DefaultWebSocketConfig(), funcRegistry, serviceRegistry)

	return funcRegistry, serviceRegistry, localPortal, httpPortal, wsPortal
}

// NewLocalPortalWithRegistries creates a local portal with injected registries
func NewLocalPortalWithRegistries(funcRegistry api.FunctionRegistry, serviceRegistry api.ServiceRegistry) api.LocalPortal {
	// TODO: Update LocalPortal constructor to support dependency injection
	return NewLocalPortal() // For now, use existing constructor
}

// NewHTTPPortalWithRegistries creates an HTTP portal with injected registries
func NewHTTPPortalWithRegistries(config *HTTPConfig, funcRegistry api.FunctionRegistry, serviceRegistry api.ServiceRegistry) api.HTTPPortal {
	// TODO: Update HTTPPortal constructor to support dependency injection
	return NewHTTPPortal(config) // For now, use existing constructor
}
