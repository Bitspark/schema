# JSON Schema Generator

The JSON Schema generator converts schema definitions into JSON Schema format using the visitor pattern. It supports multiple JSON Schema draft versions and provides extensive configuration options.

## Features

- **Multiple Draft Support**: Draft-07, Draft-2019-09, Draft-2020-12
- **Configurable Output**: Pretty printing, minification, custom indentation
- **Comprehensive Validation**: All JSON Schema validation keywords
- **Metadata Support**: Titles, descriptions, examples, defaults
- **Visitor Pattern**: Clean separation of concerns using the visitor pattern
- **Type Safety**: Full integration with the schema type system

## Basic Usage

```go
package main

import (
    "fmt"
    "defs.dev/schema/export/json"
    "defs.dev/schema/schemas"
    "defs.dev/schema/api/core"
)

func main() {
    // Create a string schema
    schema := schemas.NewStringSchema(schemas.StringSchemaConfig{
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
    output, err := generator.Generate(schema)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(output))
}

func intPtr(i int) *int { return &i }
```

Output:
```json
{
  "$schema": "https://json-schema.org/draft-07/schema#",
  "description": "A valid email address",
  "format": "email",
  "maxLength": 100,
  "minLength": 5,
  "title": "Email",
  "type": "string"
}
```

## Configuration Options

### Draft Version

```go
// Use different JSON Schema draft versions
generator := json.NewGenerator(
    json.WithDraft("draft-2019-09"),
)
```

### Output Formatting

```go
// Minified output
generator := json.NewGenerator(
    json.WithMinifyOutput(true),
    json.WithPrettyPrint(false),
)

// Custom indentation
generator := json.NewGenerator(
    json.WithIndentSize(4),
)
```

### Metadata Control

```go
// Control what metadata is included
generator := json.NewGenerator(
    json.WithIncludeTitle(true),
    json.WithIncludeDescription(true),
    json.WithIncludeExamples(false),
    json.WithIncludeDefaults(true),
)
```

### Validation Strictness

```go
// Strict mode for additional validation
generator := json.NewGenerator(
    json.WithStrictMode(true),
    json.WithIncludeAdditionalProperties(true),
)
```

### Schema Metadata

```go
// Custom schema metadata
generator := json.NewGenerator(
    json.WithSchemaURI("https://json-schema.org/draft/2020-12/schema#"),
    json.WithRootID("https://example.com/schemas/user.json"),
)
```

## Complex Schema Examples

### Object Schema

```go
// Create an object schema
objectSchema := schemas.NewObjectSchema(schemas.ObjectSchemaConfig{
    Metadata: core.SchemaMetadata{
        Name:        "User",
        Description: "A user object",
    },
    Properties: map[string]core.Schema{
        "name": schemas.NewStringSchema(schemas.StringSchemaConfig{
            Metadata: core.SchemaMetadata{Name: "Name"},
            MinLength: intPtr(1),
            MaxLength: intPtr(100),
        }),
        "email": schemas.NewStringSchema(schemas.StringSchemaConfig{
            Metadata: core.SchemaMetadata{Name: "Email"},
            Format: "email",
        }),
        "age": schemas.NewIntegerSchema(schemas.IntegerSchemaConfig{
            Metadata: core.SchemaMetadata{Name: "Age"},
            Minimum: int64Ptr(0),
            Maximum: int64Ptr(150),
        }),
    },
    Required: []string{"name", "email"},
    AdditionalProperties: boolPtr(false),
})

generator := json.NewGenerator()
output, _ := generator.Generate(objectSchema)
```

Output:
```json
{
  "$schema": "https://json-schema.org/draft-07/schema#",
  "additionalProperties": false,
  "description": "A user object",
  "properties": {
    "age": {
      "maximum": 150,
      "minimum": 0,
      "title": "Age",
      "type": "integer"
    },
    "email": {
      "format": "email",
      "title": "Email",
      "type": "string"
    },
    "name": {
      "maxLength": 100,
      "minLength": 1,
      "title": "Name",
      "type": "string"
    }
  },
  "required": ["name", "email"],
  "title": "User",
  "type": "object"
}
```

### Array Schema

```go
// Create an array schema
arraySchema := schemas.NewArraySchema(schemas.ArraySchemaConfig{
    Metadata: core.SchemaMetadata{
        Name:        "UserList",
        Description: "A list of users",
    },
    Items: objectSchema, // Use the object schema from above
    MinItems: intPtr(1),
    MaxItems: intPtr(100),
    UniqueItems: boolPtr(true),
})

generator := json.NewGenerator()
output, _ := generator.Generate(arraySchema)
```

## Integration with Export System

### Registry Registration

```go
import "defs.dev/schema/export"

// Create registry
registry := export.NewGeneratorRegistry()

// Register JSON generator
json.RegisterJSONGenerator(registry)

// Use through registry
output, err := registry.Generate("json", schema)
```

### Factory Pattern

```go
// Create generator through factory
generator, err := json.NewJSONGenerator(
    json.WithDraft("draft-2020-12"),
    json.WithPrettyPrint(true),
)
```

## Supported Schema Types

| Schema Type | JSON Schema Type | Constraints |
|-------------|------------------|-------------|
| StringSchema | `string` | minLength, maxLength, pattern, format, enum |
| IntegerSchema | `integer` | minimum, maximum, multipleOf, enum |
| NumberSchema | `number` | minimum, maximum, multipleOf |
| BooleanSchema | `boolean` | - |
| ArraySchema | `array` | items, minItems, maxItems, uniqueItems |
| ObjectSchema | `object` | properties, required, additionalProperties |

## Draft Version Differences

### Draft-07
- Uses `definitions` for schema definitions
- Supports `exclusiveMinimum`/`exclusiveMaximum` as numbers

### Draft-2019-09 / Draft-2020-12
- Uses `$defs` for schema definitions  
- Uses boolean `exclusiveMinimum`/`exclusiveMaximum` with separate `minimum`/`maximum`

## Error Handling

The generator provides detailed error information:

```go
output, err := generator.Generate(schema)
if err != nil {
    if genErr, ok := err.(*base.GenerationError); ok {
        fmt.Printf("Generator: %s\n", genErr.Generator)
        fmt.Printf("Type: %s\n", genErr.Type)
        fmt.Printf("Message: %s\n", genErr.Message)
    }
}
```

## Performance Considerations

- **Memory Efficient**: Streaming generation without large intermediate representations
- **Concurrent Safe**: Generators can be used concurrently (each has independent state)
- **Reusable**: Generators can be reused for multiple schemas
- **Configurable**: Options allow trading features for performance

## Best Practices

1. **Reuse Generators**: Create once, use multiple times
2. **Configure Appropriately**: Only enable features you need
3. **Handle Errors**: Always check for generation errors
4. **Use Registry**: For multiple output formats, use the registry system
5. **Validate Output**: Consider validating generated JSON Schema

## Helper Functions

```go
func intPtr(i int) *int { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool { return &b }
``` 