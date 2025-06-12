package native

import (
	"fmt"
	"reflect"
	"sync"

	"defs.dev/schema/annotation"
	"defs.dev/schema/api/core"
	"defs.dev/schema/builders"
	"defs.dev/schema/registry"
)

// DefaultTypeConverter implements TypeConverter with annotation support.
type DefaultTypeConverter struct {
	annotationRegistry annotation.AnnotationRegistry
	validatorRegistry  registry.ValidatorRegistry
	tagParser          TagParser
	config             ConverterConfig
	cache              map[reflect.Type]core.Schema
	mu                 sync.RWMutex
}

// NewDefaultTypeConverter creates a new type converter with default configuration.
func NewDefaultTypeConverter(
	annotationRegistry annotation.AnnotationRegistry,
	validatorRegistry registry.ValidatorRegistry,
) *DefaultTypeConverter {
	converter := &DefaultTypeConverter{
		annotationRegistry: annotationRegistry,
		validatorRegistry:  validatorRegistry,
		cache:              make(map[reflect.Type]core.Schema),
		config: ConverterConfig{
			StrictMode:          false,
			DefaultAnnotations:  true,
			ValidateAnnotations: true,
			RecursiveConversion: true,
			MaxDepth:            10,
			CacheResults:        true,
			IgnoreUnknownTags:   true,
		},
	}

	converter.tagParser = NewDefaultTagParser(annotationRegistry)
	return converter
}

// FromType implements TypeConverter.
func (c *DefaultTypeConverter) FromType(t reflect.Type) (core.Schema, error) {
	return c.FromTypeWithAnnotations(t, nil)
}

// FromValue implements TypeConverter.
func (c *DefaultTypeConverter) FromValue(v any) (core.Schema, error) {
	return c.FromValueWithAnnotations(v, nil)
}

// FromTypeName implements TypeConverter.
func (c *DefaultTypeConverter) FromTypeName(typeName string) (core.Schema, error) {
	// This would require type registry or reflection-based lookup
	return nil, fmt.Errorf("FromTypeName not implemented - need type registry")
}

// FromTypeWithAnnotations implements TypeConverter.
func (c *DefaultTypeConverter) FromTypeWithAnnotations(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error) {
	// Check cache first
	if c.config.CacheResults {
		c.mu.RLock()
		if cached, exists := c.cache[t]; exists {
			c.mu.RUnlock()
			return cached, nil
		}
		c.mu.RUnlock()
	}

	schema, err := c.convertType(t, annotations, 0)
	if err != nil {
		return nil, err
	}

	// Cache result
	if c.config.CacheResults {
		c.mu.Lock()
		c.cache[t] = schema
		c.mu.Unlock()
	}

	return schema, nil
}

// FromValueWithAnnotations implements TypeConverter.
func (c *DefaultTypeConverter) FromValueWithAnnotations(v any, annotations []annotation.Annotation) (core.Schema, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot convert nil value")
	}

	t := reflect.TypeOf(v)
	return c.FromTypeWithAnnotations(t, annotations)
}

// FromTypes implements TypeConverter.
func (c *DefaultTypeConverter) FromTypes(types map[string]reflect.Type) (map[string]core.Schema, error) {
	results := make(map[string]core.Schema, len(types))

	for name, t := range types {
		schema, err := c.FromType(t)
		if err != nil {
			return nil, fmt.Errorf("failed to convert type %s: %v", name, err)
		}
		results[name] = schema
	}

	return results, nil
}

// FromValues implements TypeConverter.
func (c *DefaultTypeConverter) FromValues(values map[string]any) (map[string]core.Schema, error) {
	results := make(map[string]core.Schema, len(values))

	for name, v := range values {
		schema, err := c.FromValue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value %s: %v", name, err)
		}
		results[name] = schema
	}

	return results, nil
}

// SetAnnotationRegistry implements TypeConverter.
func (c *DefaultTypeConverter) SetAnnotationRegistry(registry annotation.AnnotationRegistry) {
	c.annotationRegistry = registry
	if c.tagParser != nil {
		c.tagParser.SetAnnotationRegistry(registry)
	}
}

// SetValidatorRegistry implements TypeConverter.
func (c *DefaultTypeConverter) SetValidatorRegistry(registry registry.ValidatorRegistry) {
	c.validatorRegistry = registry
}

// SetStrictMode implements TypeConverter.
func (c *DefaultTypeConverter) SetStrictMode(strict bool) {
	c.config.StrictMode = strict
	if c.tagParser != nil {
		c.tagParser.SetStrictMode(strict)
	}
}

// GetSupportedTags implements TypeConverter.
func (c *DefaultTypeConverter) GetSupportedTags() []string {
	if c.tagParser != nil {
		return c.tagParser.GetSupportedTags()
	}
	return []string{}
}

// GetConfiguration implements TypeConverter.
func (c *DefaultTypeConverter) GetConfiguration() ConverterConfig {
	return c.config
}

// SetTagParser allows custom tag parser.
func (c *DefaultTypeConverter) SetTagParser(parser TagParser) {
	c.tagParser = parser
}

// Core conversion logic

func (c *DefaultTypeConverter) convertType(t reflect.Type, annotations []annotation.Annotation, depth int) (core.Schema, error) {
	if depth > c.config.MaxDepth {
		return nil, fmt.Errorf("maximum conversion depth exceeded")
	}

	switch t.Kind() {
	case reflect.String:
		return c.convertString(t, annotations)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.convertInteger(t, annotations)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return c.convertInteger(t, annotations)
	case reflect.Float32, reflect.Float64:
		return c.convertNumber(t, annotations)
	case reflect.Bool:
		return c.convertBoolean(t, annotations)
	case reflect.Slice, reflect.Array:
		return c.convertArray(t, annotations, depth)
	case reflect.Map:
		return c.convertMap(t, annotations, depth)
	case reflect.Struct:
		return c.convertStruct(t, annotations, depth)
	case reflect.Ptr:
		return c.convertPointer(t, annotations, depth)
	case reflect.Interface:
		return c.convertInterface(t, annotations)
	default:
		return nil, fmt.Errorf("unsupported type kind: %v", t.Kind())
	}
}

func (c *DefaultTypeConverter) convertString(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error) {
	builder := builders.NewStringSchema()

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertInteger(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error) {
	builder := builders.NewIntegerSchema()

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertNumber(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error) {
	builder := builders.NewNumberSchema()

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertBoolean(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error) {
	builder := builders.NewBooleanSchema()

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertArray(t reflect.Type, annotations []annotation.Annotation, depth int) (core.Schema, error) {
	elementType := t.Elem()
	elementSchema, err := c.convertType(elementType, nil, depth+1)
	if err != nil {
		return nil, fmt.Errorf("failed to convert array element type: %v", err)
	}

	builder := builders.NewArraySchema().Items(elementSchema)

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertMap(t reflect.Type, annotations []annotation.Annotation, depth int) (core.Schema, error) {
	// For maps, we create an object schema with dictionary-like behavior
	valueType := t.Elem()
	valueSchema, err := c.convertType(valueType, nil, depth+1)
	if err != nil {
		return nil, fmt.Errorf("failed to convert map value type: %v", err)
	}

	builder := builders.NewObjectSchema().Dict(valueSchema)

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertStruct(t reflect.Type, annotations []annotation.Annotation, depth int) (core.Schema, error) {
	builder := builders.NewObjectSchema()

	// Process struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse field tags to annotations
		fieldAnnotations, err := c.tagParser.ParseTags(field.Tag)
		if err != nil && c.config.StrictMode {
			return nil, fmt.Errorf("failed to parse tags for field %s: %v", field.Name, err)
		}

		// Check if field should be ignored (json:"-")
		shouldIgnore := false
		fieldName := field.Name

		for _, ann := range fieldAnnotations {
			if ann.Name() == "json" {
				if jsonData, ok := ann.Value().(map[string]interface{}); ok {
					if name, ok := jsonData["name"].(string); ok {
						if name == "-" {
							shouldIgnore = true
							break
						}
						fieldName = name
					}
				}
			}
		}

		if shouldIgnore {
			continue
		}

		// Convert field type
		fieldSchema, err := c.convertType(field.Type, fieldAnnotations, depth+1)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %v", field.Name, err)
		}

		// Add field to object
		builder = builder.Property(fieldName, fieldSchema).(*builders.ObjectBuilder)

		// Check if field is required
		for _, ann := range fieldAnnotations {
			if ann.Name() == "required" {
				if required, ok := ann.Value().(bool); ok && required {
					builder = builder.Required(fieldName).(*builders.ObjectBuilder)
				}
			}
		}
	}

	// Apply struct-level annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

func (c *DefaultTypeConverter) convertPointer(t reflect.Type, annotations []annotation.Annotation, depth int) (core.Schema, error) {
	// For pointers, convert the pointed-to type
	elemType := t.Elem()
	return c.convertType(elemType, annotations, depth+1)
}

func (c *DefaultTypeConverter) convertInterface(t reflect.Type, annotations []annotation.Annotation) (core.Schema, error) {
	// For interfaces, we create a flexible schema
	builder := builders.NewObjectSchema().Flexible()

	// Apply annotations
	for _, ann := range annotations {
		if err := c.applyAnnotationToBuilder(builder, ann); err != nil {
			return nil, fmt.Errorf("failed to apply annotation %s: %v", ann.Name(), err)
		}
	}

	return builder.Build(), nil
}

// Helper method to apply annotations to builders
func (c *DefaultTypeConverter) applyAnnotationToBuilder(builder any, ann annotation.Annotation) error {
	// This is a simplified implementation - in practice, you'd need to handle
	// different builder types and annotation types more systematically

	switch ann.Name() {
	case "description":
		// Apply description to builder if it supports it
		// Note: This would need proper builder interface methods
	case "default":
		// Apply default value
	case "examples":
		// Apply examples
	case "enum":
		// Apply enum constraints
	case "format":
		// Apply format constraints
	case "pattern":
		// Apply pattern constraints
	case "min", "max", "minLength", "maxLength", "minItems", "maxItems":
		// Apply numeric/length constraints
	}

	return nil
}
