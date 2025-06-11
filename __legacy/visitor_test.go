package schema

import (
	"fmt"
	"strings"
	"testing"
)

func TestVisitorPattern(t *testing.T) {
	// Create a complex schema to test with
	apiSchema := NewObject().
		Name("APIResponse").
		Property("user", NewObject().
			Name("User").
			Property("id", NewInteger().Min(1).Build()).
			Property("name", NewString().MinLength(1).Build()).
			Property("email", NewString().Email().Build()).
			Property("roles", NewArray().Items(NewString().Enum("admin", "user", "guest").Build()).Build()).
			Required("id", "name", "email").
			Build()).
		Property("metadata", NewObject().
			Name("Metadata").
			Property("timestamp", NewString().Build()).
			Property("version", NewString().Build()).
			Build()).
		Property("data", Union2[string, int]().Build()).
		Required("user").
		Build()

	t.Run("StringCollectorVisitor", func(t *testing.T) {
		visitor := &StringCollectorVisitor{}

		// Walk the entire schema tree and collect all string schemas
		err := Walk(apiSchema, func(schema Schema) error {
			if accepter, ok := schema.(Accepter); ok {
				return accepter.Accept(visitor)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Should find multiple string schemas: name, email, roles items, timestamp, version, union member
		if len(visitor.Strings) < 5 {
			t.Errorf("Expected at least 5 string schemas, got %d", len(visitor.Strings))
		}

		t.Logf("Found %d string schemas", len(visitor.Strings))
	})

	t.Run("RequiredFieldAnalyzer", func(t *testing.T) {
		analyzer := NewRequiredFieldAnalyzer()

		// Use traversal visitor that uses introspection methods
		traversal := NewTraversalVisitor(func(schema Schema) error {
			if accepter, ok := schema.(Accepter); ok {
				return accepter.Accept(analyzer)
			}
			return nil
		})

		err := apiSchema.(Accepter).Accept(traversal)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		// Check that we found required fields
		if len(analyzer.RequiredFields) == 0 {
			t.Error("Expected to find required fields")
		}

		// Log what we found
		for schemaName, required := range analyzer.RequiredFields {
			t.Logf("Schema '%s' has required fields: %v", schemaName, required)
		}

		// APIResponse should have "user" as required
		apiResponseRequired := analyzer.RequiredFields["APIResponse"]
		hasUser := false
		for _, field := range apiResponseRequired {
			if field == "user" {
				hasUser = true
				break
			}
		}
		if !hasUser {
			t.Error("Expected APIResponse to have 'user' as required field")
		}
	})

	t.Run("SchemaStatisticsVisitor", func(t *testing.T) {
		statsVisitor := &SchemaStatisticsVisitor{}

		// Walk the entire schema and collect statistics
		err := Walk(apiSchema, func(schema Schema) error {
			if accepter, ok := schema.(Accepter); ok {
				return accepter.Accept(statsVisitor)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Statistics collection failed: %v", err)
		}

		stats := statsVisitor.Stats

		// Log statistics
		t.Logf("Schema Statistics:")
		t.Logf("  Strings: %d", stats.StringCount)
		t.Logf("  Integers: %d", stats.IntegerCount)
		t.Logf("  Objects: %d", stats.ObjectCount)
		t.Logf("  Arrays: %d", stats.ArrayCount)
		t.Logf("  Unions: %d", stats.UnionCount)
		t.Logf("  Total Properties: %d", stats.ObjectProperties)
		t.Logf("  Required Fields: %d", stats.RequiredFields)

		// Verify we found the expected types
		if stats.ObjectCount < 3 {
			t.Errorf("Expected at least 3 objects, got %d", stats.ObjectCount)
		}
		if stats.StringCount < 5 {
			t.Errorf("Expected at least 5 strings, got %d", stats.StringCount)
		}
		if stats.UnionCount < 1 {
			t.Errorf("Expected at least 1 union, got %d", stats.UnionCount)
		}
	})

	t.Run("Direct Introspection vs Visitor", func(t *testing.T) {
		// Show how both approaches can be used for the same goal

		// Approach 1: Direct introspection (simple, specific)
		obj := apiSchema.(*ObjectSchema)
		userProperty := obj.Properties()["user"]
		userObj := userProperty.(*ObjectSchema)
		directRequired := userObj.Required()

		// Approach 2: Visitor pattern (comprehensive, flexible)
		analyzer := NewRequiredFieldAnalyzer()
		err := Walk(apiSchema, func(schema Schema) error {
			if accepter, ok := schema.(Accepter); ok {
				return accepter.Accept(analyzer)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Visitor analysis failed: %v", err)
		}

		visitorRequired := analyzer.RequiredFields["User"]

		// Both should find the same required fields for User object
		if len(directRequired) != len(visitorRequired) {
			t.Errorf("Direct introspection found %d required fields, visitor found %d",
				len(directRequired), len(visitorRequired))
		}

		t.Logf("Direct introspection found: %v", directRequired)
		t.Logf("Visitor pattern found: %v", visitorRequired)
	})
}

func TestFunctionSchemaVisitor(t *testing.T) {
	// Create a function schema with complex structure
	calcSchema := NewFunctionSchema().
		Name("calculator").
		Description("Performs mathematical operations").
		Input("operands", NewArray().Items(NewNumber().Build()).MinItems(2).Build()).
		Input("operation", NewString().Enum("add", "subtract", "multiply", "divide").Build()).
		Input("precision", NewInteger().Min(0).Max(10).Build()).
		Output(NewObject().
			Property("result", NewNumber().Build()).
			Property("precision", NewInteger().Build()).
			Required("result").
			Build()).
		Required("operands", "operation").
		Build()

	t.Run("FunctionAnalysisVisitor", func(t *testing.T) {
		// Custom visitor that analyzes function schemas using introspection
		type FunctionAnalysis struct {
			Name            string
			InputCount      int
			HasOutput       bool
			RequiredInputs  []string
			ComplexityScore int
		}

		var functions []FunctionAnalysis

		// Create visitor that analyzes functions
		visitor := NewTraversalVisitor(func(schema Schema) error {
			if funcSchema, ok := schema.(*FunctionSchema); ok {
				analysis := FunctionAnalysis{
					Name:           funcSchema.Metadata().Name,
					InputCount:     len(funcSchema.Inputs()),
					HasOutput:      funcSchema.Outputs() != nil,
					RequiredInputs: funcSchema.Required(),
				}

				// Calculate complexity based on schema structure
				complexity := len(funcSchema.Inputs()) * 2
				if funcSchema.Outputs() != nil {
					complexity += 3
				}
				if funcSchema.Errors() != nil {
					complexity += 2
				}
				analysis.ComplexityScore = complexity

				functions = append(functions, analysis)
			}
			return nil
		})

		// Apply visitor
		err := calcSchema.(Accepter).Accept(visitor)
		if err != nil {
			t.Fatalf("Function analysis failed: %v", err)
		}

		if len(functions) != 1 {
			t.Fatalf("Expected 1 function analysis, got %d", len(functions))
		}

		analysis := functions[0]
		t.Logf("Function Analysis:")
		t.Logf("  Name: %s", analysis.Name)
		t.Logf("  Input Count: %d", analysis.InputCount)
		t.Logf("  Has Output: %v", analysis.HasOutput)
		t.Logf("  Required Inputs: %v", analysis.RequiredInputs)
		t.Logf("  Complexity Score: %d", analysis.ComplexityScore)

		if analysis.InputCount != 3 {
			t.Errorf("Expected 3 inputs, got %d", analysis.InputCount)
		}
		if !analysis.HasOutput {
			t.Error("Expected function to have output")
		}
	})
}

func ExampleWalk() {
	// Create a simple schema
	userSchema := NewObject().
		Property("name", NewString().MinLength(1).Build()).
		Required("name").
		Build()

	// Walk the schema and print each node
	count := 0
	err := Walk(userSchema, func(schema Schema) error {
		count++
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Visited %d schemas\n", count)

	// Output:
	// Visited 2 schemas
}

func ExampleTraversalVisitor() {
	// Example of using visitor + introspection to transform schemas

	var modified []string

	// Create visitor that tracks string schema modifications
	enforcer := NewTraversalVisitor(func(schema Schema) error {
		if _, ok := schema.(*StringSchema); ok {
			// In a real implementation, you might create a new schema with enforced validation
			modified = append(modified, "String schema would be modified")
		}
		return nil
	})

	schema := NewObject().
		Property("title", NewString().Build()).
		Property("description", NewString().Build()).
		Build()

	err := Walk(schema, func(s Schema) error {
		if accepter, ok := s.(Accepter); ok {
			return accepter.Accept(enforcer)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Modified %d string schemas\n", len(modified))
	// Output: Modified 4 string schemas
}

// Demonstrate the combination of both patterns
func ExampleWalk_combined() {
	schema := NewObject().
		Property("user", NewObject().
			Property("name", NewString().Build()).
			Property("age", NewInteger().Build()).
			Required("name").
			Build()).
		Property("tags", NewArray().Items(NewString().Build()).Build()).
		Required("user").
		Build()

	fmt.Println("=== Direct Introspection ===")
	// Direct introspection for specific access
	obj := schema.(*ObjectSchema)
	fmt.Printf("Root object has %d properties\n", len(obj.Properties()))

	userProp := obj.Properties()["user"].(*ObjectSchema)
	fmt.Printf("User object requires: %s\n", strings.Join(userProp.Required(), ", "))

	fmt.Println("\n=== Visitor Pattern ===")
	// Visitor for comprehensive analysis
	count := 0
	err := Walk(schema, func(s Schema) error {
		count++
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Total schemas in tree: %d\n", count)

	// Output:
	// === Direct Introspection ===
	// Root object has 2 properties
	// User object requires: name
	//
	// === Visitor Pattern ===
	// Total schemas in tree: 6
}
