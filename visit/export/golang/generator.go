package golang

import (
	"fmt"
	"strings"

	"defs.dev/schema/core"
	"defs.dev/schema/visit/export/base"
)

// Generator generates Go code from schemas using the visitor pattern.
type Generator struct {
	base.BaseVisitor
	options       GoOptions
	context       *base.GenerationContext
	typeMapper    *TypeMapper
	formatter     *CodeFormatter
	enumFormatter *EnumFormatter
	importManager *ImportManager
	output        strings.Builder
}

// NewGenerator creates a new Go generator with the given options.
func NewGenerator(options GoOptions) *Generator {
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

// Generate produces Go code from a schema by accepting it as a visitor.
func (g *Generator) Generate(schema core.Schema) ([]byte, error) {
	// Reset output
	g.output.Reset()

	// Validate options
	if err := g.options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Write file header
	g.writeFileHeader()

	// Write package declaration
	g.output.WriteString(fmt.Sprintf("package %s\n\n", g.options.PackageName))

	// Generate imports
	if g.options.IncludeImports {
		imports := g.generateImports()
		if imports != "" {
			g.output.WriteString(imports)
			g.output.WriteString("\n")
		}
	}

	// Generate code by accepting the schema
	if accepter, ok := schema.(core.Accepter); ok {
		if err := accepter.Accept(g); err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}
	} else {
		return nil, fmt.Errorf("schema does not implement Accepter interface")
	}

	return []byte(g.output.String()), nil
}

// writeFileHeader writes the file header comment if configured.
func (g *Generator) writeFileHeader() {
	if g.options.FileHeader != "" {
		headerLines := strings.Split(g.options.FileHeader, "\n")
		for _, line := range headerLines {
			g.output.WriteString(fmt.Sprintf("// %s\n", line))
		}
		g.output.WriteString("\n")
	}
}

// generateImports generates the import section.
func (g *Generator) generateImports() string {
	imports := g.importManager.GetRequiredImports()
	if len(imports) == 0 {
		return ""
	}

	var importSection strings.Builder

	switch g.options.ImportStyle {
	case "single":
		for _, imp := range imports {
			importSection.WriteString(fmt.Sprintf("import \"%s\"\n", imp))
		}
	case "grouped":
		importSection.WriteString("import (\n")
		for _, imp := range imports {
			importSection.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		importSection.WriteString(")\n")
	case "goimports":
		// Let goimports handle the formatting
		importSection.WriteString("import (\n")
		for _, imp := range imports {
			importSection.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		importSection.WriteString(")\n")
	default:
		importSection.WriteString("import (\n")
		for _, imp := range imports {
			importSection.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		importSection.WriteString(")\n")
	}

	return importSection.String()
}

// VisitString generates Go code for a string schema.
func (g *Generator) VisitString(schema core.StringSchema) error {
	// Check if this is an enum
	if enumValues := schema.EnumValues(); len(enumValues) > 0 && g.options.GenerateEnums {
		return g.generateStringEnum(schema)
	}

	// For standalone string schemas, generate based on output style
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	switch g.options.OutputStyle {
	case "struct":
		return g.generateStringStruct(schema, typeName)
	case "interface":
		return g.generateStringInterface(schema, typeName)
	case "type_alias":
		return g.generateStringTypeAlias(schema, typeName)
	default:
		return g.generateStringStruct(schema, typeName)
	}
}

// VisitInteger generates Go code for an integer schema.
func (g *Generator) VisitInteger(schema core.IntegerSchema) error {
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	switch g.options.OutputStyle {
	case "struct":
		return g.generateIntegerStruct(schema, typeName)
	case "interface":
		return g.generateIntegerInterface(schema, typeName)
	case "type_alias":
		return g.generateIntegerTypeAlias(schema, typeName)
	default:
		return g.generateIntegerStruct(schema, typeName)
	}
}

// VisitBoolean generates Go code for a boolean schema.
func (g *Generator) VisitBoolean(schema core.BooleanSchema) error {
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	switch g.options.OutputStyle {
	case "struct":
		return g.generateBooleanStruct(schema, typeName)
	case "interface":
		return g.generateBooleanInterface(schema, typeName)
	case "type_alias":
		return g.generateBooleanTypeAlias(schema, typeName)
	default:
		return g.generateBooleanStruct(schema, typeName)
	}
}

// VisitArray generates Go code for an array schema.
func (g *Generator) VisitArray(schema core.ArraySchema) error {
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	switch g.options.OutputStyle {
	case "struct":
		return g.generateArrayStruct(schema, typeName)
	case "interface":
		return g.generateArrayInterface(schema, typeName)
	case "type_alias":
		return g.generateArrayTypeAlias(schema, typeName)
	default:
		return g.generateArrayStruct(schema, typeName)
	}
}

// VisitObject generates Go code for an object schema.
func (g *Generator) VisitObject(schema core.ObjectSchema) error {
	return g.generateObjectStruct(schema)
}

// VisitNumber generates Go code for a number schema.
func (g *Generator) VisitNumber(schema core.NumberSchema) error {
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	switch g.options.OutputStyle {
	case "struct":
		return g.generateNumberStruct(schema, typeName)
	case "interface":
		return g.generateNumberInterface(schema, typeName)
	case "type_alias":
		return g.generateNumberTypeAlias(schema, typeName)
	default:
		return g.generateNumberStruct(schema, typeName)
	}
}

// generateStringStruct generates a struct for a string schema.
func (g *Generator) generateStringStruct(schema core.StringSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create field
	field := Field{
		Name:         "Value",
		Type:         g.typeMapper.MapSchemaType(core.TypeString),
		OriginalName: "value",
		Required:     false,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		JSONTag:      g.typeMapper.FormatJSONTag("value"),
	}

	// Add validation tag if needed
	if g.options.IncludeValidationTags {
		field.ValidationTag = g.generateStringValidationTag(schema)
	}

	// Generate struct
	return g.generateStruct(typeName, []Field{field}, metadata.Description)
}

// generateIntegerStruct generates a struct for an integer schema.
func (g *Generator) generateIntegerStruct(schema core.IntegerSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create field
	field := Field{
		Name:         "Value",
		Type:         g.typeMapper.MapSchemaType(core.TypeInteger),
		OriginalName: "value",
		Required:     false,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		JSONTag:      g.typeMapper.FormatJSONTag("value"),
	}

	// Add validation tag if needed
	if g.options.IncludeValidationTags {
		field.ValidationTag = g.generateIntegerValidationTag(schema)
	}

	// Generate struct
	return g.generateStruct(typeName, []Field{field}, metadata.Description)
}

// generateBooleanStruct generates a struct for a boolean schema.
func (g *Generator) generateBooleanStruct(schema core.BooleanSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create field
	field := Field{
		Name:         "Value",
		Type:         g.typeMapper.MapSchemaType(core.TypeBoolean),
		OriginalName: "value",
		Required:     false,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		JSONTag:      g.typeMapper.FormatJSONTag("value"),
	}

	// Generate struct
	return g.generateStruct(typeName, []Field{field}, metadata.Description)
}

// generateArrayStruct generates a struct for an array schema.
func (g *Generator) generateArrayStruct(schema core.ArraySchema, typeName string) error {
	metadata := schema.Metadata()

	// Determine element type
	elementType := "any"
	if itemSchema := schema.ItemSchema(); itemSchema != nil {
		elementType = g.getSchemaTypeName(itemSchema)
	}

	sliceType := g.typeMapper.FormatSliceType(elementType)

	// Create field
	field := Field{
		Name:         "Items",
		Type:         sliceType,
		OriginalName: "items",
		Required:     true,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		JSONTag:      g.typeMapper.FormatJSONTag("items"),
	}

	// Generate struct
	return g.generateStruct(typeName, []Field{field}, metadata.Description)
}

// generateObjectStruct generates a struct for an object schema.
func (g *Generator) generateObjectStruct(schema core.ObjectSchema) error {
	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	// Convert properties to fields
	var fields []Field
	properties := schema.Properties()
	required := schema.Required()

	for propName, propSchema := range properties {
		field := Field{
			Name:         g.typeMapper.FormatFieldName(propName),
			Type:         g.getSchemaTypeName(propSchema),
			OriginalName: propName,
			Required:     g.isRequired(propName, required),
			Description:  g.getSchemaDescription(propSchema),
			Examples:     g.getSchemaExamples(propSchema),
			JSONTag:      g.typeMapper.FormatJSONTag(propName),
		}

		// Handle optional fields with pointers
		if !field.Required && g.options.UsePointers {
			field.Type = g.typeMapper.FormatPointerType(field.Type)
		}

		// Add validation tag if needed
		if g.options.IncludeValidationTags {
			field.ValidationTag = g.generateFieldValidationTag(propSchema, field.Required)
		}

		fields = append(fields, field)
	}

	// Generate struct
	return g.generateStruct(typeName, fields, metadata.Description)
}

// generateStruct generates a struct with the given fields.
func (g *Generator) generateStruct(name string, fields []Field, description string) error {
	// Add struct comment
	if g.options.IncludeComments && description != "" {
		commentLines := g.formatter.FormatComment(description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate struct
	structLines := g.formatter.FormatStruct(name, fields)
	for _, line := range structLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")

	// Generate additional methods if requested
	if g.options.GenerateConstructors {
		g.generateConstructor(name, fields)
	}

	if g.options.GenerateValidators {
		g.generateValidator(name, fields)
	}

	if g.options.GenerateStringers {
		g.generateStringer(name)
	}

	if g.options.GenerateGetters {
		g.generateGetters(name, fields)
	}

	if g.options.GenerateSetters {
		g.generateSetters(name, fields)
	}

	return nil
}

// generateStringTypeAlias generates a type alias for a string schema.
func (g *Generator) generateStringTypeAlias(schema core.StringSchema, typeName string) error {
	metadata := schema.Metadata()

	// Add comment
	if g.options.IncludeComments && metadata.Description != "" {
		commentLines := g.formatter.FormatComment(metadata.Description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate type alias
	aliasLines := g.formatter.FormatTypeAlias(typeName, g.typeMapper.MapSchemaType(core.TypeString))
	for _, line := range aliasLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateIntegerTypeAlias generates a type alias for an integer schema.
func (g *Generator) generateIntegerTypeAlias(schema core.IntegerSchema, typeName string) error {
	metadata := schema.Metadata()

	// Add comment
	if g.options.IncludeComments && metadata.Description != "" {
		commentLines := g.formatter.FormatComment(metadata.Description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate type alias
	aliasLines := g.formatter.FormatTypeAlias(typeName, g.typeMapper.MapSchemaType(core.TypeInteger))
	for _, line := range aliasLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateBooleanTypeAlias generates a type alias for a boolean schema.
func (g *Generator) generateBooleanTypeAlias(schema core.BooleanSchema, typeName string) error {
	metadata := schema.Metadata()

	// Add comment
	if g.options.IncludeComments && metadata.Description != "" {
		commentLines := g.formatter.FormatComment(metadata.Description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate type alias
	aliasLines := g.formatter.FormatTypeAlias(typeName, g.typeMapper.MapSchemaType(core.TypeBoolean))
	for _, line := range aliasLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateArrayTypeAlias generates a type alias for an array schema.
func (g *Generator) generateArrayTypeAlias(schema core.ArraySchema, typeName string) error {
	metadata := schema.Metadata()

	// Determine element type
	elementType := "any"
	if itemSchema := schema.ItemSchema(); itemSchema != nil {
		elementType = g.getSchemaTypeName(itemSchema)
	}

	sliceType := g.typeMapper.FormatSliceType(elementType)

	// Add comment
	if g.options.IncludeComments && metadata.Description != "" {
		commentLines := g.formatter.FormatComment(metadata.Description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate type alias
	aliasLines := g.formatter.FormatTypeAlias(typeName, sliceType)
	for _, line := range aliasLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateStringInterface generates an interface for a string schema.
func (g *Generator) generateStringInterface(schema core.StringSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create methods
	methods := []Method{
		{
			Name:        "GetValue",
			Parameters:  "",
			ReturnType:  "string",
			Description: "GetValue returns the string value",
		},
		{
			Name:        "SetValue",
			Parameters:  "value string",
			ReturnType:  "",
			Description: "SetValue sets the string value",
		},
	}

	return g.generateInterface(typeName, methods, metadata.Description)
}

// generateIntegerInterface generates an interface for an integer schema.
func (g *Generator) generateIntegerInterface(schema core.IntegerSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create methods
	methods := []Method{
		{
			Name:        "GetValue",
			Parameters:  "",
			ReturnType:  "int64",
			Description: "GetValue returns the integer value",
		},
		{
			Name:        "SetValue",
			Parameters:  "value int64",
			ReturnType:  "",
			Description: "SetValue sets the integer value",
		},
	}

	return g.generateInterface(typeName, methods, metadata.Description)
}

// generateBooleanInterface generates an interface for a boolean schema.
func (g *Generator) generateBooleanInterface(schema core.BooleanSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create methods
	methods := []Method{
		{
			Name:        "GetValue",
			Parameters:  "",
			ReturnType:  "bool",
			Description: "GetValue returns the boolean value",
		},
		{
			Name:        "SetValue",
			Parameters:  "value bool",
			ReturnType:  "",
			Description: "SetValue sets the boolean value",
		},
	}

	return g.generateInterface(typeName, methods, metadata.Description)
}

// generateArrayInterface generates an interface for an array schema.
func (g *Generator) generateArrayInterface(schema core.ArraySchema, typeName string) error {
	metadata := schema.Metadata()

	// Determine element type
	elementType := "any"
	if itemSchema := schema.ItemSchema(); itemSchema != nil {
		elementType = g.getSchemaTypeName(itemSchema)
	}

	sliceType := g.typeMapper.FormatSliceType(elementType)

	// Create methods
	methods := []Method{
		{
			Name:        "GetItems",
			Parameters:  "",
			ReturnType:  sliceType,
			Description: "GetItems returns the array items",
		},
		{
			Name:        "SetItems",
			Parameters:  fmt.Sprintf("items %s", sliceType),
			ReturnType:  "",
			Description: "SetItems sets the array items",
		},
		{
			Name:        "Len",
			Parameters:  "",
			ReturnType:  "int",
			Description: "Len returns the number of items",
		},
	}

	return g.generateInterface(typeName, methods, metadata.Description)
}

// generateInterface generates an interface with the given methods.
func (g *Generator) generateInterface(name string, methods []Method, description string) error {
	// Add interface comment
	if g.options.IncludeComments && description != "" {
		commentLines := g.formatter.FormatComment(description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate interface
	interfaceLines := g.formatter.FormatInterface(name, methods)
	for _, line := range interfaceLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateStringEnum generates a Go enum for string values.
func (g *Generator) generateStringEnum(schema core.StringSchema) error {
	metadata := schema.Metadata()
	enumName := g.typeMapper.FormatTypeName(metadata.Name)

	// Add enum comment
	if g.options.IncludeComments && metadata.Description != "" {
		commentLines := g.formatter.FormatComment(metadata.Description)
		for _, line := range commentLines {
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

// Helper methods

// getSchemaTypeName returns the Go type name for a schema.
func (g *Generator) getSchemaTypeName(schema core.Schema) string {
	metadata := schema.Metadata()
	if metadata.Name != "" {
		return g.typeMapper.FormatTypeName(metadata.Name)
	}

	// Fallback to basic type mapping
	return g.typeMapper.MapSchemaType(schema.Type())
}

// getSchemaDescription returns the description of a schema.
func (g *Generator) getSchemaDescription(schema core.Schema) string {
	return schema.Metadata().Description
}

// getSchemaExamples returns the examples of a schema.
func (g *Generator) getSchemaExamples(schema core.Schema) []any {
	return schema.Metadata().Examples
}

// isRequired checks if a property name is in the required list.
func (g *Generator) isRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}

// generateStringValidationTag generates validation tags for string schemas.
func (g *Generator) generateStringValidationTag(schema core.StringSchema) string {
	var validations []string

	if minLen := schema.MinLength(); minLen != nil {
		validations = append(validations, fmt.Sprintf("min=%d", *minLen))
	}

	if maxLen := schema.MaxLength(); maxLen != nil {
		validations = append(validations, fmt.Sprintf("max=%d", *maxLen))
	}

	if pattern := schema.Pattern(); pattern != "" {
		validations = append(validations, fmt.Sprintf("regexp=%s", pattern))
	}

	return strings.Join(validations, ",")
}

// generateIntegerValidationTag generates validation tags for integer schemas.
func (g *Generator) generateIntegerValidationTag(schema core.IntegerSchema) string {
	var validations []string

	if min := schema.Minimum(); min != nil {
		validations = append(validations, fmt.Sprintf("min=%d", *min))
	}

	if max := schema.Maximum(); max != nil {
		validations = append(validations, fmt.Sprintf("max=%d", *max))
	}

	return strings.Join(validations, ",")
}

// generateFieldValidationTag generates validation tags for object fields.
func (g *Generator) generateFieldValidationTag(schema core.Schema, required bool) string {
	var validations []string

	if required {
		validations = append(validations, "required")
	}

	// Add type-specific validations
	switch s := schema.(type) {
	case core.StringSchema:
		if tag := g.generateStringValidationTag(s); tag != "" {
			validations = append(validations, tag)
		}
	case core.IntegerSchema:
		if tag := g.generateIntegerValidationTag(s); tag != "" {
			validations = append(validations, tag)
		}
	}

	return strings.Join(validations, ",")
}

// Method generation helpers

// generateConstructor generates a constructor function.
func (g *Generator) generateConstructor(typeName string, fields []Field) {
	constructorName := fmt.Sprintf("%s%s", g.options.ConstructorPrefix, typeName)

	// Generate constructor signature
	var params []string
	for _, field := range fields {
		if field.Required {
			params = append(params, fmt.Sprintf("%s %s", strings.ToLower(field.Name), field.Type))
		}
	}

	g.output.WriteString(fmt.Sprintf("func %s(%s) *%s {\n", constructorName, strings.Join(params, ", "), typeName))
	g.output.WriteString(fmt.Sprintf("\treturn &%s{\n", typeName))

	for _, field := range fields {
		if field.Required {
			g.output.WriteString(fmt.Sprintf("\t\t%s: %s,\n", field.Name, strings.ToLower(field.Name)))
		}
	}

	g.output.WriteString("\t}\n")
	g.output.WriteString("}\n\n")
}

// generateValidator generates a validation method.
func (g *Generator) generateValidator(typeName string, fields []Field) {
	validatorName := fmt.Sprintf("%s%s", g.options.ValidatorPrefix, typeName)

	g.output.WriteString(fmt.Sprintf("func (v *%s) %s() error {\n", typeName, validatorName))
	g.output.WriteString("\t// TODO: Add validation logic\n")
	g.output.WriteString("\treturn nil\n")
	g.output.WriteString("}\n\n")
}

// generateStringer generates a String method.
func (g *Generator) generateStringer(typeName string) {
	g.output.WriteString(fmt.Sprintf("func (s *%s) String() string {\n", typeName))
	g.output.WriteString("\t// TODO: Add string representation\n")
	g.output.WriteString(fmt.Sprintf("\treturn \"%s{}\"\n", typeName))
	g.output.WriteString("}\n\n")
}

// generateGetters generates getter methods.
func (g *Generator) generateGetters(typeName string, fields []Field) {
	for _, field := range fields {
		getterName := fmt.Sprintf("Get%s", field.Name)
		g.output.WriteString(fmt.Sprintf("func (g *%s) %s() %s {\n", typeName, getterName, field.Type))
		g.output.WriteString(fmt.Sprintf("\treturn g.%s\n", field.Name))
		g.output.WriteString("}\n\n")
	}
}

// generateSetters generates setter methods.
func (g *Generator) generateSetters(typeName string, fields []Field) {
	for _, field := range fields {
		setterName := fmt.Sprintf("Set%s", field.Name)
		paramName := strings.ToLower(field.Name)
		g.output.WriteString(fmt.Sprintf("func (s *%s) %s(%s %s) {\n", typeName, setterName, paramName, field.Type))
		g.output.WriteString(fmt.Sprintf("\ts.%s = %s\n", field.Name, paramName))
		g.output.WriteString("}\n\n")
	}
}

// Interface compliance methods

// Name returns the human-readable name of the generator.
func (g *Generator) Name() string {
	return "Go Generator"
}

// Format returns the output format identifier.
func (g *Generator) Format() string {
	return "go"
}

// GetName returns the generator name (for compatibility).
func (g *Generator) GetName() string {
	return g.Name()
}

// GetDescription returns the generator description.
func (g *Generator) GetDescription() string {
	return "Generates Go structs, interfaces, and type aliases from schemas"
}

// GetSupportedFormats returns the supported output formats.
func (g *Generator) GetSupportedFormats() []string {
	return []string{"go", "golang"}
}

// Visitor pattern methods for unsupported schema types

// VisitFunction handles function schemas (not supported in Go generator).
func (g *Generator) VisitFunction(schema core.FunctionSchema) error {
	return fmt.Errorf("function schemas are not supported by the Go generator")
}

// VisitService handles service schemas (not supported in Go generator).
func (g *Generator) VisitService(schema core.ServiceSchema) error {
	return fmt.Errorf("service schemas are not supported by the Go generator")
}

// VisitUnion handles union schemas.
func (g *Generator) VisitUnion(schema core.UnionSchema) error {
	if !g.options.GenerateUnions {
		return fmt.Errorf("union schemas are disabled in options")
	}

	metadata := schema.Metadata()
	typeName := g.typeMapper.FormatTypeName(metadata.Name)

	switch g.options.UnionStyle {
	case "interface":
		return g.generateUnionInterface(schema, typeName)
	case "embedded":
		return g.generateUnionEmbedded(schema, typeName)
	case "discriminated":
		return g.generateUnionDiscriminated(schema, typeName)
	default:
		return g.generateUnionInterface(schema, typeName)
	}
}

// generateUnionInterface generates a union as an interface.
func (g *Generator) generateUnionInterface(schema core.UnionSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create a marker method for the union
	methods := []Method{
		{
			Name:        fmt.Sprintf("Is%s", typeName),
			Parameters:  "",
			ReturnType:  "",
			Description: fmt.Sprintf("Is%s is a marker method for the %s union", typeName, typeName),
		},
	}

	return g.generateInterface(typeName, methods, metadata.Description)
}

// generateUnionEmbedded generates a union using embedded structs.
func (g *Generator) generateUnionEmbedded(schema core.UnionSchema, typeName string) error {
	// TODO: Implement embedded union generation
	return fmt.Errorf("embedded union style not yet implemented")
}

// generateUnionDiscriminated generates a union with a discriminator field.
func (g *Generator) generateUnionDiscriminated(schema core.UnionSchema, typeName string) error {
	// TODO: Implement discriminated union generation
	return fmt.Errorf("discriminated union style not yet implemented")
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

// generateNumberStruct generates a struct for a number schema.
func (g *Generator) generateNumberStruct(schema core.NumberSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create field
	field := Field{
		Name:         "Value",
		Type:         g.typeMapper.MapSchemaType(core.TypeNumber),
		OriginalName: "value",
		Required:     false,
		Description:  metadata.Description,
		Examples:     metadata.Examples,
		JSONTag:      g.typeMapper.FormatJSONTag("value"),
	}

	// Add validation tag if needed
	if g.options.IncludeValidationTags {
		field.ValidationTag = g.generateNumberValidationTag(schema)
	}

	// Generate struct
	return g.generateStruct(typeName, []Field{field}, metadata.Description)
}

// generateNumberInterface generates an interface for a number schema.
func (g *Generator) generateNumberInterface(schema core.NumberSchema, typeName string) error {
	metadata := schema.Metadata()

	// Create methods
	methods := []Method{
		{
			Name:        "GetValue",
			Parameters:  "",
			ReturnType:  "float64",
			Description: "GetValue returns the number value",
		},
		{
			Name:        "SetValue",
			Parameters:  "value float64",
			ReturnType:  "",
			Description: "SetValue sets the number value",
		},
	}

	return g.generateInterface(typeName, methods, metadata.Description)
}

// generateNumberTypeAlias generates a type alias for a number schema.
func (g *Generator) generateNumberTypeAlias(schema core.NumberSchema, typeName string) error {
	metadata := schema.Metadata()

	// Add comment
	if g.options.IncludeComments && metadata.Description != "" {
		commentLines := g.formatter.FormatComment(metadata.Description)
		for _, line := range commentLines {
			g.output.WriteString(line + "\n")
		}
	}

	// Generate type alias
	aliasLines := g.formatter.FormatTypeAlias(typeName, g.typeMapper.MapSchemaType(core.TypeNumber))
	for _, line := range aliasLines {
		g.output.WriteString(line + "\n")
	}

	g.output.WriteString("\n")
	return nil
}

// generateNumberValidationTag generates validation tags for number schemas.
func (g *Generator) generateNumberValidationTag(schema core.NumberSchema) string {
	var validations []string

	if min := schema.Minimum(); min != nil {
		validations = append(validations, fmt.Sprintf("min=%f", *min))
	}

	if max := schema.Maximum(); max != nil {
		validations = append(validations, fmt.Sprintf("max=%f", *max))
	}

	return strings.Join(validations, ",")
}
