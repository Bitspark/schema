package portal

import (
	"context"
	registry2 "defs.dev/schema/runtime/registry"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"defs.dev/schema/api"
	"github.com/gorilla/websocket"
)

// WebSocketPortal implements api.WebSocketPortal
type WebSocketPortal struct {
	// Injected dependencies (shared registries)
	funcRegistry    api.FunctionRegistry
	serviceRegistry api.ServiceRegistry

	// WebSocket-specific fields (transport concerns only)
	config      *WebSocketConfig
	connections map[string]*websocket.Conn
	server      *http.Server
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
}

// WebSocketConfig holds configuration for the WebSocket portal.
type WebSocketConfig struct {
	Host              string
	Port              int
	Path              string
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	CheckOrigin       func(r *http.Request) bool
	Subprotocols      []string
	EnableCompression bool
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	PingPeriod        time.Duration
	PongWait          time.Duration
	MaxMessageSize    int64
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		Host:              "localhost",
		Port:              8081,
		Path:              "/ws",
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		HandshakeTimeout:  10 * time.Second,
		CheckOrigin:       func(r *http.Request) bool { return true },
		Subprotocols:      []string{},
		EnableCompression: false,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
		PingPeriod:        54 * time.Second,
		PongWait:          60 * time.Second,
		MaxMessageSize:    512,
	}
}

// NewWebSocketPortal creates a new WebSocket portal with injected registries
func NewWebSocketPortal(config *WebSocketConfig, funcRegistry api.FunctionRegistry, serviceRegistry api.ServiceRegistry) api.WebSocketPortal {
	if config == nil {
		config = DefaultWebSocketConfig()
	}
	if funcRegistry == nil {
		funcRegistry = registry2.NewFunctionRegistry()
	}
	if serviceRegistry == nil {
		serviceRegistry = registry2.NewServiceRegistry()
	}

	portal := &WebSocketPortal{
		funcRegistry:    funcRegistry,
		serviceRegistry: serviceRegistry,
		config:          config,
		connections:     make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			ReadBufferSize:    config.ReadBufferSize,
			WriteBufferSize:   config.WriteBufferSize,
			HandshakeTimeout:  config.HandshakeTimeout,
			CheckOrigin:       config.CheckOrigin,
			Subprotocols:      config.Subprotocols,
			EnableCompression: config.EnableCompression,
		},
	}

	return portal
}

// Apply registers a function and makes it available via WebSocket
func (p *WebSocketPortal) Apply(ctx context.Context, function api.Function) (api.Address, error) {
	if function == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}

	name := function.Name()
	if name == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}

	// Register with shared function registry (no duplication!)
	err := p.funcRegistry.Register(name, function)
	if err != nil {
		return nil, fmt.Errorf("failed to register function: %w", err)
	}

	// Generate WebSocket address
	address := p.generateAddress(name)
	return address, nil
}

// ApplyService registers a service and makes it available via WebSocket
func (p *WebSocketPortal) ApplyService(ctx context.Context, service api.Service) (api.Address, error) {
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}

	name := service.Schema().Name()
	if name == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	// Register with shared service registry (no duplication!)
	err := p.serviceRegistry.RegisterService(name, service.Schema())
	if err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	// Generate WebSocket service address
	address := p.generateServiceAddress(name)
	return address, nil
}

// ResolveFunction resolves an address to a callable function
func (p *WebSocketPortal) ResolveFunction(ctx context.Context, address api.Address) (api.Function, error) {
	if address.Scheme() != "ws" && address.Scheme() != "wss" {
		return nil, fmt.Errorf("address scheme must be 'ws' or 'wss': %s", address.String())
	}

	// Extract function name from path
	path := address.Path()
	if len(path) > 1 && path[0] == '/' {
		functionName := path[1:] // Remove leading '/'

		// Look up in shared registry
		if function, exists := p.funcRegistry.Get(functionName); exists {
			return function, nil
		}
	}

	return nil, fmt.Errorf("function not found: %s", address.String())
}

// ResolveService resolves an address to a service
func (p *WebSocketPortal) ResolveService(ctx context.Context, address api.Address) (api.Service, error) {
	if address.Scheme() != "ws" && address.Scheme() != "wss" {
		return nil, fmt.Errorf("address scheme must be 'ws' or 'wss': %s", address.String())
	}

	// Extract service name from path
	path := address.Path()
	if len(path) > 9 && path[:9] == "/service/" { // "/service/" prefix
		serviceName := path[9:]

		// Look up in shared registry
		if registeredService, exists := p.serviceRegistry.GetService(serviceName); exists {
			// TODO: Convert RegisteredService back to Service interface
			// For now, return an error indicating incomplete implementation
			_ = registeredService
			return nil, fmt.Errorf("service resolution not fully implemented yet")
		}
	}

	return nil, fmt.Errorf("service not found: %s", address.String())
}

// GenerateAddress creates a new address for the given name and metadata
func (p *WebSocketPortal) GenerateAddress(name string, metadata map[string]any) api.Address {
	return p.generateAddress(name)
}

// Schemes returns the schemes this portal handles
func (p *WebSocketPortal) Schemes() []string {
	return []string{"ws", "wss"}
}

// Close stops the WebSocket server and closes all connections
func (p *WebSocketPortal) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close all active connections
	for id, conn := range p.connections {
		conn.Close()
		delete(p.connections, id)
	}

	// Stop the server if running
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}

	return nil
}

// Health returns the current health status of the portal
func (p *WebSocketPortal) Health(ctx context.Context) error {
	// Check if server is running and responsive
	if p.server != nil {
		// Simple health check - we could extend this
		return nil
	}
	return fmt.Errorf("WebSocket server not running")
}

// Start starts the WebSocket server
func (p *WebSocketPortal) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server != nil {
		return fmt.Errorf("WebSocket server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc(p.config.Path, p.handleWebSocket)
	mux.HandleFunc("/health", p.handleHealth)

	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	p.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error or handle appropriately
		}
	}()

	return nil
}

// Stop stops the WebSocket server
func (p *WebSocketPortal) Stop(ctx context.Context) error {
	return p.Close()
}

// ListenAddress returns the address the portal is listening on
func (p *WebSocketPortal) ListenAddress() string {
	return fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
}

// BaseURL returns the base URL for this portal
func (p *WebSocketPortal) BaseURL() string {
	scheme := "ws"
	if p.config.Port == 443 {
		scheme = "wss"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, p.config.Host, p.config.Port)
}

// GetFunctionRegistry returns the underlying function registry
func (p *WebSocketPortal) GetFunctionRegistry() api.FunctionRegistry {
	return p.funcRegistry
}

// GetServiceRegistry returns the underlying service registry
func (p *WebSocketPortal) GetServiceRegistry() api.ServiceRegistry {
	return p.serviceRegistry
}

// WebSocket message types
type WSMessage struct {
	Type      string         `json:"type"`
	ID        string         `json:"id,omitempty"`
	Function  string         `json:"function,omitempty"`
	Service   string         `json:"service,omitempty"`
	Method    string         `json:"method,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Error     string         `json:"error,omitempty"`
	Timestamp int64          `json:"timestamp,omitempty"`
}

// WebSocket message types
const (
	WSMsgTypeCall     = "call"
	WSMsgTypeResponse = "response"
	WSMsgTypeError    = "error"
	WSMsgTypePing     = "ping"
	WSMsgTypePong     = "pong"
)

// handleWebSocket handles WebSocket connections
func (p *WebSocketPortal) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	// Generate connection ID
	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())

	p.mu.Lock()
	p.connections[connID] = conn
	p.mu.Unlock()

	// Clean up on connection close
	defer func() {
		p.mu.Lock()
		delete(p.connections, connID)
		p.mu.Unlock()
		conn.Close()
	}()

	// Set connection options
	conn.SetReadLimit(p.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(p.config.PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(p.config.PongWait))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(p.config.PingPeriod)
	defer ticker.Stop()

	// Handle messages
	go p.handleConnection(conn, connID)

	// Send pings
	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleConnection handles messages from a WebSocket connection
func (p *WebSocketPortal) handleConnection(conn *websocket.Conn, connID string) {
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}

		response := p.processMessage(msg)
		if response != nil {
			conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))
			if err := conn.WriteJSON(response); err != nil {
				break
			}
		}
	}
}

// processMessage processes a WebSocket message and returns a response
func (p *WebSocketPortal) processMessage(msg WSMessage) *WSMessage {
	switch msg.Type {
	case WSMsgTypeCall:
		return p.handleFunctionCall(msg)
	case WSMsgTypePing:
		return &WSMessage{
			Type:      WSMsgTypePong,
			ID:        msg.ID,
			Timestamp: time.Now().Unix(),
		}
	default:
		return &WSMessage{
			Type:      WSMsgTypeError,
			ID:        msg.ID,
			Error:     fmt.Sprintf("unknown message type: %s", msg.Type),
			Timestamp: time.Now().Unix(),
		}
	}
}

// handleFunctionCall handles function call messages
func (p *WebSocketPortal) handleFunctionCall(msg WSMessage) *WSMessage {
	ctx := context.Background()

	response := &WSMessage{
		Type:      WSMsgTypeResponse,
		ID:        msg.ID,
		Timestamp: time.Now().Unix(),
	}

	// Handle function calls
	if msg.Function != "" {
		// Look up in shared registry
		if function, exists := p.funcRegistry.Get(msg.Function); exists {
			// Call the function
			data := api.NewFunctionData(msg.Data)
			result, err := function.Call(ctx, data)
			if err != nil {
				response.Type = WSMsgTypeError
				response.Error = err.Error()
				return response
			}

			response.Data = map[string]any{
				"result": result.Value(),
			}
			return response
		}
	}

	// Handle service method calls
	if msg.Service != "" && msg.Method != "" {
		// Look up in shared registry
		if _, exists := p.serviceRegistry.GetService(msg.Service); exists {
			// Find method by name - using service methods
			var methodFound bool
			for _, methodName := range p.serviceRegistry.ListServiceMethods(msg.Service) {
				if methodName == msg.Method {
					methodFound = true
					break
				}
			}

			if !methodFound {
				response.Type = WSMsgTypeError
				response.Error = fmt.Sprintf("method not found: %s.%s", msg.Service, msg.Method)
				return response
			}

			// Call service method through registry
			result, err := p.serviceRegistry.CallServiceMethod(context.Background(), msg.Service, msg.Method, msg.Data)
			if err != nil {
				response.Type = WSMsgTypeError
				response.Error = err.Error()
				return response
			}

			response.Type = WSMsgTypeResponse
			response.Data = map[string]any{
				"result": result,
			}
			return response
		}
	}

	response.Type = WSMsgTypeError
	response.Error = "invalid message format"
	return response
}

// handleHealth handles health check requests
func (p *WebSocketPortal) handleHealth(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":      "healthy",
		"timestamp":   time.Now().Unix(),
		"functions":   p.funcRegistry.Count(),
		"services":    p.serviceRegistry.Count(),
		"connections": len(p.connections),
	})
}

// Helper methods

func (p *WebSocketPortal) generateAddress(name string) api.Address {
	scheme := "ws"
	if p.config.Port == 443 {
		scheme = "wss"
	}

	return NewAddressBuilder().
		Scheme(scheme).
		Host(p.config.Host).
		Port(p.config.Port).
		Path(fmt.Sprintf("/%s", name)).
		Build()
}

func (p *WebSocketPortal) generateServiceAddress(name string) api.Address {
	scheme := "ws"
	if p.config.Port == 443 {
		scheme = "wss"
	}

	return NewAddressBuilder().
		Scheme(scheme).
		Host(p.config.Host).
		Port(p.config.Port).
		Path(fmt.Sprintf("/service/%s", name)).
		Build()
}

// WebSocket client functionality

// CallFunction calls a remote function via WebSocket
func (p *WebSocketPortal) CallFunction(ctx context.Context, address api.Address, params api.FunctionData) (api.FunctionData, error) {
	// Parse the address
	u, err := url.Parse(address.String())
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(address.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Extract function name from path
	functionName := u.Path
	if len(functionName) > 1 && functionName[0] == '/' {
		functionName = functionName[1:]
	}

	// Create call message
	msg := WSMessage{
		Type:      WSMsgTypeCall,
		ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
		Function:  functionName,
		Data:      params.ToMap(),
		Timestamp: time.Now().Unix(),
	}

	// Send message
	if err := conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Read response
	var response WSMessage
	if err := conn.ReadJSON(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle response
	if response.Type == WSMsgTypeError {
		return nil, fmt.Errorf("remote error: %s", response.Error)
	}

	if response.Type != WSMsgTypeResponse {
		return nil, fmt.Errorf("unexpected response type: %s", response.Type)
	}

	// Extract result
	if result, ok := response.Data["result"]; ok {
		return api.NewFunctionDataValue(result), nil
	}

	return api.NewFunctionDataValue(response.Data), nil
}

// Stats returns statistics about the WebSocket portal
func (p *WebSocketPortal) Stats() WebSocketPortalStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return WebSocketPortalStats{
		FunctionCount:   p.funcRegistry.Count(),
		ServiceCount:    p.serviceRegistry.Count(),
		ConnectionCount: len(p.connections),
		IsRunning:       p.server != nil,
		Config:          *p.config,
	}
}

// WebSocketPortalStats represents statistics for the WebSocket portal
type WebSocketPortalStats struct {
	FunctionCount   int
	ServiceCount    int
	ConnectionCount int
	IsRunning       bool
	Config          WebSocketConfig
}

// WebSocketPortal interface implementation

// HandleWebSocket provides WebSocket handler integration
func (p *WebSocketPortal) HandleWebSocket() any {
	return http.HandlerFunc(p.handleWebSocket)
}

// Broadcast sends a message to all connected clients
func (p *WebSocketPortal) Broadcast(ctx context.Context, message any) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.connections) == 0 {
		return nil // No connected clients
	}

	// Convert message to JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send to all connections
	var lastErr error
	for connID, conn := range p.connections {
		conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))
		if err := conn.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
			// Remove failed connection
			delete(p.connections, connID)
			conn.Close()
			lastErr = err
		}
	}

	return lastErr
}

// SendToClient sends a message to a specific client
func (p *WebSocketPortal) SendToClient(ctx context.Context, clientID string, message any) error {
	p.mu.RLock()
	conn, exists := p.connections[clientID]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Convert message to JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))
	return conn.WriteMessage(websocket.TextMessage, jsonMessage)
}

// ListClients returns all connected client IDs
func (p *WebSocketPortal) ListClients() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	clients := make([]string, 0, len(p.connections))
	for clientID := range p.connections {
		clients = append(clients, clientID)
	}
	return clients
}
