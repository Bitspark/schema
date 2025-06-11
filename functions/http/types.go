package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"defs.dev/schema"
)

// HTTPEndpoint represents configuration for making HTTP requests to an endpoint
type HTTPEndpoint struct {
	Method  string            // HTTP method (GET, POST, PUT, DELETE)
	Path    string            // URL path (can include templates)
	BaseURL string            // Base server URL
	Headers map[string]string // Static headers
	Query   map[string]string // Static query parameters
	Timeout time.Duration     // Per-request timeout
	Auth    *AuthConfig       // Authentication configuration
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type   string // "bearer", "basic", "api_key"
	Token  string // Bearer token or API key
	Header string // Custom auth header name
	User   string // Username for basic auth
	Pass   string // Password for basic auth
}

// Config represents portal configuration
type Config struct {
	// Server configuration (provider side)
	Host      string
	Port      int
	TLSConfig *tls.Config

	// Client configuration (consumer side)
	DefaultTimeout time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	UserAgent      string

	// Shared configuration
	BasePath string // Base path for all endpoints (e.g., "/api/v1")
}

// HTTPPortal implements the Portal interface for HTTP-based functions
type HTTPPortal struct {
	// Provider side - serves functions as HTTP endpoints
	server *http.Server
	mux    *http.ServeMux

	// Consumer side - makes HTTP requests to endpoints
	client *http.Client

	// Shared configuration and state
	config     Config
	middleware []Middleware
	functions  map[string]*FunctionRegistration // address -> registration
}

// FunctionRegistration stores information about a registered function
type FunctionRegistration struct {
	Name     string
	Address  string
	Schema   *schema.FunctionSchema
	Handler  schema.FunctionHandler // For provider side
	Endpoint *HTTPEndpoint          // For consumer side
}

// Middleware represents HTTP middleware for requests and responses
type Middleware interface {
	ProcessRequest(req *http.Request) error
	ProcessResponse(resp *http.Response) error
}

// MiddlewareFunc is a function adapter for Middleware
type MiddlewareFunc struct {
	RequestFunc  func(*http.Request) error
	ResponseFunc func(*http.Response) error
}

func (m MiddlewareFunc) ProcessRequest(req *http.Request) error {
	if m.RequestFunc != nil {
		return m.RequestFunc(req)
	}
	return nil
}

func (m MiddlewareFunc) ProcessResponse(resp *http.Response) error {
	if m.ResponseFunc != nil {
		return m.ResponseFunc(resp)
	}
	return nil
}

// Function represents a callable function in the portal system
type Function interface {
	Call(ctx context.Context, params map[string]any) (any, error)
	Schema() *schema.FunctionSchema
	Address() string
}
