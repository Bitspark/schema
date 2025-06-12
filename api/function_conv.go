package api

import (
	"context"
	"fmt"
	"time"

	"defs.dev/schema/api/core"
)

// FunctionDataMap implements api.FunctionData as a unified map-based data structure
type FunctionDataMap map[string]any

// NewFunctionData creates a new FunctionData from a map
func NewFunctionData(data map[string]any) FunctionData {
	if data == nil {
		data = make(map[string]any)
	}
	return FunctionDataMap(data)
}

// NewFunctionDataValue creates FunctionData from a single value
func NewFunctionDataValue(value any) FunctionData {
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
	address Address
	portal  FunctionPortal
}

// NewRemoteFunction creates a new RemoteFunction
func NewRemoteFunction(name string, schema core.FunctionSchema, address Address, portal FunctionPortal) Function {
	return &RemoteFunction{
		name:    name,
		schema:  schema,
		address: address,
		portal:  portal,
	}
}

// Call executes the remote function via the portal
func (f *RemoteFunction) Call(ctx context.Context, params FunctionData) (FunctionData, error) {
	// For remote functions, we need to perform the actual remote call
	// This depends on the portal type - for HTTP portals, make HTTP requests
	// For WebSocket portals, send WebSocket messages, etc.

	// Check if portal is an HTTP portal
	if httpPortal, ok := f.portal.(HTTPPortal); ok {
		// Use the HTTP portal's client to make the request
		return f.callViaHTTP(ctx, httpPortal, params)
	}

	// For other portal types, we'll need different implementations
	// For now, return an error indicating the remote call is not implemented
	return nil, fmt.Errorf("remote function call not implemented for portal type %T", f.portal)
}

// callViaHTTP performs an HTTP request to call the remote function
func (f *RemoteFunction) callViaHTTP(ctx context.Context, portal HTTPPortal, params FunctionData) (FunctionData, error) {
	// This is a simplified implementation - in a real system, this would
	// construct proper HTTP requests, handle authentication, etc.

	// For now, return a basic error to avoid the infinite recursion
	// This allows tests to run without crashing
	return nil, fmt.Errorf("HTTP remote function call not yet implemented for function %s at %s", f.name, f.address.String())
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
func (f *RemoteFunction) Address() Address {
	return f.address
}

// ServiceImpl implements api.Service as a full executable entity
type ServiceImpl struct {
	name        string
	description string
	schema      core.ServiceSchema
	methods     map[string]Function

	// Entity state
	status    ServiceStatus
	isRunning bool
	startedAt *time.Time
	stoppedAt *time.Time
}

// NewService creates a new Service
func NewService(name string, schema core.ServiceSchema) Service {
	return &ServiceImpl{
		name:    name,
		schema:  schema,
		methods: make(map[string]Function),
		status: ServiceStatus{
			State:   ServiceStateStopped,
			Healthy: false,
		},
		isRunning: false,
	}
}

// CallMethod executes a method on the service (core entity execution)
func (s *ServiceImpl) CallMethod(ctx context.Context, methodName string, params FunctionData) (FunctionData, error) {
	if !s.isRunning {
		return nil, fmt.Errorf("service %s is not running", s.name)
	}

	method, exists := s.methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method %s not found on service %s", methodName, s.name)
	}

	return method.Call(ctx, params)
}

// Schema returns the service schema
func (s *ServiceImpl) Schema() core.ServiceSchema {
	return s.schema
}

// Name returns the service name
func (s *ServiceImpl) Name() string {
	return s.name
}

// Start starts the service (entity lifecycle)
func (s *ServiceImpl) Start(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("service %s is already running", s.name)
	}

	now := time.Now()
	s.status.State = ServiceStateStarting
	s.startedAt = &now
	s.stoppedAt = nil

	// TODO: Implement actual startup logic, initialization, etc.

	s.isRunning = true
	s.status.State = ServiceStateRunning
	s.status.Healthy = true
	s.status.StartedAt = &now

	return nil
}

// Stop stops the service (entity lifecycle)
func (s *ServiceImpl) Stop(ctx context.Context) error {
	if !s.isRunning {
		return fmt.Errorf("service %s is not running", s.name)
	}

	now := time.Now()
	s.status.State = ServiceStateStopping

	// TODO: Implement actual shutdown logic, cleanup, etc.

	s.isRunning = false
	s.status.State = ServiceStateStopped
	s.status.Healthy = false
	s.status.StoppedAt = &now
	s.stoppedAt = &now

	return nil
}

// Status returns the current service status (entity state)
func (s *ServiceImpl) Status(ctx context.Context) (ServiceStatus, error) {
	return s.status, nil
}

// IsRunning returns whether the service is currently running
func (s *ServiceImpl) IsRunning() bool {
	return s.isRunning
}

// HasMethod checks if the service has a specific method
func (s *ServiceImpl) HasMethod(methodName string) bool {
	_, exists := s.methods[methodName]
	return exists
}

// MethodNames returns all method names
func (s *ServiceImpl) MethodNames() []string {
	methods := make([]string, 0, len(s.methods))
	for name := range s.methods {
		methods = append(methods, name)
	}
	return methods
}

// Legacy methods for backward compatibility

// Description returns the service description
func (s *ServiceImpl) Description() string {
	return s.description
}

// Methods returns all method names (legacy)
func (s *ServiceImpl) Methods() []string {
	return s.MethodNames()
}

// GetMethod returns a method function by name (legacy)
func (s *ServiceImpl) GetMethod(name string) (Function, bool) {
	method, exists := s.methods[name]
	return method, exists
}

// AddMethod adds a method to the service (legacy)
func (s *ServiceImpl) AddMethod(name string, function Function) {
	s.methods[name] = function
}

// FunctionCallImpl implements the api.FunctionCall interface
type FunctionCallImpl struct {
	FunctionName string
	Address      Address
	Input        FunctionData
	Output       FunctionData
	Error        error
	Timestamp    int64
}

// NewFunctionCall creates a new FunctionCall record
func NewFunctionCall(functionName string, address Address, input FunctionData, output FunctionData, err error, timestamp int64) FunctionCall {
	// Convert old types to new FunctionData types
	var inputData FunctionData
	if input != nil {
		inputData = NewFunctionData(input.ToMap())
	}

	var outputData FunctionData
	if output != nil {
		outputData = NewFunctionDataValue(output.Value())
	}

	return FunctionCall{
		FunctionName: functionName,
		Address:      address,
		Input:        inputData,
		Output:       outputData,
		Error:        err,
		Timestamp:    timestamp,
	}
}
