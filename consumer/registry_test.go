package consumer

import (
	"testing"

	"defs.dev/schema/api/core"
)

// Mock schema for testing
type mockSchema struct {
	schemaType  core.SchemaType
	annotations []core.Annotation
}

func (m *mockSchema) Type() core.SchemaType                    { return m.schemaType }
func (m *mockSchema) Annotations() []core.Annotation           { return m.annotations }
func (m *mockSchema) Metadata() core.SchemaMetadata            { return core.SchemaMetadata{} }
func (m *mockSchema) Clone() core.Schema                       { return m }
func (m *mockSchema) Validate(value any) core.ValidationResult { return core.ValidationResult{} }

// Mock annotation for testing
type mockAnnotation struct {
	name  string
	value any
}

func (m *mockAnnotation) Name() string                      { return m.name }
func (m *mockAnnotation) Value() any                        { return m.value }
func (m *mockAnnotation) Schema() core.Schema               { return nil }
func (m *mockAnnotation) Validators() []string              { return []string{} }
func (m *mockAnnotation) Metadata() core.AnnotationMetadata { return core.AnnotationMetadata{} }
func (m *mockAnnotation) Validate() core.AnnotationValidationResult {
	return core.AnnotationValidationResult{Valid: true}
}
func (m *mockAnnotation) ToMap() map[string]any { return map[string]any{m.name: m.value} }

// Mock value for testing
type mockValue struct {
	val any
}

func (m *mockValue) Value() any        { return m.val }
func (m *mockValue) Copy() any         { return m.val }
func (m *mockValue) String() string    { return "mock" }
func (m *mockValue) IsNull() bool      { return m.val == nil }
func (m *mockValue) IsComposite() bool { return false }

// Mock schema consumer
type mockSchemaConsumer struct {
	name      string
	purpose   ConsumerPurpose
	condition SchemaCondition
}

func (m *mockSchemaConsumer) Name() string                       { return m.name }
func (m *mockSchemaConsumer) Purpose() ConsumerPurpose           { return m.purpose }
func (m *mockSchemaConsumer) ApplicableSchemas() SchemaCondition { return m.condition }
func (m *mockSchemaConsumer) Metadata() ConsumerMetadata {
	return ConsumerMetadata{Name: m.name, Purpose: m.purpose}
}
func (m *mockSchemaConsumer) ProcessSchema(ctx ProcessingContext) (ConsumerResult, error) {
	return NewResult("test", "schema-processed"), nil
}

// Mock value consumer
type mockValueConsumer struct {
	name      string
	purpose   ConsumerPurpose
	condition SchemaCondition
}

func (m *mockValueConsumer) Name() string                       { return m.name }
func (m *mockValueConsumer) Purpose() ConsumerPurpose           { return m.purpose }
func (m *mockValueConsumer) ApplicableSchemas() SchemaCondition { return m.condition }
func (m *mockValueConsumer) Metadata() ConsumerMetadata {
	return ConsumerMetadata{Name: m.name, Purpose: m.purpose}
}
func (m *mockValueConsumer) ProcessValue(ctx ProcessingContext, value core.Value[any]) (ConsumerResult, error) {
	return NewResult("test", "value-processed"), nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	schemaConsumer := &mockSchemaConsumer{
		name:      "test-schema",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	valueConsumer := &mockValueConsumer{
		name:      "test-value",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	// Test registration
	err := registry.RegisterSchemaConsumer(schemaConsumer)
	if err != nil {
		t.Fatalf("Failed to register schema consumer: %v", err)
	}

	err = registry.RegisterValueConsumer(valueConsumer)
	if err != nil {
		t.Fatalf("Failed to register value consumer: %v", err)
	}

	// Test retrieval
	retrieved, exists := registry.GetSchemaConsumer("test-schema")
	if !exists {
		t.Fatal("Schema consumer not found")
	}
	if retrieved.Name() != "test-schema" {
		t.Errorf("Expected name 'test-schema', got %s", retrieved.Name())
	}

	retrievedValue, exists := registry.GetValueConsumer("test-value")
	if !exists {
		t.Fatal("Value consumer not found")
	}
	if retrievedValue.Name() != "test-value" {
		t.Errorf("Expected name 'test-value', got %s", retrievedValue.Name())
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	registry := NewRegistry()

	consumer := &mockSchemaConsumer{
		name:      "duplicate",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	// First registration should succeed
	err := registry.RegisterSchemaConsumer(consumer)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration should fail
	err = registry.RegisterSchemaConsumer(consumer)
	if err == nil {
		t.Fatal("Expected error for duplicate registration")
	}
}

func TestRegistry_ListConsumers(t *testing.T) {
	registry := NewRegistry()

	schemaConsumer := &mockSchemaConsumer{
		name:      "schema1",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	valueConsumer := &mockValueConsumer{
		name:      "value1",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	registry.RegisterSchemaConsumer(schemaConsumer)
	registry.RegisterValueConsumer(valueConsumer)

	schemaNames := registry.ListSchemaConsumers()
	valueNames := registry.ListValueConsumers()

	if len(schemaNames) != 1 || schemaNames[0] != "schema1" {
		t.Errorf("Expected [schema1], got %v", schemaNames)
	}

	if len(valueNames) != 1 || valueNames[0] != "value1" {
		t.Errorf("Expected [value1], got %v", valueNames)
	}
}

func TestRegistry_GetApplicableConsumers(t *testing.T) {
	registry := NewRegistry()

	// Create consumers with different conditions
	stringConsumer := &mockSchemaConsumer{
		name:      "string-consumer",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	emailConsumer := &mockSchemaConsumer{
		name:      "email-consumer",
		purpose:   "validation",
		condition: And(Type(core.TypeString), HasAnnotation("format", "email")),
	}

	registry.RegisterSchemaConsumer(stringConsumer)
	registry.RegisterSchemaConsumer(emailConsumer)

	// Test with string schema (no annotations)
	stringSchema := &mockSchema{
		schemaType:  core.TypeString,
		annotations: []core.Annotation{},
	}

	applicable := registry.GetApplicableSchemaConsumers(stringSchema)
	if len(applicable) != 1 || applicable[0].Name() != "string-consumer" {
		t.Errorf("Expected [string-consumer], got %v", getConsumerNames(applicable))
	}

	// Test with email schema (has format annotation)
	emailSchema := &mockSchema{
		schemaType: core.TypeString,
		annotations: []core.Annotation{
			&mockAnnotation{name: "format", value: "email"},
		},
	}

	applicable = registry.GetApplicableSchemaConsumers(emailSchema)
	names := getConsumerNames(applicable)
	if len(applicable) != 2 {
		t.Errorf("Expected 2 consumers, got %d: %v", len(applicable), names)
	}

	// Both should match
	found := make(map[string]bool)
	for _, consumer := range applicable {
		found[consumer.Name()] = true
	}
	if !found["string-consumer"] || !found["email-consumer"] {
		t.Errorf("Expected both string-consumer and email-consumer, got %v", names)
	}
}

func TestRegistry_ProcessSchemaWithPurpose(t *testing.T) {
	registry := NewRegistry()

	consumer := &mockSchemaConsumer{
		name:      "test-consumer",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	registry.RegisterSchemaConsumer(consumer)

	schema := &mockSchema{
		schemaType:  core.TypeString,
		annotations: []core.Annotation{},
	}

	result, err := registry.ProcessSchemaWithPurpose("validation", schema)
	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	if result.Kind() != "test" {
		t.Errorf("Expected kind 'test', got %s", result.Kind())
	}

	if result.Value() != "schema-processed" {
		t.Errorf("Expected 'schema-processed', got %v", result.Value())
	}
}

func TestRegistry_ProcessValueWithPurpose(t *testing.T) {
	registry := NewRegistry()

	consumer := &mockValueConsumer{
		name:      "test-consumer",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	registry.RegisterValueConsumer(consumer)

	schema := &mockSchema{
		schemaType:  core.TypeString,
		annotations: []core.Annotation{},
	}

	value := &mockValue{val: "test"}

	result, err := registry.ProcessValueWithPurpose("validation", schema, value)
	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	if result.Kind() != "test" {
		t.Errorf("Expected kind 'test', got %s", result.Kind())
	}

	if result.Value() != "value-processed" {
		t.Errorf("Expected 'value-processed', got %v", result.Value())
	}
}

func TestRegistry_NoApplicableConsumers(t *testing.T) {
	registry := NewRegistry()

	schema := &mockSchema{
		schemaType:  core.TypeString,
		annotations: []core.Annotation{},
	}

	// No consumers registered
	_, err := registry.ProcessSchemaWithPurpose("validation", schema)
	if err == nil {
		t.Fatal("Expected error when no consumers available")
	}
}

func TestRegistry_ProcessSchemaAllWithPurpose(t *testing.T) {
	registry := NewRegistry()

	consumer1 := &mockSchemaConsumer{
		name:      "consumer1",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	consumer2 := &mockSchemaConsumer{
		name:      "consumer2",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	registry.RegisterSchemaConsumer(consumer1)
	registry.RegisterSchemaConsumer(consumer2)

	schema := &mockSchema{
		schemaType:  core.TypeString,
		annotations: []core.Annotation{},
	}

	results, err := registry.ProcessSchemaAllWithPurpose("validation", schema)
	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Kind() != "test" {
			t.Errorf("Expected kind 'test', got %s", result.Kind())
		}
		if result.Value() != "schema-processed" {
			t.Errorf("Expected 'schema-processed', got %v", result.Value())
		}
	}
}

func TestRegistry_ProcessSchemaWithPurposes(t *testing.T) {
	registry := NewRegistry()

	validationConsumer := &mockSchemaConsumer{
		name:      "validator",
		purpose:   "validation",
		condition: Type(core.TypeString),
	}

	formattingConsumer := &mockSchemaConsumer{
		name:      "formatter",
		purpose:   "formatting",
		condition: Type(core.TypeString),
	}

	registry.RegisterSchemaConsumer(validationConsumer)
	registry.RegisterSchemaConsumer(formattingConsumer)

	schema := &mockSchema{
		schemaType:  core.TypeString,
		annotations: []core.Annotation{},
	}

	result, err := registry.ProcessSchemaWithPurposes([]ConsumerPurpose{"validation", "formatting"}, schema)
	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful processing")
	}

	if len(result.SchemaResults) != 2 {
		t.Errorf("Expected 2 purpose results, got %d", len(result.SchemaResults))
	}

	if _, ok := result.SchemaResults["validation"]; !ok {
		t.Error("Missing validation results")
	}

	if _, ok := result.SchemaResults["formatting"]; !ok {
		t.Error("Missing formatting results")
	}
}

// Helper function to extract consumer names
func getConsumerNames(consumers []AnnotationConsumer) []string {
	names := make([]string, len(consumers))
	for i, consumer := range consumers {
		names[i] = consumer.Name()
	}
	return names
}
