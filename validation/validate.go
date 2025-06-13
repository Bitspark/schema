package validation

import (
	"defs.dev/schema/api/core"
)

// ValidateValue validates a value against a schema using consumer-driven validation.
// This is the main entry point that replaces deprecated schema.Validate() calls.
// It provides backward compatibility for existing code.
func ValidateValue(schema core.Schema, value any) ValidationResult {
	// Use the integration layer to validate with registry
	return ValidateWithRegistry(schema, value)
}

// simpleValue is a basic implementation of core.Value for validation
type simpleValue struct {
	value any
}

func (v *simpleValue) Value() any {
	return v.value
}

func (v *simpleValue) String() string {
	return ""
}

func (v *simpleValue) IsNull() bool {
	return v.value == nil
}

func (v *simpleValue) IsComposite() bool {
	return false
}

func (v *simpleValue) Copy() any {
	return v.value
}
