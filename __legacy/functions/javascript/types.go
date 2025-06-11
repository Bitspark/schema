package javascript

import (
	"time"
)

// JSFunction contains only the essential JavaScript code and metadata (implementation data)
type JSFunction struct {
	Code         string         // JavaScript source code
	FunctionName string         // Function name to call
	Timeout      *time.Duration // Function-specific timeout override (optional)
	Modules      []string       // Function-specific dependencies (future)
}

// Config represents portal-level configuration (infrastructure concerns)
type Config struct {
	Engine         string        // JavaScript engine type ("goja" for now)
	MemoryLimit    int64         // Maximum memory per engine
	StackSize      int           // Call stack limit
	DefaultTimeout time.Duration // Default execution timeout
	MaxEngines     int           // Engine pool size (future)
	SecurityPolicy SecurityLevel // Security sandbox level (future)
	PreloadModules []string      // Modules to load at startup (future)
}

// SecurityLevel represents different security sandbox levels (future implementation)
type SecurityLevel int

const (
	PermissiveSandbox SecurityLevel = iota // No restrictions
	ModerateSandbox                        // Some restrictions
	StrictSandbox                          // Maximum restrictions
)

// SecurityConfig represents security constraints (future implementation)
type SecurityConfig struct {
	AllowFileAccess   bool     // File system access
	AllowNetAccess    bool     // Network access
	AllowedDomains    []string // Whitelisted domains
	RestrictedGlobals []string // Blocked global objects
}

// ModuleSpec represents JavaScript module specifications (future implementation)
type ModuleSpec struct {
	Name    string // Module name (e.g., "lodash", "moment")
	Version string // Version constraint
	Source  string // "npm", "cdn", "local"
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		Engine:         "goja",
		MemoryLimit:    100 * 1024 * 1024, // 100MB
		StackSize:      1024,              // 1K stack frames
		DefaultTimeout: 30 * time.Second,
		MaxEngines:     5,
		SecurityPolicy: ModerateSandbox,
		PreloadModules: []string{},
	}
}
