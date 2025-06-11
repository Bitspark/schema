package schema

import (
	"testing"
)

// Test types for generic patterns
type TestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name" schema:"minlen=1,maxlen=50"`
}

type TestError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func TestList_Basic(t *testing.T) {
	schema := List[TestUser]().Build()

	if schema.Type() != TypeArray {
		t.Fatalf("Expected TypeArray, got %s", schema.Type())
	}

	arraySchema, ok := schema.(*ArraySchema)
	if !ok {
		t.Fatalf("Expected *ArraySchema, got %T", schema)
	}

	// Test item schema is correct
	if arraySchema.itemSchema.Type() != TypeObject {
		t.Errorf("Expected item schema to be TypeObject, got %s", arraySchema.itemSchema.Type())
	}
}

func TestList_WithValidation(t *testing.T) {
	schema := List[TestUser]().
		MinItems(1).
		MaxItems(10).
		UniqueItems().
		Description("List of users").
		Build()

	arraySchema := schema.(*ArraySchema)

	if arraySchema.minItems == nil || *arraySchema.minItems != 1 {
		t.Errorf("Expected minItems=1, got %v", arraySchema.minItems)
	}

	if arraySchema.maxItems == nil || *arraySchema.maxItems != 10 {
		t.Errorf("Expected maxItems=10, got %v", arraySchema.maxItems)
	}

	if !arraySchema.uniqueItems {
		t.Error("Expected uniqueItems=true")
	}

	if arraySchema.metadata.Description != "List of users" {
		t.Errorf("Expected description 'List of users', got '%s'", arraySchema.metadata.Description)
	}
}

func TestList_Validation(t *testing.T) {
	schema := List[TestUser]().MinItems(1).MaxItems(2).Build()

	// Test valid list (convert to []any for validation)
	validList := []any{
		map[string]any{
			"id":   1,
			"name": "Alice",
		},
	}

	result := schema.Validate(validList)
	if !result.Valid {
		t.Errorf("Expected valid result for valid list, got errors: %v", result.Errors)
	}

	// Test empty list (violates minItems)
	emptyList := []any{}
	result = schema.Validate(emptyList)
	if result.Valid {
		t.Error("Expected invalid result for empty list")
	}
}

func TestOptional_Basic(t *testing.T) {
	schema := Optional[TestUser]().Build()

	if schema.Type() != TypeOptional {
		t.Fatalf("Expected TypeOptional, got %s", schema.Type())
	}

	optionalSchema, ok := schema.(*OptionalSchema[TestUser])
	if !ok {
		t.Fatalf("Expected *OptionalSchema[TestUser], got %T", schema)
	}

	// Test item schema is correct
	if optionalSchema.itemSchema.Type() != TypeObject {
		t.Errorf("Expected item schema to be TypeObject, got %s", optionalSchema.itemSchema.Type())
	}
}

func TestOptional_Validation(t *testing.T) {
	schema := Optional[TestUser]().Build()

	// Test nil value (should be valid)
	result := schema.Validate(nil)
	if !result.Valid {
		t.Errorf("Expected valid result for nil value, got errors: %v", result.Errors)
	}

	// Test valid user (convert to map for validation)
	validUser := map[string]any{
		"id":   1,
		"name": "Alice",
	}
	result = schema.Validate(validUser)
	if !result.Valid {
		t.Errorf("Expected valid result for valid user, got errors: %v", result.Errors)
	}

	// Test invalid user (empty name violates minlen)
	invalidUser := map[string]any{
		"id":   1,
		"name": "",
	}
	result = schema.Validate(invalidUser)
	if result.Valid {
		t.Error("Expected invalid result for user with empty name")
	}
}

func TestOptional_JSONSchema(t *testing.T) {
	schema := Optional[TestUser]().Description("Optional user").Build()

	jsonSchema := schema.ToJSONSchema()

	// Should have oneOf with null and user schema
	oneOf, ok := jsonSchema["oneOf"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected oneOf to be []map[string]any, got %T", jsonSchema["oneOf"])
	}

	if len(oneOf) != 2 {
		t.Fatalf("Expected oneOf to have 2 items, got %d", len(oneOf))
	}

	// First should be null type
	if oneOf[0]["type"] != "null" {
		t.Errorf("Expected first oneOf item to be null type, got %v", oneOf[0]["type"])
	}

	// Second should be object type
	if oneOf[1]["type"] != "object" {
		t.Errorf("Expected second oneOf item to be object type, got %v", oneOf[1]["type"])
	}
}

func TestResult_Basic(t *testing.T) {
	schema := Result[TestUser, TestError]().Build()

	if schema.Type() != TypeResult {
		t.Fatalf("Expected TypeResult, got %s", schema.Type())
	}

	resultSchema, ok := schema.(*ResultSchema[TestUser, TestError])
	if !ok {
		t.Fatalf("Expected *ResultSchema[TestUser, TestError], got %T", schema)
	}

	// Test schemas are correct
	if resultSchema.successSchema.Type() != TypeObject {
		t.Errorf("Expected success schema to be TypeObject, got %s", resultSchema.successSchema.Type())
	}

	if resultSchema.errorSchema.Type() != TypeObject {
		t.Errorf("Expected error schema to be TypeObject, got %s", resultSchema.errorSchema.Type())
	}
}

func TestResult_JSONSchema(t *testing.T) {
	schema := Result[TestUser, TestError]().Description("User operation result").Build()

	jsonSchema := schema.ToJSONSchema()

	// Should have oneOf with success and error schemas
	oneOf, ok := jsonSchema["oneOf"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected oneOf to be []map[string]any, got %T", jsonSchema["oneOf"])
	}

	if len(oneOf) != 2 {
		t.Fatalf("Expected oneOf to have 2 items, got %d", len(oneOf))
	}

	// Check success schema structure
	successSchema := oneOf[0]
	if successSchema["type"] != "object" {
		t.Errorf("Expected success schema to be object type, got %v", successSchema["type"])
	}

	successProps, ok := successSchema["properties"].(map[string]any)
	if !ok || successProps["success"] == nil {
		t.Error("Expected success schema to have 'success' property")
	}

	// Check error schema structure
	errorSchema := oneOf[1]
	if errorSchema["type"] != "object" {
		t.Errorf("Expected error schema to be object type, got %v", errorSchema["type"])
	}

	errorProps, ok := errorSchema["properties"].(map[string]any)
	if !ok || errorProps["error"] == nil {
		t.Error("Expected error schema to have 'error' property")
	}
}

func TestMap_Basic(t *testing.T) {
	schema := Map[string, TestUser]().Build()

	if schema.Type() != TypeMap {
		t.Fatalf("Expected TypeMap, got %s", schema.Type())
	}

	mapSchema, ok := schema.(*MapSchema[string, TestUser])
	if !ok {
		t.Fatalf("Expected *MapSchema[string, TestUser], got %T", schema)
	}

	// Test schemas are correct
	if mapSchema.keySchema.Type() != TypeString {
		t.Errorf("Expected key schema to be TypeString, got %s", mapSchema.keySchema.Type())
	}

	if mapSchema.valueSchema.Type() != TypeObject {
		t.Errorf("Expected value schema to be TypeObject, got %s", mapSchema.valueSchema.Type())
	}
}

func TestMap_WithValidation(t *testing.T) {
	schema := Map[string, TestUser]().
		MinItems(1).
		MaxItems(5).
		Description("User map").
		Build()

	mapSchema := schema.(*MapSchema[string, TestUser])

	if mapSchema.minItems == nil || *mapSchema.minItems != 1 {
		t.Errorf("Expected minItems=1, got %v", mapSchema.minItems)
	}

	if mapSchema.maxItems == nil || *mapSchema.maxItems != 5 {
		t.Errorf("Expected maxItems=5, got %v", mapSchema.maxItems)
	}

	if mapSchema.metadata.Description != "User map" {
		t.Errorf("Expected description 'User map', got '%s'", mapSchema.metadata.Description)
	}
}

func TestMap_Validation(t *testing.T) {
	schema := Map[string, TestUser]().MinItems(1).MaxItems(2).Build()

	// Test valid map
	validMap := map[string]map[string]any{
		"user1": {
			"id":   1,
			"name": "Alice",
		},
	}

	result := schema.Validate(validMap)
	if !result.Valid {
		t.Errorf("Expected valid result for valid map, got errors: %v", result.Errors)
	}

	// Test empty map (violates minItems)
	emptyMap := map[string]map[string]any{}
	result = schema.Validate(emptyMap)
	if result.Valid {
		t.Error("Expected invalid result for empty map")
	}

	// Test map with invalid value
	invalidMap := map[string]map[string]any{
		"user1": {
			"id":   1,
			"name": "", // Empty name violates minlen
		},
	}
	result = schema.Validate(invalidMap)
	if result.Valid {
		t.Error("Expected invalid result for map with invalid value")
	}
}

func TestMap_JSONSchema(t *testing.T) {
	schema := Map[string, TestUser]().Description("User mapping").Build()

	jsonSchema := schema.ToJSONSchema()

	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type='object', got %v", jsonSchema["type"])
	}

	// Should have additionalProperties with user schema
	additionalProps, ok := jsonSchema["additionalProperties"].(map[string]any)
	if !ok {
		t.Fatalf("Expected additionalProperties to be map[string]any, got %T", jsonSchema["additionalProperties"])
	}

	if additionalProps["type"] != "object" {
		t.Errorf("Expected additionalProperties type='object', got %v", additionalProps["type"])
	}
}

func TestUnion_Basic(t *testing.T) {
	schema := Union[string, int]().Build()

	if schema.Type() != TypeUnion {
		t.Fatalf("Expected TypeUnion, got %s", schema.Type())
	}

	unionSchema, ok := schema.(*UnionSchema)
	if !ok {
		t.Fatalf("Expected *UnionSchema, got %T", schema)
	}

	if len(unionSchema.schemas) != 2 {
		t.Fatalf("Expected 2 schemas in union, got %d", len(unionSchema.schemas))
	}

	// Test schemas are correct types
	if unionSchema.schemas[0].Type() != TypeString {
		t.Errorf("Expected first schema to be TypeString, got %s", unionSchema.schemas[0].Type())
	}

	if unionSchema.schemas[1].Type() != TypeInteger {
		t.Errorf("Expected second schema to be TypeInteger, got %s", unionSchema.schemas[1].Type())
	}
}

func TestUnion_Validation(t *testing.T) {
	schema := Union[string, int]().Build()

	// Test string value (should match first schema)
	result := schema.Validate("hello")
	if !result.Valid {
		t.Errorf("Expected valid result for string value, got errors: %v", result.Errors)
	}

	// Test int value (should match second schema)
	result = schema.Validate(42)
	if !result.Valid {
		t.Errorf("Expected valid result for int value, got errors: %v", result.Errors)
	}

	// Test invalid value (should match neither schema)
	result = schema.Validate(3.14)
	if result.Valid {
		t.Error("Expected invalid result for float value")
	}
}

func TestUnion_JSONSchema(t *testing.T) {
	schema := Union[string, int]().Description("String or integer").Build()

	jsonSchema := schema.ToJSONSchema()

	oneOf, ok := jsonSchema["oneOf"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected oneOf to be []map[string]any, got %T", jsonSchema["oneOf"])
	}

	if len(oneOf) != 2 {
		t.Fatalf("Expected oneOf to have 2 items, got %d", len(oneOf))
	}

	// Check string schema
	if oneOf[0]["type"] != "string" {
		t.Errorf("Expected first oneOf item to be string type, got %v", oneOf[0]["type"])
	}

	// Check integer schema
	if oneOf[1]["type"] != "integer" {
		t.Errorf("Expected second oneOf item to be integer type, got %v", oneOf[1]["type"])
	}
}

// Test convenience functions
func TestStringList(t *testing.T) {
	schema := StringList().Build()

	if schema.Type() != TypeArray {
		t.Fatalf("Expected TypeArray, got %s", schema.Type())
	}

	arraySchema := schema.(*ArraySchema)
	if arraySchema.itemSchema.Type() != TypeString {
		t.Errorf("Expected item schema to be TypeString, got %s", arraySchema.itemSchema.Type())
	}
}

func TestStringOptional(t *testing.T) {
	schema := StringOptional().Build()

	if schema.Type() != TypeOptional {
		t.Fatalf("Expected TypeOptional, got %s", schema.Type())
	}

	// Test validation
	result := schema.Validate(nil)
	if !result.Valid {
		t.Error("Expected nil to be valid for optional string")
	}

	result = schema.Validate("hello")
	if !result.Valid {
		t.Error("Expected string to be valid for optional string")
	}
}

func TestStringMap(t *testing.T) {
	schema := StringMap().Build()

	if schema.Type() != TypeMap {
		t.Fatalf("Expected TypeMap, got %s", schema.Type())
	}

	mapSchema := schema.(*MapSchema[string, string])
	if mapSchema.keySchema.Type() != TypeString {
		t.Errorf("Expected key schema to be TypeString, got %s", mapSchema.keySchema.Type())
	}

	if mapSchema.valueSchema.Type() != TypeString {
		t.Errorf("Expected value schema to be TypeString, got %s", mapSchema.valueSchema.Type())
	}
}

// Test complex combinations
func TestComplexCombination(t *testing.T) {
	// Create a schema for: Optional[List[Map[string, User]]]
	innerSchema := Map[string, TestUser]().Build()
	listSchema := List[TestUser]().Build() // This will be replaced

	// Build the actual complex schema manually for now
	// In a real implementation, we might want: Optional[List[Map[string, TestUser]]]()
	optionalSchema := Optional[TestUser]().Build()

	if optionalSchema.Type() != TypeOptional {
		t.Fatalf("Expected TypeOptional, got %s", optionalSchema.Type())
	}

	// Just verify the components work individually
	if innerSchema.Type() != TypeMap {
		t.Errorf("Expected TypeMap for inner schema, got %s", innerSchema.Type())
	}

	if listSchema.Type() != TypeArray {
		t.Errorf("Expected TypeArray for list schema, got %s", listSchema.Type())
	}
}

// Benchmark generic schema creation
func BenchmarkListCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = List[TestUser]().Build()
	}
}

func BenchmarkOptionalCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Optional[TestUser]().Build()
	}
}

func BenchmarkMapCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Map[string, TestUser]().Build()
	}
}
