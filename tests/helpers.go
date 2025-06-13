package tests

import "defs.dev/schema/core"

// Helper function to create a pointer to a float64
func ptr(f float64) *float64 {
	return &f
}

func contains(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to generate JSON Schema for service schemas
func toJSONSchema(schema core.Schema) map[string]any {
	result := map[string]any{}

	// Map schema types to JSON Schema types
	switch schema.Type() {
	case core.TypeService:
		result["type"] = "object"
	default:
		result["type"] = string(schema.Type())
	}

	if desc := schema.Metadata().Description; desc != "" {
		result["description"] = desc
	}

	// Handle service schemas
	if serviceSchema, ok := schema.(core.ServiceSchema); ok {
		result["x-service"] = true

		methods := serviceSchema.Methods()
		if len(methods) > 0 {
			methodsMap := make(map[string]any)
			for _, method := range methods {
				methodsMap[method.Name()] = toJSONSchema(method.Function())
			}
			result["x-methods"] = methodsMap
		}
	}

	// Handle function schemas
	if functionSchema, ok := schema.(core.FunctionSchema); ok {
		result["type"] = "object"
		result["x-function"] = true

		// Add inputs as properties
		if inputs := functionSchema.Inputs(); inputs != nil {
			properties := make(map[string]any)
			required := []string{}

			for _, arg := range inputs.Args() {
				properties[arg.Name()] = toJSONSchema(arg.Schema())
			}

			for _, reqInput := range functionSchema.RequiredInputs() {
				required = append(required, reqInput)
			}

			if len(properties) > 0 {
				result["properties"] = properties
			}
			if len(required) > 0 {
				result["required"] = required
			}
		}

		// Add outputs
		if outputs := functionSchema.Outputs(); outputs != nil && len(outputs.Args()) > 0 {
			outputsMap := make(map[string]any)
			for _, arg := range outputs.Args() {
				outputsMap[arg.Name()] = toJSONSchema(arg.Schema())
			}
			result["x-returns"] = outputsMap
		}
	}

	// Handle object schemas
	if objectSchema, ok := schema.(core.ObjectSchema); ok {
		if properties := objectSchema.Properties(); properties != nil && len(properties) > 0 {
			propMap := make(map[string]any)
			for name, prop := range properties {
				propMap[name] = toJSONSchema(prop)
			}
			result["properties"] = propMap
		}
		if required := objectSchema.Required(); required != nil && len(required) > 0 {
			result["required"] = required
		}
	}

	// Handle string schemas
	if stringSchema, ok := schema.(core.StringSchema); ok {
		if minLen := stringSchema.MinLength(); minLen != nil {
			result["minLength"] = *minLen
		}
		if maxLen := stringSchema.MaxLength(); maxLen != nil {
			result["maxLength"] = *maxLen
		}
		if pattern := stringSchema.Pattern(); pattern != "" {
			result["pattern"] = pattern
		}
	}

	// Handle number schemas
	if numberSchema, ok := schema.(core.NumberSchema); ok {
		if min := numberSchema.Minimum(); min != nil {
			result["minimum"] = *min
		}
		if max := numberSchema.Maximum(); max != nil {
			result["maximum"] = *max
		}
	}

	return result
}

// Helper function to generate JSON Schema using a simple stub
func toJSONSchemaIntegration(schema core.Schema) map[string]any {
	// Simple stub implementation for testing
	result := map[string]any{}

	// Map schema types to JSON Schema types
	switch schema.Type() {
	case core.TypeStructure:
		result["type"] = "object"
	default:
		result["type"] = string(schema.Type())
	}

	if desc := schema.Metadata().Description; desc != "" {
		result["description"] = desc
	}

	// Handle object schemas
	if objectSchema, ok := schema.(core.ObjectSchema); ok {
		if properties := objectSchema.Properties(); properties != nil && len(properties) > 0 {
			propMap := make(map[string]any)
			for name, prop := range properties {
				propMap[name] = toJSONSchemaIntegration(prop)
			}
			result["properties"] = propMap
		}
		if required := objectSchema.Required(); required != nil && len(required) > 0 {
			// Convert to []any for JSON compatibility
			requiredAny := make([]any, len(required))
			for i, req := range required {
				requiredAny[i] = req
			}
			result["required"] = requiredAny
		}
		if !objectSchema.AdditionalProperties() {
			result["additionalProperties"] = false
		}
	}

	// Handle array schemas
	if arraySchema, ok := schema.(core.ArraySchema); ok {
		if minItems := arraySchema.MinItems(); minItems != nil {
			result["minItems"] = float64(*minItems)
		}
		if maxItems := arraySchema.MaxItems(); maxItems != nil {
			result["maxItems"] = float64(*maxItems)
		}
		if arraySchema.UniqueItemsRequired() {
			result["uniqueItems"] = true
		}
		if itemSchema := arraySchema.ItemSchema(); itemSchema != nil {
			result["items"] = toJSONSchemaIntegration(itemSchema)
		}
	}

	// Handle string schemas
	if stringSchema, ok := schema.(core.StringSchema); ok {
		if minLen := stringSchema.MinLength(); minLen != nil {
			result["minLength"] = *minLen
		}
		if maxLen := stringSchema.MaxLength(); maxLen != nil {
			result["maxLength"] = *maxLen
		}
		if pattern := stringSchema.Pattern(); pattern != "" {
			result["pattern"] = pattern
		}
	}

	// Handle integer schemas
	if integerSchema, ok := schema.(core.IntegerSchema); ok {
		if min := integerSchema.Minimum(); min != nil {
			result["minimum"] = *min
		}
		if max := integerSchema.Maximum(); max != nil {
			result["maximum"] = *max
		}
	}

	// Handle number schemas
	if numberSchema, ok := schema.(core.NumberSchema); ok {
		if min := numberSchema.Minimum(); min != nil {
			result["minimum"] = *min
		}
		if max := numberSchema.Maximum(); max != nil {
			result["maximum"] = *max
		}
	}

	return result
}

// Helper function to generate JSON Schema using a simple stub
func toJSONSchemaBasic(schema core.Schema) map[string]any {
	// Simple stub implementation for testing
	result := map[string]any{}

	// Map schema types to JSON Schema types
	switch schema.Type() {
	case core.TypeStructure:
		result["type"] = "object"
	default:
		result["type"] = string(schema.Type())
	}

	if desc := schema.Metadata().Description; desc != "" {
		result["description"] = desc
	}

	// Add type-specific properties
	switch s := schema.(type) {
	case core.StringSchema:
		if minLen := s.MinLength(); minLen != nil {
			result["minLength"] = *minLen
		}
		if maxLen := s.MaxLength(); maxLen != nil {
			result["maxLength"] = *maxLen
		}
		if pattern := s.Pattern(); pattern != "" {
			result["pattern"] = pattern
		}
	case core.IntegerSchema:
		if min := s.Minimum(); min != nil {
			result["minimum"] = *min
		}
		if max := s.Maximum(); max != nil {
			result["maximum"] = *max
		}
	case core.ObjectSchema:
		if !s.AdditionalProperties() {
			result["additionalProperties"] = false
		}
		if required := s.Required(); required != nil && len(required) > 0 {
			// Convert to []any for JSON compatibility
			requiredAny := make([]any, len(required))
			for i, req := range required {
				requiredAny[i] = req
			}
			result["required"] = requiredAny
		}
		if properties := s.Properties(); properties != nil && len(properties) > 0 {
			propMap := make(map[string]any)
			for name, prop := range properties {
				propMap[name] = toJSONSchemaBasic(prop)
			}
			result["properties"] = propMap
		}
	}

	return result
}
