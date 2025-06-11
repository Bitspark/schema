// Package integration - Reflection-based schema generation integration tests
//
// This file demonstrates how the reflection APIs (FromFunction, FromService, FromStruct)
// integrate with the portal system to provide automatic schema generation for Go code.
//
// Key capabilities demonstrated:
// - Automatic function schema generation from Go functions
// - Service method discovery and schema generation
// - HTTP portal integration with reflected schemas
// - Parameter type conversion and validation
// - End-to-end function calling through HTTP with reflection
//
// Note: Complex struct parameter conversion still needs investigation,
// but basic types and simple service methods work well.

package integration

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"defs.dev/schema"
	httpportal "defs.dev/schema/functions/http"
)

// Simple test types for reflection integration

type Order struct {
	ID       string  `json:"id" schema:"required,minlen=1"`
	Customer string  `json:"customer" schema:"required,minlen=1"`
	Amount   float64 `json:"amount" schema:"required,min=0"`
	Items    int     `json:"items" schema:"required,min=1"`
}

type OrderResult struct {
	ID        string    `json:"id"`
	Total     float64   `json:"total"`
	Tax       float64   `json:"tax"`
	Processed time.Time `json:"processed"`
}

// Simple functions for reflection testing

// CalculateTax calculates tax for an order amount
func CalculateTax(amount float64, rate float64) (float64, error) {
	if amount < 0 || rate < 0 || rate > 1 {
		return 0, fmt.Errorf("invalid amount or rate")
	}
	return math.Round(amount*rate*100) / 100, nil
}

// ProcessOrder processes an order and returns a result
func ProcessOrder(ctx context.Context, order Order) (*OrderResult, error) {
	if order.Amount <= 0 {
		return nil, fmt.Errorf("order amount must be positive")
	}

	tax, err := CalculateTax(order.Amount, 0.08) // 8% tax
	if err != nil {
		return nil, err
	}

	return &OrderResult{
		ID:        order.ID,
		Total:     order.Amount + tax,
		Tax:       tax,
		Processed: time.Now(),
	}, nil
}

// Simple service for reflection testing
type OrderService struct {
	processed int
}

func NewOrderService() *OrderService {
	return &OrderService{processed: 0}
}

func (s *OrderService) GetOrderCount(ctx context.Context) (int, error) {
	return s.processed, nil
}

func (s *OrderService) ProcessOrder(ctx context.Context, order Order) (*OrderResult, error) {
	result, err := ProcessOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	s.processed++
	return result, nil
}

// TestReflectionIntegration_BasicFunction tests a simple Go function
// with automatically generated schema through HTTP portal
func TestReflectionIntegration_BasicFunction(t *testing.T) {
	// === 1. Create a simple function ===
	calculateTax := func(ctx context.Context, income float64, rate float64) (float64, error) {
		if income < 0 || rate < 0 {
			return 0, fmt.Errorf("income and rate must be non-negative")
		}
		return income * rate, nil
	}

	// === 2. Create HTTP portal ===
	portal := httpportal.NewPortal(httpportal.Config{
		Host: "localhost",
		Port: 0,
	})

	// === 3. Start portal server ===
	err := portal.StartServer()
	if err != nil {
		t.Fatalf("Failed to start portal server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	// === 4. Register function with HTTP portal (STREAMLINED!) ===
	taxFunction, err := portal.RegisterFunction("calculateTax", calculateTax)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}
	if taxFunction == nil {
		t.Fatal("Function should not be nil")
	}

	// === 5. Resolve function as client ===
	ctx := context.Background()
	clientFunction, err := portal.ResolveFunction(ctx, taxFunction.Address())
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}
	if clientFunction == nil {
		t.Fatal("Client function should not be nil")
	}

	// === 6. Test function call via HTTP ===
	params := map[string]any{
		"param0": 100000.0, // income
		"param1": 0.15,     // rate
	}

	result, err := clientFunction.Call(ctx, params)
	if err != nil {
		t.Fatalf("Function call should succeed: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Verify the result
	expectedTax := 100000.0 * 0.15
	if result != expectedTax {
		t.Errorf("Tax calculation should be correct: expected %v, got %v", expectedTax, result)
	}

	// === 7. Test error case ===
	errorParams := map[string]any{
		"param0": -1000.0, // negative income
		"param1": 0.15,
	}

	_, err = clientFunction.Call(ctx, errorParams)
	if err == nil {
		t.Error("Should return error for negative income")
	}
}

// TestReflectionIntegration_StructFunction tests a function with struct parameters
// using automatically generated schemas
func TestReflectionIntegration_StructFunction(t *testing.T) {
	// === 1. Create HTTP portal ===
	portal := httpportal.NewPortal(httpportal.Config{
		Host: "localhost",
		Port: 0,
	})

	// === 2. Start portal server ===
	err := portal.StartServer()
	if err != nil {
		t.Fatalf("Failed to start portal server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	// === 3. Register function with struct parameter ===
	processOrderFunc, err := portal.RegisterFunction("processOrder", ProcessOrder)
	if err != nil {
		t.Fatalf("Failed to register ProcessOrder function: %v", err)
	}

	// === 4. Test the function call ===
	ctx := context.Background()
	clientFunc, err := portal.ResolveFunction(ctx, processOrderFunc.Address())
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	// === 5. Call with valid order struct data ===
	orderParams := map[string]any{
		"param0": map[string]any{
			"id":       "order-123",
			"customer": "John Doe",
			"amount":   100.50,
			"items":    3,
		},
	}

	result, err := clientFunc.Call(ctx, orderParams)
	if err != nil {
		t.Fatalf("ProcessOrder call should succeed: %v", err)
	}

	// === 6. Verify the result ===
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result)
	}

	// Check result fields
	if resultMap["id"] != "order-123" {
		t.Errorf("Expected ID 'order-123', got %v", resultMap["id"])
	}

	// Check calculated total (amount + 8% tax)
	expectedTotal := 100.50 * 1.08 // 8% tax
	if total, ok := resultMap["total"].(float64); !ok {
		t.Errorf("Expected total to be float64, got %T: %v", resultMap["total"], resultMap["total"])
	} else {
		// Use approximate comparison for floating point
		if diff := total - expectedTotal; diff > 0.001 || diff < -0.001 {
			t.Errorf("Expected total %v, got %v (diff: %v)", expectedTotal, total, diff)
		}
	}

	// Check tax
	expectedTax := 100.50 * 0.08
	if tax, ok := resultMap["tax"].(float64); !ok {
		t.Errorf("Expected tax to be float64, got %T: %v", resultMap["tax"], resultMap["tax"])
	} else {
		// Use approximate comparison for floating point
		if diff := tax - expectedTax; diff > 0.001 || diff < -0.001 {
			t.Errorf("Expected tax %v, got %v (diff: %v)", expectedTax, tax, diff)
		}
	}

	// Check that processed time exists
	if _, ok := resultMap["processed"]; !ok {
		t.Error("Expected processed timestamp to be present")
	}

	t.Logf("✓ Struct parameter conversion working: %+v", resultMap)

	// === 7. Test error case ===
	errorParams := map[string]any{
		"param0": map[string]any{
			"id":       "error-order",
			"customer": "Bad Customer",
			"amount":   -50.0, // Negative amount should cause error
			"items":    1,
		},
	}

	_, err = clientFunc.Call(ctx, errorParams)
	if err == nil {
		t.Error("Should return error for negative amount")
	} else {
		t.Logf("✓ Error handling works: %v", err)
	}
}

// MathService is a test service for integration testing
type MathService struct {
	multiplier float64
}

func (ms *MathService) Add(a, b float64) float64 {
	return a + b
}

func (ms *MathService) Multiply(ctx context.Context, value float64) (float64, error) {
	if value < 0 {
		return 0, fmt.Errorf("value must be non-negative")
	}
	return value * ms.multiplier, nil
}

func (ms *MathService) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil
}

// TestReflectionIntegration_Service tests a Go service (struct with methods)
// using automatically generated schemas
func TestReflectionIntegration_Service(t *testing.T) {
	// === 1. Create a service instance ===
	mathService := &MathService{multiplier: 2.5}

	// === 2. Create HTTP portal ===
	portal := httpportal.NewPortal(httpportal.Config{
		Host: "localhost",
		Port: 0,
	})

	// === 3. Start portal server ===
	err := portal.StartServer()
	if err != nil {
		t.Fatalf("Failed to start portal server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		portal.StopServer(ctx)
	}()

	// === 4. Register service with HTTP portal (STREAMLINED!) ===
	serviceFunctions, err := portal.RegisterService(mathService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}
	if len(serviceFunctions) != 3 {
		t.Errorf("Should register 3 methods, got %d", len(serviceFunctions))
	}

	// Verify all methods are registered
	if _, exists := serviceFunctions["Add"]; !exists {
		t.Error("Add method should be registered")
	}
	if _, exists := serviceFunctions["Multiply"]; !exists {
		t.Error("Multiply method should be registered")
	}
	if _, exists := serviceFunctions["Divide"]; !exists {
		t.Error("Divide method should be registered")
	}

	// === 5. Test each method via HTTP ===
	ctx := context.Background()

	// Test Add method
	addFunction, err := portal.ResolveFunction(ctx, serviceFunctions["Add"].Address())
	if err != nil {
		t.Fatalf("Failed to resolve Add function: %v", err)
	}

	addResult, err := addFunction.Call(ctx, map[string]any{
		"param0": 5.0,
		"param1": 3.0,
	})
	if err != nil {
		t.Fatalf("Add call should succeed: %v", err)
	}
	if addResult != 8.0 {
		t.Errorf("5 + 3 should equal 8, got %v", addResult)
	}

	// Test Multiply method
	multiplyFunction, err := portal.ResolveFunction(ctx, serviceFunctions["Multiply"].Address())
	if err != nil {
		t.Fatalf("Failed to resolve Multiply function: %v", err)
	}

	multiplyResult, err := multiplyFunction.Call(ctx, map[string]any{
		"param0": 4.0,
	})
	if err != nil {
		t.Fatalf("Multiply call should succeed: %v", err)
	}
	if multiplyResult != 10.0 {
		t.Errorf("4 * 2.5 should equal 10, got %v", multiplyResult)
	}

	// Test Divide method
	divideFunction, err := portal.ResolveFunction(ctx, serviceFunctions["Divide"].Address())
	if err != nil {
		t.Fatalf("Failed to resolve Divide function: %v", err)
	}

	divideResult, err := divideFunction.Call(ctx, map[string]any{
		"param0": 15.0,
		"param1": 3.0,
	})
	if err != nil {
		t.Fatalf("Divide call should succeed: %v", err)
	}
	if divideResult != 5.0 {
		t.Errorf("15 / 3 should equal 5, got %v", divideResult)
	}

	// Test error case for Divide
	_, err = divideFunction.Call(ctx, map[string]any{
		"param0": 10.0,
		"param1": 0.0,
	})
	if err == nil {
		t.Error("Should return error for division by zero")
	}
}

// TestReflectionIntegration_SchemaGeneration tests that reflection properly
// generates schemas with struct tags and validation
func TestReflectionIntegration_SchemaGeneration(t *testing.T) {
	// Test struct schema generation
	orderSchema := schema.FromStruct[Order]()

	// Verify schema properties
	objSchema, ok := orderSchema.(*schema.ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema, got %T", orderSchema)
	}

	properties := objSchema.Properties()
	requiredFields := objSchema.Required()

	// Check all fields are present
	expectedFields := []string{"id", "customer", "amount", "items"}
	for _, field := range expectedFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("Field %s not found in schema", field)
		}
	}

	// Check required fields
	if len(requiredFields) != 4 {
		t.Errorf("Expected 4 required fields, got %d", len(requiredFields))
	}

	// Test function schema generation
	taxSchema := schema.FromFunction(CalculateTax)

	// Verify function inputs
	inputs := taxSchema.Inputs()
	if len(inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(inputs))
	}

	// Check parameter names (should be param0, param1)
	if _, exists := inputs["param0"]; !exists {
		t.Error("Expected param0 not found")
	}
	if _, exists := inputs["param1"]; !exists {
		t.Error("Expected param1 not found")
	}

	// Verify function output
	if taxSchema.Outputs() == nil {
		t.Error("Expected function to have output schema")
	}

	t.Logf("✓ Schema generation working correctly")
	t.Logf("  - Struct fields: %v", expectedFields)
	t.Logf("  - Function inputs: %d", len(inputs))
	t.Logf("  - Function has output: %v", taxSchema.Outputs() != nil)
}
