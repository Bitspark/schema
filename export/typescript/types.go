package typescript

import (
	"fmt"
	"regexp"
	"strings"

	"defs.dev/schema/api/core"
	"defs.dev/schema/export/base"
)

// TypeMapper handles mapping from schema types to TypeScript types.
type TypeMapper struct {
	options TypeScriptOptions
	context *base.GenerationContext
}

// NewTypeMapper creates a new TypeMapper with the given options.
func NewTypeMapper(options TypeScriptOptions, context *base.GenerationContext) *TypeMapper {
	return &TypeMapper{
		options: options,
		context: context,
	}
}

// MapSchemaType maps a schema type to a TypeScript type string.
func (tm *TypeMapper) MapSchemaType(schemaType core.SchemaType) string {
	switch schemaType {
	case core.TypeString:
		return "string"
	case core.TypeInteger, core.TypeNumber:
		return "number"
	case core.TypeBoolean:
		return "boolean"
	case core.TypeArray:
		return "unknown[]" // Will be refined by the generator
	case core.TypeStructure:
		return "object" // Will be refined by the generator
	case core.TypeNull:
		return "null"
	case core.TypeAny:
		if tm.options.UseUnknownType {
			return "unknown"
		}
		return "any"
	default:
		if tm.options.UseUnknownType {
			return "unknown"
		}
		return "any"
	}
}

// FormatArrayType formats an array type according to the configured style.
func (tm *TypeMapper) FormatArrayType(elementType string) string {
	switch tm.options.ArrayStyle {
	case "Array<T>":
		return fmt.Sprintf("Array<%s>", elementType)
	case "T[]":
		return fmt.Sprintf("%s[]", elementType)
	default:
		return fmt.Sprintf("%s[]", elementType)
	}
}

// FormatOptionalProperty formats a property as optional if needed.
func (tm *TypeMapper) FormatOptionalProperty(propertyType string, isRequired bool) string {
	if isRequired || !tm.options.UseOptionalProperties {
		return propertyType
	}
	return propertyType // The '?' will be added at the property level, not type level
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
	case "snake_case":
		return base.ToSnakeCase(name)
	case "kebab-case":
		return base.ToKebabCase(name)
	default:
		return base.ToPascalCase(name)
	}
}

// FormatPropertyName formats a property name according to the naming convention.
func (tm *TypeMapper) FormatPropertyName(name string) string {
	if name == "" {
		return "unnamed"
	}

	switch tm.options.NamingConvention {
	case "PascalCase":
		return base.ToCamelCase(name) // Properties typically use camelCase even when types use PascalCase
	case "camelCase":
		return base.ToCamelCase(name)
	case "snake_case":
		return base.ToSnakeCase(name)
	case "kebab-case":
		return base.ToKebabCase(name)
	default:
		return base.ToCamelCase(name)
	}
}

// CodeFormatter handles TypeScript code formatting.
type CodeFormatter struct {
	options TypeScriptOptions
	context *base.GenerationContext
}

// NewCodeFormatter creates a new CodeFormatter with the given options.
func NewCodeFormatter(options TypeScriptOptions, context *base.GenerationContext) *CodeFormatter {
	return &CodeFormatter{
		options: options,
		context: context,
	}
}

// Indent returns the current indentation string.
func (cf *CodeFormatter) Indent() string {
	if cf.options.UseTabsForIndentation {
		return strings.Repeat("\t", cf.context.IndentLevel)
	}
	return strings.Repeat(" ", cf.context.IndentLevel*cf.options.IndentSize)
}

// IndentBy returns indentation for a specific level.
func (cf *CodeFormatter) IndentBy(level int) string {
	if cf.options.UseTabsForIndentation {
		return strings.Repeat("\t", level)
	}
	return strings.Repeat(" ", level*cf.options.IndentSize)
}

// FormatComment formats a comment according to the JSDoc style.
func (cf *CodeFormatter) FormatComment(comment string, isMultiline bool) []string {
	if comment == "" {
		return nil
	}

	if !isMultiline {
		return []string{fmt.Sprintf("%s// %s", cf.Indent(), comment)}
	}

	lines := strings.Split(comment, "\n")
	result := []string{fmt.Sprintf("%s/**", cf.Indent())}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, fmt.Sprintf("%s * %s", cf.Indent(), line))
		} else {
			result = append(result, fmt.Sprintf("%s *", cf.Indent()))
		}
	}

	result = append(result, fmt.Sprintf("%s */", cf.Indent()))
	return result
}

// FormatJSDoc formats JSDoc documentation.
func (cf *CodeFormatter) FormatJSDoc(description string, examples []any, defaultValue any) []string {
	if !cf.options.IncludeJSDoc {
		return nil
	}

	var lines []string

	lines = append(lines, fmt.Sprintf("%s/**", cf.Indent()))

	if description != "" {
		descLines := strings.Split(description, "\n")
		for _, line := range descLines {
			line = strings.TrimSpace(line)
			if line != "" {
				lines = append(lines, fmt.Sprintf("%s * %s", cf.Indent(), line))
			}
		}
		lines = append(lines, fmt.Sprintf("%s *", cf.Indent()))
	}

	if cf.options.IncludeExamples && len(examples) > 0 {
		lines = append(lines, fmt.Sprintf("%s * @example", cf.Indent()))
		for _, example := range examples {
			lines = append(lines, fmt.Sprintf("%s * %v", cf.Indent(), example))
		}
		lines = append(lines, fmt.Sprintf("%s *", cf.Indent()))
	}

	if cf.options.IncludeDefaults && defaultValue != nil {
		lines = append(lines, fmt.Sprintf("%s * @default %v", cf.Indent(), defaultValue))
		lines = append(lines, fmt.Sprintf("%s *", cf.Indent()))
	}

	lines = append(lines, fmt.Sprintf("%s */", cf.Indent()))
	return lines
}

// FormatInterface formats a TypeScript interface.
func (cf *CodeFormatter) FormatInterface(name string, properties []Property, exported bool) []string {
	var lines []string

	exportKeyword := ""
	if exported && cf.options.ExportTypes {
		exportKeyword = "export "
	}

	lines = append(lines, fmt.Sprintf("%s%sinterface %s {", cf.Indent(), exportKeyword, name))

	cf.context.PushIndent()
	for _, prop := range properties {
		propLines := cf.FormatProperty(prop)
		lines = append(lines, propLines...)
	}
	cf.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", cf.Indent()))
	return lines
}

// FormatType formats a TypeScript type alias.
func (cf *CodeFormatter) FormatType(name string, typeDefinition string, exported bool) []string {
	var lines []string

	exportKeyword := ""
	if exported && cf.options.ExportTypes {
		exportKeyword = "export "
	}

	lines = append(lines, fmt.Sprintf("%s%stype %s = %s;", cf.Indent(), exportKeyword, name, typeDefinition))
	return lines
}

// FormatProperty formats a TypeScript property.
func (cf *CodeFormatter) FormatProperty(prop Property) []string {
	var lines []string

	// Add JSDoc if enabled
	if cf.options.IncludeJSDoc && (prop.Description != "" || len(prop.Examples) > 0 || prop.DefaultValue != nil) {
		jsdocLines := cf.FormatJSDoc(prop.Description, prop.Examples, prop.DefaultValue)
		lines = append(lines, jsdocLines...)
	}

	// Format property declaration
	optional := ""
	if !prop.Required && cf.options.UseOptionalProperties {
		optional = "?"
	}

	readonly := ""
	if cf.options.StrictMode && prop.ReadOnly {
		readonly = "readonly "
	}

	propLine := fmt.Sprintf("%s%s%s%s: %s;", cf.Indent(), readonly, prop.Name, optional, prop.Type)
	lines = append(lines, propLine)

	return lines
}

// Property represents a TypeScript property.
type Property struct {
	Name         string
	Type         string
	Required     bool
	ReadOnly     bool
	Description  string
	Examples     []any
	DefaultValue any
}

// EnumFormatter handles TypeScript enum generation.
type EnumFormatter struct {
	options TypeScriptOptions
	context *base.GenerationContext
}

// NewEnumFormatter creates a new EnumFormatter.
func NewEnumFormatter(options TypeScriptOptions, context *base.GenerationContext) *EnumFormatter {
	return &EnumFormatter{
		options: options,
		context: context,
	}
}

// FormatStringEnum formats a string enum.
func (ef *EnumFormatter) FormatStringEnum(name string, values []string, exported bool) []string {
	if !ef.options.UseEnums {
		// Generate union type instead
		return ef.formatUnionType(name, values, exported)
	}

	var lines []string

	exportKeyword := ""
	if exported && ef.options.ExportTypes {
		exportKeyword = "export "
	}

	lines = append(lines, fmt.Sprintf("%s%senum %s {", ef.Indent(), exportKeyword, name))

	ef.context.PushIndent()
	for i, value := range values {
		enumKey := ef.formatEnumKey(value)
		comma := ","
		if i == len(values)-1 {
			comma = ""
		}
		lines = append(lines, fmt.Sprintf("%s%s = \"%s\"%s", ef.Indent(), enumKey, value, comma))
	}
	ef.context.PopIndent()

	lines = append(lines, fmt.Sprintf("%s}", ef.Indent()))
	return lines
}

// formatUnionType formats a union type for enum values.
func (ef *EnumFormatter) formatUnionType(name string, values []string, exported bool) []string {
	var lines []string

	exportKeyword := ""
	if exported && ef.options.ExportTypes {
		exportKeyword = "export "
	}

	unionValues := make([]string, len(values))
	for i, value := range values {
		if ef.options.UseConstAssertions {
			unionValues[i] = fmt.Sprintf("\"%s\"", value)
		} else {
			unionValues[i] = fmt.Sprintf("\"%s\"", value)
		}
	}

	typeDefinition := strings.Join(unionValues, " | ")
	if ef.options.UseConstAssertions {
		typeDefinition += " as const"
	}

	lines = append(lines, fmt.Sprintf("%s%stype %s = %s;", ef.Indent(), exportKeyword, name, typeDefinition))
	return lines
}

// formatEnumKey formats an enum key from a string value.
func (ef *EnumFormatter) formatEnumKey(value string) string {
	// Convert to a valid TypeScript identifier
	key := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(value, "_")
	key = regexp.MustCompile(`^[0-9]`).ReplaceAllString(key, "_$0")

	if key == "" {
		key = "EMPTY"
	}

	// Apply naming convention
	switch ef.options.NamingConvention {
	case "PascalCase":
		return base.ToPascalCase(key)
	case "camelCase":
		return base.ToCamelCase(key)
	case "snake_case":
		return strings.ToUpper(base.ToSnakeCase(key))
	case "kebab-case":
		return strings.ToUpper(strings.ReplaceAll(base.ToKebabCase(key), "-", "_"))
	default:
		return strings.ToUpper(base.ToSnakeCase(key))
	}
}

// Indent returns the current indentation string.
func (ef *EnumFormatter) Indent() string {
	if ef.options.UseTabsForIndentation {
		return strings.Repeat("\t", ef.context.IndentLevel)
	}
	return strings.Repeat(" ", ef.context.IndentLevel*ef.options.IndentSize)
}
