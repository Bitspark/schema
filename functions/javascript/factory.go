package javascript

import (
	"fmt"

	"defs.dev/schema"
	"defs.dev/schema/functions"
)

// NewRegistry creates a new registry with JavaScript portal using default configuration
func NewRegistry() functions.Registry {
	portal := NewPortalWithDefaults()
	return functions.NewRegistry(portal)
}

// NewRegistryWithConfig creates a new registry with JavaScript portal using custom configuration
func NewRegistryWithConfig(config Config) functions.Registry {
	portal := NewPortal(config)
	return functions.NewRegistry(portal)
}

// NewConsumer creates a new consumer with JavaScript portal registered
func NewConsumer() functions.Consumer {
	consumer := functions.NewConsumer()
	portal := NewPortalWithDefaults()
	consumer.RegisterPortal(&PortalWrapper{portal})
	return consumer
}

// NewConsumerWithConfig creates a new consumer with JavaScript portal using custom configuration
func NewConsumerWithConfig(config Config) functions.Consumer {
	consumer := functions.NewConsumer()
	portal := NewPortal(config)
	consumer.RegisterPortal(&PortalWrapper{portal})
	return consumer
}

// PortalWrapper wraps a typed JavaScriptPortal to implement Portal[any]
type PortalWrapper struct {
	*JavaScriptPortal
}

func (w *PortalWrapper) Apply(address string, funcSchema *schema.FunctionSchema, data any) schema.Function {
	// Type assert the data to JSFunction
	jsFunction, ok := data.(JSFunction)
	if !ok {
		panic(fmt.Sprintf("JavaScript PortalWrapper: expected JSFunction, got %T", data))
	}
	return w.JavaScriptPortal.Apply(address, funcSchema, jsFunction)
}

func (w *PortalWrapper) GenerateAddress(name string, data any) string {
	// Type assert the data to JSFunction
	jsFunction, ok := data.(JSFunction)
	if !ok {
		panic(fmt.Sprintf("JavaScript PortalWrapper: expected JSFunction, got %T", data))
	}
	return w.JavaScriptPortal.GenerateAddress(name, jsFunction)
}

// Utility function to create a complete JavaScript function system
func NewJavaScriptSystem(config Config) (*JavaScriptPortal, functions.Registry, functions.Consumer) {
	portal := NewPortal(config)
	registry := functions.NewRegistry(portal)
	consumer := functions.NewConsumer()
	consumer.RegisterPortal(&PortalWrapper{portal})

	return portal, registry, consumer
}

// Convenience function for quick setup with defaults
func NewDefaultJavaScriptSystem() (*JavaScriptPortal, functions.Registry, functions.Consumer) {
	return NewJavaScriptSystem(DefaultConfig())
}
