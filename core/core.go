package core

import (
	"defs.dev/schema/api"
	"defs.dev/schema/core/builders"
)

// Factory functions for creating schema builders
// These provide the main entry points for the core package

// NewString creates a new string schema builder.
func NewString() api.StringSchemaBuilder {
	return builders.NewString()
}

// TODO: Add other schema type factory functions as we implement them
// func NewNumber() api.NumberSchemaBuilder { return builders.NewNumber() }
// func NewInteger() api.IntegerSchemaBuilder { return builders.NewInteger() }
// func NewBoolean() api.BooleanSchemaBuilder { return builders.NewBoolean() }
// func NewArray() api.ArraySchemaBuilder { return builders.NewArray() }
// func NewObject() api.ObjectSchemaBuilder { return builders.NewObject() }
// func NewUnion() api.UnionSchemaBuilder { return builders.NewUnion() }
// func NewFunction() api.FunctionSchemaBuilder { return builders.NewFunction() }
