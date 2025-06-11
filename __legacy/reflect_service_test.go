package schema

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

// Test service structs

type UserService struct {
	DB     string `json:"db"`
	Logger string `json:"logger"`
	APIKey string `json:"apiKey"`
}

func (s *UserService) CreateUser(ctx context.Context, name string, email string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	return &User{
		Name:  name,
		Email: email,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return &User{
		Name:  "John Doe",
		Email: "john@example.com",
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, updates User) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return &updates, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return nil
}

func (s *UserService) ListUsers(ctx context.Context, limit *int, offset *int) ([]*User, error) {
	users := []*User{
		{Name: "User 1", Email: "user1@example.com"},
		{Name: "User 2", Email: "user2@example.com"},
	}
	return users, nil
}

// Private method - should be ignored
func (s *UserService) validateUser(user *User) error {
	return nil
}

// Method that returns a function - should be ignored
func (s *UserService) GetValidator() func(*User) error {
	return s.validateUser
}

type SimpleService struct {
	value string
}

func (s *SimpleService) Echo(message string) string {
	return fmt.Sprintf("Echo: %s", message)
}

func (s *SimpleService) Add(a, b int) int {
	return a + b
}

func TestFromService_Basic(t *testing.T) {
	userService := &UserService{
		DB:     "test-db",
		Logger: "test-logger",
		APIKey: "test-key",
	}

	serviceReflector := FromService(userService)

	// Check service type
	expectedTypeName := "UserService"
	actualTypeName := serviceReflector.ServiceType().Name()
	if serviceReflector.ServiceType().Kind() == reflect.Ptr {
		actualTypeName = serviceReflector.ServiceType().Elem().Name()
	}
	if actualTypeName != expectedTypeName {
		t.Errorf("Expected service type '%s', got '%s'", expectedTypeName, actualTypeName)
	}

	// Check methods discovered
	methods := serviceReflector.MethodNames()
	expectedMethods := []string{"CreateUser", "GetUser", "UpdateUser", "DeleteUser", "ListUsers"}

	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d: %v", len(expectedMethods), len(methods), methods)
	}

	// Check that private and invalid methods are excluded
	for _, method := range methods {
		if method == "validateUser" || method == "GetValidator" {
			t.Errorf("Private or invalid method '%s' should not be included", method)
		}
	}

	// Check that expected methods are present
	for _, expected := range expectedMethods {
		if !serviceReflector.HasMethod(expected) {
			t.Errorf("Expected method '%s' not found", expected)
		}
	}
}

func TestFromService_ValueStruct(t *testing.T) {
	// Test with value struct (not pointer)
	simpleService := SimpleService{value: "test"}
	serviceReflector := FromService(simpleService)

	methods := serviceReflector.MethodNames()

	// Value structs should also work, but might have fewer methods
	// depending on whether methods are defined on value or pointer receiver
	if len(methods) == 0 {
		t.Error("Expected at least some methods on value struct")
	}
}

func TestFromService_Panics(t *testing.T) {
	// Test nil input
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil input")
		}
	}()
	FromService(nil)
}

func TestFromService_InvalidType(t *testing.T) {
	// Test non-struct input
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-struct input")
		}
	}()
	FromService("not a struct")
}

func TestServiceReflector_Call(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	// Test CreateUser call
	ctx := context.Background()
	params := map[string]any{
		"param0": "John Doe",
		"param1": "john@example.com",
	}

	result, err := serviceReflector.Call("CreateUser", ctx, params)
	if err != nil {
		t.Errorf("Unexpected error calling CreateUser: %v", err)
	}

	user, ok := result.(*User)
	if !ok {
		t.Errorf("Expected *User result, got %T", result)
	}

	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
	}
}

func TestServiceReflector_CallWithError(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	// Test CreateUser with empty name (should cause error)
	ctx := context.Background()
	params := map[string]any{
		"param0": "", // empty name
		"param1": "john@example.com",
	}

	result, err := serviceReflector.Call("CreateUser", ctx, params)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Result could be nil or the zero value, both are acceptable on error
	if result != nil && result != (*User)(nil) {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

func TestServiceReflector_CallNonExistentMethod(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	ctx := context.Background()
	params := map[string]any{}

	result, err := serviceReflector.Call("NonExistentMethod", ctx, params)
	if err == nil {
		t.Error("Expected error for non-existent method")
	}

	if result != nil {
		t.Errorf("Expected nil result for non-existent method, got %v", result)
	}
}

func TestServiceReflector_Functions(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	functions := serviceReflector.Functions()

	// Check that we get FunctionReflector instances
	if len(functions) == 0 {
		t.Error("Expected at least one function")
	}

	for methodName, reflector := range functions {
		if reflector == nil {
			t.Errorf("Function reflector for %s is nil", methodName)
		}

		// Check that we can get schema from the reflector
		schema := reflector.Schema()
		if schema == nil {
			t.Errorf("Schema for %s is nil", methodName)
		}
	}
}

func TestServiceReflector_Schemas(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	schemas := serviceReflector.Schemas()

	// Check CreateUser schema
	createUserSchema, exists := schemas["CreateUser"]
	if !exists {
		t.Error("CreateUser schema not found")
	}

	inputs := createUserSchema.Inputs()
	if len(inputs) != 2 {
		t.Errorf("Expected 2 inputs for CreateUser, got %d", len(inputs))
	}

	// Check parameter names (should be param0, param1 since context is excluded)
	if _, exists := inputs["param0"]; !exists {
		t.Error("Expected param0 (name) to exist in CreateUser inputs")
	}
	if _, exists := inputs["param1"]; !exists {
		t.Error("Expected param1 (email) to exist in CreateUser inputs")
	}

	// Check outputs and errors
	if createUserSchema.Outputs() == nil {
		t.Error("Expected CreateUser to have outputs")
	}
	if createUserSchema.Errors() == nil {
		t.Error("Expected CreateUser to have errors")
	}
}

func TestServiceReflector_ServiceSchema(t *testing.T) {
	userService := &UserService{
		DB:     "test-db",
		Logger: "test-logger",
		APIKey: "test-key",
	}
	serviceReflector := FromService(userService)

	serviceSchema := serviceReflector.ServiceSchema()
	if serviceSchema == nil {
		t.Error("Service schema is nil")
	}

	// Check that the service schema is an object schema
	if serviceSchema.Type() != TypeObject {
		t.Errorf("Expected service schema type to be Object, got %s", serviceSchema.Type())
	}

	// The service schema should include the struct fields
	properties := serviceSchema.Properties()
	expectedFields := []string{"db", "logger", "apiKey"} // These are the JSON field names

	for _, field := range expectedFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("Expected service field '%s' not found in schema", field)
		}
	}
}

func TestServiceReflector_ServiceInfo(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	serviceInfo := serviceReflector.ServiceInfo()

	if serviceInfo.Name != "user_service" {
		t.Errorf("Expected service name 'user_service', got '%s'", serviceInfo.Name)
	}

	if serviceInfo.MethodCount == 0 {
		t.Error("Expected non-zero method count")
	}

	if len(serviceInfo.Methods) != serviceInfo.MethodCount {
		t.Errorf("Method count mismatch: reported %d, found %d", serviceInfo.MethodCount, len(serviceInfo.Methods))
	}

	// Check method info details
	createUserInfo, exists := serviceInfo.Methods["CreateUser"]
	if !exists {
		t.Error("CreateUser method info not found")
	}

	if createUserInfo.InputCount != 2 {
		t.Errorf("Expected CreateUser input count 2, got %d", createUserInfo.InputCount)
	}

	if createUserInfo.OutputCount != 1 {
		t.Errorf("Expected CreateUser output count 1, got %d", createUserInfo.OutputCount)
	}

	if !createUserInfo.HasError {
		t.Error("Expected CreateUser to have error")
	}
}

func TestServiceReflector_WithPointerMethod(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	// Test ListUsers which has pointer parameters
	ctx := context.Background()
	params := map[string]any{
		"param0": nil, // limit (*int)
		"param1": nil, // offset (*int)
	}

	result, err := serviceReflector.Call("ListUsers", ctx, params)
	if err != nil {
		t.Errorf("Unexpected error calling ListUsers: %v", err)
	}

	users, ok := result.([]*User)
	if !ok {
		t.Errorf("Expected []*User result, got %T", result)
	}

	if len(users) == 0 {
		t.Error("Expected at least one user in result")
	}
}

func TestServiceConvenienceFunctions(t *testing.T) {
	userService := &UserService{}

	// Test ListMethods
	methods := ListMethods(userService)
	if len(methods) == 0 {
		t.Error("Expected at least one method")
	}

	// Test GetMethodSchema
	schema, err := GetMethodSchema(userService, "CreateUser")
	if err != nil {
		t.Errorf("Unexpected error getting method schema: %v", err)
	}
	if schema == nil {
		t.Error("Expected non-nil schema")
	}

	// Test GetMethodSchema for non-existent method
	_, err = GetMethodSchema(userService, "NonExistentMethod")
	if err == nil {
		t.Error("Expected error for non-existent method")
	}

	// Test CallMethod
	ctx := context.Background()
	params := map[string]any{
		"param0": "John Doe",
		"param1": "john@example.com",
	}

	result, err := CallMethod(userService, "CreateUser", ctx, params)
	if err != nil {
		t.Errorf("Unexpected error calling method: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestValidateService(t *testing.T) {
	// Test valid service
	userService := &UserService{}
	err := ValidateService(userService)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	// Test nil service
	err = ValidateService(nil)
	if err == nil {
		t.Error("Expected validation error for nil service")
	}

	// Test non-struct
	err = ValidateService("not a struct")
	if err == nil {
		t.Error("Expected validation error for non-struct")
	}

	// Test struct with no valid methods
	type EmptyStruct struct{}
	empty := &EmptyStruct{}
	err = ValidateService(empty)
	if err == nil {
		t.Error("Expected validation error for struct with no methods")
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserService", "user_service"},
		{"APIClient", "a_p_i_client"},
		{"SimpleStruct", "simple_struct"},
		{"A", "a"},
		{"", ""},
	}

	for _, test := range tests {
		result := toSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("toSnakeCase(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestServiceReflector_MethodReturn_OnlyError(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	// Test DeleteUser which returns only error
	schemas := serviceReflector.Schemas()
	deleteUserSchema := schemas["DeleteUser"]

	// Should have no outputs but should have errors
	if deleteUserSchema.Outputs() != nil {
		t.Error("Expected DeleteUser to have no outputs")
	}

	if deleteUserSchema.Errors() == nil {
		t.Error("Expected DeleteUser to have errors")
	}
}

func TestServiceReflector_MethodWithStructParameter(t *testing.T) {
	userService := &UserService{}
	serviceReflector := FromService(userService)

	// Test UpdateUser which has a struct parameter
	ctx := context.Background()
	updates := User{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	params := map[string]any{
		"param0": int64(123), // id
		"param1": updates,    // User struct
	}

	result, err := serviceReflector.Call("UpdateUser", ctx, params)
	if err != nil {
		t.Errorf("Unexpected error calling UpdateUser: %v", err)
	}

	updatedUser, ok := result.(*User)
	if !ok {
		t.Errorf("Expected *User result, got %T", result)
	}

	if updatedUser.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updatedUser.Name)
	}
}

// Test struct with methods that have no context parameter
type MathService struct{}

func (m *MathService) Add(a, b int) int {
	return a + b
}

func (m *MathService) Multiply(a, b float64) float64 {
	return a * b
}

func TestServiceReflector_NoContextMethods(t *testing.T) {
	mathService := &MathService{}
	serviceReflector := FromService(mathService)

	// Test Add method (no context)
	params := map[string]any{
		"param0": 5,
		"param1": 3,
	}

	result, err := serviceReflector.Call("Add", context.Background(), params)
	if err != nil {
		t.Errorf("Unexpected error calling Add: %v", err)
	}

	sum, ok := result.(int)
	if !ok {
		t.Errorf("Expected int result, got %T", result)
	}

	if sum != 8 {
		t.Errorf("Expected sum 8, got %d", sum)
	}
}
