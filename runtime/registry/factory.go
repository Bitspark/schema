package registry

import (
	"context"

	"defs.dev/schema/api"
)

// DefaultFactory implements the api.Factory interface for creating registry components.
type DefaultFactory struct{}

// Ensure DefaultFactory implements the Factory interface at compile time
var _ api.Factory = (*DefaultFactory)(nil)

// NewDefaultFactory creates a new factory instance.
func NewDefaultFactory() *DefaultFactory {
	return &DefaultFactory{}
}

// CreateFunctionRegistry creates a new function registry instance.
func (f *DefaultFactory) CreateFunctionRegistry() api.FunctionRegistry {
	return NewFunctionRegistry()
}

// CreateServiceRegistry creates a new service registry instance.
func (f *DefaultFactory) CreateServiceRegistry() api.ServiceRegistry {
	return NewServiceRegistry()
}

// CreateConsumer creates a new consumer instance.
func (f *DefaultFactory) CreateConsumer() api.Consumer {
	return NewConsumer()
}

// Global factory instance for convenience
var DefaultRegistryFactory = NewDefaultFactory()

// Convenience functions using the default factory
func CreateFunctionRegistry() api.FunctionRegistry {
	return DefaultRegistryFactory.CreateFunctionRegistry()
}

func CreateServiceRegistry() api.ServiceRegistry {
	return DefaultRegistryFactory.CreateServiceRegistry()
}

func CreateConsumer() api.Consumer {
	return DefaultRegistryFactory.CreateConsumer()
}

// Consumer is a simple implementation of api.Consumer
type Consumer struct{}

// Ensure Consumer implements the API interface at compile time
var _ api.Consumer = (*Consumer)(nil)

// NewConsumer creates a new consumer instance.
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
