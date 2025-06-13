package typescript

import (
	"defs.dev/schema/visitors/export"
)

// optionFunc implements the export.Option interface
type optionFunc struct {
	apply func(any) error
}

func (o *optionFunc) Apply(target any) error {
	return o.apply(target)
}

// newOption creates a new option function
func newOption(fn func(*Generator)) export.Option {
	return &optionFunc{
		apply: func(g any) error {
			if tsGen, ok := g.(*Generator); ok {
				fn(tsGen)
			}
			return nil
		},
	}
}

// Factory functions for creating TypeScript generators

// New creates a new TypeScript generator with default options.
func New() export.Generator {
	return NewGenerator()
}

// NewWithOptions creates a new TypeScript generator with custom options.
func NewWithOptions(options TypeScriptOptions) export.Generator {
	return NewGenerator(WithOptions(options))
}

// Functional options for TypeScript generator configuration

// WithOptions sets custom TypeScript options.
func WithOptions(options TypeScriptOptions) export.Option {
	return newOption(func(g *Generator) {
		g.options = options.Clone()
	})
}

// WithOutputStyle sets the TypeScript output style.
func WithOutputStyle(style string) export.Option {
	return newOption(func(g *Generator) {
		g.options.OutputStyle = style
	})
}

// WithNamingConvention sets the naming convention for types and properties.
func WithNamingConvention(convention string) export.Option {
	return newOption(func(g *Generator) {
		g.options.NamingConvention = convention
	})
}

// WithComments enables or disables comment generation.
func WithComments(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeComments = enabled
	})
}

// WithExamples enables or disables example generation in comments.
func WithExamples(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeExamples = enabled
	})
}

// WithDefaults enables or disables default value generation.
func WithDefaults(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeDefaults = enabled
	})
}

// WithStrictMode enables or disables strict TypeScript mode.
func WithStrictMode(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.StrictMode = enabled
	})
}

// WithOptionalProperties enables or disables optional property syntax.
func WithOptionalProperties(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.UseOptionalProperties = enabled
	})
}

// WithIndentSize sets the indentation size.
func WithIndentSize(size int) export.Option {
	return newOption(func(g *Generator) {
		g.options.IndentSize = size
	})
}

// WithTabs enables or disables tab indentation.
func WithTabs(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.UseTabsForIndentation = enabled
	})
}

// WithImports enables or disables import statement generation.
func WithImports(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeImports = enabled
	})
}

// WithExports enables or disables type export.
func WithExports(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.ExportTypes = enabled
	})
}

// WithUnknownType enables or disables 'unknown' type usage.
func WithUnknownType(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.UseUnknownType = enabled
	})
}

// WithValidators enables or disables runtime validator generation.
func WithValidators(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.GenerateValidators = enabled
	})
}

// WithValidatorLibrary sets the validator library to use.
func WithValidatorLibrary(library string) export.Option {
	return newOption(func(g *Generator) {
		g.options.ValidatorLibrary = library
	})
}

// WithEnums enables or disables TypeScript enum generation.
func WithEnums(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.UseEnums = enabled
	})
}

// WithConstAssertions enables or disables const assertions.
func WithConstAssertions(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.UseConstAssertions = enabled
	})
}

// WithUtilityTypes enables or disables utility type generation.
func WithUtilityTypes(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeUtilityTypes = enabled
	})
}

// WithArrayStyle sets the array representation style.
func WithArrayStyle(style string) export.Option {
	return newOption(func(g *Generator) {
		g.options.ArrayStyle = style
	})
}

// WithObjectStyle sets the object representation style.
func WithObjectStyle(style string) export.Option {
	return newOption(func(g *Generator) {
		g.options.ObjectStyle = style
	})
}

// WithPartialTypes enables or disables Partial<T> type generation.
func WithPartialTypes(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.UsePartialTypes = enabled
	})
}

// WithJSDoc enables or disables JSDoc generation.
func WithJSDoc(enabled bool) export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeJSDoc = enabled
	})
}

// WithJSDocStyle sets the JSDoc comment style.
func WithJSDocStyle(style string) export.Option {
	return newOption(func(g *Generator) {
		g.options.JSDocStyle = style
	})
}

// WithFileExtension sets the file extension for generated files.
func WithFileExtension(extension string) export.Option {
	return newOption(func(g *Generator) {
		g.options.FileExtension = extension
	})
}

// WithModuleSystem sets the module system to use.
func WithModuleSystem(system string) export.Option {
	return newOption(func(g *Generator) {
		g.options.ModuleSystem = system
	})
}

// Preset configurations for common use cases

// WithInterfacePreset configures the generator for interface-based output.
func WithInterfacePreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.OutputStyle = "interface"
		g.options.UseOptionalProperties = true
		g.options.ExportTypes = true
		g.options.IncludeJSDoc = true
	})
}

// WithTypePreset configures the generator for type alias-based output.
func WithTypePreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.OutputStyle = "type"
		g.options.UseOptionalProperties = true
		g.options.ExportTypes = true
		g.options.IncludeJSDoc = true
	})
}

// WithStrictPreset configures the generator for strict TypeScript output.
func WithStrictPreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.StrictMode = true
		g.options.UseUnknownType = true
		g.options.UseOptionalProperties = true
		g.options.ExportTypes = true
		g.options.IncludeJSDoc = true
	})
}

// WithMinimalPreset configures the generator for minimal output.
func WithMinimalPreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.IncludeComments = false
		g.options.IncludeExamples = false
		g.options.IncludeDefaults = false
		g.options.IncludeJSDoc = false
		g.options.ExportTypes = false
	})
}

// WithReactPreset configures the generator for React component props.
func WithReactPreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.OutputStyle = "interface"
		g.options.NamingConvention = "PascalCase"
		g.options.UseOptionalProperties = true
		g.options.ExportTypes = true
		g.options.IncludeJSDoc = true
		g.options.StrictMode = true
	})
}

// WithNodePreset configures the generator for Node.js applications.
func WithNodePreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.ModuleSystem = "commonjs"
		g.options.UseUnknownType = true
		g.options.ExportTypes = true
		g.options.IncludeJSDoc = true
	})
}

// WithBrowserPreset configures the generator for browser applications.
func WithBrowserPreset() export.Option {
	return newOption(func(g *Generator) {
		g.options.ModuleSystem = "es6"
		g.options.UseUnknownType = true
		g.options.ExportTypes = true
		g.options.IncludeJSDoc = true
	})
}
