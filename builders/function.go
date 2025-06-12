package builders

import (
	"context"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

// FunctionSchemaBuilder implements core.FunctionSchemaBuilder for creating function schemas.
type FunctionSchemaBuilder struct {
	inputs        schemas.ArgSchemas
	outputs       schemas.ArgSchemas
	errors        core.Schema
	examples      []map[string]any
	allowNilError bool
	metadata      core.SchemaMetadata
}

// Ensure FunctionBuilder implements the API interface at compile time
var _ core.FunctionSchemaBuilder = (*FunctionSchemaBuilder)(nil)
var _ core.Builder[core.FunctionSchema] = (*FunctionSchemaBuilder)(nil)
var _ core.MetadataBuilder[core.FunctionSchemaBuilder] = (*FunctionSchemaBuilder)(nil)

// NewFunctionSchema creates a new FunctionBuilder.
func NewFunctionSchema() *FunctionSchemaBuilder {
	return &FunctionSchemaBuilder{
		inputs:   schemas.NewArgSchemas(),
		outputs:  schemas.NewArgSchemas(),
		examples: []map[string]any{},
		metadata: core.SchemaMetadata{},
	}
}

// Core builder methods (API compliance)

func (b *FunctionSchemaBuilder) Input(name string, schema core.Schema) core.FunctionSchemaBuilder {
	arg := schemas.NewArgSchema(name, schema)
	b.inputs.AddArg(arg)
	return b
}

func (b *FunctionSchemaBuilder) Output(name string, schema core.Schema) core.FunctionSchemaBuilder {
	arg := schemas.NewArgSchema(name, schema)
	b.outputs.AddArg(arg)
	return b
}

func (b *FunctionSchemaBuilder) Error(schema core.Schema) core.FunctionSchemaBuilder {
	b.errors = schema
	return b
}

func (b *FunctionSchemaBuilder) RequiredInputs(names ...string) core.FunctionSchemaBuilder {
	// Mark specified inputs as required
	for _, name := range names {
		b.inputs.SetOptionalByName(name, false)
	}
	return b
}

func (b *FunctionSchemaBuilder) RequiredOutputs(names ...string) core.FunctionSchemaBuilder {
	// Mark specified outputs as required
	for _, name := range names {
		b.outputs.SetOptionalByName(name, false)
	}
	return b
}

func (b *FunctionSchemaBuilder) Example(example map[string]any) core.FunctionSchemaBuilder {
	b.examples = append(b.examples, example)
	return b
}

// Metadata builder methods (API compliance)

func (b *FunctionSchemaBuilder) Description(desc string) core.FunctionSchemaBuilder {
	b.metadata.Description = desc
	return b
}

func (b *FunctionSchemaBuilder) Name(name string) core.FunctionSchemaBuilder {
	b.metadata.Name = name
	return b
}

func (b *FunctionSchemaBuilder) Tag(tag string) core.FunctionSchemaBuilder {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

// Extended builder methods (beyond API requirements)

// RequiredInput adds a required input parameter (convenience method)
func (b *FunctionSchemaBuilder) RequiredInput(name string, schema core.Schema) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, nil)
	b.inputs.AddArg(arg)
	return b
}

// OptionalInput adds an optional input parameter (convenience method)
func (b *FunctionSchemaBuilder) OptionalInput(name string, schema core.Schema) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", true, nil)
	b.inputs.AddArg(arg)
	return b
}

// RequiredOutput adds a required output parameter (convenience method)
func (b *FunctionSchemaBuilder) RequiredOutput(name string, schema core.Schema) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, nil)
	b.outputs.AddArg(arg)
	return b
}

// OptionalOutput adds an optional output parameter (convenience method)
func (b *FunctionSchemaBuilder) OptionalOutput(name string, schema core.Schema) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", true, nil)
	b.outputs.AddArg(arg)
	return b
}

// InputWithConstraints adds an input with constraints
func (b *FunctionSchemaBuilder) InputWithConstraints(name string, schema core.Schema, constraints ...string) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, constraints)
	b.inputs.AddArg(arg)
	return b
}

// OutputWithConstraints adds an output with constraints
func (b *FunctionSchemaBuilder) OutputWithConstraints(name string, schema core.Schema, constraints ...string) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, constraints)
	b.outputs.AddArg(arg)
	return b
}

// InputWithDescription adds an input with description
func (b *FunctionSchemaBuilder) InputWithDescription(name string, schema core.Schema, description string) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, description, false, nil)
	b.inputs.AddArg(arg)
	return b
}

// OutputWithDescription adds an output with description
func (b *FunctionSchemaBuilder) OutputWithDescription(name string, schema core.Schema, description string) *FunctionSchemaBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, description, false, nil)
	b.outputs.AddArg(arg)
	return b
}

// AdditionalInputs controls whether additional inputs are allowed
func (b *FunctionSchemaBuilder) AdditionalInputs(allowed bool) *FunctionSchemaBuilder {
	b.inputs.SetAllowAdditional(allowed)
	return b
}

// AdditionalOutputs controls whether additional outputs are allowed
func (b *FunctionSchemaBuilder) AdditionalOutputs(allowed bool) *FunctionSchemaBuilder {
	b.outputs.SetAllowAdditional(allowed)
	return b
}

// Strict disallows additional inputs and outputs (convenience method)
func (b *FunctionSchemaBuilder) Strict() *FunctionSchemaBuilder {
	b.inputs.SetAllowAdditional(false)
	b.outputs.SetAllowAdditional(false)
	return b
}

// Flexible allows additional inputs and outputs (convenience method)
func (b *FunctionSchemaBuilder) Flexible() *FunctionSchemaBuilder {
	b.inputs.SetAllowAdditional(true)
	b.outputs.SetAllowAdditional(true)
	return b
}

// AllowNilError controls whether nil error is allowed
func (b *FunctionSchemaBuilder) AllowNilError(allowed bool) *FunctionSchemaBuilder {
	b.allowNilError = allowed
	return b
}

// NonEmptyInput adds a non-empty constraint to an input parameter
func (b *FunctionSchemaBuilder) NonEmptyInput(inputName string) *FunctionSchemaBuilder {
	b.inputs.AddConstraintByName(inputName, "non_empty")
	return b
}

// Domain-specific function examples

// APIEndpoint creates a function schema for an API endpoint
func (b *FunctionSchemaBuilder) APIEndpoint() *FunctionSchemaBuilder {
	b.Name("API Endpoint")
	b.Description("HTTP API endpoint function")
	b.Tag("api")
	b.Tag("endpoint")
	return b.Strict()
}

// DatabaseQuery creates a function schema for database queries
func (b *FunctionSchemaBuilder) DatabaseQuery() *FunctionSchemaBuilder {
	b.Name("Database Query")
	b.Description("Database query function")
	b.Tag("database")
	b.Tag("query")
	return b.Strict()
}

// DataProcessor creates a function schema for data processing
func (b *FunctionSchemaBuilder) DataProcessor() *FunctionSchemaBuilder {
	b.Name("Data Processor")
	b.Description("Data processing function")
	b.Tag("processing")
	b.Tag("transform")
	return b.Flexible()
}

// Validator creates a function schema for validation functions
func (b *FunctionSchemaBuilder) Validator() *FunctionSchemaBuilder {
	b.Name("Validator")
	b.Description("Data validation function")
	b.Tag("validation")
	b.Tag("check")
	return b.Strict()
}

// ServiceCall creates a function schema for service calls
func (b *FunctionSchemaBuilder) ServiceCall() *FunctionSchemaBuilder {
	b.Name("Service Call")
	b.Description("External service call function")
	b.Tag("service")
	b.Tag("external")
	return b.Strict()
}

// EventHandler creates a function schema for event handlers
func (b *FunctionSchemaBuilder) EventHandler() *FunctionSchemaBuilder {
	b.Name("Event Handler")
	b.Description("Event handling function")
	b.Tag("event")
	b.Tag("handler")
	return b.Flexible()
}

// Common function patterns

// SimpleFunction creates a basic function with input and output
func (b *FunctionSchemaBuilder) SimpleFunction(inputName string, inputSchema core.Schema, outputName string, outputSchema core.Schema) *FunctionSchemaBuilder {
	b.RequiredInput(inputName, inputSchema)
	b.RequiredOutput(outputName, outputSchema)
	return b.Strict()
}

// NoOutputFunction creates a function that doesn't return a value
func (b *FunctionSchemaBuilder) NoOutputFunction() *FunctionSchemaBuilder {
	return b.Strict()
}

// ErrorReturningFunction creates a function that can return errors
func (b *FunctionSchemaBuilder) ErrorReturningFunction(errorSchema core.Schema) *FunctionSchemaBuilder {
	b.Error(errorSchema)
	return b.AllowNilError(false)
}

// VoidFunction creates a function with no inputs or outputs
func (b *FunctionSchemaBuilder) VoidFunction() *FunctionSchemaBuilder {
	b.AllowNilError(true)
	return b.Strict()
}

// Build creates the final FunctionSchema
func (b *FunctionSchemaBuilder) Build() core.FunctionSchema {
	schema := schemas.NewFunctionSchema(b.inputs, b.outputs)

	// Apply all builder configurations
	if b.errors != nil {
		schema = schema.WithError(b.errors)
	}

	for _, example := range b.examples {
		schema = schema.WithExample(example)
	}

	return schema.WithMetadata(b.metadata)
}

// Helper methods for creating common examples

// LoginExample creates an example for a login function
func (b *FunctionSchemaBuilder) LoginExample() *FunctionSchemaBuilder {
	example := map[string]any{
		"username": "user@example.com",
		"password": "securePassword123",
	}
	b.Example(example)
	return b
}

// SearchExample creates an example for a search function
func (b *FunctionSchemaBuilder) SearchExample() *FunctionSchemaBuilder {
	example := map[string]any{
		"query":  "search terms",
		"limit":  10,
		"offset": 0,
	}
	b.Example(example)
	return b
}

// CreateUserExample creates an example for a user creation function
func (b *FunctionSchemaBuilder) CreateUserExample() *FunctionSchemaBuilder {
	example := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	b.Example(example)
	return b
}

// CalculationExample creates an example for mathematical calculations
func (b *FunctionSchemaBuilder) CalculationExample() *FunctionSchemaBuilder {
	example := map[string]any{
		"a":        10,
		"b":        5,
		"operator": "add",
	}
	b.Example(example)
	return b
}

// FileProcessingExample creates an example for file processing
func (b *FunctionSchemaBuilder) FileProcessingExample() *FunctionSchemaBuilder {
	example := map[string]any{
		"filename": "document.pdf",
		"format":   "text",
		"options":  map[string]any{"encoding": "utf-8"},
	}
	b.Example(example)
	return b
}

type FunctionBuilder struct {
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func NewFunctionBuilder() *FunctionBuilder {
	return &FunctionBuilder{}
}

func (b *FunctionBuilder) Build() api.Function {
	f := &FunctionImpl{
		schema:  b.schema,
		handler: b.handler,
	}
	return f
}

type FunctionImpl struct {
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (f *FunctionImpl) Schema() core.FunctionSchema {
	return f.schema
}

func (f *FunctionImpl) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler(ctx, params)
}

func (f *FunctionImpl) Name() string {
	return f.schema.Metadata().Name
}

func (f *FunctionImpl) Handler() func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler
}
