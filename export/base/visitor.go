package base

import "defs.dev/schema/api/core"

// BaseVisitor provides default implementations for all SchemaVisitor methods.
// Concrete generators can embed this struct and override only the methods they need,
// following the composition pattern rather than inheritance.
type BaseVisitor struct {
	// GeneratorName is the name of the generator using this visitor
	GeneratorName string

	// Context stores generation context and state
	Context *GenerationContext
}

// NewBaseVisitor creates a new BaseVisitor with the given generator name.
func NewBaseVisitor(generatorName string) *BaseVisitor {
	return &BaseVisitor{
		GeneratorName: generatorName,
		Context:       NewGenerationContext(),
	}
}

// Default implementations return UnsupportedSchemaError to make it clear
// when a generator hasn't implemented support for a particular schema type.

// VisitString provides a default implementation for string schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitString(schema core.StringSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeString, "string schema not implemented")
}

// VisitNumber provides a default implementation for number schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitNumber(schema core.NumberSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeNumber, "number schema not implemented")
}

// VisitInteger provides a default implementation for integer schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitInteger(schema core.IntegerSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeInteger, "integer schema not implemented")
}

// VisitBoolean provides a default implementation for boolean schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitBoolean(schema core.BooleanSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeBoolean, "boolean schema not implemented")
}

// VisitArray provides a default implementation for array schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitArray(schema core.ArraySchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeArray, "array schema not implemented")
}

// VisitObject provides a default implementation for object schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitObject(schema core.ObjectSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeStructure, "object schema not implemented")
}

// VisitFunction provides a default implementation for function schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitFunction(schema core.FunctionSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeFunction, "function schema not implemented")
}

// VisitService provides a default implementation for service schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitService(schema core.ServiceSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeService, "service schema not implemented")
}

// VisitUnion provides a default implementation for union schema visitation.
// Concrete generators should override this method.
func (v *BaseVisitor) VisitUnion(schema core.UnionSchema) error {
	return NewUnsupportedSchemaError(v.GeneratorName, core.TypeUnion, "union schema not implemented")
}

// Helper methods for common visitor patterns

// VisitWithPath visits a schema with path tracking for better error reporting.
func (v *BaseVisitor) VisitWithPath(schema core.Schema, pathElement string) error {
	// Add path element
	v.Context.PushPath(pathElement)
	defer v.Context.PopPath()

	// Use the schema's Accept method to dispatch to the correct visitor method
	if accepter, ok := schema.(core.Accepter); ok {
		return accepter.Accept(v)
	}

	// Fallback for schemas that don't implement Accepter
	return NewGenerationError(v.GeneratorName, string(schema.Type()), "schema does not implement Accepter interface")
}

// VisitNested visits a nested schema and handles any generation errors by adding path context.
func (v *BaseVisitor) VisitNested(schema core.Schema, pathElement string) error {
	err := v.VisitWithPath(schema, pathElement)
	if err != nil {
		// If it's already a GenerationError, add path context
		if genErr, ok := err.(*GenerationError); ok {
			return genErr.AppendPath(pathElement)
		}
		// Otherwise, wrap it as a GenerationError with path
		return NewGenerationErrorWithCause(v.GeneratorName, string(schema.Type()), "nested generation failed", err).WithPath(pathElement)
	}
	return nil
}

// SetContext sets the generation context for this visitor.
func (v *BaseVisitor) SetContext(ctx *GenerationContext) {
	v.Context = ctx
}

// GetContext returns the current generation context.
func (v *BaseVisitor) GetContext() *GenerationContext {
	return v.Context
}

// WithContext creates a new visitor with the given context.
func (v *BaseVisitor) WithContext(ctx *GenerationContext) *BaseVisitor {
	newVisitor := *v
	newVisitor.Context = ctx
	return &newVisitor
}

// NoOpVisitor is a visitor that does nothing for all schema types.
// This is useful for testing or as a base for visitors that only need
// to implement a few methods.
type NoOpVisitor struct {
	*BaseVisitor
}

// NewNoOpVisitor creates a new NoOpVisitor.
func NewNoOpVisitor(generatorName string) *NoOpVisitor {
	return &NoOpVisitor{
		BaseVisitor: NewBaseVisitor(generatorName),
	}
}

// Override all visit methods to do nothing and return nil

func (v *NoOpVisitor) VisitString(schema core.StringSchema) error     { return nil }
func (v *NoOpVisitor) VisitNumber(schema core.NumberSchema) error     { return nil }
func (v *NoOpVisitor) VisitInteger(schema core.IntegerSchema) error   { return nil }
func (v *NoOpVisitor) VisitBoolean(schema core.BooleanSchema) error   { return nil }
func (v *NoOpVisitor) VisitArray(schema core.ArraySchema) error       { return nil }
func (v *NoOpVisitor) VisitObject(schema core.ObjectSchema) error     { return nil }
func (v *NoOpVisitor) VisitFunction(schema core.FunctionSchema) error { return nil }
func (v *NoOpVisitor) VisitService(schema core.ServiceSchema) error   { return nil }
func (v *NoOpVisitor) VisitUnion(schema core.UnionSchema) error       { return nil }

// CountingVisitor counts the number of schemas of each type visited.
// This is useful for analysis and testing.
type CountingVisitor struct {
	*BaseVisitor
	Counts map[core.SchemaType]int
}

// NewCountingVisitor creates a new CountingVisitor.
func NewCountingVisitor(generatorName string) *CountingVisitor {
	return &CountingVisitor{
		BaseVisitor: NewBaseVisitor(generatorName),
		Counts:      make(map[core.SchemaType]int),
	}
}

// Helper method to increment count for a schema type
func (v *CountingVisitor) count(schemaType core.SchemaType) {
	v.Counts[schemaType]++
}

// Override visit methods to count occurrences

func (v *CountingVisitor) VisitString(schema core.StringSchema) error {
	v.count(core.TypeString)
	return nil
}

func (v *CountingVisitor) VisitNumber(schema core.NumberSchema) error {
	v.count(core.TypeNumber)
	return nil
}

func (v *CountingVisitor) VisitInteger(schema core.IntegerSchema) error {
	v.count(core.TypeInteger)
	return nil
}

func (v *CountingVisitor) VisitBoolean(schema core.BooleanSchema) error {
	v.count(core.TypeBoolean)
	return nil
}

func (v *CountingVisitor) VisitArray(schema core.ArraySchema) error {
	v.count(core.TypeArray)
	return nil
}

func (v *CountingVisitor) VisitObject(schema core.ObjectSchema) error {
	v.count(core.TypeStructure)
	return nil
}

func (v *CountingVisitor) VisitFunction(schema core.FunctionSchema) error {
	v.count(core.TypeFunction)
	return nil
}

func (v *CountingVisitor) VisitService(schema core.ServiceSchema) error {
	v.count(core.TypeService)
	return nil
}

func (v *CountingVisitor) VisitUnion(schema core.UnionSchema) error {
	v.count(core.TypeUnion)
	return nil
}

// GetCount returns the count for a specific schema type.
func (v *CountingVisitor) GetCount(schemaType core.SchemaType) int {
	return v.Counts[schemaType]
}

// TotalCount returns the total number of schemas visited.
func (v *CountingVisitor) TotalCount() int {
	total := 0
	for _, count := range v.Counts {
		total += count
	}
	return total
}

// Reset clears all counts.
func (v *CountingVisitor) Reset() {
	v.Counts = make(map[core.SchemaType]int)
}
