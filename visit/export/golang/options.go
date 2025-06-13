package golang

import (
	"fmt"
	"strings"
)

// GoOptions contains configuration options for Go code generation.
type GoOptions struct {
	// OutputStyle determines the Go output style
	// Supported values: "struct", "interface", "type_alias"
	OutputStyle string

	// PackageName specifies the package name for generated code
	PackageName string

	// NamingConvention specifies the naming convention for types
	// Supported values: "PascalCase", "camelCase"
	NamingConvention string

	// FieldNamingConvention specifies the naming convention for struct fields
	// Supported values: "PascalCase", "camelCase"
	FieldNamingConvention string

	// IncludeComments determines whether to include comments
	IncludeComments bool

	// IncludeExamples determines whether to include example values in comments
	IncludeExamples bool

	// IncludeJSONTags determines whether to include JSON struct tags
	IncludeJSONTags bool

	// JSONTagStyle determines the JSON tag naming style
	// Supported values: "snake_case", "camelCase", "kebab-case"
	JSONTagStyle string

	// IncludeValidationTags determines whether to include validation tags
	IncludeValidationTags bool

	// ValidationTagStyle determines the validation tag format
	// Supported values: "go-playground", "ozzo", "custom"
	ValidationTagStyle string

	// IncludeXMLTags determines whether to include XML struct tags
	IncludeXMLTags bool

	// IncludeYAMLTags determines whether to include YAML struct tags
	IncludeYAMLTags bool

	// UsePointers determines whether to use pointers for optional fields
	UsePointers bool

	// UseOmitEmpty determines whether to add omitempty to JSON tags
	UseOmitEmpty bool

	// IndentStyle determines indentation style
	// Supported values: "tabs", "spaces"
	IndentStyle string

	// IndentSize specifies the number of spaces for indentation (when using spaces)
	IndentSize int

	// IncludeImports determines whether to include import statements
	IncludeImports bool

	// ImportStyle determines the import organization style
	// Supported values: "grouped", "single", "goimports"
	ImportStyle string

	// GenerateInterfaces determines whether to generate interfaces for polymorphic types
	GenerateInterfaces bool

	// InterfacePrefix specifies a prefix for generated interfaces
	InterfacePrefix string

	// InterfaceSuffix specifies a suffix for generated interfaces
	InterfaceSuffix string

	// GenerateConstructors determines whether to generate constructor functions
	GenerateConstructors bool

	// ConstructorPrefix specifies a prefix for constructor functions
	ConstructorPrefix string

	// GenerateValidators determines whether to generate validation methods
	GenerateValidators bool

	// ValidatorPrefix specifies a prefix for validation methods
	ValidatorPrefix string

	// GenerateStringers determines whether to generate String() methods
	GenerateStringers bool

	// GenerateGetters determines whether to generate getter methods
	GenerateGetters bool

	// GenerateSetters determines whether to generate setter methods
	GenerateSetters bool

	// UseGenerics determines whether to use Go generics (Go 1.18+)
	UseGenerics bool

	// GoVersion specifies the target Go version
	// Supported values: "1.18", "1.19", "1.20", "1.21", "1.22"
	GoVersion string

	// FileHeader specifies a header comment for generated files
	FileHeader string

	// ModulePath specifies the Go module path for imports
	ModulePath string

	// ExtraImports specifies additional imports to include
	ExtraImports []string

	// CustomTypeMappings allows custom type mappings
	CustomTypeMappings map[string]string

	// StructTagOptions specifies additional struct tag options
	StructTagOptions map[string]string

	// GenerateEnums determines whether to generate typed enums
	GenerateEnums bool

	// EnumStyle determines how enums are represented
	// Supported values: "const", "type", "string"
	EnumStyle string

	// GenerateUnions determines whether to generate union types
	GenerateUnions bool

	// UnionStyle determines how unions are represented
	// Supported values: "interface", "embedded", "discriminated"
	UnionStyle string
}

// DefaultGoOptions returns the default Go generation options.
func DefaultGoOptions() GoOptions {
	return GoOptions{
		OutputStyle:           "struct",
		PackageName:           "main",
		NamingConvention:      "PascalCase",
		FieldNamingConvention: "PascalCase",
		IncludeComments:       true,
		IncludeExamples:       false,
		IncludeJSONTags:       true,
		JSONTagStyle:          "snake_case",
		IncludeValidationTags: false,
		ValidationTagStyle:    "go-playground",
		IncludeXMLTags:        false,
		IncludeYAMLTags:       false,
		UsePointers:           true,
		UseOmitEmpty:          true,
		IndentStyle:           "tabs",
		IndentSize:            4,
		IncludeImports:        true,
		ImportStyle:           "goimports",
		GenerateInterfaces:    false,
		InterfacePrefix:       "",
		InterfaceSuffix:       "Interface",
		GenerateConstructors:  false,
		ConstructorPrefix:     "New",
		GenerateValidators:    false,
		ValidatorPrefix:       "Validate",
		GenerateStringers:     false,
		GenerateGetters:       false,
		GenerateSetters:       false,
		UseGenerics:           false,
		GoVersion:             "1.21",
		FileHeader:            "",
		ModulePath:            "",
		ExtraImports:          []string{},
		CustomTypeMappings:    make(map[string]string),
		StructTagOptions:      make(map[string]string),
		GenerateEnums:         true,
		EnumStyle:             "const",
		GenerateUnions:        false,
		UnionStyle:            "interface",
	}
}

// SetOption sets a configuration option by key.
func (o *GoOptions) SetOption(key string, value any) {
	switch key {
	case "output_style":
		if v, ok := value.(string); ok {
			o.OutputStyle = v
		}
	case "package_name":
		if v, ok := value.(string); ok {
			o.PackageName = v
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
	case "include_json_tags":
		if v, ok := value.(bool); ok {
			o.IncludeJSONTags = v
		}
	case "json_tag_style":
		if v, ok := value.(string); ok {
			o.JSONTagStyle = v
		}
	case "include_validation_tags":
		if v, ok := value.(bool); ok {
			o.IncludeValidationTags = v
		}
	case "validation_tag_style":
		if v, ok := value.(string); ok {
			o.ValidationTagStyle = v
		}
	case "include_xml_tags":
		if v, ok := value.(bool); ok {
			o.IncludeXMLTags = v
		}
	case "include_yaml_tags":
		if v, ok := value.(bool); ok {
			o.IncludeYAMLTags = v
		}
	case "use_pointers":
		if v, ok := value.(bool); ok {
			o.UsePointers = v
		}
	case "use_omit_empty":
		if v, ok := value.(bool); ok {
			o.UseOmitEmpty = v
		}
	case "indent_style":
		if v, ok := value.(string); ok {
			o.IndentStyle = v
		}
	case "indent_size":
		if v, ok := value.(int); ok {
			o.IndentSize = v
		}
	case "include_imports":
		if v, ok := value.(bool); ok {
			o.IncludeImports = v
		}
	case "import_style":
		if v, ok := value.(string); ok {
			o.ImportStyle = v
		}
	case "generate_interfaces":
		if v, ok := value.(bool); ok {
			o.GenerateInterfaces = v
		}
	case "interface_prefix":
		if v, ok := value.(string); ok {
			o.InterfacePrefix = v
		}
	case "interface_suffix":
		if v, ok := value.(string); ok {
			o.InterfaceSuffix = v
		}
	case "generate_constructors":
		if v, ok := value.(bool); ok {
			o.GenerateConstructors = v
		}
	case "constructor_prefix":
		if v, ok := value.(string); ok {
			o.ConstructorPrefix = v
		}
	case "generate_validators":
		if v, ok := value.(bool); ok {
			o.GenerateValidators = v
		}
	case "validator_prefix":
		if v, ok := value.(string); ok {
			o.ValidatorPrefix = v
		}
	case "generate_stringers":
		if v, ok := value.(bool); ok {
			o.GenerateStringers = v
		}
	case "generate_getters":
		if v, ok := value.(bool); ok {
			o.GenerateGetters = v
		}
	case "generate_setters":
		if v, ok := value.(bool); ok {
			o.GenerateSetters = v
		}
	case "use_generics":
		if v, ok := value.(bool); ok {
			o.UseGenerics = v
		}
	case "go_version":
		if v, ok := value.(string); ok {
			o.GoVersion = v
		}
	case "file_header":
		if v, ok := value.(string); ok {
			o.FileHeader = v
		}
	case "module_path":
		if v, ok := value.(string); ok {
			o.ModulePath = v
		}
	case "extra_imports":
		if v, ok := value.([]string); ok {
			o.ExtraImports = v
		}
	case "custom_type_mappings":
		if v, ok := value.(map[string]string); ok {
			o.CustomTypeMappings = v
		}
	case "struct_tag_options":
		if v, ok := value.(map[string]string); ok {
			o.StructTagOptions = v
		}
	case "generate_enums":
		if v, ok := value.(bool); ok {
			o.GenerateEnums = v
		}
	case "enum_style":
		if v, ok := value.(string); ok {
			o.EnumStyle = v
		}
	case "generate_unions":
		if v, ok := value.(bool); ok {
			o.GenerateUnions = v
		}
	case "union_style":
		if v, ok := value.(string); ok {
			o.UnionStyle = v
		}
	}
}

// Clone creates a deep copy of the options.
func (o GoOptions) Clone() GoOptions {
	clone := o
	// Deep copy slices and maps
	clone.ExtraImports = make([]string, len(o.ExtraImports))
	copy(clone.ExtraImports, o.ExtraImports)

	clone.CustomTypeMappings = make(map[string]string)
	for k, v := range o.CustomTypeMappings {
		clone.CustomTypeMappings[k] = v
	}

	clone.StructTagOptions = make(map[string]string)
	for k, v := range o.StructTagOptions {
		clone.StructTagOptions[k] = v
	}

	return clone
}

// Validate checks if the options are valid and returns an error if not.
func (o *GoOptions) Validate() error {
	// Validate output style
	validOutputStyles := []string{"struct", "interface", "type_alias"}
	if !contains(validOutputStyles, o.OutputStyle) {
		return &OptionsError{
			Field:   "OutputStyle",
			Value:   o.OutputStyle,
			Message: "unsupported output style",
			Valid:   validOutputStyles,
		}
	}

	// Validate naming convention
	validNamingConventions := []string{"PascalCase", "camelCase"}
	if !contains(validNamingConventions, o.NamingConvention) {
		return &OptionsError{
			Field:   "NamingConvention",
			Value:   o.NamingConvention,
			Message: "unsupported naming convention",
			Valid:   validNamingConventions,
		}
	}

	// Validate field naming convention
	validFieldNamingConventions := []string{"PascalCase", "camelCase"}
	if !contains(validFieldNamingConventions, o.FieldNamingConvention) {
		return &OptionsError{
			Field:   "FieldNamingConvention",
			Value:   o.FieldNamingConvention,
			Message: "unsupported field naming convention",
			Valid:   validFieldNamingConventions,
		}
	}

	// Validate JSON tag style
	validJSONTagStyles := []string{"snake_case", "camelCase", "kebab-case"}
	if !contains(validJSONTagStyles, o.JSONTagStyle) {
		return &OptionsError{
			Field:   "JSONTagStyle",
			Value:   o.JSONTagStyle,
			Message: "unsupported JSON tag style",
			Valid:   validJSONTagStyles,
		}
	}

	// Validate validation tag style
	validValidationTagStyles := []string{"go-playground", "ozzo", "custom"}
	if !contains(validValidationTagStyles, o.ValidationTagStyle) {
		return &OptionsError{
			Field:   "ValidationTagStyle",
			Value:   o.ValidationTagStyle,
			Message: "unsupported validation tag style",
			Valid:   validValidationTagStyles,
		}
	}

	// Validate indent style
	validIndentStyles := []string{"tabs", "spaces"}
	if !contains(validIndentStyles, o.IndentStyle) {
		return &OptionsError{
			Field:   "IndentStyle",
			Value:   o.IndentStyle,
			Message: "unsupported indent style",
			Valid:   validIndentStyles,
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
	validImportStyles := []string{"grouped", "single", "goimports"}
	if !contains(validImportStyles, o.ImportStyle) {
		return &OptionsError{
			Field:   "ImportStyle",
			Value:   o.ImportStyle,
			Message: "unsupported import style",
			Valid:   validImportStyles,
		}
	}

	// Validate Go version
	validGoVersions := []string{"1.18", "1.19", "1.20", "1.21", "1.22"}
	if !contains(validGoVersions, o.GoVersion) {
		return &OptionsError{
			Field:   "GoVersion",
			Value:   o.GoVersion,
			Message: "unsupported Go version",
			Valid:   validGoVersions,
		}
	}

	// Validate enum style
	validEnumStyles := []string{"const", "type", "string"}
	if !contains(validEnumStyles, o.EnumStyle) {
		return &OptionsError{
			Field:   "EnumStyle",
			Value:   o.EnumStyle,
			Message: "unsupported enum style",
			Valid:   validEnumStyles,
		}
	}

	// Validate union style
	validUnionStyles := []string{"interface", "embedded", "discriminated"}
	if !contains(validUnionStyles, o.UnionStyle) {
		return &OptionsError{
			Field:   "UnionStyle",
			Value:   o.UnionStyle,
			Message: "unsupported union style",
			Valid:   validUnionStyles,
		}
	}

	// Validate package name
	if o.PackageName == "" {
		return &OptionsError{
			Field:   "PackageName",
			Value:   o.PackageName,
			Message: "package name cannot be empty",
		}
	}

	return nil
}

// OptionsError represents an error in Go options configuration.
type OptionsError struct {
	Field   string
	Value   any
	Message string
	Valid   []string
}

// Error implements the error interface.
func (e *OptionsError) Error() string {
	msg := "invalid Go option"
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
