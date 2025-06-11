package websocket

import (
	"context"
	"fmt"
	"testing"
	"time"

	"defs.dev/schema"
)

func TestWebSocketPortal_EndToEnd_RealWebSocket(t *testing.T) {
	// Create a test function that we'll serve over WebSocket
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
			return schema.FromAny(nil), &WebSocketFunctionError{
				Code:    400,
				Message: "Client error",
				Details: "Invalid operation",
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
		Input("a", schema.NewInteger().Description("First number").Build()).
		Input("b", schema.NewInteger().Description("Second number").Build()).
		Input("operation", schema.NewString().Description("Operation: add, subtract, multiply").Build()).
		Output(schema.NewObject().
			Property("result", schema.NewInteger().Build()).
			Property("input", schema.NewObject().Build()).
			Build()).
		Build()

	// Create WebSocket portal for server
	serverPortal := NewServer("localhost", 8090)

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
	time.Sleep(300 * time.Millisecond)

	// Register function as WebSocket endpoint
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
	time.Sleep(200 * time.Millisecond)

	// Create client portal
	clientPortal := NewClient()

	// Resolve function to get WebSocket client
	clientFunc, err := clientPortal.ResolveFunction(context.Background(), address)
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	// === ACTUAL WEBSOCKET TESTS ===

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
					t.Fatalf("WebSocket call failed: %v", err)
				}

				resultMap := result.Value().(map[string]any)
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

		// Should be a WebSocket function error
		if wsErr, ok := err.(*WebSocketFunctionError); ok {
			if wsErr.Code != 400 {
				t.Errorf("Expected status 400, got %d", wsErr.Code)
			}
			t.Logf("✓ Error handled correctly: %v", wsErr)
		} else {
			t.Errorf("Expected WebSocketFunctionError, got %T: %v", err, err)
		}

		// Result should be nil on error
		if result.Value() != nil {
			t.Errorf("Expected nil result on error, got %v", result)
		}
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		// Create a pre-cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// This should fail due to context cancellation
		_, err := clientFunc.Call(ctx, map[string]interface{}{
			"a": 1, "b": 2, "operation": "add",
		})

		if err == nil {
			t.Error("Expected timeout error, but call succeeded")
		} else {
			t.Logf("✓ Timeout handled correctly: %v", err)
		}
	})
}

func TestWebSocketPortal_EndToEnd_Authentication(t *testing.T) {
	// Create a simple echo function
	var echoHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return schema.FromAny(map[string]any{
			"echo":        params,
			"received_at": time.Now().Unix(),
		}), nil
	}

	echoSchema := schema.NewFunctionSchema().
		Name("echo").
		Input("message", schema.NewString().Build()).
		Output(schema.NewObject().Build()).
		Build()

	// Create server
	portal := NewServer("localhost", 8091)

	if err := portal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	time.Sleep(300 * time.Millisecond)

	// Register function
	address, err := portal.GenerateAddress("echo", echoHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	_, err = portal.Apply(address, echoSchema.(*schema.FunctionSchema), echoHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Test with authentication middleware
	clientPortal := NewClient()

	// Add bearer token authentication (simplified for WebSocket)
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

	resultMap := result.Value().(map[string]any)
	echo := resultMap["echo"].(map[string]any)

	if echo["message"] != "Hello with auth!" {
		t.Errorf("Expected echo of message, got %v", echo)
	}

	t.Logf("✓ Authenticated call succeeded: %+v", result)
}

func TestWebSocketPortal_EndToEnd_Middleware(t *testing.T) {
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
		Output(schema.NewObject().Build()).
		Build()

	// Create server with metrics middleware
	portal := NewServer("localhost", 8092)

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

	time.Sleep(300 * time.Millisecond)

	// Register function
	address, err := portal.GenerateAddress("counter", counterHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	_, err = portal.Apply(address, counterSchema.(*schema.FunctionSchema), counterHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

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

		resultMap := result.Value().(map[string]any)
		callNumber := int(resultMap["call_number"].(float64))

		if callNumber != i+1 {
			t.Errorf("Expected call number %d, got %d", i+1, callNumber)
		}

		t.Logf("✓ Call %d successful: %+v", i+1, result)
	}

	// Check metrics
	metricsData := metrics.GetMetrics()
	messageCount := metricsData["message_count"].(int64)
	activeConnections := metricsData["active_connections"].(int64)

	if messageCount < int64(numCalls*2) { // Each call = request + response
		t.Errorf("Expected at least %d messages, got %d", numCalls*2, messageCount)
	}

	if activeConnections < 1 {
		t.Errorf("Expected at least 1 active connection, got %d", activeConnections)
	}

	t.Logf("✓ Metrics collected: %+v", metricsData)
}

func TestWebSocketPortal_EndToEnd_Ping(t *testing.T) {
	// Create server
	portal := NewServer("localhost", 8093)

	// Add logging to see ping/pong messages
	portal.AddMiddleware(NewLoggingMiddleware())

	if err := portal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	time.Sleep(300 * time.Millisecond)

	// Create client
	clientPortal := NewClient()

	// Test ping functionality
	serverAddress := portal.GetServerAddress()

	// Connect to the server
	err := clientPortal.ConnectTo(context.Background(), serverAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer clientPortal.DisconnectFrom(serverAddress)

	// Give more time for connection to stabilize
	time.Sleep(500 * time.Millisecond)

	// Test ping with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = clientPortal.PingEndpoint(ctx, serverAddress)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	t.Logf("✓ Ping successful")

	// Test connection status
	if !clientPortal.IsConnected(serverAddress) {
		t.Error("Expected to be connected, but IsConnected returned false")
	}

	// Get connection stats
	stats, err := clientPortal.GetConnectionStats(serverAddress)
	if err != nil {
		t.Fatalf("Failed to get connection stats: %v", err)
	}

	t.Logf("✓ Connection stats: %+v", stats)
}

func TestWebSocketPortal_EndToEnd_Concurrency(t *testing.T) {
	// Test concurrent connections and calls
	var simpleHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		return schema.FromAny(map[string]any{
			"processed": params,
			"worker_id": time.Now().UnixNano() % 1000,
		}), nil
	}

	simpleSchema := schema.NewFunctionSchema().
		Name("simple").
		Input("data", schema.NewString().Build()).
		Output(schema.NewObject().Build()).
		Build()

	// Create server
	portal := NewServer("localhost", 8094)

	// Add metrics to track concurrency
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

	time.Sleep(300 * time.Millisecond)

	// Register function
	address, err := portal.GenerateAddress("simple", simpleHandler)
	if err != nil {
		t.Fatalf("Failed to generate address: %v", err)
	}

	_, err = portal.Apply(address, simpleSchema.(*schema.FunctionSchema), simpleHandler)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Create multiple clients
	numClients := 3
	numCallsPerClient := 3

	results := make(chan error, numClients*numCallsPerClient)

	// Launch concurrent clients
	for clientID := 0; clientID < numClients; clientID++ {
		go func(id int) {
			clientPortal := NewClient()
			clientFunc, err := clientPortal.ResolveFunction(context.Background(), address)
			if err != nil {
				results <- err
				return
			}

			for callID := 0; callID < numCallsPerClient; callID++ {
				_, err := clientFunc.Call(context.Background(), map[string]interface{}{
					"data": fmt.Sprintf("client_%d_call_%d", id, callID),
				})
				results <- err
			}
		}(clientID)
	}

	// Collect results
	successCount := 0
	errorCount := 0
	for i := 0; i < numClients*numCallsPerClient; i++ {
		err := <-results
		if err != nil {
			t.Logf("Call failed: %v", err)
			errorCount++
		} else {
			successCount++
		}
	}

	t.Logf("✓ Concurrent calls completed: %d successful, %d failed", successCount, errorCount)

	if successCount == 0 {
		t.Error("No concurrent calls succeeded")
	}

	// Check metrics
	metricsData := metrics.GetMetrics()
	t.Logf("✓ Final metrics: %+v", metricsData)
}

func TestWebSocketPortal_EndToEnd_ConnectionManagement(t *testing.T) {
	// Test connection lifecycle management
	portal := NewServer("localhost", 8095)

	if err := portal.StartServer(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	time.Sleep(300 * time.Millisecond)

	serverAddress := portal.GetServerAddress()

	// Create client
	clientPortal := NewClient()

	// Test connection lifecycle
	if clientPortal.IsConnected(serverAddress) {
		t.Error("Expected not to be connected initially")
	}

	// Connect
	err := clientPortal.ConnectTo(context.Background(), serverAddress)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !clientPortal.IsConnected(serverAddress) {
		t.Error("Expected to be connected after ConnectTo")
	}

	// Get connections from server
	connections := portal.GetConnections()
	if len(connections) == 0 {
		t.Error("Expected server to have connections")
	}
	t.Logf("✓ Server connections: %d", len(connections))

	// Disconnect
	err = clientPortal.DisconnectFrom(serverAddress)
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Give some time for disconnection to propagate
	time.Sleep(100 * time.Millisecond)

	if clientPortal.IsConnected(serverAddress) {
		t.Error("Expected not to be connected after DisconnectFrom")
	}

	t.Logf("✓ Connection lifecycle test completed")
}
