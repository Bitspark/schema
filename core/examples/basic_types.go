package examples

import (
	"fmt"
	"log"

	"defs.dev/schema/core"
)

// ExampleBasicTypes demonstrates the usage of Number, Integer, and Boolean schemas
func ExampleBasicTypes() {
	fmt.Println("=== Schema Core - Basic Types Examples ===")

	// Number Schema Examples
	demonstrateNumberSchema()

	// Integer Schema Examples
	demonstrateIntegerSchema()

	// Boolean Schema Examples
	demonstrateBooleanSchema()
}

func demonstrateNumberSchema() {
	fmt.Println("📊 Number Schema Examples")
	fmt.Println("------------------------")

	// Basic number schema
	basicNumber := core.NewNumber().
		Description("Any numeric value").
		Example(42.5).
		Build()

	// Test validation
	fmt.Printf("Validating 42.5: %v\n", basicNumber.Validate(42.5).Valid)
	fmt.Printf("Validating 'text': %v\n", basicNumber.Validate("text").Valid)

	// Range-constrained number
	percentage := core.NewNumber().
		Percentage(). // Built-in helper for 0-100 range
		Build()

	fmt.Printf("Percentage validation (50.0): %v\n", percentage.Validate(50.0).Valid)
	fmt.Printf("Percentage validation (150.0): %v\n", percentage.Validate(150.0).Valid)

	// Custom range with helper methods
	positiveNumber := core.NewNumber().
		Positive().
		Description("Must be positive").
		Build()

	fmt.Printf("Positive number validation (5.5): %v\n", positiveNumber.Validate(5.5).Valid)
	fmt.Printf("Positive number validation (-1.0): %v\n", positiveNumber.Validate(-1.0).Valid)

	// JSON Schema generation
	jsonSchema := percentage.ToJSONSchema()
	fmt.Printf("Percentage JSON Schema: %+v\n", jsonSchema)

	fmt.Println()
}

func demonstrateIntegerSchema() {
	fmt.Println("🔢 Integer Schema Examples")
	fmt.Println("--------------------------")

	// Basic integer schema
	basicInteger := core.NewInteger().
		Description("Any integer value").
		Example(int64(42)).
		Build()

	// Test different integer types
	fmt.Printf("Validating int(42): %v\n", basicInteger.Validate(42).Valid)
	fmt.Printf("Validating int64(100): %v\n", basicInteger.Validate(int64(100)).Valid)
	fmt.Printf("Validating float64(42.0): %v\n", basicInteger.Validate(42.0).Valid) // Whole number
	fmt.Printf("Validating float64(42.5): %v\n", basicInteger.Validate(42.5).Valid) // Decimal

	// Port number validation using helper
	portSchema := core.NewInteger().
		Port(). // Built-in helper for 1-65535 range
		Build()

	fmt.Printf("Port validation (8080): %v\n", portSchema.Validate(8080).Valid)
	fmt.Printf("Port validation (0): %v\n", portSchema.Validate(0).Valid)
	fmt.Printf("Port validation (70000): %v\n", portSchema.Validate(70000).Valid)

	// ID schema with custom constraints
	idSchema := core.NewInteger().
		ID(). // Positive integer helper
		Description("Unique identifier").
		Build()

	fmt.Printf("ID validation (123): %v\n", idSchema.Validate(123).Valid)
	fmt.Printf("ID validation (-1): %v\n", idSchema.Validate(-1).Valid)

	// Age validation
	ageSchema := core.NewInteger().
		Age(). // 0-150 range helper
		Build()

	fmt.Printf("Age validation (25): %v\n", ageSchema.Validate(25).Valid)
	fmt.Printf("Age validation (200): %v\n", ageSchema.Validate(200).Valid)

	// Custom range
	customRange := core.NewInteger().
		Range(10, 20).
		Description("Custom range 10-20").
		Build()

	fmt.Printf("Custom range validation (15): %v\n", customRange.Validate(15).Valid)
	fmt.Printf("Custom range validation (25): %v\n", customRange.Validate(25).Valid)

	fmt.Println()
}

func demonstrateBooleanSchema() {
	fmt.Println("✅ Boolean Schema Examples")
	fmt.Println("--------------------------")

	// Basic boolean schema
	basicBool := core.NewBoolean().
		Description("Simple true/false value").
		Example(true).
		Build()

	fmt.Printf("Validating true: %v\n", basicBool.Validate(true).Valid)
	fmt.Printf("Validating false: %v\n", basicBool.Validate(false).Valid)
	fmt.Printf("Validating 'true' string: %v\n", basicBool.Validate("true").Valid) // Should fail

	// Boolean with string conversion
	flexibleBool := core.NewBoolean().
		AllowStringConversion().
		Description("Accepts boolean or string values").
		Build()

	fmt.Printf("Flexible bool - true: %v\n", flexibleBool.Validate(true).Valid)
	fmt.Printf("Flexible bool - 'true': %v\n", flexibleBool.Validate("true").Valid)
	fmt.Printf("Flexible bool - '1': %v\n", flexibleBool.Validate("1").Valid)
	fmt.Printf("Flexible bool - 'false': %v\n", flexibleBool.Validate("false").Valid)
	fmt.Printf("Flexible bool - '0': %v\n", flexibleBool.Validate("0").Valid)
	fmt.Printf("Flexible bool - 'maybe': %v\n", flexibleBool.Validate("maybe").Valid) // Should fail

	// Case-insensitive boolean (includes string conversion)
	caseInsensitive := core.NewBoolean().
		CaseInsensitive().
		Description("Case-insensitive string conversion").
		Build()

	fmt.Printf("Case insensitive - 'TRUE': %v\n", caseInsensitive.Validate("TRUE").Valid)
	fmt.Printf("Case insensitive - 'False': %v\n", caseInsensitive.Validate("False").Valid)
	fmt.Printf("Case insensitive - 'YES': %v\n", caseInsensitive.Validate("YES").Valid)
	fmt.Printf("Case insensitive - 'no': %v\n", caseInsensitive.Validate("no").Valid)
	fmt.Printf("Case insensitive - 'ON': %v\n", caseInsensitive.Validate("ON").Valid)
	fmt.Printf("Case insensitive - 'off': %v\n", caseInsensitive.Validate("off").Valid)

	// Using helper methods
	flagSchema := core.NewBoolean().
		Flag(). // Defaults to false
		Build()

	enabledSchema := core.NewBoolean().
		Enabled(). // Defaults to false, description "Whether this feature is enabled"
		Build()

	switchSchema := core.NewBoolean().
		Switch(). // Case-insensitive string conversion enabled
		Build()

	fmt.Printf("Flag schema default: %v\n", flagSchema.GenerateExample())
	fmt.Printf("Enabled schema default: %v\n", enabledSchema.GenerateExample())
	fmt.Printf("Switch validation ('TRUE'): %v\n", switchSchema.Validate("TRUE").Valid)

	// JSON Schema generation
	jsonSchema := switchSchema.ToJSONSchema()
	fmt.Printf("Switch JSON Schema: %+v\n", jsonSchema)

	fmt.Println()
}

// ExampleErrorHandling demonstrates error handling and validation details
func ExampleErrorHandling() {
	fmt.Println("🚨 Error Handling Examples")
	fmt.Println("--------------------------")

	// Number schema with constraints
	rangedNumber := core.NewNumber().
		Range(0.0, 100.0).
		Description("Value between 0 and 100").
		Build()

	// Test invalid value
	result := rangedNumber.Validate(-10.0)
	if !result.Valid {
		fmt.Printf("Validation failed for -10.0:\n")
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s (Code: %s)\n", err.Message, err.Code)
			fmt.Printf("  Expected: %s\n", err.Expected)
			fmt.Printf("  Suggestion: %s\n", err.Suggestion)
		}
	}

	// Integer overflow handling
	intSchema := core.NewInteger().Build()

	// This would cause overflow in a naive implementation
	result = intSchema.Validate(uint64(18446744073709551615)) // MaxUint64
	if !result.Valid {
		fmt.Printf("\nOverflow protection for MaxUint64:\n")
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s (Code: %s)\n", err.Message, err.Code)
		}
	}

	// Boolean string conversion error
	boolSchema := core.NewBoolean().AllowStringConversion().Build()
	result = boolSchema.Validate("invalid")
	if !result.Valid {
		fmt.Printf("\nBoolean conversion error for 'invalid':\n")
		for _, err := range result.Errors {
			fmt.Printf("  Error: %s (Code: %s)\n", err.Message, err.Code)
			fmt.Printf("  Expected: %s\n", err.Expected)
		}
	}

	fmt.Println()
}

// ExampleComplexValidation shows more complex validation scenarios
func ExampleComplexValidation() {
	fmt.Println("🔍 Complex Validation Examples")
	fmt.Println("------------------------------")

	// Chained builder methods
	priceSchema := core.NewNumber().
		NonNegative().
		Description("Product price in USD").
		Name("price").
		Tag("currency").
		Tag("required").
		Example(29.99).
		Default(0.0).
		Build()

	fmt.Printf("Price schema metadata: %+v\n", priceSchema.Metadata())
	fmt.Printf("Price validation (29.99): %v\n", priceSchema.Validate(29.99).Valid)

	// Complex integer validation
	userIdSchema := core.NewInteger().
		ID().
		Name("user_id").
		Description("Unique user identifier").
		Tag("database").
		Tag("primary_key").
		Example(int64(12345)).
		Build()

	fmt.Printf("User ID JSON Schema: %+v\n", userIdSchema.ToJSONSchema())

	// Boolean with metadata
	featureFlagSchema := core.NewBoolean().
		Switch().
		Name("feature_enabled").
		Description("Enable experimental feature").
		Tag("feature_flag").
		Tag("experimental").
		Default(false).
		Build()

	fmt.Printf("Feature flag validation ('true'): %v\n", featureFlagSchema.Validate("true").Valid)
	fmt.Printf("Feature flag default: %v\n", featureFlagSchema.GenerateExample())

	fmt.Println()
}

// RunAllExamples runs all the examples
func RunAllExamples() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Example panicked: %v", r)
		}
	}()

	ExampleBasicTypes()
	ExampleErrorHandling()
	ExampleComplexValidation()
	RunArrayExamples()

	fmt.Println("✨ All examples completed successfully!")
}
