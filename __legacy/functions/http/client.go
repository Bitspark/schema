package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"defs.dev/schema"
)

// HTTPClientFunction represents a function that makes HTTP requests
type HTTPClientFunction struct {
	address  string
	schema   *schema.FunctionSchema
	endpoint *HTTPEndpoint
	client   *http.Client
	portal   *HTTPPortal
}

// Call implements the Function interface for HTTP client functions
func (f *HTTPClientFunction) Call(ctx context.Context, params map[string]any) (any, error) {
	// Build the HTTP request
	req, err := f.buildRequest(ctx, params)
	if err != nil {
		return nil, NewSerializationError("encode request", err)
	}

	// Apply authentication if configured
	if err := f.applyAuth(req); err != nil {
		return nil, err
	}

	// Apply middleware
	for _, middleware := range f.portal.middleware {
		if err := middleware.ProcessRequest(req); err != nil {
			return nil, NewNetworkError(err, map[string]interface{}{
				"middleware": "request",
				"address":    f.address,
			})
		}
	}

	// Make the HTTP request with retries
	resp, err := f.makeRequestWithRetries(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Apply response middleware
	for _, middleware := range f.portal.middleware {
		middleware.ProcessResponse(resp) // Best effort, ignore errors
	}

	// Handle HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, f.handleHTTPError(resp.StatusCode, string(body))
	}

	// Parse response
	return f.parseResponse(resp)
}

// Schema returns the function schema
func (f *HTTPClientFunction) Schema() *schema.FunctionSchema {
	return f.schema
}

// Address returns the function address
func (f *HTTPClientFunction) Address() string {
	return f.address
}

// buildRequest builds an HTTP request from parameters
func (f *HTTPClientFunction) buildRequest(ctx context.Context, params map[string]any) (*http.Request, error) {
	// Serialize parameters to JSON
	var body io.Reader
	if params != nil {
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	// Build URL
	requestURL := f.buildURL()

	// Create request
	req, err := http.NewRequestWithContext(ctx, f.getMethod(), requestURL, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set user agent
	if f.portal.config.UserAgent != "" {
		req.Header.Set("User-Agent", f.portal.config.UserAgent)
	}

	// Apply static headers from endpoint config
	if f.endpoint != nil {
		for key, value := range f.endpoint.Headers {
			req.Header.Set(key, value)
		}

		// Apply query parameters
		if len(f.endpoint.Query) > 0 {
			q := req.URL.Query()
			for key, value := range f.endpoint.Query {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()
		}
	}

	return req, nil
}

// buildURL builds the request URL
func (f *HTTPClientFunction) buildURL() string {
	if f.endpoint != nil && f.endpoint.BaseURL != "" {
		// Use configured base URL and path
		baseURL := strings.TrimRight(f.endpoint.BaseURL, "/")
		path := strings.TrimLeft(f.endpoint.Path, "/")
		return fmt.Sprintf("%s/%s", baseURL, path)
	}

	// Use the address directly
	return f.address
}

// getMethod returns the HTTP method to use
func (f *HTTPClientFunction) getMethod() string {
	if f.endpoint != nil && f.endpoint.Method != "" {
		return f.endpoint.Method
	}
	return http.MethodPost // Default to POST for function calls
}

// applyAuth applies authentication to the request
func (f *HTTPClientFunction) applyAuth(req *http.Request) error {
	if f.endpoint == nil || f.endpoint.Auth == nil {
		return nil
	}

	auth := f.endpoint.Auth

	switch auth.Type {
	case "bearer":
		if auth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		}
	case "basic":
		if auth.User != "" && auth.Pass != "" {
			req.SetBasicAuth(auth.User, auth.Pass)
		}
	case "api_key":
		if auth.Token != "" {
			header := auth.Header
			if header == "" {
				header = "X-API-Key"
			}
			req.Header.Set(header, auth.Token)
		}
	default:
		return NewNetworkError(fmt.Errorf("unsupported auth type: %s", auth.Type), map[string]interface{}{
			"auth_type": auth.Type,
			"address":   f.address,
		})
	}

	return nil
}

// makeRequestWithRetries makes the HTTP request with retry logic
func (f *HTTPClientFunction) makeRequestWithRetries(req *http.Request) (*http.Response, error) {
	maxRetries := f.portal.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1 // At least one attempt
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Clone the request for retries (body might be consumed)
		reqClone, err := f.cloneRequest(req)
		if err != nil {
			return nil, NewSerializationError("clone request", err)
		}

		// Set timeout
		timeout := f.getRequestTimeout()
		if timeout > 0 {
			ctx, cancel := context.WithTimeout(reqClone.Context(), timeout)
			defer cancel()
			reqClone = reqClone.WithContext(ctx)
		}

		// Make the request
		resp, err := f.client.Do(reqClone)

		// Check for success
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !f.isRetryableError(err) {
			break
		}

		// Wait before retry (except on last attempt)
		if attempt < maxRetries-1 {
			retryDelay := f.portal.config.RetryDelay
			if retryDelay <= 0 {
				retryDelay = time.Second // Default 1 second
			}
			time.Sleep(retryDelay)
		}
	}

	// All retries failed
	return nil, NewNetworkError(lastErr, map[string]interface{}{
		"address":     f.address,
		"max_retries": maxRetries,
	})
}

// cloneRequest clones an HTTP request for retries
func (f *HTTPClientFunction) cloneRequest(req *http.Request) (*http.Request, error) {
	// Read body if present
	var body io.Reader
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()

		// Set body for original and clone
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		body = bytes.NewReader(bodyBytes)
	}

	// Create clone
	clone, err := http.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	for key, values := range req.Header {
		for _, value := range values {
			clone.Header.Add(key, value)
		}
	}

	return clone, nil
}

// getRequestTimeout returns the timeout for this request
func (f *HTTPClientFunction) getRequestTimeout() time.Duration {
	// Use endpoint-specific timeout if available
	if f.endpoint != nil && f.endpoint.Timeout > 0 {
		return f.endpoint.Timeout
	}

	// Use portal default timeout
	if f.portal.config.DefaultTimeout > 0 {
		return f.portal.config.DefaultTimeout
	}

	// No timeout
	return 0
}

// isRetryableError checks if an error should trigger a retry
func (f *HTTPClientFunction) isRetryableError(err error) bool {
	// Check if it's a timeout or network error
	if urlErr, ok := err.(*url.Error); ok {
		if urlErr.Timeout() {
			return true
		}
		// Network errors are generally retryable
		return true
	}

	return false
}

// handleHTTPError creates appropriate error for HTTP status codes
func (f *HTTPClientFunction) handleHTTPError(statusCode int, responseBody string) error {
	if statusCode >= 400 && statusCode < 500 {
		return NewClientError(statusCode, f.address, responseBody)
	} else if statusCode >= 500 {
		return NewServerHTTPError(statusCode, f.address, responseBody)
	}

	return NewHTTPError(statusCode, "Unexpected status code", f.address, responseBody)
}

// parseResponse parses the HTTP response
func (f *HTTPClientFunction) parseResponse(resp *http.Response) (any, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewSerializationError("read response", err)
	}

	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewSerializationError("decode response", err)
	}

	return result, nil
}

// CreateClientFunction creates a new HTTP client function
func (p *HTTPPortal) CreateClientFunction(address string, schema *schema.FunctionSchema, endpoint *HTTPEndpoint) Function {
	return &HTTPClientFunction{
		address:  address,
		schema:   schema,
		endpoint: endpoint,
		client:   p.client,
		portal:   p,
	}
}
