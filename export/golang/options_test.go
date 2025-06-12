package golang

import (
	"reflect"
	"testing"
)

func TestDefaultGoOptions(t *testing.T) {
	opts := DefaultGoOptions()

	// Test key default values
	if opts.OutputStyle != "struct" {
		t.Errorf("Expected OutputStyle 'struct', got %q", opts.OutputStyle)
	}
	if opts.PackageName != "main" {
		t.Errorf("Expected PackageName 'main', got %q", opts.PackageName)
	}
	if opts.NamingConvention != "PascalCase" {
		t.Errorf("Expected NamingConvention 'PascalCase', got %q", opts.NamingConvention)
	}
	if !opts.IncludeJSONTags {
		t.Error("Expected IncludeJSONTags to be true")
	}
	if !opts.UseOmitEmpty {
		t.Error("Expected UseOmitEmpty to be true")
	}
	if opts.GoVersion != "1.21" {
		t.Errorf("Expected GoVersion '1.21', got %q", opts.GoVersion)
	}

	// Test that slices and maps are initialized
	if opts.ExtraImports == nil {
		t.Error("Expected ExtraImports to be initialized")
	}
	if opts.CustomTypeMappings == nil {
		t.Error("Expected CustomTypeMappings to be initialized")
	}
	if opts.StructTagOptions == nil {
		t.Error("Expected StructTagOptions to be initialized")
	}
}

func TestGoOptions_Validate_ValidOptions(t *testing.T) {
	opts := DefaultGoOptions()
	if err := opts.Validate(); err != nil {
		t.Errorf("Default options should be valid, got error: %v", err)
	}
}

func TestGoOptions_Validate_InvalidOutputStyle(t *testing.T) {
	opts := DefaultGoOptions()
	opts.OutputStyle = "invalid"

	err := opts.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid output style")
		return
	}

	optErr, ok := err.(*OptionsError)
	if !ok {
		t.Errorf("Expected OptionsError, got %T", err)
		return
	}

	if optErr.Field != "OutputStyle" {
		t.Errorf("Expected error field 'OutputStyle', got %q", optErr.Field)
	}
}

func TestGoOptions_Validate_EmptyPackageName(t *testing.T) {
	opts := DefaultGoOptions()
	opts.PackageName = ""

	err := opts.Validate()
	if err == nil {
		t.Error("Expected validation error for empty package name")
		return
	}

	optErr, ok := err.(*OptionsError)
	if !ok {
		t.Errorf("Expected OptionsError, got %T", err)
		return
	}

	if optErr.Field != "PackageName" {
		t.Errorf("Expected error field 'PackageName', got %q", optErr.Field)
	}
}

func TestGoOptions_Validate_InvalidNamingConvention(t *testing.T) {
	opts := DefaultGoOptions()
	opts.NamingConvention = "snake_case"

	err := opts.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid naming convention")
		return
	}

	optErr, ok := err.(*OptionsError)
	if !ok {
		t.Errorf("Expected OptionsError, got %T", err)
		return
	}

	if optErr.Field != "NamingConvention" {
		t.Errorf("Expected error field 'NamingConvention', got %q", optErr.Field)
	}
}

func TestGoOptions_SetOption_StringOptions(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		checkFn func(GoOptions) string
	}{
		{"output_style", "interface", func(o GoOptions) string { return o.OutputStyle }},
		{"package_name", "mypackage", func(o GoOptions) string { return o.PackageName }},
		{"naming_convention", "camelCase", func(o GoOptions) string { return o.NamingConvention }},
		{"json_tag_style", "camelCase", func(o GoOptions) string { return o.JSONTagStyle }},
		{"validation_tag_style", "ozzo", func(o GoOptions) string { return o.ValidationTagStyle }},
		{"indent_style", "spaces", func(o GoOptions) string { return o.IndentStyle }},
		{"import_style", "grouped", func(o GoOptions) string { return o.ImportStyle }},
		{"enum_style", "type", func(o GoOptions) string { return o.EnumStyle }},
		{"union_style", "embedded", func(o GoOptions) string { return o.UnionStyle }},
		{"go_version", "1.22", func(o GoOptions) string { return o.GoVersion }},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.SetOption(tt.key, tt.value)

			if got := tt.checkFn(opts); got != tt.value {
				t.Errorf("SetOption(%q, %q): expected %q, got %q", tt.key, tt.value, tt.value, got)
			}
		})
	}
}

func TestGoOptions_SetOption_BoolOptions(t *testing.T) {
	tests := []struct {
		key     string
		value   bool
		checkFn func(GoOptions) bool
	}{
		{"include_comments", false, func(o GoOptions) bool { return o.IncludeComments }},
		{"include_examples", true, func(o GoOptions) bool { return o.IncludeExamples }},
		{"include_json_tags", false, func(o GoOptions) bool { return o.IncludeJSONTags }},
		{"include_validation_tags", true, func(o GoOptions) bool { return o.IncludeValidationTags }},
		{"include_xml_tags", true, func(o GoOptions) bool { return o.IncludeXMLTags }},
		{"include_yaml_tags", true, func(o GoOptions) bool { return o.IncludeYAMLTags }},
		{"use_pointers", false, func(o GoOptions) bool { return o.UsePointers }},
		{"use_omit_empty", false, func(o GoOptions) bool { return o.UseOmitEmpty }},
		{"include_imports", false, func(o GoOptions) bool { return o.IncludeImports }},
		{"generate_interfaces", true, func(o GoOptions) bool { return o.GenerateInterfaces }},
		{"generate_constructors", true, func(o GoOptions) bool { return o.GenerateConstructors }},
		{"generate_validators", true, func(o GoOptions) bool { return o.GenerateValidators }},
		{"generate_stringers", true, func(o GoOptions) bool { return o.GenerateStringers }},
		{"generate_getters", true, func(o GoOptions) bool { return o.GenerateGetters }},
		{"generate_setters", true, func(o GoOptions) bool { return o.GenerateSetters }},
		{"use_generics", true, func(o GoOptions) bool { return o.UseGenerics }},
		{"generate_enums", false, func(o GoOptions) bool { return o.GenerateEnums }},
		{"generate_unions", true, func(o GoOptions) bool { return o.GenerateUnions }},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			opts := DefaultGoOptions()
			opts.SetOption(tt.key, tt.value)

			if got := tt.checkFn(opts); got != tt.value {
				t.Errorf("SetOption(%q, %v): expected %v, got %v", tt.key, tt.value, tt.value, got)
			}
		})
	}
}

func TestGoOptions_SetOption_IntOptions(t *testing.T) {
	opts := DefaultGoOptions()
	opts.SetOption("indent_size", 2)

	if opts.IndentSize != 2 {
		t.Errorf("SetOption(indent_size, 2): expected 2, got %d", opts.IndentSize)
	}
}

func TestGoOptions_SetOption_SliceOptions(t *testing.T) {
	opts := DefaultGoOptions()
	imports := []string{"fmt", "strings"}
	opts.SetOption("extra_imports", imports)

	if !reflect.DeepEqual(opts.ExtraImports, imports) {
		t.Errorf("SetOption(extra_imports): expected %v, got %v", imports, opts.ExtraImports)
	}
}

func TestGoOptions_SetOption_MapOptions(t *testing.T) {
	opts := DefaultGoOptions()

	// Test custom type mappings
	typeMappings := map[string]string{"UUID": "string", "Time": "time.Time"}
	opts.SetOption("custom_type_mappings", typeMappings)

	if !reflect.DeepEqual(opts.CustomTypeMappings, typeMappings) {
		t.Errorf("SetOption(custom_type_mappings): expected %v, got %v", typeMappings, opts.CustomTypeMappings)
	}

	// Test struct tag options
	tagOptions := map[string]string{"db": "primary_key", "validate": "required"}
	opts.SetOption("struct_tag_options", tagOptions)

	if !reflect.DeepEqual(opts.StructTagOptions, tagOptions) {
		t.Errorf("SetOption(struct_tag_options): expected %v, got %v", tagOptions, opts.StructTagOptions)
	}
}

func TestGoOptions_SetOption_InvalidTypes(t *testing.T) {
	opts := DefaultGoOptions()
	original := opts.Clone()

	// Test invalid type for string option
	opts.SetOption("output_style", 123)
	if opts.OutputStyle != original.OutputStyle {
		t.Error("SetOption with invalid type should not change string option")
	}

	// Test invalid type for bool option
	opts.SetOption("include_comments", "true")
	if opts.IncludeComments != original.IncludeComments {
		t.Error("SetOption with invalid type should not change bool option")
	}

	// Test invalid type for int option
	opts.SetOption("indent_size", "4")
	if opts.IndentSize != original.IndentSize {
		t.Error("SetOption with invalid type should not change int option")
	}
}

func TestGoOptions_SetOption_UnknownKey(t *testing.T) {
	opts := DefaultGoOptions()
	original := opts.Clone()

	opts.SetOption("unknown_key", "value")

	if !reflect.DeepEqual(opts, original) {
		t.Error("SetOption with unknown key should not change options")
	}
}

func TestGoOptions_Clone(t *testing.T) {
	original := DefaultGoOptions()
	original.ExtraImports = []string{"fmt", "strings"}
	original.CustomTypeMappings = map[string]string{"UUID": "string"}
	original.StructTagOptions = map[string]string{"db": "primary_key"}

	cloned := original.Clone()

	// Test equality
	if !reflect.DeepEqual(original, cloned) {
		t.Error("Clone should create equal copy")
	}

	// Test deep copy of slices
	cloned.ExtraImports[0] = "log"
	if original.ExtraImports[0] == "log" {
		t.Error("Clone should deep copy slices")
	}

	// Test deep copy of maps
	cloned.CustomTypeMappings["UUID"] = "uuid.UUID"
	if original.CustomTypeMappings["UUID"] == "uuid.UUID" {
		t.Error("Clone should deep copy CustomTypeMappings")
	}

	cloned.StructTagOptions["db"] = "foreign_key"
	if original.StructTagOptions["db"] == "foreign_key" {
		t.Error("Clone should deep copy StructTagOptions")
	}
}

func TestOptionsError_Error(t *testing.T) {
	err := &OptionsError{
		Field:   "OutputStyle",
		Value:   "invalid",
		Message: "unsupported output style",
		Valid:   []string{"struct", "interface", "type_alias"},
	}

	expected := "invalid Go option for field OutputStyle: invalid - unsupported output style (valid values: struct, interface, type_alias)"
	if err.Error() != expected {
		t.Errorf("OptionsError.Error() = %q, expected %q", err.Error(), expected)
	}
}

func TestOptionsError_Error_EmptyValid(t *testing.T) {
	err := &OptionsError{
		Field:   "PackageName",
		Value:   "",
		Message: "package name cannot be empty",
		Valid:   []string{},
	}

	expected := "invalid Go option for field PackageName:  - package name cannot be empty"
	if err.Error() != expected {
		t.Errorf("OptionsError.Error() = %q, expected %q", err.Error(), expected)
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"test"}, "test", true},
		{[]string{"test"}, "Test", false}, // case sensitive
	}

	for _, tt := range tests {
		if got := contains(tt.slice, tt.item); got != tt.expected {
			t.Errorf("contains(%v, %q) = %v, expected %v", tt.slice, tt.item, got, tt.expected)
		}
	}
}
