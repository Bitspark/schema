package schemas

import (
	"fmt"
	"math"

	"defs.dev/schema/api/core"
)

// IntegerSchemaConfig holds the configuration for building an IntegerSchema.
type IntegerSchemaConfig struct {
	Metadata   core.SchemaMetadata
	Minimum    *int64
	Maximum    *int64
	DefaultVal *int64
}

// IntegerSchema is a clean, API-first implementation of integer schema validation.
// It implements core.IntegerSchema interface and provides immutable operations.
type IntegerSchema struct {
	config IntegerSchemaConfig
}

// Ensure IntegerSchema implements the API interfaces at compile time
var _ core.Schema = (*IntegerSchema)(nil)
var _ core.IntegerSchema = (*IntegerSchema)(nil)
var _ core.Accepter = (*IntegerSchema)(nil)

// NewIntegerSchema creates a new IntegerSchema with the given configuration.
func NewIntegerSchema(config IntegerSchemaConfig) *IntegerSchema {
	return &IntegerSchema{config: config}
}

// Type returns the schema type constant.
func (i *IntegerSchema) Type() core.SchemaType {
	return core.TypeInteger
}

// Metadata returns the schema metadata.
func (i *IntegerSchema) Metadata() core.SchemaMetadata {
	return i.config.Metadata
}

// Clone returns a deep copy of the IntegerSchema.
func (i *IntegerSchema) Clone() core.Schema {
	newConfig := i.config

	// Deep copy metadata examples and tags
	if i.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(i.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, i.config.Metadata.Examples)
	}

	if i.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(i.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, i.config.Metadata.Tags)
	}

	return NewIntegerSchema(newConfig)
}

// Minimum returns the minimum value constraint.
func (i *IntegerSchema) Minimum() *int64 {
	return i.config.Minimum
}

// Maximum returns the maximum value constraint.
func (i *IntegerSchema) Maximum() *int64 {
	return i.config.Maximum
}

// DefaultValue returns the default value.
func (i *IntegerSchema) DefaultValue() *int64 {
	return i.config.DefaultVal
}

// Validate validates a value against the integer schema.
func (i *IntegerSchema) Validate(value any) core.ValidationResult {
	// Handle different integer types and convert to int64
	var intVal int64
	var ok bool

	switch v := value.(type) {
	case int64:
		intVal = v
		ok = true
	case int:
		intVal = int64(v)
		ok = true
	case int32:
		intVal = int64(v)
		ok = true
	case int16:
		intVal = int64(v)
		ok = true
	case int8:
		intVal = int64(v)
		ok = true
	case uint64:
		// Check for overflow when converting uint64 to int64
		if v > math.MaxInt64 {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Integer value too large for int64",
					Code:       "overflow",
					Value:      value,
					Expected:   "integer within int64 range",
					Suggestion: fmt.Sprintf("Provide an integer ≤ %d", math.MaxInt64),
				}},
			}
		}
		intVal = int64(v)
		ok = true
	case uint:
		// Check for overflow when converting uint to int64
		if uint64(v) > math.MaxInt64 {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Integer value too large for int64",
					Code:       "overflow",
					Value:      value,
					Expected:   "integer within int64 range",
					Suggestion: fmt.Sprintf("Provide an integer ≤ %d", math.MaxInt64),
				}},
			}
		}
		intVal = int64(v)
		ok = true
	case uint32:
		intVal = int64(v)
		ok = true
	case uint16:
		intVal = int64(v)
		ok = true
	case uint8:
		intVal = int64(v)
		ok = true
	case float64:
		// Allow integers represented as floats, but only if they're whole numbers
		if v != math.Trunc(v) {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Expected integer, got decimal number",
					Code:       "not_integer",
					Value:      value,
					Expected:   "whole number",
					Suggestion: "Provide an integer value without decimal places",
				}},
			}
		}
		// Check for overflow
		if v > math.MaxInt64 || v < math.MinInt64 {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Integer value out of range for int64",
					Code:       "overflow",
					Value:      value,
					Expected:   "integer within int64 range",
					Suggestion: fmt.Sprintf("Provide an integer between %d and %d", math.MinInt64, math.MaxInt64),
				}},
			}
		}
		intVal = int64(v)
		ok = true
	case float32:
		// Same logic as float64
		f64Val := float64(v)
		if f64Val != math.Trunc(f64Val) {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Expected integer, got decimal number",
					Code:       "not_integer",
					Value:      value,
					Expected:   "whole number",
					Suggestion: "Provide an integer value without decimal places",
				}},
			}
		}
		if f64Val > math.MaxInt64 || f64Val < math.MinInt64 {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Integer value out of range for int64",
					Code:       "overflow",
					Value:      value,
					Expected:   "integer within int64 range",
					Suggestion: fmt.Sprintf("Provide an integer between %d and %d", math.MinInt64, math.MaxInt64),
				}},
			}
		}
		intVal = int64(f64Val)
		ok = true
	default:
		ok = false
	}

	if !ok {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{{
				Path:       "",
				Message:    "Expected integer",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "integer",
				Suggestion: "Provide an integer value",
			}},
		}
	}

	var errors []core.ValidationError

	// Minimum validation
	if i.config.Minimum != nil && intVal < *i.config.Minimum {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Integer too small (minimum %d)", *i.config.Minimum),
			Code:       "minimum",
			Value:      intVal,
			Expected:   fmt.Sprintf("≥ %d", *i.config.Minimum),
			Suggestion: fmt.Sprintf("Provide an integer greater than or equal to %d", *i.config.Minimum),
		})
	}

	// Maximum validation
	if i.config.Maximum != nil && intVal > *i.config.Maximum {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Integer too large (maximum %d)", *i.config.Maximum),
			Code:       "maximum",
			Value:      intVal,
			Expected:   fmt.Sprintf("≤ %d", *i.config.Maximum),
			Suggestion: fmt.Sprintf("Provide an integer less than or equal to %d", *i.config.Maximum),
		})
	}

	return core.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// GenerateExample generates an example value for the integer schema.
func (i *IntegerSchema) GenerateExample() any {
	// Use provided examples if available
	if len(i.config.Metadata.Examples) > 0 {
		return i.config.Metadata.Examples[0]
	}

	// Use default value if set
	if i.config.DefaultVal != nil {
		return *i.config.DefaultVal
	}

	// Generate based on constraints
	if i.config.Minimum != nil && i.config.Maximum != nil {
		// Return midpoint between min and max
		return (*i.config.Minimum + *i.config.Maximum) / 2
	}

	if i.config.Minimum != nil {
		// Return minimum + 1 (or minimum if it's positive)
		if *i.config.Minimum >= 0 {
			return *i.config.Minimum + 1
		}
		return *i.config.Minimum
	}

	if i.config.Maximum != nil {
		// Return maximum - 1 (or maximum if it's negative)
		if *i.config.Maximum <= 0 {
			return *i.config.Maximum - 1
		}
		return *i.config.Maximum
	}

	// Default example
	return int64(42)
}

// Accept implements the visitor pattern for schema traversal.
func (i *IntegerSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitInteger(i)
}
