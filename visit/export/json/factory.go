package json

import "defs.dev/schema/visit/export"

// Option represents a functional option for JSON Schema generation.
type Option struct {
	key   string
	value any
}

// Apply applies this option to a generator (implements export.Option).
func (o *Option) Apply(target any) error {
	if gen, ok := target.(*Generator); ok {
		gen.options.SetOption(o.key, o.value)
	}
	return nil
}

// Functional option constructors

// WithDraft sets the JSON Schema draft version.
func WithDraft(draft string) export.Option {
	return &Option{key: "draft", value: draft}
}

// WithIncludeExamples sets whether to include examples in the schema.
func WithIncludeExamples(include bool) export.Option {
	return &Option{key: "include_examples", value: include}
}

// WithIncludeDefaults sets whether to include default values in the schema.
func WithIncludeDefaults(include bool) export.Option {
	return &Option{key: "include_defaults", value: include}
}

// WithIncludeDescription sets whether to include descriptions in the schema.
func WithIncludeDescription(include bool) export.Option {
	return &Option{key: "include_description", value: include}
}

// WithStrictMode enables or disables strict validation mode.
func WithStrictMode(strict bool) export.Option {
	return &Option{key: "strict_mode", value: strict}
}

// WithPrettyPrint enables or disables pretty printing of JSON output.
func WithPrettyPrint(pretty bool) export.Option {
	return &Option{key: "pretty_print", value: pretty}
}

// WithIndentSize sets the indentation size for pretty printing.
func WithIndentSize(size int) export.Option {
	return &Option{key: "indent_size", value: size}
}

// WithIncludeFormat sets whether to include format constraints.
func WithIncludeFormat(include bool) export.Option {
	return &Option{key: "include_format", value: include}
}

// WithIncludeTitle sets whether to include titles.
func WithIncludeTitle(include bool) export.Option {
	return &Option{key: "include_title", value: include}
}

// WithIncludeAdditionalProperties sets whether to include additionalProperties.
func WithIncludeAdditionalProperties(include bool) export.Option {
	return &Option{key: "include_additional_properties", value: include}
}

// WithSchemaURI sets the $schema URI.
func WithSchemaURI(uri string) export.Option {
	return &Option{key: "schema_uri", value: uri}
}

// WithRootID sets the root $id.
func WithRootID(id string) export.Option {
	return &Option{key: "root_id", value: id}
}

// WithDefinitionsKey sets the definitions key to use.
func WithDefinitionsKey(key string) export.Option {
	return &Option{key: "definitions_key", value: key}
}

// WithMinifyOutput sets whether to minify the output.
func WithMinifyOutput(minify bool) export.Option {
	return &Option{key: "minify_output", value: minify}
}

// WithIncludeReadOnly sets whether to include readOnly properties.
func WithIncludeReadOnly(include bool) export.Option {
	return &Option{key: "include_readonly", value: include}
}

// WithIncludeWriteOnly sets whether to include writeOnly properties.
func WithIncludeWriteOnly(include bool) export.Option {
	return &Option{key: "include_writeonly", value: include}
}

// WithAllowNullableTypes sets whether to allow nullable types.
func WithAllowNullableTypes(allow bool) export.Option {
	return &Option{key: "allow_nullable_types", value: allow}
}

// Factory function for the export registry

// NewJSONGenerator creates a new JSON Schema generator (factory function).
func NewJSONGenerator(options ...any) (export.Generator, error) {
	// Convert any options to export.Option
	var exportOptions []export.Option
	for _, opt := range options {
		if exportOpt, ok := opt.(export.Option); ok {
			exportOptions = append(exportOptions, exportOpt)
		}
	}
	return NewGenerator(exportOptions...), nil
}

// RegisterJSONGenerator registers the JSON Schema generator with the export system.
func RegisterJSONGenerator(registry export.GeneratorRegistry) error {
	return registry.RegisterFactory("json", NewJSONGenerator)
}
