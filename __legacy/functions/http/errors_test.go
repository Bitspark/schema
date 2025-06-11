package http

import (
	"errors"
	"testing"
)

func TestHTTPErrorFunctions(t *testing.T) {
	t.Run("HTTPPortalError Unwrap", func(t *testing.T) {
		originalErr := errors.New("original error")
		httpErr := &HTTPPortalError{
			Type:    "TestError",
			Message: "HTTP error occurred",
			Cause:   originalErr,
		}

		unwrapped := httpErr.Unwrap()
		if unwrapped != originalErr {
			t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
		}
	})

	t.Run("NewTimeoutError", func(t *testing.T) {
		timeout := "30s"
		address := "http://example.com"

		err := NewTimeoutError(timeout, address)

		if err.Type != "TimeoutError" {
			t.Errorf("Expected type 'TimeoutError', got '%s'", err.Type)
		}

		expectedMessage := "Request timed out after 30s"
		if err.Message != expectedMessage {
			t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
		}

		if err.Details["timeout"] != timeout {
			t.Errorf("Expected timeout in details, got %v", err.Details["timeout"])
		}

		if err.Details["address"] != address {
			t.Errorf("Expected address in details, got %v", err.Details["address"])
		}
	})

	t.Run("NewServerError", func(t *testing.T) {
		originalErr := errors.New("server crashed")
		details := map[string]interface{}{
			"server_id": "http-001",
			"uptime":    3600,
		}

		err := NewServerError(originalErr, details)

		if err.Type != "ServerError" {
			t.Errorf("Expected type 'ServerError', got '%s'", err.Type)
		}

		if err.Message != "Failed to configure or start HTTP server" {
			t.Errorf("Expected message 'Failed to configure or start HTTP server', got '%s'", err.Message)
		}

		if err.Cause != originalErr {
			t.Errorf("Expected cause to be original error, got %v", err.Cause)
		}

		if err.Details["server_id"] != "http-001" {
			t.Errorf("Expected server_id in details, got %v", err.Details["server_id"])
		}
	})

	t.Run("NewAddressError", func(t *testing.T) {
		address := "invalid://address"
		originalErr := errors.New("invalid scheme")

		err := NewAddressError(address, originalErr)

		if err.Type != "AddressError" {
			t.Errorf("Expected type 'AddressError', got '%s'", err.Type)
		}

		expectedMessage := "Invalid or malformed address: invalid://address"
		if err.Message != expectedMessage {
			t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
		}

		if err.Cause != originalErr {
			t.Errorf("Expected cause to be original error, got %v", err.Cause)
		}

		if err.Details["address"] != address {
			t.Errorf("Expected address in details, got %v", err.Details["address"])
		}
	})

	t.Run("NewSerializationError", func(t *testing.T) {
		operation := "marshal"
		originalErr := errors.New("invalid character")

		err := NewSerializationError(operation, originalErr)

		if err.Type != "SerializationError" {
			t.Errorf("Expected type 'SerializationError', got '%s'", err.Type)
		}

		expectedMessage := "Failed to marshal JSON data"
		if err.Message != expectedMessage {
			t.Errorf("Expected message '%s', got '%s'", expectedMessage, err.Message)
		}

		if err.Cause != originalErr {
			t.Errorf("Expected cause to be original error, got %v", err.Cause)
		}

		if err.Details["operation"] != operation {
			t.Errorf("Expected operation in details, got %v", err.Details["operation"])
		}
	})

	t.Run("NewServerHTTPError", func(t *testing.T) {
		statusCode := 404
		address := "http://example.com/api"
		response := "Not Found"

		err := NewServerHTTPError(statusCode, address, response)

		if err.StatusCode != statusCode {
			t.Errorf("Expected status code %d, got %d", statusCode, err.StatusCode)
		}

		if err.Message != "Server error" {
			t.Errorf("Expected message 'Server error', got '%s'", err.Message)
		}

		if err.Address != address {
			t.Errorf("Expected address '%s', got '%s'", address, err.Address)
		}

		if err.Response != response {
			t.Errorf("Expected response '%s', got '%s'", response, err.Response)
		}
	})

	t.Run("IsRetryableError", func(t *testing.T) {
		// Test retryable errors
		retryableErr := &HTTPPortalError{Type: "NetworkError"}
		if !IsRetryableError(retryableErr) {
			t.Error("Expected NetworkError to be retryable")
		}

		retryableErr = &HTTPPortalError{Type: "TimeoutError"}
		if !IsRetryableError(retryableErr) {
			t.Error("Expected TimeoutError to be retryable")
		}

		retryableHttpErr := &HTTPFunctionError{StatusCode: 502}
		if !IsRetryableError(retryableHttpErr) {
			t.Error("Expected 502 error to be retryable")
		}

		retryableHttpErr = &HTTPFunctionError{StatusCode: 503}
		if !IsRetryableError(retryableHttpErr) {
			t.Error("Expected 503 error to be retryable")
		}

		retryableHttpErr = &HTTPFunctionError{StatusCode: 504}
		if !IsRetryableError(retryableHttpErr) {
			t.Error("Expected 504 error to be retryable")
		}

		// Test non-retryable errors
		nonRetryableErr := &HTTPPortalError{Type: "AddressError"}
		if IsRetryableError(nonRetryableErr) {
			t.Error("Expected AddressError to not be retryable")
		}

		nonRetryableHttpErr := &HTTPFunctionError{StatusCode: 404}
		if IsRetryableError(nonRetryableHttpErr) {
			t.Error("Expected 404 error to not be retryable")
		}

		// Test other error types
		otherErr := errors.New("network error")
		if IsRetryableError(otherErr) {
			t.Error("Expected non-HTTP error to not be retryable")
		}
	})

	t.Run("IsClientError", func(t *testing.T) {
		// Test client errors (4xx)
		clientErr := &HTTPFunctionError{StatusCode: 400}
		if !IsClientError(clientErr) {
			t.Error("Expected 400 error to be client error")
		}

		clientErr = &HTTPFunctionError{StatusCode: 404}
		if !IsClientError(clientErr) {
			t.Error("Expected 404 error to be client error")
		}

		clientErr = &HTTPFunctionError{StatusCode: 499}
		if !IsClientError(clientErr) {
			t.Error("Expected 499 error to be client error")
		}

		// Test non-client errors
		serverErr := &HTTPFunctionError{StatusCode: 500}
		if IsClientError(serverErr) {
			t.Error("Expected 500 error to not be client error")
		}

		otherErr := errors.New("network error")
		if IsClientError(otherErr) {
			t.Error("Expected non-HTTP error to not be client error")
		}
	})

	t.Run("IsServerError", func(t *testing.T) {
		// Test server errors (5xx)
		serverErr := &HTTPFunctionError{StatusCode: 500}
		if !IsServerError(serverErr) {
			t.Error("Expected 500 error to be server error")
		}

		serverErr = &HTTPFunctionError{StatusCode: 502}
		if !IsServerError(serverErr) {
			t.Error("Expected 502 error to be server error")
		}

		serverErr = &HTTPFunctionError{StatusCode: 599}
		if !IsServerError(serverErr) {
			t.Error("Expected 599 error to be server error")
		}

		// Test non-server errors
		clientErr := &HTTPFunctionError{StatusCode: 404}
		if IsServerError(clientErr) {
			t.Error("Expected 404 error to not be server error")
		}

		otherErr := errors.New("network error")
		if IsServerError(otherErr) {
			t.Error("Expected non-HTTP error to not be server error")
		}
	})
}
