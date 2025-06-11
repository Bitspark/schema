package registry

import "fmt"

// RegistryError represents errors that occur during registry operations
type RegistryError struct {
	Type    string // "not_found", "circular_ref", "invalid_params", etc.
	Name    string
	Message string
}

func (e *RegistryError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("registry error (%s) for '%s': %s", e.Type, e.Name, e.Message)
	}
	return fmt.Sprintf("registry error (%s): %s", e.Type, e.Message)
}

// NewNotFoundError creates an error for when a schema is not found
func NewNotFoundError(name string) *RegistryError {
	return &RegistryError{
		Type:    "not_found",
		Name:    name,
		Message: fmt.Sprintf("schema '%s' not found in registry", name),
	}
}

// NewCircularRefError creates an error for circular references
func NewCircularRefError(path []string) *RegistryError {
	return &RegistryError{
		Type:    "circular_ref",
		Name:    path[len(path)-1],
		Message: fmt.Sprintf("circular reference detected: %v", path),
	}
}

// NewInvalidParamsError creates an error for invalid parameters
func NewInvalidParamsError(name string, missing []string, extra []string) *RegistryError {
	msg := fmt.Sprintf("invalid parameters for schema '%s'", name)
	if len(missing) > 0 {
		msg += fmt.Sprintf(", missing: %v", missing)
	}
	if len(extra) > 0 {
		msg += fmt.Sprintf(", unexpected: %v", extra)
	}
	return &RegistryError{
		Type:    "invalid_params",
		Name:    name,
		Message: msg,
	}
}

// ParameterError represents an error with a specific parameter
type ParameterError struct {
	Parameter string
	Expected  string
	Actual    string
}

func (e *ParameterError) Error() string {
	return fmt.Sprintf("parameter '%s' error: expected %s, got %s", e.Parameter, e.Expected, e.Actual)
}

// CircularReferenceError represents a circular reference in schema resolution
type CircularReferenceError struct {
	Path []string
}

func (e *CircularReferenceError) Error() string {
	return fmt.Sprintf("circular reference detected in path: %v", e.Path)
}
