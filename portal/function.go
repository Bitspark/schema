package portal

import (
	"context"
	"fmt"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
)

// FunctionDataMap implements api.FunctionData as a unified map-based data structure
type FunctionDataMap map[string]any

// NewFunctionData creates a new FunctionData from a map
func NewFunctionData(data map[string]any) api.FunctionData {
	if data == nil {
		data = make(map[string]any)
	}
	return FunctionDataMap(data)
}

// NewFunctionDataValue creates FunctionData from a single value
func NewFunctionDataValue(value any) api.FunctionData {
	if m, ok := value.(map[string]any); ok {
		return FunctionDataMap(m)
	}
	return &FunctionDataValue{value: value}
}

// Map-like operations for FunctionDataMap

// ToMap returns the data as a map
func (f FunctionDataMap) ToMap() map[string]any {
	result := make(map[string]any)
	for k, v := range f {
		result[k] = v
	}
	return result
}

// Get retrieves a parameter value
func (f FunctionDataMap) Get(name string) (any, bool) {
	value, exists := f[name]
	return value, exists
}

// Set sets a parameter value
func (f FunctionDataMap) Set(name string, value any) {
	f[name] = value
}

// Has checks if a parameter exists
func (f FunctionDataMap) Has(name string) bool {
	_, exists := f[name]
	return exists
}

// Keys returns all parameter names
func (f FunctionDataMap) Keys() []string {
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	return keys
}

// Value operations for FunctionDataMap

// Value returns the entire map as the value
func (f FunctionDataMap) Value() any {
	return map[string]any(f)
}

// ToAny returns the map as any
func (f FunctionDataMap) ToAny() any {
	return map[string]any(f)
}

// FunctionDataValue implements api.FunctionData for single values
type FunctionDataValue struct {
	value any
}

// Map-like operations for FunctionDataValue (treat as single-key map)

// ToMap returns the value as a single-key map
func (f *FunctionDataValue) ToMap() map[string]any {
	return map[string]any{"value": f.value}
}

// Get retrieves the value if name is "value", otherwise returns false
func (f *FunctionDataValue) Get(name string) (any, bool) {
	if name == "value" {
		return f.value, true
	}
	return nil, false
}

// Set sets the value if name is "value"
func (f *FunctionDataValue) Set(name string, value any) {
	if name == "value" {
		f.value = value
	}
}

// Has checks if name is "value"
func (f *FunctionDataValue) Has(name string) bool {
	return name == "value"
}

// Keys returns ["value"]
func (f *FunctionDataValue) Keys() []string {
	return []string{"value"}
}

// Value operations for FunctionDataValue

// Value returns the stored value
func (f *FunctionDataValue) Value() any {
	return f.value
}

// ToAny returns the stored value
func (f *FunctionDataValue) ToAny() any {
	return f.value
}

// Legacy types for backward compatibility

// FunctionInputMap implements core.FunctionInput as a map (deprecated)
type FunctionInputMap map[string]any

// NewFunctionInputMap creates a new FunctionInputMap (deprecated, use NewFunctionData)
func NewFunctionInputMap(data map[string]any) FunctionInputMap {
	if data == nil {
		data = make(map[string]any)
	}
	return FunctionInputMap(data)
}

// ToMap returns the input as a map
func (f FunctionInputMap) ToMap() map[string]any {
	result := make(map[string]any)
	for k, v := range f {
		result[k] = v
	}
	return result
}

// Get retrieves a parameter value
func (f FunctionInputMap) Get(name string) (any, bool) {
	value, exists := f[name]
	return value, exists
}

// Set sets a parameter value
func (f FunctionInputMap) Set(name string, value any) {
	f[name] = value
}

// Has checks if a parameter exists
func (f FunctionInputMap) Has(name string) bool {
	_, exists := f[name]
	return exists
}

// Keys returns all parameter names
func (f FunctionInputMap) Keys() []string {
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	return keys
}

// RemoteFunction represents a function accessible via a portal
type RemoteFunction struct {
	name    string
	schema  core.FunctionSchema
	address api.Address
	portal  api.FunctionPortal
}

// NewRemoteFunction creates a new RemoteFunction
func NewRemoteFunction(name string, schema core.FunctionSchema, address api.Address, portal api.FunctionPortal) api.Function {
	return &RemoteFunction{
		name:    name,
		schema:  schema,
		address: address,
		portal:  portal,
	}
}

// Call executes the remote function via the portal
func (f *RemoteFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	// This would typically involve network communication
	// For now, we delegate back to the portal's resolve mechanism
	resolved, err := f.portal.ResolveFunction(ctx, f.address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote function: %w", err)
	}

	return resolved.Call(ctx, params)
}

// Schema returns the function schema
func (f *RemoteFunction) Schema() core.FunctionSchema {
	return f.schema
}

// Name returns the function name
func (f *RemoteFunction) Name() string {
	return f.name
}

// Address returns the function address
func (f *RemoteFunction) Address() api.Address {
	return f.address
}

// ServiceImpl implements api.Service
type ServiceImpl struct {
	name        string
	description string
	schema      core.ServiceSchema
	methods     map[string]api.Function
}

// NewService creates a new Service
func NewService(name string, schema core.ServiceSchema) api.Service {
	return &ServiceImpl{
		name:    name,
		schema:  schema,
		methods: make(map[string]api.Function),
	}
}

// Name returns the service name
func (s *ServiceImpl) Name() string {
	return s.name
}

// Description returns the service description
func (s *ServiceImpl) Description() string {
	return s.description
}

// Schema returns the service schema
func (s *ServiceImpl) Schema() core.ServiceSchema {
	return s.schema
}

// Methods returns all method names
func (s *ServiceImpl) Methods() []string {
	methods := make([]string, 0, len(s.methods))
	for name := range s.methods {
		methods = append(methods, name)
	}
	return methods
}

// GetMethod returns a method function by name
func (s *ServiceImpl) GetMethod(name string) (api.Function, bool) {
	method, exists := s.methods[name]
	return method, exists
}

// AddMethod adds a method to the service
func (s *ServiceImpl) AddMethod(name string, function api.Function) {
	s.methods[name] = function
}

// FunctionCallImpl implements the api.FunctionCall interface
type FunctionCallImpl struct {
	FunctionName string
	Address      api.Address
	Input        api.FunctionData
	Output       api.FunctionData
	Error        error
	Timestamp    int64
}

// NewFunctionCall creates a new FunctionCall record
func NewFunctionCall(functionName string, address api.Address, input api.FunctionData, output api.FunctionData, err error, timestamp int64) api.FunctionCall {
	// Convert old types to new FunctionData types
	var inputData api.FunctionData
	if input != nil {
		inputData = NewFunctionData(input.ToMap())
	}

	var outputData api.FunctionData
	if output != nil {
		outputData = NewFunctionDataValue(output.Value())
	}

	return api.FunctionCall{
		FunctionName: functionName,
		Address:      address,
		Input:        inputData,
		Output:       outputData,
		Error:        err,
		Timestamp:    timestamp,
	}
}
