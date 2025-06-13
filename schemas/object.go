package schemas

import (
	"fmt"
	"reflect"

	"defs.dev/schema/core"
)

// ObjectSchemaConfig holds the configuration for building an ObjectSchema.
type ObjectSchemaConfig struct {
	Metadata             core.SchemaMetadata
	Annotations          []core.Annotation
	Properties           map[string]core.Schema
	Required             []string
	AdditionalProperties bool
	MinProperties        *int
	MaxProperties        *int
	PatternProperties    map[string]core.Schema
	PropertyDependencies map[string][]string
	DefaultVal           map[string]any
}

// ObjectSchema is a clean, API-first implementation of object schema validation.
// It implements core.ObjectSchema interface and provides immutable operations.
type ObjectSchema struct {
	config ObjectSchemaConfig
}

// Ensure ObjectSchema implements the API interfaces at compile time
var _ core.Schema = (*ObjectSchema)(nil)
var _ core.ObjectSchema = (*ObjectSchema)(nil)
var _ core.Accepter = (*ObjectSchema)(nil)

// NewObjectSchema creates a new ObjectSchema with the given configuration.
func NewObjectSchema(config ObjectSchemaConfig) *ObjectSchema {
	return &ObjectSchema{config: config}
}

// Type returns the schema type constant.
func (o *ObjectSchema) Type() core.SchemaType {
	return core.TypeStructure
}

// Metadata returns the schema metadata.
func (o *ObjectSchema) Metadata() core.SchemaMetadata {
	return o.config.Metadata
}

// Annotations returns the annotations for this schema.
func (o *ObjectSchema) Annotations() []core.Annotation {
	if o.config.Annotations == nil {
		return nil
	}
	result := make([]core.Annotation, len(o.config.Annotations))
	copy(result, o.config.Annotations)
	return result
}

// Clone returns a deep copy of the ObjectSchema.
func (o *ObjectSchema) Clone() core.Schema {
	newConfig := o.config

	// Deep copy metadata examples and tags
	if o.config.Metadata.Examples != nil {
		newConfig.Metadata.Examples = make([]any, len(o.config.Metadata.Examples))
		copy(newConfig.Metadata.Examples, o.config.Metadata.Examples)
	}

	if o.config.Metadata.Tags != nil {
		newConfig.Metadata.Tags = make([]string, len(o.config.Metadata.Tags))
		copy(newConfig.Metadata.Tags, o.config.Metadata.Tags)
	}

	// Deep copy properties map
	if o.config.Properties != nil {
		newConfig.Properties = make(map[string]core.Schema, len(o.config.Properties))
		for k, v := range o.config.Properties {
			newConfig.Properties[k] = v // Schemas should be immutable
		}
	}

	// Deep copy required slice
	if o.config.Required != nil {
		newConfig.Required = make([]string, len(o.config.Required))
		copy(newConfig.Required, o.config.Required)
	}

	// Deep copy pattern properties
	if o.config.PatternProperties != nil {
		newConfig.PatternProperties = make(map[string]core.Schema, len(o.config.PatternProperties))
		for k, v := range o.config.PatternProperties {
			newConfig.PatternProperties[k] = v
		}
	}

	// Deep copy property dependencies
	if o.config.PropertyDependencies != nil {
		newConfig.PropertyDependencies = make(map[string][]string, len(o.config.PropertyDependencies))
		for k, v := range o.config.PropertyDependencies {
			deps := make([]string, len(v))
			copy(deps, v)
			newConfig.PropertyDependencies[k] = deps
		}
	}

	// Deep copy default value
	if o.config.DefaultVal != nil {
		newConfig.DefaultVal = make(map[string]any, len(o.config.DefaultVal))
		for k, v := range o.config.DefaultVal {
			newConfig.DefaultVal[k] = v
		}
	}

	return NewObjectSchema(newConfig)
}

// Properties returns the property schemas.
func (o *ObjectSchema) Properties() map[string]core.Schema {
	if o.config.Properties == nil {
		return make(map[string]core.Schema)
	}
	// Return a copy to maintain immutability
	result := make(map[string]core.Schema, len(o.config.Properties))
	for k, v := range o.config.Properties {
		result[k] = v
	}
	return result
}

// Required returns the list of required property names.
func (o *ObjectSchema) Required() []string {
	if o.config.Required == nil {
		return []string{}
	}
	// Return a copy to maintain immutability
	result := make([]string, len(o.config.Required))
	copy(result, o.config.Required)
	return result
}

// AdditionalProperties returns whether additional properties are allowed.
func (o *ObjectSchema) AdditionalProperties() bool {
	return o.config.AdditionalProperties
}

// MinProperties returns the minimum properties constraint.
func (o *ObjectSchema) MinProperties() *int {
	return o.config.MinProperties
}

// MaxProperties returns the maximum properties constraint.
func (o *ObjectSchema) MaxProperties() *int {
	return o.config.MaxProperties
}

// PatternProperties returns the pattern properties map.
func (o *ObjectSchema) PatternProperties() map[string]core.Schema {
	if o.config.PatternProperties == nil {
		return make(map[string]core.Schema)
	}
	// Return a copy to maintain immutability
	result := make(map[string]core.Schema, len(o.config.PatternProperties))
	for k, v := range o.config.PatternProperties {
		result[k] = v
	}
	return result
}

// PropertyDependencies returns the property dependencies map.
func (o *ObjectSchema) PropertyDependencies() map[string][]string {
	if o.config.PropertyDependencies == nil {
		return make(map[string][]string)
	}
	// Return a copy to maintain immutability
	result := make(map[string][]string, len(o.config.PropertyDependencies))
	for k, v := range o.config.PropertyDependencies {
		deps := make([]string, len(v))
		copy(deps, v)
		result[k] = deps
	}
	return result
}

// DefaultValue returns the default value.
func (o *ObjectSchema) DefaultValue() map[string]any {
	if o.config.DefaultVal == nil {
		return nil
	}
	result := make(map[string]any, len(o.config.DefaultVal))
	for k, v := range o.config.DefaultVal {
		result[k] = v
	}
	return result
}

// Note: Validation moved to consumer-driven architecture.
// Use schema/consumer.Registry.ProcessValueWithPurpose("validation", schema, value) instead.

// convertToMap converts various object-like types to map[string]any.
func (o *ObjectSchema) convertToMap(value any) (map[string]any, bool) {
	if value == nil {
		return nil, false
	}

	// Direct map of string to any
	if m, ok := value.(map[string]any); ok {
		return m, true
	}

	// Use reflection for other map types and structs
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Map:
		// Handle other map types (map[string]any, etc.)
		result := make(map[string]any)
		for _, key := range rv.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = rv.MapIndex(key).Interface()
		}
		return result, true

	case reflect.Struct:
		// Handle structs by converting to map
		result := make(map[string]any)
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.IsExported() {
				fieldValue := rv.Field(i)
				if fieldValue.CanInterface() {
					// Use json tag if available, otherwise field name
					fieldName := field.Name
					if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
						// Simple json tag parsing (could be enhanced)
						if commaIdx := len(jsonTag); commaIdx > 0 {
							fieldName = jsonTag
							if commaIdx := 0; jsonTag[commaIdx:commaIdx+1] == "," {
								fieldName = jsonTag[:commaIdx]
							}
						}
					}
					result[fieldName] = fieldValue.Interface()
				}
			}
		}
		return result, true

	default:
		return nil, false
	}
}

// buildPropertyPath constructs the error path for nested properties.
func (o *ObjectSchema) buildPropertyPath(prefix, propName string) string {
	if prefix == "" {
		return propName
	}
	if propName == "" {
		return prefix
	}
	return prefix + "." + propName
}

// matchesPattern checks if a property name matches a pattern.
// This is a simple implementation - could be enhanced with regex support.
func (o *ObjectSchema) matchesPattern(propName, pattern string) bool {
	// Simple wildcard pattern matching for now
	if pattern == "*" {
		return true
	}
	// Could add regex matching here: regexp.MustCompile(pattern).MatchString(propName)
	return propName == pattern
}

// Accept implements the visitor pattern for schema traversal.
func (o *ObjectSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitObject(o)
}
