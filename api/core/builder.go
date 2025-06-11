package core

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
	Default(value float64) NumberSchemaBuilder

	// Common number helpers
	Positive() NumberSchemaBuilder
	NonNegative() NumberSchemaBuilder
	Percentage() NumberSchemaBuilder
	Ratio() NumberSchemaBuilder
}

// IntegerSchemaBuilder defines the interface for building integer schemas.
type IntegerSchemaBuilder interface {
	Builder[IntegerSchema]
	MetadataBuilder[IntegerSchemaBuilder]

	Min(min int64) IntegerSchemaBuilder
	Max(max int64) IntegerSchemaBuilder
	Range(min, max int64) IntegerSchemaBuilder
	Example(example int64) IntegerSchemaBuilder
	Default(value int64) IntegerSchemaBuilder

	// Common integer helpers
	Positive() IntegerSchemaBuilder
	NonNegative() IntegerSchemaBuilder
	Port() IntegerSchemaBuilder
	Age() IntegerSchemaBuilder
	ID() IntegerSchemaBuilder
	Count() IntegerSchemaBuilder
}

// BooleanSchemaBuilder defines the interface for building boolean schemas.
type BooleanSchemaBuilder interface {
	Builder[BooleanSchema]
	MetadataBuilder[BooleanSchemaBuilder]

	Example(example bool) BooleanSchemaBuilder
	Default(value bool) BooleanSchemaBuilder
	AllowStringConversion() BooleanSchemaBuilder
	CaseInsensitive() BooleanSchemaBuilder

	// Common boolean helpers
	Required() BooleanSchemaBuilder
	Flag() BooleanSchemaBuilder
	Switch() BooleanSchemaBuilder
	Enabled() BooleanSchemaBuilder
	Active() BooleanSchemaBuilder
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
	Default(value []any) ArraySchemaBuilder
	Contains(schema Schema) ArraySchemaBuilder
	Length(length int) ArraySchemaBuilder
	Range(min, max int) ArraySchemaBuilder

	// Common array helpers
	NonEmpty() ArraySchemaBuilder
	StringArray() ArraySchemaBuilder
	NumberArray() ArraySchemaBuilder
	IntegerArray() ArraySchemaBuilder
	BooleanArray() ArraySchemaBuilder
	List() ArraySchemaBuilder
	Set() ArraySchemaBuilder
	Tuple(length int) ArraySchemaBuilder
	LimitedList(maxItems int) ArraySchemaBuilder
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
	Output(name string, schema Schema) FunctionSchemaBuilder
	Error(schema Schema) FunctionSchemaBuilder
	RequiredInputs(names ...string) FunctionSchemaBuilder
	RequiredOutputs(names ...string) FunctionSchemaBuilder
	Example(example map[string]any) FunctionSchemaBuilder
}

// ServiceSchemaBuilder defines the interface for building service schemas.
type ServiceSchemaBuilder interface {
	Builder[ServiceSchema]
	MetadataBuilder[ServiceSchemaBuilder]

	Method(name string, functionSchema FunctionSchema) ServiceSchemaBuilder
	FromStruct(instance any) ServiceSchemaBuilder
	Example(example map[string]any) ServiceSchemaBuilder
}

// UnionSchemaBuilder defines the interface for building union schemas.
type UnionSchemaBuilder interface {
	Builder[UnionSchema]
	MetadataBuilder[UnionSchemaBuilder]

	Schemas(schemas ...Schema) UnionSchemaBuilder
}
