package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"defs.dev/schema"
	httpportal "defs.dev/schema/functions/http"
)

func main() {
	fmt.Println("HTTP Portal Basic Example")
	fmt.Println("=========================")

	// Create a simple validation function
	var validateUserHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		p := params
		username := p["username"].(string)
		email := p["email"].(string)

		errors := []string{}
		if len(username) < 3 {
			errors = append(errors, "Username must be at least 3 characters")
		}
		if email == "" || !contains(email, "@") {
			errors = append(errors, "Valid email is required")
		}

		return schema.FromAny(map[string]any{
			"valid":  len(errors) == 0,
			"errors": errors,
		}), nil
	}

	// Define the function schema
	userValidationSchema := schema.NewFunctionSchema().
		Name("validateUser").
		Description("Validates user registration data").
		Input("username", schema.NewString().MinLength(1).Description("Username to validate").Build()).
		Input("email", schema.NewString().Description("Email address to validate").Build()).
		Output(schema.NewObject().
			Property("valid", schema.NewBoolean().Description("Whether the user data is valid").Build()).
			Property("errors", schema.NewArray().Items(schema.NewString().Build()).Description("Validation errors").Build()).
			Build()).
		Build()

	// === PROVIDER SIDE (Server) ===
	fmt.Println("\n1. Setting up HTTP Portal (Provider Side)")

	// Create HTTP portal with custom configuration
	portal := httpportal.NewPortal(httpportal.Config{
		Host:           "localhost",
		Port:           8083,
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		BasePath:       "/api",
	})

	// Add logging middleware
	portal.AddMiddleware(httpportal.NewLoggingMiddleware())

	// Start the server
	if err := portal.StartServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	fmt.Printf("✓ HTTP server started at %s\n", portal.GetServerAddress())

	// Register the function as an HTTP endpoint
	address, err := portal.GenerateAddress("validateUser", validateUserHandler)
	if err != nil {
		log.Fatalf("Failed to generate address: %v", err)
	}

	endpointFunc, err := portal.Apply(address, userValidationSchema.(*schema.FunctionSchema), validateUserHandler)
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	fmt.Printf("✓ Function registered at: %s\n", address)
	fmt.Printf("✓ Endpoint function created: %s\n", endpointFunc.Address())

	// Wait a moment for server to fully start
	time.Sleep(500 * time.Millisecond)

	// === CONSUMER SIDE (Client) ===
	fmt.Println("\n2. Using HTTP Portal (Consumer Side)")

	// Resolve the address to get a client function
	clientFunc, err := portal.ResolveFunction(context.Background(), address)
	if err != nil {
		log.Fatalf("Failed to resolve function: %v", err)
	}

	fmt.Printf("✓ Client function resolved for: %s\n", clientFunc.Address())

	// Test the function with valid data
	fmt.Println("\n3. Testing Function Calls")

	testCases := []map[string]interface{}{
		{
			"username": "johndoe",
			"email":    "john@example.com",
		},
		{
			"username": "ab", // Too short
			"email":    "invalid-email",
		},
		{
			"username": "alice",
			"email":    "",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\nTest Case %d: %+v\n", i+1, testCase)

		result, err := clientFunc.Call(context.Background(), testCase)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✓ Result: %+v\n", result)
		}
	}

	// === Demonstrate Portal Features ===
	fmt.Println("\n4. Portal Features Demo")

	// Show registered functions
	registrations := portal.GetRegisteredFunctions()
	fmt.Printf("✓ Registered functions: %d\n", len(registrations))

	for addr, reg := range registrations {
		fmt.Printf("  - %s: %s\n", reg.Name, addr)
	}

	// Show supported schemes
	schemes := portal.Scheme()
	fmt.Printf("✓ Supported schemes: %v\n", schemes)

	// Show server status
	fmt.Printf("✓ Server running: %v\n", portal.IsServerRunning())

	fmt.Println("\n=== HTTP Portal Example Complete ===")
	fmt.Println("The portal successfully demonstrated:")
	fmt.Println("• Provider side: Local function → HTTP endpoint")
	fmt.Println("• Consumer side: HTTP address → Client function")
	fmt.Println("• Unified interface: Same calling pattern for both")
	fmt.Println("• Error handling: Network and application errors")
	fmt.Println("• Middleware: Logging and extensibility")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		len(s) >= len(substr) &&
		findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
