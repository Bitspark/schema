package builders

import (
	"defs.dev/schema/api"
	"defs.dev/schema/core/schemas"
)

// FunctionBuilder implements api.FunctionSchemaBuilder for creating function schemas.
type FunctionBuilder struct {
	inputs          schemas.ArgSchemas
	outputs         schemas.ArgSchemas
	errors          api.Schema
	examples        []map[string]any
	allowNilError   bool
	validationRules []schemas.FunctionValidationRule
	metadata        api.SchemaMetadata
}

// Ensure FunctionBuilder implements the API interface at compile time
var _ api.FunctionSchemaBuilder = (*FunctionBuilder)(nil)
var _ api.Builder[api.FunctionSchema] = (*FunctionBuilder)(nil)
var _ api.MetadataBuilder[api.FunctionSchemaBuilder] = (*FunctionBuilder)(nil)

// NewFunctionBuilder creates a new FunctionBuilder.
func NewFunctionBuilder() *FunctionBuilder {
	return &FunctionBuilder{
		inputs:          schemas.NewArgSchemas(),
		outputs:         schemas.NewArgSchemas(),
		examples:        []map[string]any{},
		validationRules: []schemas.FunctionValidationRule{},
		metadata:        api.SchemaMetadata{},
	}
}

// Core builder methods (API compliance)

func (b *FunctionBuilder) Input(name string, schema api.Schema) api.FunctionSchemaBuilder {
	arg := schemas.NewArgSchema(name, schema)
	b.inputs.AddArg(arg)
	return b
}

func (b *FunctionBuilder) Output(name string, schema api.Schema) api.FunctionSchemaBuilder {
	arg := schemas.NewArgSchema(name, schema)
	b.outputs.AddArg(arg)
	return b
}

func (b *FunctionBuilder) Error(schema api.Schema) api.FunctionSchemaBuilder {
	b.errors = schema
	return b
}

func (b *FunctionBuilder) RequiredInputs(names ...string) api.FunctionSchemaBuilder {
	// Mark specified inputs as required
	for _, name := range names {
		b.inputs.SetOptionalByName(name, false)
	}
	return b
}

func (b *FunctionBuilder) RequiredOutputs(names ...string) api.FunctionSchemaBuilder {
	// Mark specified outputs as required
	for _, name := range names {
		b.outputs.SetOptionalByName(name, false)
	}
	return b
}

func (b *FunctionBuilder) Example(example map[string]any) api.FunctionSchemaBuilder {
	b.examples = append(b.examples, example)
	return b
}

// Metadata builder methods (API compliance)

func (b *FunctionBuilder) Description(desc string) api.FunctionSchemaBuilder {
	b.metadata.Description = desc
	return b
}

func (b *FunctionBuilder) Name(name string) api.FunctionSchemaBuilder {
	b.metadata.Name = name
	return b
}

func (b *FunctionBuilder) Tag(tag string) api.FunctionSchemaBuilder {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

// Extended builder methods (beyond API requirements)

// RequiredInput adds a required input parameter (convenience method)
func (b *FunctionBuilder) RequiredInput(name string, schema api.Schema) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, nil)
	b.inputs.AddArg(arg)
	return b
}

// OptionalInput adds an optional input parameter (convenience method)
func (b *FunctionBuilder) OptionalInput(name string, schema api.Schema) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", true, nil)
	b.inputs.AddArg(arg)
	return b
}

// RequiredOutput adds a required output parameter (convenience method)
func (b *FunctionBuilder) RequiredOutput(name string, schema api.Schema) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, nil)
	b.outputs.AddArg(arg)
	return b
}

// OptionalOutput adds an optional output parameter (convenience method)
func (b *FunctionBuilder) OptionalOutput(name string, schema api.Schema) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", true, nil)
	b.outputs.AddArg(arg)
	return b
}

// InputWithConstraints adds an input with constraints
func (b *FunctionBuilder) InputWithConstraints(name string, schema api.Schema, constraints ...string) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, constraints)
	b.inputs.AddArg(arg)
	return b
}

// OutputWithConstraints adds an output with constraints
func (b *FunctionBuilder) OutputWithConstraints(name string, schema api.Schema, constraints ...string) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, "", false, constraints)
	b.outputs.AddArg(arg)
	return b
}

// InputWithDescription adds an input with description
func (b *FunctionBuilder) InputWithDescription(name string, schema api.Schema, description string) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, description, false, nil)
	b.inputs.AddArg(arg)
	return b
}

// OutputWithDescription adds an output with description
func (b *FunctionBuilder) OutputWithDescription(name string, schema api.Schema, description string) *FunctionBuilder {
	arg := schemas.NewArgSchemaWithOptions(name, schema, description, false, nil)
	b.outputs.AddArg(arg)
	return b
}

// AdditionalInputs controls whether additional inputs are allowed
func (b *FunctionBuilder) AdditionalInputs(allowed bool) *FunctionBuilder {
	b.inputs.SetAllowAdditional(allowed)
	return b
}

// AdditionalOutputs controls whether additional outputs are allowed
func (b *FunctionBuilder) AdditionalOutputs(allowed bool) *FunctionBuilder {
	b.outputs.SetAllowAdditional(allowed)
	return b
}

// Strict disallows additional inputs and outputs (convenience method)
func (b *FunctionBuilder) Strict() *FunctionBuilder {
	b.inputs.SetAllowAdditional(false)
	b.outputs.SetAllowAdditional(false)
	return b
}

// Flexible allows additional inputs and outputs (convenience method)
func (b *FunctionBuilder) Flexible() *FunctionBuilder {
	b.inputs.SetAllowAdditional(true)
	b.outputs.SetAllowAdditional(true)
	return b
}

// AllowNilError controls whether nil error is allowed
func (b *FunctionBuilder) AllowNilError(allowed bool) *FunctionBuilder {
	b.allowNilError = allowed
	return b
}

// NonEmptyInput adds a non-empty constraint to an input parameter
func (b *FunctionBuilder) NonEmptyInput(inputName string) *FunctionBuilder {
	b.inputs.AddConstraintByName(inputName, "non_empty")
	return b
}

// ValidationRule adds a custom validation rule
func (b *FunctionBuilder) ValidationRule(rule schemas.FunctionValidationRule) *FunctionBuilder {
	b.validationRules = append(b.validationRules, rule)
	return b
}

// Domain-specific function examples

// APIEndpoint creates a function schema for an API endpoint
func (b *FunctionBuilder) APIEndpoint() *FunctionBuilder {
	b.Name("API Endpoint")
	b.Description("HTTP API endpoint function")
	b.Tag("api")
	b.Tag("endpoint")
	return b.Strict()
}

// DatabaseQuery creates a function schema for database queries
func (b *FunctionBuilder) DatabaseQuery() *FunctionBuilder {
	b.Name("Database Query")
	b.Description("Database query function")
	b.Tag("database")
	b.Tag("query")
	return b.Strict()
}

// DataProcessor creates a function schema for data processing
func (b *FunctionBuilder) DataProcessor() *FunctionBuilder {
	b.Name("Data Processor")
	b.Description("Data processing function")
	b.Tag("processing")
	b.Tag("transform")
	return b.Flexible()
}

// Validator creates a function schema for validation functions
func (b *FunctionBuilder) Validator() *FunctionBuilder {
	b.Name("Validator")
	b.Description("Data validation function")
	b.Tag("validation")
	b.Tag("check")
	return b.Strict()
}

// ServiceCall creates a function schema for service calls
func (b *FunctionBuilder) ServiceCall() *FunctionBuilder {
	b.Name("Service Call")
	b.Description("External service call function")
	b.Tag("service")
	b.Tag("external")
	return b.Strict()
}

// EventHandler creates a function schema for event handlers
func (b *FunctionBuilder) EventHandler() *FunctionBuilder {
	b.Name("Event Handler")
	b.Description("Event handling function")
	b.Tag("event")
	b.Tag("handler")
	return b.Flexible()
}

// Common function patterns

// SimpleFunction creates a basic function with input and output
func (b *FunctionBuilder) SimpleFunction(inputName string, inputSchema api.Schema, outputName string, outputSchema api.Schema) *FunctionBuilder {
	b.RequiredInput(inputName, inputSchema)
	b.RequiredOutput(outputName, outputSchema)
	return b.Strict()
}

// NoOutputFunction creates a function that doesn't return a value
func (b *FunctionBuilder) NoOutputFunction() *FunctionBuilder {
	return b.Strict()
}

// ErrorReturningFunction creates a function that can return errors
func (b *FunctionBuilder) ErrorReturningFunction(errorSchema api.Schema) *FunctionBuilder {
	b.Error(errorSchema)
	return b.AllowNilError(false)
}

// VoidFunction creates a function with no inputs or outputs
func (b *FunctionBuilder) VoidFunction() *FunctionBuilder {
	b.AllowNilError(true)
	return b.Strict()
}

// Build creates the final FunctionSchema
func (b *FunctionBuilder) Build() api.FunctionSchema {
	schema := schemas.NewFunctionSchema(b.inputs, b.outputs)

	// Apply all builder configurations
	if b.errors != nil {
		schema = schema.WithError(b.errors)
	}

	for _, example := range b.examples {
		schema = schema.WithExample(example)
	}

	for _, rule := range b.validationRules {
		schema = schema.WithValidationRule(rule)
	}

	return schema.WithMetadata(b.metadata)
}

// Helper methods for creating common examples

// LoginExample creates an example for a login function
func (b *FunctionBuilder) LoginExample() *FunctionBuilder {
	example := map[string]any{
		"username": "user@example.com",
		"password": "securePassword123",
	}
	b.Example(example)
	return b
}

// SearchExample creates an example for a search function
func (b *FunctionBuilder) SearchExample() *FunctionBuilder {
	example := map[string]any{
		"query":  "search terms",
		"limit":  10,
		"offset": 0,
	}
	b.Example(example)
	return b
}

// CreateUserExample creates an example for a user creation function
func (b *FunctionBuilder) CreateUserExample() *FunctionBuilder {
	example := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	b.Example(example)
	return b
}

// CalculationExample creates an example for mathematical calculations
func (b *FunctionBuilder) CalculationExample() *FunctionBuilder {
	example := map[string]any{
		"a":        10,
		"b":        5,
		"operator": "add",
	}
	b.Example(example)
	return b
}

// FileProcessingExample creates an example for file processing
func (b *FunctionBuilder) FileProcessingExample() *FunctionBuilder {
	example := map[string]any{
		"filename": "document.pdf",
		"format":   "text",
		"options":  map[string]any{"encoding": "utf-8"},
	}
	b.Example(example)
	return b
}
