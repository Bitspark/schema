package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// Test structs for struct generation
type User struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name" schema:"minlen=1,maxlen=100,desc=User display name"`
	Email    string      `json:"email" schema:"email,desc=User email address"`
	Age      *int        `json:"age,omitempty" schema:"min=0,max=150,desc=User age in years"`
	IsActive bool        `json:"is_active"`
	Tags     []string    `json:"tags,omitempty" schema:"maxitems=10,desc=User tags"`
	Profile  UserProfile `json:"profile"`
}

type UserProfile struct {
	Bio       string    `json:"bio,omitempty" schema:"maxlen=500"`
	Website   *string   `json:"website,omitempty" schema:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Company struct {
	Name      string `json:"name" schema:"required,minlen=1,maxlen=200"`
	Industry  string `json:"industry" schema:"enum=tech|finance|healthcare|retail"`
	Employees int    `json:"employees" schema:"min=1,max=1000000"`
}

func TestFromStruct_2(t *testing.T) {
	schema := FromStruct[Company]()

	schemaJSON, _ := json.Marshal(schema.ToJSONSchema())

	fmt.Println(string(schemaJSON))
}

func TestFromStruct_Basic(t *testing.T) {
	schema := FromStruct[User]()

	if schema.Type() != TypeObject {
		t.Fatalf("Expected TypeObject, got %s", schema.Type())
	}

	objSchema, ok := schema.(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected *ObjectSchema, got %T", schema)
	}

	// Test basic properties exist
	expectedProps := []string{"id", "name", "email", "age", "is_active", "tags", "profile"}
	for _, prop := range expectedProps {
		if _, exists := objSchema.properties[prop]; !exists {
			t.Errorf("Property '%s' not found in schema", prop)
		}
	}

	// Test required fields
	expectedRequired := []string{"id", "name", "email", "is_active", "profile"}
	for _, req := range expectedRequired {
		found := false
		for _, r := range objSchema.required {
			if r == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required field '%s' not found in required list", req)
		}
	}
}

func TestFromStruct_StringValidation(t *testing.T) {
	schema := FromStruct[User]()
	objSchema := schema.(*ObjectSchema)

	// Test name field validation
	nameSchema, ok := objSchema.properties["name"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for name field, got %T", objSchema.properties["name"])
	}

	if nameSchema.minLength == nil || *nameSchema.minLength != 1 {
		t.Errorf("Expected minLength=1 for name field, got %v", nameSchema.minLength)
	}

	if nameSchema.maxLength == nil || *nameSchema.maxLength != 100 {
		t.Errorf("Expected maxLength=100 for name field, got %v", nameSchema.maxLength)
	}

	if nameSchema.metadata.Description != "User display name" {
		t.Errorf("Expected description 'User display name', got '%s'", nameSchema.metadata.Description)
	}

	// Test email field format
	emailSchema, ok := objSchema.properties["email"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for email field, got %T", objSchema.properties["email"])
	}

	if emailSchema.format != "email" {
		t.Errorf("Expected format='email' for email field, got '%s'", emailSchema.format)
	}
}

func TestFromStruct_NumberValidation(t *testing.T) {
	schema := FromStruct[User]()
	objSchema := schema.(*ObjectSchema)

	// Test age field validation (should be IntegerSchema because it's *int)
	ageSchema, ok := objSchema.properties["age"].(*IntegerSchema)
	if !ok {
		t.Fatalf("Expected IntegerSchema for age field, got %T", objSchema.properties["age"])
	}

	if ageSchema.minimum == nil || *ageSchema.minimum != 0 {
		t.Errorf("Expected minimum=0 for age field, got %v", ageSchema.minimum)
	}

	if ageSchema.maximum == nil || *ageSchema.maximum != 150 {
		t.Errorf("Expected maximum=150 for age field, got %v", ageSchema.maximum)
	}
}

func TestFromStruct_ArrayValidation(t *testing.T) {
	schema := FromStruct[User]()
	objSchema := schema.(*ObjectSchema)

	// Test tags field validation
	tagsSchema, ok := objSchema.properties["tags"].(*ArraySchema)
	if !ok {
		t.Fatalf("Expected ArraySchema for tags field, got %T", objSchema.properties["tags"])
	}

	if tagsSchema.maxItems == nil || *tagsSchema.maxItems != 10 {
		t.Errorf("Expected maxItems=10 for tags field, got %v", tagsSchema.maxItems)
	}

	// Test item schema is string
	if tagsSchema.itemSchema.Type() != TypeString {
		t.Errorf("Expected string items for tags array, got %s", tagsSchema.itemSchema.Type())
	}
}

func TestFromStruct_NestedObject(t *testing.T) {
	schema := FromStruct[User]()
	objSchema := schema.(*ObjectSchema)

	// Test profile field (nested object)
	profileSchema, ok := objSchema.properties["profile"].(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema for profile field, got %T", objSchema.properties["profile"])
	}

	// Test nested properties
	expectedNestedProps := []string{"bio", "website", "created_at"}
	for _, prop := range expectedNestedProps {
		if _, exists := profileSchema.properties[prop]; !exists {
			t.Errorf("Nested property '%s' not found in profile schema", prop)
		}
	}

	// Test bio maxlength
	bioSchema, ok := profileSchema.properties["bio"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for bio field, got %T", profileSchema.properties["bio"])
	}

	if bioSchema.maxLength == nil || *bioSchema.maxLength != 500 {
		t.Errorf("Expected maxLength=500 for bio field, got %v", bioSchema.maxLength)
	}
}

func TestFromStruct_EnumValidation(t *testing.T) {
	schema := FromStruct[Company]()
	objSchema := schema.(*ObjectSchema)

	// Test industry field enum
	industrySchema, ok := objSchema.properties["industry"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for industry field, got %T", objSchema.properties["industry"])
	}

	expectedEnums := []string{"tech", "finance", "healthcare", "retail"}
	if len(industrySchema.enumValues) != len(expectedEnums) {
		t.Errorf("Expected %d enum values, got %d", len(expectedEnums), len(industrySchema.enumValues))
	}

	for i, expected := range expectedEnums {
		if i >= len(industrySchema.enumValues) || industrySchema.enumValues[i] != expected {
			t.Errorf("Expected enum value '%s' at index %d, got '%s'", expected, i, industrySchema.enumValues[i])
		}
	}
}

func TestFromStruct_RequiredFields(t *testing.T) {
	schema := FromStruct[Company]()
	objSchema := schema.(*ObjectSchema)

	// Company.Name should be required due to explicit "required" tag
	nameRequired := false
	for _, req := range objSchema.required {
		if req == "name" {
			nameRequired = true
			break
		}
	}
	if !nameRequired {
		t.Error("Expected 'name' field to be required due to explicit schema tag")
	}

	// All non-pointer fields should be required
	expectedRequired := []string{"name", "industry", "employees"}
	if len(objSchema.required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(objSchema.required))
	}
}

func TestFromStruct_JSONSchemaConversion(t *testing.T) {
	schema := FromStruct[User]()

	jsonSchema := schema.ToJSONSchema()

	// Test basic structure
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type='object', got %v", jsonSchema["type"])
	}

	properties, ok := jsonSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("Expected properties to be map[string]any, got %T", jsonSchema["properties"])
	}

	// Test email field has correct format in JSON schema
	if emailProp, exists := properties["email"]; exists {
		if emailMap, ok := emailProp.(map[string]any); ok {
			if emailMap["format"] != "email" {
				t.Errorf("Expected email format in JSON schema, got %v", emailMap["format"])
			}
		}
	}
}

// Test caching behavior
func TestFromStruct_Caching(t *testing.T) {
	// Generate schema twice for same type
	schema1 := FromStruct[User]()
	schema2 := FromStruct[User]()

	// Should be functionally equivalent but different instances (due to Clone())
	if schema1 == schema2 {
		t.Error("Expected different instances due to cloning")
	}

	// Should have same structure
	if schema1.Type() != schema2.Type() {
		t.Error("Cached schemas should have same type")
	}
}

func TestValidateStructTag(t *testing.T) {
	// Test valid struct
	if err := ValidateStructTag(reflect.TypeOf(User{})); err != nil {
		t.Errorf("Expected no error for valid struct, got: %v", err)
	}

	// Test struct with invalid tag (would need to create a struct with bad tag)
	// For now, just test that the function doesn't panic
}

// Benchmark struct generation performance
func BenchmarkFromStruct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromStruct[User]()
	}
}

func BenchmarkFromStruct_WithCache(b *testing.B) {
	// Pre-populate cache
	_ = FromStruct[User]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromStruct[User]()
	}
}
