package http

import (
	"context"
	"testing"
	"time"

	"defs.dev/schema"
)

func TestHTTPPortal_EndToEnd_RealHTTP(t *testing.T) {
	// Create a test function that we'll serve over HTTP
	var mathHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		p := params
		a := int(p["a"].(float64)) // JSON numbers come as float64
		b := int(p["b"].(float64))
		op := p["operation"].(string)

		var result int
		switch op {
		case "add":
			result = a + b
		case "subtract":
			result = a - b
		case "multiply":
			result = a * b
		default:
			return schema.FunctionOutput{}, &HTTPFunctionError{
				StatusCode: 400,
				Message:    "Invalid operation",
			}
		}

		return schema.FromAny(map[string]any{
			"result": result,
			"input":  map[string]any{"a": a, "b": b, "operation": op},
		}), nil
	}

	// Create function schema
	mathSchema := schema.NewFunctionSchema().
		Name("mathOperation").
		Description("Performs basic math operations").
		Input("a", schema.Integer().Description("First number").Build()).
		Input("b", schema.Integer().Description("Second number").Build()).
		Input("operation", schema.String().Description("Operation: add, subtract, multiply").Build()).
		Output(schema.Object().
			Property("result", schema.Integer().Build()).
			Property("input", schema.Object().Build()).
			Build()).
		Build()

	// Create HTTP portal for server
	serverPortal := NewPortal(Config{
		Host:           "localhost",
		Port:           8084, // Different port to avoid conflicts
		DefaultTimeout: 5 * time.Second,
		BasePath:       "/api/v1",
	})

	// Add logging middleware to see what's happening
	serverPortal.AddMiddleware(NewLoggingMiddleware())

	// Start server
	if err := serverPortal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		serverPortal.StopServer(ctx)
	}()

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)

	// Register function as HTTP endpoint
	address, err := serverPortal.GenerateAddress("mathOperation", mathHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	t.Logf("Generated address: %s", address)

	_, err = serverPortal.Apply(address, mathSchema.(*schema.FunctionSchema), mathHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Wait a bit more for registration to complete
	time.Sleep(100 * time.Millisecond)

	// Create client portal (can be same instance or different)
	clientPortal := NewPortal(Config{
		DefaultTimeout: 10 * time.Second,
		MaxRetries:     2,
		RetryDelay:     500 * time.Millisecond,
	})

	// Resolve function to get HTTP client
	clientFunc, err := clientPortal.ResolveFunction(context.Background(), address)
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	// === ACTUAL HTTP TESTS ===

	t.Run("ValidMathOperations", func(t *testing.T) {
		testCases := []struct {
			name     string
			params   map[string]interface{}
			expected int
		}{
			{
				name:     "Addition",
				params:   map[string]interface{}{"a": 5, "b": 3, "operation": "add"},
				expected: 8,
			},
			{
				name:     "Subtraction",
				params:   map[string]interface{}{"a": 10, "b": 4, "operation": "subtract"},
				expected: 6,
			},
			{
				name:     "Multiplication",
				params:   map[string]interface{}{"a": 7, "b": 6, "operation": "multiply"},
				expected: 42,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := clientFunc.Call(context.Background(), tc.params)
				if err != nil {
					t.Fatalf("HTTP call failed: %v", err)
				}

				resultMap := result.(map[string]interface{})
				actualResult := int(resultMap["result"].(float64))

				if actualResult != tc.expected {
					t.Errorf("Expected result %d, got %d", tc.expected, actualResult)
				}

				t.Logf("✓ %s: %d (params: %+v)", tc.name, actualResult, tc.params)
			})
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test invalid operation
		result, err := clientFunc.Call(context.Background(), map[string]interface{}{
			"a": 5, "b": 3, "operation": "divide",
		})

		if err == nil {
			t.Error("Expected error for invalid operation, but got none")
		}

		// Should be an HTTP error
		if httpErr, ok := err.(*HTTPFunctionError); ok {
			if httpErr.StatusCode != 400 {
				t.Errorf("Expected status 400, got %d", httpErr.StatusCode)
			}
			t.Logf("✓ Error handled correctly: %v", httpErr)
		} else {
			t.Errorf("Expected HTTPFunctionError, got %T: %v", err, err)
		}

		// Result should be nil on error
		if result != nil {
			t.Errorf("Expected nil result on error, got %v", result)
		}
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		// Create a slow handler that will definitely timeout
		var slowHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
			time.Sleep(200 * time.Millisecond) // Sleep longer than timeout
			return schema.FromAny(map[string]any{"result": 42}), nil
		}

		slowSchema := schema.NewFunctionSchema().
			Name("slow").
			Output(schema.Object().Build()).
			Build()

		// Use the same server portal to register the slow function
		slowAddress, err := serverPortal.GenerateAddress("slow", slowHandler)
		if err != nil {
			t.Fatalf("Failed to generate slow address: %v", err)
		}

		_, err = serverPortal.Apply(slowAddress, slowSchema.(*schema.FunctionSchema), slowHandler)
		if err != nil {
			t.Fatalf("Failed to register slow function: %v", err)
		}

		// Create a client with short timeout
		shortTimeoutPortal := NewPortal(Config{
			DefaultTimeout: 50 * time.Millisecond, // Shorter than the sleep
			MaxRetries:     1,
		})

		shortTimeoutFunc, err := shortTimeoutPortal.ResolveFunction(context.Background(), slowAddress)
		if err != nil {
			t.Fatalf("Failed to resolve slow function: %v", err)
		}

		// This should timeout
		_, err = shortTimeoutFunc.Call(context.Background(), schema.FromMap(map[string]interface{}{}))

		if err == nil {
			t.Error("Expected timeout error, but call succeeded")
		} else {
			t.Logf("✓ Timeout handled correctly: %v", err)
		}
	})
}

func TestHTTPPortal_EndToEnd_Authentication(t *testing.T) {
	// Create a simple echo function
	var echoHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return schema.FromAny(map[string]any{
			"echo":        params,
			"received_at": time.Now().Unix(),
		}), nil
	}

	echoSchema := schema.NewFunctionSchema().
		Name("echo").
		Input("message", schema.String().Build()).
		Output(schema.Object().Build()).
		Build()

	// Create server
	portal := NewPortal(Config{
		Host:           "localhost",
		Port:           0, // Use dynamic port allocation
		DefaultTimeout: 5 * time.Second,
	})

	if err := portal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	// Register function
	address, err := portal.GenerateAddress("echo", echoHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	_, err = portal.Apply(address, echoSchema.(*schema.FunctionSchema), echoHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test with authentication middleware
	clientPortal := NewPortal()

	// Add bearer token authentication
	clientPortal.AddMiddleware(NewBearerAuthMiddleware("test-token-123"))

	clientFunc, err := clientPortal.ResolveFunction(context.Background(), address)
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	// Make authenticated call
	result, err := clientFunc.Call(context.Background(), map[string]interface{}{
		"message": "Hello with auth!",
	})

	if err != nil {
		t.Fatalf("Authenticated call failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	echo := resultMap["echo"].(map[string]interface{})

	if echo["message"] != "Hello with auth!" {
		t.Errorf("Expected echo of message, got %v", echo)
	}

	t.Logf("✓ Authenticated call succeeded: %+v", result)
}

func TestHTTPPortal_EndToEnd_Middleware(t *testing.T) {
	// Create counter function to test metrics
	callCount := 0
	var counterHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		callCount++
		return schema.FromAny(map[string]any{
			"call_number": callCount,
			"timestamp":   time.Now().Unix(),
		}), nil
	}

	counterSchema := schema.NewFunctionSchema().
		Name("counter").
		Output(schema.Object().Build()).
		Build()

	// Create server with metrics middleware
	portal := NewPortal(Config{
		Host:           "localhost",
		Port:           0, // Use dynamic port allocation
		DefaultTimeout: 5 * time.Second,
	})

	// Add metrics middleware
	metrics := NewMetricsMiddleware()
	portal.AddMiddleware(metrics)

	if err := portal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	// Register function
	address, err := portal.GenerateAddress("counter", counterHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	_, err = portal.Apply(address, counterSchema.(*schema.FunctionSchema), counterHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Create client
	clientFunc, err := portal.ResolveFunction(context.Background(), address)
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	// Make multiple calls
	numCalls := 3
	for i := 0; i < numCalls; i++ {
		result, err := clientFunc.Call(context.Background(), nil)
		if err != nil {
			t.Fatalf("Call %d failed: %v", i+1, err)
		}

		resultMap := result.(map[string]interface{})
		callNumber := int(resultMap["call_number"].(float64))

		if callNumber != i+1 {
			t.Errorf("Expected call number %d, got %d", i+1, callNumber)
		}

		t.Logf("✓ Call %d successful: %+v", i+1, result)
	}

	// Check metrics
	metricsData := metrics.GetMetrics()
	requestCount := metricsData["request_count"].(int64)
	responseCount := metricsData["response_count"].(int64)

	if requestCount < int64(numCalls) {
		t.Errorf("Expected at least %d requests, got %d", numCalls, requestCount)
	}

	if responseCount < int64(numCalls) {
		t.Errorf("Expected at least %d responses, got %d", numCalls, responseCount)
	}

	t.Logf("✓ Metrics collected: %+v", metricsData)
}
