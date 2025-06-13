// Package validation provides validation consumers that implement the consumer framework.
package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"defs.dev/schema/api/core"
	"defs.dev/schema/consumer"
)

// StringValidationConsumer validates string values against string schema constraints
type StringValidationConsumer struct{}

func (c *StringValidationConsumer) Name() string {
	return "string_validator"
}

func (c *StringValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *StringValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeString)
}

func (c *StringValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual string value
	actualValue := value.Value()
	str, ok := actualValue.(string)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected string, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Cast to StringSchema to access properties
	stringSchema, ok := ctx.Schema.(core.StringSchema)
	if !ok {
		// Fallback to annotation-based validation
		for _, annotation := range ctx.Schema.Annotations() {
			switch annotation.Name() {
			case "format":
				if err := c.validateFormat(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			case "pattern":
				if err := c.validatePattern(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			case "minLength":
				if err := c.validateMinLength(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			case "maxLength":
				if err := c.validateMaxLength(str, annotation.Value(), ctx.Path); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, *err)
				}
			}
		}
		return consumer.NewResult("validation", result), nil
	}

	// Validate using schema properties
	if minLen := stringSchema.MinLength(); minLen != nil {
		if err := c.validateMinLength(str, *minLen, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	if maxLen := stringSchema.MaxLength(); maxLen != nil {
		if err := c.validateMaxLength(str, *maxLen, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	if pattern := stringSchema.Pattern(); pattern != "" {
		if err := c.validatePattern(str, pattern, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	if format := stringSchema.Format(); format != "" {
		if err := c.validateFormat(str, format, ctx.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	// Validate enum values
	if enumValues := stringSchema.EnumValues(); len(enumValues) > 0 {
		found := false
		for _, enumValue := range enumValues {
			if str == enumValue {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Code:    "enum_mismatch",
				Message: fmt.Sprintf("value '%s' is not one of the allowed values", str),
			})
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *StringValidationConsumer) validateFormat(value string, format any, path []string) *ValidationIssue {
	formatStr, ok := format.(string)
	if !ok {
		return &ValidationIssue{
			Path:    path,
			Code:    "invalid_format_annotation",
			Message: "format annotation must be a string",
		}
	}

	switch formatStr {
	case "email":
		if _, err := mail.ParseAddress(value); err != nil {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_email",
				Message: "invalid email format",
			}
		}
	case "url":
		if _, err := url.Parse(value); err != nil {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_url",
				Message: "invalid URL format",
			}
		}
	case "uuid":
		uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
		if matched, _ := regexp.MatchString(uuidPattern, strings.ToLower(value)); !matched {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_uuid",
				Message: "invalid UUID format",
			}
		}
	}
	return nil
}

func (c *StringValidationConsumer) validatePattern(value string, pattern any, path []string) *ValidationIssue {
	patternStr, ok := pattern.(string)
	if !ok {
		return &ValidationIssue{
			Path:    path,
			Code:    "invalid_pattern_annotation",
			Message: "pattern annotation must be a string",
		}
	}

	if matched, err := regexp.MatchString(patternStr, value); err != nil {
		return &ValidationIssue{
			Path:    path,
			Code:    "invalid_regex",
			Message: "invalid regular expression: " + err.Error(),
		}
	} else if !matched {
		return &ValidationIssue{
			Path:    path,
			Code:    "pattern_mismatch",
			Message: fmt.Sprintf("value does not match pattern: %s", patternStr),
		}
	}
	return nil
}

func (c *StringValidationConsumer) validateMinLength(value string, minLen any, path []string) *ValidationIssue {
	min, ok := minLen.(int)
	if !ok {
		if minFloat, ok := minLen.(float64); ok {
			min = int(minFloat)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_minlength_annotation",
				Message: "minLength annotation must be a number",
			}
		}
	}

	if len(value) < min {
		return &ValidationIssue{
			Path:    path,
			Code:    "string_too_short",
			Message: fmt.Sprintf("string length %d is less than minimum %d", len(value), min),
		}
	}
	return nil
}

func (c *StringValidationConsumer) validateMaxLength(value string, maxLen any, path []string) *ValidationIssue {
	max, ok := maxLen.(int)
	if !ok {
		if maxFloat, ok := maxLen.(float64); ok {
			max = int(maxFloat)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_maxlength_annotation",
				Message: "maxLength annotation must be a number",
			}
		}
	}

	if len(value) > max {
		return &ValidationIssue{
			Path:    path,
			Code:    "string_too_long",
			Message: fmt.Sprintf("string length %d exceeds maximum %d", len(value), max),
		}
	}
	return nil
}

func (c *StringValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "string_validator",
		Purpose:      "validation",
		Description:  "Validates string values against string schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "string", "constraints"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}

// BooleanValidationConsumer validates boolean values
type BooleanValidationConsumer struct{}

func (c *BooleanValidationConsumer) Name() string {
	return "boolean_validator"
}

func (c *BooleanValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *BooleanValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeBoolean)
}

func (c *BooleanValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual value
	actualValue := value.Value()

	// Try to cast to BooleanSchema to check for string conversion
	if _, ok := ctx.Schema.(core.BooleanSchema); ok {
		// Check if it's a boolean first
		if _, ok := actualValue.(bool); ok {
			return consumer.NewResult("validation", result), nil
		}

		// If we get here, it's not a boolean
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected boolean, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Fallback: just check if it's a boolean
	_, ok := actualValue.(bool)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected boolean, got %T", actualValue),
			Code:    "type_mismatch",
		})
	}

	return consumer.NewResult("validation", result), nil
}

func (c *BooleanValidationConsumer) convertStringToBool(str string, caseInsensitive bool) (bool, error) {
	// Always check exact matches first
	switch str {
	case "true", "1", "yes", "on", "y", "t":
		return true, nil
	case "false", "0", "no", "off", "n", "f":
		return false, nil
	}

	// For string conversion, be reasonably permissive with common case variations
	// even if CaseInsensitive is not explicitly enabled
	lowerStr := strings.ToLower(str)
	switch lowerStr {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	// If case insensitive is enabled, also check additional patterns
	if caseInsensitive {
		switch lowerStr {
		case "yes", "on", "y", "t", "1":
			return true, nil
		case "no", "off", "n", "f", "0":
			return false, nil
		}
	}

	return false, fmt.Errorf("invalid boolean string: %s", str)
}

func (c *BooleanValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "boolean_validator",
		Purpose:      "validation",
		Description:  "Validates boolean values against boolean schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "boolean"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}

// NumberValidationConsumer validates numeric values
type NumberValidationConsumer struct{}

func (c *NumberValidationConsumer) Name() string {
	return "number_validator"
}

func (c *NumberValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *NumberValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Or(consumer.Type(core.TypeNumber), consumer.Type(core.TypeInteger))
}

func (c *NumberValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual numeric value
	actualValue := value.Value()
	var numValue float64
	var ok bool

	switch v := actualValue.(type) {
	case int:
		numValue = float64(v)
		ok = true
	case int8:
		numValue = float64(v)
		ok = true
	case int16:
		numValue = float64(v)
		ok = true
	case int32:
		numValue = float64(v)
		ok = true
	case int64:
		numValue = float64(v)
		ok = true
	case uint:
		numValue = float64(v)
		ok = true
	case uint8:
		numValue = float64(v)
		ok = true
	case uint16:
		numValue = float64(v)
		ok = true
	case uint32:
		numValue = float64(v)
		ok = true
	case uint64:
		numValue = float64(v)
		ok = true
	case float64:
		numValue = v
		ok = true
	case float32:
		numValue = float64(v)
		ok = true
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected number, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "invalid numeric value",
			Code:    "invalid_number",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Check for special float values (NaN, Inf)
	if numValue != numValue { // NaN check
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "NaN is not allowed",
			Code:    "invalid_number",
		})
		return consumer.NewResult("validation", result), nil
	}

	if numValue == numValue+1 { // Infinity check
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: "Infinity is not allowed",
			Code:    "invalid_number",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Try to cast to NumberSchema first
	if numberSchema, ok := ctx.Schema.(core.NumberSchema); ok {
		if min := numberSchema.Minimum(); min != nil {
			if err := c.validateMin(numValue, *min, ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}

		if max := numberSchema.Maximum(); max != nil {
			if err := c.validateMax(numValue, *max, ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}
		return consumer.NewResult("validation", result), nil
	}

	// Try to cast to IntegerSchema
	if integerSchema, ok := ctx.Schema.(core.IntegerSchema); ok {
		// Check if it's actually an integer
		if float64(int64(numValue)) != numValue {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: "expected integer value",
				Code:    "not_integer",
			})
			return consumer.NewResult("validation", result), nil
		}

		if min := integerSchema.Minimum(); min != nil {
			if err := c.validateMin(numValue, float64(*min), ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}

		if max := integerSchema.Maximum(); max != nil {
			if err := c.validateMax(numValue, float64(*max), ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}
		return consumer.NewResult("validation", result), nil
	}

	// Fallback to annotation-based validation
	for _, annotation := range ctx.Schema.Annotations() {
		switch annotation.Name() {
		case "min":
			if err := c.validateMin(numValue, annotation.Value(), ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		case "max":
			if err := c.validateMax(numValue, annotation.Value(), ctx.Path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *NumberValidationConsumer) validateMin(value float64, min any, path []string) *ValidationIssue {
	minVal, ok := min.(float64)
	if !ok {
		if minInt, ok := min.(int); ok {
			minVal = float64(minInt)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_min_annotation",
				Message: "min annotation must be a number",
			}
		}
	}

	if value < minVal {
		return &ValidationIssue{
			Path:    path,
			Code:    "number_too_small",
			Message: fmt.Sprintf("value %g is less than minimum %g", value, minVal),
		}
	}
	return nil
}

func (c *NumberValidationConsumer) validateMax(value float64, max any, path []string) *ValidationIssue {
	maxVal, ok := max.(float64)
	if !ok {
		if maxInt, ok := max.(int); ok {
			maxVal = float64(maxInt)
		} else {
			return &ValidationIssue{
				Path:    path,
				Code:    "invalid_max_annotation",
				Message: "max annotation must be a number",
			}
		}
	}

	if value > maxVal {
		return &ValidationIssue{
			Path:    path,
			Code:    "number_too_large",
			Message: fmt.Sprintf("value %g exceeds maximum %g", value, maxVal),
		}
	}
	return nil
}

func (c *NumberValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "number_validator",
		Purpose:      "validation",
		Description:  "Validates numeric values against numeric schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "number", "constraints"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}

// FunctionValidationConsumer validates function input/output values
type FunctionValidationConsumer struct{}

func (c *FunctionValidationConsumer) Name() string {
	return "function_validator"
}

func (c *FunctionValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *FunctionValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeFunction)
}

func (c *FunctionValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	functionSchema, ok := ctx.Schema.(core.FunctionSchema)
	if !ok {
		return nil, fmt.Errorf("expected function schema, got %T", ctx.Schema)
	}

	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual function input value (should be a map)
	actualValue := value.Value()
	inputMap, ok := actualValue.(map[string]any)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected function input map, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Validate required inputs
	inputs := functionSchema.Inputs()
	for _, input := range inputs.Args() {
		if !input.Optional() {
			if _, exists := inputMap[input.Name()]; !exists {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    append(ctx.Path, input.Name()),
					Message: fmt.Sprintf("required input '%s' is missing", input.Name()),
					Code:    "missing_required_input",
				})
			}
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *FunctionValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "function_validator",
		Purpose:      "validation",
		Description:  "Validates function input/output values against function schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "function", "inputs", "outputs"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}

// ArrayValidationConsumer validates array values
type ArrayValidationConsumer struct{}

func (c *ArrayValidationConsumer) Name() string {
	return "array_validator"
}

func (c *ArrayValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *ArrayValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeArray)
}

func (c *ArrayValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// Get the actual array value
	actualValue := value.Value()

	// Check if it's an array/slice
	var arrayItems []any
	switch v := actualValue.(type) {
	case []any:
		arrayItems = v
	case []string:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	case []int:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	case []float64:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	case []bool:
		arrayItems = make([]any, len(v))
		for i, item := range v {
			arrayItems[i] = item
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected array, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
	}

	// Cast to ArraySchema to access properties
	arraySchema, ok := ctx.Schema.(core.ArraySchema)
	if !ok {
		// Fallback validation - just check it's an array
		return consumer.NewResult("validation", result), nil
	}

	// Validate array constraints
	if minItems := arraySchema.MinItems(); minItems != nil {
		if len(arrayItems) < *minItems {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("array has %d items, minimum required is %d", len(arrayItems), *minItems),
				Code:    "min_items_violation",
			})
		}
	}

	if maxItems := arraySchema.MaxItems(); maxItems != nil {
		if len(arrayItems) > *maxItems {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("array has %d items, maximum allowed is %d", len(arrayItems), *maxItems),
				Code:    "max_items_violation",
			})
		}
	}

	// Validate unique items if required
	if arraySchema.UniqueItemsRequired() {
		seen := make(map[string]bool)
		for i, item := range arrayItems {
			itemStr := fmt.Sprintf("%v", item)
			if seen[itemStr] {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationIssue{
					Path:    append(ctx.Path, fmt.Sprintf("[%d]", i)),
					Message: fmt.Sprintf("duplicate item found: %v", item),
					Code:    "unique_items_violation",
				})
			}
			seen[itemStr] = true
		}
	}

	// Validate each item against the item schema
	if itemSchema := arraySchema.ItemSchema(); itemSchema != nil {
		for i, item := range arrayItems {
			itemPath := append(ctx.Path, fmt.Sprintf("[%d]", i))

			// Use recursive validation for the item
			itemResult := ValidateWithRegistry(itemSchema, item)
			if !itemResult.Valid {
				result.Valid = false
				// Add path context to item errors
				for _, err := range itemResult.Errors {
					err.Path = itemPath
					result.Errors = append(result.Errors, err)
				}
			}
		}
	}

	// Validate contains constraint
	if containsSchema := arraySchema.ContainsSchema(); containsSchema != nil {
		containsMatched := false
		for _, item := range arrayItems {
			itemResult := ValidateWithRegistry(containsSchema, item)
			if itemResult.Valid {
				containsMatched = true
				break
			}
		}
		if !containsMatched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: "array does not contain any item matching the contains schema",
				Code:    "contains_constraint_violation",
			})
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *ArrayValidationConsumer) validateItem(ctx consumer.ProcessingContext, value core.Value[any]) ValidationResult {
	// Use recursive validation instead of simplified validation
	actualValue := value.Value()
	return ValidateWithRegistry(ctx.Schema, actualValue)
}

func (c *ArrayValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "array_validator",
		Purpose:      "validation",
		Description:  "Validates array values against array schema constraints",
		Version:      "1.0.0",
		Tags:         []string{"validation", "array", "constraints"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}

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
	default:
		// Try to handle structs by converting to map using reflection
		// For now, we'll just accept any non-primitive type as an object
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Path:    ctx.Path,
			Message: fmt.Sprintf("expected object, got %T", actualValue),
			Code:    "type_mismatch",
		})
		return consumer.NewResult("validation", result), nil
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
				Message: fmt.Sprintf("missing required property '%s'", requiredProp),
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
					Message: fmt.Sprintf("additional property '%s' is not allowed", propName),
					Code:    "additional_property_not_allowed",
				})
			}
			continue
		}

		// Validate the property value against its schema
		propPath := append(ctx.Path, propName)
		propValueObj := &simpleValue{value: propValue}
		propCtx := consumer.ProcessingContext{
			Schema: propSchema,
			Path:   propPath,
		}

		// Recursively validate the property
		propResult := c.validateProperty(propCtx, propValueObj)
		if !propResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, propResult.Errors...)
		}
	}

	return consumer.NewResult("validation", result), nil
}

func (c *ObjectValidationConsumer) validateProperty(ctx consumer.ProcessingContext, value core.Value[any]) ValidationResult {
	// This is a simplified property validation - in a full implementation,
	// we would recursively call the validation system for the property
	result := ValidationResult{Valid: true, Errors: []ValidationIssue{}}

	// Basic type checking based on schema type
	actualValue := value.Value()
	schemaType := ctx.Schema.Type()

	switch schemaType {
	case core.TypeString:
		if _, ok := actualValue.(string); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("expected string, got %T", actualValue),
				Code:    "type_mismatch",
			})
		}
	case core.TypeNumber:
		switch actualValue.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid numeric types
		default:
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("expected number, got %T", actualValue),
				Code:    "type_mismatch",
			})
		}
	case core.TypeBoolean:
		if _, ok := actualValue.(bool); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Path:    ctx.Path,
				Message: fmt.Sprintf("expected boolean, got %T", actualValue),
				Code:    "type_mismatch",
			})
		}
	}

	return result
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

// ServiceValidationConsumer validates service schemas
type ServiceValidationConsumer struct{}

func (c *ServiceValidationConsumer) Name() string {
	return "service_validator"
}

func (c *ServiceValidationConsumer) Purpose() consumer.ConsumerPurpose {
	return "validation"
}

func (c *ServiceValidationConsumer) ApplicableSchemas() consumer.SchemaCondition {
	return consumer.Type(core.TypeService)
}

func (c *ServiceValidationConsumer) ProcessValue(ctx consumer.ProcessingContext, value core.Value[any]) (consumer.ConsumerResult, error) {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationIssue{},
	}

	// For service schemas, we typically validate service metadata or configuration
	// This is a basic implementation that accepts any value for service schemas
	return consumer.NewResult("validation", result), nil
}

func (c *ServiceValidationConsumer) Metadata() consumer.ConsumerMetadata {
	return consumer.ConsumerMetadata{
		Name:         "service_validator",
		Purpose:      "validation",
		Description:  "Validates service schema definitions",
		Version:      "1.0.0",
		Tags:         []string{"validation", "service"},
		ResultKind:   "validation",
		ResultGoType: "*validation.ValidationResult",
	}
}
