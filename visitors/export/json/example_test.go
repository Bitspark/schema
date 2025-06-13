package json_test

import (
	"fmt"
	"log"

	"defs.dev/schema/core"
	"defs.dev/schema/schemas"
	"defs.dev/schema/visitors/export/json"
)

// Example demonstrates basic JSON Schema generation
func Example_basic() {
	// Create a string schema for email validation
	emailSchema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Name:        "Email",
			Description: "A valid email address",
		},
		MinLength: intPtr(5),
		MaxLength: intPtr(100),
		Format:    "email",
	})

	// Create JSON Schema generator
	generator := json.NewGenerator()

	// Generate JSON Schema
	output, err := generator.Generate(emailSchema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
	// Output:
	// {
	//   "$schema": "https://json-schema.org/draft-07/schema#",
	//   "description": "A valid email address",
	//   "format": "email",
	//   "maxLength": 100,
	//   "minLength": 5,
	//   "title": "Email",
	//   "type": "string"
	// }
}

// Example demonstrates complex object schema generation
func Example_objectSchema() {
	// Create a user object schema
	userSchema := schemas.NewObjectSchema(schemas.ObjectSchemaConfig{
		Metadata: core.SchemaMetadata{
			Name:        "User",
			Description: "A user in the system",
		},
		Properties: map[string]core.Schema{
			"id": schemas.NewIntegerSchema(schemas.IntegerSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "ID",
					Description: "Unique user identifier",
				},
				Minimum: int64Ptr(1),
			}),
			"name": schemas.NewStringSchema(schemas.StringSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "Name",
					Description: "User's full name",
				},
				MinLength: intPtr(1),
				MaxLength: intPtr(100),
			}),
			"email": schemas.NewStringSchema(schemas.StringSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "Email",
					Description: "User's email address",
				},
				Format: "email",
			}),
			"active": schemas.NewBooleanSchema(schemas.BooleanSchemaConfig{
				Metadata: core.SchemaMetadata{
					Name:        "Active",
					Description: "Whether the user is active",
				},
			}),
		},
		Required:             []string{"id", "name", "email"},
		AdditionalProperties: false,
	})

	// Generate with custom options
	generator := json.NewGenerator(
		json.WithDraft("draft-2019-09"),
		json.WithIncludeAdditionalProperties(true),
	)

	output, err := generator.Generate(userSchema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
	// Output:
	// {
	//   "$schema": "https://json-schema.org/draft/2019-09/schema#",
	//   "additionalProperties": false,
	//   "description": "A user in the system",
	//   "properties": {
	//     "active": {
	//       "description": "Whether the user is active",
	//       "title": "Active",
	//       "type": "boolean"
	//     },
	//     "email": {
	//       "description": "User's email address",
	//       "format": "email",
	//       "title": "Email",
	//       "type": "string"
	//     },
	//     "id": {
	//       "description": "Unique user identifier",
	//       "minimum": 1,
	//       "title": "ID",
	//       "type": "integer"
	//     },
	//     "name": {
	//       "description": "User's full name",
	//       "maxLength": 100,
	//       "minLength": 1,
	//       "title": "Name",
	//       "type": "string"
	//     }
	//   },
	//   "required": [
	//     "id",
	//     "name",
	//     "email"
	//   ],
	//   "title": "User",
	//   "type": "object"
	// }
}

// Example demonstrates array schema generation
func Example_arraySchema() {
	// Create an array of strings schema
	tagsSchema := schemas.NewArraySchema(schemas.ArraySchemaConfig{
		Metadata: core.SchemaMetadata{
			Name:        "Tags",
			Description: "A list of tags",
		},
		ItemSchema: schemas.NewStringSchema(schemas.StringSchemaConfig{
			Metadata: core.SchemaMetadata{
				Name: "Tag",
			},
			MinLength: intPtr(1),
			MaxLength: intPtr(50),
		}),
		MinItems:    intPtr(0),
		MaxItems:    intPtr(10),
		UniqueItems: true,
	})

	generator := json.NewGenerator()
	output, err := generator.Generate(tagsSchema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
	// Output:
	// {
	//   "$schema": "https://json-schema.org/draft-07/schema#",
	//   "description": "A list of tags",
	//   "items": {
	//     "maxLength": 50,
	//     "minLength": 1,
	//     "title": "Tag",
	//     "type": "string"
	//   },
	//   "maxItems": 10,
	//   "minItems": 0,
	//   "title": "Tags",
	//   "type": "array",
	//   "uniqueItems": true
	// }
}

// Example demonstrates minified output
func Example_minified() {
	schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Name: "CompactString",
		},
		MinLength: intPtr(1),
	})

	generator := json.NewGenerator(
		json.WithMinifyOutput(true),
		json.WithPrettyPrint(false),
	)

	output, err := generator.Generate(schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
	// Output: {"$schema":"https://json-schema.org/draft-07/schema#","minLength":1,"title":"CompactString","type":"string"}
}

// Example demonstrates different draft versions
func Example_draftVersions() {
	schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Name: "TestString",
		},
	})

	// Draft 07
	gen07 := json.NewGenerator(json.WithDraft("draft-07"))
	output07, _ := gen07.Generate(schema)
	fmt.Println("Draft 07:")
	fmt.Println(string(output07))

	// Draft 2020-12
	gen2020 := json.NewGenerator(json.WithDraft("draft-2020-12"))
	output2020, _ := gen2020.Generate(schema)
	fmt.Println("\nDraft 2020-12:")
	fmt.Println(string(output2020))

	// Output:
	// Draft 07:
	// {
	//   "$schema": "https://json-schema.org/draft-07/schema#",
	//   "title": "TestString",
	//   "type": "string"
	// }
	//
	// Draft 2020-12:
	// {
	//   "$schema": "https://json-schema.org/draft/2020-12/schema#",
	//   "title": "TestString",
	//   "type": "string"
	// }
}

// Example demonstrates metadata control
func Example_metadataControl() {
	schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
		Metadata: core.SchemaMetadata{
			Name:        "TestString",
			Description: "A test string with examples",
			Examples:    []any{"hello", "world"},
		},
		DefaultVal: stringPtr("default"),
	})

	// Include all metadata
	fullGen := json.NewGenerator(
		json.WithIncludeTitle(true),
		json.WithIncludeDescription(true),
		json.WithIncludeExamples(true),
		json.WithIncludeDefaults(true),
	)

	fullOutput, _ := fullGen.Generate(schema)
	fmt.Println("Full metadata:")
	fmt.Println(string(fullOutput))

	// Minimal metadata
	minimalGen := json.NewGenerator(
		json.WithIncludeTitle(false),
		json.WithIncludeDescription(false),
		json.WithIncludeExamples(false),
		json.WithIncludeDefaults(false),
	)

	minimalOutput, _ := minimalGen.Generate(schema)
	fmt.Println("\nMinimal metadata:")
	fmt.Println(string(minimalOutput))

	// Output:
	// Full metadata:
	// {
	//   "$schema": "https://json-schema.org/draft-07/schema#",
	//   "default": "default",
	//   "description": "A test string with examples",
	//   "examples": [
	//     "hello",
	//     "world"
	//   ],
	//   "title": "TestString",
	//   "type": "string"
	// }
	//
	// Minimal metadata:
	// {
	//   "$schema": "https://json-schema.org/draft-07/schema#",
	//   "type": "string"
	// }
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
