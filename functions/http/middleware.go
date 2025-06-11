package http

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Built-in middleware implementations

// LoggingMiddleware logs HTTP requests and responses
type LoggingMiddleware struct {
	Logger func(format string, args ...interface{})
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger: func(format string, args ...interface{}) {
			log.Printf("[HTTP] "+format, args...)
		},
	}
}

func (m *LoggingMiddleware) ProcessRequest(req *http.Request) error {
	m.Logger("Request: %s %s from %s", req.Method, req.URL.String(), req.RemoteAddr)
	return nil
}

func (m *LoggingMiddleware) ProcessResponse(resp *http.Response) error {
	url := "unknown"
	if resp.Request != nil && resp.Request.URL != nil {
		url = resp.Request.URL.String()
	}
	m.Logger("Response: %d %s for %s", resp.StatusCode, http.StatusText(resp.StatusCode), url)
	return nil
}

// AuthenticationMiddleware handles various authentication schemes
type AuthenticationMiddleware struct {
	Type   string // "bearer", "basic", "api_key"
	Secret string // Secret key or token
	Header string // Custom header name for API key auth
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

func (m *AuthenticationMiddleware) ProcessRequest(req *http.Request) error {
	switch m.Type {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+m.Secret)
	case "api_key":
		req.Header.Set(m.Header, m.Secret)
	default:
		return fmt.Errorf("unsupported auth type: %s", m.Type)
	}
	return nil
}

func (m *AuthenticationMiddleware) ProcessResponse(resp *http.Response) error {
	// Authentication middleware doesn't typically process responses
	return nil
}

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	AllowedCredentials bool
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Content-Type", "Authorization"},
		AllowedCredentials: false,
	}
}

func (m *CORSMiddleware) ProcessRequest(req *http.Request) error {
	// CORS headers are typically set on responses, not requests
	return nil
}

func (m *CORSMiddleware) ProcessResponse(resp *http.Response) error {
	if len(m.AllowedOrigins) > 0 {
		resp.Header.Set("Access-Control-Allow-Origin", strings.Join(m.AllowedOrigins, ", "))
	}
	if len(m.AllowedMethods) > 0 {
		resp.Header.Set("Access-Control-Allow-Methods", strings.Join(m.AllowedMethods, ", "))
	}
	if len(m.AllowedHeaders) > 0 {
		resp.Header.Set("Access-Control-Allow-Headers", strings.Join(m.AllowedHeaders, ", "))
	}
	if m.AllowedCredentials {
		resp.Header.Set("Access-Control-Allow-Credentials", "true")
	}
	return nil
}

// MetricsMiddleware collects request/response metrics
type MetricsMiddleware struct {
	RequestCount  int64
	ResponseCount int64
	TotalLatency  time.Duration
	AvgLatency    time.Duration
	requestTimes  map[string]time.Time
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		requestTimes: make(map[string]time.Time),
	}
}

func (m *MetricsMiddleware) ProcessRequest(req *http.Request) error {
	m.RequestCount++
	// Store request start time (using request URL as key)
	m.requestTimes[req.URL.String()] = time.Now()
	return nil
}

func (m *MetricsMiddleware) ProcessResponse(resp *http.Response) error {
	m.ResponseCount++

	// Calculate latency if we have start time
	if startTime, exists := m.requestTimes[resp.Request.URL.String()]; exists {
		latency := time.Since(startTime)
		m.TotalLatency += latency
		m.AvgLatency = m.TotalLatency / time.Duration(m.ResponseCount)

		// Clean up to prevent memory leak
		delete(m.requestTimes, resp.Request.URL.String())
	}

	return nil
}

// GetMetrics returns current metrics
func (m *MetricsMiddleware) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"request_count":  m.RequestCount,
		"response_count": m.ResponseCount,
		"total_latency":  m.TotalLatency.String(),
		"avg_latency":    m.AvgLatency.String(),
	}
}

// HeaderMiddleware adds custom headers to requests
type HeaderMiddleware struct {
	Headers map[string]string
}

// NewHeaderMiddleware creates middleware that adds custom headers
func NewHeaderMiddleware(headers map[string]string) *HeaderMiddleware {
	return &HeaderMiddleware{
		Headers: headers,
	}
}

func (m *HeaderMiddleware) ProcessRequest(req *http.Request) error {
	for key, value := range m.Headers {
		req.Header.Set(key, value)
	}
	return nil
}

func (m *HeaderMiddleware) ProcessResponse(resp *http.Response) error {
	// Header middleware typically doesn't process responses
	return nil
}

// RetryMiddleware handles retry logic (though this is better handled at client level)
type RetryMiddleware struct {
	MaxRetries int
	RetryDelay time.Duration
}

// NewRetryMiddleware creates retry middleware
func NewRetryMiddleware(maxRetries int, delay time.Duration) *RetryMiddleware {
	return &RetryMiddleware{
		MaxRetries: maxRetries,
		RetryDelay: delay,
	}
}

func (m *RetryMiddleware) ProcessRequest(req *http.Request) error {
	// Retry logic is handled at the client level in client.go
	// This middleware is mainly for configuration
	return nil
}

func (m *RetryMiddleware) ProcessResponse(resp *http.Response) error {
	// Response processing for retry decisions
	return nil
}

// ValidationMiddleware validates request/response data
type ValidationMiddleware struct {
	ValidateRequest  func(*http.Request) error
	ValidateResponse func(*http.Response) error
}

// NewValidationMiddleware creates validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

func (m *ValidationMiddleware) ProcessRequest(req *http.Request) error {
	if m.ValidateRequest != nil {
		return m.ValidateRequest(req)
	}
	return nil
}

func (m *ValidationMiddleware) ProcessResponse(resp *http.Response) error {
	if m.ValidateResponse != nil {
		return m.ValidateResponse(resp)
	}
	return nil
}

// Middleware chain helper functions

// ChainMiddleware combines multiple middlewares into one
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return MiddlewareFunc{
		RequestFunc: func(req *http.Request) error {
			for _, middleware := range middlewares {
				if err := middleware.ProcessRequest(req); err != nil {
					return err
				}
			}
			return nil
		},
		ResponseFunc: func(resp *http.Response) error {
			// Process in reverse order for responses
			for i := len(middlewares) - 1; i >= 0; i-- {
				if err := middlewares[i].ProcessResponse(resp); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// ConditionalMiddleware applies middleware based on a condition
func ConditionalMiddleware(condition func(*http.Request) bool, middleware Middleware) Middleware {
	return MiddlewareFunc{
		RequestFunc: func(req *http.Request) error {
			if condition(req) {
				return middleware.ProcessRequest(req)
			}
			return nil
		},
		ResponseFunc: func(resp *http.Response) error {
			if condition(resp.Request) {
				return middleware.ProcessResponse(resp)
			}
			return nil
		},
	}
}
