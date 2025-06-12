// Package export provides schema generation using the visitor pattern.
//
// This package separates schema generation concerns from schema validation by implementing
// various generators as visitors that traverse schema structures. This design enables:
//   - Clean separation of validation and generation logic
//   - Easy addition of new output formats without modifying schemas
//   - Consistent generation patterns across all formats
//   - Independent testing of generation logic
//
// Architecture:
//
// The export package uses the visitor pattern where:
//   - Schemas implement core.Accepter and can accept visitors
//   - Generators implement core.SchemaVisitor and visit different schema types
//   - Each generator produces output in its specific format
//
// Supported Formats:
//
//   - JSON Schema: Standard JSON Schema Draft 7/2020-12 output
//   - TypeScript: TypeScript type definitions and interfaces
//   - Python: Python type hints (dataclass, TypedDict, Pydantic)
//   - Go: Go struct definitions with JSON tags
//   - OpenAPI: OpenAPI 3.x specification components
//
// Usage:
//
//	import "defs.dev/schema/export"
//
//	// Generate JSON Schema
//	generator := export.NewJSONSchemaGenerator()
//	output, err := generator.Generate(schema)
//
//	// Generate TypeScript
//	tsGen := export.NewTypeScriptGenerator(
//		export.WithTypeScriptStyle("interface"),
//		export.WithStrictNullChecks(true),
//	)
//	tsOutput, err := tsGen.Generate(schema)
//
//	// Generate multiple formats
//	registry := export.NewGeneratorRegistry()
//	registry.Register("json", export.NewJSONSchemaGenerator())
//	registry.Register("typescript", tsGen)
//
//	outputs, err := registry.GenerateAll(schema)
//
// Generator Options:
//
// Each generator supports configuration through functional options:
//
//	jsonGen := export.NewJSONSchemaGenerator(
//		export.WithJSONSchemaDraft("draft-2020-12"),
//		export.WithIncludeExamples(true),
//	)
//
// Custom Generators:
//
// You can create custom generators by implementing the Generator interface:
//
//	type MyGenerator struct {
//		*base.BaseVisitor
//		output strings.Builder
//	}
//
//	func (g *MyGenerator) VisitString(s core.StringSchema) error {
//		g.output.WriteString("custom string representation")
//		return nil
//	}
//
//	func (g *MyGenerator) Generate(schema core.Schema) ([]byte, error) {
//		if err := schema.Accept(g); err != nil {
//			return nil, err
//		}
//		return []byte(g.output.String()), nil
//	}
//
// Error Handling:
//
// Generators return detailed errors with context about what failed during generation:
//
//	output, err := generator.Generate(schema)
//	if err != nil {
//		if genErr, ok := err.(*base.GenerationError); ok {
//			fmt.Printf("Failed to generate %s: %s", genErr.SchemaType, genErr.Message)
//		}
//	}
//
// Performance:
//
// Generators are designed to be reusable and efficient:
//   - Generators can be reused across multiple schemas
//   - Internal caches optimize repeated generation patterns
//   - Memory usage is minimized through streaming generation where possible
package export
