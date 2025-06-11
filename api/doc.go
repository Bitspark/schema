// Package api defines the core interfaces and contracts for the schema system.
//
// This package contains only interfaces, types, and constants - no implementations.
// It serves as the API layer that defines contracts between different parts of the
// schema system, enabling clean separation of concerns and better testability.
//
// # Architecture
//
// The schema system follows a layered architecture:
//
//   - schema/api - Core interfaces and contracts (this package)
//   - schema - Concrete implementations of the interfaces
//   - schema/functions - Function-specific implementations
//   - schema/registry - Registry implementations
//   - schema/generator - Code generation using the interfaces
//
// # Core Interfaces
//
// The main interfaces defined in this package are:
//
//   - Schema: Core interface that all schema types implement
//   - SchemaVisitor/Accepter: Visitor pattern for schema traversal
//   - Function/TypedFunction: Function definition and execution
//   - Registry: Function registration and management
//   - Various Builder interfaces: Fluent APIs for schema construction
//
// # Usage
//
// This package is typically imported alongside the main schema package:
//
//	import (
//		"defs.dev/schema"
//		"defs.dev/schema/api"
//	)
//
//	// Use concrete implementations from schema package
//	stringSchema := schema.NewString().MinLength(5).Build()
//
//	// Use interfaces from api package for function parameters
//	func processSchema(s api.Schema) {
//		result := s.Validate(someValue)
//		// ...
//	}
//
// # Design Principles
//
// 1. Interface First: All contracts are defined as interfaces
// 2. Separation of Concerns: Interfaces separate from implementations
// 3. Composability: Small, focused interfaces that can be composed
// 4. Type Safety: Generic interfaces where appropriate
// 5. Extensibility: Easy to add new schema types and behaviors
//
// # Visitor Pattern
//
// The package supports the visitor pattern for schema traversal:
//
//	type MyVisitor struct {
//		api.BaseVisitor // Hypothetical base implementation
//	}
//
//	func (v *MyVisitor) VisitString(s api.StringSchema) error {
//		// Process string schema
//		return nil
//	}
//
//	// Walk through a schema tree
//	schema.Accept(myVisitor)
//
// # Function System
//
// The function system provides interfaces for:
//
//   - Function definition and metadata
//   - Type-safe function calls
//   - Function registration and discovery
//   - Middleware and processing pipelines
//
// Example:
//
//	func createFunction() api.Function {
//		return schema.NewFunctionSchema().
//			Input("name", schema.NewString().Build()).
//			Output(schema.NewString().Build()).
//			Build()
//	}
package api
