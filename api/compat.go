package api

// This file provides compatibility types and interfaces to bridge the gap
// between existing schema package implementations and the new API interfaces.

// SchemaLegacy represents the existing Schema interface from the main schema package.
// This allows gradual migration without breaking existing code.
type SchemaLegacy interface {
	// Validation
	Validate(value any) ValidationResult

	// JSON Schema generation
	ToJSONSchema() map[string]any

	// Metadata
	Type() SchemaType
	Metadata() SchemaMetadata
	WithMetadata(metadata SchemaMetadata) SchemaLegacy

	// Example generation
	GenerateExample() any

	// Utilities
	Clone() SchemaLegacy
}

// SchemaVisitorLegacy represents the existing SchemaVisitor interface from the main schema package.
type SchemaVisitorLegacy interface {
	VisitString(StringSchemaLegacy) error
	VisitNumber(NumberSchemaLegacy) error
	VisitInteger(IntegerSchemaLegacy) error
	VisitBoolean(BooleanSchemaLegacy) error
	VisitArray(ArraySchemaLegacy) error
	VisitObject(ObjectSchemaLegacy) error
	VisitFunction(FunctionSchemaLegacy) error
	VisitUnion(UnionSchemaLegacy) error
}

// AccepterLegacy represents the existing Accepter interface.
type AccepterLegacy interface {
	Accept(SchemaVisitorLegacy) error
}

// Legacy schema type interfaces that match existing implementations
type StringSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	MinLength() *int
	MaxLength() *int
	Pattern() string
	Format() string
	EnumValues() []string
	DefaultValue() *string
}

type NumberSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	Minimum() *float64
	Maximum() *float64
}

type IntegerSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	Minimum() *int64
	Maximum() *int64
}

type BooleanSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy
}

type ArraySchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	ItemSchema() SchemaLegacy
	MinItems() *int
	MaxItems() *int
	UniqueItemsRequired() bool
}

type ObjectSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	Properties() map[string]SchemaLegacy
	Required() []string
	AdditionalProperties() bool
}

type FunctionSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	Inputs() map[string]SchemaLegacy
	Outputs() SchemaLegacy
	Errors() SchemaLegacy
	Required() []string
}

type UnionSchemaLegacy interface {
	SchemaLegacy
	AccepterLegacy

	// Introspection methods
	Schemas() []SchemaLegacy
}
