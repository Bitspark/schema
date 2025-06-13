package validation

import (
	"fmt"
	"reflect"

	"defs.dev/schema/consumer"
	"defs.dev/schema/core"
)

// ObjectValidationConsumer validates object values
type ObjectValidationConsumer struct{}

func (c *ObjectValidationConsumer) Name() string {
	return "object_validator"
}

func (c *ObjectValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *ObjectValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeStructure)
}

func (c *ObjectValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual object value
	actualValue := value.Value()

	// Check if it's an object (map or struct)
	var objectMap map[string]any
	switch v := actualValue.(type) {
	case map[string]any:
		objectMap = v
	case nil:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "Expected object or map",
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	default:
		// Try to handle structs by converting to map using reflection
		converted, ok := c.convertToMap(actualValue)
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: "Expected object or map",
				Code:    "type_mismatch",
			})
			return consumer.NewResult("validation", result), nil
		}
		objectMap = converted
	}

	// Cast to ObjectSchema to access properties
	objectSchema, ok := ctx.Schema.(core.ObjectSchema)
	if !ok {
		// Fallback validation - just check it's an object
		return consumer.NewResult("validation", result), nil
	}

	// Validate required properties
	required := objectSchema.Required()
	for _, requiredProp := range required {
		if _, exists := objectMap[requiredProp]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    append(ctx.Path, requiredProp),
				Message: fmt.Sprintf("Missing required property '%s'", requiredProp),
				Code:    "missing_required_property",
			})
		}
	}

	// Validate properties against their schemas
	properties := objectSchema.Properties()
	for propName, propValue := range objectMap {
		propSchema, exists := properties[propName]
		if !exists {
			// Check if additional properties are allowed
			if !objectSchema.AdditionalProperties() {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    append(ctx.Path, propName),
					Message: fmt.Sprintf("Additional property '%s' is not allowed", propName),
					Code:    "additional_property_not_allowed",
				})
			}
			continue
		}

		// Validate the property value against its schema using recursive validation
		propPath := append(ctx.Path, propName)
		propResult := ValidateWithRegistry(propSchema, propValue)
		if !propResult.Valid {
			result.Valid = false
			// Add path context to property errors
			for _, err := range propResult.Errors {
				err.Path = append(propPath, err.Path...)
				result.Errors = append(result.Errors, err)
			}
		}
	}

	return consumer.NewResult("validation", result), nil
}

// convertToMap converts various object-like types to map[string]any.
func (c *ObjectValidationConsumer) convertToMap(value any) (map[string]any, bool) {
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
						// Simple json tag parsing
						for commaIdx := 0; commaIdx < len(jsonTag); commaIdx++ {
							if jsonTag[commaIdx] == ',' {
								fieldName = jsonTag[:commaIdx]
								break
							}
						}
						if fieldName == field.Name {
							fieldName = jsonTag
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

func (c *ObjectValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "object_validator",
		Purpose:      "validation",
		Description:  "Validates object values against object schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "object", "properties"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
