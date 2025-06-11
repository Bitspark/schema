package javascript

import (
	"context"
	"fmt"
	"time"

	"defs.dev/schema"
	"github.com/dop251/goja"
)

// JavaScriptFunction implements schema.Function for JavaScript functions executed via Goja
type JavaScriptFunction struct {
	name       string
	address    string
	parameters schema.Schema
	returns    schema.Schema
	jsFunction JSFunction
	portal     *JavaScriptPortal
	vm         *goja.Runtime // Cached VM instance
}

// Implement schema.Function interface

func (f *JavaScriptFunction) Name() string {
	return f.name
}

func (f *JavaScriptFunction) Schema() *schema.FunctionSchema {
	// Build function schema from metadata using introspection
	schemaBuilder := schema.NewFunctionSchema().
		Name(f.name).
		Description(fmt.Sprintf("JavaScript function: %s", f.jsFunction.FunctionName))

	// Add inputs from parameters schema using introspection methods
	if obj, ok := f.parameters.(*schema.ObjectSchema); ok {
		for name, inputSchema := range obj.Properties() {
			schemaBuilder = schemaBuilder.Input(name, inputSchema)
		}
		for _, required := range obj.Required() {
			schemaBuilder = schemaBuilder.Required(required)
		}
	}

	if f.returns != nil {
		schemaBuilder = schemaBuilder.Output(f.returns)
	}

	return schemaBuilder.Build().(*schema.FunctionSchema)
}

func (f *JavaScriptFunction) Call(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
	// Validate input parameters - convert FunctionInput to map[string]any for validation
	if result := f.parameters.Validate(params.ToMap()); !result.Valid {
		return schema.FunctionOutput{}, NewValidationError(f.name, "parameter_validation", result.Errors)
	}

	// Get or create VM instance
	vm, err := f.getOrCreateVM()
	if err != nil {
		return schema.FunctionOutput{}, NewExecutionError(f.name, "vm_creation", err)
	}

	// Determine timeout
	timeout := f.portal.config.DefaultTimeout
	if f.jsFunction.Timeout != nil {
		timeout = *f.jsFunction.Timeout
	}

	// Execute JavaScript with timeout
	result, err := f.executeWithTimeout(ctx, vm, params, timeout)
	if err != nil {
		return schema.FunctionOutput{}, err
	}

	// Validate output (internal check)
	if f.returns != nil {
		if validationResult := f.returns.Validate(result); !validationResult.Valid {
			// Log but don't fail - this is internal validation
			logFunctionOutputValidation(f.name, validationResult.Errors)
		}
	}

	return result, nil
}

// getOrCreateVM returns a cached VM or creates a new one
func (f *JavaScriptFunction) getOrCreateVM() (*goja.Runtime, error) {
	if f.vm != nil {
		return f.vm, nil
	}

	// Create new Goja runtime
	vm := goja.New()

	// Apply memory and stack limits (future: implement when Goja supports it)
	// For now, Goja doesn't expose direct memory/stack controls

	// Load and compile JavaScript code
	_, err := vm.RunString(f.jsFunction.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to load JavaScript code: %w", err)
	}

	// Verify the function exists
	fnValue := vm.Get(f.jsFunction.FunctionName)
	if fnValue == nil || goja.IsUndefined(fnValue) {
		return nil, fmt.Errorf("function '%s' not found in JavaScript code", f.jsFunction.FunctionName)
	}

	// Verify it's callable
	if _, ok := goja.AssertFunction(fnValue); !ok {
		return nil, fmt.Errorf("'%s' is not a function in JavaScript code", f.jsFunction.FunctionName)
	}

	// Cache the VM
	f.vm = vm
	return vm, nil
}

// executeWithTimeout executes JavaScript function with timeout handling
func (f *JavaScriptFunction) executeWithTimeout(ctx context.Context, vm *goja.Runtime, params schema.FunctionInput, timeout time.Duration) (schema.FunctionOutput, error) {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Channel for result
	type result struct {
		value schema.FunctionOutput
		err   error
	}
	resultChan := make(chan result, 1)

	// Execute in goroutine to enable timeout
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- result{
					err: fmt.Errorf("javascript function panicked: %v", r),
				}
			}
		}()

		// Convert FunctionInput to map for JavaScript
		paramsMap := params.ToMap()

		// Set parameters in VM
		vm.Set("params", paramsMap)

		// Call the function
		fnValue := vm.Get(f.jsFunction.FunctionName)
		if fnValue == nil {
			resultChan <- result{
				err: fmt.Errorf("function '%s' not found", f.jsFunction.FunctionName),
			}
			return
		}

		// Assert it's a function and call it
		if fn, ok := goja.AssertFunction(fnValue); ok {
			jsResult, err := fn(goja.Undefined(), vm.ToValue(paramsMap))
			if err != nil {
				resultChan <- result{
					err: fmt.Errorf("javascript execution error: %w", err),
				}
				return
			}

			// Convert JavaScript result to Go
			goResult := jsResult.Export()
			resultChan <- result{value: schema.FromAny(goResult)}
		} else {
			resultChan <- result{
				err: fmt.Errorf("'%s' is not a function", f.jsFunction.FunctionName),
			}
		}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		if res.err != nil {
			// Check if it's a goja Exception (JavaScript runtime error)
			if jsErr, ok := res.err.(*goja.Exception); ok {
				return schema.FunctionOutput{}, NewSyntaxError(f.name, jsErr)
			}
			return schema.FunctionOutput{}, NewExecutionError(f.name, "execution", res.err)
		}
		return res.value, nil

	case <-timeoutCtx.Done():
		return schema.FunctionOutput{}, NewTimeoutError(f.name, timeout.String())
	}
}

// Helper functions (similar to local portal)

func buildParametersFromSchema(functionSchema *schema.FunctionSchema) schema.Schema {
	// Build object schema from function inputs
	inputs := functionSchema.Inputs()
	required := functionSchema.Required()

	if len(inputs) == 0 {
		return schema.NewObject().Build()
	}

	builder := schema.NewObject()
	for name, inputSchema := range inputs {
		builder = builder.Property(name, inputSchema)
	}
	builder = builder.Required(required...)

	return builder.Build()
}

func buildReturnsFromSchema(functionSchema *schema.FunctionSchema) schema.Schema {
	return functionSchema.Outputs()
}
