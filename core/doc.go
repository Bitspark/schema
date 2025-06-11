// Package core provides a clean, API-first implementation of the schema system.
//
// This package is a complete re-implementation using the interfaces defined in
// schema/api, providing better organization, performance, and extensibility.
//
// Key features:
//   - API-first design using schema/api interfaces
//   - Clean separation of concerns
//   - Immutable schema objects
//   - Type-safe generic patterns
//   - Enhanced validation and error reporting
//   - Comprehensive visitor pattern support
//
// Usage:
//
//	import "defs.dev/schema/core"
//
//	// Create a string schema
//	schema := core.NewString().
//		MinLength(3).
//		MaxLength(50).
//		Pattern(`^[a-zA-Z0-9_]+$`).
//		Build()
//
//	// Validate a value
//	result := schema.Validate("john_doe")
//	if !result.Valid {
//		// Handle validation errors
//	}
//
// The core package is designed to eventually replace the legacy schema package
// while maintaining full compatibility through the shared API interfaces.
package core
