package http

import (
	"testing"

	"defs.dev/schema"
)

func TestHTTPClientFunctions(t *testing.T) {
	t.Run("HTTPClientFunction Schema", func(t *testing.T) {
		// Create a test schema
		testSchema := schema.NewFunctionSchema().
			Name("testFunction").
			Input("message", schema.String().Build()).
			Output(schema.String().Build()).
			Build().(*schema.FunctionSchema)
		
		// Create a client function
		clientFunc := &HTTPClientFunction{
			address: "http://example.com/api/test",
			schema:  testSchema,
		}
		
		result := clientFunc.Schema()
		if result != testSchema {
			t.Error("Expected Schema() to return the assigned schema")
		}
	})
	
	t.Run("HTTPEndpointFunction Schema", func(t *testing.T) {
		// Create a test schema
		testSchema := schema.NewFunctionSchema().
			Name("endpointFunction").
			Input("id", schema.Integer().Build()).
			Output(schema.Object().Build()).
			Build().(*schema.FunctionSchema)
		
		// Create an endpoint function
		endpointFunc := &HTTPEndpointFunction{
			address: "http://api.example.com/endpoint",
			schema:  testSchema,
		}
		
		result := endpointFunc.Schema()
		if result != testSchema {
			t.Error("Expected Schema() to return the assigned schema")
		}
	})
	
	t.Run("HTTPEndpointFunction Address", func(t *testing.T) {
		address := "http://api.example.com/endpoint"
		
		endpointFunc := &HTTPEndpointFunction{
			address: address,
		}
		
		result := endpointFunc.Address()
		if result != address {
			t.Errorf("Expected Address() to return '%s', got '%s'", address, result)
		}
	})
}