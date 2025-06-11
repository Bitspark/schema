package http

import (
	"net/http"
	"time"
)

// NewPortal creates a new HTTP portal with default configuration
func NewPortal(config ...Config) *HTTPPortal {
	// Use default config if none provided
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultConfig()
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: cfg.DefaultTimeout,
	}

	// Create portal
	portal := &HTTPPortal{
		client:     client,
		config:     cfg,
		mux:        http.NewServeMux(),
		functions:  make(map[string]*FunctionRegistration),
		middleware: make([]Middleware, 0),
	}

	return portal
}

// NewServerPortal creates a portal optimized for server/provider usage
func NewServerPortal(config Config) *HTTPPortal {
	portal := NewPortal(config)

	// Start the server automatically
	if err := portal.StartServer(); err != nil {
		// In production, you'd want better error handling
		panic("Failed to start HTTP server: " + err.Error())
	}

	return portal
}

// NewClientPortal creates a portal optimized for client/consumer usage
func NewClientPortal(config Config) *HTTPPortal {
	portal := NewPortal(config)

	// Client-only portal doesn't need server functionality
	// You might configure client-specific settings here

	return portal
}

// DefaultConfig returns a default configuration for HTTP portal
func DefaultConfig() Config {
	return Config{
		Host:           "localhost",
		Port:           8080,
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     time.Second,
		UserAgent:      "go-llm4-http-portal/1.0",
		BasePath:       "/api",
	}
}

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig() Config {
	return Config{
		Host:           "localhost",
		Port:           8080,
		DefaultTimeout: 10 * time.Second,
		MaxRetries:     1,
		RetryDelay:     500 * time.Millisecond,
		UserAgent:      "go-llm4-http-portal/dev",
		BasePath:       "/api/dev",
	}
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig(host string, port int) Config {
	return Config{
		Host:           host,
		Port:           port,
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     5,
		RetryDelay:     2 * time.Second,
		UserAgent:      "go-llm4-http-portal/1.0",
		BasePath:       "/api/v1",
	}
}

// ConfigBuilder provides a fluent interface for building configurations
type ConfigBuilder struct {
	config Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// Host sets the server host
func (b *ConfigBuilder) Host(host string) *ConfigBuilder {
	b.config.Host = host
	return b
}

// Port sets the server port
func (b *ConfigBuilder) Port(port int) *ConfigBuilder {
	b.config.Port = port
	return b
}

// Timeout sets the default client timeout
func (b *ConfigBuilder) Timeout(timeout time.Duration) *ConfigBuilder {
	b.config.DefaultTimeout = timeout
	return b
}

// Retries sets the maximum number of retries
func (b *ConfigBuilder) Retries(retries int) *ConfigBuilder {
	b.config.MaxRetries = retries
	return b
}

// RetryDelay sets the delay between retries
func (b *ConfigBuilder) RetryDelay(delay time.Duration) *ConfigBuilder {
	b.config.RetryDelay = delay
	return b
}

// UserAgent sets the HTTP user agent
func (b *ConfigBuilder) UserAgent(ua string) *ConfigBuilder {
	b.config.UserAgent = ua
	return b
}

// BasePath sets the base path for all endpoints
func (b *ConfigBuilder) BasePath(path string) *ConfigBuilder {
	b.config.BasePath = path
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() Config {
	return b.config
}

// Portal creation helpers

// WithConfig creates a portal with custom configuration using builder pattern
func WithConfig() *ConfigBuilder {
	return NewConfigBuilder()
}

// QuickPortal creates a portal with minimal configuration for quick setup
func QuickPortal(host string, port int) *HTTPPortal {
	config := Config{
		Host:           host,
		Port:           port,
		DefaultTimeout: 15 * time.Second,
		MaxRetries:     2,
		RetryDelay:     time.Second,
		UserAgent:      "go-llm4-quick",
		BasePath:       "/api",
	}
	return NewPortal(config)
}

// LocalPortal creates a portal for local development (localhost:8080)
func LocalPortal() *HTTPPortal {
	return QuickPortal("localhost", 8080)
}

// Example factory functions for common scenarios

// MicroservicePortal creates a portal configured for microservice architecture
func MicroservicePortal(serviceName string, port int) *HTTPPortal {
	config := Config{
		Host:           "0.0.0.0", // Listen on all interfaces
		Port:           port,
		DefaultTimeout: 45 * time.Second,
		MaxRetries:     3,
		RetryDelay:     2 * time.Second,
		UserAgent:      "microservice-" + serviceName,
		BasePath:       "/api/" + serviceName,
	}
	return NewServerPortal(config)
}

// APIGatewayPortal creates a portal configured for API gateway usage
func APIGatewayPortal() *HTTPPortal {
	config := Config{
		Host:           "0.0.0.0",
		Port:           80,
		DefaultTimeout: 60 * time.Second,
		MaxRetries:     5,
		RetryDelay:     time.Second,
		UserAgent:      "api-gateway",
		BasePath:       "/gateway",
	}
	return NewServerPortal(config)
}
