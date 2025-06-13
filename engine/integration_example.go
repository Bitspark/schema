package engine

import (
	builders2 "defs.dev/schema/builders"
	"fmt"
)

// IntegrationExample demonstrates how to use the Schema Engine
// alongside existing schema builders
func IntegrationExample() {
	fmt.Println("=== Schema Engine Integration Example ===")

	// 1. Create Schema Engine
	fmt.Println("1. Creating Schema Engine with custom configuration...")
	config := EngineConfig{
		EnableCache:        true,
		MaxCacheSize:       500,
		CircularDepthLimit: 25,
		StrictMode:         false,
		ValidateOnRegister: false,
		EnableConcurrency:  true,
	}
	engine := NewSchemaEngineWithConfig(config)
	fmt.Printf("✓ Engine created with cache size: %d\n", config.MaxCacheSize)

	// 2. Define Schemas with Builders
	fmt.Println("\n2. Defining schemas with builders...")

	// User entity schema
	userSchema := builders2.NewObjectSchema().
		Name("User").
		Description("User entity with authentication and profile information").
		Property("id", builders2.NewIntegerSchema().
			Description("Unique user identifier").
			Min(1).
			Build()).
		Property("username", builders2.NewStringSchema().
			Description("Unique username for authentication").
			MinLength(3).
			MaxLength(50).
			Pattern("^[a-zA-Z0-9_]+$").
			Build()).
		Property("email", builders2.NewStringSchema().
			Description("User email address").
			Format("email").
			Build()).
		Property("profile", builders2.NewObjectSchema().
			Property("firstName", builders2.NewStringSchema().Build()).
			Property("lastName", builders2.NewStringSchema().Build()).
			Property("bio", builders2.NewStringSchema().MaxLength(500).Build()).
			Required("firstName", "lastName").
			Build()).
		Property("roles", builders2.NewArraySchema().
			Items(builders2.NewStringSchema().
				Enum("user", "admin", "moderator").
				Build()).
			Build()).
		Property("active", builders2.NewBooleanSchema().
			Default(true).
			Build()).
		Property("createdAt", builders2.NewStringSchema().
			Format("date-time").
			Build()).
		Required("id", "username", "email", "profile", "active", "createdAt").
		Build()

	// Order management schema
	orderSchema := builders2.NewObjectSchema().
		Name("Order").
		Description("E-commerce order with items and billing information").
		Property("id", builders2.NewIntegerSchema().
			Description("Unique order identifier").
			Min(1).
			Build()).
		Property("userId", builders2.NewIntegerSchema().
			Description("Reference to the user who placed the order").
			Min(1).
			Build()).
		Property("items", builders2.NewArraySchema().
			Description("List of ordered items").
			Items(builders2.NewObjectSchema().
				Property("productId", builders2.NewIntegerSchema().Min(1).Build()).
				Property("name", builders2.NewStringSchema().Build()).
				Property("quantity", builders2.NewIntegerSchema().Min(1).Build()).
				Property("price", builders2.NewNumberSchema().Min(0).Build()).
				Required("productId", "name", "quantity", "price").
				Build()).
			MinItems(1).
			Build()).
		Property("total", builders2.NewNumberSchema().
			Description("Total order amount").
			Min(0).
			Build()).
		Property("status", builders2.NewStringSchema().
			Description("Order processing status").
			Enum("pending", "processing", "shipped", "delivered", "cancelled").
			Default("pending").
			Build()).
		Property("shippingAddress", builders2.NewObjectSchema().
			Property("street", builders2.NewStringSchema().Build()).
			Property("city", builders2.NewStringSchema().Build()).
			Property("state", builders2.NewStringSchema().Build()).
			Property("zipCode", builders2.NewStringSchema().Build()).
			Property("country", builders2.NewStringSchema().Build()).
			Required("street", "city", "state", "zipCode", "country").
			Build()).
		Property("orderDate", builders2.NewStringSchema().
			Format("date-time").
			Build()).
		Required("id", "userId", "items", "total", "status", "orderDate").
		Build()

	fmt.Printf("✓ User schema defined with %d properties\n", len(userSchema.Properties()))
	fmt.Printf("✓ Order schema defined with %d properties\n", len(orderSchema.Properties()))

	// 3. Register Schemas in Engine
	fmt.Println("\n3. Registering schemas in engine...")

	if err := engine.RegisterSchema("User", userSchema); err != nil {
		fmt.Printf("✗ Failed to register User schema: %v\n", err)
		return
	}

	if err := engine.RegisterSchema("Order", orderSchema); err != nil {
		fmt.Printf("✗ Failed to register Order schema: %v\n", err)
		return
	}

	fmt.Printf("✓ Registered %d schemas in engine\n", len(engine.ListSchemas()))

	// 4. Create Function Definitions
	fmt.Println("\n4. Creating function definitions...")

	// User service functions
	getUserFunction := builders2.NewFunctionSchema().
		Name("getUser").
		Description("Retrieve user by ID").
		Input("userId", builders2.NewIntegerSchema().Min(1).Build()).
		Output("user", userSchema). // Reference to registered schema
		Build()

	createUserFunction := builders2.NewFunctionSchema().
		Name("createUser").
		Description("Create a new user").
		Input("userData", builders2.NewObjectSchema().
			Property("username", builders2.NewStringSchema().Build()).
			Property("email", builders2.NewStringSchema().Build()).
			Property("profile", builders2.NewObjectSchema().
				Property("firstName", builders2.NewStringSchema().Build()).
				Property("lastName", builders2.NewStringSchema().Build()).
				Build()).
			Required("username", "email", "profile").
			Build()).
		Output("user", userSchema).
		Build()

	fmt.Printf("✓ getUserFunction defined with %d inputs, %d outputs\n",
		len(getUserFunction.Inputs().Args()), len(getUserFunction.Outputs().Args()))
	fmt.Printf("✓ createUserFunction defined with %d inputs, %d outputs\n",
		len(createUserFunction.Inputs().Args()), len(createUserFunction.Outputs().Args()))

	// 5. Register Functions
	fmt.Println("\n5. Registering functions...")

	if err := engine.RegisterSchema("getUser", getUserFunction); err != nil {
		fmt.Printf("✗ Failed to register getUser function: %v\n", err)
		return
	}

	if err := engine.RegisterSchema("createUser", createUserFunction); err != nil {
		fmt.Printf("✗ Failed to register createUser function: %v\n", err)
		return
	}

	fmt.Printf("✓ Total schemas registered: %d\n", len(engine.ListSchemas()))

	// 6. Schema Resolution and References
	fmt.Println("\n6. Testing schema resolution...")

	// Resolve by name
	resolvedUser, err := engine.ResolveSchema("User")
	if err != nil {
		fmt.Printf("✗ Failed to resolve User schema: %v\n", err)
		return
	}
	fmt.Printf("✓ Resolved User schema: %s\n", resolvedUser.Metadata().Name)

	// Use references
	userRef := Ref("User")
	fmt.Printf("✓ Created reference: %s\n", userRef.FullName())

	resolvedByRef, err := engine.ResolveReference(userRef)
	if err != nil {
		fmt.Printf("✗ Failed to resolve by reference: %v\n", err)
		return
	}
	fmt.Printf("✓ Resolved by reference: %s\n", resolvedByRef.Metadata().Name)

	// 7. Annotations for Schema Metadata
	fmt.Println("\n7. Applying annotations...")

	// Validate built-in annotations
	annotations := map[string]any{
		"pattern":  "entity",
		"behavior": []string{"persistent", "cached"},
		"performance": map[string]any{
			"timeout":    "30s",
			"rate_limit": 1000,
			"async":      false,
		},
		"security": map[string]any{
			"authentication": "required",
			"authorization":  []string{"user:read", "admin:write"},
			"encryption":     "optional",
			"audit":          true,
		},
	}

	for name, value := range annotations {
		if err := engine.ValidateAnnotation(name, value); err != nil {
			fmt.Printf("✗ Invalid annotation %s: %v\n", name, err)
		} else {
			fmt.Printf("✓ Valid annotation: %s\n", name)
		}
	}

	// 8. Engine Statistics
	fmt.Println("\n8. Engine statistics...")

	stats := map[string]any{
		"schemas":     len(engine.ListSchemas()),
		"annotations": len(engine.ListAnnotations()),
		"types":       len(engine.GetAvailableTypes()),
		"config":      engine.Config(),
	}

	fmt.Printf("✓ Engine statistics: %+v\n", stats)

	// 9. Clone for Different Environments
	fmt.Println("\n9. Environment-specific configurations...")

	// Development environment
	devEngine := engine.Clone()
	devConfig := devEngine.Config()
	devConfig.StrictMode = false
	devConfig.ValidateOnRegister = false
	devEngine = devEngine.WithConfig(devConfig)

	// Production environment
	prodEngine := engine.Clone()
	prodConfig := prodEngine.Config()
	prodConfig.StrictMode = true
	prodConfig.ValidateOnRegister = true
	prodEngine = prodEngine.WithConfig(prodConfig)

	fmt.Printf("✓ Development engine: strict=%v, validate=%v\n",
		devEngine.Config().StrictMode, devEngine.Config().ValidateOnRegister)
	fmt.Printf("✓ Production engine: strict=%v, validate=%v\n",
		prodEngine.Config().StrictMode, prodEngine.Config().ValidateOnRegister)

	fmt.Println("\n=== Integration Example Complete ===")
}
