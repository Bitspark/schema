package api

import "context"

// Address represents a unique, addressable identifier for functions across different transports.
// It follows a URL-like format: scheme://[authority]/path[?query][#fragment]
type Address interface {
	// String returns the full address as a string
	String() string

	// Scheme returns the transport scheme (http, https, ws, wss, local, etc.)
	Scheme() string

	// Authority returns the authority portion (host:port for network addresses)
	Authority() string

	// Path returns the path portion
	Path() string

	// Query returns query parameters as a map
	Query() map[string]string

	// Fragment returns the fragment identifier
	Fragment() string

	// IsLocal returns true if this is a local address
	IsLocal() bool

	// IsNetwork returns true if this requires network communication
	IsNetwork() bool
}

// FunctionPortal defines the core interface for function execution portals.
// Portals provide transport abstraction for executing functions across different protocols.
type FunctionPortal interface {
	// Apply registers a function with the portal and returns its address
	Apply(ctx context.Context, function Function) (Address, error)

	// ApplyService registers an entire service with the portal
	ApplyService(ctx context.Context, service Service) (Address, error)

	// ResolveFunction resolves an address to a callable function
	ResolveFunction(ctx context.Context, address Address) (Function, error)

	// ResolveService resolves an address to a service
	ResolveService(ctx context.Context, address Address) (Service, error)

	// GenerateAddress creates a new address for the given name and metadata
	GenerateAddress(name string, metadata map[string]any) Address

	// Schemes returns the schemes this portal handles
	Schemes() []string

	// Close closes the portal and releases resources
	Close() error

	// Health returns the current health status of the portal
	Health(ctx context.Context) error
}

// LocalPortal defines the interface for local (in-process) function execution.
// It combines portal functionality with registry interfaces for local storage.
type LocalPortal interface {
	FunctionPortal
	FunctionRegistry // Local portals ARE function registries
	ServiceRegistry  // Local portals ARE service registries
}

// NetworkPortal defines the interface for network-based portals.
type NetworkPortal interface {
	FunctionPortal

	// Start starts the network portal server
	Start(ctx context.Context) error

	// Stop stops the network portal server
	Stop(ctx context.Context) error

	// ListenAddress returns the address the portal is listening on
	ListenAddress() string

	// BaseURL returns the base URL for this portal
	BaseURL() string

	// GetFunctionRegistry returns the underlying function registry (updated to new interface)
	GetFunctionRegistry() FunctionRegistry

	// GetServiceRegistry returns the underlying service registry (updated to new interface)
	GetServiceRegistry() ServiceRegistry
}

// HTTPPortal defines the interface for HTTP-based function portals.
type HTTPPortal interface {
	NetworkPortal

	// HandleHTTP provides HTTP handler integration
	HandleHTTP() any // Returns http.Handler

	// SetMiddleware sets HTTP middleware
	SetMiddleware(middleware []any)

	// EnableCORS enables CORS support
	EnableCORS(origins []string)
}

// WebSocketPortal defines the interface for WebSocket-based function portals.
type WebSocketPortal interface {
	NetworkPortal

	// HandleWebSocket provides WebSocket handler integration
	HandleWebSocket() any // Returns websocket handler

	// Broadcast sends a message to all connected clients
	Broadcast(ctx context.Context, message any) error

	// SendToClient sends a message to a specific client
	SendToClient(ctx context.Context, clientID string, message any) error

	// ListClients returns all connected client IDs
	ListClients() []string
}

// TestingPortal defines the interface for testing/mock portals.
// It combines portal functionality with registry interfaces for testing.
type TestingPortal interface {
	FunctionPortal
	FunctionRegistry // Testing portals ARE function registries for mocks
	ServiceRegistry  // Testing portals ARE service registries for mocks

	// Testing-specific functionality
	// Mock registers a mock function for testing (extends FunctionRegistry.Register)
	Mock(function Function) Address

	// Verify verifies that expected calls were made
	Verify() error

	// Reset resets all mocks and call history
	Reset()

	// CallHistory returns the history of function calls
	CallHistory() []FunctionCall
}

// FunctionCall represents a recorded function call for testing/debugging.
type FunctionCall struct {
	FunctionName string
	Address      Address
	Input        FunctionData
	Output       FunctionData
	Error        error
	Timestamp    int64
}

// PortalRegistry manages multiple portals and provides unified access.
type PortalRegistry interface {
	Registry // Base registry functionality

	// RegisterPortal registers a portal for specific schemes
	RegisterPortal(schemes []string, portal FunctionPortal) error

	// GetPortal returns a portal that can handle the given address
	GetPortal(address Address) (FunctionPortal, error)

	// ResolveFunction resolves any address to a function using appropriate portal
	ResolveFunction(ctx context.Context, address Address) (Function, error)

	// ListPortals returns all registered portals
	ListPortals() map[string]FunctionPortal

	// UnregisterPortal removes a portal
	UnregisterPortal(schemes []string) error
}

// AddressBuilder provides a fluent interface for building addresses.
type AddressBuilder interface {
	Scheme(scheme string) AddressBuilder
	Authority(authority string) AddressBuilder
	Host(host string) AddressBuilder
	Port(port int) AddressBuilder
	Path(path string) AddressBuilder
	Query(key, value string) AddressBuilder
	Fragment(fragment string) AddressBuilder
	Build() Address
}

// ServicePortal defines the interface for service execution portals.
// This is now just an alias for FunctionPortal since services are handled through FunctionPortal.
type ServicePortal interface {
	FunctionPortal
}

// Portal defines a generic interface for function execution portals.
// This is now just an alias for FunctionPortal.
type Portal interface {
	FunctionPortal
}

// PortalFactory provides a factory for creating different types of portals.
type PortalFactory interface {
	Factory // Extends the base factory

	// CreateLocalPortal creates a local portal with embedded registries
	CreateLocalPortal() LocalPortal

	// CreateHTTPPortal creates an HTTP portal
	CreateHTTPPortal(baseURL string) HTTPPortal

	// CreateWebSocketPortal creates a WebSocket portal
	CreateWebSocketPortal(baseURL string) WebSocketPortal

	// CreateTestingPortal creates a testing portal with mock capabilities
	CreateTestingPortal() TestingPortal

	// CreatePortalRegistry creates a portal registry
	CreatePortalRegistry() PortalRegistry
}
