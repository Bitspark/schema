package typescript

import (
	"fmt"
	"strings"
)

// TypeScriptOptions configures the behavior of the TypeScript generator.
type TypeScriptOptions struct {
	// OutputStyle determines the TypeScript output style
	// Supported values: "interface", "type", "class"
	OutputStyle string

	// NamingConvention specifies the naming convention for types
	// Supported values: "PascalCase", "camelCase", "snake_case", "kebab-case"
	NamingConvention string

	// IncludeComments determines whether to include JSDoc comments
	IncludeComments bool

	// IncludeExamples determines whether to include example values in comments
	IncludeExamples bool

	// IncludeDefaults determines whether to include default values
	IncludeDefaults bool

	// StrictMode enables stricter TypeScript types (e.g., readonly arrays)
	StrictMode bool

	// UseOptionalProperties uses optional properties (?) instead of union with undefined
	UseOptionalProperties bool

	// IndentSize specifies the number of spaces for indentation
	IndentSize int

	// UseTabsForIndentation uses tabs instead of spaces
	UseTabsForIndentation bool

	// IncludeImports determines whether to include import statements
	IncludeImports bool

	// ExportTypes determines whether to export generated types
	ExportTypes bool

	// UseUnknownType uses 'unknown' instead of 'any' for unknown types
	UseUnknownType bool

	// GenerateValidators determines whether to generate runtime validators
	GenerateValidators bool

	// ValidatorLibrary specifies which validation library to use
	// Supported values: "zod", "yup", "joi", "ajv"
	ValidatorLibrary string

	// UseEnums generates TypeScript enums for string enums
	UseEnums bool

	// UseConstAssertions uses 'as const' assertions for literal types
	UseConstAssertions bool

	// IncludeUtilityTypes includes utility type definitions
	IncludeUtilityTypes bool

	// ArrayStyle determines how arrays are represented
	// Supported values: "Array<T>", "T[]"
	ArrayStyle string

	// ObjectStyle determines how objects are represented
	// Supported values: "interface", "type", "Record<string, T>"
	ObjectStyle string

	// UsePartialTypes generates Partial<T> for optional object properties
	UsePartialTypes bool

	// IncludeJSDoc includes detailed JSDoc documentation
	IncludeJSDoc bool

	// JSDocStyle determines JSDoc comment style
	// Supported values: "standard", "tsdoc"
	JSDocStyle string

	// FileExtension specifies the file extension to use
	// Supported values: ".ts", ".d.ts"
	FileExtension string

	// ModuleSystem specifies the module system
	// Supported values: "es6", "commonjs", "umd", "none"
	ModuleSystem string
}

// DefaultTypeScriptOptions returns the default options for TypeScript generation.
func DefaultTypeScriptOptions() TypeScriptOptions {
	return TypeScriptOptions{
		OutputStyle:           "interface",
		NamingConvention:      "PascalCase",
		IncludeComments:       true,
		IncludeExamples:       true,
		IncludeDefaults:       true,
		StrictMode:            false,
		UseOptionalProperties: true,
		IndentSize:            2,
		UseTabsForIndentation: false,
		IncludeImports:        true,
		ExportTypes:           true,
		UseUnknownType:        true,
		GenerateValidators:    false,
		ValidatorLibrary:      "zod",
		UseEnums:              true,
		UseConstAssertions:    false,
		IncludeUtilityTypes:   false,
		ArrayStyle:            "T[]",
		ObjectStyle:           "interface",
		UsePartialTypes:       false,
		IncludeJSDoc:          true,
		JSDocStyle:            "standard",
		FileExtension:         ".ts",
		ModuleSystem:          "es6",
	}
}

// SetOption implements the option setter interface for functional options.
func (o *TypeScriptOptions) SetOption(key string, value any) {
	switch key {
	case "output_style":
		if v, ok := value.(string); ok {
			o.OutputStyle = v
		}
	case "naming_convention":
		if v, ok := value.(string); ok {
			o.NamingConvention = v
		}
	case "include_comments":
		if v, ok := value.(bool); ok {
			o.IncludeComments = v
		}
	case "include_examples":
		if v, ok := value.(bool); ok {
			o.IncludeExamples = v
		}
	case "include_defaults":
		if v, ok := value.(bool); ok {
			o.IncludeDefaults = v
		}
	case "strict_mode":
		if v, ok := value.(bool); ok {
			o.StrictMode = v
		}
	case "use_optional_properties":
		if v, ok := value.(bool); ok {
			o.UseOptionalProperties = v
		}
	case "indent_size":
		if v, ok := value.(int); ok {
			o.IndentSize = v
		}
	case "use_tabs_for_indentation":
		if v, ok := value.(bool); ok {
			o.UseTabsForIndentation = v
		}
	case "include_imports":
		if v, ok := value.(bool); ok {
			o.IncludeImports = v
		}
	case "export_types":
		if v, ok := value.(bool); ok {
			o.ExportTypes = v
		}
	case "use_unknown_type":
		if v, ok := value.(bool); ok {
			o.UseUnknownType = v
		}
	case "generate_validators":
		if v, ok := value.(bool); ok {
			o.GenerateValidators = v
		}
	case "validator_library":
		if v, ok := value.(string); ok {
			o.ValidatorLibrary = v
		}
	case "use_enums":
		if v, ok := value.(bool); ok {
			o.UseEnums = v
		}
	case "use_const_assertions":
		if v, ok := value.(bool); ok {
			o.UseConstAssertions = v
		}
	case "include_utility_types":
		if v, ok := value.(bool); ok {
			o.IncludeUtilityTypes = v
		}
	case "array_style":
		if v, ok := value.(string); ok {
			o.ArrayStyle = v
		}
	case "object_style":
		if v, ok := value.(string); ok {
			o.ObjectStyle = v
		}
	case "use_partial_types":
		if v, ok := value.(bool); ok {
			o.UsePartialTypes = v
		}
	case "include_jsdoc":
		if v, ok := value.(bool); ok {
			o.IncludeJSDoc = v
		}
	case "jsdoc_style":
		if v, ok := value.(string); ok {
			o.JSDocStyle = v
		}
	case "file_extension":
		if v, ok := value.(string); ok {
			o.FileExtension = v
		}
	case "module_system":
		if v, ok := value.(string); ok {
			o.ModuleSystem = v
		}
	}
}

// Clone creates a deep copy of the options.
func (o TypeScriptOptions) Clone() TypeScriptOptions {
	return o // struct copy is sufficient since all fields are value types
}

// Validate checks if the options are valid and returns an error if not.
func (o *TypeScriptOptions) Validate() error {
	// Validate output style
	validOutputStyles := []string{"interface", "type", "class"}
	if !contains(validOutputStyles, o.OutputStyle) {
		return &OptionsError{
			Field:   "OutputStyle",
			Value:   o.OutputStyle,
			Message: "unsupported output style",
			Valid:   validOutputStyles,
		}
	}

	// Validate naming convention
	validNamingConventions := []string{"PascalCase", "camelCase", "snake_case", "kebab-case"}
	if !contains(validNamingConventions, o.NamingConvention) {
		return &OptionsError{
			Field:   "NamingConvention",
			Value:   o.NamingConvention,
			Message: "unsupported naming convention",
			Valid:   validNamingConventions,
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

	// Validate validator library
	if o.GenerateValidators {
		validValidatorLibraries := []string{"zod", "yup", "joi", "ajv"}
		if !contains(validValidatorLibraries, o.ValidatorLibrary) {
			return &OptionsError{
				Field:   "ValidatorLibrary",
				Value:   o.ValidatorLibrary,
				Message: "unsupported validator library",
				Valid:   validValidatorLibraries,
			}
		}
	}

	// Validate array style
	validArrayStyles := []string{"Array<T>", "T[]"}
	if !contains(validArrayStyles, o.ArrayStyle) {
		return &OptionsError{
			Field:   "ArrayStyle",
			Value:   o.ArrayStyle,
			Message: "unsupported array style",
			Valid:   validArrayStyles,
		}
	}

	// Validate object style
	validObjectStyles := []string{"interface", "type", "Record<string, T>"}
	if !contains(validObjectStyles, o.ObjectStyle) {
		return &OptionsError{
			Field:   "ObjectStyle",
			Value:   o.ObjectStyle,
			Message: "unsupported object style",
			Valid:   validObjectStyles,
		}
	}

	// Validate JSDoc style
	validJSDocStyles := []string{"standard", "tsdoc"}
	if !contains(validJSDocStyles, o.JSDocStyle) {
		return &OptionsError{
			Field:   "JSDocStyle",
			Value:   o.JSDocStyle,
			Message: "unsupported JSDoc style",
			Valid:   validJSDocStyles,
		}
	}

	// Validate file extension
	validFileExtensions := []string{".ts", ".d.ts"}
	if !contains(validFileExtensions, o.FileExtension) {
		return &OptionsError{
			Field:   "FileExtension",
			Value:   o.FileExtension,
			Message: "unsupported file extension",
			Valid:   validFileExtensions,
		}
	}

	// Validate module system
	validModuleSystems := []string{"es6", "commonjs", "umd", "none"}
	if !contains(validModuleSystems, o.ModuleSystem) {
		return &OptionsError{
			Field:   "ModuleSystem",
			Value:   o.ModuleSystem,
			Message: "unsupported module system",
			Valid:   validModuleSystems,
		}
	}

	return nil
}

// OptionsError represents an error in TypeScript options configuration.
type OptionsError struct {
	Field   string
	Value   any
	Message string
	Valid   []string
}

// Error implements the error interface.
func (e *OptionsError) Error() string {
	msg := "invalid TypeScript option"
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

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
