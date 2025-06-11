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

// NewNumber creates a new number schema builder.
func NewNumber() api.NumberSchemaBuilder {
	return builders.NewNumber()
}

// NewInteger creates a new integer schema builder.
func NewInteger() api.IntegerSchemaBuilder {
	return builders.NewInteger()
}

// NewBoolean creates a new boolean schema builder.
func NewBoolean() api.BooleanSchemaBuilder {
	return builders.NewBoolean()
}

// NewArray creates a new array schema builder.
func NewArray() api.ArraySchemaBuilder {
	return builders.NewArray()
}

// NewObject creates a new object schema builder.
func NewObject() api.ObjectSchemaBuilder {
	return builders.NewObject()
}

// TODO: Add other schema type factory functions as we implement them
// func NewUnion() api.UnionSchemaBuilder { return builders.NewUnion() }
// func NewFunction() api.FunctionSchemaBuilder { return builders.NewFunction() }
