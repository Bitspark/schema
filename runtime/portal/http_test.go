package portal

import (
	"bytes"
	"context"
	"defs.dev/schema/construct/builders"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"defs.dev/schema/api"
	"defs.dev/schema/core"
)

func TestHTTPPortal_Creation(t *testing.T) {
	// Test with default config
	portal := NewHTTPPortal(nil)
	if portal == nil {
		t.Fatal("Expected portal to be created with default config")
	}

	// Test schemes
	schemes := portal.Schemes()
	if len(schemes) != 1 || schemes[0] != "http" {
		t.Errorf("Expected schemes [http], got %v", schemes)
	}

	// Test with custom config
	config := &HTTPConfig{
		Host: "localhost",
		Port: 9090,
		TLS:  &TLSConfig{CertFile: "cert.pem", KeyFile: "key.pem"},
	}
	portal = NewHTTPPortal(config)

	schemes = portal.Schemes()
	if len(schemes) != 2 || schemes[0] != "http" || schemes[1] != "https" {
		t.Errorf("Expected schemes [http, https], got %v", schemes)
	}

	if portal.BaseURL() != "https://localhost:9090" {
		t.Errorf("Expected base URL https://localhost:9090, got %s", portal.BaseURL())
	}
}

func TestHTTPPortal_FunctionRegistration(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())
	ctx := context.Background()

	// Create a test function
	testFunc := &HTTPTestFunction{
		name: "testFunc",
		schema: builders.NewFunctionSchema().
			Name("testFunc").
			Description("Test function").
			Input("name", builders.NewStringSchema().Build()).
			RequiredInputs("name").
			Output("result", builders.NewStringSchema().Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			name, _ := params.Get("name")
			return api.NewFunctionData(map[string]any{"result": fmt.Sprintf("Hello, %s!", name)}), nil
		},
	}

	// Test function registration
	address, err := portal.Apply(ctx, testFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	if address.Scheme() != "http" {
		t.Errorf("Expected scheme http, got %s", address.Scheme())
	}

	if address.Path() != "/functions/testFunc" {
		t.Errorf("Expected path /functions/testFunc, got %s", address.Path())
	}

	// Test duplicate registration
	_, err = portal.Apply(ctx, testFunc)
	if err == nil {
		t.Error("Expected error for duplicate function registration")
	}
}

func TestHTTPPortal_ServiceRegistration(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())
	ctx := context.Background()

	// Create a test service
	testService := &TestService{
		name: "TestService",
		schema: builders.NewServiceSchema().
			Name("TestService").
			Method("greet", builders.NewFunctionSchema().
				Name("greet").
				Description("Greet method").
				Input("name", builders.NewStringSchema().Build()).
				RequiredInputs("name").
				Output("greeting", builders.NewStringSchema().Build()).
				Build()).
			Build(),
		greetHandler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			name, _ := params.Get("name")
			return api.NewFunctionData(map[string]any{"greeting": fmt.Sprintf("Hello, %s!", name)}), nil
		},
	}

	// Test service registration
	address, err := portal.ApplyService(ctx, testService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	if address.Scheme() != "http" {
		t.Errorf("Expected scheme http, got %s", address.Scheme())
	}

	if address.Path() != "/services/TestService" {
		t.Errorf("Expected path /services/TestService, got %s", address.Path())
	}
}

func TestHTTPPortal_FunctionResolution(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())
	ctx := context.Background()

	// Register a function first
	testFunc := &HTTPTestFunction{
		name:   "testFunc",
		schema: builders.NewFunctionSchema().Name("testFunc").Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			return api.NewFunctionData(map[string]any{"result": "test result"}), nil
		},
	}

	address, err := portal.Apply(ctx, testFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Test local function resolution
	function, err := portal.ResolveFunction(ctx, address)
	if err != nil {
		t.Fatalf("Failed to resolve function: %v", err)
	}

	if function.Name() != "testFunc" {
		t.Errorf("Expected function name testFunc, got %s", function.Name())
	}

	// Test remote function resolution
	remoteAddress := MustNewAddress("http://example.com:8080/functions/remoteFunc")
	remoteFunction, err := portal.ResolveFunction(ctx, remoteAddress)
	if err != nil {
		t.Fatalf("Failed to resolve remote function: %v", err)
	}

	// Should be a RemoteFunction
	if remoteFunction.Name() == "" {
		t.Error("Expected remote function to have a name")
	}

	// Test unsupported scheme
	wsAddress := MustNewAddress("ws://localhost:8080/functions/testFunc")
	_, err = portal.ResolveFunction(ctx, wsAddress)
	if err == nil {
		t.Error("Expected error for unsupported scheme")
	}
}

func TestHTTPPortal_HTTPHandler(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())
	ctx := context.Background()

	// Register a test function
	testFunc := &HTTPTestFunction{
		name:   "echo",
		schema: builders.NewFunctionSchema().Name("echo").Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			message, _ := params.Get("message")
			return api.NewFunctionData(map[string]any{"echo": message}), nil
		},
	}

	_, err := portal.Apply(ctx, testFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Create test server
	server := httptest.NewServer(portal.HandleHTTP().(http.Handler))
	defer server.Close()

	// Test successful function call
	requestData := map[string]any{"message": "hello world"}
	requestBody, _ := json.Marshal(requestData)

	resp, err := http.Post(server.URL+"/functions/echo", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if result["echo"] != "hello world" {
		t.Errorf("Expected echo to be 'hello world', got %v", result["echo"])
	}
}

func TestHTTPPortal_HTTPErrorHandling(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())

	// Register a test function for the invalid JSON and method not allowed tests
	echoFunc := &HTTPTestFunction{
		name:   "echo",
		schema: nil, // Simple function without schema validation
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			return params, nil
		},
	}

	ctx := context.Background()
	_, err := portal.Apply(ctx, echoFunc)
	if err != nil {
		t.Fatalf("Failed to register echo function: %v", err)
	}

	server := httptest.NewServer(portal.HandleHTTP().(http.Handler))
	defer server.Close()

	// Test function not found
	requestBody := bytes.NewBuffer([]byte("{}"))
	resp, err := http.Post(server.URL+"/functions/nonexistent", "application/json", requestBody)
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	// Test invalid JSON
	requestBody = bytes.NewBuffer([]byte("invalid json"))
	resp, err = http.Post(server.URL+"/functions/echo", "application/json", requestBody)
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Test method not allowed
	resp, err = http.Get(server.URL + "/functions/echo")
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestHTTPPortal_CORS(t *testing.T) {
	config := DefaultHTTPConfig()
	config.CORSOrigins = []string{"https://example.com", "https://test.com"}

	portal := NewHTTPPortal(config)
	server := httptest.NewServer(portal.HandleHTTP().(http.Handler))
	defer server.Close()

	// Test CORS preflight
	req, _ := http.NewRequest("OPTIONS", server.URL+"/functions/test", nil)
	req.Header.Set("Origin", "https://example.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("Expected CORS origin https://example.com, got %s", origin)
	}

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	if methods != "GET, POST, OPTIONS" {
		t.Errorf("Expected CORS methods 'GET, POST, OPTIONS', got %s", methods)
	}
}

func TestHTTPPortal_Middleware(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())

	// Create a simple middleware
	middleware := &TestMiddleware{called: false}
	portal.SetMiddleware([]any{middleware})

	server := httptest.NewServer(portal.HandleHTTP().(http.Handler))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL + "/functions/test")
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if !middleware.called {
		t.Error("Expected middleware to be called")
	}
}

func TestHTTPPortal_StartStop(t *testing.T) {
	config := DefaultHTTPConfig()
	config.Port = 0 // Use random port

	portal := NewHTTPPortal(config)
	ctx := context.Background()

	// Test health before starting
	err := portal.Health(ctx)
	if err == nil {
		t.Error("Expected health check to fail before starting")
	}

	// Start the portal
	err = portal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start portal: %v", err)
	}

	// Test health after starting
	err = portal.Health(ctx)
	if err != nil {
		t.Errorf("Expected health check to pass after starting: %v", err)
	}

	// Test double start
	err = portal.Start(ctx)
	if err == nil {
		t.Error("Expected error for double start")
	}

	// Stop the portal
	err = portal.Stop(ctx)
	if err != nil {
		t.Errorf("Failed to stop portal: %v", err)
	}

	// Test double stop
	err = portal.Stop(ctx)
	if err != nil {
		t.Error("Expected no error for double stop")
	}
}

func TestHTTPPortal_AddressGeneration(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())

	// Test basic address generation
	address := portal.GenerateAddress("testFunc", nil)
	if address.Scheme() != "http" {
		t.Errorf("Expected scheme http, got %s", address.Scheme())
	}
	if address.Authority() != "localhost:8080" {
		t.Errorf("Expected authority localhost:8080, got %s", address.Authority())
	}
	if address.Path() != "/functions/testFunc" {
		t.Errorf("Expected path /functions/testFunc, got %s", address.Path())
	}

	// Test address generation with metadata
	metadata := map[string]any{"path": "/custom/path"}
	address = portal.GenerateAddress("testFunc", metadata)
	if address.Path() != "/custom/path" {
		t.Errorf("Expected path /custom/path, got %s", address.Path())
	}
}

func TestHTTPPortal_ExtractFunctionName(t *testing.T) {
	portal := NewHTTPPortal(DefaultHTTPConfig())

	tests := []struct {
		path     string
		expected string
	}{
		{"/functions/myFunc", "myFunc"},
		{"/services/MyService/myMethod", "MyService.myMethod"},
		{"/other/path", "/other/path"},
		{"/functions/", ""},
		{"/services/", ""},
	}

	for _, test := range tests {
		result := portal.extractFunctionName(test.path)
		if result != test.expected {
			t.Errorf("extractFunctionName(%s) = %s, expected %s", test.path, result, test.expected)
		}
	}
}

// Test helper types

type HTTPTestFunction struct {
	name    string
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (f *HTTPTestFunction) Name() string {
	return f.name
}

func (f *HTTPTestFunction) Schema() core.FunctionSchema {
	return f.schema
}

func (f *HTTPTestFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler(ctx, params)
}

type TestService struct {
	name         string
	schema       core.ServiceSchema
	greetHandler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (s *TestService) Name() string {
	return s.name
}

func (s *TestService) Schema() core.ServiceSchema {
	return s.schema
}

func (s *TestService) HasMethod(methodName string) bool {
	return methodName == "greet"
}

func (s *TestService) CallMethod(ctx context.Context, methodName string, params api.FunctionData) (api.FunctionData, error) {
	return s.greetHandler(ctx, params)
}

func (s *TestService) GetFunction(name string) (api.Function, bool) {
	if name == "greet" {
		return &HTTPTestFunction{
			name:    "greet",
			schema:  builders.NewFunctionSchema().Name("greet").Build(),
			handler: s.greetHandler,
		}, true
	}
	return nil, false
}

func (s *TestService) MethodNames() []string {
	return []string{"greet"}
}

func (s *TestService) Description() string {
	return "Test service"
}

func (s *TestService) Methods() []string {
	return []string{"greet"}
}

func (s *TestService) GetMethod(name string) (api.Function, bool) {
	return s.GetFunction(name)
}

func (s *TestService) Start(ctx context.Context) error {
	return nil
}

func (s *TestService) Stop(ctx context.Context) error {
	return nil
}

func (s *TestService) Status(ctx context.Context) (api.ServiceStatus, error) {
	return api.ServiceStatus{
		State: api.ServiceStateRunning,
	}, nil
}

func (s *TestService) IsRunning() bool {
	return true
}

type TestMiddleware struct {
	called bool
}

func (m *TestMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.called = true
		next.ServeHTTP(w, r)
	})
}

func TestDefaultHTTPConfig(t *testing.T) {
	config := DefaultHTTPConfig()

	if config.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", config.Host)
	}

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	if config.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", config.ReadTimeout)
	}

	if config.ClientTimeout != 30*time.Second {
		t.Errorf("Expected client timeout 30s, got %v", config.ClientTimeout)
	}

	if len(config.CORSOrigins) != 1 || config.CORSOrigins[0] != "*" {
		t.Errorf("Expected CORS origins [*], got %v", config.CORSOrigins)
	}
}
