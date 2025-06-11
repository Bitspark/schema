package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"defs.dev/schema"
)

// StartServer starts the HTTP server for serving functions as endpoints
func (p *HTTPPortal) StartServer() error {
	if p.server != nil {
		return fmt.Errorf("server already started")
	}

	address := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	// Create listener to capture actual port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Update config with actual port if it was 0 (dynamic allocation)
	if p.config.Port == 0 {
		if addr, ok := listener.Addr().(*net.TCPAddr); ok {
			p.config.Port = addr.Port
		}
	}

	p.server = &http.Server{
		Handler:   p.mux,
		TLSConfig: p.config.TLSConfig,
	}

	// Start server in background with the listener
	go func() {
		var err error
		if p.config.TLSConfig != nil {
			err = p.server.ServeTLS(listener, "", "")
		} else {
			err = p.server.Serve(listener)
		}

		if err != nil && err != http.ErrServerClosed {
			// Log error - in production you'd use proper logging
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

// StopServer stops the HTTP server
func (p *HTTPPortal) StopServer(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	err := p.server.Shutdown(ctx)
	p.server = nil
	return err
}

// RegisterHandler registers a function handler as an HTTP endpoint
func (p *HTTPPortal) RegisterHandler(name string, address string, funcSchema *schema.FunctionSchema, handler schema.FunctionHandler) error {
	path := p.extractPathFromAddress(address)

	// Create HTTP handler for the function
	httpHandler := p.createFunctionHandler(funcSchema, handler)

	// Apply middleware
	finalHandler := p.applyMiddleware(httpHandler)

	// Register the handler
	p.mux.HandleFunc(path, finalHandler)

	// Store registration
	p.functions[address] = &FunctionRegistration{
		Name:    name,
		Address: address,
		Schema:  funcSchema,
		Handler: handler,
	}

	return nil
}

// createFunctionHandler creates an HTTP handler for a function
func (p *HTTPPortal) createFunctionHandler(funcSchema *schema.FunctionSchema, handler schema.FunctionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST for function calls
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set response content type
		w.Header().Set("Content-Type", "application/json")

		// Parse request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			p.writeErrorResponse(w, http.StatusBadRequest, "Failed to read request body", err)
			return
		}
		defer r.Body.Close()

		// Parse JSON parameters
		var params map[string]any
		if len(body) > 0 {
			if err := json.Unmarshal(body, &params); err != nil {
				p.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON in request body", err)
				return
			}
		}

		// Validate parameters against schema (if schema validation is implemented)
		// TODO: Add schema validation here

		// Call the function
		ctx := r.Context()
		input := schema.NewFunctionInput(params)
		output, err := handler(ctx, input)
		if err != nil {
			// Check if it's an HTTPFunctionError with specific status code
			if httpErr, ok := err.(*HTTPFunctionError); ok {
				p.writeErrorResponse(w, httpErr.StatusCode, httpErr.Message, err)
			} else {
				p.writeErrorResponse(w, http.StatusInternalServerError, "Function execution failed", err)
			}
			return
		}

		// Serialize result - extract the value from FunctionOutput
		result := output.Value()
		response, err := json.Marshal(result)
		if err != nil {
			p.writeErrorResponse(w, http.StatusInternalServerError, "Failed to serialize result", err)
			return
		}

		// Write successful response
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

// applyMiddleware applies all registered middleware to an HTTP handler
func (p *HTTPPortal) applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply request middleware
		for _, middleware := range p.middleware {
			if err := middleware.ProcessRequest(r); err != nil {
				http.Error(w, "Middleware error: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Create response wrapper to capture response for middleware
		rw := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the handler
		handler(rw, r)

		// Apply response middleware (best effort - response already sent)
		resp := &http.Response{
			StatusCode: rw.statusCode,
			Header:     w.Header(),
			Request:    r, // Set the request so middleware can access it
		}

		for _, middleware := range p.middleware {
			middleware.ProcessResponse(resp) // Ignore errors for response middleware
		}
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// writeErrorResponse writes a JSON error response
func (p *HTTPPortal) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	response, _ := json.Marshal(errorResponse)
	w.Write(response)
}

// extractPathFromAddress extracts the path component from an address
func (p *HTTPPortal) extractPathFromAddress(address string) string {
	// Parse address format: http://host:port/path/functionId
	// Extract the path component

	// Remove protocol
	if strings.HasPrefix(address, "http://") {
		address = address[7:]
	} else if strings.HasPrefix(address, "https://") {
		address = address[8:]
	}

	// Find first slash after host:port
	parts := strings.SplitN(address, "/", 2)
	if len(parts) < 2 {
		return p.config.BasePath + "/unknown"
	}

	path := "/" + parts[1]

	// Ensure path starts with base path
	if p.config.BasePath != "" && !strings.HasPrefix(path, p.config.BasePath) {
		path = p.config.BasePath + path
	}

	return path
}

// GetServerAddress returns the full server address
func (p *HTTPPortal) GetServerAddress() string {
	scheme := "http"
	if p.config.TLSConfig != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, p.config.Host, p.config.Port)
}

// IsServerRunning checks if the server is currently running
func (p *HTTPPortal) IsServerRunning() bool {
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
