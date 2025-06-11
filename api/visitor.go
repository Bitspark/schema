package api

// SchemaVisitor defines the visitor interface for schema traversal.
// This enables the visitor pattern for processing different schema types.
type SchemaVisitor interface {
	VisitString(StringSchema) error
	VisitNumber(NumberSchema) error
	VisitInteger(IntegerSchema) error
	VisitBoolean(BooleanSchema) error
	VisitArray(ArraySchema) error
	VisitObject(ObjectSchema) error
	VisitFunction(FunctionSchema) error
	VisitUnion(UnionSchema) error
}

// Accepter defines the interface for schemas that can accept visitors.
// This is the other half of the visitor pattern.
type Accepter interface {
	Accept(SchemaVisitor) error
}

// StringSchema interface for string-based schemas with introspection methods.
type StringSchema interface {
	Schema
	Accepter

	// Introspection methods
	MinLength() *int
	MaxLength() *int
	Pattern() string
	Format() string
	EnumValues() []string
	DefaultValue() *string
}

// NumberSchema interface for numeric schemas with introspection methods.
type NumberSchema interface {
	Schema
	Accepter

	// Introspection methods
	Minimum() *float64
	Maximum() *float64
}

// IntegerSchema interface for integer schemas with introspection methods.
type IntegerSchema interface {
	Schema
	Accepter

	// Introspection methods
	Minimum() *int64
	Maximum() *int64
}

// BooleanSchema interface for boolean schemas.
type BooleanSchema interface {
	Schema
	Accepter
}

// ArraySchema interface for array schemas with introspection methods.
type ArraySchema interface {
	Schema
	Accepter

	// Introspection methods
	ItemSchema() Schema
	MinItems() *int
	MaxItems() *int
	UniqueItemsRequired() bool
}

// ObjectSchema interface for object schemas with introspection methods.
type ObjectSchema interface {
	Schema
	Accepter

	// Introspection methods
	Properties() map[string]Schema
	Required() []string
	AdditionalProperties() bool
}

// FunctionSchema interface for function schemas with introspection methods.
type FunctionSchema interface {
	Schema
	Accepter

	// Introspection methods
	Inputs() map[string]Schema
	Outputs() Schema
	Errors() Schema
	Required() []string
}

// UnionSchema interface for union schemas with introspection methods.
type UnionSchema interface {
	Schema
	Accepter

	// Introspection methods
	Schemas() []Schema
}
