package portal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
)

// HTTPPortal implements the api.HTTPPortal interface for HTTP-based function execution.
type HTTPPortal struct {
	mu sync.RWMutex

	// Configuration
	config *HTTPConfig

	// Server components
	server     *http.Server
	mux        *http.ServeMux
	middleware []Middleware

	// Function registry
	functions map[string]api.Function
	schemas   map[string]core.FunctionSchema

	// Client components
	client *http.Client

	// State
	running bool
	baseURL string
}

// HTTPConfig holds configuration for the HTTP portal.
type HTTPConfig struct {
	// Server configuration
	Host         string
	Port         int
	TLS          *TLSConfig
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Client configuration
	ClientTimeout time.Duration
	MaxRetries    int
	RetryDelay    time.Duration

	// CORS configuration
	CORSOrigins []string
	CORSMethods []string
	CORSHeaders []string

	// Security
	RequireAuth bool
	AuthMethods []string

	// Limits
	MaxRequestSize int64
	RateLimit      *RateLimitConfig
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	CertFile string
	KeyFile  string
	Insecure bool
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	PerClientLimit    int
}

// Middleware defines HTTP middleware interface.
type Middleware interface {
	Handle(next http.Handler) http.Handler
}

// NewHTTPPortal creates a new HTTP portal with the given configuration.
func NewHTTPPortal(config *HTTPConfig) *HTTPPortal {
	if config == nil {
		config = DefaultHTTPConfig()
	}

	mux := http.NewServeMux()

	portal := &HTTPPortal{
		config:    config,
		mux:       mux,
		functions: make(map[string]api.Function),
		schemas:   make(map[string]core.FunctionSchema),
		client: &http.Client{
			Timeout: config.ClientTimeout,
		},
	}

	// Set up server
	portal.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      portal.buildHandler(),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	// Set base URL
	scheme := "http"
	if config.TLS != nil {
		scheme = "https"
	}
	portal.baseURL = fmt.Sprintf("%s://%s:%d", scheme, config.Host, config.Port)

	return portal
}

// DefaultHTTPConfig returns default HTTP configuration.
func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Host:           "localhost",
		Port:           8080,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		ClientTimeout:  30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     time.Second,
		CORSOrigins:    []string{"*"},
		CORSMethods:    []string{"GET", "POST", "OPTIONS"},
		CORSHeaders:    []string{"Content-Type", "Authorization"},
		MaxRequestSize: 1024 * 1024, // 1MB
		RequireAuth:    false,
	}
}

// Apply registers a function with the HTTP portal.
func (h *HTTPPortal) Apply(ctx context.Context, function api.Function) (api.Address, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	name := function.Name()
	if name == "" {
		return nil, fmt.Errorf("function name is required")
	}

	if _, exists := h.functions[name]; exists {
		return nil, fmt.Errorf("function %s already registered", name)
	}

	// Register function directly
	h.functions[name] = function
	h.schemas[name] = function.Schema()

	// Register HTTP endpoint
	path := "/functions/" + name
	h.mux.HandleFunc(path, h.handleFunctionCall)

	// Generate address
	address := h.GenerateAddress(name, map[string]any{
		"path": path,
		"type": "function",
	})

	return address, nil
}

// ApplyService registers a service with the HTTP portal.
func (h *HTTPPortal) ApplyService(ctx context.Context, service api.Service) (api.Address, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	name := service.Schema().Name()

	// Register service methods as individual functions
	methods := service.Schema().Methods()
	for _, method := range methods {
		methodName := method.Name()
		functionName := name + "." + methodName

		// TODO: ServiceMethodSchema doesn't implement api.Function directly
		// We need to create a wrapper function that handles service method calls
		// For now, just store the schema
		// h.functions[functionName] = method
		h.schemas[functionName] = method.Function()

		// Register HTTP endpoint
		path := "/services/" + name + "/" + methodName
		h.mux.HandleFunc(path, h.handleFunctionCall)
	}

	// Generate service address
	address := h.GenerateAddress(name, map[string]any{
		"path": "/services/" + name,
		"type": "service",
	})

	return address, nil
}

// ResolveFunction resolves an HTTP address to a function.
func (h *HTTPPortal) ResolveFunction(ctx context.Context, address api.Address) (api.Function, error) {
	if address.Scheme() != "http" && address.Scheme() != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", address.Scheme())
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// For local functions, resolve directly
	if address.Authority() == h.getAuthority() {
		functionName := h.extractFunctionName(address.Path())
		if function, exists := h.functions[functionName]; exists {
			return function, nil
		}
	}

	// For remote functions, create HTTP client function
	return &RemoteFunction{
		name:    h.extractFunctionName(address.Path()),
		schema:  nil, // Would need to fetch schema from remote
		address: address,
		portal:  h,
	}, nil
}

// ResolveService resolves an HTTP address to a service.
func (h *HTTPPortal) ResolveService(ctx context.Context, address api.Address) (api.Service, error) {
	if address.Scheme() != "http" && address.Scheme() != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", address.Scheme())
	}

	// Create HTTP client service
	return &ServiceImpl{
		name:   h.extractServiceName(address.Path()),
		schema: nil, // Would need to fetch schema from remote
	}, nil
}

// GenerateAddress creates a new HTTP address.
func (h *HTTPPortal) GenerateAddress(name string, metadata map[string]any) api.Address {
	builder := NewAddressBuilder()

	scheme := "http"
	if h.config.TLS != nil {
		scheme = "https"
	}

	path := "/functions/" + name
	if metadata != nil {
		if p, ok := metadata["path"].(string); ok {
			path = p
		}
	}

	return builder.
		Scheme(scheme).
		Host(h.config.Host).
		Port(h.config.Port).
		Path(path).
		Build()
}

// Schemes returns the schemes this portal handles.
func (h *HTTPPortal) Schemes() []string {
	schemes := []string{"http"}
	if h.config.TLS != nil {
		schemes = append(schemes, "https")
	}
	return schemes
}

// Start starts the HTTP server.
func (h *HTTPPortal) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("HTTP portal already running")
	}

	go func() {
		var err error
		if h.config.TLS != nil {
			err = h.server.ListenAndServeTLS(h.config.TLS.CertFile, h.config.TLS.KeyFile)
		} else {
			err = h.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			// Log error (would use proper logger in production)
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	h.running = true
	return nil
}

// Stop stops the HTTP server.
func (h *HTTPPortal) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	err := h.server.Shutdown(ctx)
	h.running = false
	return err
}

// ListenAddress returns the address the server is listening on.
func (h *HTTPPortal) ListenAddress() string {
	return h.server.Addr
}

// BaseURL returns the base URL for this portal.
func (h *HTTPPortal) BaseURL() string {
	return h.baseURL
}

// HandleHTTP returns the HTTP handler.
func (h *HTTPPortal) HandleHTTP() any {
	return h.buildHandler()
}

// SetMiddleware sets HTTP middleware.
func (h *HTTPPortal) SetMiddleware(middleware []any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.middleware = make([]Middleware, 0, len(middleware))
	for _, m := range middleware {
		if mw, ok := m.(Middleware); ok {
			h.middleware = append(h.middleware, mw)
		}
	}
}

// EnableCORS enables CORS support.
func (h *HTTPPortal) EnableCORS(origins []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.CORSOrigins = origins
}

// Close closes the HTTP portal.
func (h *HTTPPortal) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return h.Stop(ctx)
}

// Health returns the health status of the portal.
func (h *HTTPPortal) Health(ctx context.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.running {
		return fmt.Errorf("HTTP portal not running")
	}

	return nil
}

// Private helper methods

func (h *HTTPPortal) buildHandler() http.Handler {
	handler := http.Handler(h.mux)

	// Apply middleware in reverse order
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i].Handle(handler)
	}

	// Apply CORS
	handler = h.corsHandler(handler)

	return handler
}

func (h *HTTPPortal) corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		if len(h.config.CORSOrigins) > 0 {
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range h.config.CORSOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *HTTPPortal) handleFunctionCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract function name from path
	functionName := h.extractFunctionName(r.URL.Path)

	h.mu.RLock()
	function, exists := h.functions[functionName]
	schema, hasSchema := h.schemas[functionName]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var requestData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create function input
	input := NewFunctionData(requestData)

	// Validate input if schema is available
	if hasSchema {
		if err := h.validateInput(input, schema); err != nil {
			http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Execute function
	ctx := r.Context()
	output, err := function.Call(ctx, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"result": output.Value(),
		"error":  nil,
	})
}

func (h *HTTPPortal) extractFunctionName(path string) string {
	// Extract function name from paths like "/functions/myFunc" or "/services/myService/myMethod"
	if len(path) >= 11 && path[:11] == "/functions/" {
		functionName := path[11:]
		return functionName
	}
	if len(path) >= 10 && path[:10] == "/services/" {
		// For services, extract "serviceName.methodName"
		parts := splitPath(path[10:])
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1]
		}
		// If we don't have enough parts, return empty string
		return ""
	}
	return path
}

func (h *HTTPPortal) getAuthority() string {
	return fmt.Sprintf("%s:%d", h.config.Host, h.config.Port)
}

func (h *HTTPPortal) extractServiceName(path string) string {
	// Extract service name from paths like "/services/myService/myMethod"
	if len(path) > 10 && path[:10] == "/services/" {
		parts := splitPath(path[10:])
		if len(parts) >= 1 {
			return parts[0]
		}
	}
	return path
}

func (h *HTTPPortal) validateInput(input api.FunctionData, schema core.FunctionSchema) error {
	// Basic validation - would be enhanced with proper schema validation
	return nil
}

func splitPath(path string) []string {
	parts := make([]string, 0)
	current := ""

	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
