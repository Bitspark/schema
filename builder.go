package schema

import "fmt"

// String builder
func String() *StringBuilder {
	return &StringBuilder{
		schema: &StringSchema{
			metadata: SchemaMetadata{},
		},
	}
}

type StringBuilder struct {
	schema *StringSchema
}

func (b *StringBuilder) Description(desc string) *StringBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *StringBuilder) Name(name string) *StringBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *StringBuilder) MinLength(min int) *StringBuilder {
	b.schema.minLength = &min
	return b
}

func (b *StringBuilder) MaxLength(max int) *StringBuilder {
	b.schema.maxLength = &max
	return b
}

func (b *StringBuilder) Pattern(pattern string) *StringBuilder {
	b.schema.pattern = pattern
	return b
}

func (b *StringBuilder) Enum(values ...string) *StringBuilder {
	b.schema.enumValues = values
	return b
}

func (b *StringBuilder) Default(value string) *StringBuilder {
	b.schema.defaultVal = &value
	return b
}

func (b *StringBuilder) Email() *StringBuilder {
	b.schema.format = "email"
	if b.schema.metadata.Description == "" {
		b.schema.metadata.Description = "Valid email address"
	}
	b.schema.metadata.Examples = []any{"user@example.com"}
	return b
}

func (b *StringBuilder) UUID() *StringBuilder {
	b.schema.format = "uuid"
	if b.schema.metadata.Description == "" {
		b.schema.metadata.Description = "UUID identifier"
	}
	b.schema.metadata.Examples = []any{"123e4567-e89b-12d3-a456-426614174000"}
	return b
}

func (b *StringBuilder) URL() *StringBuilder {
	b.schema.format = "url"
	if b.schema.metadata.Description == "" {
		b.schema.metadata.Description = "Valid URL"
	}
	b.schema.metadata.Examples = []any{"https://example.com"}
	return b
}

func (b *StringBuilder) Example(example string) *StringBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *StringBuilder) Tag(tag string) *StringBuilder {
	b.schema.metadata.Tags = append(b.schema.metadata.Tags, tag)
	return b
}

func (b *StringBuilder) Build() Schema {
	return b.schema
}

// Object builder
func Object() *ObjectBuilder {
	return &ObjectBuilder{
		schema: &ObjectSchema{
			metadata:   SchemaMetadata{},
			properties: make(map[string]Schema),
			required:   []string{},
		},
	}
}

type ObjectBuilder struct {
	schema *ObjectSchema
}

func (b *ObjectBuilder) Property(name string, schema Schema) *ObjectBuilder {
	b.schema.properties[name] = schema
	return b
}

func (b *ObjectBuilder) Required(names ...string) *ObjectBuilder {
	b.schema.required = append(b.schema.required, names...)
	return b
}

func (b *ObjectBuilder) AdditionalProperties(allowed bool) *ObjectBuilder {
	b.schema.additionalProps = allowed
	return b
}

func (b *ObjectBuilder) Description(desc string) *ObjectBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *ObjectBuilder) Name(name string) *ObjectBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *ObjectBuilder) Example(example map[string]any) *ObjectBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *ObjectBuilder) Tag(tag string) *ObjectBuilder {
	b.schema.metadata.Tags = append(b.schema.metadata.Tags, tag)
	return b
}

func (b *ObjectBuilder) Build() Schema {
	return b.schema
}

// Number builder
func Number() *NumberBuilder {
	return &NumberBuilder{
		schema: &NumberSchema{
			metadata: SchemaMetadata{},
		},
	}
}

type NumberSchema struct {
	metadata SchemaMetadata
	minimum  *float64
	maximum  *float64
}

func (s *NumberSchema) Type() SchemaType {
	return TypeNumber
}

func (s *NumberSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *NumberSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *NumberSchema) Clone() Schema {
	clone := *s
	return &clone
}

func (s *NumberSchema) Validate(value any) ValidationResult {
	var num float64
	var ok bool

	switch v := value.(type) {
	case float64:
		num, ok = v, true
	case float32:
		num, ok = float64(v), true
	case int:
		num, ok = float64(v), true
	case int32:
		num, ok = float64(v), true
	case int64:
		num, ok = float64(v), true
	}

	if !ok {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Path:       "",
				Message:    "Expected number",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "number",
				Suggestion: "Provide a numeric value",
			}},
		}
	}

	var errors []ValidationError

	if s.minimum != nil && num < *s.minimum {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Number too small (minimum %.2f)", *s.minimum),
			Code:       "min_value",
			Value:      num,
			Suggestion: fmt.Sprintf("Provide a value >= %.2f", *s.minimum),
		})
	}

	if s.maximum != nil && num > *s.maximum {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Number too large (maximum %.2f)", *s.maximum),
			Code:       "max_value",
			Value:      num,
			Suggestion: fmt.Sprintf("Provide a value <= %.2f", *s.maximum),
		})
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (s *NumberSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "number",
	}

	if s.minimum != nil {
		schema["minimum"] = *s.minimum
	}
	if s.maximum != nil {
		schema["maximum"] = *s.maximum
	}
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}
	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	return schema
}

func (s *NumberSchema) GenerateExample() any {
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}

	if s.minimum != nil && s.maximum != nil {
		return (*s.minimum + *s.maximum) / 2
	}

	if s.minimum != nil {
		return *s.minimum + 1
	}

	if s.maximum != nil {
		return *s.maximum - 1
	}

	return 42.0
}

type NumberBuilder struct {
	schema *NumberSchema
}

func (b *NumberBuilder) Min(min float64) *NumberBuilder {
	b.schema.minimum = &min
	return b
}

func (b *NumberBuilder) Max(max float64) *NumberBuilder {
	b.schema.maximum = &max
	return b
}

func (b *NumberBuilder) Range(min, max float64) *NumberBuilder {
	b.schema.minimum = &min
	b.schema.maximum = &max
	return b
}

func (b *NumberBuilder) Description(desc string) *NumberBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *NumberBuilder) Name(name string) *NumberBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *NumberBuilder) Example(example float64) *NumberBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *NumberBuilder) Build() Schema {
	return b.schema
}

// Integer builder
func Integer() *IntegerBuilder {
	return &IntegerBuilder{
		schema: &IntegerSchema{
			metadata: SchemaMetadata{},
		},
	}
}

type IntegerSchema struct {
	metadata SchemaMetadata
	minimum  *int64
	maximum  *int64
}

func (s *IntegerSchema) Type() SchemaType {
	return TypeInteger
}

func (s *IntegerSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *IntegerSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *IntegerSchema) Clone() Schema {
	clone := *s
	return &clone
}

func (s *IntegerSchema) Validate(value any) ValidationResult {
	var num int64
	var ok bool

	switch v := value.(type) {
	case int:
		num, ok = int64(v), true
	case int32:
		num, ok = int64(v), true
	case int64:
		num, ok = v, true
	case float64:
		if v == float64(int64(v)) {
			num, ok = int64(v), true
		}
	}

	if !ok {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Path:       "",
				Message:    "Expected integer",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "integer",
				Suggestion: "Provide an integer value",
			}},
		}
	}

	var errors []ValidationError

	if s.minimum != nil && num < *s.minimum {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Integer too small (minimum %d)", *s.minimum),
			Code:       "min_value",
			Value:      num,
			Suggestion: fmt.Sprintf("Provide a value >= %d", *s.minimum),
		})
	}

	if s.maximum != nil && num > *s.maximum {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Integer too large (maximum %d)", *s.maximum),
			Code:       "max_value",
			Value:      num,
			Suggestion: fmt.Sprintf("Provide a value <= %d", *s.maximum),
		})
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (s *IntegerSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "integer",
	}

	if s.minimum != nil {
		schema["minimum"] = *s.minimum
	}
	if s.maximum != nil {
		schema["maximum"] = *s.maximum
	}
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}
	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	return schema
}

func (s *IntegerSchema) GenerateExample() any {
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}

	if s.minimum != nil && s.maximum != nil {
		return (*s.minimum + *s.maximum) / 2
	}

	if s.minimum != nil {
		return *s.minimum + 1
	}

	if s.maximum != nil {
		return *s.maximum - 1
	}

	return int64(42)
}

type IntegerBuilder struct {
	schema *IntegerSchema
}

func (b *IntegerBuilder) Min(min int64) *IntegerBuilder {
	b.schema.minimum = &min
	return b
}

func (b *IntegerBuilder) Max(max int64) *IntegerBuilder {
	b.schema.maximum = &max
	return b
}

func (b *IntegerBuilder) Range(min, max int64) *IntegerBuilder {
	b.schema.minimum = &min
	b.schema.maximum = &max
	return b
}

func (b *IntegerBuilder) Description(desc string) *IntegerBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *IntegerBuilder) Name(name string) *IntegerBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *IntegerBuilder) Example(example int64) *IntegerBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *IntegerBuilder) Build() Schema {
	return b.schema
}

// Boolean builder
func Boolean() *BooleanBuilder {
	return &BooleanBuilder{
		schema: &BooleanSchema{
			metadata: SchemaMetadata{},
		},
	}
}

type BooleanSchema struct {
	metadata SchemaMetadata
}

func (s *BooleanSchema) Type() SchemaType {
	return TypeBoolean
}

func (s *BooleanSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *BooleanSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *BooleanSchema) Clone() Schema {
	clone := *s
	return &clone
}

func (s *BooleanSchema) Validate(value any) ValidationResult {
	_, ok := value.(bool)
	if !ok {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Path:       "",
				Message:    "Expected boolean",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "boolean",
				Suggestion: "Provide true or false",
			}},
		}
	}

	return ValidationResult{Valid: true}
}

func (s *BooleanSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "boolean",
	}

	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}
	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	return schema
}

func (s *BooleanSchema) GenerateExample() any {
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}
	return true
}

type BooleanBuilder struct {
	schema *BooleanSchema
}

func (b *BooleanBuilder) Description(desc string) *BooleanBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *BooleanBuilder) Name(name string) *BooleanBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *BooleanBuilder) Example(example bool) *BooleanBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *BooleanBuilder) Build() Schema {
	return b.schema
}

// Array builder
func Array() *ArrayBuilder {
	return &ArrayBuilder{
		schema: &ArraySchema{
			metadata: SchemaMetadata{},
		},
	}
}

type ArraySchema struct {
	metadata    SchemaMetadata
	itemSchema  Schema
	minItems    *int
	maxItems    *int
	uniqueItems bool
}

func (s *ArraySchema) ItemSchema() Schema {
	return s.itemSchema
}

func (s *ArraySchema) MinItems() *int {
	if s.minItems == nil {
		return nil
	}
	val := *s.minItems
	return &val
}

func (s *ArraySchema) MaxItems() *int {
	if s.maxItems == nil {
		return nil
	}
	val := *s.maxItems
	return &val
}

func (s *ArraySchema) UniqueItemsRequired() bool {
	return s.uniqueItems
}

func (s *ArraySchema) Type() SchemaType {
	return TypeArray
}

func (s *ArraySchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *ArraySchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *ArraySchema) Clone() Schema {
	clone := *s
	if s.itemSchema != nil {
		clone.itemSchema = s.itemSchema.Clone()
	}
	return &clone
}

func (s *ArraySchema) Validate(value any) ValidationResult {
	arr, ok := value.([]any)
	if !ok {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Path:       "",
				Message:    "Expected array",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "array",
				Suggestion: "Provide an array of values",
			}},
		}
	}

	var errors []ValidationError

	// Length validation
	if s.minItems != nil && len(arr) < *s.minItems {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Array too short (minimum %d items)", *s.minItems),
			Code:       "min_items",
			Value:      arr,
			Suggestion: fmt.Sprintf("Provide at least %d items", *s.minItems),
		})
	}

	if s.maxItems != nil && len(arr) > *s.maxItems {
		errors = append(errors, ValidationError{
			Path:       "",
			Message:    fmt.Sprintf("Array too long (maximum %d items)", *s.maxItems),
			Code:       "max_items",
			Value:      arr,
			Suggestion: fmt.Sprintf("Limit to %d items", *s.maxItems),
		})
	}

	// Validate each item
	if s.itemSchema != nil {
		for i, item := range arr {
			result := s.itemSchema.Validate(item)
			if !result.Valid {
				for _, err := range result.Errors {
					if err.Path == "" {
						err.Path = fmt.Sprintf("[%d]", i)
					} else {
						err.Path = fmt.Sprintf("[%d].%s", i, err.Path)
					}
					errors = append(errors, err)
				}
			}
		}
	}

	// Unique items validation
	if s.uniqueItems {
		seen := make(map[string]bool)
		for i, item := range arr {
			key := fmt.Sprintf("%v", item)
			if seen[key] {
				errors = append(errors, ValidationError{
					Path:       fmt.Sprintf("[%d]", i),
					Message:    "Duplicate item found",
					Code:       "unique_items",
					Value:      item,
					Suggestion: "Remove duplicate items",
				})
			}
			seen[key] = true
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (s *ArraySchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type": "array",
	}

	if s.itemSchema != nil {
		schema["items"] = s.itemSchema.ToJSONSchema()
	}
	if s.minItems != nil {
		schema["minItems"] = *s.minItems
	}
	if s.maxItems != nil {
		schema["maxItems"] = *s.maxItems
	}
	if s.uniqueItems {
		schema["uniqueItems"] = s.uniqueItems
	}
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}
	if len(s.metadata.Examples) > 0 {
		schema["examples"] = s.metadata.Examples
	}

	return schema
}

func (s *ArraySchema) GenerateExample() any {
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}

	if s.itemSchema == nil {
		return []any{}
	}

	// Generate a small array with examples
	count := 2
	if s.minItems != nil && *s.minItems > count {
		count = *s.minItems
	}
	if s.maxItems != nil && *s.maxItems < count {
		count = *s.maxItems
	}

	result := make([]any, count)
	for i := 0; i < count; i++ {
		result[i] = s.itemSchema.GenerateExample()
	}

	return result
}

type ArrayBuilder struct {
	schema *ArraySchema
}

func (b *ArrayBuilder) Items(itemSchema Schema) *ArrayBuilder {
	b.schema.itemSchema = itemSchema
	return b
}

func (b *ArrayBuilder) MinItems(min int) *ArrayBuilder {
	b.schema.minItems = &min
	return b
}

func (b *ArrayBuilder) MaxItems(max int) *ArrayBuilder {
	b.schema.maxItems = &max
	return b
}

func (b *ArrayBuilder) UniqueItems() *ArrayBuilder {
	b.schema.uniqueItems = true
	return b
}

func (b *ArrayBuilder) Description(desc string) *ArrayBuilder {
	b.schema.metadata.Description = desc
	return b
}

func (b *ArrayBuilder) Name(name string) *ArrayBuilder {
	b.schema.metadata.Name = name
	return b
}

func (b *ArrayBuilder) Example(example []any) *ArrayBuilder {
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, example)
	return b
}

func (b *ArrayBuilder) Build() Schema {
	return b.schema
}
