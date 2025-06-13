package tests

import (
	"context"
	builders2 "defs.dev/schema/builders"
	portal2 "defs.dev/schema/runtime/portal"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"defs.dev/schema/api"
)

// E2E Test 1: WebSocket Portal with Real-time Communication
func TestE2E_WebSocketPortalRealTimeCommunication(t *testing.T) {
	ctx := context.Background()

	// Create WebSocket portal
	wsConfig := &portal2.WebSocketConfig{
		Host:           "localhost",
		Port:           8086,
		Path:           "/ws",
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		PingPeriod:     10 * time.Second,
		PongWait:       60 * time.Second,
		MaxMessageSize: 1024 * 1024,
	}
	wsPortal := portal2.NewWebSocketPortal(wsConfig, nil, nil)

	// Register streaming functions
	streamingFunctions := []*E2ETestFunction{
		{
			name: "counter_stream",
			schema: builders2.NewFunctionSchema().
				Name("counter_stream").
				Description("Streams counting numbers").
				Input("start", builders2.NewIntegerSchema().Build()).
				Input("end", builders2.NewIntegerSchema().Build()).
				Input("interval_ms", builders2.NewIntegerSchema().Build()).
				RequiredInputs("start", "end").
				Output("count", builders2.NewIntegerSchema().Build()).
				Output("timestamp", builders2.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				start, _ := params.Get("start")
				end, _ := params.Get("end")
				intervalMs, hasInterval := params.Get("interval_ms")

				startInt := int(start.(float64))
				endInt := int(end.(float64))
				interval := 100 * time.Millisecond

				if hasInterval {
					interval = time.Duration(int(intervalMs.(float64))) * time.Millisecond
				}

				// For WebSocket streaming, we'll simulate by returning the final count
				// In a real implementation, this would stream multiple responses
				time.Sleep(interval * time.Duration(endInt-startInt))

				return api.NewFunctionData(map[string]any{
					"count":     endInt,
					"timestamp": time.Now().Format(time.RFC3339),
				}), nil
			},
		},
		{
			name: "chat_processor",
			schema: builders2.NewFunctionSchema().
				Name("chat_processor").
				Description("Processes chat messages").
				Input("message", builders2.NewStringSchema().Build()).
				Input("user_id", builders2.NewStringSchema().Build()).
				RequiredInputs("message", "user_id").
				Output("processed_message", builders2.NewStringSchema().Build()).
				Output("word_count", builders2.NewIntegerSchema().Build()).
				Output("timestamp", builders2.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				message, _ := params.Get("message")
				userID, _ := params.Get("user_id")

				messageStr := message.(string)
				words := len(strings.Fields(messageStr))

				processedMessage := fmt.Sprintf("[%s]: %s", userID.(string), messageStr)

				return api.NewFunctionData(map[string]any{
					"processed_message": processedMessage,
					"word_count":        words,
					"timestamp":         time.Now().Format(time.RFC3339),
				}), nil
			},
		},
	}

	// Register functions
	for _, fn := range streamingFunctions {
		_, err := wsPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register WebSocket function %s: %v", fn.name, err)
		}
	}

	// Start WebSocket server
	err := wsPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start WebSocket portal: %v", err)
	}
	defer wsPortal.Stop(ctx)

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)

	// Test with multiple WebSocket clients
	numClients := 3
	messagesPerClient := 5

	var wg sync.WaitGroup
	results := make(chan E2EWebSocketResult, numClients*messagesPerClient*2)

	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Connect to WebSocket
			u := url.URL{Scheme: "ws", Host: "localhost:8086", Path: "/ws"}
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				results <- E2EWebSocketResult{
					ClientID: id,
					Error:    fmt.Errorf("failed to connect: %v", err),
				}
				return
			}
			defer conn.Close()

			for message := 0; message < messagesPerClient; message++ {
				// Test counter stream
				counterRequest := map[string]any{
					"type":     "call",
					"id":       fmt.Sprintf("counter_%d_%d", id, message),
					"function": "counter_stream",
					"data": map[string]any{
						"start":       id * 10,
						"end":         id*10 + message + 1,
						"interval_ms": 50,
					},
				}

				result := testWebSocketFunction(conn, counterRequest)
				result.ClientID = id
				result.MessageID = message
				result.FunctionName = "counter_stream"
				results <- result

				// Test chat processor
				chatRequest := map[string]any{
					"type":     "call",
					"id":       fmt.Sprintf("chat_%d_%d", id, message),
					"function": "chat_processor",
					"data": map[string]any{
						"message": fmt.Sprintf("Hello from client %d, message %d", id, message),
						"user_id": fmt.Sprintf("user_%d", id),
					},
				}

				result = testWebSocketFunction(conn, chatRequest)
				result.ClientID = id
				result.MessageID = message
				result.FunctionName = "chat_processor"
				results <- result
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
			t.Errorf("WebSocket client %d message %d function %s failed: %v",
				result.ClientID, result.MessageID, result.FunctionName, result.Error)
		} else {
			successCount++
			functionCounts[result.FunctionName]++

			// Validate specific function results
			switch result.FunctionName {
			case "counter_stream":
				if count, ok := result.Response["count"].(float64); ok {
					expectedCount := result.ClientID*10 + result.MessageID + 1
					if int(count) != expectedCount {
						t.Errorf("Counter result mismatch: expected %d, got %.0f", expectedCount, count)
					}
				}
			case "chat_processor":
				if processed, ok := result.Response["processed_message"].(string); ok {
					expectedPrefix := fmt.Sprintf("[user_%d]:", result.ClientID)
					if len(processed) < len(expectedPrefix) || processed[:len(expectedPrefix)] != expectedPrefix {
						t.Errorf("Chat processor result unexpected: %s", processed)
					}
				}
			}
		}
	}

	expectedTotal := numClients * messagesPerClient * 2
	if successCount+errorCount != expectedTotal {
		t.Errorf("Result count mismatch: expected %d, got %d", expectedTotal, successCount+errorCount)
	}

	if errorCount > 0 {
		t.Errorf("Had %d errors out of %d total WebSocket messages", errorCount, expectedTotal)
	}

	t.Logf("WebSocket E2E test completed: %d successful messages, %d errors", successCount, errorCount)
	t.Logf("Function call distribution: %v", functionCounts)
}

// E2E Test 2: Mixed Portal Communication (HTTP + WebSocket)
func TestE2E_MixedPortalCommunication(t *testing.T) {
	ctx := context.Background()

	// Create HTTP portal for API services
	httpConfig := &portal2.HTTPConfig{
		Host:          "localhost",
		Port:          8087,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		ClientTimeout: 5 * time.Second,
	}
	httpPortal := portal2.NewHTTPPortal(httpConfig)

	// Create WebSocket portal for real-time services
	wsConfig := &portal2.WebSocketConfig{
		Host:           "localhost",
		Port:           8088,
		Path:           "/ws",
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		PingPeriod:     10 * time.Second,
		PongWait:       60 * time.Second,
		MaxMessageSize: 1024 * 1024,
	}
	wsPortal := portal2.NewWebSocketPortal(wsConfig, nil, nil)

	// Register HTTP functions (stateless operations)
	httpFunctions := []*E2ETestFunction{
		{
			name: "user_validator",
			schema: builders2.NewFunctionSchema().
				Name("user_validator").
				Input("username", builders2.NewStringSchema().Build()).
				Input("email", builders2.NewStringSchema().Build()).
				RequiredInputs("username", "email").
				Output("valid", builders2.NewBooleanSchema().Build()).
				Output("errors", builders2.NewArraySchema().Items(builders2.NewStringSchema().Build()).Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				username, _ := params.Get("username")
				email, _ := params.Get("email")

				errors := []string{}
				usernameStr := username.(string)
				emailStr := email.(string)

				if len(usernameStr) < 3 {
					errors = append(errors, "Username must be at least 3 characters")
				}
				if !strings.Contains(emailStr, "@") {
					errors = append(errors, "Email must contain @ symbol")
				}

				return api.NewFunctionData(map[string]any{
					"valid":  len(errors) == 0,
					"errors": errors,
				}), nil
			},
		},
		{
			name: "data_transformer",
			schema: builders2.NewFunctionSchema().
				Name("data_transformer").
				Input("data", builders2.NewObjectSchema().Build()).
				Input("transform_type", builders2.NewStringSchema().Build()).
				RequiredInputs("data", "transform_type").
				Output("transformed", builders2.NewObjectSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				data, _ := params.Get("data")
				transformType, _ := params.Get("transform_type")

				dataMap := data.(map[string]any)
				transformed := make(map[string]any)

				switch transformType.(string) {
				case "uppercase_keys":
					for k, v := range dataMap {
						transformed[strings.ToUpper(k)] = v
					}
				case "add_metadata":
					for k, v := range dataMap {
						transformed[k] = v
					}
					transformed["_metadata"] = map[string]any{
						"transformed_at": time.Now().Format(time.RFC3339),
						"transform_type": transformType,
					}
				default:
					transformed = dataMap
				}

				return api.NewFunctionData(map[string]any{
					"transformed": transformed,
				}), nil
			},
		},
	}

	// Register WebSocket functions (stateful/streaming operations)
	wsFunctions := []*E2ETestFunction{
		{
			name: "notification_sender",
			schema: builders2.NewFunctionSchema().
				Name("notification_sender").
				Input("user_id", builders2.NewStringSchema().Build()).
				Input("message", builders2.NewStringSchema().Build()).
				Input("priority", builders2.NewStringSchema().Build()).
				RequiredInputs("user_id", "message").
				Output("notification_id", builders2.NewStringSchema().Build()).
				Output("sent_at", builders2.NewStringSchema().Build()).
				Build(),
			handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
				userID, _ := params.Get("user_id")
				_, _ = params.Get("message") // Message is required but not used in this simple implementation
				priority, hasPriority := params.Get("priority")

				priorityStr := "normal"
				if hasPriority {
					priorityStr = priority.(string)
				}

				notificationID := fmt.Sprintf("notif_%s_%d", userID.(string), time.Now().Unix())

				// Simulate notification sending delay based on priority
				switch priorityStr {
				case "high":
					time.Sleep(10 * time.Millisecond)
				case "normal":
					time.Sleep(50 * time.Millisecond)
				case "low":
					time.Sleep(100 * time.Millisecond)
				}

				return api.NewFunctionData(map[string]any{
					"notification_id": notificationID,
					"sent_at":         time.Now().Format(time.RFC3339),
				}), nil
			},
		},
	}

	// Register functions with their respective portals
	for _, fn := range httpFunctions {
		_, err := httpPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register HTTP function %s: %v", fn.name, err)
		}
	}

	for _, fn := range wsFunctions {
		_, err := wsPortal.Apply(ctx, fn)
		if err != nil {
			t.Fatalf("Failed to register WebSocket function %s: %v", fn.name, err)
		}
	}

	// Start both portals
	err := httpPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start HTTP portal: %v", err)
	}
	defer httpPortal.Stop(ctx)

	err = wsPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start WebSocket portal: %v", err)
	}
	defer wsPortal.Stop(ctx)

	// Wait for servers to be ready
	time.Sleep(200 * time.Millisecond)

	// Test mixed portal workflow
	client := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Validate user data via HTTP
	validationResult := testHTTPFunction(client, "http://localhost:8087", "user_validator", map[string]any{
		"username": "testuser",
		"email":    "test@example.com",
	})
	if validationResult.Error != nil {
		t.Fatalf("User validation failed: %v", validationResult.Error)
	}

	if !validationResult.Response["valid"].(bool) {
		t.Fatalf("User validation should have passed: %v", validationResult.Response["errors"])
	}

	// Step 2: Transform user data via HTTP
	transformResult := testHTTPFunction(client, "http://localhost:8087", "data_transformer", map[string]any{
		"data": map[string]any{
			"username": "testuser",
			"email":    "test@example.com",
			"status":   "active",
		},
		"transform_type": "add_metadata",
	})
	if transformResult.Error != nil {
		t.Fatalf("Data transformation failed: %v", transformResult.Error)
	}

	transformedData := transformResult.Response["transformed"].(map[string]any)
	if _, hasMetadata := transformedData["_metadata"]; !hasMetadata {
		t.Error("Expected metadata to be added to transformed data")
	}

	// Step 3: Send notification via WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8088", Path: "/ws"}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsConn.Close()

	notificationRequest := map[string]any{
		"type":     "call",
		"id":       "notification_test",
		"function": "notification_sender",
		"data": map[string]any{
			"user_id":  "testuser",
			"message":  "Welcome! Your account has been validated and processed.",
			"priority": "high",
		},
	}

	notificationResult := testWebSocketFunction(wsConn, notificationRequest)
	if notificationResult.Error != nil {
		t.Fatalf("Notification sending failed: %v", notificationResult.Error)
	}

	if notificationID, ok := notificationResult.Response["notification_id"].(string); ok {
		if !strings.HasPrefix(notificationID, "notif_testuser_") {
			t.Errorf("Unexpected notification ID format: %s", notificationID)
		}
	} else {
		t.Error("Expected notification_id in response")
	}

	t.Logf("Mixed portal E2E test completed successfully")
	t.Logf("Validation result: %v", validationResult.Response)
	t.Logf("Transform result keys: %v", getKeys(transformedData))
	t.Logf("Notification result: %v", notificationResult.Response)
}

// Helper types and functions for WebSocket tests

type E2EWebSocketResult struct {
	ClientID     int
	MessageID    int
	FunctionName string
	Response     map[string]any
	Error        error
}

func testWebSocketFunction(conn *websocket.Conn, request map[string]any) E2EWebSocketResult {
	result := E2EWebSocketResult{}

	// Send request
	if err := conn.WriteJSON(request); err != nil {
		result.Error = fmt.Errorf("failed to send WebSocket message: %v", err)
		return result
	}

	// Read response
	var response struct {
		Type  string         `json:"type"`
		ID    string         `json:"id"`
		Data  map[string]any `json:"data"`
		Error string         `json:"error"`
	}

	if err := conn.ReadJSON(&response); err != nil {
		result.Error = fmt.Errorf("failed to read WebSocket response: %v", err)
		return result
	}

	if response.Type == "error" || response.Error != "" {
		result.Error = fmt.Errorf("function error: %v", response.Error)
		return result
	}

	if response.Type != "response" {
		result.Error = fmt.Errorf("unexpected response type: %s", response.Type)
		return result
	}

	// Extract the actual result from the data field
	if resultData, ok := response.Data["result"]; ok {
		if resultMap, ok := resultData.(map[string]any); ok {
			result.Response = resultMap
		} else {
			result.Response = map[string]any{"result": resultData}
		}
	} else {
		result.Response = response.Data
	}
	return result
}

func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
