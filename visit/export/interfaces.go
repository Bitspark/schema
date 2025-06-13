package export

import "defs.dev/schema/core"

// Generator is the main interface that all export generators must implement.
// It combines schema visitation with output generation capabilities.
type Generator interface {
	core.SchemaVisitor

	// Generate produces output from a schema by accepting it as a visitor
	Generate(schema core.Schema) ([]byte, error)

	// Name returns the human-readable name of the generator
	Name() string

	// Format returns the output format identifier (e.g., "json-schema", "typescript")
	Format() string
}

// GeneratorWithOptions supports configuration through functional options.
// This enables generators to be customized without breaking the base interface.
type GeneratorWithOptions[T any] interface {
	Generator
	WithOptions(options T) Generator
}

// MultiFileGenerator generates multiple files from a single schema.
// This is useful for generators that need to split output across multiple files
// (e.g., TypeScript with separate interface and implementation files).
type MultiFileGenerator interface {
	Generator

	// GenerateFiles returns a map of filename to file content
	GenerateFiles(schema core.Schema) (map[string][]byte, error)
}

// StreamingGenerator supports streaming generation for large schemas.
// This enables memory-efficient generation of large outputs.
type StreamingGenerator interface {
	Generator

	// GenerateStream writes output directly to a writer instead of returning bytes
	GenerateStream(schema core.Schema, writer any) error
}

// GeneratorFactory creates generators with specific configurations.
// This enables dynamic generator creation and registration.
type GeneratorFactory func(options ...any) (Generator, error)

// GeneratorRegistry manages multiple generators and enables batch generation.
type GeneratorRegistry interface {
	// Register adds a generator with a given name
	Register(name string, generator Generator) error

	// RegisterFactory adds a generator factory with a given name
	RegisterFactory(name string, factory GeneratorFactory) error

	// Get retrieves a generator by name
	Get(name string) (Generator, bool)

	// List returns all registered generator names
	List() []string

	// Generate uses a specific generator to produce output
	Generate(generatorName string, schema core.Schema) ([]byte, error)

	// GenerateAll produces output using all registered generators
	GenerateAll(schema core.Schema) (map[string][]byte, error)

	// Remove unregisters a generator
	Remove(name string) bool
}

// GenerationResult contains the output and metadata from generation.
type GenerationResult struct {
	// Output is the generated content
	Output []byte

	// Format is the output format
	Format string

	// Metadata contains additional information about the generation
	Metadata map[string]any

	// Warnings contains any non-fatal issues encountered during generation
	Warnings []string
}

// BatchGenerationResult contains results from multiple generators.
type BatchGenerationResult struct {
	// Results maps generator name to generation result
	Results map[string]*GenerationResult

	// Errors maps generator name to any errors that occurred
	Errors map[string]error

	// Summary contains overall statistics
	Summary *GenerationSummary
}

// GenerationSummary provides statistics about batch generation.
type GenerationSummary struct {
	// TotalGenerators is the number of generators that were run
	TotalGenerators int

	// SuccessfulGenerators is the number that completed without error
	SuccessfulGenerators int

	// FailedGenerators is the number that encountered errors
	FailedGenerators int

	// TotalWarnings is the total number of warnings across all generators
	TotalWarnings int
}

// Option represents a functional option for configuring generators.
type Option interface {
	Apply(target any) error
}

// Validator can validate generated output before returning it.
// This enables generators to perform self-validation.
type Validator interface {
	// Validate checks if the generated output is valid
	Validate(output []byte, format string) error
}

// Transformer can modify generated output before returning it.
// This enables post-processing like formatting, minification, etc.
type Transformer interface {
	// Transform modifies the generated output
	Transform(output []byte, format string) ([]byte, error)
}

// Plugin represents an extensible generator plugin.
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// CreateGenerator creates a new generator instance
	CreateGenerator(options ...Option) (Generator, error)

	// SupportedFormats returns the formats this plugin can generate
	SupportedFormats() []string
}
