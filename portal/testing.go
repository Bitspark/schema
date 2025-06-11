package portal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
)

// TestingPortalImpl implements api.TestingPortal
type TestingPortalImpl struct {
	functions   map[string]api.Function
	mocks       map[string]api.Function
	callHistory []api.FunctionCall
	mutex       sync.RWMutex
	idCounter   int64
}

// NewTestingPortal creates a new testing portal
func NewTestingPortal() api.TestingPortal {
	return &TestingPortalImpl{
		functions:   make(map[string]api.Function),
		mocks:       make(map[string]api.Function),
		callHistory: make([]api.FunctionCall, 0),
		idCounter:   0,
	}
}

// Apply registers a function with the portal and returns its address
func (p *TestingPortalImpl) Apply(ctx context.Context, function api.Function) (api.Address, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Create wrapped function for recording
	wrappedFunction := p.wrapFunctionForRecording(function)
	address := p.generateAddress(function.Name())
	p.functions[address.String()] = wrappedFunction

	return address, nil
}

// ApplyService registers service (basic implementation)
func (p *TestingPortalImpl) ApplyService(ctx context.Context, service api.Service) (api.Address, error) {
	address := p.generateServiceAddress(service.Schema().Name())
	return address, nil
}

// ResolveFunction resolves address to function
func (p *TestingPortalImpl) ResolveFunction(ctx context.Context, address api.Address) (api.Function, error) {
	if address.Scheme() != "test" && address.Scheme() != "mock" {
		return nil, fmt.Errorf("address is not a test address: %s", address.String())
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	addressStr := address.String()

	if function, exists := p.mocks[addressStr]; exists {
		return function, nil
	}
	if function, exists := p.functions[addressStr]; exists {
		return function, nil
	}

	return nil, fmt.Errorf("function not found: %s", addressStr)
}

// ResolveService resolves address to service
func (p *TestingPortalImpl) ResolveService(ctx context.Context, address api.Address) (api.Service, error) {
	return nil, fmt.Errorf("service resolution not implemented for testing portal")
}

// GenerateAddress creates new address
func (p *TestingPortalImpl) GenerateAddress(name string, metadata map[string]any) api.Address {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.generateAddress(name)
}

// Schemes returns supported schemes
func (p *TestingPortalImpl) Schemes() []string {
	return []string{"test", "mock"}
}

// Close closes the portal
func (p *TestingPortalImpl) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.functions = make(map[string]api.Function)
	p.mocks = make(map[string]api.Function)
	p.callHistory = make([]api.FunctionCall, 0)
	return nil
}

// Health returns health status
func (p *TestingPortalImpl) Health(ctx context.Context) error {
	return nil
}

// Mock registers a mock function
func (p *TestingPortalImpl) Mock(function api.Function) api.Address {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	wrappedFunction := p.wrapFunctionForRecording(function)
	address := p.generateMockAddress(function.Name())
	addressStr := address.String()
	p.mocks[addressStr] = wrappedFunction

	return address
}

// Verify verifies expected calls
func (p *TestingPortalImpl) Verify() error {
	return nil
}

// Reset resets mocks and history
func (p *TestingPortalImpl) Reset() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.mocks = make(map[string]api.Function)
	p.callHistory = make([]api.FunctionCall, 0)
}

// CallHistory returns call history
func (p *TestingPortalImpl) CallHistory() []api.FunctionCall {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	history := make([]api.FunctionCall, len(p.callHistory))
	copy(history, p.callHistory)
	return history
}

// Helper methods

// wrapFunctionForRecording wraps a function to record calls
func (p *TestingPortalImpl) wrapFunctionForRecording(function api.Function) api.Function {
	return &RecordingFunction{
		original: function,
		portal:   p,
	}
}

// RecordingFunction wraps a function to record calls for testing
type RecordingFunction struct {
	original api.Function
	portal   *TestingPortalImpl
}

func (rf *RecordingFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	startTime := time.Now()
	output, err := rf.original.Call(ctx, params)

	call := api.FunctionCall{
		FunctionName: rf.original.Name(),
		Address:      nil,
		Input:        params,
		Output:       output,
		Error:        err,
		Timestamp:    startTime.Unix(),
	}

	rf.portal.mutex.Lock()
	rf.portal.callHistory = append(rf.portal.callHistory, call)
	rf.portal.mutex.Unlock()

	return output, err
}

func (rf *RecordingFunction) Schema() core.FunctionSchema {
	return rf.original.Schema()
}

func (rf *RecordingFunction) Name() string {
	return rf.original.Name()
}

func (p *TestingPortalImpl) generateAddress(name string) api.Address {
	p.idCounter++
	return NewAddressBuilder().
		Scheme("test").
		Path(fmt.Sprintf("/%s", name)).
		Query("id", fmt.Sprintf("%d", p.idCounter)).
		Build()
}

func (p *TestingPortalImpl) generateServiceAddress(name string) api.Address {
	p.idCounter++
	return NewAddressBuilder().
		Scheme("test").
		Path(fmt.Sprintf("/service/%s", name)).
		Query("id", fmt.Sprintf("%d", p.idCounter)).
		Build()
}

func (p *TestingPortalImpl) generateMockAddress(name string) api.Address {
	p.idCounter++
	return NewAddressBuilder().
		Scheme("mock").
		Path(fmt.Sprintf("/%s", name)).
		Query("id", fmt.Sprintf("%d", p.idCounter)).
		Query("mock", "true").
		Build()
}

func (p *TestingPortalImpl) wrapHandlerForRecording(name string, handler func(context.Context, map[string]any) (any, error)) func(context.Context, map[string]any) (any, error) {
	return func(ctx context.Context, input map[string]any) (any, error) {
		startTime := time.Now()
		output, err := handler(ctx, input)

		// Convert to FunctionData for recording
		var inputData api.FunctionData
		if input != nil {
			inputData = NewFunctionData(input)
		}

		var outputData api.FunctionData
		if output != nil {
			outputData = NewFunctionDataValue(output)
		}

		call := api.FunctionCall{
			FunctionName: name,
			Address:      nil,
			Input:        inputData,
			Output:       outputData,
			Error:        err,
			Timestamp:    startTime.Unix(),
		}

		p.mutex.Lock()
		p.callHistory = append(p.callHistory, call)
		p.mutex.Unlock()

		return output, err
	}
}

// GetCallsForFunction returns call history for a specific function
func (p *TestingPortalImpl) GetCallsForFunction(functionName string) []api.FunctionCall {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var calls []api.FunctionCall
	for _, call := range p.callHistory {
		if call.FunctionName == functionName {
			calls = append(calls, call)
		}
	}
	return calls
}

// GetCallCount returns the number of calls for a specific function
func (p *TestingPortalImpl) GetCallCount(functionName string) int {
	return len(p.GetCallsForFunction(functionName))
}

// WasCalled returns true if a function was called at least once
func (p *TestingPortalImpl) WasCalled(functionName string) bool {
	return p.GetCallCount(functionName) > 0
}

// ClearHistory clears the call history but keeps mocks
func (p *TestingPortalImpl) ClearHistory() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.callHistory = make([]api.FunctionCall, 0)
}

// Stats returns statistics about the testing portal
func (p *TestingPortalImpl) Stats() TestingPortalStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return TestingPortalStats{
		FunctionCount: len(p.functions),
		MockCount:     len(p.mocks),
		CallCount:     len(p.callHistory),
		NextID:        p.idCounter + 1,
	}
}

// TestingPortalStats represents statistics for the testing portal
type TestingPortalStats struct {
	FunctionCount int
	MockCount     int
	CallCount     int
	NextID        int64
}
