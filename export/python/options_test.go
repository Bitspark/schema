package python

import (
	"testing"
)

func TestDefaultPythonOptions(t *testing.T) {
	options := DefaultPythonOptions()

	// Test default values
	if options.OutputStyle != "pydantic" {
		t.Errorf("Expected default OutputStyle to be 'pydantic', got %s", options.OutputStyle)
	}

	if options.PydanticVersion != "v2" {
		t.Errorf("Expected default PydanticVersion to be 'v2', got %s", options.PydanticVersion)
	}

	if options.NamingConvention != "PascalCase" {
		t.Errorf("Expected default NamingConvention to be 'PascalCase', got %s", options.NamingConvention)
	}

	if options.FieldNamingConvention != "snake_case" {
		t.Errorf("Expected default FieldNamingConvention to be 'snake_case', got %s", options.FieldNamingConvention)
	}

	if !options.IncludeComments {
		t.Error("Expected default IncludeComments to be true")
	}

	if !options.IncludeExamples {
		t.Error("Expected default IncludeExamples to be true")
	}

	if !options.IncludeDefaults {
		t.Error("Expected default IncludeDefaults to be true")
	}

	if options.StrictMode {
		t.Error("Expected default StrictMode to be false")
	}

	if !options.UseOptional {
		t.Error("Expected default UseOptional to be true")
	}

	if options.IndentSize != 4 {
		t.Errorf("Expected default IndentSize to be 4, got %d", options.IndentSize)
	}

	if options.UseTabsForIndentation {
		t.Error("Expected default UseTabsForIndentation to be false")
	}

	if !options.IncludeImports {
		t.Error("Expected default IncludeImports to be true")
	}

	if options.ImportStyle != "absolute" {
		t.Errorf("Expected default ImportStyle to be 'absolute', got %s", options.ImportStyle)
	}

	if !options.UseTypeHints {
		t.Error("Expected default UseTypeHints to be true")
	}

	if options.TypeHintStyle != "typing" {
		t.Errorf("Expected default TypeHintStyle to be 'typing', got %s", options.TypeHintStyle)
	}

	if options.GenerateValidators {
		t.Error("Expected default GenerateValidators to be false")
	}

	if options.ValidatorStyle != "pydantic" {
		t.Errorf("Expected default ValidatorStyle to be 'pydantic', got %s", options.ValidatorStyle)
	}

	if !options.UseEnums {
		t.Error("Expected default UseEnums to be true")
	}

	if options.EnumStyle != "Enum" {
		t.Errorf("Expected default EnumStyle to be 'Enum', got %s", options.EnumStyle)
	}

	if options.IncludeSerializers {
		t.Error("Expected default IncludeSerializers to be false")
	}

	if options.SerializerStyle != "dict" {
		t.Errorf("Expected default SerializerStyle to be 'dict', got %s", options.SerializerStyle)
	}

	if options.UseDataclassFeatures {
		t.Error("Expected default UseDataclassFeatures to be false")
	}

	if !options.IncludeDocstrings {
		t.Error("Expected default IncludeDocstrings to be true")
	}

	if options.DocstringStyle != "google" {
		t.Errorf("Expected default DocstringStyle to be 'google', got %s", options.DocstringStyle)
	}

	if options.FileHeader != "" {
		t.Errorf("Expected default FileHeader to be empty, got %s", options.FileHeader)
	}

	if options.ModuleName != "" {
		t.Errorf("Expected default ModuleName to be empty, got %s", options.ModuleName)
	}

	if options.BaseClass != "" {
		t.Errorf("Expected default BaseClass to be empty, got %s", options.BaseClass)
	}

	if options.UseForwardRefs {
		t.Error("Expected default UseForwardRefs to be false")
	}

	if options.PythonVersion != "3.9" {
		t.Errorf("Expected default PythonVersion to be '3.9', got %s", options.PythonVersion)
	}

	if len(options.ExtraImports) != 0 {
		t.Errorf("Expected default ExtraImports to be empty, got %v", options.ExtraImports)
	}

	if len(options.CustomTypeMapping) != 0 {
		t.Errorf("Expected default CustomTypeMapping to be empty, got %v", options.CustomTypeMapping)
	}
}

func TestPythonOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		options     PythonOptions
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default options",
			options:     DefaultPythonOptions(),
			expectError: false,
		},
		{
			name: "invalid output style",
			options: PythonOptions{
				OutputStyle: "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported output style",
		},
		{
			name: "invalid pydantic version",
			options: PythonOptions{
				OutputStyle:     "pydantic",
				PydanticVersion: "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported Pydantic version",
		},
		{
			name: "invalid naming convention",
			options: PythonOptions{
				OutputStyle:      "pydantic",
				PydanticVersion:  "v2",
				NamingConvention: "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported naming convention",
		},
		{
			name: "invalid field naming convention",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported field naming convention",
		},
		{
			name: "invalid indent size",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            -1,
			},
			expectError: true,
			errorMsg:    "indent size must be between",
		},
		{
			name: "invalid import style",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported import style",
		},
		{
			name: "invalid type hint style",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "absolute",
				TypeHintStyle:         "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported type hint style",
		},
		{
			name: "invalid validator style",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "absolute",
				TypeHintStyle:         "typing",
				ValidatorStyle:        "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported validator style",
		},
		{
			name: "invalid enum style",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "absolute",
				TypeHintStyle:         "typing",
				ValidatorStyle:        "pydantic",
				EnumStyle:             "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported enum style",
		},
		{
			name: "invalid serializer style",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "absolute",
				TypeHintStyle:         "typing",
				ValidatorStyle:        "pydantic",
				EnumStyle:             "Enum",
				SerializerStyle:       "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported serializer style",
		},
		{
			name: "invalid docstring style",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "absolute",
				TypeHintStyle:         "typing",
				ValidatorStyle:        "pydantic",
				EnumStyle:             "Enum",
				SerializerStyle:       "dict",
				DocstringStyle:        "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported docstring style",
		},
		{
			name: "invalid python version",
			options: PythonOptions{
				OutputStyle:           "pydantic",
				PydanticVersion:       "v2",
				NamingConvention:      "PascalCase",
				FieldNamingConvention: "snake_case",
				IndentSize:            4,
				ImportStyle:           "absolute",
				TypeHintStyle:         "typing",
				ValidatorStyle:        "pydantic",
				EnumStyle:             "Enum",
				SerializerStyle:       "dict",
				DocstringStyle:        "google",
				PythonVersion:         "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported Python version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestPythonOptions_SetOption(t *testing.T) {
	options := DefaultPythonOptions()

	tests := []struct {
		name     string
		key      string
		value    any
		expected any
		getter   func() any
	}{
		{
			name:     "set output style",
			key:      "output_style",
			value:    "dataclass",
			expected: "dataclass",
			getter:   func() any { return options.OutputStyle },
		},
		{
			name:     "set pydantic version",
			key:      "pydantic_version",
			value:    "v1",
			expected: "v1",
			getter:   func() any { return options.PydanticVersion },
		},
		{
			name:     "set naming convention",
			key:      "naming_convention",
			value:    "snake_case",
			expected: "snake_case",
			getter:   func() any { return options.NamingConvention },
		},
		{
			name:     "set field naming convention",
			key:      "field_naming_convention",
			value:    "camelCase",
			expected: "camelCase",
			getter:   func() any { return options.FieldNamingConvention },
		},
		{
			name:     "set include comments",
			key:      "include_comments",
			value:    false,
			expected: false,
			getter:   func() any { return options.IncludeComments },
		},
		{
			name:     "set include examples",
			key:      "include_examples",
			value:    false,
			expected: false,
			getter:   func() any { return options.IncludeExamples },
		},
		{
			name:     "set include defaults",
			key:      "include_defaults",
			value:    false,
			expected: false,
			getter:   func() any { return options.IncludeDefaults },
		},
		{
			name:     "set strict mode",
			key:      "strict_mode",
			value:    true,
			expected: true,
			getter:   func() any { return options.StrictMode },
		},
		{
			name:     "set use optional",
			key:      "use_optional",
			value:    false,
			expected: false,
			getter:   func() any { return options.UseOptional },
		},
		{
			name:     "set indent size",
			key:      "indent_size",
			value:    2,
			expected: 2,
			getter:   func() any { return options.IndentSize },
		},
		{
			name:     "set use tabs for indentation",
			key:      "use_tabs_for_indentation",
			value:    true,
			expected: true,
			getter:   func() any { return options.UseTabsForIndentation },
		},
		{
			name:     "set include imports",
			key:      "include_imports",
			value:    false,
			expected: false,
			getter:   func() any { return options.IncludeImports },
		},
		{
			name:     "set import style",
			key:      "import_style",
			value:    "relative",
			expected: "relative",
			getter:   func() any { return options.ImportStyle },
		},
		{
			name:     "set use type hints",
			key:      "use_type_hints",
			value:    false,
			expected: false,
			getter:   func() any { return options.UseTypeHints },
		},
		{
			name:     "set type hint style",
			key:      "type_hint_style",
			value:    "builtin",
			expected: "builtin",
			getter:   func() any { return options.TypeHintStyle },
		},
		{
			name:     "set generate validators",
			key:      "generate_validators",
			value:    true,
			expected: true,
			getter:   func() any { return options.GenerateValidators },
		},
		{
			name:     "set validator style",
			key:      "validator_style",
			value:    "custom",
			expected: "custom",
			getter:   func() any { return options.ValidatorStyle },
		},
		{
			name:     "set use enums",
			key:      "use_enums",
			value:    false,
			expected: false,
			getter:   func() any { return options.UseEnums },
		},
		{
			name:     "set enum style",
			key:      "enum_style",
			value:    "StrEnum",
			expected: "StrEnum",
			getter:   func() any { return options.EnumStyle },
		},
		{
			name:     "set include serializers",
			key:      "include_serializers",
			value:    true,
			expected: true,
			getter:   func() any { return options.IncludeSerializers },
		},
		{
			name:     "set serializer style",
			key:      "serializer_style",
			value:    "json",
			expected: "json",
			getter:   func() any { return options.SerializerStyle },
		},
		{
			name:     "set use dataclass features",
			key:      "use_dataclass_features",
			value:    true,
			expected: true,
			getter:   func() any { return options.UseDataclassFeatures },
		},
		{
			name:     "set dataclass options",
			key:      "dataclass_options",
			value:    []string{"frozen", "order"},
			expected: []string{"frozen", "order"},
			getter:   func() any { return options.DataclassOptions },
		},
		{
			name:     "set include docstrings",
			key:      "include_docstrings",
			value:    false,
			expected: false,
			getter:   func() any { return options.IncludeDocstrings },
		},
		{
			name:     "set docstring style",
			key:      "docstring_style",
			value:    "numpy",
			expected: "numpy",
			getter:   func() any { return options.DocstringStyle },
		},
		{
			name:     "set file header",
			key:      "file_header",
			value:    "# Generated file",
			expected: "# Generated file",
			getter:   func() any { return options.FileHeader },
		},
		{
			name:     "set module name",
			key:      "module_name",
			value:    "mymodule",
			expected: "mymodule",
			getter:   func() any { return options.ModuleName },
		},
		{
			name:     "set base class",
			key:      "base_class",
			value:    "MyBaseClass",
			expected: "MyBaseClass",
			getter:   func() any { return options.BaseClass },
		},
		{
			name:     "set use forward refs",
			key:      "use_forward_refs",
			value:    true,
			expected: true,
			getter:   func() any { return options.UseForwardRefs },
		},
		{
			name:     "set python version",
			key:      "python_version",
			value:    "3.11",
			expected: "3.11",
			getter:   func() any { return options.PythonVersion },
		},
		{
			name:     "set extra imports",
			key:      "extra_imports",
			value:    []string{"import os", "import sys"},
			expected: []string{"import os", "import sys"},
			getter:   func() any { return options.ExtraImports },
		},
		{
			name:     "set custom type mapping",
			key:      "custom_type_mapping",
			value:    map[string]string{"UUID": "uuid.UUID"},
			expected: map[string]string{"UUID": "uuid.UUID"},
			getter:   func() any { return options.CustomTypeMapping },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options.SetOption(tt.key, tt.value)

			actual := tt.getter()
			if !equal(actual, tt.expected) {
				t.Errorf("SetOption(%q, %v): expected %v, got %v", tt.key, tt.value, tt.expected, actual)
			}
		})
	}
}

func TestPythonOptions_SetOption_InvalidKey(t *testing.T) {
	options := DefaultPythonOptions()

	// Setting an unknown key should not panic or cause errors
	options.SetOption("unknown_key", "value")

	// The option should remain unchanged
	if options.OutputStyle != "pydantic" {
		t.Error("Unknown option key should not affect existing options")
	}
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			hasSubstring(s, substr))))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func equal(a, b any) bool {
	switch va := a.(type) {
	case []string:
		vb, ok := b.([]string)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, v := range va {
			if v != vb[i] {
				return false
			}
		}
		return true
	case map[string]string:
		vb, ok := b.(map[string]string)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if vb[k] != v {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
