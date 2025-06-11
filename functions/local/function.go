package local

import (
	"context"
	"encoding/json"
	"fmt"

	"defs.dev/schema"
)

// LocalFunction - schema-validated local Go function implementation
type LocalFunction struct {
	name        string
	description string
	parameters  schema.Schema
	returns     schema.Schema // For internal validation
	handler     schema.FunctionHandler
	examples    []FunctionExample
	tags        []string
}

// FunctionHandler is an alias for schema.FunctionHandler
type FunctionHandler = schema.FunctionHandler

type FunctionExample struct {
	Input       any    `json:"input"`
	Output      any    `json:"output"`
	Description string `json:"description"`
}

// Implement Function interface for LocalFunction
func (f *LocalFunction) Name() string {
	return f.name
}

func (f *LocalFunction) Schema() *schema.FunctionSchema {
	// Build function schema from local function metadata
	schemaBuilder := schema.NewFunctionSchema().
		Description(f.description).
		Name(f.name)

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

// Call implements Function interface - execute with schema validation
func (f *LocalFunction) Call(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
	// Validate input parameters - convert FunctionInput to map[string]any for validation
	if result := f.parameters.Validate(params.ToMap()); !result.Valid {
		return schema.FunctionOutput{}, &FunctionError{
			Function: f.name,
			Stage:    "parameter_validation",
			Errors:   result.Errors,
		}
	}

	// Execute handler
	output, err := f.handler(ctx, params)
	if err != nil {
		return schema.FunctionOutput{}, &FunctionError{
			Function: f.name,
			Stage:    "execution",
			Cause:    err,
		}
	}

	// Validate output (internal check)
	if f.returns != nil {
		if result := f.returns.Validate(output.Value()); !result.Valid {
			// Log but don't fail - this is internal validation
			logFunctionOutputValidation(f.name, result.Errors)
		}
	}

	return output, nil
}

// CallTyped implements typed execution by delegating to Call
func (f *LocalFunction) CallTyped(ctx context.Context, input any, output any) error {
	functionInput, err := schema.ConvertToFunctionInput(input)
	if err != nil {
		return err
	}

	result, err := f.Call(ctx, functionInput)
	if err != nil {
		return err
	}

	// Marshal and unmarshal for type conversion
	jsonData, err := json.Marshal(result.Value())
	if err != nil {
		return fmt.Errorf("result marshaling failed: %w", err)
	}

	if err := json.Unmarshal(jsonData, output); err != nil {
		return fmt.Errorf("result unmarshaling failed: %w", err)
	}

	return nil
}

// Local function builder
func NewLocalFunction(name string) *LocalFunctionBuilder {
	return &LocalFunctionBuilder{
		function: &LocalFunction{name: name},
	}
}

type LocalFunctionBuilder struct {
	function *LocalFunction
}

func (b *LocalFunctionBuilder) Description(desc string) *LocalFunctionBuilder {
	b.function.description = desc
	return b
}

func (b *LocalFunctionBuilder) Parameters(schema schema.Schema) *LocalFunctionBuilder {
	b.function.parameters = schema
	return b
}

func (b *LocalFunctionBuilder) Returns(schema schema.Schema) *LocalFunctionBuilder {
	b.function.returns = schema
	return b
}

func (b *LocalFunctionBuilder) Handler(handler schema.FunctionHandler) *LocalFunctionBuilder {
	b.function.handler = handler
	return b
}

func (b *LocalFunctionBuilder) Example(input, output any, description string) *LocalFunctionBuilder {
	b.function.examples = append(b.function.examples, FunctionExample{
		Input:       input,
		Output:      output,
		Description: description,
	})
	return b
}

func (b *LocalFunctionBuilder) Tag(tag string) *LocalFunctionBuilder {
	b.function.tags = append(b.function.tags, tag)
	return b
}

func (b *LocalFunctionBuilder) Build() schema.TypedFunction {
	if b.function.parameters == nil {
		panic("LocalFunction parameters schema is required")
	}
	if b.function.handler == nil {
		panic("LocalFunction handler is required")
	}
	return b.function
}

// FunctionError represents errors during function execution
type FunctionError struct {
	Function string
	Stage    string
	Message  string
	Errors   []schema.ValidationError
	Cause    error
}

func (e *FunctionError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("function '%s' failed at %s: %s", e.Function, e.Stage, e.Message)
	}

	if len(e.Errors) > 0 {
		// Build detailed error message showing all validation errors
		errorDetails := make([]string, len(e.Errors))
		for i, validationErr := range e.Errors {
			if validationErr.Path != "" {
				errorDetails[i] = fmt.Sprintf("'%s': %s", validationErr.Path, validationErr.Message)
			} else {
				errorDetails[i] = validationErr.Message
			}

			// Add suggestion if available
			if validationErr.Suggestion != "" {
				errorDetails[i] += fmt.Sprintf(" (suggestion: %s)", validationErr.Suggestion)
			}
		}

		if len(e.Errors) == 1 {
			return fmt.Sprintf("function '%s' validation failed at %s: %s", e.Function, e.Stage, errorDetails[0])
		} else {
			var details string
			for i, detail := range errorDetails {
				details += fmt.Sprintf("\n  %d. %s", i+1, detail)
			}
			return fmt.Sprintf("function '%s' validation failed at %s (%d errors):%s", e.Function, e.Stage, len(e.Errors), details)
		}
	}

	if e.Cause != nil {
		return fmt.Sprintf("function '%s' failed at %s: %v", e.Function, e.Stage, e.Cause)
	}

	return fmt.Sprintf("function '%s' failed at %s", e.Function, e.Stage)
}

func (e *FunctionError) Unwrap() error {
	return e.Cause
}

// Helper function for logging output validation (placeholder)
func logFunctionOutputValidation(functionName string, errors []schema.ValidationError) {
	// Provide detailed error reporting for function output validation issues
	if len(errors) == 0 {
		return
	}

	fmt.Printf("WARNING: Function '%s' output validation failed:\n", functionName)
	for i, err := range errors {
		var errorDetail string
		if err.Path != "" {
			errorDetail = fmt.Sprintf("'%s': %s", err.Path, err.Message)
		} else {
			errorDetail = err.Message
		}

		// Add suggestion if available
		if err.Suggestion != "" {
			errorDetail += fmt.Sprintf(" (suggestion: %s)", err.Suggestion)
		}

		// Add value information if available
		if err.Value != nil {
			errorDetail += fmt.Sprintf(" [got: %v]", err.Value)
		}

		// Add expected information if available
		if err.Expected != "" {
			errorDetail += fmt.Sprintf(" [expected: %s]", err.Expected)
		}

		fmt.Printf("  %d. %s\n", i+1, errorDetail)
	}
}
