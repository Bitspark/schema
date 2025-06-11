package websocket

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// LoggingMiddleware logs WebSocket messages and connection events
type LoggingMiddleware struct {
	Logger func(format string, args ...interface{})
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger: log.Printf,
	}
}

// NewCustomLoggingMiddleware creates logging middleware with a custom logger
func NewCustomLoggingMiddleware(logger func(format string, args ...interface{})) *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger: logger,
	}
}

func (m *LoggingMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	direction := "incoming"
	if conn.IsClient {
		direction = "outgoing"
	}

	m.Logger("[WebSocket] %s message: %s/%s from %s (ID: %s)",
		direction, msg.Type, msg.Function, conn.Address, msg.ID)
	return nil
}

func (m *LoggingMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	connType := "server"
	if conn.IsClient {
		connType = "client"
	}

	m.Logger("[WebSocket] %s connection established: %s (ID: %s)",
		connType, conn.Address, conn.ID)
	return nil
}

func (m *LoggingMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	connType := "server"
	if conn.IsClient {
		connType = "client"
	}

	duration := time.Since(conn.LastActivity)
	m.Logger("[WebSocket] %s connection closed: %s (ID: %s, duration: %v)",
		connType, conn.Address, conn.ID, duration)
	return nil
}

// AuthenticationMiddleware handles various authentication schemes for WebSocket
type AuthenticationMiddleware struct {
	Type      string                   // "bearer", "api_key", "custom"
	Secret    string                   // Secret key or token
	Header    string                   // Custom header name for API key auth
	Validator func(msg *Message) error // Custom validation function
}

// NewBearerAuthMiddleware creates bearer token authentication middleware
func NewBearerAuthMiddleware(token string) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		Type:   "bearer",
		Secret: token,
	}
}

// NewAPIKeyAuthMiddleware creates API key authentication middleware
func NewAPIKeyAuthMiddleware(apiKey string, header ...string) *AuthenticationMiddleware {
	headerName := "X-API-Key"
	if len(header) > 0 {
		headerName = header[0]
	}
	return &AuthenticationMiddleware{
		Type:   "api_key",
		Secret: apiKey,
		Header: headerName,
	}
}

// NewCustomAuthMiddleware creates custom authentication middleware
func NewCustomAuthMiddleware(validator func(msg *Message) error) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		Type:      "custom",
		Validator: validator,
	}
}

func (m *AuthenticationMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	switch m.Type {
	case "bearer":
		// Check for bearer token in message metadata
		if msg.Metadata.Headers != nil {
			if auth, exists := msg.Metadata.Headers["Authorization"]; exists {
				expected := "Bearer " + m.Secret
				if auth != expected {
					return NewMiddlewareError("authentication", fmt.Errorf("invalid bearer token"))
				}
			} else {
				return NewMiddlewareError("authentication", fmt.Errorf("missing authorization header"))
			}
		}
	case "api_key":
		// Check for API key in message metadata
		if msg.Metadata.Headers != nil {
			if key, exists := msg.Metadata.Headers[m.Header]; exists {
				if key != m.Secret {
					return NewMiddlewareError("authentication", fmt.Errorf("invalid API key"))
				}
			} else {
				return NewMiddlewareError("authentication", fmt.Errorf("missing API key header"))
			}
		}
	case "custom":
		// Use custom validator
		if m.Validator != nil {
			return m.Validator(msg)
		}
	}
	return nil
}

func (m *AuthenticationMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	// Connection-level authentication could be implemented here
	return nil
}

func (m *AuthenticationMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// Clean up authentication state if needed
	return nil
}

// MetricsMiddleware collects WebSocket metrics
type MetricsMiddleware struct {
	ConnectionCount    int64
	DisconnectionCount int64
	MessageCount       int64
	ErrorCount         int64
	TotalLatency       int64 // nanoseconds
	ActiveConnections  int64
	messageTimes       map[string]time.Time
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		messageTimes: make(map[string]time.Time),
	}
}

func (m *MetricsMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	atomic.AddInt64(&m.MessageCount, 1)

	// Track message timing for responses
	if msg.Type == MessageTypeCall {
		m.messageTimes[msg.ID] = time.Now()
	} else if msg.Type == MessageTypeResponse || msg.Type == MessageTypeError {
		if startTime, exists := m.messageTimes[msg.ID]; exists {
			latency := time.Since(startTime).Nanoseconds()
			atomic.AddInt64(&m.TotalLatency, latency)
			delete(m.messageTimes, msg.ID)
		}
	}

	// Track errors
	if msg.Type == MessageTypeError {
		atomic.AddInt64(&m.ErrorCount, 1)
	}

	return nil
}

func (m *MetricsMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	atomic.AddInt64(&m.ConnectionCount, 1)
	atomic.AddInt64(&m.ActiveConnections, 1)
	return nil
}

func (m *MetricsMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	atomic.AddInt64(&m.DisconnectionCount, 1)
	atomic.AddInt64(&m.ActiveConnections, -1)
	return nil
}

// GetMetrics returns current metrics
func (m *MetricsMiddleware) GetMetrics() map[string]interface{} {
	messageCount := atomic.LoadInt64(&m.MessageCount)
	totalLatency := atomic.LoadInt64(&m.TotalLatency)
	avgLatency := time.Duration(0)

	if messageCount > 0 {
		avgLatency = time.Duration(totalLatency / messageCount)
	}

	return map[string]interface{}{
		"connection_count":    atomic.LoadInt64(&m.ConnectionCount),
		"disconnection_count": atomic.LoadInt64(&m.DisconnectionCount),
		"message_count":       messageCount,
		"error_count":         atomic.LoadInt64(&m.ErrorCount),
		"active_connections":  atomic.LoadInt64(&m.ActiveConnections),
		"avg_latency":         avgLatency.String(),
		"total_latency":       time.Duration(totalLatency).String(),
	}
}

// ValidationMiddleware validates WebSocket messages
type ValidationMiddleware struct {
	ValidateMessage func(msg *Message) error
	MaxMessageSize  int64
	AllowedTypes    []MessageType
}

// NewValidationMiddleware creates validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		MaxMessageSize: 1024 * 1024, // 1MB default
		AllowedTypes: []MessageType{
			MessageTypeCall,
			MessageTypeResponse,
			MessageTypeError,
			MessageTypePing,
			MessageTypePong,
		},
	}
}

func (m *ValidationMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	// Validate message type
	if len(m.AllowedTypes) > 0 {
		typeAllowed := false
		for _, allowedType := range m.AllowedTypes {
			if msg.Type == allowedType {
				typeAllowed = true
				break
			}
		}
		if !typeAllowed {
			return NewValidationError("message_type", string(msg.Type), fmt.Errorf("message type not allowed"))
		}
	}

	// Validate message ID
	if msg.ID == "" {
		return NewValidationError("message_id", "", fmt.Errorf("message ID is required"))
	}

	// Custom validation
	if m.ValidateMessage != nil {
		return m.ValidateMessage(msg)
	}

	return nil
}

func (m *ValidationMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	// Connection validation could be implemented here
	return nil
}

func (m *ValidationMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// No validation needed for disconnection
	return nil
}

// RateLimitMiddleware implements rate limiting for WebSocket messages
type RateLimitMiddleware struct {
	MaxMessagesPerSecond int
	connectionTokens     map[string]*tokenBucket
}

type tokenBucket struct {
	tokens     int
	maxTokens  int
	lastRefill time.Time
}

// NewRateLimitMiddleware creates rate limiting middleware
func NewRateLimitMiddleware(maxMessagesPerSecond int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		MaxMessagesPerSecond: maxMessagesPerSecond,
		connectionTokens:     make(map[string]*tokenBucket),
	}
}

func (m *RateLimitMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	// Get or create token bucket for this connection
	bucket, exists := m.connectionTokens[conn.ID]
	if !exists {
		bucket = &tokenBucket{
			tokens:     m.MaxMessagesPerSecond,
			maxTokens:  m.MaxMessagesPerSecond,
			lastRefill: time.Now(),
		}
		m.connectionTokens[conn.ID] = bucket
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * m.MaxMessagesPerSecond
	bucket.tokens = min(bucket.maxTokens, bucket.tokens+tokensToAdd)
	bucket.lastRefill = now

	// Check if we have tokens
	if bucket.tokens <= 0 {
		return NewMiddlewareError("rate_limit", fmt.Errorf("rate limit exceeded"))
	}

	// Consume a token
	bucket.tokens--

	return nil
}

func (m *RateLimitMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	// Initialize token bucket for new connection
	m.connectionTokens[conn.ID] = &tokenBucket{
		tokens:     m.MaxMessagesPerSecond,
		maxTokens:  m.MaxMessagesPerSecond,
		lastRefill: time.Now(),
	}
	return nil
}

func (m *RateLimitMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// Clean up token bucket
	delete(m.connectionTokens, conn.ID)
	return nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CompressionMiddleware handles message compression (placeholder)
type CompressionMiddleware struct {
	Enabled bool
	Level   int // Compression level
}

// NewCompressionMiddleware creates compression middleware
func NewCompressionMiddleware(enabled bool, level int) *CompressionMiddleware {
	return &CompressionMiddleware{
		Enabled: enabled,
		Level:   level,
	}
}

func (m *CompressionMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	// Compression would be handled at the WebSocket frame level
	// This is mainly for configuration and metrics
	return nil
}

func (m *CompressionMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	// Set compression parameters if needed
	return nil
}

func (m *CompressionMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// Nothing to do for compression cleanup
	return nil
}

// CircuitBreakerMiddleware implements circuit breaker pattern
type CircuitBreakerMiddleware struct {
	MaxFailures     int
	ResetTimeout    time.Duration
	currentState    string // "closed", "open", "half-open"
	failureCount    int
	lastFailureTime time.Time
}

// NewCircuitBreakerMiddleware creates circuit breaker middleware
func NewCircuitBreakerMiddleware(maxFailures int, resetTimeout time.Duration) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		MaxFailures:  maxFailures,
		ResetTimeout: resetTimeout,
		currentState: "closed",
	}
}

func (m *CircuitBreakerMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	// Check circuit breaker state
	if m.currentState == "open" {
		// Check if we should transition to half-open
		if time.Since(m.lastFailureTime) > m.ResetTimeout {
			m.currentState = "half-open"
			m.failureCount = 0
		} else {
			return NewMiddlewareError("circuit_breaker", fmt.Errorf("circuit breaker is open"))
		}
	}

	// Track errors to update circuit breaker state
	if msg.Type == MessageTypeError {
		m.failureCount++
		m.lastFailureTime = time.Now()

		if m.failureCount >= m.MaxFailures {
			m.currentState = "open"
		}
	} else if msg.Type == MessageTypeResponse {
		// Success - reset failure count if in half-open state
		if m.currentState == "half-open" {
			m.currentState = "closed"
			m.failureCount = 0
		}
	}

	return nil
}

func (m *CircuitBreakerMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	// Circuit breaker is per-middleware, not per-connection
	return nil
}

func (m *CircuitBreakerMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// Nothing to do for circuit breaker cleanup
	return nil
}

// GetState returns the current circuit breaker state
func (m *CircuitBreakerMiddleware) GetState() string {
	return m.currentState
}

// ChainMiddleware creates a middleware that chains multiple middleware together
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return &chainedMiddleware{middlewares: middlewares}
}

type chainedMiddleware struct {
	middlewares []Middleware
}

func (c *chainedMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	for _, middleware := range c.middlewares {
		if err := middleware.ProcessMessage(conn, msg); err != nil {
			return err
		}
	}
	return nil
}

func (c *chainedMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	for _, middleware := range c.middlewares {
		if err := middleware.ProcessConnection(conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *chainedMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// Process in reverse order for cleanup
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		middleware := c.middlewares[i]
		middleware.ProcessDisconnection(conn) // Ignore errors in cleanup
	}
	return nil
}

// ConditionalMiddleware creates middleware that only applies under certain conditions
func ConditionalMiddleware(condition func(conn *WebSocketConnection, msg *Message) bool, middleware Middleware) Middleware {
	return &conditionalMiddleware{
		condition:  condition,
		middleware: middleware,
	}
}

type conditionalMiddleware struct {
	condition  func(conn *WebSocketConnection, msg *Message) bool
	middleware Middleware
}

func (c *conditionalMiddleware) ProcessMessage(conn *WebSocketConnection, msg *Message) error {
	if c.condition(conn, msg) {
		return c.middleware.ProcessMessage(conn, msg)
	}
	return nil
}

func (c *conditionalMiddleware) ProcessConnection(conn *WebSocketConnection) error {
	// For connection events, we can't check message condition, so we apply the middleware
	return c.middleware.ProcessConnection(conn)
}

func (c *conditionalMiddleware) ProcessDisconnection(conn *WebSocketConnection) error {
	// For disconnection events, we apply the middleware
	return c.middleware.ProcessDisconnection(conn)
}
