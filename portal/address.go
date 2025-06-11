package portal

import (
	"fmt"
	"net/url"
	"strings"

	"defs.dev/schema/api"
)

// AddressImpl implements the api.Address interface
type AddressImpl struct {
	scheme    string
	authority string
	path      string
	query     map[string]string
	fragment  string
}

// NewAddress creates a new Address from a URL string
func NewAddress(addressStr string) (api.Address, error) {
	if addressStr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	// Handle local addresses specially
	if strings.HasPrefix(addressStr, "local://") {
		return parseLocalAddress(addressStr)
	}

	// Parse as standard URL
	u, err := url.Parse(addressStr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	// Convert URL query to map
	queryMap := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			queryMap[key] = values[0] // Take first value if multiple
		}
	}

	return &AddressImpl{
		scheme:    u.Scheme,
		authority: u.Host,
		path:      u.Path,
		query:     queryMap,
		fragment:  u.Fragment,
	}, nil
}

// MustNewAddress creates a new Address, panicking on error
func MustNewAddress(addressStr string) api.Address {
	addr, err := NewAddress(addressStr)
	if err != nil {
		panic(err)
	}
	return addr
}

func parseLocalAddress(addressStr string) (api.Address, error) {
	// local://function-name/unique-id?param=value#fragment
	u, err := url.Parse(addressStr)
	if err != nil {
		return nil, fmt.Errorf("invalid local address format: %w", err)
	}

	queryMap := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			queryMap[key] = values[0]
		}
	}

	// For local addresses, if there's no path but there's a host,
	// treat the host as the path (e.g., "local://add" -> path="/add")
	path := u.Path
	if path == "" && u.Host != "" {
		path = "/" + u.Host
	}

	return &AddressImpl{
		scheme:    "local",
		authority: "", // Local addresses don't have authority
		path:      path,
		query:     queryMap,
		fragment:  u.Fragment,
	}, nil
}

// String returns the full address as a string
func (a *AddressImpl) String() string {
	var builder strings.Builder

	// Scheme
	builder.WriteString(a.scheme)
	builder.WriteString("://")

	// Authority (host:port)
	if a.authority != "" {
		builder.WriteString(a.authority)
	}

	// Path
	if a.path != "" {
		if !strings.HasPrefix(a.path, "/") && a.authority != "" {
			builder.WriteString("/")
		}
		builder.WriteString(a.path)
	}

	// Query
	if len(a.query) > 0 {
		builder.WriteString("?")
		first := true
		for key, value := range a.query {
			if !first {
				builder.WriteString("&")
			}
			builder.WriteString(url.QueryEscape(key))
			builder.WriteString("=")
			builder.WriteString(url.QueryEscape(value))
			first = false
		}
	}

	// Fragment
	if a.fragment != "" {
		builder.WriteString("#")
		builder.WriteString(a.fragment)
	}

	return builder.String()
}

// Scheme returns the transport scheme
func (a *AddressImpl) Scheme() string {
	return a.scheme
}

// Authority returns the authority portion (host:port for network addresses)
func (a *AddressImpl) Authority() string {
	return a.authority
}

// Path returns the path portion
func (a *AddressImpl) Path() string {
	return a.path
}

// Query returns query parameters as a map
func (a *AddressImpl) Query() map[string]string {
	// Return a copy to prevent modification
	result := make(map[string]string)
	for k, v := range a.query {
		result[k] = v
	}
	return result
}

// Fragment returns the fragment identifier
func (a *AddressImpl) Fragment() string {
	return a.fragment
}

// IsLocal returns true if this is a local address
func (a *AddressImpl) IsLocal() bool {
	return a.scheme == "local"
}

// IsNetwork returns true if this requires network communication
func (a *AddressImpl) IsNetwork() bool {
	switch a.scheme {
	case "local", "test", "mock":
		return false
	case "http", "https", "ws", "wss", "grpc", "tcp", "udp":
		return true
	default:
		// Default to network for unknown schemes
		return true
	}
}

// AddressBuilderImpl implements the api.AddressBuilder interface
type AddressBuilderImpl struct {
	scheme    string
	authority string
	host      string
	port      int
	path      string
	query     map[string]string
	fragment  string
}

// NewAddressBuilder creates a new AddressBuilder
func NewAddressBuilder() api.AddressBuilder {
	return &AddressBuilderImpl{
		query: make(map[string]string),
	}
}

// Scheme sets the scheme
func (b *AddressBuilderImpl) Scheme(scheme string) api.AddressBuilder {
	b.scheme = scheme
	return b
}

// Authority sets the authority directly
func (b *AddressBuilderImpl) Authority(authority string) api.AddressBuilder {
	b.authority = authority
	return b
}

// Host sets the host
func (b *AddressBuilderImpl) Host(host string) api.AddressBuilder {
	b.host = host
	return b
}

// Port sets the port
func (b *AddressBuilderImpl) Port(port int) api.AddressBuilder {
	b.port = port
	return b
}

// Path sets the path
func (b *AddressBuilderImpl) Path(path string) api.AddressBuilder {
	b.path = path
	return b
}

// Query adds a query parameter
func (b *AddressBuilderImpl) Query(key, value string) api.AddressBuilder {
	b.query[key] = value
	return b
}

// Fragment sets the fragment
func (b *AddressBuilderImpl) Fragment(fragment string) api.AddressBuilder {
	b.fragment = fragment
	return b
}

// Build creates the Address
func (b *AddressBuilderImpl) Build() api.Address {
	authority := b.authority

	// Build authority from host and port if not set directly
	if authority == "" && b.host != "" {
		if b.port > 0 {
			authority = fmt.Sprintf("%s:%d", b.host, b.port)
		} else {
			authority = b.host
		}
	}

	return &AddressImpl{
		scheme:    b.scheme,
		authority: authority,
		path:      b.path,
		query:     b.query,
		fragment:  b.fragment,
	}
}

// Utility functions for common address patterns

// LocalAddress creates a local address
func LocalAddress(functionName string) api.Address {
	return NewAddressBuilder().
		Scheme("local").
		Path("/" + functionName).
		Build()
}

// HTTPAddress creates an HTTP address
func HTTPAddress(host string, port int, path string) api.Address {
	return NewAddressBuilder().
		Scheme("http").
		Host(host).
		Port(port).
		Path(path).
		Build()
}

// HTTPSAddress creates an HTTPS address
func HTTPSAddress(host string, path string) api.Address {
	return NewAddressBuilder().
		Scheme("https").
		Host(host).
		Path(path).
		Build()
}

// WebSocketAddress creates a WebSocket address
func WebSocketAddress(host string, port int, path string) api.Address {
	return NewAddressBuilder().
		Scheme("ws").
		Host(host).
		Port(port).
		Path(path).
		Build()
}

// WebSocketSecureAddress creates a secure WebSocket address
func WebSocketSecureAddress(host string, path string) api.Address {
	return NewAddressBuilder().
		Scheme("wss").
		Host(host).
		Path(path).
		Build()
}

// ParseScheme extracts scheme from address string
func ParseScheme(addressStr string) string {
	if idx := strings.Index(addressStr, "://"); idx != -1 {
		return addressStr[:idx]
	}
	return ""
}
