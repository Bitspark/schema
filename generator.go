package schema

import (
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// GeneratorConfig configures the behavior of the random value generator
type GeneratorConfig struct {
	// MaxDepth limits how deeply nested objects/arrays can be (prevents infinite recursion)
	MaxDepth int

	// MaxItems limits the maximum number of items in arrays and objects
	MaxItems int

	// MinItems sets the minimum number of items in arrays and objects (if no schema constraint)
	MinItems int

	// StringLength controls generated string lengths
	StringLength struct {
		Min int // Default minimum string length
		Max int // Default maximum string length
	}

	// NumberRange controls generated number ranges
	NumberRange struct {
		Min float64 // Default minimum number value
		Max float64 // Default maximum number value
	}

	// IntegerRange controls generated integer ranges
	IntegerRange struct {
		Min int64 // Default minimum integer value
		Max int64 // Default maximum integer value
	}

	// OptionalProbability controls how often optional fields are included (0.0 = never, 1.0 = always)
	OptionalProbability float64

	// UnionChoice controls which option to pick for union types ("first", "random", "balanced")
	UnionChoice string

	// Seed for reproducible random generation (0 = use system time)
	Seed int64

	// PreferExamples when true, uses schema examples when available instead of generating random values
	PreferExamples bool

	// CustomGenerators allows overriding generation for specific schema types/names
	CustomGenerators map[string]func(schema Schema, config GeneratorConfig, depth int) any

	// FollowReferences when true, attempts to resolve schema references during generation
	FollowReferences bool

	// GenerateDefaults when true, generates default/empty values instead of random values
	GenerateDefaults bool

	// DefaultValues controls what default values to use for different types
	DefaultValues struct {
		String  string  // Default string value (empty string if not specified)
		Number  float64 // Default number value
		Integer int64   // Default integer value
		Boolean bool    // Default boolean value
	}

	// MinimalGeneration when true, generates minimal valid data (empty arrays, minimal objects)
	MinimalGeneration bool
}

// DefaultGeneratorConfig returns sensible defaults for the generator
func DefaultGeneratorConfig() GeneratorConfig {
	return GeneratorConfig{
		MaxDepth: 5,
		MaxItems: 10,
		MinItems: 1,
		StringLength: struct {
			Min int
			Max int
		}{Min: 3, Max: 20},
		NumberRange: struct {
			Min float64
			Max float64
		}{Min: 0, Max: 1000},
		IntegerRange: struct {
			Min int64
			Max int64
		}{Min: 0, Max: 1000},
		OptionalProbability: 0.7,
		UnionChoice:         "random",
		Seed:                0,
		PreferExamples:      true,
		CustomGenerators:    make(map[string]func(schema Schema, config GeneratorConfig, depth int) any),
		FollowReferences:    true,
		GenerateDefaults:    false,
		DefaultValues: struct {
			String  string
			Number  float64
			Integer int64
			Boolean bool
		}{
			String:  "",
			Number:  0.0,
			Integer: 0,
			Boolean: false,
		},
		MinimalGeneration: false,
	}
}

// Generator generates random values that conform to schemas
type Generator struct {
	config GeneratorConfig
	rng    *rand.Rand
}

// NewGenerator creates a new random value generator with the given configuration
func NewGenerator(config GeneratorConfig) *Generator {
	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &Generator{
		config: config,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// NewGeneratorWithDefaults creates a generator with default configuration
func NewGeneratorWithDefaults() *Generator {
	return NewGenerator(DefaultGeneratorConfig())
}

// NewDefaultValueGenerator creates a generator that produces default/empty values
func NewDefaultValueGenerator() *Generator {
	config := DefaultGeneratorConfig()
	config.GenerateDefaults = true
	config.MinimalGeneration = true
	config.OptionalProbability = 0.0 // Don't include optional fields by default
	config.MinItems = 0              // Allow empty arrays by default
	return NewGenerator(config)
}

// NewMinimalGenerator creates a generator that produces minimal valid data
func NewMinimalGenerator() *Generator {
	config := DefaultGeneratorConfig()
	config.MinimalGeneration = true
	config.OptionalProbability = 0.0 // Don't include optional fields
	config.MinItems = 0              // Allow empty arrays/objects
	return NewGenerator(config)
}

// Generate creates a random value that conforms to the given schema
func (g *Generator) Generate(schema Schema) any {
	return g.generateWithDepth(schema, 0)
}

// GenerateMany creates multiple random values that conform to the given schema
func (g *Generator) GenerateMany(schema Schema, count int) []any {
	results := make([]any, count)
	for i := 0; i < count; i++ {
		results[i] = g.Generate(schema)
	}
	return results
}

// generateWithDepth is the internal method that tracks recursion depth
func (g *Generator) generateWithDepth(schema Schema, depth int) any {
	// Prevent infinite recursion - generate a simple fallback value based on schema type
	if depth > g.config.MaxDepth {
		return g.generateFallbackValue(schema)
	}

	// Check for custom generators first
	if customGen, exists := g.config.CustomGenerators[schema.Metadata().Name]; exists {
		return customGen(schema, g.config, depth)
	}

	// Use examples if preferred and available
	if g.config.PreferExamples {
		examples := schema.Metadata().Examples
		if len(examples) > 0 {
			// Pick a random example
			return examples[g.rng.Intn(len(examples))]
		}
	}

	// Generate based on schema type
	switch schema.Type() {
	case TypeString:
		return g.generateString(schema.(*StringSchema))
	case TypeNumber:
		return g.generateNumber(schema.(*NumberSchema))
	case TypeInteger:
		return g.generateInteger(schema.(*IntegerSchema))
	case TypeBoolean:
		return g.generateBoolean(schema.(*BooleanSchema))
	case TypeObject:
		return g.generateObject(schema.(*ObjectSchema), depth)
	case TypeArray:
		return g.generateArray(schema.(*ArraySchema), depth)
	case TypeOptional:
		return g.generateOptional(schema, depth)
	case TypeResult:
		return g.generateResult(schema, depth)
	case TypeMap:
		return g.generateMap(schema, depth)
	case TypeUnion:
		return g.generateUnion(schema.(*UnionSchema), depth)
	case TypeNull:
		return nil
	case TypeAny:
		return g.generateAny(depth)
	default:
		// Fallback to GenerateExample method
		return schema.GenerateExample()
	}
}

// generateString creates a random string following string schema constraints
func (g *Generator) generateString(schema *StringSchema) any {
	// Generate default values if requested
	if g.config.GenerateDefaults {
		// Handle enum values first - use first enum value as default
		if len(schema.enumValues) > 0 {
			return schema.enumValues[0]
		}

		// For default generation, use the configured default string value
		defaultStr := g.config.DefaultValues.String

		// Ensure it meets minimum length requirements
		if schema.minLength != nil && len(defaultStr) < *schema.minLength {
			// Pad with 'a' characters to meet minimum length
			needed := *schema.minLength - len(defaultStr)
			defaultStr += strings.Repeat("a", needed)
		}

		// Ensure it doesn't exceed maximum length
		if schema.maxLength != nil && len(defaultStr) > *schema.maxLength {
			defaultStr = defaultStr[:*schema.maxLength]
		}

		return defaultStr
	}

	// Handle enum values first
	if len(schema.enumValues) > 0 {
		return schema.enumValues[g.rng.Intn(len(schema.enumValues))]
	}

	// Handle format-specific generation
	if schema.format != "" {
		return g.generateFormatString(schema.format)
	}

	// Handle pattern-based generation (simplified)
	if schema.pattern != "" {
		return g.generatePatternString(schema.pattern)
	}

	// Generate random string within length constraints
	minLen := g.config.StringLength.Min
	maxLen := g.config.StringLength.Max

	if schema.minLength != nil && *schema.minLength > minLen {
		minLen = *schema.minLength
	}
	if schema.maxLength != nil && *schema.maxLength < maxLen {
		maxLen = *schema.maxLength
	}

	// Ensure min <= max
	if minLen > maxLen {
		minLen = maxLen
	}

	length := minLen
	if maxLen > minLen {
		length = minLen + g.rng.Intn(maxLen-minLen+1)
	}

	return g.generateRandomString(length)
}

// generateNumber creates a random number following number schema constraints
func (g *Generator) generateNumber(schema *NumberSchema) any {
	// Generate default values if requested
	if g.config.GenerateDefaults {
		defaultNum := g.config.DefaultValues.Number

		// Ensure it meets constraints
		if schema.minimum != nil && defaultNum < *schema.minimum {
			return *schema.minimum
		}
		if schema.maximum != nil && defaultNum > *schema.maximum {
			return *schema.maximum
		}

		return defaultNum
	}

	min := g.config.NumberRange.Min
	max := g.config.NumberRange.Max

	if schema.minimum != nil && *schema.minimum > min {
		min = *schema.minimum
	}
	if schema.maximum != nil && *schema.maximum < max {
		max = *schema.maximum
	}

	// Ensure min <= max
	if min > max {
		min = max
	}

	return min + g.rng.Float64()*(max-min)
}

// generateInteger creates a random integer following integer schema constraints
func (g *Generator) generateInteger(schema *IntegerSchema) any {
	// Generate default values if requested
	if g.config.GenerateDefaults {
		defaultInt := g.config.DefaultValues.Integer

		// Ensure it meets constraints
		if schema.minimum != nil && defaultInt < *schema.minimum {
			return *schema.minimum
		}
		if schema.maximum != nil && defaultInt > *schema.maximum {
			return *schema.maximum
		}

		return defaultInt
	}

	min := g.config.IntegerRange.Min
	max := g.config.IntegerRange.Max

	if schema.minimum != nil && *schema.minimum > min {
		min = *schema.minimum
	}
	if schema.maximum != nil && *schema.maximum < max {
		max = *schema.maximum
	}

	// Ensure min <= max
	if min > max {
		min = max
	}

	return min + g.rng.Int63n(max-min+1)
}

// generateBoolean creates a random boolean
func (g *Generator) generateBoolean(schema *BooleanSchema) any {
	// Generate default values if requested
	if g.config.GenerateDefaults {
		return g.config.DefaultValues.Boolean
	}

	return g.rng.Float64() < 0.5
}

// generateObject creates a random object following object schema constraints
func (g *Generator) generateObject(schema *ObjectSchema, depth int) any {
	result := make(map[string]any)

	// Generate required properties first
	for _, required := range schema.required {
		if propSchema, exists := schema.properties[required]; exists {
			result[required] = g.generateWithDepth(propSchema, depth+1)
		} else {
			// If required property doesn't exist in properties, create a simple fallback
			result[required] = g.generateRandomString(5 + g.rng.Intn(10))
		}
	}

	// Generate optional properties based on probability (skip for minimal generation)
	if !g.config.MinimalGeneration {
		for propName, propSchema := range schema.properties {
			// Skip if already added as required
			if _, exists := result[propName]; exists {
				continue
			}

			// Decide whether to include this optional property
			if g.rng.Float64() < g.config.OptionalProbability {
				result[propName] = g.generateWithDepth(propSchema, depth+1)
			}
		}

		// Add additional properties if allowed (generate simple strings only)
		if schema.additionalProps && len(result) < g.config.MaxItems {
			additionalCount := g.rng.Intn(g.config.MaxItems - len(result) + 1)
			for i := 0; i < additionalCount; i++ {
				propName := "additional_" + strconv.Itoa(i)
				// Use simple string schema for additional properties to avoid constraint violations
				result[propName] = g.generateRandomString(5 + g.rng.Intn(10))
			}
		}
	}

	return result
}

// generateArray creates a random array following array schema constraints
func (g *Generator) generateArray(schema *ArraySchema, depth int) any {
	minItems := g.config.MinItems
	maxItems := g.config.MaxItems

	if schema.minItems != nil && *schema.minItems > minItems {
		minItems = *schema.minItems
	}
	if schema.maxItems != nil && *schema.maxItems < maxItems {
		maxItems = *schema.maxItems
	}

	// Ensure min <= max
	if minItems > maxItems {
		minItems = maxItems
	}

	// For minimal generation, use the smallest valid size
	count := minItems
	if !g.config.MinimalGeneration && maxItems > minItems {
		count = minItems + g.rng.Intn(maxItems-minItems+1)
	}

	result := make([]any, count)
	for i := 0; i < count; i++ {
		item := g.generateWithDepth(schema.itemSchema, depth+1)

		// Handle unique items constraint
		if schema.uniqueItems {
			maxAttempts := 50 // Prevent infinite loops
			attempts := 0
			for attempts < maxAttempts {
				isDuplicate := false
				for j := 0; j < i; j++ {
					if g.valuesEqual(result[j], item) {
						isDuplicate = true
						break
					}
				}
				if !isDuplicate {
					break
				}
				// Generate a different value by adding variation
				item = g.generateVariantValue(schema.itemSchema, depth+1, attempts)
				attempts++
			}
			// If still duplicate after many attempts, use a fallback unique value
			if attempts >= maxAttempts {
				item = g.generateUniqueItemValue(i)
			}
		}

		result[i] = item
	}

	return result
}

// generateOptional creates a random optional value
func (g *Generator) generateOptional(schema Schema, depth int) any {
	// Use reflection to get the actual optional schema
	if optSchema, ok := schema.(*OptionalSchema[any]); ok {
		// Decide whether to generate null or actual value
		if g.rng.Float64() < g.config.OptionalProbability {
			return g.generateWithDepth(optSchema.itemSchema, depth+1)
		}
		return nil
	}

	// Fallback: just use the example or nil
	if g.rng.Float64() < g.config.OptionalProbability {
		return schema.GenerateExample()
	}
	return nil
}

// generateResult creates a random result value (success or error)
func (g *Generator) generateResult(schema Schema, depth int) any {
	// Simple implementation: 70% success, 30% error
	if g.rng.Float64() < 0.7 {
		// Generate success value with simple types to avoid constraint violations
		return map[string]any{
			"success": true,
			"value":   g.generateRandomString(8),
		}
	} else {
		// Generate error value with simple types
		return map[string]any{
			"success": false,
			"error":   g.generateRandomString(8),
		}
	}
}

// generateMap creates a random map
func (g *Generator) generateMap(schema Schema, depth int) any {
	count := g.config.MinItems + g.rng.Intn(g.config.MaxItems-g.config.MinItems+1)
	result := make(map[string]any)

	for i := 0; i < count; i++ {
		key := "key_" + strconv.Itoa(i)
		// Use simple string values to avoid constraint violations
		value := g.generateRandomString(5 + g.rng.Intn(10))
		result[key] = value
	}

	return result
}

// generateUnion creates a random union value
func (g *Generator) generateUnion(schema *UnionSchema, depth int) any {
	if len(schema.schemas) == 0 {
		return nil
	}

	var chosenSchema Schema

	switch g.config.UnionChoice {
	case "first":
		chosenSchema = schema.schemas[0]
	case "balanced":
		// Try to balance between different types
		chosenSchema = schema.schemas[g.rng.Intn(len(schema.schemas))]
	default: // "random"
		chosenSchema = schema.schemas[g.rng.Intn(len(schema.schemas))]
	}

	// Generate value with fallback if the chosen schema fails
	value := g.generateWithDepth(chosenSchema, depth+1)

	// If the generated value is nil or empty for non-null schemas, try the first schema as fallback
	if value == nil && len(schema.schemas) > 0 && schema.schemas[0].Type() != TypeNull {
		value = g.generateWithDepth(schema.schemas[0], depth+1)
	}

	return value
}

// generateAny creates a random value of any type
func (g *Generator) generateAny(depth int) any {
	if depth > g.config.MaxDepth {
		// Return a simple fallback value instead of nil
		return "fallback_value"
	}

	types := []string{"string", "number", "boolean", "object", "array"}
	chosen := types[g.rng.Intn(len(types))]

	switch chosen {
	case "string":
		return g.generateRandomString(g.config.StringLength.Min + g.rng.Intn(g.config.StringLength.Max-g.config.StringLength.Min+1))
	case "number":
		return g.config.NumberRange.Min + g.rng.Float64()*(g.config.NumberRange.Max-g.config.NumberRange.Min)
	case "boolean":
		return g.rng.Float64() < 0.5
	case "object":
		count := 1 + g.rng.Intn(3) // 1-3 properties (reduced to avoid deep nesting)
		result := make(map[string]any)
		for i := 0; i < count; i++ {
			key := "prop_" + strconv.Itoa(i)
			// Use simple primitive values to avoid constraint violations and infinite recursion
			result[key] = g.generateSimplePrimitive()
		}
		return result
	case "array":
		count := 1 + g.rng.Intn(3) // 1-3 items (reduced to avoid deep nesting)
		result := make([]any, count)
		for i := 0; i < count; i++ {
			// Use simple primitive values to avoid constraint violations and infinite recursion
			result[i] = g.generateSimplePrimitive()
		}
		return result
	default:
		// Return a simple fallback instead of nil
		return "unknown_type"
	}
}

// Helper methods

// generateFormatString generates strings for specific formats
func (g *Generator) generateFormatString(format string) string {
	switch format {
	case "email":
		domains := []string{"example.com", "test.org", "demo.net"}
		users := []string{"user", "test", "admin", "demo"}
		return users[g.rng.Intn(len(users))] + "@" + domains[g.rng.Intn(len(domains))]
	case "uuid":
		return g.generateUUID()
	case "url":
		protocols := []string{"https", "http"}
		domains := []string{"example.com", "test.org", "api.demo.net"}
		return protocols[g.rng.Intn(len(protocols))] + "://" + domains[g.rng.Intn(len(domains))]
	case "date":
		year := 2020 + g.rng.Intn(5)
		month := 1 + g.rng.Intn(12)
		day := 1 + g.rng.Intn(28)
		return strconv.Itoa(year) + "-" + g.zeroPad(month) + "-" + g.zeroPad(day)
	case "time":
		hour := g.rng.Intn(24)
		minute := g.rng.Intn(60)
		second := g.rng.Intn(60)
		return g.zeroPad(hour) + ":" + g.zeroPad(minute) + ":" + g.zeroPad(second)
	case "date-time":
		return g.generateFormatString("date") + "T" + g.generateFormatString("time") + "Z"
	default:
		return g.generateRandomString(10)
	}
}

// generatePatternString attempts to generate a string matching a pattern (simplified)
func (g *Generator) generatePatternString(pattern string) string {
	// This is a simplified implementation - a full implementation would parse regex
	// For now, we'll generate some common patterns

	if strings.Contains(pattern, "[0-9]") {
		return g.generateNumericString(5)
	}
	if strings.Contains(pattern, "[a-z]") {
		return g.generateAlphaString(5, false)
	}
	if strings.Contains(pattern, "[A-Z]") {
		return g.generateAlphaString(5, true)
	}

	// Default fallback
	return g.generateRandomString(8)
}

// generateRandomString creates a random alphanumeric string
func (g *Generator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[g.rng.Intn(len(charset))]
	}
	return string(result)
}

// generateAlphaString creates a random alphabetic string
func (g *Generator) generateAlphaString(length int, uppercase bool) string {
	charset := "abcdefghijklmnopqrstuvwxyz"
	if uppercase {
		charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[g.rng.Intn(len(charset))]
	}
	return string(result)
}

// generateNumericString creates a random numeric string
func (g *Generator) generateNumericString(length int) string {
	const charset = "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[g.rng.Intn(len(charset))]
	}
	return string(result)
}

// generateUUID creates a random UUID-like string
func (g *Generator) generateUUID() string {
	return g.generateRandomString(8) + "-" +
		g.generateRandomString(4) + "-" +
		g.generateRandomString(4) + "-" +
		g.generateRandomString(4) + "-" +
		g.generateRandomString(12)
}

// zeroPad pads single digits with leading zeros
func (g *Generator) zeroPad(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	}
	return strconv.Itoa(num)
}

// generateFallbackValue generates a simple fallback value when depth limit is exceeded
func (g *Generator) generateFallbackValue(schema Schema) any {
	switch schema.Type() {
	case TypeString:
		return "fallback"
	case TypeNumber:
		return 0.0
	case TypeInteger:
		return int64(0)
	case TypeBoolean:
		return false
	case TypeObject:
		return make(map[string]any)
	case TypeArray:
		return []any{}
	case TypeNull:
		return nil
	default:
		return "fallback"
	}
}

// generateSimplePrimitive generates a simple primitive value (used to avoid infinite recursion and constraint violations)
func (g *Generator) generateSimplePrimitive() any {
	primitives := []string{"string", "number", "boolean"}
	chosen := primitives[g.rng.Intn(len(primitives))]

	switch chosen {
	case "string":
		return g.generateRandomString(3 + g.rng.Intn(8))
	case "number":
		return g.rng.Float64() * 100
	case "boolean":
		return g.rng.Float64() < 0.5
	default:
		return "fallback"
	}
}

// generateVariantValue generates a variant of a value for unique constraints
func (g *Generator) generateVariantValue(schema Schema, depth int, attempt int) any {
	// Add variation based on attempt number to increase uniqueness probability
	oldSeed := g.rng.Int63()
	g.rng.Seed(oldSeed + int64(attempt*1000)) // Change seed to get different values

	value := g.generateWithDepth(schema, depth)

	g.rng.Seed(oldSeed) // Restore original seed behavior
	return value
}

// generateUniqueItemValue generates a guaranteed unique value for array items
func (g *Generator) generateUniqueItemValue(index int) any {
	// Generate a simple unique value - prefer numbers for better compatibility
	return float64(1000 + index) // Simple unique numbers
}

// valuesEqual performs a safe equality check for complex types
func (g *Generator) valuesEqual(a, b any) bool {
	// Use reflect.DeepEqual for safe comparison of complex types
	return reflect.DeepEqual(a, b)
}

// Convenience functions

// Generate creates a random value using default configuration
func Generate(schema Schema) any {
	return NewGeneratorWithDefaults().Generate(schema)
}

// GenerateMany creates multiple random values using default configuration
func GenerateMany(schema Schema, count int) []any {
	return NewGeneratorWithDefaults().GenerateMany(schema, count)
}

// GenerateWithSeed creates a random value with a specific seed for reproducibility
func GenerateWithSeed(schema Schema, seed int64) any {
	config := DefaultGeneratorConfig()
	config.Seed = seed
	return NewGenerator(config).Generate(schema)
}

// GenerateDefaults creates default/empty values for the given schema
func GenerateDefaults(schema Schema) any {
	generator := NewDefaultValueGenerator()
	return generator.Generate(schema)
}

// GenerateMinimal creates minimal valid data for the given schema
func GenerateMinimal(schema Schema) any {
	generator := NewMinimalGenerator()
	return generator.Generate(schema)
}

// GenerateCustomDefaults creates values using custom default configuration
func GenerateCustomDefaults(schema Schema, defaultString string, defaultNumber float64, defaultInteger int64, defaultBoolean bool) any {
	config := DefaultGeneratorConfig()
	config.GenerateDefaults = true
	config.MinimalGeneration = true
	config.OptionalProbability = 0.0
	config.MinItems = 0
	config.DefaultValues.String = defaultString
	config.DefaultValues.Number = defaultNumber
	config.DefaultValues.Integer = defaultInteger
	config.DefaultValues.Boolean = defaultBoolean

	generator := NewGenerator(config)
	return generator.Generate(schema)
}
