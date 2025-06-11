package api

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

// ArgSchema represents a named argument with its schema and description.
// This is used for both function inputs and outputs to provide rich metadata.
type ArgSchema interface {
	Accepter

	// Introspection methods
	Name() string
	Description() string
	Schema() Schema
	Optional() bool
	Constraints() []string
}

// ArgSchemas represents a collection of named arguments with collection-level metadata.
type ArgSchemas interface {
	Accepter

	Args() []ArgSchema
	AllowAdditional() bool
	AdditionalSchema() Schema
	CollectionName() string
	CollectionDescription() string
}

// FunctionSchema interface for function schemas with introspection methods.
type FunctionSchema interface {
	Schema
	Accepter

	// Introspection methods
	Inputs() ArgSchemas
	Outputs() ArgSchemas
	Errors() Schema
	RequiredInputs() []string
	RequiredOutputs() []string
}

type ServiceMethodSchema interface {
	Schema
	Accepter

	Name() string
	Function() FunctionSchema
}

// ServiceSchema interface for service schemas with introspection methods.
type ServiceSchema interface {
	Schema
	Accepter

	// Introspection methods
	Name() string
	Methods() []ServiceMethodSchema
}

// UnionSchema interface for union schemas with introspection methods.
type UnionSchema interface {
	Schema
	Accepter

	// Introspection methods
	Schemas() []Schema
}
