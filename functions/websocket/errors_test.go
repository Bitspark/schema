package websocket

import (
	"errors"
	"testing"
)

func TestWebSocketErrorFunctions(t *testing.T) {
	t.Run("NewConnectionError", func(t *testing.T) {
		address := "ws://localhost:8080"
		originalErr := errors.New("connection failed")
		
		err := NewConnectionError(address, originalErr)
		
		if err.Type != "ConnectionError" {
			t.Errorf("Expected type 'ConnectionError', got '%s'", err.Type)
		}
		
		if err.Code != 1001 {
			t.Errorf("Expected code 1001, got %d", err.Code)
		}
		
		if err.Message != "Failed to establish WebSocket connection" {
			t.Errorf("Expected message 'Failed to establish WebSocket connection', got '%s'", err.Message)
		}
		
		if err.Details != originalErr.Error() {
			t.Errorf("Expected details '%s', got '%s'", originalErr.Error(), err.Details)
		}
		
		if err.Context["address"] != address {
			t.Errorf("Expected address in context, got %v", err.Context["address"])
		}
	})
	
	t.Run("NewMessageError", func(t *testing.T) {
		messageID := "msg-123"
		originalErr := errors.New("malformed JSON structure")
		
		err := NewMessageError(messageID, originalErr)
		
		if err.Type != "MessageError" {
			t.Errorf("Expected type 'MessageError', got '%s'", err.Type)
		}
		
		if err.Code != 1002 {
			t.Errorf("Expected code 1002, got %d", err.Code)
		}
		
		if err.Message != "Failed to process WebSocket message" {
			t.Errorf("Expected message 'Failed to process WebSocket message', got '%s'", err.Message)
		}
		
		if err.Details != originalErr.Error() {
			t.Errorf("Expected details '%s', got '%s'", originalErr.Error(), err.Details)
		}
		
		if err.Context["message_id"] != messageID {
			t.Errorf("Expected message_id in context, got %v", err.Context["message_id"])
		}
	})
	
	t.Run("NewProtocolError", func(t *testing.T) {
		expected := "1.1"
		actual := "1.0"
		
		err := NewProtocolError(expected, actual)
		
		if err.Type != "ProtocolError" {
			t.Errorf("Expected type 'ProtocolError', got '%s'", err.Type)
		}
		
		if err.Code != 1003 {
			t.Errorf("Expected code 1003, got %d", err.Code)
		}
		
		if err.Message != "WebSocket protocol violation" {
			t.Errorf("Expected message 'WebSocket protocol violation', got '%s'", err.Message)
		}
		
		expectedDetails := "expected 1.1, got 1.0"
		if err.Details != expectedDetails {
			t.Errorf("Expected details '%s', got '%s'", expectedDetails, err.Details)
		}
	})
	
	t.Run("NewServerError", func(t *testing.T) {
		originalErr := errors.New("server crashed")
		context := map[string]interface{}{
			"server_id": "ws-001",
			"uptime":    3600,
		}
		
		err := NewServerError(originalErr, context)
		
		if err.Type != "ServerError" {
			t.Errorf("Expected type 'ServerError', got '%s'", err.Type)
		}
		
		if err.Code != 1005 {
			t.Errorf("Expected code 1005, got %d", err.Code)
		}
		
		if err.Message != "WebSocket server error" {
			t.Errorf("Expected message 'WebSocket server error', got '%s'", err.Message)
		}
		
		if err.Details != originalErr.Error() {
			t.Errorf("Expected details '%s', got '%s'", originalErr.Error(), err.Details)
		}
		
		if err.Context["server_id"] != "ws-001" {
			t.Errorf("Expected server_id in context, got %v", err.Context["server_id"])
		}
	})
	
	t.Run("NewAddressError", func(t *testing.T) {
		address := "invalid://address"
		originalErr := errors.New("invalid scheme")
		
		err := NewAddressError(address, originalErr)
		
		if err.Type != "AddressError" {
			t.Errorf("Expected type 'AddressError', got '%s'", err.Type)
		}
		
		if err.Code != 1006 {
			t.Errorf("Expected code 1006, got %d", err.Code)
		}
		
		if err.Message != "Invalid WebSocket address" {
			t.Errorf("Expected message 'Invalid WebSocket address', got '%s'", err.Message)
		}
		
		if err.Details != originalErr.Error() {
			t.Errorf("Expected details '%s', got '%s'", originalErr.Error(), err.Details)
		}
		
		if err.Context["address"] != address {
			t.Errorf("Expected address in context, got %v", err.Context["address"])
		}
	})
	
	t.Run("NewRegistrationError", func(t *testing.T) {
		functionName := "calculateSum"
		originalErr := errors.New("function already registered")
		
		err := NewRegistrationError(functionName, originalErr)
		
		if err.Type != "RegistrationError" {
			t.Errorf("Expected type 'RegistrationError', got '%s'", err.Type)
		}
		
		if err.Code != 1007 {
			t.Errorf("Expected code 1007, got %d", err.Code)
		}
		
		if err.Message != "Failed to register function" {
			t.Errorf("Expected message 'Failed to register function', got '%s'", err.Message)
		}
		
		if err.Details != originalErr.Error() {
			t.Errorf("Expected details '%s', got '%s'", originalErr.Error(), err.Details)
		}
		
		if err.Context["function"] != functionName {
			t.Errorf("Expected function in context, got %v", err.Context["function"])
		}
	})
	
	t.Run("NewCallError", func(t *testing.T) {
		functionName := "testFunc"
		address := "ws://localhost:8080/func"
		originalErr := errors.New("function execution failed")
		
		err := NewCallError(functionName, address, originalErr)
		
		if err.Code != 500 {
			t.Errorf("Expected code 500, got %d", err.Code)
		}
		
		if err.Message != "Function call failed" {
			t.Errorf("Expected message 'Function call failed', got '%s'", err.Message)
		}
		
		if err.Address != address {
			t.Errorf("Expected address '%s', got '%s'", address, err.Address)
		}
		
		if err.Function != functionName {
			t.Errorf("Expected function '%s', got '%s'", functionName, err.Function)
		}
	})
	
	t.Run("NewClientFunctionError", func(t *testing.T) {
		address := "ws://localhost:8080/func"
		code := 404
		message := "Function not found"
		
		err := NewClientFunctionError(address, code, message)
		
		if err.Code != code {
			t.Errorf("Expected code %d, got %d", code, err.Code)
		}
		
		if err.Message != "Client error" {
			t.Errorf("Expected message 'Client error', got '%s'", err.Message)
		}
		
		if err.Address != address {
			t.Errorf("Expected address '%s', got '%s'", address, err.Address)
		}
		
		if err.Details != message {
			t.Errorf("Expected details '%s', got '%s'", message, err.Details)
		}
		
		// Test different error types
		serverErr := NewClientFunctionError(address, 500, "Internal error")
		if serverErr.Message != "Server error" {
			t.Errorf("Expected 'Server error' for 500 code, got '%s'", serverErr.Message)
		}
		
		unknownErr := NewClientFunctionError(address, 300, "Unknown")
		if unknownErr.Message != "Unknown error" {
			t.Errorf("Expected 'Unknown error' for 300 code, got '%s'", unknownErr.Message)
		}
	})
	
	t.Run("NewNetworkError", func(t *testing.T) {
		operation := "connect"
		originalErr := errors.New("network unreachable")
		
		err := NewNetworkError(operation, originalErr)
		
		if err.Type != "NetworkError" {
			t.Errorf("Expected type 'NetworkError', got '%s'", err.Type)
		}
		
		if err.Code != 1008 {
			t.Errorf("Expected code 1008, got %d", err.Code)
		}
		
		if err.Message != "Failed to establish network connection" {
			t.Errorf("Expected message 'Failed to establish network connection', got '%s'", err.Message)
		}
		
		expectedDetails := "connect - network unreachable"
		if err.Details != expectedDetails {
			t.Errorf("Expected details '%s', got '%s'", expectedDetails, err.Details)
		}
	})
	
	t.Run("NewValidationError", func(t *testing.T) {
		field := "input.amount"
		value := "-100"
		originalErr := errors.New("must be positive")
		
		err := NewValidationError(field, value, originalErr)
		
		if err.Type != "ValidationError" {
			t.Errorf("Expected type 'ValidationError', got '%s'", err.Type)
		}
		
		if err.Code != 1009 {
			t.Errorf("Expected code 1009, got %d", err.Code)
		}
		
		if err.Message != "Invalid input data" {
			t.Errorf("Expected message 'Invalid input data', got '%s'", err.Message)
		}
		
		expectedDetails := "field 'input.amount' with value '-100': must be positive"
		if err.Details != expectedDetails {
			t.Errorf("Expected details '%s', got '%s'", expectedDetails, err.Details)
		}
		
		if err.Context["field"] != field {
			t.Errorf("Expected field in context, got %v", err.Context["field"])
		}
		
		if err.Context["value"] != value {
			t.Errorf("Expected value in context, got %v", err.Context["value"])
		}
	})
	
	t.Run("NewMiddlewareError", func(t *testing.T) {
		middlewareName := "AuthMiddleware"
		originalErr := errors.New("token expired")
		
		err := NewMiddlewareError(middlewareName, originalErr)
		
		if err.Type != "MiddlewareError" {
			t.Errorf("Expected type 'MiddlewareError', got '%s'", err.Type)
		}
		
		if err.Code != 1010 {
			t.Errorf("Expected code 1010, got %d", err.Code)
		}
		
		if err.Message != "Middleware processing failed" {
			t.Errorf("Expected message 'Middleware processing failed', got '%s'", err.Message)
		}
		
		expectedDetails := "AuthMiddleware: token expired"
		if err.Details != expectedDetails {
			t.Errorf("Expected details '%s', got '%s'", expectedDetails, err.Details)
		}
		
		if err.Context["middleware"] != middlewareName {
			t.Errorf("Expected middleware in context, got %v", err.Context["middleware"])
		}
	})
	
	t.Run("NewConfigError", func(t *testing.T) {
		field := "max_connections"
		originalErr := errors.New("must be positive integer")
		
		err := NewConfigError(field, originalErr)
		
		if err.Type != "ConfigError" {
			t.Errorf("Expected type 'ConfigError', got '%s'", err.Type)
		}
		
		if err.Code != 1011 {
			t.Errorf("Expected code 1011, got %d", err.Code)
		}
		
		if err.Message != "Invalid configuration" {
			t.Errorf("Expected message 'Invalid configuration', got '%s'", err.Message)
		}
		
		expectedDetails := "field 'max_connections': must be positive integer"
		if err.Details != expectedDetails {
			t.Errorf("Expected details '%s', got '%s'", expectedDetails, err.Details)
		}
		
		if err.Context["field"] != field {
			t.Errorf("Expected field in context, got %v", err.Context["field"])
		}
	})
}

func TestWebSocketErrorImplementations(t *testing.T) {
	t.Run("WebSocketPortalError implements error interface", func(t *testing.T) {
		err := NewConnectionError("ws://test", errors.New("test"))
		
		var errorInterface error = err
		if errorInterface == nil {
			t.Error("WebSocketPortalError should implement error interface")
		}
		
		errorString := err.Error()
		if errorString == "" {
			t.Error("Error() should return non-empty string")
		}
		
		// Should contain error type and message
		if !containsSubstring(errorString, "ConnectionError") {
			t.Errorf("Error string should contain type, got: %s", errorString)
		}
	})
	
	t.Run("WebSocketFunctionError implements error interface", func(t *testing.T) {
		err := NewClientFunctionError("ws://test", 404, "Not found")
		
		var errorInterface error = err
		if errorInterface == nil {
			t.Error("WebSocketFunctionError should implement error interface")
		}
		
		errorString := err.Error()
		if errorString == "" {
			t.Error("Error() should return non-empty string")
		}
		
		// Should contain error code and message
		if !containsSubstring(errorString, "404") {
			t.Errorf("Error string should contain code, got: %s", errorString)
		}
	})
}

// Helper function to check if string contains substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && findSubstring(str, substr) != -1
}

func findSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}