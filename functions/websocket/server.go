package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"defs.dev/schema"
)

// StartServer starts the WebSocket server
func (p *WebSocketPortal) StartServer() error {
	if p.server != nil {
		return NewServerError(fmt.Errorf("server already running"), nil)
	}

	// Create HTTP server
	p.mux = http.NewServeMux()

	// Setup WebSocket upgrader
	p.upgrader = &websocket.Upgrader{
		ReadBufferSize:    p.config.ReadBufferSize,
		WriteBufferSize:   p.config.WriteBufferSize,
		HandshakeTimeout:  p.config.HandshakeTimeout,
		CheckOrigin:       p.config.CheckOrigin,
		EnableCompression: p.config.EnableCompression,
	}

	// Register WebSocket endpoint
	p.mux.HandleFunc(p.config.Path, p.handleWebSocket)

	// Create server
	address := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	p.server = &http.Server{
		Addr:         address,
		Handler:      p.mux,
		TLSConfig:    p.config.TLSConfig,
		ReadTimeout:  p.config.ReadTimeout,
		WriteTimeout: p.config.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		var err error
		if p.config.TLSConfig != nil {
			err = p.server.ListenAndServeTLS("", "")
		} else {
			err = p.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("WebSocket server error: %v\n", err)
		}
	}()

	// Wait for server to be ready
	return p.waitForServer()
}

// StopServer stops the WebSocket server
func (p *WebSocketPortal) StopServer(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	// Close all connections
	p.mu.Lock()
	for _, conn := range p.connections {
		conn.Cancel()
		conn.Conn.Close()
	}
	p.connections = make(map[string]*WebSocketConnection)
	p.mu.Unlock()

	// Shutdown server
	err := p.server.Shutdown(ctx)
	p.server = nil
	return err
}

// waitForServer waits for the server to be ready
func (p *WebSocketPortal) waitForServer() error {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		if p.IsServerRunning() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return NewServerError(fmt.Errorf("server failed to start after %d attempts", maxAttempts), nil)
}

// IsServerRunning checks if the server is currently running
func (p *WebSocketPortal) IsServerRunning() bool {
	if p.server == nil {
		return false
	}

	// Try to connect to the server
	address := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// handleWebSocket handles WebSocket connections
func (p *WebSocketPortal) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}

	// Create connection wrapper
	wsConn := p.createConnection(conn, r, false)

	// Apply connection middleware
	for _, middleware := range p.middleware {
		if err := middleware.ProcessConnection(wsConn); err != nil {
			p.closeConnection(wsConn, fmt.Sprintf("Middleware error: %v", err))
			return
		}
	}

	// Start connection handler
	go p.handleConnection(wsConn)
}

// createConnection creates a new WebSocket connection wrapper
func (p *WebSocketPortal) createConnection(conn *websocket.Conn, r *http.Request, isClient bool) *WebSocketConnection {
	ctx, cancel := context.WithCancel(context.Background())

	connID := p.generateConnectionID()
	wsConn := &WebSocketConnection{
		ID:           connID,
		Conn:         conn,
		Address:      conn.RemoteAddr().String(),
		Context:      ctx,
		Cancel:       cancel,
		LastActivity: time.Now(),
		IsClient:     isClient,
		PendingCalls: make(map[string]chan *Message),
	}

	// Configure connection
	conn.SetReadLimit(p.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(p.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))

	// Set up ping/pong handlers
	conn.SetPingHandler(func(data string) error {
		wsConn.LastActivity = time.Now()
		conn.SetReadDeadline(time.Now().Add(p.config.ReadTimeout))
		return conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(p.config.WriteTimeout))
	})

	conn.SetPongHandler(func(data string) error {
		wsConn.LastActivity = time.Now()
		conn.SetReadDeadline(time.Now().Add(p.config.ReadTimeout))
		return nil
	})

	// Store connection
	p.mu.Lock()
	p.connections[connID] = wsConn
	p.mu.Unlock()

	return wsConn
}

// handleConnection handles a WebSocket connection
func (p *WebSocketPortal) handleConnection(conn *WebSocketConnection) {
	defer p.closeConnection(conn, "Connection closed")

	// Start ping ticker if configured
	var pingTicker *time.Ticker
	if p.config.PingInterval > 0 {
		pingTicker = time.NewTicker(p.config.PingInterval)
		defer pingTicker.Stop()

		go func() {
			for {
				select {
				case <-pingTicker.C:
					if err := p.sendPing(conn); err != nil {
						return
					}
				case <-conn.Context.Done():
					return
				}
			}
		}()
	}

	// Message handling loop
	for {
		select {
		case <-conn.Context.Done():
			return
		default:
			// Read message
			messageType, data, err := conn.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket error: %v\n", err)
				}
				return
			}

			// Update activity
			conn.LastActivity = time.Now()
			conn.Conn.SetReadDeadline(time.Now().Add(p.config.ReadTimeout))

			// Only handle text messages
			if messageType != websocket.TextMessage {
				continue
			}

			// Parse message
			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				p.sendError(conn, "", NewMessageError("", err))
				continue
			}

			// Process message
			go p.processMessage(conn, &msg)
		}
	}
}

// processMessage processes a received WebSocket message
func (p *WebSocketPortal) processMessage(conn *WebSocketConnection, msg *Message) {
	fmt.Printf("[DEBUG] Processing message: ID=%s, Type=%s, Function=%s\n", msg.ID, msg.Type, msg.Function)

	// Apply message middleware
	for _, middleware := range p.middleware {
		if err := middleware.ProcessMessage(conn, msg); err != nil {
			p.sendError(conn, msg.ID, NewMiddlewareError("message_processing", err))
			return
		}
	}

	// Handle different message types
	switch msg.Type {
	case MessageTypeCall:
		fmt.Printf("[DEBUG] Routing to handleFunctionCall\n")
		p.handleFunctionCall(conn, msg)
	case MessageTypeResponse:
		fmt.Printf("[DEBUG] Routing to handleResponse\n")
		p.handleResponse(conn, msg)
	case MessageTypeError:
		fmt.Printf("[DEBUG] Routing to handleError\n")
		p.handleError(conn, msg)
	case MessageTypeRegister:
		fmt.Printf("[DEBUG] Routing to handleFunctionRegister\n")
		p.handleFunctionRegister(conn, msg)
	case MessageTypePing:
		fmt.Printf("[DEBUG] Routing to handlePing\n")
		p.handlePing(conn, msg)
	case MessageTypePong:
		fmt.Printf("[DEBUG] Routing to handlePong\n")
		p.handlePong(conn, msg)
	default:
		fmt.Printf("[DEBUG] Unknown message type: %s\n", msg.Type)
		p.sendError(conn, msg.ID, NewProtocolError("known message type", string(msg.Type)))
	}
}

// handleFunctionCall handles function call messages
func (p *WebSocketPortal) handleFunctionCall(conn *WebSocketConnection, msg *Message) {
	// Find registered function
	p.mu.RLock()
	registration, exists := p.functions[msg.Function]
	p.mu.RUnlock()

	if !exists {
		p.sendError(conn, msg.ID, &WebSocketFunctionError{
			Code:     404,
			Message:  "Function not found",
			Function: msg.Function,
		})
		return
	}

	// Call function
	ctx := conn.Context
	if p.config.DefaultTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.DefaultTimeout)
		defer cancel()
	}

	output, err := registration.Handler(ctx, schema.FromMap(msg.Params))
	if err != nil {
		// Check if it's a WebSocketFunctionError
		if wsErr, ok := err.(*WebSocketFunctionError); ok {
			p.sendError(conn, msg.ID, wsErr)
		} else {
			p.sendError(conn, msg.ID, NewCallError(msg.Function, registration.Address, err))
		}
		return
	}

	// Send response
	response := &Message{
		ID:     msg.ID,
		Type:   MessageTypeResponse,
		Result: output.Value(),
		Metadata: Metadata{
			Timestamp: time.Now(),
		},
	}

	p.sendMessage(conn, response)
}

// handleResponse handles response messages (for client-side)
func (p *WebSocketPortal) handleResponse(conn *WebSocketConnection, msg *Message) {
	fmt.Printf("[DEBUG] Handling response message: %s, type: %s\n", msg.ID, msg.Type)

	conn.mu.RLock()
	ch, exists := conn.PendingCalls[msg.ID]
	conn.mu.RUnlock()

	if exists {
		fmt.Printf("[DEBUG] Found pending call for message: %s, sending to channel\n", msg.ID)
		select {
		case ch <- msg:
			fmt.Printf("[DEBUG] Message sent to channel successfully: %s\n", msg.ID)
		case <-time.After(time.Second):
			fmt.Printf("[DEBUG] Channel blocked for message: %s\n", msg.ID)
		}
	} else {
		fmt.Printf("[DEBUG] No pending call found for message: %s\n", msg.ID)
	}
}

// handleError handles error messages
func (p *WebSocketPortal) handleError(conn *WebSocketConnection, msg *Message) {
	conn.mu.RLock()
	ch, exists := conn.PendingCalls[msg.ID]
	conn.mu.RUnlock()

	if exists {
		select {
		case ch <- msg:
		case <-time.After(time.Second):
			// Channel is blocked, ignore
		}
	}
}

// handleFunctionRegister handles function registration messages
func (p *WebSocketPortal) handleFunctionRegister(conn *WebSocketConnection, msg *Message) {
	// Function registration over WebSocket could be implemented here
	// For now, we'll send a not implemented error
	p.sendError(conn, msg.ID, &WebSocketFunctionError{
		Code:    501,
		Message: "Function registration over WebSocket not implemented",
	})
}

// handlePing handles ping messages
func (p *WebSocketPortal) handlePing(conn *WebSocketConnection, msg *Message) {
	fmt.Printf("[DEBUG] Handling ping message: %s\n", msg.ID)

	response := &Message{
		ID:   msg.ID,
		Type: MessageTypePong,
		Metadata: Metadata{
			Timestamp: time.Now(),
		},
	}

	err := p.sendMessage(conn, response)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to send pong: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Pong sent successfully for ping: %s\n", msg.ID)
	}
}

// handlePong handles pong messages
func (p *WebSocketPortal) handlePong(conn *WebSocketConnection, msg *Message) {
	// Forward pong to any pending ping calls
	conn.mu.RLock()
	ch, exists := conn.PendingCalls[msg.ID]
	conn.mu.RUnlock()

	if exists {
		select {
		case ch <- msg:
		case <-time.After(time.Second):
			// Channel is blocked, ignore
		}
	}
}

// sendMessage sends a message over WebSocket
func (p *WebSocketPortal) sendMessage(conn *WebSocketConnection, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn.Conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))
	return conn.Conn.WriteMessage(websocket.TextMessage, data)
}

// sendError sends an error message
func (p *WebSocketPortal) sendError(conn *WebSocketConnection, messageID string, err error) {
	var errorInfo *ErrorInfo

	if wsErr, ok := err.(*WebSocketFunctionError); ok {
		errorInfo = &ErrorInfo{
			Code:    wsErr.Code,
			Message: wsErr.Message,
			Details: wsErr.Details,
		}
	} else if wsErr, ok := err.(*WebSocketPortalError); ok {
		errorInfo = &ErrorInfo{
			Code:    wsErr.Code,
			Message: wsErr.Message,
			Details: wsErr.Details,
		}
	} else {
		errorInfo = &ErrorInfo{
			Code:    500,
			Message: "Internal error",
			Details: err.Error(),
		}
	}

	errorMsg := &Message{
		ID:    messageID,
		Type:  MessageTypeError,
		Error: errorInfo,
		Metadata: Metadata{
			Timestamp: time.Now(),
		},
	}

	p.sendMessage(conn, errorMsg)
}

// sendPing sends a ping message
func (p *WebSocketPortal) sendPing(conn *WebSocketConnection) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(p.config.WriteTimeout))
	return conn.Conn.WriteMessage(websocket.PingMessage, []byte{})
}

// closeConnection closes a WebSocket connection
func (p *WebSocketPortal) closeConnection(conn *WebSocketConnection, reason string) {
	// Apply disconnection middleware
	for _, middleware := range p.middleware {
		middleware.ProcessDisconnection(conn)
	}

	// Cancel context
	conn.Cancel()

	// Close WebSocket connection
	conn.Conn.Close()

	// Remove from connections map
	p.mu.Lock()
	delete(p.connections, conn.ID)
	p.mu.Unlock()

	fmt.Printf("WebSocket connection %s closed: %s\n", conn.ID, reason)
}

// generateConnectionID generates a unique connection ID
func (p *WebSocketPortal) generateConnectionID() string {
	return fmt.Sprintf("ws_%d_%s", time.Now().UnixNano(), p.generateRandomID())
}

// generateRandomID generates a random ID
func (p *WebSocketPortal) generateRandomID() string {
	// Simple random ID generation
	return fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFFFFFF)
}

// RegisterHandler registers a function handler for WebSocket calls
func (p *WebSocketPortal) RegisterHandler(name string, address string, funcSchema *schema.FunctionSchema, handler schema.FunctionHandler) error {
	// Extract function name from address
	functionName := p.extractFunctionFromAddress(address)

	// Store registration
	p.mu.Lock()
	p.functions[functionName] = &FunctionRegistration{
		Name:    name,
		Address: address,
		Schema:  funcSchema,
		Handler: handler,
	}
	p.mu.Unlock()

	return nil
}

// extractFunctionFromAddress extracts the function name from an address
func (p *WebSocketPortal) extractFunctionFromAddress(address string) string {
	// For WebSocket addresses like "ws://host:port/ws/functionName/id"
	// Extract the function name

	// Remove protocol
	if strings.HasPrefix(address, "ws://") {
		address = address[5:]
	} else if strings.HasPrefix(address, "wss://") {
		address = address[6:]
	}

	// Find path component
	parts := strings.SplitN(address, "/", 2)
	if len(parts) < 2 {
		return "unknown"
	}

	path := parts[1]

	// Remove the WebSocket path prefix
	wsPath := strings.TrimLeft(p.config.Path, "/")
	if strings.HasPrefix(path, wsPath+"/") {
		path = path[len(wsPath)+1:]
	}

	// Extract function name (first path component after ws path)
	pathParts := strings.Split(path, "/")
	if len(pathParts) > 0 {
		return pathParts[0]
	}

	return "unknown"
}

// GetServerAddress returns the full server address
func (p *WebSocketPortal) GetServerAddress() string {
	scheme := "ws"
	if p.config.TLSConfig != nil {
		scheme = "wss"
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, p.config.Host, p.config.Port, p.config.Path)
}

// GetConnections returns information about current connections
func (p *WebSocketPortal) GetConnections() []ConnectionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	connections := make([]ConnectionInfo, 0, len(p.connections))
	for _, conn := range p.connections {
		conn.mu.RLock()
		pendingCalls := len(conn.PendingCalls)
		conn.mu.RUnlock()

		connections = append(connections, ConnectionInfo{
			ID:           conn.ID,
			Address:      conn.Address,
			State:        ConnectionStateConnected,                       // Simplified for now
			ConnectedAt:  time.Now().Add(-time.Since(conn.LastActivity)), // Approximation
			LastActivity: conn.LastActivity,
			IsClient:     conn.IsClient,
			PendingCalls: pendingCalls,
		})
	}

	return connections
}
