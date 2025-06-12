package schemas

import (
	"fmt"
	"reflect"
	"sort"

	"defs.dev/schema/api/core"
)

// ObjectSchemaConfig holds the configuration for building an ObjectSchema.
type ObjectSchemaConfig struct {
	Metadata             core.SchemaMetadata
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
	return core.TypeObject
}

// Metadata returns the schema metadata.
func (o *ObjectSchema) Metadata() core.SchemaMetadata {
	return o.config.Metadata
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

// Validate validates a value against the object schema.
func (o *ObjectSchema) Validate(value any) core.ValidationResult {
	return o.validateWithPath(value, "")
}

// validateWithPath validates a value against the object schema with a given path prefix.
func (o *ObjectSchema) validateWithPath(value any, pathPrefix string) core.ValidationResult {
	// Convert to object/map
	objectValue, ok := o.convertToMap(value)
	if !ok {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{{
				Path:       pathPrefix,
				Message:    "Expected object or map",
				Code:       "type_mismatch",
				Value:      value,
				Expected:   "object or map",
				Suggestion: "Provide an object or map value",
			}},
		}
	}

	var errors []core.ValidationError

	// Property count validation
	propertyCount := len(objectValue)

	if o.config.MinProperties != nil && propertyCount < *o.config.MinProperties {
		errors = append(errors, core.ValidationError{
			Path:       pathPrefix,
			Message:    fmt.Sprintf("Object has too few properties (minimum %d)", *o.config.MinProperties),
			Code:       "min_properties",
			Value:      propertyCount,
			Expected:   fmt.Sprintf("≥ %d properties", *o.config.MinProperties),
			Suggestion: fmt.Sprintf("Add at least %d properties", *o.config.MinProperties-propertyCount),
		})
	}

	if o.config.MaxProperties != nil && propertyCount > *o.config.MaxProperties {
		errors = append(errors, core.ValidationError{
			Path:       pathPrefix,
			Message:    fmt.Sprintf("Object has too many properties (maximum %d)", *o.config.MaxProperties),
			Code:       "max_properties",
			Value:      propertyCount,
			Expected:   fmt.Sprintf("≤ %d properties", *o.config.MaxProperties),
			Suggestion: fmt.Sprintf("Remove %d properties", propertyCount-*o.config.MaxProperties),
		})
	}

	// Required properties validation
	for _, reqProp := range o.config.Required {
		if _, exists := objectValue[reqProp]; !exists {
			errors = append(errors, core.ValidationError{
				Path:       o.buildPropertyPath(pathPrefix, reqProp),
				Message:    fmt.Sprintf("Missing required property '%s'", reqProp),
				Code:       "required_property",
				Value:      objectValue,
				Expected:   fmt.Sprintf("property '%s' to be present", reqProp),
				Suggestion: fmt.Sprintf("Add the required property '%s'", reqProp),
			})
		}
	}

	// Property schema validation
	for propName, propValue := range objectValue {
		if propSchema, exists := o.config.Properties[propName]; exists {
			// Validate against defined property schema
			propResult := propSchema.Validate(propValue)
			if !propResult.Valid {
				for _, propError := range propResult.Errors {
					errors = append(errors, core.ValidationError{
						Path:       o.buildPropertyPath(pathPrefix, o.buildPropertyPath(propName, propError.Path)),
						Message:    propError.Message,
						Code:       propError.Code,
						Value:      propError.Value,
						Expected:   propError.Expected,
						Suggestion: propError.Suggestion,
						Context:    fmt.Sprintf("Property '%s'", propName),
					})
				}
			}
		} else {
			// Check pattern properties
			matched := false
			if o.config.PatternProperties != nil {
				for pattern, patternSchema := range o.config.PatternProperties {
					// Simple pattern matching (could be enhanced with regex)
					if o.matchesPattern(propName, pattern) {
						matched = true
						propResult := patternSchema.Validate(propValue)
						if !propResult.Valid {
							for _, propError := range propResult.Errors {
								errors = append(errors, core.ValidationError{
									Path:       o.buildPropertyPath(pathPrefix, o.buildPropertyPath(propName, propError.Path)),
									Message:    propError.Message,
									Code:       propError.Code,
									Value:      propError.Value,
									Expected:   propError.Expected,
									Suggestion: propError.Suggestion,
									Context:    fmt.Sprintf("Pattern property '%s' (pattern: %s)", propName, pattern),
								})
							}
						}
						break
					}
				}
			}

			// Check additional properties
			if !matched && !o.config.AdditionalProperties {
				errors = append(errors, core.ValidationError{
					Path:       o.buildPropertyPath(pathPrefix, propName),
					Message:    fmt.Sprintf("Additional property '%s' is not allowed", propName),
					Code:       "additional_property",
					Value:      propValue,
					Expected:   "only defined properties",
					Suggestion: fmt.Sprintf("Remove the property '%s' or define it in the schema", propName),
				})
			}
		}
	}

	// Property dependencies validation
	for propName, dependencies := range o.config.PropertyDependencies {
		if _, exists := objectValue[propName]; exists {
			for _, dep := range dependencies {
				if _, depExists := objectValue[dep]; !depExists {
					errors = append(errors, core.ValidationError{
						Path:       o.buildPropertyPath(pathPrefix, dep),
						Message:    fmt.Sprintf("Property '%s' requires property '%s'", propName, dep),
						Code:       "property_dependency",
						Value:      objectValue,
						Expected:   fmt.Sprintf("property '%s' when '%s' is present", dep, propName),
						Suggestion: fmt.Sprintf("Add the required dependency property '%s'", dep),
					})
				}
			}
		}
	}

	return core.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

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

// GenerateExample generates an example value for the object schema.
func (o *ObjectSchema) GenerateExample() any {
	// Use provided examples if available
	if len(o.config.Metadata.Examples) > 0 {
		return o.config.Metadata.Examples[0]
	}

	// Use default value if set
	if o.config.DefaultVal != nil {
		return o.config.DefaultVal
	}

	// Generate based on properties
	result := make(map[string]any)

	// Add required properties first
	for _, reqProp := range o.config.Required {
		if propSchema, exists := o.config.Properties[reqProp]; exists {
			result[reqProp] = propSchema.GenerateExample()
		} else {
			// Generate a basic example for unknown required property
			result[reqProp] = fmt.Sprintf("example_%s", reqProp)
		}
	}

	// Add some optional properties (limit to reasonable number)
	optionalCount := 0
	maxOptional := 3

	// Sort property names for consistent example generation
	var propNames []string
	for propName := range o.config.Properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	for _, propName := range propNames {
		// Skip if already added as required
		if _, exists := result[propName]; exists {
			continue
		}

		// Add some optional properties but not all
		if optionalCount < maxOptional {
			propSchema := o.config.Properties[propName]
			result[propName] = propSchema.GenerateExample()
			optionalCount++
		}
	}

	// If no properties defined, create a simple example
	if len(result) == 0 {
		result["example_property"] = "example_value"
	}

	return result
}

// Accept implements the visitor pattern for schema traversal.
func (o *ObjectSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitObject(o)
}
