package websocket

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

// NewPortal creates a new WebSocket portal with optional configuration
func NewPortal(configs ...Config) *WebSocketPortal {
	var config Config
	if len(configs) > 0 {
		config = configs[0]
	} else {
		config = DefaultConfig()
	}

	// Initialize portal
	portal := &WebSocketPortal{
		config:      config,
		connections: make(map[string]*WebSocketConnection),
		functions:   make(map[string]*FunctionRegistration),
		middleware:  make([]Middleware, 0),
		mu:          sync.RWMutex{},
	}

	return portal
}

// ConfigBuilder provides a fluent interface for building WebSocket configurations
type ConfigBuilder struct {
	config Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// Host sets the host address
func (b *ConfigBuilder) Host(host string) *ConfigBuilder {
	b.config.Host = host
	return b
}

// Port sets the port number
func (b *ConfigBuilder) Port(port int) *ConfigBuilder {
	b.config.Port = port
	return b
}

// Path sets the WebSocket endpoint path
func (b *ConfigBuilder) Path(path string) *ConfigBuilder {
	b.config.Path = path
	return b
}

// TLS enables TLS with the provided configuration
func (b *ConfigBuilder) TLS(tlsConfig *tls.Config) *ConfigBuilder {
	b.config.TLSConfig = tlsConfig
	return b
}

// BufferSizes sets the read and write buffer sizes
func (b *ConfigBuilder) BufferSizes(readSize, writeSize int) *ConfigBuilder {
	b.config.ReadBufferSize = readSize
	b.config.WriteBufferSize = writeSize
	return b
}

// HandshakeTimeout sets the WebSocket handshake timeout
func (b *ConfigBuilder) HandshakeTimeout(timeout time.Duration) *ConfigBuilder {
	b.config.HandshakeTimeout = timeout
	return b
}

// CheckOrigin sets the origin check function
func (b *ConfigBuilder) CheckOrigin(checkOrigin func(r *http.Request) bool) *ConfigBuilder {
	b.config.CheckOrigin = checkOrigin
	return b
}

// EnableCompression enables or disables compression
func (b *ConfigBuilder) EnableCompression(enable bool) *ConfigBuilder {
	b.config.EnableCompression = enable
	return b
}

// Timeouts sets various timeout configurations
func (b *ConfigBuilder) Timeouts(defaultTimeout, readTimeout, writeTimeout time.Duration) *ConfigBuilder {
	b.config.DefaultTimeout = defaultTimeout
	b.config.ReadTimeout = readTimeout
	b.config.WriteTimeout = writeTimeout
	return b
}

// Retry sets retry configuration
func (b *ConfigBuilder) Retry(maxRetries int, retryDelay time.Duration) *ConfigBuilder {
	b.config.MaxRetries = maxRetries
	b.config.RetryDelay = retryDelay
	return b
}

// Ping sets ping/pong configuration
func (b *ConfigBuilder) Ping(pingInterval, pongTimeout time.Duration) *ConfigBuilder {
	b.config.PingInterval = pingInterval
	b.config.PongTimeout = pongTimeout
	return b
}

// MaxMessageSize sets the maximum message size
func (b *ConfigBuilder) MaxMessageSize(size int64) *ConfigBuilder {
	b.config.MaxMessageSize = size
	return b
}

// Build returns the constructed configuration
func (b *ConfigBuilder) Build() Config {
	return b.config
}

// Preset configurations

// ServerConfig returns a configuration optimized for server use
func ServerConfig(host string, port int) Config {
	return NewConfigBuilder().
		Host(host).
		Port(port).
		Path("/ws").
		BufferSizes(4096, 4096).
		HandshakeTimeout(10*time.Second).
		Timeouts(60*time.Second, 60*time.Second, 10*time.Second).
		Ping(30*time.Second, 60*time.Second).
		MaxMessageSize(10 * 1024 * 1024). // 10MB
		EnableCompression(true).
		CheckOrigin(func(r *http.Request) bool { return true }).
		Build()
}

// ClientConfig returns a configuration optimized for client use
func ClientConfig() Config {
	return NewConfigBuilder().
		BufferSizes(2048, 2048).
		HandshakeTimeout(10*time.Second).
		Timeouts(30*time.Second, 30*time.Second, 10*time.Second).
		Retry(3, time.Second).
		MaxMessageSize(5 * 1024 * 1024). // 5MB
		EnableCompression(true).
		Build()
}

// SecureServerConfig returns a configuration for secure WebSocket server
func SecureServerConfig(host string, port int, tlsConfig *tls.Config) Config {
	config := ServerConfig(host, port)
	config.TLSConfig = tlsConfig
	return config
}

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig() Config {
	return NewConfigBuilder().
		Host("localhost").
		Port(8080).
		Path("/ws").
		BufferSizes(1024, 1024).
		HandshakeTimeout(5*time.Second).
		Timeouts(30*time.Second, 30*time.Second, 5*time.Second).
		Ping(15*time.Second, 30*time.Second).
		MaxMessageSize(1024 * 1024). // 1MB
		EnableCompression(false).    // Disabled for easier debugging
		CheckOrigin(func(r *http.Request) bool { return true }).
		Build()
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig(host string, port int) Config {
	return NewConfigBuilder().
		Host(host).
		Port(port).
		Path("/ws").
		BufferSizes(8192, 8192).
		HandshakeTimeout(10*time.Second).
		Timeouts(120*time.Second, 120*time.Second, 30*time.Second).
		Retry(5, 2*time.Second).
		Ping(60*time.Second, 120*time.Second).
		MaxMessageSize(50 * 1024 * 1024). // 50MB
		EnableCompression(true).
		CheckOrigin(func(r *http.Request) bool {
			// In production, you should implement proper origin checking
			origin := r.Header.Get("Origin")
			// Example: only allow specific origins
			allowedOrigins := []string{
				"https://yourapp.com",
				"https://www.yourapp.com",
			}
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		}).
		Build()
}

// HighPerformanceConfig returns a configuration optimized for high performance
func HighPerformanceConfig(host string, port int) Config {
	return NewConfigBuilder().
		Host(host).
		Port(port).
		Path("/ws").
		BufferSizes(16384, 16384). // Large buffers
		HandshakeTimeout(5*time.Second).
		Timeouts(60*time.Second, 60*time.Second, 5*time.Second).
		Retry(2, 500*time.Millisecond). // Fast retries
		Ping(30*time.Second, 60*time.Second).
		MaxMessageSize(100 * 1024 * 1024). // 100MB
		EnableCompression(true).
		CheckOrigin(func(r *http.Request) bool { return true }).
		Build()
}

// Convenience functions

// NewServer creates a WebSocket portal configured as a server
func NewServer(host string, port int) *WebSocketPortal {
	config := ServerConfig(host, port)
	return NewPortal(config)
}

// NewClient creates a WebSocket portal configured as a client
func NewClient() *WebSocketPortal {
	config := ClientConfig()
	return NewPortal(config)
}

// NewSecureServer creates a secure WebSocket portal with TLS
func NewSecureServer(host string, port int, tlsConfig *tls.Config) *WebSocketPortal {
	config := SecureServerConfig(host, port, tlsConfig)
	return NewPortal(config)
}

// NewDevelopmentPortal creates a portal for development use
func NewDevelopmentPortal() *WebSocketPortal {
	config := DevelopmentConfig()
	return NewPortal(config)
}

// NewProductionPortal creates a portal for production use
func NewProductionPortal(host string, port int) *WebSocketPortal {
	config := ProductionConfig(host, port)
	return NewPortal(config)
}

// NewHighPerformancePortal creates a portal optimized for high performance
func NewHighPerformancePortal(host string, port int) *WebSocketPortal {
	config := HighPerformanceConfig(host, port)
	return NewPortal(config)
}

// CreatePortalFromConfig creates a portal from an existing configuration
func CreatePortalFromConfig(config Config) *WebSocketPortal {
	return NewPortal(config)
}

// ClonePortal creates a new portal with the same configuration as an existing one
func ClonePortal(source *WebSocketPortal) *WebSocketPortal {
	return NewPortal(source.GetConfig())
}

// MergeConfigs merges multiple configurations (later configs override earlier ones)
func MergeConfigs(configs ...Config) Config {
	if len(configs) == 0 {
		return DefaultConfig()
	}

	result := configs[0]
	for i := 1; i < len(configs); i++ {
		config := configs[i]

		// Override non-zero values
		if config.Host != "" {
			result.Host = config.Host
		}
		if config.Port != 0 {
			result.Port = config.Port
		}
		if config.Path != "" {
			result.Path = config.Path
		}
		if config.TLSConfig != nil {
			result.TLSConfig = config.TLSConfig
		}
		if config.ReadBufferSize != 0 {
			result.ReadBufferSize = config.ReadBufferSize
		}
		if config.WriteBufferSize != 0 {
			result.WriteBufferSize = config.WriteBufferSize
		}
		if config.HandshakeTimeout != 0 {
			result.HandshakeTimeout = config.HandshakeTimeout
		}
		if config.CheckOrigin != nil {
			result.CheckOrigin = config.CheckOrigin
		}
		if config.DefaultTimeout != 0 {
			result.DefaultTimeout = config.DefaultTimeout
		}
		if config.MaxRetries != 0 {
			result.MaxRetries = config.MaxRetries
		}
		if config.RetryDelay != 0 {
			result.RetryDelay = config.RetryDelay
		}
		if config.PingInterval != 0 {
			result.PingInterval = config.PingInterval
		}
		if config.PongTimeout != 0 {
			result.PongTimeout = config.PongTimeout
		}
		if config.MaxMessageSize != 0 {
			result.MaxMessageSize = config.MaxMessageSize
		}
		if config.WriteTimeout != 0 {
			result.WriteTimeout = config.WriteTimeout
		}
		if config.ReadTimeout != 0 {
			result.ReadTimeout = config.ReadTimeout
		}

		// Boolean field - always override
		result.EnableCompression = config.EnableCompression
	}

	return result
}
