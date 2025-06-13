package core

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
	VisitService(ServiceSchema) error
	VisitUnion(UnionSchema) error
}

// Accepter defines the interface for schemas that can accept visitors.
// This is the other half of the visitor pattern.
type Accepter interface {
	Accept(SchemaVisitor) error
}
