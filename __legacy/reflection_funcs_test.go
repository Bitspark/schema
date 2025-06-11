package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// Sample functions for testing
func AddNumbers(a int, b int) (int, error) {
	if a < 0 || b < 0 {
		return 0, fmt.Errorf("negative numbers not allowed")
	}
	return a + b, nil
}

func CreateUser(ctx context.Context, name string, email string, age int) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	agePtr := &age
	return &User{
		Name:  name,
		Email: email,
		Age:   agePtr,
	}, nil
}

func SimpleGreeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func NoParameters() string {
	return "Hello, World!"
}

func OnlyError() error {
	return fmt.Errorf("something went wrong")
}

// User type is defined in reflection_test.go

type TestOrder struct {
	ID       string  `json:"id"`
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
	Items    int     `json:"items"`
}

type TestOrderResult struct {
	ID        string    `json:"id"`
	Total     float64   `json:"total"`
	Tax       float64   `json:"tax"`
	OrderInfo TestOrder `json:"order_info"`
}

func ProcessTestOrder(ctx context.Context, order TestOrder) (*TestOrderResult, error) {
	tax := order.Amount * 0.08
	return &TestOrderResult{
		ID:        order.ID,
		Total:     order.Amount + tax,
		Tax:       tax,
		OrderInfo: order,
	}, nil
}

func TestFromFunction_BasicTypes(t *testing.T) {
	schema := FromFunction(AddNumbers)

	// Check function name
	if schema.metadata.Name != "AddNumbers" {
		t.Errorf("Expected function name 'AddNumbers', got '%s'", schema.metadata.Name)
	}

	// Check inputs
	inputs := schema.Inputs()
	if len(inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(inputs))
	}

	// Check that param0 and param1 exist and are integers
	if _, exists := inputs["param0"]; !exists {
		t.Error("Expected param0 to exist")
	}
	if _, exists := inputs["param1"]; !exists {
		t.Error("Expected param1 to exist")
	}

	// Check required fields
	required := schema.Required()
	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}

	// Check outputs
	if schema.outputs == nil {
		t.Error("Expected outputs to be set")
	}

	// Check errors
	if schema.errors == nil {
		t.Error("Expected errors to be set")
	}
}

func TestFromFunction_WithContext(t *testing.T) {
	schema := FromFunction(CreateUser)

	// Check function name
	if schema.metadata.Name != "CreateUser" {
		t.Errorf("Expected function name 'CreateUser', got '%s'", schema.metadata.Name)
	}

	// Check inputs (should skip context parameter)
	inputs := schema.Inputs()

	if len(inputs) != 3 {
		t.Errorf("Expected 3 inputs (excluding context), got %d", len(inputs))
	}

	// Check parameter types exist
	if _, exists := inputs["param0"]; !exists {
		t.Error("Expected param0 (name) to exist")
	}
	if _, exists := inputs["param1"]; !exists {
		t.Error("Expected param1 (email) to exist")
	}
	if _, exists := inputs["param2"]; !exists {
		t.Error("Expected param2 (age) to exist")
	}
}

func TestFromFunction_SingleReturn(t *testing.T) {
	schema := FromFunction(SimpleGreeting)

	jsonSchema := schema.ToJSONSchema()

	jsonSchemaJSON, _ := json.Marshal(jsonSchema)

	fmt.Printf("%s\n", string(jsonSchemaJSON))

	// Check inputs
	inputs := schema.Inputs()
	if len(inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(inputs))
	}

	// Check outputs (single string return)
	if schema.outputs == nil {
		t.Error("Expected outputs to be set")
	}

	// Check no errors (function doesn't return error)
	if schema.errors != nil {
		t.Error("Expected errors to be nil for function that doesn't return error")
	}
}

func TestFromFunction_NoParameters(t *testing.T) {
	schema := FromFunction(NoParameters)

	// Check no inputs
	inputs := schema.Inputs()
	if len(inputs) != 0 {
		t.Errorf("Expected 0 inputs, got %d", len(inputs))
	}

	// Check outputs
	if schema.outputs == nil {
		t.Error("Expected outputs to be set")
	}
}

func TestFromFunction_OnlyError(t *testing.T) {
	schema := FromFunction(OnlyError)

	// Check no inputs
	inputs := schema.Inputs()
	if len(inputs) != 0 {
		t.Errorf("Expected 0 inputs, got %d", len(inputs))
	}

	// Check no outputs
	if schema.outputs != nil {
		t.Error("Expected outputs to be nil for function that only returns error")
	}

	// Check errors are set
	if schema.errors == nil {
		t.Error("Expected errors to be set")
	}
}

func TestFunctionReflector_Call(t *testing.T) {
	reflector := NewFunctionReflector(AddNumbers)

	// Test successful call
	params := map[string]any{
		"param0": 5,
		"param1": 3,
	}

	result, err := reflector.Call(context.Background(), params)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.Value() != 8 {
		t.Errorf("Expected result 8, got %v", result)
	}

	// Test error case
	params = map[string]any{
		"param0": -1,
		"param1": 3,
	}

	result, err = reflector.Call(context.Background(), params)
	if err == nil {
		t.Error("Expected error for negative input")
	}

	if result.Value() != 0 {
		t.Errorf("Expected result 0 when error occurs, got %v", result)
	}
}

func TestFunctionReflector_CallWithContext(t *testing.T) {
	reflector := NewFunctionReflector(CreateUser)

	params := map[string]any{
		"param0": "John Doe",
		"param1": "john@example.com",
		"param2": 30,
	}

	ctx := context.Background()
	result, err := reflector.Call(ctx, params)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	user, ok := result.Value().(*User)
	if !ok {
		t.Errorf("Expected *User, got %T", result)
	}

	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
	}
	if user.Age == nil || *user.Age != 30 {
		t.Errorf("Expected age 30, got %v", user.Age)
	}
}

func TestFromFunction_JSONSchema(t *testing.T) {
	schema := FromFunction(AddNumbers)

	jsonSchema := schema.ToJSONSchema()

	// Check basic structure
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", jsonSchema["type"])
	}

	// Check properties exist
	properties, ok := jsonSchema["properties"].(map[string]any)
	if !ok {
		t.Error("Expected properties to be a map")
	}

	if len(properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(properties))
	}

	// Check required fields
	required, ok := jsonSchema["required"].([]string)
	if !ok {
		t.Error("Expected required to be a string slice")
	}

	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}
}

func TestStructParameterConversion(t *testing.T) {
	// Create function reflector
	reflector := NewFunctionReflector(ProcessTestOrder)

	// Test data as map[string]any (simulating JSON input)
	params := FromMap(map[string]any{
		"param0": map[string]any{
			"id":       "test-123",
			"customer": "John Doe",
			"amount":   100.0,
			"items":    5,
		},
	})

	// Call the function
	ctx := context.Background()
	result, err := reflector.Call(ctx, params)
	if err != nil {
		t.Fatalf("Function call failed: %v", err)
	}

	// The main test: verify struct parameter conversion worked
	// (the function successfully received and processed the struct)
	resultValue := result.Value()
	testResult, ok := resultValue.(*TestOrderResult)
	if !ok {
		t.Fatalf("Expected result to be *TestOrderResult, got %T", resultValue)
	}

	// Check that the struct input was properly converted
	if testResult.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %v", testResult.ID)
	}

	if testResult.Total != 108.0 {
		t.Errorf("Expected total 108.0, got %v", testResult.Total)
	}

	// Most importantly: check that the input struct was properly converted
	if testResult.OrderInfo.Customer != "John Doe" {
		t.Errorf("Expected customer 'John Doe', got %v", testResult.OrderInfo.Customer)
	}

	if testResult.OrderInfo.Amount != 100.0 {
		t.Errorf("Expected amount 100.0, got %v", testResult.OrderInfo.Amount)
	}

	t.Logf("✅ Struct parameter conversion working! Input struct properly converted.")
	t.Logf("   - ID: %s", testResult.ID)
	t.Logf("   - Total: %v", testResult.Total)
	t.Logf("   - Input Customer: %s", testResult.OrderInfo.Customer)
	t.Logf("   - Input Amount: %v", testResult.OrderInfo.Amount)
}
