package api

// Builder defines the base interface for all schema builders.
type Builder[T Schema] interface {
	Build() T
}

// MetadataBuilder defines common methods for adding metadata to schemas.
type MetadataBuilder[T any] interface {
	Description(desc string) T
	Name(name string) T
	Tag(tag string) T
}

// StringSchemaBuilder defines the interface for building string schemas.
type StringSchemaBuilder interface {
	Builder[StringSchema]
	MetadataBuilder[StringSchemaBuilder]

	MinLength(min int) StringSchemaBuilder
	MaxLength(max int) StringSchemaBuilder
	Pattern(pattern string) StringSchemaBuilder
	Format(format string) StringSchemaBuilder
	Enum(values ...string) StringSchemaBuilder
	Default(value string) StringSchemaBuilder
	Example(example string) StringSchemaBuilder

	// Common formats
	Email() StringSchemaBuilder
	UUID() StringSchemaBuilder
	URL() StringSchemaBuilder
}

// NumberSchemaBuilder defines the interface for building number schemas.
type NumberSchemaBuilder interface {
	Builder[NumberSchema]
	MetadataBuilder[NumberSchemaBuilder]

	Min(min float64) NumberSchemaBuilder
	Max(max float64) NumberSchemaBuilder
	Range(min, max float64) NumberSchemaBuilder
	Example(example float64) NumberSchemaBuilder
}

// IntegerSchemaBuilder defines the interface for building integer schemas.
type IntegerSchemaBuilder interface {
	Builder[IntegerSchema]
	MetadataBuilder[IntegerSchemaBuilder]

	Min(min int64) IntegerSchemaBuilder
	Max(max int64) IntegerSchemaBuilder
	Range(min, max int64) IntegerSchemaBuilder
	Example(example int64) IntegerSchemaBuilder
}

// BooleanSchemaBuilder defines the interface for building boolean schemas.
type BooleanSchemaBuilder interface {
	Builder[BooleanSchema]
	MetadataBuilder[BooleanSchemaBuilder]

	Example(example bool) BooleanSchemaBuilder
}

// ArraySchemaBuilder defines the interface for building array schemas.
type ArraySchemaBuilder interface {
	Builder[ArraySchema]
	MetadataBuilder[ArraySchemaBuilder]

	Items(itemSchema Schema) ArraySchemaBuilder
	MinItems(min int) ArraySchemaBuilder
	MaxItems(max int) ArraySchemaBuilder
	UniqueItems() ArraySchemaBuilder
	Example(example []any) ArraySchemaBuilder
}

// ObjectSchemaBuilder defines the interface for building object schemas.
type ObjectSchemaBuilder interface {
	Builder[ObjectSchema]
	MetadataBuilder[ObjectSchemaBuilder]

	Property(name string, schema Schema) ObjectSchemaBuilder
	Required(names ...string) ObjectSchemaBuilder
	AdditionalProperties(allowed bool) ObjectSchemaBuilder
	Example(example map[string]any) ObjectSchemaBuilder
}

// FunctionSchemaBuilder defines the interface for building function schemas.
type FunctionSchemaBuilder interface {
	Builder[FunctionSchema]
	MetadataBuilder[FunctionSchemaBuilder]

	Input(name string, schema Schema) FunctionSchemaBuilder
	Output(schema Schema) FunctionSchemaBuilder
	Error(schema Schema) FunctionSchemaBuilder
	Required(names ...string) FunctionSchemaBuilder
	Example(example map[string]any) FunctionSchemaBuilder
}

// UnionSchemaBuilder defines the interface for building union schemas.
type UnionSchemaBuilder interface {
	Builder[UnionSchema]
	MetadataBuilder[UnionSchemaBuilder]

	Schemas(schemas ...Schema) UnionSchemaBuilder
}
