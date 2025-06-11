package api

import "context"

// FunctionHandler represents a local function that can be served as HTTP endpoint.
type FunctionHandler func(ctx context.Context, params FunctionInput) (FunctionOutput, error)

// TypedFunction interface represents a universal callable function concept with type safety.
type TypedFunction interface {
	Function
	CallTyped(ctx context.Context, input any, output any) error
}

// Function interface defines the contract for callable functions.
type Function interface {
	Call(ctx context.Context, params FunctionInput) (FunctionOutput, error)
	Schema() FunctionSchema
	Name() string
}

// FunctionInput represents input parameters to a function.
type FunctionInput interface {
	ToMap() map[string]any
	Get(name string) (any, bool)
	Set(name string, value any)
	Has(name string) bool
	Keys() []string
}

// FunctionOutput represents the output/result of a function call.
type FunctionOutput interface {
	Value() any
	ToAny() any
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

// Portal defines a generic interface for function execution portals.
type Portal[D any] interface {
	Execute(ctx context.Context, data D) error
	Close() error
}
