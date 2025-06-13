package registry

import (
	"defs.dev/schema/annotation"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BaseValidator provides common functionality for all validators.
type BaseValidator struct {
	name                 string
	supportedAnnotations []string
	metadata             ValidatorMetadata
}

func (b *BaseValidator) Name() string                   { return b.name }
func (b *BaseValidator) SupportedAnnotations() []string { return b.supportedAnnotations }
func (b *BaseValidator) Metadata() ValidatorMetadata    { return b.metadata }

// EmailValidator validates email addresses.
type EmailValidator struct {
	BaseValidator
}

func NewEmailValidator() *EmailValidator {
	return &EmailValidator{
		BaseValidator: BaseValidator{
			name:                 "email",
			supportedAnnotations: []string{"format"},
			metadata: ValidatorMetadata{
				Name:           "email",
				Description:    "Validates email address format",
				Category:       "format",
				Tags:           []string{"string", "format", "validation"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *EmailValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	_, err := mail.ParseAddress(str)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeInvalidFormat, fmt.Sprintf("invalid email format: %v", err)))
	}

	return ValidResult()
}

func (v *EmailValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "format" && ann.Value() == "email" {
			return v.Validate(value)
		}
	}
	return ValidResult()
}

func (v *EmailValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil // No configuration needed for email validator
}

// URLValidator validates URL format.
type URLValidator struct {
	BaseValidator
}

func NewURLValidator() *URLValidator {
	return &URLValidator{
		BaseValidator: BaseValidator{
			name:                 "url",
			supportedAnnotations: []string{"format"},
			metadata: ValidatorMetadata{
				Name:           "url",
				Description:    "Validates URL format",
				Category:       "format",
				Tags:           []string{"string", "format", "url"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *URLValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	_, err := url.Parse(str)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeInvalidFormat, fmt.Sprintf("invalid URL format: %v", err)))
	}

	return ValidResult()
}

func (v *URLValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "format" && ann.Value() == "url" {
			return v.Validate(value)
		}
	}
	return ValidResult()
}

func (v *URLValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil
}

// UUIDValidator validates UUID format.
type UUIDValidator struct {
	BaseValidator
}

func NewUUIDValidator() *UUIDValidator {
	return &UUIDValidator{
		BaseValidator: BaseValidator{
			name:                 "uuid",
			supportedAnnotations: []string{"format"},
			metadata: ValidatorMetadata{
				Name:           "uuid",
				Description:    "Validates UUID format",
				Category:       "format",
				Tags:           []string{"string", "format", "uuid"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *UUIDValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	// UUID regex pattern
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(str) {
		return InvalidResult(NewValidationError(v.name, ErrorCodeInvalidFormat, "invalid UUID format"))
	}

	return ValidResult()
}

func (v *UUIDValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "format" && ann.Value() == "uuid" {
			return v.Validate(value)
		}
	}
	return ValidResult()
}

func (v *UUIDValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil
}

// DateValidator validates date format.
type DateValidator struct {
	BaseValidator
}

func NewDateValidator() *DateValidator {
	return &DateValidator{
		BaseValidator: BaseValidator{
			name:                 "date",
			supportedAnnotations: []string{"format"},
			metadata: ValidatorMetadata{
				Name:           "date",
				Description:    "Validates date format (YYYY-MM-DD)",
				Category:       "format",
				Tags:           []string{"string", "format", "date"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *DateValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	_, err := time.Parse("2006-01-02", str)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeInvalidFormat, fmt.Sprintf("invalid date format, expected YYYY-MM-DD: %v", err)))
	}

	return ValidResult()
}

func (v *DateValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "format" && ann.Value() == "date" {
			return v.Validate(value)
		}
	}
	return ValidResult()
}

func (v *DateValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil
}

// TimeValidator validates time format.
type TimeValidator struct {
	BaseValidator
}

func NewTimeValidator() *TimeValidator {
	return &TimeValidator{
		BaseValidator: BaseValidator{
			name:                 "time",
			supportedAnnotations: []string{"format"},
			metadata: ValidatorMetadata{
				Name:           "time",
				Description:    "Validates time format (HH:MM:SS)",
				Category:       "format",
				Tags:           []string{"string", "format", "time"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *TimeValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	_, err := time.Parse("15:04:05", str)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeInvalidFormat, fmt.Sprintf("invalid time format, expected HH:MM:SS: %v", err)))
	}

	return ValidResult()
}

func (v *TimeValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "format" && ann.Value() == "time" {
			return v.Validate(value)
		}
	}
	return ValidResult()
}

func (v *TimeValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil
}

// DateTimeValidator validates datetime format.
type DateTimeValidator struct {
	BaseValidator
}

func NewDateTimeValidator() *DateTimeValidator {
	return &DateTimeValidator{
		BaseValidator: BaseValidator{
			name:                 "datetime",
			supportedAnnotations: []string{"format"},
			metadata: ValidatorMetadata{
				Name:           "datetime",
				Description:    "Validates datetime format (RFC3339)",
				Category:       "format",
				Tags:           []string{"string", "format", "datetime"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *DateTimeValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	_, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeInvalidFormat, fmt.Sprintf("invalid datetime format, expected RFC3339: %v", err)))
	}

	return ValidResult()
}

func (v *DateTimeValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "format" && ann.Value() == "datetime" {
			return v.Validate(value)
		}
	}
	return ValidResult()
}

func (v *DateTimeValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil
}

// PatternValidator validates regex patterns.
type PatternValidator struct {
	BaseValidator
	pattern *regexp.Regexp
}

func NewPatternValidator() *PatternValidator {
	return &PatternValidator{
		BaseValidator: BaseValidator{
			name:                 "pattern",
			supportedAnnotations: []string{"pattern"},
			metadata: ValidatorMetadata{
				Name:           "pattern",
				Description:    "Validates string against regex pattern",
				Category:       "pattern",
				Tags:           []string{"string", "pattern", "regex"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *PatternValidator) Validate(value any) ValidationResult {
	if v.pattern == nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeValidationFailed, "no pattern configured"))
	}

	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	if !v.pattern.MatchString(str) {
		return InvalidResult(NewValidationError(v.name, ErrorCodePatternMismatch, fmt.Sprintf("value does not match pattern %s", v.pattern.String())))
	}

	return ValidResult()
}

func (v *PatternValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "pattern" {
			patternStr, ok := ann.Value().(string)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "pattern value must be a string"))
			}

			pattern, err := regexp.Compile(patternStr)
			if err != nil {
				return InvalidResult(NewValidationError(v.name, ErrorCodeValidationFailed, fmt.Sprintf("invalid regex pattern: %v", err)))
			}

			str, ok := value.(string)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
			}

			if !pattern.MatchString(str) {
				return InvalidResult(NewValidationError(v.name, ErrorCodePatternMismatch, fmt.Sprintf("value does not match pattern %s", patternStr)))
			}
		}
	}
	return ValidResult()
}

func (v *PatternValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "pattern" {
			patternStr, ok := ann.Value().(string)
			if !ok {
				return fmt.Errorf("pattern value must be a string")
			}

			pattern, err := regexp.Compile(patternStr)
			if err != nil {
				return fmt.Errorf("invalid regex pattern: %v", err)
			}

			v.pattern = pattern
			return nil
		}
	}
	return nil
}

// MinLengthValidator validates minimum string length.
type MinLengthValidator struct {
	BaseValidator
	minLength int
}

func NewMinLengthValidator() *MinLengthValidator {
	return &MinLengthValidator{
		BaseValidator: BaseValidator{
			name:                 "minLength",
			supportedAnnotations: []string{"minLength"},
			metadata: ValidatorMetadata{
				Name:           "minLength",
				Description:    "Validates minimum string length",
				Category:       "length",
				Tags:           []string{"string", "length", "constraint"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *MinLengthValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	if len(str) < v.minLength {
		return InvalidResult(NewValidationError(v.name, ErrorCodeLengthConstraint, fmt.Sprintf("string length %d is less than minimum %d", len(str), v.minLength)))
	}

	return ValidResult()
}

func (v *MinLengthValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "minLength" {
			minLen, err := extractIntValue(ann.Value())
			if err != nil {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("minLength value must be an integer: %v", err)))
			}

			str, ok := value.(string)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
			}

			if len(str) < minLen {
				return InvalidResult(NewValidationError(v.name, ErrorCodeLengthConstraint, fmt.Sprintf("string length %d is less than minimum %d", len(str), minLen)))
			}
		}
	}
	return ValidResult()
}

func (v *MinLengthValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "minLength" {
			minLen, err := extractIntValue(ann.Value())
			if err != nil {
				return fmt.Errorf("minLength value must be an integer: %v", err)
			}
			v.minLength = minLen
			return nil
		}
	}
	return nil
}

// MaxLengthValidator validates maximum string length.
type MaxLengthValidator struct {
	BaseValidator
	maxLength int
}

func NewMaxLengthValidator() *MaxLengthValidator {
	return &MaxLengthValidator{
		BaseValidator: BaseValidator{
			name:                 "maxLength",
			supportedAnnotations: []string{"maxLength"},
			metadata: ValidatorMetadata{
				Name:           "maxLength",
				Description:    "Validates maximum string length",
				Category:       "length",
				Tags:           []string{"string", "length", "constraint"},
				SupportedTypes: []string{"string"},
			},
		},
	}
}

func (v *MaxLengthValidator) Validate(value any) ValidationResult {
	str, ok := value.(string)
	if !ok {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
	}

	if len(str) > v.maxLength {
		return InvalidResult(NewValidationError(v.name, ErrorCodeLengthConstraint, fmt.Sprintf("string length %d exceeds maximum %d", len(str), v.maxLength)))
	}

	return ValidResult()
}

func (v *MaxLengthValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "maxLength" {
			maxLen, err := extractIntValue(ann.Value())
			if err != nil {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("maxLength value must be an integer: %v", err)))
			}

			str, ok := value.(string)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be a string"))
			}

			if len(str) > maxLen {
				return InvalidResult(NewValidationError(v.name, ErrorCodeLengthConstraint, fmt.Sprintf("string length %d exceeds maximum %d", len(str), maxLen)))
			}
		}
	}
	return ValidResult()
}

func (v *MaxLengthValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "maxLength" {
			maxLen, err := extractIntValue(ann.Value())
			if err != nil {
				return fmt.Errorf("maxLength value must be an integer: %v", err)
			}
			v.maxLength = maxLen
			return nil
		}
	}
	return nil
}

// MinValidator validates minimum numeric values.
type MinValidator struct {
	BaseValidator
	min any // int, int64, float64
}

func NewMinValidator() *MinValidator {
	return &MinValidator{
		BaseValidator: BaseValidator{
			name:                 "min",
			supportedAnnotations: []string{"min"},
			metadata: ValidatorMetadata{
				Name:           "min",
				Description:    "Validates minimum numeric value",
				Category:       "numeric",
				Tags:           []string{"numeric", "constraint"},
				SupportedTypes: []string{"int", "int64", "float64", "number"},
			},
		},
	}
}

func (v *MinValidator) Validate(value any) ValidationResult {
	return v.validateMinValue(value, v.min)
}

func (v *MinValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "min" {
			return v.validateMinValue(value, ann.Value())
		}
	}
	return ValidResult()
}

func (v *MinValidator) validateMinValue(value, minValue any) ValidationResult {
	// Convert both values to float64 for comparison
	val, err := extractNumericValue(value)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("value must be numeric: %v", err)))
	}

	min, err := extractNumericValue(minValue)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("min value must be numeric: %v", err)))
	}

	if val < min {
		return InvalidResult(NewValidationError(v.name, ErrorCodeOutOfRange, fmt.Sprintf("value %v is less than minimum %v", val, min)))
	}

	return ValidResult()
}

func (v *MinValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "min" {
			v.min = ann.Value()
			return nil
		}
	}
	return nil
}

// MaxValidator validates maximum numeric values.
type MaxValidator struct {
	BaseValidator
	max any
}

func NewMaxValidator() *MaxValidator {
	return &MaxValidator{
		BaseValidator: BaseValidator{
			name:                 "max",
			supportedAnnotations: []string{"max"},
			metadata: ValidatorMetadata{
				Name:           "max",
				Description:    "Validates maximum numeric value",
				Category:       "numeric",
				Tags:           []string{"numeric", "constraint"},
				SupportedTypes: []string{"int", "int64", "float64", "number"},
			},
		},
	}
}

func (v *MaxValidator) Validate(value any) ValidationResult {
	return v.validateMaxValue(value, v.max)
}

func (v *MaxValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "max" {
			return v.validateMaxValue(value, ann.Value())
		}
	}
	return ValidResult()
}

func (v *MaxValidator) validateMaxValue(value, maxValue any) ValidationResult {
	// Convert both values to float64 for comparison
	val, err := extractNumericValue(value)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("value must be numeric: %v", err)))
	}

	max, err := extractNumericValue(maxValue)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("max value must be numeric: %v", err)))
	}

	if val > max {
		return InvalidResult(NewValidationError(v.name, ErrorCodeOutOfRange, fmt.Sprintf("value %v exceeds maximum %v", val, max)))
	}

	return ValidResult()
}

func (v *MaxValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "max" {
			v.max = ann.Value()
			return nil
		}
	}
	return nil
}

// RangeValidator validates numeric ranges.
type RangeValidator struct {
	BaseValidator
	min any
	max any
}

func NewRangeValidator() *RangeValidator {
	return &RangeValidator{
		BaseValidator: BaseValidator{
			name:                 "range",
			supportedAnnotations: []string{"range"},
			metadata: ValidatorMetadata{
				Name:           "range",
				Description:    "Validates numeric value is within range",
				Category:       "numeric",
				Tags:           []string{"numeric", "constraint", "range"},
				SupportedTypes: []string{"int", "int64", "float64", "number"},
			},
		},
	}
}

func (v *RangeValidator) Validate(value any) ValidationResult {
	return v.validateRange(value, v.min, v.max)
}

func (v *RangeValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "range" {
			rangeValue, ok := ann.Value().(map[string]any)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "range value must be an object with min and max properties"))
			}

			min, hasMin := rangeValue["min"]
			max, hasMax := rangeValue["max"]

			if !hasMin || !hasMax {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "range must have both min and max properties"))
			}

			return v.validateRange(value, min, max)
		}
	}
	return ValidResult()
}

func (v *RangeValidator) validateRange(value, minValue, maxValue any) ValidationResult {
	val, err := extractNumericValue(value)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("value must be numeric: %v", err)))
	}

	min, err := extractNumericValue(minValue)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("min value must be numeric: %v", err)))
	}

	max, err := extractNumericValue(maxValue)
	if err != nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("max value must be numeric: %v", err)))
	}

	if val < min || val > max {
		return InvalidResult(NewValidationError(v.name, ErrorCodeOutOfRange, fmt.Sprintf("value %v is outside range [%v, %v]", val, min, max)))
	}

	return ValidResult()
}

func (v *RangeValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "range" {
			rangeValue, ok := ann.Value().(map[string]any)
			if !ok {
				return fmt.Errorf("range value must be an object with min and max properties")
			}

			min, hasMin := rangeValue["min"]
			max, hasMax := rangeValue["max"]

			if !hasMin || !hasMax {
				return fmt.Errorf("range must have both min and max properties")
			}

			v.min = min
			v.max = max
			return nil
		}
	}
	return nil
}

// MinItemsValidator validates minimum array length.
type MinItemsValidator struct {
	BaseValidator
	minItems int
}

func NewMinItemsValidator() *MinItemsValidator {
	return &MinItemsValidator{
		BaseValidator: BaseValidator{
			name:                 "minItems",
			supportedAnnotations: []string{"minItems"},
			metadata: ValidatorMetadata{
				Name:           "minItems",
				Description:    "Validates minimum array length",
				Category:       "array",
				Tags:           []string{"array", "length", "constraint"},
				SupportedTypes: []string{"array", "slice"},
			},
		},
	}
}

func (v *MinItemsValidator) Validate(value any) ValidationResult {
	return v.validateMinItems(value, v.minItems)
}

func (v *MinItemsValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "minItems" {
			minItems, err := extractIntValue(ann.Value())
			if err != nil {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("minItems value must be an integer: %v", err)))
			}
			return v.validateMinItems(value, minItems)
		}
	}
	return ValidResult()
}

func (v *MinItemsValidator) validateMinItems(value any, minItems int) ValidationResult {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be an array or slice"))
	}

	length := rv.Len()
	if length < minItems {
		return InvalidResult(NewValidationError(v.name, ErrorCodeLengthConstraint, fmt.Sprintf("array length %d is less than minimum %d", length, minItems)))
	}

	return ValidResult()
}

func (v *MinItemsValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "minItems" {
			minItems, err := extractIntValue(ann.Value())
			if err != nil {
				return fmt.Errorf("minItems value must be an integer: %v", err)
			}
			v.minItems = minItems
			return nil
		}
	}
	return nil
}

// MaxItemsValidator validates maximum array length.
type MaxItemsValidator struct {
	BaseValidator
	maxItems int
}

func NewMaxItemsValidator() *MaxItemsValidator {
	return &MaxItemsValidator{
		BaseValidator: BaseValidator{
			name:                 "maxItems",
			supportedAnnotations: []string{"maxItems"},
			metadata: ValidatorMetadata{
				Name:           "maxItems",
				Description:    "Validates maximum array length",
				Category:       "array",
				Tags:           []string{"array", "length", "constraint"},
				SupportedTypes: []string{"array", "slice"},
			},
		},
	}
}

func (v *MaxItemsValidator) Validate(value any) ValidationResult {
	return v.validateMaxItems(value, v.maxItems)
}

func (v *MaxItemsValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "maxItems" {
			maxItems, err := extractIntValue(ann.Value())
			if err != nil {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, fmt.Sprintf("maxItems value must be an integer: %v", err)))
			}
			return v.validateMaxItems(value, maxItems)
		}
	}
	return ValidResult()
}

func (v *MaxItemsValidator) validateMaxItems(value any, maxItems int) ValidationResult {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be an array or slice"))
	}

	length := rv.Len()
	if length > maxItems {
		return InvalidResult(NewValidationError(v.name, ErrorCodeLengthConstraint, fmt.Sprintf("array length %d exceeds maximum %d", length, maxItems)))
	}

	return ValidResult()
}

func (v *MaxItemsValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	for _, ann := range annotations {
		if ann.Name() == "maxItems" {
			maxItems, err := extractIntValue(ann.Value())
			if err != nil {
				return fmt.Errorf("maxItems value must be an integer: %v", err)
			}
			v.maxItems = maxItems
			return nil
		}
	}
	return nil
}

// UniqueItemsValidator validates array uniqueness.
type UniqueItemsValidator struct {
	BaseValidator
}

func NewUniqueItemsValidator() *UniqueItemsValidator {
	return &UniqueItemsValidator{
		BaseValidator: BaseValidator{
			name:                 "uniqueItems",
			supportedAnnotations: []string{"uniqueItems"},
			metadata: ValidatorMetadata{
				Name:           "uniqueItems",
				Description:    "Validates that array items are unique",
				Category:       "array",
				Tags:           []string{"array", "unique", "constraint"},
				SupportedTypes: []string{"array", "slice"},
			},
		},
	}
}

func (v *UniqueItemsValidator) Validate(value any) ValidationResult {
	return v.validateUniqueItems(value)
}

func (v *UniqueItemsValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "uniqueItems" {
			unique, ok := ann.Value().(bool)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "uniqueItems value must be a boolean"))
			}
			if unique {
				return v.validateUniqueItems(value)
			}
		}
	}
	return ValidResult()
}

func (v *UniqueItemsValidator) validateUniqueItems(value any) ValidationResult {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "value must be an array or slice"))
	}

	seen := make(map[any]bool)
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if seen[item] {
			return InvalidResult(NewValidationError(v.name, ErrorCodeValidationFailed, fmt.Sprintf("duplicate item found at index %d", i)))
		}
		seen[item] = true
	}

	return ValidResult()
}

func (v *UniqueItemsValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil // No configuration needed
}

// RequiredValidator validates required fields.
type RequiredValidator struct {
	BaseValidator
}

func NewRequiredValidator() *RequiredValidator {
	return &RequiredValidator{
		BaseValidator: BaseValidator{
			name:                 "required",
			supportedAnnotations: []string{"required"},
			metadata: ValidatorMetadata{
				Name:           "required",
				Description:    "Validates that field is present and not empty",
				Category:       "presence",
				Tags:           []string{"required", "presence", "constraint"},
				SupportedTypes: []string{"any"},
			},
		},
	}
}

func (v *RequiredValidator) Validate(value any) ValidationResult {
	return v.validateRequired(value)
}

func (v *RequiredValidator) ValidateWithAnnotations(value any, annotations []annotation.Annotation) ValidationResult {
	for _, ann := range annotations {
		if ann.Name() == "required" {
			required, ok := ann.Value().(bool)
			if !ok {
				return InvalidResult(NewValidationError(v.name, ErrorCodeTypeError, "required value must be a boolean"))
			}
			if required {
				return v.validateRequired(value)
			}
		}
	}
	return ValidResult()
}

func (v *RequiredValidator) validateRequired(value any) ValidationResult {
	if value == nil {
		return InvalidResult(NewValidationError(v.name, ErrorCodeRequiredMissing, "required field is nil"))
	}

	// Check for empty values based on type
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		if rv.Len() == 0 {
			return InvalidResult(NewValidationError(v.name, ErrorCodeRequiredMissing, "required string field is empty"))
		}
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		if rv.Len() == 0 {
			return InvalidResult(NewValidationError(v.name, ErrorCodeRequiredMissing, "required collection field is empty"))
		}
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return InvalidResult(NewValidationError(v.name, ErrorCodeRequiredMissing, "required field is nil"))
		}
	}

	return ValidResult()
}

func (v *RequiredValidator) ConfigureFromAnnotations(annotations []annotation.Annotation) error {
	return nil // No configuration needed
}

// Helper functions for value extraction.
func extractIntValue(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

func extractNumericValue(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to numeric value", value)
	}
}
