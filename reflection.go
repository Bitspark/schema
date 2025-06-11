// Package schema - struct generation and reflection utilities
package schema

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// FromStruct generates a schema from a Go struct type using reflection and struct tags.
// Example: userSchema := schema.FromStruct[User]()
func FromStruct[T any]() Schema {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	return generateSchemaFromType(typ)
}

// StructAnalyzer handles the analysis of Go structs and generation of schemas.
type StructAnalyzer struct {
	tagParser    *TagParser
	typeRegistry *TypeRegistry
	cache        sync.Map
}

// TagParser parses and validates struct tags for schema generation.
type TagParser struct {
	validators map[string]func(string) (any, error)
}

// TypeRegistry manages custom type mappings for struct generation.
type TypeRegistry struct {
	mappings map[reflect.Type]func() Schema
	mu       sync.RWMutex
}

// Global instances
var (
	defaultAnalyzer = &StructAnalyzer{
		tagParser:    newTagParser(),
		typeRegistry: newTypeRegistry(),
	}
)

// generateSchemaFromType generates a schema from a reflect.Type.
func generateSchemaFromType(typ reflect.Type) Schema {
	// Check cache first
	if cached, ok := defaultAnalyzer.cache.Load(typ); ok {
		return cached.(Schema).Clone()
	}

	schema := generateSchemaFromTypeUncached(typ)
	defaultAnalyzer.cache.Store(typ, schema)
	return schema.Clone()
}

// generateSchemaFromTypeUncached generates schema without caching.
func generateSchemaFromTypeUncached(typ reflect.Type) Schema {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		return generateSchemaFromType(typ.Elem())
	}

	// Check for custom type mapping
	if mapping := defaultAnalyzer.typeRegistry.GetMapping(typ); mapping != nil {
		return mapping()
	}

	switch typ.Kind() {
	case reflect.String:
		return String().Build()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Integer().Build()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Integer().Build()
	case reflect.Float32, reflect.Float64:
		return Number().Build()
	case reflect.Bool:
		return Boolean().Build()
	case reflect.Slice, reflect.Array:
		itemSchema := generateSchemaFromType(typ.Elem())
		return Array().Items(itemSchema).Build()
	case reflect.Map:
		// Map types are treated as objects with additional properties
		if typ.Key().Kind() == reflect.String {
			return Object().AdditionalProperties(true).
				Name(fmt.Sprintf("Map[string]%s", getTypeName(typ.Elem()))).
				Build()
		}
		return Object().AdditionalProperties(true).Build()
	case reflect.Struct:
		return generateObjectSchemaFromStruct(typ)
	case reflect.Interface:
		// Interfaces are treated as "any" type
		return Object().AdditionalProperties(true).
			Name("any").
			Description("Any value").
			Build()
	default:
		// Fallback to object for unknown types
		return Object().AdditionalProperties(true).
			Name(typ.String()).
			Build()
	}
}

// generateObjectSchemaFromStruct generates an object schema from a struct type.
func generateObjectSchemaFromStruct(typ reflect.Type) Schema {
	builder := Object().Name(typ.Name())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded fields
		if field.Anonymous {
			embeddedSchema := generateSchemaFromType(field.Type)
			if objSchema, ok := embeddedSchema.(*ObjectSchema); ok {
				// Merge embedded object properties
				for propName, propSchema := range objSchema.properties {
					builder.Property(propName, propSchema)
				}
				// Merge required fields
				builder.Required(objSchema.required...)
			}
			continue
		}

		// Get JSON field name
		jsonName := getJSONFieldName(field)
		if jsonName == "-" {
			continue // Skip fields marked with json:"-"
		}

		// Generate field schema
		fieldSchema := generateFieldSchema(field)
		builder.Property(jsonName, fieldSchema)

		// Check if field is required (not a pointer and no omitempty)
		if isRequiredField(field) {
			builder.Required(jsonName)
		}
	}

	return builder.Build()
}

// generateFieldSchema generates a schema for a specific struct field.
func generateFieldSchema(field reflect.StructField) Schema {
	// Start with base schema from field type
	baseSchema := generateSchemaFromType(field.Type)

	// Parse schema tags
	schemaTag := field.Tag.Get("schema")
	if schemaTag == "" {
		return baseSchema
	}

	// Apply schema tag modifications
	return applySchemaTag(baseSchema, schemaTag, field)
}

// applySchemaTag applies schema tag directives to a base schema.
func applySchemaTag(baseSchema Schema, tagValue string, field reflect.StructField) Schema {
	tags := parseSchemaTag(tagValue)

	// Clone the base schema for modification
	schema := baseSchema.Clone()

	// Apply modifications based on schema type
	switch s := schema.(type) {
	case *StringSchema:
		return applyStringTags(s, tags, field)
	case *NumberSchema:
		return applyNumberTags(s, tags, field)
	case *IntegerSchema:
		return applyIntegerTags(s, tags, field)
	case *BooleanSchema:
		return applyBooleanTags(s, tags, field)
	case *ArraySchema:
		return applyArrayTags(s, tags, field)
	case *ObjectSchema:
		return applyObjectTags(s, tags, field)
	}

	return schema
}

// parseSchemaTag parses a schema tag value into a map of directives.
func parseSchemaTag(tagValue string) map[string]string {
	tags := make(map[string]string)
	parts := strings.Split(tagValue, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			tags[part] = "true"
		}
	}

	return tags
}

// Tag application functions for different schema types

func applyStringTags(schema *StringSchema, tags map[string]string, field reflect.StructField) Schema {
	builder := String()

	// Copy existing properties
	if schema.minLength != nil {
		builder.MinLength(*schema.minLength)
	}
	if schema.maxLength != nil {
		builder.MaxLength(*schema.maxLength)
	}
	if schema.pattern != "" {
		builder.Pattern(schema.pattern)
	}
	if schema.format != "" {
		// Format will be overridden by tags if present
	}
	if len(schema.enumValues) > 0 {
		builder.Enum(schema.enumValues...)
	}

	// Apply tag directives
	for key, value := range tags {
		switch key {
		case "minlen", "minlength":
			if val, err := strconv.Atoi(value); err == nil {
				builder.MinLength(val)
			}
		case "maxlen", "maxlength":
			if val, err := strconv.Atoi(value); err == nil {
				builder.MaxLength(val)
			}
		case "pattern":
			builder.Pattern(value)
		case "email":
			builder.Email()
		case "url":
			builder.URL()
		case "uuid":
			builder.UUID()
		case "format":
			switch value {
			case "email":
				builder.Email()
			case "url":
				builder.URL()
			case "uuid":
				builder.UUID()
			default:
				// Custom format
				builder.schema.format = value
			}
		case "enum":
			values := strings.Split(value, "|")
			builder.Enum(values...)
		case "desc", "description":
			builder.Description(value)
		case "example":
			builder.Example(value)
		}
	}

	return builder.Build()
}

func applyNumberTags(schema *NumberSchema, tags map[string]string, field reflect.StructField) Schema {
	builder := Number()

	// Copy existing properties
	if schema.minimum != nil {
		builder.Min(*schema.minimum)
	}
	if schema.maximum != nil {
		builder.Max(*schema.maximum)
	}

	// Apply tag directives
	for key, value := range tags {
		switch key {
		case "min", "minimum":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				builder.Min(val)
			}
		case "max", "maximum":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				builder.Max(val)
			}
		case "desc", "description":
			builder.Description(value)
		case "example":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				builder.Example(val)
			}
		}
	}

	return builder.Build()
}

func applyIntegerTags(schema *IntegerSchema, tags map[string]string, field reflect.StructField) Schema {
	builder := Integer()

	// Copy existing properties
	if schema.minimum != nil {
		builder.Min(*schema.minimum)
	}
	if schema.maximum != nil {
		builder.Max(*schema.maximum)
	}

	// Apply tag directives
	for key, value := range tags {
		switch key {
		case "min", "minimum":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				builder.Min(val)
			}
		case "max", "maximum":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				builder.Max(val)
			}
		case "desc", "description":
			builder.Description(value)
		case "example":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				builder.Example(val)
			}
		}
	}

	return builder.Build()
}

func applyBooleanTags(schema *BooleanSchema, tags map[string]string, field reflect.StructField) Schema {
	builder := Boolean()

	// Apply tag directives
	for key, value := range tags {
		switch key {
		case "desc", "description":
			builder.Description(value)
		case "example":
			if val, err := strconv.ParseBool(value); err == nil {
				builder.Example(val)
			}
		}
	}

	return builder.Build()
}

func applyArrayTags(schema *ArraySchema, tags map[string]string, field reflect.StructField) Schema {
	builder := Array()

	// Copy existing properties
	if schema.itemSchema != nil {
		builder.Items(schema.itemSchema)
	}
	if schema.minItems != nil {
		builder.MinItems(*schema.minItems)
	}
	if schema.maxItems != nil {
		builder.MaxItems(*schema.maxItems)
	}
	if schema.uniqueItems {
		builder.UniqueItems()
	}

	// Apply tag directives
	for key, value := range tags {
		switch key {
		case "minitems":
			if val, err := strconv.Atoi(value); err == nil {
				builder.MinItems(val)
			}
		case "maxitems":
			if val, err := strconv.Atoi(value); err == nil {
				builder.MaxItems(val)
			}
		case "unique":
			builder.UniqueItems()
		case "items_enum":
			// For arrays where items should be enum values
			if schema.itemSchema != nil {
				if _, ok := schema.itemSchema.(*StringSchema); ok {
					values := strings.Split(value, "|")
					newStringSchema := String().Enum(values...).Build()
					builder.Items(newStringSchema)
				}
			}
		case "desc", "description":
			builder.Description(value)
		}
	}

	return builder.Build()
}

func applyObjectTags(schema *ObjectSchema, tags map[string]string, field reflect.StructField) Schema {
	// For now, object-level tags are limited
	// Most object configuration happens at the property level
	for key, value := range tags {
		switch key {
		case "desc", "description":
			metadata := schema.metadata
			metadata.Description = value
			return schema.WithMetadata(metadata)
		}
	}

	return schema
}

// Helper functions

// getJSONFieldName extracts the JSON field name from struct field tags.
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name
	}

	parts := strings.Split(jsonTag, ",")
	name := strings.TrimSpace(parts[0])
	if name == "" {
		return field.Name
	}

	return name
}

// isRequiredField determines if a struct field should be required in the schema.
func isRequiredField(field reflect.StructField) bool {
	// Check schema tag for explicit required/optional
	schemaTag := field.Tag.Get("schema")
	if strings.Contains(schemaTag, "required") {
		return true
	}
	if strings.Contains(schemaTag, "optional") {
		return false
	}

	// Check JSON tag for omitempty
	jsonTag := field.Tag.Get("json")
	if strings.Contains(jsonTag, "omitempty") {
		return false
	}

	// Pointer types are optional by default
	if field.Type.Kind() == reflect.Ptr {
		return false
	}

	// Non-pointer types are required by default
	return true
}

// getTypeName returns a human-readable name for a type.
func getTypeName(typ reflect.Type) string {
	if typ.Name() != "" {
		return typ.Name()
	}
	return typ.String()
}

// Type registry functions

func newTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		mappings: make(map[reflect.Type]func() Schema),
	}
}

func (tr *TypeRegistry) RegisterMapping(typ reflect.Type, schemaFactory func() Schema) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.mappings[typ] = schemaFactory
}

func (tr *TypeRegistry) GetMapping(typ reflect.Type) func() Schema {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return tr.mappings[typ]
}

// RegisterTypeMapping registers a custom schema mapping for a specific type.
func RegisterTypeMapping(typ reflect.Type, schemaFactory func() Schema) {
	defaultAnalyzer.typeRegistry.RegisterMapping(typ, schemaFactory)
}

// Tag parser functions

func newTagParser() *TagParser {
	return &TagParser{
		validators: map[string]func(string) (any, error){
			"min":         parseNumber,
			"max":         parseNumber,
			"minlen":      parseInt,
			"maxlen":      parseInt,
			"pattern":     parseString,
			"format":      parseString,
			"enum":        parseEnum,
			"description": parseString,
			"example":     parseString,
		},
	}
}

func parseNumber(value string) (any, error) {
	return strconv.ParseFloat(value, 64)
}

func parseInt(value string) (any, error) {
	return strconv.Atoi(value)
}

func parseString(value string) (any, error) {
	return value, nil
}

func parseEnum(value string) (any, error) {
	return strings.Split(value, "|"), nil
}

// Example usage and validation

// ValidateStructTag validates that a struct has valid schema tags.
func ValidateStructTag(typ reflect.Type) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("type %s is not a struct", typ.String())
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		schemaTag := field.Tag.Get("schema")
		if schemaTag == "" {
			continue
		}

		if err := validateFieldTag(field, schemaTag); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	return nil
}

// validateFieldTag validates a single field's schema tag.
func validateFieldTag(field reflect.StructField, tagValue string) error {
	tags := parseSchemaTag(tagValue)

	// Basic validation - ensure no contradictory tags
	if _, hasRequired := tags["required"]; hasRequired {
		if _, hasOptional := tags["optional"]; hasOptional {
			return fmt.Errorf("cannot have both 'required' and 'optional' tags")
		}
	}

	// Type-specific validation
	switch field.Type.Kind() {
	case reflect.String:
		return validateStringTags(tags)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return validateIntegerTags(tags)
	case reflect.Float32, reflect.Float64:
		return validateNumberTags(tags)
	}

	return nil
}

func validateStringTags(tags map[string]string) error {
	if minLenStr, ok := tags["minlen"]; ok {
		if _, err := strconv.Atoi(minLenStr); err != nil {
			return fmt.Errorf("invalid minlen value: %s", minLenStr)
		}
	}

	if maxLenStr, ok := tags["maxlen"]; ok {
		if _, err := strconv.Atoi(maxLenStr); err != nil {
			return fmt.Errorf("invalid maxlen value: %s", maxLenStr)
		}
	}

	if pattern, ok := tags["pattern"]; ok {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %s", pattern)
		}
	}

	return nil
}

func validateIntegerTags(tags map[string]string) error {
	if minStr, ok := tags["min"]; ok {
		if _, err := strconv.ParseInt(minStr, 10, 64); err != nil {
			return fmt.Errorf("invalid min value: %s", minStr)
		}
	}

	if maxStr, ok := tags["max"]; ok {
		if _, err := strconv.ParseInt(maxStr, 10, 64); err != nil {
			return fmt.Errorf("invalid max value: %s", maxStr)
		}
	}

	return nil
}

func validateNumberTags(tags map[string]string) error {
	if minStr, ok := tags["min"]; ok {
		if _, err := strconv.ParseFloat(minStr, 64); err != nil {
			return fmt.Errorf("invalid min value: %s", minStr)
		}
	}

	if maxStr, ok := tags["max"]; ok {
		if _, err := strconv.ParseFloat(maxStr, 64); err != nil {
			return fmt.Errorf("invalid max value: %s", maxStr)
		}
	}

	return nil
}
