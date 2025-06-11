package schema

import (
	"context"
	"encoding/json"
	"fmt"
)

// FunctionHandler represents a local function that can be served as HTTP endpoint
type FunctionHandler func(ctx context.Context, params FunctionInput) (FunctionOutput, error)

// FunctionSchema represents a function signature as a first-class schema type
type FunctionSchema struct {
	metadata SchemaMetadata
	inputs   map[string]Schema
	outputs  Schema
	errors   Schema
	required []string
}

// Introspection methods for FunctionSchema
func (s *FunctionSchema) Inputs() map[string]Schema {
	// Return a copy to prevent external mutation
	inputs := make(map[string]Schema)
	for k, v := range s.inputs {
		inputs[k] = v
	}
	return inputs
}

func (s *FunctionSchema) Outputs() Schema {
	return s.outputs
}

func (s *FunctionSchema) Errors() Schema {
	return s.errors
}

func (s *FunctionSchema) Required() []string {
	// Return a copy to prevent external mutation
	return append([]string(nil), s.required...)
}

func (s *FunctionSchema) Type() SchemaType {
	return TypeFunction
}

func (s *FunctionSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *FunctionSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *FunctionSchema) Clone() Schema {
	clone := *s
	clone.inputs = make(map[string]Schema)
	for k, v := range s.inputs {
		clone.inputs[k] = v.Clone()
	}
	if s.outputs != nil {
		clone.outputs = s.outputs.Clone()
	}
	if s.errors != nil {
		clone.errors = s.errors.Clone()
	}
	clone.required = append([]string(nil), s.required...)
	return &clone
}

func (s *FunctionSchema) Validate(value any) ValidationResult {
	// Function schemas validate function signatures, not values
	// This would validate that a given value conforms to the function signature
	// For now, we'll implement basic validation
	return ValidationResult{Valid: true}
}

func (s *FunctionSchema) ToJSONSchema() map[string]any {
	properties := make(map[string]any)
	for name, inputSchema := range s.inputs {
		properties[name] = inputSchema.ToJSONSchema()
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(s.required) > 0 {
		schema["required"] = s.required
	}

	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}

	// Add function-specific metadata
	if s.outputs != nil {
		schema["returns"] = s.outputs.ToJSONSchema()
	}
	if s.errors != nil {
		schema["errors"] = s.errors.ToJSONSchema()
	}

	return schema
}

func (s *FunctionSchema) GenerateExample() any {
	example := make(map[string]any)
	for name, inputSchema := range s.inputs {
		example[name] = inputSchema.GenerateExample()
	}
	return example
}

// Function schema builder for creating function signatures
func NewFunctionSchema() *FunctionSchemaBuilder {
	return &FunctionSchemaBuilder{
		schema: &FunctionSchema{
			metadata: SchemaMetadata{},
			inputs:   make(map[string]Schema),
		},
	}
}

type FunctionSchemaBuilder struct {
	schema *FunctionSchema
}

func (b *FunctionSchemaBuilder) Input(name string, schema Schema) *FunctionSchemaBuilder {
	b.schema.inputs[name] = schema
	return b
}

func (b *FunctionSchemaBuilder) Output(schema Schema) *FunctionSchemaBuilder {
	b.schema.outputs = schema
	return b
}

func (b *FunctionSchemaBuilder) Error(schema Schema) *FunctionSchemaBuilder {
	b.schema.errors = schema
	return b
}

func (b *FunctionSchemaBuilder) Required(names ...string) *FunctionSchemaBuilder {
	b.schema.required = append(b.schema.required, names...)
	return b
}

func (b *FunctionSchemaBuilder) Description(desc string) *FunctionSchemaBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *FunctionSchemaBuilder) Name(name string) *FunctionSchemaBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *FunctionSchemaBuilder) Example(example map[string]any) *FunctionSchemaBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *FunctionSchemaBuilder) Tag(tag string) *FunctionSchemaBuilder {
	b.schema.metadata.Tags = append(b.schema.metadata.Tags, tag)
	return b
}

func (b *FunctionSchemaBuilder) Build() Schema {
	return b.schema
}

// TypedFunction interface - universal callable function concept
type TypedFunction interface {
	Function

	CallTyped(ctx context.Context, input any, output any) error
}

// Function interface for backward compatibility - to be deprecated
type Function interface {
	Call(ctx context.Context, params FunctionInput) (FunctionOutput, error)
	Schema() *FunctionSchema
	Name() string
}

func Typed(fn Function) TypedFunction {
	return &FunctionTyper{Function: fn}
}

// FunctionTyper wraps a Function to implement the new TypedFunction interface
type FunctionTyper struct {
	Function
}

// NewFunctionAdapter creates an adapter from a legacy function
func NewFunctionAdapter(fn Function) TypedFunction {
	return &FunctionTyper{Function: fn}
}

// Call implements the new Function interface using FunctionInput/FunctionOutput
func (fa *FunctionTyper) Call(ctx context.Context, params FunctionInput) (FunctionOutput, error) {
	return fa.Function.Call(ctx, params)
}

// CallTyped implements typed execution by delegating to Call
func (fa *FunctionTyper) CallTyped(ctx context.Context, input any, output any) error {
	functionInput, err := ConvertToFunctionInput(input)
	if err != nil {
		return err
	}

	result, err := fa.Call(ctx, functionInput)
	if err != nil {
		return err
	}

	// Marshal and unmarshal for type conversion
	jsonData, err := json.Marshal(result.Value())
	if err != nil {
		return fmt.Errorf("result marshaling failed: %w", err)
	}

	if err := json.Unmarshal(jsonData, output); err != nil {
		return fmt.Errorf("result unmarshaling failed: %w", err)
	}

	return nil
}

// Schema delegates to the legacy function
func (fa *FunctionTyper) Schema() *FunctionSchema {
	return fa.Function.Schema()
}

// Name delegates to the legacy function
func (fa *FunctionTyper) Name() string {
	return fa.Function.Name()
}
