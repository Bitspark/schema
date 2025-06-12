package python

import (
	"fmt"
	"strings"
	"testing"

	"defs.dev/schema/api/core"
)

// Simplified mock schema that implements the minimum required interfaces
type mockSchema struct {
	schemaType  core.SchemaType
	title       string
	description string
	example     any
	defaultVal  any
}

func (s *mockSchema) Type() core.SchemaType { return s.schemaType }
func (s *mockSchema) ToJSONSchema() map[string]any {
	return map[string]any{"type": string(s.schemaType)}
}
func (s *mockSchema) Metadata() core.SchemaMetadata {
	return core.SchemaMetadata{
		Name:        s.title,
		Description: s.description,
		Examples:    []any{s.example},
	}
}
func (s *mockSchema) Annotations() []core.Annotation {
	return []core.Annotation{}
}
func (s *mockSchema) GenerateExample() any { return s.example }
func (s *mockSchema) Clone() core.Schema   { return s }

func (s *mockSchema) Accept(visitor core.SchemaVisitor) error {
	switch s.schemaType {
	case core.TypeString:
		return visitor.VisitString(&mockStringSchema{mockSchema: s})
	case core.TypeInteger:
		return visitor.VisitInteger(&mockIntegerSchema{mockSchema: s})
	case core.TypeBoolean:
		return visitor.VisitBoolean(&mockBooleanSchema{mockSchema: s})
	case core.TypeArray:
		return visitor.VisitArray(&mockArraySchema{mockSchema: s})
	case core.TypeStructure:
		return visitor.VisitObject(&mockObjectSchema{mockSchema: s})
	default:
		return nil
	}
}

// Specific schema type implementations
type mockStringSchema struct {
	*mockSchema
	minLength *int
	maxLength *int
	pattern   string
	format    string
	enum      []string
	defVal    *string
}

func (s *mockStringSchema) MinLength() *int       { return s.minLength }
func (s *mockStringSchema) MaxLength() *int       { return s.maxLength }
func (s *mockStringSchema) Pattern() string       { return s.pattern }
func (s *mockStringSchema) Format() string        { return s.format }
func (s *mockStringSchema) EnumValues() []string  { return s.enum }
func (s *mockStringSchema) DefaultValue() *string { return s.defVal }
func (s *mockStringSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitString(s)
}

type mockIntegerSchema struct {
	*mockSchema
	minimum *int64
	maximum *int64
}

func (s *mockIntegerSchema) Minimum() *int64 { return s.minimum }
func (s *mockIntegerSchema) Maximum() *int64 { return s.maximum }
func (s *mockIntegerSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitInteger(s)
}

type mockBooleanSchema struct {
	*mockSchema
}

func (s *mockBooleanSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitBoolean(s)
}

type mockArraySchema struct {
	*mockSchema
	items    core.Schema
	minItems *int
	maxItems *int
	unique   bool
}

func (s *mockArraySchema) ItemSchema() core.Schema   { return s.items }
func (s *mockArraySchema) MinItems() *int            { return s.minItems }
func (s *mockArraySchema) MaxItems() *int            { return s.maxItems }
func (s *mockArraySchema) UniqueItemsRequired() bool { return s.unique }
func (s *mockArraySchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitArray(s)
}

type mockObjectSchema struct {
	*mockSchema
	properties map[string]core.Schema
	required   []string
	additional bool
}

func (s *mockObjectSchema) Properties() map[string]core.Schema { return s.properties }
func (s *mockObjectSchema) Required() []string                 { return s.required }
func (s *mockObjectSchema) AdditionalProperties() bool         { return s.additional }
func (s *mockObjectSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitObject(s)
}

// invalidSchema for testing error handling
type invalidSchema struct {
	*mockSchema
}

// Override Accept to return an error
func (s *invalidSchema) Accept(visitor core.SchemaVisitor) error {
	return fmt.Errorf("mock error for testing")
}

func TestGenerator_Name(t *testing.T) {
	generator := NewGenerator(DefaultPythonOptions())

	if got := generator.Name(); got != "python" {
		t.Errorf("Name() = %v, want %v", got, "python")
	}
}

func TestGenerator_Format(t *testing.T) {
	generator := NewGenerator(DefaultPythonOptions())

	if got := generator.Format(); got != "python" {
		t.Errorf("Format() = %v, want %v", got, "python")
	}
}

func TestGenerator_VisitString(t *testing.T) {
	tests := []struct {
		name     string
		schema   *mockStringSchema
		options  PythonOptions
		contains []string
	}{
		{
			name: "basic string with pydantic",
			schema: &mockStringSchema{
				mockSchema: &mockSchema{
					schemaType:  core.TypeString,
					title:       "Name",
					description: "A person's name",
					example:     "John Doe",
				},
			},
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IncludeDocstrings:     true,
				IncludeComments:       true,
				IncludeImports:        true,
				UseTypeHints:          true,
				TypeHintStyle:         "typing",
				UseOptional:           true,
				UseEnums:              true,
				EnumStyle:             "Enum",
				IndentSize:            4,
				ImportStyle:           "absolute",
				ValidatorStyle:        "pydantic",
				SerializerStyle:       "dict",
				DocstringStyle:        "google",
				PythonVersion:         "3.9",
			},
			contains: []string{"class Name", "str", "A person's name"},
		},
		{
			name: "string with enum",
			schema: &mockStringSchema{
				mockSchema: &mockSchema{
					schemaType: core.TypeString,
					title:      "Status",
				},
				enum: []string{"active", "inactive", "pending"},
			},
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				UseEnums:              true,
				EnumStyle:             "Enum",
				IncludeImports:        true,
				UseTypeHints:          true,
				TypeHintStyle:         "typing",
				UseOptional:           true,
				IndentSize:            4,
				ImportStyle:           "absolute",
				ValidatorStyle:        "pydantic",
				SerializerStyle:       "dict",
				DocstringStyle:        "google",
				PythonVersion:         "3.9",
			},
			contains: []string{"class Status", "Enum"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.options)

			output, err := generator.Generate(tt.schema)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			result := string(output)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Generate() output missing %q\nGot:\n%s", expected, result)
				}
			}
		})
	}
}

func TestGenerator_VisitInteger(t *testing.T) {
	schema := &mockIntegerSchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeInteger,
			title:       "Age",
			description: "Person's age",
			example:     25,
		},
		minimum: int64Ptr(0),
		maximum: int64Ptr(120),
	}

	generator := NewGenerator(PythonOptions{
		OutputStyle:           "pydantic",
		PydanticVersion:       "v2",
		NamingConvention:      "PascalCase",
		FieldNamingConvention: "snake_case",
		IncludeDocstrings:     true,
		IncludeImports:        true,
		UseTypeHints:          true,
		TypeHintStyle:         "typing",
		UseOptional:           true,
		UseEnums:              true,
		EnumStyle:             "Enum",
		IndentSize:            4,
		ImportStyle:           "absolute",
		ValidatorStyle:        "pydantic",
		SerializerStyle:       "dict",
		DocstringStyle:        "google",
		PythonVersion:         "3.9",
	})

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(output)
	expected := []string{"class Age", "int", "Person's age"}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Generate() output missing %q\nGot:\n%s", exp, result)
		}
	}
}

func TestGenerator_VisitBoolean(t *testing.T) {
	schema := &mockBooleanSchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeBoolean,
			title:       "IsActive",
			description: "Whether the item is active",
			defaultVal:  true,
		},
	}

	generator := NewGenerator(PythonOptions{
		OutputStyle:           "pydantic",
		PydanticVersion:       "v2",
		NamingConvention:      "PascalCase",
		FieldNamingConvention: "snake_case",
		IncludeDefaults:       true,
		IncludeDocstrings:     true,
		IncludeImports:        true,
		UseTypeHints:          true,
		TypeHintStyle:         "typing",
		UseOptional:           true,
		UseEnums:              true,
		EnumStyle:             "Enum",
		IndentSize:            4,
		ImportStyle:           "absolute",
		ValidatorStyle:        "pydantic",
		SerializerStyle:       "dict",
		DocstringStyle:        "google",
		PythonVersion:         "3.9",
	})

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(output)
	expected := []string{"class IsActive", "bool", "Whether the item is active"}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Generate() output missing %q\nGot:\n%s", exp, result)
		}
	}
}

func TestGenerator_VisitArray(t *testing.T) {
	itemSchema := &mockSchema{
		schemaType: core.TypeString,
		title:      "Item",
	}

	schema := &mockArraySchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeArray,
			title:       "Items",
			description: "List of items",
		},
		items:    itemSchema,
		minItems: intPtr(1),
		maxItems: intPtr(10),
	}

	generator := NewGenerator(PythonOptions{
		OutputStyle:           "pydantic",
		PydanticVersion:       "v2",
		NamingConvention:      "PascalCase",
		FieldNamingConvention: "snake_case",
		UseTypeHints:          true,
		IncludeDocstrings:     true,
		IncludeImports:        true,
		TypeHintStyle:         "typing",
		UseOptional:           true,
		UseEnums:              true,
		EnumStyle:             "Enum",
		IndentSize:            4,
		ImportStyle:           "absolute",
		ValidatorStyle:        "pydantic",
		SerializerStyle:       "dict",
		DocstringStyle:        "google",
		PythonVersion:         "3.9",
	})

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(output)
	expected := []string{"class Items", "List[", "List of items"}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Generate() output missing %q\nGot:\n%s", exp, result)
		}
	}
}

func TestGenerator_VisitObject(t *testing.T) {
	properties := map[string]core.Schema{
		"name": &mockSchema{schemaType: core.TypeString, title: "Name"},
		"age":  &mockSchema{schemaType: core.TypeInteger, title: "Age"},
	}

	schema := &mockObjectSchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeStructure,
			title:       "Person",
			description: "A person object",
		},
		properties: properties,
		required:   []string{"name"},
	}

	generator := NewGenerator(PythonOptions{
		OutputStyle:           "pydantic",
		PydanticVersion:       "v2",
		NamingConvention:      "PascalCase",
		FieldNamingConvention: "snake_case",
		UseTypeHints:          true,
		IncludeDocstrings:     true,
		IncludeImports:        true,
		TypeHintStyle:         "typing",
		UseOptional:           true,
		UseEnums:              true,
		EnumStyle:             "Enum",
		IndentSize:            4,
		ImportStyle:           "absolute",
		ValidatorStyle:        "pydantic",
		SerializerStyle:       "dict",
		DocstringStyle:        "google",
		PythonVersion:         "3.9",
	})

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(output)
	expected := []string{"class Person", "name:", "age:", "A person object"}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Generate() output missing %q\nGot:\n%s", exp, result)
		}
	}
}

func TestGenerator_OutputStyles(t *testing.T) {
	schema := &mockSchema{
		schemaType:  core.TypeString,
		title:       "TestField",
		description: "A test field",
	}

	tests := []struct {
		name        string
		outputStyle string
		contains    []string
		notContains []string
	}{
		{
			name:        "pydantic style",
			outputStyle: "pydantic",
			contains:    []string{"BaseModel", "class TestField"},
			notContains: []string{"@dataclass", "NamedTuple"},
		},
		{
			name:        "dataclass style",
			outputStyle: "dataclass",
			contains:    []string{"@dataclass", "class TestField"},
			notContains: []string{"BaseModel", "NamedTuple"},
		},
		{
			name:        "class style",
			outputStyle: "class",
			contains:    []string{"class TestField"},
			notContains: []string{"BaseModel", "@dataclass", "NamedTuple"},
		},
		{
			name:        "namedtuple style",
			outputStyle: "namedtuple",
			contains:    []string{"namedtuple", "TestField"},
			notContains: []string{"BaseModel", "@dataclass", "class TestField"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := DefaultPythonOptions()
			options.OutputStyle = tt.outputStyle

			generator := NewGenerator(options)
			output, err := generator.Generate(schema)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			result := string(output)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Generate() output missing %q for style %s\nGot:\n%s", expected, tt.outputStyle, result)
				}
			}

			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Generate() output should not contain %q for style %s\nGot:\n%s", notExpected, tt.outputStyle, result)
				}
			}
		})
	}
}

func TestGenerator_InvalidOptions(t *testing.T) {
	schema := &mockSchema{schemaType: core.TypeString, title: "Test"}

	// Test with invalid output style
	options := DefaultPythonOptions()
	options.OutputStyle = "invalid"

	generator := NewGenerator(options)
	_, err := generator.Generate(schema)
	if err == nil {
		t.Error("Generate() should return error for invalid output style")
	}
}

func TestGenerator_ErrorHandling(t *testing.T) {
	// Test with schema that doesn't implement Accepter properly
	schema := &invalidSchema{
		mockSchema: &mockSchema{
			schemaType: core.TypeString,
			title:      "Invalid",
		},
	}

	generator := NewGenerator(DefaultPythonOptions())
	_, err := generator.Generate(schema)
	if err == nil {
		t.Error("Generate() should return error when Accept method fails")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
