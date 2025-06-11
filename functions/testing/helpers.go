package integration

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"defs.dev/schema"
	httpportal "defs.dev/schema/functions/http"
	jsportal "defs.dev/schema/functions/javascript"
)

// TestPortalManager helps manage multiple portals in integration tests
type TestPortalManager struct {
	httpPortals []*httpportal.HTTPPortal
	jsPortals   []*jsportal.JavaScriptPortal
	t           *testing.T
}

// NewTestPortalManager creates a new portal manager for testing
func NewTestPortalManager(t *testing.T) *TestPortalManager {
	return &TestPortalManager{
		httpPortals: make([]*httpportal.HTTPPortal, 0),
		jsPortals:   make([]*jsportal.JavaScriptPortal, 0),
		t:           t,
	}
}

// NewHTTPPortal creates a new HTTP portal with a unique port
func (m *TestPortalManager) NewHTTPPortal(basePath string) *httpportal.HTTPPortal {
	port := findAvailablePort(m.t)
	portal := httpportal.NewPortal(httpportal.Config{
		Host:           "localhost",
		Port:           port,
		DefaultTimeout: 10 * time.Second,
		BasePath:       basePath,
	})

	m.httpPortals = append(m.httpPortals, portal)
	m.t.Logf("Created HTTP portal on port %d with base path %s", port, basePath)
	return portal
}

// NewHTTPClientPortal creates a new HTTP client portal (no server)
func (m *TestPortalManager) NewHTTPClientPortal() *httpportal.HTTPPortal {
	portal := httpportal.NewPortal(httpportal.Config{
		DefaultTimeout: 15 * time.Second,
		MaxRetries:     2,
		RetryDelay:     500 * time.Millisecond,
	})

	// Don't add to managed portals since this doesn't have a server to stop
	m.t.Logf("Created HTTP client portal")
	return portal
}

// NewJavaScriptPortal creates a new JavaScript portal
func (m *TestPortalManager) NewJavaScriptPortal() *jsportal.JavaScriptPortal {
	portal := jsportal.NewPortalWithDefaults()
	m.jsPortals = append(m.jsPortals, portal)
	m.t.Logf("Created JavaScript portal")
	return portal
}

// StartHTTPServers starts all HTTP portal servers
func (m *TestPortalManager) StartHTTPServers() error {
	for i, portal := range m.httpPortals {
		if err := portal.StartServer(); err != nil {
			return fmt.Errorf("failed to start HTTP server %d: %w", i, err)
		}
		m.t.Logf("Started HTTP server %d", i)
	}

	// Wait for servers to be ready
	time.Sleep(200 * time.Millisecond)
	return nil
}

// Cleanup stops all servers and cleans up resources
func (m *TestPortalManager) Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Stop HTTP servers
	for i, portal := range m.httpPortals {
		if err := portal.StopServer(ctx); err != nil {
			m.t.Logf("Warning: Failed to stop HTTP server %d: %v", i, err)
		} else {
			m.t.Logf("Stopped HTTP server %d", i)
		}
	}

	// Clean up JavaScript portals
	for i, portal := range m.jsPortals {
		portal.Clear()
		m.t.Logf("Cleared JavaScript portal %d", i)
	}
}

// ChainBuilder helps build complex function call chains
type ChainBuilder struct {
	stages []ChainStage
	t      *testing.T
}

type ChainStage struct {
	Name       string
	Portal     interface{} // HTTPPortal or JavaScriptPortal
	Address    string
	Schema     *schema.FunctionSchema
	Function   interface{} // Can be schema.Function or http.Function depending on portal
	IsEndpoint bool        // true if this is the final endpoint, false if it's a bridge
}

// NewChainBuilder creates a new chain builder
func NewChainBuilder(t *testing.T) *ChainBuilder {
	return &ChainBuilder{
		stages: make([]ChainStage, 0),
		t:      t,
	}
}

// AddJavaScriptStage adds a JavaScript function as a stage
func (b *ChainBuilder) AddJavaScriptStage(name string, portal *jsportal.JavaScriptPortal, jsFunc jsportal.JSFunction, schema *schema.FunctionSchema) *ChainBuilder {
	address := portal.GenerateAddress(name, jsFunc)
	function := portal.Apply(address, schema, jsFunc)

	b.stages = append(b.stages, ChainStage{
		Name:       name,
		Portal:     portal,
		Address:    address,
		Schema:     schema,
		Function:   function,
		IsEndpoint: true,
	})

	b.t.Logf("Added JavaScript stage: %s -> %s", name, address)
	return b
}

// AddHTTPBridgeStage adds an HTTP stage that calls the previous stage
func (b *ChainBuilder) AddHTTPBridgeStage(name string, portal *httpportal.HTTPPortal, schema *schema.FunctionSchema) *ChainBuilder {
	if len(b.stages) == 0 {
		panic("Cannot add HTTP bridge stage without a previous stage")
	}

	prevStage := b.stages[len(b.stages)-1]

	// Create HTTP handler that calls the previous stage
	handler := func(ctx context.Context, params map[string]any) (any, error) {
		return callFunction(prevStage.Function, ctx, params)
	}

	address, _ := portal.GenerateAddress(name, handler)
	function, _ := portal.Apply(address, schema, handler)

	b.stages = append(b.stages, ChainStage{
		Name:       name,
		Portal:     portal,
		Address:    address,
		Schema:     schema,
		Function:   function,
		IsEndpoint: false,
	})

	b.t.Logf("Added HTTP bridge stage: %s -> %s (calls %s)", name, address, prevStage.Address)
	return b
}

// GetFinalAddress returns the address of the last stage in the chain
func (b *ChainBuilder) GetFinalAddress() string {
	if len(b.stages) == 0 {
		return ""
	}
	return b.stages[len(b.stages)-1].Address
}

// GetStages returns all stages in the chain
func (b *ChainBuilder) GetStages() []ChainStage {
	return b.stages
}

// CreateJavaScriptFunction creates a common JavaScript function for testing
func CreateJavaScriptFunction(functionName, code string, timeout *time.Duration) jsportal.JSFunction {
	return jsportal.JSFunction{
		Code:         code,
		FunctionName: functionName,
		Timeout:      timeout,
	}
}

// CreateStandardSchema creates a standard function schema for testing
func CreateStandardSchema(name, description string, inputs map[string]schema.Schema, output schema.Schema, required []string) *schema.FunctionSchema {
	builder := schema.NewFunctionSchema().
		Name(name).
		Description(description)

	for inputName, inputSchema := range inputs {
		builder = builder.Input(inputName, inputSchema)
	}

	if output != nil {
		builder = builder.Output(output)
	}

	if len(required) > 0 {
		builder = builder.Required(required...)
	}

	return builder.Build().(*schema.FunctionSchema)
}

// Utility functions

// findAvailablePort finds an available port for testing
func findAvailablePort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

// WaitForServerReady waits for an HTTP server to be ready
func WaitForServerReady(address string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server at %s", address)
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

// Common JavaScript Functions for Testing

// DiscountCalculatorJS creates a discount calculation JavaScript function
func DiscountCalculatorJS() jsportal.JSFunction {
	return CreateJavaScriptFunction("calculateDiscount", `
		function calculateDiscount(params) {
			const { customerTier, purchaseAmount, promoCode } = params;
			
			let discount = 0;
			
			// Base discount by tier
			switch(customerTier) {
				case 'bronze': discount = 0.05; break;
				case 'silver': discount = 0.10; break;
				case 'gold': discount = 0.15; break;
				case 'platinum': discount = 0.20; break;
				default: discount = 0;
			}
			
			// Bonus for large purchases
			if (purchaseAmount > 1000) {
				discount += 0.05;
			}
			
			// Promo code bonus
			if (promoCode === 'SAVE20') {
				discount += 0.20;
			} else if (promoCode === 'SAVE10') {
				discount += 0.10;
			}
			
			// Cap at 50% discount
			discount = Math.min(discount, 0.50);
			
			const discountAmount = purchaseAmount * discount;
			const finalAmount = purchaseAmount - discountAmount;
			
			return {
				originalAmount: purchaseAmount,
				discountPercent: Math.round(discount * 100),
				discountAmount: Math.round(discountAmount * 100) / 100,
				finalAmount: Math.round(finalAmount * 100) / 100,
				appliedTier: customerTier,
				appliedPromo: promoCode || null
			};
		}
	`, timePtr(5*time.Second))
}

// OrderValidatorJS creates an order validation JavaScript function
func OrderValidatorJS() jsportal.JSFunction {
	return CreateJavaScriptFunction("validateOrder", `
		function validateOrder(params) {
			const { orderId, items, customerId } = params;
			
			const errors = [];
			
			// Validate order ID
			if (!orderId || orderId.length < 5) {
				errors.push("Order ID must be at least 5 characters");
			}
			
			// Validate items
			if (!items || items.length === 0) {
				errors.push("Order must have at least one item");
			} else {
				items.forEach((item, index) => {
					if (!item.sku || item.quantity <= 0 || item.price <= 0) {
						errors.push("Item " + (index + 1) + " has invalid data");
					}
				});
			}
			
			// Validate customer
			if (!customerId) {
				errors.push("Customer ID is required");
			}
			
			const isValid = errors.length === 0;
			const totalAmount = items ? items.reduce((sum, item) => 
				sum + (item.quantity * item.price), 0) : 0;
			
			return {
				valid: isValid,
				errors: errors,
				orderId: orderId,
				customerId: customerId,
				totalAmount: Math.round(totalAmount * 100) / 100,
				itemCount: items ? items.length : 0
			};
		}
	`, nil)
}

// Common Schemas for Testing

// DiscountSchema creates a schema for discount calculation
func DiscountSchema() *schema.FunctionSchema {
	return CreateStandardSchema(
		"calculateDiscount",
		"Calculates customer discount based on tier and purchase amount",
		map[string]schema.Schema{
			"customerTier":   schema.String().Description("Customer tier: bronze, silver, gold, platinum").Build(),
			"purchaseAmount": schema.Number().Description("Purchase amount in dollars").Build(),
			"promoCode":      schema.String().Description("Optional promo code").Build(),
		},
		schema.Object().
			Property("originalAmount", schema.Number().Build()).
			Property("discountPercent", schema.Integer().Build()).
			Property("discountAmount", schema.Number().Build()).
			Property("finalAmount", schema.Number().Build()).
			Property("appliedTier", schema.String().Build()).
			Property("appliedPromo", schema.String().Build()).
			Build(),
		[]string{"customerTier", "purchaseAmount"},
	)
}

// OrderValidationSchema creates a schema for order validation
func OrderValidationSchema() *schema.FunctionSchema {
	return CreateStandardSchema(
		"validateOrder",
		"Validates order data",
		map[string]schema.Schema{
			"orderId":    schema.String().Build(),
			"customerId": schema.String().Build(),
			"items": schema.Array().Items(
				schema.Object().
					Property("sku", schema.String().Build()).
					Property("quantity", schema.Integer().Build()).
					Property("price", schema.Number().Build()).
					Build(),
			).Build(),
		},
		schema.Object().
			Property("valid", schema.Boolean().Build()).
			Property("errors", schema.Array().Items(schema.String().Build()).Build()).
			Property("orderId", schema.String().Build()).
			Property("customerId", schema.String().Build()).
			Property("totalAmount", schema.Number().Build()).
			Property("itemCount", schema.Integer().Build()).
			Build(),
		[]string{"orderId", "customerId", "items"},
	)
}

// timePtr returns a pointer to a time.Duration value
func timePtr(d time.Duration) *time.Duration {
	return &d
}

// callFunction is a helper that can call either schema.Function or http.Function
func callFunction(function interface{}, ctx context.Context, params map[string]any) (any, error) {
	// Try schema.Function interface first
	if schemaFunc, ok := function.(schema.TypedFunction); ok {
		return schemaFunc.Call(ctx, params)
	}

	// Try http.Function interface
	if httpFunc, ok := function.(httpportal.Function); ok {
		return httpFunc.Call(ctx, params)
	}

	return nil, fmt.Errorf("unsupported function type: %T", function)
}
