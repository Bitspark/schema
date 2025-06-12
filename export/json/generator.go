package json

import (
	"encoding/json"
	"fmt"

	"defs.dev/schema/api/core"
	"defs.dev/schema/export"
	"defs.dev/schema/export/base"
)

// Generator generates JSON Schema from schema types using the visitor pattern.
type Generator struct {
	*base.BaseVisitor
	options JSONSchemaOptions
	result  map[string]any
}

// NewGenerator creates a new JSON Schema generator with the given options.
func NewGenerator(options ...export.Option) *Generator {
	g := &Generator{
		BaseVisitor: base.NewBaseVisitor("json"),
		options:     DefaultJSONSchemaOptions(),
		result:      make(map[string]any),
	}

	// Apply functional options
	for _, opt := range options {
		opt.Apply(g)
	}

	// Validate options
	if err := g.options.Validate(); err != nil {
		panic(fmt.Sprintf("invalid JSON Schema options: %v", err))
	}

	return g
}

// OptionApplier allows setting options on the JSON generator.
type OptionApplier interface {
	ApplyToJSONGenerator(*Generator)
}

// Generate generates JSON Schema for the given schema (implements export.Generator).
func (g *Generator) Generate(s core.Schema) ([]byte, error) {
	// Reset result
	g.result = make(map[string]any)

	// Accept the visitor pattern
	if accepter, ok := s.(core.Accepter); ok {
		if err := accepter.Accept(g); err != nil {
			return nil, base.NewGenerationError("json", string(s.Type()), err.Error())
		}
	} else {
		return nil, base.NewGenerationError("json", string(s.Type()), "schema does not implement Accepter interface")
	}

	// Add schema metadata
	g.addSchemaMetadata()

	// Convert to JSON
	result, err := g.marshalResult()
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

// Name returns the generator name (implements export.Generator).
func (g *Generator) Name() string {
	return "JSON Schema Generator"
}

// Format returns the output format (implements export.Generator).
func (g *Generator) Format() string {
	return "json-schema"
}

// GetOptions returns the current generator options.
func (g *Generator) GetOptions() JSONSchemaOptions {
	return g.options.Clone()
}

// SetOptions allows updating generator options.
func (g *Generator) SetOptions(options JSONSchemaOptions) {
	g.options = options.Clone()
}

// addSchemaMetadata adds top-level schema metadata.
func (g *Generator) addSchemaMetadata() {
	if g.options.SchemaURI != "" {
		g.result["$schema"] = g.options.SchemaURI
	}
	if g.options.RootID != "" {
		g.result["$id"] = g.options.RootID
	}
}

// marshalResult converts the result to JSON string with appropriate formatting.
func (g *Generator) marshalResult() (string, error) {
	var data []byte
	var err error

	if g.options.PrettyPrint && !g.options.MinifyOutput {
		data, err = json.MarshalIndent(g.result, "", "  ")
	} else {
		data, err = json.Marshal(g.result)
	}

	if err != nil {
		return "", base.NewGenerationError("json", "marshal", "failed to marshal JSON")
	}

	return string(data), nil
}

// Visitor methods for different schema types

// VisitString generates JSON Schema for string types.
func (g *Generator) VisitString(s core.StringSchema) error {
	jsonSchema := map[string]any{
		"type": "string",
	}

	// Add constraints using the schema's introspection methods
	if minLen := s.MinLength(); minLen != nil {
		jsonSchema["minLength"] = *minLen
	}
	if maxLen := s.MaxLength(); maxLen != nil {
		jsonSchema["maxLength"] = *maxLen
	}
	if pattern := s.Pattern(); pattern != "" {
		jsonSchema["pattern"] = pattern
	}
	if g.options.IncludeFormat && s.Format() != "" {
		jsonSchema["format"] = s.Format()
	}
	if enum := s.EnumValues(); len(enum) > 0 {
		jsonSchema["enum"] = enum
	}
	if g.options.IncludeDefaults && s.DefaultValue() != nil {
		jsonSchema["default"] = *s.DefaultValue()
	}

	g.addCommonMetadata(jsonSchema, s)
	g.result = jsonSchema
	return nil
}

// VisitInteger generates JSON Schema for integer types.
func (g *Generator) VisitInteger(s core.IntegerSchema) error {
	jsonSchema := map[string]any{
		"type": "integer",
	}

	// Add constraints using the schema's introspection methods
	if min := s.Minimum(); min != nil {
		jsonSchema["minimum"] = *min
	}
	if max := s.Maximum(); max != nil {
		jsonSchema["maximum"] = *max
	}

	g.addCommonMetadata(jsonSchema, s)
	g.result = jsonSchema
	return nil
}

// VisitNumber generates JSON Schema for number types.
func (g *Generator) VisitNumber(s core.NumberSchema) error {
	jsonSchema := map[string]any{
		"type": "number",
	}

	// Add constraints using the schema's introspection methods
	if min := s.Minimum(); min != nil {
		jsonSchema["minimum"] = *min
	}
	if max := s.Maximum(); max != nil {
		jsonSchema["maximum"] = *max
	}

	g.addCommonMetadata(jsonSchema, s)
	g.result = jsonSchema
	return nil
}

// VisitBoolean generates JSON Schema for boolean types.
func (g *Generator) VisitBoolean(s core.BooleanSchema) error {
	jsonSchema := map[string]any{
		"type": "boolean",
	}

	g.addCommonMetadata(jsonSchema, s)
	g.result = jsonSchema
	return nil
}

// VisitArray generates JSON Schema for array types.
func (g *Generator) VisitArray(s core.ArraySchema) error {
	jsonSchema := map[string]any{
		"type": "array",
	}

	// Generate schema for items if present
	if itemSchema := s.ItemSchema(); itemSchema != nil {
		itemsGenerator := NewGenerator(
			WithSchemaURI(""), // Don't add $schema to nested schemas
		)
		itemsJSON, err := itemsGenerator.Generate(itemSchema)
		if err != nil {
			return fmt.Errorf("failed to generate items schema: %w", err)
		}

		var itemsSchema any
		if err := json.Unmarshal([]byte(itemsJSON), &itemsSchema); err != nil {
			return fmt.Errorf("failed to parse items schema: %w", err)
		}
		jsonSchema["items"] = itemsSchema
	}

	// Add array constraints
	if minItems := s.MinItems(); minItems != nil {
		jsonSchema["minItems"] = *minItems
	}
	if maxItems := s.MaxItems(); maxItems != nil {
		jsonSchema["maxItems"] = *maxItems
	}
	if s.UniqueItemsRequired() {
		jsonSchema["uniqueItems"] = true
	}

	g.addCommonMetadata(jsonSchema, s)
	g.result = jsonSchema
	return nil
}

// VisitObject generates JSON Schema for object types.
func (g *Generator) VisitObject(s core.ObjectSchema) error {
	jsonSchema := map[string]any{
		"type": "object",
	}

	// Generate properties
	properties := s.Properties()
	if len(properties) > 0 {
		propsJSON := make(map[string]any)
		for name, prop := range properties {
			propGenerator := NewGenerator(
				WithSchemaURI(""), // Don't add $schema to nested schemas
			)
			propJSON, err := propGenerator.Generate(prop)
			if err != nil {
				return fmt.Errorf("failed to generate property %s: %w", name, err)
			}

			var propSchema any
			if err := json.Unmarshal([]byte(propJSON), &propSchema); err != nil {
				return fmt.Errorf("failed to parse property %s schema: %w", name, err)
			}
			propsJSON[name] = propSchema
		}
		jsonSchema["properties"] = propsJSON
	}

	// Add required fields
	if required := s.Required(); len(required) > 0 {
		jsonSchema["required"] = required
	}

	// Add additional properties
	if g.options.IncludeAdditionalProperties {
		jsonSchema["additionalProperties"] = s.AdditionalProperties()
	}

	g.addCommonMetadata(jsonSchema, s)
	g.result = jsonSchema
	return nil
}

// addCommonMetadata adds common metadata from schema to JSON Schema.
func (g *Generator) addCommonMetadata(jsonSchema map[string]any, s core.Schema) {
	metadata := s.Metadata()

	if g.options.IncludeTitle && metadata.Name != "" {
		jsonSchema["title"] = metadata.Name
	}

	if g.options.IncludeDescription && metadata.Description != "" {
		jsonSchema["description"] = metadata.Description
	}

	if g.options.IncludeExamples && len(metadata.Examples) > 0 {
		jsonSchema["examples"] = metadata.Examples
	}
}
