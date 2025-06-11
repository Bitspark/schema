// Package schema - service reflection utilities
package schema

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// FromService generates a ServiceReflector from a struct instance with methods
// Example: serviceReflector := schema.FromService(userService)
func FromService(instance any) *ServiceReflector {
	if instance == nil {
		panic("FromService expects a non-nil instance")
	}

	instanceValue := reflect.ValueOf(instance)
	instanceType := instanceValue.Type()

	// Handle pointer to struct
	if instanceType.Kind() == reflect.Ptr {
		if instanceValue.IsNil() {
			panic("FromService expects a non-nil pointer")
		}
		// Keep the pointer type for method discovery, but note the element type
		instanceType = instanceValue.Type()
	} else if instanceType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("FromService expects a struct or pointer to struct, got %T", instance))
	}

	return generateServiceReflector(instanceValue, instanceType)
}

// ServiceReflector wraps a struct instance and provides access to its methods as functions
type ServiceReflector struct {
	instance      reflect.Value
	instanceType  reflect.Type
	methods       map[string]*BoundMethod
	serviceSchema *ObjectSchema
}

// BoundMethod represents a method bound to a specific struct instance
type BoundMethod struct {
	reflector    *FunctionReflector
	method       reflect.Method
	methodName   string
	originalName string
}

// generateServiceReflector creates a ServiceReflector from a struct instance
func generateServiceReflector(instanceValue reflect.Value, instanceType reflect.Type) *ServiceReflector {
	sr := &ServiceReflector{
		instance:     instanceValue,
		instanceType: instanceType,
		methods:      make(map[string]*BoundMethod),
	}

	// Discover and bind all valid methods
	methods := discoverMethods(instanceType)
	for methodName, method := range methods {
		boundMethod := createBoundMethod(instanceValue, method)
		sr.methods[methodName] = boundMethod
	}

	return sr
}

// discoverMethods finds all exported methods on a type that are suitable for service reflection
func discoverMethods(instanceType reflect.Type) map[string]reflect.Method {
	methods := make(map[string]reflect.Method)

	// Check methods on the type itself
	for i := 0; i < instanceType.NumMethod(); i++ {
		method := instanceType.Method(i)

		// Only include exported methods that are valid service methods
		if method.IsExported() && isValidServiceMethod(method) {
			methods[method.Name] = method
		}
	}

	// If it's not a pointer type, also check methods on the pointer type
	if instanceType.Kind() != reflect.Ptr {
		ptrType := reflect.PtrTo(instanceType)
		for i := 0; i < ptrType.NumMethod(); i++ {
			method := ptrType.Method(i)

			// Only include exported methods that are valid service methods
			if method.IsExported() && isValidServiceMethod(method) && methods[method.Name].Name == "" {
				methods[method.Name] = method
			}
		}
	}

	return methods
}

// isValidServiceMethod checks if a method is suitable for service reflection
func isValidServiceMethod(method reflect.Method) bool {
	methodType := method.Type

	// Must have at least receiver
	if methodType.NumIn() < 1 {
		return false
	}

	// Skip methods that return functions or channels (likely not service methods)
	for i := 0; i < methodType.NumOut(); i++ {
		out := methodType.Out(i)
		if out.Kind() == reflect.Func || out.Kind() == reflect.Chan {
			return false
		}
	}

	// Skip common non-service methods
	switch method.Name {
	case "String", "GoString", "Error", "Format":
		return false
	}

	return true
}

// createBoundMethod creates a BoundMethod by binding a method to its instance
func createBoundMethod(instance reflect.Value, method reflect.Method) *BoundMethod {
	methodType := method.Type
	methodValue := method.Func

	// Create a wrapper function that binds the receiver
	// Original signature: func(receiver ReceiverType, param1 Type1, ...) (RetType, error)
	// Bound signature:    func(param1 Type1, ...) (RetType, error)

	boundFunc := func(args []reflect.Value) []reflect.Value {
		// Handle value vs pointer receiver
		receiver := instance

		// If the method expects a pointer receiver but we have a value, take address
		if methodType.In(0).Kind() == reflect.Ptr && instance.Kind() != reflect.Ptr {
			if instance.CanAddr() {
				receiver = instance.Addr()
			} else {
				// Create a new addressable value
				newInstance := reflect.New(instance.Type())
				newInstance.Elem().Set(instance)
				receiver = newInstance
			}
		}

		// Prepend the receiver (instance) to the arguments
		fullArgs := append([]reflect.Value{receiver}, args...)
		return methodValue.Call(fullArgs)
	}

	// Create input types (excluding receiver)
	var inputTypes []reflect.Type
	for i := 1; i < methodType.NumIn(); i++ { // Skip receiver at index 0
		inputTypes = append(inputTypes, methodType.In(i))
	}

	// Create output types
	var outputTypes []reflect.Type
	for i := 0; i < methodType.NumOut(); i++ {
		outputTypes = append(outputTypes, methodType.Out(i))
	}

	// Create a synthetic function type for reflection
	syntheticFunc := createSyntheticFunction(boundFunc, inputTypes, outputTypes, method.Name)

	// Use existing function reflection on the synthetic function
	reflector := NewFunctionReflector(syntheticFunc)

	// Override the function name in the schema metadata
	schema := reflector.Schema()
	schema.metadata.Name = method.Name

	return &BoundMethod{
		reflector:    reflector,
		method:       method,
		methodName:   method.Name,
		originalName: method.Name,
	}
}

// createSyntheticFunction creates a function value that can be used with function reflection
func createSyntheticFunction(boundFunc func([]reflect.Value) []reflect.Value, inputTypes []reflect.Type, outputTypes []reflect.Type, methodName string) any {
	// Create function type
	funcType := reflect.FuncOf(inputTypes, outputTypes, false)

	// Create function value
	funcValue := reflect.MakeFunc(funcType, boundFunc)

	return funcValue.Interface()
}

// Core ServiceReflector methods

// Functions returns a map of method names to their FunctionReflector instances
func (sr *ServiceReflector) Functions() map[string]*FunctionReflector {
	functions := make(map[string]*FunctionReflector)
	for name, boundMethod := range sr.methods {
		functions[name] = boundMethod.reflector
	}
	return functions
}

// Schemas returns a map of method names to their FunctionSchema instances
func (sr *ServiceReflector) Schemas() map[string]*FunctionSchema {
	schemas := make(map[string]*FunctionSchema)
	for name, boundMethod := range sr.methods {
		schemas[name] = boundMethod.reflector.Schema()
	}
	return schemas
}

// MethodNames returns a list of all available method names
func (sr *ServiceReflector) MethodNames() []string {
	names := make([]string, 0, len(sr.methods))
	for name := range sr.methods {
		names = append(names, name)
	}
	return names
}

// Call executes a method by name with flexible parameter format
func (sr *ServiceReflector) Call(methodName string, ctx context.Context, input FunctionInput) (any, error) {
	boundMethod, exists := sr.methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	// Delegate to the FunctionReflector.Call
	output, err := boundMethod.reflector.Call(ctx, input)
	if err != nil {
		return nil, err
	}

	// Extract the actual value from FunctionOutput for backward compatibility
	return output.Value(), nil
}

// ServiceSchema returns the schema of the service struct itself (its fields)
func (sr *ServiceReflector) ServiceSchema() *ObjectSchema {
	if sr.serviceSchema == nil {
		// Use existing struct reflection to analyze the service struct
		structType := sr.instanceType
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}

		schema := generateSchemaFromType(structType)
		if objSchema, ok := schema.(*ObjectSchema); ok {
			sr.serviceSchema = objSchema
		} else {
			// Fallback - create empty object schema
			sr.serviceSchema = Object().Name(structType.Name()).Build().(*ObjectSchema)
		}
	}
	return sr.serviceSchema
}

// ServiceType returns the reflect.Type of the service
func (sr *ServiceReflector) ServiceType() reflect.Type {
	return sr.instanceType
}

// Convenience methods for portal integration

// AsFunctions returns all service methods as Function interface instances
func (sr *ServiceReflector) AsFunctions() map[string]TypedFunction {
	functions := make(map[string]TypedFunction)
	for name, boundMethod := range sr.methods {
		functions[name] = boundMethod.reflector.AsFunction()
	}
	return functions
}

// ToJSONSchemas converts all method schemas to JSON Schema format
func (sr *ServiceReflector) ToJSONSchemas() map[string]map[string]any {
	schemas := make(map[string]map[string]any)

	for methodName, boundMethod := range sr.methods {
		schema := boundMethod.reflector.Schema()
		schemas[methodName] = schema.ToJSONSchema()
	}

	return schemas
}

// HasMethod checks if a method exists on the service
func (sr *ServiceReflector) HasMethod(methodName string) bool {
	_, exists := sr.methods[methodName]
	return exists
}

// GetMethod returns a specific BoundMethod by name
func (sr *ServiceReflector) GetMethod(methodName string) (*BoundMethod, bool) {
	method, exists := sr.methods[methodName]
	return method, exists
}

// ServiceInfo returns comprehensive information about the service
func (sr *ServiceReflector) ServiceInfo() *ServiceInfo {
	methodInfo := make(map[string]*MethodInfo)

	for name, boundMethod := range sr.methods {
		schema := boundMethod.reflector.Schema()
		methodInfo[name] = &MethodInfo{
			Name:        name,
			Schema:      schema,
			InputCount:  len(schema.Inputs()),
			OutputCount: getOutputCount(schema),
			HasError:    schema.Errors() != nil,
		}
	}

	return &ServiceInfo{
		Name:        sr.getServiceName(),
		Type:        sr.instanceType,
		MethodCount: len(sr.methods),
		Methods:     methodInfo,
		Schema:      sr.ServiceSchema(),
	}
}

// ServiceInfo provides comprehensive information about a service
type ServiceInfo struct {
	Name        string                 `json:"name"`
	Type        reflect.Type           `json:"-"`
	MethodCount int                    `json:"method_count"`
	Methods     map[string]*MethodInfo `json:"methods"`
	Schema      *ObjectSchema          `json:"schema"`
}

// MethodInfo provides information about a specific method
type MethodInfo struct {
	Name        string          `json:"name"`
	Schema      *FunctionSchema `json:"schema"`
	InputCount  int             `json:"input_count"`
	OutputCount int             `json:"output_count"`
	HasError    bool            `json:"has_error"`
}

// Helper methods

// getServiceName extracts a service name from the type
func (sr *ServiceReflector) getServiceName() string {
	typeName := ""

	// Get the underlying struct type name
	if sr.instanceType.Kind() == reflect.Ptr {
		typeName = sr.instanceType.Elem().Name()
	} else {
		typeName = sr.instanceType.Name()
	}

	// Convert from PascalCase to snake_case for service name
	return toSnakeCase(typeName)
}

// getOutputCount counts the number of outputs (excluding error)
func getOutputCount(schema *FunctionSchema) int {
	if schema.Outputs() == nil {
		return 0
	}
	// For now, count as 1 output (could be enhanced to count object properties)
	return 1
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r | 0x20) // Convert to lowercase
	}

	return result.String()
}

// Utility functions for working with ServiceReflector

// CallMethod is a convenience function for calling service methods
func CallMethod(service any, methodName string, ctx context.Context, params map[string]any) (any, error) {
	serviceReflector := FromService(service)
	input := NewFunctionInput(params)
	return serviceReflector.Call(methodName, ctx, input)
}

// GetMethodSchema is a convenience function for getting a method's schema
func GetMethodSchema(service any, methodName string) (*FunctionSchema, error) {
	serviceReflector := FromService(service)

	if !serviceReflector.HasMethod(methodName) {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	schemas := serviceReflector.Schemas()
	return schemas[methodName], nil
}

// ListMethods is a convenience function for listing all methods on a service
func ListMethods(service any) []string {
	serviceReflector := FromService(service)
	return serviceReflector.MethodNames()
}

// ValidateService checks if a service instance is valid for reflection
func ValidateService(service any) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	serviceType := reflect.TypeOf(service)

	// Check if it's a struct or pointer to struct
	if serviceType.Kind() == reflect.Ptr {
		if reflect.ValueOf(service).IsNil() {
			return fmt.Errorf("service pointer cannot be nil")
		}
		serviceType = serviceType.Elem()
	}

	if serviceType.Kind() != reflect.Struct {
		return fmt.Errorf("service must be a struct or pointer to struct, got %T", service)
	}

	// Check if it has any exportable methods
	methodCount := 0

	// Check methods on the type itself
	actualType := reflect.TypeOf(service)
	for i := 0; i < actualType.NumMethod(); i++ {
		method := actualType.Method(i)
		if method.IsExported() && isValidServiceMethod(method) {
			methodCount++
		}
	}

	// If it's not a pointer type, also check methods on the pointer type
	if serviceType.Kind() != reflect.Ptr {
		ptrType := reflect.PointerTo(serviceType)
		for i := 0; i < ptrType.NumMethod(); i++ {
			method := ptrType.Method(i)
			if method.IsExported() && isValidServiceMethod(method) {
				methodCount++
			}
		}
	}

	if methodCount == 0 {
		return fmt.Errorf("service %s has no valid exportable methods", serviceType.Name())
	}

	return nil
}
