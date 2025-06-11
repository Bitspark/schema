package javascript

import (
	"fmt"
	"strings"

	"defs.dev/schema"
)

// JSFunctionError represents errors during JavaScript function execution
type JSFunctionError struct {
	Function string
	Stage    string
	Message  string
	Errors   []schema.ValidationError
	Cause    error
}

func (e *JSFunctionError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("javascript function '%s' failed at %s: %s", e.Function, e.Stage, e.Message)
	}

	if len(e.Errors) > 0 {
		// Build detailed error message showing all validation errors
		errorDetails := make([]string, len(e.Errors))
		for i, validationErr := range e.Errors {
			if validationErr.Path != "" {
				errorDetails[i] = fmt.Sprintf("'%s': %s", validationErr.Path, validationErr.Message)
			} else {
				errorDetails[i] = validationErr.Message
			}

			// Add suggestion if available
			if validationErr.Suggestion != "" {
				errorDetails[i] += fmt.Sprintf(" (suggestion: %s)", validationErr.Suggestion)
			}
		}

		if len(e.Errors) == 1 {
			return fmt.Sprintf("javascript function '%s' validation failed at %s: %s", e.Function, e.Stage, errorDetails[0])
		} else {
			var details string
			for i, detail := range errorDetails {
				details += fmt.Sprintf("\n  %d. %s", i+1, detail)
			}
			return fmt.Sprintf("javascript function '%s' validation failed at %s with %d errors:%s", e.Function, e.Stage, len(e.Errors), details)
		}
	}

	if e.Cause != nil {
		return fmt.Sprintf("javascript function '%s' failed at %s: %v", e.Function, e.Stage, e.Cause)
	}

	return fmt.Sprintf("javascript function '%s' failed at %s", e.Function, e.Stage)
}

func (e *JSFunctionError) Unwrap() error {
	return e.Cause
}

// JSPortalError represents errors in portal operations
type JSPortalError struct {
	Address string
	Message string
	Cause   error
}

func (e *JSPortalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("javascript portal error for address '%s': %s (caused by: %v)",
			e.Address, e.Message, e.Cause)
	}
	return fmt.Sprintf("javascript portal error for address '%s': %s", e.Address, e.Message)
}

func (e *JSPortalError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a JSFunctionError for validation failures
func NewValidationError(functionName, stage string, errors []schema.ValidationError) *JSFunctionError {
	return &JSFunctionError{
		Function: functionName,
		Stage:    stage,
		Errors:   errors,
	}
}

// NewExecutionError creates a JSFunctionError for execution failures
func NewExecutionError(functionName, stage string, cause error) *JSFunctionError {
	return &JSFunctionError{
		Function: functionName,
		Stage:    stage,
		Cause:    cause,
	}
}

// NewTimeoutError creates a JSFunctionError for timeout failures
func NewTimeoutError(functionName string, timeout string) *JSFunctionError {
	return &JSFunctionError{
		Function: functionName,
		Stage:    "timeout",
		Message:  fmt.Sprintf("execution timeout after %s", timeout),
	}
}

// NewSyntaxError creates a JSFunctionError for JavaScript syntax errors
func NewSyntaxError(functionName string, cause error) *JSFunctionError {
	return &JSFunctionError{
		Function: functionName,
		Stage:    "syntax_error",
		Message:  "JavaScript syntax error",
		Cause:    cause,
	}
}

// logFunctionOutputValidation logs validation errors for internal output validation
func logFunctionOutputValidation(functionName string, errors []schema.ValidationError) {
	errorDetails := make([]string, len(errors))
	for i, err := range errors {
		if err.Path != "" {
			errorDetails[i] = fmt.Sprintf("%s: %s", err.Path, err.Message)
		} else {
			errorDetails[i] = err.Message
		}
	}

	// In a real implementation, you'd use a proper logger
	fmt.Printf("[WARN] JavaScript function '%s' output validation failed: %s\n",
		functionName, strings.Join(errorDetails, ", "))
}
