package typescript

import (
	"fmt"
	"strings"

	"defs.dev/schema/core"
	"defs.dev/schema/visitors/export"
	"defs.dev/schema/visitors/export/base"
)

// Generator generates TypeScript code from schema types using the visitor pattern.
type Generator struct {
	*base.BaseVisitor
	options   TypeScriptOptions
	context   *base.GenerationContext
	mapper    *TypeMapper
	formatter *CodeFormatter
	result    []string
}

// NewGenerator creates a new TypeScript generator with the given options.
func NewGenerator(options ...export.Option) *Generator {
	g := &Generator{
		BaseVisitor: base.NewBaseVisitor("typescript"),
		options:     DefaultTypeScriptOptions(),
		result:      make([]string, 0),
	}

	// Apply functional options
	for _, opt := range options {
		opt.Apply(g)
	}

	// Validate options
	if err := g.options.Validate(); err != nil {
		panic(fmt.Sprintf("invalid TypeScript options: %v", err))
	}

	g.context = base.NewGenerationContext()
	g.mapper = NewTypeMapper(g.options, g.context)
	g.formatter = NewCodeFormatter(g.options, g.context)

	return g
}

// Generate generates TypeScript code for the given schema (implements export.Generator).
func (g *Generator) Generate(s core.Schema) ([]byte, error) {
	// Reset result
	g.result = make([]string, 0)
	g.context = base.NewGenerationContext()

	// Accept the visitor pattern
	if accepter, ok := s.(core.Accepter); ok {
		if err := accepter.Accept(g); err != nil {
			return nil, base.NewGenerationError("typescript", string(s.Type()), err.Error())
		}
	} else {
		return nil, base.NewGenerationError("typescript", string(s.Type()), "schema does not implement Accepter interface")
	}

	// Join all lines and return as bytes
	output := strings.Join(g.result, "\n")
	return []byte(output), nil
}

// Name returns the generator name (implements export.Generator).
func (g *Generator) Name() string {
	return "TypeScript Generator"
}

// Format returns the output format (implements export.Generator).
func (g *Generator) Format() string {
	return "typescript"
}

// GetOptions returns the current generator options.
func (g *Generator) GetOptions() TypeScriptOptions {
	return g.options.Clone()
}

// SetOptions allows updating generator options.
func (g *Generator) SetOptions(options TypeScriptOptions) {
	g.options = options.Clone()
	g.mapper = NewTypeMapper(g.options, g.context)
	g.formatter = NewCodeFormatter(g.options, g.context)
}

// Visitor methods for different schema types

// VisitString generates TypeScript for string types.
func (g *Generator) VisitString(s core.StringSchema) error {
	metadata := s.Metadata()
	typeName := g.mapper.FormatTypeName(metadata.Name)

	if typeName == "" || typeName == "UnnamedType" {
		// For unnamed strings, just return the base type
		g.addSimpleType("string")
		return nil
	}

	// Check if this is an enum
	if enumValues := s.EnumValues(); len(enumValues) > 0 {
		return g.generateStringEnum(typeName, enumValues, metadata)
	}

	// Generate type alias for named string types
	return g.generateStringType(typeName, s, metadata)
}

// VisitInteger generates TypeScript for integer types.
func (g *Generator) VisitInteger(s core.IntegerSchema) error {
	metadata := s.Metadata()
	typeName := g.mapper.FormatTypeName(metadata.Name)

	if typeName == "" || typeName == "UnnamedType" {
		g.addSimpleType("number")
		return nil
	}

	return g.generateNumberType(typeName, "number", metadata)
}

// VisitNumber generates TypeScript for number types.
func (g *Generator) VisitNumber(s core.NumberSchema) error {
	metadata := s.Metadata()
	typeName := g.mapper.FormatTypeName(metadata.Name)

	if typeName == "" || typeName == "UnnamedType" {
		g.addSimpleType("number")
		return nil
	}

	return g.generateNumberType(typeName, "number", metadata)
}

// VisitBoolean generates TypeScript for boolean types.
func (g *Generator) VisitBoolean(s core.BooleanSchema) error {
	metadata := s.Metadata()
	typeName := g.mapper.FormatTypeName(metadata.Name)

	if typeName == "" || typeName == "UnnamedType" {
		g.addSimpleType("boolean")
		return nil
	}

	return g.generateBooleanType(typeName, metadata)
}

// VisitArray generates TypeScript for array types.
func (g *Generator) VisitArray(s core.ArraySchema) error {
	metadata := s.Metadata()
	typeName := g.mapper.FormatTypeName(metadata.Name)

	// Generate the item type
	var itemType string
	if itemSchema := s.ItemSchema(); itemSchema != nil {
		itemGenerator := NewGenerator()
		itemOutput, err := itemGenerator.Generate(itemSchema)
		if err != nil {
			return fmt.Errorf("failed to generate item type: %w", err)
		}
		itemType = strings.TrimSpace(string(itemOutput))
	} else {
		itemType = g.mapper.MapSchemaType(core.TypeAny)
	}

	arrayType := g.mapper.FormatArrayType(itemType)

	if typeName == "" || typeName == "UnnamedType" {
		g.addSimpleType(arrayType)
		return nil
	}

	return g.generateArrayType(typeName, arrayType, metadata)
}

// VisitObject generates TypeScript for object types.
func (g *Generator) VisitObject(s core.ObjectSchema) error {
	metadata := s.Metadata()
	typeName := g.mapper.FormatTypeName(metadata.Name)

	if typeName == "" || typeName == "UnnamedType" {
		g.addSimpleType("object")
		return nil
	}

	return g.generateObjectType(typeName, s, metadata)
}

// Helper methods for generating different TypeScript constructs

// addSimpleType adds a simple type without declaration.
func (g *Generator) addSimpleType(typeStr string) {
	g.result = append(g.result, typeStr)
}

// generateStringEnum generates a TypeScript enum for string values.
func (g *Generator) generateStringEnum(name string, values []string, metadata core.SchemaMetadata) error {
	// Add JSDoc if enabled
	if g.options.IncludeJSDoc && metadata.Description != "" {
		jsdocLines := g.formatter.FormatJSDoc(metadata.Description, metadata.Examples, nil)
		g.result = append(g.result, jsdocLines...)
	}

	enumFormatter := NewEnumFormatter(g.options, g.context)
	enumLines := enumFormatter.FormatStringEnum(name, values, true)
	g.result = append(g.result, enumLines...)

	return nil
}

// generateStringType generates a TypeScript type alias for string types.
func (g *Generator) generateStringType(name string, s core.StringSchema, metadata core.SchemaMetadata) error {
	// Add JSDoc if enabled
	if g.options.IncludeJSDoc && metadata.Description != "" {
		jsdocLines := g.formatter.FormatJSDoc(metadata.Description, metadata.Examples, s.DefaultValue())
		g.result = append(g.result, jsdocLines...)
	}

	typeLines := g.formatter.FormatType(name, "string", true)
	g.result = append(g.result, typeLines...)

	return nil
}

// generateNumberType generates a TypeScript type alias for number types.
func (g *Generator) generateNumberType(name string, baseType string, metadata core.SchemaMetadata) error {
	// Add JSDoc if enabled
	if g.options.IncludeJSDoc && metadata.Description != "" {
		jsdocLines := g.formatter.FormatJSDoc(metadata.Description, metadata.Examples, nil)
		g.result = append(g.result, jsdocLines...)
	}

	typeLines := g.formatter.FormatType(name, baseType, true)
	g.result = append(g.result, typeLines...)

	return nil
}

// generateBooleanType generates a TypeScript type alias for boolean types.
func (g *Generator) generateBooleanType(name string, metadata core.SchemaMetadata) error {
	// Add JSDoc if enabled
	if g.options.IncludeJSDoc && metadata.Description != "" {
		jsdocLines := g.formatter.FormatJSDoc(metadata.Description, metadata.Examples, nil)
		g.result = append(g.result, jsdocLines...)
	}

	typeLines := g.formatter.FormatType(name, "boolean", true)
	g.result = append(g.result, typeLines...)

	return nil
}

// generateArrayType generates a TypeScript type alias for array types.
func (g *Generator) generateArrayType(name string, arrayType string, metadata core.SchemaMetadata) error {
	// Add JSDoc if enabled
	if g.options.IncludeJSDoc && metadata.Description != "" {
		jsdocLines := g.formatter.FormatJSDoc(metadata.Description, metadata.Examples, nil)
		g.result = append(g.result, jsdocLines...)
	}

	finalArrayType := arrayType
	if g.options.StrictMode {
		// Make arrays readonly in strict mode
		finalArrayType = fmt.Sprintf("readonly %s", arrayType)
	}

	typeLines := g.formatter.FormatType(name, finalArrayType, true)
	g.result = append(g.result, typeLines...)

	return nil
}

// generateObjectType generates a TypeScript interface or type for object types.
func (g *Generator) generateObjectType(name string, s core.ObjectSchema, metadata core.SchemaMetadata) error {
	// Add JSDoc if enabled
	if g.options.IncludeJSDoc && metadata.Description != "" {
		jsdocLines := g.formatter.FormatJSDoc(metadata.Description, metadata.Examples, nil)
		g.result = append(g.result, jsdocLines...)
	}

	properties := s.Properties()
	required := s.Required()
	requiredMap := make(map[string]bool)
	for _, req := range required {
		requiredMap[req] = true
	}

	// Generate properties
	var props []Property
	for propName, propSchema := range properties {
		propType, err := g.generatePropertyType(propSchema)
		if err != nil {
			return fmt.Errorf("failed to generate property %s: %w", propName, err)
		}

		propMetadata := propSchema.Metadata()
		prop := Property{
			Name:         g.mapper.FormatPropertyName(propName),
			Type:         propType,
			Required:     requiredMap[propName],
			ReadOnly:     false, // Could be extended to support readonly properties
			Description:  propMetadata.Description,
			Examples:     propMetadata.Examples,
			DefaultValue: nil, // Could be extended to support default values
		}
		props = append(props, prop)
	}

	// Generate interface or type based on options
	switch g.options.OutputStyle {
	case "interface":
		interfaceLines := g.formatter.FormatInterface(name, props, true)
		g.result = append(g.result, interfaceLines...)
	case "type":
		typeDefinition := g.formatObjectAsType(props)
		typeLines := g.formatter.FormatType(name, typeDefinition, true)
		g.result = append(g.result, typeLines...)
	default:
		interfaceLines := g.formatter.FormatInterface(name, props, true)
		g.result = append(g.result, interfaceLines...)
	}

	return nil
}

// generatePropertyType generates the TypeScript type for a property.
func (g *Generator) generatePropertyType(propSchema core.Schema) (string, error) {
	propGenerator := NewGenerator()
	propOutput, err := propGenerator.Generate(propSchema)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(propOutput)), nil
}

// formatObjectAsType formats object properties as a type definition.
func (g *Generator) formatObjectAsType(props []Property) string {
	if len(props) == 0 {
		return "object"
	}

	var propStrings []string
	for _, prop := range props {
		optional := ""
		if !prop.Required && g.options.UseOptionalProperties {
			optional = "?"
		}

		readonly := ""
		if g.options.StrictMode && prop.ReadOnly {
			readonly = "readonly "
		}

		propStr := fmt.Sprintf("%s%s%s: %s", readonly, prop.Name, optional, prop.Type)
		propStrings = append(propStrings, propStr)
	}

	return fmt.Sprintf("{\n  %s\n}", strings.Join(propStrings, ";\n  "))
}
