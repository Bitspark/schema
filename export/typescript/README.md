# TypeScript Generator

The TypeScript generator converts schema definitions into TypeScript interfaces, types, and enums using the visitor pattern. It provides extensive configuration options for customizing the generated TypeScript code to match your project's conventions and requirements.

## Features

- **Multiple Output Styles**: Generate interfaces, type aliases, or classes
- **Flexible Naming Conventions**: Support for PascalCase, camelCase, snake_case, and kebab-case
- **Rich Documentation**: JSDoc comments with examples and default values
- **Enum Support**: Generate TypeScript enums or union types
- **Strict Mode**: Optional readonly arrays and stricter type definitions
- **Customizable Formatting**: Configurable indentation, array styles, and module systems
- **Preset Configurations**: Ready-made configurations for React, Node.js, and other environments

## Quick Start

```go
package main

import (
    "fmt"
    "defs.dev/schema/export/typescript"
)

func main() {
    // Create a generator with default options
    generator := typescript.New()
    
    // Generate TypeScript code from a schema
    output, err := generator.Generate(mySchema)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(output))
}
```

## Configuration Options

### Output Styles

```go
// Interface style (default)
generator := typescript.NewGenerator(
    typescript.WithOutputStyle("interface"),
)
// Generates: export interface User { name: string; }

// Type alias style
generator := typescript.NewGenerator(
    typescript.WithOutputStyle("type"),
)
// Generates: export type User = { name: string; }
```

### Naming Conventions

```go
// PascalCase (default for types)
generator := typescript.NewGenerator(
    typescript.WithNamingConvention("PascalCase"),
)
// user_profile -> UserProfile

// camelCase
generator := typescript.NewGenerator(
    typescript.WithNamingConvention("camelCase"),
)
// user_profile -> userProfile

// snake_case
generator := typescript.NewGenerator(
    typescript.WithNamingConvention("snake_case"),
)
// UserProfile -> user_profile

// kebab-case
generator := typescript.NewGenerator(
    typescript.WithNamingConvention("kebab-case"),
)
// UserProfile -> user-profile
```

### Array Styles

```go
// T[] style (default)
generator := typescript.NewGenerator(
    typescript.WithArrayStyle("T[]"),
)
// Generates: string[]

// Array<T> style
generator := typescript.NewGenerator(
    typescript.WithArrayStyle("Array<T>"),
)
// Generates: Array<string>
```

### Enum Generation

```go
// TypeScript enums (default)
generator := typescript.NewGenerator(
    typescript.WithEnums(true),
)
// Generates:
// export enum Status {
//   ACTIVE = "active",
//   INACTIVE = "inactive"
// }

// Union types
generator := typescript.NewGenerator(
    typescript.WithEnums(false),
)
// Generates: export type Status = "active" | "inactive";
```

### Documentation Options

```go
// Enable JSDoc comments (default)
generator := typescript.NewGenerator(
    typescript.WithJSDoc(true),
    typescript.WithExamples(true),
    typescript.WithDefaults(true),
)
// Generates:
// /**
//  * User profile information
//  * @example { name: "John Doe", age: 30 }
//  * @default { name: "", age: 0 }
//  */
// export interface User {
//   name: string;
//   age: number;
// }
```

### Strict Mode

```go
// Enable strict TypeScript features
generator := typescript.NewGenerator(
    typescript.WithStrictMode(true),
    typescript.WithUnknownType(true),
)
// Generates readonly arrays and uses 'unknown' instead of 'any'
```

## Preset Configurations

### React Components

```go
generator := typescript.NewGenerator(
    typescript.WithReactPreset(),
)
// Optimized for React component props:
// - Interface style
// - PascalCase naming
// - Optional properties
// - Strict mode
// - JSDoc comments
```

### Node.js Applications

```go
generator := typescript.NewGenerator(
    typescript.WithNodePreset(),
)
// Optimized for Node.js:
// - CommonJS modules
// - Unknown type preference
// - Export types
// - JSDoc comments
```

### Browser Applications

```go
generator := typescript.NewGenerator(
    typescript.WithBrowserPreset(),
)
// Optimized for browser:
// - ES6 modules
// - Unknown type preference
// - Export types
// - JSDoc comments
```

### Minimal Output

```go
generator := typescript.NewGenerator(
    typescript.WithMinimalPreset(),
)
// Minimal output:
// - No comments
// - No examples
// - No exports
// - Compact formatting
```

### Strict TypeScript

```go
generator := typescript.NewGenerator(
    typescript.WithStrictPreset(),
)
// Strict TypeScript:
// - Strict mode enabled
// - Unknown type preference
// - Optional properties
// - Readonly arrays
// - Full documentation
```

## Advanced Configuration

### Custom Options

```go
options := typescript.TypeScriptOptions{
    OutputStyle:             "interface",
    NamingConvention:        "PascalCase",
    IncludeComments:         true,
    IncludeExamples:         true,
    IncludeDefaults:         true,
    StrictMode:              false,
    UseOptionalProperties:   true,
    IndentSize:              2,
    UseTabsForIndentation:   false,
    IncludeImports:          true,
    ExportTypes:             true,
    UseUnknownType:          true,
    GenerateValidators:      false,
    ValidatorLibrary:        "zod",
    UseEnums:                true,
    UseConstAssertions:      false,
    IncludeUtilityTypes:     false,
    ArrayStyle:              "T[]",
    ObjectStyle:             "interface",
    UsePartialTypes:         false,
    IncludeJSDoc:            true,
    JSDocStyle:              "standard",
    FileExtension:           ".ts",
    ModuleSystem:            "es6",
}

generator := typescript.NewWithOptions(options)
```

### Functional Options

```go
generator := typescript.NewGenerator(
    typescript.WithOutputStyle("type"),
    typescript.WithNamingConvention("camelCase"),
    typescript.WithStrictMode(true),
    typescript.WithIndentSize(4),
    typescript.WithTabs(false),
    typescript.WithUnknownType(true),
    typescript.WithEnums(false),
    typescript.WithArrayStyle("Array<T>"),
    typescript.WithJSDoc(true),
    typescript.WithExports(true),
)
```

## Integration Examples

### With Registry

```go
import (
    "defs.dev/schema/export"
    "defs.dev/schema/export/typescript"
)

// Create registry
registry := export.NewRegistry()

// Register TypeScript generator
registry.Register("typescript", typescript.New())

// Generate TypeScript
output, err := registry.Generate("typescript", mySchema)
if err != nil {
    panic(err)
}
```

### Batch Generation

```go
// Generate multiple formats
generators := map[string]export.Generator{
    "typescript-interface": typescript.NewGenerator(typescript.WithInterfacePreset()),
    "typescript-type":      typescript.NewGenerator(typescript.WithTypePreset()),
    "typescript-react":     typescript.NewGenerator(typescript.WithReactPreset()),
}

for name, gen := range generators {
    output, err := gen.Generate(mySchema)
    if err != nil {
        fmt.Printf("Error generating %s: %v\n", name, err)
        continue
    }
    
    // Save to file
    filename := fmt.Sprintf("types.%s.ts", name)
    err = os.WriteFile(filename, output, 0644)
    if err != nil {
        fmt.Printf("Error writing %s: %v\n", filename, err)
    }
}
```

### Dynamic Configuration

```go
func createGenerator(config map[string]interface{}) export.Generator {
    var options []typescript.Option
    
    if style, ok := config["outputStyle"].(string); ok {
        options = append(options, typescript.WithOutputStyle(style))
    }
    
    if convention, ok := config["namingConvention"].(string); ok {
        options = append(options, typescript.WithNamingConvention(convention))
    }
    
    if strict, ok := config["strictMode"].(bool); ok {
        options = append(options, typescript.WithStrictMode(strict))
    }
    
    return typescript.NewGenerator(options...)
}

// Usage
config := map[string]interface{}{
    "outputStyle":       "interface",
    "namingConvention":  "PascalCase",
    "strictMode":        true,
}

generator := createGenerator(config)
```

## Output Examples

### String Schema

**Input Schema**: String with enum values
```go
// Schema: name="Status", enum=["active", "inactive", "pending"]
```

**Generated TypeScript**:
```typescript
/**
 * Status enumeration
 * @example "active"
 */
export enum Status {
  ACTIVE = "active",
  INACTIVE = "inactive",
  PENDING = "pending"
}
```

### Object Schema

**Input Schema**: User object with properties
```go
// Schema: name="User", properties={id: integer, name: string, email?: string}
```

**Generated TypeScript**:
```typescript
/**
 * User object
 */
export interface User {
  /**
   * User ID
   */
  id: number;
  
  /**
   * User name
   */
  name: string;
  
  /**
   * User email
   */
  email?: string;
}
```

### Array Schema

**Input Schema**: Array of strings
```go
// Schema: name="Tags", items=string
```

**Generated TypeScript**:
```typescript
/**
 * List of tags
 */
export type Tags = string[];
```

## Error Handling

The generator provides detailed error messages for common issues:

```go
generator := typescript.NewGenerator()
output, err := generator.Generate(schema)
if err != nil {
    switch e := err.(type) {
    case *base.GenerationError:
        fmt.Printf("Generation error in %s for %s: %s\n", 
            e.Generator, e.SchemaType, e.Message)
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

## Best Practices

1. **Use Presets**: Start with preset configurations and customize as needed
2. **Consistent Naming**: Choose one naming convention and stick to it
3. **Documentation**: Enable JSDoc for better developer experience
4. **Strict Mode**: Use strict mode for better type safety
5. **Validation**: Always validate options before using them
6. **Error Handling**: Implement proper error handling for generation failures

## Supported Schema Types

- ✅ String (with enum support)
- ✅ Integer/Number
- ✅ Boolean
- ✅ Array
- ✅ Object
- ✅ Null
- ✅ Any/Unknown

## Future Enhancements

- Runtime validator generation (Zod, Yup, Joi, AJV)
- Class generation with methods
- Utility type generation
- OpenAPI integration
- Multi-file output support
- Import/export optimization 