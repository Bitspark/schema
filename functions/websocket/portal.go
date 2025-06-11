package websocket

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"

	"defs.dev/schema"
)

// Portal interface implementation for WebSocket

// Apply registers a function with the WebSocket portal
func (p *WebSocketPortal) Apply(address string, funcSchema *schema.FunctionSchema, data any) (Function, error) {
	// Validate inputs
	if address == "" {
		return nil, NewAddressError(address, fmt.Errorf("address cannot be empty"))
	}
	if funcSchema == nil {
		return nil, NewValidationError("funcSchema", "nil", fmt.Errorf("function schema is required"))
	}

	// Extract handler from data
	handler, ok := data.(schema.FunctionHandler)
	if !ok {
		return nil, NewValidationError("data", fmt.Sprintf("%T", data), fmt.Errorf("expected schema.FunctionHandler"))
	}

	// Register the handler
	err := p.RegisterHandler(funcSchema.Metadata().Name, address, funcSchema, handler)
	if err != nil {
		return nil, NewRegistrationError(funcSchema.Metadata().Name, err)
	}

	// Return a function that represents the endpoint (metadata only)
	return &WebSocketEndpointFunction{
		address: address,
		schema:  funcSchema,
		portal:  p,
	}, nil
}

// GenerateAddress creates a unique WebSocket address for a function
func (p *WebSocketPortal) GenerateAddress(name string, data any) (string, error) {
	// Generate a unique identifier
	id, err := generateUniqueID()
	if err != nil {
		return "", NewAddressError("", err)
	}

	// Build address components
	scheme := "ws"
	if p.config.TLSConfig != nil {
		scheme = "wss"
	}

	host := p.config.Host
	if host == "" {
		host = "localhost"
	}

	port := p.config.Port
	if port == 0 {
		if scheme == "wss" {
			port = 443
		} else {
			port = 80
		}
	}

	path := p.config.Path
	if path == "" {
		path = "/ws"
	}

	// Remove leading slash from path for construction
	path = strings.TrimLeft(path, "/")

	// Construct address
	address := fmt.Sprintf("%s://%s:%d/%s/%s/%s", scheme, host, port, path, name, id)

	return address, nil
}

// Scheme returns the schemes supported by this portal
func (p *WebSocketPortal) Scheme() []string {
	return []string{"ws", "wss"}
}

// ResolveFunction resolves an address back to a callable client function
func (p *WebSocketPortal) ResolveFunction(ctx context.Context, address string) (Function, error) {
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
	// This ensures we can make WebSocket calls to test the server
	var funcSchema *schema.FunctionSchema

	// Check if we have schema information from local registration
	functionName := p.extractFunctionFromAddress(address)
	if registration, exists := p.functions[functionName]; exists {
		funcSchema = registration.Schema
	}

	return p.CreateClientFunction(address, funcSchema, &WebSocketEndpoint{
		URL: address,
	}), nil
}

// Helper methods

// isSchemeSupported checks if a URL scheme is supported
func (p *WebSocketPortal) isSchemeSupported(scheme string) bool {
	for _, s := range p.Scheme() {
		if s == scheme {
			return true
		}
	}
	return false
}

// AddMiddleware adds middleware to the portal
func (p *WebSocketPortal) AddMiddleware(middleware Middleware) {
	p.middleware = append(p.middleware, middleware)
}

// GetRegisteredFunctions returns all registered functions
func (p *WebSocketPortal) GetRegisteredFunctions() map[string]*FunctionRegistration {
	// Return a copy to prevent external modification
	result := make(map[string]*FunctionRegistration)
	p.mu.RLock()
	for k, v := range p.functions {
		result[k] = v
	}
	p.mu.RUnlock()
	return result
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

// parseAddressComponents parses a WebSocket address into components
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
		if scheme == "wss" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return scheme, host, port, path, nil
}

// ValidateAddress validates a WebSocket address format
func ValidateAddress(address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return NewAddressError(address, err)
	}

	if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" {
		return NewAddressError(address, fmt.Errorf("invalid scheme: %s", parsedURL.Scheme))
	}

	if parsedURL.Host == "" {
		return NewAddressError(address, fmt.Errorf("missing host"))
	}

	return nil
}

// Close closes the portal and all its resources
func (p *WebSocketPortal) Close() error {
	// Close all connections
	if err := p.CloseAllConnections(); err != nil {
		return err
	}

	// Stop server if running
	if p.server != nil {
		ctx := context.Background()
		return p.StopServer(ctx)
	}

	return nil
}

// GetStatus returns the current status of the portal
func (p *WebSocketPortal) GetStatus() map[string]interface{} {
	p.mu.RLock()
	connectionCount := len(p.connections)
	functionCount := len(p.functions)
	p.mu.RUnlock()

	status := map[string]interface{}{
		"type":             "websocket",
		"server_running":   p.IsServerRunning(),
		"server_address":   p.GetServerAddress(),
		"connection_count": connectionCount,
		"function_count":   functionCount,
		"config": map[string]interface{}{
			"host":               p.config.Host,
			"port":               p.config.Port,
			"path":               p.config.Path,
			"tls_enabled":        p.config.TLSConfig != nil,
			"ping_interval":      p.config.PingInterval.String(),
			"default_timeout":    p.config.DefaultTimeout.String(),
			"max_message_size":   p.config.MaxMessageSize,
			"enable_compression": p.config.EnableCompression,
		},
	}

	return status
}

// SetConfig updates the portal configuration
func (p *WebSocketPortal) SetConfig(config Config) error {
	// Validate configuration
	if config.Port <= 0 || config.Port > 65535 {
		return NewConfigError("port", fmt.Errorf("invalid port: %d", config.Port))
	}

	if config.Host == "" {
		return NewConfigError("host", fmt.Errorf("host cannot be empty"))
	}

	if config.Path == "" {
		return NewConfigError("path", fmt.Errorf("path cannot be empty"))
	}

	if config.MaxMessageSize <= 0 {
		return NewConfigError("max_message_size", fmt.Errorf("must be positive"))
	}

	// Update configuration
	p.config = config

	return nil
}

// GetConfig returns the current configuration
func (p *WebSocketPortal) GetConfig() Config {
	return p.config
}

// HealthCheck performs a health check on the portal
func (p *WebSocketPortal) HealthCheck(ctx context.Context) error {
	// Check if server is running (if configured as server)
	if p.server != nil && !p.IsServerRunning() {
		return NewServerError(fmt.Errorf("server not running"), nil)
	}

	// Check connection health
	p.mu.RLock()
	unhealthyConnections := 0
	for _, conn := range p.connections {
		select {
		case <-conn.Context.Done():
			unhealthyConnections++
		default:
			// Connection appears healthy
		}
	}
	connectionCount := len(p.connections)
	p.mu.RUnlock()

	if unhealthyConnections > 0 {
		return NewConnectionError("health_check", fmt.Errorf("%d of %d connections unhealthy", unhealthyConnections, connectionCount))
	}

	return nil
}

// Statistics returns detailed portal statistics
func (p *WebSocketPortal) Statistics() map[string]interface{} {
	p.mu.RLock()
	connections := make([]map[string]interface{}, 0, len(p.connections))
	for _, conn := range p.connections {
		conn.mu.RLock()
		connStats := map[string]interface{}{
			"id":            conn.ID,
			"address":       conn.Address,
			"is_client":     conn.IsClient,
			"last_activity": conn.LastActivity,
			"pending_calls": len(conn.PendingCalls),
		}
		conn.mu.RUnlock()
		connections = append(connections, connStats)
	}

	functions := make([]map[string]interface{}, 0, len(p.functions))
	for name, reg := range p.functions {
		functions = append(functions, map[string]interface{}{
			"name":    name,
			"address": reg.Address,
			"schema":  reg.Schema.Metadata().Name,
		})
	}
	p.mu.RUnlock()

	return map[string]interface{}{
		"connections":    connections,
		"functions":      functions,
		"middleware":     len(p.middleware),
		"server_running": p.IsServerRunning(),
	}
}
