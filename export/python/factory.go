package python

import (
	"defs.dev/schema/export"
)

// Option represents a functional option for configuring the Python generator.
type Option func(*PythonOptions)

// WithOutputStyle sets the output style for Python generation.
func WithOutputStyle(style string) Option {
	return func(o *PythonOptions) {
		o.OutputStyle = style
	}
}

// WithPydanticVersion sets the Pydantic version to target.
func WithPydanticVersion(version string) Option {
	return func(o *PythonOptions) {
		o.PydanticVersion = version
	}
}

// WithNamingConvention sets the naming convention for classes.
func WithNamingConvention(convention string) Option {
	return func(o *PythonOptions) {
		o.NamingConvention = convention
	}
}

// WithFieldNamingConvention sets the naming convention for fields.
func WithFieldNamingConvention(convention string) Option {
	return func(o *PythonOptions) {
		o.FieldNamingConvention = convention
	}
}

// WithComments enables or disables comment generation.
func WithComments(include bool) Option {
	return func(o *PythonOptions) {
		o.IncludeComments = include
	}
}

// WithExamples enables or disables example generation.
func WithExamples(include bool) Option {
	return func(o *PythonOptions) {
		o.IncludeExamples = include
	}
}

// WithDefaults enables or disables default value generation.
func WithDefaults(include bool) Option {
	return func(o *PythonOptions) {
		o.IncludeDefaults = include
	}
}

// WithStrictMode enables or disables strict mode.
func WithStrictMode(strict bool) Option {
	return func(o *PythonOptions) {
		o.StrictMode = strict
	}
}

// WithOptional sets whether to use Optional[T] instead of T | None.
func WithOptional(useOptional bool) Option {
	return func(o *PythonOptions) {
		o.UseOptional = useOptional
	}
}

// WithIndentSize sets the indentation size.
func WithIndentSize(size int) Option {
	return func(o *PythonOptions) {
		o.IndentSize = size
	}
}

// WithTabs enables or disables tab indentation.
func WithTabs(useTabs bool) Option {
	return func(o *PythonOptions) {
		o.UseTabsForIndentation = useTabs
	}
}

// WithImports enables or disables import generation.
func WithImports(include bool) Option {
	return func(o *PythonOptions) {
		o.IncludeImports = include
	}
}

// WithImportStyle sets the import style.
func WithImportStyle(style string) Option {
	return func(o *PythonOptions) {
		o.ImportStyle = style
	}
}

// WithTypeHints enables or disables type hints.
func WithTypeHints(useTypeHints bool) Option {
	return func(o *PythonOptions) {
		o.UseTypeHints = useTypeHints
	}
}

// WithTypeHintStyle sets the type hint style.
func WithTypeHintStyle(style string) Option {
	return func(o *PythonOptions) {
		o.TypeHintStyle = style
	}
}

// WithValidators enables or disables validator generation.
func WithValidators(generate bool) Option {
	return func(o *PythonOptions) {
		o.GenerateValidators = generate
	}
}

// WithValidatorStyle sets the validator style.
func WithValidatorStyle(style string) Option {
	return func(o *PythonOptions) {
		o.ValidatorStyle = style
	}
}

// WithEnums enables or disables enum generation.
func WithEnums(useEnums bool) Option {
	return func(o *PythonOptions) {
		o.UseEnums = useEnums
	}
}

// WithEnumStyle sets the enum style.
func WithEnumStyle(style string) Option {
	return func(o *PythonOptions) {
		o.EnumStyle = style
	}
}

// WithSerializers enables or disables serializer generation.
func WithSerializers(include bool) Option {
	return func(o *PythonOptions) {
		o.IncludeSerializers = include
	}
}

// WithSerializerStyle sets the serializer style.
func WithSerializerStyle(style string) Option {
	return func(o *PythonOptions) {
		o.SerializerStyle = style
	}
}

// WithDataclassFeatures enables or disables dataclass features.
func WithDataclassFeatures(use bool) Option {
	return func(o *PythonOptions) {
		o.UseDataclassFeatures = use
	}
}

// WithDataclassOptions sets the dataclass options.
func WithDataclassOptions(options []string) Option {
	return func(o *PythonOptions) {
		o.DataclassOptions = options
	}
}

// WithDocstrings enables or disables docstring generation.
func WithDocstrings(include bool) Option {
	return func(o *PythonOptions) {
		o.IncludeDocstrings = include
	}
}

// WithDocstringStyle sets the docstring style.
func WithDocstringStyle(style string) Option {
	return func(o *PythonOptions) {
		o.DocstringStyle = style
	}
}

// WithFileHeader sets the file header comment.
func WithFileHeader(header string) Option {
	return func(o *PythonOptions) {
		o.FileHeader = header
	}
}

// WithModuleName sets the module name.
func WithModuleName(name string) Option {
	return func(o *PythonOptions) {
		o.ModuleName = name
	}
}

// WithBaseClass sets the base class for generated classes.
func WithBaseClass(baseClass string) Option {
	return func(o *PythonOptions) {
		o.BaseClass = baseClass
	}
}

// WithForwardRefs enables or disables forward references.
func WithForwardRefs(use bool) Option {
	return func(o *PythonOptions) {
		o.UseForwardRefs = use
	}
}

// WithPythonVersion sets the target Python version.
func WithPythonVersion(version string) Option {
	return func(o *PythonOptions) {
		o.PythonVersion = version
	}
}

// WithExtraImports sets additional imports.
func WithExtraImports(imports []string) Option {
	return func(o *PythonOptions) {
		o.ExtraImports = imports
	}
}

// WithCustomTypeMapping sets custom type mappings.
func WithCustomTypeMapping(mapping map[string]string) Option {
	return func(o *PythonOptions) {
		o.CustomTypeMapping = mapping
	}
}

// NewPythonGenerator creates a new Python generator with the given options.
func NewPythonGenerator(opts ...Option) *Generator {
	options := DefaultPythonOptions()

	for _, opt := range opts {
		opt(&options)
	}

	return NewGenerator(options)
}

// Preset configurations

// PydanticV2Preset returns options for Pydantic v2 generation.
func PydanticV2Preset() []Option {
	return []Option{
		WithOutputStyle("pydantic"),
		WithPydanticVersion("v2"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("snake_case"),
		WithTypeHints(true),
		WithTypeHintStyle("typing"),
		WithOptional(true),
		WithEnums(true),
		WithEnumStyle("Enum"),
		WithDocstrings(true),
		WithDocstringStyle("google"),
		WithComments(true),
		WithExamples(true),
		WithDefaults(true),
		WithImports(true),
		WithPythonVersion("3.9"),
	}
}

// PydanticV1Preset returns options for Pydantic v1 generation.
func PydanticV1Preset() []Option {
	return []Option{
		WithOutputStyle("pydantic"),
		WithPydanticVersion("v1"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("snake_case"),
		WithTypeHints(true),
		WithTypeHintStyle("typing"),
		WithOptional(true),
		WithEnums(true),
		WithEnumStyle("Enum"),
		WithDocstrings(true),
		WithDocstringStyle("google"),
		WithComments(true),
		WithExamples(true),
		WithDefaults(true),
		WithImports(true),
		WithPythonVersion("3.8"),
	}
}

// DataclassPreset returns options for dataclass generation.
func DataclassPreset() []Option {
	return []Option{
		WithOutputStyle("dataclass"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("snake_case"),
		WithTypeHints(true),
		WithTypeHintStyle("typing"),
		WithOptional(true),
		WithEnums(true),
		WithEnumStyle("Enum"),
		WithDocstrings(true),
		WithDocstringStyle("google"),
		WithComments(true),
		WithExamples(true),
		WithDefaults(true),
		WithImports(true),
		WithDataclassFeatures(true),
		WithDataclassOptions([]string{"frozen"}),
		WithPythonVersion("3.9"),
	}
}

// ModernPythonPreset returns options for modern Python (3.10+) generation.
func ModernPythonPreset() []Option {
	return []Option{
		WithOutputStyle("pydantic"),
		WithPydanticVersion("v2"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("snake_case"),
		WithTypeHints(true),
		WithTypeHintStyle("builtin"),
		WithOptional(false), // Use T | None instead of Optional[T]
		WithEnums(true),
		WithEnumStyle("StrEnum"),
		WithDocstrings(true),
		WithDocstringStyle("google"),
		WithComments(true),
		WithExamples(true),
		WithDefaults(true),
		WithImports(true),
		WithPythonVersion("3.10"),
	}
}

// MinimalPreset returns options for minimal Python generation.
func MinimalPreset() []Option {
	return []Option{
		WithOutputStyle("class"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("snake_case"),
		WithTypeHints(false),
		WithOptional(false),
		WithEnums(false),
		WithDocstrings(false),
		WithComments(false),
		WithExamples(false),
		WithDefaults(false),
		WithImports(false),
		WithPythonVersion("3.8"),
	}
}

// StrictPreset returns options for strict Python generation.
func StrictPreset() []Option {
	return []Option{
		WithOutputStyle("pydantic"),
		WithPydanticVersion("v2"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("snake_case"),
		WithTypeHints(true),
		WithTypeHintStyle("typing"),
		WithOptional(true),
		WithStrictMode(true),
		WithEnums(true),
		WithEnumStyle("Enum"),
		WithDocstrings(true),
		WithDocstringStyle("sphinx"),
		WithComments(true),
		WithExamples(true),
		WithDefaults(true),
		WithImports(true),
		WithValidators(true),
		WithValidatorStyle("pydantic"),
		WithPythonVersion("3.9"),
	}
}

// Factory function for integration with the export system

// CreatePythonGenerator creates a Python generator from a map of options.
func CreatePythonGenerator(options map[string]any) (export.Generator, error) {
	pythonOptions := DefaultPythonOptions()

	// Apply options from map
	for key, value := range options {
		pythonOptions.SetOption(key, value)
	}

	// Validate options
	if err := pythonOptions.Validate(); err != nil {
		return nil, err
	}

	generator := NewGenerator(pythonOptions)
	return generator, nil
}

// PythonGeneratorFactory creates a Python generator factory function.
func PythonGeneratorFactory(options ...any) (export.Generator, error) {
	// Convert options to map if needed
	optionsMap := make(map[string]any)

	for _, opt := range options {
		if optMap, ok := opt.(map[string]any); ok {
			for k, v := range optMap {
				optionsMap[k] = v
			}
		}
	}

	return CreatePythonGenerator(optionsMap)
}
