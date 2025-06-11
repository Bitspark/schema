package schemas

import (
	"fmt"
	"reflect"

	"defs.dev/schema/api"
)

// ArraySchemaConfig holds the configuration for building an ArraySchema.
type ArraySchemaConfig struct {
	Metadata       api.SchemaMetadata
	ItemSchema     api.Schema
	MinItems       *int
	MaxItems       *int
	UniqueItems    bool
	ContainsSchema api.Schema
	DefaultVal     []any
}

// ArraySchema is a clean, API-first implementation of array schema validation.
// It implements api.ArraySchema interface and provides immutable operations.
type ArraySchema struct {
	config ArraySchemaConfig
}

// Ensure ArraySchema implements the API interfaces at compile time
var _ api.Schema = (*ArraySchema)(nil)
var _ api.ArraySchema = (*ArraySchema)(nil)
var _ api.Accepter = (*ArraySchema)(nil)

// NewArraySchema creates a new ArraySchema with the given configuration.
func NewArraySchema(config ArraySchemaConfig) *ArraySchema {
	return &ArraySchema{config: config}
}

// Type returns the schema type constant.
func (a *ArraySchema) Type() api.SchemaType {
	return api.TypeArray
}

// Metadata returns the schema metadata.
func (a *ArraySchema) Metadata() api.SchemaMetadata {
	return a.config.Metadata
}

// Clone returns a deep copy of the ArraySchema.
func (a *ArraySchema) Clone() api.Schema {
	newConfig := a.config

	// Deep copy metadata examples and tags
	if a.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(a.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, a.config.Metadata.Examples)
	}

	if a.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(a.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, a.config.Metadata.Tags)
	}

	// Deep copy default value
	if a.config.DefaultVal != nil {
		newConfig.DefaultVal = make([]any, len(a.config.DefaultVal))
		copy(newConfig.DefaultVal, a.config.DefaultVal)
	}

	// Note: ItemSchema and ContainsSchema are not deeply cloned as they should be immutable
	// If deep cloning is needed, it should be done at the configuration level

	return NewArraySchema(newConfig)
}

// ItemSchema returns the schema for array items.
func (a *ArraySchema) ItemSchema() api.Schema {
	return a.config.ItemSchema
}

// MinItems returns the minimum items constraint.
func (a *ArraySchema) MinItems() *int {
	return a.config.MinItems
}

// MaxItems returns the maximum items constraint.
func (a *ArraySchema) MaxItems() *int {
	return a.config.MaxItems
}

// UniqueItemsRequired returns whether unique items are required.
func (a *ArraySchema) UniqueItemsRequired() bool {
	return a.config.UniqueItems
}

// ContainsSchema returns the contains constraint schema.
func (a *ArraySchema) ContainsSchema() api.Schema {
	return a.config.ContainsSchema
}

// DefaultValue returns the default value.
func (a *ArraySchema) DefaultValue() []any {
	if a.config.DefaultVal == nil {
		return nil
	}
	result := make([]any, len(a.config.DefaultVal))
	copy(result, a.config.DefaultVal)
	return result
}

// Validate validates a value against the array schema.
func (a *ArraySchema) Validate(value any) api.ValidationResult {
	// Convert to slice/array
	arrayValue, ok := a.convertToSlice(value)
	if !ok {
		return api.ValidationResult{
			Valid: false,
			Errors: []api.ValidationError{{
				Path:       "",
				Message:    "Expected array or slice",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "array or slice",
				Suggestion: "Provide an array or slice value",
			}},
		}
	}

	var errors []api.ValidationError

	// Length validation
	length := len(arrayValue)

	if a.config.MinItems != nil && length < *a.config.MinItems {
		errors = append(errors, api.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Array too short (minimum %d items)", *a.config.MinItems),
			Code:       "min_items",
			Value:      length,
			Expected:   fmt.Sprintf("≥ %d items", *a.config.MinItems),
			Suggestion: fmt.Sprintf("Provide at least %d items", *a.config.MinItems),
		})
	}

	if a.config.MaxItems != nil && length > *a.config.MaxItems {
		errors = append(errors, api.ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Array too long (maximum %d items)", *a.config.MaxItems),
			Code:       "max_items",
			Value:      length,
			Expected:   fmt.Sprintf("≤ %d items", *a.config.MaxItems),
			Suggestion: fmt.Sprintf("Limit to %d items", *a.config.MaxItems),
		})
	}

	// Unique items validation
	if a.config.UniqueItems {
		if !a.areItemsUnique(arrayValue) {
			errors = append(errors, api.ValidationError{
				Path:       "",
				Message:    "Array items must be unique",
				Code:       "unique_items",
				Value:      arrayValue,
				Expected:   "unique items",
				Suggestion: "Remove duplicate items from the array",
			})
		}
	}

	// Item schema validation
	if a.config.ItemSchema != nil {
		for i, item := range arrayValue {
			itemResult := a.config.ItemSchema.Validate(item)
			if !itemResult.Valid {
				for _, itemError := range itemResult.Errors {
					errors = append(errors, api.ValidationError{
						Path:       fmt.Sprintf("[%d]%s", i, itemError.Path),
						Message:    itemError.Message,
						Code:       itemError.Code,
						Value:      itemError.Value,
						Expected:   itemError.Expected,
						Suggestion: itemError.Suggestion,
						Context:    fmt.Sprintf("Array item %d", i),
					})
				}
			}
		}
	}

	// Contains schema validation
	if a.config.ContainsSchema != nil {
		containsValid := false
		for _, item := range arrayValue {
			if a.config.ContainsSchema.Validate(item).Valid {
				containsValid = true
				break
			}
		}
		if !containsValid {
			errors = append(errors, api.ValidationError{
				Path:       "",
				Message:    "Array must contain at least one item matching the contains schema",
				Code:       "contains",
				Value:      arrayValue,
				Expected:   "at least one matching item",
				Suggestion: "Add an item that matches the required schema",
			})
		}
	}

	return api.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// convertToSlice converts various array-like types to []any.
func (a *ArraySchema) convertToSlice(value any) ([]any, bool) {
	if value == nil {
		return nil, false
	}

	// Direct slice of any
	if slice, ok := value.([]any); ok {
		return slice, true
	}

	// Use reflection for other slice types
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	length := rv.Len()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		result[i] = rv.Index(i).Interface()
	}

	return result, true
}

// areItemsUnique checks if all items in the array are unique.
func (a *ArraySchema) areItemsUnique(items []any) bool {
	seen := make(map[any]bool)
	for _, item := range items {
		// Use deep equal for complex types
		key := a.getUniqueKey(item)
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

// getUniqueKey generates a comparable key for uniqueness checking.
func (a *ArraySchema) getUniqueKey(item any) any {
	// For simple types, return as-is
	switch v := item.(type) {
	case nil, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, string:
		return v
	default:
		// For complex types, use string representation
		// This is a simple approach; a more sophisticated implementation
		// might use content-based hashing
		return fmt.Sprintf("%+v", v)
	}
}

// ToJSONSchema generates a JSON Schema representation of the array schema.
func (a *ArraySchema) ToJSONSchema() map[string]any {
	jsonSchema := map[string]any{
		"type": "array",
	}

	// Add constraints
	if a.config.MinItems != nil {
		jsonSchema["minItems"] = *a.config.MinItems
	}

	if a.config.MaxItems != nil {
		jsonSchema["maxItems"] = *a.config.MaxItems
	}

	if a.config.UniqueItems {
		jsonSchema["uniqueItems"] = true
	}

	// Add item schema
	if a.config.ItemSchema != nil {
		jsonSchema["items"] = a.config.ItemSchema.ToJSONSchema()
	}

	// Add contains schema
	if a.config.ContainsSchema != nil {
		jsonSchema["contains"] = a.config.ContainsSchema.ToJSONSchema()
	}

	// Add metadata
	if a.config.Metadata.Description != "" {
		jsonSchema["description"] = a.config.Metadata.Description
	}

	if len(a.config.Metadata.Examples) > 0 {
		if len(a.config.Metadata.Examples) == 1 {
			jsonSchema["example"] = a.config.Metadata.Examples[0]
		} else {
			jsonSchema["examples"] = a.config.Metadata.Examples
		}
	}

	if len(a.config.Metadata.Tags) > 0 {
		jsonSchema["tags"] = a.config.Metadata.Tags
	}

	if a.config.DefaultVal != nil {
		jsonSchema["default"] = a.config.DefaultVal
	}

	return jsonSchema
}

// GenerateExample generates an example value for the array schema.
func (a *ArraySchema) GenerateExample() any {
	// Use provided examples if available
	if len(a.config.Metadata.Examples) > 0 {
		return a.config.Metadata.Examples[0]
	}

	// Use default value if set
	if a.config.DefaultVal != nil {
		return a.config.DefaultVal
	}

	// Generate based on constraints
	minLength := 1
	if a.config.MinItems != nil && *a.config.MinItems > 0 {
		minLength = *a.config.MinItems
	}

	maxLength := 3
	if a.config.MaxItems != nil {
		maxLength = *a.config.MaxItems
		if maxLength > 5 {
			maxLength = 5 // Cap for reasonable examples
		}
	}

	// Choose length between min and max
	length := minLength
	if maxLength > minLength {
		length = (minLength + maxLength) / 2
	}

	var result []any

	// Generate items using item schema if available
	if a.config.ItemSchema != nil {
		for i := 0; i < length; i++ {
			example := a.config.ItemSchema.GenerateExample()

			// Ensure uniqueness if required
			if a.config.UniqueItems {
				// Simple uniqueness for examples - just modify slightly
				if i > 0 {
					switch v := example.(type) {
					case string:
						example = fmt.Sprintf("%s_%d", v, i)
					case int64:
						example = v + int64(i)
					case float64:
						example = v + float64(i)
					}
				}
			}

			result = append(result, example)
		}
	} else {
		// Generic examples without item schema
		for i := 0; i < length; i++ {
			result = append(result, fmt.Sprintf("item_%d", i+1))
		}
	}

	return result
}

// Accept implements the visitor pattern for schema traversal.
func (a *ArraySchema) Accept(visitor api.SchemaVisitor) error {
	return visitor.VisitArray(a)
}
