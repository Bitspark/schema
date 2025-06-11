package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"defs.dev/schema"
	ws "defs.dev/schema/functions/websocket"
)

func main() {
	fmt.Println("WebSocket Portal Example")
	fmt.Println("========================")

	// Create a simple calculator function
	var calculator schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		p := params
		operation := p["operation"].(string)
		a := p["a"].(float64)
		b := p["b"].(float64)

		var result float64
		switch operation {
		case "add":
			result = a + b
		case "subtract":
			result = a - b
		case "multiply":
			result = a * b
		case "divide":
			if b == 0 {
				return schema.FromAny(nil), &ws.WebSocketFunctionError{
					Code:    400,
					Message: "Division by zero not allowed",
				}
			}
			result = a / b
		default:
			return schema.FromAny(nil), &ws.WebSocketFunctionError{
				Code:    400,
				Message: "Invalid operation",
			}
		}

		return schema.FromAny(map[string]any{
			"result":    result,
			"operation": operation,
			"operands":  []float64{a, b},
		}), nil
	}

	// Create function schema
	calcSchema := schema.NewFunctionSchema().
		Name("calculator").
		Description("A simple calculator function").
		Input("operation", schema.NewString().Description("Operation: add, subtract, multiply, divide").Build()).
		Input("a", schema.NewNumber().Description("First number").Build()).
		Input("b", schema.NewNumber().Description("Second number").Build()).
		Output(schema.NewObject().
			Property("result", schema.NewNumber().Build()).
			Property("operation", schema.NewString().Build()).
			Property("operands", schema.NewArray().Items(schema.NewNumber().Build()).Build()).
			Build()).
		Build()

	// 1. Create and start server
	fmt.Println("\n1. Creating WebSocket server...")
	server := ws.NewServer("localhost", 8080)

	// Add logging middleware
	server.AddMiddleware(ws.NewLoggingMiddleware())

	// Start server
	if err := server.StartServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.StopServer(ctx)
	}()

	fmt.Printf("Server started at: %s\n", server.GetServerAddress())

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// 2. Register function
	fmt.Println("\n2. Registering calculator function...")
	address, err := server.GenerateAddress("calculator", calculator)
	if err != nil {
		log.Fatalf("Failed to generate address: %v", err)
	}

	fmt.Printf("Generated address: %s\n", address)

	_, err = server.Apply(address, calcSchema.(*schema.FunctionSchema), calculator)
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	fmt.Println("Function registered successfully!")

	// Wait for registration to complete
	time.Sleep(200 * time.Millisecond)

	// 3. Create client and resolve function
	fmt.Println("\n3. Creating WebSocket client...")
	client := ws.NewClient()

	calcFunc, err := client.ResolveFunction(context.Background(), address)
	if err != nil {
		log.Fatalf("Failed to resolve function: %v", err)
	}

	fmt.Println("Function resolved successfully!")

	// 4. Make function calls
	fmt.Println("\n4. Making function calls...")

	testCases := []struct {
		name      string
		operation string
		a, b      float64
		expectErr bool
	}{
		{"Addition", "add", 10, 5, false},
		{"Subtraction", "subtract", 10, 3, false},
		{"Multiplication", "multiply", 4, 7, false},
		{"Division", "divide", 15, 3, false},
		{"Division by zero", "divide", 10, 0, true},
		{"Invalid operation", "modulo", 10, 3, true},
	}

	for _, tc := range testCases {
		fmt.Printf("\n• %s: %.1f %s %.1f\n", tc.name, tc.a, tc.operation, tc.b)

		result, err := calcFunc.Call(context.Background(), map[string]interface{}{
			"operation": tc.operation,
			"a":         tc.a,
			"b":         tc.b,
		})

		if tc.expectErr {
			if err != nil {
				fmt.Printf("  ✓ Expected error: %v\n", err)
			} else {
				fmt.Printf("  ✗ Expected error but got result: %v\n", result)
			}
		} else {
			if err != nil {
				fmt.Printf("  ✗ Unexpected error: %v\n", err)
			} else {
				resultMap := result.Value().(map[string]any)
				resultValue := resultMap["result"].(float64)
				fmt.Printf("  ✓ Result: %.2f\n", resultValue)
			}
		}
	}

	// 5. Test ping functionality
	fmt.Println("\n5. Testing ping functionality...")
	err = client.PingEndpoint(context.Background(), server.GetServerAddress())
	if err != nil {
		fmt.Printf("  ✗ Ping failed: %v\n", err)
	} else {
		fmt.Println("  ✓ Ping successful!")
	}

	// 6. Show server statistics
	fmt.Println("\n6. Server statistics:")
	stats := server.Statistics()
	fmt.Printf("  • Functions: %d\n", len(stats["functions"].([]map[string]any)))
	fmt.Printf("  • Connections: %d\n", len(stats["connections"].([]map[string]any)))
	fmt.Printf("  • Middleware: %d\n", stats["middleware"].(int))

	// 7. Test connection management
	fmt.Println("\n7. Testing connection management...")
	serverAddr := server.GetServerAddress()

	if client.IsConnected(serverAddr) {
		fmt.Println("  ✓ Client is connected to server")

		connStats, err := client.GetConnectionStats(serverAddr)
		if err != nil {
			fmt.Printf("  ✗ Failed to get connection stats: %v\n", err)
		} else {
			fmt.Printf("  • Connection ID: %s\n", connStats.ID)
			fmt.Printf("  • Last activity: %v ago\n", time.Since(connStats.LastActivity).Round(time.Millisecond))
			fmt.Printf("  • Pending calls: %d\n", connStats.PendingCalls)
		}
	} else {
		fmt.Println("  ✗ Client not connected to server")
	}

	// 8. Graceful cleanup
	fmt.Println("\n8. Cleaning up...")
	if err := client.CloseAllConnections(); err != nil {
		fmt.Printf("  ✗ Failed to close client connections: %v\n", err)
	} else {
		fmt.Println("  ✓ Client connections closed")
	}

	fmt.Println("\nWebSocket Portal Example completed successfully!")
	fmt.Println("This demonstrated:")
	fmt.Println("  • Server creation and function registration")
	fmt.Println("  • Client connection and function resolution")
	fmt.Println("  • Real-time function calls over WebSocket")
	fmt.Println("  • Error handling and validation")
	fmt.Println("  • Ping/pong functionality")
	fmt.Println("  • Connection management and statistics")
}
