package local

import (
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"defs.dev/schema"
	"defs.dev/schema/functions"
)

// Portal implements functions.Portal[schema.FunctionHandler] for in-process Go functions
type Portal struct {
	functions map[string]*LocalFunction // address -> function
	mu        sync.RWMutex
}

// NewPortal creates a new local portal
func NewPortal() *Portal {
	return &Portal{
		functions: make(map[string]*LocalFunction),
	}
}

func (p *Portal) Scheme() string {
	return "local"
}

func (p *Portal) GenerateAddress(name string, data schema.FunctionHandler) string {
	// Generate unique ID based on function pointer
	id := generateHandlerID(data)
	return fmt.Sprintf("local://%s-%s", name, id[:8])
}

func (p *Portal) Apply(address string, schema *schema.FunctionSchema, data schema.FunctionHandler) schema.Function {
	// Extract name from address
	name := extractNameFromLocalAddress(address)

	// Build LocalFunction using existing constructor but with schema parameters
	localFunc := &LocalFunction{
		name:        name,
		description: schema.Metadata().Description,
		parameters:  buildParametersFromSchema(schema),
		returns:     buildReturnsFromSchema(schema),
		handler:     data,
		examples:    extractExamplesFromSchema(schema),
		tags:        extractTagsFromSchema(schema),
	}

	// Store for resolution
	p.mu.Lock()
	p.functions[address] = localFunc
	p.mu.Unlock()

	return localFunc
}

func (p *Portal) ResolveFunction(ctx context.Context, address string) (schema.Function, error) {
	p.mu.RLock()
	function, exists := p.functions[address]
	p.mu.RUnlock()

	if !exists {
		return nil, &functions.PortalError{
			Scheme:  "local",
			Address: address,
			Message: "function not found",
		}
	}

	return function, nil
}

// Helper functions

func generateHandlerID(handler schema.FunctionHandler) string {
	// Use reflection to get function pointer value
	h := sha256.New()
	funcValue := reflect.ValueOf(handler)

	// Write function pointer value to hash
	h.Write([]byte(fmt.Sprintf("%v", funcValue.Pointer())))

	// Return hex encoded hash
	return fmt.Sprintf("%x", h.Sum(nil))
}

func extractNameFromLocalAddress(address string) string {
	// local://functionName-12345678
	parts := strings.SplitN(strings.TrimPrefix(address, "local://"), "-", 2)
	if len(parts) >= 1 {
		return parts[0]
	}
	return "unknown"
}

func buildParametersFromSchema(functionSchema *schema.FunctionSchema) schema.Schema {
	// Build object schema from function inputs
	inputs := functionSchema.Inputs()
	required := functionSchema.Required()

	if len(inputs) == 0 {
		return schema.Object().Build()
	}

	builder := schema.Object()
	for name, inputSchema := range inputs {
		builder = builder.Property(name, inputSchema)
	}
	builder = builder.Required(required...)

	return builder.Build()
}

func buildReturnsFromSchema(functionSchema *schema.FunctionSchema) schema.Schema {
	return functionSchema.Outputs()
}

func extractExamplesFromSchema(functionSchema *schema.FunctionSchema) []FunctionExample {
	// For now, return empty slice
	// In the future, this could extract examples from schema metadata
	return []FunctionExample{}
}

func extractTagsFromSchema(functionSchema *schema.FunctionSchema) []string {
	// For now, return empty slice
	// In the future, this could extract tags from schema metadata
	return []string{}
}
