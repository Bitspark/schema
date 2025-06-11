package examples

import (
	"encoding/json"
	"fmt"

	"defs.dev/schema/core"
	"defs.dev/schema/core/builders"
)

// ObjectExamples demonstrates various ObjectSchema usage patterns
func ObjectExamples() {
	fmt.Println("=== Object Schema Examples ===")
	fmt.Println()

	basicObjectExamples()
	fmt.Println()

	propertyExamples()
	fmt.Println()

	requiredPropertiesExamples()
	fmt.Println()

	additionalPropertiesExamples()
	fmt.Println()

	nestedObjectExamples()
	fmt.Println()

	constraintsExamples()
	fmt.Println()

	realWorldExamples()
	fmt.Println()

	fmt.Println("✨ All Object Schema examples completed successfully!")
}

// RunObjectExamples is the main entry point for object schema examples
func RunObjectExamples() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error in object examples: %v\n", r)
		}
	}()
	ObjectExamples()
}

func basicObjectExamples() {
	fmt.Println("--- Basic Object Validation ---")

	// Simple object schema (allows any properties)
	schema := core.NewObject().Build()

	// Test various object types
	testValues := []any{
		map[string]any{}, // Empty object
		map[string]any{"name": "John", "age": 30},
		map[string]any{"foo": "bar", "nested": map[string]any{"key": "value"}},
		"not an object", // Invalid
		42,              // Invalid
		nil,             // Invalid
	}

	for _, value := range testValues {
		result := schema.Validate(value)
		status := "✅ Valid"
		if !result.Valid {
			status = "❌ Invalid"
		}
		fmt.Printf("Object %v: %s\n", value, status)
		if !result.Valid && len(result.Errors) > 0 {
			fmt.Printf("  Error: %s\n", result.Errors[0].Message)
		}
	}
}

func propertyExamples() {
	fmt.Println("--- Object Property Schemas ---")

	// Object with typed properties
	personSchema := core.NewObject().
		Description("A person object").
		Property("name", core.NewString().MinLength(2).Build()).
		Property("age", core.NewInteger().Min(0).Max(150).Build()).
		Property("email", core.NewString().Email().Build()).
		Build()

	testCases := []struct {
		name  string
		value any
	}{
		{
			"Valid person",
			map[string]any{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
			},
		},
		{
			"Invalid name (too short)",
			map[string]any{
				"name":  "J",
				"age":   30,
				"email": "john@example.com",
			},
		},
		{
			"Invalid age (negative)",
			map[string]any{
				"name":  "John Doe",
				"age":   -5,
				"email": "john@example.com",
			},
		},
		{
			"Invalid email format",
			map[string]any{
				"name":  "John Doe",
				"age":   30,
				"email": "not-an-email",
			},
		},
		{
			"Missing properties (valid - optional by default)",
			map[string]any{
				"name": "John",
			},
		},
	}

	for _, tc := range testCases {
		result := personSchema.Validate(tc.value)
		status := "✅ Valid"
		if !result.Valid {
			status = "❌ Invalid"
		}
		fmt.Printf("%s: %s\n", tc.name, status)
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}
}

func requiredPropertiesExamples() {
	fmt.Println("--- Required Properties ---")

	// Schema with required properties
	userSchema := core.NewObject().
		Property("username", core.NewString().MinLength(3).Build()).
		Property("email", core.NewString().Email().Build()).
		Property("age", core.NewInteger().Min(13).Build()).
		Required("username", "email"). // Age is optional
		Build()

	testCases := []struct {
		name  string
		value any
	}{
		{
			"Valid with all properties",
			map[string]any{
				"username": "john_doe",
				"email":    "john@example.com",
				"age":      25,
			},
		},
		{
			"Valid without optional age",
			map[string]any{
				"username": "jane_doe",
				"email":    "jane@example.com",
			},
		},
		{
			"Missing required username",
			map[string]any{
				"email": "missing@example.com",
				"age":   30,
			},
		},
		{
			"Missing required email",
			map[string]any{
				"username": "missing_email",
				"age":      25,
			},
		},
	}

	for _, tc := range testCases {
		result := userSchema.Validate(tc.value)
		status := "✅ Valid"
		if !result.Valid {
			status = "❌ Invalid"
		}
		fmt.Printf("%s: %s\n", tc.name, status)
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}
}

func additionalPropertiesExamples() {
	fmt.Println("--- Additional Properties Control ---")

	// Strict schema (no additional properties)
	strictSchema := core.NewObject().
		Property("name", core.NewString().Build()).
		Property("age", core.NewInteger().Build()).
		AdditionalProperties(false).
		Build()

	// Flexible schema (allows additional properties)
	flexibleSchema := core.NewObject().
		Property("name", core.NewString().Build()).
		Property("age", core.NewInteger().Build()).
		AdditionalProperties(true).
		Build()

	testData := map[string]any{
		"name":  "John",
		"age":   30,
		"extra": "not allowed in strict schema",
	}

	// Test strict schema
	result := strictSchema.Validate(testData)
	status := "✅ Valid"
	if !result.Valid {
		status = "❌ Invalid"
	}
	fmt.Printf("Strict schema validation: %s\n", status)
	if !result.Valid {
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s\n", err.Message)
		}
	}

	// Test flexible schema
	result = flexibleSchema.Validate(testData)
	status = "✅ Valid"
	if !result.Valid {
		status = "❌ Invalid"
	}
	fmt.Printf("Flexible schema validation: %s\n", status)
}

func nestedObjectExamples() {
	fmt.Println("--- Nested Objects ---")

	// Create nested address schema
	addressSchema := core.NewObject().
		Property("street", core.NewString().MinLength(5).Build()).
		Property("city", core.NewString().MinLength(2).Build()).
		Property("country", core.NewString().Enum("US", "CA", "UK", "DE").Build()).
		Property("postal_code", core.NewString().Pattern(`^\d{5}(-\d{4})?$`).Build()).
		Required("street", "city", "country").
		AdditionalProperties(false).
		Build()

	// Create person schema with nested address
	personSchema := core.NewObject().
		Property("name", core.NewString().MinLength(2).Build()).
		Property("age", core.NewInteger().Min(0).Build()).
		Property("address", addressSchema).
		Property("work_address", addressSchema). // Optional work address
		Required("name", "address").
		Build()

	testCases := []struct {
		name  string
		value any
	}{
		{
			"Valid person with address",
			map[string]any{
				"name": "Alice Smith",
				"age":  28,
				"address": map[string]any{
					"street":      "123 Main Street",
					"city":        "Springfield",
					"country":     "US",
					"postal_code": "12345",
				},
			},
		},
		{
			"Valid with work address too",
			map[string]any{
				"name": "Bob Johnson",
				"age":  35,
				"address": map[string]any{
					"street":      "456 Oak Avenue",
					"city":        "Portland",
					"country":     "US",
					"postal_code": "97201",
				},
				"work_address": map[string]any{
					"street":  "789 Business Blvd",
					"city":    "Portland",
					"country": "US",
				},
			},
		},
		{
			"Invalid nested address (missing required field)",
			map[string]any{
				"name": "Charlie Brown",
				"address": map[string]any{
					"street": "321 Pine Street",
					"city":   "Seattle",
					// Missing required "country"
				},
			},
		},
		{
			"Invalid nested address (invalid country)",
			map[string]any{
				"name": "Diana Prince",
				"address": map[string]any{
					"street":  "999 Hero Lane",
					"city":    "Metropolis",
					"country": "XX", // Invalid country code
				},
			},
		},
	}

	for _, tc := range testCases {
		result := personSchema.Validate(tc.value)
		status := "✅ Valid"
		if !result.Valid {
			status = "❌ Invalid"
		}
		fmt.Printf("%s: %s\n", tc.name, status)
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}
}

func constraintsExamples() {
	fmt.Println("--- Object Constraints ---")

	// Schema with property count constraints
	configSchema := builders.NewObjectBuilder().
		MinProperties(1). // Must have at least 1 property
		MaxProperties(5). // Cannot have more than 5 properties
		Description("Configuration object with size limits").
		Build()

	testCases := []struct {
		name  string
		value any
	}{
		{
			"Valid - within limits",
			map[string]any{
				"debug":   true,
				"timeout": 30,
				"retries": 3,
			},
		},
		{
			"Invalid - too few properties",
			map[string]any{}, // Empty object
		},
		{
			"Invalid - too many properties",
			map[string]any{
				"prop1": 1,
				"prop2": 2,
				"prop3": 3,
				"prop4": 4,
				"prop5": 5,
				"prop6": 6, // 6th property exceeds limit
			},
		},
		{
			"Valid - exactly at minimum",
			map[string]any{
				"enabled": true,
			},
		},
		{
			"Valid - exactly at maximum",
			map[string]any{
				"prop1": 1,
				"prop2": 2,
				"prop3": 3,
				"prop4": 4,
				"prop5": 5,
			},
		},
	}

	for _, tc := range testCases {
		result := configSchema.Validate(tc.value)
		status := "✅ Valid"
		if !result.Valid {
			status = "❌ Invalid"
		}
		fmt.Printf("%s: %s\n", tc.name, status)
		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  Error: %s\n", err.Message)
			}
		}
	}
}

func realWorldExamples() {
	fmt.Println("--- Real-World Use Cases ---")

	// 1. API Request Schema
	fmt.Println("API Request Schema:")
	apiRequestSchema := core.NewObject().
		Description("API request for creating a user").
		Property("user", core.NewObject().
			Property("name", core.NewString().MinLength(2).MaxLength(50).Build()).
			Property("email", core.NewString().Email().Build()).
			Property("age", core.NewInteger().Min(13).Max(120).Build()).
			Required("name", "email").
			Build()).
		Property("metadata", core.NewObject().
			AdditionalProperties(true). // Allow flexible metadata
			Build()).
		Required("user").
		AdditionalProperties(false).
		Build()

	// Generate JSON Schema
	jsonSchema := apiRequestSchema.ToJSONSchema()
	jsonBytes, _ := json.MarshalIndent(jsonSchema, "", "  ")
	fmt.Printf("JSON Schema:\n%s\n\n", string(jsonBytes))

	// 2. Database Record Schema
	fmt.Println("Database Record Schema:")
	dbRecordBuilder := builders.NewObjectBuilder()
	dbRecordBuilder.RequiredProperty("id", core.NewInteger().Min(1).Build())
	dbRecordBuilder.RequiredProperty("created_at", core.NewString().Format("date-time").Build())
	dbRecordBuilder.OptionalProperty("updated_at", core.NewString().Format("date-time").Build())
	dbRecordBuilder.OptionalProperty("deleted_at", core.NewString().Format("date-time").Build())
	dbRecordBuilder.OptionalProperty("created_by", core.NewString().Build())
	dbRecordBuilder.OptionalProperty("modified_by", core.NewString().Build())
	dbRecordBuilder.OptionalProperty("modified_at", core.NewString().Format("date-time").Build())
	dbRecordSchema := dbRecordBuilder.Build()

	// Test with valid database record
	dbRecord := map[string]any{
		"id":          123,
		"created_at":  "2023-01-15T10:30:00Z",
		"updated_at":  "2023-01-16T14:20:00Z",
		"created_by":  "admin",
		"modified_by": "admin",
		"modified_at": "2023-01-16T14:20:00Z",
	}

	result := dbRecordSchema.Validate(dbRecord)
	status := "✅ Valid"
	if !result.Valid {
		status = "❌ Invalid"
	}
	fmt.Printf("Database record validation: %s\n", status)

	// 3. Configuration Schema
	fmt.Println("\nConfiguration Schema:")
	configSchema := builders.NewObjectBuilder().
		ConfigExample().
		RequiredProperty("app_name", core.NewString().MinLength(1).Build()).
		OptionalProperty("debug", core.NewBoolean().Default(false).Build()).
		OptionalProperty("port", core.NewInteger().Port().Build()).
		OptionalProperty("database", core.NewObject().
			Property("host", core.NewString().Build()).
			Property("port", core.NewInteger().Port().Build()).
			Property("name", core.NewString().Build()).
			Required("host", "name").
			Build()).
		Build()

	configData := map[string]any{
		"app_name": "MyApp",
		"debug":    true,
		"port":     8080,
		"database": map[string]any{
			"host": "localhost",
			"port": 5432,
			"name": "myapp_db",
		},
		"custom_setting": "allowed", // Additional properties allowed
	}

	result = configSchema.Validate(configData)
	status = "✅ Valid"
	if !result.Valid {
		status = "❌ Invalid"
	}
	fmt.Printf("Configuration validation: %s\n", status)

	// 4. Example Generation
	fmt.Println("\nGenerated Examples:")

	userExample := apiRequestSchema.GenerateExample()
	fmt.Printf("API Request Example: %v\n", userExample)

	configExample := configSchema.GenerateExample()
	fmt.Printf("Configuration Example: %v\n", configExample)

	// 5. Schema Introspection
	fmt.Println("\nSchema Introspection:")
	properties := apiRequestSchema.Properties()
	fmt.Printf("API Request Schema has %d properties:\n", len(properties))
	for propName := range properties {
		fmt.Printf("  - %s\n", propName)
	}

	required := apiRequestSchema.Required()
	fmt.Printf("Required properties: %v\n", required)
	fmt.Printf("Additional properties allowed: %v\n", apiRequestSchema.AdditionalProperties())
}
