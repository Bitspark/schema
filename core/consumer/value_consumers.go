// Package consumer provides the generic consumer framework.
// Specific consumer implementations should be in their respective packages.
package consumer

import (
	"defs.dev/schema/core"
)

// Common consumer purposes - these are just examples, users can define their own
const (
	PurposeValidation    ConsumerPurpose = "validation"
	PurposeFormatting    ConsumerPurpose = "formatting"
	PurposeGeneration    ConsumerPurpose = "generation"
	PurposeDocumentation ConsumerPurpose = "documentation"
	PurposeTransform     ConsumerPurpose = "transform"
	PurposeAnalysis      ConsumerPurpose = "analysis"
)

// ExampleStringFormattingConsumer demonstrates how to implement a value consumer.
// Real implementations should be in their respective packages (e.g., validation, formatting).
type ExampleStringFormattingConsumer struct{}

func (c *ExampleStringFormattingConsumer) Name() string {
	return "example_string_formatter"
}

func (c *ExampleStringFormattingConsumer) Purpose() ConsumerPurpose {
	return PurposeFormatting
}

func (c *ExampleStringFormattingConsumer) ApplicableSchemas() SchemaCondition {
	return Type(core.TypeString)
}

func (c *ExampleStringFormattingConsumer) ProcessValue(ctx ProcessingContext, value core.Value[any]) (ConsumerResult, error) {
	actualValue := value.Value()
	str, ok := actualValue.(string)
	if !ok {
		return NewResult("error", "expected string"), nil
	}

	// Simple example: just return the string as-is
	return NewResult("formatting", str), nil
}

func (c *ExampleStringFormattingConsumer) Metadata() ConsumerMetadata {
	return ConsumerMetadata{
		Name:         "example_string_formatter",
		Purpose:      PurposeFormatting,
		Description:  "Example string formatter - real implementations should be in specific packages",
		Version:      "1.0.0",
		Tags:         []string{"example", "formatting", "string"},
		ResultKind:   "formatting",
		ResultGoType: "string",
	}
}

// RegisterExampleConsumers registers example consumers for testing/demonstration
func RegisterExampleConsumers(registry Registry) error {
	consumers := []ValueConsumer{
		&ExampleStringFormattingConsumer{},
	}

	for _, consumer := range consumers {
		if err := registry.RegisterValueConsumer(consumer); err != nil {
			return err
		}
	}

	return nil
}
