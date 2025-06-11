package functions

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"defs.dev/schema"
)

// Consumer provides universal function calling by address
type Consumer interface {
	// Call function by address (universal)
	CallAt(ctx context.Context, address string, params schema.FunctionInput) (schema.FunctionOutput, error)

	// Register additional portal for address resolution
	RegisterPortal(portal Portal[any])

	// List registered portals
	Portals() []string
}

// UniversalConsumer implements Consumer with multiple portal support
type UniversalConsumer struct {
	portals map[string]Portal[any] // scheme -> portal
	mu      sync.RWMutex
}

// NewConsumer creates a new universal consumer
func NewConsumer() Consumer {
	return &UniversalConsumer{
		portals: make(map[string]Portal[any]),
	}
}

func (c *UniversalConsumer) CallAt(ctx context.Context, address string, params schema.FunctionInput) (schema.FunctionOutput, error) {
	scheme := extractScheme(address)
	if scheme == "" {
		return schema.FunctionOutput{}, &ConsumerError{
			Address: address,
			Message: "invalid address format: no scheme found",
		}
	}

	c.mu.RLock()
	portal, exists := c.portals[scheme]
	c.mu.RUnlock()

	if !exists {
		return schema.FunctionOutput{}, &ConsumerError{
			Address: address,
			Message: "no portal registered for scheme: " + scheme,
		}
	}

	// Use portal to resolve address to function
	function, err := portal.ResolveFunction(ctx, address)
	if err != nil {
		return schema.FunctionOutput{}, &ConsumerError{
			Address: address,
			Message: "failed to resolve function",
			Cause:   err,
		}
	}

	// Call the resolved function
	return function.Call(ctx, params)
}

func (c *UniversalConsumer) RegisterPortal(portal Portal[any]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	scheme := portal.Scheme()
	c.portals[scheme] = portal
}

func (c *UniversalConsumer) Portals() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	schemes := make([]string, 0, len(c.portals))
	for scheme := range c.portals {
		schemes = append(schemes, scheme)
	}

	return schemes
}

// extractScheme extracts the scheme from an address
func extractScheme(address string) string {
	if idx := strings.Index(address, "://"); idx != -1 {
		return address[:idx]
	}

	// Try parsing as URL for more complex cases
	if u, err := url.Parse(address); err == nil && u.Scheme != "" {
		return u.Scheme
	}

	return ""
}
