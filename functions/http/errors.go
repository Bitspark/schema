package http

import (
	"fmt"
	"net/http"
)

// HTTPPortalError represents an error from the HTTP portal
type HTTPPortalError struct {
	Type    string
	Message string
	Cause   error
	Details map[string]interface{}
}

func (e *HTTPPortalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("HTTP Portal %s: %s - %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("HTTP Portal %s: %s", e.Type, e.Message)
}

func (e *HTTPPortalError) Unwrap() error {
	return e.Cause
}

// HTTPFunctionError represents an error from an HTTP function call
type HTTPFunctionError struct {
	StatusCode int
	Message    string
	Response   string
	Address    string
}

func (e *HTTPFunctionError) Error() string {
	return fmt.Sprintf("HTTP Function Error [%d] at %s: %s", e.StatusCode, e.Address, e.Message)
}

// Error constructors

// NewNetworkError creates an error for network-related issues
func NewNetworkError(cause error, details map[string]interface{}) *HTTPPortalError {
	return &HTTPPortalError{
		Type:    "NetworkError",
		Message: "Failed to establish network connection",
		Cause:   cause,
		Details: details,
	}
}

// NewTimeoutError creates an error for timeout issues
func NewTimeoutError(timeout string, address string) *HTTPPortalError {
	return &HTTPPortalError{
		Type:    "TimeoutError",
		Message: fmt.Sprintf("Request timed out after %s", timeout),
		Details: map[string]interface{}{
			"timeout": timeout,
			"address": address,
		},
	}
}

// NewServerError creates an error for HTTP server setup issues
func NewServerError(cause error, details map[string]interface{}) *HTTPPortalError {
	return &HTTPPortalError{
		Type:    "ServerError",
		Message: "Failed to configure or start HTTP server",
		Cause:   cause,
		Details: details,
	}
}

// NewAddressError creates an error for invalid address issues
func NewAddressError(address string, cause error) *HTTPPortalError {
	return &HTTPPortalError{
		Type:    "AddressError",
		Message: fmt.Sprintf("Invalid or malformed address: %s", address),
		Cause:   cause,
		Details: map[string]interface{}{
			"address": address,
		},
	}
}

// NewSerializationError creates an error for JSON encoding/decoding issues
func NewSerializationError(operation string, cause error) *HTTPPortalError {
	return &HTTPPortalError{
		Type:    "SerializationError",
		Message: fmt.Sprintf("Failed to %s JSON data", operation),
		Cause:   cause,
		Details: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewHTTPError creates an error for HTTP response issues
func NewHTTPError(statusCode int, message string, address string, response string) *HTTPFunctionError {
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return &HTTPFunctionError{
		StatusCode: statusCode,
		Message:    message,
		Address:    address,
		Response:   response,
	}
}

// NewClientError creates an error for HTTP 4xx client errors
func NewClientError(statusCode int, address string, response string) *HTTPFunctionError {
	return NewHTTPError(statusCode, "Client error", address, response)
}

// NewServerHTTPError creates an error for HTTP 5xx server errors
func NewServerHTTPError(statusCode int, address string, response string) *HTTPFunctionError {
	return NewHTTPError(statusCode, "Server error", address, response)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	switch e := err.(type) {
	case *HTTPPortalError:
		return e.Type == "NetworkError" || e.Type == "TimeoutError"
	case *HTTPFunctionError:
		// Retry on 5xx server errors, but not 4xx client errors
		return e.StatusCode >= 500
	default:
		return false
	}
}

// IsClientError checks if an error is a client error (4xx)
func IsClientError(err error) bool {
	if httpErr, ok := err.(*HTTPFunctionError); ok {
		return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
	}
	return false
}

// IsServerError checks if an error is a server error (5xx)
func IsServerError(err error) bool {
	if httpErr, ok := err.(*HTTPFunctionError); ok {
		return httpErr.StatusCode >= 500
	}
	return false
}
