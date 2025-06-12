package python

import (
	"fmt"
	"strings"

	"defs.dev/schema/api/core"
	"defs.dev/schema/export/base"
)

// Generator generates Python code from schema definitions using the visitor pattern.
type Generator struct {
	base.BaseVisitor
	options       PythonOptions
	context       *base.GenerationContext
	typeMapper    *TypeMapper
	formatter     *CodeFormatter
	enumFormatter *EnumFormatter
	importManager *ImportManager
	output        strings.Builder
}

// NewGenerator creates a new Python generator with the given options.
func NewGenerator(options PythonOptions) *Generator {
	context := base.NewGenerationContext()

	return &Generator{
		options:       options,
		context:       context,
		typeMapper:    NewTypeMapper(options, context),
		formatter:     NewCodeFormatter(options, context),
		enumFormatter: NewEnumFormatter(options, context),
		importManager: NewImportManager(options),
	}
}

// Generate generates Python code from a schema.
func (g *Generator) Generate(schema core.Schema) ([]byte, error) {
	// Reset output
	g.output.Reset()

	// Validate options
	if err := g.options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Add file header if specified
	if g.options.FileHeader != "" {
		g.writeFileHeader()
	}

	// Generate the schema using visitor pattern
	if accepter, ok := schema.(core.Accepter); ok {
		if err := accepter.Accept(g); err != nil {
			return nil, fmt.Errorf("failed to generate Python code: %w", err)
		}
	} else {
		return nil, fmt.Errorf("schema does not support visitor pattern")
	}

	// Add imports at the beginning
	result := g.output.String()
	if g.options.IncludeImports {
		result = g.generateImports() + "\n\n" + result
	}

	return []byte(result), nil
}

// writeFileHeader writes the file header comment.
func (g *Generator) writeFileHeader() {
	lines := strings.Split(g.options.FileHeader, "\n")
	for _, line := range lines {
		g.output.WriteString(fmt.Sprintf("# %s\n", line))
	}
	g.output.WriteString("\n")
}

// generateImports generates the import statements.
func (g *Generator) generateImports() string {
	imports := g.importManager.GetRequiredImports()
	if len(imports) == 0 {
		return ""
	}

	var result strings.Builder
	for _, imp := range imports {
		result.WriteString(imp + "\n")
	}

	return result.String()
}

// VisitString generates Python code for a string schema.
func (g *Generator) VisitString(schema core.StringSchema) error {
	// Check if this is an enum
	if enumValues := schema.EnumValues(); len(enumValues) > 0 && g.options.UseEnums {
		return g.generateStringEnum(schema)
	}

	// For standalone string schemas, generate based on output style
	metadata := schema.Metadata()
	className := g.typeMapper.FormatClassName(metadata.Name)

	// Create a field representing this string type
	field := Field{
		Name:         "value",
		Type:         g.typeMapper.MapSchemaType(core.TypeString),
		Required:     true,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		DefaultValue: schema.DefaultValue(),
	}

	// Add class docstring if enabled
	if g.options.IncludeDocstrings && metadata.Description != "" {
		docLines := g.formatter.FormatDocstring(metadata.Description, metadata.Examples, nil)
		for _, line := range docLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate the model based on output style
	var modelLines []string
	switch g.options.OutputStyle {
	case "pydantic":
		modelLines = g.formatter.FormatPydanticModel(className, []Field{field}, g.options.BaseClass)
	case "dataclass":
		modelLines = g.formatter.FormatDataclass(className, []Field{field}, g.options.BaseClass)
	case "class":
		modelLines = g.generatePlainClass(className, []Field{field}, g.options.BaseClass)
	case "namedtuple":
		modelLines = g.generateNamedTuple(className, []Field{field})
	default:
		// Fallback to type alias for unknown styles
		if g.options.IncludeComments {
			g.writeSchemaComments(metadata.Description, metadata.Examples, nil)
		}
		pythonType := g.typeMapper.MapSchemaType(core.TypeString)
		g.output.WriteString(fmt.Sprintf("%s = %s\n", className, pythonType))
		return nil
	}

	for _, line := range modelLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// VisitInteger generates Python code for an integer schema.
func (g *Generator) VisitInteger(schema core.IntegerSchema) error {
	// For standalone integer schemas, generate based on output style
	metadata := schema.Metadata()
	className := g.typeMapper.FormatClassName(metadata.Name)

	// Create a field representing this integer type
	field := Field{
		Name:         "value",
		Type:         g.typeMapper.MapSchemaType(core.TypeInteger),
		Required:     true,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		DefaultValue: nil, // Integer schemas don't have a DefaultValue method in the interface
	}

	// Add class docstring if enabled
	if g.options.IncludeDocstrings && metadata.Description != "" {
		docLines := g.formatter.FormatDocstring(metadata.Description, metadata.Examples, nil)
		for _, line := range docLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate the model based on output style
	var modelLines []string
	switch g.options.OutputStyle {
	case "pydantic":
		modelLines = g.formatter.FormatPydanticModel(className, []Field{field}, g.options.BaseClass)
	case "dataclass":
		modelLines = g.formatter.FormatDataclass(className, []Field{field}, g.options.BaseClass)
	case "class":
		modelLines = g.generatePlainClass(className, []Field{field}, g.options.BaseClass)
	case "namedtuple":
		modelLines = g.generateNamedTuple(className, []Field{field})
	default:
		// Fallback to type alias for unknown styles
		if g.options.IncludeComments {
			g.writeSchemaComments(metadata.Description, metadata.Examples, nil)
		}
		pythonType := g.typeMapper.MapSchemaType(core.TypeInteger)
		g.output.WriteString(fmt.Sprintf("%s = %s\n", className, pythonType))
		return nil
	}

	for _, line := range modelLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// VisitNumber generates Python code for a number schema.
func (g *Generator) VisitNumber(schema core.NumberSchema) error {
	// For standalone number schemas, generate a type alias
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatClassName(metadata.Name)
	pythonType := g.typeMapper.MapSchemaType(core.TypeNumber)

	// Add constraints as comments if enabled
	if g.options.IncludeComments {
		g.writeSchemaComments(metadata.Description, metadata.Examples, nil)
	}

	g.output.WriteString(fmt.Sprintf("%s = %s\n", typeName, pythonType))
	return nil
}

// VisitBoolean generates Python code for a boolean schema.
func (g *Generator) VisitBoolean(schema core.BooleanSchema) error {
	// For standalone boolean schemas, generate based on output style
	metadata := schema.Metadata()
	className := g.typeMapper.FormatClassName(metadata.Name)

	// Create a field representing this boolean type
	field := Field{
		Name:         "value",
		Type:         g.typeMapper.MapSchemaType(core.TypeBoolean),
		Required:     true,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		DefaultValue: nil, // Boolean schemas don't have a DefaultValue method in the interface
	}

	// Add class docstring if enabled
	if g.options.IncludeDocstrings && metadata.Description != "" {
		docLines := g.formatter.FormatDocstring(metadata.Description, metadata.Examples, nil)
		for _, line := range docLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate the model based on output style
	var modelLines []string
	switch g.options.OutputStyle {
	case "pydantic":
		modelLines = g.formatter.FormatPydanticModel(className, []Field{field}, g.options.BaseClass)
	case "dataclass":
		modelLines = g.formatter.FormatDataclass(className, []Field{field}, g.options.BaseClass)
	case "class":
		modelLines = g.generatePlainClass(className, []Field{field}, g.options.BaseClass)
	case "namedtuple":
		modelLines = g.generateNamedTuple(className, []Field{field})
	default:
		// Fallback to type alias for unknown styles
		if g.options.IncludeComments {
			g.writeSchemaComments(metadata.Description, metadata.Examples, nil)
		}
		pythonType := g.typeMapper.MapSchemaType(core.TypeBoolean)
		g.output.WriteString(fmt.Sprintf("%s = %s\n", className, pythonType))
		return nil
	}

	for _, line := range modelLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// VisitArray generates Python code for an array schema.
func (g *Generator) VisitArray(schema core.ArraySchema) error {
	metadata := schema.Metadata()
	className := g.typeMapper.FormatClassName(metadata.Name)

	// Determine element type
	elementType := "Any"
	if itemSchema := schema.ItemSchema(); itemSchema != nil {
		elementType = g.getSchemaTypeName(itemSchema)
	}

	pythonType := g.typeMapper.FormatListType(elementType)

	// Create a field representing this array type
	field := Field{
		Name:         "value",
		Type:         pythonType,
		Required:     true,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		DefaultValue: nil,
	}

	// Add class docstring if enabled
	if g.options.IncludeDocstrings && metadata.Description != "" {
		docLines := g.formatter.FormatDocstring(metadata.Description, metadata.Examples, nil)
		for _, line := range docLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate the model based on output style
	var modelLines []string
	switch g.options.OutputStyle {
	case "pydantic":
		modelLines = g.formatter.FormatPydanticModel(className, []Field{field}, g.options.BaseClass)
	case "dataclass":
		modelLines = g.formatter.FormatDataclass(className, []Field{field}, g.options.BaseClass)
	case "class":
		modelLines = g.generatePlainClass(className, []Field{field}, g.options.BaseClass)
	case "namedtuple":
		modelLines = g.generateNamedTuple(className, []Field{field})
	default:
		// Fallback to type alias for unknown styles
		if g.options.IncludeComments {
			g.writeSchemaComments(metadata.Description, metadata.Examples, nil)
		}
		g.output.WriteString(fmt.Sprintf("%s = %s\n", className, pythonType))
		return nil
	}

	for _, line := range modelLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// VisitObject generates Python code for an object schema.
func (g *Generator) VisitObject(schema core.ObjectSchema) error {
	return g.generateObjectModel(schema)
}

// generateObjectModel generates a Python model for an object schema.
func (g *Generator) generateObjectModel(schema core.ObjectSchema) error {
	metadata := schema.Metadata()
	className := g.typeMapper.FormatClassName(metadata.Name)

	// Convert properties to fields
	var fields []Field
	properties := schema.Properties()
	required := schema.Required()

	for propName, propSchema := range properties {
		field := Field{
			Name:         g.typeMapper.FormatFieldName(propName),
			Type:         g.getSchemaTypeName(propSchema),
			Required:     g.isRequired(propName, required),
			Description:  g.getSchemaDescription(propSchema),
			Examples:     g.getSchemaExamples(propSchema),
			DefaultValue: g.getSchemaDefault(propSchema),
		}

		// Handle optional fields
		if !field.Required {
			field.Type = g.typeMapper.FormatOptionalType(field.Type)
		}

		fields = append(fields, field)
	}

	// Add class docstring if enabled
	if g.options.IncludeDocstrings && metadata.Description != "" {
		docLines := g.formatter.FormatDocstring(metadata.Description, metadata.Examples, nil)
		for _, line := range docLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate the model based on output style
	var modelLines []string
	switch g.options.OutputStyle {
	case "pydantic":
		modelLines = g.formatter.FormatPydanticModel(className, fields, g.options.BaseClass)
	case "dataclass":
		modelLines = g.formatter.FormatDataclass(className, fields, g.options.BaseClass)
	case "class":
		modelLines = g.generatePlainClass(className, fields, g.options.BaseClass)
	case "namedtuple":
		modelLines = g.generateNamedTuple(className, fields)
	default:
		modelLines = g.formatter.FormatPydanticModel(className, fields, g.options.BaseClass)
	}

	for _, line := range modelLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateStringEnum generates a Python enum for string values.
func (g *Generator) generateStringEnum(schema core.StringSchema) error {
	metadata := schema.Metadata()
	enumName := g.typeMapper.FormatClassName(metadata.Name)

	// Add enum docstring if enabled
	if g.options.IncludeDocstrings && metadata.Description != "" {
		docLines := g.formatter.FormatDocstring(metadata.Description, metadata.Examples, nil)
		for _, line := range docLines {
			g.output.WriteString(line + "\n")
		}
	}

	enumValues := schema.EnumValues()
	enumLines := g.enumFormatter.FormatStringEnum(enumName, enumValues)
	for _, line := range enumLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generatePlainClass generates a plain Python class.
func (g *Generator) generatePlainClass(name string, fields []Field, baseClass string) []string {
	var lines []string

	// Class definition
	classLine := fmt.Sprintf("class %s", name)
	if baseClass != "" {
		classLine += fmt.Sprintf("(%s)", baseClass)
	}
	classLine += ":"
	lines = append(lines, classLine)

	g.context.PushIndent()

	// Constructor
	lines = append(lines, g.generateConstructor(fields)...)

	g.context.PopIndent()
	return lines
}

// generateConstructor generates a constructor for a plain class.
func (g *Generator) generateConstructor(fields []Field) []string {
	var lines []string

	// Method signature
	params := []string{"self"}
	for _, field := range fields {
		param := fmt.Sprintf("%s: %s", field.Name, field.Type)
		if !field.Required || field.DefaultValue != nil {
			if field.DefaultValue != nil {
				param += fmt.Sprintf(" = %v", field.DefaultValue)
			} else {
				param += " = None"
			}
		}
		params = append(params, param)
	}

	lines = append(lines, fmt.Sprintf("%sdef __init__(%s):", g.formatter.Indent(), strings.Join(params, ", ")))

	g.context.PushIndent()

	// Field assignments
	if len(fields) == 0 {
		lines = append(lines, fmt.Sprintf("%spass", g.formatter.Indent()))
	} else {
		for _, field := range fields {
			lines = append(lines, fmt.Sprintf("%sself.%s = %s", g.formatter.Indent(), field.Name, field.Name))
		}
	}

	g.context.PopIndent()
	return lines
}

// generateNamedTuple generates a named tuple.
func (g *Generator) generateNamedTuple(name string, fields []Field) []string {
	var lines []string

	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fieldNames[i] = fmt.Sprintf("'%s'", field.Name)
	}

	lines = append(lines, fmt.Sprintf("%s = namedtuple('%s', [%s])",
		name, name, strings.Join(fieldNames, ", ")))

	// Add import for namedtuple
	g.importManager.AddImport("from collections import namedtuple")

	return lines
}

// getSchemaTypeName returns the Python type name for a schema.
func (g *Generator) getSchemaTypeName(schema core.Schema) string {
	switch s := schema.(type) {
	case core.StringSchema:
		enumValues := s.EnumValues()
		if len(enumValues) > 0 && g.options.UseEnums {
			metadata := s.Metadata()
			return g.typeMapper.FormatClassName(metadata.Name)
		}
		return g.typeMapper.MapSchemaType(core.TypeString)
	case core.IntegerSchema:
		return g.typeMapper.MapSchemaType(core.TypeInteger)
	case core.NumberSchema:
		return g.typeMapper.MapSchemaType(core.TypeNumber)
	case core.BooleanSchema:
		return g.typeMapper.MapSchemaType(core.TypeBoolean)
	case core.ArraySchema:
		elementType := "Any"
		if itemSchema := s.ItemSchema(); itemSchema != nil {
			elementType = g.getSchemaTypeName(itemSchema)
		}
		return g.typeMapper.FormatListType(elementType)
	case core.ObjectSchema:
		metadata := s.Metadata()
		return g.typeMapper.FormatClassName(metadata.Name)
	default:
		return "Any"
	}
}

// getSchemaDescription returns the description of a schema.
func (g *Generator) getSchemaDescription(schema core.Schema) string {
	metadata := schema.Metadata()
	return metadata.Description
}

// getSchemaExamples returns the examples of a schema.
func (g *Generator) getSchemaExamples(schema core.Schema) []any {
	metadata := schema.Metadata()
	return metadata.Examples
}

// getSchemaDefault returns the default value of a schema.
func (g *Generator) getSchemaDefault(schema core.Schema) any {
	// Default values are not available through the metadata interface
	// This would need to be handled differently based on the specific schema type
	return nil
}

// isRequired checks if a property is required.
func (g *Generator) isRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}

// writeSchemaComments writes schema information as comments.
func (g *Generator) writeSchemaComments(description string, examples []any, defaultValue any) {
	if description != "" {
		g.output.WriteString(fmt.Sprintf("# %s\n", description))
	}

	if g.options.IncludeExamples && len(examples) > 0 {
		g.output.WriteString("# Examples: ")
		exampleStrs := make([]string, len(examples))
		for i, example := range examples {
			exampleStrs[i] = fmt.Sprintf("%v", example)
		}
		g.output.WriteString(strings.Join(exampleStrs, ", ") + "\n")
	}

	if g.options.IncludeDefaults && defaultValue != nil {
		g.output.WriteString(fmt.Sprintf("# Default: %v\n", defaultValue))
	}
}

// Name returns the name of the generator.
func (g *Generator) Name() string {
	return "python"
}

// Format returns the output format identifier.
func (g *Generator) Format() string {
	return "python"
}

// GetName returns the name of the generator (deprecated, use Name).
func (g *Generator) GetName() string {
	return g.Name()
}

// GetDescription returns the description of the generator.
func (g *Generator) GetDescription() string {
	return "Generates Python code from schema definitions"
}

// GetSupportedFormats returns the supported output formats.
func (g *Generator) GetSupportedFormats() []string {
	return []string{"pydantic", "dataclass", "class", "namedtuple"}
}

// VisitFunction generates Python code for a function schema.
func (g *Generator) VisitFunction(schema core.FunctionSchema) error {
	// Function schemas are not directly supported in Python generation
	// This could be extended to generate function signatures or stubs
	return nil
}

// VisitService generates Python code for a service schema.
func (g *Generator) VisitService(schema core.ServiceSchema) error {
	// Service schemas are not directly supported in Python generation
	// This could be extended to generate service classes or interfaces
	return nil
}

// VisitUnion generates Python code for a union schema.
func (g *Generator) VisitUnion(schema core.UnionSchema) error {
	// Union schemas could be represented as Union types in Python
	// For now, we'll generate a type alias to Any
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatClassName(metadata.Name)

	// Add constraints as comments if enabled
	if g.options.IncludeComments {
		g.writeSchemaComments(metadata.Description, metadata.Examples, nil)
	}

	g.output.WriteString(fmt.Sprintf("%s = Any  # Union type\n", typeName))
	return nil
}

// Clone creates a copy of the generator with new options.
func (g *Generator) Clone(options map[string]any) (*Generator, error) {
	newOptions := g.options.Clone()

	// Apply new options
	for key, value := range options {
		newOptions.SetOption(key, value)
	}

	// Validate new options
	if err := newOptions.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return NewGenerator(newOptions), nil
}
