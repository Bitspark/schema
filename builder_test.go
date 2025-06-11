package schema

import (
	"testing"
)

func TestStringBuilderTag(t *testing.T) {
	schema := String().
		Tag("input").
		Tag("validation").
		Build().(*StringSchema)
	
	if len(schema.metadata.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(schema.metadata.Tags))
	}
	
	if schema.metadata.Tags[0] != "input" {
		t.Errorf("Expected first tag 'input', got '%s'", schema.metadata.Tags[0])
	}
	
	if schema.metadata.Tags[1] != "validation" {
		t.Errorf("Expected second tag 'validation', got '%s'", schema.metadata.Tags[1])
	}
}

func TestObjectBuilderExample(t *testing.T) {
	example := map[string]any{
		"name": "John",
		"age":  30,
	}
	
	schema := Object().
		Example(example).
		Build().(*ObjectSchema)
	
	if len(schema.metadata.Examples) == 0 {
		t.Error("Expected example to be set")
	}
	
	exampleMap := schema.metadata.Examples[0].(map[string]any)
	if exampleMap["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", exampleMap["name"])
	}
	
	if exampleMap["age"] != 30 {
		t.Errorf("Expected age 30, got %v", exampleMap["age"])
	}
}

func TestObjectBuilderTag(t *testing.T) {
	schema := Object().
		Tag("entity").
		Tag("user").
		Build().(*ObjectSchema)
	
	if len(schema.metadata.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(schema.metadata.Tags))
	}
	
	if schema.metadata.Tags[0] != "entity" {
		t.Errorf("Expected first tag 'entity', got '%s'", schema.metadata.Tags[0])
	}
	
	if schema.metadata.Tags[1] != "user" {
		t.Errorf("Expected second tag 'user', got '%s'", schema.metadata.Tags[1])
	}
}

func TestNumberBuilderDescription(t *testing.T) {
	description := "A numeric value for calculations"
	
	schema := Number().
		Description(description).
		Build().(*NumberSchema)
	
	if schema.metadata.Description != description {
		t.Errorf("Expected description '%s', got '%s'", description, schema.metadata.Description)
	}
}

func TestNumberBuilderName(t *testing.T) {
	name := "price"
	
	schema := Number().
		Name(name).
		Build().(*NumberSchema)
	
	if schema.metadata.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, schema.metadata.Name)
	}
}

func TestIntegerBuilderName(t *testing.T) {
	name := "count"
	
	schema := Integer().
		Name(name).
		Build().(*IntegerSchema)
	
	if schema.metadata.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, schema.metadata.Name)
	}
}

func TestArrayBuilderExample(t *testing.T) {
	example := []any{"apple", "banana", "cherry"}
	
	schema := Array().
		Example(example).
		Build().(*ArraySchema)
	
	if len(schema.metadata.Examples) == 0 {
		t.Error("Expected example to be set")
	}
	
	exampleSlice := schema.metadata.Examples[0].([]any)
	if len(exampleSlice) != 3 {
		t.Errorf("Expected 3 items in example, got %d", len(exampleSlice))
	}
	
	if exampleSlice[0] != "apple" {
		t.Errorf("Expected first item 'apple', got %v", exampleSlice[0])
	}
}

func TestArrayBuilderName(t *testing.T) {
	name := "fruits"
	
	schema := Array().
		Name(name).
		Build().(*ArraySchema)
	
	if schema.metadata.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, schema.metadata.Name)
	}
}

func TestArrayBuilderDescription(t *testing.T) {
	description := "A list of items"
	
	schema := Array().
		Description(description).
		Build().(*ArraySchema)
	
	if schema.metadata.Description != description {
		t.Errorf("Expected description '%s', got '%s'", description, schema.metadata.Description)
	}
}

func TestUnionSchemaWithMetadata(t *testing.T) {
	// Create a UnionSchema directly
	stringSchema := String().Build()
	numberSchema := Number().Build()
	
	original := &UnionSchema{
		metadata: SchemaMetadata{},
		schemas:  []Schema{stringSchema, numberSchema},
	}
	
	metadata := SchemaMetadata{
		Name:        "test-union",
		Description: "Test union schema",
		Tags:        []string{"flexible", "multi-type"},
	}
	
	result := original.WithMetadata(metadata)
	resultUnion := result.(*UnionSchema)
	
	// Verify original is not modified
	if original.metadata.Name == "test-union" {
		t.Error("Original schema was modified, expected clone")
	}
	
	// Verify clone has the metadata
	if resultUnion.metadata.Name != "test-union" {
		t.Errorf("Expected name 'test-union', got %s", resultUnion.metadata.Name)
	}
	
	if resultUnion.metadata.Description != "Test union schema" {
		t.Errorf("Expected description 'Test union schema', got %s", resultUnion.metadata.Description)
	}
	
	if len(resultUnion.metadata.Tags) != 2 || resultUnion.metadata.Tags[0] != "flexible" {
		t.Errorf("Expected tags [flexible, multi-type], got %v", resultUnion.metadata.Tags)
	}
}