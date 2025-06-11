package registry_test

import (
	"fmt"
	"log"

	"defs.dev/schema"
	"defs.dev/schema/registry"
)

func ExampleRegistry_usage() {
	// Create a new registry
	reg := registry.New()

	// Define some basic schemas
	reg.Define("User", schema.Object().
		Property("id", schema.Integer().Min(1).Build()).
		Property("name", schema.String().MinLength(1).Build()).
		Property("email", schema.String().Email().Build()).
		Required("id", "name", "email").
		Build())

	reg.Define("Product", schema.Object().
		Property("id", schema.Integer().Min(1).Build()).
		Property("name", schema.String().MinLength(1).Build()).
		Property("price", schema.Number().Min(0).Build()).
		Required("id", "name", "price").
		Build())

	// Define parameterized schemas
	reg.Define("List", schema.Array().
		Items(registry.Param("T")).
		Build(), "T")

	// Create a simple null schema
	nullSchema := &schema.StringSchema{} // Placeholder - we'll create a proper null schema
	nullSchema = nullSchema.WithMetadata(schema.SchemaMetadata{Name: "null"}).(*schema.StringSchema)

	reg.Define("Optional", schema.Object().
		Property("value", registry.Param("T")).
		Property("hasValue", schema.Boolean().Build()).
		Required("hasValue").
		Build(), "T")

	reg.Define("ApiResponse", schema.Object().
		Property("success", schema.Boolean().Build()).
		Property("data", registry.Param("T")).
		Property("message", schema.String().Build()).
		Required("success").
		Build(), "T")

	// Use basic schemas
	user, _ := reg.Get("User")
	fmt.Printf("User schema type: %s\n", user.Type())

	// Use parameterized schemas
	userList, err := reg.Build("List").WithParam("T", reg.Ref("User")).Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User list schema type: %s\n", userList.Type())

	// Complex nested example
	userListForResponse, err := reg.Build("List").WithParam("T", reg.Ref("User")).Build()
	if err != nil {
		log.Fatal(err)
	}
	userResponse, err := reg.Build("ApiResponse").WithParam("T", userListForResponse).Build()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User response schema type: %s\n", userResponse.Type())

	// Validation example
	validUser := map[string]any{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	userRef := reg.Ref("User")
	result := userRef.Validate(validUser)
	fmt.Printf("User validation: %t\n", result.Valid)

	// List all schemas
	schemas := reg.List()
	fmt.Printf("Total schemas: %d\n", len(schemas))

	// Output:
	// User schema type: object
	// User list schema type: array
	// User response schema type: object
	// User validation: true
	// Total schemas: 5
}

func ExampleRegistry_realWorldUsage() {
	reg := registry.New()

	// Define parameterized schemas first
	reg.Define("List", schema.Array().
		Items(registry.Param("T")).
		Build(), "T")

	// Define business domain schemas
	reg.Define("Address", schema.Object().
		Property("street", schema.String().MinLength(1).Build()).
		Property("city", schema.String().MinLength(1).Build()).
		Property("zipCode", schema.String().Pattern("^[0-9]{5}$").Build()).
		Required("street", "city", "zipCode").
		Build())

	reg.Define("Customer", schema.Object().
		Property("id", schema.Integer().Min(1).Build()).
		Property("name", schema.String().MinLength(1).Build()).
		Property("email", schema.String().Email().Build()).
		Property("address", reg.Ref("Address")).
		Required("id", "name", "email").
		Build())

	// Define generic utility schemas
	reg.Define("Page", schema.Object().
		Property("items", registry.Param("T")).
		Property("totalCount", schema.Integer().Min(0).Build()).
		Property("pageSize", schema.Integer().Min(1).Build()).
		Property("currentPage", schema.Integer().Min(1).Build()).
		Required("items", "totalCount", "pageSize", "currentPage").
		Build(), "T")

	// Create a paginated customer list
	customerList, err := reg.Build("List").WithParam("T", reg.Ref("Customer")).Build()
	if err != nil {
		log.Fatal(err)
	}
	customerPage, err := reg.Build("Page").WithParam("T", customerList).Build()

	if err != nil {
		log.Fatal(err)
	}

	// Validate some data
	sampleCustomer := map[string]any{
		"id":    123,
		"name":  "Alice Johnson",
		"email": "alice@example.com",
		"address": map[string]any{
			"street":  "123 Main St",
			"city":    "Springfield",
			"zipCode": "12345",
		},
	}

	result := reg.Ref("Customer").Validate(sampleCustomer)
	fmt.Printf("Customer validation: %t\n", result.Valid)

	samplePage := map[string]any{
		"items":       []any{sampleCustomer},
		"totalCount":  1,
		"pageSize":    10,
		"currentPage": 1,
	}

	result = customerPage.Validate(samplePage)
	fmt.Printf("Customer page validation: %t\n", result.Valid)

	// Show registry capabilities
	fmt.Printf("Customer has parameters: %v\n", reg.Parameters("Customer"))
	fmt.Printf("Page has parameters: %v\n", reg.Parameters("Page"))
	fmt.Printf("Registry contains %d schemas\n", len(reg.List()))

	// Output:
	// Customer validation: true
	// Customer page validation: false
	// Customer has parameters: []
	// Page has parameters: [T]
	// Registry contains 4 schemas
}
