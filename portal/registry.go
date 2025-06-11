package portal

import (
	"context"
	"fmt"
	"sync"

	"defs.dev/schema/api"
)

// PortalRegistryImpl implements api.PortalRegistry
type PortalRegistryImpl struct {
	portals map[string]api.FunctionPortal // scheme -> portal
	mutex   sync.RWMutex
}

// NewPortalRegistry creates a new portal registry
func NewPortalRegistry() api.PortalRegistry {
	return &PortalRegistryImpl{
		portals: make(map[string]api.FunctionPortal),
	}
}

// RegisterPortal registers a portal for specific schemes
func (r *PortalRegistryImpl) RegisterPortal(schemes []string, portal api.FunctionPortal) error {
	if len(schemes) == 0 {
		return fmt.Errorf("at least one scheme must be provided")
	}

	if portal == nil {
		return fmt.Errorf("portal cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check for conflicts
	for _, scheme := range schemes {
		if existing, exists := r.portals[scheme]; exists {
			return fmt.Errorf("scheme %s is already registered to portal %T", scheme, existing)
		}
	}

	// Register the portal for all schemes
	for _, scheme := range schemes {
		r.portals[scheme] = portal
	}

	return nil
}

// GetPortal returns a portal that can handle the given address
func (r *PortalRegistryImpl) GetPortal(address api.Address) (api.FunctionPortal, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	scheme := address.Scheme()
	portal, exists := r.portals[scheme]
	if !exists {
		return nil, fmt.Errorf("no portal registered for scheme: %s", scheme)
	}

	return portal, nil
}

// ResolveFunction resolves any address to a function using appropriate portal
func (r *PortalRegistryImpl) ResolveFunction(ctx context.Context, address api.Address) (api.Function, error) {
	portal, err := r.GetPortal(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get portal: %w", err)
	}

	return portal.ResolveFunction(ctx, address)
}

// ListPortals returns all registered portals
func (r *PortalRegistryImpl) ListPortals() map[string]api.FunctionPortal {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Return a copy to prevent modification
	result := make(map[string]api.FunctionPortal)
	for scheme, portal := range r.portals {
		result[scheme] = portal
	}
	return result
}

// Close closes all registered portals
func (r *PortalRegistryImpl) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var errors []error
	for scheme, portal := range r.portals {
		if err := portal.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close portal for scheme %s: %w", scheme, err))
		}
	}

	// Clear all portals
	r.portals = make(map[string]api.FunctionPortal)

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d portals: %v", len(errors), errors)
	}

	return nil
}

// Additional utility methods

// RegisterLocalPortal registers a local portal for the "local" scheme
func (r *PortalRegistryImpl) RegisterLocalPortal(portal api.LocalPortal) error {
	return r.RegisterPortal([]string{"local"}, portal)
}

// RegisterTestingPortal registers a testing portal for "test" and "mock" schemes
func (r *PortalRegistryImpl) RegisterTestingPortal(portal api.TestingPortal) error {
	return r.RegisterPortal([]string{"test", "mock"}, portal)
}

// RegisterHTTPPortal registers an HTTP portal for "http" and "https" schemes
func (r *PortalRegistryImpl) RegisterHTTPPortal(portal api.HTTPPortal) error {
	return r.RegisterPortal([]string{"http", "https"}, portal)
}

// RegisterWebSocketPortal registers a WebSocket portal for "ws" and "wss" schemes
func (r *PortalRegistryImpl) RegisterWebSocketPortal(portal api.WebSocketPortal) error {
	return r.RegisterPortal([]string{"ws", "wss"}, portal)
}

// GetSupportedSchemes returns all schemes currently supported by registered portals
func (r *PortalRegistryImpl) GetSupportedSchemes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	schemes := make([]string, 0, len(r.portals))
	for scheme := range r.portals {
		schemes = append(schemes, scheme)
	}
	return schemes
}

// SupportsScheme returns true if the registry has a portal for the given scheme
func (r *PortalRegistryImpl) SupportsScheme(scheme string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.portals[scheme]
	return exists
}

// UnregisterScheme removes a portal registration for a specific scheme
func (r *PortalRegistryImpl) UnregisterScheme(scheme string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.portals[scheme]; !exists {
		return fmt.Errorf("scheme %s is not registered", scheme)
	}

	delete(r.portals, scheme)
	return nil
}

// Health checks the health of all registered portals
func (r *PortalRegistryImpl) Health(ctx context.Context) map[string]error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	health := make(map[string]error)
	for scheme, portal := range r.portals {
		health[scheme] = portal.Health(ctx)
	}
	return health
}

// Stats returns statistics about the portal registry
func (r *PortalRegistryImpl) Stats() PortalRegistryStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return PortalRegistryStats{
		RegisteredPortals: len(r.portals),
		SupportedSchemes:  r.GetSupportedSchemes(),
	}
}

// PortalRegistryStats represents statistics for the portal registry
type PortalRegistryStats struct {
	RegisteredPortals int
	SupportedSchemes  []string
}

// NewDefaultPortalRegistry creates a registry with common portals pre-registered
func NewDefaultPortalRegistry() api.PortalRegistry {
	registry := NewPortalRegistry().(*PortalRegistryImpl)

	// Register common portals
	localPortal := NewLocalPortal()
	testingPortal := NewTestingPortal()

	registry.RegisterLocalPortal(localPortal)
	registry.RegisterTestingPortal(testingPortal)

	return registry
}
