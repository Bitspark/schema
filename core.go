package schema

import (
	"defs.dev/schema/api"
	"defs.dev/schema/builders"
	"defs.dev/schema/portal"
	"defs.dev/schema/registry"
)

// Factory functions for creating schema builders
// These provide the main entry points for the core package

// NewString creates a new string schema builder.
func NewString() api.StringSchemaBuilder {
	return builders.NewStringSchema()
}

// NewNumber creates a new number schema builder.
func NewNumber() api.NumberSchemaBuilder {
	return builders.NewNumberSchema()
}

// NewInteger creates a new integer schema builder.
func NewInteger() api.IntegerSchemaBuilder {
	return builders.NewIntegerSchema()
}

// NewBoolean creates a new boolean schema builder.
func NewBoolean() api.BooleanSchemaBuilder {
	return builders.NewBooleanSchema()
}

// NewArray creates a new array schema builder.
func NewArray() api.ArraySchemaBuilder {
	return builders.NewArraySchema()
}

// NewObject creates a new object schema builder.
func NewObject() api.ObjectSchemaBuilder {
	return builders.NewObject()
}

// NewFunction creates a new function schema builder.
func NewFunction() api.FunctionSchemaBuilder {
	return builders.NewFunctionSchema()
}

// NewService creates a new ServiceBuilder for building service schemas.
func NewService() api.ServiceSchemaBuilder {
	return builders.NewServiceSchema()
}

// Registry and Factory functions

// NewFunctionRegistry creates a new function registry for managing callable functions.
func NewFunctionRegistry() api.Registry {
	return registry.NewFunctionRegistry()
}

// NewServiceRegistry creates a new service registry for managing services and their methods.
func NewServiceRegistry() *registry.ServiceRegistry {
	return registry.NewServiceRegistry()
}

// NewFactory creates a new factory for creating registries and other components.
func NewFactory() api.Factory {
	return registry.NewFactory()
}

// TODO: Add other schema type factory functions as we implement them
// func NewUnion() api.UnionSchemaBuilder { return builders.NewUnion() }

// Portal system factory functions

// NewLocalPortal creates a new local portal for in-process function execution
func NewLocalPortal() api.LocalPortal {
	return portal.NewLocalPortal()
}

// NewTestingPortal creates a new testing portal for mock/stub functionality
func NewTestingPortal() api.TestingPortal {
	return portal.NewTestingPortal()
}

// NewPortalRegistry creates a new portal registry for managing multiple portals
func NewPortalRegistry() api.PortalRegistry {
	return portal.NewPortalRegistry()
}

// NewDefaultPortalRegistry creates a portal registry with common portals pre-registered
func NewDefaultPortalRegistry() api.PortalRegistry {
	return portal.NewDefaultPortalRegistry()
}

// NewHTTPPortal creates a new HTTP portal with default configuration
func NewHTTPPortal() api.HTTPPortal {
	return portal.NewHTTPPortal(nil)
}

// NewHTTPPortalWithConfig creates a new HTTP portal with custom configuration
func NewHTTPPortalWithConfig(config any) api.HTTPPortal {
	if httpConfig, ok := config.(*portal.HTTPConfig); ok {
		return portal.NewHTTPPortal(httpConfig)
	}
	return portal.NewHTTPPortal(nil)
}

// NewWebSocketPortal creates a new WebSocket portal with default configuration
func NewWebSocketPortal() api.WebSocketPortal {
	return portal.NewWebSocketPortal(nil)
}

// NewWebSocketPortalWithConfig creates a new WebSocket portal with custom configuration
func NewWebSocketPortalWithConfig(config any) api.WebSocketPortal {
	if wsConfig, ok := config.(*portal.WebSocketConfig); ok {
		return portal.NewWebSocketPortal(wsConfig)
	}
	return portal.NewWebSocketPortal(nil)
}

// Address system factory functions

// NewAddress creates a new Address from a URL string
func NewAddress(addressStr string) (api.Address, error) {
	return portal.NewAddress(addressStr)
}

// MustNewAddress creates a new Address, panicking on error
func MustNewAddress(addressStr string) api.Address {
	return portal.MustNewAddress(addressStr)
}

// NewAddressBuilder creates a new AddressBuilder for fluent address construction
func NewAddressBuilder() api.AddressBuilder {
	return portal.NewAddressBuilder()
}

// LocalAddress creates a local address for a function name
func LocalAddress(functionName string) api.Address {
	return portal.LocalAddress(functionName)
}

// HTTPAddress creates an HTTP address
func HTTPAddress(host string, port int, path string) api.Address {
	return portal.HTTPAddress(host, port, path)
}

// HTTPSAddress creates an HTTPS address
func HTTPSAddress(host string, path string) api.Address {
	return portal.HTTPSAddress(host, path)
}

// WebSocketAddress creates a WebSocket address
func WebSocketAddress(host string, port int, path string) api.Address {
	return portal.WebSocketAddress(host, port, path)
}

// WebSocketSecureAddress creates a secure WebSocket address
func WebSocketSecureAddress(host string, path string) api.Address {
	return portal.WebSocketSecureAddress(host, path)
}

// Function utilities

// NewFunctionData creates a new FunctionData from a map
func NewFunctionData(data map[string]any) api.FunctionData {
	return portal.NewFunctionData(data)
}

// NewFunctionDataValue creates FunctionData from a single value
func NewFunctionDataValue(value any) api.FunctionData {
	return portal.NewFunctionDataValue(value)
}

// Legacy function utilities (deprecated)

// NewFunctionInputMap creates a new FunctionInputMap (deprecated, use NewFunctionData)
func NewFunctionInputMap() portal.FunctionInputMap {
	return make(portal.FunctionInputMap)
}

// NewFunctionOutput creates a new FunctionOutput with the given value (deprecated, use NewFunctionData)
// This function has been removed because api.FunctionOutput no longer exists in the new API.
// Use NewFunctionDataValue instead.
func NewFunctionOutput(value any) any {
	return portal.NewFunctionDataValue(value)
}
