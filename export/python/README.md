# Python Generator

The Python generator creates Python code from schema definitions using the visitor pattern. It supports multiple output styles including Pydantic models, dataclasses, plain classes, and named tuples.

## Features

- **Multiple Output Styles**: Pydantic models, dataclasses, plain classes, named tuples
- **Pydantic Support**: Both v1 and v2 with proper imports and features
- **Type Hints**: Full support for Python type hints with configurable styles
- **Naming Conventions**: Flexible naming conventions for classes and fields
- **Documentation**: Rich docstring generation with multiple styles
- **Enums**: Support for Python enums, StrEnum, and Literal types
- **Modern Python**: Support for Python 3.8+ features including union types
- **Customization**: Extensive configuration options and preset configurations

## Quick Start

```go
package main

import (
    "fmt"
    "defs.dev/schema/export/python"
    "defs.dev/schema/api/core"
)

func main() {
    // Create a generator with default options
    generator := python.NewPythonGenerator()
    
    // Or use a preset
    generator = python.NewPythonGenerator(python.PydanticV2Preset()...)
    
    // Generate Python code from a schema
    result, err := generator.Generate(schema)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result)
}
```

## Output Styles

### Pydantic Models

Generate Pydantic models with full validation support:

```go
generator := python.NewPythonGenerator(
    python.WithOutputStyle("pydantic"),
    python.WithPydanticVersion("v2"),
)
```

**Output:**
```python
from pydantic import BaseModel
from typing import Optional

class User(BaseModel):
    """
    User model with profile information.
    
    Examples:
        {"name": "John Doe", "age": 30}
    """
    name: str
    age: int
    email: Optional[str] = None
```

### Dataclasses

Generate Python dataclasses:

```go
generator := python.NewPythonGenerator(
    python.WithOutputStyle("dataclass"),
    python.WithDataclassFeatures(true),
    python.WithDataclassOptions([]string{"frozen", "slots"}),
)
```

**Output:**
```python
from dataclasses import dataclass
from typing import Optional

@dataclass(frozen, slots)
class User:
    """
    User model with profile information.
    """
    name: str
    age: int
    email: Optional[str] = None
```

### Plain Classes

Generate simple Python classes:

```go
generator := python.NewPythonGenerator(
    python.WithOutputStyle("class"),
)
```

**Output:**
```python
class User:
    def __init__(self, name: str, age: int, email: Optional[str] = None):
        self.name = name
        self.age = age
        self.email = email
```

### Named Tuples

Generate named tuples for immutable data:

```go
generator := python.NewPythonGenerator(
    python.WithOutputStyle("namedtuple"),
)
```

**Output:**
```python
from collections import namedtuple

User = namedtuple('User', ['name', 'age', 'email'])
```

## Configuration Options

### Basic Options

```go
generator := python.NewPythonGenerator(
    // Output style
    python.WithOutputStyle("pydantic"),           // pydantic, dataclass, class, namedtuple
    python.WithPydanticVersion("v2"),             // v1, v2
    
    // Naming conventions
    python.WithNamingConvention("PascalCase"),    // PascalCase, snake_case
    python.WithFieldNamingConvention("snake_case"), // snake_case, camelCase
    
    // Documentation
    python.WithDocstrings(true),
    python.WithDocstringStyle("google"),          // google, numpy, sphinx
    python.WithComments(true),
    python.WithExamples(true),
    python.WithDefaults(true),
)
```

### Type Hints

```go
generator := python.NewPythonGenerator(
    python.WithTypeHints(true),
    python.WithTypeHintStyle("typing"),           // typing, builtin
    python.WithOptional(true),                    // Use Optional[T] vs T | None
    python.WithPythonVersion("3.10"),            // 3.8, 3.9, 3.10, 3.11, 3.12
)
```

### Enums

```go
generator := python.NewPythonGenerator(
    python.WithEnums(true),
    python.WithEnumStyle("Enum"),                 // Enum, StrEnum, Literal
)
```

**Enum Output:**
```python
from enum import Enum

class Status(Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    PENDING = "pending"
```

**StrEnum Output (Python 3.11+):**
```python
from enum import StrEnum

class Status(StrEnum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    PENDING = "pending"
```

**Literal Output:**
```python
from typing import Literal

Status = Literal["active", "inactive", "pending"]
```

### Formatting

```go
generator := python.NewPythonGenerator(
    python.WithIndentSize(4),
    python.WithTabs(false),                       // Use spaces instead of tabs
    python.WithImports(true),
    python.WithImportStyle("absolute"),           // absolute, relative
)
```

### Advanced Options

```go
generator := python.NewPythonGenerator(
    python.WithStrictMode(true),
    python.WithValidators(true),
    python.WithValidatorStyle("pydantic"),        // pydantic, custom, none
    python.WithSerializers(true),
    python.WithSerializerStyle("dict"),           // dict, json, both
    python.WithForwardRefs(true),
    python.WithBaseClass("CustomBase"),
    python.WithFileHeader("# Generated code - do not edit"),
    python.WithModuleName("models"),
    python.WithExtraImports([]string{"from datetime import datetime"}),
    python.WithCustomTypeMapping(map[string]string{
        "timestamp": "datetime",
    }),
)
```

## Preset Configurations

### Pydantic V2 (Default)

```go
generator := python.NewPythonGenerator(python.PydanticV2Preset()...)
```

Modern Pydantic v2 with type hints, enums, and Google-style docstrings.

### Pydantic V1

```go
generator := python.NewPythonGenerator(python.PydanticV1Preset()...)
```

Legacy Pydantic v1 support for older codebases.

### Dataclass

```go
generator := python.NewPythonGenerator(python.DataclassPreset()...)
```

Python dataclasses with frozen and slots options.

### Modern Python

```go
generator := python.NewPythonGenerator(python.ModernPythonPreset()...)
```

Uses Python 3.10+ features like builtin generics and union types.

### Minimal

```go
generator := python.NewPythonGenerator(python.MinimalPreset()...)
```

Minimal output with no type hints or documentation.

### Strict

```go
generator := python.NewPythonGenerator(python.StrictPreset()...)
```

Strict mode with validators and comprehensive documentation.

## Integration with Export System

### Registry Registration

```go
import "defs.dev/schema/export/base"

registry := base.NewGeneratorRegistry()
python.RegisterPythonGenerator(registry)

// Use with options
generator, err := registry.Create("python", map[string]any{
    "output_style": "pydantic",
    "pydantic_version": "v2",
    "naming_convention": "PascalCase",
})
```

### Batch Generation

```go
generators := map[string]base.Generator{
    "pydantic": python.NewPythonGenerator(python.PydanticV2Preset()...),
    "dataclass": python.NewPythonGenerator(python.DataclassPreset()...),
}

results, err := registry.GenerateBatch(schemas, generators)
```

## Examples

### Complete User Model

**Input Schema:**
```go
userSchema := &core.ObjectSchema{
    Name: "User",
    Description: "User profile with authentication details",
    Properties: map[string]core.Schema{
        "id": &core.IntegerSchema{
            Name: "id",
            Description: "Unique user identifier",
            Examples: []any{1, 42, 123},
        },
        "username": &core.StringSchema{
            Name: "username",
            Description: "Unique username for login",
            MinLength: ptr(3),
            MaxLength: ptr(50),
            Pattern: "^[a-zA-Z0-9_]+$",
            Examples: []any{"john_doe", "alice123"},
        },
        "email": &core.StringSchema{
            Name: "email",
            Description: "User email address",
            Format: "email",
            Examples: []any{"john@example.com"},
        },
        "status": &core.StringSchema{
            Name: "status",
            Description: "Account status",
            Enum: []string{"active", "inactive", "suspended"},
            Default: ptr("active"),
        },
        "profile": &core.ObjectSchema{
            Name: "profile",
            Description: "User profile information",
            Properties: map[string]core.Schema{
                "first_name": &core.StringSchema{Name: "first_name"},
                "last_name": &core.StringSchema{Name: "last_name"},
                "age": &core.IntegerSchema{
                    Name: "age",
                    Minimum: ptr(int64(0)),
                    Maximum: ptr(int64(150)),
                },
            },
            Required: []string{"first_name", "last_name"},
        },
    },
    Required: []string{"id", "username", "email"},
}
```

**Generated Pydantic Output:**
```python
from pydantic import BaseModel
from typing import Optional
from enum import Enum

class Status(Enum):
    """
    Account status
    
    Examples:
        "active"
    
    Default:
        "active"
    """
    ACTIVE = "active"
    INACTIVE = "inactive"
    SUSPENDED = "suspended"

class Profile(BaseModel):
    """
    User profile information
    """
    first_name: str
    last_name: str
    age: Optional[int] = None

class User(BaseModel):
    """
    User profile with authentication details
    
    Examples:
        {"id": 1, "username": "john_doe", "email": "john@example.com"}
    """
    id: int
    username: str
    email: str
    status: Optional[Status] = Status.ACTIVE
    profile: Optional[Profile] = None
```

### API Response Model

**Generated with Modern Python preset:**
```python
from pydantic import BaseModel
from enum import StrEnum

class ResponseStatus(StrEnum):
    SUCCESS = "success"
    ERROR = "error"

class ApiResponse(BaseModel):
    """
    Standard API response format
    """
    status: ResponseStatus
    data: dict[str, any] | None = None
    message: str | None = None
    errors: list[str] | None = None
```

## Best Practices

### 1. Choose the Right Output Style

- **Pydantic**: For APIs with validation requirements
- **Dataclass**: For data containers with immutability needs
- **Plain Class**: For simple data structures
- **Named Tuple**: For immutable, lightweight data

### 2. Use Appropriate Presets

Start with presets and customize as needed:

```go
// Start with a preset
options := python.PydanticV2Preset()

// Add customizations
options = append(options, 
    python.WithFileHeader("# Auto-generated models"),
    python.WithBaseClass("BaseModel"),
)

generator := python.NewPythonGenerator(options...)
```

### 3. Configure for Your Python Version

```go
// For modern Python (3.10+)
generator := python.NewPythonGenerator(
    python.WithPythonVersion("3.10"),
    python.WithTypeHintStyle("builtin"),
    python.WithOptional(false), // Use T | None
)

// For older Python (3.8+)
generator := python.NewPythonGenerator(
    python.WithPythonVersion("3.8"),
    python.WithTypeHintStyle("typing"),
    python.WithOptional(true), // Use Optional[T]
)
```

### 4. Organize Generated Code

```go
generator := python.NewPythonGenerator(
    python.WithFileHeader("# Generated by schema export system"),
    python.WithModuleName("api.models"),
    python.WithImportStyle("absolute"),
)
```

### 5. Handle Complex Types

```go
generator := python.NewPythonGenerator(
    python.WithCustomTypeMapping(map[string]string{
        "uuid": "UUID",
        "datetime": "datetime",
        "decimal": "Decimal",
    }),
    python.WithExtraImports([]string{
        "from uuid import UUID",
        "from datetime import datetime",
        "from decimal import Decimal",
    }),
)
```

## Error Handling

The generator provides detailed error messages for configuration issues:

```go
generator := python.NewPythonGenerator(
    python.WithOutputStyle("invalid"), // This will cause an error
)

result, err := generator.Generate(schema)
if err != nil {
    // Error: invalid Python option for field OutputStyle: invalid - 
    // unsupported output style (valid values: pydantic, dataclass, class, namedtuple)
    fmt.Printf("Error: %v\n", err)
}
```

## Performance Considerations

- Use minimal presets for large schemas when documentation isn't needed
- Disable examples and comments for production builds
- Consider using dataclasses for better performance than Pydantic in some cases

## Migration Guide

### From Embedded ToJSONSchema()

The Python generator replaces embedded `ToJSONSchema()` methods:

**Before:**
```go
schema.ToJSONSchema() // Limited customization
```

**After:**
```go
generator := python.NewPythonGenerator(
    python.WithOutputStyle("pydantic"),
    python.WithDocstrings(true),
    // Full customization available
)
result, err := generator.Generate(schema)
```

This provides much more flexibility and follows the visitor pattern architecture properly. 