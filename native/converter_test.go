package native

import (
	"defs.dev/schema/annotation"
	"defs.dev/schema/runtime/registry"
	"reflect"
	"testing"

	"defs.dev/schema/core"
)

func TestDefaultTypeConverter_BasicTypes(t *testing.T) {
	// Setup
	annotationReg := annotation.NewRegistry()
	validatorReg := registry.NewDefaultValidatorRegistry(annotationReg)
	converter := NewDefaultTypeConverter(annotationReg, validatorReg)

	tests := []struct {
		name     string
		input    any
		wantType core.SchemaType
	}{
		{"string", "hello", core.TypeString},
		{"int", 42, core.TypeInteger},
		{"float64", 3.14, core.TypeNumber},
		{"bool", true, core.TypeBoolean},
		{"slice", []string{"a", "b"}, core.TypeArray},
		{"map", map[string]int{"key": 1}, core.TypeStructure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := converter.FromValue(tt.input)
			if err != nil {
				t.Fatalf("FromValue() error = %v", err)
			}
			if schema == nil {
				t.Fatal("FromValue() returned nil schema")
			}
			if schema.Type() != tt.wantType {
				t.Errorf("FromValue() type = %v, want %v", schema.Type(), tt.wantType)
			}
		})
	}
}

func TestDefaultTypeConverter_StructConversion(t *testing.T) {
	// Setup
	annotationReg := annotation.NewRegistry()
	validatorReg := registry.NewDefaultValidatorRegistry(annotationReg)
	converter := NewDefaultTypeConverter(annotationReg, validatorReg)

	// Test struct
	type TestStruct struct {
		Name  string `json:"name" required:"true"`
		Email string `json:"email" format:"email"`
		Age   int    `json:"age" min:"0"`
	}

	schema, err := converter.FromType(reflect.TypeOf(TestStruct{}))
	if err != nil {
		t.Fatalf("FromType() error = %v", err)
	}

	if schema.Type() != core.TypeStructure {
		t.Errorf("FromType() type = %v, want object", schema.Type())
	}

	// Basic validation that we got a schema
	if schema == nil {
		t.Error("Schema is nil")
	}
}

func TestDefaultTagParser_BasicTags(t *testing.T) {
	// Setup
	annotationReg := annotation.NewRegistry()
	parser := NewDefaultTagParser(annotationReg)

	tests := []struct {
		name    string
		tag     reflect.StructTag
		wantLen int
	}{
		{"json tag", `json:"name"`, 1},
		{"multiple tags", `json:"name" required:"true" format:"email"`, 3},
		{"empty tag", ``, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations, err := parser.ParseTags(tt.tag)
			if err != nil {
				t.Fatalf("ParseTags() error = %v", err)
			}
			if len(annotations) != tt.wantLen {
				t.Errorf("ParseTags() len = %v, want %v", len(annotations), tt.wantLen)
			}
		})
	}
}

func TestDefaultTypeConverter_Configuration(t *testing.T) {
	// Setup
	annotationReg := annotation.NewRegistry()
	validatorReg := registry.NewDefaultValidatorRegistry(annotationReg)
	converter := NewDefaultTypeConverter(annotationReg, validatorReg)

	// Test configuration
	config := converter.GetConfiguration()
	if config.MaxDepth != 10 {
		t.Errorf("GetConfiguration() MaxDepth = %v, want 10", config.MaxDepth)
	}

	// Test strict mode
	converter.SetStrictMode(true)
	config = converter.GetConfiguration()
	if !config.StrictMode {
		t.Error("SetStrictMode(true) did not set strict mode")
	}

	// Test supported tags
	tags := converter.GetSupportedTags()
	if len(tags) == 0 {
		t.Error("GetSupportedTags() returned empty slice")
	}
}
