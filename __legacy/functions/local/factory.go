package local

import (
	"defs.dev/schema"
	"defs.dev/schema/functions"
)

// NewRegistry creates a registry that uses the local portal for in-process functions
func NewRegistry() functions.Registry {
	return functions.NewRegistry(NewPortal())
}

// NewConsumer creates a consumer with the local portal pre-registered
func NewConsumer() functions.Consumer {
	consumer := functions.NewConsumer()
	consumer.RegisterPortal(&functions.PortalWrapper[schema.FunctionHandler]{Portal: NewPortal()})
	return consumer
}
