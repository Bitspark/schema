package schemas

import (
	"strconv"
	"strings"

	"defs.dev/schema/api/core"
)

// BooleanSchemaConfig holds the configuration for building a BooleanSchema.
type BooleanSchemaConfig struct {
	Metadata        core.SchemaMetadata
	DefaultVal      *bool
	AllowStringConv bool // Allow conversion from string ("true", "false", "1", "0")
	CaseInsensitive bool // Case-insensitive string conversion
}

// BooleanSchema is a clean, API-first implementation of boolean schema validation.
// It implements core.BooleanSchema interface and provides immutable operations.
type BooleanSchema struct {
	config BooleanSchemaConfig
}

// Ensure BooleanSchema implements the API interfaces at compile time
var _ core.Schema = (*BooleanSchema)(nil)
var _ core.BooleanSchema = (*BooleanSchema)(nil)
var _ core.Accepter = (*BooleanSchema)(nil)

// NewBooleanSchema creates a new BooleanSchema with the given configuration.
func NewBooleanSchema(config BooleanSchemaConfig) *BooleanSchema {
	return &BooleanSchema{config: config}
}

// Type returns the schema type constant.
func (b *BooleanSchema) Type() core.SchemaType {
	return core.TypeBoolean
}

// Metadata returns the schema metadata.
func (b *BooleanSchema) Metadata() core.SchemaMetadata {
	return b.config.Metadata
}

// Clone returns a deep copy of the BooleanSchema.
func (b *BooleanSchema) Clone() core.Schema {
	newConfig := b.config

	// Deep copy metadata examples and tags
	if b.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(b.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, b.config.Metadata.Examples)
	}

	if b.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(b.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, b.config.Metadata.Tags)
	}

	return NewBooleanSchema(newConfig)
}

// DefaultValue returns the default value.
func (b *BooleanSchema) DefaultValue() *bool {
	return b.config.DefaultVal
}

// AllowStringConversion returns whether string-to-bool conversion is allowed.
func (b *BooleanSchema) AllowStringConversion() bool {
	return b.config.AllowStringConv
}

// CaseInsensitive returns whether string conversion is case-insensitive.
func (b *BooleanSchema) CaseInsensitive() bool {
	return b.config.CaseInsensitive
}

// Validate validates a value against the boolean schema.
func (b *BooleanSchema) Validate(value any) core.ValidationResult {
	// Direct boolean validation
	if _, ok := value.(bool); ok {
		return core.ValidationResult{
			Valid:  true,
			Errors: nil,
		}
	}

	// String conversion if allowed
	if b.config.AllowStringConv {
		if strVal, ok := value.(string); ok {
			convertedVal, err := b.convertStringToBool(strVal)
			if err == nil {
				// Store the converted value for potential use
				return core.ValidationResult{
					Valid:  true,
					Errors: nil,
					Metadata: map[string]any{
						"converted_value": convertedVal,
						"original_value":  strVal,
					},
				}
			}
			// If conversion failed, return specific error
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "Invalid boolean string format",
					Code:       "invalid_boolean_string",
					Value:      value,
					Expected:   "true, false, 1, or 0",
					Suggestion: "Use 'true', 'false', '1', or '0'",
				}},
			}
		}
	}

	// Type mismatch error
	expectedTypes := "boolean"
	if b.config.AllowStringConv {
		expectedTypes = "boolean or string (true/false/1/0)"
	}

	return core.ValidationResult{
		Valid: false,
		Errors: []core.ValidationError{{
			Path:       "",
			Message:    "Expected boolean",
			Code:       "type_mismatch",
			Value:      value,
			Expected:   expectedTypes,
			Suggestion: "Provide a boolean value (true or false)",
		}},
	}
}

// convertStringToBool converts a string to boolean using various formats.
func (b *BooleanSchema) convertStringToBool(str string) (bool, error) {
	// Handle case insensitivity
	convertStr := str
	if b.config.CaseInsensitive {
		convertStr = strings.ToLower(str)
	}

	// Standard boolean strings
	switch convertStr {
	case "true", "1", "yes", "on", "y", "t":
		return true, nil
	case "false", "0", "no", "off", "n", "f":
		return false, nil
	}

	// Try Go's standard strconv.ParseBool as fallback
	return strconv.ParseBool(convertStr)
}

// ToJSONSchema generates a JSON Schema representation of the boolean schema.
func (b *BooleanSchema) ToJSONSchema() map[string]any {
	jsonSchema := map[string]any{
		"type": "boolean",
	}

	// Add metadata
	if b.config.Metadata.Description != "" {
		jsonSchema["description"] = b.config.Metadata.Description
	}

	if len(b.config.Metadata.Examples) > 0 {
		if len(b.config.Metadata.Examples) == 1 {
			jsonSchema["example"] = b.config.Metadata.Examples[0]
		} else {
			jsonSchema["examples"] = b.config.Metadata.Examples
		}
	}

	if len(b.config.Metadata.Tags) > 0 {
		jsonSchema["tags"] = b.config.Metadata.Tags
	}

	if b.config.DefaultVal != nil {
		jsonSchema["default"] = *b.config.DefaultVal
	}

	// Add custom properties for string conversion support
	if b.config.AllowStringConv {
		jsonSchema["x-allow-string-conversion"] = true
		if b.config.CaseInsensitive {
			jsonSchema["x-case-insensitive"] = true
		}
	}

	return jsonSchema
}

// GenerateExample generates an example value for the boolean schema.
func (b *BooleanSchema) GenerateExample() any {
	// Use provided examples if available
	if len(b.config.Metadata.Examples) > 0 {
		return b.config.Metadata.Examples[0]
	}

	// Use default value if set
	if b.config.DefaultVal != nil {
		return *b.config.DefaultVal
	}

	// Default example
	return true
}

// Accept implements the visitor pattern for schema traversal.
func (b *BooleanSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitBoolean(b)
}
