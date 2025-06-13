package schemas

import (
	"fmt"
	"strings"

	"defs.dev/schema/core"
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

func (s *ServiceMethodSchema) Annotations() []core.Annotation {
	return []core.Annotation{}
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

// Annotations returns the annotations of the schema.
func (s *ServiceSchema) Annotations() []core.Annotation {
	return []core.Annotation{}
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
