package golang

import (
	"defs.dev/schema/visitors/export"
)

// Option represents a functional option for configuring the Go generator.
type Option func(*GoOptions)

// WithOutputStyle sets the output style.
func WithOutputStyle(style string) Option {
	return func(o *GoOptions) {
		o.OutputStyle = style
	}
}

// WithPackageName sets the package name.
func WithPackageName(name string) Option {
	return func(o *GoOptions) {
		o.PackageName = name
	}
}

// WithNamingConvention sets the naming convention.
func WithNamingConvention(convention string) Option {
	return func(o *GoOptions) {
		o.NamingConvention = convention
	}
}

// WithFieldNamingConvention sets the field naming convention.
func WithFieldNamingConvention(convention string) Option {
	return func(o *GoOptions) {
		o.FieldNamingConvention = convention
	}
}

// WithComments enables or disables comments.
func WithComments(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeComments = enabled
	}
}

// WithExamples enables or disables examples in comments.
func WithExamples(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeExamples = enabled
	}
}

// WithJSONTags enables or disables JSON struct tags.
func WithJSONTags(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeJSONTags = enabled
	}
}

// WithJSONTagStyle sets the JSON tag naming style.
func WithJSONTagStyle(style string) Option {
	return func(o *GoOptions) {
		o.JSONTagStyle = style
	}
}

// WithValidationTags enables or disables validation tags.
func WithValidationTags(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeValidationTags = enabled
	}
}

// WithValidationTagStyle sets the validation tag style.
func WithValidationTagStyle(style string) Option {
	return func(o *GoOptions) {
		o.ValidationTagStyle = style
	}
}

// WithXMLTags enables or disables XML struct tags.
func WithXMLTags(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeXMLTags = enabled
	}
}

// WithYAMLTags enables or disables YAML struct tags.
func WithYAMLTags(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeYAMLTags = enabled
	}
}

// WithPointers enables or disables pointers for optional fields.
func WithPointers(enabled bool) Option {
	return func(o *GoOptions) {
		o.UsePointers = enabled
	}
}

// WithOmitEmpty enables or disables omitempty in JSON tags.
func WithOmitEmpty(enabled bool) Option {
	return func(o *GoOptions) {
		o.UseOmitEmpty = enabled
	}
}

// WithIndentStyle sets the indentation style.
func WithIndentStyle(style string) Option {
	return func(o *GoOptions) {
		o.IndentStyle = style
	}
}

// WithIndentSize sets the indentation size.
func WithIndentSize(size int) Option {
	return func(o *GoOptions) {
		o.IndentSize = size
	}
}

// WithImports enables or disables import statements.
func WithImports(enabled bool) Option {
	return func(o *GoOptions) {
		o.IncludeImports = enabled
	}
}

// WithImportStyle sets the import organization style.
func WithImportStyle(style string) Option {
	return func(o *GoOptions) {
		o.ImportStyle = style
	}
}

// WithInterfaces enables or disables interface generation.
func WithInterfaces(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateInterfaces = enabled
	}
}

// WithInterfacePrefix sets the interface prefix.
func WithInterfacePrefix(prefix string) Option {
	return func(o *GoOptions) {
		o.InterfacePrefix = prefix
	}
}

// WithInterfaceSuffix sets the interface suffix.
func WithInterfaceSuffix(suffix string) Option {
	return func(o *GoOptions) {
		o.InterfaceSuffix = suffix
	}
}

// WithConstructors enables or disables constructor generation.
func WithConstructors(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateConstructors = enabled
	}
}

// WithConstructorPrefix sets the constructor prefix.
func WithConstructorPrefix(prefix string) Option {
	return func(o *GoOptions) {
		o.ConstructorPrefix = prefix
	}
}

// WithValidators enables or disables validator generation.
func WithValidators(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateValidators = enabled
	}
}

// WithValidatorPrefix sets the validator prefix.
func WithValidatorPrefix(prefix string) Option {
	return func(o *GoOptions) {
		o.ValidatorPrefix = prefix
	}
}

// WithStringers enables or disables String method generation.
func WithStringers(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateStringers = enabled
	}
}

// WithGetters enables or disables getter method generation.
func WithGetters(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateGetters = enabled
	}
}

// WithSetters enables or disables setter method generation.
func WithSetters(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateSetters = enabled
	}
}

// WithGenerics enables or disables Go generics.
func WithGenerics(enabled bool) Option {
	return func(o *GoOptions) {
		o.UseGenerics = enabled
	}
}

// WithGoVersion sets the target Go version.
func WithGoVersion(version string) Option {
	return func(o *GoOptions) {
		o.GoVersion = version
	}
}

// WithFileHeader sets the file header comment.
func WithFileHeader(header string) Option {
	return func(o *GoOptions) {
		o.FileHeader = header
	}
}

// WithModulePath sets the Go module path.
func WithModulePath(path string) Option {
	return func(o *GoOptions) {
		o.ModulePath = path
	}
}

// WithExtraImports sets additional imports.
func WithExtraImports(imports []string) Option {
	return func(o *GoOptions) {
		o.ExtraImports = imports
	}
}

// WithCustomTypeMappings sets custom type mappings.
func WithCustomTypeMappings(mappings map[string]string) Option {
	return func(o *GoOptions) {
		o.CustomTypeMappings = mappings
	}
}

// WithStructTagOptions sets additional struct tag options.
func WithStructTagOptions(options map[string]string) Option {
	return func(o *GoOptions) {
		o.StructTagOptions = options
	}
}

// WithEnums enables or disables enum generation.
func WithEnums(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateEnums = enabled
	}
}

// WithEnumStyle sets the enum style.
func WithEnumStyle(style string) Option {
	return func(o *GoOptions) {
		o.EnumStyle = style
	}
}

// WithUnions enables or disables union generation.
func WithUnions(enabled bool) Option {
	return func(o *GoOptions) {
		o.GenerateUnions = enabled
	}
}

// WithUnionStyle sets the union style.
func WithUnionStyle(style string) Option {
	return func(o *GoOptions) {
		o.UnionStyle = style
	}
}

// NewGoGenerator creates a new Go generator with the given options.
func NewGoGenerator(options ...Option) export.Generator {
	opts := DefaultGoOptions()

	// Apply functional options
	for _, option := range options {
		option(&opts)
	}

	return NewGenerator(opts)
}

// NewGoGeneratorWithOptions creates a new Go generator with explicit options.
func NewGoGeneratorWithOptions(options GoOptions) export.Generator {
	return NewGenerator(options)
}

// Preset configurations

// BasicGoGenerator creates a basic Go generator with minimal configuration.
func BasicGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("struct"),
		WithPackageName("main"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(true),
		WithJSONTags(true),
		WithJSONTagStyle("snake_case"),
		WithIndentStyle("tabs"),
		WithImports(true),
		WithImportStyle("goimports"),
	)
}

// MinimalGoGenerator creates a minimal Go generator with no extra features.
func MinimalGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("struct"),
		WithPackageName("main"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(false),
		WithJSONTags(false),
		WithValidationTags(false),
		WithXMLTags(false),
		WithYAMLTags(false),
		WithPointers(false),
		WithIndentStyle("tabs"),
		WithImports(false),
		WithConstructors(false),
		WithValidators(false),
		WithStringers(false),
		WithGetters(false),
		WithSetters(false),
		WithEnums(false),
		WithUnions(false),
	)
}

// FullFeaturedGoGenerator creates a Go generator with all features enabled.
func FullFeaturedGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("struct"),
		WithPackageName("main"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(true),
		WithExamples(true),
		WithJSONTags(true),
		WithJSONTagStyle("snake_case"),
		WithValidationTags(true),
		WithValidationTagStyle("go-playground"),
		WithXMLTags(true),
		WithYAMLTags(true),
		WithPointers(true),
		WithOmitEmpty(true),
		WithIndentStyle("tabs"),
		WithImports(true),
		WithImportStyle("goimports"),
		WithInterfaces(true),
		WithInterfaceSuffix("Interface"),
		WithConstructors(true),
		WithConstructorPrefix("New"),
		WithValidators(true),
		WithValidatorPrefix("Validate"),
		WithStringers(true),
		WithGetters(true),
		WithSetters(true),
		WithGenerics(true),
		WithGoVersion("1.21"),
		WithEnums(true),
		WithEnumStyle("const"),
		WithUnions(true),
		WithUnionStyle("interface"),
	)
}

// APIGoGenerator creates a Go generator optimized for API models.
func APIGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("struct"),
		WithPackageName("models"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(true),
		WithExamples(false),
		WithJSONTags(true),
		WithJSONTagStyle("snake_case"),
		WithValidationTags(true),
		WithValidationTagStyle("go-playground"),
		WithXMLTags(false),
		WithYAMLTags(false),
		WithPointers(true),
		WithOmitEmpty(true),
		WithIndentStyle("tabs"),
		WithImports(true),
		WithImportStyle("goimports"),
		WithInterfaces(false),
		WithConstructors(true),
		WithConstructorPrefix("New"),
		WithValidators(true),
		WithValidatorPrefix("Validate"),
		WithStringers(false),
		WithGetters(false),
		WithSetters(false),
		WithGenerics(false),
		WithGoVersion("1.21"),
		WithEnums(true),
		WithEnumStyle("const"),
		WithUnions(false),
	)
}

// ConfigGoGenerator creates a Go generator optimized for configuration structs.
func ConfigGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("struct"),
		WithPackageName("config"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(true),
		WithExamples(true),
		WithJSONTags(true),
		WithJSONTagStyle("snake_case"),
		WithValidationTags(false),
		WithXMLTags(false),
		WithYAMLTags(true),
		WithPointers(true),
		WithOmitEmpty(true),
		WithIndentStyle("tabs"),
		WithImports(true),
		WithImportStyle("goimports"),
		WithInterfaces(false),
		WithConstructors(false),
		WithValidators(false),
		WithStringers(false),
		WithGetters(false),
		WithSetters(false),
		WithGenerics(false),
		WithGoVersion("1.21"),
		WithEnums(true),
		WithEnumStyle("const"),
		WithUnions(false),
	)
}

// InterfaceGoGenerator creates a Go generator that generates interfaces.
func InterfaceGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("interface"),
		WithPackageName("interfaces"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(true),
		WithExamples(false),
		WithJSONTags(false),
		WithValidationTags(false),
		WithXMLTags(false),
		WithYAMLTags(false),
		WithPointers(false),
		WithIndentStyle("tabs"),
		WithImports(true),
		WithImportStyle("goimports"),
		WithInterfaces(true),
		WithInterfaceSuffix(""),
		WithConstructors(false),
		WithValidators(false),
		WithStringers(false),
		WithGetters(false),
		WithSetters(false),
		WithGenerics(true),
		WithGoVersion("1.21"),
		WithEnums(false),
		WithUnions(true),
		WithUnionStyle("interface"),
	)
}

// TypeAliasGoGenerator creates a Go generator that generates type aliases.
func TypeAliasGoGenerator() export.Generator {
	return NewGoGenerator(
		WithOutputStyle("type_alias"),
		WithPackageName("types"),
		WithNamingConvention("PascalCase"),
		WithFieldNamingConvention("PascalCase"),
		WithComments(true),
		WithExamples(false),
		WithJSONTags(false),
		WithValidationTags(false),
		WithXMLTags(false),
		WithYAMLTags(false),
		WithPointers(false),
		WithIndentStyle("tabs"),
		WithImports(true),
		WithImportStyle("goimports"),
		WithInterfaces(false),
		WithConstructors(false),
		WithValidators(false),
		WithStringers(false),
		WithGetters(false),
		WithSetters(false),
		WithGenerics(true),
		WithGoVersion("1.21"),
		WithEnums(true),
		WithEnumStyle("type"),
		WithUnions(false),
	)
}

// Helper functions for creating generators with specific configurations

// CreateGoGenerator creates a Go generator with a map of options.
func CreateGoGenerator(options map[string]any) (export.Generator, error) {
	opts := DefaultGoOptions()

	// Apply options from map
	for key, value := range options {
		opts.SetOption(key, value)
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return NewGenerator(opts), nil
}

// CreateGoGeneratorFromPreset creates a Go generator from a preset name.
func CreateGoGeneratorFromPreset(preset string) (export.Generator, error) {
	switch preset {
	case "basic":
		return BasicGoGenerator(), nil
	case "minimal":
		return MinimalGoGenerator(), nil
	case "full":
		return FullFeaturedGoGenerator(), nil
	case "api":
		return APIGoGenerator(), nil
	case "config":
		return ConfigGoGenerator(), nil
	case "interface":
		return InterfaceGoGenerator(), nil
	case "type_alias":
		return TypeAliasGoGenerator(), nil
	default:
		return nil, &OptionsError{
			Field:   "preset",
			Value:   preset,
			Message: "unknown preset",
			Valid:   []string{"basic", "minimal", "full", "api", "config", "interface", "type_alias"},
		}
	}
}

// GetAvailablePresets returns a list of available preset names.
func GetAvailablePresets() []string {
	return []string{
		"basic",
		"minimal",
		"full",
		"api",
		"config",
		"interface",
		"type_alias",
	}
}

// GetPresetDescription returns a description of a preset.
func GetPresetDescription(preset string) string {
	descriptions := map[string]string{
		"basic":      "Basic Go generator with minimal configuration",
		"minimal":    "Minimal Go generator with no extra features",
		"full":       "Full-featured Go generator with all options enabled",
		"api":        "Go generator optimized for API models with validation",
		"config":     "Go generator optimized for configuration structs with YAML support",
		"interface":  "Go generator that creates interfaces instead of structs",
		"type_alias": "Go generator that creates type aliases instead of structs",
	}

	if desc, exists := descriptions[preset]; exists {
		return desc
	}

	return "Unknown preset"
}

// Factory function for integration with the export system

// GoGeneratorFactory is a factory function that creates Go generators.
func GoGeneratorFactory(options ...any) (export.Generator, error) {
	// Handle different option types
	if len(options) == 0 {
		return BasicGoGenerator(), nil
	}

	// If first option is a string, treat it as a preset
	if preset, ok := options[0].(string); ok {
		return CreateGoGeneratorFromPreset(preset)
	}

	// If first option is a map, use it as options
	if optMap, ok := options[0].(map[string]any); ok {
		return CreateGoGenerator(optMap)
	}

	// If first option is GoOptions, use it directly
	if goOpts, ok := options[0].(GoOptions); ok {
		return NewGenerator(goOpts), nil
	}

	// Default to basic generator
	return BasicGoGenerator(), nil
}
