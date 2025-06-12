package core

// ValidationResult represents the result of validating a value against a schema.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// ValidationError represents a single validation error with context and suggestions.
type ValidationError struct {
	Path       string `json:"path"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Value      any    `json:"value,omitempty"`
	Expected   string `json:"expected,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Context    string `json:"context,omitempty"`
}
