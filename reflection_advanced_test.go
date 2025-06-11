// Package schema - Advanced reflection testing
//
// This file contains comprehensive tests for the advanced reflection functionality
// in reflection.go that wasn't well covered by existing tests. It includes:
//
// - Complex struct types with various schema tags
// - Tag parsing edge cases and validation
// - Type registry custom mappings
// - Embedded struct handling
// - Interface and map type handling
// - Pointer type optionality
// - JSON tag processing
// - Schema tag application (url, uuid, format, examples, etc.)
// - Helper function testing
// - Caching behavior validation
//
// These tests complement the existing reflection_test.go to provide
// comprehensive coverage of the struct reflection functionality.

package schema

import (
	"reflect"
	"testing"
	"time"
)

// Test structs for advanced reflection testing

type AdvancedUser struct {
	// Basic types with various tags
	Name        string         `json:"name" schema:"minlen=1,maxlen=100,desc=User name"`
	Email       string         `json:"email" schema:"email,desc=Email address"`
	Website     *string        `json:"website,omitempty" schema:"url,desc=Personal website"`
	Age         *int           `json:"age,omitempty" schema:"min=0,max=150"`
	Score       float64        `json:"score" schema:"min=0.0,max=100.0"`
	IsActive    bool           `json:"is_active" schema:"desc=Account status"`
	Tags        []string       `json:"tags,omitempty" schema:"maxitems=10,desc=User tags"`
	Preferences map[string]any `json:"preferences,omitempty"`

	// Enum testing
	Status string `json:"status" schema:"enum=active|inactive|pending"`

	// Pattern testing
	PhoneNumber string `json:"phone" schema:"pattern=^\\\\+?[1-9]\\\\d+$,desc=E.164 phone format"`

	// Complex nested types
	Profile   AdvancedProfile   `json:"profile"`
	Addresses []AdvancedAddress `json:"addresses,omitempty" schema:"maxitems=5"`

	// Time and custom types
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Edge cases
	Interface any `json:"data,omitempty"`
	Embedded      // Anonymous embedded struct
}

type AdvancedProfile struct {
	Bio         string            `json:"bio,omitempty" schema:"maxlen=500"`
	Avatar      string            `json:"avatar,omitempty" schema:"url"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
}

type AdvancedAddress struct {
	Street  string `json:"street" schema:"required,minlen=1"`
	City    string `json:"city" schema:"required,minlen=1"`
	Country string `json:"country" schema:"required,enum=US|CA|UK|DE|FR"`
	ZipCode string `json:"zip_code,omitempty" schema:"pattern=^[0-9]{5}(-[0-9]{4})?$"`
}

type Embedded struct {
	EmbeddedField string `json:"embedded_field" schema:"desc=Embedded field"`
}

// Custom type for type registry testing
type CustomID string

func TestFromStruct_AdvancedComplexTypes(t *testing.T) {
	schema := FromStruct[AdvancedUser]()

	objSchema, ok := schema.(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected *ObjectSchema, got %T", schema)
	}

	// Test that all fields are present
	expectedFields := []string{
		"name", "email", "website", "age", "score", "is_active",
		"tags", "preferences", "status", "phone", "profile",
		"addresses", "created_at", "updated_at", "data", "embedded_field",
	}

	for _, field := range expectedFields {
		if _, exists := objSchema.properties[field]; !exists {
			t.Errorf("Expected field '%s' not found in schema", field)
		}
	}

	// Test required fields
	expectedRequired := []string{"name", "email", "score", "is_active", "status", "phone", "profile", "created_at"}
	for _, required := range expectedRequired {
		found := false
		for _, r := range objSchema.required {
			if r == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field '%s' not found in required list", required)
		}
	}
}

func TestFromStruct_AdvancedTagParsing(t *testing.T) {
	schema := FromStruct[AdvancedUser]()
	objSchema := schema.(*ObjectSchema)

	// Test string validation with minlen/maxlen
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

	if nameSchema.metadata.Description != "User name" {
		t.Errorf("Expected description 'User name', got '%s'", nameSchema.metadata.Description)
	}

	// Test email format
	emailSchema, ok := objSchema.properties["email"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for email field, got %T", objSchema.properties["email"])
	}

	if emailSchema.format != "email" {
		t.Errorf("Expected format='email' for email field, got '%s'", emailSchema.format)
	}

	// Test enum validation
	statusSchema, ok := objSchema.properties["status"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for status field, got %T", objSchema.properties["status"])
	}

	expectedEnums := []string{"active", "inactive", "pending"}
	if len(statusSchema.enumValues) != len(expectedEnums) {
		t.Errorf("Expected %d enum values, got %d", len(expectedEnums), len(statusSchema.enumValues))
	}

	for i, expected := range expectedEnums {
		if i >= len(statusSchema.enumValues) || statusSchema.enumValues[i] != expected {
			t.Errorf("Expected enum value '%s' at index %d, got '%s'", expected, i, statusSchema.enumValues[i])
		}
	}

	// Test pattern validation
	phoneSchema, ok := objSchema.properties["phone"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for phone field, got %T", objSchema.properties["phone"])
	}

	expectedPattern := "^\\\\+?[1-9]\\\\d+$"
	if phoneSchema.pattern != expectedPattern {
		t.Errorf("Expected pattern '%s', got '%s'", expectedPattern, phoneSchema.pattern)
	}
}

func TestFromStruct_AdvancedNumberValidation(t *testing.T) {
	schema := FromStruct[AdvancedUser]()
	objSchema := schema.(*ObjectSchema)

	// Test integer validation with min/max
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

	// Test float validation
	scoreSchema, ok := objSchema.properties["score"].(*NumberSchema)
	if !ok {
		t.Fatalf("Expected NumberSchema for score field, got %T", objSchema.properties["score"])
	}

	if scoreSchema.minimum == nil || *scoreSchema.minimum != 0.0 {
		t.Errorf("Expected minimum=0.0 for score field, got %v", scoreSchema.minimum)
	}

	if scoreSchema.maximum == nil || *scoreSchema.maximum != 100.0 {
		t.Errorf("Expected maximum=100.0 for score field, got %v", scoreSchema.maximum)
	}
}

func TestFromStruct_AdvancedArrayValidation(t *testing.T) {
	schema := FromStruct[AdvancedUser]()
	objSchema := schema.(*ObjectSchema)

	// Test array with maxitems
	tagsSchema, ok := objSchema.properties["tags"].(*ArraySchema)
	if !ok {
		t.Fatalf("Expected ArraySchema for tags field, got %T", objSchema.properties["tags"])
	}

	if tagsSchema.maxItems == nil || *tagsSchema.maxItems != 10 {
		t.Errorf("Expected maxItems=10 for tags field, got %v", tagsSchema.maxItems)
	}

	// Test array items are strings
	if tagsSchema.itemSchema.Type() != TypeString {
		t.Errorf("Expected string items for tags array, got %s", tagsSchema.itemSchema.Type())
	}

	// Test nested object arrays
	addressesSchema, ok := objSchema.properties["addresses"].(*ArraySchema)
	if !ok {
		t.Fatalf("Expected ArraySchema for addresses field, got %T", objSchema.properties["addresses"])
	}

	if addressesSchema.maxItems == nil || *addressesSchema.maxItems != 5 {
		t.Errorf("Expected maxItems=5 for addresses field, got %v", addressesSchema.maxItems)
	}

	// Check that array items are Address objects
	addressItemSchema, ok := addressesSchema.itemSchema.(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema for address items, got %T", addressesSchema.itemSchema)
	}

	// Verify Address object has expected properties
	expectedAddressProps := []string{"street", "city", "country", "zip_code"}
	for _, prop := range expectedAddressProps {
		if _, exists := addressItemSchema.properties[prop]; !exists {
			t.Errorf("Expected property '%s' not found in Address schema", prop)
		}
	}
}

func TestFromStruct_AdvancedEmbeddedStructs(t *testing.T) {
	schema := FromStruct[AdvancedUser]()
	objSchema := schema.(*ObjectSchema)

	// Test that embedded field is properly merged
	if _, exists := objSchema.properties["embedded_field"]; !exists {
		t.Error("Expected embedded_field to be merged into parent schema")
	}

	embeddedSchema, ok := objSchema.properties["embedded_field"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for embedded_field, got %T", objSchema.properties["embedded_field"])
	}

	if embeddedSchema.metadata.Description != "Embedded field" {
		t.Errorf("Expected description 'Embedded field', got '%s'", embeddedSchema.metadata.Description)
	}
}

func TestFromStruct_AdvancedInterfaceTypes(t *testing.T) {
	schema := FromStruct[AdvancedUser]()
	objSchema := schema.(*ObjectSchema)

	// Test interface{} field
	dataSchema, ok := objSchema.properties["data"].(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema for interface field, got %T", objSchema.properties["data"])
	}

	// Check if additional properties are allowed by checking the schema properties
	if dataSchema.Type() != TypeObject {
		t.Error("Expected interface field to be object type")
	}
}

func TestFromStruct_AdvancedMapTypes(t *testing.T) {
	schema := FromStruct[AdvancedUser]()
	objSchema := schema.(*ObjectSchema)

	// Test map[string]any field
	prefsSchema, ok := objSchema.properties["preferences"].(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema for map field, got %T", objSchema.properties["preferences"])
	}

	if prefsSchema.Type() != TypeObject {
		t.Error("Expected map field to be object type")
	}
}

func TestTypeRegistry(t *testing.T) {
	// Register a custom type mapping
	customType := reflect.TypeOf(CustomID(""))
	RegisterTypeMapping(customType, func() Schema {
		return NewString().
			MinLength(5).
			MaxLength(20).
			Pattern("^[A-Z0-9]+$").
			Description("Custom ID format").
			Build()
	})

	// Test struct with custom type
	type TestStruct struct {
		ID CustomID `json:"id"`
	}

	schema := FromStruct[TestStruct]()
	objSchema := schema.(*ObjectSchema)

	idSchema, ok := objSchema.properties["id"].(*StringSchema)
	if !ok {
		t.Fatalf("Expected StringSchema for custom ID, got %T", objSchema.properties["id"])
	}

	if idSchema.minLength == nil || *idSchema.minLength != 5 {
		t.Errorf("Expected minLength=5 for custom ID, got %v", idSchema.minLength)
	}

	if idSchema.maxLength == nil || *idSchema.maxLength != 20 {
		t.Errorf("Expected maxLength=20 for custom ID, got %v", idSchema.maxLength)
	}

	if idSchema.pattern != "^[A-Z0-9]+$" {
		t.Errorf("Expected pattern '^[A-Z0-9]+$' for custom ID, got '%s'", idSchema.pattern)
	}

	if idSchema.metadata.Description != "Custom ID format" {
		t.Errorf("Expected description 'Custom ID format', got '%s'", idSchema.metadata.Description)
	}
}

func TestValidateStructTag_Advanced(t *testing.T) {
	// Test valid struct
	err := ValidateStructTag(reflect.TypeOf(AdvancedUser{}))
	if err != nil {
		t.Errorf("Unexpected validation error for valid struct: %v", err)
	}

	// Test struct with invalid tags
	type InvalidStruct struct {
		BadMinLen   string  `schema:"minlen=invalid"`
		BadMaxLen   string  `schema:"maxlen=notanumber"`
		BadPattern  string  `schema:"pattern=[invalid"`
		BadIntMin   int     `schema:"min=notint"`
		BadFloatMax float64 `schema:"max=notfloat"`
		Conflicting string  `schema:"required,optional"`
	}

	err = ValidateStructTag(reflect.TypeOf(InvalidStruct{}))
	if err == nil {
		t.Error("Expected validation error for struct with invalid tags")
	}
}

func TestTagParsing_EdgeCases(t *testing.T) {
	// Test parseSchemaTag with various formats
	testCases := []struct {
		input    string
		expected map[string]string
	}{
		{
			"minlen=5,maxlen=10",
			map[string]string{"minlen": "5", "maxlen": "10"},
		},
		{
			"required,email,desc=Email address",
			map[string]string{"required": "true", "email": "true", "desc": "Email address"},
		},
		{
			"enum=one|two|three,desc=Choose one",
			map[string]string{"enum": "one|two|three", "desc": "Choose one"},
		},
		{
			"  minlen = 1 , maxlen = 100  ",
			map[string]string{"minlen": "1", "maxlen": "100"},
		},
		{
			"",
			map[string]string{},
		},
	}

	for i, tc := range testCases {
		result := parseSchemaTag(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("Test case %d: expected %d tags, got %d", i, len(tc.expected), len(result))
			continue
		}

		for key, expectedValue := range tc.expected {
			if actualValue, exists := result[key]; !exists {
				t.Errorf("Test case %d: expected key '%s' not found", i, key)
			} else if actualValue != expectedValue {
				t.Errorf("Test case %d: for key '%s', expected '%s', got '%s'", i, key, expectedValue, actualValue)
			}
		}
	}
}

func TestFromStruct_PointerTypes(t *testing.T) {
	type PointerStruct struct {
		RequiredString string  `json:"required_string"`
		OptionalString *string `json:"optional_string,omitempty"`
		OptionalInt    *int    `json:"optional_int,omitempty"`
		OptionalBool   *bool   `json:"optional_bool,omitempty"`
	}

	schema := FromStruct[PointerStruct]()
	objSchema := schema.(*ObjectSchema)

	// Check required fields
	expectedRequired := []string{"required_string"}
	if len(objSchema.required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(objSchema.required))
	}

	for _, required := range expectedRequired {
		found := false
		for _, r := range objSchema.required {
			if r == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field '%s' not found", required)
		}
	}

	// Check that pointer types are not required
	optionalFields := []string{"optional_string", "optional_int", "optional_bool"}
	for _, optional := range optionalFields {
		for _, required := range objSchema.required {
			if required == optional {
				t.Errorf("Optional field '%s' should not be required", optional)
			}
		}
	}
}

func TestFromStruct_JSONTags(t *testing.T) {
	type JSONTagStruct struct {
		Field1    string `json:"custom_name"`
		Field2    string `json:"field2,omitempty"`
		Field3    string `json:"-"`          // Should be ignored
		Field4    string `json:",omitempty"` // Uses field name but optional
		NoJSONTag string // Uses field name
	}

	schema := FromStruct[JSONTagStruct]()
	objSchema := schema.(*ObjectSchema)

	// Check expected fields
	expectedFields := map[string]bool{
		"custom_name": true,
		"field2":      true,
		"Field4":      true,
		"NoJSONTag":   true,
	}

	if len(objSchema.properties) != len(expectedFields) {
		t.Errorf("Expected %d properties, got %d", len(expectedFields), len(objSchema.properties))
	}

	for fieldName := range expectedFields {
		if _, exists := objSchema.properties[fieldName]; !exists {
			t.Errorf("Expected field '%s' not found", fieldName)
		}
	}

	// Check that Field3 is ignored
	if _, exists := objSchema.properties["Field3"]; exists {
		t.Error("Field3 should be ignored due to json:\"-\" tag")
	}

	// Check required fields (omitempty should make fields optional)
	expectedRequired := []string{"custom_name", "NoJSONTag"}
	if len(objSchema.required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(objSchema.required))
	}
}

func TestFromStruct_TimeTypes(t *testing.T) {
	type TimeStruct struct {
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
	}

	schema := FromStruct[TimeStruct]()
	objSchema := schema.(*ObjectSchema)

	// time.Time should be treated as a struct (object schema)
	createdSchema, ok := objSchema.properties["created_at"].(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema for time.Time field, got %T", objSchema.properties["created_at"])
	}

	if createdSchema.Type() != TypeObject {
		t.Errorf("Expected time.Time to be treated as object, got %s", createdSchema.Type())
	}

	// *time.Time should also be treated as an object
	updatedSchema, ok := objSchema.properties["updated_at"].(*ObjectSchema)
	if !ok {
		t.Fatalf("Expected ObjectSchema for *time.Time field, got %T", objSchema.properties["updated_at"])
	}

	if updatedSchema.Type() != TypeObject {
		t.Errorf("Expected *time.Time to be treated as object, got %s", updatedSchema.Type())
	}
}

func TestFromStruct_Caching_Advanced(t *testing.T) {
	// Test that caching works across multiple calls
	schema1 := FromStruct[AdvancedUser]()
	schema2 := FromStruct[AdvancedUser]()

	// Schemas should be equivalent but different instances (due to Clone())
	if schema1 == schema2 {
		t.Error("Expected different schema instances due to cloning")
	}

	// But they should have the same structure
	obj1 := schema1.(*ObjectSchema)
	obj2 := schema2.(*ObjectSchema)

	if len(obj1.properties) != len(obj2.properties) {
		t.Error("Cached schemas should have same number of properties")
	}

	if len(obj1.required) != len(obj2.required) {
		t.Error("Cached schemas should have same number of required fields")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test getJSONFieldName
	testField := reflect.StructField{
		Name: "TestField",
		Tag:  `json:"custom_name,omitempty"`,
	}

	jsonName := getJSONFieldName(testField)
	if jsonName != "custom_name" {
		t.Errorf("Expected JSON field name 'custom_name', got '%s'", jsonName)
	}

	// Test getTypeName
	typeName := getTypeName(reflect.TypeOf(""))
	if typeName != "string" {
		t.Errorf("Expected type name 'string', got '%s'", typeName)
	}

	typeName = getTypeName(reflect.TypeOf([]string{}))
	if typeName != "[]string" {
		t.Errorf("Expected type name '[]string', got '%s'", typeName)
	}
}

func TestSchemaTagApplication(t *testing.T) {
	// Test different tag application scenarios
	type TaggedStruct struct {
		URLField     string   `schema:"url"`
		UUIDField    string   `schema:"uuid"`
		FormatField  string   `schema:"format=custom"`
		ExampleStr   string   `schema:"example=test value"`
		ExampleInt   int      `schema:"example=42"`
		ExampleFloat float64  `schema:"example=3.14"`
		ExampleBool  bool     `schema:"example=true"`
		UniqueArray  []string `schema:"unique"`
	}

	schema := FromStruct[TaggedStruct]()
	objSchema := schema.(*ObjectSchema)

	// Test URL format
	urlSchema := objSchema.properties["URLField"].(*StringSchema)
	if urlSchema.format != "url" {
		t.Errorf("Expected URL format 'url', got '%s'", urlSchema.format)
	}

	// Test UUID format
	uuidSchema := objSchema.properties["UUIDField"].(*StringSchema)
	if uuidSchema.format != "uuid" {
		t.Errorf("Expected UUID format 'uuid', got '%s'", uuidSchema.format)
	}

	// Test custom format
	formatSchema := objSchema.properties["FormatField"].(*StringSchema)
	if formatSchema.format != "custom" {
		t.Errorf("Expected custom format 'custom', got '%s'", formatSchema.format)
	}

	// Test array with unique items
	arraySchema := objSchema.properties["UniqueArray"].(*ArraySchema)
	if !arraySchema.uniqueItems {
		t.Error("Expected unique items to be true")
	}
}
