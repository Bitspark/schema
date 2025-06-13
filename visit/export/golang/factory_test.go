package golang

import (
	"reflect"
	"testing"
)

func TestNewGoGenerator(t *testing.T) {
	generator := NewGoGenerator()

	if generator == nil {
		t.Error("NewGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}

	if generator.Format() != "go" {
		t.Errorf("Expected generator format 'go', got %q", generator.Format())
	}
}

func TestNewGoGeneratorWithOptions(t *testing.T) {
	opts := DefaultGoOptions()
	opts.OutputStyle = "interface"
	opts.PackageName = "testpkg"

	generator := NewGoGeneratorWithOptions(opts)
	if generator == nil {
		t.Error("NewGoGeneratorWithOptions should not return nil")
	}
}

func TestBasicGoGenerator(t *testing.T) {
	generator := BasicGoGenerator()

	if generator == nil {
		t.Error("BasicGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

func TestMinimalGoGenerator(t *testing.T) {
	generator := MinimalGoGenerator()

	if generator == nil {
		t.Error("MinimalGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

func TestFullFeaturedGoGenerator(t *testing.T) {
	generator := FullFeaturedGoGenerator()

	if generator == nil {
		t.Error("FullFeaturedGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

func TestAPIGoGenerator(t *testing.T) {
	generator := APIGoGenerator()

	if generator == nil {
		t.Error("APIGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

func TestConfigGoGenerator(t *testing.T) {
	generator := ConfigGoGenerator()

	if generator == nil {
		t.Error("ConfigGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

func TestInterfaceGoGenerator(t *testing.T) {
	generator := InterfaceGoGenerator()

	if generator == nil {
		t.Error("InterfaceGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

func TestTypeAliasGoGenerator(t *testing.T) {
	generator := TypeAliasGoGenerator()

	if generator == nil {
		t.Error("TypeAliasGoGenerator should not return nil")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}

// Test Functional Options

func TestWithOutputStyle(t *testing.T) {
	tests := []string{"struct", "interface", "type_alias"}

	for _, style := range tests {
		t.Run(style, func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithOutputStyle(style)
			option(&opts)

			if opts.OutputStyle != style {
				t.Errorf("WithOutputStyle(%q): expected %q, got %q", style, style, opts.OutputStyle)
			}
		})
	}
}

func TestWithPackageName(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithPackageName("testpkg")
	option(&opts)

	if opts.PackageName != "testpkg" {
		t.Errorf("WithPackageName: expected 'testpkg', got %q", opts.PackageName)
	}
}

func TestWithNamingConvention(t *testing.T) {
	tests := []string{"PascalCase", "camelCase"}

	for _, convention := range tests {
		t.Run(convention, func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithNamingConvention(convention)
			option(&opts)

			if opts.NamingConvention != convention {
				t.Errorf("WithNamingConvention(%q): expected %q, got %q", convention, convention, opts.NamingConvention)
			}
		})
	}
}

func TestWithComments(t *testing.T) {
	tests := []bool{true, false}

	for i, value := range tests {
		t.Run(string(rune('0'+i)), func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithComments(value)
			option(&opts)

			if opts.IncludeComments != value {
				t.Errorf("WithComments(%v): expected %v, got %v", value, value, opts.IncludeComments)
			}
		})
	}
}

func TestWithJSONTags(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithJSONTags(false)
	option(&opts)

	if opts.IncludeJSONTags {
		t.Error("WithJSONTags(false): expected false, got true")
	}
}

func TestWithJSONTagStyle(t *testing.T) {
	tests := []string{"snake_case", "camelCase", "kebab-case"}

	for _, style := range tests {
		t.Run(style, func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithJSONTagStyle(style)
			option(&opts)

			if opts.JSONTagStyle != style {
				t.Errorf("WithJSONTagStyle(%q): expected %q, got %q", style, style, opts.JSONTagStyle)
			}
		})
	}
}

func TestWithValidationTags(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithValidationTags(true)
	option(&opts)

	if !opts.IncludeValidationTags {
		t.Error("WithValidationTags(true): expected true, got false")
	}
}

func TestWithPointers(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithPointers(false)
	option(&opts)

	if opts.UsePointers {
		t.Error("WithPointers(false): expected false, got true")
	}
}

func TestWithOmitEmpty(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithOmitEmpty(false)
	option(&opts)

	if opts.UseOmitEmpty {
		t.Error("WithOmitEmpty(false): expected false, got true")
	}
}

func TestWithIndentStyle(t *testing.T) {
	tests := []string{"tabs", "spaces"}

	for _, style := range tests {
		t.Run(style, func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithIndentStyle(style)
			option(&opts)

			if opts.IndentStyle != style {
				t.Errorf("WithIndentStyle(%q): expected %q, got %q", style, style, opts.IndentStyle)
			}
		})
	}
}

func TestWithIndentSize(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithIndentSize(2)
	option(&opts)

	if opts.IndentSize != 2 {
		t.Errorf("WithIndentSize(2): expected 2, got %d", opts.IndentSize)
	}
}

func TestWithConstructors(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithConstructors(true)
	option(&opts)

	if !opts.GenerateConstructors {
		t.Error("WithConstructors(true): expected true, got false")
	}
}

func TestWithValidators(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithValidators(true)
	option(&opts)

	if !opts.GenerateValidators {
		t.Error("WithValidators(true): expected true, got false")
	}
}

func TestWithStringers(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithStringers(true)
	option(&opts)

	if !opts.GenerateStringers {
		t.Error("WithStringers(true): expected true, got false")
	}
}

func TestWithGenerics(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithGenerics(true)
	option(&opts)

	if !opts.UseGenerics {
		t.Error("WithGenerics(true): expected true, got false")
	}
}

func TestWithGoVersion(t *testing.T) {
	tests := []string{"1.18", "1.19", "1.20", "1.21", "1.22"}

	for _, version := range tests {
		t.Run(version, func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithGoVersion(version)
			option(&opts)

			if opts.GoVersion != version {
				t.Errorf("WithGoVersion(%q): expected %q, got %q", version, version, opts.GoVersion)
			}
		})
	}
}

func TestWithExtraImports(t *testing.T) {
	imports := []string{"fmt", "strings", "time"}
	opts := DefaultGoOptions()
	option := WithExtraImports(imports)
	option(&opts)

	if !reflect.DeepEqual(opts.ExtraImports, imports) {
		t.Errorf("WithExtraImports: expected %v, got %v", imports, opts.ExtraImports)
	}
}

func TestWithCustomTypeMappings(t *testing.T) {
	mappings := map[string]string{
		"UUID":      "string",
		"Timestamp": "time.Time",
	}
	opts := DefaultGoOptions()
	option := WithCustomTypeMappings(mappings)
	option(&opts)

	if !reflect.DeepEqual(opts.CustomTypeMappings, mappings) {
		t.Errorf("WithCustomTypeMappings: expected %v, got %v", mappings, opts.CustomTypeMappings)
	}
}

func TestWithEnums(t *testing.T) {
	opts := DefaultGoOptions()
	option := WithEnums(false)
	option(&opts)

	if opts.GenerateEnums {
		t.Error("WithEnums(false): expected false, got true")
	}
}

func TestWithEnumStyle(t *testing.T) {
	tests := []string{"const", "type", "string"}

	for _, style := range tests {
		t.Run(style, func(t *testing.T) {
			opts := DefaultGoOptions()
			option := WithEnumStyle(style)
			option(&opts)

			if opts.EnumStyle != style {
				t.Errorf("WithEnumStyle(%q): expected %q, got %q", style, style, opts.EnumStyle)
			}
		})
	}
}

// Test Factory Functions

func TestCreateGoGenerator(t *testing.T) {
	tests := []struct {
		name        string
		options     map[string]any
		expectError bool
	}{
		{
			name:        "valid options",
			options:     map[string]any{"output_style": "interface", "package_name": "testpkg"},
			expectError: false,
		},
		{
			name:        "invalid output style",
			options:     map[string]any{"output_style": "invalid"},
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
			generator, err := CreateGoGenerator(tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				if generator != nil {
					t.Error("Expected nil generator on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if generator == nil {
					t.Error("Expected non-nil generator")
				}
			}
		})
	}
}

func TestCreateGoGeneratorFromPreset(t *testing.T) {
	tests := []struct {
		name        string
		preset      string
		expectError bool
	}{
		{
			name:        "basic preset",
			preset:      "basic",
			expectError: false,
		},
		{
			name:        "minimal preset",
			preset:      "minimal",
			expectError: false,
		},
		{
			name:        "api preset",
			preset:      "api",
			expectError: false,
		},
		{
			name:        "invalid preset",
			preset:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := CreateGoGeneratorFromPreset(tt.preset)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				if generator != nil {
					t.Error("Expected nil generator on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if generator == nil {
					t.Error("Expected non-nil generator")
				}
			}
		})
	}
}

func TestGoGeneratorFactory(t *testing.T) {
	tests := []struct {
		name        string
		options     []any
		expectError bool
	}{
		{
			name:        "valid options map",
			options:     []any{map[string]any{"output_style": "interface"}},
			expectError: false,
		},
		{
			name:        "no options",
			options:     []any{},
			expectError: false,
		},
		{
			name:        "non-map option",
			options:     []any{"invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := GoGeneratorFactory(tt.options...)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				if generator != nil {
					t.Error("Expected nil generator on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if generator == nil {
					t.Error("Expected non-nil generator")
				}
			}
		})
	}
}

func TestGetAvailablePresets(t *testing.T) {
	presets := GetAvailablePresets()

	if len(presets) == 0 {
		t.Error("Expected non-empty list of presets")
	}

	// Check for some expected presets
	expectedPresets := []string{"basic", "minimal", "api"}
	for _, expected := range expectedPresets {
		found := false
		for _, preset := range presets {
			if preset == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected preset %q not found in list", expected)
		}
	}
}

func TestGetPresetDescription(t *testing.T) {
	tests := []struct {
		preset      string
		expectEmpty bool
	}{
		{"basic", false},
		{"minimal", false},
		{"api", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			desc := GetPresetDescription(tt.preset)

			if tt.expectEmpty {
				if desc != "" {
					t.Errorf("Expected empty description for unknown preset, got %q", desc)
				}
			} else {
				if desc == "" {
					t.Errorf("Expected non-empty description for preset %q", tt.preset)
				}
			}
		})
	}
}

func TestMultipleFunctionalOptions(t *testing.T) {
	generator := NewGoGenerator(
		WithOutputStyle("interface"),
		WithPackageName("testpkg"),
		WithComments(false),
		WithJSONTags(false),
		WithValidationTags(true),
	)

	if generator == nil {
		t.Error("Expected non-nil generator")
	}

	if generator.Name() != "Go Generator" {
		t.Errorf("Expected generator name 'Go Generator', got %q", generator.Name())
	}
}
