package http

import (
	"context"
	"fmt"

	"defs.dev/schema"
)

// Convenience methods for registering reflected functions and services

// RegisterFunction registers a Go function using reflection with the HTTP portal
func (p *HTTPPortal) RegisterFunction(name string, fn any) (Function, error) {
	// Create function reflector
	reflector := schema.NewFunctionReflector(fn)

	// Create handler that uses unified signature
	var handler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return reflector.Call(ctx, params)
	}

	// Generate address
	address, err := p.GenerateAddress(name, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address for function %s: %w", name, err)
	}

	// Register with portal
	return p.Apply(address, reflector.Schema(), handler)
}

// RegisterService registers all methods of a Go service struct using reflection
func (p *HTTPPortal) RegisterService(service any) (map[string]Function, error) {
	// Create service reflector
	serviceReflector := schema.FromService(service)

	// Get all method names and schemas
	schemas := serviceReflector.Schemas()
	functions := make(map[string]Function, len(schemas))

	// Register each method
	for methodName := range schemas {
		// Create handler that uses unified signature
		handler := func(mn string) schema.FunctionHandler {
			return func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
				output, err := serviceReflector.Call(mn, ctx, params)
				if err != nil {
					return schema.FunctionOutput{}, err
				}
				return schema.FromAny(output), nil
			}
		}(methodName)

		// Generate address for this method
		address, err := p.GenerateAddress(methodName, handler)
		if err != nil {
			return nil, fmt.Errorf("failed to generate address for method %s: %w", methodName, err)
		}

		// Register with portal
		function, err := p.Apply(address, schemas[methodName], handler)
		if err != nil {
			return nil, fmt.Errorf("failed to register method %s: %w", methodName, err)
		}

		functions[methodName] = function
	}

	return functions, nil
}

// RegisterFunctions registers multiple Go functions using reflection
func (p *HTTPPortal) RegisterFunctions(functions map[string]any) (map[string]Function, error) {
	results := make(map[string]Function, len(functions))

	for name, fn := range functions {
		function, err := p.RegisterFunction(name, fn)
		if err != nil {
			return nil, fmt.Errorf("failed to register function %s: %w", name, err)
		}
		results[name] = function
	}

	return results, nil
}

// RegisterFromFunctionReflector registers an existing FunctionReflector
func (p *HTTPPortal) RegisterFromFunctionReflector(name string, reflector *schema.FunctionReflector) (Function, error) {
	// Create handler that uses unified signature
	var handler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return reflector.Call(ctx, params)
	}

	// Generate address
	address, err := p.GenerateAddress(name, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address for reflector %s: %w", name, err)
	}

	// Register with portal
	return p.Apply(address, reflector.Schema(), handler)
}

// RegisterFromServiceReflector registers all methods from an existing ServiceReflector
func (p *HTTPPortal) RegisterFromServiceReflector(serviceReflector *schema.ServiceReflector) (map[string]Function, error) {
	// Get all method names and schemas
	schemas := serviceReflector.Schemas()
	functions := make(map[string]Function, len(schemas))

	// Register each method
	for methodName := range schemas {
		// Create handler that uses unified signature
		handler := func(mn string) schema.FunctionHandler {
			return func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
				output, err := serviceReflector.Call(mn, ctx, params)
				if err != nil {
					return schema.FunctionOutput{}, err
				}
				return schema.FromAny(output), nil
			}
		}(methodName)

		// Generate address for this method
		address, err := p.GenerateAddress(methodName, handler)
		if err != nil {
			return nil, fmt.Errorf("failed to generate address for method %s: %w", methodName, err)
		}

		// Register with portal
		function, err := p.Apply(address, schemas[methodName], handler)
		if err != nil {
			return nil, fmt.Errorf("failed to register method %s: %w", methodName, err)
		}

		functions[methodName] = function
	}

	return functions, nil
}

// Convenience method to register schema.Function interface directly
func (p *HTTPPortal) RegisterSchemaFunction(name string, function schema.TypedFunction) (Function, error) {
	// Create handler from the schema.Function
	var handler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return function.Call(ctx, params)
	}

	// Generate address
	address, err := p.GenerateAddress(name, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address for schema function %s: %w", name, err)
	}

	// Register with portal
	return p.Apply(address, function.Schema(), handler)
}
