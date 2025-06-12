package schemas

import (
	"fmt"
	"strings"

	"defs.dev/schema/api/core"
)

// FunctionConfig holds the configuration for building a Function.
type FunctionConfig struct {
	Metadata core.SchemaMetadata
}

// ArgSchema represents a named argument with its schema and description.
// This is used for both function inputs and outputs to provide rich metadata.
type ArgSchema struct {
	name        string      `json:"name"`
	description string      `json:"description,omitempty"`
	schema      core.Schema `json:"schema"`
	optional    bool        `json:"optional,omitempty"`
	constraints []string    `json:"constraints,omitempty"`
}

// Ensure ArgSchema implements the API interface at compile time
var _ core.ArgSchema = (*ArgSchema)(nil)

// NewArgSchema creates a new ArgSchema with the given parameters.
func NewArgSchema(name string, schema core.Schema) ArgSchema {
	return ArgSchema{
		name:   name,
		schema: schema,
	}
}

// NewArgSchemaWithOptions creates a new ArgSchema with all options.
func NewArgSchemaWithOptions(name string, schema core.Schema, description string, optional bool, constraints []string) ArgSchema {
	return ArgSchema{
		name:        name,
		description: description,
		schema:      schema,
		optional:    optional,
		constraints: append([]string(nil), constraints...), // copy slice
	}
}

// API interface implementation for ArgSchema
func (a *ArgSchema) Accept(visitor core.SchemaVisitor) error {
	// ArgSchema delegates to its underlying schema
	if accepter, ok := a.schema.(core.Accepter); ok {
		return accepter.Accept(visitor)
	}
	return nil
}

func (a *ArgSchema) Name() string          { return a.name }
func (a *ArgSchema) Description() string   { return a.description }
func (a *ArgSchema) Schema() core.Schema   { return a.schema }
func (a *ArgSchema) Optional() bool        { return a.optional }
func (a *ArgSchema) Constraints() []string { return append([]string(nil), a.constraints...) }

// ArgSchemas represents a collection of named arguments with collection-level metadata.
type ArgSchemas struct {
	args                  []ArgSchema `json:"args"`
	allowAdditional       bool        `json:"allowAdditional,omitempty"`
	additionalSchema      core.Schema `json:"additionalSchema,omitempty"`
	collectionName        string      `json:"collectionName,omitempty"`
	collectionDescription string      `json:"collectionDescription,omitempty"`
}

// Ensure ArgSchemas implements the API interface at compile time
var _ core.ArgSchemas = (*ArgSchemas)(nil)

// NewArgSchemas creates a new ArgSchemas with the given arguments.
func NewArgSchemas() ArgSchemas {
	return ArgSchemas{
		args:            []ArgSchema{},
		allowAdditional: false,
	}
}

// NewArgSchemasWithArgs creates a new ArgSchemas with the provided arguments.
func NewArgSchemasWithArgs(args []ArgSchema) ArgSchemas {
	return ArgSchemas{
		args:            append([]ArgSchema(nil), args...), // copy slice
		allowAdditional: false,
	}
}

// AddArg adds a new argument to the ArgSchemas.
func (a *ArgSchemas) AddArg(arg ArgSchema) {
	a.args = append(a.args, arg)
}

// SetAllowAdditional sets whether additional arguments are allowed.
func (a *ArgSchemas) SetAllowAdditional(allow bool) {
	a.allowAdditional = allow
}

// SetCollectionMetadata sets the collection-level metadata.
func (a *ArgSchemas) SetCollectionMetadata(name, description string) {
	a.collectionName = name
	a.collectionDescription = description
}

// SetAdditionalSchema sets the schema for additional arguments.
func (a *ArgSchemas) SetAdditionalSchema(schema core.Schema) {
	a.additionalSchema = schema
}

// GetArgs returns the internal args slice for modification (used by FunctionBuilder).
func (a *ArgSchemas) GetArgs() *[]ArgSchema {
	return &a.args
}

// SetOptionalByName sets the optional flag for an argument by name.
func (a *ArgSchemas) SetOptionalByName(name string, optional bool) {
	for i := range a.args {
		if a.args[i].name == name {
			a.args[i].optional = optional
			break
		}
	}
}

// AddConstraintByName adds a constraint to an argument by name.
func (a *ArgSchemas) AddConstraintByName(name string, constraint string) {
	for i := range a.args {
		if a.args[i].name == name {
			a.args[i].constraints = append(a.args[i].constraints, constraint)
			break
		}
	}
}

// API interface implementation for ArgSchemas
func (a *ArgSchemas) Accept(visitor core.SchemaVisitor) error {
	// Visit each arg schema
	for _, arg := range a.args {
		if err := arg.Accept(visitor); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArgSchemas) Args() []core.ArgSchema {
	args := make([]core.ArgSchema, len(a.args))
	for i, arg := range a.args {
		args[i] = &arg
	}
	return args
}

func (a *ArgSchemas) AllowAdditional() bool         { return a.allowAdditional }
func (a *ArgSchemas) AdditionalSchema() core.Schema { return a.additionalSchema }
func (a *ArgSchemas) CollectionName() string        { return a.collectionName }
func (a *ArgSchemas) CollectionDescription() string { return a.collectionDescription }

// ToMap converts ArgSchemas to a map[string]core.Schema for compatibility.
func (args ArgSchemas) ToMap() map[string]core.Schema {
	result := make(map[string]core.Schema)
	for _, arg := range args.args {
		result[arg.Name()] = arg.Schema()
	}
	return result
}

// Names returns the names of all arguments.
func (args ArgSchemas) Names() []string {
	names := make([]string, len(args.args))
	for i, arg := range args.args {
		names[i] = arg.Name()
	}
	return names
}

// RequiredNames returns the names of all required arguments.
func (args ArgSchemas) RequiredNames() []string {
	var required []string
	for _, arg := range args.args {
		if !arg.Optional() {
			required = append(required, arg.Name())
		}
	}
	return required
}

// Get returns the ArgSchema with the given name, if it exists.
func (args ArgSchemas) Get(name string) (*ArgSchema, bool) {
	for _, arg := range args.args {
		if arg.Name() == name {
			return &arg, true
		}
	}
	return nil, false
}

// FunctionSchema represents a function signature as a first-class schema type.
// It validates function inputs, outputs, and error schemas, supporting both
// structural validation and runtime type checking.
//
// Note: Function validation is handled by the consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) for validation.
type FunctionSchema struct {
	metadata          core.SchemaMetadata
	annotations       []core.Annotation
	inputs            ArgSchemas
	outputs           ArgSchemas
	errors            core.Schema
	additionalInputs  bool
	additionalOutputs bool
	examples          []map[string]any
	allowNilError     bool
}

// NewFunctionSchema creates a new FunctionSchema with the provided inputs and outputs.
func NewFunctionSchema(inputs ArgSchemas, outputs ArgSchemas) *FunctionSchema {
	return &FunctionSchema{
		metadata:          core.SchemaMetadata{},
		inputs:            inputs,
		outputs:           outputs,
		additionalInputs:  false,
		additionalOutputs: false,
		examples:          []map[string]any{},
	}
}

// Core Schema interface implementation

func (s *FunctionSchema) Type() core.SchemaType {
	return core.TypeFunction
}

func (s *FunctionSchema) Metadata() core.SchemaMetadata {
	return s.metadata
}

func (s *FunctionSchema) Annotations() []core.Annotation {
	if s.annotations == nil {
		return nil
	}
	result := make([]core.Annotation, len(s.annotations))
	copy(result, s.annotations)
	return result
}

// Note: Validation moved to consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) instead.

func (s *FunctionSchema) Clone() core.Schema {
	// Deep clone inputs
	clonedInputs := ArgSchemas{
		args:                  make([]ArgSchema, len(s.inputs.args)),
		allowAdditional:       s.inputs.allowAdditional,
		additionalSchema:      s.inputs.additionalSchema,
		collectionName:        s.inputs.collectionName,
		collectionDescription: s.inputs.collectionDescription,
	}
	copy(clonedInputs.args, s.inputs.args)

	// Deep clone outputs
	clonedOutputs := ArgSchemas{
		args:                  make([]ArgSchema, len(s.outputs.args)),
		allowAdditional:       s.outputs.allowAdditional,
		additionalSchema:      s.outputs.additionalSchema,
		collectionName:        s.outputs.collectionName,
		collectionDescription: s.outputs.collectionDescription,
	}
	copy(clonedOutputs.args, s.outputs.args)

	// Deep clone examples
	clonedExamples := make([]map[string]any, len(s.examples))
	for i, example := range s.examples {
		clonedExample := make(map[string]any)
		for k, v := range example {
			clonedExample[k] = v
		}
		clonedExamples[i] = clonedExample
	}

	// Clone metadata
	clonedMetadata := core.SchemaMetadata{
		Name:        s.metadata.Name,
		Description: s.metadata.Description,
		Examples:    append([]any(nil), s.metadata.Examples...),
		Tags:        append([]string(nil), s.metadata.Tags...),
	}

	if s.metadata.Properties != nil {
		clonedMetadata.Properties = make(map[string]string)
		for k, v := range s.metadata.Properties {
			clonedMetadata.Properties[k] = v
		}
	}

	return &FunctionSchema{
		metadata:          clonedMetadata,
		inputs:            clonedInputs,
		outputs:           clonedOutputs,
		errors:            s.errors,
		additionalInputs:  s.additionalInputs,
		additionalOutputs: s.additionalOutputs,
		examples:          clonedExamples,
		allowNilError:     s.allowNilError,
	}
}

// FunctionSchema interface implementation (API compliance)

func (s *FunctionSchema) Inputs() core.ArgSchemas {
	return &s.inputs
}

func (s *FunctionSchema) Outputs() core.ArgSchemas {
	return &s.outputs
}

func (s *FunctionSchema) Errors() core.Schema {
	return s.errors
}

func (s *FunctionSchema) RequiredInputs() []string {
	return s.inputs.RequiredNames()
}

func (s *FunctionSchema) RequiredOutputs() []string {
	return s.outputs.RequiredNames()
}

// Visitor pattern support

func (s *FunctionSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitFunction(s)
}

// Additional utility methods

// WithMetadata creates a new FunctionSchema with updated metadata
func (s *FunctionSchema) WithMetadata(metadata core.SchemaMetadata) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.metadata = metadata
	return clone
}

// WithInput adds or updates an input parameter
func (s *FunctionSchema) WithInput(name string, schema core.Schema) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.inputs.args = append(clone.inputs.args, ArgSchema{name: name, schema: schema})
	return clone
}

// WithOutput adds or updates an output parameter
func (s *FunctionSchema) WithOutput(name string, schema core.Schema) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.outputs.args = append(clone.outputs.args, ArgSchema{name: name, schema: schema})
	return clone
}

// WithError sets the error schema
func (s *FunctionSchema) WithError(schema core.Schema) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.errors = schema
	return clone
}

// WithAdditionalInputs controls whether additional inputs are allowed
func (s *FunctionSchema) WithAdditionalInputs(allowed bool) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.additionalInputs = allowed
	return clone
}

// WithAdditionalOutputs controls whether additional outputs are allowed
func (s *FunctionSchema) WithAdditionalOutputs(allowed bool) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.additionalOutputs = allowed
	return clone
}

// WithExample adds an example input set
func (s *FunctionSchema) WithExample(example map[string]any) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.examples = append(clone.examples, example)
	return clone
}

// Introspection methods

// AdditionalInputs returns whether additional inputs are allowed
func (s *FunctionSchema) AdditionalInputs() bool {
	return s.additionalInputs
}

// AdditionalOutputs returns whether additional outputs are allowed
func (s *FunctionSchema) AdditionalOutputs() bool {
	return s.additionalOutputs
}

// Examples returns function input examples
func (s *FunctionSchema) Examples() []map[string]any {
	return append([]map[string]any(nil), s.examples...)
}

// AllowNilError returns whether nil error schema is allowed
func (s *FunctionSchema) AllowNilError() bool {
	return s.allowNilError
}

// String representation for debugging
func (s *FunctionSchema) String() string {
	inputNames := s.inputs.Names()
	outputNames := s.outputs.Names()

	name := s.metadata.Name
	if name == "" {
		name = "anonymous"
	}

	return fmt.Sprintf("FunctionSchema(%s: inputs=[%s] outputs=[%s])",
		name,
		strings.Join(inputNames, ", "),
		strings.Join(outputNames, ", "))
}

// Note: All validation functionality has been moved to the consumer-driven architecture.
// Use schema/consumer.Registry for validation, formatting, and other processing needs.
