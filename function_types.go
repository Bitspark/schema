package schema

import (
	"fmt"
)

// FunctionInput represents the input parameters to a function call
// Currently wraps map[string]any but can be extended in the future
// to include metadata, validation, or other features
type FunctionInput map[string]any

// FunctionOutput represents the output result from a function call
// Wraps any value but can be extended in the future
// to include metadata, status information, or other features
type FunctionOutput struct {
	value any
}

// NewFunctionInput creates a new FunctionInput from a map
func NewFunctionInput(params map[string]any) FunctionInput {
	if params == nil {
		return make(FunctionInput)
	}
	return FunctionInput(params)
}

// ToMap converts FunctionInput back to map[string]any for backward compatibility
func (fi FunctionInput) ToMap() map[string]any {
	return map[string]any(fi)
}

// Get retrieves a parameter value by name
func (fi FunctionInput) Get(name string) (any, bool) {
	value, exists := fi[name]
	return value, exists
}

// Set sets a parameter value
func (fi FunctionInput) Set(name string, value any) {
	fi[name] = value
}

// Has checks if a parameter exists
func (fi FunctionInput) Has(name string) bool {
	_, exists := fi[name]
	return exists
}

// Keys returns all parameter names
func (fi FunctionInput) Keys() []string {
	keys := make([]string, 0, len(fi))
	for k := range fi {
		keys = append(keys, k)
	}
	return keys
}

// NewFunctionOutput creates a new FunctionOutput from any value
func NewFunctionOutput(result any) FunctionOutput {
	return FunctionOutput{value: result}
}

// Value returns the underlying value of FunctionOutput
func (fo FunctionOutput) Value() any {
	return fo.value
}

// Conversion helpers for backward compatibility

// FromMap converts map[string]any to FunctionInput
func FromMap(params map[string]any) FunctionInput {
	return NewFunctionInput(params)
}

// FromAny converts any value to FunctionOutput
func FromAny(result any) FunctionOutput {
	return NewFunctionOutput(result)
}

// ToAny converts FunctionOutput to any for backward compatibility
func (fo FunctionOutput) ToAny() any {
	return fo.Value()
}

// Conversion functions for seamless interoperability

// ConvertToFunctionInput converts various input types to FunctionInput
func ConvertToFunctionInput(params any) (FunctionInput, error) {
	if params == nil {
		return make(FunctionInput), nil
	}

	switch p := params.(type) {
	case FunctionInput:
		return p, nil
	case map[string]any:
		return NewFunctionInput(p), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to FunctionInput", params)
	}
}

// ConvertToFunctionOutput converts any value to FunctionOutput
func ConvertToFunctionOutput(result any) FunctionOutput {
	if fo, ok := result.(FunctionOutput); ok {
		return fo
	}
	return NewFunctionOutput(result)
}
