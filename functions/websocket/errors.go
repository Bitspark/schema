package websocket

import (
	"fmt"
)

// WebSocketPortalError represents errors specific to the WebSocket portal
type WebSocketPortalError struct {
	Type    string                 `json:"type"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *WebSocketPortalError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("WebSocket Portal %s: %s - %s", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("WebSocket Portal %s: %s", e.Type, e.Message)
}

// WebSocketFunctionError represents errors from function execution over WebSocket
type WebSocketFunctionError struct {
	Code     int                    `json:"code"`
	Message  string                 `json:"message"`
	Address  string                 `json:"address,omitempty"`
	Function string                 `json:"function,omitempty"`
	Details  string                 `json:"details,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *WebSocketFunctionError) Error() string {
	location := e.Address
	if e.Function != "" {
		location = e.Function + " at " + e.Address
	}

	if e.Details != "" {
		return fmt.Sprintf("WebSocket Function Error [%d] at %s: %s - %s", e.Code, location, e.Message, e.Details)
	}
	return fmt.Sprintf("WebSocket Function Error [%d] at %s: %s", e.Code, location, e.Message)
}

// Connection errors
func NewConnectionError(address string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "ConnectionError",
		Code:    1001,
		Message: "Failed to establish WebSocket connection",
		Details: err.Error(),
		Context: map[string]interface{}{
			"address": address,
		},
	}
}

// Message errors
func NewMessageError(messageID string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "MessageError",
		Code:    1002,
		Message: "Failed to process WebSocket message",
		Details: err.Error(),
		Context: map[string]interface{}{
			"message_id": messageID,
		},
	}
}

// Protocol errors
func NewProtocolError(expected, actual string) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "ProtocolError",
		Code:    1003,
		Message: "WebSocket protocol violation",
		Details: fmt.Sprintf("expected %s, got %s", expected, actual),
	}
}

// Timeout errors
func NewTimeoutError(operation string, timeout interface{}) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "TimeoutError",
		Code:    1004,
		Message: "Operation timed out",
		Details: fmt.Sprintf("%s operation exceeded timeout %v", operation, timeout),
		Context: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
		},
	}
}

// Server errors
func NewServerError(err error, context map[string]interface{}) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "ServerError",
		Code:    1005,
		Message: "WebSocket server error",
		Details: err.Error(),
		Context: context,
	}
}

// Address errors
func NewAddressError(address string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "AddressError",
		Code:    1006,
		Message: "Invalid WebSocket address",
		Details: err.Error(),
		Context: map[string]interface{}{
			"address": address,
		},
	}
}

// Function registration errors
func NewRegistrationError(name string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "RegistrationError",
		Code:    1007,
		Message: "Failed to register function",
		Details: err.Error(),
		Context: map[string]interface{}{
			"function": name,
		},
	}
}

// Function call errors
func NewCallError(function, address string, err error) *WebSocketFunctionError {
	return &WebSocketFunctionError{
		Code:     500,
		Message:  "Function call failed",
		Address:  address,
		Function: function,
		Details:  err.Error(),
	}
}

// Client function errors
func NewClientFunctionError(address string, code int, message string) *WebSocketFunctionError {
	var errorType string
	switch {
	case code >= 400 && code < 500:
		errorType = "Client error"
	case code >= 500:
		errorType = "Server error"
	default:
		errorType = "Unknown error"
	}

	return &WebSocketFunctionError{
		Code:    code,
		Message: errorType,
		Address: address,
		Details: message,
	}
}

// Network errors
func NewNetworkError(operation string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "NetworkError",
		Code:    1008,
		Message: "Failed to establish network connection",
		Details: fmt.Sprintf("%s - %s", operation, err.Error()),
	}
}

// Validation errors
func NewValidationError(field, value string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "ValidationError",
		Code:    1009,
		Message: "Invalid input data",
		Details: fmt.Sprintf("field '%s' with value '%s': %s", field, value, err.Error()),
		Context: map[string]interface{}{
			"field": field,
			"value": value,
		},
	}
}

// Middleware errors
func NewMiddlewareError(middleware string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "MiddlewareError",
		Code:    1010,
		Message: "Middleware processing failed",
		Details: fmt.Sprintf("%s: %s", middleware, err.Error()),
		Context: map[string]interface{}{
			"middleware": middleware,
		},
	}
}

// Configuration errors
func NewConfigError(field string, err error) *WebSocketPortalError {
	return &WebSocketPortalError{
		Type:    "ConfigError",
		Code:    1011,
		Message: "Invalid configuration",
		Details: fmt.Sprintf("field '%s': %s", field, err.Error()),
		Context: map[string]interface{}{
			"field": field,
		},
	}
}
