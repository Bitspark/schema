package export

import (
	"fmt"

	"defs.dev/schema/visit/export/base"
)

// GeneratorFactoryFunc is a function that creates a generator with the given options.
type GeneratorFactoryFunc func(options ...Option) (Generator, error)

// FactoryRegistry maintains a registry of generator factories.
type FactoryRegistry struct {
	factories map[string]GeneratorFactoryFunc
}

// NewFactoryRegistry creates a new FactoryRegistry.
func NewFactoryRegistry() *FactoryRegistry {
	return &FactoryRegistry{
		factories: make(map[string]GeneratorFactoryFunc),
	}
}

// Register registers a generator factory with the given name.
func (r *FactoryRegistry) Register(name string, factory GeneratorFactoryFunc) error {
	if name == "" {
		return fmt.Errorf("generator name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("generator factory cannot be nil")
	}

	r.factories[name] = factory
	return nil
}

// Unregister removes a generator factory from the registry.
func (r *FactoryRegistry) Unregister(name string) bool {
	if _, exists := r.factories[name]; exists {
		delete(r.factories, name)
		return true
	}
	return false
}

// Create creates a generator using the factory registered with the given name.
func (r *FactoryRegistry) Create(name string, options ...Option) (Generator, error) {
	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("no factory registered for generator: %s", name)
	}

	return factory(options...)
}

// List returns all registered generator names.
func (r *FactoryRegistry) List() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// HasFactory returns true if a factory is registered for the given name.
func (r *FactoryRegistry) HasFactory(name string) bool {
	_, exists := r.factories[name]
	return exists
}

// DefaultFactoryRegistry is the global factory registry.
var DefaultFactoryRegistry = NewFactoryRegistry()

// Convenience functions for the default factory registry

// RegisterGenerator registers a generator factory in the default registry.
func RegisterGenerator(name string, factory GeneratorFactoryFunc) error {
	return DefaultFactoryRegistry.Register(name, factory)
}

// UnregisterGenerator removes a generator factory from the default registry.
func UnregisterGenerator(name string) bool {
	return DefaultFactoryRegistry.Unregister(name)
}

// CreateGenerator creates a generator using the default factory registry.
func CreateGenerator(name string, options ...Option) (Generator, error) {
	return DefaultFactoryRegistry.Create(name, options...)
}

// ListGenerators returns all registered generator names from the default registry.
func ListGenerators() []string {
	return DefaultFactoryRegistry.List()
}

// HasGenerator returns true if a factory is registered for the given name in the default registry.
func HasGenerator(name string) bool {
	return DefaultFactoryRegistry.HasFactory(name)
}

// Option types and implementations

// OptionFunc is a functional option for configuring generators.
type OptionFunc func(any) error

// Apply applies the option to the given target.
func (f OptionFunc) Apply(target any) error {
	return f(target)
}

// Ensure OptionFunc implements Option interface
var _ Option = OptionFunc(nil)

// String option helpers

// WithStringOption creates an option that sets a string value.
func WithStringOption(key, value string) Option {
	return OptionFunc(func(target any) error {
		if setter, ok := target.(interface{ SetOption(string, any) }); ok {
			setter.SetOption(key, value)
			return nil
		}
		return fmt.Errorf("target does not support string options")
	})
}

// WithBoolOption creates an option that sets a boolean value.
func WithBoolOption(key string, value bool) Option {
	return OptionFunc(func(target any) error {
		if setter, ok := target.(interface{ SetOption(string, any) }); ok {
			setter.SetOption(key, value)
			return nil
		}
		return fmt.Errorf("target does not support boolean options")
	})
}

// WithIntOption creates an option that sets an integer value.
func WithIntOption(key string, value int) Option {
	return OptionFunc(func(target any) error {
		if setter, ok := target.(interface{ SetOption(string, any) }); ok {
			setter.SetOption(key, value)
			return nil
		}
		return fmt.Errorf("target does not support integer options")
	})
}

// WithOption creates a generic option that sets an arbitrary value.
func WithOption(key string, value any) Option {
	return OptionFunc(func(target any) error {
		if setter, ok := target.(interface{ SetOption(string, any) }); ok {
			setter.SetOption(key, value)
			return nil
		}
		return fmt.Errorf("target does not support options")
	})
}

// Common option keys (these can be used by generators)
const (
	// General options
	OptionIndentSize   = "indent_size"
	OptionIndentString = "indent_string"
	OptionMaxWidth     = "max_width"
	OptionComments     = "include_comments"
	OptionExamples     = "include_examples"
	OptionMetadata     = "include_metadata"

	// Naming conventions
	OptionNamingStyle = "naming_style"
	OptionCaseStyle   = "case_style"

	// Output format options
	OptionPrettyPrint = "pretty_print"
	OptionMinify      = "minify"
	OptionValidate    = "validate_output"

	// Language-specific options
	OptionLanguageVersion = "language_version"
	OptionStrictMode      = "strict_mode"
	OptionExportStyle     = "export_style"
)

// Common option values
const (
	// Naming styles
	NamingCamelCase  = "camelCase"
	NamingPascalCase = "PascalCase"
	NamingSnakeCase  = "snake_case"
	NamingKebabCase  = "kebab-case"

	// Export styles
	ExportNamed   = "named"
	ExportDefault = "default"
	ExportBoth    = "both"
	ExportNone    = "none"
)

// Common option creators

// WithIndentSize sets the indentation size (number of spaces or tabs).
func WithIndentSize(size int) Option {
	return WithIntOption(OptionIndentSize, size)
}

// WithIndentString sets the indentation string (e.g., "  ", "\t").
func WithIndentString(indent string) Option {
	return WithStringOption(OptionIndentString, indent)
}

// WithMaxWidth sets the maximum line width for formatting.
func WithMaxWidth(width int) Option {
	return WithIntOption(OptionMaxWidth, width)
}

// WithComments enables or disables comment generation.
func WithComments(enabled bool) Option {
	return WithBoolOption(OptionComments, enabled)
}

// WithExamples enables or disables example generation.
func WithExamples(enabled bool) Option {
	return WithBoolOption(OptionExamples, enabled)
}

// WithMetadata enables or disables metadata inclusion.
func WithMetadata(enabled bool) Option {
	return WithBoolOption(OptionMetadata, enabled)
}

// WithNamingStyle sets the naming convention style.
func WithNamingStyle(style string) Option {
	return WithStringOption(OptionNamingStyle, style)
}

// WithPrettyPrint enables or disables pretty printing.
func WithPrettyPrint(enabled bool) Option {
	return WithBoolOption(OptionPrettyPrint, enabled)
}

// WithValidation enables or disables output validation.
func WithValidation(enabled bool) Option {
	return WithBoolOption(OptionValidate, enabled)
}

// WithStrictMode enables or disables strict mode.
func WithStrictMode(enabled bool) Option {
	return WithBoolOption(OptionStrictMode, enabled)
}

// Builder pattern for complex generator configurations

// GeneratorBuilder provides a fluent interface for building generators.
type GeneratorBuilder struct {
	generatorType string
	options       []Option
	errors        *base.ErrorCollector
}

// NewGeneratorBuilder creates a new GeneratorBuilder for the specified generator type.
func NewGeneratorBuilder(generatorType string) *GeneratorBuilder {
	return &GeneratorBuilder{
		generatorType: generatorType,
		options:       make([]Option, 0),
		errors:        base.NewErrorCollector(),
	}
}

// WithOption adds an option to the builder.
func (b *GeneratorBuilder) WithOption(option Option) *GeneratorBuilder {
	b.options = append(b.options, option)
	return b
}

// WithStringOption adds a string option to the builder.
func (b *GeneratorBuilder) WithStringOption(key, value string) *GeneratorBuilder {
	return b.WithOption(WithStringOption(key, value))
}

// WithBoolOption adds a boolean option to the builder.
func (b *GeneratorBuilder) WithBoolOption(key string, value bool) *GeneratorBuilder {
	return b.WithOption(WithBoolOption(key, value))
}

// WithIntOption adds an integer option to the builder.
func (b *GeneratorBuilder) WithIntOption(key string, value int) *GeneratorBuilder {
	return b.WithOption(WithIntOption(key, value))
}

// IndentSize sets the indentation size.
func (b *GeneratorBuilder) IndentSize(size int) *GeneratorBuilder {
	return b.WithOption(WithIndentSize(size))
}

// IndentString sets the indentation string.
func (b *GeneratorBuilder) IndentString(indent string) *GeneratorBuilder {
	return b.WithOption(WithIndentString(indent))
}

// MaxWidth sets the maximum line width.
func (b *GeneratorBuilder) MaxWidth(width int) *GeneratorBuilder {
	return b.WithOption(WithMaxWidth(width))
}

// Comments enables or disables comment generation.
func (b *GeneratorBuilder) Comments(enabled bool) *GeneratorBuilder {
	return b.WithOption(WithComments(enabled))
}

// Examples enables or disables example generation.
func (b *GeneratorBuilder) Examples(enabled bool) *GeneratorBuilder {
	return b.WithOption(WithExamples(enabled))
}

// Metadata enables or disables metadata inclusion.
func (b *GeneratorBuilder) Metadata(enabled bool) *GeneratorBuilder {
	return b.WithOption(WithMetadata(enabled))
}

// NamingStyle sets the naming convention style.
func (b *GeneratorBuilder) NamingStyle(style string) *GeneratorBuilder {
	return b.WithOption(WithNamingStyle(style))
}

// PrettyPrint enables or disables pretty printing.
func (b *GeneratorBuilder) PrettyPrint(enabled bool) *GeneratorBuilder {
	return b.WithOption(WithPrettyPrint(enabled))
}

// Validation enables or disables output validation.
func (b *GeneratorBuilder) Validation(enabled bool) *GeneratorBuilder {
	return b.WithOption(WithValidation(enabled))
}

// StrictMode enables or disables strict mode.
func (b *GeneratorBuilder) StrictMode(enabled bool) *GeneratorBuilder {
	return b.WithOption(WithStrictMode(enabled))
}

// Build creates the generator using the configured options.
func (b *GeneratorBuilder) Build() (Generator, error) {
	if b.errors.HasErrors() {
		return nil, b.errors.Error()
	}

	return CreateGenerator(b.generatorType, b.options...)
}

// MustBuild creates the generator and panics if there are any errors.
func (b *GeneratorBuilder) MustBuild() Generator {
	generator, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build generator: %v", err))
	}
	return generator
}

// Validation helpers

// ValidateOptions validates that all required options are present and valid.
func ValidateOptions(options []Option, required []string) error {
	// This is a placeholder implementation
	// In a real implementation, you would validate specific option requirements
	if len(required) == 0 {
		return nil
	}

	// For now, just return nil - individual generators can implement their own validation
	return nil
}

// ApplyOptions applies a list of options to a target object.
func ApplyOptions(target any, options []Option) error {
	collector := base.NewErrorCollector()

	for _, option := range options {
		if err := option.Apply(target); err != nil {
			collector.Add(err)
		}
	}

	return collector.Error()
}
