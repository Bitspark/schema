package javascript

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"defs.dev/schema"
	"defs.dev/schema/functions"
)

// JavaScriptPortal implements functions.Portal[JSFunction] for JavaScript function execution
type JavaScriptPortal struct {
	config    Config
	functions map[string]*JavaScriptFunction // address -> function
	mu        sync.RWMutex
}

// NewPortal creates a new JavaScript portal with the given configuration
func NewPortal(config Config) *JavaScriptPortal {
	return &JavaScriptPortal{
		config:    config,
		functions: make(map[string]*JavaScriptFunction),
	}
}

// NewPortalWithDefaults creates a new JavaScript portal with default configuration
func NewPortalWithDefaults() *JavaScriptPortal {
	return NewPortal(DefaultConfig())
}

// Portal interface implementation

func (p *JavaScriptPortal) Scheme() string {
	return "js"
}

func (p *JavaScriptPortal) GenerateAddress(name string, data JSFunction) string {
	// Generate unique ID based on JavaScript code and function name
	id := generateCodeID(data.Code, data.FunctionName)
	engine := p.config.Engine
	if engine == "" {
		engine = "goja"
	}
	return fmt.Sprintf("js://%s/%s/%s", engine, name, id[:8])
}

func (p *JavaScriptPortal) Apply(address string, schema *schema.FunctionSchema, data JSFunction) schema.Function {
	// Extract name from address
	name := extractNameFromJSAddress(address)

	// Create JavaScriptFunction
	jsFunc := &JavaScriptFunction{
		name:       name,
		address:    address,
		parameters: buildParametersFromSchema(schema),
		returns:    buildReturnsFromSchema(schema),
		jsFunction: data,
		portal:     p,
	}

	// Store for resolution
	p.mu.Lock()
	p.functions[address] = jsFunc
	p.mu.Unlock()

	return jsFunc
}

func (p *JavaScriptPortal) ResolveFunction(ctx context.Context, address string) (schema.Function, error) {
	p.mu.RLock()
	function, exists := p.functions[address]
	p.mu.RUnlock()

	if !exists {
		return nil, &functions.PortalError{
			Scheme:  "js",
			Address: address,
			Message: "javascript function not found",
		}
	}

	return function, nil
}

// Helper functions

func generateCodeID(code, functionName string) string {
	// Create hash from code and function name for unique identification
	h := sha256.New()
	h.Write([]byte(code))
	h.Write([]byte(functionName))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func extractNameFromJSAddress(address string) string {
	// Parse js://engine/functionName/id to extract functionName
	// Example: js://goja/calculateDiscount/abc12345
	parts := strings.SplitN(strings.TrimPrefix(address, "js://"), "/", 3)
	if len(parts) >= 2 {
		return parts[1] // functionName
	}
	return "unknown"
}

// Configuration and management methods

func (p *JavaScriptPortal) Config() Config {
	return p.config
}

func (p *JavaScriptPortal) SetConfig(config Config) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
}

// Stats returns portal statistics
func (p *JavaScriptPortal) Stats() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]any{
		"engine":          p.config.Engine,
		"functions_count": len(p.functions),
		"memory_limit":    p.config.MemoryLimit,
		"default_timeout": p.config.DefaultTimeout.String(),
		"max_engines":     p.config.MaxEngines,
		"security_policy": p.config.SecurityPolicy,
	}
}

// Clear removes all registered functions (useful for testing)
func (p *JavaScriptPortal) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.functions = make(map[string]*JavaScriptFunction)
}
