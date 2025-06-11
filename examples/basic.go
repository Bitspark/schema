package examples

import (
	"defs.dev/schema"
	"fmt"

	"defs.dev/schema/api"
)

// BasicStringSchemaExample demonstrates creating and using a StringSchema
// with the new core package and API interfaces.
func BasicStringSchemaExample() {
	// Create a string schema using the core package
	usernameSchema := schema.NewString().
		MinLength(3).
		MaxLength(20).
		Pattern(`^[a-zA-Z0-9_]+$`).
		Description("Username for the system").
		Example("john_doe").
		Build()

	// The schema implements api.StringSchema interface
	var schema api.StringSchema = usernameSchema

	// Test validation with valid input
	result := schema.Validate("john_doe")
	fmt.Printf("Validating 'john_doe': Valid=%t\n", result.Valid)

	// Test validation with invalid input (too short)
	result = schema.Validate("jo")
	fmt.Printf("Validating 'jo': Valid=%t\n", result.Valid)
	if !result.Valid {
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s (Code: %s)\n", err.Message, err.Code)
			fmt.Printf("  Suggestion: %s\n", err.Suggestion)
		}
	}

	// Test validation with invalid input (pattern mismatch)
	result = schema.Validate("john-doe")
	fmt.Printf("Validating 'john-doe': Valid=%t\n", result.Valid)
	if !result.Valid {
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s (Code: %s)\n", err.Message, err.Code)
		}
	}

	// Generate JSON Schema
	jsonSchema := schema.ToJSONSchema()
	fmt.Printf("JSON Schema: %+v\n", jsonSchema)

	// Generate example
	example := schema.GenerateExample()
	fmt.Printf("Generated example: %v\n", example)
}

// EmailSchemaExample demonstrates creating an email schema with format validation.
func EmailSchemaExample() {
	emailSchema := schema.NewString().
		Email().
		Description("User's email address").
		Build()

	// Test with valid email
	result := emailSchema.Validate("user@example.com")
	fmt.Printf("Email validation 'user@example.com': Valid=%t\n", result.Valid)

	// Test with invalid email
	result = emailSchema.Validate("not-an-email")
	fmt.Printf("Email validation 'not-an-email': Valid=%t\n", result.Valid)
	if !result.Valid {
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s\n", err.Message)
			fmt.Printf("  Suggestion: %s\n", err.Suggestion)
		}
	}
}

// SchemaCompositionExample demonstrates using schemas as api.Schema interface.
func SchemaCompositionExample() {
	// Create different types of schemas
	nameSchema := schema.NewString().MinLength(1).Description("Person's name").Build()
	emailSchema := schema.NewString().Email().Build()

	// Function that accepts any schema via the API interface
	validateAndPrint := func(name string, schema api.Schema, value any) {
		result := schema.Validate(value)
		fmt.Printf("Validating %s with value %v: Valid=%t\n", name, value, result.Valid)

		if !result.Valid {
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err.Message)
			}
		}

		fmt.Printf("  Schema type: %s\n", schema.Type())
		fmt.Printf("  Generated example: %v\n", schema.GenerateExample())
	}

	// Use the same function with different schema types
	validateAndPrint("name", nameSchema, "John")
	validateAndPrint("name", nameSchema, "")
	validateAndPrint("email", emailSchema, "john@example.com")
	validateAndPrint("email", emailSchema, "invalid-email")
}
