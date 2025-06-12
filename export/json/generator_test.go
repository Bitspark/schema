package json

import (
	"encoding/json"
	"strings"
	"testing"

	"defs.dev/schema/api/core"
	"defs.dev/schema/schemas"
)

func TestJSONGenerator_Basic(t *testing.T) {
	tests := []struct {
		name     string
		schema   core.Schema
		contains []string // strings that should be present in output
	}{
		{
			name: "string schema",
			schema: schemas.NewStringSchema(schemas.StringSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "TestString",
					Description: "A test string",
				},
				MinLength: intPtr(1),
				MaxLength: intPtr(100),
				Format:    "email",
			}),
			contains: []string{`"type": "string"`, `"minLength": 1`, `"maxLength": 100`, `"format": "email"`},
		},
		{
			name: "integer schema",
			schema: schemas.NewIntegerSchema(schemas.IntegerSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "TestInteger",
					Description: "A test integer",
				},
				Minimum: int64Ptr(0),
				Maximum: int64Ptr(1000),
			}),
			contains: []string{`"type": "integer"`, `"minimum": 0`, `"maximum": 1000`},
		},
		{
			name: "boolean schema",
			schema: schemas.NewBooleanSchema(schemas.BooleanSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "TestBoolean",
					Description: "A test boolean",
				},
			}),
			contains: []string{`"type": "boolean"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()

			output, err := generator.Generate(tt.schema)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(output) == 0 {
				t.Fatal("Generate() returned empty output")
			}

			// Verify it's valid JSON
			var result map[string]any
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("Generated output is not valid JSON: %v", err)
			}

			outputStr := string(output)
			for _, contains := range tt.contains {
				if !strings.Contains(outputStr, contains) {
					t.Errorf("Output does not contain expected string %q\nOutput: %s", contains, outputStr)
				}
			}
		})
	}
}

func TestJSONGenerator_Options(t *testing.T) {
	schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Name:        "TestString",
			Description: "A test string",
		},
	})

	t.Run("default options", func(t *testing.T) {
		generator := NewGenerator()
		output, err := generator.Generate(schema)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Should be pretty printed by default
		if !strings.Contains(string(output), "\n") {
			t.Error("Expected pretty printed output")
		}
	})

	t.Run("minified output", func(t *testing.T) {
		generator := NewGenerator(WithMinifyOutput(true), WithPrettyPrint(false))
		output, err := generator.Generate(schema)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Should be minified (no whitespace)
		if strings.Contains(string(output), "\n") || strings.Contains(string(output), "  ") {
			t.Error("Expected minified output")
		}
	})

	t.Run("custom draft", func(t *testing.T) {
		generator := NewGenerator(WithDraft("draft-2019-09"))
		output, err := generator.Generate(schema)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "draft/2019-09") {
			t.Error("Expected draft-2019-09 schema URI")
		}
	})
}

func TestJSONGenerator_Interface(t *testing.T) {
	generator := NewGenerator()

	// Test generator interface methods
	if name := generator.Name(); name == "" {
		t.Error("Name() should return non-empty string")
	}

	if format := generator.Format(); format == "" {
		t.Error("Format() should return non-empty string")
	}

	// Test options
	options := generator.GetOptions()
	if options.Draft == "" {
		t.Error("GetOptions() should return valid options")
	}
}

func TestJSONGenerator_Visitor(t *testing.T) {
	generator := NewGenerator()

	// Test visitor interface methods exist (compile-time check)
	var visitor core.SchemaVisitor = generator
	if visitor == nil {
		t.Error("Generator should implement SchemaVisitor")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
