package validation

// ValidationResult represents the result of a validation operation.
// This is the canonical result type that both schema-level and value-level
// validators should return via consumer.NewResult("validation", ValidationResult{...}).
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors,omitempty"`
	Warnings []ValidationIssue `json:"warnings,omitempty"`
}

// ValidationIssue represents a single validation error or warning.
type ValidationIssue struct {
	Path    []string `json:"path"` // empty = root
	Code    string   `json:"code"`
	Message string   `json:"message"`
}

// NewValidationResult creates a valid ValidationResult.
func NewValidationResult() ValidationResult {
	return ValidationResult{Valid: true}
}

// NewValidationError creates a ValidationResult with an error.
func NewValidationError(path []string, code, message string) ValidationResult {
	return ValidationResult{
		Valid: false,
		Errors: []ValidationIssue{{
			Path:    append([]string(nil), path...), // copy slice
			Code:    code,
			Message: message,
		}},
	}
}

// NewValidationWarning creates a ValidationResult with a warning.
func NewValidationWarning(path []string, code, message string) ValidationResult {
	return ValidationResult{
		Valid: true,
		Warnings: []ValidationIssue{{
			Path:    append([]string(nil), path...), // copy slice
			Code:    code,
			Message: message,
		}},
	}
}

// AddError adds an error to the ValidationResult.
func (r *ValidationResult) AddError(path []string, code, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationIssue{
		Path:    append([]string(nil), path...), // copy slice
		Code:    code,
		Message: message,
	})
}

// AddWarning adds a warning to the ValidationResult.
func (r *ValidationResult) AddWarning(path []string, code, message string) {
	r.Warnings = append(r.Warnings, ValidationIssue{
		Path:    append([]string(nil), path...), // copy slice
		Code:    code,
		Message: message,
	})
}

// Merge combines multiple ValidationResults.
func (r *ValidationResult) Merge(other ValidationResult) {
	if !other.Valid {
		r.Valid = false
	}
	r.Errors = append(r.Errors, other.Errors...)
	r.Warnings = append(r.Warnings, other.Warnings...)
}
