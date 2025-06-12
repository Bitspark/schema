package native

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"defs.dev/schema/annotation"
)

// DefaultTagParser implements TagParser for common struct tags.
type DefaultTagParser struct {
	annotationRegistry annotation.AnnotationRegistry
	strictMode         bool
	supportedTags      []string
}

// NewDefaultTagParser creates a new tag parser with default configuration.
func NewDefaultTagParser(annotationRegistry annotation.AnnotationRegistry) *DefaultTagParser {
	return &DefaultTagParser{
		annotationRegistry: annotationRegistry,
		strictMode:         false,
		supportedTags: []string{
			"json", "validate", "format", "pattern", "min", "max", "minLength", "maxLength",
			"minItems", "maxItems", "uniqueItems", "required", "default", "enum",
			"description", "example", "deprecated", "title", "category", "tags",
		},
	}
}

// ParseTags implements TagParser.
func (p *DefaultTagParser) ParseTags(tags reflect.StructTag) ([]annotation.Annotation, error) {
	var annotations []annotation.Annotation

	// Parse each supported tag
	for _, tagKey := range p.supportedTags {
		if tagValue, ok := tags.Lookup(tagKey); ok {
			ann, err := p.ParseTag(tagKey, tagValue)
			if err != nil {
				if p.strictMode {
					return nil, fmt.Errorf("failed to parse tag %s: %v", tagKey, err)
				}
				// In non-strict mode, continue processing other tags
				continue
			}
			if ann != nil {
				annotations = append(annotations, ann)
			}
		}
	}

	return annotations, nil
}

// ParseTag implements TagParser.
func (p *DefaultTagParser) ParseTag(key, value string) (annotation.Annotation, error) {
	switch key {
	case "json":
		return p.parseJSONTag(value)
	case "validate":
		return p.parseValidateTag(value)
	case "format":
		return p.parseFormatTag(value)
	case "pattern":
		return p.parsePatternTag(value)
	case "min":
		return p.parseMinTag(value)
	case "max":
		return p.parseMaxTag(value)
	case "minLength":
		return p.parseMinLengthTag(value)
	case "maxLength":
		return p.parseMaxLengthTag(value)
	case "minItems":
		return p.parseMinItemsTag(value)
	case "maxItems":
		return p.parseMaxItemsTag(value)
	case "uniqueItems":
		return p.parseUniqueItemsTag(value)
	case "required":
		return p.parseRequiredTag(value)
	case "default":
		return p.parseDefaultTag(value)
	case "enum":
		return p.parseEnumTag(value)
	case "description":
		return p.parseDescriptionTag(value)
	case "example":
		return p.parseExampleTag(value)
	case "deprecated":
		return p.parseDeprecatedTag(value)
	case "title":
		return p.parseTitleTag(value)
	case "category":
		return p.parseCategoryTag(value)
	case "tags":
		return p.parseTagsTag(value)
	default:
		if p.strictMode {
			return nil, fmt.Errorf("unsupported tag: %s", key)
		}
		return nil, nil // Ignore unknown tags in non-strict mode
	}
}

// GetSupportedTags implements TagParser.
func (p *DefaultTagParser) GetSupportedTags() []string {
	return append([]string{}, p.supportedTags...) // Return a copy
}

// HasTag implements TagParser.
func (p *DefaultTagParser) HasTag(key string) bool {
	for _, tag := range p.supportedTags {
		if tag == key {
			return true
		}
	}
	return false
}

// SetAnnotationRegistry implements TagParser.
func (p *DefaultTagParser) SetAnnotationRegistry(registry annotation.AnnotationRegistry) {
	p.annotationRegistry = registry
}

// SetStrictMode implements TagParser.
func (p *DefaultTagParser) SetStrictMode(strict bool) {
	p.strictMode = strict
}

// Tag parsing methods

func (p *DefaultTagParser) parseJSONTag(value string) (annotation.Annotation, error) {
	// Parse json tag: "name,omitempty"
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return nil, nil
	}

	jsonName := parts[0]
	omitempty := false

	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "omitempty":
			omitempty = true
		case "-":
			// Field should be ignored
			return nil, nil
		}
	}

	// Create annotation with json metadata
	metadata := map[string]any{
		"name":      jsonName,
		"omitempty": omitempty,
	}

	return p.annotationRegistry.Create("json", metadata)
}

func (p *DefaultTagParser) parseValidateTag(value string) (annotation.Annotation, error) {
	// Parse validate tag: "required,min=1,max=100"
	validators := strings.Split(value, ",")
	validatorList := make([]string, 0, len(validators))

	for _, validator := range validators {
		validator = strings.TrimSpace(validator)
		if validator != "" {
			validatorList = append(validatorList, validator)
		}
	}

	return p.annotationRegistry.Create("validators", validatorList)
}

func (p *DefaultTagParser) parseFormatTag(value string) (annotation.Annotation, error) {
	return p.annotationRegistry.Create("format", value)
}

func (p *DefaultTagParser) parsePatternTag(value string) (annotation.Annotation, error) {
	return p.annotationRegistry.Create("pattern", value)
}

func (p *DefaultTagParser) parseMinTag(value string) (annotation.Annotation, error) {
	minValue, err := parseNumericValue(value)
	if err != nil {
		return nil, fmt.Errorf("invalid min value: %v", err)
	}
	return p.annotationRegistry.Create("min", minValue)
}

func (p *DefaultTagParser) parseMaxTag(value string) (annotation.Annotation, error) {
	maxValue, err := parseNumericValue(value)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	}
	return p.annotationRegistry.Create("max", maxValue)
}

func (p *DefaultTagParser) parseMinLengthTag(value string) (annotation.Annotation, error) {
	minLength, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid minLength value: %v", err)
	}
	return p.annotationRegistry.Create("minLength", minLength)
}

func (p *DefaultTagParser) parseMaxLengthTag(value string) (annotation.Annotation, error) {
	maxLength, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid maxLength value: %v", err)
	}
	return p.annotationRegistry.Create("maxLength", maxLength)
}

func (p *DefaultTagParser) parseMinItemsTag(value string) (annotation.Annotation, error) {
	minItems, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid minItems value: %v", err)
	}
	return p.annotationRegistry.Create("minItems", minItems)
}

func (p *DefaultTagParser) parseMaxItemsTag(value string) (annotation.Annotation, error) {
	maxItems, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid maxItems value: %v", err)
	}
	return p.annotationRegistry.Create("maxItems", maxItems)
}

func (p *DefaultTagParser) parseUniqueItemsTag(value string) (annotation.Annotation, error) {
	uniqueItems, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid uniqueItems value: %v", err)
	}
	return p.annotationRegistry.Create("uniqueItems", uniqueItems)
}

func (p *DefaultTagParser) parseRequiredTag(value string) (annotation.Annotation, error) {
	required, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid required value: %v", err)
	}
	return p.annotationRegistry.Create("required", required)
}

func (p *DefaultTagParser) parseDefaultTag(value string) (annotation.Annotation, error) {
	// Try to parse as different types
	if parsed, err := parseAnyValue(value); err == nil {
		return p.annotationRegistry.Create("default", parsed)
	}
	// Fall back to string value
	return p.annotationRegistry.Create("default", value)
}

func (p *DefaultTagParser) parseEnumTag(value string) (annotation.Annotation, error) {
	// Parse comma-separated enum values
	enumValues := strings.Split(value, ",")
	for i, val := range enumValues {
		enumValues[i] = strings.TrimSpace(val)
	}
	return p.annotationRegistry.Create("enum", enumValues)
}

func (p *DefaultTagParser) parseDescriptionTag(value string) (annotation.Annotation, error) {
	return p.annotationRegistry.Create("description", value)
}

func (p *DefaultTagParser) parseExampleTag(value string) (annotation.Annotation, error) {
	// Try to parse as appropriate type
	if parsed, err := parseAnyValue(value); err == nil {
		return p.annotationRegistry.Create("examples", []any{parsed})
	}
	// Fall back to string
	return p.annotationRegistry.Create("examples", []any{value})
}

func (p *DefaultTagParser) parseDeprecatedTag(value string) (annotation.Annotation, error) {
	deprecated, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid deprecated value: %v", err)
	}
	return p.annotationRegistry.Create("deprecated", deprecated)
}

func (p *DefaultTagParser) parseTitleTag(value string) (annotation.Annotation, error) {
	return p.annotationRegistry.Create("title", value)
}

func (p *DefaultTagParser) parseCategoryTag(value string) (annotation.Annotation, error) {
	return p.annotationRegistry.Create("category", value)
}

func (p *DefaultTagParser) parseTagsTag(value string) (annotation.Annotation, error) {
	// Parse comma-separated tags
	tags := strings.Split(value, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return p.annotationRegistry.Create("tags", tags)
}

// Helper functions for parsing values

func parseNumericValue(value string) (any, error) {
	// Try integer first
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal, nil
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal, nil
	}

	return nil, fmt.Errorf("not a valid numeric value: %s", value)
}

func parseAnyValue(value string) (any, error) {
	// Try boolean
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal, nil
	}

	// Try integer
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal, nil
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal, nil
	}

	// Return as string
	return value, nil
}
