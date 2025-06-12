package schemas

import (
	"fmt"
	"math"

	"defs.dev/schema/api/core"
)

// NumberSchemaConfig holds the configuration for building a NumberSchema.
type NumberSchemaConfig struct {
	Metadata   core.SchemaMetadata
	Minimum    *float64
	Maximum    *float64
	DefaultVal *float64
}

// NumberSchema is a clean, API-first implementation of number schema validation.
// It implements core.NumberSchema interface and provides immutable operations.
type NumberSchema struct {
	config NumberSchemaConfig
}

// Ensure NumberSchema implements the API interfaces at compile time
var _ core.Schema = (*NumberSchema)(nil)
var _ core.NumberSchema = (*NumberSchema)(nil)
var _ core.Accepter = (*NumberSchema)(nil)

// NewNumberSchema creates a new NumberSchema with the given configuration.
func NewNumberSchema(config NumberSchemaConfig) *NumberSchema {
	return &NumberSchema{config: config}
}

// Type returns the schema type constant.
func (n *NumberSchema) Type() core.SchemaType {
	return core.TypeNumber
}

// Metadata returns the schema metadata.
func (n *NumberSchema) Metadata() core.SchemaMetadata {
	return n.config.Metadata
}

// Clone returns a deep copy of the NumberSchema.
func (n *NumberSchema) Clone() core.Schema {
	newConfig := n.config

	// Deep copy metadata examples and tags
	if n.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(n.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, n.config.Metadata.Examples)
	}

	if n.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(n.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, n.config.Metadata.Tags)
	}

	return NewNumberSchema(newConfig)
}

// Minimum returns the minimum value constraint.
func (n *NumberSchema) Minimum() *float64 {
	return n.config.Minimum
}

// Maximum returns the maximum value constraint.
func (n *NumberSchema) Maximum() *float64 {
	return n.config.Maximum
}

// DefaultValue returns the default value.
func (n *NumberSchema) DefaultValue() *float64 {
	return n.config.DefaultVal
}

// Validate validates a value against the number schema.
func (n *NumberSchema) Validate(value any) core.ValidationResult {
	// Handle different numeric types
	var num float64
	var ok bool

	switch v := value.(type) {
	case float64:
		num = v
		ok = true
	case float32:
		num = float64(v)
		ok = true
	case int:
		num = float64(v)
		ok = true
	case int32:
		num = float64(v)
		ok = true
	case int64:
		num = float64(v)
		ok = true
	case uint:
		num = float64(v)
		ok = true
	case uint32:
		num = float64(v)
		ok = true
	case uint64:
		num = float64(v)
		ok = true
	default:
		ok = false
	}

	if !ok {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{{
				Path:       "",
				Message:    "Expected number",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "number",
				Suggestion: "Provide a numeric value",
			}},
		}
	}

	var errors []core.ValidationError

	// Check for special float values
	if math.IsNaN(num) {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    "NaN (Not a Number) is not allowed",
			Code:       "invalid_number",
			Value:      num,
			Expected:   "finite number",
			Suggestion: "Provide a finite numeric value",
		})
	}

	if math.IsInf(num, 0) {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    "Infinite values are not allowed",
			Code:       "invalid_number",
			Value:      num,
			Expected:   "finite number",
			Suggestion: "Provide a finite numeric value",
		})
	}

	// Minimum validation
	if n.config.Minimum != nil && num < *n.config.Minimum {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Number too small (minimum %g)", *n.config.Minimum),
			Code:       "minimum",
			Value:      num,
			Expected:   fmt.Sprintf("≥ %g", *n.config.Minimum),
			Suggestion: fmt.Sprintf("Provide a number greater than or equal to %g", *n.config.Minimum),
		})
	}

	// Maximum validation
	if n.config.Maximum != nil && num > *n.config.Maximum {
		errors = append(errors, core.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Number too large (maximum %g)", *n.config.Maximum),
			Code:       "maximum",
			Value:      num,
			Expected:   fmt.Sprintf("≤ %g", *n.config.Maximum),
			Suggestion: fmt.Sprintf("Provide a number less than or equal to %g", *n.config.Maximum),
		})
	}

	return core.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// GenerateExample generates an example value for the number schema.
func (n *NumberSchema) GenerateExample() any {
	// Use provided examples if available
	if len(n.config.Metadata.Examples) > 0 {
		return n.config.Metadata.Examples[0]
	}

	// Use default value if set
	if n.config.DefaultVal != nil {
		return *n.config.DefaultVal
	}

	// Generate based on constraints
	if n.config.Minimum != nil && n.config.Maximum != nil {
		// Return midpoint between min and max
		return (*n.config.Minimum + *n.config.Maximum) / 2
	}

	if n.config.Minimum != nil {
		// Return minimum + 1 (or minimum if it's positive)
		if *n.config.Minimum >= 0 {
			return *n.config.Minimum + 1
		}
		return *n.config.Minimum
	}

	if n.config.Maximum != nil {
		// Return maximum - 1 (or maximum if it's negative)
		if *n.config.Maximum <= 0 {
			return *n.config.Maximum - 1
		}
		return *n.config.Maximum
	}

	// Default example
	return 42.0
}

// Accept implements the visitor pattern for schema traversal.
func (n *NumberSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitNumber(n)
}
