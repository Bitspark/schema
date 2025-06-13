package python

import (
	"strings"
	"testing"

	"defs.dev/schema/core"
)

func TestNewPythonGenerator(t *testing.T) {
	// Test with default options
	generator := NewPythonGenerator()

	if generator.Name() != "python" {
		t.Errorf("Expected generator name to be 'python', got %s", generator.Name())
	}

	if generator.Format() != "python" {
		t.Errorf("Expected generator format to be 'python', got %s", generator.Format())
	}
}

func TestNewPythonGenerator_WithOptions(t *testing.T) {
	// Test with custom options
	generator := NewPythonGenerator(
		WithOutputStyle("dataclass"),
		WithNamingConvention("snake_case"),
		WithIndentSize(2),
		WithComments(false),
	)

	if generator.Name() != "python" {
		t.Errorf("Expected generator name to be 'python', got %s", generator.Name())
	}

	// Test that options were applied (we can't directly access them, but we can test generation)
	schema := &mockSchema{
		schemaType: core.TypeString,
		title:      "TestField",
	}

	output, err := generator.Generate(schema)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "@dataclass") {
		t.Error("Expected dataclass output style to be applied")
	}
}

func TestPydanticV2Preset(t *testing.T) {
	options := PydanticV2Preset()

	// Apply options to a default config
	config := DefaultPythonOptions()
	for _, opt := range options {
		opt(&config)
	}

	// Test that preset values were applied
	if config.OutputStyle != "pydantic" {
		t.Errorf("Expected OutputStyle to be 'pydantic', got %s", config.OutputStyle)
	}

	if config.PydanticVersion != "v2" {
		t.Errorf("Expected PydanticVersion to be 'v2', got %s", config.PydanticVersion)
	}

	if config.NamingConvention != "PascalCase" {
		t.Errorf("Expected NamingConvention to be 'PascalCase', got %s", config.NamingConvention)
	}

	if config.FieldNamingConvention != "snake_case" {
		t.Errorf("Expected FieldNamingConvention to be 'snake_case', got %s", config.FieldNamingConvention)
	}

	if !config.UseTypeHints {
		t.Error("Expected UseTypeHints to be true")
	}

	if config.TypeHintStyle != "typing" {
		t.Errorf("Expected TypeHintStyle to be 'typing', got %s", config.TypeHintStyle)
	}

	if !config.UseOptional {
		t.Error("Expected UseOptional to be true")
	}

	if !config.UseEnums {
		t.Error("Expected UseEnums to be true")
	}

	if config.EnumStyle != "Enum" {
		t.Errorf("Expected EnumStyle to be 'Enum', got %s", config.EnumStyle)
	}

	if !config.IncludeDocstrings {
		t.Error("Expected IncludeDocstrings to be true")
	}

	if config.DocstringStyle != "google" {
		t.Errorf("Expected DocstringStyle to be 'google', got %s", config.DocstringStyle)
	}

	if !config.IncludeComments {
		t.Error("Expected IncludeComments to be true")
	}

	if !config.IncludeExamples {
		t.Error("Expected IncludeExamples to be true")
	}

	if !config.IncludeDefaults {
		t.Error("Expected IncludeDefaults to be true")
	}

	if !config.IncludeImports {
		t.Error("Expected IncludeImports to be true")
	}

	if config.PythonVersion != "3.9" {
		t.Errorf("Expected PythonVersion to be '3.9', got %s", config.PythonVersion)
	}
}

func TestPydanticV1Preset(t *testing.T) {
	options := PydanticV1Preset()

	// Apply options to a default config
	config := DefaultPythonOptions()
	for _, opt := range options {
		opt(&config)
	}

	// Test key differences from v2
	if config.PydanticVersion != "v1" {
		t.Errorf("Expected PydanticVersion to be 'v1', got %s", config.PydanticVersion)
	}

	if config.PythonVersion != "3.8" {
		t.Errorf("Expected PythonVersion to be '3.8', got %s", config.PythonVersion)
	}
}

func TestDataclassPreset(t *testing.T) {
	options := DataclassPreset()

	// Apply options to a default config
	config := DefaultPythonOptions()
	for _, opt := range options {
		opt(&config)
	}

	// Test dataclass-specific settings
	if config.OutputStyle != "dataclass" {
		t.Errorf("Expected OutputStyle to be 'dataclass', got %s", config.OutputStyle)
	}

	if !config.UseDataclassFeatures {
		t.Error("Expected UseDataclassFeatures to be true")
	}

	if len(config.DataclassOptions) == 0 || config.DataclassOptions[0] != "frozen" {
		t.Errorf("Expected DataclassOptions to contain 'frozen', got %v", config.DataclassOptions)
	}

	if config.PythonVersion != "3.9" {
		t.Errorf("Expected PythonVersion to be '3.9', got %s", config.PythonVersion)
	}
}

func TestModernPythonPreset(t *testing.T) {
	options := ModernPythonPreset()

	// Apply options to a default config
	config := DefaultPythonOptions()
	for _, opt := range options {
		opt(&config)
	}

	// Test modern Python features
	if config.TypeHintStyle != "builtin" {
		t.Errorf("Expected TypeHintStyle to be 'builtin', got %s", config.TypeHintStyle)
	}

	if config.UseOptional {
		t.Error("Expected UseOptional to be false (use T | None instead)")
	}

	if config.EnumStyle != "StrEnum" {
		t.Errorf("Expected EnumStyle to be 'StrEnum', got %s", config.EnumStyle)
	}

	if config.PythonVersion != "3.10" {
		t.Errorf("Expected PythonVersion to be '3.10', got %s", config.PythonVersion)
	}
}

func TestMinimalPreset(t *testing.T) {
	options := MinimalPreset()

	// Apply options to a default config
	config := DefaultPythonOptions()
	for _, opt := range options {
		opt(&config)
	}

	// Test minimal settings
	if config.OutputStyle != "class" {
		t.Errorf("Expected OutputStyle to be 'class', got %s", config.OutputStyle)
	}

	if config.UseTypeHints {
		t.Error("Expected UseTypeHints to be false")
	}

	if config.UseOptional {
		t.Error("Expected UseOptional to be false")
	}

	if config.UseEnums {
		t.Error("Expected UseEnums to be false")
	}

	if config.IncludeDocstrings {
		t.Error("Expected IncludeDocstrings to be false")
	}

	if config.IncludeComments {
		t.Error("Expected IncludeComments to be false")
	}

	if config.IncludeExamples {
		t.Error("Expected IncludeExamples to be false")
	}

	if config.IncludeDefaults {
		t.Error("Expected IncludeDefaults to be false")
	}

	if config.IncludeImports {
		t.Error("Expected IncludeImports to be false")
	}

	if config.PythonVersion != "3.8" {
		t.Errorf("Expected PythonVersion to be '3.8', got %s", config.PythonVersion)
	}
}

func TestStrictPreset(t *testing.T) {
	options := StrictPreset()

	// Apply options to a default config
	config := DefaultPythonOptions()
	for _, opt := range options {
		opt(&config)
	}

	// Test strict settings
	if !config.StrictMode {
		t.Error("Expected StrictMode to be true")
	}

	if !config.GenerateValidators {
		t.Error("Expected GenerateValidators to be true")
	}

	if config.ValidatorStyle != "pydantic" {
		t.Errorf("Expected ValidatorStyle to be 'pydantic', got %s", config.ValidatorStyle)
	}

	if config.DocstringStyle != "sphinx" {
		t.Errorf("Expected DocstringStyle to be 'sphinx', got %s", config.DocstringStyle)
	}
}

func TestCreatePythonGenerator(t *testing.T) {
	tests := []struct {
		name        string
		options     map[string]any
		expectError bool
	}{
		{
			name: "valid options",
			options: map[string]any{
				"output_style":      "pydantic",
				"pydantic_version":  "v2",
				"naming_convention": "PascalCase",
			},
			expectError: false,
		},
		{
			name: "invalid output style",
			options: map[string]any{
				"output_style": "invalid",
			},
			expectError: true,
		},
		{
			name:        "empty options",
			options:     map[string]any{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := CreatePythonGenerator(tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if generator == nil {
					t.Error("Expected generator, got nil")
				} else {
					if generator.Name() != "python" {
						t.Errorf("Expected generator name to be 'python', got %s", generator.Name())
					}

					if generator.Format() != "python" {
						t.Errorf("Expected generator format to be 'python', got %s", generator.Format())
					}
				}
			}
		})
	}
}

func TestPythonGeneratorFactory(t *testing.T) {
	tests := []struct {
		name        string
		options     []any
		expectError bool
	}{
		{
			name:        "valid options map",
			options:     []any{map[string]any{"output_style": "dataclass"}},
			expectError: false,
		},
		{
			name: "multiple options maps",
			options: []any{
				map[string]any{"output_style": "pydantic"},
				map[string]any{"pydantic_version": "v1"},
			},
			expectError: false,
		},
		{
			name:        "invalid options",
			options:     []any{map[string]any{"output_style": "invalid"}},
			expectError: true,
		},
		{
			name:        "no options",
			options:     []any{},
			expectError: false,
		},
		{
			name:        "non-map option",
			options:     []any{"invalid"},
			expectError: false, // Should be ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := PythonGeneratorFactory(tt.options...)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if generator == nil {
					t.Error("Expected generator, got nil")
				} else {
					if generator.Name() != "python" {
						t.Errorf("Expected generator name to be 'python', got %s", generator.Name())
					}
				}
			}
		})
	}
}

func TestFunctionalOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(PythonOptions) bool
	}{
		{
			name:   "WithOutputStyle",
			option: WithOutputStyle("dataclass"),
			expected: func(o PythonOptions) bool {
				return o.OutputStyle == "dataclass"
			},
		},
		{
			name:   "WithPydanticVersion",
			option: WithPydanticVersion("v1"),
			expected: func(o PythonOptions) bool {
				return o.PydanticVersion == "v1"
			},
		},
		{
			name:   "WithNamingConvention",
			option: WithNamingConvention("snake_case"),
			expected: func(o PythonOptions) bool {
				return o.NamingConvention == "snake_case"
			},
		},
		{
			name:   "WithComments",
			option: WithComments(false),
			expected: func(o PythonOptions) bool {
				return !o.IncludeComments
			},
		},
		{
			name:   "WithStrictMode",
			option: WithStrictMode(true),
			expected: func(o PythonOptions) bool {
				return o.StrictMode
			},
		},
		{
			name:   "WithIndentSize",
			option: WithIndentSize(2),
			expected: func(o PythonOptions) bool {
				return o.IndentSize == 2
			},
		},
		{
			name:   "WithTypeHints",
			option: WithTypeHints(false),
			expected: func(o PythonOptions) bool {
				return !o.UseTypeHints
			},
		},
		{
			name:   "WithEnums",
			option: WithEnums(false),
			expected: func(o PythonOptions) bool {
				return !o.UseEnums
			},
		},
		{
			name:   "WithPythonVersion",
			option: WithPythonVersion("3.11"),
			expected: func(o PythonOptions) bool {
				return o.PythonVersion == "3.11"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := DefaultPythonOptions()
			tt.option(&options)

			if !tt.expected(options) {
				t.Errorf("Option %s was not applied correctly", tt.name)
			}
		})
	}
}
