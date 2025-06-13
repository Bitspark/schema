package python

import (
	"fmt"
	"regexp"
	"strings"

	"defs.dev/schema/core"
	"defs.dev/schema/visitors/export/base"
)

// TypeMapper handles mapping from schema types to Python types.
type TypeMapper struct {
	options PythonOptions
	context *base.GenerationContext
}

// NewTypeMapper creates a new TypeMapper with the given options.
func NewTypeMapper(options PythonOptions, context *base.GenerationContext) *TypeMapper {
	return &TypeMapper{
		options: options,
		context: context,
	}
}

// MapSchemaType maps a schema type to a Python type string.
func (tm *TypeMapper) MapSchemaType(schemaType core.SchemaType) string {
	// Check custom type mapping first
	if customType, exists := tm.options.CustomTypeMapping[string(schemaType)]; exists {
		return customType
	}

	switch schemaType {
	case core.TypeString:
		return "str"
	case core.TypeInteger:
		return "int"
	case core.TypeNumber:
		return "float"
	case core.TypeBoolean:
		return "bool"
	case core.TypeArray:
		if tm.options.TypeHintStyle == "builtin" && tm.isPython39Plus() {
			return "list"
		}
		return "List"
	case core.TypeStructure:
		if tm.options.TypeHintStyle == "builtin" && tm.isPython39Plus() {
			return "dict"
		}
		return "Dict"
	case core.TypeNull:
		return "None"
	case core.TypeAny:
		return "Any"
	default:
		return "Any"
	}
}

// FormatListType formats a list type according to the configured style.
func (tm *TypeMapper) FormatListType(elementType string) string {
	if tm.options.TypeHintStyle == "builtin" && tm.isPython39Plus() {
		return fmt.Sprintf("list[%s]", elementType)
	}
	return fmt.Sprintf("List[%s]", elementType)
}

// FormatDictType formats a dict type according to the configured style.
func (tm *TypeMapper) FormatDictType(keyType, valueType string) string {
	if tm.options.TypeHintStyle == "builtin" && tm.isPython39Plus() {
		return fmt.Sprintf("dict[%s, %s]", keyType, valueType)
	}
	return fmt.Sprintf("Dict[%s, %s]", keyType, valueType)
}

// FormatOptionalType formats an optional type according to the configured style.
func (tm *TypeMapper) FormatOptionalType(baseType string) string {
	if tm.options.UseOptional {
		return fmt.Sprintf("Optional[%s]", baseType)
	}
	if tm.isPython310Plus() {
		return fmt.Sprintf("%s | None", baseType)
	}
	return fmt.Sprintf("Union[%s, None]", baseType)
}

// FormatClassName formats a class name according to the naming convention.
func (tm *TypeMapper) FormatClassName(name string) string {
	if name == "" {
		return "UnnamedModel"
	}

	switch tm.options.NamingConvention {
	case "PascalCase":
		return base.ToPascalCase(name)
	case "snake_case":
		return base.ToSnakeCase(name)
	default:
		return base.ToPascalCase(name)
	}
}

// FormatFieldName formats a field name according to the field naming convention.
func (tm *TypeMapper) FormatFieldName(name string) string {
	if name == "" {
		return "unnamed"
	}

	switch tm.options.FieldNamingConvention {
	case "snake_case":
		return base.ToSnakeCase(name)
	case "camelCase":
		return base.ToCamelCase(name)
	default:
		return base.ToSnakeCase(name)
	}
}

// isPython39Plus checks if the target Python version is 3.9 or higher.
func (tm *TypeMapper) isPython39Plus() bool {
	return tm.options.PythonVersion >= "3.9"
}

// isPython310Plus checks if the target Python version is 3.10 or higher.
func (tm *TypeMapper) isPython310Plus() bool {
	return tm.options.PythonVersion >= "3.10"
}

// CodeFormatter handles Python code formatting.
type CodeFormatter struct {
	options PythonOptions
	context *base.GenerationContext
}

// NewCodeFormatter creates a new CodeFormatter with the given options.
func NewCodeFormatter(options PythonOptions, context *base.GenerationContext) *CodeFormatter {
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

// FormatDocstring formats a docstring according to the configured style.
func (cf *CodeFormatter) FormatDocstring(description string, examples []any, defaultValue any) []string {
	if !cf.options.IncludeDocstrings || description == "" {
		return nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s\"\"\"", cf.Indent()))

	switch cf.options.DocstringStyle {
	case "google":
		lines = append(lines, cf.formatGoogleDocstring(description, examples, defaultValue)...)
	case "numpy":
		lines = append(lines, cf.formatNumpyDocstring(description, examples, defaultValue)...)
	case "sphinx":
		lines = append(lines, cf.formatSphinxDocstring(description, examples, defaultValue)...)
	default:
		lines = append(lines, cf.formatGoogleDocstring(description, examples, defaultValue)...)
	}

	lines = append(lines, fmt.Sprintf("%s\"\"\"", cf.Indent()))
	return lines
}

// formatGoogleDocstring formats a Google-style docstring.
func (cf *CodeFormatter) formatGoogleDocstring(description string, examples []any, defaultValue any) []string {
	var lines []string

	// Description
	descLines := strings.Split(description, "\n")
	for _, line := range descLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, fmt.Sprintf("%s%s", cf.Indent(), line))
		}
	}

	// Examples
	if cf.options.IncludeExamples && len(examples) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%sExamples:", cf.Indent()))
		for _, example := range examples {
			lines = append(lines, fmt.Sprintf("%s    %v", cf.Indent(), example))
		}
	}

	// Default value
	if cf.options.IncludeDefaults && defaultValue != nil {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%sDefault:", cf.Indent()))
		lines = append(lines, fmt.Sprintf("%s    %v", cf.Indent(), defaultValue))
	}

	return lines
}

// formatNumpyDocstring formats a NumPy-style docstring.
func (cf *CodeFormatter) formatNumpyDocstring(description string, examples []any, defaultValue any) []string {
	var lines []string

	// Description
	descLines := strings.Split(description, "\n")
	for _, line := range descLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, fmt.Sprintf("%s%s", cf.Indent(), line))
		}
	}

	// Examples
	if cf.options.IncludeExamples && len(examples) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%sExamples", cf.Indent()))
		lines = append(lines, fmt.Sprintf("%s--------", cf.Indent()))
		for _, example := range examples {
			lines = append(lines, fmt.Sprintf("%s%v", cf.Indent(), example))
		}
	}

	return lines
}

// formatSphinxDocstring formats a Sphinx-style docstring.
func (cf *CodeFormatter) formatSphinxDocstring(description string, examples []any, defaultValue any) []string {
	var lines []string

	// Description
	descLines := strings.Split(description, "\n")
	for _, line := range descLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, fmt.Sprintf("%s%s", cf.Indent(), line))
		}
	}

	// Examples
	if cf.options.IncludeExamples && len(examples) > 0 {
		lines = append(lines, "")
		for _, example := range examples {
			lines = append(lines, fmt.Sprintf("%s:example: %v", cf.Indent(), example))
		}
	}

	// Default value
	if cf.options.IncludeDefaults && defaultValue != nil {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%s:default: %v", cf.Indent(), defaultValue))
	}

	return lines
}

// FormatPydanticModel formats a Pydantic model.
func (cf *CodeFormatter) FormatPydanticModel(name string, fields []Field, baseClass string) []string {
	var lines []string

	// Class definition
	classLine := fmt.Sprintf("%sclass %s", cf.Indent(), name)
	if baseClass != "" {
		classLine += fmt.Sprintf("(%s)", baseClass)
	} else if cf.options.PydanticVersion == "v2" {
		classLine += "(BaseModel)"
	} else {
		classLine += "(BaseModel)"
	}
	classLine += ":"
	lines = append(lines, classLine)

	cf.context.PushIndent()

	// Fields
	if len(fields) == 0 {
		lines = append(lines, fmt.Sprintf("%spass", cf.Indent()))
	} else {
		for _, field := range fields {
			fieldLines := cf.FormatPydanticField(field)
			lines = append(lines, fieldLines...)
		}
	}

	cf.context.PopIndent()
	return lines
}

// FormatDataclass formats a dataclass.
func (cf *CodeFormatter) FormatDataclass(name string, fields []Field, baseClass string) []string {
	var lines []string

	// Decorator
	decorator := "@dataclass"
	if len(cf.options.DataclassOptions) > 0 {
		decorator += fmt.Sprintf("(%s)", strings.Join(cf.options.DataclassOptions, ", "))
	}
	lines = append(lines, fmt.Sprintf("%s%s", cf.Indent(), decorator))

	// Class definition
	classLine := fmt.Sprintf("%sclass %s", cf.Indent(), name)
	if baseClass != "" {
		classLine += fmt.Sprintf("(%s)", baseClass)
	}
	classLine += ":"
	lines = append(lines, classLine)

	cf.context.PushIndent()

	// Fields
	if len(fields) == 0 {
		lines = append(lines, fmt.Sprintf("%spass", cf.Indent()))
	} else {
		for _, field := range fields {
			fieldLines := cf.FormatDataclassField(field)
			lines = append(lines, fieldLines...)
		}
	}

	cf.context.PopIndent()
	return lines
}

// FormatPydanticField formats a Pydantic field.
func (cf *CodeFormatter) FormatPydanticField(field Field) []string {
	var lines []string

	// Add docstring if enabled
	if cf.options.IncludeDocstrings && field.Description != "" {
		docLines := cf.FormatDocstring(field.Description, field.Examples, field.DefaultValue)
		lines = append(lines, docLines...)
	}

	// Format field declaration
	fieldLine := fmt.Sprintf("%s%s: %s", cf.Indent(), field.Name, field.Type)

	// Add default value or Field() configuration
	if field.DefaultValue != nil {
		fieldLine += fmt.Sprintf(" = %v", field.DefaultValue)
	} else if !field.Required {
		fieldLine += " = None"
	}

	lines = append(lines, fieldLine)
	return lines
}

// FormatDataclassField formats a dataclass field.
func (cf *CodeFormatter) FormatDataclassField(field Field) []string {
	var lines []string

	// Add docstring if enabled
	if cf.options.IncludeDocstrings && field.Description != "" {
		docLines := cf.FormatDocstring(field.Description, field.Examples, field.DefaultValue)
		lines = append(lines, docLines...)
	}

	// Format field declaration
	fieldLine := fmt.Sprintf("%s%s: %s", cf.Indent(), field.Name, field.Type)

	// Add default value
	if field.DefaultValue != nil {
		fieldLine += fmt.Sprintf(" = %v", field.DefaultValue)
	} else if !field.Required {
		fieldLine += " = None"
	}

	lines = append(lines, fieldLine)
	return lines
}

// Field represents a Python field.
type Field struct {
	Name         string
	Type         string
	Required     bool
	Description  string
	Examples     []any
	DefaultValue any
}

// EnumFormatter handles Python enum generation.
type EnumFormatter struct {
	options PythonOptions
	context *base.GenerationContext
}

// NewEnumFormatter creates a new EnumFormatter.
func NewEnumFormatter(options PythonOptions, context *base.GenerationContext) *EnumFormatter {
	return &EnumFormatter{
		options: options,
		context: context,
	}
}

// FormatStringEnum formats a string enum.
func (ef *EnumFormatter) FormatStringEnum(name string, values []string) []string {
	switch ef.options.EnumStyle {
	case "Enum":
		return ef.formatEnum(name, values)
	case "StrEnum":
		return ef.formatStrEnum(name, values)
	case "Literal":
		return ef.formatLiteral(name, values)
	default:
		return ef.formatEnum(name, values)
	}
}

// formatEnum formats a standard Enum.
func (ef *EnumFormatter) formatEnum(name string, values []string) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf("%sclass %s(Enum):", ef.Indent(), name))

	ef.context.PushIndent()
	for _, value := range values {
		enumKey := ef.formatEnumKey(value)
		lines = append(lines, fmt.Sprintf("%s%s = \"%s\"", ef.Indent(), enumKey, value))
	}
	ef.context.PopIndent()

	return lines
}

// formatStrEnum formats a StrEnum (Python 3.11+).
func (ef *EnumFormatter) formatStrEnum(name string, values []string) []string {
	var lines []string

	lines = append(lines, fmt.Sprintf("%sclass %s(StrEnum):", ef.Indent(), name))

	ef.context.PushIndent()
	for _, value := range values {
		enumKey := ef.formatEnumKey(value)
		lines = append(lines, fmt.Sprintf("%s%s = \"%s\"", ef.Indent(), enumKey, value))
	}
	ef.context.PopIndent()

	return lines
}

// formatLiteral formats a Literal type.
func (ef *EnumFormatter) formatLiteral(name string, values []string) []string {
	var lines []string

	quotedValues := make([]string, len(values))
	for i, value := range values {
		quotedValues[i] = fmt.Sprintf("\"%s\"", value)
	}

	literalType := fmt.Sprintf("Literal[%s]", strings.Join(quotedValues, ", "))
	lines = append(lines, fmt.Sprintf("%s%s = %s", ef.Indent(), name, literalType))

	return lines
}

// formatEnumKey formats an enum key from a string value.
func (ef *EnumFormatter) formatEnumKey(value string) string {
	// Convert to a valid Python identifier
	key := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(value, "_")
	key = regexp.MustCompile(`^[0-9]`).ReplaceAllString(key, "_$0")

	if key == "" {
		key = "EMPTY"
	}

	return strings.ToUpper(base.ToSnakeCase(key))
}

// Indent returns the current indentation string.
func (ef *EnumFormatter) Indent() string {
	if ef.options.UseTabsForIndentation {
		return strings.Repeat("\t", ef.context.IndentLevel)
	}
	return strings.Repeat(" ", ef.context.IndentLevel*ef.options.IndentSize)
}

// ImportManager handles Python import generation.
type ImportManager struct {
	options PythonOptions
	imports map[string]bool
}

// NewImportManager creates a new ImportManager.
func NewImportManager(options PythonOptions) *ImportManager {
	return &ImportManager{
		options: options,
		imports: make(map[string]bool),
	}
}

// AddImport adds an import to the manager.
func (im *ImportManager) AddImport(importPath string) {
	im.imports[importPath] = true
}

// GetRequiredImports returns the required imports based on the options.
func (im *ImportManager) GetRequiredImports() []string {
	var imports []string

	// Add imports based on output style
	switch im.options.OutputStyle {
	case "pydantic":
		if im.options.PydanticVersion == "v2" {
			imports = append(imports, "from pydantic import BaseModel")
		} else {
			imports = append(imports, "from pydantic import BaseModel")
		}
	case "dataclass":
		imports = append(imports, "from dataclasses import dataclass")
	}

	// Add typing imports
	if im.options.UseTypeHints {
		if im.options.TypeHintStyle == "typing" || !im.isPython39Plus() {
			typingImports := []string{}
			if im.options.UseOptional {
				typingImports = append(typingImports, "Optional")
			}
			if !im.isPython39Plus() {
				typingImports = append(typingImports, "List", "Dict")
			}
			if !im.isPython310Plus() && !im.options.UseOptional {
				typingImports = append(typingImports, "Union")
			}
			if len(typingImports) > 0 {
				imports = append(imports, fmt.Sprintf("from typing import %s", strings.Join(typingImports, ", ")))
			}
		}
	}

	// Add enum imports
	if im.options.UseEnums {
		switch im.options.EnumStyle {
		case "Enum":
			imports = append(imports, "from enum import Enum")
		case "StrEnum":
			imports = append(imports, "from enum import StrEnum")
		case "Literal":
			imports = append(imports, "from typing import Literal")
		}
	}

	// Add extra imports
	imports = append(imports, im.options.ExtraImports...)

	// Add manually added imports
	for imp := range im.imports {
		imports = append(imports, imp)
	}

	return imports
}

// isPython39Plus checks if the target Python version is 3.9 or higher.
func (im *ImportManager) isPython39Plus() bool {
	return im.options.PythonVersion >= "3.9"
}

// isPython310Plus checks if the target Python version is 3.10 or higher.
func (im *ImportManager) isPython310Plus() bool {
	return im.options.PythonVersion >= "3.10"
}
