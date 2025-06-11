package functions

// Factory functions for creating different types of registries
//
// Note: Specific portal implementations are now in their own packages:
// - defs.dev/schema/functions/local
// - defs.dev/schema/functions/http
// - defs.dev/schema/functions/db
// - etc.
//
// Import the specific package and use its NewRegistry() function:
//   import "defs.dev/schema/functions/local"
//   registry := local.NewRegistry()
