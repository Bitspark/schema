package schemas

import (
	"fmt"
	"reflect"
	"strings"

	"defs.dev/schema/api/core"
)

// ServiceMethodSchema represents a single method in a service schema.
type ServiceMethodSchema struct {
	name     string              `json:"name"`
	function core.FunctionSchema `json:"function"`
	metadata core.SchemaMetadata `json:"metadata,omitempty"`
}

// Ensure ServiceMethodSchema implements the API interface at compile time
var _ core.ServiceMethodSchema = (*ServiceMethodSchema)(nil)

// NewServiceMethodSchema creates a new ServiceMethodSchema.
func NewServiceMethodSchema(name string, functionSchema core.FunctionSchema) *ServiceMethodSchema {
	return &ServiceMethodSchema{
		name:     name,
		function: functionSchema,
		metadata: core.SchemaMetadata{},
	}
}

// Core Schema interface implementation for ServiceMethodSchema

func (s *ServiceMethodSchema) Type() core.SchemaType {
	return core.TypeFunction // Service method is essentially a function
}

func (s *ServiceMethodSchema) Metadata() core.SchemaMetadata {
	return s.metadata
}

func (s *ServiceMethodSchema) Validate(value any) core.ValidationResult {
	// Delegate validation to the underlying function schema
	return s.function.Validate(value)
}

func (s *ServiceMethodSchema) ToJSONSchema() map[string]any {
	schema := s.function.ToJSONSchema()
	// Add service method specific metadata
	if s.name != "" {
		schema["x-method-name"] = s.name
	}
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}
	return schema
}

func (s *ServiceMethodSchema) GenerateExample() any {
	return s.function.GenerateExample()
}

func (s *ServiceMethodSchema) Clone() core.Schema {
	return &ServiceMethodSchema{
		name:     s.name,
		function: s.function.Clone().(core.FunctionSchema),
		metadata: s.metadata,
	}
}

// ServiceMethodSchema interface implementation

func (s *ServiceMethodSchema) Name() string {
	return s.name
}

func (s *ServiceMethodSchema) Function() core.FunctionSchema {
	return s.function
}

// Visitor pattern support

func (s *ServiceMethodSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitFunction(s.function)
}

// ServiceSchema represents a service contract with multiple methods.
type ServiceSchema struct {
	name     string                 `json:"name"`
	methods  []*ServiceMethodSchema `json:"methods"`
	metadata core.SchemaMetadata    `json:"metadata,omitempty"`
}

// Ensure ServiceSchema implements the API interface at compile time
var _ core.ServiceSchema = (*ServiceSchema)(nil)
var _ core.Schema = (*ServiceSchema)(nil)
var _ core.Accepter = (*ServiceSchema)(nil)

// NewServiceSchema creates a new ServiceSchema.
func NewServiceSchema(name string) *ServiceSchema {
	return &ServiceSchema{
		name:     name,
		methods:  []*ServiceMethodSchema{},
		metadata: core.SchemaMetadata{},
	}
}

// Core Schema interface implementation

func (s *ServiceSchema) Type() core.SchemaType {
	return core.TypeService
}

func (s *ServiceSchema) Metadata() core.SchemaMetadata {
	return s.metadata
}

func (s *ServiceSchema) Validate(value any) core.ValidationResult {
	// Service schema validation could validate:
	// 1. Service instance conformity
	// 2. Method availability
	// 3. Service contract compliance

	switch v := value.(type) {
	case map[string]any:
		return s.validateServiceData(v)
	default:
		// Try reflection-based validation for struct instances
		return s.validateServiceStruct(value)
	}
}

// validateServiceData validates service data as a map
func (s *ServiceSchema) validateServiceData(data map[string]any) core.ValidationResult {
	var errors []core.ValidationError

	// Check if all required methods are represented
	methodMap := make(map[string]bool)
	for _, method := range s.methods {
		methodMap[method.name] = false
	}

	// Mark methods as found and validate their data
	for key, value := range data {
		if _, exists := methodMap[key]; exists {
			methodMap[key] = true
			// Could validate method-specific data here
			_ = value
		}
	}

	// Check for missing required methods
	for methodName, found := range methodMap {
		if !found {
			errors = append(errors, core.ValidationError{
				Path:       methodName,
				Message:    fmt.Sprintf("required service method '%s' not found", methodName),
				Code:       "missing_service_method",
				Expected:   fmt.Sprintf("method '%s' implementation", methodName),
				Suggestion: fmt.Sprintf("implement method '%s' in the service", methodName),
				Context:    "service_validation",
			})
		}
	}

	return core.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Metadata: map[string]any{
			"service_name":   s.name,
			"method_count":   len(s.methods),
			"validated_type": "service_data",
		},
	}
}

// validateServiceStruct validates a service struct using reflection
func (s *ServiceSchema) validateServiceStruct(instance any) core.ValidationResult {
	if instance == nil {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{{
				Path:       "",
				Message:    "service instance cannot be nil",
				Code:       "nil_service_instance",
				Expected:   "non-nil service instance",
				Suggestion: "provide a valid service instance",
				Context:    "service_validation",
			}},
		}
	}

	var errors []core.ValidationError
	instanceType := reflect.TypeOf(instance)
	instanceValue := reflect.ValueOf(instance)

	// Handle pointer to struct
	if instanceType.Kind() == reflect.Ptr {
		if instanceValue.IsNil() {
			return core.ValidationResult{
				Valid: false,
				Errors: []core.ValidationError{{
					Path:       "",
					Message:    "service instance pointer is nil",
					Code:       "nil_service_pointer",
					Expected:   "non-nil service pointer",
					Suggestion: "provide a valid service instance",
					Context:    "service_validation",
				}},
			}
		}
		instanceType = instanceType.Elem()
		instanceValue = instanceValue.Elem()
	}

	if instanceType.Kind() != reflect.Struct {
		return core.ValidationResult{
			Valid: false,
			Errors: []core.ValidationError{{
				Path:       "",
				Message:    fmt.Sprintf("service instance must be a struct, got %T", instance),
				Code:       "invalid_service_type",
				Value:      instance,
				Expected:   "struct instance",
				Suggestion: "provide a struct instance that implements the service",
				Context:    "service_validation",
			}},
		}
	}

	// Check if all service methods are available on the struct
	for _, methodSchema := range s.methods {
		_, found := instanceType.MethodByName(methodSchema.name)
		if !found {
			errors = append(errors, core.ValidationError{
				Path:       methodSchema.name,
				Message:    fmt.Sprintf("service method '%s' not found on struct", methodSchema.name),
				Code:       "missing_service_method",
				Expected:   fmt.Sprintf("method '%s' on struct", methodSchema.name),
				Suggestion: fmt.Sprintf("implement method '%s' on the service struct", methodSchema.name),
				Context:    "service_validation",
			})
		}
	}

	return core.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
		Metadata: map[string]any{
			"service_name":   s.name,
			"service_type":   instanceType.Name(),
			"method_count":   len(s.methods),
			"validated_type": "service_struct",
		},
	}
}

func (s *ServiceSchema) ToJSONSchema() map[string]any {
	schema := map[string]any{
		"type":       "object",
		"x-service":  true,
		"properties": make(map[string]any),
	}

	// Add service metadata
	if s.name != "" {
		schema["title"] = s.name
	}
	if s.metadata.Description != "" {
		schema["description"] = s.metadata.Description
	}

	// Add methods as properties
	properties := schema["properties"].(map[string]any)
	for _, method := range s.methods {
		properties[method.name] = method.ToJSONSchema()
	}

	// Add service-specific metadata
	methodNames := make([]string, len(s.methods))
	for i, method := range s.methods {
		methodNames[i] = method.name
	}
	schema["x-methods"] = methodNames

	if len(s.metadata.Tags) > 0 {
		schema["x-tags"] = s.metadata.Tags
	}

	return schema
}

func (s *ServiceSchema) GenerateExample() any {
	example := make(map[string]any)

	// Generate examples for each method
	for _, method := range s.methods {
		example[method.name] = method.GenerateExample()
	}

	return example
}

func (s *ServiceSchema) Clone() core.Schema {
	clonedMethods := make([]*ServiceMethodSchema, len(s.methods))
	for i, method := range s.methods {
		clonedMethods[i] = method.Clone().(*ServiceMethodSchema)
	}

	// Clone metadata
	clonedMetadata := core.SchemaMetadata{
		Name:        s.metadata.Name,
		Description: s.metadata.Description,
		Examples:    append([]any(nil), s.metadata.Examples...),
		Tags:        append([]string(nil), s.metadata.Tags...),
	}

	if s.metadata.Properties != nil {
		clonedMetadata.Properties = make(map[string]string)
		for k, v := range s.metadata.Properties {
			clonedMetadata.Properties[k] = v
		}
	}

	return &ServiceSchema{
		name:     s.name,
		methods:  clonedMethods,
		metadata: clonedMetadata,
	}
}

// ServiceSchema interface implementation

func (s *ServiceSchema) Name() string {
	return s.name
}

func (s *ServiceSchema) Methods() []core.ServiceMethodSchema {
	methods := make([]core.ServiceMethodSchema, len(s.methods))
	for i, method := range s.methods {
		methods[i] = method
	}
	return methods
}

// Visitor pattern support

func (s *ServiceSchema) Accept(visitor core.SchemaVisitor) error {
	return visitor.VisitService(s)
}

// Additional utility methods

// WithMetadata creates a new ServiceSchema with updated metadata
func (s *ServiceSchema) WithMetadata(metadata core.SchemaMetadata) *ServiceSchema {
	clone := s.Clone().(*ServiceSchema)
	clone.metadata = metadata
	return clone
}

// WithMethod adds or updates a method in the service schema
func (s *ServiceSchema) WithMethod(name string, functionSchema core.FunctionSchema) *ServiceSchema {
	clone := s.Clone().(*ServiceSchema)

	// Check if method already exists and update it
	for i, method := range clone.methods {
		if method.name == name {
			clone.methods[i] = NewServiceMethodSchema(name, functionSchema)
			return clone
		}
	}

	// Add new method
	clone.methods = append(clone.methods, NewServiceMethodSchema(name, functionSchema))
	return clone
}

// WithName creates a new ServiceSchema with updated name
func (s *ServiceSchema) WithName(name string) *ServiceSchema {
	clone := s.Clone().(*ServiceSchema)
	clone.name = name
	return clone
}

// Introspection methods

// MethodNames returns a list of all method names
func (s *ServiceSchema) MethodNames() []string {
	names := make([]string, len(s.methods))
	for i, method := range s.methods {
		names[i] = method.name
	}
	return names
}

// GetMethod returns a method schema by name
func (s *ServiceSchema) GetMethod(name string) (core.ServiceMethodSchema, bool) {
	for _, method := range s.methods {
		if method.name == name {
			return method, true
		}
	}
	return nil, false
}

// HasMethod checks if a method exists in the service
func (s *ServiceSchema) HasMethod(name string) bool {
	_, exists := s.GetMethod(name)
	return exists
}

// MethodCount returns the number of methods in the service
func (s *ServiceSchema) MethodCount() int {
	return len(s.methods)
}

// String representation for debugging
func (s *ServiceSchema) String() string {
	methodNames := make([]string, len(s.methods))
	for i, method := range s.methods {
		methodNames[i] = method.name
	}

	name := s.name
	if name == "" {
		name = "anonymous"
	}

	return fmt.Sprintf("ServiceSchema(%s: methods=[%s])", name, strings.Join(methodNames, ", "))
}
