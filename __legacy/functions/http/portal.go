package http

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"

	"defs.dev/schema"
)

// Portal interface methods

// Apply transforms a local function into an HTTP endpoint and returns its address
func (p *HTTPPortal) Apply(address string, funcSchema *schema.FunctionSchema, data any) (Function, error) {
	// For HTTP portal, data should be a FunctionHandler
	handler, ok := data.(schema.FunctionHandler)
	if !ok {
		return nil, fmt.Errorf("HTTP portal requires FunctionHandler, got %T", data)
	}

	// Register the handler as an HTTP endpoint
	if err := p.RegisterHandler(funcSchema.Metadata().Name, address, funcSchema, handler); err != nil {
		return nil, NewServerError(err, map[string]interface{}{
			"address": address,
			"name":    funcSchema.Metadata().Name,
		})
	}

	// Return a function that represents the endpoint (metadata only)
	return &HTTPEndpointFunction{
		address: address,
		schema:  funcSchema,
		portal:  p,
	}, nil
}

// GenerateAddress creates a unique HTTP address for a function
func (p *HTTPPortal) GenerateAddress(name string, data any) (string, error) {
	// Generate a unique identifier
	id, err := generateUniqueID()
	if err != nil {
		return "", NewAddressError("", err)
	}

	// Build address components
	scheme := "http"
	if p.config.TLSConfig != nil {
		scheme = "https"
	}

	host := p.config.Host
	if host == "" {
		host = "localhost"
	}

	port := p.config.Port
	if port == 0 {
		if scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}

	basePath := p.config.BasePath
	if basePath == "" {
		basePath = "/api"
	}

	// Remove leading slash from basePath for construction
	basePath = strings.TrimLeft(basePath, "/")

	// Construct address
	address := fmt.Sprintf("%s://%s:%d/%s/%s/%s", scheme, host, port, basePath, name, id)

	return address, nil
}

// Scheme returns the schemes supported by this portal
func (p *HTTPPortal) Scheme() []string {
	return []string{"http", "https"}
}

// ResolveFunction resolves an address back to a callable client function
func (p *HTTPPortal) ResolveFunction(ctx context.Context, address string) (Function, error) {
	// Parse the address to validate it
	parsedURL, err := url.Parse(address)
	if err != nil {
		return nil, NewAddressError(address, err)
	}

	// Check if scheme is supported
	if !p.isSchemeSupported(parsedURL.Scheme) {
		return nil, NewAddressError(address, fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme))
	}

	// Always create a client function - even for locally registered functions
	// This ensures we can make HTTP calls to test the server
	var funcSchema *schema.FunctionSchema

	// Check if we have schema information from local registration
	if registration, exists := p.functions[address]; exists {
		funcSchema = registration.Schema
	}

	return p.CreateClientFunction(address, funcSchema, &HTTPEndpoint{
		BaseURL: fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host),
		Path:    parsedURL.Path,
		Method:  "POST",
	}), nil
}

// Helper methods

// isSchemeSupported checks if a URL scheme is supported
func (p *HTTPPortal) isSchemeSupported(scheme string) bool {
	for _, s := range p.Scheme() {
		if s == scheme {
			return true
		}
	}
	return false
}

// AddMiddleware adds middleware to the portal
func (p *HTTPPortal) AddMiddleware(middleware Middleware) {
	p.middleware = append(p.middleware, middleware)
}

// GetRegisteredFunctions returns all registered functions
func (p *HTTPPortal) GetRegisteredFunctions() map[string]*FunctionRegistration {
	// Return a copy to prevent external modification
	result := make(map[string]*FunctionRegistration)
	for k, v := range p.functions {
		result[k] = v
	}
	return result
}

// HTTPEndpointFunction represents a function endpoint (metadata only)
type HTTPEndpointFunction struct {
	address string
	schema  *schema.FunctionSchema
	portal  *HTTPPortal
}

// Call is not implemented for endpoint functions (they represent the server side)
func (f *HTTPEndpointFunction) Call(ctx context.Context, params map[string]any) (any, error) {
	return nil, fmt.Errorf("HTTPEndpointFunction represents a server endpoint, use a client function to call it")
}

// Schema returns the function schema
func (f *HTTPEndpointFunction) Schema() *schema.FunctionSchema {
	return f.schema
}

// Address returns the function address
func (f *HTTPEndpointFunction) Address() string {
	return f.address
}

// Utility functions

// generateUniqueID generates a unique identifier for addresses
func generateUniqueID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex string
	return fmt.Sprintf("%x", bytes), nil
}

// parseAddressComponents parses an HTTP address into components
func parseAddressComponents(address string) (scheme, host, port, path string, err error) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return "", "", "", "", err
	}

	scheme = parsedURL.Scheme
	host = parsedURL.Hostname()
	port = parsedURL.Port()
	path = parsedURL.Path

	// Set default ports if not specified
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return scheme, host, port, path, nil
}

// ValidateAddress validates an HTTP address format
func ValidateAddress(address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return NewAddressError(address, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return NewAddressError(address, fmt.Errorf("invalid scheme: %s", parsedURL.Scheme))
	}

	if parsedURL.Host == "" {
		return NewAddressError(address, fmt.Errorf("missing host"))
	}

	return nil
}
