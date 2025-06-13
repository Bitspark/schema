package python

import (
	"fmt"
	"strings"
)

// PythonOptions configures the behavior of the Python generator.
type PythonOptions struct {
	// OutputStyle determines the Python output style
	// Supported values: "pydantic", "dataclass", "class", "namedtuple"
	OutputStyle string

	// PydanticVersion specifies the Pydantic version to target
	// Supported values: "v1", "v2"
	PydanticVersion string

	// NamingConvention specifies the naming convention for classes and fields
	// Supported values: "PascalCase", "snake_case"
	NamingConvention string

	// FieldNamingConvention specifies the naming convention for fields
	// Supported values: "snake_case", "camelCase"
	FieldNamingConvention string

	// IncludeComments determines whether to include docstrings and comments
	IncludeComments bool

	// IncludeExamples determines whether to include example values in docstrings
	IncludeExamples bool

	// IncludeDefaults determines whether to include default values
	IncludeDefaults bool

	// StrictMode enables stricter type hints and validation
	StrictMode bool

	// UseOptional uses Optional[T] instead of T | None for optional fields
	UseOptional bool

	// IndentSize specifies the number of spaces for indentation
	IndentSize int

	// UseTabsForIndentation uses tabs instead of spaces
	UseTabsForIndentation bool

	// IncludeImports determines whether to include import statements
	IncludeImports bool

	// ImportStyle determines the import style
	// Supported values: "absolute", "relative"
	ImportStyle string

	// UseTypeHints enables type hints in generated code
	UseTypeHints bool

	// TypeHintStyle determines the type hint style
	// Supported values: "typing", "builtin" (Python 3.9+)
	TypeHintStyle string

	// GenerateValidators determines whether to generate validation methods
	GenerateValidators bool

	// ValidatorStyle determines the validation approach
	// Supported values: "pydantic", "custom", "none"
	ValidatorStyle string

	// UseEnums generates Python enums for string enums
	UseEnums bool

	// EnumStyle determines how enums are represented
	// Supported values: "Enum", "StrEnum", "Literal"
	EnumStyle string

	// IncludeSerializers determines whether to include serialization methods
	IncludeSerializers bool

	// SerializerStyle determines the serialization approach
	// Supported values: "dict", "json", "both"
	SerializerStyle string

	// UseDataclassFeatures enables dataclass-specific features
	UseDataclassFeatures bool

	// DataclassOptions specifies dataclass configuration
	// Supported values: "frozen", "slots", "kw_only"
	DataclassOptions []string

	// IncludeDocstrings includes detailed docstrings
	IncludeDocstrings bool

	// DocstringStyle determines docstring format
	// Supported values: "google", "numpy", "sphinx"
	DocstringStyle string

	// FileHeader specifies a header comment for generated files
	FileHeader string

	// ModuleName specifies the module name for imports
	ModuleName string

	// BaseClass specifies a base class for generated classes
	BaseClass string

	// UseForwardRefs enables forward references for recursive types
	UseForwardRefs bool

	// PythonVersion specifies the target Python version
	// Supported values: "3.8", "3.9", "3.10", "3.11", "3.12"
	PythonVersion string

	// ExtraImports specifies additional imports to include
	ExtraImports []string

	// CustomTypeMapping allows custom type mappings
	CustomTypeMapping map[string]string
}

// DefaultPythonOptions returns the default options for Python generation.
func DefaultPythonOptions() PythonOptions {
	return PythonOptions{
		OutputStyle:           "pydantic",
		PydanticVersion:       "v2",
		NamingConvention:      "PascalCase",
		FieldNamingConvention: "snake_case",
		IncludeComments:       true,
		IncludeExamples:       true,
		IncludeDefaults:       true,
		StrictMode:            false,
		UseOptional:           true,
		IndentSize:            4,
		UseTabsForIndentation: false,
		IncludeImports:        true,
		ImportStyle:           "absolute",
		UseTypeHints:          true,
		TypeHintStyle:         "typing",
		GenerateValidators:    false,
		ValidatorStyle:        "pydantic",
		UseEnums:              true,
		EnumStyle:             "Enum",
		IncludeSerializers:    false,
		SerializerStyle:       "dict",
		UseDataclassFeatures:  false,
		DataclassOptions:      []string{},
		IncludeDocstrings:     true,
		DocstringStyle:        "google",
		FileHeader:            "",
		ModuleName:            "",
		BaseClass:             "",
		UseForwardRefs:        false,
		PythonVersion:         "3.9",
		ExtraImports:          []string{},
		CustomTypeMapping:     make(map[string]string),
	}
}

// SetOption implements the option setter interface for functional options.
func (o *PythonOptions) SetOption(key string, value any) {
	switch key {
	case "output_style":
		if v, ok := value.(string); ok {
			o.OutputStyle = v
		}
	case "pydantic_version":
		if v, ok := value.(string); ok {
			o.PydanticVersion = v
		}
	case "naming_convention":
		if v, ok := value.(string); ok {
			o.NamingConvention = v
		}
	case "field_naming_convention":
		if v, ok := value.(string); ok {
			o.FieldNamingConvention = v
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
	case "use_optional":
		if v, ok := value.(bool); ok {
			o.UseOptional = v
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
	case "import_style":
		if v, ok := value.(string); ok {
			o.ImportStyle = v
		}
	case "use_type_hints":
		if v, ok := value.(bool); ok {
			o.UseTypeHints = v
		}
	case "type_hint_style":
		if v, ok := value.(string); ok {
			o.TypeHintStyle = v
		}
	case "generate_validators":
		if v, ok := value.(bool); ok {
			o.GenerateValidators = v
		}
	case "validator_style":
		if v, ok := value.(string); ok {
			o.ValidatorStyle = v
		}
	case "use_enums":
		if v, ok := value.(bool); ok {
			o.UseEnums = v
		}
	case "enum_style":
		if v, ok := value.(string); ok {
			o.EnumStyle = v
		}
	case "include_serializers":
		if v, ok := value.(bool); ok {
			o.IncludeSerializers = v
		}
	case "serializer_style":
		if v, ok := value.(string); ok {
			o.SerializerStyle = v
		}
	case "use_dataclass_features":
		if v, ok := value.(bool); ok {
			o.UseDataclassFeatures = v
		}
	case "dataclass_options":
		if v, ok := value.([]string); ok {
			o.DataclassOptions = v
		}
	case "include_docstrings":
		if v, ok := value.(bool); ok {
			o.IncludeDocstrings = v
		}
	case "docstring_style":
		if v, ok := value.(string); ok {
			o.DocstringStyle = v
		}
	case "file_header":
		if v, ok := value.(string); ok {
			o.FileHeader = v
		}
	case "module_name":
		if v, ok := value.(string); ok {
			o.ModuleName = v
		}
	case "base_class":
		if v, ok := value.(string); ok {
			o.BaseClass = v
		}
	case "use_forward_refs":
		if v, ok := value.(bool); ok {
			o.UseForwardRefs = v
		}
	case "python_version":
		if v, ok := value.(string); ok {
			o.PythonVersion = v
		}
	case "extra_imports":
		if v, ok := value.([]string); ok {
			o.ExtraImports = v
		}
	case "custom_type_mapping":
		if v, ok := value.(map[string]string); ok {
			o.CustomTypeMapping = v
		}
	}
}

// Clone creates a deep copy of the options.
func (o PythonOptions) Clone() PythonOptions {
	clone := o
	// Deep copy slices and maps
	clone.DataclassOptions = make([]string, len(o.DataclassOptions))
	copy(clone.DataclassOptions, o.DataclassOptions)

	clone.ExtraImports = make([]string, len(o.ExtraImports))
	copy(clone.ExtraImports, o.ExtraImports)

	clone.CustomTypeMapping = make(map[string]string)
	for k, v := range o.CustomTypeMapping {
		clone.CustomTypeMapping[k] = v
	}

	return clone
}

// Validate checks if the options are valid and returns an error if not.
func (o *PythonOptions) Validate() error {
	// Validate output style
	validOutputStyles := []string{"pydantic", "dataclass", "class", "namedtuple"}
	if !contains(validOutputStyles, o.OutputStyle) {
		return &OptionsError{
			Field:   "OutputStyle",
			Value:   o.OutputStyle,
			Message: "unsupported output style",
			Valid:   validOutputStyles,
		}
	}

	// Validate Pydantic version
	validPydanticVersions := []string{"v1", "v2"}
	if !contains(validPydanticVersions, o.PydanticVersion) {
		return &OptionsError{
			Field:   "PydanticVersion",
			Value:   o.PydanticVersion,
			Message: "unsupported Pydantic version",
			Valid:   validPydanticVersions,
		}
	}

	// Validate naming convention
	validNamingConventions := []string{"PascalCase", "snake_case"}
	if !contains(validNamingConventions, o.NamingConvention) {
		return &OptionsError{
			Field:   "NamingConvention",
			Value:   o.NamingConvention,
			Message: "unsupported naming convention",
			Valid:   validNamingConventions,
		}
	}

	// Validate field naming convention
	validFieldNamingConventions := []string{"snake_case", "camelCase"}
	if !contains(validFieldNamingConventions, o.FieldNamingConvention) {
		return &OptionsError{
			Field:   "FieldNamingConvention",
			Value:   o.FieldNamingConvention,
			Message: "unsupported field naming convention",
			Valid:   validFieldNamingConventions,
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

	// Validate import style
	validImportStyles := []string{"absolute", "relative"}
	if !contains(validImportStyles, o.ImportStyle) {
		return &OptionsError{
			Field:   "ImportStyle",
			Value:   o.ImportStyle,
			Message: "unsupported import style",
			Valid:   validImportStyles,
		}
	}

	// Validate type hint style
	validTypeHintStyles := []string{"typing", "builtin"}
	if !contains(validTypeHintStyles, o.TypeHintStyle) {
		return &OptionsError{
			Field:   "TypeHintStyle",
			Value:   o.TypeHintStyle,
			Message: "unsupported type hint style",
			Valid:   validTypeHintStyles,
		}
	}

	// Validate validator style
	validValidatorStyles := []string{"pydantic", "custom", "none"}
	if !contains(validValidatorStyles, o.ValidatorStyle) {
		return &OptionsError{
			Field:   "ValidatorStyle",
			Value:   o.ValidatorStyle,
			Message: "unsupported validator style",
			Valid:   validValidatorStyles,
		}
	}

	// Validate enum style
	validEnumStyles := []string{"Enum", "StrEnum", "Literal"}
	if !contains(validEnumStyles, o.EnumStyle) {
		return &OptionsError{
			Field:   "EnumStyle",
			Value:   o.EnumStyle,
			Message: "unsupported enum style",
			Valid:   validEnumStyles,
		}
	}

	// Validate serializer style
	validSerializerStyles := []string{"dict", "json", "both"}
	if !contains(validSerializerStyles, o.SerializerStyle) {
		return &OptionsError{
			Field:   "SerializerStyle",
			Value:   o.SerializerStyle,
			Message: "unsupported serializer style",
			Valid:   validSerializerStyles,
		}
	}

	// Validate docstring style
	validDocstringStyles := []string{"google", "numpy", "sphinx"}
	if !contains(validDocstringStyles, o.DocstringStyle) {
		return &OptionsError{
			Field:   "DocstringStyle",
			Value:   o.DocstringStyle,
			Message: "unsupported docstring style",
			Valid:   validDocstringStyles,
		}
	}

	// Validate Python version
	validPythonVersions := []string{"3.8", "3.9", "3.10", "3.11", "3.12"}
	if !contains(validPythonVersions, o.PythonVersion) {
		return &OptionsError{
			Field:   "PythonVersion",
			Value:   o.PythonVersion,
			Message: "unsupported Python version",
			Valid:   validPythonVersions,
		}
	}

	// Validate dataclass options
	validDataclassOptions := []string{"frozen", "slots", "kw_only"}
	for _, opt := range o.DataclassOptions {
		if !contains(validDataclassOptions, opt) {
			return &OptionsError{
				Field:   "DataclassOptions",
				Value:   opt,
				Message: "unsupported dataclass option",
				Valid:   validDataclassOptions,
			}
		}
	}

	return nil
}

// OptionsError represents an error in Python options configuration.
type OptionsError struct {
	Field   string
	Value   any
	Message string
	Valid   []string
}

// Error implements the error interface.
func (e *OptionsError) Error() string {
	msg := "invalid Python option"
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
