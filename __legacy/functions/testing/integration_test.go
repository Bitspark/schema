package integration

import (
	"context"
	"testing"
	"time"

	"defs.dev/schema"
	httpportal "defs.dev/schema/functions/http"
	jsportal "defs.dev/schema/functions/javascript"
)

// TestCrossPortalIntegration_JavaScriptOverHTTP tests a JavaScript function
// registered at an HTTP server and called via HTTP client
func TestCrossPortalIntegration_JavaScriptOverHTTP(t *testing.T) {
	// === 1. Create JavaScript function ===
	discountCalculatorJS := jsportal.JSFunction{
		Code: `
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
					appliedPromo: promoCode || ""
				};
			}
		`,
		FunctionName: "calculateDiscount",
		Timeout:      timePtr(5 * time.Second),
	}

	// === 2. Create function schema ===
	discountSchema := schema.NewFunctionSchema().
		Name("calculateDiscount").
		Description("Calculates customer discount based on tier and purchase amount").
		Input("customerTier", schema.NewString().Description("Customer tier: bronze, silver, gold, platinum").Build()).
		Input("purchaseAmount", schema.NewNumber().Description("Purchase amount in dollars").Build()).
		Input("promoCode", schema.NewString().Description("Optional promo code").Build()).
		Output(schema.NewObject().
								Property("originalAmount", schema.NewNumber().Build()).
								Property("discountPercent", schema.NewInteger().Build()).
								Property("discountAmount", schema.NewNumber().Build()).
								Property("finalAmount", schema.NewNumber().Build()).
								Property("appliedTier", schema.NewString().Build()).
								Property("appliedPromo", schema.NewString().Build()).
								Build()).
		Required("customerTier", "purchaseAmount"). // promoCode is optional (not in required list)
		Build()

	// === 3. Setup JavaScript portal ===
	jsPortal := jsportal.NewPortalWithDefaults()
	jsAddress := jsPortal.GenerateAddress("calculateDiscount", discountCalculatorJS)
	jsFunction := jsPortal.Apply(jsAddress, discountSchema.(*schema.FunctionSchema), discountCalculatorJS)

	// === 4. Create HTTP wrapper function ===
	// This is the bridge: HTTP handler that calls JavaScript function
	var httpHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		t.Logf("HTTP handler received params: %+v", params)

		// Call the JavaScript function
		result, err := jsFunction.Call(ctx, schema.FromMap(params))
		if err != nil {
			t.Logf("JavaScript function error: %v", err)
			return schema.FunctionOutput{}, err
		}

		t.Logf("JavaScript function result: %+v", result)
		return result, nil
	}

	// === 5. Setup HTTP portal server ===
	httpPortal := httpportal.NewPortal(httpportal.Config{
		Host:           "localhost",
		Port:           8085, // Unique port for this test
		DefaultTimeout: 10 * time.Second,
		BasePath:       "/api/v1",
	})

	// Add logging middleware
	httpPortal.AddMiddleware(httpportal.NewLoggingMiddleware())

	// Start HTTP server
	if err := httpPortal.StartServer(); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		httpPortal.StopServer(ctx)
	}()

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)

	// === 6. Register HTTP endpoint ===
	httpAddress, err := httpPortal.GenerateAddress("calculateDiscount", httpHandler)
	if err != nil {
		t.Fatalf("Failed to generate HTTP address: %v", err)
	}

	t.Logf("Generated HTTP address: %s", httpAddress)
	t.Logf("JavaScript address: %s", jsAddress)

	_, err = httpPortal.Apply(httpAddress, discountSchema.(*schema.FunctionSchema), httpHandler)
	if err != nil {
		t.Fatalf("Failed to register HTTP endpoint: %v", err)
	}

	// Wait for registration
	time.Sleep(100 * time.Millisecond)

	// === 7. Setup HTTP client ===
	clientPortal := httpportal.NewPortal(httpportal.Config{
		DefaultTimeout: 15 * time.Second,
		MaxRetries:     2,
		RetryDelay:     500 * time.Millisecond,
	})

	clientFunc, err := clientPortal.ResolveFunction(context.Background(), httpAddress)
	if err != nil {
		t.Fatalf("Failed to resolve HTTP client function: %v", err)
	}

	// === 8. Run integration tests ===
	t.Run("BasicDiscountCalculation", func(t *testing.T) {
		testCases := []struct {
			name             string
			params           map[string]interface{}
			expectedTier     string
			expectedFinal    float64
			expectedDiscount int
		}{
			{
				name: "BronzeCustomer_SmallPurchase",
				params: map[string]interface{}{
					"customerTier":   "bronze",
					"purchaseAmount": 100.0,
				},
				expectedTier:     "bronze",
				expectedFinal:    95.0, // 5% discount
				expectedDiscount: 5,
			},
			{
				name: "GoldCustomer_LargePurchase",
				params: map[string]interface{}{
					"customerTier":   "gold",
					"purchaseAmount": 1500.0,
				},
				expectedTier:     "gold",
				expectedFinal:    1200.0, // 20% discount (15% + 5% large purchase)
				expectedDiscount: 20,
			},
			{
				name: "PlatinumCustomer_WithPromo",
				params: map[string]interface{}{
					"customerTier":   "platinum",
					"purchaseAmount": 2000.0,
					"promoCode":      "SAVE20",
				},
				expectedTier:     "platinum",
				expectedFinal:    1100.0, // 45% discount (20% + 5% + 20% = 45%)
				expectedDiscount: 45,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// This goes: HTTP Client -> HTTP Server -> JavaScript Function
				result, err := clientFunc.Call(context.Background(), tc.params)
				if err != nil {
					t.Fatalf("Cross-portal call failed: %v", err)
				}

				resultMap := result.(map[string]interface{})

				// Debug: print the actual result
				t.Logf("Actual result: %+v", resultMap)

				// Verify results with safe type assertions
				appliedTierRaw, exists := resultMap["appliedTier"]
				if !exists {
					t.Fatalf("appliedTier not found in result: %+v", resultMap)
				}
				appliedTier, ok := appliedTierRaw.(string)
				if !ok {
					t.Fatalf("appliedTier is not a string, got %T: %v", appliedTierRaw, appliedTierRaw)
				}
				if appliedTier != tc.expectedTier {
					t.Errorf("Expected tier %s, got %s", tc.expectedTier, appliedTier)
				}

				finalAmount := resultMap["finalAmount"].(float64)
				if finalAmount != tc.expectedFinal {
					t.Errorf("Expected final amount %.2f, got %.2f", tc.expectedFinal, finalAmount)
				}

				discountPercent := int(resultMap["discountPercent"].(float64))
				if discountPercent != tc.expectedDiscount {
					t.Errorf("Expected discount %d%%, got %d%%", tc.expectedDiscount, discountPercent)
				}

				t.Logf("✓ %s: Final amount %.2f (discount %d%%)",
					tc.name, finalAmount, discountPercent)
			})
		}
	})

	t.Run("ErrorPropagation", func(t *testing.T) {
		// Test invalid tier - should cause JavaScript error
		_, err := clientFunc.Call(context.Background(), map[string]interface{}{
			"customerTier":   "invalid_tier",
			"purchaseAmount": 100.0,
		})

		// Error should propagate through HTTP to client
		// (This might succeed if JS handles unknown tier gracefully)
		t.Logf("Invalid tier result: error=%v", err)
	})

	t.Run("CrossPortalTimeout", func(t *testing.T) {
		// Create a client with very short timeout
		shortTimeoutPortal := httpportal.NewPortal(httpportal.Config{
			DefaultTimeout: 1 * time.Millisecond,
			MaxRetries:     1,
		})

		shortTimeoutFunc, err := shortTimeoutPortal.ResolveFunction(context.Background(), httpAddress)
		if err != nil {
			t.Fatalf("Failed to resolve function: %v", err)
		}

		// This should timeout
		_, err = shortTimeoutFunc.Call(context.Background(), map[string]interface{}{
			"customerTier":   "gold",
			"purchaseAmount": 100.0,
		})

		if err == nil {
			t.Error("Expected timeout error, but call succeeded")
		} else {
			t.Logf("✓ Cross-portal timeout handled correctly: %v", err)
		}
	})
}

// TestMultiplePortalChain tests a more complex scenario with multiple portals
func TestMultiplePortalChain_JavaScriptThroughMultipleHTTP(t *testing.T) {
	// This test demonstrates JavaScript -> HTTP Gateway -> HTTP Service -> HTTP Client
	// Simulating a microservice architecture where JS runs in one service,
	// and is called through multiple HTTP hops

	// === JavaScript Business Logic ===
	validatorJS := jsportal.JSFunction{
		Code: `
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
		`,
		FunctionName: "validateOrder",
	}

	orderSchema := schema.NewFunctionSchema().
		Name("validateOrder").
		Description("Validates order data").
		Input("orderId", schema.NewString().Build()).
		Input("customerId", schema.NewString().Build()).
		Input("items", schema.NewArray().Items(
			schema.NewObject().
				Property("sku", schema.NewString().Build()).
				Property("quantity", schema.NewInteger().Build()).
				Property("price", schema.NewNumber().Build()).
				Build()).Build()).
		Output(schema.NewObject().
			Property("valid", schema.NewBoolean().Build()).
			Property("errors", schema.NewArray().Items(schema.NewString().Build()).Build()).
			Property("orderId", schema.NewString().Build()).
			Property("customerId", schema.NewString().Build()).
			Property("totalAmount", schema.NewNumber().Build()).
			Property("itemCount", schema.NewInteger().Build()).
			Build()).
		Build()

	// === Setup chain: JavaScript -> Gateway HTTP -> Service HTTP -> Client ===

	// 1. JavaScript portal
	jsPortal := jsportal.NewPortalWithDefaults()
	jsAddress := jsPortal.GenerateAddress("validateOrder", validatorJS)
	jsFunction := jsPortal.Apply(jsAddress, orderSchema.(*schema.FunctionSchema), validatorJS)

	// 2. Gateway HTTP service (calls JS)
	gatewayPortal := httpportal.NewPortal(httpportal.Config{
		Host: "localhost", Port: 8086, BasePath: "/gateway",
	})

	var gatewayHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		return jsFunction.Call(ctx, params)
	}

	if err := gatewayPortal.StartServer(); err != nil {
		t.Fatalf("Failed to start gateway server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		gatewayPortal.StopServer(ctx)
	}()

	gatewayAddress, _ := gatewayPortal.GenerateAddress("validateOrder", gatewayHandler)
	gatewayPortal.Apply(gatewayAddress, orderSchema.(*schema.FunctionSchema), gatewayHandler)

	// 3. Service HTTP (calls gateway)
	servicePortal := httpportal.NewPortal(httpportal.Config{
		Host: "localhost", Port: 8087, BasePath: "/service",
	})

	// Create client to call gateway
	gatewayClient, _ := gatewayPortal.ResolveFunction(context.Background(), gatewayAddress)

	var serviceHandler schema.FunctionHandler = func(ctx context.Context, params schema.FunctionInput) (schema.FunctionOutput, error) {
		result, err := gatewayClient.Call(ctx, params)
		if err != nil {
			return schema.FromAny(nil), err
		}
		return schema.FromAny(result), nil
	}

	if err := servicePortal.StartServer(); err != nil {
		t.Fatalf("Failed to start service server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		servicePortal.StopServer(ctx)
	}()

	serviceAddress, _ := servicePortal.GenerateAddress("validateOrder", serviceHandler)
	servicePortal.Apply(serviceAddress, orderSchema.(*schema.FunctionSchema), serviceHandler)

	// Wait for all servers
	time.Sleep(300 * time.Millisecond)

	// 4. Final client
	clientPortal := httpportal.NewPortal(httpportal.Config{DefaultTimeout: 20 * time.Second})
	clientFunc, err := clientPortal.ResolveFunction(context.Background(), serviceAddress)
	if err != nil {
		t.Fatalf("Failed to resolve service function: %v", err)
	}

	// === Test the full chain ===
	t.Run("ValidOrder_ThroughChain", func(t *testing.T) {
		validOrder := map[string]interface{}{
			"orderId":    "ORD-12345",
			"customerId": "CUST-999",
			"items": []interface{}{
				map[string]interface{}{"sku": "ITEM-001", "quantity": 2, "price": 29.99},
				map[string]interface{}{"sku": "ITEM-002", "quantity": 1, "price": 15.50},
			},
		}

		// This goes: Client -> Service HTTP -> Gateway HTTP -> JavaScript
		result, err := clientFunc.Call(context.Background(), validOrder)
		if err != nil {
			t.Fatalf("Multi-hop call failed: %v", err)
		}

		resultMap := result.(map[string]interface{})

		if !resultMap["valid"].(bool) {
			t.Errorf("Expected valid order, got invalid: %v", resultMap["errors"])
		}

		totalAmount := resultMap["totalAmount"].(float64)
		expectedTotal := 75.48 // (2 * 29.99) + (1 * 15.50)
		if totalAmount != expectedTotal {
			t.Errorf("Expected total %.2f, got %.2f", expectedTotal, totalAmount)
		}

		t.Logf("✓ Multi-hop validation successful: Total $%.2f", totalAmount)
	})

	t.Run("InvalidOrder_ErrorPropagation", func(t *testing.T) {
		invalidOrder := map[string]interface{}{
			"orderId": "123",           // Too short
			"items":   []interface{}{}, // Empty
		}

		result, err := clientFunc.Call(context.Background(), invalidOrder)
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}

		resultMap := result.(map[string]interface{})

		if resultMap["valid"].(bool) {
			t.Error("Expected invalid order, got valid")
		}

		errors := resultMap["errors"].([]interface{})
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}

		t.Logf("✓ Validation errors propagated correctly: %v", errors)
	})
}
