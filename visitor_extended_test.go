package schema

import (
	"testing"
)

func TestBaseVisitorMethods(t *testing.T) {
	visitor := &BaseVisitor{}

	// Test all visitor methods return nil and don't panic
	t.Run("VisitNumber", func(t *testing.T) {
		err := visitor.VisitNumber(NewNumber().Build().(*NumberSchema))
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("VisitBoolean", func(t *testing.T) {
		err := visitor.VisitBoolean(NewBoolean().Build().(*BooleanSchema))
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("VisitFunction", func(t *testing.T) {
		functionSchema := NewFunctionSchema().
			Input("test", NewString().Build()).
			Output(NewString().Build()).
			Build().(*FunctionSchema)

		err := visitor.VisitFunction(functionSchema)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})
}

func TestSchemaAcceptVisitor(t *testing.T) {
	// Test that schemas can accept visitors
	visitor := &BaseVisitor{}

	t.Run("StringSchema Accept", func(t *testing.T) {
		stringSchema := NewString().Build().(*StringSchema)
		err := stringSchema.Accept(visitor)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})
}

// Custom test visitor to verify visitor pattern works
type testVisitor struct {
	visitedTypes []SchemaType
	BaseVisitor
}

func (v *testVisitor) VisitString(schema *StringSchema) error {
	v.visitedTypes = append(v.visitedTypes, schema.Type())
	return nil
}

func (v *testVisitor) VisitNumber(schema *NumberSchema) error {
	v.visitedTypes = append(v.visitedTypes, schema.Type())
	return nil
}

func (v *testVisitor) VisitBoolean(schema *BooleanSchema) error {
	v.visitedTypes = append(v.visitedTypes, schema.Type())
	return nil
}

func TestCustomVisitorPattern(t *testing.T) {
	visitor := &testVisitor{}

	// Visit different schema types
	NewString().Build().(*StringSchema).Accept(visitor)
	NewNumber().Build().(*NumberSchema).Accept(visitor)
	NewBoolean().Build().(*BooleanSchema).Accept(visitor)

	expectedTypes := []SchemaType{TypeString, TypeNumber, TypeBoolean}
	if len(visitor.visitedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d visited types, got %d", len(expectedTypes), len(visitor.visitedTypes))
	}

	for i, expected := range expectedTypes {
		if i >= len(visitor.visitedTypes) || visitor.visitedTypes[i] != expected {
			t.Errorf("Expected type %s at index %d, got %s", expected, i, visitor.visitedTypes[i])
		}
	}
}
