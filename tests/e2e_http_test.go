package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"defs.dev/schema/api"
	"defs.dev/schema/api/core"
	"defs.dev/schema/builders"
	"defs.dev/schema/portal"
)

// E2E Test 1: Single HTTP Server with Multiple Clients
func TestE2E_HTTPPortalSingleServerMultipleClients(t *testing.T) {
	ctx := context.Background()

	// Create HTTP portal server
	serverConfig := &portal.HTTPConfig{
		Host:          "localhost",
		Port:          8081,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		ClientTimeout: 5 * time.Second,
	}

	httpPortal := portal.NewHTTPPortal(serverConfig)

	// Create and register multiple functions
	functions := []*E2ETestFunction{
		{
			name: "calculator_add",
			schema: builders.NewFunctionSchema().
				Name("calculator_add").
				Description("Adds two numbers").
				Input("a", builders.NewNumberSchema().Build()).
				Input("b", builders.NewNumberSchema().Build()).
				RequiredInputs("a", "b").
				Output("result", builders.NewNumberSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				a, _ := params.Get("a")
				b, _ := params.Get("b")
				result := a.(float64) + b.(float64)
				return portal.NewFunctionData(map[string]any{"result": result}), nil
			},
		},
		{
			name: "string_processor",
			schema: builders.NewFunctionSchema().
				Name("string_processor").
				Description("Processes strings with various operations").
				Input("text", builders.NewStringSchema().Build()).
				Input("operation", builders.NewStringSchema().Build()).
				RequiredInputs("text", "operation").
				Output("processed", builders.NewStringSchema().Build()).
				Output("length", builders.NewIntegerSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				text, _ := params.Get("text")
				operation, _ := params.Get("operation")

				textStr := text.(string)
				var processed string

				switch operation.(string) {
				case "uppercase":
					processed = fmt.Sprintf("UPPER: %s", textStr)
				case "lowercase":
					processed = fmt.Sprintf("lower: %s", textStr)
				case "reverse":
					runes := []rune(textStr)
					for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
						runes[i], runes[j] = runes[j], runes[i]
					}
					processed = string(runes)
				default:
					processed = fmt.Sprintf("PROCESSED: %s", textStr)
				}

				return portal.NewFunctionData(map[string]any{
					"processed": processed,
					"length":    len(textStr),
				}), nil
			},
		},
		{
			name: "data_aggregator",
			schema: builders.NewFunctionSchema().
				Name("data_aggregator").
				Description("Aggregates array data").
				Input("numbers", builders.NewArraySchema().Items(builders.NewNumberSchema().Build()).Build()).
				RequiredInputs("numbers").
				Output("sum", builders.NewNumberSchema().Build()).
				Output("average", builders.NewNumberSchema().Build()).
				Output("count", builders.NewIntegerSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				numbers, _ := params.Get("numbers")
				numbersArray := numbers.([]any)

				sum := 0.0
				count := len(numbersArray)

				for _, num := range numbersArray {
					sum += num.(float64)
				}

				average := 0.0
				if count > 0 {
					average = sum / float64(count)
				}

				return portal.NewFunctionData(map[string]any{
					"sum":     sum,
					"average": average,
					"count":   count,
				}), nil
			},
		},
	}

	// Register all functions
	for _, fn := range functions {
		_, err := httpPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register function %s: %v", fn.name, err)
		}
	}

	// Start HTTP server
	err := httpPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start HTTP portal: %v", err)
	}
	defer httpPortal.Stop(ctx)

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Test with multiple concurrent clients
	numClients := 5
	requestsPerClient := 3

	var wg sync.WaitGroup
	results := make(chan E2ETestResult, numClients*requestsPerClient*len(functions))

	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}
			baseURL := fmt.Sprintf("http://localhost:8081")

			for request := 0; request < requestsPerClient; request++ {
				// Test calculator function
				calcResult := testHTTPFunction(client, baseURL, "calculator_add", map[string]any{
					"a": float64(10 + id),
					"b": float64(5 + request),
				})
				calcResult.ClientID = id
				calcResult.RequestID = request
				results <- calcResult

				// Test string processor function
				stringResult := testHTTPFunction(client, baseURL, "string_processor", map[string]any{
					"text":      fmt.Sprintf("Hello from client %d request %d", id, request),
					"operation": "uppercase",
				})
				stringResult.ClientID = id
				stringResult.RequestID = request
				results <- stringResult

				// Test data aggregator function
				dataResult := testHTTPFunction(client, baseURL, "data_aggregator", map[string]any{
					"numbers": []any{float64(id), float64(request), float64(id + request)},
				})
				dataResult.ClientID = id
				dataResult.RequestID = request
				results <- dataResult
			}
		}(clientID)
	}

	wg.Wait()
	close(results)

	// Analyze results
	successCount := 0
	errorCount := 0
	functionCounts := make(map[string]int)

	for result := range results {
		if result.Error != nil {
			errorCount++
			t.Errorf("Client %d Request %d Function %s failed: %v",
				result.ClientID, result.RequestID, result.FunctionName, result.Error)
		} else {
			successCount++
			functionCounts[result.FunctionName]++

			// Validate specific function results
			switch result.FunctionName {
			case "calculator_add":
				if resultData, ok := result.Response["result"].(float64); ok {
					expectedResult := float64(10+result.ClientID) + float64(5+result.RequestID)
					if resultData != expectedResult {
						t.Errorf("Calculator result mismatch: expected %.2f, got %.2f", expectedResult, resultData)
					}
				}
			case "string_processor":
				if processed, ok := result.Response["processed"].(string); ok {
					expectedPrefix := "UPPER: Hello from client"
					if len(processed) < len(expectedPrefix) || processed[:len(expectedPrefix)] != expectedPrefix {
						t.Errorf("String processor result unexpected: %s", processed)
					}
				}
			case "data_aggregator":
				if count, ok := result.Response["count"].(float64); ok {
					if count != 3 {
						t.Errorf("Data aggregator count mismatch: expected 3, got %.0f", count)
					}
				}
			}
		}
	}

	expectedTotal := numClients * requestsPerClient * len(functions)
	if successCount+errorCount != expectedTotal {
		t.Errorf("Result count mismatch: expected %d, got %d", expectedTotal, successCount+errorCount)
	}

	if errorCount > 0 {
		t.Errorf("Had %d errors out of %d total requests", errorCount, expectedTotal)
	}

	t.Logf("E2E HTTP test completed: %d successful requests, %d errors", successCount, errorCount)
	t.Logf("Function call distribution: %v", functionCounts)
}

// E2E Test 2: Multiple HTTP Servers with Cross-Server Communication
func TestE2E_HTTPPortalMultipleServersWithCommunication(t *testing.T) {
	ctx := context.Background()

	// Create first HTTP portal (Math Service)
	mathConfig := &portal.HTTPConfig{
		Host:          "localhost",
		Port:          8082,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		ClientTimeout: 5 * time.Second,
	}
	mathPortal := portal.NewHTTPPortal(mathConfig)

	// Create second HTTP portal (String Service)
	stringConfig := &portal.HTTPConfig{
		Host:          "localhost",
		Port:          8083,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		ClientTimeout: 5 * time.Second,
	}
	stringPortal := portal.NewHTTPPortal(stringConfig)

	// Create third HTTP portal (Orchestrator Service)
	orchestratorConfig := &portal.HTTPConfig{
		Host:          "localhost",
		Port:          8084,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		ClientTimeout: 5 * time.Second,
	}
	orchestratorPortal := portal.NewHTTPPortal(orchestratorConfig)

	// Register math functions
	mathFunctions := []*E2ETestFunction{
		{
			name: "multiply",
			schema: builders.NewFunctionSchema().
				Name("multiply").
				Input("a", builders.NewNumberSchema().Build()).
				Input("b", builders.NewNumberSchema().Build()).
				RequiredInputs("a", "b").
				Output("result", builders.NewNumberSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				a, _ := params.Get("a")
				b, _ := params.Get("b")
				result := a.(float64) * b.(float64)
				return portal.NewFunctionData(map[string]any{"result": result}), nil
			},
		},
		{
			name: "power",
			schema: builders.NewFunctionSchema().
				Name("power").
				Input("base", builders.NewNumberSchema().Build()).
				Input("exponent", builders.NewNumberSchema().Build()).
				RequiredInputs("base", "exponent").
				Output("result", builders.NewNumberSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				base, _ := params.Get("base")
				exponent, _ := params.Get("exponent")

				// Simple power calculation
				result := 1.0
				baseVal := base.(float64)
				expVal := int(exponent.(float64))

				for i := 0; i < expVal; i++ {
					result *= baseVal
				}

				return portal.NewFunctionData(map[string]any{"result": result}), nil
			},
		},
	}

	// Register string functions
	stringFunctions := []*E2ETestFunction{
		{
			name: "format_number",
			schema: builders.NewFunctionSchema().
				Name("format_number").
				Input("number", builders.NewNumberSchema().Build()).
				Input("format", builders.NewStringSchema().Build()).
				RequiredInputs("number").
				Output("formatted", builders.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				number, _ := params.Get("number")
				format, hasFormat := params.Get("format")

				formatStr := "%.2f"
				if hasFormat {
					formatStr = format.(string)
				}

				formatted := fmt.Sprintf(formatStr, number.(float64))
				return portal.NewFunctionData(map[string]any{"formatted": formatted}), nil
			},
		},
		{
			name: "create_report",
			schema: builders.NewFunctionSchema().
				Name("create_report").
				Input("title", builders.NewStringSchema().Build()).
				Input("data", builders.NewArraySchema().Build()).
				RequiredInputs("title", "data").
				Output("report", builders.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				title, _ := params.Get("title")
				data, _ := params.Get("data")

				dataArray := data.([]any)
				report := fmt.Sprintf("=== %s ===\n", title.(string))
				for i, item := range dataArray {
					report += fmt.Sprintf("%d. %v\n", i+1, item)
				}
				report += fmt.Sprintf("Total items: %d", len(dataArray))

				return portal.NewFunctionData(map[string]any{"report": report}), nil
			},
		},
	}

	// Register orchestrator function that calls other services
	orchestratorFunctions := []*E2ETestFunction{
		{
			name: "complex_calculation",
			schema: builders.NewFunctionSchema().
				Name("complex_calculation").
				Input("numbers", builders.NewArraySchema().Items(builders.NewNumberSchema().Build()).Build()).
				RequiredInputs("numbers").
				Output("calculation_report", builders.NewStringSchema().Build()).
				Output("total_operations", builders.NewIntegerSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				numbers, _ := params.Get("numbers")
				numbersArray := numbers.([]any)

				client := &http.Client{Timeout: 5 * time.Second}
				results := []string{}
				operationCount := 0

				// Perform calculations using math service
				for i, num := range numbersArray {
					// Call multiply function
					multiplyResult, err := callHTTPFunction(client, "http://localhost:8082", "multiply", map[string]any{
						"a": num,
						"b": 2.0,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to call multiply: %v", err)
					}
					operationCount++

					// Call power function
					powerResult, err := callHTTPFunction(client, "http://localhost:8082", "power", map[string]any{
						"base":     num,
						"exponent": 2.0,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to call power: %v", err)
					}
					operationCount++

					// Format results using string service
					multiplyFormatted, err := callHTTPFunction(client, "http://localhost:8083", "format_number", map[string]any{
						"number": multiplyResult["result"],
						"format": "%.1f",
					})
					if err != nil {
						return nil, fmt.Errorf("failed to format multiply result: %v", err)
					}
					operationCount++

					powerFormatted, err := callHTTPFunction(client, "http://localhost:8083", "format_number", map[string]any{
						"number": powerResult["result"],
						"format": "%.1f",
					})
					if err != nil {
						return nil, fmt.Errorf("failed to format power result: %v", err)
					}
					operationCount++

					results = append(results, fmt.Sprintf("Number %d: %.1f * 2 = %s, %.1f^2 = %s",
						i+1, num.(float64), multiplyFormatted["formatted"], num.(float64), powerFormatted["formatted"]))
				}

				// Create final report using string service
				reportResult, err := callHTTPFunction(client, "http://localhost:8083", "create_report", map[string]any{
					"title": "Complex Calculation Results",
					"data":  results,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to create report: %v", err)
				}
				operationCount++

				return portal.NewFunctionData(map[string]any{
					"calculation_report": reportResult["report"],
					"total_operations":   operationCount,
				}), nil
			},
		},
	}

	// Register functions with their respective portals
	for _, fn := range mathFunctions {
		_, err := mathPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register math function %s: %v", fn.name, err)
		}
	}

	for _, fn := range stringFunctions {
		_, err := stringPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register string function %s: %v", fn.name, err)
		}
	}

	for _, fn := range orchestratorFunctions {
		_, err := orchestratorPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register orchestrator function %s: %v", fn.name, err)
		}
	}

	// Start all servers
	err := mathPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start math portal: %v", err)
	}
	defer mathPortal.Stop(ctx)

	err = stringPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start string portal: %v", err)
	}
	defer stringPortal.Stop(ctx)

	err = orchestratorPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start orchestrator portal: %v", err)
	}
	defer orchestratorPortal.Stop(ctx)

	// Wait for all servers to be ready
	time.Sleep(200 * time.Millisecond)

	// Test direct service calls first
	client := &http.Client{Timeout: 10 * time.Second}

	// Test math service
	mathResult := testHTTPFunction(client, "http://localhost:8082", "multiply", map[string]any{
		"a": 6.0,
		"b": 7.0,
	})
	if mathResult.Error != nil {
		t.Fatalf("Math service test failed: %v", mathResult.Error)
	}
	if mathResult.Response["result"].(float64) != 42.0 {
		t.Errorf("Math result incorrect: expected 42, got %v", mathResult.Response["result"])
	}

	// Test string service
	stringResult := testHTTPFunction(client, "http://localhost:8083", "format_number", map[string]any{
		"number": 123.456,
		"format": "%.1f",
	})
	if stringResult.Error != nil {
		t.Fatalf("String service test failed: %v", stringResult.Error)
	}
	if stringResult.Response["formatted"].(string) != "123.5" {
		t.Errorf("String result incorrect: expected '123.5', got %v", stringResult.Response["formatted"])
	}

	// Test complex orchestrator function (cross-service communication)
	orchestratorResult := testHTTPFunction(client, "http://localhost:8084", "complex_calculation", map[string]any{
		"numbers": []any{3.0, 4.0, 5.0},
	})
	if orchestratorResult.Error != nil {
		t.Fatalf("Orchestrator service test failed: %v", orchestratorResult.Error)
	}

	// Validate orchestrator results
	if report, ok := orchestratorResult.Response["calculation_report"].(string); ok {
		if len(report) < 50 { // Basic sanity check
			t.Errorf("Report seems too short: %s", report)
		}
		t.Logf("Generated report:\n%s", report)
	} else {
		t.Error("Expected calculation_report in response")
	}

	if operations, ok := orchestratorResult.Response["total_operations"].(float64); ok {
		expectedOps := 3*4 + 1 // 4 operations per number + 1 final report
		if int(operations) != expectedOps {
			t.Errorf("Expected %d operations, got %.0f", expectedOps, operations)
		}
	} else {
		t.Error("Expected total_operations in response")
	}

	t.Logf("Multi-server E2E test completed successfully")
}

// E2E Test 3: Load Testing with Concurrent Requests
func TestE2E_HTTPPortalLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ctx := context.Background()

	// Create HTTP portal for load testing
	loadConfig := &portal.HTTPConfig{
		Host:          "localhost",
		Port:          8085,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		ClientTimeout: 10 * time.Second,
	}
	loadPortal := portal.NewHTTPPortal(loadConfig)

	// Register a CPU-intensive function
	cpuFunction := &E2ETestFunction{
		name: "fibonacci",
		schema: builders.NewFunctionSchema().
			Name("fibonacci").
			Input("n", builders.NewIntegerSchema().Build()).
			RequiredInputs("n").
			Output("result", builders.NewIntegerSchema().Build()).
			Output("computed_in_ms", builders.NewIntegerSchema().Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			start := time.Now()
			n, _ := params.Get("n")
			nInt := int(n.(float64))

			// Limit to reasonable values for testing
			if nInt > 40 {
				nInt = 40
			}

			result := fibonacci(nInt)
			duration := time.Since(start).Milliseconds()

			return portal.NewFunctionData(map[string]any{
				"result":         result,
				"computed_in_ms": duration,
			}), nil
		},
	}

	_, err := loadPortal.Apply(ctx, cpuFunction)
	if err != nil {
		t.Fatalf("Failed to register fibonacci function: %v", err)
	}

	// Start server
	err = loadPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start load portal: %v", err)
	}
	defer loadPortal.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	// Load test parameters
	numClients := 10
	requestsPerClient := 20

	var wg sync.WaitGroup
	results := make(chan E2ETestResult, numClients*requestsPerClient)
	startTime := time.Now()

	// Launch concurrent clients
	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client := &http.Client{Timeout: 15 * time.Second}
			baseURL := "http://localhost:8085"

			for request := 0; request < requestsPerClient; request++ {
				// Vary the fibonacci number to create different load patterns
				fibNumber := 20 + (id*requestsPerClient+request)%15

				result := testHTTPFunction(client, baseURL, "fibonacci", map[string]any{
					"n": float64(fibNumber),
				})
				result.ClientID = id
				result.RequestID = request
				result.StartTime = time.Now()
				results <- result
			}
		}(clientID)
	}

	wg.Wait()
	close(results)
	totalDuration := time.Since(startTime)

	// Analyze load test results
	successCount := 0
	errorCount := 0
	totalResponseTime := time.Duration(0)
	maxResponseTime := time.Duration(0)
	minResponseTime := time.Hour // Start with a large value

	for result := range results {
		if result.Error != nil {
			errorCount++
			t.Logf("Load test error from client %d: %v", result.ClientID, result.Error)
		} else {
			successCount++

			// Calculate response time (this is approximate since we don't have precise timing)
			if computeTime, ok := result.Response["computed_in_ms"].(float64); ok {
				responseTime := time.Duration(computeTime) * time.Millisecond
				totalResponseTime += responseTime

				if responseTime > maxResponseTime {
					maxResponseTime = responseTime
				}
				if responseTime < minResponseTime {
					minResponseTime = responseTime
				}
			}
		}
	}

	totalRequests := numClients * requestsPerClient
	successRate := float64(successCount) / float64(totalRequests) * 100
	avgResponseTime := totalResponseTime / time.Duration(successCount)
	requestsPerSecond := float64(totalRequests) / totalDuration.Seconds()

	// Validate load test results
	if successRate < 95.0 {
		t.Errorf("Success rate too low: %.2f%% (expected >= 95%%)", successRate)
	}

	if avgResponseTime > 5*time.Second {
		t.Errorf("Average response time too high: %v (expected < 5s)", avgResponseTime)
	}

	if requestsPerSecond < 1.0 {
		t.Errorf("Throughput too low: %.2f req/s (expected >= 1 req/s)", requestsPerSecond)
	}

	t.Logf("Load test results:")
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Successful: %d (%.2f%%)", successCount, successRate)
	t.Logf("  Failed: %d", errorCount)
	t.Logf("  Total duration: %v", totalDuration)
	t.Logf("  Requests/second: %.2f", requestsPerSecond)
	t.Logf("  Avg response time: %v", avgResponseTime)
	t.Logf("  Min response time: %v", minResponseTime)
	t.Logf("  Max response time: %v", maxResponseTime)
}

// Helper types and functions

type E2ETestFunction struct {
	name    string
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (f *E2ETestFunction) Name() string {
	return f.name
}

func (f *E2ETestFunction) Schema() core.FunctionSchema {
	return f.schema
}

func (f *E2ETestFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler(ctx, params)
}

type E2ETestResult struct {
	ClientID     int
	RequestID    int
	FunctionName string
	Response     map[string]any
	Error        error
	StartTime    time.Time
}

func testHTTPFunction(client *http.Client, baseURL, functionName string, params map[string]any) E2ETestResult {
	result := E2ETestResult{
		FunctionName: functionName,
		StartTime:    time.Now(),
	}

	response, err := callHTTPFunction(client, baseURL, functionName, params)
	if err != nil {
		result.Error = err
		return result
	}

	result.Response = response
	return result
}

func callHTTPFunction(client *http.Client, baseURL, functionName string, params map[string]any) (map[string]any, error) {
	url := fmt.Sprintf("%s/functions/%s", baseURL, functionName)

	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var response struct {
		Result map[string]any `json:"result"`
		Error  any            `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("function error: %v", response.Error)
	}

	return response.Result, nil
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
