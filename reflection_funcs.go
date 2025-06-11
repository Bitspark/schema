// Package schema - function reflection utilities
package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// FromFunction generates a FunctionSchema from a Go function using reflection
// Example: userSchema := schema.FromFunction(CreateUser)
func FromFunction(fn any) *FunctionSchema {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf("FromFunction expects a function, got %T", fn))
	}

	return generateSchemaFromFunction(fnType, fn)
}

// generateSchemaFromFunction creates a FunctionSchema from a reflect.Type of a function
func generateSchemaFromFunction(fnType reflect.Type, fn any) *FunctionSchema {
	schema := &FunctionSchema{
		metadata: SchemaMetadata{},
		inputs:   make(map[string]Schema),
	}

	// Try to get function name from runtime
	if fnPtr := reflect.ValueOf(fn).Pointer(); fnPtr != 0 {
		if runtimeFunc := runtime.FuncForPC(fnPtr); runtimeFunc != nil {
			name := runtimeFunc.Name()
			// Clean up the name (remove package path, just keep function name)
			if lastDot := strings.LastIndex(name, "."); lastDot != -1 {
				name = name[lastDot+1:]
			}
			schema.metadata.Name = name
		}
	}

	// Process input parameters
	processInputs(fnType, schema)

	// Process output parameters
	processOutputs(fnType, schema)

	return schema
}

// processInputs analyzes function input parameters and adds them to the schema
func processInputs(fnType reflect.Type, schema *FunctionSchema) {
	paramIndex := 0
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)

		// Skip context.Context parameters as they're infrastructure
		if isContextType(paramType) {
			continue
		}

		// Generate parameter name (Go doesn't preserve parameter names in reflection)
		paramName := fmt.Sprintf("param%d", paramIndex)

		// Generate schema for this parameter type
		paramSchema := generateSchemaFromType(paramType)

		// Add to inputs
		schema.inputs[paramName] = paramSchema

		// For non-pointer types, mark as required
		if paramType.Kind() != reflect.Ptr && !isOptionalType(paramType) {
			schema.required = append(schema.required, paramName)
		}

		paramIndex++
	}
}

// processOutputs analyzes function return values and sets the output schema
func processOutputs(fnType reflect.Type, schema *FunctionSchema) {
	numOut := fnType.NumOut()
	if numOut == 0 {
		return
	}

	// Handle the common Go pattern: (result, error)
	if numOut == 2 && fnType.Out(1).Implements(errorInterface) {
		// First return value is the result
		resultType := fnType.Out(0)
		schema.outputs = generateSchemaFromType(resultType)

		// Second return value is error
		schema.errors = String().
			Description("Error message if the function fails").
			Build()
		return
	}

	// Handle single return value
	if numOut == 1 {
		returnType := fnType.Out(0)

		// If it's just an error, set only errors
		if returnType.Implements(errorInterface) {
			schema.errors = String().
				Description("Error message if the function fails").
				Build()
			return
		}

		// Otherwise it's a result
		schema.outputs = generateSchemaFromType(returnType)
		return
	}

	// Handle multiple return values (not following error convention)
	// Create an object schema with numbered fields
	builder := Object().Name("MultipleReturns")
	for i := 0; i < numOut; i++ {
		returnType := fnType.Out(i)
		fieldName := fmt.Sprintf("result%d", i)
		fieldSchema := generateSchemaFromType(returnType)
		builder.Property(fieldName, fieldSchema)
		builder.Required(fieldName)
	}
	schema.outputs = builder.Build()
}

// Helper functions

// isContextType checks if a type is context.Context
func isContextType(t reflect.Type) bool {
	// Check for context.Context interface by string representation
	// This is more reliable than checking PkgPath and Name separately
	return t.String() == "context.Context" ||
		(t.Kind() == reflect.Interface && t.PkgPath() == "context" && t.Name() == "Context")
}

// isOptionalType checks if a type should be considered optional
func isOptionalType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Interface:
		return true
	default:
		return false
	}
}

// Error interface for checking return types
var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// FunctionReflector wraps a function with its schema for easier use
type FunctionReflector struct {
	fn     reflect.Value
	schema *FunctionSchema
}

// NewFunctionReflector creates a new function reflector
func NewFunctionReflector(fn any) *FunctionReflector {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic(fmt.Sprintf("NewFunctionReflector expects a function, got %T", fn))
	}

	return &FunctionReflector{
		fn:     fnValue,
		schema: FromFunction(fn),
	}
}

// Schema returns the function's schema
func (fr *FunctionReflector) Schema() *FunctionSchema {
	return fr.schema
}

// Call invokes the function with the given parameters (new Function interface)
func (fr *FunctionReflector) Call(ctx context.Context, params FunctionInput) (FunctionOutput, error) {
	result, err := fr.callWithMap(ctx, params.ToMap())
	return NewFunctionOutput(result), err
}

// CallTyped implements typed execution by delegating to Call
func (fr *FunctionReflector) CallTyped(ctx context.Context, input any, output any) error {
	functionInput, err := ConvertToFunctionInput(input)
	if err != nil {
		return err
	}

	result, err := fr.Call(ctx, functionInput)
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

// callWithMap is the internal implementation that works with map parameters
func (fr *FunctionReflector) callWithMap(ctx context.Context, params map[string]any) (any, error) {
	fnType := fr.fn.Type()
	args := make([]reflect.Value, 0, fnType.NumIn())

	// Build argument list
	paramIndex := 0
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)

		// Handle context.Context parameters
		if isContextType(paramType) {
			args = append(args, reflect.ValueOf(ctx))
			continue
		}

		// Get parameter value from params map
		paramName := fmt.Sprintf("param%d", paramIndex)
		paramValue, exists := params[paramName]
		if !exists {
			return nil, fmt.Errorf("missing required parameter: %s", paramName)
		}

		// Convert to appropriate type
		argValue, err := convertToType(paramValue, paramType)
		if err != nil {
			return nil, fmt.Errorf("parameter %s: %w", paramName, err)
		}

		args = append(args, argValue)
		paramIndex++
	}

	// Call the function
	results := fr.fn.Call(args)

	// Handle results according to Go conventions
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		result := results[0]
		if result.Type().Implements(errorInterface) {
			if result.IsNil() {
				return nil, nil
			}
			return nil, result.Interface().(error)
		}
		return result.Interface(), nil
	case 2:
		// Typical (result, error) pattern
		result := results[0]
		errResult := results[1]

		var err error
		if !errResult.IsNil() {
			err = errResult.Interface().(error)
		}

		return result.Interface(), err
	default:
		// Multiple returns - return as slice
		values := make([]any, len(results))
		for i, result := range results {
			values[i] = result.Interface()
		}
		return values, nil
	}
}

// Name returns the function name from schema metadata
func (fr *FunctionReflector) Name() string {
	return fr.schema.metadata.Name
}

// Convenience methods for portal integration

// AsFunction returns this reflector as a Function interface
func (fr *FunctionReflector) AsFunction() TypedFunction {
	return fr
}

// convertToType converts a value to the target reflect.Type
func convertToType(value any, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	valueReflect := reflect.ValueOf(value)

	// Direct assignment if types match
	if valueReflect.Type().AssignableTo(targetType) {
		return valueReflect, nil
	}

	// Try conversion if types are convertible
	if valueReflect.Type().ConvertibleTo(targetType) {
		return valueReflect.Convert(targetType), nil
	}

	// Handle struct conversion from map[string]any
	if targetType.Kind() == reflect.Struct && valueReflect.Kind() == reflect.Map {
		return convertMapToStruct(value, targetType)
	}

	// Handle pointer to struct conversion
	if targetType.Kind() == reflect.Ptr && targetType.Elem().Kind() == reflect.Struct && valueReflect.Kind() == reflect.Map {
		structValue, err := convertMapToStruct(value, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		// Create a pointer to the struct
		ptrValue := reflect.New(targetType.Elem())
		ptrValue.Elem().Set(structValue)
		return ptrValue, nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %v", value, targetType)
}

// convertMapToStruct converts a map[string]any to a struct
func convertMapToStruct(value any, structType reflect.Type) (reflect.Value, error) {
	mapValue, ok := value.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("expected map[string]any, got %T", value)
	}

	// Create new instance of the struct
	structValue := reflect.New(structType).Elem()

	// Iterate over struct fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the field name from json tag, or use the field name
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// Handle json tag like "id,omitempty" -> "id"
			if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
				fieldName = jsonTag[:commaIdx]
			} else {
				fieldName = jsonTag
			}
		}

		// Get value from map
		mapFieldValue, exists := mapValue[fieldName]
		if !exists {
			// Check if field is required or has a default value
			continue
		}

		// Convert and set the field value
		convertedValue, err := convertToType(mapFieldValue, field.Type)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("field %s: %w", fieldName, err)
		}

		fieldValue.Set(convertedValue)
	}

	return structValue, nil
}
