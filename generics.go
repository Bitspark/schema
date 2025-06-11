// Package schema - Generic schema patterns for type-safe, ergonomic schema construction
package schema

import (
	"fmt"
	"reflect"
)

// Generic Schema Patterns
// These provide type-safe, ergonomic APIs for common schema patterns

// List creates a schema for []T with type safety and rich validation options.
// Example: userListSchema := schema.List[User]()
func List[T any]() *ListBuilder[T] {
	itemSchema := FromStruct[T]()
	return &ListBuilder[T]{
		schema: &ArraySchema{
			metadata:    SchemaMetadata{Name: fmt.Sprintf("List[%s]", getGenericTypeName[T]())},
			itemSchema:  itemSchema,
			minItems:    nil,
			maxItems:    nil,
			uniqueItems: false,
		},
	}
}

// ListBuilder provides a fluent interface for building List[T] schemas.
type ListBuilder[T any] struct {
	schema *ArraySchema
}

func (b *ListBuilder[T]) MinItems(min int) *ListBuilder[T] {
	b.schema.minItems = &min
	return b
}

func (b *ListBuilder[T]) MaxItems(max int) *ListBuilder[T] {
	b.schema.maxItems = &max
	return b
}

func (b *ListBuilder[T]) UniqueItems() *ListBuilder[T] {
	b.schema.uniqueItems = true
	return b
}

func (b *ListBuilder[T]) Description(desc string) *ListBuilder[T] {
	b.schema.metadata.Description = desc
	return b
}

func (b *ListBuilder[T]) Name(name string) *ListBuilder[T] {
	b.schema.metadata.Name = name
	return b
}

func (b *ListBuilder[T]) Example(example []T) *ListBuilder[T] {
	// Convert to []any for storage
	var anyExample []any
	for _, item := range example {
		anyExample = append(anyExample, item)
	}
	b.schema.metadata.Examples = append(b.schema.metadata.Examples, anyExample)
	return b
}

func (b *ListBuilder[T]) Tag(tag string) *ListBuilder[T] {
	b.schema.metadata.Tags = append(b.schema.metadata.Tags, tag)
	return b
}

func (b *ListBuilder[T]) Build() Schema {
	return b.schema
}

// Optional creates a schema for *T with smart nil handling and validation.
// Example: optionalUserSchema := schema.Optional[User]()
func Optional[T any]() *OptionalBuilder[T] {
	itemSchema := FromStruct[T]()
	return &OptionalBuilder[T]{
		itemSchema: itemSchema,
		metadata:   SchemaMetadata{Name: fmt.Sprintf("Optional[%s]", getGenericTypeName[T]())},
	}
}

// OptionalBuilder provides a fluent interface for building Optional[T] schemas.
type OptionalBuilder[T any] struct {
	itemSchema Schema
	metadata   SchemaMetadata
}

func (b *OptionalBuilder[T]) Description(desc string) *OptionalBuilder[T] {
	b.metadata.Description = desc
	return b
}

func (b *OptionalBuilder[T]) Name(name string) *OptionalBuilder[T] {
	b.metadata.Name = name
	return b
}

func (b *OptionalBuilder[T]) Example(example *T) *OptionalBuilder[T] {
	if example != nil {
		b.metadata.Examples = append(b.metadata.Examples, *example)
	} else {
		b.metadata.Examples = append(b.metadata.Examples, nil)
	}
	return b
}

func (b *OptionalBuilder[T]) Tag(tag string) *OptionalBuilder[T] {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

func (b *OptionalBuilder[T]) Build() Schema {
	return &OptionalSchema[T]{
		metadata:   b.metadata,
		itemSchema: b.itemSchema,
	}
}

// OptionalSchema represents an optional/nullable schema that can be null or of type T.
type OptionalSchema[T any] struct {
	metadata   SchemaMetadata
	itemSchema Schema
}

func (s *OptionalSchema[T]) Type() SchemaType {
	return TypeOptional
}

func (s *OptionalSchema[T]) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *OptionalSchema[T]) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *OptionalSchema[T]) Clone() Schema {
	clone := *s
	clone.itemSchema = s.itemSchema.Clone()
	return &clone
}

func (s *OptionalSchema[T]) Validate(value any) ValidationResult {
	// Handle nil values
	if value == nil {
		return ValidationResult{Valid: true}
	}

	// Validate non-nil values against item schema
	return s.itemSchema.Validate(value)
}

func (s *OptionalSchema[T]) ToJSONSchema() map[string]any {
	itemJSON := s.itemSchema.ToJSONSchema()

	// Create oneOf with null and the item type
	return map[string]any{
		"oneOf": []map[string]any{
			{"type": "null"},
			itemJSON,
		},
		"description": s.metadata.Description,
	}
}

func (s *OptionalSchema[T]) GenerateExample() any {
	// 50% chance of null, 50% chance of real example
	if len(s.metadata.Examples) > 0 {
		return s.metadata.Examples[0]
	}

	// Generate example from item schema
	return s.itemSchema.GenerateExample()
}

// Result creates a schema for Result[T, E] patterns (success/failure).
// Example: resultSchema := schema.Result[User, ValidationError]()
func Result[T, E any]() *ResultBuilder[T, E] {
	successSchema := FromStruct[T]()
	errorSchema := FromStruct[E]()

	return &ResultBuilder[T, E]{
		successSchema: successSchema,
		errorSchema:   errorSchema,
		metadata:      SchemaMetadata{Name: fmt.Sprintf("Result[%s, %s]", getGenericTypeName[T](), getGenericTypeName[E]())},
	}
}

// ResultBuilder provides a fluent interface for building Result[T, E] schemas.
type ResultBuilder[T, E any] struct {
	successSchema Schema
	errorSchema   Schema
	metadata      SchemaMetadata
}

func (b *ResultBuilder[T, E]) Description(desc string) *ResultBuilder[T, E] {
	b.metadata.Description = desc
	return b
}

func (b *ResultBuilder[T, E]) Name(name string) *ResultBuilder[T, E] {
	b.metadata.Name = name
	return b
}

func (b *ResultBuilder[T, E]) Tag(tag string) *ResultBuilder[T, E] {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

func (b *ResultBuilder[T, E]) Build() Schema {
	return &ResultSchema[T, E]{
		metadata:      b.metadata,
		successSchema: b.successSchema,
		errorSchema:   b.errorSchema,
	}
}

// ResultSchema represents a Result[T, E] that can be either success (T) or error (E).
type ResultSchema[T, E any] struct {
	metadata      SchemaMetadata
	successSchema Schema
	errorSchema   Schema
}

func (s *ResultSchema[T, E]) Type() SchemaType {
	return TypeResult
}

func (s *ResultSchema[T, E]) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *ResultSchema[T, E]) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *ResultSchema[T, E]) Clone() Schema {
	clone := *s
	clone.successSchema = s.successSchema.Clone()
	clone.errorSchema = s.errorSchema.Clone()
	return &clone
}

func (s *ResultSchema[T, E]) Validate(value any) ValidationResult {
	// Try to validate as success type first
	if result := s.successSchema.Validate(value); result.Valid {
		return result
	}

	// If not success, try error type
	return s.errorSchema.Validate(value)
}

func (s *ResultSchema[T, E]) ToJSONSchema() map[string]any {
	successJSON := s.successSchema.ToJSONSchema()
	errorJSON := s.errorSchema.ToJSONSchema()

	return map[string]any{
		"oneOf": []map[string]any{
			{
				"type": "object",
				"properties": map[string]any{
					"success": successJSON,
				},
				"required":             []string{"success"},
				"additionalProperties": false,
			},
			{
				"type": "object",
				"properties": map[string]any{
					"error": errorJSON,
				},
				"required":             []string{"error"},
				"additionalProperties": false,
			},
		},
		"description": s.metadata.Description,
	}
}

func (s *ResultSchema[T, E]) GenerateExample() any {
	// Generate success example by default
	return map[string]any{
		"success": s.successSchema.GenerateExample(),
	}
}

// Map creates a schema for map[K]V with typed keys and values.
// Example: userMapSchema := schema.Map[string, User]()
func Map[K comparable, V any]() *MapBuilder[K, V] {
	keySchema := generateSchemaForType[K]()
	valueSchema := FromStruct[V]()

	return &MapBuilder[K, V]{
		keySchema:   keySchema,
		valueSchema: valueSchema,
		metadata:    SchemaMetadata{Name: fmt.Sprintf("Map[%s, %s]", getGenericTypeName[K](), getGenericTypeName[V]())},
	}
}

// MapBuilder provides a fluent interface for building Map[K, V] schemas.
type MapBuilder[K comparable, V any] struct {
	keySchema   Schema
	valueSchema Schema
	metadata    SchemaMetadata
	minItems    *int
	maxItems    *int
}

func (b *MapBuilder[K, V]) MinItems(min int) *MapBuilder[K, V] {
	b.minItems = &min
	return b
}

func (b *MapBuilder[K, V]) MaxItems(max int) *MapBuilder[K, V] {
	b.maxItems = &max
	return b
}

func (b *MapBuilder[K, V]) Description(desc string) *MapBuilder[K, V] {
	b.metadata.Description = desc
	return b
}

func (b *MapBuilder[K, V]) Name(name string) *MapBuilder[K, V] {
	b.metadata.Name = name
	return b
}

func (b *MapBuilder[K, V]) Tag(tag string) *MapBuilder[K, V] {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

func (b *MapBuilder[K, V]) Build() Schema {
	return &MapSchema[K, V]{
		metadata:    b.metadata,
		keySchema:   b.keySchema,
		valueSchema: b.valueSchema,
		minItems:    b.minItems,
		maxItems:    b.maxItems,
	}
}

// MapSchema represents a Map[K, V] schema with typed keys and values.
type MapSchema[K comparable, V any] struct {
	metadata    SchemaMetadata
	keySchema   Schema
	valueSchema Schema
	minItems    *int
	maxItems    *int
}

func (s *MapSchema[K, V]) Type() SchemaType {
	return TypeMap
}

func (s *MapSchema[K, V]) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *MapSchema[K, V]) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *MapSchema[K, V]) Clone() Schema {
	clone := *s
	clone.keySchema = s.keySchema.Clone()
	clone.valueSchema = s.valueSchema.Clone()
	return &clone
}

func (s *MapSchema[K, V]) Validate(value any) ValidationResult {
	// Try to handle different map types more flexibly
	var mapLen int
	var keyValuePairs []struct{ key, value any }

	switch v := value.(type) {
	case map[K]V:
		mapLen = len(v)
		for key, val := range v {
			keyValuePairs = append(keyValuePairs, struct{ key, value any }{key, val})
		}
	case map[string]any:
		if getGenericTypeName[K]() == "string" {
			mapLen = len(v)
			for key, val := range v {
				keyValuePairs = append(keyValuePairs, struct{ key, value any }{key, val})
			}
		} else {
			return ValidationResult{
				Valid:  false,
				Errors: []ValidationError{{Path: "", Message: fmt.Sprintf("expected map[%s]%s, got %T", getGenericTypeName[K](), getGenericTypeName[V](), value)}},
			}
		}
	case map[string]map[string]any:
		if getGenericTypeName[K]() == "string" {
			mapLen = len(v)
			for key, val := range v {
				keyValuePairs = append(keyValuePairs, struct{ key, value any }{key, val})
			}
		} else {
			return ValidationResult{
				Valid:  false,
				Errors: []ValidationError{{Path: "", Message: fmt.Sprintf("expected map[%s]%s, got %T", getGenericTypeName[K](), getGenericTypeName[V](), value)}},
			}
		}
	default:
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Path: "", Message: fmt.Sprintf("expected map[%s]%s, got %T", getGenericTypeName[K](), getGenericTypeName[V](), value)}},
		}
	}

	var errors []ValidationError

	// Validate map size
	if s.minItems != nil && mapLen < *s.minItems {
		errors = append(errors, ValidationError{
			Path:    "",
			Message: fmt.Sprintf("map has %d items, minimum is %d", mapLen, *s.minItems),
		})
	}

	if s.maxItems != nil && mapLen > *s.maxItems {
		errors = append(errors, ValidationError{
			Path:    "",
			Message: fmt.Sprintf("map has %d items, maximum is %d", mapLen, *s.maxItems),
		})
	}

	// Validate each key-value pair
	for _, pair := range keyValuePairs {
		// Validate key
		if keyResult := s.keySchema.Validate(pair.key); !keyResult.Valid {
			for _, err := range keyResult.Errors {
				errors = append(errors, ValidationError{
					Path:    fmt.Sprintf("key(%v).%s", pair.key, err.Path),
					Message: err.Message,
				})
			}
		}

		// Validate value
		if valResult := s.valueSchema.Validate(pair.value); !valResult.Valid {
			for _, err := range valResult.Errors {
				errors = append(errors, ValidationError{
					Path:    fmt.Sprintf("[%v].%s", pair.key, err.Path),
					Message: err.Message,
				})
			}
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (s *MapSchema[K, V]) ToJSONSchema() map[string]any {
	valueJSON := s.valueSchema.ToJSONSchema()

	result := map[string]any{
		"type":                 "object",
		"additionalProperties": valueJSON,
		"description":          s.metadata.Description,
	}

	if s.minItems != nil {
		result["minProperties"] = *s.minItems
	}

	if s.maxItems != nil {
		result["maxProperties"] = *s.maxItems
	}

	return result
}

func (s *MapSchema[K, V]) GenerateExample() any {
	// Generate a simple example map
	key := s.keySchema.GenerateExample()
	value := s.valueSchema.GenerateExample()

	return map[string]any{
		fmt.Sprintf("%v", key): value,
	}
}

// Union creates a schema for union types (oneOf in JSON Schema).
// Example: unionSchema := schema.Union[string, int]()
func Union[T1, T2 any]() *UnionBuilder2[T1, T2] {
	schema1 := generateSchemaForType[T1]()
	schema2 := generateSchemaForType[T2]()

	return &UnionBuilder2[T1, T2]{
		schemas:  []Schema{schema1, schema2},
		metadata: SchemaMetadata{Name: fmt.Sprintf("Union[%s, %s]", getGenericTypeName[T1](), getGenericTypeName[T2]())},
	}
}

// UnionBuilder2 provides a fluent interface for building Union[T1, T2] schemas.
type UnionBuilder2[T1, T2 any] struct {
	schemas  []Schema
	metadata SchemaMetadata
}

func (b *UnionBuilder2[T1, T2]) Description(desc string) *UnionBuilder2[T1, T2] {
	b.metadata.Description = desc
	return b
}

func (b *UnionBuilder2[T1, T2]) Name(name string) *UnionBuilder2[T1, T2] {
	b.metadata.Name = name
	return b
}

func (b *UnionBuilder2[T1, T2]) Tag(tag string) *UnionBuilder2[T1, T2] {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

func (b *UnionBuilder2[T1, T2]) Build() Schema {
	return &UnionSchema{
		metadata: b.metadata,
		schemas:  b.schemas,
	}
}

// UnionSchema represents a union type that can be one of several types.
type UnionSchema struct {
	metadata SchemaMetadata
	schemas  []Schema
}

// Introspection methods for UnionSchema
func (s *UnionSchema) Schemas() []Schema {
	// Return a copy to prevent external mutation
	schemas := make([]Schema, len(s.schemas))
	copy(schemas, s.schemas)
	return schemas
}

func (s *UnionSchema) Type() SchemaType {
	return TypeUnion
}

func (s *UnionSchema) Metadata() SchemaMetadata {
	return s.metadata
}

func (s *UnionSchema) WithMetadata(metadata SchemaMetadata) Schema {
	clone := *s
	clone.metadata = metadata
	return &clone
}

func (s *UnionSchema) Clone() Schema {
	clone := *s
	clone.schemas = make([]Schema, len(s.schemas))
	for i, schema := range s.schemas {
		clone.schemas[i] = schema.Clone()
	}
	return &clone
}

func (s *UnionSchema) Validate(value any) ValidationResult {
	var allErrors []ValidationError

	// Value is valid if it matches any of the union schemas
	for i, schema := range s.schemas {
		if result := schema.Validate(value); result.Valid {
			return ValidationResult{Valid: true}
		} else {
			// Collect errors for debugging
			for _, err := range result.Errors {
				allErrors = append(allErrors, ValidationError{
					Path:    fmt.Sprintf("union[%d].%s", i, err.Path),
					Message: err.Message,
				})
			}
		}
	}

	return ValidationResult{
		Valid: false,
		Errors: []ValidationError{{
			Path:    "",
			Message: fmt.Sprintf("value does not match any union type (%d schemas tried)", len(s.schemas)),
		}},
	}
}

func (s *UnionSchema) ToJSONSchema() map[string]any {
	var oneOf []map[string]any
	for _, schema := range s.schemas {
		oneOf = append(oneOf, schema.ToJSONSchema())
	}

	return map[string]any{
		"oneOf":       oneOf,
		"description": s.metadata.Description,
	}
}

func (s *UnionSchema) GenerateExample() any {
	// Generate example from first schema
	if len(s.schemas) > 0 {
		return s.schemas[0].GenerateExample()
	}
	return nil
}

// Helper functions for generic patterns

// generateSchemaForType generates a schema for a specific type T using reflection.
func generateSchemaForType[T any]() Schema {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	return generateSchemaFromType(typ)
}

// getGenericTypeName returns a human-readable name for a generic type T.
func getGenericTypeName[T any]() string {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	return getTypeName(typ)
}

// Convenience functions for common patterns

// StringList creates a schema for []string with validation.
func StringList() *ListBuilder[string] {
	return List[string]()
}

// IntList creates a schema for []int with validation.
func IntList() *ListBuilder[int] {
	return List[int]()
}

// StringMap creates a schema for map[string]string.
func StringMap() *MapBuilder[string, string] {
	return Map[string, string]()
}

// StringOptional creates a schema for *string.
func StringOptional() *OptionalBuilder[string] {
	return Optional[string]()
}

// IntOptional creates a schema for *int.
func IntOptional() *OptionalBuilder[int] {
	return Optional[int]()
}

// BoolOptional creates a schema for *bool.
func BoolOptional() *OptionalBuilder[bool] {
	return Optional[bool]()
}
