package json

import (
	"fmt"
	"strings"
)

// JSONSchemaOptions configures the behavior of the JSON Schema generator.
type JSONSchemaOptions struct {
	// Draft specifies the JSON Schema draft version to generate
	// Supported values: "draft-07", "draft-2019-09", "draft-2020-12"
	Draft string

	// IncludeExamples determines whether to include example values in the schema
	IncludeExamples bool

	// IncludeDefaults determines whether to include default values in the schema
	IncludeDefaults bool

	// IncludeDescription determines whether to include description metadata
	IncludeDescription bool

	// StrictMode enables stricter validation rules
	StrictMode bool

	// PrettyPrint enables JSON formatting with indentation
	PrettyPrint bool

	// IndentSize specifies the number of spaces for indentation (when PrettyPrint is true)
	IndentSize int

	// IncludeFormat determines whether to include format constraints for strings
	IncludeFormat bool

	// IncludeTitle determines whether to include title metadata
	IncludeTitle bool

	// IncludeAdditionalProperties controls additionalProperties behavior for objects
	// When true, includes "additionalProperties": false for strict objects
	IncludeAdditionalProperties bool

	// SchemaURI sets the $schema URI in the generated schema
	SchemaURI string

	// RootID sets the $id for the root schema
	RootID string

	// DefinitionsKey specifies the key to use for schema definitions
	// Common values: "definitions" (draft-07), "$defs" (draft-2019-09+)
	DefinitionsKey string

	// MinifyOutput removes unnecessary whitespace and formatting
	MinifyOutput bool

	// IncludeReadOnly determines whether to include readOnly properties
	IncludeReadOnly bool

	// IncludeWriteOnly determines whether to include writeOnly properties
	IncludeWriteOnly bool

	// AllowNullableTypes enables nullable type generation (type: ["string", "null"])
	AllowNullableTypes bool
}

// DefaultJSONSchemaOptions returns the default options for JSON Schema generation.
func DefaultJSONSchemaOptions() JSONSchemaOptions {
	return JSONSchemaOptions{
		Draft:                       "draft-07",
		IncludeExamples:             true,
		IncludeDefaults:             true,
		IncludeDescription:          true,
		StrictMode:                  false,
		PrettyPrint:                 true,
		IndentSize:                  2,
		IncludeFormat:               true,
		IncludeTitle:                true,
		IncludeAdditionalProperties: true,
		SchemaURI:                   "https://json-schema.org/draft-07/schema#",
		RootID:                      "",
		DefinitionsKey:              "definitions",
		MinifyOutput:                false,
		IncludeReadOnly:             true,
		IncludeWriteOnly:            true,
		AllowNullableTypes:          false,
	}
}

// SetOption implements the option setter interface for functional options.
func (o *JSONSchemaOptions) SetOption(key string, value any) {
	switch key {
	case "draft":
		if v, ok := value.(string); ok {
			o.Draft = v
			o.updateSchemaURI()
			o.updateDefinitionsKey()
		}
	case "include_examples":
		if v, ok := value.(bool); ok {
			o.IncludeExamples = v
		}
	case "include_defaults":
		if v, ok := value.(bool); ok {
			o.IncludeDefaults = v
		}
	case "include_description":
		if v, ok := value.(bool); ok {
			o.IncludeDescription = v
		}
	case "strict_mode":
		if v, ok := value.(bool); ok {
			o.StrictMode = v
		}
	case "pretty_print":
		if v, ok := value.(bool); ok {
			o.PrettyPrint = v
		}
	case "indent_size":
		if v, ok := value.(int); ok {
			o.IndentSize = v
		}
	case "include_format":
		if v, ok := value.(bool); ok {
			o.IncludeFormat = v
		}
	case "include_title":
		if v, ok := value.(bool); ok {
			o.IncludeTitle = v
		}
	case "include_additional_properties":
		if v, ok := value.(bool); ok {
			o.IncludeAdditionalProperties = v
		}
	case "schema_uri":
		if v, ok := value.(string); ok {
			o.SchemaURI = v
		}
	case "root_id":
		if v, ok := value.(string); ok {
			o.RootID = v
		}
	case "definitions_key":
		if v, ok := value.(string); ok {
			o.DefinitionsKey = v
		}
	case "minify_output":
		if v, ok := value.(bool); ok {
			o.MinifyOutput = v
		}
	case "include_readonly":
		if v, ok := value.(bool); ok {
			o.IncludeReadOnly = v
		}
	case "include_writeonly":
		if v, ok := value.(bool); ok {
			o.IncludeWriteOnly = v
		}
	case "allow_nullable_types":
		if v, ok := value.(bool); ok {
			o.AllowNullableTypes = v
		}
	}
}

// updateSchemaURI updates the schema URI based on the draft version.
func (o *JSONSchemaOptions) updateSchemaURI() {
	if o.SchemaURI == "" || isDefaultSchemaURI(o.SchemaURI) {
		switch o.Draft {
		case "draft-07":
			o.SchemaURI = "https://json-schema.org/draft-07/schema#"
		case "draft-2019-09":
			o.SchemaURI = "https://json-schema.org/draft/2019-09/schema#"
		case "draft-2020-12":
			o.SchemaURI = "https://json-schema.org/draft/2020-12/schema#"
		default:
			o.SchemaURI = "https://json-schema.org/draft-07/schema#"
		}
	}
}

// updateDefinitionsKey updates the definitions key based on the draft version.
func (o *JSONSchemaOptions) updateDefinitionsKey() {
	if o.DefinitionsKey == "" || isDefaultDefinitionsKey(o.DefinitionsKey) {
		switch o.Draft {
		case "draft-07":
			o.DefinitionsKey = "definitions"
		case "draft-2019-09", "draft-2020-12":
			o.DefinitionsKey = "$defs"
		default:
			o.DefinitionsKey = "definitions"
		}
	}
}

// isDefaultSchemaURI checks if the URI is one of the default schema URIs.
func isDefaultSchemaURI(uri string) bool {
	defaults := []string{
		"https://json-schema.org/draft-07/schema#",
		"https://json-schema.org/draft/2019-09/schema#",
		"https://json-schema.org/draft/2020-12/schema#",
	}
	for _, defaultURI := range defaults {
		if uri == defaultURI {
			return true
		}
	}
	return false
}

// isDefaultDefinitionsKey checks if the key is one of the default definitions keys.
func isDefaultDefinitionsKey(key string) bool {
	return key == "definitions" || key == "$defs"
}

// Clone creates a deep copy of the options.
func (o JSONSchemaOptions) Clone() JSONSchemaOptions {
	return o // struct copy is sufficient since all fields are value types
}

// Validate checks if the options are valid and returns an error if not.
func (o *JSONSchemaOptions) Validate() error {
	// Validate draft version
	validDrafts := []string{"draft-07", "draft-2019-09", "draft-2020-12"}
	validDraft := false
	for _, draft := range validDrafts {
		if o.Draft == draft {
			validDraft = true
			break
		}
	}
	if !validDraft {
		return &OptionsError{
			Field:   "Draft",
			Value:   o.Draft,
			Message: "unsupported JSON Schema draft version",
			Valid:   validDrafts,
		}
	}

	// Validate indent size
	if o.IndentSize < 0 || o.IndentSize > 10 {
		return &OptionsError{
			Field:   "IndentSize",
			Value:   o.IndentSize,
			Message: "indent size must be between 0 and 10",
		}
	}

	// Validate definitions key for draft version
	if o.Draft == "draft-07" && o.DefinitionsKey != "definitions" {
		return &OptionsError{
			Field:   "DefinitionsKey",
			Value:   o.DefinitionsKey,
			Message: "draft-07 requires 'definitions' key",
		}
	}

	if (o.Draft == "draft-2019-09" || o.Draft == "draft-2020-12") && o.DefinitionsKey != "$defs" {
		return &OptionsError{
			Field:   "DefinitionsKey",
			Value:   o.DefinitionsKey,
			Message: "draft-2019-09 and draft-2020-12 require '$defs' key",
		}
	}

	return nil
}

// OptionsError represents an error in JSON Schema options configuration.
type OptionsError struct {
	Field   string
	Value   any
	Message string
	Valid   []string
}

// Error implements the error interface.
func (e *OptionsError) Error() string {
	msg := "invalid JSON Schema option"
	if e.Field != "" {
		msg += " for field " + e.Field
	}
	if e.Value != nil {
		msg += ": " + fmt.Sprintf("%v", e.Value)
	}
	if e.Message != "" {
		msg += " - " + e.Message
	}
	if len(e.Valid) > 0 {
		msg += " (valid values: " + strings.Join(e.Valid, ", ") + ")"
	}
	return msg
}
