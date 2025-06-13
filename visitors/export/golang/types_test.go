package golang

import (
	"strings"
	"testing"

	"defs.dev/schema/core"
	"defs.dev/schema/visitors/export/base"
)

func TestTypeMapper_MapSchemaType(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	mapper := NewTypeMapper(opts, context)

	tests := []struct {
		schemaType core.SchemaType
		expected   string
	}{
		{core.TypeString, "string"},
		{core.TypeInteger, "int64"},
		{core.TypeNumber, "float64"},
		{core.TypeBoolean, "bool"},
		{core.TypeArray, "[]any"},
		{core.TypeStructure, "any"},
	}

	for _, tt := range tests {
		t.Run(string(tt.schemaType), func(t *testing.T) {
			result := mapper.MapSchemaType(tt.schemaType)
			if result != tt.expected {
				t.Errorf("MapSchemaType(%s): expected %q, got %q", tt.schemaType, tt.expected, result)
			}
		})
	}
}

func TestTypeMapper_FormatSliceType(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	mapper := NewTypeMapper(opts, context)

	tests := []struct {
		elementType string
		expected    string
	}{
		{"string", "[]string"},
		{"int64", "[]int64"},
		{"User", "[]User"},
		{"*User", "[]*User"},
	}

	for _, tt := range tests {
		t.Run(tt.elementType, func(t *testing.T) {
			result := mapper.FormatSliceType(tt.elementType)
			if result != tt.expected {
				t.Errorf("FormatSliceType(%q): expected %q, got %q", tt.elementType, tt.expected, result)
			}
		})
	}
}

func TestTypeMapper_FormatMapType(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	mapper := NewTypeMapper(opts, context)

	tests := []struct {
		keyType   string
		valueType string
		expected  string
	}{
		{"string", "any", "map[string]any"},
		{"string", "User", "map[string]User"},
		{"int", "string", "map[int]string"},
	}

	for _, tt := range tests {
		t.Run(tt.keyType+"_"+tt.valueType, func(t *testing.T) {
			result := mapper.FormatMapType(tt.keyType, tt.valueType)
			if result != tt.expected {
				t.Errorf("FormatMapType(%q, %q): expected %q, got %q", tt.keyType, tt.valueType, tt.expected, result)
			}
		})
	}
}

func TestTypeMapper_FormatPointerType(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	mapper := NewTypeMapper(opts, context)

	tests := []struct {
		baseType string
		expected string
	}{
		{"string", "*string"},
		{"User", "*User"},
		{"[]string", "*[]string"},
	}

	for _, tt := range tests {
		t.Run(tt.baseType, func(t *testing.T) {
			result := mapper.FormatPointerType(tt.baseType)
			if result != tt.expected {
				t.Errorf("FormatPointerType(%q): expected %q, got %q", tt.baseType, tt.expected, result)
			}
		})
	}
}

func TestTypeMapper_FormatTypeName(t *testing.T) {
	tests := []struct {
		name       string
		convention string
		expected   string
	}{
		{"user_name", "PascalCase", "UserName"},
		{"user_name", "camelCase", "userName"},
		{"UserName", "PascalCase", "UserName"},
		{"UserName", "camelCase", "username"},
		{"API", "PascalCase", "API"},
		{"API", "camelCase", "api"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.convention, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.NamingConvention = tt.convention
			context := base.NewGenerationContext()
			mapper := NewTypeMapper(opts, context)

			result := mapper.FormatTypeName(tt.name)
			if result != tt.expected {
				t.Errorf("FormatTypeName(%q) with %s: expected %q, got %q", tt.name, tt.convention, tt.expected, result)
			}
		})
	}
}

func TestTypeMapper_FormatFieldName(t *testing.T) {
	tests := []struct {
		name       string
		convention string
		expected   string
	}{
		{"user_name", "PascalCase", "UserName"},
		{"user_name", "camelCase", "userName"},
		{"UserName", "PascalCase", "UserName"},
		{"UserName", "camelCase", "username"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.convention, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.FieldNamingConvention = tt.convention
			context := base.NewGenerationContext()
			mapper := NewTypeMapper(opts, context)

			result := mapper.FormatFieldName(tt.name)
			if result != tt.expected {
				t.Errorf("FormatFieldName(%q) with %s: expected %q, got %q", tt.name, tt.convention, tt.expected, result)
			}
		})
	}
}

func TestTypeMapper_FormatJSONTag(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		expected string
	}{
		{"user_name", "snake_case", "user_name"},
		{"user_name", "camelCase", "userName"},
		{"user_name", "kebab-case", "user-name"},
		{"UserName", "snake_case", "user_name"},
		{"UserName", "camelCase", "username"},
		{"UserName", "kebab-case", "user-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.style, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.JSONTagStyle = tt.style
			context := base.NewGenerationContext()
			mapper := NewTypeMapper(opts, context)

			result := mapper.FormatJSONTag(tt.name)
			if result != tt.expected {
				t.Errorf("FormatJSONTag(%q) with %s: expected %q, got %q", tt.name, tt.style, tt.expected, result)
			}
		})
	}
}

func TestCodeFormatter_FormatStruct(t *testing.T) {
	opts := DefaultGoOptions()
	opts.IncludeImports = false
	context := base.NewGenerationContext()
	formatter := NewCodeFormatter(opts, context)

	fields := []Field{
		{
			Name:         "Name",
			Type:         "string",
			OriginalName: "name",
			Required:     false,
			Description:  "User's name",
			JSONTag:      "name",
		},
		{
			Name:         "Age",
			Type:         "int",
			OriginalName: "age",
			Required:     true,
			Description:  "User's age",
			JSONTag:      "age",
		},
	}

	lines := formatter.FormatStruct("User", fields)

	// Check that struct is properly formatted
	structFound := false
	for _, line := range lines {
		if strings.Contains(line, "type User struct") {
			structFound = true
			break
		}
	}
	if !structFound {
		t.Error("Expected struct definition not found")
	}

	// Check that fields are included
	nameFieldFound := false
	ageFieldFound := false
	for _, line := range lines {
		if strings.Contains(line, "Name string") {
			nameFieldFound = true
		}
		if strings.Contains(line, "Age int") {
			ageFieldFound = true
		}
	}
	if !nameFieldFound {
		t.Error("Expected Name field not found")
	}
	if !ageFieldFound {
		t.Error("Expected Age field not found")
	}
}

func TestCodeFormatter_FormatStructTags(t *testing.T) {
	tests := []struct {
		name     string
		options  GoOptions
		field    Field
		expected string
	}{
		{
			name: "JSON tags only",
			options: func() GoOptions {
				opts := DefaultGoOptions()
				opts.IncludeXMLTags = false
				opts.IncludeYAMLTags = false
				return opts
			}(),
			field: Field{
				Name:         "Name",
				OriginalName: "name",
				Required:     false,
				JSONTag:      "name",
			},
			expected: `json:"name,omitempty"`,
		},
		{
			name: "JSON tags without omitempty",
			options: func() GoOptions {
				opts := DefaultGoOptions()
				opts.UseOmitEmpty = false
				opts.IncludeXMLTags = false
				opts.IncludeYAMLTags = false
				return opts
			}(),
			field: Field{
				Name:         "Name",
				OriginalName: "name",
				Required:     false,
				JSONTag:      "name",
			},
			expected: `json:"name"`,
		},
		{
			name: "Required field with omitempty enabled",
			options: func() GoOptions {
				opts := DefaultGoOptions()
				opts.UseOmitEmpty = true
				opts.IncludeXMLTags = false
				opts.IncludeYAMLTags = false
				return opts
			}(),
			field: Field{
				Name:         "Name",
				OriginalName: "name",
				Required:     true,
				JSONTag:      "name",
			},
			expected: `json:"name"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := base.NewGenerationContext()
			formatter := NewCodeFormatter(tt.options, context)

			result := formatter.FormatStructTags(tt.field)
			if result != tt.expected {
				t.Errorf("FormatStructTags(): expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCodeFormatter_FormatInterface(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	formatter := NewCodeFormatter(opts, context)

	methods := []Method{
		{
			Name:        "GetName",
			Parameters:  "",
			ReturnType:  "string",
			Description: "Returns the name",
		},
		{
			Name:        "SetName",
			Parameters:  "name string",
			ReturnType:  "",
			Description: "Sets the name",
		},
	}

	lines := formatter.FormatInterface("User", methods)

	// Check that interface is properly formatted
	interfaceFound := false
	for _, line := range lines {
		if strings.Contains(line, "type User interface") {
			interfaceFound = true
			break
		}
	}
	if !interfaceFound {
		t.Error("Expected interface definition not found")
	}

	// Check that methods are included
	getNameFound := false
	setNameFound := false
	for _, line := range lines {
		if strings.Contains(line, "GetName() string") {
			getNameFound = true
		}
		if strings.Contains(line, "SetName(name string)") {
			setNameFound = true
		}
	}
	if !getNameFound {
		t.Error("Expected GetName method not found")
	}
	if !setNameFound {
		t.Error("Expected SetName method not found")
	}
}

func TestCodeFormatter_FormatTypeAlias(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	formatter := NewCodeFormatter(opts, context)

	lines := formatter.FormatTypeAlias("UserID", "string")

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	expected := "type UserID = string"
	if !strings.Contains(lines[0], expected) {
		t.Errorf("Expected %q, got %q", expected, lines[0])
	}
}

func TestEnumFormatter_FormatStringEnum(t *testing.T) {
	tests := []struct {
		name   string
		style  string
		values []string
	}{
		{
			name:   "const style",
			style:  "const",
			values: []string{"active", "inactive", "pending"},
		},
		{
			name:   "type style",
			style:  "type",
			values: []string{"red", "green", "blue"},
		},
		{
			name:   "string style",
			style:  "string",
			values: []string{"small", "medium", "large"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.EnumStyle = tt.style
			context := base.NewGenerationContext()
			formatter := NewEnumFormatter(opts, context)

			lines := formatter.FormatStringEnum("Status", tt.values)

			if len(lines) == 0 {
				t.Error("Expected non-empty output")
			}

			// Check that type definition exists
			typeDefFound := false
			for _, line := range lines {
				if strings.Contains(line, "type Status") {
					typeDefFound = true
					break
				}
			}
			if !typeDefFound {
				t.Error("Expected type definition not found")
			}
		})
	}
}

func TestImportManager_AddImport(t *testing.T) {
	opts := DefaultGoOptions()
	opts.IncludeJSONTags = false // Disable automatic imports
	manager := NewImportManager(opts)

	manager.AddImport("fmt")
	manager.AddImport("strings")
	manager.AddImport("fmt") // duplicate

	imports := manager.GetRequiredImports()

	if len(imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(imports))
	}

	// Check that both imports are present
	fmtFound := false
	stringsFound := false
	for _, imp := range imports {
		if imp == "fmt" {
			fmtFound = true
		}
		if imp == "strings" {
			stringsFound = true
		}
	}

	if !fmtFound {
		t.Error("Expected 'fmt' import not found")
	}
	if !stringsFound {
		t.Error("Expected 'strings' import not found")
	}
}

func TestImportManager_GetRequiredImports(t *testing.T) {
	opts := DefaultGoOptions()
	opts.IncludeJSONTags = false // Disable automatic imports
	manager := NewImportManager(opts)

	// Initially should be empty
	imports := manager.GetRequiredImports()
	if len(imports) != 0 {
		t.Errorf("Expected 0 imports initially, got %d", len(imports))
	}

	// Add some imports
	manager.AddImport("time")
	manager.AddImport("encoding/json")

	imports = manager.GetRequiredImports()
	if len(imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(imports))
	}
}

func TestField_Struct(t *testing.T) {
	field := Field{
		Name:          "UserName",
		Type:          "string",
		OriginalName:  "user_name",
		Required:      true,
		Description:   "The user's name",
		Examples:      []any{"John", "Jane"},
		DefaultValue:  "Unknown",
		JSONTag:       "user_name",
		XMLTag:        "UserName",
		YAMLTag:       "user_name",
		ValidationTag: "required,min=1",
	}

	// Test that all fields are properly set
	if field.Name != "UserName" {
		t.Errorf("Expected Name 'UserName', got %q", field.Name)
	}
	if field.Type != "string" {
		t.Errorf("Expected Type 'string', got %q", field.Type)
	}
	if field.OriginalName != "user_name" {
		t.Errorf("Expected OriginalName 'user_name', got %q", field.OriginalName)
	}
	if !field.Required {
		t.Error("Expected Required to be true")
	}
	if field.Description != "The user's name" {
		t.Errorf("Expected Description 'The user's name', got %q", field.Description)
	}
}

func TestMethod_Struct(t *testing.T) {
	method := Method{
		Name:        "GetUserName",
		Parameters:  "",
		ReturnType:  "string",
		Description: "Returns the user's name",
	}

	// Test that all fields are properly set
	if method.Name != "GetUserName" {
		t.Errorf("Expected Name 'GetUserName', got %q", method.Name)
	}
	if method.Parameters != "" {
		t.Errorf("Expected empty Parameters, got %q", method.Parameters)
	}
	if method.ReturnType != "string" {
		t.Errorf("Expected ReturnType 'string', got %q", method.ReturnType)
	}
	if method.Description != "Returns the user's name" {
		t.Errorf("Expected Description 'Returns the user's name', got %q", method.Description)
	}
}

func TestCodeFormatter_IndentationStyles(t *testing.T) {
	tests := []struct {
		style    string
		size     int
		expected string
	}{
		{"tabs", 1, "\t"},
		{"spaces", 2, "  "},
		{"spaces", 4, "    "},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.IndentStyle = tt.style
			opts.IndentSize = tt.size
			context := base.NewGenerationContext()
			context.PushIndent() // Set indent level to 1
			formatter := NewCodeFormatter(opts, context)

			result := formatter.Indent()
			if result != tt.expected {
				t.Errorf("Indent() with %s/%d: expected %q, got %q", tt.style, tt.size, tt.expected, result)
			}
		})
	}
}

func TestCodeFormatter_FormatComment(t *testing.T) {
	opts := DefaultGoOptions()
	context := base.NewGenerationContext()
	formatter := NewCodeFormatter(opts, context)

	tests := []struct {
		comment  string
		expected []string
	}{
		{
			comment:  "Single line comment",
			expected: []string{"// Single line comment"},
		},
		{
			comment:  "Multi line\ncomment here",
			expected: []string{"// Multi line", "// comment here"},
		},
		{
			comment:  "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			result := formatter.FormatComment(tt.comment)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], line)
				}
			}
		})
	}
}
