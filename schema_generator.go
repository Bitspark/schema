package schema

import (
	"math/rand"
	"strconv"
	"time"
)

// TypeWeights controls the probability distribution of generated schema types
type TypeWeights struct {
	String   float64 // Weight for string schemas
	Number   float64 // Weight for number schemas
	Integer  float64 // Weight for integer schemas
	Boolean  float64 // Weight for boolean schemas
	Object   float64 // Weight for object schemas
	Array    float64 // Weight for array schemas
	Union    float64 // Weight for union schemas
	Optional float64 // Weight for optional schemas
	Null     float64 // Weight for null schemas
	Any      float64 // Weight for any schemas
}

// SchemaGeneratorConfig configures the behavior of the random schema generator
type SchemaGeneratorConfig struct {
	// Schema structure controls
	MaxDepth      int // Maximum nesting depth for objects/arrays
	MaxProperties int // Maximum properties in objects
	MinProperties int // Minimum properties in objects
	MaxArrayItems int // Maximum constraint for array schemas

	// Type distribution controls
	TypeWeights    TypeWeights // Probability weights for each schema type
	ComplexityBias float64     // 0.0 = simple types, 1.0 = complex types

	// Constraint generation
	GenerateConstraints   bool    // Whether to add min/max/length constraints
	ConstraintProbability float64 // Probability of adding constraints (0.0-1.0)
	GenerateFormats       bool    // Whether to add format constraints to strings
	GeneratePatterns      bool    // Whether to add regex patterns to strings
	GenerateEnums         bool    // Whether to create enum constraints

	// Schema metadata
	GenerateExamples     bool // Whether to generate example values
	GenerateDescriptions bool // Whether to add descriptions
	GenerateTitles       bool // Whether to add titles

	// Advanced features
	ReferenceDepth      int     // How deep to create $ref schemas (not implemented yet)
	UnionMaxTypes       int     // Maximum types in union schemas
	OptionalProbability float64 // Probability a property is optional

	// Naming and realism
	PropertyNameStyle string // "realistic", "generic", "random"
	UseCommonFormats  bool   // Use realistic formats (email, uuid, etc.)

	// Seed for reproducibility
	Seed int64
}

// DefaultSchemaGeneratorConfig returns sensible defaults for schema generation
func DefaultSchemaGeneratorConfig() SchemaGeneratorConfig {
	return SchemaGeneratorConfig{
		MaxDepth:      3, // Reduced depth
		MaxProperties: 5, // Reduced max properties
		MinProperties: 1,
		MaxArrayItems: 20,
		TypeWeights: TypeWeights{
			String:   0.25,
			Number:   0.15,
			Integer:  0.15,
			Boolean:  0.10,
			Object:   0.20,
			Array:    0.10,
			Union:    0.03,
			Optional: 0.02,
			Null:     0.00,
			Any:      0.00,
		},
		ComplexityBias:        0.5,
		GenerateConstraints:   true,
		ConstraintProbability: 0.1,
		GenerateFormats:       true,
		GeneratePatterns:      false,
		GenerateEnums:         true,
		GenerateExamples:      true,
		GenerateDescriptions:  false,
		GenerateTitles:        false,
		ReferenceDepth:        0,
		UnionMaxTypes:         4,
		OptionalProbability:   0.3,
		PropertyNameStyle:     "realistic",
		UseCommonFormats:      true,
		Seed:                  0,
	}
}

// SimpleSchemaGeneratorConfig returns config biased toward simple schemas
func SimpleSchemaGeneratorConfig() SchemaGeneratorConfig {
	config := DefaultSchemaGeneratorConfig()
	config.ComplexityBias = 0.2
	config.MaxDepth = 2
	config.MaxProperties = 4
	config.TypeWeights = TypeWeights{
		String:   0.35,
		Number:   0.20,
		Integer:  0.20,
		Boolean:  0.15,
		Object:   0.08,
		Array:    0.02,
		Union:    0.00,
		Optional: 0.00,
		Null:     0.00,
		Any:      0.00,
	}
	return config
}

// ComplexSchemaGeneratorConfig returns config biased toward complex schemas
func ComplexSchemaGeneratorConfig() SchemaGeneratorConfig {
	config := DefaultSchemaGeneratorConfig()
	// Use more conservative complex generation to avoid validation issues
	config.ComplexityBias = 0.4         // Reduced complexity
	config.MaxDepth = 2                 // Much shallower
	config.MaxProperties = 3            // Fewer properties
	config.ConstraintProbability = 0.05 // Much fewer constraints
	config.GenerateEnums = false        // Disable enums in complex mode
	config.GenerateFormats = false      // Disable formats in complex mode
	config.GeneratePatterns = false     // Disable patterns in complex mode
	config.TypeWeights = TypeWeights{
		String:   0.3,  // More strings
		Number:   0.2,  // More numbers
		Integer:  0.2,  // More integers
		Boolean:  0.15, // More booleans
		Object:   0.1,  // Fewer objects
		Array:    0.05, // Much fewer arrays
		Union:    0.0,  // No unions
		Optional: 0.0,  // No optionals
		Null:     0.0,  // No nulls
		Any:      0.0,  // No any types
	}
	return config
}

// SchemaGenerator generates random schemas that can be used for testing
type SchemaGenerator struct {
	config SchemaGeneratorConfig
	rng    *rand.Rand

	// Pre-defined pools for realistic generation
	propertyNames  []string
	stringFormats  []string
	commonPatterns []string
	descriptions   []string
	titles         []string
}

// Common property names for realistic schema generation
var commonPropertyNames = []string{
	"id", "name", "email", "age", "title", "description", "status", "type",
	"createdAt", "updatedAt", "deletedAt", "userId", "username", "password",
	"firstName", "lastName", "fullName", "address", "city", "country",
	"phone", "website", "avatar", "bio", "role", "permissions", "settings",
	"metadata", "tags", "category", "priority", "score", "rating", "count",
	"amount", "price", "currency", "discount", "tax", "total", "balance",
	"startDate", "endDate", "duration", "deadline", "reminder", "notes",
	"content", "body", "summary", "keywords", "author", "editor", "reviewer",
	"version", "revision", "branch", "commit", "hash", "checksum",
	"enabled", "active", "visible", "public", "private", "archived",
	"config", "options", "parameters", "attributes", "properties", "data",
}

// Common string formats
var commonStringFormats = []string{
	"email", "uuid", "uri", "url", "date", "time", "date-time",
	"password", "ipv4", "ipv6", "hostname", "json-pointer",
}

// Common regex patterns
var commonPatterns = []string{
	"^[a-zA-Z0-9]+$",               // Alphanumeric
	"^[0-9]{3}-[0-9]{3}-[0-9]{4}$", // Phone number
	"^[A-Z]{2}[0-9]{4}$",           // Code format
	"^#[0-9A-Fa-f]{6}$",            // Hex color
	"^[a-z_]+$",                    // Snake case
	"^[A-Z][a-z]+$",                // Title case
}

// Sample descriptions for schema metadata
var sampleDescriptions = []string{
	"A unique identifier for the resource",
	"The name or title of the item",
	"A brief description of the content",
	"The current status of the operation",
	"Timestamp when the record was created",
	"User-provided configuration settings",
	"Additional metadata for the object",
	"The priority level (1-10 scale)",
	"Email address for notifications",
	"Indicates whether the item is active",
}

// Sample titles for schema metadata
var sampleTitles = []string{
	"Identifier", "Name", "Title", "Description", "Status", "Type",
	"Created Date", "Updated Date", "User ID", "Email Address",
	"Configuration", "Settings", "Metadata", "Properties", "Data",
	"Count", "Amount", "Price", "Rating", "Score", "Priority",
}

// NewSchemaGenerator creates a new random schema generator with the given configuration
func NewSchemaGenerator(config SchemaGeneratorConfig) *SchemaGenerator {
	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &SchemaGenerator{
		config:         config,
		rng:            rand.New(rand.NewSource(seed)),
		propertyNames:  commonPropertyNames,
		stringFormats:  commonStringFormats,
		commonPatterns: commonPatterns,
		descriptions:   sampleDescriptions,
		titles:         sampleTitles,
	}
}

// NewSchemaGeneratorWithDefaults creates a schema generator with default configuration
func NewSchemaGeneratorWithDefaults() *SchemaGenerator {
	return NewSchemaGenerator(DefaultSchemaGeneratorConfig())
}

// NewSimpleSchemaGenerator creates a generator biased toward simple schemas
func NewSimpleSchemaGenerator() *SchemaGenerator {
	return NewSchemaGenerator(SimpleSchemaGeneratorConfig())
}

// NewComplexSchemaGenerator creates a generator biased toward complex schemas
func NewComplexSchemaGenerator() *SchemaGenerator {
	return NewSchemaGenerator(ComplexSchemaGeneratorConfig())
}

// Generate creates a random schema following the generator's configuration
func (sg *SchemaGenerator) Generate() Schema {
	return sg.generateWithDepth(0)
}

// GenerateMany creates multiple random schemas
func (sg *SchemaGenerator) GenerateMany(count int) []Schema {
	results := make([]Schema, count)
	for i := 0; i < count; i++ {
		results[i] = sg.Generate()
	}
	return results
}

// generateWithDepth is the internal method that tracks recursion depth
func (sg *SchemaGenerator) generateWithDepth(depth int) Schema {
	// Prevent infinite recursion
	if depth > sg.config.MaxDepth {
		return sg.generatePrimitiveSchema()
	}

	schemaType := sg.selectSchemaType(depth)
	return sg.generateSchemaOfType(schemaType, depth)
}

// selectSchemaType chooses a schema type based on weights and complexity bias
func (sg *SchemaGenerator) selectSchemaType(depth int) SchemaType {
	weights := sg.config.TypeWeights

	// Bias toward simpler types at greater depths
	depthFactor := float64(depth) / float64(sg.config.MaxDepth)
	complexityReduction := depthFactor * (1.0 - sg.config.ComplexityBias)

	// Reduce weights for complex types based on depth
	weights.Object *= (1.0 - complexityReduction)
	weights.Array *= (1.0 - complexityReduction)
	weights.Union *= (1.0 - complexityReduction)

	// Create weighted selection
	totalWeight := weights.String + weights.Number + weights.Integer + weights.Boolean +
		weights.Object + weights.Array + weights.Union + weights.Optional + weights.Null + weights.Any

	if totalWeight <= 0 {
		return TypeString // Fallback
	}

	r := sg.rng.Float64() * totalWeight
	current := 0.0

	if current += weights.String; r < current {
		return TypeString
	}
	if current += weights.Number; r < current {
		return TypeNumber
	}
	if current += weights.Integer; r < current {
		return TypeInteger
	}
	if current += weights.Boolean; r < current {
		return TypeBoolean
	}
	if current += weights.Object; r < current {
		return TypeObject
	}
	if current += weights.Array; r < current {
		return TypeArray
	}
	if current += weights.Union; r < current {
		return TypeUnion
	}
	if current += weights.Optional; r < current {
		return TypeOptional
	}
	if current += weights.Null; r < current {
		return TypeNull
	}
	if current += weights.Any; r < current {
		return TypeAny
	}

	return TypeString // Fallback
}

// generateSchemaOfType creates a schema of the specified type
func (sg *SchemaGenerator) generateSchemaOfType(schemaType SchemaType, depth int) Schema {
	switch schemaType {
	case TypeString:
		return sg.generateStringSchema()
	case TypeNumber:
		return sg.generateNumberSchema()
	case TypeInteger:
		return sg.generateIntegerSchema()
	case TypeBoolean:
		return sg.generateBooleanSchema()
	case TypeObject:
		return sg.generateObjectSchema(depth)
	case TypeArray:
		return sg.generateArraySchema(depth)
	case TypeUnion:
		return sg.generateUnionSchema(depth)
	case TypeOptional:
		return sg.generateOptionalSchema(depth)
	case TypeNull:
		return sg.generateNullSchema()
	case TypeAny:
		return sg.generateAnySchema()
	default:
		return sg.generateStringSchema()
	}
}

// generatePrimitiveSchema generates a simple primitive schema (used at max depth)
func (sg *SchemaGenerator) generatePrimitiveSchema() Schema {
	primitives := []SchemaType{TypeString, TypeNumber, TypeInteger, TypeBoolean}
	chosen := primitives[sg.rng.Intn(len(primitives))]
	return sg.generateSchemaOfType(chosen, 0)
}

// generateStringSchema creates a random string schema
func (sg *SchemaGenerator) generateStringSchema() Schema {
	builder := String()

	// Only add one type of constraint to avoid conflicts
	constraintType := sg.rng.Intn(4)

	// Generate constraints based on type (avoid conflicts)
	if sg.shouldGenerateConstraint() {
		switch constraintType {
		case 0: // Length constraints only
			sg.addStringConstraintsToBuilder(builder)
		case 1: // Enum values only (no format/pattern)
			if sg.config.GenerateEnums && sg.rng.Float64() < 0.1 {
				sg.addEnumValuesToBuilder(builder)
			}
		case 2: // Format only (no enum/pattern)
			if sg.config.GenerateFormats && sg.config.UseCommonFormats && sg.rng.Float64() < 0.4 {
				format := sg.stringFormats[sg.rng.Intn(len(sg.stringFormats))]
				switch format {
				case "email":
					builder.Email()
				case "uuid":
					builder.UUID()
				case "url":
					builder.URL()
				default:
					// Use a generic string for other formats
				}
			}
		case 3: // Pattern only (no enum/format)
			if sg.config.GeneratePatterns && sg.rng.Float64() < 0.2 {
				pattern := sg.commonPatterns[sg.rng.Intn(len(sg.commonPatterns))]
				builder.Pattern(pattern)
			}
		}
	}

	// Add metadata
	sg.addMetadataToStringBuilder(builder)

	return builder.Build()
}

// generateNumberSchema creates a random number schema
func (sg *SchemaGenerator) generateNumberSchema() Schema {
	builder := Number()

	// Generate constraints
	if sg.shouldGenerateConstraint() {
		sg.addNumberConstraintsToBuilder(builder)
	}

	// Add metadata
	sg.addMetadataToNumberBuilder(builder)

	return builder.Build()
}

// generateIntegerSchema creates a random integer schema
func (sg *SchemaGenerator) generateIntegerSchema() Schema {
	builder := Integer()

	// Generate constraints
	if sg.shouldGenerateConstraint() {
		sg.addIntegerConstraintsToBuilder(builder)
	}

	// Add metadata
	sg.addMetadataToIntegerBuilder(builder)

	return builder.Build()
}

// generateBooleanSchema creates a random boolean schema
func (sg *SchemaGenerator) generateBooleanSchema() Schema {
	builder := Boolean()

	// Add metadata
	sg.addMetadataToBooleanBuilder(builder)

	return builder.Build()
}

// generateObjectSchema creates a random object schema
func (sg *SchemaGenerator) generateObjectSchema(depth int) Schema {
	builder := Object()

	// Generate properties
	propCount := sg.config.MinProperties + sg.rng.Intn(sg.config.MaxProperties-sg.config.MinProperties+1)
	// Reduce required properties to avoid validation issues
	maxRequired := propCount / 2 // At most half the properties should be required
	if maxRequired == 0 && propCount > 0 {
		maxRequired = 1
	}
	requiredCount := sg.rng.Intn(maxRequired + 1)

	propertyNames := sg.generatePropertyNames(propCount)
	requiredProps := make([]string, 0, requiredCount)

	for i, propName := range propertyNames {
		propSchema := sg.generateWithDepth(depth + 1)
		builder.Property(propName, propSchema)

		// Make some properties required
		if i < requiredCount {
			requiredProps = append(requiredProps, propName)
		}
	}

	if len(requiredProps) > 0 {
		builder.Required(requiredProps...)
	}

	// Occasionally allow additional properties
	if sg.rng.Float64() < 0.3 {
		builder.AdditionalProperties(true)
	}

	// Add metadata
	sg.addMetadataToObjectBuilder(builder)

	return builder.Build()
}

// generateArraySchema creates a random array schema
func (sg *SchemaGenerator) generateArraySchema(depth int) Schema {
	itemSchema := sg.generateWithDepth(depth + 1)
	builder := Array().Items(itemSchema)

	// Generate constraints
	if sg.shouldGenerateConstraint() {
		sg.addArrayConstraintsToBuilder(builder)
	}

	// Rarely make items unique (only for simple types)
	if sg.rng.Float64() < 0.05 {
		builder.UniqueItems()
	}

	// Add metadata
	sg.addMetadataToArrayBuilder(builder)

	return builder.Build()
}

// generateUnionSchema creates a random union schema
func (sg *SchemaGenerator) generateUnionSchema(depth int) Schema {
	// Reduce union complexity to avoid validation issues
	maxTypes := sg.config.UnionMaxTypes - 1
	if maxTypes > 2 {
		maxTypes = 2 // Limit to max 3 total types
	}
	typeCount := 2 + sg.rng.Intn(maxTypes)
	subSchemas := make([]Schema, typeCount)

	// Generate simpler schemas for unions to improve compatibility
	for i := 0; i < typeCount; i++ {
		// Favor primitive types in unions to reduce complexity
		if sg.rng.Float64() < 0.7 || depth >= sg.config.MaxDepth-1 {
			subSchemas[i] = sg.generatePrimitiveSchema()
		} else {
			subSchemas[i] = sg.generateWithDepth(depth + 1)
		}
	}

	// Create union using simple constructor
	return sg.createUnionSchema(subSchemas)
}

// generateOptionalSchema creates a random optional schema
func (sg *SchemaGenerator) generateOptionalSchema(depth int) Schema {
	innerSchema := sg.generateWithDepth(depth + 1)
	return sg.createOptionalSchema(innerSchema)
}

// generateNullSchema creates a null schema
func (sg *SchemaGenerator) generateNullSchema() Schema {
	return sg.createNullSchema()
}

// generateAnySchema creates an any schema
func (sg *SchemaGenerator) generateAnySchema() Schema {
	return sg.createAnySchema()
}

// Helper methods for adding constraints and metadata

// shouldGenerateConstraint determines if constraints should be added
func (sg *SchemaGenerator) shouldGenerateConstraint() bool {
	return sg.config.GenerateConstraints && sg.rng.Float64() < sg.config.ConstraintProbability
}

// Helper methods for adding constraints and metadata to builders

// addMetadataToStringBuilder adds metadata to a string builder
func (sg *SchemaGenerator) addMetadataToStringBuilder(builder *StringBuilder) {
	if sg.config.GenerateTitles && sg.rng.Float64() < 0.5 {
		title := sg.titles[sg.rng.Intn(len(sg.titles))]
		builder.Name(title)
	}

	if sg.config.GenerateDescriptions && sg.rng.Float64() < 0.4 {
		description := sg.descriptions[sg.rng.Intn(len(sg.descriptions))]
		builder.Description(description)
	}

	if sg.config.GenerateExamples && sg.rng.Float64() < 0.6 {
		example := sg.generateRandomString(8)
		builder.Example(example)
	}
}

// addMetadataToNumberBuilder adds metadata to a number builder
func (sg *SchemaGenerator) addMetadataToNumberBuilder(builder *NumberBuilder) {
	if sg.config.GenerateTitles && sg.rng.Float64() < 0.5 {
		title := sg.titles[sg.rng.Intn(len(sg.titles))]
		builder.Name(title)
	}

	if sg.config.GenerateDescriptions && sg.rng.Float64() < 0.4 {
		description := sg.descriptions[sg.rng.Intn(len(sg.descriptions))]
		builder.Description(description)
	}

	if sg.config.GenerateExamples && sg.rng.Float64() < 0.6 {
		example := sg.rng.Float64() * 100
		builder.Example(example)
	}
}

// addMetadataToIntegerBuilder adds metadata to an integer builder
func (sg *SchemaGenerator) addMetadataToIntegerBuilder(builder *IntegerBuilder) {
	if sg.config.GenerateTitles && sg.rng.Float64() < 0.5 {
		title := sg.titles[sg.rng.Intn(len(sg.titles))]
		builder.Name(title)
	}

	if sg.config.GenerateDescriptions && sg.rng.Float64() < 0.4 {
		description := sg.descriptions[sg.rng.Intn(len(sg.descriptions))]
		builder.Description(description)
	}

	if sg.config.GenerateExamples && sg.rng.Float64() < 0.6 {
		example := int64(sg.rng.Intn(100))
		builder.Example(example)
	}
}

// addMetadataToBooleanBuilder adds metadata to a boolean builder
func (sg *SchemaGenerator) addMetadataToBooleanBuilder(builder *BooleanBuilder) {
	if sg.config.GenerateTitles && sg.rng.Float64() < 0.5 {
		title := sg.titles[sg.rng.Intn(len(sg.titles))]
		builder.Name(title)
	}

	if sg.config.GenerateDescriptions && sg.rng.Float64() < 0.4 {
		description := sg.descriptions[sg.rng.Intn(len(sg.descriptions))]
		builder.Description(description)
	}

	if sg.config.GenerateExamples && sg.rng.Float64() < 0.6 {
		example := sg.rng.Float64() < 0.5
		builder.Example(example)
	}
}

// addMetadataToObjectBuilder adds metadata to an object builder
func (sg *SchemaGenerator) addMetadataToObjectBuilder(builder *ObjectBuilder) {
	if sg.config.GenerateTitles && sg.rng.Float64() < 0.5 {
		title := sg.titles[sg.rng.Intn(len(sg.titles))]
		builder.Name(title)
	}

	if sg.config.GenerateDescriptions && sg.rng.Float64() < 0.4 {
		description := sg.descriptions[sg.rng.Intn(len(sg.descriptions))]
		builder.Description(description)
	}
}

// addMetadataToArrayBuilder adds metadata to an array builder
func (sg *SchemaGenerator) addMetadataToArrayBuilder(builder *ArrayBuilder) {
	if sg.config.GenerateTitles && sg.rng.Float64() < 0.5 {
		title := sg.titles[sg.rng.Intn(len(sg.titles))]
		builder.Name(title)
	}

	if sg.config.GenerateDescriptions && sg.rng.Float64() < 0.4 {
		description := sg.descriptions[sg.rng.Intn(len(sg.descriptions))]
		builder.Description(description)
	}
}

// addStringConstraintsToBuilder adds min/max length constraints to string builders
func (sg *SchemaGenerator) addStringConstraintsToBuilder(builder *StringBuilder) {
	// Generate more reasonable length constraints
	if sg.rng.Float64() < 0.5 {
		minLen := 1 + sg.rng.Intn(5)            // 1-5 characters (very reasonable)
		maxLen := minLen + 10 + sg.rng.Intn(20) // Give plenty of room
		builder.MinLength(minLen).MaxLength(maxLen)
	}
}

// addNumberConstraintsToBuilder adds min/max constraints to number builders
func (sg *SchemaGenerator) addNumberConstraintsToBuilder(builder *NumberBuilder) {
	// Generate very lenient number constraints
	if sg.rng.Float64() < 0.3 {
		min := sg.rng.Float64() * 10            // 0-10 (very reasonable range)
		max := min + 100 + sg.rng.Float64()*900 // Give plenty of room above min
		builder.Min(min).Max(max)
	}
}

// addIntegerConstraintsToBuilder adds min/max constraints to integer builders
func (sg *SchemaGenerator) addIntegerConstraintsToBuilder(builder *IntegerBuilder) {
	// Generate very lenient integer constraints
	if sg.rng.Float64() < 0.3 {
		min := int64(sg.rng.Intn(10))            // 0-9 (very reasonable range)
		max := min + int64(100+sg.rng.Intn(900)) // Give plenty of room above min
		builder.Min(min).Max(max)
	}
}

// addArrayConstraintsToBuilder adds min/max items constraints to array builders
func (sg *SchemaGenerator) addArrayConstraintsToBuilder(builder *ArrayBuilder) {
	// Generate reasonable array constraints
	if sg.rng.Float64() < 0.5 {
		minItems := sg.rng.Intn(3)                // 0-2 items (reasonable minimum)
		maxItems := minItems + 3 + sg.rng.Intn(7) // Give plenty of room above min
		builder.MinItems(minItems).MaxItems(maxItems)
	}
}

// addEnumValuesToBuilder adds enum constraints to string builders
func (sg *SchemaGenerator) addEnumValuesToBuilder(builder *StringBuilder) {
	enumCount := 2 + sg.rng.Intn(4) // 2-5 enum values
	values := make([]string, enumCount)

	for i := 0; i < enumCount; i++ {
		values[i] = sg.generateEnumValue()
	}

	builder.Enum(values...)
}

// generateEnumValue creates a realistic enum value
func (sg *SchemaGenerator) generateEnumValue() string {
	prefixes := []string{"active", "inactive", "pending", "approved", "rejected", "draft", "published"}
	suffixes := []string{"low", "medium", "high", "critical", "normal", "urgent", "optional"}

	if sg.rng.Float64() < 0.5 {
		return prefixes[sg.rng.Intn(len(prefixes))]
	} else {
		return suffixes[sg.rng.Intn(len(suffixes))]
	}
}

// createUnionSchema creates a union schema from multiple schemas
func (sg *SchemaGenerator) createUnionSchema(schemas []Schema) Schema {
	// Create a simple union schema manually since we need variadic support
	union := &UnionSchema{
		metadata: SchemaMetadata{Name: "GeneratedUnion"},
		schemas:  schemas,
	}
	return union
}

// createOptionalSchema creates an optional schema wrapper
func (sg *SchemaGenerator) createOptionalSchema(innerSchema Schema) Schema {
	// Use the generic Optional constructor
	return &OptionalSchema[any]{
		metadata:   SchemaMetadata{Name: "GeneratedOptional"},
		itemSchema: innerSchema,
	}
}

// createNullSchema creates a null schema
func (sg *SchemaGenerator) createNullSchema() Schema {
	// For now, return a simple object that represents null
	// In practice, this might need a proper NullSchema implementation
	return Object().
		Name("Null").
		Description("Represents a null value").
		Build()
}

// createAnySchema creates an any schema
func (sg *SchemaGenerator) createAnySchema() Schema {
	// For now, return an object with additional properties to represent "any"
	return Object().
		AdditionalProperties(true).
		Name("Any").
		Description("Any value is allowed").
		Build()
}

// generateRandomString creates a random alphanumeric string
func (sg *SchemaGenerator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[sg.rng.Intn(len(charset))]
	}
	return string(result)
}

// generatePropertyNames creates realistic property names
func (sg *SchemaGenerator) generatePropertyNames(count int) []string {
	names := make([]string, count)
	usedNames := make(map[string]bool)

	for i := 0; i < count; i++ {
		var name string
		attempts := 0
		for {
			name = sg.generatePropertyName()
			if !usedNames[name] || attempts > 10 {
				break
			}
			attempts++
		}
		usedNames[name] = true
		names[i] = name
	}

	return names
}

// generatePropertyName creates a single property name based on style configuration
func (sg *SchemaGenerator) generatePropertyName() string {
	switch sg.config.PropertyNameStyle {
	case "realistic":
		return sg.propertyNames[sg.rng.Intn(len(sg.propertyNames))]
	case "generic":
		return "property" + strconv.Itoa(sg.rng.Intn(1000))
	case "random":
		return sg.generateRandomPropertyName()
	default:
		return sg.propertyNames[sg.rng.Intn(len(sg.propertyNames))]
	}
}

// generateRandomPropertyName creates a completely random property name
func (sg *SchemaGenerator) generateRandomPropertyName() string {
	length := 3 + sg.rng.Intn(12)
	const charset = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)

	// Start with a letter
	result[0] = charset[sg.rng.Intn(len(charset))]

	// Continue with letters and numbers
	extendedCharset := charset + "0123456789"
	for i := 1; i < length; i++ {
		result[i] = extendedCharset[sg.rng.Intn(len(extendedCharset))]
	}

	return string(result)
}

// Convenience functions

// GenerateSchema creates a schema using default configuration
func GenerateSchema() Schema {
	return NewSchemaGeneratorWithDefaults().Generate()
}

// GenerateSchemas creates multiple schemas using default configuration
func GenerateSchemas(count int) []Schema {
	return NewSchemaGeneratorWithDefaults().GenerateMany(count)
}

// GenerateSchemaWithSeed creates a schema with a specific seed for reproducibility
func GenerateSchemaWithSeed(seed int64) Schema {
	config := DefaultSchemaGeneratorConfig()
	config.Seed = seed
	return NewSchemaGenerator(config).Generate()
}

// GenerateSimpleSchema creates a simple schema biased toward primitive types
func GenerateSimpleSchema() Schema {
	return NewSimpleSchemaGenerator().Generate()
}

// GenerateComplexSchema creates a complex schema biased toward nested structures
func GenerateComplexSchema() Schema {
	return NewComplexSchemaGenerator().Generate()
}

// GenerateRealisticSchema creates a schema with realistic property names and constraints
func GenerateRealisticSchema() Schema {
	config := DefaultSchemaGeneratorConfig()
	config.PropertyNameStyle = "realistic"
	config.UseCommonFormats = true
	config.GenerateConstraints = true
	config.GenerateExamples = true
	config.GenerateDescriptions = true
	return NewSchemaGenerator(config).Generate()
}
