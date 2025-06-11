package schema

import "fmt"

// SchemaVisitor defines the visitor interface for schema traversal
type SchemaVisitor interface {
	VisitString(*StringSchema) error
	VisitNumber(*NumberSchema) error
	VisitInteger(*IntegerSchema) error
	VisitBoolean(*BooleanSchema) error
	VisitArray(*ArraySchema) error
	VisitObject(*ObjectSchema) error
	VisitFunction(*FunctionSchema) error
	VisitUnion(*UnionSchema) error
}

// Accept method for each schema type to support visitor pattern
func (s *StringSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitString(s)
}

func (s *NumberSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitNumber(s)
}

func (s *IntegerSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitInteger(s)
}

func (s *BooleanSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitBoolean(s)
}

func (s *ArraySchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitArray(s)
}

func (s *ObjectSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitObject(s)
}

func (s *FunctionSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitFunction(s)
}

func (s *UnionSchema) Accept(visitor SchemaVisitor) error {
	return visitor.VisitUnion(s)
}

// Accepter interface - schemas that can accept visitors
type Accepter interface {
	Accept(SchemaVisitor) error
}

// BaseVisitor provides default implementations for all visitor methods
type BaseVisitor struct{}

func (v *BaseVisitor) VisitString(*StringSchema) error     { return nil }
func (v *BaseVisitor) VisitNumber(*NumberSchema) error     { return nil }
func (v *BaseVisitor) VisitInteger(*IntegerSchema) error   { return nil }
func (v *BaseVisitor) VisitBoolean(*BooleanSchema) error   { return nil }
func (v *BaseVisitor) VisitArray(*ArraySchema) error       { return nil }
func (v *BaseVisitor) VisitObject(*ObjectSchema) error     { return nil }
func (v *BaseVisitor) VisitFunction(*FunctionSchema) error { return nil }
func (v *BaseVisitor) VisitUnion(*UnionSchema) error       { return nil }

// TraversalVisitor recursively visits all schemas in a tree
// This demonstrates how visitor + introspection methods work together
type TraversalVisitor struct {
	BaseVisitor
	handler func(Schema) error
}

func NewTraversalVisitor(handler func(Schema) error) *TraversalVisitor {
	return &TraversalVisitor{handler: handler}
}

func (v *TraversalVisitor) VisitArray(schema *ArraySchema) error {
	// Call handler for this schema
	if err := v.handler(schema); err != nil {
		return err
	}

	// Use introspection method to get item schema and traverse it
	if itemSchema := schema.ItemSchema(); itemSchema != nil {
		if accepter, ok := itemSchema.(Accepter); ok {
			return accepter.Accept(v)
		}
	}
	return nil
}

func (v *TraversalVisitor) VisitObject(schema *ObjectSchema) error {
	// Call handler for this schema
	if err := v.handler(schema); err != nil {
		return err
	}

	// Use introspection method to get properties and traverse them
	for _, propSchema := range schema.Properties() {
		if accepter, ok := propSchema.(Accepter); ok {
			if err := accepter.Accept(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *TraversalVisitor) VisitFunction(schema *FunctionSchema) error {
	// Call handler for this schema
	if err := v.handler(schema); err != nil {
		return err
	}

	// Use introspection methods to traverse inputs and outputs
	for _, inputSchema := range schema.Inputs() {
		if accepter, ok := inputSchema.(Accepter); ok {
			if err := accepter.Accept(v); err != nil {
				return err
			}
		}
	}

	if outputs := schema.Outputs(); outputs != nil {
		if accepter, ok := outputs.(Accepter); ok {
			if err := accepter.Accept(v); err != nil {
				return err
			}
		}
	}

	if errors := schema.Errors(); errors != nil {
		if accepter, ok := errors.(Accepter); ok {
			if err := accepter.Accept(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *TraversalVisitor) VisitUnion(schema *UnionSchema) error {
	// Call handler for this schema
	if err := v.handler(schema); err != nil {
		return err
	}

	// Use introspection method to get union schemas and traverse them
	for _, unionSchema := range schema.Schemas() {
		if accepter, ok := unionSchema.(Accepter); ok {
			if err := accepter.Accept(v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Handle primitive types by calling handler
func (v *TraversalVisitor) VisitString(schema *StringSchema) error {
	return v.handler(schema)
}

func (v *TraversalVisitor) VisitNumber(schema *NumberSchema) error {
	return v.handler(schema)
}

func (v *TraversalVisitor) VisitInteger(schema *IntegerSchema) error {
	return v.handler(schema)
}

func (v *TraversalVisitor) VisitBoolean(schema *BooleanSchema) error {
	return v.handler(schema)
}

// Convenience function to walk a schema tree
func Walk(schema Schema, handler func(Schema) error) error {
	visitor := NewTraversalVisitor(handler)
	if accepter, ok := schema.(Accepter); ok {
		return accepter.Accept(visitor)
	}
	return handler(schema)
}

// Example visitors that use introspection methods

// StringCollectorVisitor collects all string schemas with their metadata
type StringCollectorVisitor struct {
	BaseVisitor
	Strings []*StringSchema
}

func (v *StringCollectorVisitor) VisitString(schema *StringSchema) error {
	v.Strings = append(v.Strings, schema)
	return nil
}

// RequiredFieldAnalyzer finds all required fields in object schemas
type RequiredFieldAnalyzer struct {
	BaseVisitor
	RequiredFields map[string][]string // schema name -> required fields
}

func NewRequiredFieldAnalyzer() *RequiredFieldAnalyzer {
	return &RequiredFieldAnalyzer{
		RequiredFields: make(map[string][]string),
	}
}

func (v *RequiredFieldAnalyzer) VisitObject(schema *ObjectSchema) error {
	name := schema.Metadata().Name
	if name == "" {
		name = fmt.Sprintf("unnamed_object_%p", schema)
	}

	// Use introspection method to get required fields
	v.RequiredFields[name] = schema.Required()
	return nil
}

// SchemaStatisticsVisitor collects statistics about a schema tree
type SchemaStatisticsVisitor struct {
	BaseVisitor
	Stats SchemaStats
}

type SchemaStats struct {
	StringCount   int
	NumberCount   int
	IntegerCount  int
	BooleanCount  int
	ArrayCount    int
	ObjectCount   int
	FunctionCount int
	UnionCount    int

	ObjectProperties int // Total properties across all objects
	RequiredFields   int // Total required fields
	ArraysWithItems  int // Arrays that have item constraints
}

func (v *SchemaStatisticsVisitor) VisitString(*StringSchema) error {
	v.Stats.StringCount++
	return nil
}

func (v *SchemaStatisticsVisitor) VisitNumber(*NumberSchema) error {
	v.Stats.NumberCount++
	return nil
}

func (v *SchemaStatisticsVisitor) VisitInteger(*IntegerSchema) error {
	v.Stats.IntegerCount++
	return nil
}

func (v *SchemaStatisticsVisitor) VisitBoolean(*BooleanSchema) error {
	v.Stats.BooleanCount++
	return nil
}

func (v *SchemaStatisticsVisitor) VisitArray(schema *ArraySchema) error {
	v.Stats.ArrayCount++

	// Use introspection methods to gather more detailed stats
	if schema.ItemSchema() != nil {
		v.Stats.ArraysWithItems++
	}
	return nil
}

func (v *SchemaStatisticsVisitor) VisitObject(schema *ObjectSchema) error {
	v.Stats.ObjectCount++

	// Use introspection methods to gather more detailed stats
	v.Stats.ObjectProperties += len(schema.Properties())
	v.Stats.RequiredFields += len(schema.Required())
	return nil
}

func (v *SchemaStatisticsVisitor) VisitFunction(schema *FunctionSchema) error {
	v.Stats.FunctionCount++

	// Could use introspection methods to analyze function complexity
	// v.Stats.FunctionInputCount += len(schema.Inputs())
	return nil
}

func (v *SchemaStatisticsVisitor) VisitUnion(schema *UnionSchema) error {
	v.Stats.UnionCount++
	return nil
}
