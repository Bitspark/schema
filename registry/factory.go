package registry

import (
	"context"

	"defs.dev/schema/api"
)

// Factory implements api.Factory for creating registries and other components.
type Factory struct{}

// Ensure Factory implements the API interface at compile time
var _ api.Factory = (*Factory)(nil)

// NewFactory creates a new factory instance.
func NewFactory() *Factory {
	return &Factory{}
}

// CreateRegistry creates a new function registry.
func (f *Factory) CreateRegistry() api.Registry {
	return NewFunctionRegistry()
}

// CreateConsumer creates a new function consumer.
func (f *Factory) CreateConsumer() api.Consumer {
	return NewConsumer()
}

// Consumer implements api.Consumer for executing functions.
type Consumer struct{}

// Ensure Consumer implements the API interface at compile time
var _ api.Consumer = (*Consumer)(nil)

// NewConsumer creates a new consumer.
func NewConsumer() *Consumer {
	return &Consumer{}
}

// Consume executes a function with the given input.
func (c *Consumer) Consume(ctx context.Context, fn api.Function, input any) (any, error) {
	// Convert input to FunctionData interface
	var params api.FunctionData

	switch v := input.(type) {
	case api.FunctionData:
		params = v
	case map[string]any:
		params = FunctionInputMap(v)
	default:
		// For other types, create a single-value input
		params = FunctionInputMap{"value": input}
	}

	output, err := fn.Call(ctx, params)
	if err != nil {
		return nil, err
	}

	return output.ToAny(), nil
}

// FunctionInputMap implements api.FunctionData
type FunctionInputMap map[string]any

// Ensure FunctionInputMap implements the API interface at compile time
var _ api.FunctionData = (FunctionInputMap)(nil)

func (f FunctionInputMap) ToMap() map[string]any {
	return map[string]any(f)
}

func (f FunctionInputMap) Get(name string) (any, bool) {
	value, exists := f[name]
	return value, exists
}

func (f FunctionInputMap) Set(name string, value any) {
	f[name] = value
}

func (f FunctionInputMap) Has(name string) bool {
	_, exists := f[name]
	return exists
}

func (f FunctionInputMap) Keys() []string {
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	return keys
}

func (f FunctionInputMap) Value() any {
	return map[string]any(f)
}

func (f FunctionInputMap) ToAny() any {
	return map[string]any(f)
}
