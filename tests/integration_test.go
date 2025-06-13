package tests

import (
	"context"
	builders2 "defs.dev/schema/builders"
	"defs.dev/schema/consumers/validation"
	"fmt"
	"testing"
	"time"

	"defs.dev/schema/api"
	"defs.dev/schema/core"
	"defs.dev/schema/portal"
	"defs.dev/schema/registry"
)

// Integration Test 1: End-to-End Service with Complex Generics
func TestIntegration_ComplexServiceWithGenerics(t *testing.T) {
	// Create a complex service schema with generics
	userSchema := builders2.NewObjectSchema().
		Property("id", builders2.NewIntegerSchema().Build()).
		Property("name", builders2.NewStringSchema().Build()).
		Property("email", builders2.NewStringSchema().Build()).
		Property("tags", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
		Required("id", "name", "email").
		Build()

	// Create a generic Result[User, Error] schema
	resultSchema := builders2.NewObjectSchema().
		Property("success", builders2.NewBooleanSchema().Build()).
		Property("data", userSchema).
		Property("error", builders2.NewStringSchema().Build()).
		Required("success").
		Build()

	// Create a service with multiple methods using complex schemas
	userService := builders2.NewServiceSchema().
		Name("UserService").
		Description("A comprehensive user management service").
		Tag("users").
		Tag("management").
		Example(map[string]any{
			"name":        "UserService",
			"description": "Manages user operations",
		}).
		Method("createUser", builders2.NewFunctionSchema().
			Name("createUser").
			Description("Creates a new user").
			Input("userData", userSchema).
			RequiredInputs("userData").
			Output("result", resultSchema).
			Example(map[string]any{
				"input": map[string]any{
					"userData": map[string]any{
						"id":    1,
						"name":  "John Doe",
						"email": "john@example.com",
						"tags":  []string{"admin", "active"},
					},
				},
				"output": map[string]any{
					"result": map[string]any{
						"success": true,
						"data": map[string]any{
							"id":    1,
							"name":  "John Doe",
							"email": "john@example.com",
							"tags":  []string{"admin", "active"},
						},
					},
				},
			}).
			Build()).
		Method("getUserById", builders2.NewFunctionSchema().
			Name("getUserById").
			Description("Retrieves a user by ID").
			Input("id", builders2.NewIntegerSchema().Build()).
			RequiredInputs("id").
			Output("result", resultSchema).
			Build()).
		Method("updateUserTags", builders2.NewFunctionSchema().
			Name("updateUserTags").
			Description("Updates user tags with array operations").
			Input("userId", builders2.NewIntegerSchema().Build()).
			Input("tagsToAdd", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
			Input("tagsToRemove", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
			RequiredInputs("userId").
			Output("result", resultSchema).
			Build()).
		Build()

	// Test service schema validation
	if userService.Name() != "UserService" {
		t.Errorf("Expected service name 'UserService', got %s", userService.Name())
	}

	methods := userService.Methods()
	if len(methods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(methods))
	}

	// Test method introspection
	createUserMethod := methods[0]
	if createUserMethod.Name() != "createUser" {
		t.Errorf("Expected method name 'createUser', got %s", createUserMethod.Name())
	}

	// Test complex input validation
	validUserData := map[string]any{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
		"tags":  []string{"admin", "active"},
	}

	// Use the function schema directly for validation
	functionSchema := createUserMethod.Function()
	result := validation.ValidateValue(functionSchema, map[string]any{"userData": validUserData})
	if !result.Valid {
		t.Errorf("Expected valid user data, got errors: %v", result.Errors)
	}

	// Test invalid data
	invalidUserData := map[string]any{
		"id":   "not-a-number",
		"name": "",
		// missing required email
		"tags": "not-an-array",
	}

	result = validation.ValidateValue(functionSchema, map[string]any{"userData": invalidUserData})
	if result.Valid {
		t.Error("Expected validation to fail for invalid user data")
	}

	// Should have multiple validation errors
	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 validation errors, got %d", len(result.Errors))
	}

	t.Logf("Complex service with generics test passed with %d methods", len(methods))
}

// Integration Test 2: Multi-Portal Service Registration and Resolution
func TestIntegration_MultiPortalServiceRegistration(t *testing.T) {
	ctx := context.Background()

	// Create portal registry
	registry := portal.NewPortalRegistry()

	// Create different types of portals
	localPortal := portal.NewLocalPortal()
	testingPortal := portal.NewTestingPortal()
	httpPortal := portal.NewHTTPPortal(nil)

	// Register portals
	err := registry.RegisterPortal([]string{"local"}, localPortal)
	if err != nil {
		t.Fatalf("Failed to register local portal: %v", err)
	}

	err = registry.RegisterPortal([]string{"test", "mock"}, testingPortal)
	if err != nil {
		t.Fatalf("Failed to register testing portal: %v", err)
	}

	err = registry.RegisterPortal([]string{"http"}, httpPortal)
	if err != nil {
		t.Fatalf("Failed to register HTTP portal: %v", err)
	}

	// Create a complex calculation service
	calculatorService := &TestCalculatorService{
		name:      "CalculatorService",
		isRunning: false,
		status: api.ServiceStatus{
			State:   api.ServiceStateStopped,
			Healthy: false,
		},
		schema: builders2.NewServiceSchema().
			Name("CalculatorService").
			Description("Advanced mathematical operations service").
			Method("add", builders2.NewFunctionSchema().
				Name("add").
				Description("Adds two numbers").
				Input("a", builders2.NewNumberSchema().Build()).
				Input("b", builders2.NewNumberSchema().Build()).
				RequiredInputs("a", "b").
				Output("result", builders2.NewNumberSchema().Build()).
				Build()).
			Method("multiply", builders2.NewFunctionSchema().
				Name("multiply").
				Description("Multiplies two numbers").
				Input("a", builders2.NewNumberSchema().Build()).
				Input("b", builders2.NewNumberSchema().Build()).
				RequiredInputs("a", "b").
				Output("result", builders2.NewNumberSchema().Build()).
				Build()).
			Method("factorial", builders2.NewFunctionSchema().
				Name("factorial").
				Description("Calculates factorial of a number").
				Input("n", builders2.NewIntegerSchema().Build()).
				RequiredInputs("n").
				Output("result", builders2.NewIntegerSchema().Build()).
				Build()).
			Build(),
	}

	// Register service with different portals
	localAddr, err := localPortal.ApplyService(ctx, calculatorService)
	if err != nil {
		t.Fatalf("Failed to register service with local portal: %v", err)
	}

	testAddr, err := testingPortal.ApplyService(ctx, calculatorService)
	if err != nil {
		t.Fatalf("Failed to register service with testing portal: %v", err)
	}

	httpAddr, err := httpPortal.ApplyService(ctx, calculatorService)
	if err != nil {
		t.Fatalf("Failed to register service with HTTP portal: %v", err)
	}

	// Test address schemes
	if localAddr.Scheme() != "local" {
		t.Errorf("Expected local scheme, got %s", localAddr.Scheme())
	}

	if testAddr.Scheme() != "test" {
		t.Errorf("Expected test scheme, got %s", testAddr.Scheme())
	}

	if httpAddr.Scheme() != "http" {
		t.Errorf("Expected http scheme, got %s", httpAddr.Scheme())
	}

	// Test portal resolution through registry
	resolvedPortal, err := registry.GetPortal(localAddr)
	if err != nil {
		t.Fatalf("Failed to resolve local portal: %v", err)
	}

	if resolvedPortal != localPortal {
		t.Error("Expected to get the same local portal instance")
	}

	// Test service resolution
	resolvedService, err := resolvedPortal.ResolveService(ctx, localAddr)
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	if resolvedService.Schema().Name() != "CalculatorService" {
		t.Errorf("Expected service name 'CalculatorService', got %s", resolvedService.Schema().Name())
	}

	// Test health checks across all portals
	healthResults := registry.(*portal.PortalRegistryImpl).Health(ctx)
	for scheme, healthErr := range healthResults {
		if healthErr != nil {
			// HTTP portal is expected to be unhealthy since we're not starting a server
			if scheme == "http" {
				t.Logf("Portal %s is unhealthy as expected: %v", scheme, healthErr)
			} else {
				t.Errorf("Portal %s is unhealthy: %v", scheme, healthErr)
			}
		}
	}

	t.Logf("Multi-portal registration test passed with %d portals", len(registry.ListPortals()))
}

// Integration Test 3: Complex Function Composition with Validation
func TestIntegration_ComplexFunctionComposition(t *testing.T) {
	ctx := context.Background()

	// Create a data processing pipeline with multiple functions

	// Step 1: Data validation function
	validateDataFunc := &TestFunction{
		name: "validateData",
		schema: builders2.NewFunctionSchema().
			Name("validateData").
			Description("Validates input data structure").
			Input("data", builders2.NewObjectSchema().
				Property("items", builders2.NewArraySchema().
					Items(builders2.NewObjectSchema().
						Property("id", builders2.NewIntegerSchema().Build()).
						Property("value", builders2.NewNumberSchema().Build()).
						Property("metadata", builders2.NewObjectSchema().
							Property("category", builders2.NewStringSchema().Build()).
							Property("tags", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
							Build()).
						Required("id", "value").
						Build()).
					Build()).
				Property("config", builders2.NewObjectSchema().
					Property("threshold", builders2.NewNumberSchema().Build()).
					Property("includeMetadata", builders2.NewBooleanSchema().Build()).
					Build()).
				Required("items").
				Build()).
			RequiredInputs("data").
			Output("isValid", builders2.NewBooleanSchema().Build()).
			Output("errors", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			data, _ := params.Get("data")
			dataMap := data.(map[string]any)

			errors := []string{}
			items, hasItems := dataMap["items"]
			if !hasItems {
				errors = append(errors, "missing items")
			} else {
				itemsArray := items.([]any)
				for i, item := range itemsArray {
					itemMap := item.(map[string]any)
					if _, hasID := itemMap["id"]; !hasID {
						errors = append(errors, fmt.Sprintf("item[%d] missing id", i))
					}
					if _, hasValue := itemMap["value"]; !hasValue {
						errors = append(errors, fmt.Sprintf("item[%d] missing value", i))
					}
				}
			}

			return api.NewFunctionData(map[string]any{
				"isValid": len(errors) == 0,
				"errors":  errors,
			}), nil
		},
	}

	// Step 2: Data transformation function
	transformDataFunc := &TestFunction{
		name: "transformData",
		schema: builders2.NewFunctionSchema().
			Name("transformData").
			Description("Transforms and enriches data").
			Input("data", builders2.NewObjectSchema().
				Property("items", builders2.NewArraySchema().Build()).
				Property("config", builders2.NewObjectSchema().Build()).
				Required("items").
				Build()).
			RequiredInputs("data").
			Output("transformedData", builders2.NewObjectSchema().
				Property("processedItems", builders2.NewArraySchema().Build()).
				Property("summary", builders2.NewObjectSchema().
					Property("totalItems", builders2.NewIntegerSchema().Build()).
					Property("averageValue", builders2.NewNumberSchema().Build()).
					Build()).
				Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			data, _ := params.Get("data")
			dataMap := data.(map[string]any)
			items := dataMap["items"].([]any)

			processedItems := make([]any, len(items))
			totalValue := 0.0

			for i, item := range items {
				itemMap := item.(map[string]any)
				value := itemMap["value"].(float64)
				totalValue += value

				processedItems[i] = map[string]any{
					"id":             itemMap["id"],
					"originalValue":  value,
					"processedValue": value * 1.1, // 10% increase
					"timestamp":      time.Now().Unix(),
				}
			}

			return api.NewFunctionData(map[string]any{
				"transformedData": map[string]any{
					"processedItems": processedItems,
					"summary": map[string]any{
						"totalItems":   len(items),
						"averageValue": totalValue / float64(len(items)),
					},
				},
			}), nil
		},
	}

	// Step 3: Aggregation function
	aggregateFunc := &TestFunction{
		name: "aggregateResults",
		schema: builders2.NewFunctionSchema().
			Name("aggregateResults").
			Description("Aggregates processed data").
			Input("transformedData", builders2.NewObjectSchema().Build()).
			RequiredInputs("transformedData").
			Output("aggregation", builders2.NewObjectSchema().
				Property("totalProcessedValue", builders2.NewNumberSchema().Build()).
				Property("itemCount", builders2.NewIntegerSchema().Build()).
				Property("processingTimestamp", builders2.NewIntegerSchema().Build()).
				Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			transformedData, _ := params.Get("transformedData")
			dataMap := transformedData.(map[string]any)
			processedItems := dataMap["processedItems"].([]any)

			totalProcessedValue := 0.0
			for _, item := range processedItems {
				itemMap := item.(map[string]any)
				totalProcessedValue += itemMap["processedValue"].(float64)
			}

			return api.NewFunctionData(map[string]any{
				"aggregation": map[string]any{
					"totalProcessedValue": totalProcessedValue,
					"itemCount":           len(processedItems),
					"processingTimestamp": time.Now().Unix(),
				},
			}), nil
		},
	}

	// Register functions with local portal
	localPortal := portal.NewLocalPortal()

	validateAddr, err := localPortal.Apply(ctx, validateDataFunc)
	if err != nil {
		t.Fatalf("Failed to register validate function: %v", err)
	}

	transformAddr, err := localPortal.Apply(ctx, transformDataFunc)
	if err != nil {
		t.Fatalf("Failed to register transform function: %v", err)
	}

	aggregateAddr, err := localPortal.Apply(ctx, aggregateFunc)
	if err != nil {
		t.Fatalf("Failed to register aggregate function: %v", err)
	}

	// Test the complete pipeline
	testData := map[string]any{
		"items": []any{
			map[string]any{
				"id":    1,
				"value": 10.5,
				"metadata": map[string]any{
					"category": "A",
					"tags":     []string{"important", "processed"},
				},
			},
			map[string]any{
				"id":    2,
				"value": 20.3,
				"metadata": map[string]any{
					"category": "B",
					"tags":     []string{"standard"},
				},
			},
			map[string]any{
				"id":    3,
				"value": 15.7,
			},
		},
		"config": map[string]any{
			"threshold":       10.0,
			"includeMetadata": true,
		},
	}

	// Step 1: Validate
	validateFunc, err := localPortal.ResolveFunction(ctx, validateAddr)
	if err != nil {
		t.Fatalf("Failed to resolve validate function: %v", err)
	}

	validateResult, err := validateFunc.Call(ctx, api.NewFunctionData(map[string]any{"data": testData}))
	if err != nil {
		t.Fatalf("Failed to call validate function: %v", err)
	}

	isValid, _ := validateResult.Get("isValid")
	if !isValid.(bool) {
		errors, _ := validateResult.Get("errors")
		t.Fatalf("Data validation failed: %v", errors)
	}

	// Step 2: Transform
	transformFunc, err := localPortal.ResolveFunction(ctx, transformAddr)
	if err != nil {
		t.Fatalf("Failed to resolve transform function: %v", err)
	}

	transformResult, err := transformFunc.Call(ctx, api.NewFunctionData(map[string]any{"data": testData}))
	if err != nil {
		t.Fatalf("Failed to call transform function: %v", err)
	}

	transformedData, _ := transformResult.Get("transformedData")
	transformedMap := transformedData.(map[string]any)
	summary := transformedMap["summary"].(map[string]any)

	if summary["totalItems"].(int) != 3 {
		t.Errorf("Expected 3 total items, got %v", summary["totalItems"])
	}

	// Step 3: Aggregate
	aggregateFuncResolved, err := localPortal.ResolveFunction(ctx, aggregateAddr)
	if err != nil {
		t.Fatalf("Failed to resolve aggregate function: %v", err)
	}

	aggregateResult, err := aggregateFuncResolved.Call(ctx, api.NewFunctionData(map[string]any{"transformedData": transformedData}))
	if err != nil {
		t.Fatalf("Failed to call aggregate function: %v", err)
	}

	aggregation, _ := aggregateResult.Get("aggregation")
	aggregationMap := aggregation.(map[string]any)

	if aggregationMap["itemCount"].(int) != 3 {
		t.Errorf("Expected 3 items in aggregation, got %v", aggregationMap["itemCount"])
	}

	// Verify the processing increased values by 10%
	expectedTotal := (10.5 + 20.3 + 15.7) * 1.1
	actualTotal := aggregationMap["totalProcessedValue"].(float64)
	if actualTotal < expectedTotal-0.1 || actualTotal > expectedTotal+0.1 {
		t.Errorf("Expected total processed value around %.2f, got %.2f", expectedTotal, actualTotal)
	}

	t.Logf("Complex function composition test passed with pipeline processing %d items", 3)
}

// Integration Test 4: Service Registry with Function Registry Integration
func TestIntegration_ServiceAndFunctionRegistryIntegration(t *testing.T) {
	// Create registries
	funcRegistry := registry.NewFunctionRegistry()
	serviceRegistry := registry.NewServiceRegistry()

	// Create a comprehensive analytics service
	analyticsService := builders2.NewServiceSchema().
		Name("AnalyticsService").
		Description("Advanced analytics and reporting service").
		Tag("analytics").
		Tag("reporting").
		Method("calculateMetrics", builders2.NewFunctionSchema().
			Name("calculateMetrics").
			Description("Calculates various metrics from data").
			Input("dataset", builders2.NewArraySchema().
				Items(builders2.NewObjectSchema().
					Property("timestamp", builders2.NewIntegerSchema().Build()).
					Property("value", builders2.NewNumberSchema().Build()).
					Property("category", builders2.NewStringSchema().Build()).
					Required("timestamp", "value").
					Build()).
				Build()).
			Input("metricTypes", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
			RequiredInputs("dataset", "metricTypes").
			Output("metrics", builders2.NewObjectSchema().
				Property("mean", builders2.NewNumberSchema().Build()).
				Property("median", builders2.NewNumberSchema().Build()).
				Property("standardDeviation", builders2.NewNumberSchema().Build()).
				Property("categoryBreakdown", builders2.NewObjectSchema().Build()).
				Build()).
			Build()).
		Method("generateReport", builders2.NewFunctionSchema().
			Name("generateReport").
			Description("Generates a formatted report").
			Input("metrics", builders2.NewObjectSchema().Build()).
			Input("format", builders2.NewStringSchema().Build()).
			RequiredInputs("metrics", "format").
			Output("report", builders2.NewStringSchema().Build()).
			Output("metadata", builders2.NewObjectSchema().
				Property("generatedAt", builders2.NewIntegerSchema().Build()).
				Property("format", builders2.NewStringSchema().Build()).
				Property("size", builders2.NewIntegerSchema().Build()).
				Build()).
			Build()).
		Method("exportData", builders2.NewFunctionSchema().
			Name("exportData").
			Description("Exports data in various formats").
			Input("data", builders2.NewObjectSchema().Build()).
			Input("exportFormat", builders2.NewStringSchema().Build()).
			Input("options", builders2.NewObjectSchema().
				Property("includeHeaders", builders2.NewBooleanSchema().Build()).
				Property("compression", builders2.NewStringSchema().Build()).
				Build()).
			RequiredInputs("data", "exportFormat").
			Output("exportedData", builders2.NewStringSchema().Build()).
			Output("exportInfo", builders2.NewObjectSchema().
				Property("format", builders2.NewStringSchema().Build()).
				Property("size", builders2.NewIntegerSchema().Build()).
				Property("checksum", builders2.NewStringSchema().Build()).
				Build()).
			Build()).
		Build()

	// Register the service
	err := serviceRegistry.RegisterService("AnalyticsService", analyticsService)
	if err != nil {
		t.Fatalf("Failed to register analytics service: %v", err)
	}

	// Set service metadata separately
	err = serviceRegistry.SetServiceMetadata("AnalyticsService", api.ServiceMetadata{
		Version:     "1.0.0",
		Tags:        []string{"analytics", "reporting", "data"},
		Description: "Comprehensive analytics service for data processing",
		Owner:       "data-team",
	})
	if err != nil {
		t.Fatalf("Failed to set service metadata: %v", err)
	}

	// Create standalone utility functions
	utilityFunctions := []*TestFunction{
		{
			name: "validateDataset",
			schema: builders2.NewFunctionSchema().
				Name("validateDataset").
				Description("Validates dataset structure and content").
				Input("dataset", builders2.NewArraySchema().Build()).
				RequiredInputs("dataset").
				Output("isValid", builders2.NewBooleanSchema().Build()).
				Output("validationErrors", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				dataset, _ := params.Get("dataset")
				dataArray := dataset.([]any)
				errors := []string{}

				for i, item := range dataArray {
					itemMap := item.(map[string]any)
					if _, hasTimestamp := itemMap["timestamp"]; !hasTimestamp {
						errors = append(errors, fmt.Sprintf("item[%d] missing timestamp", i))
					}
					if _, hasValue := itemMap["value"]; !hasValue {
						errors = append(errors, fmt.Sprintf("item[%d] missing value", i))
					}
				}

				return api.NewFunctionData(map[string]any{
					"isValid":          len(errors) == 0,
					"validationErrors": errors,
				}), nil
			},
		},
		{
			name: "formatTimestamp",
			schema: builders2.NewFunctionSchema().
				Name("formatTimestamp").
				Description("Formats Unix timestamp to human-readable format").
				Input("timestamp", builders2.NewIntegerSchema().Build()).
				Input("format", builders2.NewStringSchema().Build()).
				RequiredInputs("timestamp").
				Output("formatted", builders2.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				timestamp, _ := params.Get("timestamp")
				format, hasFormat := params.Get("format")

				ts := time.Unix(int64(timestamp.(int)), 0)
				var formatted string

				if hasFormat && format.(string) == "iso" {
					formatted = ts.Format(time.RFC3339)
				} else {
					formatted = ts.Format("2006-01-02 15:04:05")
				}

				return api.NewFunctionData(map[string]any{
					"formatted": formatted,
				}), nil
			},
		},
		{
			name: "calculateHash",
			schema: builders2.NewFunctionSchema().
				Name("calculateHash").
				Description("Calculates hash of input data").
				Input("data", builders2.NewStringSchema().Build()).
				Input("algorithm", builders2.NewStringSchema().Build()).
				RequiredInputs("data").
				Output("hash", builders2.NewStringSchema().Build()).
				Output("algorithm", builders2.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				data, _ := params.Get("data")
				algorithm, hasAlgorithm := params.Get("algorithm")

				alg := "sha256"
				if hasAlgorithm {
					alg = algorithm.(string)
				}

				// Simple hash simulation
				hash := fmt.Sprintf("%s-hash-of-%s", alg, data.(string)[:min(10, len(data.(string)))])

				return api.NewFunctionData(map[string]any{
					"hash":      hash,
					"algorithm": alg,
				}), nil
			},
		},
	}

	// Register utility functions
	for _, fn := range utilityFunctions {
		err := funcRegistry.Register(fn.name, fn)
		if err != nil {
			t.Fatalf("Failed to register function %s: %v", fn.name, err)
		}
	}

	// Test service introspection
	services := serviceRegistry.ListServices()
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if services[0] != "AnalyticsService" {
		t.Errorf("Expected service 'AnalyticsService', got %s", services[0])
	}

	// Test service method introspection
	methods := serviceRegistry.ListServiceMethods("AnalyticsService")
	expectedMethods := []string{"calculateMetrics", "generateReport", "exportData"}
	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(methods))
	}

	for _, expectedMethod := range expectedMethods {
		found := false
		for _, method := range methods {
			if method == expectedMethod {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected method %s not found", expectedMethod)
		}
	}

	// Test function registry integration
	functions := funcRegistry.List()
	if len(functions) != 3 {
		t.Errorf("Expected 3 utility functions, got %d", len(functions))
	}

	// Test cross-registry functionality
	// Get a service method and validate its schema
	calculateMetricsMethod, exists := serviceRegistry.GetServiceMethod("AnalyticsService", "calculateMetrics")
	if !exists {
		t.Fatal("Expected to find calculateMetrics method")
	}

	// Test method schema validation
	testDataset := []any{
		map[string]any{
			"timestamp": 1640995200,
			"value":     100.5,
			"category":  "sales",
		},
		map[string]any{
			"timestamp": 1640995260,
			"value":     150.3,
			"category":  "marketing",
		},
	}

	methodInput := map[string]any{
		"dataset":     testDataset,
		"metricTypes": []string{"mean", "median"},
	}

	functionSchema := calculateMetricsMethod.Schema()
	validationResult := validation.ValidateValue(functionSchema, methodInput)
	if !validationResult.Valid {
		t.Errorf("Expected valid method input, got errors: %v", validationResult.Errors)
	}

	// Test utility function validation
	validateFunc, exists := funcRegistry.Get("validateDataset")
	if !exists {
		t.Fatal("Expected to find validateDataset function")
	}

	utilityInput := map[string]any{"dataset": testDataset}
	utilityFunctionSchema := validateFunc.Schema()
	utilityValidationResult := validation.ValidateValue(utilityFunctionSchema, utilityInput)
	if !utilityValidationResult.Valid {
		t.Errorf("Expected valid utility input, got errors: %v", utilityValidationResult.Errors)
	}

	// Test registry statistics
	serviceStats := serviceRegistry.Count()
	functionStats := funcRegistry.Count()

	if serviceStats != 1 {
		t.Errorf("Expected 1 service in registry, got %d", serviceStats)
	}

	if functionStats != 3 {
		t.Errorf("Expected 3 functions in registry, got %d", functionStats)
	}

	// Test method count across service
	methodCount := serviceRegistry.MethodCount()
	if methodCount != 3 {
		t.Errorf("Expected 3 methods across all services, got %d", methodCount)
	}

	t.Logf("Service and function registry integration test passed with %d services and %d functions", serviceStats, functionStats)
}

// Integration Test 5: Advanced Generic Schema Composition
func TestIntegration_AdvancedGenericSchemaComposition(t *testing.T) {
	// Create complex nested generic schemas

	// Define a generic Result[T, E] pattern
	createResultSchema := func(successSchema, errorSchema core.Schema) core.Schema {
		return builders2.NewObjectSchema().
			Property("success", builders2.NewBooleanSchema().Build()).
			Property("data", successSchema).
			Property("error", errorSchema).
			Property("timestamp", builders2.NewIntegerSchema().Build()).
			Required("success", "timestamp").
			Build()
	}

	// Define a generic List[T] pattern
	createListSchema := func(itemSchema core.Schema) core.Schema {
		return builders2.NewObjectSchema().
			Property("items", builders2.NewArraySchema().Items(itemSchema).Build()).
			Property("totalCount", builders2.NewIntegerSchema().Build()).
			Property("hasMore", builders2.NewBooleanSchema().Build()).
			Required("items", "totalCount", "hasMore").
			Build()
	}

	// Define a generic Map[K, V] pattern
	createMapSchema := func(keySchema, valueSchema core.Schema) core.Schema {
		return builders2.NewObjectSchema().
			Property("entries", builders2.NewArraySchema().
				Items(builders2.NewObjectSchema().
					Property("key", keySchema).
					Property("value", valueSchema).
					Required("key", "value").
					Build()).
				Build()).
			Property("size", builders2.NewIntegerSchema().Build()).
			Required("entries", "size").
			Build()
	}

	// Create base entity schema
	entitySchema := builders2.NewObjectSchema().
		Property("id", builders2.NewStringSchema().Build()).
		Property("type", builders2.NewStringSchema().Build()).
		Property("attributes", builders2.NewObjectSchema().Build()).
		Property("relationships", builders2.NewArraySchema().
			Items(builders2.NewObjectSchema().
				Property("type", builders2.NewStringSchema().Build()).
				Property("id", builders2.NewStringSchema().Build()).
				Required("type", "id").
				Build()).
			Build()).
		Required("id", "type").
		Build()

	// Create error schema
	errorSchema := builders2.NewObjectSchema().
		Property("code", builders2.NewStringSchema().Build()).
		Property("message", builders2.NewStringSchema().Build()).
		Property("details", builders2.NewObjectSchema().Build()).
		Required("code", "message").
		Build()

	// Compose complex schemas using generics

	// Result[List[Entity], Error]
	entityListResultSchema := createResultSchema(
		createListSchema(entitySchema),
		errorSchema,
	)

	// Map[String, Result[Entity, Error]]
	entityMapSchema := createMapSchema(
		builders2.NewStringSchema().Build(),
		createResultSchema(entitySchema, errorSchema),
	)

	// Result[Map[String, List[Entity]], Error]
	complexNestedSchema := createResultSchema(
		createMapSchema(
			builders2.NewStringSchema().Build(),
			createListSchema(entitySchema),
		),
		errorSchema,
	)

	// Test schema validation with complex nested data

	// Test 1: Valid entity list result
	validEntityListResult := map[string]any{
		"success":   true,
		"timestamp": 1640995200,
		"data": map[string]any{
			"items": []any{
				map[string]any{
					"id":   "entity-1",
					"type": "user",
					"attributes": map[string]any{
						"name":  "John Doe",
						"email": "john@example.com",
					},
					"relationships": []any{
						map[string]any{
							"type": "group",
							"id":   "group-1",
						},
					},
				},
				map[string]any{
					"id":   "entity-2",
					"type": "product",
					"attributes": map[string]any{
						"name":  "Widget",
						"price": 29.99,
					},
					"relationships": []any{},
				},
			},
			"totalCount": 2,
			"hasMore":    false,
		},
	}

	result := validation.ValidateValue(entityListResultSchema, validEntityListResult)
	if !result.Valid {
		t.Errorf("Expected valid entity list result, got errors: %v", result.Errors)
	}

	// Test 2: Valid entity map
	validEntityMap := map[string]any{
		"entries": []any{
			map[string]any{
				"key": "user-123",
				"value": map[string]any{
					"success":   true,
					"timestamp": 1640995200,
					"data": map[string]any{
						"id":   "user-123",
						"type": "user",
						"attributes": map[string]any{
							"name": "Alice Smith",
						},
						"relationships": []any{},
					},
				},
			},
			map[string]any{
				"key": "user-456",
				"value": map[string]any{
					"success":   false,
					"timestamp": 1640995200,
					"error": map[string]any{
						"code":    "NOT_FOUND",
						"message": "User not found",
						"details": map[string]any{
							"userId": "user-456",
						},
					},
				},
			},
		},
		"size": 2,
	}

	result = validation.ValidateValue(entityMapSchema, validEntityMap)
	if !result.Valid {
		t.Errorf("Expected valid entity map, got errors: %v", result.Errors)
	}

	// Test 3: Complex nested schema
	validComplexNested := map[string]any{
		"success":   true,
		"timestamp": 1640995200,
		"data": map[string]any{
			"entries": []any{
				map[string]any{
					"key": "users",
					"value": map[string]any{
						"items": []any{
							map[string]any{
								"id":            "user-1",
								"type":          "user",
								"attributes":    map[string]any{"name": "John"},
								"relationships": []any{},
							},
						},
						"totalCount": 1,
						"hasMore":    false,
					},
				},
				map[string]any{
					"key": "products",
					"value": map[string]any{
						"items": []any{
							map[string]any{
								"id":            "product-1",
								"type":          "product",
								"attributes":    map[string]any{"name": "Widget"},
								"relationships": []any{},
							},
							map[string]any{
								"id":            "product-2",
								"type":          "product",
								"attributes":    map[string]any{"name": "Gadget"},
								"relationships": []any{},
							},
						},
						"totalCount": 2,
						"hasMore":    true,
					},
				},
			},
			"size": 2,
		},
	}

	result = validation.ValidateValue(complexNestedSchema, validComplexNested)
	if !result.Valid {
		t.Errorf("Expected valid complex nested schema, got errors: %v", result.Errors)
	}

	// Test 4: Invalid data with detailed error paths
	invalidComplexNested := map[string]any{
		"success":   true,
		"timestamp": "not-a-number", // Invalid timestamp
		"data": map[string]any{
			"entries": []any{
				map[string]any{
					"key": "users",
					"value": map[string]any{
						"items": []any{
							map[string]any{
								"id":   "user-1",
								"type": "user",
								// Missing required attributes
								"relationships": "not-an-array", // Invalid type
							},
						},
						"totalCount": "not-a-number", // Invalid type
						"hasMore":    false,
					},
				},
			},
			// Missing required size property
		},
	}

	result = validation.ValidateValue(complexNestedSchema, invalidComplexNested)
	if result.Valid {
		t.Error("Expected validation to fail for invalid complex nested data")
	}

	// Should have multiple validation errors with proper paths
	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 validation errors, got %d", len(result.Errors))
	}

	// Check that error paths are properly nested
	hasNestedPath := false
	for _, err := range result.Errors {
		if len(err.Path) > 10 { // Look for deeply nested paths
			hasNestedPath = true
			break
		}
	}
	if !hasNestedPath {
		t.Error("Expected to find deeply nested validation error paths")
	}

	// Test JSON Schema generation for complex schemas
	jsonSchema := toJSONSchema(complexNestedSchema)
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected JSON schema type 'object', got %v", jsonSchema["type"])
	}

	properties, hasProperties := jsonSchema["properties"]
	if !hasProperties {
		t.Error("Expected JSON schema to have properties")
	}

	propertiesMap := properties.(map[string]any)
	if _, hasData := propertiesMap["data"]; !hasData {
		t.Error("Expected JSON schema to have 'data' property")
	}

	t.Logf("Advanced generic schema composition test passed with complex nested validation")
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Test helper types

type TestCalculatorService struct {
	name      string
	schema    core.ServiceSchema
	isRunning bool
	status    api.ServiceStatus
}

// Implement the enhanced Service interface

func (s *TestCalculatorService) CallMethod(ctx context.Context, methodName string, params api.FunctionData) (api.FunctionData, error) {
	if !s.isRunning {
		return nil, fmt.Errorf("service %s is not running", s.name)
	}

	function, exists := s.GetFunction(methodName)
	if !exists {
		return nil, fmt.Errorf("method %s not found on service %s", methodName, s.name)
	}

	return function.Call(ctx, params)
}

func (s *TestCalculatorService) Schema() core.ServiceSchema {
	return s.schema
}

func (s *TestCalculatorService) Name() string {
	return s.name
}

func (s *TestCalculatorService) Start(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("service %s is already running", s.name)
	}
	s.isRunning = true
	now := time.Now()
	s.status = api.ServiceStatus{
		State:     api.ServiceStateRunning,
		StartedAt: &now,
		Healthy:   true,
	}
	return nil
}

func (s *TestCalculatorService) Stop(ctx context.Context) error {
	if !s.isRunning {
		return fmt.Errorf("service %s is not running", s.name)
	}
	s.isRunning = false
	now := time.Now()
	s.status = api.ServiceStatus{
		State:     api.ServiceStateStopped,
		StoppedAt: &now,
		Healthy:   false,
	}
	return nil
}

func (s *TestCalculatorService) Status(ctx context.Context) (api.ServiceStatus, error) {
	return s.status, nil
}

func (s *TestCalculatorService) IsRunning() bool {
	return s.isRunning
}

func (s *TestCalculatorService) HasMethod(methodName string) bool {
	_, exists := s.GetFunction(methodName)
	return exists
}

func (s *TestCalculatorService) MethodNames() []string {
	return []string{"add", "multiply", "factorial"}
}

// Legacy method for backward compatibility
func (s *TestCalculatorService) GetFunction(name string) (api.Function, bool) {
	switch name {
	case "add":
		return &TestFunction{
			name: "add",
			schema: builders2.NewFunctionSchema().
				Name("add").
				Input("a", builders2.NewNumberSchema().Build()).
				Input("b", builders2.NewNumberSchema().Build()).
				RequiredInputs("a", "b").
				Output("result", builders2.NewNumberSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				a, _ := params.Get("a")
				b, _ := params.Get("b")
				result := a.(float64) + b.(float64)
				return api.NewFunctionData(map[string]any{"result": result}), nil
			},
		}, true
	case "multiply":
		return &TestFunction{
			name: "multiply",
			schema: builders2.NewFunctionSchema().
				Name("multiply").
				Input("a", builders2.NewNumberSchema().Build()).
				Input("b", builders2.NewNumberSchema().Build()).
				RequiredInputs("a", "b").
				Output("result", builders2.NewNumberSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				a, _ := params.Get("a")
				b, _ := params.Get("b")
				result := a.(float64) * b.(float64)
				return api.NewFunctionData(map[string]any{"result": result}), nil
			},
		}, true
	case "factorial":
		return &TestFunction{
			name: "factorial",
			schema: builders2.NewFunctionSchema().
				Name("factorial").
				Input("n", builders2.NewIntegerSchema().Build()).
				RequiredInputs("n").
				Output("result", builders2.NewIntegerSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				n, _ := params.Get("n")
				num := n.(int)
				result := 1
				for i := 2; i <= num; i++ {
					result *= i
				}
				return api.NewFunctionData(map[string]any{"result": result}), nil
			},
		}, true
	}
	return nil, false
}

type TestFunction struct {
	name    string
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (f *TestFunction) Name() string {
	return f.name
}

func (f *TestFunction) Schema() core.FunctionSchema {
	return f.schema
}

func (f *TestFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler(ctx, params)
}
