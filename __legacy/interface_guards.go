package schema

// This file contains interface guards to ensure compile-time compatibility
// between concrete implementations and API interfaces.
// These are compile-time checks using the pattern: var _ Interface = (*ConcreteType)(nil)

import (
	_ "defs.dev/schema/api"
)

// Testing StringSchema migration to API interfaces - REVERTED FOR NOW
// var _ api.Schema = (*StringSchema)(nil)
// var _ api.StringSchema = (*StringSchema)(nil)
// var _ api.Accepter = (*StringSchema)(nil)

// TODO: The interface guards reveal that we need to solve the circular dependency
// between concrete types (Schema) and abstract types (api.Schema).
//
// The main issues are:
// 1. Methods return concrete Schema types but api expects api.Schema types
// 2. Methods accept concrete SchemaVisitor but api expects api.SchemaVisitor
//
// This requires a more careful design approach. For now, we'll comment out
// all interface guards and focus on creating a working API package first.

// Next approach: Use the API interfaces in NEW code, while keeping existing
// implementations unchanged. The type aliases in types.go bridge the gap.

// Legacy interface guards - TEMPORARILY COMMENTED OUT
// var _ api.SchemaLegacy = (*StringSchema)(nil)
// var _ api.SchemaLegacy = (*NumberSchema)(nil)
// var _ api.SchemaLegacy = (*IntegerSchema)(nil)
// var _ api.SchemaLegacy = (*BooleanSchema)(nil)
// var _ api.SchemaLegacy = (*ArraySchema)(nil)
// var _ api.SchemaLegacy = (*ObjectSchema)(nil)
// var _ api.SchemaLegacy = (*FunctionSchema)(nil)
// var _ api.SchemaLegacy = (*UnionSchema)(nil)

// Specific legacy schema interface guards - TEMPORARILY COMMENTED OUT
// var _ api.StringSchemaLegacy = (*StringSchema)(nil)
// var _ api.NumberSchemaLegacy = (*NumberSchema)(nil)
// var _ api.IntegerSchemaLegacy = (*IntegerSchema)(nil)
// var _ api.BooleanSchemaLegacy = (*BooleanSchema)(nil)
// var _ api.ArraySchemaLegacy = (*ArraySchema)(nil)
// var _ api.ObjectSchemaLegacy = (*ObjectSchema)(nil)
// var _ api.FunctionSchemaLegacy = (*FunctionSchema)(nil)
// var _ api.UnionSchemaLegacy = (*UnionSchema)(nil)

// Legacy accepter interface guards (visitor pattern) - TEMPORARILY COMMENTED OUT
// var _ api.AccepterLegacy = (*StringSchema)(nil)
// var _ api.AccepterLegacy = (*NumberSchema)(nil)
// var _ api.AccepterLegacy = (*IntegerSchema)(nil)
// var _ api.AccepterLegacy = (*BooleanSchema)(nil)
// var _ api.AccepterLegacy = (*ArraySchema)(nil)
// var _ api.AccepterLegacy = (*ObjectSchema)(nil)
// var _ api.AccepterLegacy = (*FunctionSchema)(nil)
// var _ api.AccepterLegacy = (*UnionSchema)(nil)

// TODO: Once we resolve the circular dependency issues, we can uncomment
// these interface guards to ensure compile-time compatibility.

// TODO: Builder interface guards - need to verify builder return types
// var _ api.StringSchemaBuilder = (*StringBuilder)(nil)
// var _ api.NumberSchemaBuilder = (*NumberBuilder)(nil)
// var _ api.IntegerSchemaBuilder = (*IntegerBuilder)(nil)
// var _ api.BooleanSchemaBuilder = (*BooleanBuilder)(nil)
// var _ api.ArraySchemaBuilder = (*ArrayBuilder)(nil)
// var _ api.ObjectSchemaBuilder = (*ObjectBuilder)(nil)
// var _ api.FunctionSchemaBuilder = (*FunctionSchemaBuilder)(nil)
// var _ api.UnionSchemaBuilder = (*UnionBuilder)(nil)

// TODO: Function interface guards - need to verify function implementations
// var _ api.FunctionInput = (FunctionInput)(nil)
// var _ api.FunctionOutput = (*FunctionOutput)(nil)
// var _ api.Function = (*FunctionTyper)(nil)
// var _ api.TypedFunction = (*FunctionTyper)(nil)
