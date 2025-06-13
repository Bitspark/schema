package builders

import (
	"fmt"
	"reflect"

	"defs.dev/schema/core"
	"defs.dev/schema/schemas"
)

// ServiceBuilder implements core.ServiceSchemaBuilder for creating service schemas.
type ServiceBuilder struct {
	name     string
	methods  map[string]core.FunctionSchema
	examples []map[string]any
	metadata core.SchemaMetadata
}

// Ensure ServiceBuilder implements the API interface at compile time
var _ core.ServiceSchemaBuilder = (*ServiceBuilder)(nil)
var _ core.Builder[core.ServiceSchema] = (*ServiceBuilder)(nil)
var _ core.MetadataBuilder[core.ServiceSchemaBuilder] = (*ServiceBuilder)(nil)

// NewServiceSchema creates a new ServiceBuilder.
func NewServiceSchema() *ServiceBuilder {
	return &ServiceBuilder{
		name:     "",
		methods:  make(map[string]core.FunctionSchema),
		examples: []map[string]any{},
		metadata: core.SchemaMetadata{},
	}
}

// Core builder methods (API compliance)

func (b *ServiceBuilder) Method(name string, functionSchema core.FunctionSchema) core.ServiceSchemaBuilder {
	b.methods[name] = functionSchema
	return b
}

func (b *ServiceBuilder) FromStruct(instance any) core.ServiceSchemaBuilder {
	if instance == nil {
		return b
	}

	instanceType := reflect.TypeOf(instance)
	instanceValue := reflect.ValueOf(instance)

	// Handle pointer to struct
	if instanceType.Kind() == reflect.Ptr {
		if instanceValue.IsNil() {
			return b
		}
		instanceType = instanceType.Elem()
		instanceValue = instanceValue.Elem()
	}

	if instanceType.Kind() != reflect.Struct {
		return b
	}

	// Set service name from struct type
	if b.name == "" {
		b.name = instanceType.Name()
	}

	// Discover methods from the struct
	actualType := reflect.TypeOf(instance)
	for i := 0; i < actualType.NumMethod(); i++ {
		method := actualType.Method(i)

		// Only include exported methods that look like service methods
		if method.IsExported() && isValidServiceMethod(method) {
			// Create a basic function schema for the method
			// This is a simplified version - in a real implementation,
			// you'd want more sophisticated reflection
			functionSchema := createFunctionSchemaFromMethod(method)
			b.methods[method.Name] = functionSchema
		}
	}

	return b
}

func (b *ServiceBuilder) Example(example map[string]any) core.ServiceSchemaBuilder {
	b.examples = append(b.examples, example)
	return b
}

// Metadata builder methods (API compliance)

func (b *ServiceBuilder) Description(desc string) core.ServiceSchemaBuilder {
	b.metadata.Description = desc
	return b
}

func (b *ServiceBuilder) Name(name string) core.ServiceSchemaBuilder {
	b.name = name
	b.metadata.Name = name
	return b
}

func (b *ServiceBuilder) Tag(tag string) core.ServiceSchemaBuilder {
	b.metadata.Tags = append(b.metadata.Tags, tag)
	return b
}

// Build creates the final ServiceSchema
func (b *ServiceBuilder) Build() core.ServiceSchema {
	serviceSchema := schemas.NewServiceSchema(b.name)

	// Add all methods
	for methodName, functionSchema := range b.methods {
		serviceSchema = serviceSchema.WithMethod(methodName, functionSchema)
	}

	// Prepare metadata with examples
	metadata := b.metadata
	if len(b.examples) > 0 {
		// Convert examples to []any for metadata
		metadataExamples := make([]any, len(b.examples))
		for i, example := range b.examples {
			metadataExamples[i] = example
		}
		metadata.Examples = metadataExamples
	}

	// Apply metadata
	if metadata.Description != "" || len(metadata.Tags) > 0 || len(metadata.Examples) > 0 {
		serviceSchema = serviceSchema.WithMetadata(metadata)
	}

	return serviceSchema
}

// Helper methods

// isValidServiceMethod checks if a method is suitable for service schema generation
func isValidServiceMethod(method reflect.Method) bool {
	methodType := method.Type

	// Must have at least receiver
	if methodType.NumIn() < 1 {
		return false
	}

	// Skip methods that return functions or channels (likely not service methods)
	for i := 0; i < methodType.NumOut(); i++ {
		out := methodType.Out(i)
		if out.Kind() == reflect.Func || out.Kind() == reflect.Chan {
			return false
		}
	}

	// Skip common non-service methods
	switch method.Name {
	case "String", "GoString", "Error", "Format":
		return false
	}

	return true
}

// createFunctionSchemaFromMethod creates a basic function schema from a reflect.Method
func createFunctionSchemaFromMethod(method reflect.Method) core.FunctionSchema {
	// This is a simplified implementation
	// In a real system, you'd want more sophisticated type analysis

	builder := NewFunctionSchema()
	builder.Name(method.Name)

	methodType := method.Type

	// Add inputs (skip receiver at index 0)
	for i := 1; i < methodType.NumIn(); i++ {
		inputType := methodType.In(i)
		inputName := fmt.Sprintf("param%d", i-1)

		// Create a basic schema based on the type
		inputSchema := createSchemaFromType(inputType)
		builder.Input(inputName, inputSchema)
	}

	// Add outputs
	if methodType.NumOut() > 0 {
		for i := 0; i < methodType.NumOut(); i++ {
			outputType := methodType.Out(i)
			outputName := fmt.Sprintf("output%d", i)

			// Special handling for error type
			if outputType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				// This is an error output
				errorSchema := createSchemaFromType(outputType)
				builder.Error(errorSchema)
			} else {
				// Regular output
				outputSchema := createSchemaFromType(outputType)
				builder.Output(outputName, outputSchema)
			}
		}
	}

	return builder.Build()
}

// createSchemaFromType creates a basic schema from a reflect.Type
func createSchemaFromType(t reflect.Type) core.Schema {
	// This is a very basic implementation
	// In a real system, you'd want comprehensive type mapping

	switch t.Kind() {
	case reflect.String:
		return NewStringSchema().Build()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewIntegerSchema().Build()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewIntegerSchema().Build()
	case reflect.Float32, reflect.Float64:
		return NewNumberSchema().Build()
	case reflect.Bool:
		return NewBooleanSchema().Build()
	case reflect.Slice, reflect.Array:
		itemSchema := createSchemaFromType(t.Elem())
		return NewArraySchema().Items(itemSchema).Build()
	case reflect.Map:
		// For now, treat maps as objects with additional properties
		return NewObject().AdditionalProperties(true).Build()
	case reflect.Struct:
		// For structs, create an object schema with properties based on fields
		builder := NewObject()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				fieldSchema := createSchemaFromType(field.Type)
				builder.Property(field.Name, fieldSchema)
			}
		}
		return builder.Build()
	case reflect.Ptr:
		// For pointers, create schema for the element type
		return createSchemaFromType(t.Elem())
	case reflect.Interface:
		// For interfaces, use a flexible object schema
		return NewObject().AdditionalProperties(true).Build()
	default:
		// Fallback to a flexible object schema
		return NewObject().AdditionalProperties(true).Build()
	}
}

// Extended builder methods (beyond API requirements)

// ServiceName sets the service name (alias for Name)
func (b *ServiceBuilder) ServiceName(name string) *ServiceBuilder {
	b.Name(name)
	return b
}

// AddMethod adds a method with a function schema (alias for Method)
func (b *ServiceBuilder) AddMethod(name string, functionSchema core.FunctionSchema) *ServiceBuilder {
	b.Method(name, functionSchema)
	return b
}

// SimpleMethod creates a simple method with basic input/output
func (b *ServiceBuilder) SimpleMethod(name string, inputSchema core.Schema, outputSchema core.Schema) *ServiceBuilder {
	functionSchema := NewFunctionSchema().
		Input("input", inputSchema).
		Output("output", outputSchema).
		Name(name).
		Build()

	b.Method(name, functionSchema)
	return b
}

// VoidMethod creates a method that takes no inputs and returns no outputs
func (b *ServiceBuilder) VoidMethod(name string) *ServiceBuilder {
	functionSchema := NewFunctionSchema().
		Name(name).
		Build()

	b.Method(name, functionSchema)
	return b
}

// Domain-specific service examples

// RESTService creates a service schema for REST API services
func (b *ServiceBuilder) RESTService() *ServiceBuilder {
	b.Tag("rest")
	b.Tag("api")
	b.Description("REST API service")
	return b
}

// DatabaseService creates a service schema for database services
func (b *ServiceBuilder) DatabaseService() *ServiceBuilder {
	b.Tag("database")
	b.Tag("persistence")
	b.Description("Database service")
	return b
}

// BusinessService creates a service schema for business logic services
func (b *ServiceBuilder) BusinessService() *ServiceBuilder {
	b.Tag("business")
	b.Tag("domain")
	b.Description("Business logic service")
	return b
}

// ValidationService creates a service schema for validation services
func (b *ServiceBuilder) ValidationService() *ServiceBuilder {
	b.Tag("validation")
	b.Tag("rules")
	b.Description("Validation service")
	return b
}

// Common service patterns

// CRUDService creates a basic CRUD service with standard methods
func (b *ServiceBuilder) CRUDService(entityName string) *ServiceBuilder {
	b.Name(fmt.Sprintf("%sService", entityName))
	b.Description(fmt.Sprintf("CRUD service for %s entities", entityName))

	// Create
	b.SimpleMethod("Create",
		NewObject().Build(), // Input: entity data
		NewObject().Build()) // Output: created entity

	// Read
	b.SimpleMethod("Get",
		NewStringSchema().Build(), // Input: ID
		NewObject().Build())       // Output: entity

	// Update
	b.SimpleMethod("Update",
		NewObject().Build(), // Input: updated entity
		NewObject().Build()) // Output: updated entity

	// Delete
	b.SimpleMethod("Delete",
		NewStringSchema().Build(),  // Input: ID
		NewBooleanSchema().Build()) // Output: success

	// List
	b.SimpleMethod("List",
		NewObject().Build(),      // Input: query parameters
		NewArraySchema().Build()) // Output: array of entities

	return b
}

// EventService creates a service for handling events
func (b *ServiceBuilder) EventService() *ServiceBuilder {
	b.Name("EventService")
	b.Description("Event handling service")
	b.Tag("events")
	b.Tag("messaging")

	b.SimpleMethod("Publish",
		NewObject().Build(),        // Input: event data
		NewBooleanSchema().Build()) // Output: success

	b.SimpleMethod("Subscribe",
		NewStringSchema().Build(),  // Input: event type
		NewBooleanSchema().Build()) // Output: success

	return b
}
