package websocket

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"

	"defs.dev/schema"
)

// CreateClientFunction creates a client function for making WebSocket calls
func (p *WebSocketPortal) CreateClientFunction(address string, funcSchema *schema.FunctionSchema, endpoint *WebSocketEndpoint) Function {
	return &WebSocketClientFunction{
		address:  address,
		schema:   funcSchema,
		portal:   p,
		endpoint: endpoint,
	}
}

// CallFunction makes a function call over WebSocket
func (p *WebSocketPortal) CallFunction(ctx context.Context, address string, params schema.FunctionInput) (schema.FunctionOutput, error) {
	// Parse address to get connection details
	parsedURL, err := url.Parse(address)
	if err != nil {
		return schema.FromAny(nil), NewAddressError(address, err)
	}

	// Get or create connection
	conn, err := p.getOrCreateConnection(parsedURL)
	if err != nil {
		return schema.FromAny(nil), err
	}

	// Extract function name from address
	functionName := p.extractFunctionFromAddress(address)

	// Generate message ID
	messageID := p.generateMessageID()

	// Create call message
	msg := &Message{
		ID:       messageID,
		Type:     MessageTypeCall,
		Function: functionName,
		Params:   params,
		Metadata: Metadata{
			Timestamp: time.Now(),
		},
	}

	// Create response channel
	respCh := make(chan *Message, 1)

	// Register pending call
	conn.mu.Lock()
	conn.PendingCalls[messageID] = respCh
	conn.mu.Unlock()

	// Clean up pending call when done
	defer func() {
		conn.mu.Lock()
		delete(conn.PendingCalls, messageID)
		conn.mu.Unlock()
		close(respCh)
	}()

	// Send message
	if err := p.sendMessage(conn, msg); err != nil {
		return schema.FromAny(nil), NewNetworkError("send message", err)
	}

	// Wait for response with timeout
	timeout := p.config.DefaultTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	select {
	case response := <-respCh:
		if response.Type == MessageTypeError {
			return schema.FromAny(nil), &WebSocketFunctionError{
				Code:     response.Error.Code,
				Message:  response.Error.Message,
				Address:  address,
				Function: functionName,
				Details:  response.Error.Details,
			}
		}
		return schema.FromAny(response.Result), nil

	case <-ctx.Done():
		return schema.FromAny(nil), NewTimeoutError("function call", "context cancelled")

	case <-time.After(timeout):
		return schema.FromAny(nil), NewTimeoutError("function call", timeout)
	}
}

// getOrCreateConnection gets an existing connection or creates a new one
func (p *WebSocketPortal) getOrCreateConnection(parsedURL *url.URL) (*WebSocketConnection, error) {
	// For WebSocket, we connect to the base endpoint, not the full function path
	// Extract the base WebSocket endpoint (scheme://host:port/ws)
	baseURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   "/ws", // Use the standard WebSocket path
	}

	// Create connection key using the base endpoint
	connKey := baseURL.String()

	// Check for existing connection
	p.mu.RLock()
	if conn, exists := p.connections[connKey]; exists {
		p.mu.RUnlock()
		return conn, nil
	}
	p.mu.RUnlock()

	// Create new connection to the base endpoint
	return p.createClientConnection(baseURL)
}

// createClientConnection creates a new client WebSocket connection
func (p *WebSocketPortal) createClientConnection(parsedURL *url.URL) (*WebSocketConnection, error) {
	// Create WebSocket dialer
	dialer := &websocket.Dialer{
		HandshakeTimeout:  p.config.HandshakeTimeout,
		ReadBufferSize:    p.config.ReadBufferSize,
		WriteBufferSize:   p.config.WriteBufferSize,
		EnableCompression: p.config.EnableCompression,
		TLSClientConfig:   p.config.TLSConfig,
	}

	// Prepare headers
	headers := make(map[string][]string)

	// Apply middleware to add headers (if any authentication middleware exists)
	for _, _ = range p.middleware {
		// This is a simplified approach - in a real implementation you might want
		// middleware to modify the dialer or headers directly
	}

	// Connect to WebSocket
	conn, _, err := dialer.Dial(parsedURL.String(), headers)
	if err != nil {
		return nil, NewConnectionError(parsedURL.String(), err)
	}

	// Create connection wrapper
	wsConn := p.createConnection(conn, nil, true)

	// Override connection key for clients
	connKey := parsedURL.String()

	p.mu.Lock()
	delete(p.connections, wsConn.ID) // Remove old ID
	p.connections[connKey] = wsConn  // Use URL as key
	wsConn.ID = connKey              // Update connection ID
	p.mu.Unlock()

	// Apply connection middleware
	for _, middleware := range p.middleware {
		if err := middleware.ProcessConnection(wsConn); err != nil {
			wsConn.Cancel()
			conn.Close()
			return nil, NewMiddlewareError("connection", err)
		}
	}

	// Start connection handler
	go p.handleConnection(wsConn)

	return wsConn, nil
}

// generateMessageID generates a unique message ID
func (p *WebSocketPortal) generateMessageID() string {
	return fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), p.generateRandomID())
}

// ConnectTo explicitly connects to a WebSocket endpoint
func (p *WebSocketPortal) ConnectTo(ctx context.Context, address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return NewAddressError(address, err)
	}

	_, err = p.createClientConnection(parsedURL)
	return err
}

// DisconnectFrom disconnects from a WebSocket endpoint
func (p *WebSocketPortal) DisconnectFrom(address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return NewAddressError(address, err)
	}

	connKey := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	p.mu.Lock()
	conn, exists := p.connections[connKey]
	if exists {
		delete(p.connections, connKey)
	}
	p.mu.Unlock()

	if exists {
		conn.Cancel()
		conn.Conn.Close()
	}

	return nil
}

// IsConnected checks if connected to an endpoint
func (p *WebSocketPortal) IsConnected(address string) bool {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return false
	}

	connKey := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	p.mu.RLock()
	_, exists := p.connections[connKey]
	p.mu.RUnlock()

	return exists
}

// SendMessage sends a raw message to a WebSocket endpoint
func (p *WebSocketPortal) SendMessage(address string, msg *Message) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return NewAddressError(address, err)
	}

	conn, err := p.getOrCreateConnection(parsedURL)
	if err != nil {
		return err
	}

	return p.sendMessage(conn, msg)
}

// SetConnectionHandler sets a handler for connection events
func (p *WebSocketPortal) SetConnectionHandler(handler func(event string, conn *WebSocketConnection)) {
	// This could be implemented to handle connection events
	// For now, it's a placeholder for future extensibility
}

// PingEndpoint sends a ping to an endpoint
func (p *WebSocketPortal) PingEndpoint(ctx context.Context, address string) error {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return NewAddressError(address, err)
	}

	conn, err := p.getOrCreateConnection(parsedURL)
	if err != nil {
		return err
	}

	// Generate message ID
	messageID := p.generateMessageID()

	// Create ping message
	msg := &Message{
		ID:   messageID,
		Type: MessageTypePing,
		Metadata: Metadata{
			Timestamp: time.Now(),
		},
	}

	// Create response channel
	respCh := make(chan *Message, 1)

	// Register pending call
	conn.mu.Lock()
	conn.PendingCalls[messageID] = respCh
	conn.mu.Unlock()

	// Clean up pending call when done
	defer func() {
		conn.mu.Lock()
		delete(conn.PendingCalls, messageID)
		conn.mu.Unlock()
		close(respCh)
	}()

	// Send ping
	if err := p.sendMessage(conn, msg); err != nil {
		return NewNetworkError("send ping", err)
	}

	// Wait for pong
	select {
	case response := <-respCh:
		if response.Type != MessageTypePong {
			return NewProtocolError("pong", string(response.Type))
		}
		return nil

	case <-ctx.Done():
		return NewTimeoutError("ping", "context cancelled")
	}
}

// GetConnectionStats returns connection statistics
func (p *WebSocketPortal) GetConnectionStats(address string) (*ConnectionInfo, error) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return nil, NewAddressError(address, err)
	}

	connKey := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	p.mu.RLock()
	conn, exists := p.connections[connKey]
	p.mu.RUnlock()

	if !exists {
		return nil, NewConnectionError(address, fmt.Errorf("not connected"))
	}

	conn.mu.RLock()
	pendingCalls := len(conn.PendingCalls)
	conn.mu.RUnlock()

	return &ConnectionInfo{
		ID:           conn.ID,
		Address:      conn.Address,
		State:        ConnectionStateConnected,
		ConnectedAt:  time.Now().Add(-time.Since(conn.LastActivity)), // Approximation
		LastActivity: conn.LastActivity,
		IsClient:     conn.IsClient,
		PendingCalls: pendingCalls,
	}, nil
}

// CloseAllConnections closes all active connections
func (p *WebSocketPortal) CloseAllConnections() error {
	p.mu.Lock()
	connections := make([]*WebSocketConnection, 0, len(p.connections))
	for _, conn := range p.connections {
		connections = append(connections, conn)
	}
	p.connections = make(map[string]*WebSocketConnection)
	p.mu.Unlock()

	// Close all connections
	for _, conn := range connections {
		conn.Cancel()
		conn.Conn.Close()
	}

	return nil
}

// SetRetryConfig configures retry behavior for client connections
func (p *WebSocketPortal) SetRetryConfig(maxRetries int, retryDelay time.Duration) {
	p.config.MaxRetries = maxRetries
	p.config.RetryDelay = retryDelay
}

// callWithRetry makes a function call with retry logic
func (p *WebSocketPortal) callWithRetry(ctx context.Context, address string, params schema.FunctionInput) (schema.FunctionOutput, error) {
	var lastErr error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return schema.FromAny(nil), ctx.Err()
			case <-time.After(p.config.RetryDelay):
			}
		}

		result, err := p.CallFunction(ctx, address, params)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if wsErr, ok := err.(*WebSocketFunctionError); ok {
			// Don't retry client errors (4xx)
			if wsErr.Code >= 400 && wsErr.Code < 500 {
				break
			}
		}

		if wsErr, ok := err.(*WebSocketPortalError); ok {
			// Don't retry certain portal errors
			if wsErr.Type == "ValidationError" || wsErr.Type == "ProtocolError" {
				break
			}
		}
	}

	return schema.FromAny(nil), lastErr
}
