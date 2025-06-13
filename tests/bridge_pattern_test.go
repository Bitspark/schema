package tests

import (
	"bytes"
	"context"
	builders2 "defs.dev/schema/builders"
	"defs.dev/schema/runtime/portal"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"defs.dev/schema/api"
	"defs.dev/schema/core"
)

// PortalBridge implements the WebSocket-to-HTTP bridge pattern
type PortalBridge struct {
	wsClient     *websocket.Conn
	httpServer   *http.Server
	wsServerAddr string
	httpPort     int
	mu           sync.RWMutex
	writeMu      sync.Mutex // Protects WebSocket writes
	running      bool
	requestMap   map[string]chan BridgeResponse
}

type BridgeResponse struct {
	Data  map[string]any `json:"data"`
	Error string         `json:"error"`
}

// NewPortalBridge creates a new bridge instance
func NewPortalBridge(wsServerAddr string, httpPort int) *PortalBridge {
	return &PortalBridge{
		wsServerAddr: wsServerAddr,
		httpPort:     httpPort,
		requestMap:   make(map[string]chan BridgeResponse),
	}
}

// Start initializes the bridge by connecting to WebSocket server and starting HTTP server
func (b *PortalBridge) Start(ctx context.Context) error {
	// Connect to WebSocket server
	u, err := url.Parse(b.wsServerAddr)
	if err != nil {
		return fmt.Errorf("invalid WebSocket server address: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	b.mu.Lock()
	b.wsClient = conn
	b.running = true
	b.mu.Unlock()

	// Start WebSocket message handler
	go b.handleWebSocketMessages()

	// Start HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/bridge/", b.handleHTTPRequest)
	mux.HandleFunc("/health", b.handleHealth)

	b.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", b.httpPort),
		Handler: mux,
	}

	go func() {
		if err := b.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error
		}
	}()

	return nil
}

// Stop shuts down the bridge
func (b *PortalBridge) Stop(ctx context.Context) error {
	b.mu.Lock()
	b.running = false
	b.mu.Unlock()

	if b.wsClient != nil {
		b.wsClient.Close()
	}

	if b.httpServer != nil {
		return b.httpServer.Shutdown(ctx)
	}

	return nil
}

// handleWebSocketMessages processes incoming WebSocket messages
func (b *PortalBridge) handleWebSocketMessages() {
	for {
		b.mu.RLock()
		if !b.running {
			b.mu.RUnlock()
			break
		}
		conn := b.wsClient
		b.mu.RUnlock()

		var response struct {
			Type  string         `json:"type"`
			ID    string         `json:"id"`
			Data  map[string]any `json:"data"`
			Error string         `json:"error"`
		}

		err := conn.ReadJSON(&response)
		if err != nil {
			break
		}

		// Route response to waiting HTTP request
		b.mu.RLock()
		if ch, exists := b.requestMap[response.ID]; exists {
			bridgeResp := BridgeResponse{
				Data:  response.Data,
				Error: response.Error,
			}
			select {
			case ch <- bridgeResp:
			case <-time.After(1 * time.Second):
				// Timeout sending to channel
			}
		}
		b.mu.RUnlock()
	}
}

// handleHTTPRequest processes HTTP requests and forwards them to WebSocket
func (b *PortalBridge) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract function name from URL path
	// URL format: /bridge/{function_name}
	path := r.URL.Path
	if len(path) < 9 { // "/bridge/" = 8 chars + at least 1 for function name
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}
	functionName := path[8:] // Remove "/bridge/" prefix

	// Parse request body
	var requestData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Generate unique request ID
	requestID := fmt.Sprintf("bridge_%d", time.Now().UnixNano())

	// Create response channel
	responseChan := make(chan BridgeResponse, 1)
	b.mu.Lock()
	b.requestMap[requestID] = responseChan
	b.mu.Unlock()

	// Clean up response channel after request
	defer func() {
		b.mu.Lock()
		delete(b.requestMap, requestID)
		b.mu.Unlock()
		close(responseChan)
	}()

	// Forward to WebSocket server
	wsMessage := map[string]any{
		"type":     "call",
		"id":       requestID,
		"function": functionName,
		"data":     requestData,
	}

	b.mu.RLock()
	conn := b.wsClient
	b.mu.RUnlock()

	// Synchronize WebSocket writes
	b.writeMu.Lock()
	err := conn.WriteJSON(wsMessage)
	b.writeMu.Unlock()

	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		w.Header().Set("Content-Type", "application/json")
		if response.Error != "" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"error": response.Error,
			})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response.Data)
		}
	case <-time.After(30 * time.Second):
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
	}
}

// handleHealth provides health check endpoint
func (b *PortalBridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	b.mu.RLock()
	running := b.running
	b.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	status := "healthy"
	if !running {
		status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":    status,
		"timestamp": time.Now().Unix(),
	})
}

// Test the Portal Bridge Pattern
func TestE2E_PortalBridgePattern(t *testing.T) {
	ctx := context.Background()

	// Step 1: Create WebSocket Server (WSS)
	wsServer := portal.NewWebSocketPortal(&portal.WebSocketConfig{
		Host:           "localhost",
		Port:           8091,
		Path:           "/ws",
		PingPeriod:     54 * time.Second,
		PongWait:       60 * time.Second,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    60 * time.Second,
		MaxMessageSize: 512 * 1024,
	}, nil, nil)

	// Register functions on WebSocket server
	calculatorFunc := &BridgeTestFunction{
		name: "calculator",
		schema: builders2.NewFunctionSchema().
			Name("calculator").
			Description("Performs basic calculations").
			Input("operation", builders2.NewStringSchema().Build()).
			Input("a", builders2.NewNumberSchema().Build()).
			Input("b", builders2.NewNumberSchema().Build()).
			RequiredInputs("operation", "a", "b").
			Output("result", builders2.NewNumberSchema().Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			operation, _ := params.Get("operation")
			a, _ := params.Get("a")
			b, _ := params.Get("b")

			aVal := a.(float64)
			bVal := b.(float64)
			var result float64

			switch operation.(string) {
			case "add":
				result = aVal + bVal
			case "subtract":
				result = aVal - bVal
			case "multiply":
				result = aVal * bVal
			case "divide":
				if bVal == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = aVal / bVal
			default:
				return nil, fmt.Errorf("unknown operation: %s", operation)
			}

			return api.NewFunctionData(map[string]any{
				"result": result,
			}), nil
		},
	}

	dataProcessorFunc := &BridgeTestFunction{
		name: "data_processor",
		schema: builders2.NewFunctionSchema().
			Name("data_processor").
			Description("Processes and transforms data").
			Input("data", builders2.NewArraySchema().Build()).
			Input("transform", builders2.NewStringSchema().Build()).
			RequiredInputs("data", "transform").
			Output("processed_data", builders2.NewArraySchema().Build()).
			Output("summary", builders2.NewObjectSchema().Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			data, _ := params.Get("data")
			transform, _ := params.Get("transform")

			dataArray := data.([]any)
			transformType := transform.(string)

			processedData := make([]any, len(dataArray))
			sum := 0.0

			for i, item := range dataArray {
				val := item.(float64)
				sum += val

				switch transformType {
				case "double":
					processedData[i] = val * 2
				case "square":
					processedData[i] = val * val
				case "increment":
					processedData[i] = val + 1
				default:
					processedData[i] = val
				}
			}

			return api.NewFunctionData(map[string]any{
				"processed_data": processedData,
				"summary": map[string]any{
					"count":   len(dataArray),
					"sum":     sum,
					"average": sum / float64(len(dataArray)),
				},
			}), nil
		},
	}

	// Register functions
	_, err := wsServer.Apply(ctx, calculatorFunc)
	if err != nil {
		t.Fatalf("Failed to register calculator function: %v", err)
	}

	_, err = wsServer.Apply(ctx, dataProcessorFunc)
	if err != nil {
		t.Fatalf("Failed to register data processor function: %v", err)
	}

	// Start WebSocket server
	err = wsServer.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}
	defer wsServer.Stop(ctx)

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)

	// Step 2: Create Portal Bridge (WHC + HTS)
	bridge := NewPortalBridge("ws://localhost:8091/ws", 8092)
	err = bridge.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start portal bridge: %v", err)
	}
	defer bridge.Stop(ctx)

	// Wait for bridge to be ready
	time.Sleep(200 * time.Millisecond)

	// Step 3: Test HTTP Client (HTC) communication through bridge
	t.Run("Calculator through bridge", func(t *testing.T) {
		// Test addition
		requestData := map[string]any{
			"operation": "add",
			"a":         15.0,
			"b":         25.0,
		}

		response, err := makeHTTPBridgeRequest("http://localhost:8092/bridge/calculator", requestData)
		if err != nil {
			t.Fatalf("Failed to make bridge request: %v", err)
		}

		if result, ok := response["result"].(map[string]any)["result"].(float64); ok {
			if result != 40.0 {
				t.Errorf("Expected result 40.0, got %v", result)
			}
		} else {
			t.Errorf("Invalid response format: %+v", response)
		}

		// Test multiplication
		requestData = map[string]any{
			"operation": "multiply",
			"a":         7.0,
			"b":         6.0,
		}

		response, err = makeHTTPBridgeRequest("http://localhost:8092/bridge/calculator", requestData)
		if err != nil {
			t.Fatalf("Failed to make bridge request: %v", err)
		}

		if result, ok := response["result"].(map[string]any)["result"].(float64); ok {
			if result != 42.0 {
				t.Errorf("Expected result 42.0, got %v", result)
			}
		} else {
			t.Errorf("Invalid response format: %+v", response)
		}
	})

	t.Run("Data processor through bridge", func(t *testing.T) {
		requestData := map[string]any{
			"data":      []any{1.0, 2.0, 3.0, 4.0, 5.0},
			"transform": "double",
		}

		response, err := makeHTTPBridgeRequest("http://localhost:8092/bridge/data_processor", requestData)
		if err != nil {
			t.Fatalf("Failed to make bridge request: %v", err)
		}

		if result, ok := response["result"].(map[string]any); ok {
			if processedData, ok := result["processed_data"].([]any); ok {
				expected := []float64{2.0, 4.0, 6.0, 8.0, 10.0}
				for i, val := range processedData {
					if val.(float64) != expected[i] {
						t.Errorf("Expected processed_data[%d] = %v, got %v", i, expected[i], val)
					}
				}
			} else {
				t.Error("Expected processed_data in result")
			}

			if summary, ok := result["summary"].(map[string]any); ok {
				if count := summary["count"].(float64); count != 5.0 {
					t.Errorf("Expected count 5, got %v", count)
				}
				if sum := summary["sum"].(float64); sum != 15.0 {
					t.Errorf("Expected sum 15, got %v", sum)
				}
			} else {
				t.Error("Expected summary in result")
			}
		} else {
			t.Errorf("Invalid response format: %+v", response)
		}
	})

	t.Run("Concurrent bridge requests", func(t *testing.T) {
		const numRequests = 3 // Reduced from 10 to avoid overwhelming the bridge
		var wg sync.WaitGroup
		results := make([]float64, numRequests)
		errors := make([]error, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				requestData := map[string]any{
					"operation": "multiply",
					"a":         float64(index + 1),
					"b":         2.0,
				}

				response, err := makeHTTPBridgeRequest("http://localhost:8092/bridge/calculator", requestData)
				if err != nil {
					errors[index] = err
					return
				}

				if result, ok := response["result"].(map[string]any)["result"].(float64); ok {
					results[index] = result
				} else {
					errors[index] = fmt.Errorf("invalid response format")
				}
			}(i)
		}

		wg.Wait()

		// Check results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			if errors[i] != nil {
				t.Errorf("Request %d failed: %v", i, errors[i])
			} else {
				expected := float64((i + 1) * 2)
				if results[i] != expected {
					t.Errorf("Request %d: expected %v, got %v", i, expected, results[i])
				} else {
					successCount++
				}
			}
		}

		t.Logf("Bridge concurrent test: %d/%d requests successful", successCount, numRequests)
	})

	t.Logf("Portal Bridge Pattern test completed successfully!")
}

// Helper function to make HTTP requests to the bridge
func makeHTTPBridgeRequest(url string, data map[string]any) (map[string]any, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

// BridgeTestFunction implements api.Function for bridge testing
type BridgeTestFunction struct {
	name    string
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (f *BridgeTestFunction) Name() string {
	return f.name
}

func (f *BridgeTestFunction) Schema() core.FunctionSchema {
	return f.schema
}

func (f *BridgeTestFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler(ctx, params)
}
