package functions

import (
	"context"
	"fmt"

	"defs.dev/schema"
)

// Portal transforms implementation data D into callable Function
type Portal[D any] interface {
	// Apply creates a Function from address, schema, and implementation data
	Apply(address string, schema *schema.FunctionSchema, data D) schema.Function

	// GenerateAddress creates a unique address for the given name and data
	GenerateAddress(name string, data D) string

	// Scheme returns the URI scheme this portal handles (e.g., "local", "https", "postgres")
	Scheme() string

	// ResolveFunction resolves an address back to a callable function (for consumer)
	ResolveFunction(ctx context.Context, address string) (schema.Function, error)
}

// PortalError represents errors in portal operations
type PortalError struct {
	Scheme  string
	Address string
	Message string
	Cause   error
}

func (e *PortalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("portal error for %s address '%s': %s (caused by: %v)",
			e.Scheme, e.Address, e.Message, e.Cause)
	}
	return fmt.Sprintf("portal error for %s address '%s': %s", e.Scheme, e.Address, e.Message)
}

func (e *PortalError) Unwrap() error {
	return e.Cause
}

// RegistryError represents errors in function registry operations
type RegistryError struct {
	Name    string
	Address string
	Type    string // "conflict", "not_found", "invalid"
	Message string
}

func (e *RegistryError) Error() string {
	if e.Address != "" {
		return fmt.Sprintf("registry error (%s) for function '%s' at address '%s': %s",
			e.Type, e.Name, e.Address, e.Message)
	}
	return fmt.Sprintf("registry error (%s) for function '%s': %s", e.Type, e.Name, e.Message)
}

// ConsumerError represents errors in consumer operations
type ConsumerError struct {
	Address string
	Message string
	Cause   error
}

func (e *ConsumerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("consumer error for address '%s': %s (caused by: %v)",
			e.Address, e.Message, e.Cause)
	}
	return fmt.Sprintf("consumer error for address '%s': %s", e.Address, e.Message)
}

func (e *ConsumerError) Unwrap() error {
	return e.Cause
}
