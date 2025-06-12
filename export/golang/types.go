package golang

import (
	"fmt"
	"strings"

	"defs.dev/schema/api/core"
	"defs.dev/schema/export/base"
)

// TypeMapper handles Go type mapping and naming conventions.
type TypeMapper struct {
	options GoOptions
	context *base.GenerationContext
}

// NewTypeMapper creates a new TypeMapper with the given options.
func NewTypeMapper(options GoOptions, context *base.GenerationContext) *TypeMapper {
	return &TypeMapper{
		options: options,
		context: context,
	}
}

// MapSchemaType maps a schema type to a Go type.
func (tm *TypeMapper) MapSchemaType(schemaType core.SchemaType) string {
	// Check for custom type mappings first
	if customType, exists := tm.options.CustomTypeMappings[string(schemaType)]; exists {
		return customType
	}

	switch schemaType {
	case core.TypeString:
		return "string"
	case core.TypeInteger:
		return "int64"
	case core.TypeNumber:
		return "float64"
	case core.TypeBoolean:
		return "bool"
	case core.TypeArray:
		return "[]interface{}" // Will be specialized based on item type
	case core.TypeObject:
		return "interface{}" // Will be specialized to struct
	default:
		return "interface{}"
	}
}

// FormatSliceType formats a Go slice type.
func (tm *TypeMapper) FormatSliceType(elementType string) string {
	return fmt.Sprintf("[]%s", elementType)
}

// FormatMapType formats a Go map type.
func (tm *TypeMapper) FormatMapType(keyType, valueType string) string {
	return fmt.Sprintf("map[%s]%s", keyType, valueType)
}

// FormatPointerType formats a Go pointer type.
func (tm *TypeMapper) FormatPointerType(baseType string) string {
	return fmt.Sprintf("*%s", baseType)
}

// FormatTypeName formats a type name according to the naming convention.
func (tm *TypeMapper) FormatTypeName(name string) string {
	if name == "" {
		return "UnnamedType"
	}

	switch tm.options.NamingConvention {
	case "PascalCase":
		return base.ToPascalCase(name)
	case "camelCase":
		return base.ToCamelCase(name)
	default:
		return base.ToPascalCase(name)
	}
}

// FormatFieldName formats a field name according to the field naming convention.
func (tm *TypeMapper) FormatFieldName(name string) string {
	if name == "" {
		return "Unnamed"
	}

	switch tm.options.FieldNamingConvention {
	case "PascalCase":
		return base.ToPascalCase(name)
	case "camelCase":
		return base.ToCamelCase(name)
	default:
		return base.ToPascalCase(name)
	}
}

// FormatJSONTag formats a JSON tag according to the JSON tag style.
func (tm *TypeMapper) FormatJSONTag(name string) string {
	if name == "" {
		return ""
	}

	switch tm.options.JSONTagStyle {
	case "snake_case":
		return base.ToSnakeCase(name)
	case "camelCase":
		return base.ToCamelCase(name)
	case "kebab-case":
		return base.ToKebabCase(name)
	default:
		return base.ToSnakeCase(name)
	}
}

// isGoVersion118Plus checks if the target Go version is 1.18 or higher.
func (tm *TypeMapper) isGoVersion118Plus() bool {
	return tm.options.GoVersion >= "1.18"
}

// isGoVersion121Plus checks if the target Go version is 1.21 or higher.
func (tm *TypeMapper) isGoVersion121Plus() bool {
	return tm.options.GoVersion >= "1.21"
}

// CodeFormatter handles Go code formatting.
type CodeFormatter struct {
	options GoOptions
	context *base.GenerationContext
}

// NewCodeFormatter creates a new CodeFormatter with the given options.
func NewCodeFormatter(options GoOptions, context *base.GenerationContext) *CodeFormatter {
	return &CodeFormatter{
		options: options,
		context: context,
	}
}

// Indent returns the current indentation string.
func (cf *CodeFormatter) Indent() string {
	if cf.options.IndentStyle == "tabs" {
		return strings.Repeat("\t", cf.context.IndentLevel)
	}
	return strings.Repeat(" ", cf.context.IndentLevel*cf.options.IndentSize)
}

// IndentBy returns indentation for a specific level.
func (cf *CodeFormatter) IndentBy(level int) string {
	if cf.options.IndentStyle == "tabs" {
		return strings.Repeat("\t", level)
	}
	return strings.Repeat(" ", level*cf.options.IndentSize)
}

// FormatComment formats a comment according to Go conventions.
func (cf *CodeFormatter) FormatComment(comment string) []string {
	if !cf.options.IncludeComments || comment == "" {
		return nil
	}

	var lines []string
	commentLines := strings.Split(comment, "\n")

	for _, line := range commentLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, fmt.Sprintf("%s// %s", cf.Indent(), line))
		} else {
			lines = append(lines, fmt.Sprintf("%s//", cf.Indent()))
		}
	}

	return lines
}

// FormatStruct formats a Go struct definition.
func (cf *CodeFormatter) FormatStruct(name string, fields []Field) []string {
	var lines []string

	// Struct definition
	lines = append(lines, fmt.Sprintf("%stype %s struct {", cf.Indent(), name))

	cf.context.PushIndent()

	// Fields
	if len(fields) == 0 {
		// Empty struct
		lines = append(lines, fmt.Sprintf("%s// Empty struct", cf.Indent()))
	} else {
		for _, field := range fields {
			fieldLines := cf.FormatStructField(field)
			lines = append(lines, fieldLines...)
		}
	}

	cf.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", cf.Indent()))
	return lines
}

// FormatStructField formats a single struct field.
func (cf *CodeFormatter) FormatStructField(field Field) []string {
	var lines []string

	// Field comment
	if cf.options.IncludeComments && field.Description != "" {
		commentLines := cf.FormatComment(field.Description)
		lines = append(lines, commentLines...)
	}

	// Field definition
	fieldLine := fmt.Sprintf("%s%s %s", cf.Indent(), field.Name, field.Type)

	// Struct tags
	tags := cf.FormatStructTags(field)
	if tags != "" {
		fieldLine += fmt.Sprintf(" `%s`", tags)
	}

	lines = append(lines, fieldLine)
	return lines
}

// FormatStructTags formats struct tags for a field.
func (cf *CodeFormatter) FormatStructTags(field Field) string {
	var tags []string

	// JSON tag
	if cf.options.IncludeJSONTags {
		jsonTag := field.JSONTag
		if jsonTag == "" {
			jsonTag = field.OriginalName
		}

		jsonValue := jsonTag
		if cf.options.UseOmitEmpty && !field.Required {
			jsonValue += ",omitempty"
		}
		tags = append(tags, fmt.Sprintf(`json:"%s"`, jsonValue))
	}

	// XML tag
	if cf.options.IncludeXMLTags {
		xmlTag := field.XMLTag
		if xmlTag == "" {
			xmlTag = field.OriginalName
		}
		tags = append(tags, fmt.Sprintf(`xml:"%s"`, xmlTag))
	}

	// YAML tag
	if cf.options.IncludeYAMLTags {
		yamlTag := field.YAMLTag
		if yamlTag == "" {
			yamlTag = field.OriginalName
		}
		tags = append(tags, fmt.Sprintf(`yaml:"%s"`, yamlTag))
	}

	// Validation tag
	if cf.options.IncludeValidationTags && field.ValidationTag != "" {
		switch cf.options.ValidationTagStyle {
		case "go-playground":
			tags = append(tags, fmt.Sprintf(`validate:"%s"`, field.ValidationTag))
		case "ozzo":
			tags = append(tags, fmt.Sprintf(`validation:"%s"`, field.ValidationTag))
		case "custom":
			tags = append(tags, field.ValidationTag)
		}
	}

	// Custom struct tag options
	for key, value := range cf.options.StructTagOptions {
		tags = append(tags, fmt.Sprintf(`%s:"%s"`, key, value))
	}

	return strings.Join(tags, " ")
}

// FormatInterface formats a Go interface definition.
func (cf *CodeFormatter) FormatInterface(name string, methods []Method) []string {
	var lines []string

	// Interface definition
	lines = append(lines, fmt.Sprintf("%stype %s interface {", cf.Indent(), name))

	cf.context.PushIndent()

	// Methods
	if len(methods) == 0 {
		// Empty interface
		lines = append(lines, fmt.Sprintf("%s// Empty interface", cf.Indent()))
	} else {
		for _, method := range methods {
			methodLines := cf.FormatInterfaceMethod(method)
			lines = append(lines, methodLines...)
		}
	}

	cf.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", cf.Indent()))
	return lines
}

// FormatInterfaceMethod formats a single interface method.
func (cf *CodeFormatter) FormatInterfaceMethod(method Method) []string {
	var lines []string

	// Method comment
	if cf.options.IncludeComments && method.Description != "" {
		commentLines := cf.FormatComment(method.Description)
		lines = append(lines, commentLines...)
	}

	// Method signature
	methodLine := fmt.Sprintf("%s%s(%s)", cf.Indent(), method.Name, method.Parameters)
	if method.ReturnType != "" {
		methodLine += fmt.Sprintf(" %s", method.ReturnType)
	}

	lines = append(lines, methodLine)
	return lines
}

// FormatTypeAlias formats a Go type alias.
func (cf *CodeFormatter) FormatTypeAlias(name, targetType string) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%stype %s = %s", cf.Indent(), name, targetType))
	return lines
}

// Field represents a struct field.
type Field struct {
	Name          string
	Type          string
	OriginalName  string
	Required      bool
	Description   string
	Examples      []any
	DefaultValue  any
	JSONTag       string
	XMLTag        string
	YAMLTag       string
	ValidationTag string
}

// Method represents an interface method.
type Method struct {
	Name        string
	Parameters  string
	ReturnType  string
	Description string
}

// EnumFormatter handles Go enum formatting.
type EnumFormatter struct {
	options GoOptions
	context *base.GenerationContext
}

// NewEnumFormatter creates a new EnumFormatter with the given options.
func NewEnumFormatter(options GoOptions, context *base.GenerationContext) *EnumFormatter {
	return &EnumFormatter{
		options: options,
		context: context,
	}
}

// FormatStringEnum formats a string enum according to the enum style.
func (ef *EnumFormatter) FormatStringEnum(name string, values []string) []string {
	switch ef.options.EnumStyle {
	case "const":
		return ef.formatConstEnum(name, values)
	case "type":
		return ef.formatTypeEnum(name, values)
	case "string":
		return ef.formatStringTypeEnum(name, values)
	default:
		return ef.formatConstEnum(name, values)
	}
}

// formatConstEnum formats an enum using const declarations.
func (ef *EnumFormatter) formatConstEnum(name string, values []string) []string {
	var lines []string

	// Type definition
	lines = append(lines, fmt.Sprintf("%stype %s string", ef.Indent(), name))
	lines = append(lines, "")

	// Constants
	lines = append(lines, fmt.Sprintf("%sconst (", ef.Indent()))

	ef.context.PushIndent()
	for i, value := range values {
		constName := fmt.Sprintf("%s%s", name, base.ToPascalCase(value))
		if i == 0 {
			lines = append(lines, fmt.Sprintf("%s%s %s = %q", ef.Indent(), constName, name, value))
		} else {
			lines = append(lines, fmt.Sprintf("%s%s = %q", ef.Indent(), constName, value))
		}
	}
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s)", ef.Indent()))
	return lines
}

// formatTypeEnum formats an enum using a custom type.
func (ef *EnumFormatter) formatTypeEnum(name string, values []string) []string {
	var lines []string

	// Type definition
	lines = append(lines, fmt.Sprintf("%stype %s int", ef.Indent(), name))
	lines = append(lines, "")

	// Constants
	lines = append(lines, fmt.Sprintf("%sconst (", ef.Indent()))

	ef.context.PushIndent()
	for i, value := range values {
		constName := fmt.Sprintf("%s%s", name, base.ToPascalCase(value))
		if i == 0 {
			lines = append(lines, fmt.Sprintf("%s%s %s = iota", ef.Indent(), constName, name))
		} else {
			lines = append(lines, fmt.Sprintf("%s%s", ef.Indent(), constName))
		}
	}
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s)", ef.Indent()))

	// String method
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%sfunc (e %s) String() string {", ef.Indent(), name))

	ef.context.PushIndent()
	lines = append(lines, fmt.Sprintf("%sswitch e {", ef.Indent()))

	ef.context.PushIndent()
	for _, value := range values {
		constName := fmt.Sprintf("%s%s", name, base.ToPascalCase(value))
		lines = append(lines, fmt.Sprintf("%scase %s:", ef.Indent(), constName))
		lines = append(lines, fmt.Sprintf("%s\treturn %q", ef.Indent(), value))
	}
	lines = append(lines, fmt.Sprintf("%sdefault:", ef.Indent()))
	lines = append(lines, fmt.Sprintf("%s\treturn \"unknown\"", ef.Indent()))
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", ef.Indent()))
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", ef.Indent()))
	return lines
}

// formatStringTypeEnum formats an enum using string type with validation.
func (ef *EnumFormatter) formatStringTypeEnum(name string, values []string) []string {
	var lines []string

	// Type definition
	lines = append(lines, fmt.Sprintf("%stype %s string", ef.Indent(), name))
	lines = append(lines, "")

	// Validation method
	lines = append(lines, fmt.Sprintf("%sfunc (e %s) IsValid() bool {", ef.Indent(), name))

	ef.context.PushIndent()
	lines = append(lines, fmt.Sprintf("%sswitch e {", ef.Indent()))

	ef.context.PushIndent()
	for _, value := range values {
		lines = append(lines, fmt.Sprintf("%scase %q:", ef.Indent(), value))
	}
	lines = append(lines, fmt.Sprintf("%s\treturn true", ef.Indent()))
	lines = append(lines, fmt.Sprintf("%sdefault:", ef.Indent()))
	lines = append(lines, fmt.Sprintf("%s\treturn false", ef.Indent()))
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", ef.Indent()))
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", ef.Indent()))
	return lines
}

// Indent returns the current indentation string.
func (ef *EnumFormatter) Indent() string {
	if ef.options.IndentStyle == "tabs" {
		return strings.Repeat("\t", ef.context.IndentLevel)
	}
	return strings.Repeat(" ", ef.context.IndentLevel*ef.options.IndentSize)
}

// ImportManager handles Go import management.
type ImportManager struct {
	options GoOptions
	imports map[string]bool
}

// NewImportManager creates a new ImportManager with the given options.
func NewImportManager(options GoOptions) *ImportManager {
	return &ImportManager{
		options: options,
		imports: make(map[string]bool),
	}
}

// AddImport adds an import to the manager.
func (im *ImportManager) AddImport(importPath string) {
	if importPath != "" {
		im.imports[importPath] = true
	}
}

// GetRequiredImports returns the required imports based on the options.
func (im *ImportManager) GetRequiredImports() []string {
	var imports []string

	// Add standard library imports based on usage
	if im.options.IncludeJSONTags {
		im.AddImport("encoding/json")
	}

	if im.options.IncludeXMLTags {
		im.AddImport("encoding/xml")
	}

	if im.options.GenerateValidators {
		switch im.options.ValidationTagStyle {
		case "go-playground":
			im.AddImport("github.com/go-playground/validator/v10")
		case "ozzo":
			im.AddImport("github.com/go-ozzo/ozzo-validation/v4")
		}
	}

	// Add extra imports
	for _, extraImport := range im.options.ExtraImports {
		im.AddImport(extraImport)
	}

	// Convert map to sorted slice
	for importPath := range im.imports {
		imports = append(imports, importPath)
	}

	return imports
}
