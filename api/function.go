package api

import (
	"context"

	"defs.dev/schema/api/core"
)

// FunctionData represents unified data for function inputs and outputs.
// This provides symmetry with the schema system and simplifies the API.
type FunctionData interface {
	// Map-like operations for structured data access
	ToMap() map[string]any
	Get(name string) (any, bool)
	Set(name string, value any)
	Has(name string) bool
	Keys() []string

	// Value operations for direct data access
	Value() any
	ToAny() any
}

// Function interface defines the contract for callable functions.
type Function interface {
	Call(ctx context.Context, params FunctionData) (FunctionData, error)
	Schema() core.FunctionSchema
	Name() string
}

// FunctionInputMap provides a concrete implementation interface for FunctionInput.
type FunctionInputMap map[string]any

// FunctionOutputValue provides a concrete implementation interface for FunctionOutput.
type FunctionOutputValue struct {
	value any
}

// Consumer defines the interface for consuming/executing functions.
type Consumer interface {
	Consume(ctx context.Context, fn Function, input any) (any, error)
}
