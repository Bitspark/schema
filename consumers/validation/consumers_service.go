package validation

import (
	"defs.dev/schema/consumer"
	"defs.dev/schema/core"
)

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
