package golang

import (
	"strings"
	"testing"

	"defs.dev/schema/api/core"
)

// Mock schema implementations for testing

type mockSchema struct {
	schemaType  core.SchemaType
	title       string
	description string
	example     any
}

func (m *mockSchema) Type() core.SchemaType {
	return m.schemaType
}

func (m *mockSchema) Metadata() core.SchemaMetadata {
	return core.SchemaMetadata{
		Name:        m.title,
		Description: m.description,
		Examples:    []any{m.example},
	}
}

func (m *mockSchema) Validate(value any) core.ValidationResult {
	return core.ValidationResult{Valid: true}
}

func (m *mockSchema) ToJSONSchema() map[string]any {
	return map[string]any{"type": string(m.schemaType)}
}

func (m *mockSchema) GenerateExample() any {
	return m.example
}

func (m *mockSchema) Clone() core.Schema {
	return m
}

// Mock string schema
type mockStringSchema struct {
	*mockSchema
	minLength    *int
	maxLength    *int
	pattern      string
	format       string
	enumValues   []string
	defaultValue *string
}

func (m *mockStringSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitString(m)
}

func (m *mockStringSchema) MinLength() *int {
	return m.minLength
}

func (m *mockStringSchema) MaxLength() *int {
	return m.maxLength
}

func (m *mockStringSchema) Pattern() string {
	return m.pattern
}

func (m *mockStringSchema) Format() string {
	return m.format
}

func (m *mockStringSchema) EnumValues() []string {
	return m.enumValues
}

func (m *mockStringSchema) DefaultValue() *string {
	return m.defaultValue
}

// Mock integer schema
type mockIntegerSchema struct {
	*mockSchema
	minimum *int64
	maximum *int64
}

func (m *mockIntegerSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitInteger(m)
}

func (m *mockIntegerSchema) Minimum() *int64 {
	return m.minimum
}

func (m *mockIntegerSchema) Maximum() *int64 {
	return m.maximum
}

// Mock number schema
type mockNumberSchema struct {
	*mockSchema
	minimum *float64
	maximum *float64
}

func (m *mockNumberSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitNumber(m)
}

func (m *mockNumberSchema) Minimum() *float64 {
	return m.minimum
}

func (m *mockNumberSchema) Maximum() *float64 {
	return m.maximum
}

// Mock boolean schema
type mockBooleanSchema struct {
	*mockSchema
}

func (m *mockBooleanSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitBoolean(m)
}

// Helper functions
func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

// Tests

func TestGenerator_VisitString(t *testing.T) {
	tests := []struct {
		name        string
		schema      core.Schema
		options     GoOptions
		contains    []string
		notContains []string
	}{
		{
			name: "basic string struct",
			schema: &mockStringSchema{
				mockSchema: &mockSchema{
					schemaType:  core.TypeString,
					title:       "Name",
					description: "A person's name",
					example:     "John Doe",
				},
			},
			options: func() GoOptions {
				opts := DefaultGoOptions()
				opts.OutputStyle = "struct"
				opts.IncludeImports = false
				return opts
			}(),
			contains:    []string{"type Name struct", "Value string", `json:"value,omitempty"`},
			notContains: []string{"interface", "type Name ="},
		},
		{
			name: "string type alias",
			schema: &mockStringSchema{
				mockSchema: &mockSchema{
					schemaType:  core.TypeString,
					title:       "UserID",
					description: "Unique user identifier",
					example:     "user123",
				},
			},
			options: func() GoOptions {
				opts := DefaultGoOptions()
				opts.OutputStyle = "type_alias"
				opts.IncludeImports = false
				return opts
			}(),
			contains:    []string{"type UserID = string"},
			notContains: []string{"struct", "interface"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.options)

			output, err := generator.Generate(tt.schema)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			outputStr := string(output)

			for _, contains := range tt.contains {
				if !strings.Contains(outputStr, contains) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", contains, outputStr)
				}
			}

			for _, notContains := range tt.notContains {
				if strings.Contains(outputStr, notContains) {
					t.Errorf("Expected output to NOT contain %q, but it did.\nOutput:\n%s", notContains, outputStr)
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

	opts := DefaultGoOptions()
	opts.IncludeImports = false
	generator := NewGenerator(opts)

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	outputStr := string(output)
	expectedContains := []string{
		"type Age struct",
		"Value int64",
		`json:"value,omitempty"`,
	}

	for _, expected := range expectedContains {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
		}
	}
}

func TestGenerator_VisitNumber(t *testing.T) {
	schema := &mockNumberSchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeNumber,
			title:       "Price",
			description: "Product price",
			example:     19.99,
		},
		minimum: float64Ptr(0.0),
		maximum: float64Ptr(1000.0),
	}

	opts := DefaultGoOptions()
	opts.IncludeImports = false
	generator := NewGenerator(opts)

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	outputStr := string(output)
	expectedContains := []string{
		"type Price struct",
		"Value float64",
		`json:"value,omitempty"`,
	}

	for _, expected := range expectedContains {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
		}
	}
}

func TestGenerator_VisitBoolean(t *testing.T) {
	schema := &mockBooleanSchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeBoolean,
			title:       "IsActive",
			description: "Whether the user is active",
			example:     true,
		},
	}

	opts := DefaultGoOptions()
	opts.IncludeImports = false
	generator := NewGenerator(opts)

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	outputStr := string(output)
	expectedContains := []string{
		"type IsActive struct",
		"Value bool",
		`json:"value,omitempty"`,
	}

	for _, expected := range expectedContains {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
		}
	}
}

func TestGenerator_OutputStyles(t *testing.T) {
	schema := &mockStringSchema{
		mockSchema: &mockSchema{
			schemaType:  core.TypeString,
			title:       "TestField",
			description: "Test field",
			example:     "test",
		},
	}

	tests := []struct {
		name        string
		outputStyle string
		contains    []string
		notContains []string
	}{
		{
			name:        "struct style",
			outputStyle: "struct",
			contains:    []string{"type TestField struct", "Value string"},
			notContains: []string{"interface", "type TestField ="},
		},
		{
			name:        "type alias style",
			outputStyle: "type_alias",
			contains:    []string{"type TestField = string"},
			notContains: []string{"struct", "interface"},
		},
		{
			name:        "interface style",
			outputStyle: "interface",
			contains:    []string{"type TestField interface", "GetValue() string", "SetValue(value string)"},
			notContains: []string{"struct", "type TestField ="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.OutputStyle = tt.outputStyle
			opts.IncludeImports = false
			generator := NewGenerator(opts)

			output, err := generator.Generate(schema)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			outputStr := string(output)

			for _, contains := range tt.contains {
				if !strings.Contains(outputStr, contains) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", contains, outputStr)
				}
			}

			for _, notContains := range tt.notContains {
				if strings.Contains(outputStr, notContains) {
					t.Errorf("Expected output to NOT contain %q, but it did.\nOutput:\n%s", notContains, outputStr)
				}
			}
		})
	}
}

func TestGenerator_Name(t *testing.T) {
	generator := NewGenerator(DefaultGoOptions())

	if name := generator.Name(); name != "Go Generator" {
		t.Errorf("Expected name 'Go Generator', got %q", name)
	}
}

func TestGenerator_Format(t *testing.T) {
	generator := NewGenerator(DefaultGoOptions())

	if format := generator.Format(); format != "go" {
		t.Errorf("Expected format 'go', got %q", format)
	}
}

func TestGoOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		options     GoOptions
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid options",
			options:     DefaultGoOptions(),
			expectError: false,
		},
		{
			name: "invalid output style",
			options: GoOptions{
				OutputStyle: "invalid",
				PackageName: "main",
			},
			expectError: true,
			errorMsg:    "unsupported output style",
		},
		{
			name: "empty package name",
			options: func() GoOptions {
				opts := DefaultGoOptions()
				opts.PackageName = ""
				return opts
			}(),
			expectError: true,
			errorMsg:    "package name cannot be empty",
		},
		{
			name: "invalid naming convention",
			options: GoOptions{
				OutputStyle:      "struct",
				PackageName:      "main",
				NamingConvention: "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported naming convention",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
			}
		})
	}
}
