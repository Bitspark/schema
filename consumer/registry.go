package consumer

import (
	"fmt"
	"sync"
	"time"

	"defs.dev/schema/api/core"
)

// Registry manages both AnnotationConsumers and ValueConsumers with purpose-based selection.
type Registry interface {
	// Registration
	RegisterSchemaConsumer(consumer AnnotationConsumer) error
	RegisterValueConsumer(consumer ValueConsumer) error

	// Discovery
	GetSchemaConsumer(name string) (AnnotationConsumer, bool)
	GetValueConsumer(name string) (ValueConsumer, bool)
	ListSchemaConsumers() []string
	ListValueConsumers() []string
	ListByPurpose(purpose ConsumerPurpose) ([]string, []string) // schema consumers, value consumers

	// Schema processing
	GetApplicableSchemaConsumers(schema core.Schema) []AnnotationConsumer
	GetApplicableSchemaConsumersByPurpose(schema core.Schema, purpose ConsumerPurpose) []AnnotationConsumer
	ProcessSchemaWithPurpose(purpose ConsumerPurpose, schema core.Schema) (ConsumerResult, error)
	ProcessSchemaAllWithPurpose(purpose ConsumerPurpose, schema core.Schema) ([]ConsumerResult, error)
	ProcessSchemaWithPurposes(purposes []ConsumerPurpose, schema core.Schema) (ProcessingResult, error)

	// Value processing
	GetApplicableValueConsumers(schema core.Schema) []ValueConsumer
	GetApplicableValueConsumersByPurpose(schema core.Schema, purpose ConsumerPurpose) []ValueConsumer
	ProcessValueWithPurpose(purpose ConsumerPurpose, schema core.Schema, value core.Value[any]) (ConsumerResult, error)
	ProcessValueAllWithPurpose(purpose ConsumerPurpose, schema core.Schema, value core.Value[any]) ([]ConsumerResult, error)
	ProcessValueWithPurposes(purposes []ConsumerPurpose, schema core.Schema, value core.Value[any]) (ProcessingResult, error)

	// Combined processing
	ProcessWithContext(ctx ProcessingContext) (ProcessingResult, error)
}

// ProcessingResult contains results from multiple consumer types and purposes.
type ProcessingResult struct {
	Success       bool                                 `json:"success"`
	SchemaResults map[ConsumerPurpose][]ConsumerResult `json:"schema_results"`
	ValueResults  map[ConsumerPurpose][]ConsumerResult `json:"value_results"`
	Errors        map[ConsumerPurpose][]error          `json:"errors,omitempty"`
	ExecutedAt    time.Time                            `json:"executed_at"`
	Duration      time.Duration                        `json:"duration"`
}

// RegistryImpl implements the Registry interface.
type RegistryImpl struct {
	mu              sync.RWMutex
	schemaConsumers map[string]AnnotationConsumer
	valueConsumers  map[string]ValueConsumer
	schemaByPurpose map[ConsumerPurpose][]AnnotationConsumer
	valueByPurpose  map[ConsumerPurpose][]ValueConsumer
	conditionCache  map[cacheKey]bool // cache for condition matching
}

type cacheKey struct {
	consumerName string
	schemaHash   string // simplified, could use better hashing
}

// NewRegistry creates a new consumer registry.
func NewRegistry() Registry {
	return &RegistryImpl{
		schemaConsumers: make(map[string]AnnotationConsumer),
		valueConsumers:  make(map[string]ValueConsumer),
		schemaByPurpose: make(map[ConsumerPurpose][]AnnotationConsumer),
		valueByPurpose:  make(map[ConsumerPurpose][]ValueConsumer),
		conditionCache:  make(map[cacheKey]bool),
	}
}

// RegisterSchemaConsumer registers an AnnotationConsumer.
func (r *RegistryImpl) RegisterSchemaConsumer(consumer AnnotationConsumer) error {
	if consumer == nil {
		return fmt.Errorf("consumer cannot be nil")
	}
	if consumer.Name() == "" {
		return fmt.Errorf("consumer name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := consumer.Name()
	if _, exists := r.schemaConsumers[name]; exists {
		return fmt.Errorf("schema consumer %s already registered", name)
	}

	r.schemaConsumers[name] = consumer
	purpose := consumer.Purpose()
	r.schemaByPurpose[purpose] = append(r.schemaByPurpose[purpose], consumer)

	return nil
}

// RegisterValueConsumer registers a ValueConsumer.
func (r *RegistryImpl) RegisterValueConsumer(consumer ValueConsumer) error {
	if consumer == nil {
		return fmt.Errorf("consumer cannot be nil")
	}
	if consumer.Name() == "" {
		return fmt.Errorf("consumer name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := consumer.Name()
	if _, exists := r.valueConsumers[name]; exists {
		return fmt.Errorf("value consumer %s already registered", name)
	}

	r.valueConsumers[name] = consumer
	purpose := consumer.Purpose()
	r.valueByPurpose[purpose] = append(r.valueByPurpose[purpose], consumer)

	return nil
}

// GetSchemaConsumer retrieves a schema consumer by name.
func (r *RegistryImpl) GetSchemaConsumer(name string) (AnnotationConsumer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	consumer, exists := r.schemaConsumers[name]
	return consumer, exists
}

// GetValueConsumer retrieves a value consumer by name.
func (r *RegistryImpl) GetValueConsumer(name string) (ValueConsumer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	consumer, exists := r.valueConsumers[name]
	return consumer, exists
}

// ListSchemaConsumers returns all registered schema consumer names.
func (r *RegistryImpl) ListSchemaConsumers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.schemaConsumers))
	for name := range r.schemaConsumers {
		names = append(names, name)
	}
	return names
}

// ListValueConsumers returns all registered value consumer names.
func (r *RegistryImpl) ListValueConsumers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.valueConsumers))
	for name := range r.valueConsumers {
		names = append(names, name)
	}
	return names
}

// ListByPurpose returns consumer names grouped by purpose.
func (r *RegistryImpl) ListByPurpose(purpose ConsumerPurpose) ([]string, []string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var schemaNames, valueNames []string

	if schemaConsumers, ok := r.schemaByPurpose[purpose]; ok {
		for _, consumer := range schemaConsumers {
			schemaNames = append(schemaNames, consumer.Name())
		}
	}

	if valueConsumers, ok := r.valueByPurpose[purpose]; ok {
		for _, consumer := range valueConsumers {
			valueNames = append(valueNames, consumer.Name())
		}
	}

	return schemaNames, valueNames
}

// GetApplicableSchemaConsumers finds all schema consumers that match the schema.
func (r *RegistryImpl) GetApplicableSchemaConsumers(schema core.Schema) []AnnotationConsumer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var applicable []AnnotationConsumer
	for _, consumer := range r.schemaConsumers {
		if r.checkCondition(consumer.Name(), schema, consumer.ApplicableSchemas()) {
			applicable = append(applicable, consumer)
		}
	}
	return applicable
}

// GetApplicableSchemaConsumersByPurpose finds schema consumers that match schema and purpose.
func (r *RegistryImpl) GetApplicableSchemaConsumersByPurpose(schema core.Schema, purpose ConsumerPurpose) []AnnotationConsumer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var applicable []AnnotationConsumer
	if consumers, ok := r.schemaByPurpose[purpose]; ok {
		for _, consumer := range consumers {
			if r.checkCondition(consumer.Name(), schema, consumer.ApplicableSchemas()) {
				applicable = append(applicable, consumer)
			}
		}
	}
	return applicable
}

// GetApplicableValueConsumers finds all value consumers that match the schema.
func (r *RegistryImpl) GetApplicableValueConsumers(schema core.Schema) []ValueConsumer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var applicable []ValueConsumer
	for _, consumer := range r.valueConsumers {
		if r.checkCondition(consumer.Name(), schema, consumer.ApplicableSchemas()) {
			applicable = append(applicable, consumer)
		}
	}
	return applicable
}

// GetApplicableValueConsumersByPurpose finds value consumers that match schema and purpose.
func (r *RegistryImpl) GetApplicableValueConsumersByPurpose(schema core.Schema, purpose ConsumerPurpose) []ValueConsumer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var applicable []ValueConsumer
	if consumers, ok := r.valueByPurpose[purpose]; ok {
		for _, consumer := range consumers {
			if r.checkCondition(consumer.Name(), schema, consumer.ApplicableSchemas()) {
				applicable = append(applicable, consumer)
			}
		}
	}
	return applicable
}

// ProcessSchemaWithPurpose processes a schema with the first matching consumer of the given purpose.
func (r *RegistryImpl) ProcessSchemaWithPurpose(purpose ConsumerPurpose, schema core.Schema) (ConsumerResult, error) {
	consumers := r.GetApplicableSchemaConsumersByPurpose(schema, purpose)
	if len(consumers) == 0 {
		return nil, fmt.Errorf("no applicable schema consumers found for purpose %s", purpose)
	}

	ctx := ProcessingContext{
		Schema: schema,
		Path:   []string{},
	}

	result, err := consumers[0].ProcessSchema(ctx)
	if err != nil {
		return nil, NewConsumerError(consumers[0].Name(), purpose, ctx.Path, err)
	}
	return result, nil
}

// ProcessSchemaAllWithPurpose processes a schema with all matching consumers of the given purpose.
func (r *RegistryImpl) ProcessSchemaAllWithPurpose(purpose ConsumerPurpose, schema core.Schema) ([]ConsumerResult, error) {
	consumers := r.GetApplicableSchemaConsumersByPurpose(schema, purpose)
	if len(consumers) == 0 {
		return nil, fmt.Errorf("no applicable schema consumers found for purpose %s", purpose)
	}

	ctx := ProcessingContext{
		Schema: schema,
		Path:   []string{},
	}

	var results []ConsumerResult
	var errors []error

	for _, consumer := range consumers {
		result, err := consumer.ProcessSchema(ctx)
		if err != nil {
			errors = append(errors, NewConsumerError(consumer.Name(), purpose, ctx.Path, err))
		} else {
			results = append(results, result)
		}
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("some consumers failed: %v", errors)
	}

	return results, nil
}

// ProcessValueWithPurpose processes a value with the first matching consumer of the given purpose.
func (r *RegistryImpl) ProcessValueWithPurpose(purpose ConsumerPurpose, schema core.Schema, value core.Value[any]) (ConsumerResult, error) {
	consumers := r.GetApplicableValueConsumersByPurpose(schema, purpose)
	if len(consumers) == 0 {
		return nil, fmt.Errorf("no applicable value consumers found for purpose %s", purpose)
	}

	ctx := ProcessingContext{
		Schema: schema,
		Value:  value,
		Path:   []string{},
	}

	result, err := consumers[0].ProcessValue(ctx, value)
	if err != nil {
		return nil, NewConsumerError(consumers[0].Name(), purpose, ctx.Path, err)
	}
	return result, nil
}

// ProcessValueAllWithPurpose processes a value with all matching consumers of the given purpose.
func (r *RegistryImpl) ProcessValueAllWithPurpose(purpose ConsumerPurpose, schema core.Schema, value core.Value[any]) ([]ConsumerResult, error) {
	consumers := r.GetApplicableValueConsumersByPurpose(schema, purpose)
	if len(consumers) == 0 {
		return nil, fmt.Errorf("no applicable value consumers found for purpose %s", purpose)
	}

	ctx := ProcessingContext{
		Schema: schema,
		Value:  value,
		Path:   []string{},
	}

	var results []ConsumerResult
	var errors []error

	for _, consumer := range consumers {
		result, err := consumer.ProcessValue(ctx, value)
		if err != nil {
			errors = append(errors, NewConsumerError(consumer.Name(), purpose, ctx.Path, err))
		} else {
			results = append(results, result)
		}
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("some consumers failed: %v", errors)
	}

	return results, nil
}

// ProcessSchemaWithPurposes processes a schema with multiple purposes.
func (r *RegistryImpl) ProcessSchemaWithPurposes(purposes []ConsumerPurpose, schema core.Schema) (ProcessingResult, error) {
	start := time.Now()
	result := ProcessingResult{
		Success:       true,
		SchemaResults: make(map[ConsumerPurpose][]ConsumerResult),
		ValueResults:  make(map[ConsumerPurpose][]ConsumerResult),
		Errors:        make(map[ConsumerPurpose][]error),
		ExecutedAt:    start,
	}

	for _, purpose := range purposes {
		results, err := r.ProcessSchemaAllWithPurpose(purpose, schema)
		if err != nil {
			result.Success = false
			result.Errors[purpose] = []error{err}
		} else {
			result.SchemaResults[purpose] = results
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// ProcessValueWithPurposes processes a value with multiple purposes.
func (r *RegistryImpl) ProcessValueWithPurposes(purposes []ConsumerPurpose, schema core.Schema, value core.Value[any]) (ProcessingResult, error) {
	start := time.Now()
	result := ProcessingResult{
		Success:       true,
		SchemaResults: make(map[ConsumerPurpose][]ConsumerResult),
		ValueResults:  make(map[ConsumerPurpose][]ConsumerResult),
		Errors:        make(map[ConsumerPurpose][]error),
		ExecutedAt:    start,
	}

	for _, purpose := range purposes {
		results, err := r.ProcessValueAllWithPurpose(purpose, schema, value)
		if err != nil {
			result.Success = false
			result.Errors[purpose] = []error{err}
		} else {
			result.ValueResults[purpose] = results
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// ProcessWithContext processes using the provided context which may specify both schema and value processing.
func (r *RegistryImpl) ProcessWithContext(ctx ProcessingContext) (ProcessingResult, error) {
	start := time.Now()
	result := ProcessingResult{
		Success:       true,
		SchemaResults: make(map[ConsumerPurpose][]ConsumerResult),
		ValueResults:  make(map[ConsumerPurpose][]ConsumerResult),
		Errors:        make(map[ConsumerPurpose][]error),
		ExecutedAt:    start,
	}

	// Extract purposes from context options
	var purposes []ConsumerPurpose
	if purposesOption, ok := ctx.Options["purposes"]; ok {
		if p, ok := purposesOption.([]ConsumerPurpose); ok {
			purposes = p
		}
	}

	if len(purposes) == 0 {
		result.Duration = time.Since(start)
		return result, fmt.Errorf("no purposes specified in context")
	}

	// Process schema consumers
	if ctx.Schema != nil {
		for _, purpose := range purposes {
			results, err := r.ProcessSchemaAllWithPurpose(purpose, ctx.Schema)
			if err != nil {
				result.Success = false
				result.Errors[purpose] = append(result.Errors[purpose], err)
			} else {
				result.SchemaResults[purpose] = results
			}
		}
	}

	// Process value consumers
	if ctx.Value != nil && ctx.Schema != nil {
		for _, purpose := range purposes {
			results, err := r.ProcessValueAllWithPurpose(purpose, ctx.Schema, ctx.Value)
			if err != nil {
				result.Success = false
				result.Errors[purpose] = append(result.Errors[purpose], err)
			} else {
				result.ValueResults[purpose] = results
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// checkCondition checks if a schema matches a condition, with caching.
func (r *RegistryImpl) checkCondition(consumerName string, schema core.Schema, condition SchemaCondition) bool {
	// Simple cache key (could be improved with better hashing)
	key := cacheKey{
		consumerName: consumerName,
		schemaHash:   fmt.Sprintf("%p", schema), // simplified hash
	}

	if cached, exists := r.conditionCache[key]; exists {
		return cached
	}

	matches := condition.Matches(schema)
	r.conditionCache[key] = matches
	return matches
}
