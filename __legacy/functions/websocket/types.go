package websocket

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"defs.dev/schema"
)

// WebSocketPortal implements the Portal interface for WebSocket communication
type WebSocketPortal struct {
	config      Config
	server      *http.Server
	upgrader    *websocket.Upgrader
	connections map[string]*WebSocketConnection
	functions   map[string]*FunctionRegistration
	middleware  []Middleware
	mux         *http.ServeMux
	mu          sync.RWMutex
}

// Config holds WebSocket portal configuration
type Config struct {
	Host              string
	Port              int
	Path              string // WebSocket endpoint path (e.g., "/ws")
	TLSConfig         *tls.Config
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	CheckOrigin       func(r *http.Request) bool
	EnableCompression bool
	DefaultTimeout    time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
	PingInterval      time.Duration
	PongTimeout       time.Duration
	MaxMessageSize    int64
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
}

// DefaultConfig returns a default WebSocket configuration
func DefaultConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              8080,
		Path:              "/ws",
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		HandshakeTimeout:  10 * time.Second,
		CheckOrigin:       func(r *http.Request) bool { return true }, // Allow all origins for now
		EnableCompression: true,
		DefaultTimeout:    30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		PingInterval:      30 * time.Second,
		PongTimeout:       60 * time.Second,
		MaxMessageSize:    1024 * 1024, // 1MB
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
	}
}

// FunctionRegistration stores information about a registered function
type FunctionRegistration struct {
	Name    string
	Address string
	Schema  *schema.FunctionSchema
	Handler schema.FunctionHandler
}

// WebSocketConnection represents a WebSocket connection with metadata
type WebSocketConnection struct {
	ID           string
	Conn         *websocket.Conn
	Address      string
	Context      context.Context
	Cancel       context.CancelFunc
	LastActivity time.Time
	IsClient     bool
	PendingCalls map[string]chan *Message // For request/response correlation
	mu           sync.RWMutex
}

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	MessageTypeCall     MessageType = "call"
	MessageTypeResponse MessageType = "response"
	MessageTypeError    MessageType = "error"
	MessageTypeRegister MessageType = "register"
	MessageTypePing     MessageType = "ping"
	MessageTypePong     MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	ID       string         `json:"id"`
	Type     MessageType    `json:"type"`
	Function string         `json:"function,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
	Result   any            `json:"result,omitempty"`
	Error    *ErrorInfo     `json:"error,omitempty"`
	Metadata Metadata       `json:"metadata,omitempty"`
}

// ErrorInfo contains error details in WebSocket messages
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Metadata contains additional message metadata
type Metadata struct {
	Timestamp time.Time              `json:"timestamp"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// WebSocketEndpoint represents a WebSocket endpoint configuration
type WebSocketEndpoint struct {
	URL         string
	Headers     map[string]string
	Subprotocol string
}

// Middleware interface for WebSocket request/response processing
type Middleware interface {
	ProcessMessage(conn *WebSocketConnection, msg *Message) error
	ProcessConnection(conn *WebSocketConnection) error
	ProcessDisconnection(conn *WebSocketConnection) error
}

// Function represents a callable function over WebSocket
type Function interface {
	Call(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error)
	Schema() *schema.FunctionSchema
	Address() string
}

// WebSocketClientFunction represents a client-side function that makes WebSocket calls
type WebSocketClientFunction struct {
	address  string
	schema   *schema.FunctionSchema
	portal   *WebSocketPortal
	endpoint *WebSocketEndpoint
}

// Call implements the Function interface for client functions
func (f *WebSocketClientFunction) Call(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
	return f.portal.CallFunction(ctx, f.address, params)
}

// Schema returns the function schema
func (f *WebSocketClientFunction) Schema() *schema.FunctionSchema {
	return f.schema
}

// Address returns the function address
func (f *WebSocketClientFunction) Address() string {
	return f.address
}

// WebSocketEndpointFunction represents a server-side function endpoint
type WebSocketEndpointFunction struct {
	address string
	schema  *schema.FunctionSchema
	portal  *WebSocketPortal
}

// Call is not implemented for endpoint functions (they represent the server side)
func (f *WebSocketEndpointFunction) Call(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
	return schema.FromAny(nil), &WebSocketFunctionError{
		Code:    400,
		Message: "WebSocketEndpointFunction represents a server endpoint, use a client function to call it",
	}
}

// Schema returns the function schema
func (f *WebSocketEndpointFunction) Schema() *schema.FunctionSchema {
	return f.schema
}

// Address returns the function address
func (f *WebSocketEndpointFunction) Address() string {
	return f.address
}

// ConnectionState represents the state of a WebSocket connection
type ConnectionState int

const (
	ConnectionStateConnecting ConnectionState = iota
	ConnectionStateConnected
	ConnectionStateClosing
	ConnectionStateClosed
	ConnectionStateError
)

// String returns the string representation of ConnectionState
func (s ConnectionState) String() string {
	switch s {
	case ConnectionStateConnecting:
		return "connecting"
	case ConnectionStateConnected:
		return "connected"
	case ConnectionStateClosing:
		return "closing"
	case ConnectionStateClosed:
		return "closed"
	case ConnectionStateError:
		return "error"
	default:
		return "unknown"
	}
}

// ConnectionInfo provides information about a WebSocket connection
type ConnectionInfo struct {
	ID           string
	Address      string
	State        ConnectionState
	ConnectedAt  time.Time
	LastActivity time.Time
	IsClient     bool
	PendingCalls int
}
