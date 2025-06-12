package core

// ValueVisitor defines the visitor interface for value traversal.
// This enables the visitor pattern for processing different value types.
type ValueVisitor interface {
	VisitString(StringValue) error
	VisitNumber(NumberValue) error
	VisitInteger(IntegerValue) error
	VisitBoolean(BooleanValue) error
	VisitArray(ArrayValue[any]) error
	VisitObject(StructureValue[any]) error
	VisitMap(MapValue[any, any]) error
}

// ValueAccepter defines the interface for values that can accept visitors.
// This is the other half of the visitor pattern for values.
type ValueAccepter interface {
	AcceptValue(ValueVisitor) error
}
