package schemas

import (
	"fmt"
	"reflect"
	"strings"

	"defs.dev/schema/api"
)

// ArgSchema represents a named argument with its schema and description.
// This is used for both function inputs and outputs to provide rich metadata.
type ArgSchema struct {
	name        string     `json:"name"`
	description string     `json:"description,omitempty"`
	schema      api.Schema `json:"schema"`
	optional    bool       `json:"optional,omitempty"`
	constraints []string   `json:"constraints,omitempty"`
}

// Ensure ArgSchema implements the API interface at compile time
var _ api.ArgSchema = (*ArgSchema)(nil)

// NewArgSchema creates a new ArgSchema with the given parameters.
func NewArgSchema(name string, schema api.Schema) ArgSchema {
	return ArgSchema{
		name:   name,
		schema: schema,
	}
}

// NewArgSchemaWithOptions creates a new ArgSchema with all options.
func NewArgSchemaWithOptions(name string, schema api.Schema, description string, optional bool, constraints []string) ArgSchema {
	return ArgSchema{
		name:        name,
		description: description,
		schema:      schema,
		optional:    optional,
		constraints: append([]string(nil), constraints...), // copy slice
	}
}

// API interface implementation for ArgSchema
func (a *ArgSchema) Accept(visitor api.SchemaVisitor) error {
	// ArgSchema delegates to its underlying schema
	if accepter, ok := a.schema.(api.Accepter); ok {
		return accepter.Accept(visitor)
	}
	return nil
}

func (a *ArgSchema) Name() string          { return a.name }
func (a *ArgSchema) Description() string   { return a.description }
func (a *ArgSchema) Schema() api.Schema    { return a.schema }
func (a *ArgSchema) Optional() bool        { return a.optional }
func (a *ArgSchema) Constraints() []string { return append([]string(nil), a.constraints...) }

// ArgSchemas represents a collection of named arguments with collection-level metadata.
type ArgSchemas struct {
	args                  []ArgSchema `json:"args"`
	allowAdditional       bool        `json:"allowAdditional,omitempty"`
	additionalSchema      api.Schema  `json:"additionalSchema,omitempty"`
	collectionName        string      `json:"collectionName,omitempty"`
	collectionDescription string      `json:"collectionDescription,omitempty"`
}

// Ensure ArgSchemas implements the API interface at compile time
var _ api.ArgSchemas = (*ArgSchemas)(nil)

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
func (a *ArgSchemas) SetAdditionalSchema(schema api.Schema) {
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
func (a *ArgSchemas) Accept(visitor api.SchemaVisitor) error {
	// Visit each arg schema
	for _, arg := range a.args {
		if err := arg.Accept(visitor); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArgSchemas) Args() []api.ArgSchema {
	args := make([]api.ArgSchema, len(a.args))
	for i, arg := range a.args {
		args[i] = &arg
	}
	return args
}

func (a *ArgSchemas) AllowAdditional() bool         { return a.allowAdditional }
func (a *ArgSchemas) AdditionalSchema() api.Schema  { return a.additionalSchema }
func (a *ArgSchemas) CollectionName() string        { return a.collectionName }
func (a *ArgSchemas) CollectionDescription() string { return a.collectionDescription }

// ToMap converts ArgSchemas to a map[string]api.Schema for compatibility.
func (args ArgSchemas) ToMap() map[string]api.Schema {
	result := make(map[string]api.Schema)
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
type FunctionSchema struct {
	metadata          api.SchemaMetadata
	inputs            ArgSchemas
	outputs           ArgSchemas
	errors            api.Schema
	additionalInputs  bool
	additionalOutputs bool
	examples          []map[string]any
	allowNilError     bool
	inputConstraints  map[string][]string
	outputConstraints map[string][]string
	validationRules   []FunctionValidationRule
}

// FunctionValidationRule defines custom validation logic for function schemas
type FunctionValidationRule func(inputs map[string]any, functionSchema *FunctionSchema) *api.ValidationError

// NewFunctionSchema creates a new FunctionSchema with the provided inputs and outputs.
func NewFunctionSchema(inputs ArgSchemas, outputs ArgSchemas) *FunctionSchema {
	return &FunctionSchema{
		metadata:          api.SchemaMetadata{},
		inputs:            inputs,
		outputs:           outputs,
		additionalInputs:  false,
		additionalOutputs: false,
		examples:          []map[string]any{},
		inputConstraints:  make(map[string][]string),
		outputConstraints: make(map[string][]string),
		validationRules:   []FunctionValidationRule{},
	}
}

// Core Schema interface implementation

func (s *FunctionSchema) Type() api.SchemaType {
	return api.TypeFunction
}

func (s *FunctionSchema) Metadata() api.SchemaMetadata {
	return s.metadata
}

func (s *FunctionSchema) Validate(value any) api.ValidationResult {
	// Functions can validate in several contexts:
	// 1. Function call inputs (map[string]any)
	// 2. Function definition/signature validation
	// 3. Function execution context

	switch v := value.(type) {
	case map[string]any:
		return s.validateInputs(v)
	default:
		// Try to convert structs to maps for validation
		if inputs, err := s.convertToInputMap(value); err == nil {
			return s.validateInputs(inputs)
		}

		return api.ValidationResult{
			Valid: false,
			Errors: []api.ValidationError{{
				Path:       "",
				Message:    fmt.Sprintf("function input must be a map or struct, got %T", value),
				Code:       "invalid_function_input_type",
				Value:      value,
				Expected:   "map[string]any or struct",
				Suggestion: "provide function inputs as key-value pairs",
				Context:    "function_input_validation",
			}},
		}
	}
}

// validateInputs validates function input parameters
func (s *FunctionSchema) validateInputs(inputs map[string]any) api.ValidationResult {
	var errors []api.ValidationError

	// Validate each provided input against its schema
	for name, value := range inputs {
		inputSchema, exists := s.inputs.ToMap()[name]
		if !exists {
			if !s.additionalInputs {
				errors = append(errors, api.ValidationError{
					Path:       name,
					Message:    fmt.Sprintf("unexpected input '%s'", name),
					Code:       "unexpected_input",
					Value:      value,
					Expected:   fmt.Sprintf("one of: %s", strings.Join(s.inputs.Names(), ", ")),
					Suggestion: fmt.Sprintf("remove '%s' or add it to the function schema", name),
					Context:    "function_input_validation",
				})
			}
			continue
		}

		// Validate input value against its schema
		result := inputSchema.Validate(value)
		if !result.Valid {
			for _, err := range result.Errors {
				// Prefix path with input name
				path := name
				if err.Path != "" {
					path = fmt.Sprintf("%s.%s", name, err.Path)
				}

				errors = append(errors, api.ValidationError{
					Path:       path,
					Message:    fmt.Sprintf("input '%s': %s", name, err.Message),
					Code:       err.Code,
					Value:      err.Value,
					Expected:   err.Expected,
					Suggestion: err.Suggestion,
					Context:    "function_input_validation",
				})
			}
		}
	}

	// Apply custom validation rules
	for _, rule := range s.validationRules {
		if err := rule(inputs, s); err != nil {
			errors = append(errors, *err)
		}
	}

	// Check input constraints
	for inputName, constraints := range s.inputConstraints {
		value, exists := inputs[inputName]
		if exists {
			for _, constraint := range constraints {
				if err := s.validateConstraint(inputName, value, constraint); err != nil {
					errors = append(errors, *err)
				}
			}
		}
	}

	return api.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Metadata: map[string]any{
			"function_name":  s.metadata.Name,
			"inputs_count":   len(inputs),
			"required_count": len(s.inputs.RequiredNames()),
		},
	}
}

// validateConstraint validates individual input constraints
func (s *FunctionSchema) validateConstraint(inputName string, value any, constraint string) *api.ValidationError {
	switch constraint {
	case "non_empty":
		if s.isEmpty(value) {
			return &api.ValidationError{
				Path:       inputName,
				Message:    fmt.Sprintf("input '%s' cannot be empty", inputName),
				Code:       "empty_input_value",
				Value:      value,
				Expected:   "non-empty value",
				Suggestion: fmt.Sprintf("provide a non-empty value for '%s'", inputName),
				Context:    "function_constraint_validation",
			}
		}
	case "unique":
		// This would require comparison with other function calls - skip for now
	default:
		return &api.ValidationError{
			Path:       inputName,
			Message:    fmt.Sprintf("unknown constraint '%s' for input '%s'", constraint, inputName),
			Code:       "unknown_constraint",
			Expected:   "valid constraint type",
			Suggestion: "use supported constraints like 'non_empty'",
			Context:    "function_constraint_validation",
		}
	}
	return nil
}

// isEmpty checks if a value is considered empty
func (s *FunctionSchema) isEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			return rv.Len() == 0
		case reflect.Ptr, reflect.Interface:
			return rv.IsNil()
		}
	}
	return false
}

// convertToInputMap converts structs to input maps using reflection
func (s *FunctionSchema) convertToInputMap(value any) (map[string]any, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot convert %T to input map", value)
	}

	result := make(map[string]any)
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get field name from JSON tag or use field name
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		result[fieldName] = fieldValue.Interface()
	}

	return result, nil
}

// getInputNames returns a list of all input parameter names
func (s *FunctionSchema) getInputNames() []string {
	names := make([]string, 0, len(s.inputs.args))
	for name := range s.inputs.ToMap() {
		names = append(names, name)
	}
	return names
}

func (s *FunctionSchema) ToJSONSchema() map[string]any {
	// Function schemas are represented as objects with input properties
	properties := make(map[string]any)
	for name, inputSchema := range s.inputs.ToMap() {
		properties[name] = inputSchema.ToJSONSchema()
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
		"x-function": true, // Mark as function schema
	}

	if len(s.inputs.RequiredNames()) > 0 {
		schema["required"] = s.inputs.RequiredNames()
	}

	if !s.additionalInputs {
		schema["additionalProperties"] = false
	}

	// Add metadata
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}

	if s.metadata.Name != "" {
		schema["title"] = s.metadata.Name
	}

	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	if len(s.metadata.Tags) > 0 {
		schema["tags"] = s.metadata.Tags
	}

	// Add function-specific metadata
	if len(s.outputs.args) > 0 {
		schema["x-returns"] = s.outputs.ToMap()
	}

	if s.errors != nil {
		schema["x-errors"] = s.errors.ToJSONSchema()
	}

	if len(s.examples) > 0 {
		schema["x-input-examples"] = s.examples
	}

	if len(s.inputConstraints) > 0 {
		schema["x-input-constraints"] = s.inputConstraints
	}

	return schema
}

func (s *FunctionSchema) GenerateExample() any {
	// If we have explicit examples, return the first one
	if len(s.examples) > 0 {
		return s.examples[0]
	}

	// Generate example from input schemas
	example := make(map[string]any)
	for name, inputSchema := range s.inputs.ToMap() {
		example[name] = inputSchema.GenerateExample()
	}

	// Ensure all required inputs are present
	for _, required := range s.inputs.RequiredNames() {
		if _, exists := example[required]; !exists {
			if inputSchema, exists := s.inputs.ToMap()[required]; exists {
				example[required] = inputSchema.GenerateExample()
			} else {
				example[required] = fmt.Sprintf("example_%s", required)
			}
		}
	}

	return example
}

func (s *FunctionSchema) Clone() api.Schema {
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

	// Deep clone constraints - remove these since they're now in ArgSchema
	clonedInputConstraints := make(map[string][]string)
	for k, v := range s.inputConstraints {
		clonedInputConstraints[k] = append([]string(nil), v...)
	}

	clonedOutputConstraints := make(map[string][]string)
	for k, v := range s.outputConstraints {
		clonedOutputConstraints[k] = append([]string(nil), v...)
	}

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
	clonedMetadata := api.SchemaMetadata{
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
		inputConstraints:  clonedInputConstraints,
		outputConstraints: clonedOutputConstraints,
		validationRules:   append([]FunctionValidationRule(nil), s.validationRules...),
	}
}

// FunctionSchema interface implementation (API compliance)

func (s *FunctionSchema) Inputs() api.ArgSchemas {
	return &s.inputs
}

func (s *FunctionSchema) Outputs() api.ArgSchemas {
	return &s.outputs
}

func (s *FunctionSchema) Errors() api.Schema {
	return s.errors
}

func (s *FunctionSchema) RequiredInputs() []string {
	return s.inputs.RequiredNames()
}

func (s *FunctionSchema) RequiredOutputs() []string {
	return s.outputs.RequiredNames()
}

// Visitor pattern support

func (s *FunctionSchema) Accept(visitor api.SchemaVisitor) error {
	return visitor.VisitFunction(s)
}

// Additional utility methods

// WithMetadata creates a new FunctionSchema with updated metadata
func (s *FunctionSchema) WithMetadata(metadata api.SchemaMetadata) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.metadata = metadata
	return clone
}

// WithInput adds or updates an input parameter
func (s *FunctionSchema) WithInput(name string, schema api.Schema) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.inputs.args = append(clone.inputs.args, ArgSchema{name: name, schema: schema})
	return clone
}

// WithOutput adds or updates an output parameter
func (s *FunctionSchema) WithOutput(name string, schema api.Schema) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.outputs.args = append(clone.outputs.args, ArgSchema{name: name, schema: schema})
	return clone
}

// WithError sets the error schema
func (s *FunctionSchema) WithError(schema api.Schema) *FunctionSchema {
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

// WithConstraint adds a constraint to an input parameter
func (s *FunctionSchema) WithConstraint(inputName string, constraint string) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.inputConstraints[inputName] = append(clone.inputConstraints[inputName], constraint)
	return clone
}

// WithValidationRule adds a custom validation rule
func (s *FunctionSchema) WithValidationRule(rule FunctionValidationRule) *FunctionSchema {
	clone := s.Clone().(*FunctionSchema)
	clone.validationRules = append(clone.validationRules, rule)
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

// InputConstraints returns input parameter constraints
func (s *FunctionSchema) InputConstraints() map[string][]string {
	constraints := make(map[string][]string)
	for k, v := range s.inputConstraints {
		constraints[k] = append([]string(nil), v...)
	}
	return constraints
}

// OutputConstraints returns output parameter constraints
func (s *FunctionSchema) OutputConstraints() map[string][]string {
	constraints := make(map[string][]string)
	for k, v := range s.outputConstraints {
		constraints[k] = append([]string(nil), v...)
	}
	return constraints
}

// AllowNilError returns whether nil error schema is allowed
func (s *FunctionSchema) AllowNilError() bool {
	return s.allowNilError
}

// Validation helpers

// ValidateOutput validates a function's output value
func (s *FunctionSchema) ValidateOutput(value any) api.ValidationResult {
	if len(s.outputs.args) == 0 {
		if value == nil {
			return api.ValidationResult{Valid: true}
		}
		if !s.additionalOutputs {
			return api.ValidationResult{
				Valid: false,
				Errors: []api.ValidationError{{
					Path:       "output",
					Message:    "function output is not defined in schema",
					Code:       "undefined_output",
					Value:      value,
					Expected:   "nil (no output expected)",
					Suggestion: "ensure function does not return a value or define output schema",
					Context:    "function_output_validation",
				}},
			}
		}
	}

	for _, output := range s.outputs.args {
		result := output.schema.Validate(value)
		if !result.Valid {
			// Prefix errors with output context
			for i := range result.Errors {
				result.Errors[i].Context = "function_output_validation"
				if result.Errors[i].Path == "" {
					result.Errors[i].Path = "output"
				} else {
					result.Errors[i].Path = fmt.Sprintf("output.%s", result.Errors[i].Path)
				}
			}
		}
		return result
	}

	return api.ValidationResult{Valid: true}
}

// ValidateError validates a function's error value
func (s *FunctionSchema) ValidateError(value any) api.ValidationResult {
	if s.errors == nil {
		if value == nil {
			return api.ValidationResult{Valid: true}
		}
		if !s.allowNilError {
			return api.ValidationResult{
				Valid: false,
				Errors: []api.ValidationError{{
					Path:       "error",
					Message:    "function error is not defined in schema",
					Code:       "undefined_error",
					Value:      value,
					Expected:   "nil (no error expected)",
					Suggestion: "ensure function does not return an error or define error schema",
					Context:    "function_error_validation",
				}},
			}
		}
	}

	if s.errors != nil {
		result := s.errors.Validate(value)
		if !result.Valid {
			// Prefix errors with error context
			for i := range result.Errors {
				result.Errors[i].Context = "function_error_validation"
				if result.Errors[i].Path == "" {
					result.Errors[i].Path = "error"
				} else {
					result.Errors[i].Path = fmt.Sprintf("error.%s", result.Errors[i].Path)
				}
			}
		}
		return result
	}

	return api.ValidationResult{Valid: true}
}

// String representation for debugging
func (s *FunctionSchema) String() string {
	inputNames := make([]string, 0, len(s.inputs.args))
	for name := range s.inputs.ToMap() {
		inputNames = append(inputNames, name)
	}

	name := s.metadata.Name
	if name == "" {
		name = "anonymous"
	}

	return fmt.Sprintf("FunctionSchema(%s: inputs=%v, required=%v)", name, inputNames, s.inputs.RequiredNames())
}
